package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/MrCodeEU/facepass/pkg/pam"
)

type MockAuthenticator struct {
	Result pam.AuthResult
}

func (m *MockAuthenticator) Authenticate(username string) pam.AuthResult {
	return m.Result
}

func (m *MockAuthenticator) SetTimeout(seconds int)      {}
func (m *MockAuthenticator) SetMaxAttempts(attempts int) {}

func TestRunAuthentication(t *testing.T) {
	tests := []struct {
		name     string
		result   pam.AuthResult
		expected int
	}{
		{
			name: "Success",
			result: pam.AuthResult{
				Success:    true,
				Confidence: 0.95,
			},
			expected: 0,
		},
		{
			name: "NotEnrolled",
			result: pam.AuthResult{
				Success: false,
				Error:   &pam.AuthError{Code: pam.ErrCodeNotEnrolled},
				Reason:  "User not enrolled",
			},
			expected: 2,
		},
		{
			name: "Timeout",
			result: pam.AuthResult{
				Success: false,
				Error:   &pam.AuthError{Code: pam.ErrCodeTimeout},
				Reason:  "Timeout",
			},
			expected: 2,
		},
		{
			name: "CameraError",
			result: pam.AuthResult{
				Success: false,
				Error:   &pam.AuthError{Code: pam.ErrCodeCamera},
				Reason:  "Camera error",
			},
			expected: 3,
		},
		{
			name: "LivenessFailed",
			result: pam.AuthResult{
				Success: false,
				Error:   &pam.AuthError{Code: pam.ErrCodeLiveness},
				Reason:  "Liveness failed",
			},
			expected: 1,
		},
		{
			name: "NotRecognized",
			result: pam.AuthResult{
				Success: false,
				Error:   &pam.AuthError{Code: pam.ErrCodeNotRecognized},
				Reason:  "Not recognized",
			},
			expected: 1,
		},
		{
			name: "NoFace",
			result: pam.AuthResult{
				Success: false,
				Error:   &pam.AuthError{Code: pam.ErrCodeNoFace},
				Reason:  "No face",
			},
			expected: 2,
		},
		{
			name: "GenericError",
			result: pam.AuthResult{
				Success: false,
				Error:   fmt.Errorf("some random error"),
				Reason:  "Random error",
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockAuthenticator{Result: tt.result}
			code := runAuthentication(mock, "testuser", time.Now())
			if code != tt.expected {
				t.Errorf("runAuthentication() = %d, want %d", code, tt.expected)
			}
		})
	}
}
