package auth

import (
	"fmt"
	"os"
	"time"
)

// GitAuthConfig represents authentication configuration for git operations
type GitAuthConfig struct {
	AuthMethod GitAuthMethod `json:"auth_method"`
	Username   string        `json:"username,omitempty"`
	Password   string        `json:"password,omitempty"`
	Token      string        `json:"token,omitempty"`
	SSHKeyPath string        `json:"ssh_key_path,omitempty"`
	SSHKey     string        `json:"ssh_key,omitempty"`
	ExpiresAt  *time.Time    `json:"expires_at,omitempty"`
}

// GitAuthMethod represents the git authentication method
type GitAuthMethod string

const (
	GitAuthMethodSSH   GitAuthMethod = "ssh"
	GitAuthMethodHTTPS GitAuthMethod = "https"
	GitAuthMethodToken GitAuthMethod = "token"
	GitAuthMethodNone  GitAuthMethod = "none"
)

// Git Authentication Methods

// SetGitHTTPSAuth sets HTTPS authentication for git operations
func (m *Manager) SetGitHTTPSAuth(username, password string, scope AuthScope) error {
	gitAuthConfig := &GitAuthConfig{
		AuthMethod: GitAuthMethodHTTPS,
		Username:   username,
		Password:   password,
	}

	return m.saveGitAuthConfig(gitAuthConfig, scope)
}

// SetGitTokenAuth sets token authentication for git operations
func (m *Manager) SetGitTokenAuth(token string, scope AuthScope) error {
	gitAuthConfig := &GitAuthConfig{
		AuthMethod: GitAuthMethodToken,
		Token:      token,
	}

	return m.saveGitAuthConfig(gitAuthConfig, scope)
}

// SetGitSSHAuth sets SSH authentication for git operations
func (m *Manager) SetGitSSHKeyFileAuth(sshKeyPath string, scope AuthScope) error {
	gitAuthConfig := &GitAuthConfig{
		AuthMethod: GitAuthMethodSSH,
		SSHKeyPath: sshKeyPath,
	}

	return m.saveGitAuthConfig(gitAuthConfig, scope)
}

func (m *Manager) SetGitSSHKeyAuth(sshKey string, scope AuthScope) error {
	gitAuthConfig := &GitAuthConfig{
		AuthMethod: GitAuthMethodSSH,
		SSHKey:     sshKey,
	}

	return m.saveGitAuthConfig(gitAuthConfig, scope)
}

func (m *Manager) SetGitBasicAuth(username, password string, scope AuthScope) error {
	gitAuthConfig := &GitAuthConfig{
		AuthMethod: GitAuthMethodHTTPS,
		Username:   username,
		Password:   password,
	}

	return m.saveGitAuthConfig(gitAuthConfig, scope)
}

// GetGitAuthConfig retrieves git authentication configuration
func (m *Manager) GetGitAuthConfig(scope AuthScope) (*GitAuthConfig, error) {
	key := m.getGitAuthKey(scope)
	var gitAuthConfig GitAuthConfig
	var err error
	switch scope {
	case AuthScopeGlobal:
		err = m.globalAuthStore.Get(key, &gitAuthConfig)
	case AuthScopeProject:
		err = m.localAuthStore.Get(key, &gitAuthConfig)
	}
	if err != nil {
		return nil, err
	}
	return &gitAuthConfig, nil
}

// RemoveGitAuth removes git authentication configuration
func (m *Manager) RemoveGitAuth(scope AuthScope) error {
	key := m.getGitAuthKey(scope)
	switch scope {
	case AuthScopeGlobal:
		return m.globalAuthStore.Delete(key)
	case AuthScopeProject:
		return m.localAuthStore.Delete(key)
	}
	return nil
}

// GetGitSSHKey gets the SSH key path from config or stored git auth
func (m *Manager) GetGitSSHKey(scope AuthScope) (string, error) {
	// First try to get from stored git auth config
	gitAuthConfig, err := m.GetGitAuthConfig(scope)
	if err != nil {
		return "", err
	}

	if gitAuthConfig.AuthMethod != GitAuthMethodSSH {
		return "", fmt.Errorf("git auth method is not SSH")
	}

	if gitAuthConfig.SSHKeyPath != "" {
		// Read the ssh key file
		sshKey, err := os.ReadFile(gitAuthConfig.SSHKeyPath)
		if err != nil {
			return "", err
		}
		return string(sshKey), nil
	} else if gitAuthConfig.SSHKey != "" {
		return gitAuthConfig.SSHKey, nil
	} else {
		return "", fmt.Errorf("no SSH key found")
	}
}

// GetGitHTTPSAuth gets HTTPS authentication credentials
func (m *Manager) GetGitHTTPSAuth(scope AuthScope) (username, password string, err error) {
	gitAuthConfig, err := m.GetGitAuthConfig(scope)
	if err != nil {
		return "", "", err
	}

	if gitAuthConfig.AuthMethod != GitAuthMethodHTTPS {
		return "", "", fmt.Errorf("git auth method is not HTTPS")
	}

	return gitAuthConfig.Username, gitAuthConfig.Password, nil
}

// GetGitTokenAuth gets token authentication
func (m *Manager) GetGitTokenAuth(scope AuthScope) (string, error) {
	gitAuthConfig, err := m.GetGitAuthConfig(scope)
	if err != nil {
		return "", err
	}

	if gitAuthConfig.AuthMethod != GitAuthMethodToken {
		return "", fmt.Errorf("git auth method is not token")
	}

	return gitAuthConfig.Token, nil
}
