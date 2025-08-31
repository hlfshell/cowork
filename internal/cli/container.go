package cli

import (
	"github.com/spf13/cobra"
)

// addContainerCommands adds container configuration commands
func addContainerCommands(app *App) *cobra.Command {
	containerCmd := &cobra.Command{
		Use:   "container",
		Short: "Manage container runtime settings",
		Long:  "Configure Docker/Podman settings for agent containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.showContainerHelp(cmd)
		},
	}

	// Runtime command to set docker/podman
	runtimeCmd := &cobra.Command{
		Use:   "runtime [docker|podman]",
		Short: "Set container runtime",
		Long:  "Set the container runtime to use (docker or podman)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return app.showContainerRuntime(cmd)
			}
			return app.setContainerRuntime(cmd, args[0])
		},
	}

	// Registry command for registry settings
	registryCmd := &cobra.Command{
		Use:   "registry",
		Short: "Manage container registry settings",
		Long:  "Configure container registry settings including authentication",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.showRegistryHelp(cmd)
		},
	}

	// Login subcommand for registry
	loginCmd := &cobra.Command{
		Use:   "login [registry-url]",
		Short: "Login to container registry",
		Long:  "Authenticate with a container registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.loginToRegistry(cmd, args[0])
		},
	}
	loginCmd.Flags().String("username", "", "Registry username")
	loginCmd.Flags().String("password", "", "Registry password")

	// Logout subcommand for registry
	logoutCmd := &cobra.Command{
		Use:   "logout [registry-url]",
		Short: "Logout from container registry",
		Long:  "Remove authentication credentials for a container registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.logoutFromRegistry(cmd, args[0])
		},
	}

	// Resource limits command
	resourcesCmd := &cobra.Command{
		Use:   "resources",
		Short: "Configure container resource limits",
		Long:  "Set resource limits for agent containers (memory, CPU, etc.)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.showResourcesHelp(cmd)
		},
	}

	// Memory subcommand
	memoryCmd := &cobra.Command{
		Use:   "memory [limit]",
		Short: "Set memory limit for containers",
		Long:  "Set memory limit for agent containers (e.g., 2g, 512m)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return app.showMemoryLimit(cmd)
			}
			return app.setMemoryLimit(cmd, args[0])
		},
	}

	// CPU subcommand
	cpuCmd := &cobra.Command{
		Use:   "cpu [limit]",
		Short: "Set CPU limit for containers",
		Long:  "Set CPU limit for agent containers (e.g., 2, 0.5)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return app.showCPULimit(cmd)
			}
			return app.setCPULimit(cmd, args[0])
		},
	}

	// Network subcommand
	networkCmd := &cobra.Command{
		Use:   "network [network-name]",
		Short: "Set container network",
		Long:  "Set the network for agent containers",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return app.showContainerNetwork(cmd)
			}
			return app.setContainerNetwork(cmd, args[0])
		},
	}

	registryCmd.AddCommand(loginCmd, logoutCmd)
	resourcesCmd.AddCommand(memoryCmd, cpuCmd, networkCmd)
	containerCmd.AddCommand(runtimeCmd, registryCmd, resourcesCmd)

	return containerCmd
}

// Container command implementations
func (app *App) showContainerHelp(cmd *cobra.Command) error {
	cmd.Println("🐳 Container Configuration")
	cmd.Println("=========================")
	cmd.Println()
	cmd.Println("Manage Docker/Podman settings for agent containers.")
	cmd.Println()
	cmd.Println("Commands:")
	cmd.Println("  cw config container runtime [docker|podman]")
	cmd.Println("  cw config container registry login <registry-url>")
	cmd.Println("  cw config container resources memory <limit>")
	cmd.Println("  cw config container resources cpu <limit>")
	cmd.Println("  cw config container resources network <network-name>")
	cmd.Println()
	cmd.Println("Examples:")
	cmd.Println("  cw config container runtime docker")
	cmd.Println("  cw config container registry login docker.io --username user --password pass")
	cmd.Println("  cw config container resources memory 2g")
	cmd.Println("  cw config container resources cpu 1.5")
	cmd.Println("  cw config container resources network cowork-net")

	return nil
}

func (app *App) showContainerRuntime(cmd *cobra.Command) error {
	_, err := app.configManager.Load()
	if err != nil {
		return err
	}

	runtime := app.configManager.GetContainerRuntime()
	cmd.Printf("Container runtime: %s\n", runtime)
	return nil
}

func (app *App) setContainerRuntime(cmd *cobra.Command, runtime string) error {
	if runtime != "docker" && runtime != "podman" {
		return cmd.Help()
	}

	_, err := app.configManager.Load()
	if err != nil {
		return err
	}

	err = app.configManager.SetContainerRuntime(runtime)
	if err != nil {
		return err
	}

	cmd.Printf("✅ Container runtime set to: %s\n", runtime)
	return nil
}

func (app *App) showRegistryHelp(cmd *cobra.Command) error {
	cmd.Println("📦 Container Registry Management")
	cmd.Println("==============================")
	cmd.Println()
	cmd.Println("Commands:")
	cmd.Println("  cw config container registry login <registry-url>")
	cmd.Println("  cw config container registry logout <registry-url>")
	cmd.Println()
	cmd.Println("Examples:")
	cmd.Println("  cw config container registry login docker.io --username user --password pass")
	cmd.Println("  cw config container registry login ghcr.io --username user --password token")
	cmd.Println("  cw config container registry logout docker.io")

	return nil
}

func (app *App) loginToRegistry(cmd *cobra.Command, registry string) error {
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")

	if username == "" || password == "" {
		return cmd.Help()
	}

	_, err := app.configManager.Load()
	if err != nil {
		return err
	}

	err = app.configManager.SetRegistryAuth(registry, username, password)
	if err != nil {
		return err
	}

	cmd.Printf("✅ Logged in to registry: %s\n", registry)
	return nil
}

func (app *App) logoutFromRegistry(cmd *cobra.Command, registry string) error {
	_, err := app.configManager.Load()
	if err != nil {
		return err
	}

	err = app.configManager.RemoveRegistryAuth(registry)
	if err != nil {
		return err
	}

	cmd.Printf("✅ Logged out from registry: %s\n", registry)
	return nil
}

func (app *App) showResourcesHelp(cmd *cobra.Command) error {
	cmd.Println("⚡ Container Resource Management")
	cmd.Println("==============================")
	cmd.Println()
	cmd.Println("Commands:")
	cmd.Println("  cw config container resources memory [limit]")
	cmd.Println("  cw config container resources cpu [limit]")
	cmd.Println("  cw config container resources network [network-name]")
	cmd.Println()
	cmd.Println("Examples:")
	cmd.Println("  cw config container resources memory 2g     # Set 2GB memory limit")
	cmd.Println("  cw config container resources memory 512m   # Set 512MB memory limit")
	cmd.Println("  cw config container resources cpu 2         # Set 2 CPU cores")
	cmd.Println("  cw config container resources cpu 0.5       # Set 0.5 CPU cores")
	cmd.Println("  cw config container resources network host  # Use host networking")

	return nil
}

func (app *App) showMemoryLimit(cmd *cobra.Command) error {
	_, err := app.configManager.Load()
	if err != nil {
		return err
	}

	limit := app.configManager.GetMemoryLimit()
	cmd.Printf("Memory limit: %s\n", limit)
	return nil
}

func (app *App) setMemoryLimit(cmd *cobra.Command, limit string) error {
	_, err := app.configManager.Load()
	if err != nil {
		return err
	}

	err = app.configManager.SetMemoryLimit(limit)
	if err != nil {
		return err
	}

	cmd.Printf("✅ Memory limit set to: %s\n", limit)
	return nil
}

func (app *App) showCPULimit(cmd *cobra.Command) error {
	_, err := app.configManager.Load()
	if err != nil {
		return err
	}

	limit := app.configManager.GetCPULimit()
	cmd.Printf("CPU limit: %s\n", limit)
	return nil
}

func (app *App) setCPULimit(cmd *cobra.Command, limit string) error {
	_, err := app.configManager.Load()
	if err != nil {
		return err
	}

	err = app.configManager.SetCPULimit(limit)
	if err != nil {
		return err
	}

	cmd.Printf("✅ CPU limit set to: %s\n", limit)
	return nil
}

func (app *App) showContainerNetwork(cmd *cobra.Command) error {
	_, err := app.configManager.Load()
	if err != nil {
		return err
	}

	network := app.configManager.GetContainerNetwork()
	cmd.Printf("Container network: %s\n", network)
	return nil
}

func (app *App) setContainerNetwork(cmd *cobra.Command, network string) error {
	_, err := app.configManager.Load()
	if err != nil {
		return err
	}

	err = app.configManager.SetContainerNetwork(network)
	if err != nil {
		return err
	}

	cmd.Printf("✅ Container network set to: %s\n", network)
	return nil
}
