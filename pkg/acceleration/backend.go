// Package acceleration provides GPU/NPU acceleration support for face recognition.
// It supports multiple backends: AMD ROCm, NVIDIA CUDA, and Intel OpenVINO
// through ONNX Runtime execution providers.
package acceleration

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/MrCodeEU/facepass/pkg/logging"
)

// Backend represents an acceleration backend type.
type Backend string

const (
	// BackendCPU is the default CPU-only backend (always available).
	BackendCPU Backend = "cpu"

	// BackendROCm is the AMD ROCm backend for AMD GPUs.
	// Status: Primary supported backend (tested).
	BackendROCm Backend = "rocm"

	// BackendCUDA is the NVIDIA CUDA backend.
	// Status: Implemented but needs community testing.
	// WARNING: This backend has not been tested by the maintainers.
	// Please report issues at: https://github.com/MrCodeEU/facepass/issues
	BackendCUDA Backend = "cuda"

	// BackendOpenVINO is the Intel OpenVINO backend for Intel GPUs/NPUs.
	// Status: Implemented but needs community testing.
	// WARNING: This backend has not been tested by the maintainers.
	// Please report issues at: https://github.com/MrCodeEU/facepass/issues
	BackendOpenVINO Backend = "openvino"

	// BackendAuto automatically selects the best available backend.
	BackendAuto Backend = "auto"
)

// BackendInfo contains information about an acceleration backend.
type BackendInfo struct {
	Backend     Backend
	Name        string
	Available   bool
	Tested      bool // Whether this backend has been tested by maintainers
	Version     string
	DeviceName  string
	DeviceCount int
	Warning     string // Warning message for untested backends
}

// Config holds acceleration configuration.
type Config struct {
	PreferredBackend Backend
	FallbackToCPU    bool
	DeviceIndex      int // For multi-GPU systems
	EnableProfiling  bool
	ModelPath        string
}

// DefaultConfig returns default acceleration configuration.
func DefaultConfig() Config {
	return Config{
		PreferredBackend: BackendAuto,
		FallbackToCPU:    true,
		DeviceIndex:      0,
		EnableProfiling:  false,
		ModelPath:        "/usr/share/facepass/models/onnx",
	}
}

// Manager manages acceleration backends.
type Manager struct {
	config          Config
	activeBackend   Backend
	availableBackends map[Backend]*BackendInfo
	mu              sync.RWMutex
	initialized     bool
}

// Global manager instance
var (
	globalManager *Manager
	managerOnce   sync.Once
)

// GetManager returns the global acceleration manager.
func GetManager() *Manager {
	managerOnce.Do(func() {
		globalManager = &Manager{
			config:            DefaultConfig(),
			availableBackends: make(map[Backend]*BackendInfo),
		}
	})
	return globalManager
}

// Initialize initializes the acceleration manager with the given config.
func (m *Manager) Initialize(cfg Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = cfg

	// Detect available backends
	m.detectBackends()

	// Select the best backend
	backend := m.selectBackend(cfg.PreferredBackend)
	m.activeBackend = backend

	m.initialized = true

	info := m.availableBackends[backend]
	if info != nil {
		logging.Infof("Acceleration initialized: %s (%s)", info.Name, info.DeviceName)
		if info.Warning != "" {
			logging.Warnf("Backend warning: %s", info.Warning)
		}
	}

	return nil
}

// detectBackends detects all available acceleration backends.
func (m *Manager) detectBackends() {
	// CPU is always available
	m.availableBackends[BackendCPU] = &BackendInfo{
		Backend:     BackendCPU,
		Name:        "CPU (dlib)",
		Available:   true,
		Tested:      true,
		DeviceName:  getCPUName(),
		DeviceCount: runtime.NumCPU(),
	}

	// Detect ROCm (AMD)
	if rocmInfo := detectROCm(); rocmInfo != nil {
		m.availableBackends[BackendROCm] = rocmInfo
	}

	// Detect CUDA (NVIDIA)
	if cudaInfo := detectCUDA(); cudaInfo != nil {
		m.availableBackends[BackendCUDA] = cudaInfo
	}

	// Detect OpenVINO (Intel)
	if openvinoInfo := detectOpenVINO(); openvinoInfo != nil {
		m.availableBackends[BackendOpenVINO] = openvinoInfo
	}
}

// selectBackend selects the best available backend.
func (m *Manager) selectBackend(preferred Backend) Backend {
	// If specific backend requested, try to use it
	if preferred != BackendAuto {
		if info, ok := m.availableBackends[preferred]; ok && info.Available {
			return preferred
		}
		if m.config.FallbackToCPU {
			logging.Warnf("Requested backend %s not available, falling back to CPU", preferred)
			return BackendCPU
		}
		return BackendCPU
	}

	// Auto-select: prefer ROCm > CUDA > OpenVINO > CPU
	// (ROCm first because it's the tested backend)
	priorities := []Backend{BackendROCm, BackendCUDA, BackendOpenVINO, BackendCPU}

	for _, backend := range priorities {
		if info, ok := m.availableBackends[backend]; ok && info.Available {
			return backend
		}
	}

	return BackendCPU
}

// GetActiveBackend returns the currently active backend.
func (m *Manager) GetActiveBackend() Backend {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeBackend
}

// GetBackendInfo returns information about a specific backend.
func (m *Manager) GetBackendInfo(backend Backend) *BackendInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.availableBackends[backend]
}

// GetAllBackends returns information about all detected backends.
func (m *Manager) GetAllBackends() map[Backend]*BackendInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[Backend]*BackendInfo)
	for k, v := range m.availableBackends {
		result[k] = v
	}
	return result
}

// IsAccelerated returns true if using GPU/NPU acceleration.
func (m *Manager) IsAccelerated() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeBackend != BackendCPU
}

// detectROCm detects AMD ROCm availability.
func detectROCm() *BackendInfo {
	info := &BackendInfo{
		Backend:   BackendROCm,
		Name:      "AMD ROCm",
		Available: false,
		Tested:    true, // Primary supported backend
	}

	// Check for ROCm installation
	rocmPath := os.Getenv("ROCM_PATH")
	if rocmPath == "" {
		rocmPath = "/opt/rocm"
	}

	// Check if ROCm directory exists
	if _, err := os.Stat(rocmPath); os.IsNotExist(err) {
		return nil
	}

	// Check for rocm-smi to get device info
	cmd := exec.Command("rocm-smi", "--showproductname")
	output, err := cmd.Output()
	if err != nil {
		// Try alternative detection
		cmd = exec.Command("rocminfo")
		output, err = cmd.Output()
		if err != nil {
			return nil
		}
	}

	// Parse device info
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "GPU") || strings.Contains(line, "gfx") {
			info.DeviceName = strings.TrimSpace(line)
			info.DeviceCount++
		}
	}

	if info.DeviceCount == 0 {
		// Check for any AMD GPU via /sys
		devices, _ := filepath.Glob("/sys/class/drm/card*/device/vendor")
		for _, dev := range devices {
			vendor, _ := os.ReadFile(dev)
			if strings.TrimSpace(string(vendor)) == "0x1002" { // AMD vendor ID
				info.DeviceCount++
			}
		}
	}

	if info.DeviceCount > 0 {
		info.Available = true
		info.Version = getROCmVersion(rocmPath)
		if info.DeviceName == "" {
			info.DeviceName = fmt.Sprintf("AMD GPU (%d device(s))", info.DeviceCount)
		}
	}

	return info
}

// getROCmVersion gets the ROCm version.
func getROCmVersion(rocmPath string) string {
	versionFile := filepath.Join(rocmPath, ".info", "version")
	if data, err := os.ReadFile(versionFile); err == nil {
		return strings.TrimSpace(string(data))
	}

	// Try alternative
	versionFile = filepath.Join(rocmPath, "version")
	if data, err := os.ReadFile(versionFile); err == nil {
		return strings.TrimSpace(string(data))
	}

	return "unknown"
}

// detectCUDA detects NVIDIA CUDA availability.
func detectCUDA() *BackendInfo {
	info := &BackendInfo{
		Backend:   BackendCUDA,
		Name:      "NVIDIA CUDA",
		Available: false,
		Tested:    false, // Not tested by maintainers
		Warning:   "CUDA support has not been tested by the maintainers. Please report issues at: https://github.com/MrCodeEU/facepass/issues",
	}

	// Check for nvidia-smi
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,driver_version", "--format=csv,noheader")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > 0 && lines[0] != "" {
		parts := strings.Split(lines[0], ",")
		if len(parts) >= 1 {
			info.DeviceName = strings.TrimSpace(parts[0])
		}
		if len(parts) >= 2 {
			info.Version = strings.TrimSpace(parts[1])
		}
		info.DeviceCount = len(lines)
		info.Available = true
	}

	return info
}

// detectOpenVINO detects Intel OpenVINO availability.
func detectOpenVINO() *BackendInfo {
	info := &BackendInfo{
		Backend:   BackendOpenVINO,
		Name:      "Intel OpenVINO",
		Available: false,
		Tested:    false, // Not tested by maintainers
		Warning:   "OpenVINO support has not been tested by the maintainers. Please report issues at: https://github.com/MrCodeEU/facepass/issues",
	}

	// Check for OpenVINO environment
	openvinoPath := os.Getenv("INTEL_OPENVINO_DIR")
	if openvinoPath == "" {
		// Check common installation paths
		commonPaths := []string{
			"/opt/intel/openvino",
			"/opt/intel/openvino_2024",
			"/opt/intel/openvino_2023",
		}
		for _, p := range commonPaths {
			if _, err := os.Stat(p); err == nil {
				openvinoPath = p
				break
			}
		}
	}

	if openvinoPath == "" {
		return nil
	}

	info.Available = true
	info.Version = getOpenVINOVersion(openvinoPath)

	// Detect Intel GPU/NPU
	info.DeviceName = detectIntelDevice()
	if info.DeviceName != "" {
		info.DeviceCount = 1
	}

	return info
}

// getOpenVINOVersion gets the OpenVINO version.
func getOpenVINOVersion(path string) string {
	versionFile := filepath.Join(path, "version.txt")
	if data, err := os.ReadFile(versionFile); err == nil {
		return strings.TrimSpace(string(data))
	}
	return "unknown"
}

// detectIntelDevice detects Intel GPU or NPU.
func detectIntelDevice() string {
	// Check for Intel GPU via /sys
	devices, _ := filepath.Glob("/sys/class/drm/card*/device/vendor")
	for _, dev := range devices {
		vendor, _ := os.ReadFile(dev)
		if strings.TrimSpace(string(vendor)) == "0x8086" { // Intel vendor ID
			// Try to get device name
			deviceDir := filepath.Dir(dev)
			if nameData, err := os.ReadFile(filepath.Join(deviceDir, "device")); err == nil {
				return fmt.Sprintf("Intel GPU (device: %s)", strings.TrimSpace(string(nameData)))
			}
			return "Intel GPU"
		}
	}

	// Check for NPU
	if _, err := os.Stat("/dev/accel/accel0"); err == nil {
		return "Intel NPU"
	}

	return "Intel (CPU inference)"
}

// getCPUName returns the CPU name.
func getCPUName() string {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return "Unknown CPU"
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "model name") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return "Unknown CPU"
}

// ErrBackendNotAvailable is returned when a requested backend is not available.
var ErrBackendNotAvailable = errors.New("acceleration backend not available")

// ErrNotInitialized is returned when the manager is not initialized.
var ErrNotInitialized = errors.New("acceleration manager not initialized")
