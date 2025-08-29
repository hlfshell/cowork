package auth

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

// SecureStore provides secure storage for authentication credentials
type SecureStore struct {
	storePath string
	key       []byte
}

// NewSecureStore creates a new secure store instance
func NewSecureStore() (*SecureStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	storePath := filepath.Join(homeDir, ".config", "cowork", "auth")
	if err := os.MkdirAll(storePath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create auth directory: %w", err)
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

// Set stores an authentication configuration securely
func (s *SecureStore) Set(key string, config *AuthConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	encrypted, err := s.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	filePath := filepath.Join(s.storePath, s.sanitizeKey(key)+".enc")
	return os.WriteFile(filePath, encrypted, 0600)
}

// Get retrieves an authentication configuration
func (s *SecureStore) Get(key string) (*AuthConfig, error) {
	filePath := filepath.Join(s.storePath, s.sanitizeKey(key)+".enc")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("authentication not found for key: %s", key)
		}
		return nil, fmt.Errorf("failed to read auth file: %w", err)
	}

	decrypted, err := s.decrypt(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	var config AuthConfig
	if err := json.Unmarshal(decrypted, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// Delete removes an authentication configuration
func (s *SecureStore) Delete(key string) error {
	filePath := filepath.Join(s.storePath, s.sanitizeKey(key)+".enc")
	return os.Remove(filePath)
}

// List lists all authentication configurations with the given prefix
func (s *SecureStore) List(prefix string) ([]*AuthConfig, error) {
	entries, err := os.ReadDir(s.storePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth directory: %w", err)
	}

	var configs []*AuthConfig
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".enc") {
			continue
		}

		// Extract the original key from the sanitized filename
		sanitizedKey := strings.TrimSuffix(entry.Name(), ".enc")

		if !strings.HasPrefix(sanitizedKey, s.sanitizeKey(prefix)) {
			continue
		}

		// Read the file directly and decrypt it
		filePath := filepath.Join(s.storePath, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			// Skip invalid entries
			continue
		}

		decrypted, err := s.decrypt(data)
		if err != nil {
			// Skip invalid entries
			continue
		}

		var config AuthConfig
		if err := json.Unmarshal(decrypted, &config); err != nil {
			// Skip invalid entries
			continue
		}

		// Only include configs that have valid provider types
		if config.ProviderType == "" {
			// Skip invalid entries
			continue
		}

		configs = append(configs, &config)
	}

	return configs, nil
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
