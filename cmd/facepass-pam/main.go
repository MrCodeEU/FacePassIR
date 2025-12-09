package main

import (
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/MrCodeEU/facepass/pkg/config"
	"github.com/MrCodeEU/facepass/pkg/logging"
	"github.com/MrCodeEU/facepass/pkg/pam"
)

const version = "0.2.0"

func main() {
	// PAM module entry point
	// This binary is called by PAM during authentication
	// Exit codes:
	//   0 = authentication successful
	//   1 = authentication failed
	//   2 = user not enrolled (fallback to password)
	//   3 = system error (fallback to password)

	startTime := time.Now()

	// Get the username from PAM environment
	username := os.Getenv("PAM_USER")
	if username == "" {
		// Fallback: try to get current user
		currentUser, err := user.Current()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Could not determine username")
			os.Exit(3)
		}
		username = currentUser.Username
	}

	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		// Try default config path
		cfg, err = config.Load("/etc/facepass/facepass.yaml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "FacePass: Configuration error: %v\n", err)
			os.Exit(3)
		}
	}

	// Initialize logging (to file for PAM, stdout would interfere)
	logging.Initialize(cfg.Logging.Level, cfg.Logging.File)

	logging.Infof("FacePass PAM v%s starting authentication for: %s", version, username)

	// Check if face auth is enabled
	if !cfg.Auth.Enabled {
		fmt.Fprintln(os.Stderr, "FacePass: Face authentication disabled")
		os.Exit(2)
	}

	// Create authenticator
	auth, err := pam.NewPAMAuthenticator(cfg)
	if err != nil {
		logging.Errorf("Failed to initialize authenticator: %v", err)
		fmt.Fprintf(os.Stderr, "FacePass: Initialization error\n")
		os.Exit(3)
	}
	defer auth.Close()

	// Override timeout if set in environment
	if pamTimeout := os.Getenv("PAM_FACEPASS_TIMEOUT"); pamTimeout != "" {
		var timeout int
		if _, err := fmt.Sscanf(pamTimeout, "%d", &timeout); err == nil && timeout > 0 {
			auth.SetTimeout(timeout)
		}
	}

	// Perform authentication
	fmt.Fprintf(os.Stderr, "FacePass: Authenticating %s (look at camera)...\n", username)

	result := auth.Authenticate(username)

	// Handle result
	if result.Success {
		logging.Infof("Authentication successful for %s (confidence: %.2f, duration: %v)",
			username, result.Confidence, result.Duration)
		fmt.Fprintf(os.Stderr, "FacePass: Authentication successful (%.0f%% confidence)\n",
			result.Confidence*100)
		os.Exit(0)
	}

	// Authentication failed
	logging.Warnf("Authentication failed for %s: %s (duration: %v)",
		username, result.Reason, time.Since(startTime))

	// Determine exit code based on error type
	if authErr, ok := result.Error.(*pam.AuthError); ok {
		switch authErr.Code {
		case pam.ErrCodeNotEnrolled:
			fmt.Fprintln(os.Stderr, "FacePass: User not enrolled")
			os.Exit(2)
		case pam.ErrCodeTimeout:
			fmt.Fprintln(os.Stderr, "FacePass: Timeout, falling back to password")
			os.Exit(2)
		case pam.ErrCodeCamera:
			fmt.Fprintln(os.Stderr, "FacePass: Camera error, falling back to password")
			os.Exit(3)
		case pam.ErrCodeLiveness:
			fmt.Fprintf(os.Stderr, "FacePass: %s\n", pam.GetErrorMessage(authErr.Code))
			os.Exit(1)
		case pam.ErrCodeNotRecognized:
			fmt.Fprintln(os.Stderr, "FacePass: Face not recognized")
			os.Exit(1)
		case pam.ErrCodeNoFace:
			fmt.Fprintln(os.Stderr, "FacePass: No face detected, falling back to password")
			os.Exit(2)
		default:
			fmt.Fprintf(os.Stderr, "FacePass: %s\n", result.Reason)
			os.Exit(1)
		}
	}

	// Generic failure
	fmt.Fprintf(os.Stderr, "FacePass: Authentication failed: %s\n", result.Reason)
	os.Exit(1)
}
