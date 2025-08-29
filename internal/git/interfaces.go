package git

import (
	"context"
	"time"

	"github.com/hlfshell/cowork/internal/types"
)

// ProviderType represents the type of Git hosting provider
type ProviderType string

const (
	// ProviderGitHub represents GitHub
	ProviderGitHub ProviderType = "github"
	// ProviderGitLab represents GitLab
	ProviderGitLab ProviderType = "gitlab"
	// ProviderBitbucket represents Bitbucket
	ProviderBitbucket ProviderType = "bitbucket"
)

// String returns the string representation of the provider type
func (pt ProviderType) String() string {
	return string(pt)
}

// IsValid checks if the provider type is valid
func (pt ProviderType) IsValid() bool {
	switch pt {
	case ProviderGitHub, ProviderGitLab, ProviderBitbucket:
		return true
	default:
		return false
	}
}

// GitProvider defines the interface for Git hosting platform operations
// This interface enables duck typing across different provider implementations
type GitProvider interface {
	// GetProviderType returns the type of this provider
	GetProviderType() ProviderType

	// TestAuth verifies if the provided authentication is valid
	// This should perform a minimal API call to verify credentials without making changes
	TestAuth(ctx context.Context) error

	// GetRepositoryInfo retrieves basic information about a repository
	GetRepositoryInfo(ctx context.Context, owner, repo string) (*Repository, error)

	// GetIssues retrieves issues from a repository with optional filtering
	GetIssues(ctx context.Context, owner, repo string, options *IssueListOptions) ([]*Issue, error)

	// GetIssue retrieves a specific issue by ID
	GetIssue(ctx context.Context, owner, repo string, issueNumber int) (*Issue, error)

	// CreateIssue creates a new issue in the repository
	CreateIssue(ctx context.Context, owner, repo string, issue *CreateIssueRequest) (*Issue, error)

	// UpdateIssue updates an existing issue
	UpdateIssue(ctx context.Context, owner, repo string, issueNumber int, updates *UpdateIssueRequest) (*Issue, error)

	// GetPullRequests retrieves pull requests from a repository with optional filtering
	GetPullRequests(ctx context.Context, owner, repo string, options *PullRequestListOptions) ([]*PullRequest, error)

	// GetPullRequest retrieves a specific pull request by number
	GetPullRequest(ctx context.Context, owner, repo string, prNumber int) (*PullRequest, error)

	// GetPullRequestByBranch retrieves a pull request by source branch name
	GetPullRequestByBranch(ctx context.Context, owner, repo string, branchName string) (*PullRequest, error)

	// GetPullRequestByIssue retrieves a pull request that closes a specific issue
	GetPullRequestByIssue(ctx context.Context, owner, repo string, issueNumber int) (*PullRequest, error)

	// CreatePullRequest creates a new pull request
	CreatePullRequest(ctx context.Context, owner, repo string, pr *CreatePullRequestRequest) (*PullRequest, error)

	// UpdatePullRequest updates an existing pull request
	UpdatePullRequest(ctx context.Context, owner, repo string, prNumber int, updates *UpdatePullRequestRequest) (*PullRequest, error)

	// GetPullRequestReviews retrieves reviews for a pull request
	GetPullRequestReviews(ctx context.Context, owner, repo string, prNumber int) ([]*Review, error)

	// GetPullRequestComments retrieves comments for a pull request
	GetPullRequestComments(ctx context.Context, owner, repo string, prNumber int) ([]*Comment, error)

	// GetIssueComments retrieves comments for an issue
	GetIssueComments(ctx context.Context, owner, repo string, issueNumber int) ([]*Comment, error)

	// CreateComment creates a comment on an issue or pull request
	CreateComment(ctx context.Context, owner, repo string, issueNumber int, comment *CreateCommentRequest) (*Comment, error)

	// GetLabels retrieves available labels for a repository
	GetLabels(ctx context.Context, owner, repo string) ([]*Label, error)
}

// GitOperationsInterface defines the interface for local Git operations
type GitOperationsInterface interface {
	// CloneRepository clones a repository using the specified isolation level
	CloneRepository(req *types.CreateWorkspaceRequest, workspacePath string) error

	// GetRepositoryInfo retrieves information about a Git repository
	GetRepositoryInfo(repoPath string) (*RepositoryInfo, error)
}

// Repository contains basic information about a repository
type Repository struct {
	// Repository owner/organization
	Owner string `json:"owner"`
	// Repository name
	Name string `json:"name"`
	// Full repository name (owner/name)
	FullName string `json:"full_name"`
	// Repository description
	Description string `json:"description"`
	// Repository URL
	URL string `json:"url"`
	// Whether the repository is private
	Private bool `json:"private"`
	// Default branch name - ie main, master, dev, etc
	DefaultBranch string `json:"default_branch"`
	// Repository creation time
	CreatedAt time.Time `json:"created_at"`
	// Last update time
	UpdatedAt time.Time `json:"updated_at"`
}

// Issue represents an issue in a repository
type Issue struct {
	// Issue number
	Number int `json:"number"`
	// Issue title
	Title string `json:"title"`
	// Issue body/description
	Body string `json:"body"`
	// Issue state (open, closed)
	State string `json:"state"`
	// Issue author
	Author *User `json:"author"`
	// Issue assignees
	Assignees []*User `json:"assignees"`
	// Issue labels
	Labels []*Label `json:"labels"`
	// Issue creation time
	CreatedAt time.Time `json:"created_at"`
	// Issue update time
	UpdatedAt time.Time `json:"updated_at"`
	// Issue close time (if closed)
	ClosedAt *time.Time `json:"closed_at"`
	// Comments
	Comments Comment `json:"comments"`
	// Issue URL
	URL string `json:"url"`
}

// PullRequest represents a pull request in a repository
type PullRequest struct {
	// Pull request number
	Number int `json:"number"`
	// Pull request title
	Title string `json:"title"`
	// Pull request body/description
	Body string `json:"body"`
	// Pull request state (open, closed, merged)
	State string `json:"state"`
	// Whether the pull request is merged
	Merged bool `json:"merged"`
	// Merge time (if merged)
	MergedAt *time.Time `json:"merged_at"`
	// Pull request author
	Author *User `json:"author"`
	// Pull request assignees
	Assignees []*User `json:"assignees"`
	// Pull request labels
	Labels []*Label `json:"labels"`
	// Source branch information
	Head *Branch `json:"head"`
	// Target branch information
	Base *Branch `json:"base"`
	// Pull request creation time
	CreatedAt time.Time `json:"created_at"`
	// Pull request update time
	UpdatedAt time.Time `json:"updated_at"`
	// Pull request close time (if closed)
	ClosedAt *time.Time `json:"closed_at"`
	// Comments
	Comments Comment `json:"comments"`
	// Pull request URL
	URL string `json:"url"`
	// Whether the pull request is draft
	Draft bool `json:"draft"`
	// Mergeable status
	Mergeable *bool `json:"mergeable"`
	// Requested reviewers
	RequestedReviewers []*User `json:"requested_reviewers"`
}

// Branch contains information about a branch
type Branch struct {
	// Branch name
	Ref string `json:"ref"`
	// Branch SHA
	SHA string `json:"sha"`
	// Repository information
	Repo *Repository `json:"repo"`
	// User who owns the branch
	User *User `json:"user"`
}

// User represents a user in the Git hosting platform
type User struct {
	// User ID
	ID int `json:"id"`
	// Username
	Login string `json:"login"`
	// Display name
	Name string `json:"name"`
	// Email address
	Email string `json:"email"`
	// User type (User, Organization)
	Type string `json:"type"`
}

// Label represents a label in a repository
type Label struct {
	// Label ID
	ID int `json:"id"`
	// Label name
	Name string `json:"name"`
	// Label description
	Description string `json:"description"`
	// Label URL
	URL string `json:"url"`
}

// Review represents a pull request review
type Review struct {
	// Review ID
	ID int `json:"id"`
	// Review author
	User *User `json:"user"`
	// Review body/comment
	Body string `json:"body"`
	// Review state (approved, changes_requested, commented)
	State string `json:"state"`
	// Review submission time
	SubmittedAt time.Time `json:"submitted_at"`
	// Review commit SHA
	CommitID string `json:"commit_id"`
	// Review URL
	URL string `json:"url"`
}

// Comment represents a comment on an issue or pull request
type Comment struct {
	// Comment ID
	ID int `json:"id"`
	// Comment author
	User *User `json:"user"`
	// Comment body
	Body string `json:"body"`
	// Comment creation time
	CreatedAt time.Time `json:"created_at"`
	// Comment update time
	UpdatedAt time.Time `json:"updated_at"`
	// Comment URL
	URL string `json:"url"`
}

// IssueListOptions contains options for listing issues
type IssueListOptions struct {
	// Filter by state (open, closed, all)
	State string `json:"state"`
	// Filter by assignee
	Assignee string `json:"assignee"`
	// Filter by creator
	Creator string `json:"creator"`
	// Filter by mentioned user
	Mentioned string `json:"mentioned"`
	// Filter by labels (comma-separated)
	Labels string `json:"labels"`
	// Filter by since time
	Since time.Time `json:"since"`
	// Sort by (created, updated, comments)
	Sort string `json:"sort"`
	// Sort direction (asc, desc)
	Direction string `json:"direction"`
	// Page number for pagination
	Page int `json:"page"`
	// Number of items per page
	PerPage int `json:"per_page"`
}

// PullRequestListOptions contains options for listing pull requests
type PullRequestListOptions struct {
	// Filter by state (open, closed, all)
	State string `json:"state"`
	// Filter by head branch
	Head string `json:"head"`
	// Filter by base branch
	Base string `json:"base"`
	// Filter by sort (created, updated, popularity, long-running)
	Sort string `json:"sort"`
	// Sort direction (asc, desc)
	Direction string `json:"direction"`
	// Page number for pagination
	Page int `json:"page"`
	// Number of items per page
	PerPage int `json:"per_page"`
}

// CreateIssueRequest contains data for creating a new issue
type CreateIssueRequest struct {
	// Issue title
	Title string `json:"title"`
	// Issue body/description
	Body string `json:"body"`
	// Issue assignees
	Assignees []string `json:"assignees"`
	// Issue labels
	Labels []string `json:"labels"`
}

// UpdateIssueRequest contains data for updating an issue
type UpdateIssueRequest struct {
	// Issue title
	Title *string `json:"title"`
	// Issue body/description
	Body *string `json:"body"`
	// Issue state (open, closed)
	State *string `json:"state"`
	// Issue assignees
	Assignees *[]string `json:"assignees"`
	// Issue labels
	Labels *[]string `json:"labels"`
}

// CreatePullRequestRequest contains data for creating a new pull request
type CreatePullRequestRequest struct {
	// Pull request title
	Title string `json:"title"`
	// Pull request body/description
	Body string `json:"body"`
	// Source branch
	Head string `json:"head"`
	// Target branch
	Base string `json:"base"`
	// Whether the pull request is draft
	Draft bool `json:"draft"`
	// Requested reviewers
	RequestedReviewers []string `json:"requested_reviewers"`
	// Requested teams
	RequestedTeams []string `json:"requested_teams"`
	// Pull request labels
	Labels []string `json:"labels"`
}

// UpdatePullRequestRequest contains data for updating a pull request
type UpdatePullRequestRequest struct {
	// Pull request title
	Title *string `json:"title"`
	// Pull request body/description
	Body *string `json:"body"`
	// Pull request state (open, closed)
	State *string `json:"state"`
	// Target branch
	Base *string `json:"base"`
	// Whether the pull request is draft
	Draft *bool `json:"draft"`
	// Requested reviewers
	RequestedReviewers *[]string `json:"requested_reviewers"`
	// Requested teams
	RequestedTeams *[]string `json:"requested_teams"`
	// Pull request labels
	Labels *[]string `json:"labels"`
}

// CreateCommentRequest contains data for creating a comment
type CreateCommentRequest struct {
	// Comment body
	Body string `json:"body"`
}

// CoworkProvider defines the interface for Git provider operations that integrate with the cowork task system
// This interface extends GitProvider with task-specific operations and workflow management
type CoworkProvider interface {
	// Embed the base GitProvider interface
	GitProvider

	// Issue and Task Management
	// ScanOpenIssues scans all open issues assigned to the current user
	// Returns issues that should be converted to tasks
	ScanOpenIssues(ctx context.Context, owner, repo string) ([]*Issue, error)

	// CreateTaskFromIssue creates a new task from a provider issue
	// Returns the created task and any error encountered
	CreateTaskFromIssue(ctx context.Context, owner, repo string, issue *Issue) (*types.Task, error)

	// GetTaskByIssue retrieves a task that was created from a specific issue
	// Returns nil if no task exists for the issue
	GetTaskByIssue(ctx context.Context, owner, repo string, issueNumber int) (*types.Task, error)

	// Workspace Management
	// CreateWorkspaceForTask creates a workspace for a task when it starts
	// This should create a branch based on the issue and set up the workspace
	CreateWorkspaceForTask(ctx context.Context, task *types.Task, owner, repo string) (*types.Workspace, error)

	// GenerateBranchName generates a branch name for a task based on the issue
	// This should follow provider-specific naming conventions
	GenerateBranchName(issue *Issue) string

	// Pull Request Management
	// CreatePullRequestForTask creates a pull request for a completed task
	// If a PR already exists, it should be updated instead
	CreatePullRequestForTask(ctx context.Context, task *types.Task, owner, repo string, workspace *types.Workspace) (*PullRequest, error)

	// GetPullRequestForTask retrieves the pull request associated with a task
	// Returns nil if no PR exists for the task
	GetPullRequestForTask(ctx context.Context, task *types.Task, owner, repo string) (*PullRequest, error)

	// LinkPullRequestToIssue links a pull request to the original issue
	// This should add appropriate comments and references
	LinkPullRequestToIssue(ctx context.Context, owner, repo string, pr *PullRequest, issue *Issue) error

	// PR Monitoring and Updates
	// ScanPullRequestsForTasks scans pull requests associated with known tasks
	// Returns PRs that need attention (reviews, comments, etc.)
	ScanPullRequestsForTasks(ctx context.Context, owner, repo string, knownTaskIDs []string) ([]*PullRequest, error)

	// GetPullRequestUpdates retrieves recent updates for a pull request
	// This includes new comments, reviews, status changes, etc.
	GetPullRequestUpdates(ctx context.Context, owner, repo string, prNumber int, since time.Time) (*PullRequestUpdate, error)

	// UpdateTaskFromPullRequest updates a task based on pull request changes
	// This should handle new comments, review requests, etc.
	UpdateTaskFromPullRequest(ctx context.Context, task *types.Task, pr *PullRequest, updates *PullRequestUpdate) error

	// Task Status Synchronization
	// SyncTaskStatusToProvider updates the provider (issue/PR) with task status changes
	// This should update issue labels, PR status, etc.
	SyncTaskStatusToProvider(ctx context.Context, task *types.Task, owner, repo string) error

	// GetProviderMetadata retrieves provider-specific metadata for a task
	// This includes issue labels, PR status, review information, etc.
	GetProviderMetadata(ctx context.Context, task *types.Task, owner, repo string) (map[string]interface{}, error)
}

// PullRequestUpdate represents updates to a pull request
type PullRequestUpdate struct {
	// Pull request number
	PRNumber int `json:"pr_number"`

	// New comments since last check
	NewComments []*Comment `json:"new_comments"`

	// New reviews since last check
	NewReviews []*Review `json:"new_reviews"`

	// Status changes (mergeable, draft, etc.)
	StatusChanges map[string]interface{} `json:"status_changes"`

	// Last update time
	UpdatedAt time.Time `json:"updated_at"`
}
