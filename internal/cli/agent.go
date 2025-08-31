package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func addAgentCommands(app *App) *cobra.Command {
	agentCmd := &cobra.Command{
		Use:   "agent",
		Short: "Manage agent configuration(s)",
		Long:  "Manage agent configuration(s). Agents are the CLI coding agents and their individual setups.",
	}

	// Add env commands
	envCmd := addEnvCommands(app)
	agentCmd.AddCommand(envCmd)

	return agentCmd
}

func addEnvCommands(app *App) *cobra.Command {
	envCmd := &cobra.Command{
		Use:   "env",
		Short: "Manage agent environment variables",
		Long:  "Manage agent environment variables",
	}

	// Set command
	setCmd := &cobra.Command{
		Use:   "set <key> <value> OR <key>=<value>",
		Short: "Set an agent environment variable",
		Long:  "Set an agent environment variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("invalid number of arguments - you must specify the key and value either by 'set key value' or 'set key=value'")
			} else if len(args) == 1 {
				// Split the argument by "=" if possible
				keyValue := args[0]
				if idx := strings.Index(keyValue, "="); idx != -1 {
					key := keyValue[:idx]
					value := keyValue[idx+1:]
					return app.configManager.SetEnvVar(key, value)
				} else {
					return fmt.Errorf("please use 'set key=value' format or provide both key and value as separate arguments")
				}
			} else if len(args) == 2 {
				key := args[0]
				value := args[1]
				return app.configManager.SetEnvVar(key, value)
			} else {
				return fmt.Errorf("invalid number of arguments. Use 'set key value' or 'set key=value'")
			}
		},
	}
	envCmd.AddCommand(setCmd)

	// Get command
	getCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get an agent environment variable",
		Long:  "Get an agent environment variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value, err := app.configManager.GetEnvVar(key)
			if err != nil {
				return err
			}
			cmd.Printf("%s=%s\n", key, value)
			return nil
		},
	}
	envCmd.AddCommand(getCmd)

	// List command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all agent environment variables",
		Long:  "List all agent environment variables",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("invalid number of arguments. Use 'list' or 'list <key>'")
			}
			env, err := app.configManager.GetEnvVars()
			if err != nil {
				return err
			}
			for key, value := range env {
				cmd.Printf("%s=%s\n", key, value)
			}
			return nil
		},
	}
	envCmd.AddCommand(listCmd)

	// Delete command
	deleteCmd := &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete an agent environment variable",
		Long:  "Delete an agent environment variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.configManager.DeleteEnvVar(args[0])
		},
	}
	envCmd.AddCommand(deleteCmd)

	// Import command
	importCmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import agent environment variables from a file",
		Long:  "Import agent environment variables from a file. Follows typical .env file format - key=value per line.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.loadEnvFile(cmd, args[0])
		},
	}
	envCmd.AddCommand(importCmd)

	return envCmd
}
