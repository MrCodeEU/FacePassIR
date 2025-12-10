package pam

import (
	"github.com/MrCodeEU/facepass/pkg/camera"
	"github.com/MrCodeEU/facepass/pkg/liveness"
	"github.com/MrCodeEU/facepass/pkg/recognition"
	"github.com/MrCodeEU/facepass/pkg/storage"
)

// MockCamera implements Camera interface for testing
type MockCamera struct {
	CaptureFunc          func() (*camera.Frame, error)
	HasIREmitterFunc     func() bool
	EnableIREmitterFunc  func() error
	DisableIREmitterFunc func() error
	CloseFunc            func() error
	OpenFunc             func(device string) error
	SetResolutionFunc    func(width, height int) error
	GetDeviceInfoFunc    func() camera.DeviceInfo
	StartStreamingFunc   func() error
	StopStreamingFunc    func() error
	ReadFrameFunc        func() (*camera.Frame, error)
}

func (m *MockCamera) Capture() (*camera.Frame, error) {
	if m.CaptureFunc != nil {
		return m.CaptureFunc()
	}
	return &camera.Frame{}, nil
}

func (m *MockCamera) StartStreaming() error {
	if m.StartStreamingFunc != nil {
		return m.StartStreamingFunc()
	}
	return nil
}

func (m *MockCamera) StopStreaming() error {
	if m.StopStreamingFunc != nil {
		return m.StopStreamingFunc()
	}
	return nil
}

func (m *MockCamera) ReadFrame() (*camera.Frame, error) {
	if m.ReadFrameFunc != nil {
		return m.ReadFrameFunc()
	}
	// Fallback to CaptureFunc if ReadFrameFunc is not set
	if m.CaptureFunc != nil {
		return m.CaptureFunc()
	}
	return &camera.Frame{}, nil
}

func (m *MockCamera) HasIREmitter() bool {
	if m.HasIREmitterFunc != nil {
		return m.HasIREmitterFunc()
	}
	return false
}

func (m *MockCamera) EnableIREmitter() error {
	if m.EnableIREmitterFunc != nil {
		return m.EnableIREmitterFunc()
	}
	return nil
}

func (m *MockCamera) DisableIREmitter() error {
	if m.DisableIREmitterFunc != nil {
		return m.DisableIREmitterFunc()
	}
	return nil
}

func (m *MockCamera) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockCamera) Open(device string) error {
	if m.OpenFunc != nil {
		return m.OpenFunc(device)
	}
	return nil
}

func (m *MockCamera) SetResolution(width, height int) error {
	if m.SetResolutionFunc != nil {
		return m.SetResolutionFunc(width, height)
	}
	return nil
}

func (m *MockCamera) GetDeviceInfo() camera.DeviceInfo {
	if m.GetDeviceInfoFunc != nil {
		return m.GetDeviceInfoFunc()
	}
	return camera.DeviceInfo{}
}

// MockRecognizer implements Recognizer interface for testing
type MockRecognizer struct {
	FindBestMatchFunc    func(embedding recognition.Embedding, knownEmbeddings []recognition.Embedding) (int, float64, bool)
	CloseFunc            func() error
	LoadModelsFunc       func(path string) error
	SetToleranceFunc     func(tolerance float64)
	DetectSingleFaceFunc func(data []byte) (*recognition.Face, error)
	GetEmbeddingFunc     func(face *recognition.Face, label string) recognition.Embedding
}

func (m *MockRecognizer) FindBestMatch(embedding recognition.Embedding, knownEmbeddings []recognition.Embedding) (int, float64, bool) {
	if m.FindBestMatchFunc != nil {
		return m.FindBestMatchFunc(embedding, knownEmbeddings)
	}
	return -1, 1.0, false
}

func (m *MockRecognizer) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockRecognizer) LoadModels(path string) error {
	if m.LoadModelsFunc != nil {
		return m.LoadModelsFunc(path)
	}
	return nil
}

func (m *MockRecognizer) SetTolerance(tolerance float64) {
	if m.SetToleranceFunc != nil {
		m.SetToleranceFunc(tolerance)
	}
}

func (m *MockRecognizer) DetectSingleFace(data []byte) (*recognition.Face, error) {
	if m.DetectSingleFaceFunc != nil {
		return m.DetectSingleFaceFunc(data)
	}
	return nil, nil
}

func (m *MockRecognizer) GetEmbedding(face *recognition.Face, label string) recognition.Embedding {
	if m.GetEmbeddingFunc != nil {
		return m.GetEmbeddingFunc(face, label)
	}
	return recognition.Embedding{}
}

// MockStorage implements Storage interface for testing
type MockStorage struct {
	UserExistsFunc     func(username string) bool
	LoadUserFunc       func(username string) (*storage.UserFaceData, error)
	UpdateLastUsedFunc func(username string) error
}

func (m *MockStorage) UserExists(username string) bool {
	if m.UserExistsFunc != nil {
		return m.UserExistsFunc(username)
	}
	return false
}

func (m *MockStorage) LoadUser(username string) (*storage.UserFaceData, error) {
	if m.LoadUserFunc != nil {
		return m.LoadUserFunc(username)
	}
	return nil, nil
}

func (m *MockStorage) UpdateLastUsed(username string) error {
	if m.UpdateLastUsedFunc != nil {
		return m.UpdateLastUsedFunc(username)
	}
	return nil
}

// MockLiveness implements LivenessChecker interface for testing
type MockLiveness struct {
	DetectFunc     func(frames []liveness.Frame) liveness.Result
	QuickCheckFunc func(frames []liveness.Frame) (bool, float64)
}

func (m *MockLiveness) Detect(frames []liveness.Frame) liveness.Result {
	if m.DetectFunc != nil {
		return m.DetectFunc(frames)
	}
	return liveness.Result{IsLive: false}
}

func (m *MockLiveness) QuickCheck(frames []liveness.Frame) (bool, float64) {
	if m.QuickCheckFunc != nil {
		return m.QuickCheckFunc(frames)
	}
	return false, 0.0
}
