// Package logging provides a centralized logging system for FacePass.
// It wraps logrus to provide consistent logging across all components.
package logging

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// Logger is the application-wide logger instance.
var Logger *logrus.Logger

// Fields is an alias for logrus.Fields for convenience.
type Fields = logrus.Fields

func init() {
	Logger = logrus.New()
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	Logger.SetOutput(os.Stderr)
	Logger.SetLevel(logrus.InfoLevel)
}

// Init initializes the logger with the specified configuration.
func Init(level string, logFile string) error {
	// Set log level
	switch level {
	case "debug":
		Logger.SetLevel(logrus.DebugLevel)
	case "info":
		Logger.SetLevel(logrus.InfoLevel)
	case "warn":
		Logger.SetLevel(logrus.WarnLevel)
	case "error":
		Logger.SetLevel(logrus.ErrorLevel)
	default:
		Logger.SetLevel(logrus.InfoLevel)
	}

	// Set up file logging if specified
	if logFile != "" {
		// Ensure directory exists
		logDir := filepath.Dir(logFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		// Open log file
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}

		// Write to both file and stderr
		Logger.SetOutput(io.MultiWriter(os.Stderr, file))
	}

	return nil
}

// SetLevel sets the logging level.
func SetLevel(level string) {
	switch level {
	case "debug":
		Logger.SetLevel(logrus.DebugLevel)
	case "info":
		Logger.SetLevel(logrus.InfoLevel)
	case "warn":
		Logger.SetLevel(logrus.WarnLevel)
	case "error":
		Logger.SetLevel(logrus.ErrorLevel)
	}
}

// Debug logs a debug message.
func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

// Debugf logs a formatted debug message.
func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

// Info logs an info message.
func Info(args ...interface{}) {
	Logger.Info(args...)
}

// Infof logs a formatted info message.
func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

// Warn logs a warning message.
func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

// Error logs an error message.
func Error(args ...interface{}) {
	Logger.Error(args...)
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

// Fatal logs a fatal message and exits.
func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits.
func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}

// WithFields returns an entry with fields attached.
func WithFields(fields Fields) *logrus.Entry {
	return Logger.WithFields(fields)
}

// WithField returns an entry with a single field attached.
func WithField(key string, value interface{}) *logrus.Entry {
	return Logger.WithField(key, value)
}

// WithError returns an entry with an error attached.
func WithError(err error) *logrus.Entry {
	return Logger.WithError(err)
}

// Component returns a logger entry for a specific component.
func Component(name string) *logrus.Entry {
	return Logger.WithField("component", name)
}
