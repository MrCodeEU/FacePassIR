// Package recognition provides face detection and recognition functionality.
// It uses dlib/go-face for face detection, landmark extraction, and embedding generation.
package recognition

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/Kagami/go-face"
	"github.com/MrCodeEU/facepass/pkg/logging"
)

// Face represents a detected face in an image.
type Face struct {
	BoundingBox Rectangle
	Landmarks   []Point
	Confidence  float64
	Descriptor  Descriptor
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

// Descriptor is a 128-dimensional face descriptor from dlib.
type Descriptor = face.Descriptor

// Embedding represents a face embedding with metadata.
type Embedding struct {
	Vector  Descriptor `json:"vector"`
	Quality float64    `json:"quality"`
	Angle   string     `json:"angle"` // "front", "left", "right", "up", "down"
}

// ErrNoFaceDetected is returned when no face is found in the image.
var ErrNoFaceDetected = errors.New("no face detected")

// ErrMultipleFaces is returned when multiple faces are detected.
var ErrMultipleFaces = errors.New("multiple faces detected")

// ErrModelNotLoaded is returned when models are not loaded.
var ErrModelNotLoaded = errors.New("recognition models not loaded")

// ErrLowQuality is returned when face quality is below threshold.
var ErrLowQuality = errors.New("face quality too low")

// DlibRecognizer implements face recognition using dlib via go-face.
type DlibRecognizer struct {
	rec       *face.Recognizer
	modelPath string
	loaded    bool
	mu        sync.RWMutex
	tolerance float64
}

// NewRecognizer creates a new DlibRecognizer instance.
func NewRecognizer() *DlibRecognizer {
	return &DlibRecognizer{
		tolerance: 0.4, // Default tolerance for face matching
	}
}

// SetTolerance sets the tolerance for face matching.
// Lower values are more strict (fewer false positives).
func (r *DlibRecognizer) SetTolerance(tolerance float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tolerance = tolerance
}

// LoadModels loads the dlib face recognition models from the specified path.
// The path should contain:
// - shape_predictor_5_face_landmarks.dat
// - dlib_face_recognition_resnet_model_v1.dat
// - mmod_human_face_detector.dat (optional, for CNN detection)
func (r *DlibRecognizer) LoadModels(modelPath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.loaded {
		return nil
	}

	logging.Infof("Loading face recognition models from: %s", modelPath)

	rec, err := face.NewRecognizer(modelPath)
	if err != nil {
		return fmt.Errorf("failed to load models: %w", err)
	}

	r.rec = rec
	r.modelPath = modelPath
	r.loaded = true

	logging.Info("Face recognition models loaded successfully")
	return nil
}

// IsLoaded returns true if models are loaded.
func (r *DlibRecognizer) IsLoaded() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.loaded
}

// Close releases the recognizer resources.
func (r *DlibRecognizer) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.rec != nil {
		r.rec.Close()
		r.rec = nil
	}
	r.loaded = false
	return nil
}

// DetectFaces detects all faces in an image.
// Returns a slice of Face structs with bounding boxes and descriptors.
func (r *DlibRecognizer) DetectFaces(imageData []byte) ([]Face, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.loaded {
		return nil, ErrModelNotLoaded
	}

	// Recognize faces in the image
	faces, err := r.rec.Recognize(imageData)
	if err != nil {
		return nil, fmt.Errorf("face detection failed: %w", err)
	}

	if len(faces) == 0 {
		return nil, ErrNoFaceDetected
	}

	result := make([]Face, len(faces))
	for i, f := range faces {
		rect := f.Rectangle
		result[i] = Face{
			BoundingBox: Rectangle{
				X:      rect.Min.X,
				Y:      rect.Min.Y,
				Width:  rect.Dx(),
				Height: rect.Dy(),
			},
			Descriptor: f.Descriptor,
			Confidence: 1.0, // go-face doesn't provide confidence, assume high
		}
	}

	logging.Debugf("Detected %d face(s) in image", len(result))
	return result, nil
}

// DetectSingleFace detects exactly one face in the image.
// Returns an error if no face or multiple faces are detected.
func (r *DlibRecognizer) DetectSingleFace(imageData []byte) (*Face, error) {
	faces, err := r.DetectFaces(imageData)
	if err != nil {
		return nil, err
	}

	if len(faces) == 0 {
		return nil, ErrNoFaceDetected
	}

	if len(faces) > 1 {
		return nil, ErrMultipleFaces
	}

	return &faces[0], nil
}

// GetEmbedding extracts the face embedding from a detected face.
func (r *DlibRecognizer) GetEmbedding(f *Face, angle string) Embedding {
	return Embedding{
		Vector:  f.Descriptor,
		Quality: f.Confidence,
		Angle:   angle,
	}
}

// RecognizeFace detects a face and returns its embedding.
// This is a convenience method that combines detection and embedding extraction.
func (r *DlibRecognizer) RecognizeFace(imageData []byte, angle string) (*Embedding, error) {
	face, err := r.DetectSingleFace(imageData)
	if err != nil {
		return nil, err
	}

	embedding := r.GetEmbedding(face, angle)
	return &embedding, nil
}

// CompareFaces computes the Euclidean distance between two face embeddings.
// Returns the distance (lower = more similar).
// Typical threshold: 0.4-0.6 (faces with distance < threshold are considered the same person)
func (r *DlibRecognizer) CompareFaces(emb1, emb2 Embedding) float64 {
	return EuclideanDistance(emb1.Vector, emb2.Vector)
}

// Match checks if two faces match within the configured tolerance.
func (r *DlibRecognizer) Match(emb1, emb2 Embedding) bool {
	r.mu.RLock()
	tolerance := r.tolerance
	r.mu.RUnlock()

	distance := r.CompareFaces(emb1, emb2)
	return distance < tolerance
}

// FindBestMatch finds the best matching embedding from a list.
// Returns the index of the best match, the distance, and whether it's within tolerance.
func (r *DlibRecognizer) FindBestMatch(probe Embedding, gallery []Embedding) (int, float64, bool) {
	r.mu.RLock()
	tolerance := r.tolerance
	r.mu.RUnlock()

	if len(gallery) == 0 {
		return -1, math.MaxFloat64, false
	}

	bestIdx := 0
	bestDist := math.MaxFloat64

	for i, emb := range gallery {
		dist := r.CompareFaces(probe, emb)
		if dist < bestDist {
			bestDist = dist
			bestIdx = i
		}
	}

	return bestIdx, bestDist, bestDist < tolerance
}

// EuclideanDistance calculates the Euclidean distance between two descriptors.
func EuclideanDistance(d1, d2 Descriptor) float64 {
	if len(d1) != len(d2) {
		return math.MaxFloat64
	}

	var sum float64
	for i := range d1 {
		diff := float64(d1[i] - d2[i])
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

// AverageEmbedding computes the average of multiple embeddings.
// This is useful for combining multiple angles of the same face.
func AverageEmbedding(embeddings []Embedding) Embedding {
	if len(embeddings) == 0 {
		return Embedding{}
	}

	if len(embeddings) == 1 {
		return embeddings[0]
	}

	// Initialize with zeros
	var avgVector Descriptor
	for i := range avgVector {
		avgVector[i] = 0
	}

	// Sum all vectors
	for _, emb := range embeddings {
		for i, v := range emb.Vector {
			avgVector[i] += v
		}
	}

	// Divide by count
	count := float32(len(embeddings))
	for i := range avgVector {
		avgVector[i] /= count
	}

	// Average quality
	var avgQuality float64
	for _, emb := range embeddings {
		avgQuality += emb.Quality
	}
	avgQuality /= float64(len(embeddings))

	return Embedding{
		Vector:  avgVector,
		Quality: avgQuality,
		Angle:   "averaged",
	}
}
