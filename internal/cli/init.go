package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hlfshell/cowork/internal/git"
	"github.com/spf13/cobra"
)

// initializeProject initializes cowork for the current project
func (app *App) initializeProject(cmd *cobra.Command) error {
	// Check if we're in a Git repository
	repoInfo, err := git.DetectCurrentRepository()
	if err != nil {
		return fmt.Errorf("failed to detect Git repository: %w", err)
	}

	cmd.Printf("✅ Detected Git repository: %s\n", repoInfo.Path)
	cmd.Printf("   Current branch: %s\n", repoInfo.CurrentBranch)

	// Check if already initialized
	cwDir := filepath.Join(repoInfo.Path, ".cw")
	if _, err := os.Stat(cwDir); err == nil {
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			return fmt.Errorf("cowork is already initialized in this repository. Use --force to reinitialize")
		}
		cmd.Printf("ℹ️  Reinitializing existing cowork setup\n")
	}

	// Create .cw directory
	if err := os.MkdirAll(cwDir, 0755); err != nil {
		return fmt.Errorf("failed to create .cw directory: %w", err)
	}
	cmd.Printf("✅ Created .cw directory: %s\n", cwDir)

	// Create workspaces directory
	workspacesDir := filepath.Join(cwDir, "workspaces")
	if err := os.MkdirAll(workspacesDir, 0755); err != nil {
		return fmt.Errorf("failed to create workspaces directory: %w", err)
	}
	cmd.Printf("✅ Created workspaces directory: %s\n", workspacesDir)

	// Check if project config exists, create if not
	projectConfigPath := filepath.Join(repoInfo.Path, ".cwconfig")
	if _, err := os.Stat(projectConfigPath); os.IsNotExist(err) {
		// Create default project configuration
		defaultConfig := app.configManager.GetDefaultConfig()
		if err := app.configManager.SaveProject(defaultConfig); err != nil {
			return fmt.Errorf("failed to create project configuration: %w", err)
		}
		cmd.Printf("✅ Created project configuration: %s\n", projectConfigPath)
	} else {
		cmd.Printf("ℹ️  Project configuration already exists: %s\n", projectConfigPath)
	}

	cmd.Printf("\n🎉 Project initialized successfully!\n")
	cmd.Printf("   You can now use: cw task start <task-name>\n")
	cmd.Printf("   Configuration: cw config show\n")

	return nil
}
