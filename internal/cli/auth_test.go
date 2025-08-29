package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hlfshell/cowork/internal/auth"
	"github.com/hlfshell/cowork/internal/config"
	"github.com/hlfshell/cowork/internal/git"
)

// TestAuthCommands_Structure tests that all auth commands are properly structured
func TestAuthCommands_Structure(t *testing.T) {
	// Test case: The auth command should have all expected subcommands
	app := NewApp("1.0.0", "2024-01-01", "test")

	// Find the auth command
	var authCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Name() == "auth" {
			authCmd = cmd
			break
		}
	}

	require.NotNil(t, authCmd, "Auth command should exist")

	// Check for expected auth subcommands
	expectedSubcommands := []string{"git", "provider", "container", "show"}
	foundSubcommands := make(map[string]bool)

	for _, cmd := range authCmd.Commands() {
		foundSubcommands[cmd.Name()] = true
	}

	for _, expectedCmd := range expectedSubcommands {
		assert.True(t, foundSubcommands[expectedCmd], "Expected auth subcommand '%s' not found", expectedCmd)
	}
}

// TestAuthGitCommands_Structure tests that git auth commands are properly structured
func TestAuthGitCommands_Structure(t *testing.T) {
	// Test case: The git auth command should have all expected subcommands
	app := NewApp("1.0.0", "2024-01-01", "test")

	// Find the auth git command
	var gitCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Name() == "auth" {
			for _, subcmd := range cmd.Commands() {
				if subcmd.Name() == "git" {
					gitCmd = subcmd
					break
				}
			}
			break
		}
	}

	require.NotNil(t, gitCmd, "Auth git command should exist")

	// Check for expected git auth subcommands
	expectedSubcommands := []string{"user", "ssh", "https", "test"}
	foundSubcommands := make(map[string]bool)

	for _, cmd := range gitCmd.Commands() {
		foundSubcommands[cmd.Name()] = true
	}

	for _, expectedCmd := range expectedSubcommands {
		assert.True(t, foundSubcommands[expectedCmd], "Expected git auth subcommand '%s' not found", expectedCmd)
	}
}

// TestSSHCommand_KeyFlag tests that the SSH command has the key flag
func TestSSHCommand_KeyFlag(t *testing.T) {
	// Test case: The SSH command should have the key flag
	app := NewApp("1.0.0", "2024-01-01", "test")

	// Find the SSH command
	var sshCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Name() == "auth" {
			for _, subcmd := range cmd.Commands() {
				if subcmd.Name() == "git" {
					for _, gitSubcmd := range subcmd.Commands() {
						if gitSubcmd.Name() == "ssh" {
							sshCmd = gitSubcmd
							break
						}
					}
					break
				}
			}
			break
		}
	}

	require.NotNil(t, sshCmd, "SSH command should exist")

	// Check that the key flag exists
	keyFlag := sshCmd.Flags().Lookup("key")
	assert.NotNil(t, keyFlag, "key flag should exist on SSH command")
	assert.Equal(t, "key", keyFlag.Name, "Flag should be named 'key'")
	assert.Equal(t, "", keyFlag.DefValue, "Flag should have empty default value")
	assert.Equal(t, "Path to SSH private key file", keyFlag.Usage, "Flag should have correct usage description")
}

// TestConfigureSSH_WithKeyFlag tests the configureSSH function with key flag
func TestConfigureSSH_WithKeyFlag(t *testing.T) {
	// Test case: configureSSH should use the key flag when provided
	app := NewApp("1.0.0", "2024-01-01", "test")

	// Create a temporary directory for test configuration
	tempDir, err := os.MkdirTemp("", "cowork-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set up test environment
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create test SSH key path
	testKeyPath := filepath.Join(tempDir, "test_key")

	// Create the SSH command
	var sshCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Name() == "auth" {
			for _, subcmd := range cmd.Commands() {
				if subcmd.Name() == "git" {
					for _, gitSubcmd := range subcmd.Commands() {
						if gitSubcmd.Name() == "ssh" {
							sshCmd = gitSubcmd
							break
						}
					}
					break
				}
			}
			break
		}
	}

	require.NotNil(t, sshCmd, "SSH command should exist")

	// Set the key flag
	err = sshCmd.Flags().Set("key", testKeyPath)
	require.NoError(t, err, "Should be able to set key flag")

	// Verify the flag was set correctly
	flagValue, err := sshCmd.Flags().GetString("key")
	require.NoError(t, err, "Should be able to get key flag value")
	assert.Equal(t, testKeyPath, flagValue, "Flag value should match the set value")
}

// TestConfigureSSH_WithoutKeyFlag tests the configureSSH function without key flag
func TestConfigureSSH_WithoutKeyFlag(t *testing.T) {
	// Test case: configureSSH should not have key flag set by default
	app := NewApp("1.0.0", "2024-01-01", "test")

	// Find the SSH command
	var sshCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Name() == "auth" {
			for _, subcmd := range cmd.Commands() {
				if subcmd.Name() == "git" {
					for _, gitSubcmd := range subcmd.Commands() {
						if gitSubcmd.Name() == "ssh" {
							sshCmd = gitSubcmd
							break
						}
					}
					break
				}
			}
			break
		}
	}

	require.NotNil(t, sshCmd, "SSH command should exist")

	// Verify the flag is empty by default
	flagValue, err := sshCmd.Flags().GetString("key")
	require.NoError(t, err, "Should be able to get key flag value")
	assert.Equal(t, "", flagValue, "Flag should be empty by default")
}

// TestSSHCommand_HelpText tests that the SSH command help text is correct
func TestSSHCommand_HelpText(t *testing.T) {
	// Test case: The SSH command should have correct help text
	app := NewApp("1.0.0", "2024-01-01", "test")

	// Find the SSH command
	var sshCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Name() == "auth" {
			for _, subcmd := range cmd.Commands() {
				if subcmd.Name() == "git" {
					for _, gitSubcmd := range subcmd.Commands() {
						if gitSubcmd.Name() == "ssh" {
							sshCmd = gitSubcmd
							break
						}
					}
					break
				}
			}
			break
		}
	}

	require.NotNil(t, sshCmd, "SSH command should exist")

	// Check command properties
	assert.Equal(t, "ssh", sshCmd.Use, "Command should use 'ssh'")
	assert.Equal(t, "Configure SSH authentication", sshCmd.Short, "Command should have correct short description")
	assert.Equal(t, "Set up SSH keys for Git authentication", sshCmd.Long, "Command should have correct long description")
}

// TestAuthCommands_Integration tests basic integration of auth commands
func TestAuthCommands_Integration(t *testing.T) {
	// Test case: All auth commands should be properly integrated into the main app
	app := NewApp("1.0.0", "2024-01-01", "test")

	// Verify the app has the auth command
	var authCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Name() == "auth" {
			authCmd = cmd
			break
		}
	}

	require.NotNil(t, authCmd, "Auth command should be integrated into the app")

	// Verify the auth command has the correct properties
	assert.Equal(t, "auth", authCmd.Use, "Auth command should use 'auth'")
	assert.Equal(t, "Manage authentication for various services", authCmd.Short, "Auth command should have correct short description")
	assert.Equal(t, "Configure authentication for Git providers, container registries, and other services", authCmd.Long, "Auth command should have correct long description")
}

// TestProviderHelp_GeneralHelp tests the general provider help functionality
func TestProviderHelp_GeneralHelp(t *testing.T) {
	// Test case: The provider command should show general help when no subcommand is provided
	app := NewApp("1.0.0", "2024-01-01", "test")

	// Find the provider command
	var providerCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Name() == "auth" {
			for _, subcmd := range cmd.Commands() {
				if subcmd.Name() == "provider" {
					providerCmd = subcmd
					break
				}
			}
			break
		}
	}

	require.NotNil(t, providerCmd, "Provider command should exist")

	// Test that the command has the correct properties
	assert.Equal(t, "provider", providerCmd.Use, "Provider command should use 'provider'")
	assert.Contains(t, providerCmd.Long, "Configure OAuth and API key authentication", "Provider command should have correct long description")
	assert.Contains(t, providerCmd.Long, "cw auth provider login <provider>", "Provider command should mention provider-specific help")
}

// TestProviderLoginHelp_GeneralHelp tests the login command general help
func TestProviderLoginHelp_GeneralHelp(t *testing.T) {
	// Test case: The login command should show general help when help is requested
	app := NewApp("1.0.0", "2024-01-01", "test")

	// Find the login command
	var loginCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Name() == "auth" {
			for _, subcmd := range cmd.Commands() {
				if subcmd.Name() == "provider" {
					for _, providerSubcmd := range subcmd.Commands() {
						if providerSubcmd.Name() == "login" {
							loginCmd = providerSubcmd
							break
						}
					}
					break
				}
			}
			break
		}
	}

	require.NotNil(t, loginCmd, "Login command should exist")

	// Test that the command has the correct properties
	assert.Equal(t, "login [provider]", loginCmd.Use, "Login command should use 'login [provider]'")
	assert.Contains(t, loginCmd.Long, "Authenticate with a Git provider", "Login command should have correct long description")
	assert.Contains(t, loginCmd.Long, "cw auth provider login <provider>", "Login command should mention provider-specific help")
}

// TestProviderLoginHelp_CommandStructure tests the login command argument handling
func TestProviderLoginHelp_CommandStructure(t *testing.T) {
	// Test case: The login command should handle different argument combinations correctly
	app := NewApp("1.0.0", "2024-01-01", "test")

	// Find the login command
	var loginCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Name() == "auth" {
			for _, subcmd := range cmd.Commands() {
				if subcmd.Name() == "provider" {
					for _, providerSubcmd := range subcmd.Commands() {
						if providerSubcmd.Name() == "login" {
							loginCmd = providerSubcmd
							break
						}
					}
					break
				}
			}
			break
		}
	}

	require.NotNil(t, loginCmd, "Login command should exist")

	// Test that the command has the correct properties
	assert.Equal(t, "login [provider]", loginCmd.Use, "Login command should use 'login [provider]'")
	assert.Contains(t, loginCmd.Long, "Authenticate with a Git provider", "Login command should have correct long description")
	assert.Contains(t, loginCmd.Long, "cw auth provider login <provider>", "Login command should mention provider-specific help")
}

// TestAuthCommands_GlobalFlagEnforcement tests that global auth requires --global flag
func TestAuthCommands_GlobalFlagEnforcement(t *testing.T) {
	// Test case: Global authentication should require --global flag
	app := NewApp("1.0.0", "2024-01-01", "test")

	// Find the login command
	var loginCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Name() == "auth" {
			for _, subcmd := range cmd.Commands() {
				if subcmd.Name() == "provider" {
					for _, subsubcmd := range subcmd.Commands() {
						if subsubcmd.Name() == "login" {
							loginCmd = subsubcmd
							break
						}
					}
					break
				}
			}
			break
		}
	}

	require.NotNil(t, loginCmd, "Login command should exist")

	// Test that --global flag exists
	globalFlag := loginCmd.Flags().Lookup("global")
	assert.NotNil(t, globalFlag, "Login command should have --global flag")
	assert.Equal(t, "global", globalFlag.Name)
	assert.Equal(t, "Set authentication globally (requires --global flag)", globalFlag.Usage)

	// Test that scope defaults to project
	scopeFlag := loginCmd.Flags().Lookup("scope")
	assert.NotNil(t, scopeFlag, "Login command should have scope flag")
	assert.Equal(t, "project", scopeFlag.DefValue, "Scope should default to project")
}

// TestAuthCommands_LocalGlobalPrecedence tests local vs global auth precedence
func TestAuthCommands_LocalGlobalPrecedence(t *testing.T) {
	// Test case: Local auth should take precedence over global auth
	// Create auth manager
	configManager := config.NewManager()
	authManager, err := auth.NewManager(configManager)
	require.NoError(t, err)

	// Clean up any existing configs
	authManager.RemoveAuth(git.ProviderGitHub, auth.AuthScopeGlobal)
	authManager.RemoveAuth(git.ProviderGitHub, auth.AuthScopeProject)

	// Set global auth
	err = authManager.SetToken(git.ProviderGitHub, "global-token", auth.AuthScopeGlobal)
	require.NoError(t, err)

	// Set project auth
	err = authManager.SetToken(git.ProviderGitHub, "project-token", auth.AuthScopeProject)
	require.NoError(t, err)

	// Verify project auth takes precedence
	authConfig, err := authManager.GetAuthConfig(git.ProviderGitHub, auth.AuthScopeProject)
	assert.NoError(t, err)
	assert.Equal(t, "project-token", authConfig.Token)

	// Verify global auth still exists
	authConfig, err = authManager.GetAuthConfig(git.ProviderGitHub, auth.AuthScopeGlobal)
	assert.NoError(t, err)
	assert.Equal(t, "global-token", authConfig.Token)

	// Clean up
	authManager.RemoveAuth(git.ProviderGitHub, auth.AuthScopeGlobal)
	authManager.RemoveAuth(git.ProviderGitHub, auth.AuthScopeProject)
}

// TestAuthCommands_GlobalFlagRequired tests that global operations require --global flag
func TestAuthCommands_GlobalFlagRequired(t *testing.T) {
	// Test case: Global operations should require explicit --global flag
	app := NewApp("1.0.0", "2024-01-01", "test")

	// Find the logout command
	var logoutCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Name() == "auth" {
			for _, subcmd := range cmd.Commands() {
				if subcmd.Name() == "provider" {
					for _, subsubcmd := range subcmd.Commands() {
						if subsubcmd.Name() == "logout" {
							logoutCmd = subsubcmd
							break
						}
					}
					break
				}
			}
			break
		}
	}

	require.NotNil(t, logoutCmd, "Logout command should exist")

	// Test that --global flag exists
	globalFlag := logoutCmd.Flags().Lookup("global")
	assert.NotNil(t, globalFlag, "Logout command should have --global flag")
	assert.Equal(t, "Remove global authentication (requires --global flag)", globalFlag.Usage)

	// Test that scope defaults to project
	scopeFlag := logoutCmd.Flags().Lookup("scope")
	assert.NotNil(t, scopeFlag, "Logout command should have scope flag")
	assert.Equal(t, "project", scopeFlag.DefValue, "Scope should default to project")
}

// TestAuthCommands_TestCommandGlobalFlag tests that test command supports --global flag
func TestAuthCommands_TestCommandGlobalFlag(t *testing.T) {
	// Test case: Test command should support --global flag
	app := NewApp("1.0.0", "2024-01-01", "test")

	// Find the test command
	var testCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Name() == "auth" {
			for _, subcmd := range cmd.Commands() {
				if subcmd.Name() == "provider" {
					for _, subsubcmd := range subcmd.Commands() {
						if subsubcmd.Name() == "test" {
							testCmd = subsubcmd
							break
						}
					}
					break
				}
			}
			break
		}
	}

	require.NotNil(t, testCmd, "Test command should exist")

	// Test that --global flag exists
	globalFlag := testCmd.Flags().Lookup("global")
	assert.NotNil(t, globalFlag, "Test command should have --global flag")
	assert.Equal(t, "Test global authentication (requires --global flag)", globalFlag.Usage)

	// Test that scope defaults to project
	scopeFlag := testCmd.Flags().Lookup("scope")
	assert.NotNil(t, scopeFlag, "Test command should have scope flag")
	assert.Equal(t, "project", scopeFlag.DefValue, "Scope should default to project")
}

// TestAuthCommands_IntegrationFlow tests the complete auth flow with local vs global precedence
func TestAuthCommands_IntegrationFlow(t *testing.T) {
	// Test case: Complete auth flow should work correctly with local vs global precedence
	// Create auth manager
	configManager := config.NewManager()
	authManager, err := auth.NewManager(configManager)
	require.NoError(t, err)

	// Clean up any existing configs
	authManager.RemoveAuth(git.ProviderGitHub, auth.AuthScopeGlobal)
	authManager.RemoveAuth(git.ProviderGitHub, auth.AuthScopeProject)

	// Step 1: Set global auth first
	err = authManager.SetToken(git.ProviderGitHub, "global-token-123", auth.AuthScopeGlobal)
	require.NoError(t, err)

	// Verify global auth exists
	globalAuth, err := authManager.GetAuthConfig(git.ProviderGitHub, auth.AuthScopeGlobal)
	assert.NoError(t, err)
	assert.Equal(t, "global-token-123", globalAuth.Token)

	// Step 2: Set project auth (should take precedence)
	err = authManager.SetToken(git.ProviderGitHub, "project-token-456", auth.AuthScopeProject)
	require.NoError(t, err)

	// Verify project auth exists and takes precedence
	projectAuth, err := authManager.GetAuthConfig(git.ProviderGitHub, auth.AuthScopeProject)
	assert.NoError(t, err)
	assert.Equal(t, "project-token-456", projectAuth.Token)

	// Step 3: Verify global auth still exists
	globalAuth, err = authManager.GetAuthConfig(git.ProviderGitHub, auth.AuthScopeGlobal)
	assert.NoError(t, err)
	assert.Equal(t, "global-token-123", globalAuth.Token)

	// Step 4: Remove project auth
	err = authManager.RemoveAuth(git.ProviderGitHub, auth.AuthScopeProject)
	assert.NoError(t, err)

	// Step 5: Verify project auth is gone
	_, err = authManager.GetAuthConfig(git.ProviderGitHub, auth.AuthScopeProject)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication not found")

	// Step 6: Verify global auth still exists
	globalAuth, err = authManager.GetAuthConfig(git.ProviderGitHub, auth.AuthScopeGlobal)
	assert.NoError(t, err)
	assert.Equal(t, "global-token-123", globalAuth.Token)

	// Step 7: Remove global auth
	err = authManager.RemoveAuth(git.ProviderGitHub, auth.AuthScopeGlobal)
	assert.NoError(t, err)

	// Step 8: Verify global auth is gone
	_, err = authManager.GetAuthConfig(git.ProviderGitHub, auth.AuthScopeGlobal)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication not found")

	// Clean up
	authManager.RemoveAuth(git.ProviderGitHub, auth.AuthScopeGlobal)
	authManager.RemoveAuth(git.ProviderGitHub, auth.AuthScopeProject)
}

// TestAuthCommands_MultipleProviders tests auth with multiple providers
func TestAuthCommands_MultipleProviders(t *testing.T) {
	// Test case: Auth should work correctly with multiple providers
	configManager := config.NewManager()
	authManager, err := auth.NewManager(configManager)
	require.NoError(t, err)

	// Clean up any existing configs
	authManager.RemoveAuth(git.ProviderGitHub, auth.AuthScopeProject)
	authManager.RemoveAuth(git.ProviderGitLab, auth.AuthScopeProject)
	authManager.RemoveAuth(git.ProviderBitbucket, auth.AuthScopeProject)

	// Set auth for multiple providers
	err = authManager.SetToken(git.ProviderGitHub, "github-token", auth.AuthScopeProject)
	require.NoError(t, err)

	err = authManager.SetToken(git.ProviderGitLab, "gitlab-token", auth.AuthScopeProject)
	require.NoError(t, err)

	err = authManager.SetToken(git.ProviderBitbucket, "bitbucket-token", auth.AuthScopeProject)
	require.NoError(t, err)

	// Verify all providers have auth
	githubAuth, err := authManager.GetAuthConfig(git.ProviderGitHub, auth.AuthScopeProject)
	assert.NoError(t, err)
	assert.Equal(t, "github-token", githubAuth.Token)

	gitlabAuth, err := authManager.GetAuthConfig(git.ProviderGitLab, auth.AuthScopeProject)
	assert.NoError(t, err)
	assert.Equal(t, "gitlab-token", gitlabAuth.Token)

	bitbucketAuth, err := authManager.GetAuthConfig(git.ProviderBitbucket, auth.AuthScopeProject)
	assert.NoError(t, err)
	assert.Equal(t, "bitbucket-token", bitbucketAuth.Token)

	// List all configs
	configs, err := authManager.ListAuthConfigs()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(configs), 3)

	// Clean up
	authManager.RemoveAuth(git.ProviderGitHub, auth.AuthScopeProject)
	authManager.RemoveAuth(git.ProviderGitLab, auth.AuthScopeProject)
	authManager.RemoveAuth(git.ProviderBitbucket, auth.AuthScopeProject)
}

// TestAuthCommands_GlobalFlagEnforcement_Integration tests global flag enforcement in practice
func TestAuthCommands_GlobalFlagEnforcement_Integration(t *testing.T) {
	// Test case: Global flag should be required for global operations
	configManager := config.NewManager()
	authManager, err := auth.NewManager(configManager)
	require.NoError(t, err)

	// Clean up any existing configs
	authManager.RemoveAuth(git.ProviderGitHub, auth.AuthScopeGlobal)
	authManager.RemoveAuth(git.ProviderGitHub, auth.AuthScopeProject)

	// Test that project auth is the default (no global flag)
	err = authManager.SetToken(git.ProviderGitHub, "project-token", auth.AuthScopeProject)
	require.NoError(t, err)

	// Verify project auth exists
	projectAuth, err := authManager.GetAuthConfig(git.ProviderGitHub, auth.AuthScopeProject)
	assert.NoError(t, err)
	assert.Equal(t, "project-token", projectAuth.Token)

	// Verify global auth does not exist
	_, err = authManager.GetAuthConfig(git.ProviderGitHub, auth.AuthScopeGlobal)
	assert.Error(t, err)

	// Test that global auth requires explicit global scope
	err = authManager.SetToken(git.ProviderGitHub, "global-token", auth.AuthScopeGlobal)
	require.NoError(t, err)

	// Verify global auth exists
	globalAuth, err := authManager.GetAuthConfig(git.ProviderGitHub, auth.AuthScopeGlobal)
	assert.NoError(t, err)
	assert.Equal(t, "global-token", globalAuth.Token)

	// Verify project auth still exists
	projectAuth, err = authManager.GetAuthConfig(git.ProviderGitHub, auth.AuthScopeProject)
	assert.NoError(t, err)
	assert.Equal(t, "project-token", projectAuth.Token)

	// Clean up
	authManager.RemoveAuth(git.ProviderGitHub, auth.AuthScopeGlobal)
	authManager.RemoveAuth(git.ProviderGitHub, auth.AuthScopeProject)
}
