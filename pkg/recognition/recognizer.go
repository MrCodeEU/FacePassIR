// Package recognition provides face detection and recognition functionality.
// It uses dlib/go-face for face detection, landmark extraction, and embedding generation.
package recognition

import (
	"errors"
)

// Face represents a detected face in an image.
type Face struct {
	BoundingBox Rectangle
	Landmarks   []Point
	Confidence  float64
}

// Rectangle represents a bounding box.
type Rectangle struct {
	X, Y          int
	Width, Height int
}

// Point represents a 2D point.
type Point struct {
	X, Y int
}

// Embedding represents a face embedding (128-dimensional vector).
type Embedding struct {
	Vector  []float32
	Quality float64
	Angle   string // "front", "left", "right", "up", "down"
}

// Recognizer defines the interface for face recognition operations.
type Recognizer interface {
	LoadModels(path string) error
	DetectFaces(imageData []byte) ([]Face, error)
	RecognizeFace(face Face, imageData []byte) (Embedding, error)
	CompareFaces(emb1, emb2 Embedding) (float64, error)
	Close() error
}

// ErrNoFaceDetected is returned when no face is found in the image.
var ErrNoFaceDetected = errors.New("no face detected")

// ErrMultipleFaces is returned when multiple faces are detected.
var ErrMultipleFaces = errors.New("multiple faces detected")

// ErrModelNotLoaded is returned when models are not loaded.
var ErrModelNotLoaded = errors.New("recognition models not loaded")

// ErrLowQuality is returned when face quality is below threshold.
var ErrLowQuality = errors.New("face quality too low")

// TODO: Implement face recognition functionality
// - Load dlib models
// - Face detection
// - Landmark extraction
// - Embedding generation
// - Face comparison
