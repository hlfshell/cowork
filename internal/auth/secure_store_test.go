package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hlfshell/cowork/internal/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecureStore(t *testing.T) {
	// Test creating a new secure store
	store, err := NewSecureStore()
	
	assert.NoError(t, err)
	assert.NotNil(t, store)
	assert.NotEmpty(t, store.storePath)
	assert.Len(t, store.key, 32) // 256-bit key
}

func TestSetAndGet(t *testing.T) {
	// Setup
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Create test config
	config := &AuthConfig{
		ProviderType: git.ProviderGitHub,
		AuthMethod:   AuthMethodToken,
		Token:        "test-token-123",
		Username:     "testuser",
		Password:     "testpass",
	}

	// Test setting config
	err = store.Set("test-key", config)
	assert.NoError(t, err)

	// Test getting config
	retrievedConfig, err := store.Get("test-key")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedConfig)
	assert.Equal(t, config.ProviderType, retrievedConfig.ProviderType)
	assert.Equal(t, config.AuthMethod, retrievedConfig.AuthMethod)
	assert.Equal(t, config.Token, retrievedConfig.Token)
	assert.Equal(t, config.Username, retrievedConfig.Username)
	assert.Equal(t, config.Password, retrievedConfig.Password)
}

func TestGetNonExistent(t *testing.T) {
	// Setup
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Test getting non-existent config
	_, err = store.Get("non-existent-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication not found")
}

func TestDelete(t *testing.T) {
	// Setup
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Create test config
	config := &AuthConfig{
		ProviderType: git.ProviderGitLab,
		AuthMethod:   AuthMethodBasic,
		Username:     "testuser",
		Password:     "testpass",
	}

	// Set config
	err = store.Set("delete-test-key", config)
	require.NoError(t, err)

	// Verify it exists
	_, err = store.Get("delete-test-key")
	assert.NoError(t, err)

	// Delete config
	err = store.Delete("delete-test-key")
	assert.NoError(t, err)

	// Verify it's gone
	_, err = store.Get("delete-test-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication not found")
}

func TestList(t *testing.T) {
	// Setup
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Clean up any existing configs
	store.Delete("global:github")
	store.Delete("global:gitlab")

	// Create test configs
	config1 := &AuthConfig{
		ProviderType: git.ProviderGitHub,
		AuthMethod:   AuthMethodToken,
		Token:        "token1",
	}

	config2 := &AuthConfig{
		ProviderType: git.ProviderGitLab,
		AuthMethod:   AuthMethodBasic,
		Username:     "user2",
		Password:     "pass2",
	}

	// Set configs
	err = store.Set("global:github", config1)
	require.NoError(t, err)
	
	err = store.Set("global:gitlab", config2)
	require.NoError(t, err)

	// List global configs
	configs, err := store.List("global:")
	assert.NoError(t, err)
	assert.Len(t, configs, 2)

	// Verify configs
	foundGitHub := false
	foundGitLab := false
	for _, config := range configs {
		if config.ProviderType == git.ProviderGitHub {
			foundGitHub = true
			assert.Equal(t, AuthMethodToken, config.AuthMethod)
			assert.Equal(t, "token1", config.Token)
		} else if config.ProviderType == git.ProviderGitLab {
			foundGitLab = true
			assert.Equal(t, AuthMethodBasic, config.AuthMethod)
			assert.Equal(t, "user2", config.Username)
			assert.Equal(t, "pass2", config.Password)
		}
	}
	assert.True(t, foundGitHub)
	assert.True(t, foundGitLab)

	// Clean up
	store.Delete("global:github")
	store.Delete("global:gitlab")
}

func TestListEmpty(t *testing.T) {
	// Setup
	store, err := NewSecureStore()
	require.NoError(t, err)

	// List with no configs
	configs, err := store.List("empty:")
	assert.NoError(t, err)
	assert.Len(t, configs, 0)
}

func TestEncryptionDecryption(t *testing.T) {
	// Setup
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Test data
	testData := []byte("sensitive-data-that-needs-encryption")

	// Encrypt
	encrypted, err := store.encrypt(testData)
	assert.NoError(t, err)
	assert.NotNil(t, encrypted)
	assert.NotEqual(t, testData, encrypted) // Should be different

	// Decrypt
	decrypted, err := store.decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, testData, decrypted)
}

func TestEncryptionDecryptionWithEmptyData(t *testing.T) {
	// Setup
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Test with empty data
	testData := []byte{}

	// Encrypt
	encrypted, err := store.encrypt(testData)
	assert.NoError(t, err)
	assert.NotNil(t, encrypted)

	// Decrypt
	decrypted, err := store.decrypt(encrypted)
	assert.NoError(t, err)
	assert.Nil(t, decrypted) // Empty data should return nil
}

func TestDecryptInvalidData(t *testing.T) {
	// Setup
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Test with invalid data
	invalidData := []byte("invalid-encrypted-data")

	// Decrypt should fail
	_, err = store.decrypt(invalidData)
	assert.Error(t, err)
}

func TestDecryptShortData(t *testing.T) {
	// Setup
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Test with data too short
	shortData := []byte("short")

	// Decrypt should fail
	_, err = store.decrypt(shortData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ciphertext too short")
}

func TestSanitizeKey(t *testing.T) {
	// Setup
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Test valid keys
	assert.Equal(t, "valid-key", store.sanitizeKey("valid-key"))
	assert.Equal(t, "valid_key", store.sanitizeKey("valid_key"))
	assert.Equal(t, "valid123", store.sanitizeKey("valid123"))

	// Test keys with invalid characters
	assert.Equal(t, "invalid_key", store.sanitizeKey("invalid/key"))
	assert.Equal(t, "invalid_key", store.sanitizeKey("invalid\\key"))
	assert.Equal(t, "invalid_key", store.sanitizeKey("invalid:key"))
	assert.Equal(t, "invalid_key", store.sanitizeKey("invalid key"))
	assert.Equal(t, "invalid_key", store.sanitizeKey("invalid@key"))

	// Test keys with mixed characters
	assert.Equal(t, "mixed_valid_123", store.sanitizeKey("mixed/valid:123"))
}

func TestLoadOrGenerateKey(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "cowork-key-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	keyPath := filepath.Join(tempDir, "test.key")

	// Test generating new key
	key1, err := loadOrGenerateKey(keyPath)
	assert.NoError(t, err)
	assert.Len(t, key1, 32)

	// Test loading existing key
	key2, err := loadOrGenerateKey(keyPath)
	assert.NoError(t, err)
	assert.Len(t, key2, 32)
	assert.Equal(t, key1, key2) // Should be the same key

	// Verify key file exists
	_, err = os.Stat(keyPath)
	assert.NoError(t, err)
}

func TestLoadOrGenerateKeyWithInvalidFile(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "cowork-key-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	keyPath := filepath.Join(tempDir, "invalid.key")

	// Create invalid key file (wrong size)
	err = os.WriteFile(keyPath, []byte("invalid-key"), 0600)
	require.NoError(t, err)

	// Should generate new key
	key, err := loadOrGenerateKey(keyPath)
	assert.NoError(t, err)
	assert.Len(t, key, 32)

	// Verify new key file was created
	fileInfo, err := os.Stat(keyPath)
	assert.NoError(t, err)
	assert.Equal(t, int64(32), fileInfo.Size())
}

func TestSecureStoreWithSpecialCharacters(t *testing.T) {
	// Setup
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Test config with special characters
	config := &AuthConfig{
		ProviderType: git.ProviderBitbucket,
		AuthMethod:   AuthMethodToken,
		Token:        "token-with-special-chars!@#$%^&*()",
		Username:     "user@domain.com",
		Password:     "pass/with\\special:chars",
	}

	// Test setting and getting with special characters in key
	err = store.Set("special:key/with\\chars", config)
	assert.NoError(t, err)

	retrievedConfig, err := store.Get("special:key/with\\chars")
	assert.NoError(t, err)
	assert.Equal(t, config.Token, retrievedConfig.Token)
	assert.Equal(t, config.Username, retrievedConfig.Username)
	assert.Equal(t, config.Password, retrievedConfig.Password)
}

func TestSecureStoreConcurrency(t *testing.T) {
	// Setup
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Test concurrent access
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			config := &AuthConfig{
				ProviderType: git.ProviderGitHub,
				AuthMethod:   AuthMethodToken,
				Token:        fmt.Sprintf("token-%d", id),
			}

			key := fmt.Sprintf("concurrent-key-%d", id)
			
			// Set config
			err := store.Set(key, config)
			assert.NoError(t, err)

			// Get config
			retrievedConfig, err := store.Get(key)
			assert.NoError(t, err)
			assert.Equal(t, config.Token, retrievedConfig.Token)

			// Delete config
			err = store.Delete(key)
			assert.NoError(t, err)

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Benchmark tests
func BenchmarkSet(b *testing.B) {
	store, err := NewSecureStore()
	require.NoError(b, err)

	config := &AuthConfig{
		ProviderType: git.ProviderGitHub,
		AuthMethod:   AuthMethodToken,
		Token:        "benchmark-token",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark-key-%d", i)
		err := store.Set(key, config)
		require.NoError(b, err)
	}
}

func BenchmarkGet(b *testing.B) {
	store, err := NewSecureStore()
	require.NoError(b, err)

	config := &AuthConfig{
		ProviderType: git.ProviderGitHub,
		AuthMethod:   AuthMethodToken,
		Token:        "benchmark-token",
	}

	// Setup
	err = store.Set("benchmark-key", config)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.Get("benchmark-key")
		require.NoError(b, err)
	}
}

func BenchmarkEncrypt(b *testing.B) {
	store, err := NewSecureStore()
	require.NoError(b, err)

	testData := []byte("benchmark-encryption-data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.encrypt(testData)
		require.NoError(b, err)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	store, err := NewSecureStore()
	require.NoError(b, err)

	testData := []byte("benchmark-decryption-data")
	encrypted, err := store.encrypt(testData)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.decrypt(encrypted)
		require.NoError(b, err)
	}
}
