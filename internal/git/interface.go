package git

import "github.com/hlfshell/cowork/internal/types"

// GitOperationsInterface defines the interface for Git operations
type GitOperationsInterface interface {
	// CloneRepository clones a repository using the specified isolation level
	CloneRepository(req *types.CreateWorkspaceRequest, workspacePath string) error

	// GetRepositoryInfo retrieves information about a Git repository
	GetRepositoryInfo(repoPath string) (*RepositoryInfo, error)
}
