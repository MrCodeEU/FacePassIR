package main

import (
	"fmt"
	"os"
	"os/user"
)

const version = "0.1.0"

func main() {
	// PAM module entry point
	// This binary is called by PAM during authentication

	// Get the username from PAM (usually passed via environment or stdin)
	username := os.Getenv("PAM_USER")
	if username == "" {
		// Fallback: try to get current user
		currentUser, err := user.Current()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Could not determine username")
			os.Exit(1)
		}
		username = currentUser.Username
	}

	// TODO: Implement actual face recognition authentication
	// 1. Load user's face data from storage
	// 2. Initialize camera (enable IR if available)
	// 3. Capture frames
	// 4. Perform liveness detection
	// 5. Extract face embeddings
	// 6. Compare with stored embeddings
	// 7. Return success (exit 0) or failure (exit 1)

	fmt.Fprintf(os.Stderr, "FacePass PAM v%s - Authenticating user: %s\n", version, username)
	fmt.Fprintln(os.Stderr, "Face recognition not implemented yet, falling back to password...")

	// Exit with failure to trigger password fallback
	os.Exit(1)
}
