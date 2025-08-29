package workspace

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hlfshell/cowork/internal/container"
	"github.com/hlfshell/cowork/internal/git"
	"github.com/hlfshell/cowork/internal/types"
)

// Manager implements the WorkspaceManager interface
type Manager struct {
	// Git operations handler
	gitOps git.GitOperationsInterface

	// Base directory for all workspaces
	baseDir string

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Container manager for workspace containers
	containerManager container.ContainerManager
}

// NewManager creates a new workspace manager for the current project
func NewManager(gitTimeoutSeconds int) (*Manager, error) {
	// Detect the current repository
	repoInfo, err := git.DetectCurrentRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to detect current repository: %w", err)
	}

	// Create project-specific workspace directory
	projectDir := filepath.Join(repoInfo.Path, ".cw", "workspaces")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create project workspace directory: %w", err)
	}

	// Detect and initialize container manager
	containerManager, err := container.DetectEngine()
	if err != nil {
		// Container engine not available, but we can still create workspaces without containers
		fmt.Printf("Warning: No container engine detected: %v\n", err)
		containerManager = nil
	}

	return &Manager{
		gitOps:           git.NewGitOperations(gitTimeoutSeconds),
		containerManager: containerManager,
		baseDir:          projectDir,
	}, nil
}

// CreateWorkspace creates a new isolated workspace
func (m *Manager) CreateWorkspace(req *types.CreateWorkspaceRequest) (*types.Workspace, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Check if a workspace with this task name already exists
	existingWorkspaces, err := DiscoverWorkspaces(m.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing workspaces: %w", err)
	}

	for _, ws := range existingWorkspaces {
		if ws.TaskName == req.TaskName {
			return nil, fmt.Errorf("workspace with task name '%s' already exists", req.TaskName)
		}
	}

	// Generate a unique workspace ID
	// If this is a task-created workspace, use the same ID as the task
	var workspaceID int
	var taskID int
	var isTaskWorkspace bool

	if req.TaskID > 0 {
		// Task-created workspace: use the same ID as the task
		workspaceID = types.GenerateWorkspaceID(req.TaskID)
		taskID = req.TaskID
		isTaskWorkspace = true
	} else {
		// Standalone workspace: generate a new ID
		workspaceID = types.GenerateWorkspaceID()
		taskID = 0
		isTaskWorkspace = false
	}

	// Create the workspace path (convert integer ID to string for directory name)
	workspacePath := filepath.Join(m.baseDir, fmt.Sprintf("%d", workspaceID))

	// Create the workspace object
	workspace := &types.Workspace{
		ID:              workspaceID,
		TaskName:        req.TaskName,
		Description:     req.Description,
		TicketID:        req.TicketID,
		Path:            workspacePath,
		SourceRepo:      req.SourceRepo,
		BaseBranch:      req.BaseBranch,
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
		Status:          types.WorkspaceStatusCreating,
		Metadata:        req.Metadata,
		TaskID:          taskID,
		IsTaskWorkspace: isTaskWorkspace,
		ContainerConfig: req.ContainerConfig,
	}

	// Clone the repository
	if err := m.gitOps.CloneRepository(req, workspacePath); err != nil {
		// Remove the workspace directory on failure
		os.RemoveAll(workspacePath)
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get repository information to populate branch name
	repoInfo, err := m.gitOps.GetRepositoryInfo(workspacePath)
	if err != nil {
		// This is not a critical error, but we should log it
		fmt.Printf("Warning: failed to get repository info: %v\n", err)
	} else {
		workspace.BranchName = repoInfo.CurrentBranch
	}

	// Update status to ready
	workspace.Status = types.WorkspaceStatusReady
	workspace.LastActivity = time.Now()

	// Save workspace metadata
	if err := SaveWorkspaceMetadata(workspacePath, workspace); err != nil {
		// This is not a critical error, but we should log it
		fmt.Printf("Warning: failed to save workspace metadata: %v\n", err)
	}

	return workspace, nil
}

// GetWorkspace retrieves a workspace by ID
func (m *Manager) GetWorkspace(workspaceID int) (*types.Workspace, error) {
	workspacePath := filepath.Join(m.baseDir, fmt.Sprintf("%d", workspaceID))

	// Check if the workspace directory exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace not found: %d", workspaceID)
	}

	// Load workspace metadata
	workspace, err := LoadWorkspaceMetadata(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace metadata: %w", err)
	}

	return workspace, nil
}

// ListWorkspaces returns all workspaces
func (m *Manager) ListWorkspaces() ([]*types.Workspace, error) {
	return DiscoverWorkspaces(m.baseDir)
}

// DeleteWorkspace removes a workspace and cleans up its resources
func (m *Manager) DeleteWorkspace(workspaceID int) error {
	workspacePath := filepath.Join(m.baseDir, fmt.Sprintf("%d", workspaceID))

	// Check if the workspace directory exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return fmt.Errorf("workspace not found: %d", workspaceID)
	}

	// Load workspace metadata to get the path
	workspace, err := LoadWorkspaceMetadata(workspacePath)
	if err != nil {
		return fmt.Errorf("failed to load workspace metadata: %w", err)
	}

	// Update status to cleaning
	workspace.Status = types.WorkspaceStatusCleaning
	if err := UpdateWorkspaceMetadata(workspace); err != nil {
		// This is not a critical error, but we should log it
		fmt.Printf("Warning: failed to update workspace status: %v\n", err)
	}

	// Stop and remove container if it exists
	if workspace.ContainerID != "" && m.containerManager != nil {
		ctx := context.Background()

		// Stop the container
		if err := m.containerManager.Stop(ctx, workspace.ContainerID, 30); err != nil {
			fmt.Printf("Warning: failed to stop container %s: %v\n", workspace.ContainerID, err)
		}

		// Remove the container
		if err := m.containerManager.Remove(ctx, workspace.ContainerID, true); err != nil {
			fmt.Printf("Warning: failed to remove container %s: %v\n", workspace.ContainerID, err)
		}
	}

	// Remove the workspace directory
	if err := os.RemoveAll(workspace.Path); err != nil {
		return fmt.Errorf("failed to remove workspace directory: %w", err)
	}

	return nil
}

// UpdateWorkspaceStatus updates the status of a workspace
func (m *Manager) UpdateWorkspaceStatus(workspaceID int, status types.WorkspaceStatus) error {
	if !status.IsValid() {
		return fmt.Errorf("invalid workspace status: %s", status)
	}

	workspacePath := filepath.Join(m.baseDir, fmt.Sprintf("%d", workspaceID))

	// Check if the workspace directory exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return fmt.Errorf("workspace not found: %d", workspaceID)
	}

	// Load workspace metadata
	workspace, err := LoadWorkspaceMetadata(workspacePath)
	if err != nil {
		return fmt.Errorf("failed to load workspace metadata: %w", err)
	}

	// Update the status
	workspace.Status = status
	workspace.LastActivity = time.Now()

	// Save the updated metadata
	return UpdateWorkspaceMetadata(workspace)
}

// GetWorkspaceByTaskName retrieves a workspace by task name
func (m *Manager) GetWorkspaceByTaskName(taskName string) (*types.Workspace, error) {
	workspaces, err := DiscoverWorkspaces(m.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to discover workspaces: %w", err)
	}

	for _, workspace := range workspaces {
		if workspace.TaskName == taskName {
			return workspace, nil
		}
	}

	return nil, fmt.Errorf("workspace not found with task name: %s", taskName)
}

// GetBaseDirectory returns the base directory for workspaces
func (m *Manager) GetBaseDirectory() string {
	return m.baseDir
}

// CleanupOrphanedWorkspaces removes workspaces that exist on disk but not in memory
func (m *Manager) CleanupOrphanedWorkspaces() error {
	// Get all directories in the base directory
	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		return fmt.Errorf("failed to read base directory: %w", err)
	}

	// Check each directory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		workspaceID := entry.Name()
		workspacePath := filepath.Join(m.baseDir, workspaceID)

		// Try to load metadata for this workspace
		_, err := LoadWorkspaceMetadata(workspacePath)
		if err != nil {
			// This workspace has invalid or missing metadata, remove it
			if err := os.RemoveAll(workspacePath); err != nil {
				fmt.Printf("Warning: failed to remove orphaned workspace %s: %v\n", workspaceID, err)
			} else {
				fmt.Printf("Removed orphaned workspace: %s\n", workspaceID)
			}
		}
	}

	return nil
}

// StartContainer starts the container associated with a workspace
func (m *Manager) StartContainer(ctx context.Context, workspaceID int) error {
	if m.containerManager == nil {
		return fmt.Errorf("no container engine available")
	}

	// Get the workspace
	workspace, err := m.GetWorkspace(workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if workspace has container configuration
	if workspace.ContainerConfig == nil {
		return fmt.Errorf("workspace does not have container configuration")
	}

	// Check if container is already running
	if workspace.ContainerID != "" {
		// Check if container is actually running
		containerInfo, err := m.containerManager.Inspect(ctx, workspace.ContainerID)
		if err == nil && containerInfo.Status == "running" {
			return fmt.Errorf("container is already running")
		}
	}

	// Prepare container run options
	runOptions := container.RunOptions{
		Image:       workspace.ContainerConfig.Image,
		Name:        workspace.ContainerConfig.Name,
		Command:     workspace.ContainerConfig.Command,
		Args:        workspace.ContainerConfig.Args,
		WorkingDir:  workspace.ContainerConfig.WorkingDir,
		Environment: workspace.ContainerConfig.Environment,
		Ports:       workspace.ContainerConfig.Ports,
		Volumes:     workspace.ContainerConfig.Volumes,
		Network:     workspace.ContainerConfig.Network,
		User:        workspace.ContainerConfig.User,
		Privileged:  workspace.ContainerConfig.Privileged,
		Detached:    workspace.ContainerConfig.Detached,
		Remove:      workspace.ContainerConfig.Remove,
		TTY:         workspace.ContainerConfig.TTY,
		Interactive: workspace.ContainerConfig.Interactive,
		Labels:      workspace.ContainerConfig.Labels,
	}

	// Add workspace volume mount
	if runOptions.Volumes == nil {
		runOptions.Volumes = make(map[string]string)
	}
	runOptions.Volumes[workspace.Path] = "/workspace"

	// Generate container name if not provided
	if runOptions.Name == "" {
		runOptions.Name = fmt.Sprintf("cowork-workspace-%d", workspaceID)
	}

	// Add workspace-specific labels
	if runOptions.Labels == nil {
		runOptions.Labels = make(map[string]string)
	}
	runOptions.Labels["cowork.workspace.id"] = fmt.Sprintf("%d", workspaceID)
	runOptions.Labels["cowork.workspace.name"] = workspace.TaskName
	runOptions.Labels["cowork.workspace.branch"] = workspace.BranchName

	// Run the container
	containerID, err := m.containerManager.Run(ctx, runOptions)
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Update workspace with container ID
	workspace.ContainerID = containerID
	workspace.Status = types.WorkspaceStatusActive
	workspace.LastActivity = time.Now()

	// Update container status
	containerInfo, err := m.containerManager.Inspect(ctx, containerID)
	if err == nil {
		workspace.ContainerStatus = &types.ContainerStatus{
			ID:      containerInfo.ID,
			Name:    containerInfo.Name,
			Image:   containerInfo.Image,
			Status:  containerInfo.Status,
			Created: containerInfo.Created,
			Ports:   containerInfo.Ports,
			Labels:  containerInfo.Labels,
		}
	}

	// Save updated workspace metadata
	return UpdateWorkspaceMetadata(workspace)
}

// StopContainer stops the container associated with a workspace
func (m *Manager) StopContainer(ctx context.Context, workspaceID int, timeoutSeconds int) error {
	if m.containerManager == nil {
		return fmt.Errorf("no container engine available")
	}

	// Get the workspace
	workspace, err := m.GetWorkspace(workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if workspace has a container
	if workspace.ContainerID == "" {
		return fmt.Errorf("workspace does not have an associated container")
	}

	// Stop the container
	if err := m.containerManager.Stop(ctx, workspace.ContainerID, timeoutSeconds); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	// Update workspace status
	workspace.Status = types.WorkspaceStatusReady
	workspace.LastActivity = time.Now()

	// Update container status
	containerInfo, err := m.containerManager.Inspect(ctx, workspace.ContainerID)
	if err == nil {
		workspace.ContainerStatus = &types.ContainerStatus{
			ID:      containerInfo.ID,
			Name:    containerInfo.Name,
			Image:   containerInfo.Image,
			Status:  containerInfo.Status,
			Created: containerInfo.Created,
			Ports:   containerInfo.Ports,
			Labels:  containerInfo.Labels,
		}
	}

	// Save updated workspace metadata
	return UpdateWorkspaceMetadata(workspace)
}

// GetContainerStatus retrieves the current status of the container associated with a workspace
func (m *Manager) GetContainerStatus(ctx context.Context, workspaceID int) (*types.ContainerStatus, error) {
	if m.containerManager == nil {
		return nil, fmt.Errorf("no container engine available")
	}

	// Get the workspace
	workspace, err := m.GetWorkspace(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if workspace has a container
	if workspace.ContainerID == "" {
		return nil, fmt.Errorf("workspace does not have an associated container")
	}

	// Get container information
	containerInfo, err := m.containerManager.Inspect(ctx, workspace.ContainerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	// Update workspace container status
	workspace.ContainerStatus = &types.ContainerStatus{
		ID:      containerInfo.ID,
		Name:    containerInfo.Name,
		Image:   containerInfo.Image,
		Status:  containerInfo.Status,
		Created: containerInfo.Created,
		Ports:   containerInfo.Ports,
		Labels:  containerInfo.Labels,
	}

	// Save updated workspace metadata
	if err := UpdateWorkspaceMetadata(workspace); err != nil {
		fmt.Printf("Warning: failed to update workspace metadata: %v\n", err)
	}

	return workspace.ContainerStatus, nil
}

// ExecInContainer executes a command in the container associated with a workspace
func (m *Manager) ExecInContainer(ctx context.Context, workspaceID int, command []string) error {
	if m.containerManager == nil {
		return fmt.Errorf("no container engine available")
	}

	// Get the workspace
	workspace, err := m.GetWorkspace(workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if workspace has a container
	if workspace.ContainerID == "" {
		return fmt.Errorf("workspace does not have an associated container")
	}

	// Execute command in container
	execOptions := container.ExecOptions{
		TTY:         true,
		Interactive: true,
	}

	return m.containerManager.Exec(ctx, workspace.ContainerID, command, execOptions)
}

// GetContainerLogs retrieves logs from the container associated with a workspace
func (m *Manager) GetContainerLogs(ctx context.Context, workspaceID int, follow bool, tail int) (string, error) {
	if m.containerManager == nil {
		return "", fmt.Errorf("no container engine available")
	}

	// Get the workspace
	workspace, err := m.GetWorkspace(workspaceID)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if workspace has a container
	if workspace.ContainerID == "" {
		return "", fmt.Errorf("workspace does not have an associated container")
	}

	// Get container logs
	logOptions := container.LogOptions{
		Follow:     follow,
		Timestamps: true,
		Tail:       tail,
	}

	logsReader, err := m.containerManager.Logs(ctx, workspace.ContainerID, logOptions)
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}
	defer logsReader.Close()

	// Read all logs
	logsBytes, err := io.ReadAll(logsReader)
	if err != nil {
		return "", fmt.Errorf("failed to read container logs: %w", err)
	}

	return string(logsBytes), nil
}
