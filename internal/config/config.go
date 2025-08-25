package config

import (
	"fmt"
	"os"
	"path/filepath"

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
}

// WorkspaceConfig contains workspace-related configuration
type WorkspaceConfig struct {
	// Default isolation level for new workspaces
	DefaultIsolationLevel string `yaml:"default_isolation_level" default:"full-clone"`

	// Base directory for workspace storage (relative to project root)
	BaseDirectory string `yaml:"base_directory" default:".cw/workspaces"`

	// Maximum number of workspaces to keep per project
	MaxWorkspaces int `yaml:"max_workspaces" default:"10"`

	// Auto-cleanup orphaned workspaces on startup
	AutoCleanupOrphaned bool `yaml:"auto_cleanup_orphaned" default:"true"`

	// Default branch to use when creating workspaces
	DefaultBranch string `yaml:"default_branch" default:"main"`

	// Workspace naming pattern
	NamingPattern string `yaml:"naming_pattern" default:"task/{task_name}"`

	// Auto-save workspace metadata
	AutoSaveMetadata bool `yaml:"auto_save_metadata" default:"true"`

	// Workspace timeout in minutes (0 = no timeout)
	TimeoutMinutes int `yaml:"timeout_minutes" default:"0"`
}

// GitConfig contains Git-related configuration
type GitConfig struct {
	// Git timeout in seconds
	TimeoutSeconds int `yaml:"timeout_seconds" default:"300"`

	// Default remote name
	DefaultRemote string `yaml:"default_remote" default:"origin"`

	// Auto-fetch before cloning
	AutoFetch bool `yaml:"auto_fetch" default:"true"`

	// Shallow clone depth (0 = full clone)
	ShallowDepth int `yaml:"shallow_depth" default:"0"`

	// Git user configuration
	User GitUserConfig `yaml:"user"`

	// Git credential helper
	CredentialHelper string `yaml:"credential_helper" default:""`
}

// GitUserConfig contains Git user configuration
type GitUserConfig struct {
	Name  string `yaml:"name" default:""`
	Email string `yaml:"email" default:""`
}

// AgentConfig contains AI agent configuration
type AgentConfig struct {
	// Default agent type
	DefaultAgent string `yaml:"default_agent" default:"cursor"`

	// Agent timeout in minutes
	TimeoutMinutes int `yaml:"timeout_minutes" default:"30"`

	// Maximum concurrent agents per workspace
	MaxConcurrent int `yaml:"max_concurrent" default:"1"`

	// Agent-specific configurations
	Cursor AgentTypeConfig `yaml:"cursor"`
	Claude AgentTypeConfig `yaml:"claude"`
	Gemini AgentTypeConfig `yaml:"gemini"`
	Custom AgentTypeConfig `yaml:"custom"`
}

// AgentTypeConfig contains configuration for a specific agent type
type AgentTypeConfig struct {
	// Whether this agent type is enabled
	Enabled bool `yaml:"enabled" default:"true"`

	// Command to run the agent
	Command string `yaml:"command" default:""`

	// Arguments to pass to the agent
	Args []string `yaml:"args" default:"[]"`

	// Environment variables
	Env map[string]string `yaml:"env" default:"{}"`

	// Working directory
	WorkingDir string `yaml:"working_dir" default:""`

	// Timeout in minutes (0 = use global timeout)
	TimeoutMinutes int `yaml:"timeout_minutes" default:"0"`
}

// ContainerConfig contains container-related configuration
type ContainerConfig struct {
	// Container engine to use (docker, podman)
	Engine string `yaml:"engine" default:"docker"`

	// Default container image
	DefaultImage string `yaml:"default_image" default:"golang:latest"`

	// Container timeout in minutes
	TimeoutMinutes int `yaml:"timeout_minutes" default:"60"`

	// Auto-start containers for workspaces
	AutoStart bool `yaml:"auto_start" default:"false"`

	// Container resource limits
	Resources ContainerResources `yaml:"resources"`

	// Container networking
	Network ContainerNetwork `yaml:"network"`
}

// ContainerResources contains container resource limits
type ContainerResources struct {
	// Memory limit in MB
	MemoryMB int `yaml:"memory_mb" default:"2048"`

	// CPU limit (0 = no limit)
	CPULimit float64 `yaml:"cpu_limit" default:"0"`

	// Disk space limit in GB
	DiskGB int `yaml:"disk_gb" default:"10"`
}

// ContainerNetwork contains container networking configuration
type ContainerNetwork struct {
	// Network mode (bridge, host, none)
	Mode string `yaml:"mode" default:"bridge"`

	// Port mappings
	Ports []string `yaml:"ports" default:"[]"`

	// Extra hosts
	ExtraHosts []string `yaml:"extra_hosts" default:"[]"`
}

// UIConfig contains UI/CLI configuration
type UIConfig struct {
	// Output format (text, json, yaml)
	OutputFormat string `yaml:"output_format" default:"text"`

	// Color output (auto, always, never)
	Color string `yaml:"color" default:"auto"`

	// Verbose output
	Verbose bool `yaml:"verbose" default:"false"`

	// Show progress bars
	ShowProgress bool `yaml:"show_progress" default:"true"`

	// Interactive mode
	Interactive bool `yaml:"interactive" default:"true"`

	// Confirmation prompts
	ConfirmPrompts bool `yaml:"confirm_prompts" default:"true"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	// Log level (debug, info, warn, error)
	Level string `yaml:"level" default:"info"`

	// Log format (text, json)
	Format string `yaml:"format" default:"text"`

	// Log file path (empty = stdout)
	File string `yaml:"file" default:""`

	// Include timestamps
	IncludeTimestamp bool `yaml:"include_timestamp" default:"true"`

	// Include caller information
	IncludeCaller bool `yaml:"include_caller" default:"false"`
}

// AuthConfig contains authentication configuration for various services
type AuthConfig struct {
	// Git authentication settings
	Git GitAuthConfig `yaml:"git"`

	// Container registry authentication settings
	Container ContainerAuthConfig `yaml:"container"`

	// Future authentication settings can be added here
	// Cloud CloudAuthConfig `yaml:"cloud"`
	// AI AIAuthConfig `yaml:"ai"`
}

// GitAuthConfig contains Git authentication configuration
type GitAuthConfig struct {
	// Git user configuration
	User GitUserConfig `yaml:"user"`

	// SSH key configuration
	SSH SSHConfig `yaml:"ssh"`

	// HTTPS authentication
	HTTPS HTTPSConfig `yaml:"https"`

	// Git credential helper
	CredentialHelper string `yaml:"credential_helper" default:"cache"`

	// Default authentication method (ssh, https, token)
	DefaultMethod string `yaml:"default_method" default:"ssh"`
}

// SSHConfig contains SSH authentication configuration
type SSHConfig struct {
	// SSH key path
	KeyPath string `yaml:"key_path" default:"~/.ssh/id_rsa"`

	// SSH key passphrase (encrypted)
	Passphrase string `yaml:"passphrase" default:""`

	// SSH agent socket
	AgentSocket string `yaml:"agent_socket" default:""`

	// Use SSH agent
	UseAgent bool `yaml:"use_agent" default:"true"`

	// SSH known hosts file
	KnownHostsFile string `yaml:"known_hosts_file" default:"~/.ssh/known_hosts"`

	// Strict host key checking
	StrictHostKeyChecking bool `yaml:"strict_host_key_checking" default:"true"`
}

// HTTPSConfig contains HTTPS authentication configuration
type HTTPSConfig struct {
	// Username for HTTPS authentication
	Username string `yaml:"username" default:""`

	// Personal access token (encrypted)
	Token string `yaml:"token" default:""`

	// Token type (github, gitlab, generic)
	TokenType string `yaml:"token_type" default:"github"`

	// Store credentials in credential helper
	StoreCredentials bool `yaml:"store_credentials" default:"true"`

	// Credential helper timeout in seconds
	HelperTimeout int `yaml:"helper_timeout" default:"900"`
}

// ContainerAuthConfig contains container registry authentication configuration
type ContainerAuthConfig struct {
	// Default registry
	DefaultRegistry string `yaml:"default_registry" default:"docker.io"`

	// Registry configurations
	Registries map[string]RegistryConfig `yaml:"registries" default:"{}"`

	// Use credential helper
	UseCredentialHelper bool `yaml:"use_credential_helper" default:"true"`

	// Credential helper timeout in seconds
	HelperTimeout int `yaml:"helper_timeout" default:"900"`
}

// RegistryConfig contains configuration for a specific container registry
type RegistryConfig struct {
	// Registry URL
	URL string `yaml:"url"`

	// Username for authentication
	Username string `yaml:"username" default:""`

	// Password/token (encrypted)
	Password string `yaml:"password" default:""`

	// Authentication method (basic, token, oauth)
	AuthMethod string `yaml:"auth_method" default:"basic"`

	// Skip TLS verification
	Insecure bool `yaml:"insecure" default:"false"`

	// Registry namespace/organization
	Namespace string `yaml:"namespace" default:""`

	// Registry API version
	APIVersion string `yaml:"api_version" default:"v2"`
}

// Manager handles configuration loading and merging
type Manager struct {
	GlobalConfigPath  string
	ProjectConfigPath string
	config            *Config
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "~"
	}

	return &Manager{
		GlobalConfigPath:  filepath.Join(homeDir, ".config", ".cwconfig"),
		ProjectConfigPath: ".cwconfig",
	}
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

	m.config = config
	return config, nil
}

// SaveGlobal saves the configuration to the global config file
func (m *Manager) SaveGlobal(config *Config) error {
	// Ensure the directory exists
	configDir := filepath.Dir(m.GlobalConfigPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return m.SaveConfig(config, m.GlobalConfigPath)
}

// SaveProject saves the configuration to the project config file
func (m *Manager) SaveProject(config *Config) error {
	return m.SaveConfig(config, m.ProjectConfigPath)
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// loadGlobalConfig loads configuration from the global config file
func (m *Manager) loadGlobalConfig(config *Config) error {
	if _, err := os.Stat(m.GlobalConfigPath); os.IsNotExist(err) {
		// Global config doesn't exist, create it with defaults
		return m.SaveGlobal(config)
	}

	data, err := os.ReadFile(m.GlobalConfigPath)
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
	if _, err := os.Stat(m.ProjectConfigPath); os.IsNotExist(err) {
		// Project config doesn't exist, that's fine
		return nil
	}

	data, err := os.ReadFile(m.ProjectConfigPath)
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
					KeyPath:                "~/.ssh/id_rsa",
					Passphrase:             "",
					AgentSocket:            "",
					UseAgent:               true,
					KnownHostsFile:         "~/.ssh/known_hosts",
					StrictHostKeyChecking:  true,
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
				DefaultRegistry:      "docker.io",
				Registries:          map[string]RegistryConfig{},
				UseCredentialHelper: true,
				HelperTimeout:       900,
			},
		},
	}
}
