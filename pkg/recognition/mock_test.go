package recognition

import (
	"github.com/Kagami/go-face"
)

type MockFaceEngine struct {
	RecognizeFunc func(data []byte) ([]face.Face, error)
	CloseFunc     func()
}

func (m *MockFaceEngine) Recognize(data []byte) ([]face.Face, error) {
	if m.RecognizeFunc != nil {
		return m.RecognizeFunc(data)
	}
	return nil, nil
}

func (m *MockFaceEngine) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}
