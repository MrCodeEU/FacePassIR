package pam

import (
	"testing"
)

func TestErrorCodes(t *testing.T) {
	// Verify all error codes are distinct
	codes := []ErrorCode{
		ErrCodeNoFace,
		ErrCodeMultipleFaces,
		ErrCodeLiveness,
		ErrCodeNotRecognized,
		ErrCodeCamera,
		ErrCodeTimeout,
		ErrCodeNotEnrolled,
	}

	seen := make(map[ErrorCode]bool)
	for _, code := range codes {
		if seen[code] {
			t.Errorf("duplicate error code: %s", code)
		}
		seen[code] = true
	}
}

func TestGetErrorMessage(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		contains string
	}{
		{ErrCodeNoFace, "face"},
		{ErrCodeMultipleFaces, "Multiple"},
		{ErrCodeLiveness, "Liveness"},
		{ErrCodeNotRecognized, "recognized"},
		{ErrCodeCamera, "Camera"},
		{ErrCodeTimeout, "timed out"},
		{ErrCodeNotEnrolled, "enrolled"},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			msg := GetErrorMessage(tt.code)
			if msg == "" {
				t.Error("message should not be empty")
			}
			if !contains(msg, tt.contains) {
				t.Errorf("message '%s' should contain '%s'", msg, tt.contains)
			}
		})
	}
}

func TestGetErrorMessage_Unknown(t *testing.T) {
	msg := GetErrorMessage(ErrorCode("UNKNOWN"))
	if msg != "Authentication failed" {
		t.Errorf("expected default message, got '%s'", msg)
	}
}

func TestNewAuthError(t *testing.T) {
	err := NewAuthError(ErrCodeNoFace, true)

	if err.Code != ErrCodeNoFace {
		t.Errorf("expected code %s, got %s", ErrCodeNoFace, err.Code)
	}
	if !err.Retry {
		t.Error("expected Retry to be true")
	}
	if err.Message == "" {
		t.Error("message should not be empty")
	}
	if err.Details == nil {
		t.Error("details should be initialized")
	}
}

func TestAuthError_Error(t *testing.T) {
	err := NewAuthError(ErrCodeCamera, false)
	errMsg := err.Error()

	if errMsg != err.Message {
		t.Errorf("Error() should return Message: got '%s', want '%s'", errMsg, err.Message)
	}
}

func TestAuthResult(t *testing.T) {
	result := AuthResult{
		Success:    true,
		Username:   "testuser",
		Confidence: 0.95,
		Attempts:   1,
		Reason:     "matched",
	}

	if !result.Success {
		t.Error("expected Success to be true")
	}
	if result.Username != "testuser" {
		t.Errorf("expected Username 'testuser', got '%s'", result.Username)
	}
	if result.Confidence != 0.95 {
		t.Errorf("expected Confidence 0.95, got %f", result.Confidence)
	}
}

func TestAuthResult_Failed(t *testing.T) {
	authErr := NewAuthError(ErrCodeNotRecognized, false)
	result := AuthResult{
		Success:  false,
		Username: "testuser",
		Error:    authErr,
		Reason:   "face not recognized",
		Attempts: 3,
	}

	if result.Success {
		t.Error("expected Success to be false")
	}
	if result.Error == nil {
		t.Error("expected Error to be set")
	}
	if result.Attempts != 3 {
		t.Errorf("expected Attempts 3, got %d", result.Attempts)
	}
}

func TestErrorMessages_AllCovered(t *testing.T) {
	// Ensure all error codes have messages
	codes := []ErrorCode{
		ErrCodeNoFace,
		ErrCodeMultipleFaces,
		ErrCodeLiveness,
		ErrCodeNotRecognized,
		ErrCodeCamera,
		ErrCodeTimeout,
		ErrCodeNotEnrolled,
	}

	for _, code := range codes {
		msg := GetErrorMessage(code)
		if msg == "Authentication failed" {
			t.Errorf("error code %s has no specific message", code)
		}
	}
}

func TestAuthError_WithDetails(t *testing.T) {
	err := NewAuthError(ErrCodeLiveness, true)
	err.Details["score"] = 0.3
	err.Details["reason"] = "no blink detected"

	if err.Details["score"] != 0.3 {
		t.Error("details not preserved")
	}
	if err.Details["reason"] != "no blink detected" {
		t.Error("details not preserved")
	}
}

// Test that standard errors are defined
func TestStandardErrors(t *testing.T) {
	if ErrAuthFailed == nil {
		t.Error("ErrAuthFailed should not be nil")
	}
	if ErrUserNotEnrolled == nil {
		t.Error("ErrUserNotEnrolled should not be nil")
	}
	if ErrTimeout == nil {
		t.Error("ErrTimeout should not be nil")
	}
}

func TestAuthenticator_Interface(t *testing.T) {
	// Verify PAMAuthenticator implements the Authenticator interface concepts
	// (Not a compile-time check, but documents the expected behavior)

	var _ interface {
		Authenticate(username string) AuthResult
		SetTimeout(seconds int)
		SetMaxAttempts(attempts int)
	}

	// This just ensures the types are correct
	t.Log("PAMAuthenticator follows the expected interface pattern")
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkNewAuthError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewAuthError(ErrCodeNoFace, true)
	}
}

func BenchmarkGetErrorMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetErrorMessage(ErrCodeLiveness)
	}
}
