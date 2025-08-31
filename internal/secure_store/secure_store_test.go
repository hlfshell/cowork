package secure_store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthConfig represents the old AuthConfig structure for backward compatibility tests
type TestAuthConfig struct {
	ProviderType string     `json:"provider_type"`
	AuthMethod   string     `json:"auth_method"`
	Token        string     `json:"token,omitempty"`
	Username     string     `json:"username,omitempty"`
	Password     string     `json:"password,omitempty"`
	BaseURL      string     `json:"base_url,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

func TestNewSecureStore(t *testing.T) {
	// Test creating a new secure store
	store, err := NewSecureStore("test-store")

	assert.NoError(t, err)
	assert.NotNil(t, store)
	assert.NotEmpty(t, store.storePath)
	assert.Len(t, store.key, 32) // 256-bit key
}

func TestSetAndGetString(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Test setting string
	err = store.SetString("test-string-key", "test-value")
	assert.NoError(t, err)

	// Test getting string
	value, err := store.GetString("test-string-key")
	assert.NoError(t, err)
	assert.Equal(t, "test-value", value)
}

func TestSetAndGetInt(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Test setting int
	err = store.SetInt("test-int-key", 42)
	assert.NoError(t, err)

	// Test getting int
	value, err := store.GetInt("test-int-key")
	assert.NoError(t, err)
	assert.Equal(t, 42, value)
}

func TestSetAndGetBool(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Test setting bool
	err = store.SetBool("test-bool-key", true)
	assert.NoError(t, err)

	// Test getting bool
	value, err := store.GetBool("test-bool-key")
	assert.NoError(t, err)
	assert.Equal(t, true, value)
}

func TestSetAndGetFloat(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Test setting float
	err = store.SetFloat("test-float-key", 3.14159)
	assert.NoError(t, err)

	// Test getting float
	value, err := store.GetFloat("test-float-key")
	assert.NoError(t, err)
	assert.Equal(t, 3.14159, value)
}

func TestSetAndGetBytes(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Test setting bytes
	testBytes := []byte("test-bytes-data")
	err = store.SetBytes("test-bytes-key", testBytes)
	assert.NoError(t, err)

	// Test getting bytes
	value, err := store.GetBytes("test-bytes-key")
	assert.NoError(t, err)
	assert.Equal(t, testBytes, value)
}

func TestSetAndGetStruct(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Create test config
	config := &TestAuthConfig{
		ProviderType: "github",
		AuthMethod:   "token",
		Token:        "test-token-123",
		Username:     "testuser",
		Password:     "testpass",
	}

	// Test setting struct
	err = store.Set("test-struct-key", config)
	assert.NoError(t, err)

	// Test getting struct
	var retrievedConfig TestAuthConfig
	err = store.Get("test-struct-key", &retrievedConfig)
	assert.NoError(t, err)
	assert.Equal(t, config.ProviderType, retrievedConfig.ProviderType)
	assert.Equal(t, config.AuthMethod, retrievedConfig.AuthMethod)
	assert.Equal(t, config.Token, retrievedConfig.Token)
	assert.Equal(t, config.Username, retrievedConfig.Username)
	assert.Equal(t, config.Password, retrievedConfig.Password)
}

func TestGetNonExistent(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Test getting non-existent string
	_, err = store.GetString("non-existent-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "value not found")

	// Test getting non-existent struct
	var config TestAuthConfig
	err = store.Get("non-existent-key", &config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "value not found")
}

func TestDelete(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Set value
	err = store.SetString("delete-test-key", "test-value")
	require.NoError(t, err)

	// Verify it exists
	value, err := store.GetString("delete-test-key")
	assert.NoError(t, err)
	assert.Equal(t, "test-value", value)

	// Delete value
	err = store.Delete("delete-test-key")
	assert.NoError(t, err)

	// Verify it's gone
	_, err = store.GetString("delete-test-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "value not found")
}

func TestList(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Clean up any existing keys
	store.Delete("global:github")
	store.Delete("global:gitlab")

	// Create test values
	err = store.SetString("global:github", "github-token")
	require.NoError(t, err)

	err = store.SetString("global:gitlab", "gitlab-token")
	require.NoError(t, err)

	// List global keys
	keys, err := store.List("global:")
	assert.NoError(t, err)
	assert.Len(t, keys, 2)

	// Verify keys contain expected values
	foundGitHub := false
	foundGitLab := false
	for _, key := range keys {
		if key == "global_github" {
			foundGitHub = true
		} else if key == "global_gitlab" {
			foundGitLab = true
		}
	}
	assert.True(t, foundGitHub)
	assert.True(t, foundGitLab)

	// Clean up
	store.Delete("global:github")
	store.Delete("global:gitlab")
}

func TestListWithValues(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Clean up any existing keys
	store.Delete("config:setting1")
	store.Delete("config:setting2")

	// Create test values
	err = store.SetInt("config:setting1", 100)
	require.NoError(t, err)

	err = store.SetBool("config:setting2", true)
	require.NoError(t, err)

	// List with values (using interface{} type)
	values, err := store.ListWithValues("config:", interface{}(nil))
	assert.NoError(t, err)
	assert.Len(t, values, 2)

	// Clean up
	store.Delete("config:setting1")
	store.Delete("config:setting2")
}

func TestListEmpty(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// List with no values
	keys, err := store.List("empty:")
	assert.NoError(t, err)
	assert.Len(t, keys, 0)
}

func TestExists(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Test non-existent key
	exists := store.Exists("non-existent-key")
	assert.False(t, exists)

	// Set a key
	err = store.SetString("test-exists-key", "value")
	require.NoError(t, err)

	// Test existing key
	exists = store.Exists("test-exists-key")
	assert.True(t, exists)

	// Clean up
	store.Delete("test-exists-key")
}

func TestClear(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Set some values
	err = store.SetString("clear:key1", "value1")
	require.NoError(t, err)
	err = store.SetString("clear:key2", "value2")
	require.NoError(t, err)

	// Verify they exist
	assert.True(t, store.Exists("clear:key1"))
	assert.True(t, store.Exists("clear:key2"))

	// Clear all values
	err = store.Clear()
	assert.NoError(t, err)

	// Verify they're gone
	assert.False(t, store.Exists("clear:key1"))
	assert.False(t, store.Exists("clear:key2"))
}

func TestGetSize(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Clear any existing values
	store.Clear()

	// Test empty store
	size, err := store.GetSize()
	assert.NoError(t, err)
	assert.Equal(t, 0, size)

	// Add some values
	err = store.SetString("size:key1", "value1")
	require.NoError(t, err)
	err = store.SetString("size:key2", "value2")
	require.NoError(t, err)

	// Test size with values
	size, err = store.GetSize()
	assert.NoError(t, err)
	assert.Equal(t, 2, size)

	// Clean up
	store.Clear()
}

func TestEncryptionDecryption(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
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
	store, err := NewSecureStore("test-store")
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
	// JSON marshaling of empty slice results in null, so we expect nil
	assert.Nil(t, decrypted)
}

func TestEncryptionDecryptionWithEmptyString(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Test with empty string
	testData := ""

	// Encrypt
	encrypted, err := store.encrypt([]byte(testData))
	assert.NoError(t, err)
	assert.NotNil(t, encrypted)

	// Decrypt
	decrypted, err := store.decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, testData, string(decrypted))
}

func TestDecryptInvalidData(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Test with invalid data
	invalidData := []byte("invalid-encrypted-data")

	// Decrypt should fail
	_, err = store.decrypt(invalidData)
	assert.Error(t, err)
}

func TestDecryptShortData(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
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
	store, err := NewSecureStore("test-store")
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
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Test value with special characters
	testValue := "value-with-special-chars!@#$%^&*()"

	// Test setting and getting with special characters in key
	err = store.SetString("special:key/with\\chars", testValue)
	assert.NoError(t, err)

	retrievedValue, err := store.GetString("special:key/with\\chars")
	assert.NoError(t, err)
	assert.Equal(t, testValue, retrievedValue)
}

func TestSecureStoreConcurrency(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			testValue := fmt.Sprintf("value-%d", id)
			key := fmt.Sprintf("concurrent-key-%d", id)

			// Set value
			err := store.SetString(key, testValue)
			assert.NoError(t, err)

			// Get value
			retrievedValue, err := store.GetString(key)
			assert.NoError(t, err)
			assert.Equal(t, testValue, retrievedValue)

			// Delete value
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

func TestTypeConversion(t *testing.T) {
	// Setup
	store, err := NewSecureStore("test-store")
	require.NoError(t, err)

	// Test storing as one type and retrieving as another
	err = store.SetInt("conversion-key", 42)
	require.NoError(t, err)

	// Try to get as string (should fail gracefully)
	_, err = store.GetString("conversion-key")
	assert.Error(t, err)

	// Get as int (should work)
	value, err := store.GetInt("conversion-key")
	assert.NoError(t, err)
	assert.Equal(t, 42, value)

	// Clean up
	store.Delete("conversion-key")
}

// Benchmark tests
func BenchmarkSetString(b *testing.B) {
	store, err := NewSecureStore("benchmark-store")
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark-key-%d", i)
		err := store.SetString(key, "benchmark-value")
		require.NoError(b, err)
	}
}

func BenchmarkGetString(b *testing.B) {
	store, err := NewSecureStore("benchmark-store")
	require.NoError(b, err)

	// Setup
	err = store.SetString("benchmark-key", "benchmark-value")
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.GetString("benchmark-key")
		require.NoError(b, err)
	}
}

func BenchmarkSetStruct(b *testing.B) {
	store, err := NewSecureStore("benchmark-store")
	require.NoError(b, err)

	config := &TestAuthConfig{
		ProviderType: "github",
		AuthMethod:   "token",
		Token:        "benchmark-token",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark-key-%d", i)
		err := store.Set(key, config)
		require.NoError(b, err)
	}
}

func BenchmarkGetStruct(b *testing.B) {
	store, err := NewSecureStore("benchmark-store")
	require.NoError(b, err)

	config := &TestAuthConfig{
		ProviderType: "github",
		AuthMethod:   "token",
		Token:        "benchmark-token",
	}

	// Setup
	err = store.Set("benchmark-key", config)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var retrievedConfig TestAuthConfig
		err := store.Get("benchmark-key", &retrievedConfig)
		require.NoError(b, err)
	}
}

func BenchmarkEncrypt(b *testing.B) {
	store, err := NewSecureStore("benchmark-store")
	require.NoError(b, err)

	testData := []byte("benchmark-encryption-data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.encrypt(testData)
		require.NoError(b, err)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	store, err := NewSecureStore("benchmark-store")
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
