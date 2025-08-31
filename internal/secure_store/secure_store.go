package secure_store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SecureStore provides secure storage for generic key-value pairs
type SecureStore struct {
	storePath string
	key       []byte
}

// NewSecureStore creates a new secure store instance
func NewSecureStore(store string, baseDir string) (*SecureStore, error) {
	storePath := filepath.Join(baseDir, ".cowork", store)
	if err := os.MkdirAll(storePath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	// Generate or load encryption key
	keyPath := filepath.Join(storePath, ".key")
	key, err := loadOrGenerateKey(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load/generate key: %w", err)
	}

	return &SecureStore{
		storePath: storePath,
		key:       key,
	}, nil
}

// Set stores a value securely with automatic type conversion
func (s *SecureStore) Set(key string, value interface{}) error {
	// Convert value to JSON for storage
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	encrypted, err := s.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	filePath := filepath.Join(s.storePath, s.sanitizeKey(key)+".enc")
	return os.WriteFile(filePath, encrypted, 0600)
}

// Get retrieves a value and attempts to unmarshal it into the provided interface
func (s *SecureStore) Get(key string, value interface{}) error {
	filePath := filepath.Join(s.storePath, s.sanitizeKey(key)+".enc")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("value not found for key: %s", key)
		}
		return fmt.Errorf("failed to read file: %w", err)
	}

	decrypted, err := s.decrypt(data)
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}

	if err := json.Unmarshal(decrypted, value); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// SetString stores a string value
func (s *SecureStore) SetString(key string, value string) error {
	return s.Set(key, value)
}

// GetString retrieves a string value
func (s *SecureStore) GetString(key string) (string, error) {
	var value string
	err := s.Get(key, &value)
	return value, err
}

// SetInt stores an integer value
func (s *SecureStore) SetInt(key string, value int) error {
	return s.Set(key, value)
}

// GetInt retrieves an integer value
func (s *SecureStore) GetInt(key string) (int, error) {
	var value int
	err := s.Get(key, &value)
	return value, err
}

// SetBool stores a boolean value
func (s *SecureStore) SetBool(key string, value bool) error {
	return s.Set(key, value)
}

// GetBool retrieves a boolean value
func (s *SecureStore) GetBool(key string) (bool, error) {
	var value bool
	err := s.Get(key, &value)
	return value, err
}

// SetFloat stores a float64 value
func (s *SecureStore) SetFloat(key string, value float64) error {
	return s.Set(key, value)
}

// GetFloat retrieves a float64 value
func (s *SecureStore) GetFloat(key string) (float64, error) {
	var value float64
	err := s.Get(key, &value)
	return value, err
}

// SetBytes stores raw bytes
func (s *SecureStore) SetBytes(key string, value []byte) error {
	return s.Set(key, value)
}

// GetBytes retrieves raw bytes
func (s *SecureStore) GetBytes(key string) ([]byte, error) {
	var value []byte
	err := s.Get(key, &value)
	return value, err
}

// Delete removes a stored value
func (s *SecureStore) Delete(key string) error {
	filePath := filepath.Join(s.storePath, s.sanitizeKey(key)+".enc")
	return os.Remove(filePath)
}

// List lists all keys with the given prefix
func (s *SecureStore) List(prefix string) ([]string, error) {
	entries, err := os.ReadDir(s.storePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read store directory: %w", err)
	}

	var keys []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".enc") {
			continue
		}

		// Extract the original key from the sanitized filename
		sanitizedKey := strings.TrimSuffix(entry.Name(), ".enc")

		if !strings.HasPrefix(sanitizedKey, s.sanitizeKey(prefix)) {
			continue
		}

		// For now, we can't easily reconstruct the original key from sanitized
		// So we'll return the sanitized key. In a more sophisticated implementation,
		// we could store a mapping of sanitized to original keys
		keys = append(keys, sanitizedKey)
	}

	return keys, nil
}

// ListWithValues lists all keys with the given prefix and returns their values
// This is useful when you need both keys and values, but requires knowing the expected type
func (s *SecureStore) ListWithValues(prefix string, valueType interface{}) (map[string]interface{}, error) {
	keys, err := s.List(prefix)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, key := range keys {
		// Create a new instance of the value type for each key
		value := valueType
		err := s.Get(key, &value)
		if err != nil {
			// Skip invalid entries
			continue
		}
		result[key] = value
	}

	return result, nil
}

// Exists checks if a key exists in the store
func (s *SecureStore) Exists(key string) bool {
	filePath := filepath.Join(s.storePath, s.sanitizeKey(key)+".enc")
	_, err := os.Stat(filePath)
	return err == nil
}

// Clear removes all stored values
func (s *SecureStore) Clear() error {
	entries, err := os.ReadDir(s.storePath)
	if err != nil {
		return fmt.Errorf("failed to read store directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".enc") {
			continue
		}

		filePath := filepath.Join(s.storePath, entry.Name())
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("failed to remove file %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// GetSize returns the number of stored items
func (s *SecureStore) GetSize() (int, error) {
	entries, err := os.ReadDir(s.storePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read store directory: %w", err)
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".enc") {
			count++
		}
	}

	return count, nil
}

// encrypt encrypts data using AES-256-GCM
func (s *SecureStore) encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// decrypt decrypts data using AES-256-GCM
func (s *SecureStore) decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// sanitizeKey sanitizes a key for use as a filename
func (s *SecureStore) sanitizeKey(key string) string {
	// Replace invalid characters with underscores
	sanitized := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, key)
	return sanitized
}

// loadOrGenerateKey loads an existing key or generates a new one
func loadOrGenerateKey(keyPath string) ([]byte, error) {
	// Try to load existing key
	if data, err := os.ReadFile(keyPath); err == nil {
		if len(data) == 32 {
			return data, nil
		}
	}

	// Generate new key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Save key
	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return nil, fmt.Errorf("failed to save key: %w", err)
	}

	return key, nil
}
