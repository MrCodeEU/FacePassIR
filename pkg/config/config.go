// Package config provides configuration management for FacePass.
// It loads configuration from YAML files with sensible defaults.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds all FacePass configuration.
type Config struct {
	Camera    CameraConfig    `yaml:"camera"`
	Recognition RecognitionConfig `yaml:"recognition"`
	Liveness  LivenessConfig  `yaml:"liveness_detection"`
	Auth      AuthConfig      `yaml:"auth"`
	Storage   StorageConfig   `yaml:"storage"`
	Logging   LoggingConfig   `yaml:"logging"`
}

// CameraConfig holds camera settings.
type CameraConfig struct {
	Device           string `yaml:"device"`
	Width            int    `yaml:"width"`
	Height           int    `yaml:"height"`
	FPS              int    `yaml:"fps"`
	PreferIR         bool   `yaml:"prefer_ir"`
	IRDevice         string `yaml:"ir_device"`
	RGBDevice        string `yaml:"rgb_device"`
	IREmitterEnabled bool   `yaml:"ir_emitter_enabled"`
	IREmitterTool    string `yaml:"ir_emitter_tool"`
}

// RecognitionConfig holds face recognition settings.
type RecognitionConfig struct {
	ConfidenceThreshold float64 `yaml:"confidence_threshold"`
	Tolerance           float64 `yaml:"tolerance"`
	ModelPath           string  `yaml:"model_path"`
}

// LivenessConfig holds liveness detection settings.
type LivenessConfig struct {
	Level             string  `yaml:"level"`
	BlinkRequired     bool    `yaml:"blink_required"`
	ConsistencyCheck  bool    `yaml:"consistency_check"`
	ChallengeResponse bool    `yaml:"challenge_response"`
	IRAnalysis        bool    `yaml:"ir_analysis"`
	TextureAnalysis   bool    `yaml:"texture_analysis"`
	MinLivenessScore  float64 `yaml:"min_liveness_score"`
	MaxAuthTime       int     `yaml:"max_authentication_time"`
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	Timeout         int  `yaml:"timeout"`
	MaxAttempts     int  `yaml:"max_attempts"`
	FallbackEnabled bool `yaml:"fallback_enabled"`
}

// StorageConfig holds storage settings.
type StorageConfig struct {
	DataDir           string `yaml:"data_dir"`
	EncryptionEnabled bool   `yaml:"encryption_enabled"`
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		Camera: CameraConfig{
			Device:           "/dev/video0",
			Width:            640,
			Height:           480,
			FPS:              30,
			PreferIR:         true,
			IRDevice:         "/dev/video2",
			RGBDevice:        "/dev/video0",
			IREmitterEnabled: true,
			IREmitterTool:    "linux-enable-ir-emitter",
		},
		Recognition: RecognitionConfig{
			ConfidenceThreshold: 0.6,
			Tolerance:           0.4,
			ModelPath:           filepath.Join(homeDir, ".local/share/facepass/models"),
		},
		Liveness: LivenessConfig{
			Level:             "standard",
			BlinkRequired:     true,
			ConsistencyCheck:  true,
			ChallengeResponse: false,
			IRAnalysis:        true,
			TextureAnalysis:   true,
			MinLivenessScore:  0.7,
			MaxAuthTime:       10,
		},
		Auth: AuthConfig{
			Timeout:         10,
			MaxAttempts:     3,
			FallbackEnabled: true,
		},
		Storage: StorageConfig{
			DataDir:           filepath.Join(homeDir, ".local/share/facepass"),
			EncryptionEnabled: true,
		},
		Logging: LoggingConfig{
			Level: "info",
			File:  filepath.Join(homeDir, ".local/share/facepass/facepass.log"),
		},
	}
}

// Load loads configuration from the specified file.
func Load(path string) (*Config, error) {
	config := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return config, err
	}

	return config, nil
}

// LoadDefault tries to load configuration from default locations.
func LoadDefault() (*Config, error) {
	// Try system config first
	if _, err := os.Stat("/etc/facepass/facepass.yaml"); err == nil {
		return Load("/etc/facepass/facepass.yaml")
	}

	// Try user config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultConfig(), nil
	}

	userConfig := filepath.Join(homeDir, ".config/facepass/facepass.yaml")
	if _, err := os.Stat(userConfig); err == nil {
		return Load(userConfig)
	}

	// Return defaults
	return DefaultConfig(), nil
}

// ExpandPath expands ~ and environment variables in a path.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	}
	return os.ExpandEnv(path)
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	// Validate camera settings
	if c.Camera.Width <= 0 || c.Camera.Height <= 0 {
		return fmt.Errorf("invalid camera resolution: %dx%d", c.Camera.Width, c.Camera.Height)
	}
	if c.Camera.FPS <= 0 {
		return fmt.Errorf("invalid camera FPS: %d", c.Camera.FPS)
	}

	// Validate recognition settings
	if c.Recognition.ConfidenceThreshold < 0 || c.Recognition.ConfidenceThreshold > 1 {
		return fmt.Errorf("confidence_threshold must be between 0 and 1, got %f", c.Recognition.ConfidenceThreshold)
	}
	if c.Recognition.Tolerance < 0 || c.Recognition.Tolerance > 1 {
		return fmt.Errorf("tolerance must be between 0 and 1, got %f", c.Recognition.Tolerance)
	}

	// Validate liveness settings
	validLevels := map[string]bool{"basic": true, "standard": true, "strict": true, "paranoid": true}
	if !validLevels[c.Liveness.Level] {
		return fmt.Errorf("invalid liveness level: %s (must be basic, standard, strict, or paranoid)", c.Liveness.Level)
	}
	if c.Liveness.MinLivenessScore < 0 || c.Liveness.MinLivenessScore > 1 {
		return fmt.Errorf("min_liveness_score must be between 0 and 1, got %f", c.Liveness.MinLivenessScore)
	}

	// Validate auth settings
	if c.Auth.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %d", c.Auth.Timeout)
	}
	if c.Auth.MaxAttempts <= 0 {
		return fmt.Errorf("max_attempts must be positive, got %d", c.Auth.MaxAttempts)
	}

	// Validate logging level
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.Logging.Level)
	}

	return nil
}

// ExpandPaths expands all paths in the configuration.
func (c *Config) ExpandPaths() {
	c.Camera.Device = ExpandPath(c.Camera.Device)
	c.Camera.IRDevice = ExpandPath(c.Camera.IRDevice)
	c.Camera.RGBDevice = ExpandPath(c.Camera.RGBDevice)
	c.Recognition.ModelPath = ExpandPath(c.Recognition.ModelPath)
	c.Storage.DataDir = ExpandPath(c.Storage.DataDir)
	c.Logging.File = ExpandPath(c.Logging.File)
}

// EnsureDirectories creates necessary directories for storage and logging.
func (c *Config) EnsureDirectories() error {
	// Create storage directory
	if err := os.MkdirAll(c.Storage.DataDir, 0700); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Create users subdirectory
	usersDir := filepath.Join(c.Storage.DataDir, "users")
	if err := os.MkdirAll(usersDir, 0700); err != nil {
		return fmt.Errorf("failed to create users directory: %w", err)
	}

	// Create models directory
	if err := os.MkdirAll(c.Recognition.ModelPath, 0755); err != nil {
		return fmt.Errorf("failed to create models directory: %w", err)
	}

	// Create log directory
	logDir := filepath.Dir(c.Logging.File)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	return nil
}

// GetUserDataPath returns the path for a user's face data file.
func (c *Config) GetUserDataPath(username string) string {
	return filepath.Join(c.Storage.DataDir, "users", username+".json")
}
