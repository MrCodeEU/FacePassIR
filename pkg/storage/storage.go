// Package storage provides secure storage for face embeddings.
// Embeddings are encrypted at rest using NaCl secretbox.
package storage

import (
	"errors"
	"time"
)

// UserFaceData contains all face data for a user.
type UserFaceData struct {
	Username   string               `json:"username"`
	Embeddings []StoredEmbedding    `json:"embeddings"`
	EnrolledAt time.Time            `json:"enrolled_at"`
	LastUsed   time.Time            `json:"last_used"`
	Metadata   map[string]string    `json:"metadata"`
}

// StoredEmbedding represents a stored face embedding with metadata.
type StoredEmbedding struct {
	Vector  []float32 `json:"vector"`
	Angle   string    `json:"angle"`
	Quality float64   `json:"quality"`
}

// Storage defines the interface for face data persistence.
type Storage interface {
	SaveUser(user UserFaceData) error
	LoadUser(username string) (UserFaceData, error)
	DeleteUser(username string) error
	ListUsers() ([]string, error)
	UserExists(username string) bool
}

// ErrUserNotFound is returned when the user is not enrolled.
var ErrUserNotFound = errors.New("user not found")

// ErrUserExists is returned when trying to enroll an existing user.
var ErrUserExists = errors.New("user already enrolled")

// ErrStorageAccess is returned when storage cannot be accessed.
var ErrStorageAccess = errors.New("failed to access storage")

// ErrEncryption is returned when encryption/decryption fails.
var ErrEncryption = errors.New("encryption error")

// TODO: Implement storage functionality
// - File-based storage (JSON)
// - NaCl secretbox encryption
// - Per-user data files
// - Machine-specific key derivation
