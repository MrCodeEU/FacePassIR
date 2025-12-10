package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MrCodeEU/facepass/pkg/recognition"
)

func TestNewFileStorage(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		dataDir    string
		encryption bool
		wantErr    bool
	}{
		{
			name:       "without encryption",
			dataDir:    filepath.Join(tmpDir, "test1"),
			encryption: false,
			wantErr:    false,
		},
		{
			name:       "with encryption",
			dataDir:    filepath.Join(tmpDir, "test2"),
			encryption: true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := NewFileStorage(tt.dataDir, tt.encryption)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFileStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if fs == nil {
				t.Error("NewFileStorage returned nil")
			}

			// Check directories were created
			usersDir := filepath.Join(tt.dataDir, "users")
			if _, err := os.Stat(usersDir); os.IsNotExist(err) {
				t.Error("users directory was not created")
			}
		})
	}
}

func TestFileStorage_SaveAndLoadUser(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := NewFileStorage(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// Create test user data
	userData := UserFaceData{
		Username:   "testuser",
		Embeddings: createTestEmbeddings(3),
		EnrolledAt: time.Now(),
		LastUsed:   time.Now(),
		Metadata:   map[string]string{"device": "webcam"},
	}

	// Save user
	err = fs.SaveUser(userData)
	if err != nil {
		t.Fatalf("SaveUser failed: %v", err)
	}

	// Load user
	loaded, err := fs.LoadUser("testuser")
	if err != nil {
		t.Fatalf("LoadUser failed: %v", err)
	}

	// Verify loaded data
	if loaded.Username != userData.Username {
		t.Errorf("username mismatch: got %s, want %s", loaded.Username, userData.Username)
	}
	if len(loaded.Embeddings) != len(userData.Embeddings) {
		t.Errorf("embeddings count mismatch: got %d, want %d", len(loaded.Embeddings), len(userData.Embeddings))
	}
	if loaded.Metadata["device"] != "webcam" {
		t.Error("metadata not preserved")
	}
}

func TestFileStorage_SaveAndLoadUser_Encrypted(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := NewFileStorage(tmpDir, true)
	if err != nil {
		t.Fatalf("failed to create encrypted storage: %v", err)
	}

	userData := UserFaceData{
		Username:   "encrypteduser",
		Embeddings: createTestEmbeddings(2),
		EnrolledAt: time.Now(),
		LastUsed:   time.Now(),
		Metadata:   map[string]string{"test": "value"},
	}

	// Save with encryption
	err = fs.SaveUser(userData)
	if err != nil {
		t.Fatalf("SaveUser (encrypted) failed: %v", err)
	}

	// Load with decryption
	loaded, err := fs.LoadUser("encrypteduser")
	if err != nil {
		t.Fatalf("LoadUser (encrypted) failed: %v", err)
	}

	if loaded.Username != userData.Username {
		t.Errorf("username mismatch after encryption: got %s, want %s", loaded.Username, userData.Username)
	}

	// Verify the file is encrypted (not valid JSON)
	filePath := filepath.Join(tmpDir, "users", "encrypteduser.enc")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read encrypted file: %v", err)
	}

	// First byte should not be '{' if encrypted
	if len(data) > 0 && data[0] == '{' {
		t.Error("file does not appear to be encrypted")
	}
}

func TestFileStorage_LoadUser_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := NewFileStorage(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	_, err = fs.LoadUser("nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestFileStorage_DeleteUser(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := NewFileStorage(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// Create user
	userData := UserFaceData{
		Username:   "todelete",
		Embeddings: createTestEmbeddings(1),
		EnrolledAt: time.Now(),
	}
	if err := fs.SaveUser(userData); err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	// Verify user exists
	if !fs.UserExists("todelete") {
		t.Error("user should exist after save")
	}

	// Delete user
	err = fs.DeleteUser("todelete")
	if err != nil {
		t.Errorf("DeleteUser failed: %v", err)
	}

	// Verify user is gone
	if fs.UserExists("todelete") {
		t.Error("user should not exist after delete")
	}
}

func TestFileStorage_DeleteUser_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := NewFileStorage(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	err = fs.DeleteUser("nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestFileStorage_ListUsers(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := NewFileStorage(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// Initially empty
	users, err := fs.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}

	// Add some users
	for _, name := range []string{"alice", "bob", "charlie"} {
		userData := UserFaceData{
			Username:   name,
			Embeddings: createTestEmbeddings(1),
			EnrolledAt: time.Now(),
		}
		if err := fs.SaveUser(userData); err != nil {
			t.Fatalf("failed to save user %s: %v", name, err)
		}
	}

	// List users
	users, err = fs.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}

	// Check all users are listed
	userMap := make(map[string]bool)
	for _, u := range users {
		userMap[u] = true
	}
	for _, name := range []string{"alice", "bob", "charlie"} {
		if !userMap[name] {
			t.Errorf("user %s not in list", name)
		}
	}
}

func TestFileStorage_UserExists(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := NewFileStorage(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// User doesn't exist yet
	if fs.UserExists("testuser") {
		t.Error("user should not exist initially")
	}

	// Create user
	userData := UserFaceData{
		Username:   "testuser",
		Embeddings: createTestEmbeddings(1),
		EnrolledAt: time.Now(),
	}
	if err := fs.SaveUser(userData); err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	// User now exists
	if !fs.UserExists("testuser") {
		t.Error("user should exist after save")
	}
}

func TestFileStorage_AddEmbedding(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := NewFileStorage(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// Create user with 1 embedding
	userData := UserFaceData{
		Username:   "testuser",
		Embeddings: createTestEmbeddings(1),
		EnrolledAt: time.Now(),
	}
	if err := fs.SaveUser(userData); err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	// Add another embedding
	newEmb := createTestEmbeddings(1)[0]
	if err := fs.AddEmbedding("testuser", newEmb); err != nil {
		t.Fatalf("AddEmbedding failed: %v", err)
	}

	// Load and verify
	loaded, err := fs.LoadUser("testuser")
	if err != nil {
		t.Fatalf("LoadUser failed: %v", err)
	}
	if len(loaded.Embeddings) != 2 {
		t.Errorf("expected 2 embeddings, got %d", len(loaded.Embeddings))
	}
}

func TestFileStorage_UpdateLastUsed(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := NewFileStorage(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// Create user
	oldTime := time.Now().Add(-1 * time.Hour)
	userData := UserFaceData{
		Username:   "testuser",
		Embeddings: createTestEmbeddings(1),
		EnrolledAt: oldTime,
		LastUsed:   oldTime,
	}
	if err := fs.SaveUser(userData); err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	// Update last used
	time.Sleep(10 * time.Millisecond)
	if err := fs.UpdateLastUsed("testuser"); err != nil {
		t.Fatalf("UpdateLastUsed failed: %v", err)
	}

	// Verify
	loaded, err := fs.LoadUser("testuser")
	if err != nil {
		t.Fatalf("LoadUser failed: %v", err)
	}
	if !loaded.LastUsed.After(oldTime) {
		t.Error("LastUsed was not updated")
	}
}

func TestFileStorage_CreateUser(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := NewFileStorage(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	embeddings := createTestEmbeddings(3)
	metadata := map[string]string{"camera": "ir"}

	err = fs.CreateUser("newuser", embeddings, metadata)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Verify
	loaded, err := fs.LoadUser("newuser")
	if err != nil {
		t.Fatalf("LoadUser failed: %v", err)
	}
	if loaded.Username != "newuser" {
		t.Errorf("expected username 'newuser', got %s", loaded.Username)
	}
	if len(loaded.Embeddings) != 3 {
		t.Errorf("expected 3 embeddings, got %d", len(loaded.Embeddings))
	}
	if loaded.Metadata["camera"] != "ir" {
		t.Error("metadata not preserved")
	}
}

func TestFileStorage_CreateUser_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := NewFileStorage(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// Create user first time
	err = fs.CreateUser("existinguser", createTestEmbeddings(1), nil)
	if err != nil {
		t.Fatalf("first CreateUser failed: %v", err)
	}

	// Try to create again
	err = fs.CreateUser("existinguser", createTestEmbeddings(1), nil)
	if err != ErrUserExists {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

func TestFileStorage_GetAllEmbeddings(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := NewFileStorage(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	embeddings := createTestEmbeddings(5)
	if err := fs.CreateUser("testuser", embeddings, nil); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Get embeddings
	result, err := fs.GetAllEmbeddings("testuser")
	if err != nil {
		t.Fatalf("GetAllEmbeddings failed: %v", err)
	}
	if len(result) != 5 {
		t.Errorf("expected 5 embeddings, got %d", len(result))
	}
}

func TestEncryptDecrypt(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := NewFileStorage(tmpDir, true)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	plaintext := []byte("This is a test message for encryption")

	// Encrypt
	ciphertext, err := fs.encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// Ciphertext should be different from plaintext
	if string(ciphertext) == string(plaintext) {
		t.Error("ciphertext should differ from plaintext")
	}

	// Decrypt
	decrypted, err := fs.decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	// Should match original
	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted text doesn't match: got %s, want %s", string(decrypted), string(plaintext))
	}
}

func TestDecrypt_InvalidData(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := NewFileStorage(tmpDir, true)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// Too short
	_, err = fs.decrypt([]byte("short"))
	if err != ErrEncryption {
		t.Errorf("expected ErrEncryption for short data, got %v", err)
	}

	// Invalid ciphertext
	invalidData := make([]byte, 100)
	_, err = fs.decrypt(invalidData)
	if err != ErrEncryption {
		t.Errorf("expected ErrEncryption for invalid data, got %v", err)
	}
}

// Helper function to create test embeddings
func createTestEmbeddings(count int) []recognition.Embedding {
	embeddings := make([]recognition.Embedding, count)
	for i := 0; i < count; i++ {
		var vector recognition.Descriptor
		for j := range vector {
			vector[j] = float32(i*128+j) / 1000.0
		}
		embeddings[i] = recognition.Embedding{
			Vector:  vector,
			Quality: 0.9,
			Angle:   "front",
		}
	}
	return embeddings
}

// Benchmark tests
func BenchmarkFileStorage_SaveUser(b *testing.B) {
	tmpDir := b.TempDir()
	fs, _ := NewFileStorage(tmpDir, false)

	userData := UserFaceData{
		Username:   "benchuser",
		Embeddings: createTestEmbeddings(5),
		EnrolledAt: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fs.SaveUser(userData)
	}
}

func BenchmarkFileStorage_LoadUser(b *testing.B) {
	tmpDir := b.TempDir()
	fs, _ := NewFileStorage(tmpDir, false)

	userData := UserFaceData{
		Username:   "benchuser",
		Embeddings: createTestEmbeddings(5),
		EnrolledAt: time.Now(),
	}
	_ = fs.SaveUser(userData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = fs.LoadUser("benchuser")
	}
}

func BenchmarkEncryptDecrypt(b *testing.B) {
	tmpDir := b.TempDir()
	fs, _ := NewFileStorage(tmpDir, true)

	data := []byte("benchmark encryption data that is reasonably sized")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encrypted, _ := fs.encrypt(data)
		_, _ = fs.decrypt(encrypted)
	}
}
