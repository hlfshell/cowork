package cli

import (
	"fmt"
	"strings"

	"github.com/hlfshell/cowork/internal/auth"
	"github.com/hlfshell/cowork/internal/cli/providers"
	"github.com/hlfshell/cowork/internal/git"
	"github.com/spf13/cobra"
)

func (app *App) listProviders(cmd *cobra.Command) error {
	cmd.Println("🔗 Available Providers")
	cmd.Println("======================")
	cmd.Println()
	cmd.Println("Git Providers:")

	availableProviders := providers.GetAvailableProviders()

	cmd.Println("  github     - GitHub (GitHub.com)")
	cmd.Println("  gitlab     - GitLab (GitLab.com or self-hosted)")
	cmd.Println("  bitbucket  - Bitbucket (Bitbucket.org)")

	cmd.Println()
	cmd.Println("Usage:")
	cmd.Println("  cowork config provider <provider-name>")
	cmd.Println()
	cmd.Println("Examples:")
	for _, provider := range availableProviders {
		cmd.Printf("  cowork config provider %s\n", provider)
	}

	return nil
}

func (app *App) showProviderHelp(cmd *cobra.Command, providerName string) error {
	// Get help content from embedded providers
	helpContent, exists := providers.GetProviderHelp(providerName)
	if !exists {
		availableProviders := providers.GetAvailableProviders()
		return fmt.Errorf("unknown provider: %s. Available providers: %s", providerName, strings.Join(availableProviders, ", "))
	}

	cmd.Println(helpContent)
	return nil
}

func (app *App) loginProvider(cmd *cobra.Command, providerName string) error {
	// Validate provider
	availableProviders := providers.GetAvailableProviders()
	validProvider := false
	for _, provider := range availableProviders {
		if provider == providerName {
			validProvider = true
			break
		}
	}
	if !validProvider {
		return fmt.Errorf("unknown provider: %s. Available providers: %s", providerName, strings.Join(availableProviders, ", "))
	}

	// Check if any authentication flags are provided
	token, _ := cmd.Flags().GetString("token")
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")

	// If no authentication flags are provided, show help
	if token == "" && username == "" && password == "" {
		return app.showProviderHelp(cmd, providerName)
	}

	// Get authentication method and scope
	method, _ := cmd.Flags().GetString("method")
	targetScope, _ := cmd.Flags().GetString("scope")

	var scope auth.AuthScope
	if targetScope == "" {
		scope = auth.AuthScopeProject
	} else if targetScope != "global" && targetScope != "project" {
		return fmt.Errorf("invalid scope: %s. Valid scopes: global, project", targetScope)
	} else {
		scope = auth.AuthScope(targetScope)
	}

	cmd.Printf("🔐 Logging into %s provider...\n", providerName)
	cmd.Printf("Method: %s\n", method)
	cmd.Printf("Scope: %s\n", scope)

	switch method {
	case "token":
		return app.loginWithToken(cmd, providerName, scope)
	case "basic":
		return app.loginWithBasic(cmd, providerName, scope)
	default:
		return fmt.Errorf("unsupported authentication method: %s. Supported methods: token, basic", method)
	}
}

func (app *App) loginWithToken(cmd *cobra.Command, providerName string, scope auth.AuthScope) error {
	token, _ := cmd.Flags().GetString("token")

	// If no token provided, prompt for it
	if token == "" {
		cmd.Printf("Enter your %s token: ", providerName)
		fmt.Scanln(&token)
		if token == "" {
			return fmt.Errorf("token is required for %s authentication", providerName)
		}
	}

	// Convert provider name to git.ProviderType
	var providerType git.ProviderType
	switch providerName {
	case "github":
		providerType = git.ProviderGitHub
	case "gitlab":
		providerType = git.ProviderGitLab
	case "bitbucket":
		providerType = git.ProviderBitbucket
	default:
		return fmt.Errorf("unsupported provider type: %s", providerName)
	}

	// Convert scope string to auth.AuthScope
	var authScope auth.AuthScope
	switch scope {
	case "global":
		authScope = auth.AuthScopeGlobal
	case "project":
		authScope = auth.AuthScopeProject
	default:
		return fmt.Errorf("invalid scope: %s. Valid scopes: global, project", scope)
	}

	cmd.Printf("🔐 Logging into %s provider...\n", providerName)
	cmd.Printf("Method: token\n")
	cmd.Printf("Scope: %s\n", scope)

	// Create auth manager
	authManager, err := auth.NewManager(app.configManager)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}

	// Store the token
	if err := authManager.SetToken(providerType, token, authScope); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	cmd.Printf("✅ Token stored successfully for %s\n", providerName)
	cmd.Printf("📝 Authentication configured for %s scope\n", scope)

	// Test the authentication
	ctx := cmd.Context()
	if err := authManager.TestAuth(ctx, providerType, authScope); err != nil {
		cmd.Printf("⚠️  Warning: Token stored but validation failed: %v\n", err)
		cmd.Printf("   You may want to check your token permissions\n")
	} else {
		cmd.Printf("🔍 Token validated successfully with %s API\n", providerName)
	}

	return nil
}

func (app *App) loginWithBasic(cmd *cobra.Command, providerName string, scope auth.AuthScope) error {
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")

	// If no credentials provided, prompt for them
	if username == "" {
		cmd.Printf("Enter your %s username: ", providerName)
		fmt.Scanln(&username)
		if username == "" {
			return fmt.Errorf("username is required for %s basic authentication", providerName)
		}
	}

	if password == "" {
		cmd.Printf("Enter your %s password/app password: ", providerName)
		fmt.Scanln(&password)
		if password == "" {
			return fmt.Errorf("password is required for %s basic authentication", providerName)
		}
	}

	// Convert provider name to git.ProviderType
	var providerType git.ProviderType
	switch providerName {
	case "github":
		providerType = git.ProviderGitHub
	case "gitlab":
		providerType = git.ProviderGitLab
	case "bitbucket":
		providerType = git.ProviderBitbucket
	default:
		return fmt.Errorf("unsupported provider type: %s", providerName)
	}

	// Convert scope string to auth.AuthScope
	cmd.Printf("🔐 Logging into %s provider...\n", providerName)
	cmd.Printf("Method: basic\n")
	cmd.Printf("Scope: %s\n", scope)

	// Create auth manager
	authManager, err := auth.NewManager(app.configManager)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}

	// Store the credentials
	if err := authManager.SetBasicAuth(providerType, username, password, scope); err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}

	cmd.Printf("✅ Credentials stored successfully for %s\n", providerName)
	cmd.Printf("📝 Authentication configured for %s scope\n", scope)

	// Test the authentication
	ctx := cmd.Context()
	if err := authManager.TestAuth(ctx, providerType, scope); err != nil {
		cmd.Printf("⚠️  Warning: Credentials stored but validation failed: %v\n", err)
		cmd.Printf("   You may want to check your credentials\n")
	} else {
		cmd.Printf("🔍 Credentials validated successfully with %s API\n", providerName)
	}

	return nil
}

// configureGitHub handles GitHub configuration
func (app *App) configureGitHub(cmd *cobra.Command) error {
	cmd.Println("🐙 GitHub Configuration")
	cmd.Println("======================")
	cmd.Println()
	cmd.Println("GitHub authentication methods:")
	cmd.Println("  1. Personal Access Token (recommended)")
	cmd.Println("  2. SSH Key")
	cmd.Println("  3. Basic Authentication")
	cmd.Println()
	cmd.Println("Commands:")
	cmd.Println("  cowork config provider github login   - Authenticate with GitHub")
	cmd.Println("  cowork config provider github test    - Test GitHub authentication")
	cmd.Println()
	cmd.Println("Use --scope project or --scope global to specify scope (defaults to project)")
	cmd.Println()
	cmd.Println("For help getting a GitHub token, see:")
	cmd.Println("  https://github.com/settings/tokens")

	return nil
}

// loginGitHub handles GitHub authentication
func (app *App) loginGitHub(cmd *cobra.Command) error {
	method, _ := cmd.Flags().GetString("method")
	token, _ := cmd.Flags().GetString("token")
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")
	scope, _ := cmd.Flags().GetString("scope")

	cmd.Printf("🔐 Logging into GitHub...\n")
	cmd.Printf("Method: %s\n", method)
	cmd.Printf("Scope: %s\n", scope)

	// Create auth manager
	authManager, err := auth.NewManager(app.configManager)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}

	// Convert scope string to auth.AuthScope
	var authScope auth.AuthScope
	switch scope {
	case "global":
		authScope = auth.AuthScopeGlobal
	case "project":
		authScope = auth.AuthScopeProject
	default:
		return fmt.Errorf("invalid scope: %s. Valid scopes: global, project", scope)
	}

	switch method {
	case "token":
		if token == "" {
			cmd.Printf("Enter your GitHub token: ")
			fmt.Scanln(&token)
			if token == "" {
				return fmt.Errorf("token is required for GitHub authentication")
			}
		}

		if err := authManager.SetToken(git.ProviderGitHub, token, authScope); err != nil {
			return fmt.Errorf("failed to store token: %w", err)
		}

		cmd.Printf("✅ Token stored successfully for GitHub\n")

	case "basic":
		if username == "" {
			cmd.Printf("Enter your GitHub username: ")
			fmt.Scanln(&username)
			if username == "" {
				return fmt.Errorf("username is required for GitHub basic authentication")
			}
		}

		if password == "" {
			cmd.Printf("Enter your GitHub password/token: ")
			fmt.Scanln(&password)
			if password == "" {
				return fmt.Errorf("password is required for GitHub basic authentication")
			}
		}

		if err := authManager.SetBasicAuth(git.ProviderGitHub, username, password, authScope); err != nil {
			return fmt.Errorf("failed to store credentials: %w", err)
		}

		cmd.Printf("✅ Credentials stored successfully for GitHub\n")

	default:
		return fmt.Errorf("unsupported authentication method: %s. Supported methods: token, basic, ssh", method)
	}

	// Test the authentication
	ctx := cmd.Context()
	if err := authManager.TestAuth(ctx, git.ProviderGitHub, authScope); err != nil {
		cmd.Printf("⚠️  Warning: Authentication stored but validation failed: %v\n", err)
		cmd.Printf("   You may want to check your credentials\n")
	} else {
		cmd.Printf("🔍 Authentication validated successfully with GitHub API\n")
	}

	return nil
}

// testGitHub tests GitHub authentication
func (app *App) testGitHub(cmd *cobra.Command) error {
	scope, _ := cmd.Flags().GetString("scope")

	cmd.Printf("🧪 Testing GitHub authentication...\n")
	cmd.Printf("Scope: %s\n", scope)

	// Create auth manager
	authManager, err := auth.NewManager(app.configManager)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}

	// Convert scope string to auth.AuthScope
	var authScope auth.AuthScope
	switch scope {
	case "global":
		authScope = auth.AuthScopeGlobal
	case "project":
		authScope = auth.AuthScopeProject
	default:
		return fmt.Errorf("invalid scope: %s. Valid scopes: global, project", scope)
	}

	// Test authentication
	ctx := cmd.Context()
	if err := authManager.TestAuth(ctx, git.ProviderGitHub, authScope); err != nil {
		return fmt.Errorf("GitHub authentication test failed: %w", err)
	}

	cmd.Printf("✅ GitHub authentication test passed!\n")
	return nil
}

// configureGitLab handles GitLab configuration
func (app *App) configureGitLab(cmd *cobra.Command) error {
	cmd.Println("🦊 GitLab Configuration")
	cmd.Println("=======================")
	cmd.Println()
	cmd.Println("GitLab authentication methods:")
	cmd.Println("  1. Personal Access Token (recommended)")
	cmd.Println("  2. SSH Key")
	cmd.Println("  3. Basic Authentication")
	cmd.Println()
	cmd.Println("Commands:")
	cmd.Println("  cw git provider gitlab login   - Authenticate with GitLab")
	cmd.Println("  cw git provider gitlab test    - Test GitLab authentication")
	cmd.Println()
	cmd.Println("For help getting a GitLab token, see:")
	cmd.Println("  https://gitlab.com/-/profile/personal_access_tokens")

	return nil
}

// loginGitLab handles GitLab authentication
func (app *App) loginGitLab(cmd *cobra.Command) error {
	method, _ := cmd.Flags().GetString("method")
	token, _ := cmd.Flags().GetString("token")
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")
	scope, _ := cmd.Flags().GetString("scope")
	keyFile, _ := cmd.Flags().GetString("key-file")

	cmd.Printf("🔐 Logging into GitLab...\n")
	cmd.Printf("Method: %s\n", method)
	cmd.Printf("Scope: %s\n", scope)

	// Create auth manager
	authManager, err := auth.NewManager(app.configManager)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}

	// Convert scope string to auth.AuthScope
	var authScope auth.AuthScope
	switch scope {
	case "global":
		authScope = auth.AuthScopeGlobal
	case "project":
		authScope = auth.AuthScopeProject
	default:
		return fmt.Errorf("invalid scope: %s. Valid scopes: global, project", scope)
	}

	switch method {
	case "token":
		if token == "" {
			cmd.Printf("Enter your GitLab token: ")
			fmt.Scanln(&token)
			if token == "" {
				return fmt.Errorf("token is required for GitLab authentication")
			}
		}

		if err := authManager.SetToken(git.ProviderGitLab, token, authScope); err != nil {
			return fmt.Errorf("failed to store token: %w", err)
		}

		cmd.Printf("✅ Token stored successfully for GitLab\n")

	case "basic":
		if username == "" {
			cmd.Printf("Enter your GitLab username: ")
			fmt.Scanln(&username)
			if username == "" {
				return fmt.Errorf("username is required for GitLab basic authentication")
			}
		}

		if password == "" {
			cmd.Printf("Enter your GitLab password/token: ")
			fmt.Scanln(&password)
			if password == "" {
				return fmt.Errorf("password is required for GitLab basic authentication")
			}
		}

		if err := authManager.SetBasicAuth(git.ProviderGitLab, username, password, authScope); err != nil {
			return fmt.Errorf("failed to store credentials: %w", err)
		}

		cmd.Printf("✅ Credentials stored successfully for GitLab\n")

	case "ssh":
		if keyFile == "" {
			keyFile = "~/.ssh/id_rsa"
		}

		cmd.Printf("SSH key authentication for GitLab\n")
		cmd.Printf("Key file: %s\n", keyFile)
		cmd.Printf("Note: SSH keys are managed through git configuration\n")
		cmd.Printf("Make sure your SSH key is added to your GitLab account\n")

	default:
		return fmt.Errorf("unsupported authentication method: %s. Supported methods: token, basic, ssh", method)
	}

	// Test the authentication
	ctx := cmd.Context()
	if err := authManager.TestAuth(ctx, git.ProviderGitLab, authScope); err != nil {
		cmd.Printf("⚠️  Warning: Authentication stored but validation failed: %v\n", err)
		cmd.Printf("   You may want to check your credentials\n")
	} else {
		cmd.Printf("🔍 Authentication validated successfully with GitLab API\n")
	}

	return nil
}

// testGitLab tests GitLab authentication
func (app *App) testGitLab(cmd *cobra.Command) error {
	scope, _ := cmd.Flags().GetString("scope")

	cmd.Printf("🧪 Testing GitLab authentication...\n")
	cmd.Printf("Scope: %s\n", scope)

	// Create auth manager
	authManager, err := auth.NewManager(app.configManager)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}

	// Convert scope string to auth.AuthScope
	var authScope auth.AuthScope
	switch scope {
	case "global":
		authScope = auth.AuthScopeGlobal
	case "project":
		authScope = auth.AuthScopeProject
	default:
		return fmt.Errorf("invalid scope: %s. Valid scopes: global, project", scope)
	}

	// Test authentication
	ctx := cmd.Context()
	if err := authManager.TestAuth(ctx, git.ProviderGitLab, authScope); err != nil {
		return fmt.Errorf("GitLab authentication test failed: %w", err)
	}

	cmd.Printf("✅ GitLab authentication test passed!\n")
	return nil
}

// configureBitbucket handles Bitbucket configuration
func (app *App) configureBitbucket(cmd *cobra.Command) error {
	cmd.Println("🪣 Bitbucket Configuration")
	cmd.Println("==========================")
	cmd.Println()
	cmd.Println("Bitbucket authentication methods:")
	cmd.Println("  1. App Password (recommended)")
	cmd.Println("  2. SSH Key")
	cmd.Println("  3. Personal Access Token")
	cmd.Println()
	cmd.Println("Commands:")
	cmd.Println("  cw git provider bitbucket login   - Authenticate with Bitbucket")
	cmd.Println("  cw git provider bitbucket test    - Test Bitbucket authentication")
	cmd.Println()
	cmd.Println("For help getting a Bitbucket app password, see:")
	cmd.Println("  https://bitbucket.org/account/settings/app-passwords/")

	return nil
}

// loginBitbucket handles Bitbucket authentication
func (app *App) loginBitbucket(cmd *cobra.Command) error {
	method, _ := cmd.Flags().GetString("method")
	token, _ := cmd.Flags().GetString("token")
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")
	scope, _ := cmd.Flags().GetString("scope")
	keyFile, _ := cmd.Flags().GetString("key-file")

	cmd.Printf("🔐 Logging into Bitbucket...\n")
	cmd.Printf("Method: %s\n", method)
	cmd.Printf("Scope: %s\n", scope)

	// Create auth manager
	authManager, err := auth.NewManager(app.configManager)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}

	// Convert scope string to auth.AuthScope
	var authScope auth.AuthScope
	switch scope {
	case "global":
		authScope = auth.AuthScopeGlobal
	case "project":
		authScope = auth.AuthScopeProject
	default:
		return fmt.Errorf("invalid scope: %s. Valid scopes: global, project", scope)
	}

	switch method {
	case "token":
		if token == "" {
			cmd.Printf("Enter your Bitbucket token: ")
			fmt.Scanln(&token)
			if token == "" {
				return fmt.Errorf("token is required for Bitbucket authentication")
			}
		}

		if err := authManager.SetToken(git.ProviderBitbucket, token, authScope); err != nil {
			return fmt.Errorf("failed to store token: %w", err)
		}

		cmd.Printf("✅ Token stored successfully for Bitbucket\n")

	case "basic":
		if username == "" {
			cmd.Printf("Enter your Bitbucket username: ")
			fmt.Scanln(&username)
			if username == "" {
				return fmt.Errorf("username is required for Bitbucket basic authentication")
			}
		}

		if password == "" {
			cmd.Printf("Enter your Bitbucket app password: ")
			fmt.Scanln(&password)
			if password == "" {
				return fmt.Errorf("app password is required for Bitbucket basic authentication")
			}
		}

		if err := authManager.SetBasicAuth(git.ProviderBitbucket, username, password, authScope); err != nil {
			return fmt.Errorf("failed to store credentials: %w", err)
		}

		cmd.Printf("✅ Credentials stored successfully for Bitbucket\n")

	case "ssh":
		if keyFile == "" {
			keyFile = "~/.ssh/id_rsa"
		}

		cmd.Printf("SSH key authentication for Bitbucket\n")
		cmd.Printf("Key file: %s\n", keyFile)
		cmd.Printf("Note: SSH keys are managed through git configuration\n")
		cmd.Printf("Make sure your SSH key is added to your Bitbucket account\n")

	default:
		return fmt.Errorf("unsupported authentication method: %s. Supported methods: token, basic, ssh", method)
	}

	// Test the authentication
	ctx := cmd.Context()
	if err := authManager.TestAuth(ctx, git.ProviderBitbucket, authScope); err != nil {
		cmd.Printf("⚠️  Warning: Authentication stored but validation failed: %v\n", err)
		cmd.Printf("   You may want to check your credentials\n")
	} else {
		cmd.Printf("🔍 Authentication validated successfully with Bitbucket API\n")
	}

	return nil
}

// testBitbucket tests Bitbucket authentication
func (app *App) testBitbucket(cmd *cobra.Command) error {
	scope, _ := cmd.Flags().GetString("scope")

	cmd.Printf("🧪 Testing Bitbucket authentication...\n")
	cmd.Printf("Scope: %s\n", scope)

	// Create auth manager
	authManager, err := auth.NewManager(app.configManager)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}

	// Convert scope string to auth.AuthScope
	var authScope auth.AuthScope
	switch scope {
	case "global":
		authScope = auth.AuthScopeGlobal
	case "project":
		authScope = auth.AuthScopeProject
	default:
		return fmt.Errorf("invalid scope: %s. Valid scopes: global, project", scope)
	}

	// Test authentication
	ctx := cmd.Context()
	if err := authManager.TestAuth(ctx, git.ProviderBitbucket, authScope); err != nil {
		return fmt.Errorf("Bitbucket authentication test failed: %w", err)
	}

	cmd.Printf("✅ Bitbucket authentication test passed!\n")
	return nil
}
