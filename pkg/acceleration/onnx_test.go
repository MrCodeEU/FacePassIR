package acceleration

import (
	"testing"
)

func TestDefaultONNXConfig(t *testing.T) {
	cfg := DefaultONNXConfig()
	if cfg.ModelPath != "/usr/share/facepass/models/onnx" {
		t.Errorf("Expected default model path /usr/share/facepass/models/onnx, got %s", cfg.ModelPath)
	}
	if cfg.NumThreads != 0 {
		t.Errorf("Expected NumThreads 0, got %d", cfg.NumThreads)
	}
}

func TestNewONNXEngine_Error(t *testing.T) {
	// Test with invalid model path to trigger error
	cfg := DefaultONNXConfig()
	cfg.ModelPath = "/invalid/path"

	engine, err := NewONNXEngine(cfg)
	if err == nil {
		t.Error("Expected error for invalid model path")
	}
	if engine != nil {
		t.Error("Expected nil engine")
	}
}
