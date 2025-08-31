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
				// return app.loginProvider(cmd, args[0])
				return nil
			case "test":
				// return app.testProvider(cmd, args[0])
				return nil
			default:
				return fmt.Errorf("unknown action: %s. Valid actions: login, test", args[1])
			}
		},
	}
	configCmd.AddCommand(providerCmd)

	// Add flags for provider command (these will be used for login and test)
	// providerCmd.Flags().String("method", "token", "Authentication method (token, basic)")
	// providerCmd.Flags().String("token", "", "API token for token authentication")
	// providerCmd.Flags().String("username", "", "Username for basic authentication")
	// providerCmd.Flags().String("password", "", "Password for basic authentication")
	// providerCmd.Flags().String("scope", "global", "Scope for authentication (global, project)")

	// Agent command
	// agentCmd := &cobra.Command{
	// 	Use:   "agent",
	// 	Short: "Configure AI agent settings",
	// 	Long:  "Configure settings for AI agents (docker image, command structure, keys, etc.)",
	// 	RunE: func(cmd *cobra.Command, args []string) error {
	// 		return app.configureAgent(cmd)
	// 	},
	// }

	// Add git commands
	gitCmd := addGitCommands(app)
	configCmd.AddCommand(gitCmd)

	// Add agent env commands
	agentCmd := addAgentCommands(app)
	configCmd.AddCommand(agentCmd)

	// // Save command
	saveCmd := &cobra.Command{
		Use:   "save [filename]",
		Short: "Save configuration to YAML file",
		Long:  "Save current local configuration settings to a YAML file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.saveConfig(cmd, args[0])
		},
	}

	// // Load command
	loadCmd := &cobra.Command{
		Use:   "load [filename]",
		Short: "Load configuration from YAML file",
		Long:  "Load configuration settings from a YAML file into local settings",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.loadConfig(cmd, args[0])
		},
	}

	configCmd.AddCommand(showCmd, saveCmd, loadCmd)
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

// saveConfig saves the current configuration settings to a YAML file
func (app *App) saveConfig(cmd *cobra.Command, filename string) error {
	app.configManager.Load()
	app.configManager.SaveToFile(filename)
	return nil
}

// loadConfig loads the configuration settings from a YAML file
func (app *App) loadConfig(cmd *cobra.Command, filename string) error {
	app.configManager.LoadFromFile(filename)
	return nil
}

// setEnvVar sets an environment variable
func (app *App) setEnvVar(cmd *cobra.Command, args []string) error {
	// Parse the key-value pair
	var key, value string

	if len(args) == 1 {
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

	err = app.configManager.SetEnvVar(key, value)
	if err != nil {
		return fmt.Errorf("failed to set environment variable: %w", err)
	}

	cmd.Printf("✅ Environment variable '%s' set successfully\n", key)

	return nil
}

// getEnvVar retrieves an environment variable value
func (app *App) getEnvVar(cmd *cobra.Command, key string) error {
	if key == "" {
		return fmt.Errorf("environment variable key cannot be empty")
	}

	// Load configuration first to ensure it's initialized
	_, err := app.configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	value, err := app.configManager.GetEnvVar(key)
	if err != nil {
		return fmt.Errorf("failed to get environment variable: %w", err)
	}
	cmd.Printf("%s=%s\n", key, value)
	return nil

}

func (app *App) listEnvVars(cmd *cobra.Command) error {
	// Load configuration first to ensure it's initialized
	_, err := app.configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	envVars, err := app.configManager.GetEnvVars()
	if err != nil {
		return fmt.Errorf("failed to get environment variables: %w", err)
	}

	if len(envVars) == 0 {
		cmd.Println("No environment variables found.")
		return nil
	}

	cmd.Println("Environment variables:")
	for key, value := range envVars {
		cmd.Printf("  %s=%s\n", key, value)
	}
	return nil
}

// loadEnvFile loads environment variables from a file
func (app *App) loadEnvFile(cmd *cobra.Command, filename string) error {
	// Load configuration first to ensure it's initialized
	_, err := app.configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	err = app.configManager.SetEnvFromFile(filename)
	if err != nil {
		return fmt.Errorf("failed to load environment variables from file: %w", err)
	}

	cmd.Printf("✅ Environment variables loaded from %s\n", filename)
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
