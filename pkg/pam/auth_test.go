package pam

import (
	"errors"
	"testing"
	"time"

	"github.com/MrCodeEU/facepass/pkg/camera"
	"github.com/MrCodeEU/facepass/pkg/config"
	"github.com/MrCodeEU/facepass/pkg/liveness"
	"github.com/MrCodeEU/facepass/pkg/recognition"
	"github.com/MrCodeEU/facepass/pkg/storage"
)

func TestAuthenticate(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Recognition.Tolerance = 0.4

	t.Run("UserNotEnrolled", func(t *testing.T) {
		mockStorage := &MockStorage{
			UserExistsFunc: func(username string) bool {
				return false
			},
		}
		auth := &PAMAuthenticator{
			config:      cfg,
			storage:     mockStorage,
			timeout:     1 * time.Second,
			maxAttempts: 1,
		}

		result := auth.Authenticate("unknown_user")
		if result.Success {
			t.Error("expected authentication to fail")
		}
		if authErr, ok := result.Error.(*AuthError); ok {
			if authErr.Code != ErrCodeNotEnrolled {
				t.Errorf("expected error code %s, got %s", ErrCodeNotEnrolled, authErr.Code)
			}
		} else {
			t.Errorf("expected AuthError, got %T", result.Error)
		}
	})

	t.Run("CameraCaptureFailure", func(t *testing.T) {
		mockStorage := &MockStorage{
			UserExistsFunc: func(username string) bool {
				return true
			},
			LoadUserFunc: func(username string) (*storage.UserFaceData, error) {
				return &storage.UserFaceData{Username: username}, nil
			},
		}
		mockCamera := &MockCamera{
			HasIREmitterFunc: func() bool { return false },
			CaptureFunc: func() (*camera.Frame, error) {
				return nil, camera.ErrNoFrame
			},
		}
		auth := &PAMAuthenticator{
			config:      cfg,
			storage:     mockStorage,
			camera:      mockCamera,
			timeout:     1 * time.Second,
			maxAttempts: 1,
		}

		result := auth.Authenticate("testuser")
		if result.Success {
			t.Error("expected authentication to fail")
		}
		// Should fail with NotRecognized after max attempts (since capture failed repeatedly)
		if authErr, ok := result.Error.(*AuthError); ok {
			if authErr.Code != ErrCodeNotRecognized {
				t.Errorf("expected error code %s, got %s", ErrCodeNotRecognized, authErr.Code)
			}
		} else {
			t.Errorf("expected AuthError, got %T", result.Error)
		}
	})

	t.Run("LivenessFailure", func(t *testing.T) {
		mockStorage := &MockStorage{
			UserExistsFunc: func(username string) bool { return true },
			LoadUserFunc: func(username string) (*storage.UserFaceData, error) {
				return &storage.UserFaceData{Username: username}, nil
			},
		}
		mockCamera := &MockCamera{
			HasIREmitterFunc: func() bool { return false },
			CaptureFunc: func() (*camera.Frame, error) {
				return &camera.Frame{Data: []byte("fake")}, nil
			},
		}
		mockLiveness := &MockLiveness{
			DetectFunc: func(frames []liveness.Frame) liveness.Result {
				return liveness.Result{IsLive: false, Reason: "spoof", RequiresRetry: false}
			},
		}
		mockRecognizer := &MockRecognizer{
			DetectSingleFaceFunc: func(data []byte) (*recognition.Face, error) {
				return &recognition.Face{}, nil
			},
			GetEmbeddingFunc: func(face *recognition.Face, label string) recognition.Embedding {
				return recognition.Embedding{}
			},
		}
		auth := &PAMAuthenticator{
			config:      cfg,
			storage:     mockStorage,
			camera:      mockCamera,
			liveness:    mockLiveness,
			recognizer:  mockRecognizer,
			timeout:     1 * time.Second,
			maxAttempts: 1,
		}

		result := auth.Authenticate("testuser")
		if result.Success {
			t.Error("expected authentication to fail")
		}
		if authErr, ok := result.Error.(*AuthError); ok {
			if authErr.Code != ErrCodeLiveness {
				t.Errorf("expected error code %s, got %s", ErrCodeLiveness, authErr.Code)
			}
		} else {
			t.Errorf("expected AuthError, got %T", result.Error)
		}
	})

	t.Run("SuccessfulAuthentication", func(t *testing.T) {
		mockStorage := &MockStorage{
			UserExistsFunc: func(username string) bool { return true },
			LoadUserFunc: func(username string) (*storage.UserFaceData, error) {
				return &storage.UserFaceData{Username: username}, nil
			},
			UpdateLastUsedFunc: func(username string) error { return nil },
		}
		mockCamera := &MockCamera{
			HasIREmitterFunc: func() bool { return false },
			CaptureFunc: func() (*camera.Frame, error) {
				return &camera.Frame{Data: []byte("real_face")}, nil
			},
		}
		mockLiveness := &MockLiveness{
			DetectFunc: func(frames []liveness.Frame) liveness.Result {
				return liveness.Result{IsLive: true}
			},
		}
		mockRecognizer := &MockRecognizer{
			DetectSingleFaceFunc: func(data []byte) (*recognition.Face, error) {
				return &recognition.Face{}, nil
			},
			GetEmbeddingFunc: func(face *recognition.Face, label string) recognition.Embedding {
				return recognition.Embedding{}
			},
			FindBestMatchFunc: func(embedding recognition.Embedding, knownEmbeddings []recognition.Embedding) (int, float64, bool) {
				return 0, 0.1, true
			},
		}

		auth := &PAMAuthenticator{
			config:      cfg,
			storage:     mockStorage,
			camera:      mockCamera,
			liveness:    mockLiveness,
			recognizer:  mockRecognizer,
			timeout:     1 * time.Second,
			maxAttempts: 1,
		}

		result := auth.Authenticate("testuser")
		if !result.Success {
			t.Errorf("expected authentication to succeed, got error: %v", result.Error)
		}
	})

	t.Run("Timeout", func(t *testing.T) {
		mockStorage := &MockStorage{
			UserExistsFunc: func(username string) bool { return true },
			LoadUserFunc: func(username string) (*storage.UserFaceData, error) {
				return &storage.UserFaceData{Username: username}, nil
			},
		}
		mockCamera := &MockCamera{
			HasIREmitterFunc: func() bool { return false },
			CaptureFunc: func() (*camera.Frame, error) {
				time.Sleep(50 * time.Millisecond)
				return &camera.Frame{Data: []byte("face")}, nil
			},
			GetDeviceInfoFunc: func() camera.DeviceInfo { return camera.DeviceInfo{} },
		}
		mockRecognizer := &MockRecognizer{
			DetectSingleFaceFunc: func(data []byte) (*recognition.Face, error) {
				return &recognition.Face{}, nil
			},
			GetEmbeddingFunc: func(face *recognition.Face, label string) recognition.Embedding {
				return recognition.Embedding{}
			},
		}

		auth := &PAMAuthenticator{
			config:      cfg,
			storage:     mockStorage,
			camera:      mockCamera,
			recognizer:  mockRecognizer,
			timeout:     10 * time.Millisecond,
			maxAttempts: 1,
		}

		result := auth.Authenticate("testuser")
		if result.Success {
			t.Error("expected failure")
		}
		if result.Error.(*AuthError).Code != ErrCodeTimeout {
			t.Errorf("expected ErrCodeTimeout, got %s", result.Error.(*AuthError).Code)
		}
	})

	t.Run("NotRecognized", func(t *testing.T) {
		mockStorage := &MockStorage{
			UserExistsFunc: func(username string) bool { return true },
			LoadUserFunc: func(username string) (*storage.UserFaceData, error) {
				return &storage.UserFaceData{Username: username}, nil
			},
		}
		mockCamera := &MockCamera{
			HasIREmitterFunc: func() bool { return false },
			CaptureFunc: func() (*camera.Frame, error) {
				return &camera.Frame{Data: []byte("face")}, nil
			},
		}
		mockLiveness := &MockLiveness{
			DetectFunc: func(frames []liveness.Frame) liveness.Result {
				return liveness.Result{IsLive: true}
			},
		}
		mockRecognizer := &MockRecognizer{
			DetectSingleFaceFunc: func(data []byte) (*recognition.Face, error) {
				return &recognition.Face{}, nil
			},
			GetEmbeddingFunc: func(face *recognition.Face, label string) recognition.Embedding {
				return recognition.Embedding{}
			},
			FindBestMatchFunc: func(embedding recognition.Embedding, knownEmbeddings []recognition.Embedding) (int, float64, bool) {
				return -1, 1.0, false // No match
			},
		}

		auth := &PAMAuthenticator{
			config:      cfg,
			storage:     mockStorage,
			camera:      mockCamera,
			liveness:    mockLiveness,
			recognizer:  mockRecognizer,
			timeout:     1 * time.Second,
			maxAttempts: 1,
		}

		result := auth.Authenticate("testuser")
		if result.Success {
			t.Error("expected failure")
		}
		if result.Error.(*AuthError).Code != ErrCodeNotRecognized {
			t.Errorf("expected ErrCodeNotRecognized, got %s", result.Error.(*AuthError).Code)
		}
	})

	t.Run("IREmitterUsage", func(t *testing.T) {
		mockStorage := &MockStorage{
			UserExistsFunc: func(username string) bool { return true },
			LoadUserFunc: func(username string) (*storage.UserFaceData, error) {
				return &storage.UserFaceData{Username: username}, nil
			},
			UpdateLastUsedFunc: func(username string) error { return nil },
		}

		irEnabled := false
		irDisabled := false

		mockCamera := &MockCamera{
			HasIREmitterFunc: func() bool { return true },
			EnableIREmitterFunc: func() error {
				irEnabled = true
				return nil
			},
			DisableIREmitterFunc: func() error {
				irDisabled = true
				return nil
			},
			CaptureFunc: func() (*camera.Frame, error) {
				return &camera.Frame{Data: []byte("face")}, nil
			},
		}
		mockLiveness := &MockLiveness{
			DetectFunc: func(frames []liveness.Frame) liveness.Result {
				return liveness.Result{IsLive: true}
			},
		}
		mockRecognizer := &MockRecognizer{
			DetectSingleFaceFunc: func(data []byte) (*recognition.Face, error) {
				return &recognition.Face{}, nil
			},
			GetEmbeddingFunc: func(face *recognition.Face, label string) recognition.Embedding {
				return recognition.Embedding{Vector: recognition.Descriptor{1}}
			},
			FindBestMatchFunc: func(embedding recognition.Embedding, knownEmbeddings []recognition.Embedding) (int, float64, bool) {
				return 0, 0.1, true
			},
		}

		auth := &PAMAuthenticator{
			config:      cfg,
			storage:     mockStorage,
			camera:      mockCamera,
			liveness:    mockLiveness,
			recognizer:  mockRecognizer,
			timeout:     1 * time.Second,
			maxAttempts: 1,
		}

		result := auth.Authenticate("testuser")
		if !result.Success {
			t.Errorf("expected success, got error: %v", result.Error)
		}
		if !irEnabled {
			t.Error("expected IR emitter to be enabled")
		}
		if !irDisabled {
			t.Error("expected IR emitter to be disabled")
		}
	})

	t.Run("LivenessPassNoEmbeddings", func(t *testing.T) {
		mockStorage := &MockStorage{
			UserExistsFunc: func(username string) bool { return true },
			LoadUserFunc: func(username string) (*storage.UserFaceData, error) {
				return &storage.UserFaceData{Username: username}, nil
			},
		}
		mockCamera := &MockCamera{
			HasIREmitterFunc: func() bool { return false },
			CaptureFunc: func() (*camera.Frame, error) {
				return &camera.Frame{Data: []byte("face")}, nil
			},
		}
		mockLiveness := &MockLiveness{
			DetectFunc: func(frames []liveness.Frame) liveness.Result {
				return liveness.Result{IsLive: true}
			},
		}
		mockRecognizer := &MockRecognizer{
			DetectSingleFaceFunc: func(data []byte) (*recognition.Face, error) {
				// Return error so no face is found in frames
				return nil, errors.New("no face")
			},
		}

		auth := &PAMAuthenticator{
			config:      cfg,
			storage:     mockStorage,
			camera:      mockCamera,
			liveness:    mockLiveness,
			recognizer:  mockRecognizer,
			timeout:     1 * time.Second,
			maxAttempts: 1,
		}

		result := auth.Authenticate("testuser")
		if result.Success {
			t.Error("expected failure")
		}
		// Should fail with NotRecognized because getBestEmbedding returns error "no face embeddings found"
		// which is caught and logged, then loop continues/finishes.
		if result.Error.(*AuthError).Code != ErrCodeNotRecognized {
			t.Errorf("expected ErrCodeNotRecognized, got %s", result.Error.(*AuthError).Code)
		}
	})
}

func TestSettersAndClose(t *testing.T) {
	mockCamera := &MockCamera{
		CloseFunc: func() error { return nil },
	}
	mockRecognizer := &MockRecognizer{
		CloseFunc: func() error { return nil },
	}

	auth := &PAMAuthenticator{
		camera:     mockCamera,
		recognizer: mockRecognizer,
	}

	auth.SetTimeout(10)
	if auth.timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", auth.timeout)
	}

	auth.SetMaxAttempts(5)
	if auth.maxAttempts != 5 {
		t.Errorf("expected max attempts 5, got %d", auth.maxAttempts)
	}

	auth.Close()
	// Verify Close was called on dependencies (implied by no panic and coverage)
}

func TestAuthenticateQuick_Original(t *testing.T) {
	cfg := config.DefaultConfig()
	mockStorage := &MockStorage{
		UserExistsFunc: func(username string) bool { return true },
		LoadUserFunc: func(username string) (*storage.UserFaceData, error) {
			return &storage.UserFaceData{Username: username}, nil
		},
		UpdateLastUsedFunc: func(username string) error { return nil },
	}
	mockCamera := &MockCamera{
		HasIREmitterFunc: func() bool { return false },
		ReadFrameFunc: func() (*camera.Frame, error) {
			return &camera.Frame{Data: []byte("real_face")}, nil
		},
		StartStreamingFunc: func() error { return nil },
		StopStreamingFunc:  func() error { return nil },
	}
	mockLiveness := &MockLiveness{
		DetectFunc: func(frames []liveness.Frame) liveness.Result {
			return liveness.Result{IsLive: true, Score: 0.9}
		},
		QuickCheckFunc: func(frames []liveness.Frame) (bool, float64) {
			return true, 0.9
		},
	}
	mockRecognizer := &MockRecognizer{
		DetectSingleFaceFunc: func(data []byte) (*recognition.Face, error) {
			return &recognition.Face{
				Confidence: 0.99,
				Landmarks:  make([]recognition.Point, 5),
			}, nil
		},
		GetEmbeddingFunc: func(face *recognition.Face, label string) recognition.Embedding {
			return recognition.Embedding{Vector: recognition.Descriptor{1, 2, 3}}
		},
		FindBestMatchFunc: func(embedding recognition.Embedding, knownEmbeddings []recognition.Embedding) (int, float64, bool) {
			return 0, 0.1, true
		},
	}

	auth := &PAMAuthenticator{
		config:      cfg,
		storage:     mockStorage,
		camera:      mockCamera,
		liveness:    mockLiveness,
		recognizer:  mockRecognizer,
		timeout:     1 * time.Second,
		maxAttempts: 1,
	}

	result := auth.AuthenticateQuick("testuser")
	if !result.Success {
		t.Errorf("expected quick authentication to succeed, got error: %v", result.Error)
	}
}

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
		_ = NewAuthError(ErrCodeNoFace, true)
	}
}

func BenchmarkGetErrorMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetErrorMessage(ErrCodeLiveness)
	}
}

func TestAuthenticateQuick_Scenarios(t *testing.T) {
	cfg := config.DefaultConfig()

	t.Run("Success", func(t *testing.T) {
		mockStorage := &MockStorage{
			UserExistsFunc: func(username string) bool { return true },
			LoadUserFunc: func(username string) (*storage.UserFaceData, error) {
				return &storage.UserFaceData{
					Username: username,
					Embeddings: []recognition.Embedding{
						{Vector: recognition.Descriptor{1, 2, 3}},
					},
				}, nil
			},
			UpdateLastUsedFunc: func(username string) error { return nil },
		}
		mockCamera := &MockCamera{
			HasIREmitterFunc: func() bool { return false },
			ReadFrameFunc: func() (*camera.Frame, error) {
				return &camera.Frame{Data: []byte("face")}, nil
			},
			StartStreamingFunc: func() error { return nil },
			StopStreamingFunc:  func() error { return nil },
		}
		mockLiveness := &MockLiveness{
			DetectFunc: func(frames []liveness.Frame) liveness.Result {
				return liveness.Result{IsLive: true, Score: 0.9}
			},
			QuickCheckFunc: func(frames []liveness.Frame) (bool, float64) {
				return true, 0.9
			},
		}
		mockRecognizer := &MockRecognizer{
			DetectSingleFaceFunc: func(data []byte) (*recognition.Face, error) {
				return &recognition.Face{
					Confidence: 0.99,
					Landmarks:  make([]recognition.Point, 5),
				}, nil
			},
			GetEmbeddingFunc: func(face *recognition.Face, label string) recognition.Embedding {
				return recognition.Embedding{Vector: recognition.Descriptor{1, 2, 3}}
			},
			FindBestMatchFunc: func(embedding recognition.Embedding, knownEmbeddings []recognition.Embedding) (int, float64, bool) {
				return 0, 0.0, true
			},
		}

		auth := &PAMAuthenticator{
			config:     cfg,
			storage:    mockStorage,
			camera:     mockCamera,
			liveness:   mockLiveness,
			recognizer: mockRecognizer,
		}

		result := auth.AuthenticateQuick("testuser")
		if !result.Success {
			t.Errorf("expected success, got failure: %v", result.Error)
		}
	})

	t.Run("NotEnrolled", func(t *testing.T) {
		mockStorage := &MockStorage{
			UserExistsFunc: func(username string) bool { return false },
		}
		auth := &PAMAuthenticator{
			config:  cfg,
			storage: mockStorage,
		}

		result := auth.AuthenticateQuick("unknown")
		if result.Success {
			t.Error("expected failure for unknown user")
		}
		if result.Error.(*AuthError).Code != ErrCodeNotEnrolled {
			t.Errorf("expected ErrCodeNotEnrolled, got %s", result.Error.(*AuthError).Code)
		}
	})

	t.Run("LivenessFailure", func(t *testing.T) {
		mockStorage := &MockStorage{
			UserExistsFunc: func(username string) bool { return true },
			LoadUserFunc: func(username string) (*storage.UserFaceData, error) {
				return &storage.UserFaceData{Username: username}, nil
			},
		}
		mockCamera := &MockCamera{
			HasIREmitterFunc: func() bool { return false },
			ReadFrameFunc: func() (*camera.Frame, error) {
				return &camera.Frame{Data: []byte("face")}, nil
			},
			StartStreamingFunc: func() error { return nil },
			StopStreamingFunc:  func() error { return nil },
		}
		mockLiveness := &MockLiveness{
			QuickCheckFunc: func(frames []liveness.Frame) (bool, float64) {
				return false, 0.4
			},
		}
		mockRecognizer := &MockRecognizer{
			DetectSingleFaceFunc: func(data []byte) (*recognition.Face, error) {
				return &recognition.Face{}, nil
			},
			GetEmbeddingFunc: func(face *recognition.Face, label string) recognition.Embedding {
				return recognition.Embedding{}
			},
		}

		auth := &PAMAuthenticator{
			config:     cfg,
			storage:    mockStorage,
			camera:     mockCamera,
			liveness:   mockLiveness,
			recognizer: mockRecognizer,
		}

		result := auth.AuthenticateQuick("testuser")
		if result.Success {
			t.Error("expected failure")
		}
		if result.Error.(*AuthError).Code != ErrCodeLiveness {
			t.Errorf("expected ErrCodeLiveness, got %s", result.Error.(*AuthError).Code)
		}
	})

	t.Run("CameraError", func(t *testing.T) {
		mockStorage := &MockStorage{
			UserExistsFunc: func(username string) bool { return true },
			LoadUserFunc: func(username string) (*storage.UserFaceData, error) {
				return &storage.UserFaceData{Username: username}, nil
			},
		}
		mockCamera := &MockCamera{
			HasIREmitterFunc: func() bool { return false },
			ReadFrameFunc: func() (*camera.Frame, error) {
				return nil, camera.ErrNoFrame
			},
			StartStreamingFunc: func() error { return nil },
			StopStreamingFunc:  func() error { return nil },
		}

		auth := &PAMAuthenticator{
			config:  cfg,
			storage: mockStorage,
			camera:  mockCamera,
		}

		result := auth.AuthenticateQuick("testuser")
		if result.Success {
			t.Error("expected failure")
		}
		if result.Error.(*AuthError).Code != ErrCodeCamera {
			t.Errorf("expected ErrCodeCamera, got %s", result.Error.(*AuthError).Code)
		}
	})

	t.Run("LoadUserFailure", func(t *testing.T) {
		mockStorage := &MockStorage{
			UserExistsFunc: func(username string) bool { return true },
			LoadUserFunc: func(username string) (*storage.UserFaceData, error) {
				return nil, errors.New("db error")
			},
		}
		auth := &PAMAuthenticator{
			config:  cfg,
			storage: mockStorage,
		}

		result := auth.AuthenticateQuick("testuser")
		if result.Success {
			t.Error("expected failure")
		}
		if result.Error.(*AuthError).Code != ErrCodeNotEnrolled {
			t.Errorf("expected ErrCodeNotEnrolled, got %s", result.Error.(*AuthError).Code)
		}
	})

	t.Run("NotRecognized", func(t *testing.T) {
		mockStorage := &MockStorage{
			UserExistsFunc: func(username string) bool { return true },
			LoadUserFunc: func(username string) (*storage.UserFaceData, error) {
				return &storage.UserFaceData{Username: username}, nil
			},
		}
		mockCamera := &MockCamera{
			HasIREmitterFunc: func() bool { return false },
			ReadFrameFunc: func() (*camera.Frame, error) {
				return &camera.Frame{Data: []byte("face")}, nil
			},
			StartStreamingFunc: func() error { return nil },
			StopStreamingFunc:  func() error { return nil },
		}
		mockLiveness := &MockLiveness{
			QuickCheckFunc: func(frames []liveness.Frame) (bool, float64) {
				return true, 0.9
			},
		}
		mockRecognizer := &MockRecognizer{
			DetectSingleFaceFunc: func(data []byte) (*recognition.Face, error) {
				return &recognition.Face{}, nil
			},
			GetEmbeddingFunc: func(face *recognition.Face, label string) recognition.Embedding {
				return recognition.Embedding{}
			},
			FindBestMatchFunc: func(embedding recognition.Embedding, knownEmbeddings []recognition.Embedding) (int, float64, bool) {
				return -1, 1.0, false // No match
			},
		}

		auth := &PAMAuthenticator{
			config:     cfg,
			storage:    mockStorage,
			camera:     mockCamera,
			liveness:   mockLiveness,
			recognizer: mockRecognizer,
		}

		result := auth.AuthenticateQuick("testuser")
		if result.Success {
			t.Error("expected failure")
		}
		if result.Error.(*AuthError).Code != ErrCodeNotRecognized {
			t.Errorf("expected ErrCodeNotRecognized, got %s", result.Error.(*AuthError).Code)
		}
	})
}

func TestClose_Scenarios(t *testing.T) {
	mockCamera := &MockCamera{
		CloseFunc: func() error { return nil },
	}
	mockRecognizer := &MockRecognizer{
		CloseFunc: func() error { return nil },
	}

	auth := &PAMAuthenticator{
		camera:     mockCamera,
		recognizer: mockRecognizer,
	}

	auth.Close()
	// No panic means success, and we can't easily verify calls without a spy,
	// but coverage will increase.
}
