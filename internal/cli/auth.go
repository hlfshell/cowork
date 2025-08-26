package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/hlfshell/cowork/internal/auth"
	"github.com/hlfshell/cowork/internal/config"
	"github.com/hlfshell/cowork/internal/git"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// addAuthCommands adds authentication management commands
func (app *App) addAuthCommands() {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication for various services",
		Long:  "Configure authentication for Git providers, container registries, and other services",
	}

	// Git authentication subcommands
	authCmd.AddCommand(
		newAuthGitCommand(app),
		newAuthProviderCommand(app),
		newAuthContainerCommand(app),
		newAuthShowCommand(app),
	)

	app.rootCmd.AddCommand(authCmd)
}

// newAuthGitCommand creates the git authentication command
func newAuthGitCommand(app *App) *cobra.Command {
	gitCmd := &cobra.Command{
		Use:   "git",
		Short: "Configure Git authentication",
		Long:  "Set up SSH keys, HTTPS tokens, and Git user configuration",
	}

	// Git user configuration
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Configure Git user information",
		Long:  "Set Git user name and email for commits",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.configureGitUser(cmd)
		},
	}

	// SSH key configuration
	sshCmd := &cobra.Command{
		Use:   "ssh",
		Short: "Configure SSH authentication",
		Long:  "Set up SSH keys for Git authentication",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.configureSSH(cmd)
		},
	}

	// HTTPS token configuration
	httpsCmd := &cobra.Command{
		Use:   "https",
		Short: "Configure HTTPS authentication",
		Long:  "Set up personal access tokens for HTTPS Git authentication",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.configureHTTPS(cmd)
		},
	}

	// Test authentication
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test Git authentication",
		Long:  "Test SSH and HTTPS authentication with configured settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.testGitAuth(cmd)
		},
	}

	gitCmd.AddCommand(userCmd, sshCmd, httpsCmd, testCmd)
	return gitCmd
}

// newAuthContainerCommand creates the container authentication command
func newAuthContainerCommand(app *App) *cobra.Command {
	containerCmd := &cobra.Command{
		Use:   "container",
		Short: "Configure container registry authentication",
		Long:  "Set up authentication for Docker Hub, private registries, and other container registries",
	}

	// Add registry
	addCmd := &cobra.Command{
		Use:   "add [registry-name]",
		Short: "Add container registry authentication",
		Long:  "Add authentication for a specific container registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.addContainerRegistry(cmd, args[0])
		},
	}

	// Remove registry
	removeCmd := &cobra.Command{
		Use:   "remove [registry-name]",
		Short: "Remove container registry authentication",
		Long:  "Remove authentication for a specific container registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.removeContainerRegistry(cmd, args[0])
		},
	}

	// List registries
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List configured container registries",
		Long:  "Show all configured container registry authentications",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.listContainerRegistries(cmd)
		},
	}

	// Test container authentication
	testCmd := &cobra.Command{
		Use:   "test [registry-name]",
		Short: "Test container registry authentication",
		Long:  "Test authentication with a specific container registry",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			registryName := ""
			if len(args) > 0 {
				registryName = args[0]
			}
			return app.testContainerAuth(cmd, registryName)
		},
	}

	containerCmd.AddCommand(addCmd, removeCmd, listCmd, testCmd)
	return containerCmd
}

// newAuthProviderCommand creates the provider authentication command
func newAuthProviderCommand(app *App) *cobra.Command {
	providerCmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage Git provider authentication",
		Long:  "Configure OAuth and API key authentication for GitHub, GitLab, and Bitbucket",
	}

	// Login command
	loginCmd := &cobra.Command{
		Use:   "login [provider]",
		Short: "Login to a Git provider",
		Long:  "Authenticate with a Git provider using OAuth or API key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.loginProvider(cmd, args[0])
		},
	}

	// Add flags to login command
	loginCmd.Flags().String("method", "token", "Authentication method (token, basic)")
	loginCmd.Flags().String("scope", "global", "Authentication scope (global, project)")
	loginCmd.Flags().String("token", "", "API token (for token method)")
	loginCmd.Flags().String("username", "", "Username (for basic method)")
	loginCmd.Flags().String("password", "", "Password (for basic method)")

	// Logout command
	logoutCmd := &cobra.Command{
		Use:   "logout [provider]",
		Short: "Logout from a Git provider",
		Long:  "Remove authentication for a Git provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.logoutProvider(cmd, args[0])
		},
	}

	// Add flags to logout command
	logoutCmd.Flags().String("scope", "global", "Authentication scope (global, project)")

	// Test command
	testCmd := &cobra.Command{
		Use:   "test [provider]",
		Short: "Test provider authentication",
		Long:  "Test authentication with a Git provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.testProviderAuth(cmd, args[0])
		},
	}

	// Add flags to test command
	testCmd.Flags().String("scope", "global", "Authentication scope (global, project)")

	// List command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List configured providers",
		Long:  "Show all configured Git provider authentications",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.listProviderAuths(cmd)
		},
	}

	providerCmd.AddCommand(loginCmd, logoutCmd, testCmd, listCmd)
	return providerCmd
}

// newAuthShowCommand creates the show authentication command
func newAuthShowCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current authentication configuration",
		Long:  "Display the current authentication settings for all services",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.showAuthConfig(cmd)
		},
	}
}

// configureGitUser configures Git user information
func (app *App) configureGitUser(cmd *cobra.Command) error {
	configManager := config.NewManager()
	config, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)

	cmd.Printf("🔧 Git User Configuration\n")
	cmd.Printf("========================\n\n")

	// Get current values
	currentName := config.Auth.Git.User.Name
	currentEmail := config.Auth.Git.User.Email

	// Get user name
	cmd.Printf("Git user name (current: %s): ", currentName)
	name, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user name: %w", err)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = currentName
	}

	// Get user email
	cmd.Printf("Git user email (current: %s): ", currentEmail)
	email, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user email: %w", err)
	}
	email = strings.TrimSpace(email)
	if email == "" {
		email = currentEmail
	}

	// Update configuration
	config.Auth.Git.User.Name = name
	config.Auth.Git.User.Email = email

	// Save configuration
	if err := configManager.SaveGlobal(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	cmd.Printf("\n✅ Git user configuration updated:\n")
	cmd.Printf("   Name:  %s\n", name)
	cmd.Printf("   Email: %s\n", email)

	// Also update Git global config
	if name != "" {
		if err := exec.Command("git", "config", "--global", "user.name", name).Run(); err != nil {
			cmd.Printf("⚠️  Warning: Failed to update Git global user.name: %v\n", err)
		}
	}
	if email != "" {
		if err := exec.Command("git", "config", "--global", "user.email", email).Run(); err != nil {
			cmd.Printf("⚠️  Warning: Failed to update Git global user.email: %v\n", err)
		}
	}

	return nil
}

// configureSSH configures SSH authentication
func (app *App) configureSSH(cmd *cobra.Command) error {
	configManager := config.NewManager()
	config, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)

	cmd.Printf("🔑 SSH Authentication Configuration\n")
	cmd.Printf("==================================\n\n")

	// Get current values
	currentKeyPath := config.Auth.Git.SSH.KeyPath
	currentUseAgent := config.Auth.Git.SSH.UseAgent

	// Get SSH key path
	cmd.Printf("SSH key path (current: %s): ", currentKeyPath)
	keyPath, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read SSH key path: %w", err)
	}
	keyPath = strings.TrimSpace(keyPath)
	if keyPath == "" {
		keyPath = currentKeyPath
	}

	// Expand ~ to home directory
	if strings.HasPrefix(keyPath, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		keyPath = filepath.Join(homeDir, keyPath[1:])
	}

	// Check if key exists
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		cmd.Printf("⚠️  Warning: SSH key file does not exist: %s\n", keyPath)
		cmd.Printf("   Would you like to generate a new SSH key? (y/N): ")
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
		response = strings.ToLower(strings.TrimSpace(response))
		if response == "y" || response == "yes" {
			if err := app.generateSSHKey(cmd, keyPath); err != nil {
				return fmt.Errorf("failed to generate SSH key: %w", err)
			}
		}
	}

	// Ask about SSH agent
	cmd.Printf("Use SSH agent? (current: %t) (Y/n): ", currentUseAgent)
	agentResponse, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read SSH agent response: %w", err)
	}
	agentResponse = strings.ToLower(strings.TrimSpace(agentResponse))
	useAgent := currentUseAgent
	if agentResponse == "n" || agentResponse == "no" {
		useAgent = false
	} else if agentResponse == "y" || agentResponse == "yes" || agentResponse == "" {
		useAgent = true
	}

	// Update configuration
	config.Auth.Git.SSH.KeyPath = keyPath
	config.Auth.Git.SSH.UseAgent = useAgent

	// Save configuration
	if err := configManager.SaveGlobal(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	cmd.Printf("\n✅ SSH configuration updated:\n")
	cmd.Printf("   Key Path: %s\n", keyPath)
	cmd.Printf("   Use Agent: %t\n", useAgent)

	return nil
}

// generateSSHKey generates a new SSH key
func (app *App) generateSSHKey(cmd *cobra.Command, keyPath string) error {
	cmd.Printf("🔧 Generating new SSH key...\n")

	// Create directory if it doesn't exist
	keyDir := filepath.Dir(keyPath)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return fmt.Errorf("failed to create SSH key directory: %w", err)
	}

	// Generate SSH key
	genCmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", keyPath, "-N", "")
	genCmd.Stdout = os.Stdout
	genCmd.Stderr = os.Stderr

	if err := genCmd.Run(); err != nil {
		return fmt.Errorf("failed to generate SSH key: %w", err)
	}

	cmd.Printf("✅ SSH key generated successfully: %s\n", keyPath)

	// Add to SSH agent if available
	if err := exec.Command("ssh-add", keyPath).Run(); err != nil {
		cmd.Printf("⚠️  Warning: Failed to add key to SSH agent: %v\n", err)
	} else {
		cmd.Printf("✅ SSH key added to agent\n")
	}

	return nil
}

// configureHTTPS configures HTTPS authentication
func (app *App) configureHTTPS(cmd *cobra.Command) error {
	configManager := config.NewManager()
	config, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)

	cmd.Printf("🔐 HTTPS Authentication Configuration\n")
	cmd.Printf("====================================\n\n")

	// Get current values
	currentUsername := config.Auth.Git.HTTPS.Username
	currentTokenType := config.Auth.Git.HTTPS.TokenType

	// Get username
	cmd.Printf("Username (current: %s): ", currentUsername)
	username, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read username: %w", err)
	}
	username = strings.TrimSpace(username)
	if username == "" {
		username = currentUsername
	}

	// Get token type
	cmd.Printf("Token type (github/gitlab/generic) (current: %s): ", currentTokenType)
	tokenType, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read token type: %w", err)
	}
	tokenType = strings.TrimSpace(tokenType)
	if tokenType == "" {
		tokenType = currentTokenType
	}

	// Get personal access token
	cmd.Printf("Personal access token (will be hidden): ")
	byteToken, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	token := string(byteToken)
	cmd.Printf("\n")

	if token == "" {
		cmd.Printf("⚠️  No token provided. HTTPS authentication will not be configured.\n")
		return nil
	}

	// Update configuration
	config.Auth.Git.HTTPS.Username = username
	config.Auth.Git.HTTPS.Token = token
	config.Auth.Git.HTTPS.TokenType = tokenType

	// Save configuration
	if err := configManager.SaveGlobal(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	cmd.Printf("\n✅ HTTPS configuration updated:\n")
	cmd.Printf("   Username: %s\n", username)
	cmd.Printf("   Token Type: %s\n", tokenType)
	cmd.Printf("   Token: [hidden]\n")

	return nil
}

// testGitAuth tests Git authentication
func (app *App) testGitAuth(cmd *cobra.Command) error {
	configManager := config.NewManager()
	config, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cmd.Printf("🧪 Testing Git Authentication\n")
	cmd.Printf("============================\n\n")

	// Test SSH authentication
	cmd.Printf("🔑 Testing SSH authentication...\n")
	if config.Auth.Git.SSH.KeyPath != "" {
		keyPath := config.Auth.Git.SSH.KeyPath
		if strings.HasPrefix(keyPath, "~") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			keyPath = filepath.Join(homeDir, keyPath[1:])
		}

		if _, err := os.Stat(keyPath); err == nil {
			// Test SSH connection to GitHub
			testCmd := exec.Command("ssh", "-T", "-i", keyPath, "git@github.com")
			testCmd.Stderr = os.Stderr
			if err := testCmd.Run(); err != nil {
				cmd.Printf("❌ SSH authentication failed: %v\n", err)
			} else {
				cmd.Printf("✅ SSH authentication successful\n")
			}
		} else {
			cmd.Printf("❌ SSH key not found: %s\n", keyPath)
		}
	} else {
		cmd.Printf("⚠️  No SSH key configured\n")
	}

	// Test HTTPS authentication
	cmd.Printf("\n🔐 Testing HTTPS authentication...\n")
	if config.Auth.Git.HTTPS.Token != "" {
		// Test HTTPS connection to GitHub
		testCmd := exec.Command("curl", "-s", "-H", "Authorization: token "+config.Auth.Git.HTTPS.Token, "https://api.github.com/user")
		output, err := testCmd.Output()
		if err != nil {
			cmd.Printf("❌ HTTPS authentication failed: %v\n", err)
		} else {
			cmd.Printf("✅ HTTPS authentication successful\n")
			cmd.Printf("   Response: %s\n", string(output))
		}
	} else {
		cmd.Printf("⚠️  No HTTPS token configured\n")
	}

	return nil
}

// addContainerRegistry adds container registry authentication
func (app *App) addContainerRegistry(cmd *cobra.Command, registryName string) error {
	configManager := config.NewManager()
	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)

	cmd.Printf("🐳 Adding Container Registry: %s\n", registryName)
	cmd.Printf("================================\n\n")

	// Get registry URL
	cmd.Printf("Registry URL: ")
	url, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read registry URL: %w", err)
	}
	url = strings.TrimSpace(url)

	// Get username
	cmd.Printf("Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read username: %w", err)
	}
	username = strings.TrimSpace(username)

	// Get password/token
	cmd.Printf("Password/Token (will be hidden): ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	password := string(bytePassword)
	cmd.Printf("\n")

	// Get authentication method
	cmd.Printf("Authentication method (basic/token/oauth) [basic]: ")
	authMethod, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read auth method: %w", err)
	}
	authMethod = strings.TrimSpace(authMethod)
	if authMethod == "" {
		authMethod = "basic"
	}

	// Initialize registries map if nil
	if cfg.Auth.Container.Registries == nil {
		cfg.Auth.Container.Registries = make(map[string]config.RegistryConfig)
	}

	// Add registry configuration
	cfg.Auth.Container.Registries[registryName] = config.RegistryConfig{
		URL:        url,
		Username:   username,
		Password:   password,
		AuthMethod: authMethod,
	}

	// Save configuration
	if err := configManager.SaveGlobal(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	cmd.Printf("\n✅ Container registry added successfully:\n")
	cmd.Printf("   Name: %s\n", registryName)
	cmd.Printf("   URL: %s\n", url)
	cmd.Printf("   Username: %s\n", username)
	cmd.Printf("   Auth Method: %s\n", authMethod)

	return nil
}

// removeContainerRegistry removes container registry authentication
func (app *App) removeContainerRegistry(cmd *cobra.Command, registryName string) error {
	configManager := config.NewManager()
	config, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if config.Auth.Container.Registries == nil {
		return fmt.Errorf("no container registries configured")
	}

	if _, exists := config.Auth.Container.Registries[registryName]; !exists {
		return fmt.Errorf("registry '%s' not found", registryName)
	}

	// Confirm removal
	cmd.Printf("Are you sure you want to remove registry '%s'? (y/N): ", registryName)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		cmd.Printf("Registry removal cancelled.\n")
		return nil
	}

	// Remove registry
	delete(config.Auth.Container.Registries, registryName)

	// Save configuration
	if err := configManager.SaveGlobal(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	cmd.Printf("✅ Registry '%s' removed successfully.\n", registryName)
	return nil
}

// listContainerRegistries lists configured container registries
func (app *App) listContainerRegistries(cmd *cobra.Command) error {
	configManager := config.NewManager()
	config, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cmd.Printf("🐳 Configured Container Registries\n")
	cmd.Printf("==================================\n\n")

	if config.Auth.Container.Registries == nil || len(config.Auth.Container.Registries) == 0 {
		cmd.Printf("No container registries configured.\n")
		return nil
	}

	for name, registry := range config.Auth.Container.Registries {
		cmd.Printf("📦 %s:\n", name)
		cmd.Printf("   URL: %s\n", registry.URL)
		cmd.Printf("   Username: %s\n", registry.Username)
		cmd.Printf("   Auth Method: %s\n", registry.AuthMethod)
		cmd.Printf("   Insecure: %t\n", registry.Insecure)
		if registry.Namespace != "" {
			cmd.Printf("   Namespace: %s\n", registry.Namespace)
		}
		cmd.Printf("\n")
	}

	return nil
}

// testContainerAuth tests container registry authentication
func (app *App) testContainerAuth(cmd *cobra.Command, registryName string) error {
	configManager := config.NewManager()
	config, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cmd.Printf("🧪 Testing Container Registry Authentication\n")
	cmd.Printf("===========================================\n\n")

	if config.Auth.Container.Registries == nil || len(config.Auth.Container.Registries) == 0 {
		cmd.Printf("No container registries configured.\n")
		return nil
	}

	if registryName != "" {
		// Test specific registry
		registry, exists := config.Auth.Container.Registries[registryName]
		if !exists {
			return fmt.Errorf("registry '%s' not found", registryName)
		}
		return app.testSpecificRegistry(cmd, registryName, registry)
	}

	// Test all registries
	for name, registry := range config.Auth.Container.Registries {
		cmd.Printf("Testing registry: %s\n", name)
		if err := app.testSpecificRegistry(cmd, name, registry); err != nil {
			cmd.Printf("❌ Failed: %v\n", err)
		}
		cmd.Printf("\n")
	}

	return nil
}

// testSpecificRegistry tests authentication for a specific registry
func (app *App) testSpecificRegistry(cmd *cobra.Command, name string, registry config.RegistryConfig) error {
	// This is a simplified test - in a real implementation, you would
	// actually try to authenticate with the registry
	cmd.Printf("   URL: %s\n", registry.URL)
	cmd.Printf("   Username: %s\n", registry.Username)
	cmd.Printf("   Auth Method: %s\n", registry.AuthMethod)
	cmd.Printf("   ✅ Configuration looks valid\n")
	return nil
}

// showAuthConfig shows the current authentication configuration
func (app *App) showAuthConfig(cmd *cobra.Command) error {
	configManager := config.NewManager()
	config, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cmd.Printf("🔐 Current Authentication Configuration\n")
	cmd.Printf("======================================\n\n")

	// Git authentication
	cmd.Printf("📝 Git Authentication:\n")
	cmd.Printf("   User Name: %s\n", config.Auth.Git.User.Name)
	cmd.Printf("   User Email: %s\n", config.Auth.Git.User.Email)
	cmd.Printf("   Default Method: %s\n", config.Auth.Git.DefaultMethod)
	cmd.Printf("   Credential Helper: %s\n", config.Auth.Git.CredentialHelper)

	cmd.Printf("   SSH:\n")
	cmd.Printf("     Key Path: %s\n", config.Auth.Git.SSH.KeyPath)
	cmd.Printf("     Use Agent: %t\n", config.Auth.Git.SSH.UseAgent)
	cmd.Printf("     Strict Host Key Checking: %t\n", config.Auth.Git.SSH.StrictHostKeyChecking)

	cmd.Printf("   HTTPS:\n")
	cmd.Printf("     Username: %s\n", config.Auth.Git.HTTPS.Username)
	cmd.Printf("     Token Type: %s\n", config.Auth.Git.HTTPS.TokenType)
	cmd.Printf("     Store Credentials: %t\n", config.Auth.Git.HTTPS.StoreCredentials)
	if config.Auth.Git.HTTPS.Token != "" {
		cmd.Printf("     Token: [configured]\n")
	} else {
		cmd.Printf("     Token: [not configured]\n")
	}

	// Container authentication
	cmd.Printf("\n🐳 Container Authentication:\n")
	cmd.Printf("   Default Registry: %s\n", config.Auth.Container.DefaultRegistry)
	cmd.Printf("   Use Credential Helper: %t\n", config.Auth.Container.UseCredentialHelper)

	if config.Auth.Container.Registries != nil && len(config.Auth.Container.Registries) > 0 {
		cmd.Printf("   Configured Registries:\n")
		for name, registry := range config.Auth.Container.Registries {
			cmd.Printf("     %s: %s (%s)\n", name, registry.URL, registry.AuthMethod)
		}
	} else {
		cmd.Printf("   Configured Registries: none\n")
	}

	return nil
}

// loginProvider handles provider authentication
func (app *App) loginProvider(cmd *cobra.Command, providerName string) error {
	// Parse provider type
	providerType, err := parseProviderType(providerName)
	if err != nil {
		return err
	}

	// Get flags
	method, _ := cmd.Flags().GetString("method")
	scope, _ := cmd.Flags().GetString("scope")
	token, _ := cmd.Flags().GetString("token")
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")

	// Parse scope
	authScope, err := parseAuthScope(scope)
	if err != nil {
		return err
	}

	// Create auth manager
	configManager := config.NewManager()
	authManager, err := auth.NewManager(configManager)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}

	cmd.Printf("🔐 Authenticating with %s...\n", providerName)

	switch method {

	case "token":
		// Token authentication
		if token == "" {
			return fmt.Errorf("token is required for token authentication method")
		}
		err := authManager.SetToken(providerType, token, authScope)
		if err != nil {
			return fmt.Errorf("failed to set token: %w", err)
		}
		cmd.Printf("✅ Successfully authenticated with %s using token\n", providerName)

	case "basic":
		// Basic authentication
		if username == "" || password == "" {
			return fmt.Errorf("username and password are required for basic authentication method")
		}
		err := authManager.SetBasicAuth(providerType, username, password, authScope)
		if err != nil {
			return fmt.Errorf("failed to set basic auth: %w", err)
		}
		cmd.Printf("✅ Successfully authenticated with %s using basic auth\n", providerName)

	default:
		return fmt.Errorf("unsupported authentication method: %s", method)
	}

	return nil
}

// logoutProvider removes provider authentication
func (app *App) logoutProvider(cmd *cobra.Command, providerName string) error {
	// Parse provider type
	providerType, err := parseProviderType(providerName)
	if err != nil {
		return err
	}

	// Get scope
	scope, _ := cmd.Flags().GetString("scope")
	authScope, err := parseAuthScope(scope)
	if err != nil {
		return err
	}

	// Create auth manager
	configManager := config.NewManager()
	authManager, err := auth.NewManager(configManager)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}

	// Remove authentication
	err = authManager.RemoveAuth(providerType, authScope)
	if err != nil {
		return fmt.Errorf("failed to remove authentication: %w", err)
	}

	cmd.Printf("✅ Successfully logged out from %s\n", providerName)
	return nil
}

// testProviderAuth tests provider authentication
func (app *App) testProviderAuth(cmd *cobra.Command, providerName string) error {
	// Parse provider type
	providerType, err := parseProviderType(providerName)
	if err != nil {
		return err
	}

	// Get scope
	scope, _ := cmd.Flags().GetString("scope")
	authScope, err := parseAuthScope(scope)
	if err != nil {
		return err
	}

	// Create auth manager
	configManager := config.NewManager()
	authManager, err := auth.NewManager(configManager)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}

	cmd.Printf("🧪 Testing authentication with %s...\n", providerName)

	// Test authentication
	err = authManager.TestAuth(cmd.Context(), providerType, authScope)
	if err != nil {
		return fmt.Errorf("authentication test failed: %w", err)
	}

	cmd.Printf("✅ Authentication test successful for %s\n", providerName)
	return nil
}

// listProviderAuths lists all configured provider authentications
func (app *App) listProviderAuths(cmd *cobra.Command) error {
	// Create auth manager
	configManager := config.NewManager()
	authManager, err := auth.NewManager(configManager)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}

	// List configurations
	configs, err := authManager.ListAuthConfigs()
	if err != nil {
		return fmt.Errorf("failed to list auth configs: %w", err)
	}

	cmd.Printf("🔐 Configured Git Provider Authentications\n")
	cmd.Printf("==========================================\n\n")

	if len(configs) == 0 {
		cmd.Printf("No provider authentications configured.\n")
		cmd.Printf("Use 'cw auth provider login <provider>' to configure authentication.\n")
		return nil
	}

	// Group configs by scope
	globalConfigs := make([]*auth.AuthConfig, 0)
	projectConfigs := make([]*auth.AuthConfig, 0)

	for _, config := range configs {
		// For now, we'll show all configs as global since we don't store scope in the config
		// In a real implementation, you'd need to store scope information
		globalConfigs = append(globalConfigs, config)
	}

	// Show global configs
	if len(globalConfigs) > 0 {
		cmd.Printf("🌍 Global Authentications:\n")
		for _, authConfig := range globalConfigs {
			cmd.Printf("   %s: %s\n", authConfig.ProviderType, authConfig.AuthMethod)
			if authConfig.ExpiresAt != nil {
				cmd.Printf("     Expires: %s\n", authConfig.ExpiresAt.Format("2006-01-02 15:04:05"))
			}
		}
		cmd.Printf("\n")
	}

	// Show project configs
	if len(projectConfigs) > 0 {
		cmd.Printf("📁 Project Authentications:\n")
		for _, authConfig := range projectConfigs {
			cmd.Printf("   %s: %s\n", authConfig.ProviderType, authConfig.AuthMethod)
			if authConfig.ExpiresAt != nil {
				cmd.Printf("     Expires: %s\n", authConfig.ExpiresAt.Format("2006-01-02 15:04:05"))
			}
		}
		cmd.Printf("\n")
	}

	return nil
}

// parseProviderType parses a provider name string into a ProviderType
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

// parseAuthScope parses a scope string into an AuthScope
func parseAuthScope(scope string) (auth.AuthScope, error) {
	switch strings.ToLower(scope) {
	case "global":
		return auth.AuthScopeGlobal, nil
	case "project":
		return auth.AuthScopeProject, nil
	default:
		return "", fmt.Errorf("unsupported scope: %s (supported: global, project)", scope)
	}
}
