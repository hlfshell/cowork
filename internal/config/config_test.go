package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewManager tests creating a new configuration manager
func TestNewManager(t *testing.T) {
	// Test case: Creating a new configuration manager should succeed
	manager := NewManager()

	if manager == nil {
		t.Fatal("Expected manager to not be nil")
	}

	// Verify paths are set correctly
	if manager.GlobalConfigPath == "" {
		t.Error("Expected global config path to be set")
	}

	if manager.ProjectConfigPath != ".cwconfig" {
		t.Errorf("Expected project config path '.cwconfig', got '%s'", manager.ProjectConfigPath)
	}

	// Verify global config path contains .config directory
	if !strings.Contains(manager.GlobalConfigPath, ".config") {
		t.Errorf("Expected global config path to contain '.config', got '%s'", manager.GlobalConfigPath)
	}
}

// TestManager_Load_WithNoConfigFiles tests loading configuration when no config files exist
func TestManager_Load_WithNoConfigFiles(t *testing.T) {
	// Test case: Loading configuration when no config files exist should create default global config
	tempDir := t.TempDir()

	// Create a temporary manager with custom paths
	manager := &Manager{
		GlobalConfigPath:  filepath.Join(tempDir, "global", ".cwconfig"),
		ProjectConfigPath: filepath.Join(tempDir, "project", ".cwconfig"),
	}

	config, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	if config == nil {
		t.Fatal("Expected configuration to not be nil")
	}

	// Verify default values are set
	if config.Workspace.DefaultIsolationLevel != "full-clone" {
		t.Errorf("Expected default isolation level 'full-clone', got '%s'", config.Workspace.DefaultIsolationLevel)
	}

	if config.Workspace.BaseDirectory != ".cw/workspaces" {
		t.Errorf("Expected base directory '.cw/workspaces', got '%s'", config.Workspace.BaseDirectory)
	}

	if config.Workspace.MaxWorkspaces != 10 {
		t.Errorf("Expected max workspaces 10, got %d", config.Workspace.MaxWorkspaces)
	}

	if config.Git.TimeoutSeconds != 300 {
		t.Errorf("Expected git timeout 300 seconds, got %d", config.Git.TimeoutSeconds)
	}

	if config.Agent.DefaultAgent != "cursor" {
		t.Errorf("Expected default agent 'cursor', got '%s'", config.Agent.DefaultAgent)
	}

	// Verify global config file was created
	if _, err := os.Stat(manager.GlobalConfigPath); os.IsNotExist(err) {
		t.Error("Expected global config file to be created")
	}
}

// TestManager_Load_WithGlobalConfigOnly tests loading configuration with only global config
func TestManager_Load_WithGlobalConfigOnly(t *testing.T) {
	// Test case: Loading configuration with only global config should merge correctly
	tempDir := t.TempDir()

	// Create a temporary manager with custom paths
	manager := &Manager{
		GlobalConfigPath:  filepath.Join(tempDir, "global", ".cwconfig"),
		ProjectConfigPath: filepath.Join(tempDir, "project", ".cwconfig"),
	}

	// Create global config directory
	globalConfigDir := filepath.Dir(manager.GlobalConfigPath)
	if err := os.MkdirAll(globalConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create global config directory: %v", err)
	}

	// Create a custom global config
	globalConfig := `workspace:
  default_isolation_level: "worktree"
  max_workspaces: 5
  auto_cleanup_orphaned: false
git:
  timeout_seconds: 600
  auto_fetch: false
agent:
  default_agent: "claude"
  timeout_minutes: 45
ui:
  output_format: "json"
  verbose: true
logging:
  level: "debug"
  include_caller: true`

	if err := os.WriteFile(manager.GlobalConfigPath, []byte(globalConfig), 0644); err != nil {
		t.Fatalf("Failed to write global config file: %v", err)
	}

	config, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	if config == nil {
		t.Fatal("Expected configuration to not be nil")
	}

	// Verify global config values are applied
	if config.Workspace.DefaultIsolationLevel != "worktree" {
		t.Errorf("Expected isolation level 'worktree', got '%s'", config.Workspace.DefaultIsolationLevel)
	}

	if config.Workspace.MaxWorkspaces != 5 {
		t.Errorf("Expected max workspaces 5, got %d", config.Workspace.MaxWorkspaces)
	}

	if config.Workspace.AutoCleanupOrphaned {
		t.Error("Expected auto cleanup orphaned to be false")
	}

	if config.Git.TimeoutSeconds != 600 {
		t.Errorf("Expected git timeout 600 seconds, got %d", config.Git.TimeoutSeconds)
	}

	if config.Git.AutoFetch {
		t.Error("Expected auto fetch to be false")
	}

	if config.Agent.DefaultAgent != "claude" {
		t.Errorf("Expected default agent 'claude', got '%s'", config.Agent.DefaultAgent)
	}

	if config.Agent.TimeoutMinutes != 45 {
		t.Errorf("Expected agent timeout 45 minutes, got %d", config.Agent.TimeoutMinutes)
	}

	if config.UI.OutputFormat != "json" {
		t.Errorf("Expected output format 'json', got '%s'", config.UI.OutputFormat)
	}

	if !config.UI.Verbose {
		t.Error("Expected verbose to be true")
	}

	if config.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", config.Logging.Level)
	}

	if !config.Logging.IncludeCaller {
		t.Error("Expected include caller to be true")
	}

	// Verify default values are still set for unspecified fields
	if config.Workspace.BaseDirectory != ".cw/workspaces" {
		t.Errorf("Expected base directory '.cw/workspaces', got '%s'", config.Workspace.BaseDirectory)
	}

	if config.Container.Engine != "docker" {
		t.Errorf("Expected container engine 'docker', got '%s'", config.Container.Engine)
	}
}

// TestManager_Load_WithProjectConfig tests loading configuration with project config overriding global
func TestManager_Load_WithProjectConfig(t *testing.T) {
	// Test case: Loading configuration with project config should override global config
	tempDir := t.TempDir()

	// Create a temporary manager with custom paths
	manager := &Manager{
		GlobalConfigPath:  filepath.Join(tempDir, "global", ".cwconfig"),
		ProjectConfigPath: filepath.Join(tempDir, "project", ".cwconfig"),
	}

	// Create global config directory
	globalConfigDir := filepath.Dir(manager.GlobalConfigPath)
	if err := os.MkdirAll(globalConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create global config directory: %v", err)
	}

	// Create project config directory
	projectConfigDir := filepath.Dir(manager.ProjectConfigPath)
	if err := os.MkdirAll(projectConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create project config directory: %v", err)
	}

	// Create global config
	globalConfig := `workspace:
  default_isolation_level: "worktree"
  max_workspaces: 5
  auto_cleanup_orphaned: false
git:
  timeout_seconds: 600
  auto_fetch: false
agent:
  default_agent: "claude"
  timeout_minutes: 45`

	if err := os.WriteFile(manager.GlobalConfigPath, []byte(globalConfig), 0644); err != nil {
		t.Fatalf("Failed to write global config file: %v", err)
	}

	// Create project config that overrides some values
	projectConfig := `workspace:
  default_isolation_level: "full-clone"
  max_workspaces: 3
  auto_cleanup_orphaned: true
git:
  timeout_seconds: 300
agent:
  default_agent: "cursor"
  timeout_minutes: 30
ui:
  output_format: "yaml"
  verbose: true`

	if err := os.WriteFile(manager.ProjectConfigPath, []byte(projectConfig), 0644); err != nil {
		t.Fatalf("Failed to write project config file: %v", err)
	}

	config, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	if config == nil {
		t.Fatal("Expected configuration to not be nil")
	}

	// Verify project config values override global config
	if config.Workspace.DefaultIsolationLevel != "full-clone" {
		t.Errorf("Expected isolation level 'full-clone', got '%s'", config.Workspace.DefaultIsolationLevel)
	}

	if config.Workspace.MaxWorkspaces != 3 {
		t.Errorf("Expected max workspaces 3, got %d", config.Workspace.MaxWorkspaces)
	}

	if !config.Workspace.AutoCleanupOrphaned {
		t.Error("Expected auto cleanup orphaned to be true")
	}

	if config.Git.TimeoutSeconds != 300 {
		t.Errorf("Expected git timeout 300 seconds, got %d", config.Git.TimeoutSeconds)
	}

	if config.Agent.DefaultAgent != "cursor" {
		t.Errorf("Expected default agent 'cursor', got '%s'", config.Agent.DefaultAgent)
	}

	if config.Agent.TimeoutMinutes != 30 {
		t.Errorf("Expected agent timeout 30 minutes, got %d", config.Agent.TimeoutMinutes)
	}

	if config.UI.OutputFormat != "yaml" {
		t.Errorf("Expected output format 'yaml', got '%s'", config.UI.OutputFormat)
	}

	if !config.UI.Verbose {
		t.Error("Expected verbose to be true")
	}

	// Verify global config values that weren't overridden are still set
	if config.Git.AutoFetch {
		t.Error("Expected auto fetch to be false (from global config)")
	}
}

// TestManager_Load_WithInvalidGlobalConfig tests loading configuration with invalid global config
func TestManager_Load_WithInvalidGlobalConfig(t *testing.T) {
	// Test case: Loading configuration with invalid global config should fail
	tempDir := t.TempDir()

	// Create a temporary manager with custom paths
	manager := &Manager{
		GlobalConfigPath:  filepath.Join(tempDir, "global", ".cwconfig"),
		ProjectConfigPath: filepath.Join(tempDir, "project", ".cwconfig"),
	}

	// Create global config directory
	globalConfigDir := filepath.Dir(manager.GlobalConfigPath)
	if err := os.MkdirAll(globalConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create global config directory: %v", err)
	}

	// Create invalid YAML config
	invalidConfig := `workspace:
  default_isolation_level: "worktree"
  max_workspaces: invalid_value
git:
  timeout_seconds: 600`

	if err := os.WriteFile(manager.GlobalConfigPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to write invalid global config file: %v", err)
	}

	_, err := manager.Load()
	if err == nil {
		t.Error("Expected error when loading invalid global config, got nil")
	}

	expectedError := "failed to parse global config file"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestManager_Load_WithInvalidProjectConfig tests loading configuration with invalid project config
func TestManager_Load_WithInvalidProjectConfig(t *testing.T) {
	// Test case: Loading configuration with invalid project config should fail
	tempDir := t.TempDir()

	// Create a temporary manager with custom paths
	manager := &Manager{
		GlobalConfigPath:  filepath.Join(tempDir, "global", ".cwconfig"),
		ProjectConfigPath: filepath.Join(tempDir, "project", ".cwconfig"),
	}

	// Create global config directory
	globalConfigDir := filepath.Dir(manager.GlobalConfigPath)
	if err := os.MkdirAll(globalConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create global config directory: %v", err)
	}

	// Create project config directory
	projectConfigDir := filepath.Dir(manager.ProjectConfigPath)
	if err := os.MkdirAll(projectConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create project config directory: %v", err)
	}

	// Create valid global config
	globalConfig := `workspace:
  default_isolation_level: "worktree"
  max_workspaces: 5`

	if err := os.WriteFile(manager.GlobalConfigPath, []byte(globalConfig), 0644); err != nil {
		t.Fatalf("Failed to write global config file: %v", err)
	}

	// Create invalid project config
	invalidProjectConfig := `workspace:
  default_isolation_level: "full-clone"
  max_workspaces: invalid_value`

	if err := os.WriteFile(manager.ProjectConfigPath, []byte(invalidProjectConfig), 0644); err != nil {
		t.Fatalf("Failed to write invalid project config file: %v", err)
	}

	_, err := manager.Load()
	if err == nil {
		t.Error("Expected error when loading invalid project config, got nil")
	}

	expectedError := "failed to parse project config file"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestManager_SaveGlobal tests saving configuration to global config file
func TestManager_SaveGlobal(t *testing.T) {
	// Test case: Saving configuration to global config file should succeed
	tempDir := t.TempDir()

	// Create a temporary manager with custom paths
	manager := &Manager{
		GlobalConfigPath:  filepath.Join(tempDir, "global", ".cwconfig"),
		ProjectConfigPath: filepath.Join(tempDir, "project", ".cwconfig"),
	}

	// Create a custom configuration
	config := &Config{
		Workspace: WorkspaceConfig{
			DefaultIsolationLevel: "worktree",
			BaseDirectory:        ".custom/workspaces",
			MaxWorkspaces:        15,
			AutoCleanupOrphaned:  false,
			DefaultBranch:        "develop",
			NamingPattern:        "feature/{task_name}",
			AutoSaveMetadata:     false,
			TimeoutMinutes:       120,
		},
		Git: GitConfig{
			TimeoutSeconds:   900,
			DefaultRemote:    "upstream",
			AutoFetch:        false,
			ShallowDepth:     10,
			CredentialHelper: "cache",
			User: GitUserConfig{
				Name:  "Test User",
				Email: "test@example.com",
			},
		},
		Agent: AgentConfig{
			DefaultAgent:    "gemini",
			TimeoutMinutes:  60,
			MaxConcurrent:   3,
		},
		Container: ContainerConfig{
			Engine:          "podman",
			DefaultImage:    "node:latest",
			TimeoutMinutes:  90,
			AutoStart:       true,
		},
		UI: UIConfig{
			OutputFormat:    "json",
			Color:          "never",
			Verbose:        true,
			ShowProgress:   false,
			Interactive:    false,
			ConfirmPrompts: false,
		},
		Logging: LoggingConfig{
			Level:            "debug",
			Format:           "json",
			File:             "/tmp/cowork.log",
			IncludeTimestamp: false,
			IncludeCaller:    true,
		},
	}

	// Save the configuration
	err := manager.SaveGlobal(config)
	if err != nil {
		t.Fatalf("Failed to save global configuration: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(manager.GlobalConfigPath); os.IsNotExist(err) {
		t.Error("Expected global config file to be created")
	}

	// Load the configuration back and verify it matches
	loadedConfig, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load saved configuration: %v", err)
	}

	// Verify workspace config
	if loadedConfig.Workspace.DefaultIsolationLevel != config.Workspace.DefaultIsolationLevel {
		t.Errorf("Expected isolation level '%s', got '%s'", config.Workspace.DefaultIsolationLevel, loadedConfig.Workspace.DefaultIsolationLevel)
	}

	if loadedConfig.Workspace.BaseDirectory != config.Workspace.BaseDirectory {
		t.Errorf("Expected base directory '%s', got '%s'", config.Workspace.BaseDirectory, loadedConfig.Workspace.BaseDirectory)
	}

	if loadedConfig.Workspace.MaxWorkspaces != config.Workspace.MaxWorkspaces {
		t.Errorf("Expected max workspaces %d, got %d", config.Workspace.MaxWorkspaces, loadedConfig.Workspace.MaxWorkspaces)
	}

	if loadedConfig.Workspace.AutoCleanupOrphaned != config.Workspace.AutoCleanupOrphaned {
		t.Errorf("Expected auto cleanup orphaned %t, got %t", config.Workspace.AutoCleanupOrphaned, loadedConfig.Workspace.AutoCleanupOrphaned)
	}

	// Verify git config
	if loadedConfig.Git.TimeoutSeconds != config.Git.TimeoutSeconds {
		t.Errorf("Expected git timeout %d seconds, got %d", config.Git.TimeoutSeconds, loadedConfig.Git.TimeoutSeconds)
	}

	if loadedConfig.Git.DefaultRemote != config.Git.DefaultRemote {
		t.Errorf("Expected default remote '%s', got '%s'", config.Git.DefaultRemote, loadedConfig.Git.DefaultRemote)
	}

	if loadedConfig.Git.User.Name != config.Git.User.Name {
		t.Errorf("Expected git user name '%s', got '%s'", config.Git.User.Name, loadedConfig.Git.User.Name)
	}

	// Verify agent config
	if loadedConfig.Agent.DefaultAgent != config.Agent.DefaultAgent {
		t.Errorf("Expected default agent '%s', got '%s'", config.Agent.DefaultAgent, loadedConfig.Agent.DefaultAgent)
	}

	// Verify container config
	if loadedConfig.Container.Engine != config.Container.Engine {
		t.Errorf("Expected container engine '%s', got '%s'", config.Container.Engine, loadedConfig.Container.Engine)
	}

	// Verify UI config
	if loadedConfig.UI.OutputFormat != config.UI.OutputFormat {
		t.Errorf("Expected output format '%s', got '%s'", config.UI.OutputFormat, loadedConfig.UI.OutputFormat)
	}

	// Verify logging config
	if loadedConfig.Logging.Level != config.Logging.Level {
		t.Errorf("Expected log level '%s', got '%s'", config.Logging.Level, loadedConfig.Logging.Level)
	}
}

// TestManager_SaveProject tests saving configuration to project config file
func TestManager_SaveProject(t *testing.T) {
	// Test case: Saving configuration to project config file should succeed
	tempDir := t.TempDir()

	// Create a temporary manager with custom paths
	manager := &Manager{
		GlobalConfigPath:  filepath.Join(tempDir, "global", ".cwconfig"),
		ProjectConfigPath: filepath.Join(tempDir, "project", ".cwconfig"),
	}

	// Create project config directory
	projectConfigDir := filepath.Dir(manager.ProjectConfigPath)
	if err := os.MkdirAll(projectConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create project config directory: %v", err)
	}

	// Create a custom configuration
	config := &Config{
		Workspace: WorkspaceConfig{
			DefaultIsolationLevel: "full-clone",
			MaxWorkspaces:        5,
		},
		Git: GitConfig{
			TimeoutSeconds: 300,
		},
		Agent: AgentConfig{
			DefaultAgent: "cursor",
		},
	}

	// Save the configuration
	err := manager.SaveProject(config)
	if err != nil {
		t.Fatalf("Failed to save project configuration: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(manager.ProjectConfigPath); os.IsNotExist(err) {
		t.Error("Expected project config file to be created")
	}

	// Load the configuration back and verify it matches
	loadedConfig, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load saved configuration: %v", err)
	}

	// Verify the project config values are applied
	if loadedConfig.Workspace.DefaultIsolationLevel != config.Workspace.DefaultIsolationLevel {
		t.Errorf("Expected isolation level '%s', got '%s'", config.Workspace.DefaultIsolationLevel, loadedConfig.Workspace.DefaultIsolationLevel)
	}

	if loadedConfig.Workspace.MaxWorkspaces != config.Workspace.MaxWorkspaces {
		t.Errorf("Expected max workspaces %d, got %d", config.Workspace.MaxWorkspaces, loadedConfig.Workspace.MaxWorkspaces)
	}

	if loadedConfig.Git.TimeoutSeconds != config.Git.TimeoutSeconds {
		t.Errorf("Expected git timeout %d seconds, got %d", config.Git.TimeoutSeconds, loadedConfig.Git.TimeoutSeconds)
	}

	if loadedConfig.Agent.DefaultAgent != config.Agent.DefaultAgent {
		t.Errorf("Expected default agent '%s', got '%s'", config.Agent.DefaultAgent, loadedConfig.Agent.DefaultAgent)
	}
}

// TestManager_GetConfig tests getting the current configuration
func TestManager_GetConfig(t *testing.T) {
	// Test case: Getting configuration should return the loaded configuration
	tempDir := t.TempDir()

	// Create a temporary manager with custom paths
	manager := &Manager{
		GlobalConfigPath:  filepath.Join(tempDir, "global", ".cwconfig"),
		ProjectConfigPath: filepath.Join(tempDir, "project", ".cwconfig"),
	}

	// Initially, config should be nil
	if manager.GetConfig() != nil {
		t.Error("Expected config to be nil before loading")
	}

	// Load configuration
	config, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Get the configuration
	retrievedConfig := manager.GetConfig()
	if retrievedConfig == nil {
		t.Fatal("Expected retrieved config to not be nil")
	}

	// Verify it's the same configuration
	if retrievedConfig != config {
		t.Error("Expected retrieved config to be the same as loaded config")
	}

	// Verify some values to ensure it's the correct config
	if retrievedConfig.Workspace.DefaultIsolationLevel != "full-clone" {
		t.Errorf("Expected default isolation level 'full-clone', got '%s'", retrievedConfig.Workspace.DefaultIsolationLevel)
	}
}

// TestManager_GetDefaultConfig tests getting the default configuration
func TestManager_GetDefaultConfig(t *testing.T) {
	// Test case: Getting default configuration should return sensible defaults
	manager := NewManager()
	defaultConfig := manager.GetDefaultConfig()

	if defaultConfig == nil {
		t.Fatal("Expected default config to not be nil")
	}

	// Verify workspace defaults
	if defaultConfig.Workspace.DefaultIsolationLevel != "full-clone" {
		t.Errorf("Expected default isolation level 'full-clone', got '%s'", defaultConfig.Workspace.DefaultIsolationLevel)
	}

	if defaultConfig.Workspace.BaseDirectory != ".cw/workspaces" {
		t.Errorf("Expected default base directory '.cw/workspaces', got '%s'", defaultConfig.Workspace.BaseDirectory)
	}

	if defaultConfig.Workspace.MaxWorkspaces != 10 {
		t.Errorf("Expected default max workspaces 10, got %d", defaultConfig.Workspace.MaxWorkspaces)
	}

	if !defaultConfig.Workspace.AutoCleanupOrphaned {
		t.Error("Expected default auto cleanup orphaned to be true")
	}

	// Verify git defaults
	if defaultConfig.Git.TimeoutSeconds != 300 {
		t.Errorf("Expected default git timeout 300 seconds, got %d", defaultConfig.Git.TimeoutSeconds)
	}

	if defaultConfig.Git.DefaultRemote != "origin" {
		t.Errorf("Expected default remote 'origin', got '%s'", defaultConfig.Git.DefaultRemote)
	}

	if !defaultConfig.Git.AutoFetch {
		t.Error("Expected default auto fetch to be true")
	}

	// Verify agent defaults
	if defaultConfig.Agent.DefaultAgent != "cursor" {
		t.Errorf("Expected default agent 'cursor', got '%s'", defaultConfig.Agent.DefaultAgent)
	}

	if defaultConfig.Agent.TimeoutMinutes != 30 {
		t.Errorf("Expected default agent timeout 30 minutes, got %d", defaultConfig.Agent.TimeoutMinutes)
	}

	if defaultConfig.Agent.MaxConcurrent != 1 {
		t.Errorf("Expected default max concurrent 1, got %d", defaultConfig.Agent.MaxConcurrent)
	}

	// Verify container defaults
	if defaultConfig.Container.Engine != "docker" {
		t.Errorf("Expected default container engine 'docker', got '%s'", defaultConfig.Container.Engine)
	}

	if defaultConfig.Container.DefaultImage != "golang:latest" {
		t.Errorf("Expected default container image 'golang:latest', got '%s'", defaultConfig.Container.DefaultImage)
	}

	// Verify UI defaults
	if defaultConfig.UI.OutputFormat != "text" {
		t.Errorf("Expected default output format 'text', got '%s'", defaultConfig.UI.OutputFormat)
	}

	if defaultConfig.UI.Color != "auto" {
		t.Errorf("Expected default color 'auto', got '%s'", defaultConfig.UI.Color)
	}

	if defaultConfig.UI.Verbose {
		t.Error("Expected default verbose to be false")
	}

	if !defaultConfig.UI.ShowProgress {
		t.Error("Expected default show progress to be true")
	}

	// Verify logging defaults
	if defaultConfig.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", defaultConfig.Logging.Level)
	}

	if defaultConfig.Logging.Format != "text" {
		t.Errorf("Expected default log format 'text', got '%s'", defaultConfig.Logging.Format)
	}

	if defaultConfig.Logging.File != "" {
		t.Errorf("Expected default log file to be empty, got '%s'", defaultConfig.Logging.File)
	}

	if !defaultConfig.Logging.IncludeTimestamp {
		t.Error("Expected default include timestamp to be true")
	}

	if defaultConfig.Logging.IncludeCaller {
		t.Error("Expected default include caller to be false")
	}
}

// TestManager_GetDefaultConfig_WithAuth tests getting the default configuration with auth
func TestManager_GetDefaultConfig_WithAuth(t *testing.T) {
	// Test case: Default configuration should include auth settings
	manager := NewManager()
	defaultConfig := manager.GetDefaultConfig()

	// Test Git auth configuration
	if defaultConfig.Auth.Git.User.Name != "" {
		t.Errorf("Expected empty Git user name, got: %s", defaultConfig.Auth.Git.User.Name)
	}
	if defaultConfig.Auth.Git.User.Email != "" {
		t.Errorf("Expected empty Git user email, got: %s", defaultConfig.Auth.Git.User.Email)
	}
	if defaultConfig.Auth.Git.DefaultMethod != "ssh" {
		t.Errorf("Expected default Git method to be 'ssh', got: %s", defaultConfig.Auth.Git.DefaultMethod)
	}
	if defaultConfig.Auth.Git.CredentialHelper != "cache" {
		t.Errorf("Expected default credential helper to be 'cache', got: %s", defaultConfig.Auth.Git.CredentialHelper)
	}

	// Test SSH configuration
	if defaultConfig.Auth.Git.SSH.KeyPath != "~/.ssh/id_rsa" {
		t.Errorf("Expected default SSH key path to be '~/.ssh/id_rsa', got: %s", defaultConfig.Auth.Git.SSH.KeyPath)
	}
	if !defaultConfig.Auth.Git.SSH.UseAgent {
		t.Error("Expected default SSH use agent to be true")
	}
	if !defaultConfig.Auth.Git.SSH.StrictHostKeyChecking {
		t.Error("Expected default SSH strict host key checking to be true")
	}

	// Test HTTPS configuration
	if defaultConfig.Auth.Git.HTTPS.TokenType != "github" {
		t.Errorf("Expected default token type to be 'github', got: %s", defaultConfig.Auth.Git.HTTPS.TokenType)
	}
	if !defaultConfig.Auth.Git.HTTPS.StoreCredentials {
		t.Error("Expected default store credentials to be true")
	}
	if defaultConfig.Auth.Git.HTTPS.HelperTimeout != 900 {
		t.Errorf("Expected default helper timeout to be 900, got: %d", defaultConfig.Auth.Git.HTTPS.HelperTimeout)
	}

	// Test Container auth configuration
	if defaultConfig.Auth.Container.DefaultRegistry != "docker.io" {
		t.Errorf("Expected default registry to be 'docker.io', got: %s", defaultConfig.Auth.Container.DefaultRegistry)
	}
	if !defaultConfig.Auth.Container.UseCredentialHelper {
		t.Error("Expected default use credential helper to be true")
	}
	if defaultConfig.Auth.Container.HelperTimeout != 900 {
		t.Errorf("Expected default helper timeout to be 900, got: %d", defaultConfig.Auth.Container.HelperTimeout)
	}
	if defaultConfig.Auth.Container.Registries == nil {
		t.Error("Expected registries map to be initialized")
	}
}

// TestManager_Load_WithAuthConfig tests loading configuration with auth settings
func TestManager_Load_WithAuthConfig(t *testing.T) {
	// Test case: Loading configuration should properly merge auth settings
	manager := NewManager()

	// Create temporary directory for test
	tempDir := t.TempDir()
	manager.GlobalConfigPath = filepath.Join(tempDir, ".cwconfig")

	// Create global config with auth settings
	globalConfig := &Config{
		Auth: AuthConfig{
			Git: GitAuthConfig{
				User: GitUserConfig{
					Name:  "Test User",
					Email: "test@example.com",
				},
				SSH: SSHConfig{
					KeyPath: "~/.ssh/test_key",
					UseAgent: false,
				},
				HTTPS: HTTPSConfig{
					Username: "testuser",
					TokenType: "gitlab",
				},
				DefaultMethod: "https",
			},
			Container: ContainerAuthConfig{
				DefaultRegistry: "test.registry.com",
				Registries: map[string]RegistryConfig{
					"test-registry": {
						URL:        "https://test.registry.com",
						Username:   "testuser",
						AuthMethod: "basic",
					},
				},
			},
		},
	}

	// Save global config
	if err := manager.SaveGlobal(globalConfig); err != nil {
		t.Fatalf("Failed to save global config: %v", err)
	}

	// Load configuration
	config, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Verify Git auth settings
	if config.Auth.Git.User.Name != "Test User" {
		t.Errorf("Expected Git user name to be 'Test User', got: %s", config.Auth.Git.User.Name)
	}
	if config.Auth.Git.User.Email != "test@example.com" {
		t.Errorf("Expected Git user email to be 'test@example.com', got: %s", config.Auth.Git.User.Email)
	}
	if config.Auth.Git.SSH.KeyPath != "~/.ssh/test_key" {
		t.Errorf("Expected SSH key path to be '~/.ssh/test_key', got: %s", config.Auth.Git.SSH.KeyPath)
	}
	if config.Auth.Git.SSH.UseAgent {
		t.Error("Expected SSH use agent to be false")
	}
	if config.Auth.Git.HTTPS.Username != "testuser" {
		t.Errorf("Expected HTTPS username to be 'testuser', got: %s", config.Auth.Git.HTTPS.Username)
	}
	if config.Auth.Git.HTTPS.TokenType != "gitlab" {
		t.Errorf("Expected token type to be 'gitlab', got: %s", config.Auth.Git.HTTPS.TokenType)
	}
	if config.Auth.Git.DefaultMethod != "https" {
		t.Errorf("Expected default method to be 'https', got: %s", config.Auth.Git.DefaultMethod)
	}

	// Verify Container auth settings
	if config.Auth.Container.DefaultRegistry != "test.registry.com" {
		t.Errorf("Expected default registry to be 'test.registry.com', got: %s", config.Auth.Container.DefaultRegistry)
	}
	if len(config.Auth.Container.Registries) != 1 {
		t.Errorf("Expected 1 registry, got: %d", len(config.Auth.Container.Registries))
	}
	testRegistry, exists := config.Auth.Container.Registries["test-registry"]
	if !exists {
		t.Error("Expected test-registry to exist")
	}
	if testRegistry.URL != "https://test.registry.com" {
		t.Errorf("Expected registry URL to be 'https://test.registry.com', got: %s", testRegistry.URL)
	}
	if testRegistry.Username != "testuser" {
		t.Errorf("Expected registry username to be 'testuser', got: %s", testRegistry.Username)
	}
	if testRegistry.AuthMethod != "basic" {
		t.Errorf("Expected registry auth method to be 'basic', got: %s", testRegistry.AuthMethod)
	}
}

// TestManager_Load_WithProjectAuthOverride tests project auth configuration override
func TestManager_Load_WithProjectAuthOverride(t *testing.T) {
	// Test case: Project auth config should override global auth config
	manager := NewManager()

	// Create temporary directory for test
	tempDir := t.TempDir()
	manager.GlobalConfigPath = filepath.Join(tempDir, ".cwconfig")
	manager.ProjectConfigPath = filepath.Join(tempDir, ".cwconfig")

	// Create global config
	globalConfig := &Config{
		Auth: AuthConfig{
			Git: GitAuthConfig{
				User: GitUserConfig{
					Name:  "Global User",
					Email: "global@example.com",
				},
				DefaultMethod: "ssh",
			},
			Container: ContainerAuthConfig{
				DefaultRegistry: "global.registry.com",
			},
		},
	}

	// Create project config with overrides
	projectConfig := &Config{
		Auth: AuthConfig{
			Git: GitAuthConfig{
				User: GitUserConfig{
					Name:  "Project User",
					Email: "project@example.com",
				},
				DefaultMethod: "https",
			},
			Container: ContainerAuthConfig{
				DefaultRegistry: "project.registry.com",
				Registries: map[string]RegistryConfig{
					"project-registry": {
						URL:        "https://project.registry.com",
						Username:   "projectuser",
						AuthMethod: "token",
					},
				},
			},
		},
	}

	// Save both configs
	if err := manager.SaveGlobal(globalConfig); err != nil {
		t.Fatalf("Failed to save global config: %v", err)
	}
	if err := manager.SaveProject(projectConfig); err != nil {
		t.Fatalf("Failed to save project config: %v", err)
	}

	// Load configuration
	config, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Verify project overrides took effect
	if config.Auth.Git.User.Name != "Project User" {
		t.Errorf("Expected Git user name to be 'Project User', got: %s", config.Auth.Git.User.Name)
	}
	if config.Auth.Git.User.Email != "project@example.com" {
		t.Errorf("Expected Git user email to be 'project@example.com', got: %s", config.Auth.Git.User.Email)
	}
	if config.Auth.Git.DefaultMethod != "https" {
		t.Errorf("Expected default method to be 'https', got: %s", config.Auth.Git.DefaultMethod)
	}
	if config.Auth.Container.DefaultRegistry != "project.registry.com" {
		t.Errorf("Expected default registry to be 'project.registry.com', got: %s", config.Auth.Container.DefaultRegistry)
	}

	// Verify project registry was added
	if len(config.Auth.Container.Registries) != 1 {
		t.Errorf("Expected 1 registry, got: %d", len(config.Auth.Container.Registries))
	}
	projectRegistry, exists := config.Auth.Container.Registries["project-registry"]
	if !exists {
		t.Error("Expected project-registry to exist")
	}
	if projectRegistry.URL != "https://project.registry.com" {
		t.Errorf("Expected registry URL to be 'https://project.registry.com', got: %s", projectRegistry.URL)
	}
	if projectRegistry.Username != "projectuser" {
		t.Errorf("Expected registry username to be 'projectuser', got: %s", projectRegistry.Username)
	}
	if projectRegistry.AuthMethod != "token" {
		t.Errorf("Expected registry auth method to be 'token', got: %s", projectRegistry.AuthMethod)
	}
}

// TestAuthConfig_Validation tests auth configuration validation
func TestAuthConfig_Validation(t *testing.T) {
	// Test case: Auth configuration should have valid default values
	manager := NewManager()
	defaultConfig := manager.GetDefaultConfig()

	// Test Git auth validation
	if defaultConfig.Auth.Git.DefaultMethod != "ssh" && defaultConfig.Auth.Git.DefaultMethod != "https" {
		t.Errorf("Invalid default Git method: %s", defaultConfig.Auth.Git.DefaultMethod)
	}

	// Test SSH config validation
	if defaultConfig.Auth.Git.SSH.KeyPath == "" {
		t.Error("SSH key path should not be empty")
	}
	if defaultConfig.Auth.Git.SSH.KnownHostsFile == "" {
		t.Error("SSH known hosts file should not be empty")
	}

	// Test HTTPS config validation
	if defaultConfig.Auth.Git.HTTPS.TokenType != "github" && 
	   defaultConfig.Auth.Git.HTTPS.TokenType != "gitlab" && 
	   defaultConfig.Auth.Git.HTTPS.TokenType != "generic" {
		t.Errorf("Invalid token type: %s", defaultConfig.Auth.Git.HTTPS.TokenType)
	}
	if defaultConfig.Auth.Git.HTTPS.HelperTimeout <= 0 {
		t.Error("Helper timeout should be positive")
	}

	// Test Container auth validation
	if defaultConfig.Auth.Container.DefaultRegistry == "" {
		t.Error("Default registry should not be empty")
	}
	if defaultConfig.Auth.Container.HelperTimeout <= 0 {
		t.Error("Helper timeout should be positive")
	}
	if defaultConfig.Auth.Container.Registries == nil {
		t.Error("Registries map should be initialized")
	}
}

// TestRegistryConfig_Validation tests registry configuration validation
func TestRegistryConfig_Validation(t *testing.T) {
	// Test case: Registry configuration should have valid structure
	registry := RegistryConfig{
		URL:        "https://test.registry.com",
		Username:   "testuser",
		Password:   "testpass",
		AuthMethod: "basic",
		Insecure:   false,
		Namespace:  "testnamespace",
		APIVersion: "v2",
	}

	// Test basic validation
	if registry.URL == "" {
		t.Error("Registry URL should not be empty")
	}
	if registry.AuthMethod != "basic" && 
	   registry.AuthMethod != "token" && 
	   registry.AuthMethod != "oauth" {
		t.Errorf("Invalid auth method: %s", registry.AuthMethod)
	}
	if registry.APIVersion != "v1" && registry.APIVersion != "v2" {
		t.Errorf("Invalid API version: %s", registry.APIVersion)
	}
}
