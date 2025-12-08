// Package camera provides camera access and frame capture functionality.
// It supports both regular RGB cameras and IR cameras with emitter control.
package camera

import (
	"errors"
	"time"
)

// Frame represents a single camera frame.
type Frame struct {
	Data      []byte
	Width     int
	Height    int
	Format    string // "RGB", "GRAY", "IR"
	Timestamp time.Time
}

// DeviceInfo contains information about a camera device.
type DeviceInfo struct {
	Path       string
	Name       string
	Driver     string
	IsIR       bool
	HasEmitter bool
}

// Camera defines the interface for camera operations.
type Camera interface {
	Open(device string) error
	Close() error
	Capture() (Frame, error)
	GetDeviceInfo() DeviceInfo
	SetResolution(width, height int) error
}

// IREmitter represents an IR emitter control interface.
type IREmitter struct {
	Available bool
	Device    string
	Enabled   bool
}

// ErrCameraNotFound is returned when the camera device is not found.
var ErrCameraNotFound = errors.New("camera device not found")

// ErrCameraNotOpen is returned when trying to capture from a closed camera.
var ErrCameraNotOpen = errors.New("camera not open")

// ErrNoFrame is returned when no frame could be captured.
var ErrNoFrame = errors.New("failed to capture frame")

// TODO: Implement camera functionality
// - V4L2 camera access
// - IR emitter detection and control
// - Frame capture and preprocessing
