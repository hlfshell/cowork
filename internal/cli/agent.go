package cli

import (
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
		Long:  "Set an agent environment variable. Defaults to project scope unless --scope global is specified.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.setEnvVar(cmd, args)
		},
	}
	setCmd.Flags().String("scope", "project", "Scope for environment variable (global, project)")
	envCmd.AddCommand(setCmd)

	// Get command
	getCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get an agent environment variable",
		Long:  "Get an agent environment variable. Use --scope to filter by scope, or omit to check both scopes and show which one was used.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.getEnvVar(cmd, args[0])
		},
	}
	getCmd.Flags().String("scope", "", "Scope to check (project, global). If not specified, checks both and shows which scope was used.")
	envCmd.AddCommand(getCmd)

	// List command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all agent environment variables",
		Long:  "List all agent environment variables. Use --scope to filter by scope.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.listEnvVars(cmd)
		},
	}
	listCmd.Flags().String("scope", "all", "Scope to list (project, global, all)")
	envCmd.AddCommand(listCmd)

	// Delete command
	deleteCmd := &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete an agent environment variable",
		Long:  "Delete an agent environment variable. Use --scope to specify scope (defaults to project).",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.deleteEnvVar(cmd, args[0])
		},
	}
	deleteCmd.Flags().String("scope", "project", "Scope to delete from (project, global, both)")
	envCmd.AddCommand(deleteCmd)

	// Import command
	importCmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import agent environment variables from a file",
		Long:  "Import agent environment variables from a file. Follows typical .env file format - key=value per line. Defaults to project scope unless --scope global is specified.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.loadEnvFile(cmd, args[0])
		},
	}
	importCmd.Flags().String("scope", "project", "Scope for imported environment variables (global, project)")
	envCmd.AddCommand(importCmd)

	return envCmd
}
