package liveness

import (
	"testing"
	"time"

	"github.com/MrCodeEU/facepass/pkg/recognition"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Level != LevelStandard {
		t.Errorf("expected level standard, got %s", cfg.Level)
	}
	if !cfg.RequireBlink {
		t.Error("expected RequireBlink to be true")
	}
	if !cfg.RequireConsistency {
		t.Error("expected RequireConsistency to be true")
	}
	if cfg.MinScore != 0.7 {
		t.Errorf("expected MinScore 0.7, got %f", cfg.MinScore)
	}
}

func TestConfigFromLevel(t *testing.T) {
	tests := []struct {
		level           Level
		expectBlink     bool
		expectChallenge bool
		expectIR        bool
		minScore        float64
	}{
		{LevelBasic, true, false, false, 0.5},
		{LevelStandard, true, false, false, 0.7},
		{LevelStrict, true, true, true, 0.8},
		{LevelParanoid, true, true, true, 0.9},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			cfg := ConfigFromLevel(tt.level)

			if cfg.RequireBlink != tt.expectBlink {
				t.Errorf("RequireBlink: got %v, want %v", cfg.RequireBlink, tt.expectBlink)
			}
			if cfg.RequireChallenge != tt.expectChallenge {
				t.Errorf("RequireChallenge: got %v, want %v", cfg.RequireChallenge, tt.expectChallenge)
			}
			if cfg.EnableIRAnalysis != tt.expectIR {
				t.Errorf("EnableIRAnalysis: got %v, want %v", cfg.EnableIRAnalysis, tt.expectIR)
			}
			if cfg.MinScore != tt.minScore {
				t.Errorf("MinScore: got %f, want %f", cfg.MinScore, tt.minScore)
			}
		})
	}
}

func TestNewDetector(t *testing.T) {
	cfg := DefaultConfig()
	detector := NewDetector(cfg)

	if detector == nil {
		t.Fatal("NewDetector returned nil")
	}
	if detector.blinkThreshold != 0.2 {
		t.Errorf("expected blink threshold 0.2, got %f", detector.blinkThreshold)
	}
}

func TestDetector_Detect_InsufficientFrames(t *testing.T) {
	detector := NewDetector(DefaultConfig())

	// Less than 3 frames
	frames := []Frame{
		createTestFrame(true),
		createTestFrame(true),
	}

	result := detector.Detect(frames)

	if result.IsLive {
		t.Error("should not be live with insufficient frames")
	}
	if result.Reason != "insufficient frames" {
		t.Errorf("expected 'insufficient frames' reason, got %s", result.Reason)
	}
	if !result.RequiresRetry {
		t.Error("should require retry")
	}
}

func TestDetector_CheckConsistency(t *testing.T) {
	detector := NewDetector(DefaultConfig())

	tests := []struct {
		name       string
		embeddings [][]float32
		expected   bool
	}{
		{
			name:       "insufficient embeddings",
			embeddings: [][]float32{{1, 2, 3}},
			expected:   false,
		},
		{
			name: "consistent embeddings",
			embeddings: createConsistentEmbeddings(5, 0.01),
			expected:   true,
		},
		{
			name: "identical embeddings (static)",
			embeddings: createIdenticalEmbeddings(5),
			expected:   false,
		},
		{
			name: "very different embeddings",
			embeddings: createDifferentEmbeddings(5),
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.CheckConsistency(tt.embeddings)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDetector_DetectMovement(t *testing.T) {
	detector := NewDetector(DefaultConfig())

	tests := []struct {
		name     string
		frames   []Frame
		expected bool
	}{
		{
			name:     "insufficient frames",
			frames:   []Frame{createTestFrame(true)},
			expected: false,
		},
		{
			name:     "frames with movement",
			frames:   createFramesWithMovement(5, 0.1),
			expected: true,
		},
		{
			name:     "static frames",
			frames:   createStaticFrames(5),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectMovement(tt.frames)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDetector_CheckFacePresence(t *testing.T) {
	detector := NewDetector(DefaultConfig())

	tests := []struct {
		name     string
		frames   []Frame
		expected bool
	}{
		{
			name:     "empty frames",
			frames:   []Frame{},
			expected: false,
		},
		{
			name: "all faces detected",
			frames: []Frame{
				{FaceFound: true},
				{FaceFound: true},
				{FaceFound: true},
			},
			expected: true,
		},
		{
			name: "70% faces detected (threshold)",
			frames: []Frame{
				{FaceFound: true},
				{FaceFound: true},
				{FaceFound: true},
				{FaceFound: true},
				{FaceFound: true},
				{FaceFound: true},
				{FaceFound: true},
				{FaceFound: false},
				{FaceFound: false},
				{FaceFound: false},
			},
			expected: true,
		},
		{
			name: "less than 70% faces",
			frames: []Frame{
				{FaceFound: true},
				{FaceFound: true},
				{FaceFound: false},
				{FaceFound: false},
				{FaceFound: false},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.CheckFacePresence(tt.frames)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDetector_DetectBlink(t *testing.T) {
	detector := NewDetector(DefaultConfig())

	tests := []struct {
		name     string
		frames   []Frame
		expected bool
	}{
		{
			name:     "insufficient frames",
			frames:   createFramesWithEAR(3, 0.3),
			expected: false,
		},
		{
			name:     "blink detected via EAR",
			frames:   createFramesWithBlink(10),
			expected: true,
		},
		{
			name:     "no blink - constant EAR",
			frames:   createFramesWithEAR(10, 0.3),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectBlink(tt.frames)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDetector_QuickCheck(t *testing.T) {
	detector := NewDetector(DefaultConfig())

	tests := []struct {
		name         string
		frames       []Frame
		expectLive   bool
		minScore     float64
	}{
		{
			name:       "insufficient frames",
			frames:     []Frame{createTestFrame(true)},
			expectLive: false,
			minScore:   0.0,
		},
		{
			name:       "good frames",
			frames:     createGoodFrames(5),
			expectLive: true,
			minScore:   0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isLive, score := detector.QuickCheck(tt.frames)
			if isLive != tt.expectLive {
				t.Errorf("isLive: expected %v, got %v", tt.expectLive, isLive)
			}
			if score < tt.minScore {
				t.Errorf("score too low: expected >= %f, got %f", tt.minScore, score)
			}
		})
	}
}

func TestCalculateEyeAspectRatio(t *testing.T) {
	tests := []struct {
		name      string
		landmarks []Point
		expected  float64
	}{
		{
			name:      "no landmarks",
			landmarks: []Point{},
			expected:  0.5,
		},
		{
			name: "2-point landmarks",
			landmarks: []Point{
				{X: 0, Y: 0},
				{X: 10, Y: 0},
			},
			expected: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateEyeAspectRatio(tt.landmarks)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestEmbeddingDistance(t *testing.T) {
	tests := []struct {
		name     string
		e1       []float32
		e2       []float32
		expected float64
	}{
		{
			name:     "identical embeddings",
			e1:       []float32{1, 2, 3},
			e2:       []float32{1, 2, 3},
			expected: 0,
		},
		{
			name:     "different lengths",
			e1:       []float32{1, 2, 3},
			e2:       []float32{1, 2},
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := embeddingDistance(tt.e1, tt.e2)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestAverageEmbedding(t *testing.T) {
	tests := []struct {
		name       string
		embeddings [][]float32
		wantNil    bool
	}{
		{
			name:       "empty",
			embeddings: [][]float32{},
			wantNil:    true,
		},
		{
			name:       "single embedding",
			embeddings: [][]float32{{1, 2, 3}},
			wantNil:    false,
		},
		{
			name:       "multiple embeddings",
			embeddings: [][]float32{{1, 2, 3}, {5, 6, 7}},
			wantNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := averageEmbedding(tt.embeddings)
			if tt.wantNil && result != nil {
				t.Error("expected nil result")
			}
			if !tt.wantNil && result == nil {
				t.Error("unexpected nil result")
			}
		})
	}
}

func TestDistance(t *testing.T) {
	p1 := Point{X: 0, Y: 0}
	p2 := Point{X: 3, Y: 4}

	result := distance(p1, p2)
	expected := 5.0

	if result != expected {
		t.Errorf("expected %f, got %f", expected, result)
	}
}

// Helper functions to create test data

func createTestFrame(faceFound bool) Frame {
	var emb recognition.Embedding
	if faceFound {
		for i := range emb.Vector {
			emb.Vector[i] = float32(i) / 128.0
		}
	}
	return Frame{
		Data:      []byte{1, 2, 3},
		Embedding: emb,
		FaceFound: faceFound,
		Timestamp: time.Now(),
	}
}

func createConsistentEmbeddings(count int, variance float64) [][]float32 {
	embeddings := make([][]float32, count)
	base := make([]float32, 128)
	for i := range base {
		base[i] = float32(i) / 128.0
	}

	for i := 0; i < count; i++ {
		emb := make([]float32, 128)
		for j := range emb {
			emb[j] = base[j] + float32(variance)*float32(i)/float32(count)
		}
		embeddings[i] = emb
	}
	return embeddings
}

func createIdenticalEmbeddings(count int) [][]float32 {
	embeddings := make([][]float32, count)
	base := make([]float32, 128)
	for i := range base {
		base[i] = float32(i) / 128.0
	}

	for i := 0; i < count; i++ {
		emb := make([]float32, 128)
		copy(emb, base)
		embeddings[i] = emb
	}
	return embeddings
}

func createDifferentEmbeddings(count int) [][]float32 {
	embeddings := make([][]float32, count)
	for i := 0; i < count; i++ {
		emb := make([]float32, 128)
		for j := range emb {
			emb[j] = float32(i*128+j) / 100.0 // Very different values
		}
		embeddings[i] = emb
	}
	return embeddings
}

func createFramesWithMovement(count int, movement float64) []Frame {
	frames := make([]Frame, count)
	for i := 0; i < count; i++ {
		var emb recognition.Embedding
		for j := range emb.Vector {
			emb.Vector[j] = float32(j)/128.0 + float32(movement)*float32(i)
		}
		frames[i] = Frame{
			Embedding: emb,
			FaceFound: true,
			Timestamp: time.Now(),
		}
	}
	return frames
}

func createStaticFrames(count int) []Frame {
	frames := make([]Frame, count)
	var baseEmb recognition.Embedding
	for j := range baseEmb.Vector {
		baseEmb.Vector[j] = float32(j) / 128.0
	}

	for i := 0; i < count; i++ {
		frames[i] = Frame{
			Embedding: baseEmb,
			FaceFound: true,
			Timestamp: time.Now(),
		}
	}
	return frames
}

func createFramesWithEAR(count int, ear float64) []Frame {
	frames := make([]Frame, count)
	for i := 0; i < count; i++ {
		var emb recognition.Embedding
		for j := range emb.Vector {
			emb.Vector[j] = float32(j) / 128.0
		}
		frames[i] = Frame{
			Embedding:      emb,
			FaceFound:      true,
			EyeAspectRatio: ear,
			Timestamp:      time.Now(),
		}
	}
	return frames
}

func createFramesWithBlink(count int) []Frame {
	frames := make([]Frame, count)
	for i := 0; i < count; i++ {
		var emb recognition.Embedding
		for j := range emb.Vector {
			// Add slight variation
			emb.Vector[j] = float32(j)/128.0 + float32(i)*0.001
		}

		ear := 0.3 // Normal open eye
		if i == count/2 {
			ear = 0.15 // Blink (closed eye)
		}

		frames[i] = Frame{
			Embedding:      emb,
			FaceFound:      true,
			EyeAspectRatio: ear,
			Timestamp:      time.Now(),
		}
	}
	return frames
}

func createGoodFrames(count int) []Frame {
	frames := make([]Frame, count)
	for i := 0; i < count; i++ {
		var emb recognition.Embedding
		for j := range emb.Vector {
			emb.Vector[j] = float32(j)/128.0 + float32(i)*0.02 // Some movement
		}
		frames[i] = Frame{
			Embedding: emb,
			FaceFound: true,
			Timestamp: time.Now(),
		}
	}
	return frames
}

// Benchmark tests
func BenchmarkDetector_Detect(b *testing.B) {
	detector := NewDetector(DefaultConfig())
	frames := createGoodFrames(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.Detect(frames)
	}
}

func BenchmarkDetector_CheckConsistency(b *testing.B) {
	detector := NewDetector(DefaultConfig())
	embeddings := createConsistentEmbeddings(10, 0.01)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.CheckConsistency(embeddings)
	}
}

func BenchmarkEmbeddingDistance(b *testing.B) {
	e1 := make([]float32, 128)
	e2 := make([]float32, 128)
	for i := range e1 {
		e1[i] = float32(i) / 128.0
		e2[i] = float32(i+1) / 128.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		embeddingDistance(e1, e2)
	}
}
