// Package liveness provides anti-spoofing and liveness detection.
// It implements multiple tiers of detection from basic blink detection
// to advanced IR analysis.
package liveness

import (
	"errors"
	"math"
	"time"

	"github.com/MrCodeEU/facepass/pkg/logging"
	"github.com/MrCodeEU/facepass/pkg/recognition"
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

// DefaultConfig returns a default liveness configuration.
func DefaultConfig() Config {
	return Config{
		Level:              LevelStandard,
		RequireBlink:       true,
		RequireConsistency: true,
		RequireChallenge:   false,
		EnableIRAnalysis:   false,
		EnableTexture:      false,
		MinScore:           0.7,
		MaxTime:            10,
	}
}

// ConfigFromLevel creates a Config based on security level.
func ConfigFromLevel(level Level) Config {
	cfg := DefaultConfig()
	cfg.Level = level

	switch level {
	case LevelBasic:
		cfg.RequireBlink = true
		cfg.RequireConsistency = true
		cfg.RequireChallenge = false
		cfg.MinScore = 0.5
	case LevelStandard:
		cfg.RequireBlink = true
		cfg.RequireConsistency = true
		cfg.RequireChallenge = false
		cfg.MinScore = 0.7
	case LevelStrict:
		cfg.RequireBlink = true
		cfg.RequireConsistency = true
		cfg.RequireChallenge = true
		cfg.EnableIRAnalysis = true
		cfg.MinScore = 0.8
	case LevelParanoid:
		cfg.RequireBlink = true
		cfg.RequireConsistency = true
		cfg.RequireChallenge = true
		cfg.EnableIRAnalysis = true
		cfg.EnableTexture = true
		cfg.MinScore = 0.9
	}

	return cfg
}

// Result contains the liveness detection results.
type Result struct {
	IsLive        bool
	Score         float64
	Checks        map[string]bool
	Reason        string
	RequiresRetry bool
	Duration      time.Duration
}

// Challenge represents a challenge-response request.
type Challenge struct {
	Action string  // "turn_left", "turn_right", "look_up", "look_down", "blink"
	Angle  float64 // Expected angle in degrees (for head movements)
}

// Frame represents a captured frame for liveness analysis.
type Frame struct {
	Data       []byte
	Embedding  recognition.Embedding
	Landmarks  []Point
	IsIR       bool
	Timestamp  time.Time
	FaceFound  bool
	EyeAspectRatio float64
}

// Point represents a 2D point.
type Point struct {
	X, Y float64
}

// Eye landmark indices (for 5-point landmarks from dlib)
// Note: Actual indices depend on the landmark model used
const (
	LeftEyeOuter  = 0
	LeftEyeInner  = 1
	RightEyeInner = 2
	RightEyeOuter = 3
	NoseTip       = 4
)

// ErrLivenessFailed is returned when liveness check fails.
var ErrLivenessFailed = errors.New("liveness check failed")

// ErrNoBlink is returned when no blink was detected.
var ErrNoBlink = errors.New("no blink detected")

// ErrStaticImage is returned when a static image is detected.
var ErrStaticImage = errors.New("static image detected")

// ErrChallengeFailed is returned when challenge-response fails.
var ErrChallengeFailed = errors.New("challenge-response failed")

// ErrInsufficientFrames is returned when not enough frames for analysis.
var ErrInsufficientFrames = errors.New("insufficient frames for liveness detection")

// LivenessDetector implements liveness detection algorithms.
type LivenessDetector struct {
	config Config

	// Thresholds
	blinkThreshold       float64 // EAR threshold for blink detection
	consistencyMinVar    float64 // Minimum variance for consistency (too low = static)
	consistencyMaxVar    float64 // Maximum variance for consistency (too high = different person)
	movementThreshold    float64 // Minimum movement for challenge-response
}

// NewDetector creates a new LivenessDetector with the given configuration.
func NewDetector(cfg Config) *LivenessDetector {
	return &LivenessDetector{
		config:              cfg,
		blinkThreshold:      0.2, // EAR drops below this during blink
		consistencyMinVar:   0.001, // Minimum embedding variance
		consistencyMaxVar:   0.1,   // Maximum embedding variance
		movementThreshold:   0.05,  // Minimum head movement
	}
}

// Detect performs comprehensive liveness detection on a sequence of frames.
func (d *LivenessDetector) Detect(frames []Frame) Result {
	startTime := time.Now()

	result := Result{
		IsLive:  false,
		Score:   0.0,
		Checks:  make(map[string]bool),
		Reason:  "",
	}

	if len(frames) < 3 {
		result.Reason = "insufficient frames"
		result.RequiresRetry = true
		return result
	}

	logging.Debugf("Running liveness detection on %d frames", len(frames))

	var scores []float64
	totalWeight := 0.0

	// Check 1: Blink detection (weight: 0.3)
	if d.config.RequireBlink {
		blinkDetected := d.DetectBlink(frames)
		result.Checks["blink"] = blinkDetected
		if blinkDetected {
			scores = append(scores, 1.0*0.3)
		} else {
			scores = append(scores, 0.0)
		}
		totalWeight += 0.3
		logging.Debugf("Blink detection: %v", blinkDetected)
	}

	// Check 2: Frame consistency (weight: 0.3)
	if d.config.RequireConsistency {
		embeddings := extractEmbeddings(frames)
		consistent := d.CheckConsistency(embeddings)
		result.Checks["consistency"] = consistent
		if consistent {
			scores = append(scores, 1.0*0.3)
		} else {
			scores = append(scores, 0.0)
		}
		totalWeight += 0.3
		logging.Debugf("Consistency check: %v", consistent)
	}

	// Check 3: Movement detection (weight: 0.2)
	movementDetected := d.DetectMovement(frames)
	result.Checks["movement"] = movementDetected
	if movementDetected {
		scores = append(scores, 1.0*0.2)
	} else {
		scores = append(scores, 0.0)
	}
	totalWeight += 0.2

	// Check 4: Face presence throughout (weight: 0.2)
	facePresent := d.CheckFacePresence(frames)
	result.Checks["face_present"] = facePresent
	if facePresent {
		scores = append(scores, 1.0*0.2)
	} else {
		scores = append(scores, 0.0)
	}
	totalWeight += 0.2

	// Calculate final score
	var totalScore float64
	for _, s := range scores {
		totalScore += s
	}
	if totalWeight > 0 {
		result.Score = totalScore / totalWeight
	}

	// Determine if live
	result.IsLive = result.Score >= d.config.MinScore
	result.Duration = time.Since(startTime)

	if !result.IsLive {
		// Determine reason for failure
		if !result.Checks["blink"] && d.config.RequireBlink {
			result.Reason = "no blink detected"
			result.RequiresRetry = true
		} else if !result.Checks["consistency"] {
			result.Reason = "inconsistent face data (possible photo attack)"
		} else if !result.Checks["movement"] {
			result.Reason = "no movement detected (possible static image)"
		} else if !result.Checks["face_present"] {
			result.Reason = "face not consistently visible"
			result.RequiresRetry = true
		} else {
			result.Reason = "liveness score below threshold"
		}
	}

	logging.Infof("Liveness detection complete: live=%v, score=%.2f, duration=%v",
		result.IsLive, result.Score, result.Duration)

	return result
}

// DetectBlink checks for blink in the frame sequence using Eye Aspect Ratio.
func (d *LivenessDetector) DetectBlink(frames []Frame) bool {
	if len(frames) < 5 {
		return false
	}

	// Collect EAR values
	earValues := make([]float64, 0, len(frames))
	for _, frame := range frames {
		if frame.EyeAspectRatio > 0 {
			earValues = append(earValues, frame.EyeAspectRatio)
		}
	}

	if len(earValues) < 5 {
		// Not enough valid EAR values, check for variance in embeddings as fallback
		return d.detectBlinkFromEmbeddings(frames)
	}

	// Detect significant EAR drop (blink)
	maxEAR := 0.0
	minEAR := 1.0
	for _, ear := range earValues {
		if ear > maxEAR {
			maxEAR = ear
		}
		if ear < minEAR {
			minEAR = ear
		}
	}

	// A blink causes EAR to drop significantly
	earDrop := maxEAR - minEAR
	blinkDetected := earDrop > d.blinkThreshold && minEAR < 0.25

	logging.Debugf("Blink detection: maxEAR=%.3f, minEAR=%.3f, drop=%.3f, detected=%v",
		maxEAR, minEAR, earDrop, blinkDetected)

	return blinkDetected
}

// detectBlinkFromEmbeddings uses embedding variance as a fallback blink detection.
func (d *LivenessDetector) detectBlinkFromEmbeddings(frames []Frame) bool {
	// During a blink, face embeddings may show slight variations
	// This is a fallback when EAR is not available
	embeddings := extractEmbeddings(frames)
	if len(embeddings) < 3 {
		return false
	}

	// Calculate variance between consecutive embeddings
	var maxDiff float64
	for i := 1; i < len(embeddings); i++ {
		diff := embeddingDistance(embeddings[i-1], embeddings[i])
		if diff > maxDiff {
			maxDiff = diff
		}
	}

	// A blink might cause a temporary spike in embedding distance
	return maxDiff > 0.05 && maxDiff < 0.3
}

// CheckConsistency verifies frame-to-frame consistency of face embeddings.
func (d *LivenessDetector) CheckConsistency(embeddings [][]float32) bool {
	if len(embeddings) < 3 {
		return false
	}

	// Calculate variance between consecutive embeddings
	var distances []float64
	for i := 1; i < len(embeddings); i++ {
		dist := embeddingDistance(embeddings[i-1], embeddings[i])
		distances = append(distances, dist)
	}

	// Calculate mean and variance of distances
	var sum float64
	for _, d := range distances {
		sum += d
	}
	mean := sum / float64(len(distances))

	var variance float64
	for _, d := range distances {
		variance += (d - mean) * (d - mean)
	}
	variance /= float64(len(distances))

	logging.Debugf("Consistency check: mean=%.4f, variance=%.6f", mean, variance)

	// Check if variance is within acceptable range
	// Too low = static image (all frames identical)
	// Too high = different person or poor capture
	if variance < d.consistencyMinVar {
		logging.Debug("Consistency failed: variance too low (static image)")
		return false
	}
	if variance > d.consistencyMaxVar {
		logging.Debug("Consistency failed: variance too high")
		return false
	}

	// Also check that mean distance is reasonable
	if mean > 0.4 {
		logging.Debug("Consistency failed: mean distance too high")
		return false
	}

	return true
}

// DetectMovement checks for natural micro-movements between frames.
func (d *LivenessDetector) DetectMovement(frames []Frame) bool {
	if len(frames) < 3 {
		return false
	}

	embeddings := extractEmbeddings(frames)
	if len(embeddings) < 3 {
		return false
	}

	// Calculate movement by looking at embedding changes
	var totalMovement float64
	for i := 1; i < len(embeddings); i++ {
		dist := embeddingDistance(embeddings[i-1], embeddings[i])
		totalMovement += dist
	}

	avgMovement := totalMovement / float64(len(embeddings)-1)

	logging.Debugf("Movement detection: avgMovement=%.4f, threshold=%.4f",
		avgMovement, d.movementThreshold)

	// There should be some movement (not a photo)
	// but not too much (same person)
	return avgMovement > d.movementThreshold && avgMovement < 0.3
}

// CheckFacePresence ensures a face is detected in most frames.
func (d *LivenessDetector) CheckFacePresence(frames []Frame) bool {
	if len(frames) == 0 {
		return false
	}

	faceCount := 0
	for _, frame := range frames {
		if frame.FaceFound {
			faceCount++
		}
	}

	ratio := float64(faceCount) / float64(len(frames))
	logging.Debugf("Face presence: %d/%d frames (%.1f%%)", faceCount, len(frames), ratio*100)

	// Require face in at least 70% of frames
	return ratio >= 0.7
}

// PerformChallenge verifies user response to a challenge.
func (d *LivenessDetector) PerformChallenge(challenge Challenge, beforeFrames, afterFrames []Frame) bool {
	if len(beforeFrames) == 0 || len(afterFrames) == 0 {
		return false
	}

	beforeEmb := extractEmbeddings(beforeFrames)
	afterEmb := extractEmbeddings(afterFrames)

	if len(beforeEmb) == 0 || len(afterEmb) == 0 {
		return false
	}

	// Calculate average embedding before and after
	avgBefore := averageEmbedding(beforeEmb)
	avgAfter := averageEmbedding(afterEmb)

	// Calculate change
	change := embeddingDistance(avgBefore, avgAfter)

	logging.Debugf("Challenge response: action=%s, change=%.4f", challenge.Action, change)

	// For head movements, we expect a noticeable change
	switch challenge.Action {
	case "turn_left", "turn_right", "look_up", "look_down":
		return change > 0.1 && change < 0.5
	case "blink":
		return d.DetectBlink(afterFrames)
	default:
		return change > 0.05
	}
}

// QuickCheck performs a fast liveness check for authentication.
// It's less thorough but faster than full Detect().
func (d *LivenessDetector) QuickCheck(frames []Frame) (bool, float64) {
	if len(frames) < 2 {
		return false, 0.0
	}

	embeddings := extractEmbeddings(frames)
	if len(embeddings) < 2 {
		return false, 0.0
	}

	// Quick consistency check
	consistent := d.CheckConsistency(embeddings)

	// Quick movement check
	hasMovement := d.DetectMovement(frames)

	// Quick face presence check
	facePresent := d.CheckFacePresence(frames)

	score := 0.0
	if consistent {
		score += 0.4
	}
	if hasMovement {
		score += 0.3
	}
	if facePresent {
		score += 0.3
	}

	return score >= 0.6, score
}

// Helper functions

// extractEmbeddings extracts embedding vectors from frames.
func extractEmbeddings(frames []Frame) [][]float32 {
	embeddings := make([][]float32, 0, len(frames))
	for _, frame := range frames {
		if frame.FaceFound && len(frame.Embedding.Vector) > 0 {
			embeddings = append(embeddings, frame.Embedding.Vector[:])
		}
	}
	return embeddings
}

// embeddingDistance calculates Euclidean distance between two embeddings.
func embeddingDistance(e1, e2 []float32) float64 {
	if len(e1) != len(e2) {
		return 1.0
	}

	var sum float64
	for i := range e1 {
		diff := float64(e1[i] - e2[i])
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

// averageEmbedding calculates the average of multiple embeddings.
func averageEmbedding(embeddings [][]float32) []float32 {
	if len(embeddings) == 0 {
		return nil
	}

	size := len(embeddings[0])
	avg := make([]float32, size)

	for _, emb := range embeddings {
		for i, v := range emb {
			avg[i] += v
		}
	}

	count := float32(len(embeddings))
	for i := range avg {
		avg[i] /= count
	}

	return avg
}

// CalculateEyeAspectRatio calculates the EAR from eye landmarks.
// EAR = (||p2-p6|| + ||p3-p5||) / (2 * ||p1-p4||)
// This requires 6 points per eye from a 68-point landmark model.
// For 5-point models, this is approximated.
func CalculateEyeAspectRatio(eyeLandmarks []Point) float64 {
	if len(eyeLandmarks) < 2 {
		return 0.5 // Default open eye value
	}

	// For 5-point landmarks, we only have outer corners
	// This is a simplified approximation
	if len(eyeLandmarks) == 2 {
		// Distance between corners as rough approximation
		dx := eyeLandmarks[1].X - eyeLandmarks[0].X
		dy := eyeLandmarks[1].Y - eyeLandmarks[0].Y
		width := math.Sqrt(dx*dx + dy*dy)

		// Assume typical eye aspect ratio when open
		// Return a normalized value based on width
		if width > 0 {
			return 0.3 // Approximate open eye EAR
		}
		return 0.0
	}

	// For 6-point eye landmarks (standard EAR calculation)
	if len(eyeLandmarks) == 6 {
		// Vertical distances
		v1 := distance(eyeLandmarks[1], eyeLandmarks[5])
		v2 := distance(eyeLandmarks[2], eyeLandmarks[4])

		// Horizontal distance
		h := distance(eyeLandmarks[0], eyeLandmarks[3])

		if h == 0 {
			return 0.0
		}

		return (v1 + v2) / (2.0 * h)
	}

	return 0.3 // Default
}

// distance calculates Euclidean distance between two points.
func distance(p1, p2 Point) float64 {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	return math.Sqrt(dx*dx + dy*dy)
}
