package config

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
