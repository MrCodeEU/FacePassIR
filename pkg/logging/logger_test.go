package logging

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestInit(t *testing.T) {
	// Reset logger before tests
	Logger = logrus.New()

	tests := []struct {
		name    string
		level   string
		logFile string
		wantErr bool
	}{
		{
			name:    "debug level",
			level:   "debug",
			logFile: "",
			wantErr: false,
		},
		{
			name:    "info level",
			level:   "info",
			logFile: "",
			wantErr: false,
		},
		{
			name:    "warn level",
			level:   "warn",
			logFile: "",
			wantErr: false,
		},
		{
			name:    "error level",
			level:   "error",
			logFile: "",
			wantErr: false,
		},
		{
			name:    "unknown level defaults to info",
			level:   "unknown",
			logFile: "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Logger = logrus.New()
			err := Init(tt.level, tt.logFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInit_WithLogFile(t *testing.T) {
	Logger = logrus.New()
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	err := Init("info", logFile)
	if err != nil {
		t.Fatalf("Init with log file failed: %v", err)
	}

	// Check log file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("log file was not created")
	}
}

func TestInit_CreateDirectory(t *testing.T) {
	Logger = logrus.New()
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "subdir", "nested", "test.log")

	err := Init("info", logFile)
	if err != nil {
		t.Fatalf("Init with nested log file failed: %v", err)
	}

	// Check directories and file were created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("nested log file was not created")
	}
}

func TestSetLevel(t *testing.T) {
	Logger = logrus.New()

	tests := []struct {
		level    string
		expected logrus.Level
	}{
		{"debug", logrus.DebugLevel},
		{"info", logrus.InfoLevel},
		{"warn", logrus.WarnLevel},
		{"error", logrus.ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			SetLevel(tt.level)
			if Logger.GetLevel() != tt.expected {
				t.Errorf("expected level %v, got %v", tt.expected, Logger.GetLevel())
			}
		})
	}
}

func TestLoggingFunctions(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer
	Logger = logrus.New()
	Logger.SetOutput(&buf)
	Logger.SetLevel(logrus.DebugLevel)
	Logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})

	// Test Debug
	buf.Reset()
	Debug("debug message")
	if !strings.Contains(buf.String(), "debug message") {
		t.Error("Debug message not logged")
	}

	// Test Debugf
	buf.Reset()
	Debugf("debug %s", "formatted")
	if !strings.Contains(buf.String(), "debug formatted") {
		t.Error("Debugf message not logged")
	}

	// Test Info
	buf.Reset()
	Info("info message")
	if !strings.Contains(buf.String(), "info message") {
		t.Error("Info message not logged")
	}

	// Test Infof
	buf.Reset()
	Infof("info %d", 42)
	if !strings.Contains(buf.String(), "info 42") {
		t.Error("Infof message not logged")
	}

	// Test Warn
	buf.Reset()
	Warn("warn message")
	if !strings.Contains(buf.String(), "warn message") {
		t.Error("Warn message not logged")
	}

	// Test Warnf
	buf.Reset()
	Warnf("warn %s", "test")
	if !strings.Contains(buf.String(), "warn test") {
		t.Error("Warnf message not logged")
	}

	// Test Error
	buf.Reset()
	Error("error message")
	if !strings.Contains(buf.String(), "error message") {
		t.Error("Error message not logged")
	}

	// Test Errorf
	buf.Reset()
	Errorf("error %s", "occurred")
	if !strings.Contains(buf.String(), "error occurred") {
		t.Error("Errorf message not logged")
	}
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	Logger = logrus.New()
	Logger.SetOutput(&buf)
	Logger.SetLevel(logrus.InfoLevel)
	Logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})

	WithFields(Fields{
		"user": "testuser",
		"action": "login",
	}).Info("user action")

	output := buf.String()
	if !strings.Contains(output, "user=testuser") {
		t.Error("user field not in output")
	}
	if !strings.Contains(output, "action=login") {
		t.Error("action field not in output")
	}
	if !strings.Contains(output, "user action") {
		t.Error("message not in output")
	}
}

func TestWithField(t *testing.T) {
	var buf bytes.Buffer
	Logger = logrus.New()
	Logger.SetOutput(&buf)
	Logger.SetLevel(logrus.InfoLevel)
	Logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})

	WithField("key", "value").Info("test message")

	output := buf.String()
	if !strings.Contains(output, "key=value") {
		t.Error("field not in output")
	}
}

func TestWithError(t *testing.T) {
	var buf bytes.Buffer
	Logger = logrus.New()
	Logger.SetOutput(&buf)
	Logger.SetLevel(logrus.ErrorLevel)
	Logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})

	testErr := &testError{msg: "test error"}
	WithError(testErr).Error("operation failed")

	output := buf.String()
	if !strings.Contains(output, "test error") {
		t.Error("error not in output")
	}
}

func TestComponent(t *testing.T) {
	var buf bytes.Buffer
	Logger = logrus.New()
	Logger.SetOutput(&buf)
	Logger.SetLevel(logrus.InfoLevel)
	Logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})

	Component("storage").Info("initialized")

	output := buf.String()
	if !strings.Contains(output, "component=storage") {
		t.Error("component field not in output")
	}
	if !strings.Contains(output, "initialized") {
		t.Error("message not in output")
	}
}

func TestLogLevel_Filtering(t *testing.T) {
	var buf bytes.Buffer
	Logger = logrus.New()
	Logger.SetOutput(&buf)
	Logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})

	// Set to error level - should only log errors
	Logger.SetLevel(logrus.ErrorLevel)

	buf.Reset()
	Debug("debug")
	if buf.Len() > 0 {
		t.Error("Debug should not be logged at Error level")
	}

	buf.Reset()
	Info("info")
	if buf.Len() > 0 {
		t.Error("Info should not be logged at Error level")
	}

	buf.Reset()
	Warn("warn")
	if buf.Len() > 0 {
		t.Error("Warn should not be logged at Error level")
	}

	buf.Reset()
	Error("error")
	if buf.Len() == 0 {
		t.Error("Error should be logged at Error level")
	}
}

// Helper type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// Benchmark tests
func BenchmarkInfo(b *testing.B) {
	Logger = logrus.New()
	Logger.SetOutput(&bytes.Buffer{})
	Logger.SetLevel(logrus.InfoLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("benchmark message")
	}
}

func BenchmarkInfof(b *testing.B) {
	Logger = logrus.New()
	Logger.SetOutput(&bytes.Buffer{})
	Logger.SetLevel(logrus.InfoLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Infof("benchmark message %d", i)
	}
}

func BenchmarkWithFields(b *testing.B) {
	Logger = logrus.New()
	Logger.SetOutput(&bytes.Buffer{})
	Logger.SetLevel(logrus.InfoLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WithFields(Fields{
			"key1": "value1",
			"key2": "value2",
		}).Info("message")
	}
}
