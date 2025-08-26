package auth

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hlfshell/cowork/internal/config"
	"github.com/hlfshell/cowork/internal/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	// Test creating a new auth manager
	configManager := config.NewManager()
	authManager, err := NewManager(configManager)

	assert.NoError(t, err)
	assert.NotNil(t, authManager)
	assert.NotNil(t, authManager.authStore)
}

func TestSetToken(t *testing.T) {
	// Setup
	configManager := config.NewManager()
	authManager, err := NewManager(configManager)
	require.NoError(t, err)

	// Clean up any existing configs
	authManager.RemoveAuth(git.ProviderGitHub, AuthScopeGlobal)

	// Test setting a token
	err = authManager.SetToken(git.ProviderGitHub, "test-token", AuthScopeGlobal)
	assert.NoError(t, err)

	// Verify the token was saved
	authConfig, err := authManager.GetAuthConfig(git.ProviderGitHub, AuthScopeGlobal)
	assert.NoError(t, err)
	assert.Equal(t, git.ProviderGitHub, authConfig.ProviderType)
	assert.Equal(t, AuthMethodToken, authConfig.AuthMethod)
	assert.Equal(t, "test-token", authConfig.Token)

	// Clean up
	authManager.RemoveAuth(git.ProviderGitHub, AuthScopeGlobal)
}

func TestSetBasicAuth(t *testing.T) {
	// Setup
	configManager := config.NewManager()
	authManager, err := NewManager(configManager)
	require.NoError(t, err)

	// Clean up any existing configs
	authManager.RemoveAuth(git.ProviderGitLab, AuthScopeProject)

	// Test setting basic auth
	err = authManager.SetBasicAuth(git.ProviderGitLab, "testuser", "testpass", AuthScopeProject)
	assert.NoError(t, err)

	// Verify the credentials were saved
	authConfig, err := authManager.GetAuthConfig(git.ProviderGitLab, AuthScopeProject)
	assert.NoError(t, err)
	assert.Equal(t, git.ProviderGitLab, authConfig.ProviderType)
	assert.Equal(t, AuthMethodBasic, authConfig.AuthMethod)
	assert.Equal(t, "testuser", authConfig.Username)
	assert.Equal(t, "testpass", authConfig.Password)

	// Clean up
	authManager.RemoveAuth(git.ProviderGitLab, AuthScopeProject)
}

func TestGetAuthConfig(t *testing.T) {
	// Setup
	configManager := config.NewManager()
	authManager, err := NewManager(configManager)
	require.NoError(t, err)

	// Clean up any existing configs
	authManager.RemoveAuth(git.ProviderBitbucket, AuthScopeGlobal)

	// Test getting non-existent config
	_, err = authManager.GetAuthConfig(git.ProviderBitbucket, AuthScopeGlobal)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication not found")

	// Set a config and then get it
	err = authManager.SetToken(git.ProviderBitbucket, "test-token", AuthScopeGlobal)
	require.NoError(t, err)

	authConfig, err := authManager.GetAuthConfig(git.ProviderBitbucket, AuthScopeGlobal)
	assert.NoError(t, err)
	assert.Equal(t, git.ProviderBitbucket, authConfig.ProviderType)
	assert.Equal(t, AuthMethodToken, authConfig.AuthMethod)
	assert.Equal(t, "test-token", authConfig.Token)

	// Clean up
	authManager.RemoveAuth(git.ProviderBitbucket, AuthScopeGlobal)
}

func TestRemoveAuth(t *testing.T) {
	// Setup
	configManager := config.NewManager()
	authManager, err := NewManager(configManager)
	require.NoError(t, err)

	// Clean up any existing configs
	authManager.RemoveAuth(git.ProviderGitHub, AuthScopeGlobal)

	// Set a config
	err = authManager.SetToken(git.ProviderGitHub, "test-token", AuthScopeGlobal)
	require.NoError(t, err)

	// Verify it exists
	_, err = authManager.GetAuthConfig(git.ProviderGitHub, AuthScopeGlobal)
	assert.NoError(t, err)

	// Remove it
	err = authManager.RemoveAuth(git.ProviderGitHub, AuthScopeGlobal)
	assert.NoError(t, err)

	// Verify it's gone
	_, err = authManager.GetAuthConfig(git.ProviderGitHub, AuthScopeGlobal)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication not found")
}

func TestListAuthConfigs(t *testing.T) {
	// Setup
	configManager := config.NewManager()
	authManager, err := NewManager(configManager)
	require.NoError(t, err)

	// Set a single config and verify it's listed
	err = authManager.SetToken(git.ProviderGitHub, "test-token", AuthScopeGlobal)
	require.NoError(t, err)

	// List all configs
	configs, err := authManager.ListAuthConfigs()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(configs), 1)

	// Find our test config
	var found bool
	for _, config := range configs {
		if config.ProviderType == git.ProviderGitHub && config.AuthMethod == AuthMethodToken {
			assert.Equal(t, "test-token", config.Token)
			found = true
			break
		}
	}
	assert.True(t, found, "GitHub token config not found in list")

	// Clean up
	authManager.RemoveAuth(git.ProviderGitHub, AuthScopeGlobal)
}

func TestAuthenticateProvider(t *testing.T) {
	// Setup
	configManager := config.NewManager()
	authManager, err := NewManager(configManager)
	require.NoError(t, err)

	// Test token authentication (should fail since it requires manual input)
	_, err = authManager.AuthenticateProvider(context.Background(), git.ProviderGitHub, AuthMethodToken, AuthScopeGlobal)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token authentication requires manual token input")

	// Test basic authentication (should fail since it requires manual input)
	_, err = authManager.AuthenticateProvider(context.Background(), git.ProviderGitHub, AuthMethodBasic, AuthScopeGlobal)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "basic authentication requires manual credentials input")

	// Test unsupported method
	_, err = authManager.AuthenticateProvider(context.Background(), git.ProviderGitHub, "unsupported", AuthScopeGlobal)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported authentication method")
}

func TestTestAuth(t *testing.T) {
	// Setup
	configManager := config.NewManager()
	authManager, err := NewManager(configManager)
	require.NoError(t, err)

	// Test with no authentication configured
	err = authManager.TestAuth(context.Background(), git.ProviderGitHub, AuthScopeGlobal)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no authentication configured")

	// Set a token and test (this will fail because the provider will fail to authenticate)
	err = authManager.SetToken(git.ProviderGitHub, "invalid-token", AuthScopeGlobal)
	require.NoError(t, err)

	err = authManager.TestAuth(context.Background(), git.ProviderGitHub, AuthScopeGlobal)
	// This will fail because the GitHub provider will fail to authenticate with an invalid token
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GitHub authentication failed")
}

func TestGenerateState(t *testing.T) {
	// Setup
	configManager := config.NewManager()
	authManager, err := NewManager(configManager)
	require.NoError(t, err)

	// Generate state
	state1, err := authManager.generateState()
	assert.NoError(t, err)
	assert.NotEmpty(t, state1)
	assert.Len(t, state1, 44) // base64 encoded 32 bytes (with padding)

	// Generate another state
	state2, err := authManager.generateState()
	assert.NoError(t, err)
	assert.NotEmpty(t, state2)
	assert.NotEqual(t, state1, state2) // Should be different
}

func TestCreateProvider(t *testing.T) {
	// Setup
	configManager := config.NewManager()
	authManager, err := NewManager(configManager)
	require.NoError(t, err)

	authConfig := &AuthConfig{
		ProviderType: git.ProviderGitHub,
		AuthMethod:   AuthMethodToken,
		Token:        "test-token",
	}

	// Test creating GitHub provider
	provider, err := authManager.createProvider(git.ProviderGitHub, authConfig)
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, git.ProviderGitHub, provider.GetProviderType())

	// Test creating GitLab provider
	authConfig.ProviderType = git.ProviderGitLab
	provider, err = authManager.createProvider(git.ProviderGitLab, authConfig)
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, git.ProviderGitLab, provider.GetProviderType())

	// Test creating Bitbucket provider
	authConfig.ProviderType = git.ProviderBitbucket
	provider, err = authManager.createProvider(git.ProviderBitbucket, authConfig)
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, git.ProviderBitbucket, provider.GetProviderType())

	// Test unsupported provider
	_, err = authManager.createProvider("unsupported", authConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported provider type")
}

func TestGetAuthKey(t *testing.T) {
	// Setup
	configManager := config.NewManager()
	authManager, err := NewManager(configManager)
	require.NoError(t, err)

	// Test global scope
	key := authManager.getAuthKey(git.ProviderGitHub, AuthScopeGlobal)
	assert.Equal(t, "github_global", key)

	// Test project scope
	key = authManager.getAuthKey(git.ProviderGitLab, AuthScopeProject)
	assert.Equal(t, "gitlab_project", key)

	// Test Bitbucket
	key = authManager.getAuthKey(git.ProviderBitbucket, AuthScopeGlobal)
	assert.Equal(t, "bitbucket_global", key)
}

// Test helper functions
func TestParseProviderType(t *testing.T) {
	// Test valid providers
	providerType, err := parseProviderType("github")
	assert.NoError(t, err)
	assert.Equal(t, git.ProviderGitHub, providerType)

	providerType, err = parseProviderType("GITHUB")
	assert.NoError(t, err)
	assert.Equal(t, git.ProviderGitHub, providerType)

	providerType, err = parseProviderType("gitlab")
	assert.NoError(t, err)
	assert.Equal(t, git.ProviderGitLab, providerType)

	providerType, err = parseProviderType("bitbucket")
	assert.NoError(t, err)
	assert.Equal(t, git.ProviderBitbucket, providerType)

	// Test invalid provider
	_, err = parseProviderType("unsupported")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported provider")
}

func TestParseAuthScope(t *testing.T) {
	// Test valid scopes
	scope, err := parseAuthScope("global")
	assert.NoError(t, err)
	assert.Equal(t, AuthScopeGlobal, scope)

	scope, err = parseAuthScope("GLOBAL")
	assert.NoError(t, err)
	assert.Equal(t, AuthScopeGlobal, scope)

	scope, err = parseAuthScope("project")
	assert.NoError(t, err)
	assert.Equal(t, AuthScopeProject, scope)

	// Test invalid scope
	_, err = parseAuthScope("unsupported")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported scope")
}

// Helper functions for testing
func parseProviderType(providerName string) (git.ProviderType, error) {
	switch strings.ToLower(providerName) {
	case "github":
		return git.ProviderGitHub, nil
	case "gitlab":
		return git.ProviderGitLab, nil
	case "bitbucket":
		return git.ProviderBitbucket, nil
	default:
		return "", fmt.Errorf("unsupported provider: %s (supported: github, gitlab, bitbucket)", providerName)
	}
}

func parseAuthScope(scope string) (AuthScope, error) {
	switch strings.ToLower(scope) {
	case "global":
		return AuthScopeGlobal, nil
	case "project":
		return AuthScopeProject, nil
	default:
		return "", fmt.Errorf("unsupported scope: %s (supported: global, project)", scope)
	}
}

// Integration test with temporary directory
func TestAuthManagerIntegration(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "cowork-auth-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Setup
	configManager := config.NewManager()
	authManager, err := NewManager(configManager)
	require.NoError(t, err)

	// Clean up any existing configs
	authManager.RemoveAuth(git.ProviderGitHub, AuthScopeGlobal)

	// Test full workflow
	// 1. Set token
	err = authManager.SetToken(git.ProviderGitHub, "integration-test-token", AuthScopeGlobal)
	assert.NoError(t, err)

	// 2. Get token
	authConfig, err := authManager.GetAuthConfig(git.ProviderGitHub, AuthScopeGlobal)
	assert.NoError(t, err)
	assert.Equal(t, "integration-test-token", authConfig.Token)

	// 3. List configs
	configs, err := authManager.ListAuthConfigs()
	assert.NoError(t, err)
	assert.Len(t, configs, 1)

	// 4. Remove token
	err = authManager.RemoveAuth(git.ProviderGitHub, AuthScopeGlobal)
	assert.NoError(t, err)

	// 5. Verify removal
	_, err = authManager.GetAuthConfig(git.ProviderGitHub, AuthScopeGlobal)
	assert.Error(t, err)
}

// Benchmark tests
func BenchmarkSetToken(b *testing.B) {
	configManager := config.NewManager()
	authManager, err := NewManager(configManager)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := authManager.SetToken(git.ProviderGitHub, "benchmark-token", AuthScopeGlobal)
		require.NoError(b, err)
	}

	// Clean up
	authManager.RemoveAuth(git.ProviderGitHub, AuthScopeGlobal)
}

func BenchmarkGetAuthConfig(b *testing.B) {
	configManager := config.NewManager()
	authManager, err := NewManager(configManager)
	require.NoError(b, err)

	// Setup
	err = authManager.SetToken(git.ProviderGitHub, "benchmark-token", AuthScopeGlobal)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := authManager.GetAuthConfig(git.ProviderGitHub, AuthScopeGlobal)
		require.NoError(b, err)
	}

	// Clean up
	authManager.RemoveAuth(git.ProviderGitHub, AuthScopeGlobal)
}
