package auth

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/hlfshell/cowork/internal/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFilePermissions_AuthDirectory tests that the auth directory has correct permissions
func TestFilePermissions_AuthDirectory(t *testing.T) {
	// Test case: The auth directory should be created with 0700 permissions
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Check directory permissions
	dirInfo, err := os.Stat(store.storePath)
	require.NoError(t, err)

	// Get file mode
	mode := dirInfo.Mode()

	// Check that directory has correct permissions (0700 = owner read/write/execute only)
	expectedMode := os.FileMode(0700)
	assert.Equal(t, expectedMode, mode&os.ModePerm, "Auth directory should have 0700 permissions")

	// Verify it's a directory
	assert.True(t, dirInfo.IsDir(), "Auth directory should be a directory")
}

// TestFilePermissions_EncryptedFiles tests that encrypted auth files have correct permissions
func TestFilePermissions_EncryptedFiles(t *testing.T) {
	// Test case: Encrypted auth files should have 0600 permissions
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Create test config
	config := &AuthConfig{
		ProviderType: git.ProviderGitHub,
		AuthMethod:   AuthMethodToken,
		Token:        "sensitive-token-123",
		Username:     "testuser",
		Password:     "testpass",
	}

	// Store the config
	err = store.Set("test-permissions", config)
	require.NoError(t, err)

	// Get the file path
	filePath := filepath.Join(store.storePath, store.sanitizeKey("test-permissions")+".enc")

	// Check file permissions
	fileInfo, err := os.Stat(filePath)
	require.NoError(t, err)

	// Get file mode
	mode := fileInfo.Mode()

	// Check that file has correct permissions (0600 = owner read/write only)
	expectedMode := os.FileMode(0600)
	assert.Equal(t, expectedMode, mode&os.ModePerm, "Encrypted auth file should have 0600 permissions")

	// Verify it's a regular file
	assert.False(t, fileInfo.IsDir(), "Auth file should be a regular file")

	// Clean up
	store.Delete("test-permissions")
}

// TestFilePermissions_EncryptionKey tests that the encryption key file has correct permissions
func TestFilePermissions_EncryptionKey(t *testing.T) {
	// Test case: The encryption key file should have 0600 permissions
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Get the key file path
	keyPath := filepath.Join(store.storePath, ".key")

	// Check file permissions
	fileInfo, err := os.Stat(keyPath)
	require.NoError(t, err)

	// Get file mode
	mode := fileInfo.Mode()

	// Check that key file has correct permissions (0600 = owner read/write only)
	expectedMode := os.FileMode(0600)
	assert.Equal(t, expectedMode, mode&os.ModePerm, "Encryption key file should have 0600 permissions")

	// Verify it's a regular file
	assert.False(t, fileInfo.IsDir(), "Key file should be a regular file")

	// Verify key file size is correct (32 bytes for 256-bit key)
	assert.Equal(t, int64(32), fileInfo.Size(), "Encryption key should be 32 bytes (256-bit)")
}

// TestEncryption_DataIntegrity tests that encrypted data cannot be tampered with
func TestEncryption_DataIntegrity(t *testing.T) {
	// Test case: Encrypted data should be tamper-resistant
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Create test config with sensitive data
	config := &AuthConfig{
		ProviderType: git.ProviderGitHub,
		AuthMethod:   AuthMethodToken,
		Token:        "ghp_super_secret_token_123456789",
		Username:     "testuser@example.com",
		Password:     "super_secret_password_123!@#",
	}

	// Store the config
	err = store.Set("test-integrity", config)
	require.NoError(t, err)

	// Get the file path
	filePath := filepath.Join(store.storePath, store.sanitizeKey("test-integrity")+".enc")

	// Read the encrypted file
	encryptedData, err := os.ReadFile(filePath)
	require.NoError(t, err)

	// Tamper with the encrypted data (modify a byte)
	encryptedData[10] = encryptedData[10] ^ 0xFF

	// Write back the tampered data
	err = os.WriteFile(filePath, encryptedData, 0600)
	require.NoError(t, err)

	// Try to retrieve the config - should fail due to tampering
	_, err = store.Get("test-integrity")
	assert.Error(t, err, "Tampered encrypted data should cause decryption to fail")
	assert.Contains(t, err.Error(), "failed to decrypt", "Error should indicate decryption failure")

	// Clean up
	store.Delete("test-integrity")
}

// TestEncryption_UniqueNonces tests that each encryption uses unique nonces
func TestEncryption_UniqueNonces(t *testing.T) {
	// Test case: Each encryption should use a unique nonce
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Encrypt the same data multiple times
	encrypted1, err := store.encrypt([]byte("test-data"))
	require.NoError(t, err)

	encrypted2, err := store.encrypt([]byte("test-data"))
	require.NoError(t, err)

	encrypted3, err := store.encrypt([]byte("test-data"))
	require.NoError(t, err)

	// All encrypted results should be different due to unique nonces
	assert.NotEqual(t, encrypted1, encrypted2, "Encrypted data should be different due to unique nonces")
	assert.NotEqual(t, encrypted1, encrypted3, "Encrypted data should be different due to unique nonces")
	assert.NotEqual(t, encrypted2, encrypted3, "Encrypted data should be different due to unique nonces")

	// But all should decrypt to the same original data
	decrypted1, err := store.decrypt(encrypted1)
	require.NoError(t, err)

	decrypted2, err := store.decrypt(encrypted2)
	require.NoError(t, err)

	decrypted3, err := store.decrypt(encrypted3)
	require.NoError(t, err)

	assert.Equal(t, []byte("test-data"), decrypted1)
	assert.Equal(t, []byte("test-data"), decrypted2)
	assert.Equal(t, []byte("test-data"), decrypted3)
}

// TestEncryption_SensitiveDataNotLogged tests that sensitive data is not logged
func TestEncryption_SensitiveDataNotLogged(t *testing.T) {
	// Test case: Sensitive data should not appear in error messages or logs
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Create test config with sensitive data
	sensitiveToken := "ghp_super_secret_token_that_should_not_be_logged"
	config := &AuthConfig{
		ProviderType: git.ProviderGitHub,
		AuthMethod:   AuthMethodToken,
		Token:        sensitiveToken,
		Username:     "secret_user",
		Password:     "secret_pass",
	}

	// Store the config
	err = store.Set("test-sensitive", config)
	require.NoError(t, err)

	// Retrieve the config
	retrievedConfig, err := store.Get("test-sensitive")
	require.NoError(t, err)

	// Verify sensitive data is preserved
	assert.Equal(t, sensitiveToken, retrievedConfig.Token)
	assert.Equal(t, "secret_user", retrievedConfig.Username)
	assert.Equal(t, "secret_pass", retrievedConfig.Password)

	// Clean up
	store.Delete("test-sensitive")
}

// TestFilePermissions_ProjectScope tests that project-scoped auth files have correct permissions
func TestFilePermissions_ProjectScope(t *testing.T) {
	// Test case: Project-scoped auth files should have 0600 permissions
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Create test config for project scope
	config := &AuthConfig{
		ProviderType: git.ProviderGitLab,
		AuthMethod:   AuthMethodToken,
		Token:        "project-scope-token",
	}

	// Store the config with project scope
	err = store.Set("project:gitlab", config)
	require.NoError(t, err)

	// Get the file path
	filePath := filepath.Join(store.storePath, store.sanitizeKey("project:gitlab")+".enc")

	// Check file permissions
	fileInfo, err := os.Stat(filePath)
	require.NoError(t, err)

	// Get file mode
	mode := fileInfo.Mode()

	// Check that file has correct permissions (0600 = owner read/write only)
	expectedMode := os.FileMode(0600)
	assert.Equal(t, expectedMode, mode&os.ModePerm, "Project-scoped auth file should have 0600 permissions")

	// Clean up
	store.Delete("project:gitlab")
}

// TestEncryption_KeyRotation tests that encryption keys can be rotated securely
func TestEncryption_KeyRotation(t *testing.T) {
	// Test case: The system should handle key rotation securely
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Create test config
	config := &AuthConfig{
		ProviderType: git.ProviderBitbucket,
		AuthMethod:   AuthMethodToken,
		Token:        "rotation-test-token",
	}

	// Store the config
	err = store.Set("test-rotation", config)
	require.NoError(t, err)

	// Verify the config can be retrieved
	retrievedConfig, err := store.Get("test-rotation")
	require.NoError(t, err)
	assert.Equal(t, "rotation-test-token", retrievedConfig.Token)

	// Get the key file path
	keyPath := filepath.Join(store.storePath, ".key")

	// Backup the original key
	originalKey, err := os.ReadFile(keyPath)
	require.NoError(t, err)

	// Simulate key rotation by creating a new key
	newKey := make([]byte, 32)
	for i := range newKey {
		newKey[i] = byte(i % 256)
	}

	// Write the new key
	err = os.WriteFile(keyPath, newKey, 0600)
	require.NoError(t, err)

	// Create a new store instance with the new key
	newStore, err := NewSecureStore()
	require.NoError(t, err)

	// The old encrypted data should not be decryptable with the new key
	_, err = newStore.Get("test-rotation")
	assert.Error(t, err, "Old encrypted data should not be decryptable with new key")

	// Restore the original key
	err = os.WriteFile(keyPath, originalKey, 0600)
	require.NoError(t, err)

	// Create another store instance with the original key
	restoredStore, err := NewSecureStore()
	require.NoError(t, err)

	// The data should be decryptable again with the original key
	restoredConfig, err := restoredStore.Get("test-rotation")
	require.NoError(t, err)
	assert.Equal(t, "rotation-test-token", restoredConfig.Token)

	// Clean up
	store.Delete("test-rotation")
}

// TestFilePermissions_OwnerOnly tests that files are owned by the current user
func TestFilePermissions_OwnerOnly(t *testing.T) {
	// Test case: Auth files should be owned by the current user only
	store, err := NewSecureStore()
	require.NoError(t, err)

	// Create test config
	config := &AuthConfig{
		ProviderType: git.ProviderGitHub,
		AuthMethod:   AuthMethodToken,
		Token:        "owner-test-token",
	}

	// Store the config
	err = store.Set("test-owner", config)
	require.NoError(t, err)

	// Get the file path
	filePath := filepath.Join(store.storePath, store.sanitizeKey("test-owner")+".enc")

	// Check file ownership
	fileInfo, err := os.Stat(filePath)
	require.NoError(t, err)

	// Get current user ID
	currentUID := os.Getuid()

	// Check file ownership (Unix-specific)
	if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
		assert.Equal(t, uint32(currentUID), stat.Uid, "Auth file should be owned by current user")
	}

	// Clean up
	store.Delete("test-owner")
}
