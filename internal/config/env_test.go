package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestManager_SaveEnv_WithLocalCwDir tests saving environment variables to local .cw directory
func TestManager_SaveEnv_WithLocalCwDir(t *testing.T) {
	// Test case: Saving environment variables to local .cw directory should succeed
	// and create encrypted files with proper permissions

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "cowork-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(tempDir))

	// Create .cw directory
	cwDir := filepath.Join(tempDir, ".cw")
	require.NoError(t, os.MkdirAll(cwDir, 0755))

	// Create manager
	manager := &Manager{
		GlobalConfigPath:  filepath.Join(tempDir, "global", ".cwconfig"),
		ProjectConfigPath: filepath.Join(tempDir, "project", ".cwconfig"),
		config: &Config{
			Env: map[string]string{
				"OPENAI_API_KEY":    "sk-test-key-123",
				"ANTHROPIC_API_KEY": "sk-ant-test-key-456",
				"GEMINI_API_KEY":    "gemini-test-key-789",
			},
		},
	}

	// Save environment variables
	err = manager.SaveEnv()
	require.NoError(t, err)

	// Verify encrypted file was created
	envFile := filepath.Join(cwDir, "env", "env.enc")
	assert.FileExists(t, envFile)

	// Verify key file was created
	keyFile := filepath.Join(cwDir, "env", ".key")
	assert.FileExists(t, keyFile)

	// Check file permissions
	fileInfo, err := os.Stat(envFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), fileInfo.Mode()&os.ModePerm)

	keyInfo, err := os.Stat(keyFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), keyInfo.Mode()&os.ModePerm)
}

// TestManager_SaveEnv_WithGlobalCwDir tests saving environment variables to global .cw directory
func TestManager_SaveEnv_WithGlobalCwDir(t *testing.T) {
	// Test case: When no local .cw directory exists, environment variables should be saved
	// to the global .cw directory

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "cowork-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(tempDir))

	// Create manager with custom global path
	globalCwDir := filepath.Join(tempDir, ".config", "cowork", ".cw")
	manager := &Manager{
		GlobalConfigPath:  filepath.Join(tempDir, "global", ".cwconfig"),
		ProjectConfigPath: filepath.Join(tempDir, "project", ".cwconfig"),
		config: &Config{
			Env: map[string]string{
				"OPENAI_API_KEY":    "sk-test-key-123",
				"ANTHROPIC_API_KEY": "sk-ant-test-key-456",
			},
		},
	}

	// Save environment variables
	err = manager.SaveEnv()
	require.NoError(t, err)

	// Verify encrypted file was created in global directory
	envFile := filepath.Join(globalCwDir, "env", "env.enc")
	assert.FileExists(t, envFile)

	// Verify key file was created
	keyFile := filepath.Join(globalCwDir, "env", ".key")
	assert.FileExists(t, keyFile)
}

// TestManager_LoadEnv_WithLocalCwDir tests loading environment variables from local .cw directory
func TestManager_LoadEnv_WithLocalCwDir(t *testing.T) {
	// Test case: Loading environment variables from local .cw directory should succeed
	// and populate the config with the correct values

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "cowork-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(tempDir))

	// Create .cw directory
	cwDir := filepath.Join(tempDir, ".cw")
	require.NoError(t, os.MkdirAll(cwDir, 0755))

	// Create manager with test environment variables
	testEnv := map[string]string{
		"OPENAI_API_KEY":    "sk-test-key-123",
		"ANTHROPIC_API_KEY": "sk-ant-test-key-456",
		"GEMINI_API_KEY":    "gemini-test-key-789",
	}

	manager := &Manager{
		GlobalConfigPath:  filepath.Join(tempDir, "global", ".cwconfig"),
		ProjectConfigPath: filepath.Join(tempDir, "project", ".cwconfig"),
		config: &Config{
			Env: testEnv,
		},
	}

	// Save environment variables
	err = manager.SaveEnv()
	require.NoError(t, err)

	// Create new manager to test loading
	newManager := &Manager{
		GlobalConfigPath:  filepath.Join(tempDir, "global", ".cwconfig"),
		ProjectConfigPath: filepath.Join(tempDir, "project", ".cwconfig"),
		config: &Config{
			Env: map[string]string{},
		},
	}

	// Load environment variables
	err = newManager.LoadEnv()
	require.NoError(t, err)

	// Verify environment variables were loaded correctly
	assert.Equal(t, testEnv, newManager.config.Env)
}

// TestManager_LoadEnv_WithGlobalCwDir tests loading environment variables from global .cw directory
func TestManager_LoadEnv_WithGlobalCwDir(t *testing.T) {
	// Test case: When no local .cw directory exists, environment variables should be loaded
	// from the global .cw directory

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "cowork-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(tempDir))

	// Create global .cw directory
	globalCwDir := filepath.Join(tempDir, ".config", "cowork", ".cw")
	require.NoError(t, os.MkdirAll(globalCwDir, 0755))

	// Create manager with test environment variables
	testEnv := map[string]string{
		"OPENAI_API_KEY":    "sk-test-key-123",
		"ANTHROPIC_API_KEY": "sk-ant-test-key-456",
	}

	manager := &Manager{
		GlobalConfigPath:  filepath.Join(tempDir, "global", ".cwconfig"),
		ProjectConfigPath: filepath.Join(tempDir, "project", ".cwconfig"),
		config: &Config{
			Env: testEnv,
		},
	}

	// Save environment variables
	err = manager.SaveEnv()
	require.NoError(t, err)

	// Create new manager to test loading
	newManager := &Manager{
		GlobalConfigPath:  filepath.Join(tempDir, "global", ".cwconfig"),
		ProjectConfigPath: filepath.Join(tempDir, "project", ".cwconfig"),
		config: &Config{
			Env: map[string]string{},
		},
	}

	// Load environment variables
	err = newManager.LoadEnv()
	require.NoError(t, err)

	// Verify environment variables were loaded correctly
	assert.Equal(t, testEnv, newManager.config.Env)
}

// TestManager_LoadEnv_WithNoEnvFile tests loading when no encrypted env file exists
func TestManager_LoadEnv_WithNoEnvFile(t *testing.T) {
	// Test case: Loading environment variables when no encrypted file exists should
	// not return an error and should not modify the existing config

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "cowork-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(tempDir))

	// Create manager with empty environment
	manager := &Manager{
		GlobalConfigPath:  filepath.Join(tempDir, "global", ".cwconfig"),
		ProjectConfigPath: filepath.Join(tempDir, "project", ".cwconfig"),
		config: &Config{
			Env: map[string]string{},
		},
	}

	// Load environment variables (should not error)
	err = manager.LoadEnv()
	require.NoError(t, err)

	// Verify environment is still empty
	assert.Empty(t, manager.config.Env)
}

// TestManager_SaveEnv_And_LoadEnv_Encryption tests that saved environment variables are properly encrypted
func TestManager_SaveEnv_And_LoadEnv_Encryption(t *testing.T) {
	// Test case: Environment variables should be properly encrypted and decrypted
	// with data integrity preserved

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "cowork-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(tempDir))

	// Create .cw directory
	cwDir := filepath.Join(tempDir, ".cw")
	require.NoError(t, os.MkdirAll(cwDir, 0755))

	// Create manager with sensitive environment variables
	sensitiveEnv := map[string]string{
		"OPENAI_API_KEY":    "sk-super-secret-key-that-should-be-encrypted",
		"ANTHROPIC_API_KEY": "sk-ant-super-secret-key-that-should-be-encrypted",
		"GEMINI_API_KEY":    "gemini-super-secret-key-that-should-be-encrypted",
		"OTHER_SECRET":      "another-super-secret-value",
	}

	manager := &Manager{
		GlobalConfigPath:  filepath.Join(tempDir, "global", ".cwconfig"),
		ProjectConfigPath: filepath.Join(tempDir, "project", ".cwconfig"),
		config: &Config{
			Env: sensitiveEnv,
		},
	}

	// Save environment variables
	err = manager.SaveEnv()
	require.NoError(t, err)

	// Verify the encrypted file is not readable as plain text
	envFile := filepath.Join(cwDir, "env", "env.enc")
	encryptedData, err := os.ReadFile(envFile)
	require.NoError(t, err)

	// The encrypted data should not contain the original sensitive values
	encryptedString := string(encryptedData)
	assert.NotContains(t, encryptedString, "sk-super-secret-key-that-should-be-encrypted")
	assert.NotContains(t, encryptedString, "sk-ant-super-secret-key-that-should-be-encrypted")
	assert.NotContains(t, encryptedString, "gemini-super-secret-key-that-should-be-encrypted")

	// Create new manager to test loading
	newManager := &Manager{
		GlobalConfigPath:  filepath.Join(tempDir, "global", ".cwconfig"),
		ProjectConfigPath: filepath.Join(tempDir, "project", ".cwconfig"),
		config: &Config{
			Env: map[string]string{},
		},
	}

	// Load environment variables
	err = newManager.LoadEnv()
	require.NoError(t, err)

	// Verify all sensitive environment variables were loaded correctly
	assert.Equal(t, sensitiveEnv, newManager.config.Env)
}

// TestManager_AddEnvVar_RemoveEnvVar tests adding and removing environment variables
func TestManager_AddEnvVar_RemoveEnvVar(t *testing.T) {
	// Test case: Adding and removing environment variables should work correctly

	manager := &Manager{
		config: &Config{
			Env: map[string]string{},
		},
	}

	// Test adding environment variables
	manager.AddEnvVar("TEST_KEY_1", "test_value_1")
	manager.AddEnvVar("TEST_KEY_2", "test_value_2")

	// Verify environment variables were added
	assert.Equal(t, "test_value_1", manager.config.Env["TEST_KEY_1"])
	assert.Equal(t, "test_value_2", manager.config.Env["TEST_KEY_2"])

	// Test removing environment variables
	manager.RemoveEnvVar("TEST_KEY_1")

	// Verify environment variable was removed
	assert.NotContains(t, manager.config.Env, "TEST_KEY_1")
	assert.Equal(t, "test_value_2", manager.config.Env["TEST_KEY_2"])
}

// TestManager_GetEnv tests getting environment variables
func TestManager_GetEnv(t *testing.T) {
	// Test case: Getting environment variables should return the correct map

	testEnv := map[string]string{
		"KEY_1": "value_1",
		"KEY_2": "value_2",
	}

	manager := &Manager{
		config: &Config{
			Env: testEnv,
		},
	}

	// Get environment variables
	result := manager.GetEnv()

	// Verify the returned map matches the original
	assert.Equal(t, testEnv, result)
}
