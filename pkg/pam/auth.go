// Package pam provides PAM (Pluggable Authentication Modules) integration.
// It handles the authentication flow when called by PAM.
package pam

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MrCodeEU/facepass/pkg/camera"
	"github.com/MrCodeEU/facepass/pkg/config"
	"github.com/MrCodeEU/facepass/pkg/liveness"
	"github.com/MrCodeEU/facepass/pkg/logging"
	"github.com/MrCodeEU/facepass/pkg/recognition"
	"github.com/MrCodeEU/facepass/pkg/storage"
)

// AuthResult represents the result of an authentication attempt.
type AuthResult struct {
	Success    bool
	Error      error
	Duration   time.Duration
	Attempts   int
	Reason     string
	Username   string
	Confidence float64
}

// ErrorCode represents a specific authentication error type.
type ErrorCode string

const (
	ErrCodeNoFace        ErrorCode = "NO_FACE"
	ErrCodeMultipleFaces ErrorCode = "MULTIPLE_FACES"
	ErrCodeLiveness      ErrorCode = "LIVENESS_FAILED"
	ErrCodeNotRecognized ErrorCode = "NOT_RECOGNIZED"
	ErrCodeCamera        ErrorCode = "CAMERA_ERROR"
	ErrCodeTimeout       ErrorCode = "TIMEOUT"
	ErrCodeNotEnrolled   ErrorCode = "NOT_ENROLLED"
)

// AuthError is a structured authentication error.
type AuthError struct {
	Code    ErrorCode
	Message string
	Retry   bool
	Details map[string]interface{}
}

func (e *AuthError) Error() string {
	return e.Message
}

// Authenticator defines the interface for PAM authentication.
type Authenticator interface {
	// Authenticate performs face recognition authentication.
	Authenticate(username string) AuthResult

	// SetTimeout sets the authentication timeout.
	SetTimeout(seconds int)

	// SetMaxAttempts sets the maximum number of attempts.
	SetMaxAttempts(attempts int)
}

// User-friendly error messages
var errorMessages = map[ErrorCode]string{
	ErrCodeNoFace:        "Please position your face in front of the camera",
	ErrCodeMultipleFaces: "Multiple faces detected. Please ensure only you are in frame",
	ErrCodeLiveness:      "Liveness check failed. Please blink and try again",
	ErrCodeNotRecognized: "Face not recognized. Falling back to password...",
	ErrCodeCamera:        "Camera error. Please check your camera connection",
	ErrCodeTimeout:       "Face recognition timed out. Please enter your password",
	ErrCodeNotEnrolled:   "No face data enrolled for this user",
}

// GetErrorMessage returns a user-friendly message for an error code.
func GetErrorMessage(code ErrorCode) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "Authentication failed"
}

// NewAuthError creates a new authentication error.
func NewAuthError(code ErrorCode, retry bool) *AuthError {
	return &AuthError{
		Code:    code,
		Message: GetErrorMessage(code),
		Retry:   retry,
		Details: make(map[string]interface{}),
	}
}

// ErrAuthFailed is returned when authentication fails.
var ErrAuthFailed = errors.New("authentication failed")

// ErrUserNotEnrolled is returned when user has no face data.
var ErrUserNotEnrolled = errors.New("user not enrolled")

// ErrTimeout is returned when authentication times out.
var ErrTimeout = errors.New("authentication timeout")

// Camera defines the interface for camera operations.
type Camera interface {
	Capture() (*camera.Frame, error)
	StartStreaming() error
	StopStreaming() error
	ReadFrame() (*camera.Frame, error)
	HasIREmitter() bool
	EnableIREmitter() error
	DisableIREmitter() error
	Close() error
	Open(device string) error
	SetResolution(width, height int) error
	GetDeviceInfo() camera.DeviceInfo
}

// Recognizer defines the interface for face recognition.
type Recognizer interface {
	FindBestMatch(embedding recognition.Embedding, knownEmbeddings []recognition.Embedding) (int, float64, bool)
	Close() error
	LoadModels(path string) error
	SetTolerance(tolerance float64)
	DetectSingleFace(data []byte) (*recognition.Face, error)
	GetEmbedding(face *recognition.Face, label string) recognition.Embedding
}

// Storage defines the interface for user data storage.
type Storage interface {
	UserExists(username string) bool
	LoadUser(username string) (*storage.UserFaceData, error)
	UpdateLastUsed(username string) error
}

// LivenessChecker defines the interface for liveness detection.
type LivenessChecker interface {
	Detect(frames []liveness.Frame) liveness.Result
	QuickCheck(frames []liveness.Frame) (bool, float64)
}

// PAMAuthenticator implements the Authenticator interface for PAM.
type PAMAuthenticator struct {
	config     *config.Config
	storage    Storage
	recognizer Recognizer
	camera     Camera
	liveness   LivenessChecker

	timeout     time.Duration
	maxAttempts int
}

// NewPAMAuthenticator creates a new PAM authenticator.
func NewPAMAuthenticator(cfg *config.Config) (*PAMAuthenticator, error) {
	auth := &PAMAuthenticator{
		config:      cfg,
		timeout:     time.Duration(cfg.Auth.Timeout) * time.Second,
		maxAttempts: cfg.Auth.MaxAttempts,
	}

	// Initialize storage
	store, err := storage.NewFileStorage(cfg.Storage.DataDir, cfg.Storage.EncryptionEnabled)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	auth.storage = store

	// Initialize recognizer
	auth.recognizer = recognition.NewRecognizer()
	if err := auth.recognizer.LoadModels(cfg.Recognition.ModelPath); err != nil {
		return nil, fmt.Errorf("failed to load recognition models: %w", err)
	}
	auth.recognizer.SetTolerance(cfg.Recognition.Tolerance)

	// Initialize camera
	auth.camera = camera.NewCamera()
	if err := auth.camera.Open(cfg.Camera.Device); err != nil {
		return nil, fmt.Errorf("failed to open camera: %w", err)
	}
	if err := auth.camera.SetResolution(cfg.Camera.Width, cfg.Camera.Height); err != nil {
		logging.Warnf("Failed to set camera resolution: %v", err)
	}

	// Initialize liveness detector
	livenessLevel := liveness.Level(cfg.Liveness.Level)
	livenessCfg := liveness.ConfigFromLevel(livenessLevel)
	// Apply configured thresholds
	livenessCfg.MovementThreshold = cfg.Liveness.Thresholds.Movement
	livenessCfg.DepthThreshold = cfg.Liveness.Thresholds.Depth
	livenessCfg.ConsistencyThreshold = cfg.Liveness.Thresholds.Consistency
	auth.liveness = liveness.NewDetector(livenessCfg)

	return auth, nil
}

// Close releases all resources.
func (a *PAMAuthenticator) Close() {
	if a.camera != nil {
		_ = a.camera.Close()
	}
	if a.recognizer != nil {
		_ = a.recognizer.Close()
	}
}

// SetTimeout sets the authentication timeout.
func (a *PAMAuthenticator) SetTimeout(seconds int) {
	a.timeout = time.Duration(seconds) * time.Second
}

// SetMaxAttempts sets the maximum number of attempts.
func (a *PAMAuthenticator) SetMaxAttempts(attempts int) {
	a.maxAttempts = attempts
}

// Authenticate performs face recognition authentication.
func (a *PAMAuthenticator) Authenticate(username string) AuthResult {
	startTime := time.Now()
	result := AuthResult{
		Success:  false,
		Username: username,
	}

	logging.Infof("Starting authentication for user: %s", username)

	// Check if user is enrolled
	if !a.storage.UserExists(username) {
		result.Error = NewAuthError(ErrCodeNotEnrolled, false)
		result.Reason = "user not enrolled"
		logging.Warnf("User not enrolled: %s", username)
		return result
	}

	// Load user embeddings
	userData, err := a.storage.LoadUser(username)
	if err != nil {
		result.Error = NewAuthError(ErrCodeNotEnrolled, false)
		result.Reason = "failed to load user data"
		return result
	}

	// Enable IR emitter if available
	if a.camera.HasIREmitter() {
		if err := a.camera.EnableIREmitter(); err != nil {
			logging.Warnf("Failed to enable IR emitter: %v", err)
		}
	}
	defer func() {
		_ = a.camera.DisableIREmitter()
	}()

	// Authentication loop with timeout and retries
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	// Start streaming for faster capture
	if err := a.camera.StartStreaming(); err != nil {
		logging.Warnf("Failed to start streaming, falling back to single capture: %v", err)
	}
	defer func() {
		_ = a.camera.StopStreaming()
	}()

	for attempt := 1; attempt <= a.maxAttempts; attempt++ {
		result.Attempts = attempt
		logging.Debugf("Authentication attempt %d/%d", attempt, a.maxAttempts)

		// Check for timeout
		select {
		case <-ctx.Done():
			result.Error = NewAuthError(ErrCodeTimeout, false)
			result.Reason = "authentication timed out"
			result.Duration = time.Since(startTime)
			return result
		default:
		}

		// Capture 30 frames (approx 1.5s at 20fps) for liveness detection
		frames, err := a.captureFramesForLiveness(ctx, 30)
		if err != nil {
			if ctx.Err() != nil {
				result.Error = NewAuthError(ErrCodeTimeout, false)
				result.Reason = "authentication timed out"
				result.Duration = time.Since(startTime)
				return result
			}
			logging.Warnf("Frame capture failed on attempt %d: %v", attempt, err)
			continue
		}

		// Perform liveness detection
		livenessResult := a.liveness.Detect(frames)
		if !livenessResult.IsLive {
			result.Error = NewAuthError(ErrCodeLiveness, livenessResult.RequiresRetry)
			result.Reason = livenessResult.Reason
			if !livenessResult.RequiresRetry {
				// Definite failure (e.g., photo attack)
				logging.Errorf("SECURITY ALERT: Liveness check failed - potential spoofing attempt detected: %s", livenessResult.Reason)
				result.Duration = time.Since(startTime)
				return result
			}
			logging.Warnf("Liveness check failed (retrying): %s", livenessResult.Reason)
			continue
		}

		// Get face embedding from frames
		embedding, err := a.getBestEmbedding(frames)
		if err != nil {
			logging.Warnf("Failed to get embedding on attempt %d: %v", attempt, err)
			continue
		}

		// Compare with stored embeddings
		idx, distance, matched := a.recognizer.FindBestMatch(*embedding, userData.Embeddings)
		if matched {
			result.Success = true
			result.Confidence = 1.0 - distance
			result.Duration = time.Since(startTime)
			logging.Infof("Authentication successful for %s (match index: %d, distance: %.4f)",
				username, idx, distance)

			// Update last used timestamp
			if err := a.storage.UpdateLastUsed(username); err != nil {
				logging.Warnf("Failed to update last used timestamp: %v", err)
			}

			return result
		}

		logging.Debugf("Face not matched (distance: %.4f, threshold: %.4f)", distance, a.config.Recognition.Tolerance)
	}

	// All attempts failed
	result.Error = NewAuthError(ErrCodeNotRecognized, false)
	result.Reason = "face not recognized after maximum attempts"
	result.Duration = time.Since(startTime)
	return result
}

// captureFramesForLiveness captures multiple frames for liveness detection.
func (a *PAMAuthenticator) captureFramesForLiveness(ctx context.Context, count int) ([]liveness.Frame, error) {
	var frames []liveness.Frame

	for i := 0; i < count; i++ {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return frames, ctx.Err()
		default:
		}

		// Capture frame (uses ReadFrame which handles streaming or fallback)
		camFrame, err := a.camera.ReadFrame()
		if err != nil {
			logging.Warnf("Failed to capture frame %d: %v", i, err)
			continue
		}

		// Convert to liveness frame
		liveFrame := liveness.Frame{
			Data:      camFrame.Data,
			IsIR:      a.camera.GetDeviceInfo().IsIR,
			Timestamp: camFrame.Timestamp,
			FaceFound: false,
		}

		// Detect face and get embedding
		face, err := a.recognizer.DetectSingleFace(camFrame.Data)
		if err == nil {
			liveFrame.FaceFound = true
			liveFrame.Embedding = a.recognizer.GetEmbedding(face, "auth")

			// Convert landmarks
			var landmarks []liveness.Point
			for _, p := range face.Landmarks {
				landmarks = append(landmarks, liveness.Point{X: float64(p.X), Y: float64(p.Y)})
			}
			liveFrame.Landmarks = landmarks

			// Calculate EAR
			if len(landmarks) >= 5 {
				// For 5-point landmarks: 0,1 are left eye; 2,3 are right eye
				leftEye := landmarks[0:2]
				rightEye := landmarks[2:4]
				leftEAR := liveness.CalculateEyeAspectRatio(leftEye)
				rightEAR := liveness.CalculateEyeAspectRatio(rightEye)
				liveFrame.EyeAspectRatio = (leftEAR + rightEAR) / 2.0
			}
		}

		frames = append(frames, liveFrame)

		// No sleep needed when streaming
	}

	if len(frames) < 5 {
		return nil, fmt.Errorf("insufficient frames captured: %d", len(frames))
	}

	return frames, nil
}

// getBestEmbedding extracts the best quality embedding from frames.
func (a *PAMAuthenticator) getBestEmbedding(frames []liveness.Frame) (*recognition.Embedding, error) {
	var embeddings []recognition.Embedding

	for _, frame := range frames {
		if frame.FaceFound && len(frame.Embedding.Vector) > 0 {
			embeddings = append(embeddings, frame.Embedding)
		}
	}

	if len(embeddings) == 0 {
		return nil, errors.New("no face embeddings found")
	}

	// Use averaged embedding for better accuracy
	avgEmb := recognition.AverageEmbedding(embeddings)
	return &avgEmb, nil
}

// AuthenticateQuick performs a quick authentication with fewer checks.
// Used for subsequent authentications within a session.
func (a *PAMAuthenticator) AuthenticateQuick(username string) AuthResult {
	startTime := time.Now()
	result := AuthResult{
		Success:  false,
		Username: username,
		Attempts: 1,
	}

	// Check if user is enrolled
	if !a.storage.UserExists(username) {
		result.Error = NewAuthError(ErrCodeNotEnrolled, false)
		result.Reason = "user not enrolled"
		return result
	}

	// Load user embeddings
	userData, err := a.storage.LoadUser(username)
	if err != nil {
		result.Error = NewAuthError(ErrCodeNotEnrolled, false)
		result.Reason = "failed to load user data"
		return result
	}

	// Capture a few frames
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Start streaming for faster capture
	if err := a.camera.StartStreaming(); err != nil {
		logging.Warnf("Failed to start streaming, falling back to single capture: %v", err)
	}
	defer func() {
		_ = a.camera.StopStreaming()
	}()

	frames, err := a.captureFramesForLiveness(ctx, 10)
	if err != nil {
		result.Error = NewAuthError(ErrCodeCamera, true)
		result.Reason = "failed to capture frames"
		result.Duration = time.Since(startTime)
		return result
	}

	// Quick liveness check
	isLive, score := a.liveness.QuickCheck(frames)
	if !isLive {
		result.Error = NewAuthError(ErrCodeLiveness, true)
		result.Reason = fmt.Sprintf("quick liveness check failed (score: %.2f)", score)
		result.Duration = time.Since(startTime)
		return result
	}

	// Get embedding
	embedding, err := a.getBestEmbedding(frames)
	if err != nil {
		result.Error = NewAuthError(ErrCodeNoFace, true)
		result.Reason = "no face detected"
		result.Duration = time.Since(startTime)
		return result
	}

	// Match
	idx, distance, matched := a.recognizer.FindBestMatch(*embedding, userData.Embeddings)
	if matched {
		result.Success = true
		result.Confidence = 1.0 - distance
		result.Duration = time.Since(startTime)
		logging.Debugf("Quick auth successful for %s (idx: %d, dist: %.4f)", username, idx, distance)
		return result
	}

	result.Error = NewAuthError(ErrCodeNotRecognized, false)
	result.Reason = "face not recognized"
	result.Duration = time.Since(startTime)
	return result
}
