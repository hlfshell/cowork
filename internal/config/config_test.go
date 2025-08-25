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
