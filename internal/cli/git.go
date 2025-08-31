package cli

import (
	"fmt"
	"os"

	"github.com/hlfshell/cowork/internal/auth"
	"github.com/spf13/cobra"
)

// addGitCommands adds git configuration and authentication commands
func addGitCommands(app *App) *cobra.Command {
	gitCmd := &cobra.Command{
		Use:   "git",
		Short: "Manage git configuration and authentication",
		Long:  "Configure git settings, authentication methods, and manage git provider credentials",
	}

	// Add auth command
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage git authentication",
		Long:  "Configure authentication for git operations including SSH keys, HTTPS tokens, and provider credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.manageGitAuth(cmd, args)
		},
	}
	authCmd.Flags().String("ssh", "", "SSH key for git authentication")
	authCmd.Flags().String("ssh-key", "", "SSH key filepath for git authentication")
	authCmd.Flags().String("token", "", "Token for git authentication")
	authCmd.Flags().String("username", "", "Username for git authentication")
	authCmd.Flags().String("password", "", "Password for git authentication")
	authCmd.Flags().String("scope", "global", "Scope for git authentication (global, project)")

	// Add provider commands
	providerCmd := addGitProviderCommands(app)

	gitCmd.AddCommand(authCmd, providerCmd)

	return gitCmd
}

// addGitProviderCommands adds provider-specific git commands
func addGitProviderCommands(app *App) *cobra.Command {
	providerCmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage git provider authentication",
		Long:  "Configure authentication for specific git providers (GitHub, GitLab, Bitbucket)",
	}

	// GitHub command
	githubCmd := &cobra.Command{
		Use:   "github [action]",
		Short: "Manage GitHub authentication",
		Long:  "Configure authentication for GitHub using tokens or SSH keys",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return app.configureGitHub(cmd)
			}
			switch args[0] {
			case "login":
				return app.loginGitHub(cmd)
			case "test":
				return app.testGitHub(cmd)
			default:
				return fmt.Errorf("unknown action: %s. Valid actions: login, test", args[0])
			}
		},
	}

	// GitLab command
	gitlabCmd := &cobra.Command{
		Use:   "gitlab [action]",
		Short: "Manage GitLab authentication",
		Long:  "Configure authentication for GitLab using tokens or SSH keys",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return app.configureGitLab(cmd)
			}
			switch args[0] {
			case "login":
				return app.loginGitLab(cmd)
			case "test":
				return app.testGitLab(cmd)
			default:
				return fmt.Errorf("unknown action: %s. Valid actions: login, test", args[0])
			}
		},
	}

	// Bitbucket command
	bitbucketCmd := &cobra.Command{
		Use:   "bitbucket [action]",
		Short: "Manage Bitbucket authentication",
		Long:  "Configure authentication for Bitbucket using tokens or SSH keys",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return app.configureBitbucket(cmd)
			}
			switch args[0] {
			case "login":
				return app.loginBitbucket(cmd)
			case "test":
				return app.testBitbucket(cmd)
			default:
				return fmt.Errorf("unknown action: %s. Valid actions: login, test", args[0])
			}
		},
	}

	// Add flags for provider commands
	for _, cmd := range []*cobra.Command{githubCmd, gitlabCmd, bitbucketCmd} {
		cmd.Flags().String("method", "token", "Authentication method (token, basic, ssh)")
		cmd.Flags().String("token", "", "API token for token authentication")
		cmd.Flags().String("username", "", "Username for basic authentication")
		cmd.Flags().String("password", "", "Password for basic authentication")
		cmd.Flags().String("scope", "global", "Scope for authentication (global, project)")
		cmd.Flags().String("key-file", "", "SSH key file path for SSH authentication")
	}

	providerCmd.AddCommand(githubCmd, gitlabCmd, bitbucketCmd)

	return providerCmd
}

func (app *App) showGitAuthHelp(cmd *cobra.Command) {
	cmd.Println("🔐 Git Authentication Management")
	cmd.Println("===============================")
	cmd.Println()
	cmd.Println("Available authentication methods:")
	cmd.Println("  1. SSH Keys (recommended for git operations)")
	cmd.Println("\t cowork config git auth --ssh <key> OR --ssh-key <path to key>")
	cmd.Println("  2. API Tokens (for provider integration)")
	cmd.Println("\t cowork config git auth --token <token>")
	cmd.Println("  3. Basic Authentication (username/password)")
	cmd.Println("\t cowork config git auth --username <username> --password <password>")
	cmd.Println()
}

// manageGitAuth handles the main git authentication management
func (app *App) manageGitAuth(cmd *cobra.Command, args []string) error {
	ssh, _ := cmd.Flags().GetString("ssh")
	sshKey, _ := cmd.Flags().GetString("ssh-key")
	token, _ := cmd.Flags().GetString("token")
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")
	targetScope, _ := cmd.Flags().GetString("scope")

	var scope auth.AuthScope
	if targetScope == "" {
		scope = auth.AuthScopeProject
	} else {
		scope = auth.AuthScope(targetScope)
	}

	if ssh == "" && sshKey == "" && token == "" && username == "" && password == "" {
		app.showGitAuthHelp(cmd)
		return nil
	} else if ssh != "" {
		return app.setGitSSHKey(cmd, ssh, scope)
	} else if sshKey != "" {
		data, err := os.ReadFile(sshKey)
		if err != nil {
			return fmt.Errorf("failed to read SSH key file: %w", err)
		}
		return app.setGitSSHKey(cmd, string(data), scope)
	} else if token != "" {
		return app.setGitToken(cmd, token, scope)
	} else if username != "" && password != "" {
		return app.setGitBasicAuth(cmd, username, password, scope)
	}

	return nil

}

func (app *App) setGitSSHKey(cmd *cobra.Command, ssh string, scope auth.AuthScope) error {
	authManager, err := auth.NewManager(app.configManager)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}
	authManager.SetGitSSHKeyAuth(ssh, scope)
	return nil
}

func (app *App) setGitToken(cmd *cobra.Command, token string, scope auth.AuthScope) error {
	authManager, err := auth.NewManager(app.configManager)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}
	authManager.SetGitTokenAuth(token, scope)
	return nil
}

func (app *App) setGitBasicAuth(cmd *cobra.Command, username, password string, scope auth.AuthScope) error {
	authManager, err := auth.NewManager(app.configManager)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}
	authManager.SetGitBasicAuth(username, password, scope)
	return nil
}
