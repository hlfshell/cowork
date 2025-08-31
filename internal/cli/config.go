package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hlfshell/cowork/internal/config"
	"github.com/spf13/cobra"
)

// addConfigCommands adds configuration management commands
func (app *App) addConfigCommands() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration settings",
		Long:  "Configure global and local cw setups. All settings will follow the priority of direct comand flags > local project .cw configs > global .cw configs",
	}

	// Show command
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration settings",
		Long:  "Display the current configuration settings in a human readable manner. Never show keys.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.showConfig(cmd)
		},
	}

	// // Auth command
	// authCmd := &cobra.Command{
	// 	Use:   "auth",
	// 	Short: "Manage authentication settings",
	// 	Long:  "Handle authentication for git providers (GitHub, GitLab, Bitbucket) and API keys for agents",
	// 	RunE: func(cmd *cobra.Command, args []string) error {
	// 		return app.manageAuth(cmd)
	// 	},
	// }

	// Provider command
	providerCmd := &cobra.Command{
		Use:   "provider [provider-name] [action]",
		Short: "Manage provider authentication",
		Long:  "Configure authentication for specific providers (GitHub, GitLab, Bitbucket)",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return app.listProviders(cmd)
			}
			if len(args) == 1 {
				return app.showProviderHelp(cmd, args[0])
			}
			// args[0] is provider name, args[1] is action
			switch args[1] {
			case "login":
				return app.loginProvider(cmd, args[0])
			case "test":
				return app.testProvider(cmd, args[0])
			default:
				return fmt.Errorf("unknown action: %s. Valid actions: login, test", args[1])
			}
		},
	}
	configCmd.AddCommand(providerCmd)

	// Add git commands
	gitCmd := addGitCommands(app)
	configCmd.AddCommand(gitCmd)

	// Add container commands
	containerCmd := addContainerCommands(app)
	configCmd.AddCommand(containerCmd)

	// Add agent env commands
	agentCmd := addAgentCommands(app)
	configCmd.AddCommand(agentCmd)

	configCmd.AddCommand(showCmd)
	app.rootCmd.AddCommand(configCmd)
}

// showConfig shows the current configuration settings
func (app *App) showConfig(cmd *cobra.Command) error {
	cmd.Println("🔍 Showing configuration settings")
	cmd.Println("============================")

	app.configManager.Load()
	cmd.Println(app.configManager.HumanReadable())

	return nil
}

// setEnvVar sets an environment variable with scope support
func (app *App) setEnvVar(cmd *cobra.Command, args []string) error {
	// Parse the key-value pair
	var key, value string

	if len(args) == 0 {
		return fmt.Errorf("invalid number of arguments - you must specify the key and value either by 'set key value' or 'set key=value'")
	} else if len(args) == 1 {
		// Single argument: check if it's in key=value format
		keyValue := args[0]
		if idx := strings.Index(keyValue, "="); idx != -1 {
			key = keyValue[:idx]
			value = keyValue[idx+1:]
		} else {
			return fmt.Errorf("please use 'set key=value' format or provide both key and value as separate arguments")
		}
	} else if len(args) == 2 {
		// Two arguments: key and value separately
		key = args[0]
		value = args[1]
	} else {
		return fmt.Errorf("invalid number of arguments. Use 'set key=value' or 'set key value'")
	}

	if key == "" {
		return fmt.Errorf("environment variable key cannot be empty")
	}

	// Load configuration first to ensure it's initialized
	_, err := app.configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine scope
	scopeFlag, _ := cmd.Flags().GetString("scope")
	if scopeFlag == "global" {
		err = app.configManager.SetEnvVarScoped(key, value, config.EnvScopeGlobal)
	} else {
		err = app.configManager.SetEnvVarScoped(key, value, config.EnvScopeProject)
	}
	if err != nil {
		return fmt.Errorf("failed to set environment variable: %w", err)
	}

	cmd.Printf("✅ Environment variable '%s' set successfully (%s scope)\n", key, scopeFlag)

	return nil
}

// getEnvVar retrieves an environment variable value with optional scope filtering
func (app *App) getEnvVar(cmd *cobra.Command, key string) error {
	if key == "" {
		return fmt.Errorf("environment variable key cannot be empty")
	}

	// Load configuration first to ensure it's initialized
	_, err := app.configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	scopeFlag, _ := cmd.Flags().GetString("scope")

	if scopeFlag == "" {
		// No scope specified - check both and show which scope was used
		value, scope, err := app.configManager.GetEnvVarWithFallback(key)
		if err != nil {
			return fmt.Errorf("environment variable '%s' not found in any scope", key)
		}
		cmd.Printf("%s=%s (%s scope)\n", key, value, scope)
	} else {
		// Specific scope requested
		var value string
		switch scopeFlag {
		case "project":
			value, err = app.configManager.GetEnvVarScoped(key, config.EnvScopeProject)
			if err != nil {
				return fmt.Errorf("environment variable '%s' not found in project scope", key)
			}
			cmd.Printf("%s=%s (project scope)\n", key, value)
		case "global":
			value, err = app.configManager.GetEnvVarScoped(key, config.EnvScopeGlobal)
			if err != nil {
				return fmt.Errorf("environment variable '%s' not found in global scope", key)
			}
			cmd.Printf("%s=%s (global scope)\n", key, value)
		default:
			return fmt.Errorf("invalid scope: %s. Valid scopes: project, global", scopeFlag)
		}
	}

	return nil
}

func (app *App) listEnvVars(cmd *cobra.Command) error {
	// Load configuration first to ensure it's initialized
	_, err := app.configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	scope, _ := cmd.Flags().GetString("scope")

	switch scope {
	case "project":
		envVars, err := app.configManager.GetEnvVarsScoped(config.EnvScopeProject)
		if err != nil {
			return fmt.Errorf("failed to get project environment variables: %w", err)
		}
		if len(envVars) == 0 {
			cmd.Println("No project environment variables found.")
			return nil
		}
		cmd.Println("Project environment variables:")
		for key, value := range envVars {
			cmd.Printf("  %s=%s\n", key, value)
		}
	case "global":
		envVars, err := app.configManager.GetEnvVarsScoped(config.EnvScopeGlobal)
		if err != nil {
			return fmt.Errorf("failed to get global environment variables: %w", err)
		}
		if len(envVars) == 0 {
			cmd.Println("No global environment variables found.")
			return nil
		}
		cmd.Println("Global environment variables:")
		for key, value := range envVars {
			cmd.Printf("  %s=%s\n", key, value)
		}
	case "all":
		// Show both project and global
		projectEnvVars, err := app.configManager.GetEnvVarsScoped(config.EnvScopeProject)
		if err != nil {
			return fmt.Errorf("failed to get project environment variables: %w", err)
		}
		globalEnvVars, err := app.configManager.GetEnvVarsScoped(config.EnvScopeGlobal)
		if err != nil {
			return fmt.Errorf("failed to get global environment variables: %w", err)
		}

		if len(projectEnvVars) == 0 && len(globalEnvVars) == 0 {
			cmd.Println("No environment variables found.")
			return nil
		}

		if len(projectEnvVars) > 0 {
			cmd.Println("Project environment variables:")
			for key, value := range projectEnvVars {
				cmd.Printf("  %s=%s\n", key, value)
			}
		}

		if len(globalEnvVars) > 0 {
			if len(projectEnvVars) > 0 {
				cmd.Println()
			}
			cmd.Println("Global environment variables:")
			for key, value := range globalEnvVars {
				cmd.Printf("  %s=%s\n", key, value)
			}
		}
	default:
		return fmt.Errorf("invalid scope: %s. Valid scopes: project, global, all", scope)
	}

	return nil
}

// deleteEnvVar deletes an environment variable with scope support
func (app *App) deleteEnvVar(cmd *cobra.Command, key string) error {
	if key == "" {
		return fmt.Errorf("environment variable key cannot be empty")
	}

	// Load configuration first to ensure it's initialized
	_, err := app.configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	scope, _ := cmd.Flags().GetString("scope")

	switch scope {
	case "project":
		err = app.configManager.DeleteEnvVarScoped(key, config.EnvScopeProject)
		if err != nil {
			return fmt.Errorf("failed to delete environment variable from project scope: %w", err)
		}
		cmd.Printf("✅ Environment variable '%s' deleted from project scope\n", key)
	case "global":
		err = app.configManager.DeleteEnvVarScoped(key, config.EnvScopeGlobal)
		if err != nil {
			return fmt.Errorf("failed to delete environment variable from global scope: %w", err)
		}
		cmd.Printf("✅ Environment variable '%s' deleted from global scope\n", key)
	case "both":
		// Try to delete from both scopes, don't error if not found in one
		projectErr := app.configManager.DeleteEnvVarScoped(key, config.EnvScopeProject)
		globalErr := app.configManager.DeleteEnvVarScoped(key, config.EnvScopeGlobal)

		if projectErr != nil && globalErr != nil {
			return fmt.Errorf("failed to delete environment variable from both scopes: project: %v, global: %v", projectErr, globalErr)
		}

		if projectErr == nil && globalErr == nil {
			cmd.Printf("✅ Environment variable '%s' deleted from both scopes\n", key)
		} else if projectErr == nil {
			cmd.Printf("✅ Environment variable '%s' deleted from project scope (not found in global)\n", key)
		} else {
			cmd.Printf("✅ Environment variable '%s' deleted from global scope (not found in project)\n", key)
		}
	default:
		return fmt.Errorf("invalid scope: %s. Valid scopes: project, global, both", scope)
	}

	return nil
}

// loadEnvFile loads environment variables from a file with scope support
func (app *App) loadEnvFile(cmd *cobra.Command, filename string) error {
	// Load configuration first to ensure it's initialized
	_, err := app.configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine scope
	scopeFlag, _ := cmd.Flags().GetString("scope")

	// TODO: Need to implement SetEnvFromFileScoped in config manager
	// For now, use the existing method which only supports project scope
	if scopeFlag == "global" {
		return fmt.Errorf("global scope for file import not yet implemented")
	}

	err = app.configManager.SetEnvFromFile(filename)
	if err != nil {
		return fmt.Errorf("failed to load environment variables from file: %w", err)
	}

	cmd.Printf("✅ Environment variables loaded from %s (%s scope)\n", filename, scopeFlag)
	return nil
}

// manageEnv manages environment variables
func (app *App) manageEnv(cmd *cobra.Command) error {
	fmt.Println("🌍 Environment Variable Management")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("Environment variables are stored encrypted and passed to containers at runtime.")
	fmt.Println("These are local-only (not global) for security reasons.")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  cw config env set <key>=<value>")
	fmt.Println("  cw config env get <key>")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  cw config env set OPENAI_API_KEY=sk-...")
	fmt.Println("  cw config env set ANTHROPIC_API_KEY=sk-ant-...")
	fmt.Println("  cw config env get OPENAI_API_KEY")

	return nil
}

// ensureGlobalConfigExists checks if the global configuration file exists and creates it if it doesn't
func ensureGlobalConfigExists(configManager *config.Manager) error {
	// Check if global config file exists
	if _, err := os.Stat(configManager.GlobalConfigPath); os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		configDir := filepath.Dir(configManager.GlobalConfigPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Create default configuration
		defaultConfig := configManager.GetDefaultConfig()
		if err := configManager.SaveGlobal(defaultConfig); err != nil {
			return fmt.Errorf("failed to create default global configuration: %w", err)
		}

		fmt.Printf("✅ Created default global configuration at: %s\n", configManager.GlobalConfigPath)
	}
	return nil
}
