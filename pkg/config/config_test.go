package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Check camera defaults
	if cfg.Camera.Device != "/dev/video0" {
		t.Errorf("expected camera device /dev/video0, got %s", cfg.Camera.Device)
	}
	if cfg.Camera.Width != 640 {
		t.Errorf("expected camera width 640, got %d", cfg.Camera.Width)
	}
	if cfg.Camera.Height != 480 {
		t.Errorf("expected camera height 480, got %d", cfg.Camera.Height)
	}
	if cfg.Camera.FPS != 30 {
		t.Errorf("expected camera FPS 30, got %d", cfg.Camera.FPS)
	}

	// Check recognition defaults
	if cfg.Recognition.ConfidenceThreshold != 0.6 {
		t.Errorf("expected confidence threshold 0.6, got %f", cfg.Recognition.ConfidenceThreshold)
	}
	if cfg.Recognition.Tolerance != 0.4 {
		t.Errorf("expected tolerance 0.4, got %f", cfg.Recognition.Tolerance)
	}

	// Check auth defaults
	if cfg.Auth.Timeout != 10 {
		t.Errorf("expected timeout 10, got %d", cfg.Auth.Timeout)
	}
	if cfg.Auth.MaxAttempts != 3 {
		t.Errorf("expected max attempts 3, got %d", cfg.Auth.MaxAttempts)
	}
	if !cfg.Auth.FallbackEnabled {
		t.Error("expected fallback to be enabled by default")
	}

	// Check liveness defaults
	if cfg.Liveness.Level != "standard" {
		t.Errorf("expected liveness level 'standard', got %s", cfg.Liveness.Level)
	}
	if !cfg.Liveness.BlinkRequired {
		t.Error("expected blink to be required by default")
	}

	// Check storage defaults
	if !cfg.Storage.EncryptionEnabled {
		t.Error("expected encryption to be enabled by default")
	}

	// Check logging defaults
	if cfg.Logging.Level != "info" {
		t.Errorf("expected log level 'info', got %s", cfg.Logging.Level)
	}
}

func TestLoad(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	configContent := `
camera:
  device: /dev/video1
  width: 1280
  height: 720
  fps: 60

recognition:
  confidence_threshold: 0.8
  tolerance: 0.3
  model_path: /custom/models

auth:
  timeout: 15
  max_attempts: 5
  fallback_enabled: false

liveness_detection:
  level: strict
  blink_required: true
  min_liveness_score: 0.8

storage:
  data_dir: /custom/data
  encryption_enabled: true

logging:
  level: debug
  file: /var/log/facepass.log
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Test loading
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify loaded values
	if cfg.Camera.Device != "/dev/video1" {
		t.Errorf("expected camera device /dev/video1, got %s", cfg.Camera.Device)
	}
	if cfg.Camera.Width != 1280 {
		t.Errorf("expected camera width 1280, got %d", cfg.Camera.Width)
	}
	if cfg.Camera.Height != 720 {
		t.Errorf("expected camera height 720, got %d", cfg.Camera.Height)
	}
	if cfg.Recognition.ConfidenceThreshold != 0.8 {
		t.Errorf("expected confidence threshold 0.8, got %f", cfg.Recognition.ConfidenceThreshold)
	}
	if cfg.Auth.Timeout != 15 {
		t.Errorf("expected timeout 15, got %d", cfg.Auth.Timeout)
	}
	if cfg.Auth.FallbackEnabled {
		t.Error("expected fallback to be disabled")
	}
	if cfg.Liveness.Level != "strict" {
		t.Errorf("expected liveness level 'strict', got %s", cfg.Liveness.Level)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("expected log level 'debug', got %s", cfg.Logging.Level)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")

	// Should return default config with error
	if cfg == nil {
		t.Error("expected default config on error")
	}
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	if err := os.WriteFile(configPath, []byte("invalid: [yaml: content"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg, err := Load(configPath)
	if cfg == nil {
		t.Error("expected default config on error")
	}
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadDefault(t *testing.T) {
	// This should return defaults since no config files exist in test environment
	cfg, err := LoadDefault()

	if cfg == nil {
		t.Fatal("LoadDefault returned nil")
	}
	// Error might be nil if returning defaults
	_ = err

	// Verify it has default values
	if cfg.Camera.Width != 640 {
		t.Errorf("expected default camera width 640, got %d", cfg.Camera.Width)
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "tilde expansion",
			input:    "~/test/path",
			contains: "/test/path",
		},
		{
			name:     "no expansion needed",
			input:    "/absolute/path",
			contains: "/absolute/path",
		},
		{
			name:     "relative path",
			input:    "relative/path",
			contains: "relative/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandPath(tt.input)
			if tt.input == "~/test/path" {
				// Should not contain tilde anymore
				if result[0] == '~' {
					t.Error("tilde was not expanded")
				}
			}
			if tt.input != "~/test/path" && result != tt.input {
				t.Errorf("unexpected expansion: got %s", result)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		modify    func(*Config)
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid default config",
			modify:    func(c *Config) {},
			wantError: false,
		},
		{
			name: "invalid camera width",
			modify: func(c *Config) {
				c.Camera.Width = 0
			},
			wantError: true,
			errorMsg:  "invalid camera resolution",
		},
		{
			name: "invalid camera height",
			modify: func(c *Config) {
				c.Camera.Height = -1
			},
			wantError: true,
			errorMsg:  "invalid camera resolution",
		},
		{
			name: "invalid camera FPS",
			modify: func(c *Config) {
				c.Camera.FPS = 0
			},
			wantError: true,
			errorMsg:  "invalid camera FPS",
		},
		{
			name: "confidence threshold too high",
			modify: func(c *Config) {
				c.Recognition.ConfidenceThreshold = 1.5
			},
			wantError: true,
			errorMsg:  "confidence_threshold must be between 0 and 1",
		},
		{
			name: "confidence threshold negative",
			modify: func(c *Config) {
				c.Recognition.ConfidenceThreshold = -0.1
			},
			wantError: true,
			errorMsg:  "confidence_threshold must be between 0 and 1",
		},
		{
			name: "tolerance too high",
			modify: func(c *Config) {
				c.Recognition.Tolerance = 2.0
			},
			wantError: true,
			errorMsg:  "tolerance must be between 0 and 1",
		},
		{
			name: "invalid liveness level",
			modify: func(c *Config) {
				c.Liveness.Level = "invalid"
			},
			wantError: true,
			errorMsg:  "invalid liveness level",
		},
		{
			name: "valid liveness level basic",
			modify: func(c *Config) {
				c.Liveness.Level = "basic"
			},
			wantError: false,
		},
		{
			name: "valid liveness level paranoid",
			modify: func(c *Config) {
				c.Liveness.Level = "paranoid"
			},
			wantError: false,
		},
		{
			name: "min liveness score too high",
			modify: func(c *Config) {
				c.Liveness.MinLivenessScore = 1.5
			},
			wantError: true,
			errorMsg:  "min_liveness_score must be between 0 and 1",
		},
		{
			name: "timeout zero",
			modify: func(c *Config) {
				c.Auth.Timeout = 0
			},
			wantError: true,
			errorMsg:  "timeout must be positive",
		},
		{
			name: "max attempts zero",
			modify: func(c *Config) {
				c.Auth.MaxAttempts = 0
			},
			wantError: true,
			errorMsg:  "max_attempts must be positive",
		},
		{
			name: "invalid log level",
			modify: func(c *Config) {
				c.Logging.Level = "invalid"
			},
			wantError: true,
			errorMsg:  "invalid log level",
		},
		{
			name: "valid log level debug",
			modify: func(c *Config) {
				c.Logging.Level = "debug"
			},
			wantError: false,
		},
		{
			name: "valid log level warn",
			modify: func(c *Config) {
				c.Logging.Level = "warn"
			},
			wantError: false,
		},
		{
			name: "valid log level error",
			modify: func(c *Config) {
				c.Logging.Level = "error"
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modify(cfg)

			err := cfg.Validate()
			if tt.wantError {
				if err == nil {
					t.Error("expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("error message doesn't contain '%s': %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConfig_ExpandPaths(t *testing.T) {
	cfg := DefaultConfig()

	// Set paths with tilde
	cfg.Storage.DataDir = "~/facepass/data"
	cfg.Logging.File = "~/facepass/log.txt"

	cfg.ExpandPaths()

	// Check that tilde was expanded
	if cfg.Storage.DataDir[0] == '~' {
		t.Error("Storage.DataDir tilde was not expanded")
	}
	if cfg.Logging.File[0] == '~' {
		t.Error("Logging.File tilde was not expanded")
	}
}

func TestConfig_EnsureDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Storage.DataDir = filepath.Join(tmpDir, "data")
	cfg.Recognition.ModelPath = filepath.Join(tmpDir, "models")
	cfg.Logging.File = filepath.Join(tmpDir, "logs", "facepass.log")

	err := cfg.EnsureDirectories()
	if err != nil {
		t.Fatalf("EnsureDirectories failed: %v", err)
	}

	// Check directories were created
	if _, err := os.Stat(cfg.Storage.DataDir); os.IsNotExist(err) {
		t.Error("storage data dir was not created")
	}

	usersDir := filepath.Join(cfg.Storage.DataDir, "users")
	if _, err := os.Stat(usersDir); os.IsNotExist(err) {
		t.Error("users dir was not created")
	}

	if _, err := os.Stat(cfg.Recognition.ModelPath); os.IsNotExist(err) {
		t.Error("models dir was not created")
	}

	logDir := filepath.Dir(cfg.Logging.File)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Error("log dir was not created")
	}
}

func TestConfig_GetUserDataPath(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Storage.DataDir = "/var/lib/facepass"

	path := cfg.GetUserDataPath("testuser")
	expected := "/var/lib/facepass/users/testuser.json"

	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestConfig_GetUserDataPath_SpecialCharacters(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Storage.DataDir = "/data"

	tests := []struct {
		username string
		expected string
	}{
		{"john", "/data/users/john.json"},
		{"john.doe", "/data/users/john.doe.json"},
		{"user123", "/data/users/user123.json"},
	}

	for _, tt := range tests {
		t.Run(tt.username, func(t *testing.T) {
			path := cfg.GetUserDataPath(tt.username)
			if path != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, path)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DefaultConfig()
	}
}

func BenchmarkConfig_Validate(b *testing.B) {
	cfg := DefaultConfig()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cfg.Validate()
	}
}

func BenchmarkExpandPath(b *testing.B) {
	path := "~/test/path/to/file"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ExpandPath(path)
	}
}
