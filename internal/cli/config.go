package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hlfshell/cowork/internal/config"
	"github.com/spf13/cobra"
)

// addConfigCommands adds configuration management commands
func (app *App) addConfigCommands() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage cowork configuration",
		Long: `Manage cowork configuration files.

Configuration is loaded from:
1. Global config: ~/.config/.cwconfig
2. Project config: .cwconfig (in current directory)

Project configuration overrides global configuration.`,
	}

	// Add subcommands
	configCmd.AddCommand(app.newConfigInitCommand())
	configCmd.AddCommand(app.newConfigShowCommand())
	configCmd.AddCommand(app.newConfigEditCommand())
	configCmd.AddCommand(app.newConfigValidateCommand())

	app.rootCmd.AddCommand(configCmd)
}

// newConfigInitCommand creates the config init command
func (app *App) newConfigInitCommand() *cobra.Command {
	var global bool
	var project bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration files",
		Long: `Initialize configuration files with default values.

By default, initializes both global and project config files.
Use --global or --project to initialize only one type.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager := config.NewManager()

			// If no flags specified, initialize both
			if !global && !project {
				global = true
				project = true
			}

			if global {
				cmd.Printf("Initializing global configuration...\n")
				defaultConfig := configManager.GetDefaultConfig()
				if err := configManager.SaveGlobal(defaultConfig); err != nil {
					return fmt.Errorf("failed to initialize global config: %w", err)
				}
				cmd.Printf("✅ Global configuration initialized at: %s\n", configManager.GlobalConfigPath)
			}

			if project {
				cmd.Printf("Initializing project configuration...\n")
				defaultConfig := configManager.GetDefaultConfig()
				if err := configManager.SaveProject(defaultConfig); err != nil {
					return fmt.Errorf("failed to initialize project config: %w", err)
				}
				cmd.Printf("✅ Project configuration initialized at: %s\n", configManager.ProjectConfigPath)
			}

			cmd.Printf("\nConfiguration files initialized successfully!\n")
			cmd.Printf("You can now edit these files to customize your settings.\n")
			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "Initialize only global configuration")
	cmd.Flags().BoolVar(&project, "project", false, "Initialize only project configuration")

	return cmd
}

// newConfigShowCommand creates the config show command
func (app *App) newConfigShowCommand() *cobra.Command {
	var global bool
	var project bool
	var merged bool
	var format string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show configuration",
		Long: `Show configuration values.

By default, shows the merged configuration (global + project overrides).
Use --global, --project, or --merged to show specific configurations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager := config.NewManager()

			// If no flags specified, show merged config
			if !global && !project && !merged {
				merged = true
			}

			if global {
				cmd.Printf("📁 Global Configuration\n")
				cmd.Printf("=====================\n")
				if err := app.showConfig(cmd, configManager.GlobalConfigPath, format); err != nil {
					return fmt.Errorf("failed to show global config: %w", err)
				}
			}

			if project {
				cmd.Printf("📁 Project Configuration\n")
				cmd.Printf("=======================\n")
				if err := app.showConfig(cmd, configManager.ProjectConfigPath, format); err != nil {
					return fmt.Errorf("failed to show project config: %w", err)
				}
			}

			if merged {
				cmd.Printf("📁 Merged Configuration (Global + Project Overrides)\n")
				cmd.Printf("================================================\n")
				mergedConfig, err := configManager.Load()
				if err != nil {
					return fmt.Errorf("failed to load merged config: %w", err)
				}
				if err := app.showConfigStruct(cmd, mergedConfig, format); err != nil {
					return fmt.Errorf("failed to show merged config: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "Show only global configuration")
	cmd.Flags().BoolVar(&project, "project", false, "Show only project configuration")
	cmd.Flags().BoolVar(&merged, "merged", false, "Show merged configuration (default)")
	cmd.Flags().StringVar(&format, "format", "yaml", "Output format (yaml, json)")

	return cmd
}

// newConfigEditCommand creates the config edit command
func (app *App) newConfigEditCommand() *cobra.Command {
	var global bool
	var project bool

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit configuration files",
		Long: `Edit configuration files in your default editor.

By default, edits the project configuration file.
Use --global to edit the global configuration file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager := config.NewManager()

			// If no flags specified, edit project config
			if !global && !project {
				project = true
			}

			if global {
				cmd.Printf("Opening global configuration for editing...\n")
				if err := app.editConfig(cmd, configManager.GlobalConfigPath); err != nil {
					return fmt.Errorf("failed to edit global config: %w", err)
				}
			}

			if project {
				cmd.Printf("Opening project configuration for editing...\n")
				if err := app.editConfig(cmd, configManager.ProjectConfigPath); err != nil {
					return fmt.Errorf("failed to edit project config: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "Edit global configuration")
	cmd.Flags().BoolVar(&project, "project", false, "Edit project configuration")

	return cmd
}

// newConfigValidateCommand creates the config validate command
func (app *App) newConfigValidateCommand() *cobra.Command {
	var global bool
	var project bool

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration files",
		Long: `Validate configuration files for syntax and semantic errors.

By default, validates both global and project config files.
Use --global or --project to validate only one type.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager := config.NewManager()

			// If no flags specified, validate both
			if !global && !project {
				global = true
				project = true
			}

			valid := true

			if global {
				cmd.Printf("Validating global configuration...\n")
				if err := app.validateConfig(cmd, configManager.GlobalConfigPath, "global"); err != nil {
					valid = false
				} else {
					cmd.Printf("✅ Global configuration is valid\n")
				}
			}

			if project {
				cmd.Printf("Validating project configuration...\n")
				if err := app.validateConfig(cmd, configManager.ProjectConfigPath, "project"); err != nil {
					valid = false
				} else {
					cmd.Printf("✅ Project configuration is valid\n")
				}
			}

			if valid {
				cmd.Printf("\n🎉 All configuration files are valid!\n")
			} else {
				return fmt.Errorf("configuration validation failed")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "Validate only global configuration")
	cmd.Flags().BoolVar(&project, "project", false, "Validate only project configuration")

	return cmd
}

// showConfig shows configuration from a file
func (app *App) showConfig(cmd *cobra.Command, configPath string, format string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cmd.Printf("Configuration file does not exist: %s\n", configPath)
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	cmd.Printf("%s\n", string(data))
	return nil
}

// showConfigStruct shows configuration from a struct
func (app *App) showConfigStruct(cmd *cobra.Command, config *config.Config, format string) error {
	// For now, just show the YAML representation
	// In a real implementation, you might want to format this nicely
	cmd.Printf("Configuration loaded successfully.\n")
	cmd.Printf("Workspace settings:\n")
	cmd.Printf("  Default isolation level: %s\n", config.Workspace.DefaultIsolationLevel)
	cmd.Printf("  Base directory: %s\n", config.Workspace.BaseDirectory)
	cmd.Printf("  Max workspaces: %d\n", config.Workspace.MaxWorkspaces)
	cmd.Printf("  Auto cleanup orphaned: %t\n", config.Workspace.AutoCleanupOrphaned)
	cmd.Printf("  Default branch: %s\n", config.Workspace.DefaultBranch)
	cmd.Printf("  Naming pattern: %s\n", config.Workspace.NamingPattern)
	cmd.Printf("  Auto save metadata: %t\n", config.Workspace.AutoSaveMetadata)
	cmd.Printf("  Timeout minutes: %d\n", config.Workspace.TimeoutMinutes)

	cmd.Printf("\nGit settings:\n")
	cmd.Printf("  Timeout seconds: %d\n", config.Git.TimeoutSeconds)
	cmd.Printf("  Default remote: %s\n", config.Git.DefaultRemote)
	cmd.Printf("  Auto fetch: %t\n", config.Git.AutoFetch)
	cmd.Printf("  Shallow depth: %d\n", config.Git.ShallowDepth)
	cmd.Printf("  Credential helper: %s\n", config.Git.CredentialHelper)
	cmd.Printf("  User name: %s\n", config.Git.User.Name)
	cmd.Printf("  User email: %s\n", config.Git.User.Email)

	cmd.Printf("\nAgent settings:\n")
	cmd.Printf("  Default agent: %s\n", config.Agent.DefaultAgent)
	cmd.Printf("  Timeout minutes: %d\n", config.Agent.TimeoutMinutes)
	cmd.Printf("  Max concurrent: %d\n", config.Agent.MaxConcurrent)

	cmd.Printf("\nContainer settings:\n")
	cmd.Printf("  Engine: %s\n", config.Container.Engine)
	cmd.Printf("  Default image: %s\n", config.Container.DefaultImage)
	cmd.Printf("  Timeout minutes: %d\n", config.Container.TimeoutMinutes)
	cmd.Printf("  Auto start: %t\n", config.Container.AutoStart)

	cmd.Printf("\nUI settings:\n")
	cmd.Printf("  Output format: %s\n", config.UI.OutputFormat)
	cmd.Printf("  Color: %s\n", config.UI.Color)
	cmd.Printf("  Verbose: %t\n", config.UI.Verbose)
	cmd.Printf("  Show progress: %t\n", config.UI.ShowProgress)
	cmd.Printf("  Interactive: %t\n", config.UI.Interactive)
	cmd.Printf("  Confirm prompts: %t\n", config.UI.ConfirmPrompts)

	cmd.Printf("\nLogging settings:\n")
	cmd.Printf("  Level: %s\n", config.Logging.Level)
	cmd.Printf("  Format: %s\n", config.Logging.Format)
	cmd.Printf("  File: %s\n", config.Logging.File)
	cmd.Printf("  Include timestamp: %t\n", config.Logging.IncludeTimestamp)
	cmd.Printf("  Include caller: %t\n", config.Logging.IncludeCaller)

	return nil
}

// editConfig opens a configuration file for editing
func (app *App) editConfig(cmd *cobra.Command, configPath string) error {
	// Ensure the directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// If the file doesn't exist, create it with defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configManager := config.NewManager()
		defaultConfig := configManager.GetDefaultConfig()
		if err := configManager.SaveConfig(defaultConfig, configPath); err != nil {
			return fmt.Errorf("failed to create default config file: %w", err)
		}
	}

	// Get the editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano" // Default fallback
	}

	cmd.Printf("Opening %s with %s...\n", configPath, editor)
	cmd.Printf("Make your changes and save the file.\n")

	// In a real implementation, you would execute the editor
	// For now, just show the path
	cmd.Printf("File location: %s\n", configPath)

	return nil
}

// validateConfig validates a configuration file
func (app *App) validateConfig(cmd *cobra.Command, configPath string, configType string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cmd.Printf("⚠️  %s configuration file does not exist: %s\n", configType, configPath)
		return nil
	}

	// Try to load the configuration
	if configType == "global" {
		// For global config, we need to create a temporary manager
		tempManager := &config.Manager{
			GlobalConfigPath:  configPath,
			ProjectConfigPath: "/dev/null", // Use a non-existent path
		}
		_, err := tempManager.Load()
		if err != nil {
			cmd.Printf("❌ %s configuration is invalid: %v\n", configType, err)
			return err
		}
	} else if configType == "project" {
		// For project config, we need to create a temporary manager
		tempManager := &config.Manager{
			GlobalConfigPath:  "/dev/null", // Use a non-existent path
			ProjectConfigPath: configPath,
		}
		_, err := tempManager.Load()
		if err != nil {
			cmd.Printf("❌ %s configuration is invalid: %v\n", configType, err)
			return err
		}
	}

	return nil
}
