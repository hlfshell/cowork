package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hlfshell/cowork/internal/secure_store"
	"gopkg.in/yaml.v3"
)

// Config represents the configuration structure for cowork
type Config struct {
	// Workspace settings
	Workspace WorkspaceConfig `yaml:"workspace"`

	// Git settings
	Git GitConfig `yaml:"git"`

	// Agent settings
	Agent AgentConfig `yaml:"agent"`

	// Container settings
	Container ContainerConfig `yaml:"container"`

	// UI/CLI settings
	UI UIConfig `yaml:"ui"`

	// Logging settings
	Logging LoggingConfig `yaml:"logging"`

	// Authentication settings
	Auth AuthConfig `yaml:"auth"`

	// Environment variables (encrypted)
	envStore *secure_store.SecureStore
	Env      map[string]string `yaml:"env" default:"{}"`
}

// Manager handles configuration loading and merging
type Manager struct {
	GlobalConfigPath  string
	ProjectConfigPath string
	config            *Config
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	globalConfigPath := filepath.Join(homeDir, ".config", ".cowork")
	err = os.MkdirAll(globalConfigPath, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create global config directory: %w", err)
	}

	localDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}
	projectConfigPath := filepath.Join(localDir, ".cowork")
	err = os.MkdirAll(projectConfigPath, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create project config directory: %w", err)
	}

	return &Manager{
		GlobalConfigPath:  globalConfigPath,
		ProjectConfigPath: projectConfigPath,
	}, nil
}

// Load loads and merges configuration from global and project files
func (m *Manager) Load() (*Config, error) {
	// Start with default configuration
	config := m.GetDefaultConfig()

	// Load global configuration
	if err := m.loadGlobalConfig(config); err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	// Load project configuration (if exists)
	if err := m.loadProjectConfig(config); err != nil {
		return nil, fmt.Errorf("failed to load project config: %w", err)
	}

	// Load environment variables
	if config.envStore == nil {
		var err error
		config.envStore, err = secure_store.NewSecureStore(".env", m.ProjectConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create environment store: %w", err)
		}
	}

	// Load env vars directly from the config's store, not manager state
	if env_vars, err := m.getEnvVarsFromStore(config.envStore); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	} else {
		config.Env = env_vars
	}

	// Only set manager's config after everything is loaded successfully
	m.config = config
	return config, nil
}

// getEnvVarsFromStore loads environment variables from a specific store
// This is used during Load() to avoid dependency on manager state
func (m *Manager) getEnvVarsFromStore(store *secure_store.SecureStore) (map[string]string, error) {
	keys, err := store.List("")
	if err != nil {
		return nil, err
	}

	envVars := make(map[string]string)
	for _, key := range keys {
		var value string
		err := store.Get(key, &value)
		if err != nil {
			return nil, err
		}
		envVars[key] = value
	}
	return envVars, nil
}

// SaveGlobal saves the configuration to the global config file
func (m *Manager) SaveGlobal(config *Config) error {
	// Ensure the directory exists
	if err := os.MkdirAll(m.GlobalConfigPath, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFilePath := filepath.Join(m.GlobalConfigPath, "config.yaml")
	return m.SaveConfig(config, configFilePath)
}

// SaveProject saves the configuration to the project config file
func (m *Manager) SaveProject(config *Config) error {
	configFilePath := filepath.Join(m.ProjectConfigPath, "config.yaml")
	return m.SaveConfig(config, configFilePath)
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// loadGlobalConfig loads configuration from the global config file
func (m *Manager) loadGlobalConfig(config *Config) error {
	configFilePath := filepath.Join(m.GlobalConfigPath, "config.yaml")
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		// Global config doesn't exist, create it with defaults
		return m.SaveGlobal(config)
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to read global config file: %w", err)
	}

	var globalConfig Config
	if err := yaml.Unmarshal(data, &globalConfig); err != nil {
		return fmt.Errorf("failed to parse global config file: %w", err)
	}

	// Merge global config into default config
	m.mergeConfig(config, &globalConfig)
	return nil
}

// loadProjectConfig loads configuration from the project config file
func (m *Manager) loadProjectConfig(config *Config) error {
	configFilePath := filepath.Join(m.ProjectConfigPath, "config.yaml")
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		// Project config doesn't exist, that's fine
		return nil
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to read project config file: %w", err)
	}

	var projectConfig Config
	if err := yaml.Unmarshal(data, &projectConfig); err != nil {
		return fmt.Errorf("failed to parse project config file: %w", err)
	}

	// Merge project config into current config (project overrides global)
	m.mergeConfig(config, &projectConfig)
	return nil
}

// SaveConfig saves configuration to a file
func (m *Manager) SaveConfig(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// mergeConfig merges source config into target config
func (m *Manager) mergeConfig(target, source *Config) {
	// This is a simple merge - in a real implementation, you might want
	// more sophisticated merging logic for nested structures

	if source.Workspace.DefaultIsolationLevel != "" {
		target.Workspace.DefaultIsolationLevel = source.Workspace.DefaultIsolationLevel
	}
	if source.Workspace.BaseDirectory != "" {
		target.Workspace.BaseDirectory = source.Workspace.BaseDirectory
	}
	if source.Workspace.MaxWorkspaces != 0 {
		target.Workspace.MaxWorkspaces = source.Workspace.MaxWorkspaces
	}
	target.Workspace.AutoCleanupOrphaned = source.Workspace.AutoCleanupOrphaned
	if source.Workspace.DefaultBranch != "" {
		target.Workspace.DefaultBranch = source.Workspace.DefaultBranch
	}
	if source.Workspace.NamingPattern != "" {
		target.Workspace.NamingPattern = source.Workspace.NamingPattern
	}
	target.Workspace.AutoSaveMetadata = source.Workspace.AutoSaveMetadata
	if source.Workspace.TimeoutMinutes != 0 {
		target.Workspace.TimeoutMinutes = source.Workspace.TimeoutMinutes
	}

	// Git config
	if source.Git.TimeoutSeconds != 0 {
		target.Git.TimeoutSeconds = source.Git.TimeoutSeconds
	}
	if source.Git.DefaultRemote != "" {
		target.Git.DefaultRemote = source.Git.DefaultRemote
	}
	target.Git.AutoFetch = source.Git.AutoFetch
	if source.Git.ShallowDepth != 0 {
		target.Git.ShallowDepth = source.Git.ShallowDepth
	}
	if source.Git.User.Name != "" {
		target.Git.User.Name = source.Git.User.Name
	}
	if source.Git.User.Email != "" {
		target.Git.User.Email = source.Git.User.Email
	}
	if source.Git.CredentialHelper != "" {
		target.Git.CredentialHelper = source.Git.CredentialHelper
	}

	// Agent config
	if source.Agent.DefaultAgent != "" {
		target.Agent.DefaultAgent = source.Agent.DefaultAgent
	}
	if source.Agent.TimeoutMinutes != 0 {
		target.Agent.TimeoutMinutes = source.Agent.TimeoutMinutes
	}
	if source.Agent.MaxConcurrent != 0 {
		target.Agent.MaxConcurrent = source.Agent.MaxConcurrent
	}

	// Container config
	if source.Container.Engine != "" {
		target.Container.Engine = source.Container.Engine
	}
	if source.Container.DefaultImage != "" {
		target.Container.DefaultImage = source.Container.DefaultImage
	}
	if source.Container.TimeoutMinutes != 0 {
		target.Container.TimeoutMinutes = source.Container.TimeoutMinutes
	}
	target.Container.AutoStart = source.Container.AutoStart

	// UI config
	if source.UI.OutputFormat != "" {
		target.UI.OutputFormat = source.UI.OutputFormat
	}
	if source.UI.Color != "" {
		target.UI.Color = source.UI.Color
	}
	target.UI.Verbose = source.UI.Verbose
	target.UI.ShowProgress = source.UI.ShowProgress
	target.UI.Interactive = source.UI.Interactive
	target.UI.ConfirmPrompts = source.UI.ConfirmPrompts

	// Logging config
	if source.Logging.Level != "" {
		target.Logging.Level = source.Logging.Level
	}
	if source.Logging.Format != "" {
		target.Logging.Format = source.Logging.Format
	}
	if source.Logging.File != "" {
		target.Logging.File = source.Logging.File
	}
	target.Logging.IncludeTimestamp = source.Logging.IncludeTimestamp
	target.Logging.IncludeCaller = source.Logging.IncludeCaller

	// Auth config
	if source.Auth.Git.User.Name != "" {
		target.Auth.Git.User.Name = source.Auth.Git.User.Name
	}
	if source.Auth.Git.User.Email != "" {
		target.Auth.Git.User.Email = source.Auth.Git.User.Email
	}
	if source.Auth.Git.SSH.KeyPath != "" {
		target.Auth.Git.SSH.KeyPath = source.Auth.Git.SSH.KeyPath
	}
	if source.Auth.Git.SSH.Passphrase != "" {
		target.Auth.Git.SSH.Passphrase = source.Auth.Git.SSH.Passphrase
	}
	if source.Auth.Git.SSH.AgentSocket != "" {
		target.Auth.Git.SSH.AgentSocket = source.Auth.Git.SSH.AgentSocket
	}
	target.Auth.Git.SSH.UseAgent = source.Auth.Git.SSH.UseAgent
	if source.Auth.Git.SSH.KnownHostsFile != "" {
		target.Auth.Git.SSH.KnownHostsFile = source.Auth.Git.SSH.KnownHostsFile
	}
	target.Auth.Git.SSH.StrictHostKeyChecking = source.Auth.Git.SSH.StrictHostKeyChecking
	if source.Auth.Git.HTTPS.Username != "" {
		target.Auth.Git.HTTPS.Username = source.Auth.Git.HTTPS.Username
	}
	if source.Auth.Git.HTTPS.Token != "" {
		target.Auth.Git.HTTPS.Token = source.Auth.Git.HTTPS.Token
	}
	if source.Auth.Git.HTTPS.TokenType != "" {
		target.Auth.Git.HTTPS.TokenType = source.Auth.Git.HTTPS.TokenType
	}
	target.Auth.Git.HTTPS.StoreCredentials = source.Auth.Git.HTTPS.StoreCredentials
	if source.Auth.Git.HTTPS.HelperTimeout != 0 {
		target.Auth.Git.HTTPS.HelperTimeout = source.Auth.Git.HTTPS.HelperTimeout
	}
	if source.Auth.Git.CredentialHelper != "" {
		target.Auth.Git.CredentialHelper = source.Auth.Git.CredentialHelper
	}
	if source.Auth.Git.DefaultMethod != "" {
		target.Auth.Git.DefaultMethod = source.Auth.Git.DefaultMethod
	}
	if source.Auth.Container.DefaultRegistry != "" {
		target.Auth.Container.DefaultRegistry = source.Auth.Container.DefaultRegistry
	}
	if len(source.Auth.Container.Registries) > 0 {
		if target.Auth.Container.Registries == nil {
			target.Auth.Container.Registries = make(map[string]RegistryConfig)
		}
		for name, registry := range source.Auth.Container.Registries {
			target.Auth.Container.Registries[name] = registry
		}
	}
	target.Auth.Container.UseCredentialHelper = source.Auth.Container.UseCredentialHelper
	if source.Auth.Container.HelperTimeout != 0 {
		target.Auth.Container.HelperTimeout = source.Auth.Container.HelperTimeout
	}
}

// GetDefaultConfig returns the default configuration
func (m *Manager) GetDefaultConfig() *Config {
	return &Config{
		Workspace: WorkspaceConfig{
			DefaultIsolationLevel: "full-clone",
			BaseDirectory:         ".cw/workspaces",
			MaxWorkspaces:         10,
			AutoCleanupOrphaned:   true,
			DefaultBranch:         "main",
			NamingPattern:         "task/{task_name}",
			AutoSaveMetadata:      true,
			TimeoutMinutes:        0,
		},
		Git: GitConfig{
			TimeoutSeconds:   300,
			DefaultRemote:    "origin",
			AutoFetch:        true,
			ShallowDepth:     0,
			CredentialHelper: "",
			User: GitUserConfig{
				Name:  "",
				Email: "",
			},
		},
		Agent: AgentConfig{
			DefaultAgent:   "cursor",
			TimeoutMinutes: 30,
			MaxConcurrent:  1,
			Cursor: AgentTypeConfig{
				Enabled:        true,
				Command:        "",
				Args:           []string{},
				Env:            map[string]string{},
				WorkingDir:     "",
				TimeoutMinutes: 0,
			},
			Claude: AgentTypeConfig{
				Enabled:        true,
				Command:        "",
				Args:           []string{},
				Env:            map[string]string{},
				WorkingDir:     "",
				TimeoutMinutes: 0,
			},
			Gemini: AgentTypeConfig{
				Enabled:        true,
				Command:        "",
				Args:           []string{},
				Env:            map[string]string{},
				WorkingDir:     "",
				TimeoutMinutes: 0,
			},
			Custom: AgentTypeConfig{
				Enabled:        false,
				Command:        "",
				Args:           []string{},
				Env:            map[string]string{},
				WorkingDir:     "",
				TimeoutMinutes: 0,
			},
		},
		Container: ContainerConfig{
			Engine:         "docker",
			DefaultImage:   "golang:latest",
			TimeoutMinutes: 60,
			AutoStart:      false,
			Resources: ContainerResources{
				MemoryMB: 2048,
				CPULimit: 0,
				DiskGB:   10,
			},
			Network: ContainerNetwork{
				Mode:       "bridge",
				Ports:      []string{},
				ExtraHosts: []string{},
			},
		},
		UI: UIConfig{
			OutputFormat:   "text",
			Color:          "auto",
			Verbose:        false,
			ShowProgress:   true,
			Interactive:    true,
			ConfirmPrompts: true,
		},
		Logging: LoggingConfig{
			Level:            "info",
			Format:           "text",
			File:             "",
			IncludeTimestamp: true,
			IncludeCaller:    false,
		},
		Auth: AuthConfig{
			Git: GitAuthConfig{
				User: GitUserConfig{
					Name:  "",
					Email: "",
				},
				SSH: SSHConfig{
					KeyPath:               "~/.ssh/id_rsa",
					Passphrase:            "",
					AgentSocket:           "",
					UseAgent:              true,
					KnownHostsFile:        "~/.ssh/known_hosts",
					StrictHostKeyChecking: true,
				},
				HTTPS: HTTPSConfig{
					Username:         "",
					Token:            "",
					TokenType:        "github",
					StoreCredentials: true,
					HelperTimeout:    900,
				},
				CredentialHelper: "cache",
				DefaultMethod:    "ssh",
			},
			Container: ContainerAuthConfig{
				DefaultRegistry:     "docker.io",
				Registries:          map[string]RegistryConfig{},
				UseCredentialHelper: true,
				HelperTimeout:       900,
			},
		},
		Env: map[string]string{},
	}
}

// HumanReadable returns a human readable string of the config
func (m *Manager) HumanReadable() string {
	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return string(data)
}

func (m *Manager) SaveToFile(filename string) error {
	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

func (m *Manager) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, m.config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return nil
}
