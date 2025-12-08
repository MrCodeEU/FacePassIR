// Package config provides configuration management for FacePass.
// It loads configuration from YAML files with sensible defaults.
package config

import (
	"os"
	"path/filepath"

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
