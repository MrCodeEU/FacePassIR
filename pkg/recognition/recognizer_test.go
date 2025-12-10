package recognition

import (
	"errors"
	"image"
	"testing"

	"github.com/Kagami/go-face"
)

func TestNewRecognizer(t *testing.T) {
	rec := NewRecognizer()
	if rec == nil {
		t.Fatal("NewRecognizer returned nil")
	}
	if rec.tolerance != 0.4 {
		t.Errorf("expected default tolerance 0.4, got %f", rec.tolerance)
	}
}

func TestSetTolerance(t *testing.T) {
	rec := NewRecognizer()
	rec.SetTolerance(0.6)
	if rec.tolerance != 0.6 {
		t.Errorf("expected tolerance 0.6, got %f", rec.tolerance)
	}
}

func TestIsLoaded(t *testing.T) {
	rec := NewRecognizer()
	if rec.IsLoaded() {
		t.Error("expected IsLoaded to be false initially")
	}
}

func TestEuclideanDistance(t *testing.T) {
	tests := []struct {
		name     string
		d1       Descriptor
		d2       Descriptor
		expected float64
	}{
		{
			name:     "identical",
			d1:       Descriptor{1, 2, 3},
			d2:       Descriptor{1, 2, 3},
			expected: 0.0,
		},
		{
			name:     "different",
			d1:       Descriptor{1, 2, 3},
			d2:       Descriptor{4, 6, 8},
			expected: 7.0710678, // sqrt(3^2 + 4^2 + 5^2) = sqrt(9+16+25) = sqrt(50)
		},
	}

	// Fill the rest of descriptor with 0s
	for i := range tests {
		for j := 3; j < 128; j++ {
			tests[i].d1[j] = 0
			tests[i].d2[j] = 0
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := EuclideanDistance(tt.d1, tt.d2)
			// Check with epsilon
			if dist < tt.expected-0.0001 || dist > tt.expected+0.0001 {
				t.Errorf("expected %f, got %f", tt.expected, dist)
			}
		})
	}
}

func TestAverageEmbedding(t *testing.T) {
	d1 := Descriptor{1, 2, 3}
	d2 := Descriptor{3, 4, 5}
	// Fill rest with 0
	for i := 3; i < 128; i++ {
		d1[i] = 0
		d2[i] = 0
	}

	embeddings := []Embedding{
		{Vector: d1},
		{Vector: d2},
	}

	avg := AverageEmbedding(embeddings)

	if avg.Vector[0] != 2.0 || avg.Vector[1] != 3.0 || avg.Vector[2] != 4.0 {
		t.Errorf("expected [2, 3, 4], got [%f, %f, %f]", avg.Vector[0], avg.Vector[1], avg.Vector[2])
	}
}

func TestMatch(t *testing.T) {
	rec := NewRecognizer()
	rec.SetTolerance(0.5)

	d1 := Descriptor{1, 2, 3}
	d2 := Descriptor{1.1, 2.1, 3.1} // Close
	d3 := Descriptor{10, 20, 30}    // Far

	// Fill rest
	for i := 3; i < 128; i++ {
		d1[i] = 0
		d2[i] = 0
		d3[i] = 0
	}

	e1 := Embedding{Vector: d1}
	e2 := Embedding{Vector: d2}
	e3 := Embedding{Vector: d3}

	if !rec.Match(e1, e2) {
		t.Error("expected match for close descriptors")
	}

	if rec.Match(e1, e3) {
		t.Error("expected no match for far descriptors")
	}
}

func TestLoadModels(t *testing.T) {
	r := NewRecognizer()

	// Mock factory
	r.factory = func(path string) (FaceEngine, error) {
		return &MockFaceEngine{}, nil
	}

	err := r.LoadModels("/tmp/models")
	if err != nil {
		t.Errorf("LoadModels failed: %v", err)
	}
	if !r.IsLoaded() {
		t.Error("Expected loaded to be true")
	}

	// Load again (should be no-op)
	err = r.LoadModels("/tmp/models")
	if err != nil {
		t.Errorf("LoadModels failed on second call: %v", err)
	}
}

func TestLoadModels_Failure(t *testing.T) {
	r := NewRecognizer()

	// Mock factory failure
	r.factory = func(path string) (FaceEngine, error) {
		return nil, errors.New("load failed")
	}

	err := r.LoadModels("/tmp/models")
	if err == nil {
		t.Error("Expected LoadModels to fail")
	}
	if r.IsLoaded() {
		t.Error("Expected loaded to be false")
	}
}

func TestDetectFaces(t *testing.T) {
	r := NewRecognizer()
	mockEngine := &MockFaceEngine{
		RecognizeFunc: func(data []byte) ([]face.Face, error) {
			return []face.Face{
				{
					Rectangle:  image.Rect(0, 0, 100, 100),
					Descriptor: face.Descriptor{1, 2, 3},
				},
			}, nil
		},
	}
	r.factory = func(path string) (FaceEngine, error) {
		return mockEngine, nil
	}
	_ = r.LoadModels("dummy")

	faces, err := r.DetectFaces([]byte("image"))
	if err != nil {
		t.Fatalf("DetectFaces failed: %v", err)
	}
	if len(faces) != 1 {
		t.Errorf("Expected 1 face, got %d", len(faces))
	}
	if faces[0].BoundingBox.Width != 100 {
		t.Errorf("Expected width 100, got %d", faces[0].BoundingBox.Width)
	}
}

func TestDetectFaces_NotLoaded(t *testing.T) {
	r := NewRecognizer()
	_, err := r.DetectFaces([]byte("image"))
	if err != ErrModelNotLoaded {
		t.Errorf("Expected ErrModelNotLoaded, got %v", err)
	}
}

func TestDetectFaces_NoFace(t *testing.T) {
	r := NewRecognizer()
	mockEngine := &MockFaceEngine{
		RecognizeFunc: func(data []byte) ([]face.Face, error) {
			return []face.Face{}, nil
		},
	}
	r.factory = func(path string) (FaceEngine, error) {
		return mockEngine, nil
	}
	_ = r.LoadModels("dummy")

	_, err := r.DetectFaces([]byte("image"))
	if err != ErrNoFaceDetected {
		t.Errorf("Expected ErrNoFaceDetected, got %v", err)
	}
}

func TestDetectFaces_Error(t *testing.T) {
	r := NewRecognizer()
	mockEngine := &MockFaceEngine{
		RecognizeFunc: func(data []byte) ([]face.Face, error) {
			return nil, errors.New("engine error")
		},
	}
	r.factory = func(path string) (FaceEngine, error) {
		return mockEngine, nil
	}
	_ = r.LoadModels("dummy")

	_, err := r.DetectFaces([]byte("image"))
	if err == nil {
		t.Error("Expected error")
	}
}

func TestDetectSingleFace(t *testing.T) {
	r := NewRecognizer()
	mockEngine := &MockFaceEngine{
		RecognizeFunc: func(data []byte) ([]face.Face, error) {
			return []face.Face{
				{Rectangle: image.Rect(0, 0, 100, 100)},
			}, nil
		},
	}
	r.factory = func(path string) (FaceEngine, error) {
		return mockEngine, nil
	}
	_ = r.LoadModels("dummy")

	face, err := r.DetectSingleFace([]byte("image"))
	if err != nil {
		t.Fatalf("DetectSingleFace failed: %v", err)
	}
	if face == nil {
		t.Fatal("Expected face, got nil")
	}
}

func TestDetectSingleFace_Multiple(t *testing.T) {
	r := NewRecognizer()
	mockEngine := &MockFaceEngine{
		RecognizeFunc: func(data []byte) ([]face.Face, error) {
			return []face.Face{
				{Rectangle: image.Rect(0, 0, 100, 100)},
				{Rectangle: image.Rect(100, 100, 200, 200)},
			}, nil
		},
	}
	r.factory = func(path string) (FaceEngine, error) {
		return mockEngine, nil
	}
	_ = r.LoadModels("dummy")

	_, err := r.DetectSingleFace([]byte("image"))
	if err != ErrMultipleFaces {
		t.Errorf("Expected ErrMultipleFaces, got %v", err)
	}
}

func TestClose(t *testing.T) {
	r := NewRecognizer()
	closed := false
	mockEngine := &MockFaceEngine{
		CloseFunc: func() { closed = true },
	}
	r.factory = func(path string) (FaceEngine, error) {
		return mockEngine, nil
	}
	_ = r.LoadModels("dummy")

	err := r.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
	if !closed {
		t.Error("Expected engine to be closed")
	}
	if r.IsLoaded() {
		t.Error("Expected loaded to be false")
	}
}

func TestRecognizeFace(t *testing.T) {
	r := NewRecognizer()
	mockEngine := &MockFaceEngine{
		RecognizeFunc: func(data []byte) ([]face.Face, error) {
			return []face.Face{
				{
					Rectangle:  image.Rect(0, 0, 100, 100),
					Descriptor: face.Descriptor{1, 2, 3},
				},
			}, nil
		},
	}
	r.factory = func(path string) (FaceEngine, error) {
		return mockEngine, nil
	}
	_ = r.LoadModels("dummy")

	emb, err := r.RecognizeFace([]byte("image"), "front")
	if err != nil {
		t.Fatalf("RecognizeFace failed: %v", err)
	}
	if emb.Angle != "front" {
		t.Errorf("Expected angle front, got %s", emb.Angle)
	}
}

func TestFindBestMatch(t *testing.T) {
	r := NewRecognizer()

	probe := Embedding{Vector: face.Descriptor{1, 0, 0}}
	gallery := []Embedding{
		{Vector: face.Descriptor{0, 1, 0}},   // Dist sqrt(2) ~ 1.41
		{Vector: face.Descriptor{1, 0.1, 0}}, // Dist 0.1
	}

	idx, dist, match := r.FindBestMatch(probe, gallery)
	if idx != 1 {
		t.Errorf("Expected index 1, got %d", idx)
	}
	if !match {
		t.Error("Expected match")
	}
	if dist > 0.2 {
		t.Errorf("Expected small distance, got %f", dist)
	}

	// Empty gallery
	idx, _, _ = r.FindBestMatch(probe, nil)
	if idx != -1 {
		t.Errorf("Expected -1 for empty gallery, got %d", idx)
	}
}
