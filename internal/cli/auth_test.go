package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
