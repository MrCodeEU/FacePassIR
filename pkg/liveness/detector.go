// Package liveness provides anti-spoofing and liveness detection.
// It implements multiple tiers of detection from basic blink detection
// to advanced IR analysis.
package liveness

import (
	"errors"
)

// Level represents the liveness detection security level.
type Level string

const (
	LevelBasic    Level = "basic"    // Blink + consistency
	LevelStandard Level = "standard" // + challenge-response
	LevelStrict   Level = "strict"   // + IR analysis
	LevelParanoid Level = "paranoid" // All checks + manual review flag
)

// Config holds liveness detection configuration.
type Config struct {
	Level              Level
	RequireBlink       bool
	RequireConsistency bool
	RequireChallenge   bool
	EnableIRAnalysis   bool
	EnableTexture      bool
	MinScore           float64
	MaxTime            int // seconds
}

// Result contains the liveness detection results.
type Result struct {
	IsLive        bool
	Score         float64
	Checks        map[string]bool
	Reason        string
	RequiresRetry bool
}

// Challenge represents a challenge-response request.
type Challenge struct {
	Action string  // "turn_left", "turn_right", "look_up", "look_down"
	Angle  float64 // Expected angle in degrees
}

// Detector defines the interface for liveness detection.
type Detector interface {
	// Detect performs liveness detection on captured frames.
	Detect(frames []Frame, config Config) Result

	// DetectBlink checks for blink in the frame sequence.
	DetectBlink(frames []Frame) bool

	// CheckConsistency verifies frame-to-frame consistency.
	CheckConsistency(embeddings [][]float32) bool

	// PerformChallenge executes a challenge-response sequence.
	PerformChallenge(challenge Challenge, frames []Frame) bool

	// AnalyzeIR performs IR-specific liveness analysis.
	AnalyzeIR(irFrames []Frame) float64

	// AnalyzeTexture checks for screen/print artifacts.
	AnalyzeTexture(frame Frame) bool
}

// Frame represents a captured frame for liveness analysis.
type Frame struct {
	Data      []byte
	Landmarks []Point
	IsIR      bool
}

// Point represents a 2D point.
type Point struct {
	X, Y float64
}

// ErrLivenessFailed is returned when liveness check fails.
var ErrLivenessFailed = errors.New("liveness check failed")

// ErrNoBlink is returned when no blink was detected.
var ErrNoBlink = errors.New("no blink detected")

// ErrStaticImage is returned when a static image is detected.
var ErrStaticImage = errors.New("static image detected")

// ErrChallengeFailed is returned when challenge-response fails.
var ErrChallengeFailed = errors.New("challenge-response failed")

// TODO: Implement liveness detection
// - Eye Aspect Ratio (EAR) calculation for blink detection
// - Multi-frame embedding variance analysis
// - Challenge-response with head movement tracking
// - IR reflection pattern analysis
// - FFT-based texture analysis for screen/print detection
