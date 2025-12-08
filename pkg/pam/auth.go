// Package pam provides PAM (Pluggable Authentication Modules) integration.
// It handles the authentication flow when called by PAM.
package pam

import (
	"errors"
	"time"
)

// AuthResult represents the result of an authentication attempt.
type AuthResult struct {
	Success   bool
	Error     error
	Duration  time.Duration
	Attempts  int
	Reason    string
}

// ErrorCode represents a specific authentication error type.
type ErrorCode string

const (
	ErrCodeNoFace       ErrorCode = "NO_FACE"
	ErrCodeMultipleFaces ErrorCode = "MULTIPLE_FACES"
	ErrCodeLiveness     ErrorCode = "LIVENESS_FAILED"
	ErrCodeNotRecognized ErrorCode = "NOT_RECOGNIZED"
	ErrCodeCamera       ErrorCode = "CAMERA_ERROR"
	ErrCodeTimeout      ErrorCode = "TIMEOUT"
	ErrCodeNotEnrolled  ErrorCode = "NOT_ENROLLED"
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
	ErrCodeNoFace:       "Please position your face in front of the camera",
	ErrCodeMultipleFaces: "Multiple faces detected. Please ensure only you are in frame",
	ErrCodeLiveness:     "Liveness check failed. Please blink and try again",
	ErrCodeNotRecognized: "Face not recognized. Falling back to password...",
	ErrCodeCamera:       "Camera error. Please check your camera connection",
	ErrCodeTimeout:      "Face recognition timed out. Please enter your password",
	ErrCodeNotEnrolled:  "No face data enrolled for this user",
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

// TODO: Implement PAM authentication
// - Read username from PAM environment
// - Load user face data
// - Initialize camera
// - Capture frames with timeout
// - Perform liveness detection
// - Compare face embeddings
// - Return result to PAM
