package cli

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed providers
var providerDocs embed.FS

// getProviderHelpContent returns the help content for a specific provider
func getProviderHelpContent(providerName string) (string, error) {
	// Read the markdown file from the embedded filesystem
	content, err := fs.ReadFile(providerDocs, fmt.Sprintf("providers/%s.md", providerName))
	if err != nil {
		return "", fmt.Errorf("failed to read help content for %s: %w", providerName, err)
	}
	return string(content), nil
}

// formatProviderHelp formats the markdown content for terminal display
func formatProviderHelp(providerName, content string) string {
	// Convert markdown to terminal-friendly format
	formatted := content

	// Replace markdown headers with terminal-friendly headers
	formatted = strings.ReplaceAll(formatted, "# GitHub Authentication Help", "🔐 GITHUB AUTHENTICATION HELP")
	formatted = strings.ReplaceAll(formatted, "# GitLab Authentication Help", "🔐 GITLAB AUTHENTICATION HELP")
	formatted = strings.ReplaceAll(formatted, "# Bitbucket Authentication Help", "🔐 BITBUCKET AUTHENTICATION HELP")

	// Replace section headers
	formatted = strings.ReplaceAll(formatted, "## What", "📋 What")
	formatted = strings.ReplaceAll(formatted, "## How to get", "🔑 How to get")
	formatted = strings.ReplaceAll(formatted, "## How to use", "🚀 How to use")
	formatted = strings.ReplaceAll(formatted, "## Useful links", "🔗 Useful links")
	formatted = strings.ReplaceAll(formatted, "## Security notes", "⚠️  Security notes")

	// Add separator line after the main header
	formatted = strings.ReplaceAll(formatted, "🔐 GITHUB AUTHENTICATION HELP", "🔐 GITHUB AUTHENTICATION HELP\n"+strings.Repeat("=", 40))
	formatted = strings.ReplaceAll(formatted, "🔐 GITLAB AUTHENTICATION HELP", "🔐 GITLAB AUTHENTICATION HELP\n"+strings.Repeat("=", 40))
	formatted = strings.ReplaceAll(formatted, "🔐 BITBUCKET AUTHENTICATION HELP", "🔐 BITBUCKET AUTHENTICATION HELP\n"+strings.Repeat("=", 40))

	// Add newlines after sections for better readability
	formatted = strings.ReplaceAll(formatted, "## ", "\n## ")

	// Ensure there's a newline at the end
	if !strings.HasSuffix(formatted, "\n") {
		formatted += "\n"
	}

	return formatted
}

// showProviderHelpFromMarkdown displays the help content for a specific provider
func (app *App) showProviderHelpFromMarkdown(cmd *cobra.Command, providerName string) error {
	content, err := getProviderHelpContent(providerName)
	if err != nil {
		cmd.Printf("❌ %v\n", err)
		return err
	}

	formatted := formatProviderHelp(providerName, content)
	cmd.Printf("%s", formatted)
	return nil
}
