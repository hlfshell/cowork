package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hlfshell/cowork/internal/types"
)

const (
	// MetadataFileName is the name of the metadata file in each workspace
	MetadataFileName = ".cw-workspace.json"
)

// WorkspaceMetadata represents the metadata stored for each workspace
type WorkspaceMetadata struct {
	// Unique identifier for the workspace
	ID string `json:"id"`

	// Human-readable name for the task
	TaskName string `json:"task_name"`

	// Description of what the task is trying to accomplish
	Description string `json:"description,omitempty"`

	// External ticket ID (e.g., GitHub #123)
	TicketID string `json:"ticket_id,omitempty"`

	// Git branch name for this workspace
	BranchName string `json:"branch_name"`

	// Source repository URL
	SourceRepo string `json:"source_repo"`

	// Base branch (usually main/master)
	BaseBranch string `json:"base_branch"`

	// Creation timestamp
	CreatedAt time.Time `json:"created_at"`

	// Last activity timestamp
	LastActivity time.Time `json:"last_activity"`

	// Current status of the workspace
	Status types.WorkspaceStatus `json:"status"`

	// Associated container ID (if any)
	ContainerID string `json:"container_id,omitempty"`

	// Metadata for the workspace
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SaveWorkspaceMetadata saves workspace metadata to a file within the workspace directory
func SaveWorkspaceMetadata(workspacePath string, workspace *types.Workspace) error {
	metadata := &WorkspaceMetadata{
		ID:           workspace.ID,
		TaskName:     workspace.TaskName,
		Description:  workspace.Description,
		TicketID:     workspace.TicketID,
		BranchName:   workspace.BranchName,
		SourceRepo:   workspace.SourceRepo,
		BaseBranch:   workspace.BaseBranch,
		CreatedAt:    workspace.CreatedAt,
		LastActivity: workspace.LastActivity,
		Status:       workspace.Status,
		ContainerID:  workspace.ContainerID,
		Metadata:     workspace.Metadata,
	}

	metadataPath := filepath.Join(workspacePath, MetadataFileName)

	// Create the metadata file
	file, err := os.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer file.Close()

	// Encode the metadata as JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(metadata); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	return nil
}

// LoadWorkspaceMetadata loads workspace metadata from a file within the workspace directory
func LoadWorkspaceMetadata(workspacePath string) (*types.Workspace, error) {
	metadataPath := filepath.Join(workspacePath, MetadataFileName)

	// Read the metadata file
	file, err := os.Open(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open metadata file: %w", err)
	}
	defer file.Close()

	// Decode the metadata from JSON
	var metadata WorkspaceMetadata
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	// Convert metadata to workspace
	workspace := &types.Workspace{
		ID:           metadata.ID,
		TaskName:     metadata.TaskName,
		Description:  metadata.Description,
		TicketID:     metadata.TicketID,
		Path:         workspacePath,
		BranchName:   metadata.BranchName,
		SourceRepo:   metadata.SourceRepo,
		BaseBranch:   metadata.BaseBranch,
		CreatedAt:    metadata.CreatedAt,
		LastActivity: metadata.LastActivity,
		Status:       metadata.Status,
		ContainerID:  metadata.ContainerID,
		Metadata:     metadata.Metadata,
	}

	return workspace, nil
}

// DiscoverWorkspaces scans the workspaces directory and loads all workspace metadata
func DiscoverWorkspaces(workspacesDir string) ([]*types.Workspace, error) {
	var workspaces []*types.Workspace

	// Check if the workspaces directory exists
	if _, err := os.Stat(workspacesDir); os.IsNotExist(err) {
		return workspaces, nil // Return empty slice if directory doesn't exist
	}

	// Read all entries in the workspaces directory
	entries, err := os.ReadDir(workspacesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspaces directory: %w", err)
	}

	// Process each workspace directory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		workspacePath := filepath.Join(workspacesDir, entry.Name())

		// Try to load metadata for this workspace
		workspace, err := LoadWorkspaceMetadata(workspacePath)
		if err != nil {
			// Skip workspaces with invalid metadata
			continue
		}

		workspaces = append(workspaces, workspace)
	}

	return workspaces, nil
}

// UpdateWorkspaceMetadata updates the metadata file for a workspace
func UpdateWorkspaceMetadata(workspace *types.Workspace) error {
	return SaveWorkspaceMetadata(workspace.Path, workspace)
}

// TruncateDescription truncates a description to a reasonable length for display
func TruncateDescription(description string, maxLength int) string {
	if len(description) <= maxLength {
		return description
	}

	// Try to truncate at a word boundary
	truncated := description[:maxLength]
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > maxLength*3/4 { // Only use word boundary if it's not too far back
		truncated = truncated[:lastSpace]
	}

	return truncated + "..."
}

// ReadDescriptionFromFile reads a description from a file
func ReadDescriptionFromFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read description file: %w", err)
	}
	return strings.TrimSpace(string(content)), nil
}
