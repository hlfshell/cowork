package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/hlfshell/cowork/internal/config"
	"github.com/hlfshell/cowork/internal/git"
	gitprovider "github.com/hlfshell/cowork/internal/git/providers"
	"github.com/hlfshell/cowork/internal/secure_store"
)

// Manager handles authentication for all Git providers and git operations
type Manager struct {
	configManager   *config.Manager
	globalAuthStore *secure_store.SecureStore
	localAuthStore  *secure_store.SecureStore
}

// AuthConfig represents authentication configuration for a provider
type AuthConfig struct {
	ProviderType git.ProviderType `json:"provider_type"`
	AuthMethod   AuthMethod       `json:"auth_method"`
	Token        string           `json:"token,omitempty"`
	Username     string           `json:"username,omitempty"`
	Password     string           `json:"password,omitempty"`
	BaseURL      string           `json:"base_url,omitempty"`
	ExpiresAt    *time.Time       `json:"expires_at,omitempty"`
}

// AuthMethod represents the authentication method
type AuthMethod string

const (
	AuthMethodToken AuthMethod = "token"
	AuthMethodBasic AuthMethod = "basic"
	AuthMethodSSH   AuthMethod = "ssh"
)

// AuthScope represents the authentication scope (global or project)
type AuthScope string

const (
	AuthScopeGlobal  AuthScope = "global"
	AuthScopeProject AuthScope = "project"
)

// NewManager creates a new authentication manager
func NewManager(configManager *config.Manager) (*Manager, error) {
	globalAuthStore, err := secure_store.NewSecureStore("auth", configManager.GlobalConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create secure store: %w", err)
	}
	localAuthStore, err := secure_store.NewSecureStore("auth", configManager.ProjectConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create secure store: %w", err)
	}

	return &Manager{
		configManager:   configManager,
		globalAuthStore: globalAuthStore,
		localAuthStore:  localAuthStore,
	}, nil
}

func (m *Manager) getAuthStore(scope AuthScope) *secure_store.SecureStore {
	if scope == AuthScopeGlobal {
		return m.globalAuthStore
	}
	return m.localAuthStore
}

// AuthenticateProvider authenticates with a specific provider using the specified method
func (m *Manager) AuthenticateProvider(ctx context.Context, providerType git.ProviderType, method AuthMethod, scope AuthScope) (*AuthConfig, error) {
	switch method {
	case AuthMethodToken:
		return m.authenticateToken(ctx, providerType, scope)
	case AuthMethodBasic:
		return m.authenticateBasic(ctx, providerType, scope)
	default:
		return nil, fmt.Errorf("unsupported authentication method: %s", method)
	}
}

// authenticateToken performs token authentication for the specified provider
func (m *Manager) authenticateToken(ctx context.Context, providerType git.ProviderType, scope AuthScope) (*AuthConfig, error) {
	return nil, fmt.Errorf("token authentication requires manual token input")
}

// authenticateBasic performs basic authentication for the specified provider
func (m *Manager) authenticateBasic(ctx context.Context, providerType git.ProviderType, scope AuthScope) (*AuthConfig, error) {
	return nil, fmt.Errorf("basic authentication requires manual credentials input")
}

// SetToken sets a token for the specified provider
func (m *Manager) SetToken(providerType git.ProviderType, token string, scope AuthScope) error {
	authConfig := &AuthConfig{
		ProviderType: providerType,
		AuthMethod:   AuthMethodToken,
		Token:        token,
	}

	return m.saveAuthConfig(authConfig, scope)
}

// SetBasicAuth sets basic authentication credentials for the specified provider
func (m *Manager) SetBasicAuth(providerType git.ProviderType, username, password string, scope AuthScope) error {
	authConfig := &AuthConfig{
		ProviderType: providerType,
		AuthMethod:   AuthMethodBasic,
		Username:     username,
		Password:     password,
	}

	return m.saveAuthConfig(authConfig, scope)
}

// GetAuthConfig retrieves authentication configuration for the specified provider
func (m *Manager) GetAuthConfig(providerType git.ProviderType, scope AuthScope) (*AuthConfig, error) {
	key := m.getAuthKey(providerType, scope)
	var authConfig AuthConfig
	var err error
	switch scope {
	case AuthScopeGlobal:
		err = m.globalAuthStore.Get(key, &authConfig)
	case AuthScopeProject:
		err = m.localAuthStore.Get(key, &authConfig)
	}
	if err != nil {
		return nil, err
	}
	return &authConfig, nil
}

// RemoveAuth removes authentication configuration for the specified provider
func (m *Manager) RemoveAuth(providerType git.ProviderType, scope AuthScope) error {
	key := m.getAuthKey(providerType, scope)
	switch scope {
	case AuthScopeGlobal:
		return m.globalAuthStore.Delete(key)
	case AuthScopeProject:
		return m.localAuthStore.Delete(key)
	}
	return nil
}

// ListAuthConfigs lists all authentication configurations
func (m *Manager) ListAuthConfigs() ([]*AuthConfig, error) {
	// Try to get configs for all known provider types and scopes
	var result []*AuthConfig

	providers := []git.ProviderType{git.ProviderGitHub, git.ProviderGitLab, git.ProviderBitbucket}
	scopes := []AuthScope{AuthScopeGlobal, AuthScopeProject}

	for _, provider := range providers {
		for _, scope := range scopes {
			config, err := m.GetAuthConfig(provider, scope)
			if err == nil {
				result = append(result, config)
			}
			// Ignore errors - configs that don't exist will return errors
		}
	}

	return result, nil
}

// TestAuth tests authentication with the specified provider
func (m *Manager) TestAuth(ctx context.Context, providerType git.ProviderType, scope AuthScope) error {
	authConfig, err := m.GetAuthConfig(providerType, scope)
	if err != nil {
		return fmt.Errorf("no authentication configured")
	}

	provider, err := m.createProvider(providerType, authConfig)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	return provider.TestAuth(ctx)
}

// createProvider creates a provider instance with the given authentication
func (m *Manager) createProvider(providerType git.ProviderType, authConfig *AuthConfig) (git.GitProvider, error) {
	switch providerType {
	case git.ProviderGitHub:
		return gitprovider.NewGitHubProvider(authConfig.Token, authConfig.BaseURL)
	case git.ProviderGitLab:
		return gitprovider.NewGitLabProvider(authConfig.Token, authConfig.BaseURL)
	case git.ProviderBitbucket:
		return gitprovider.NewBitbucketProvider(authConfig.Token, authConfig.BaseURL)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

// saveGitAuthConfig saves git authentication configuration
func (m *Manager) saveGitAuthConfig(gitAuthConfig *GitAuthConfig, scope AuthScope) error {
	key := m.getGitAuthKey(scope)
	switch scope {
	case AuthScopeGlobal:
		return m.globalAuthStore.Set(key, gitAuthConfig)
	case AuthScopeProject:
		return m.localAuthStore.Set(key, gitAuthConfig)
	}
	return nil
}

// getGitAuthKey generates a key for storing git authentication configuration
func (m *Manager) getGitAuthKey(scope AuthScope) string {
	return fmt.Sprintf("git_%s", string(scope))
}

// generateState generates a random state string for CSRF protection
func (m *Manager) generateState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// saveAuthConfig saves authentication configuration
func (m *Manager) saveAuthConfig(authConfig *AuthConfig, scope AuthScope) error {
	key := m.getAuthKey(authConfig.ProviderType, scope)
	switch scope {
	case AuthScopeGlobal:
		return m.globalAuthStore.Set(key, authConfig)
	case AuthScopeProject:
		return m.localAuthStore.Set(key, authConfig)
	}
	return nil
}

// getAuthKey generates a key for storing authentication configuration
func (m *Manager) getAuthKey(providerType git.ProviderType, scope AuthScope) string {
	return fmt.Sprintf("%s_%s", string(providerType), string(scope))
}
