// Package storage provides secure storage for face embeddings.
// Embeddings are encrypted at rest using NaCl secretbox.
package storage

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MrCodeEU/facepass/pkg/logging"
	"github.com/MrCodeEU/facepass/pkg/recognition"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	// NonceSize is the size of the nonce used for encryption
	NonceSize = 24
	// KeySize is the size of the encryption key
	KeySize = 32
)

// UserFaceData contains all face data for a user.
type UserFaceData struct {
	Username   string                  `json:"username"`
	Embeddings []recognition.Embedding `json:"embeddings"`
	EnrolledAt time.Time               `json:"enrolled_at"`
	LastUsed   time.Time               `json:"last_used"`
	Metadata   map[string]string       `json:"metadata"`
}

// ErrUserNotFound is returned when the user is not enrolled.
var ErrUserNotFound = errors.New("user not found")

// ErrUserExists is returned when trying to enroll an existing user.
var ErrUserExists = errors.New("user already enrolled")

// ErrStorageAccess is returned when storage cannot be accessed.
var ErrStorageAccess = errors.New("failed to access storage")

// ErrEncryption is returned when encryption/decryption fails.
var ErrEncryption = errors.New("encryption error")

// FileStorage implements Storage interface using file-based storage.
type FileStorage struct {
	dataDir           string
	encryptionEnabled bool
	encryptionKey     [KeySize]byte
}

// NewFileStorage creates a new FileStorage instance.
func NewFileStorage(dataDir string, encryptionEnabled bool) (*FileStorage, error) {
	fs := &FileStorage{
		dataDir:           dataDir,
		encryptionEnabled: encryptionEnabled,
	}

	// Derive encryption key from machine-specific information
	if encryptionEnabled {
		key, err := deriveKey()
		if err != nil {
			return nil, fmt.Errorf("failed to derive encryption key: %w", err)
		}
		fs.encryptionKey = key
	}

	// Ensure directories exist
	usersDir := filepath.Join(dataDir, "users")
	if err := os.MkdirAll(usersDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create users directory: %w", err)
	}

	return fs, nil
}

// deriveKey derives an encryption key from machine-specific information.
// This ties the encrypted data to this specific machine.
func deriveKey() ([KeySize]byte, error) {
	var key [KeySize]byte

	// Combine multiple sources of machine identity
	var identity strings.Builder

	// Machine ID (Linux specific)
	if machineID, err := os.ReadFile("/etc/machine-id"); err == nil {
		identity.Write(machineID)
	}

	// Hostname
	if hostname, err := os.Hostname(); err == nil {
		identity.WriteString(hostname)
	}

	// User ID
	identity.WriteString(fmt.Sprintf("%d", os.Getuid()))

	// Add a constant salt for additional security
	identity.WriteString("facepass-v1-salt")

	// Hash to derive key
	hash := sha256.Sum256([]byte(identity.String()))
	copy(key[:], hash[:])

	return key, nil
}

// getUserPath returns the file path for a user's data.
func (fs *FileStorage) getUserPath(username string) string {
	filename := username + ".json"
	if fs.encryptionEnabled {
		filename = username + ".enc"
	}
	return filepath.Join(fs.dataDir, "users", filename)
}

// SaveUser saves user face data to storage.
func (fs *FileStorage) SaveUser(user UserFaceData) error {
	path := fs.getUserPath(user.Username)

	// Marshal to JSON
	data, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	// Encrypt if enabled
	if fs.encryptionEnabled {
		data, err = fs.encrypt(data)
		if err != nil {
			return fmt.Errorf("failed to encrypt user data: %w", err)
		}
	}

	// Write to file
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write user data: %w", err)
	}

	logging.Debugf("Saved user data for: %s", user.Username)
	return nil
}

// LoadUser loads user face data from storage.
func (fs *FileStorage) LoadUser(username string) (*UserFaceData, error) {
	path := fs.getUserPath(username)

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to read user data: %w", err)
	}

	// Decrypt if enabled
	if fs.encryptionEnabled {
		data, err = fs.decrypt(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt user data: %w", err)
		}
	}

	// Unmarshal JSON
	var user UserFaceData
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	logging.Debugf("Loaded user data for: %s", username)
	return &user, nil
}

// DeleteUser removes user face data from storage.
func (fs *FileStorage) DeleteUser(username string) error {
	path := fs.getUserPath(username)

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to delete user data: %w", err)
	}

	logging.Infof("Deleted user data for: %s", username)
	return nil
}

// ListUsers returns a list of all enrolled usernames.
func (fs *FileStorage) ListUsers() ([]string, error) {
	usersDir := filepath.Join(fs.dataDir, "users")

	entries, err := os.ReadDir(usersDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	var users []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Handle both encrypted and unencrypted files
		if strings.HasSuffix(name, ".json") {
			users = append(users, strings.TrimSuffix(name, ".json"))
		} else if strings.HasSuffix(name, ".enc") {
			users = append(users, strings.TrimSuffix(name, ".enc"))
		}
	}

	return users, nil
}

// UserExists checks if a user is enrolled.
func (fs *FileStorage) UserExists(username string) bool {
	path := fs.getUserPath(username)
	_, err := os.Stat(path)
	return err == nil
}

// AddEmbedding adds a new embedding to an existing user.
func (fs *FileStorage) AddEmbedding(username string, embedding recognition.Embedding) error {
	user, err := fs.LoadUser(username)
	if err != nil {
		return err
	}

	user.Embeddings = append(user.Embeddings, embedding)
	user.LastUsed = time.Now()

	return fs.SaveUser(*user)
}

// UpdateLastUsed updates the last used timestamp for a user.
func (fs *FileStorage) UpdateLastUsed(username string) error {
	user, err := fs.LoadUser(username)
	if err != nil {
		return err
	}

	user.LastUsed = time.Now()
	return fs.SaveUser(*user)
}

// encrypt encrypts data using NaCl secretbox.
func (fs *FileStorage) encrypt(plaintext []byte) ([]byte, error) {
	// Generate random nonce
	var nonce [NonceSize]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	// Encrypt
	encrypted := secretbox.Seal(nonce[:], plaintext, &nonce, &fs.encryptionKey)
	return encrypted, nil
}

// decrypt decrypts data using NaCl secretbox.
func (fs *FileStorage) decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < NonceSize {
		return nil, ErrEncryption
	}

	// Extract nonce
	var nonce [NonceSize]byte
	copy(nonce[:], ciphertext[:NonceSize])

	// Decrypt
	plaintext, ok := secretbox.Open(nil, ciphertext[NonceSize:], &nonce, &fs.encryptionKey)
	if !ok {
		return nil, ErrEncryption
	}

	return plaintext, nil
}

// CreateUser creates a new user with initial embeddings.
func (fs *FileStorage) CreateUser(username string, embeddings []recognition.Embedding, metadata map[string]string) error {
	if fs.UserExists(username) {
		return ErrUserExists
	}

	if metadata == nil {
		metadata = make(map[string]string)
	}

	user := UserFaceData{
		Username:   username,
		Embeddings: embeddings,
		EnrolledAt: time.Now(),
		LastUsed:   time.Now(),
		Metadata:   metadata,
	}

	return fs.SaveUser(user)
}

// GetAllEmbeddings returns all embeddings for a user.
func (fs *FileStorage) GetAllEmbeddings(username string) ([]recognition.Embedding, error) {
	user, err := fs.LoadUser(username)
	if err != nil {
		return nil, err
	}
	return user.Embeddings, nil
}
