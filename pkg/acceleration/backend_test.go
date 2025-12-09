package acceleration

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.PreferredBackend != BackendAuto {
		t.Errorf("expected PreferredBackend Auto, got %s", cfg.PreferredBackend)
	}
	if !cfg.FallbackToCPU {
		t.Error("expected FallbackToCPU to be true")
	}
	if cfg.DeviceIndex != 0 {
		t.Errorf("expected DeviceIndex 0, got %d", cfg.DeviceIndex)
	}
	if cfg.EnableProfiling {
		t.Error("expected EnableProfiling to be false")
	}
}

func TestGetManager(t *testing.T) {
	manager := GetManager()
	if manager == nil {
		t.Fatal("GetManager returned nil")
	}

	// Should return same instance
	manager2 := GetManager()
	if manager != manager2 {
		t.Error("GetManager should return singleton")
	}
}

func TestManager_Initialize(t *testing.T) {
	manager := &Manager{
		availableBackends: make(map[Backend]*BackendInfo),
	}

	cfg := DefaultConfig()
	err := manager.Initialize(cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if !manager.initialized {
		t.Error("manager should be initialized")
	}

	// CPU should always be available
	cpuInfo := manager.GetBackendInfo(BackendCPU)
	if cpuInfo == nil {
		t.Error("CPU backend info should not be nil")
	}
	if !cpuInfo.Available {
		t.Error("CPU backend should be available")
	}
}

func TestManager_GetActiveBackend(t *testing.T) {
	manager := &Manager{
		availableBackends: make(map[Backend]*BackendInfo),
		activeBackend:     BackendCPU,
	}

	if manager.GetActiveBackend() != BackendCPU {
		t.Errorf("expected CPU backend, got %s", manager.GetActiveBackend())
	}
}

func TestManager_IsAccelerated(t *testing.T) {
	tests := []struct {
		name     string
		backend  Backend
		expected bool
	}{
		{"CPU", BackendCPU, false},
		{"ROCm", BackendROCm, true},
		{"CUDA", BackendCUDA, true},
		{"OpenVINO", BackendOpenVINO, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &Manager{
				activeBackend: tt.backend,
			}
			if manager.IsAccelerated() != tt.expected {
				t.Errorf("IsAccelerated for %s: expected %v", tt.name, tt.expected)
			}
		})
	}
}

func TestManager_GetAllBackends(t *testing.T) {
	manager := &Manager{
		availableBackends: map[Backend]*BackendInfo{
			BackendCPU: {Backend: BackendCPU, Available: true},
		},
	}

	backends := manager.GetAllBackends()
	if len(backends) != 1 {
		t.Errorf("expected 1 backend, got %d", len(backends))
	}
	if _, ok := backends[BackendCPU]; !ok {
		t.Error("CPU backend should be in map")
	}
}

func TestManager_selectBackend(t *testing.T) {
	tests := []struct {
		name       string
		preferred  Backend
		available  map[Backend]*BackendInfo
		fallback   bool
		expected   Backend
	}{
		{
			name:      "auto with only CPU",
			preferred: BackendAuto,
			available: map[Backend]*BackendInfo{
				BackendCPU: {Backend: BackendCPU, Available: true},
			},
			fallback: true,
			expected: BackendCPU,
		},
		{
			name:      "auto prefers ROCm over CPU",
			preferred: BackendAuto,
			available: map[Backend]*BackendInfo{
				BackendCPU:  {Backend: BackendCPU, Available: true},
				BackendROCm: {Backend: BackendROCm, Available: true},
			},
			fallback: true,
			expected: BackendROCm,
		},
		{
			name:      "specific backend available",
			preferred: BackendCUDA,
			available: map[Backend]*BackendInfo{
				BackendCPU:  {Backend: BackendCPU, Available: true},
				BackendCUDA: {Backend: BackendCUDA, Available: true},
			},
			fallback: true,
			expected: BackendCUDA,
		},
		{
			name:      "specific backend not available with fallback",
			preferred: BackendCUDA,
			available: map[Backend]*BackendInfo{
				BackendCPU: {Backend: BackendCPU, Available: true},
			},
			fallback: true,
			expected: BackendCPU,
		},
		{
			name:      "specific backend not available without fallback",
			preferred: BackendCUDA,
			available: map[Backend]*BackendInfo{
				BackendCPU: {Backend: BackendCPU, Available: true},
			},
			fallback: false,
			expected: BackendCPU,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &Manager{
				availableBackends: tt.available,
				config: Config{
					FallbackToCPU: tt.fallback,
				},
			}
			result := manager.selectBackend(tt.preferred)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestBackendInfo(t *testing.T) {
	info := &BackendInfo{
		Backend:     BackendROCm,
		Name:        "AMD ROCm",
		Available:   true,
		Tested:      true,
		Version:     "5.7",
		DeviceName:  "AMD Radeon RX 7900",
		DeviceCount: 1,
		Warning:     "",
	}

	if info.Backend != BackendROCm {
		t.Error("Backend mismatch")
	}
	if !info.Available {
		t.Error("Should be available")
	}
	if !info.Tested {
		t.Error("Should be tested")
	}
}

func TestGetCPUName(t *testing.T) {
	name := getCPUName()
	// Should return something, even if "Unknown CPU"
	if name == "" {
		t.Error("getCPUName returned empty string")
	}
}

func TestDetectIREmitter_Unavailable(t *testing.T) {
	// This test just ensures the detection functions don't panic
	// Actual detection depends on hardware

	// These should return nil or non-nil based on system
	rocmInfo := detectROCm()
	_ = rocmInfo // May be nil on non-ROCm systems

	cudaInfo := detectCUDA()
	_ = cudaInfo // May be nil on non-CUDA systems

	openvinoInfo := detectOpenVINO()
	_ = openvinoInfo // May be nil on non-OpenVINO systems
}

func TestBackendConstants(t *testing.T) {
	// Verify backend constants are distinct
	backends := []Backend{
		BackendCPU,
		BackendROCm,
		BackendCUDA,
		BackendOpenVINO,
		BackendAuto,
	}

	seen := make(map[Backend]bool)
	for _, b := range backends {
		if seen[b] {
			t.Errorf("duplicate backend constant: %s", b)
		}
		seen[b] = true
	}
}

func TestBackendStrings(t *testing.T) {
	tests := []struct {
		backend  Backend
		expected string
	}{
		{BackendCPU, "cpu"},
		{BackendROCm, "rocm"},
		{BackendCUDA, "cuda"},
		{BackendOpenVINO, "openvino"},
		{BackendAuto, "auto"},
	}

	for _, tt := range tests {
		t.Run(string(tt.backend), func(t *testing.T) {
			if string(tt.backend) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.backend))
			}
		})
	}
}

// Benchmark tests
func BenchmarkManager_GetActiveBackend(b *testing.B) {
	manager := &Manager{
		activeBackend: BackendCPU,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetActiveBackend()
	}
}

func BenchmarkManager_IsAccelerated(b *testing.B) {
	manager := &Manager{
		activeBackend: BackendROCm,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.IsAccelerated()
	}
}
