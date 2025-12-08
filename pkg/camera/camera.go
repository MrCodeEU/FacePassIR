// Package camera provides camera access and frame capture functionality.
// It supports both regular RGB cameras and IR cameras with emitter control.
package camera

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/MrCodeEU/facepass/pkg/logging"
)

// Frame represents a single camera frame.
type Frame struct {
	Data      []byte
	Width     int
	Height    int
	Format    string // "JPEG", "RGB", "GRAY"
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

// ErrCameraNotFound is returned when the camera device is not found.
var ErrCameraNotFound = errors.New("camera device not found")

// ErrCameraNotOpen is returned when trying to capture from a closed camera.
var ErrCameraNotOpen = errors.New("camera not open")

// ErrNoFrame is returned when no frame could be captured.
var ErrNoFrame = errors.New("failed to capture frame")

// ErrCaptureTimeout is returned when frame capture times out.
var ErrCaptureTimeout = errors.New("capture timeout")

// V4L2Camera implements camera access using v4l2 tools.
type V4L2Camera struct {
	device     string
	width      int
	height     int
	isOpen     bool
	irEmitter  *IREmitter
	deviceInfo DeviceInfo
}

// IREmitter represents an IR emitter control interface.
type IREmitter struct {
	Available bool
	Device    string
	Enabled   bool
	Tool      string // "linux-enable-ir-emitter" or "sysfs"
}

// NewCamera creates a new V4L2Camera instance.
func NewCamera() *V4L2Camera {
	return &V4L2Camera{
		width:  640,
		height: 480,
	}
}

// Open opens the camera device.
func (c *V4L2Camera) Open(device string) error {
	// Check if device exists
	if _, err := os.Stat(device); os.IsNotExist(err) {
		return ErrCameraNotFound
	}

	c.device = device
	c.isOpen = true

	// Get device info
	c.deviceInfo = c.getDeviceInfo()

	// Detect IR emitter
	c.irEmitter = detectIREmitter()

	logging.Infof("Opened camera: %s", device)
	if c.irEmitter != nil && c.irEmitter.Available {
		logging.Info("IR emitter detected")
	}

	return nil
}

// Close closes the camera device.
func (c *V4L2Camera) Close() error {
	if c.irEmitter != nil && c.irEmitter.Enabled {
		c.DisableIREmitter()
	}
	c.isOpen = false
	logging.Debug("Camera closed")
	return nil
}

// IsOpen returns true if the camera is open.
func (c *V4L2Camera) IsOpen() bool {
	return c.isOpen
}

// SetResolution sets the capture resolution.
func (c *V4L2Camera) SetResolution(width, height int) error {
	c.width = width
	c.height = height
	return nil
}

// GetDeviceInfo returns information about the camera device.
func (c *V4L2Camera) GetDeviceInfo() DeviceInfo {
	return c.deviceInfo
}

// getDeviceInfo queries the camera device for information.
func (c *V4L2Camera) getDeviceInfo() DeviceInfo {
	info := DeviceInfo{
		Path: c.device,
	}

	// Try to get device info using v4l2-ctl
	cmd := exec.Command("v4l2-ctl", "-d", c.device, "--info")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Card type") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					info.Name = strings.TrimSpace(parts[1])
				}
			}
			if strings.Contains(line, "Driver name") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					info.Driver = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	// Check if this is an IR camera (heuristic based on name)
	nameLower := strings.ToLower(info.Name)
	info.IsIR = strings.Contains(nameLower, "ir") ||
		strings.Contains(nameLower, "infrared") ||
		strings.Contains(nameLower, "depth")

	return info
}

// Capture captures a single frame from the camera.
func (c *V4L2Camera) Capture() (*Frame, error) {
	if !c.isOpen {
		return nil, ErrCameraNotOpen
	}

	// Create temporary file for captured frame
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("facepass_frame_%d.jpg", time.Now().UnixNano()))
	defer os.Remove(tmpFile)

	// Use ffmpeg to capture a single frame
	// This is more reliable than direct v4l2 access in Go
	cmd := exec.Command("ffmpeg",
		"-f", "v4l2",
		"-video_size", fmt.Sprintf("%dx%d", c.width, c.height),
		"-i", c.device,
		"-frames:v", "1",
		"-y", // Overwrite output file
		tmpFile,
	)

	// Suppress ffmpeg output
	cmd.Stdout = nil
	cmd.Stderr = nil

	// Run with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			// Try alternative method using v4l2-ctl + convert
			return c.captureAlternative()
		}
	case <-time.After(5 * time.Second):
		cmd.Process.Kill()
		return nil, ErrCaptureTimeout
	}

	// Read the captured frame
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read captured frame: %w", err)
	}

	// Decode to get dimensions
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode frame: %w", err)
	}

	bounds := img.Bounds()

	return &Frame{
		Data:      data,
		Width:     bounds.Dx(),
		Height:    bounds.Dy(),
		Format:    "JPEG",
		Timestamp: time.Now(),
	}, nil
}

// captureAlternative tries an alternative capture method.
func (c *V4L2Camera) captureAlternative() (*Frame, error) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("facepass_frame_%d.ppm", time.Now().UnixNano()))
	jpgFile := filepath.Join(tmpDir, fmt.Sprintf("facepass_frame_%d.jpg", time.Now().UnixNano()))
	defer os.Remove(tmpFile)
	defer os.Remove(jpgFile)

	// Try using v4l2-ctl to capture a raw frame
	cmd := exec.Command("v4l2-ctl",
		"-d", c.device,
		"--set-fmt-video=width="+fmt.Sprintf("%d", c.width)+",height="+fmt.Sprintf("%d", c.height)+",pixelformat=YUYV",
		"--stream-mmap",
		"--stream-count=1",
		"--stream-to="+tmpFile,
	)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("v4l2-ctl capture failed: %w", err)
	}

	// Convert to JPEG using ImageMagick if available
	convertCmd := exec.Command("convert", tmpFile, jpgFile)
	if err := convertCmd.Run(); err != nil {
		return nil, fmt.Errorf("image conversion failed: %w", err)
	}

	data, err := os.ReadFile(jpgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read converted frame: %w", err)
	}

	return &Frame{
		Data:      data,
		Width:     c.width,
		Height:    c.height,
		Format:    "JPEG",
		Timestamp: time.Now(),
	}, nil
}

// CaptureMultiple captures multiple frames over a duration.
func (c *V4L2Camera) CaptureMultiple(count int, interval time.Duration) ([]*Frame, error) {
	frames := make([]*Frame, 0, count)

	for i := 0; i < count; i++ {
		frame, err := c.Capture()
		if err != nil {
			logging.Warnf("Failed to capture frame %d: %v", i, err)
			continue
		}
		frames = append(frames, frame)

		if i < count-1 {
			time.Sleep(interval)
		}
	}

	if len(frames) == 0 {
		return nil, ErrNoFrame
	}

	return frames, nil
}

// EnableIREmitter enables the IR emitter if available.
func (c *V4L2Camera) EnableIREmitter() error {
	if c.irEmitter == nil || !c.irEmitter.Available {
		return nil
	}

	logging.Debug("Enabling IR emitter")

	// Try linux-enable-ir-emitter first
	if c.irEmitter.Tool == "linux-enable-ir-emitter" {
		cmd := exec.Command("linux-enable-ir-emitter", "run")
		if err := cmd.Run(); err == nil {
			c.irEmitter.Enabled = true
			logging.Info("IR emitter enabled via linux-enable-ir-emitter")
			return nil
		}
	}

	// Try sysfs control
	if c.irEmitter.Device != "" {
		if err := os.WriteFile(c.irEmitter.Device, []byte("1"), 0644); err == nil {
			c.irEmitter.Enabled = true
			logging.Info("IR emitter enabled via sysfs")
			return nil
		}
	}

	return errors.New("failed to enable IR emitter")
}

// DisableIREmitter disables the IR emitter.
func (c *V4L2Camera) DisableIREmitter() error {
	if c.irEmitter == nil || !c.irEmitter.Enabled {
		return nil
	}

	logging.Debug("Disabling IR emitter")

	if c.irEmitter.Tool == "linux-enable-ir-emitter" {
		cmd := exec.Command("linux-enable-ir-emitter", "run", "--disable")
		cmd.Run() // Ignore errors
	}

	if c.irEmitter.Device != "" {
		os.WriteFile(c.irEmitter.Device, []byte("0"), 0644)
	}

	c.irEmitter.Enabled = false
	return nil
}

// HasIREmitter returns true if IR emitter is available.
func (c *V4L2Camera) HasIREmitter() bool {
	return c.irEmitter != nil && c.irEmitter.Available
}

// detectIREmitter detects if an IR emitter is available.
func detectIREmitter() *IREmitter {
	emitter := &IREmitter{
		Available: false,
	}

	// Check for linux-enable-ir-emitter
	if _, err := exec.LookPath("linux-enable-ir-emitter"); err == nil {
		emitter.Available = true
		emitter.Tool = "linux-enable-ir-emitter"
		logging.Debug("Found linux-enable-ir-emitter")
		return emitter
	}

	// Check for sysfs IR emitter control
	devices, err := filepath.Glob("/sys/class/video4linux/video*/device/ir_emitter")
	if err == nil && len(devices) > 0 {
		emitter.Available = true
		emitter.Device = devices[0]
		emitter.Tool = "sysfs"
		logging.Debugf("Found IR emitter sysfs control: %s", devices[0])
		return emitter
	}

	return emitter
}

// ListCameras returns a list of available camera devices.
func ListCameras() ([]DeviceInfo, error) {
	var cameras []DeviceInfo

	// List video devices
	devices, err := filepath.Glob("/dev/video*")
	if err != nil {
		return nil, err
	}

	for _, device := range devices {
		cam := NewCamera()
		if err := cam.Open(device); err != nil {
			continue
		}
		cameras = append(cameras, cam.GetDeviceInfo())
		cam.Close()
	}

	return cameras, nil
}

// ToImage converts a Frame to a Go image.Image.
func (f *Frame) ToImage() (image.Image, error) {
	return jpeg.Decode(bytes.NewReader(f.Data))
}
