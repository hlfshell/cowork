package types

import (
	"fmt"
	"time"
)

// WorkflowState represents the state of an issue/PR workflow
type WorkflowState string

const (
	// WorkflowStateQueued indicates the issue is queued for processing
	WorkflowStateQueued WorkflowState = "queued"

	// WorkflowStateWorkspaceReady indicates the workspace is ready for implementation
	WorkflowStateWorkspaceReady WorkflowState = "workspace_ready"

	// WorkflowStateImplementing indicates the coding agent is implementing changes
	WorkflowStateImplementing WorkflowState = "implementing"

	// WorkflowStatePROpen indicates a pull request has been opened
	WorkflowStatePROpen WorkflowState = "pr_open"

	// WorkflowStateRevising indicates the PR is being revised based on feedback
	WorkflowStateRevising WorkflowState = "revising"

	// WorkflowStateMerged indicates the PR has been merged
	WorkflowStateMerged WorkflowState = "merged"

	// WorkflowStateClosed indicates the PR has been closed without merging
	WorkflowStateClosed WorkflowState = "closed"

	// WorkflowStateAborted indicates the workflow was manually stopped
	WorkflowStateAborted WorkflowState = "aborted"
)

// String returns the string representation of the workflow state
func (ws WorkflowState) String() string {
	return string(ws)
}

// IsValid checks if the workflow state is valid
func (ws WorkflowState) IsValid() bool {
	switch ws {
	case WorkflowStateQueued, WorkflowStateWorkspaceReady, WorkflowStateImplementing,
		WorkflowStatePROpen, WorkflowStateRevising, WorkflowStateMerged,
		WorkflowStateClosed, WorkflowStateAborted:
		return true
	default:
		return false
	}
}

// IsTerminal checks if the workflow state is a terminal state
func (ws WorkflowState) IsTerminal() bool {
	switch ws {
	case WorkflowStateMerged, WorkflowStateClosed, WorkflowStateAborted:
		return true
	default:
		return false
	}
}

// CanTransitionTo checks if the workflow can transition to the target state
func (ws WorkflowState) CanTransitionTo(target WorkflowState) bool {
	validTransitions := map[WorkflowState][]WorkflowState{
		WorkflowStateQueued: {
			WorkflowStateWorkspaceReady,
			WorkflowStateAborted,
		},
		WorkflowStateWorkspaceReady: {
			WorkflowStateImplementing,
			WorkflowStateAborted,
		},
		WorkflowStateImplementing: {
			WorkflowStatePROpen,
			WorkflowStateAborted,
		},
		WorkflowStatePROpen: {
			WorkflowStateRevising,
			WorkflowStateMerged,
			WorkflowStateClosed,
			WorkflowStateAborted,
		},
		WorkflowStateRevising: {
			WorkflowStatePROpen,
			WorkflowStateMerged,
			WorkflowStateClosed,
			WorkflowStateAborted,
		},
		WorkflowStateMerged:  {}, // Terminal state
		WorkflowStateClosed:  {}, // Terminal state
		WorkflowStateAborted: {}, // Terminal state
	}

	allowed, exists := validTransitions[ws]
	if !exists {
		return false
	}

	for _, state := range allowed {
		if state == target {
			return true
		}
	}
	return false
}

// FeedbackIntent represents the intent of feedback on a PR
type FeedbackIntent string

const (
	// FeedbackIntentAsk represents questions or clarifications
	FeedbackIntentAsk FeedbackIntent = "ask"

	// FeedbackIntentChange represents requested code changes
	FeedbackIntentChange FeedbackIntent = "change"

	// FeedbackIntentBlocker represents must-fix issues (failing checks, conflicts)
	FeedbackIntentBlocker FeedbackIntent = "blocker"
)

// String returns the string representation of the feedback intent
func (fi FeedbackIntent) String() string {
	return string(fi)
}

// IsValid checks if the feedback intent is valid
func (fi FeedbackIntent) IsValid() bool {
	switch fi {
	case FeedbackIntentAsk, FeedbackIntentChange, FeedbackIntentBlocker:
		return true
	default:
		return false
	}
}

// Workflow represents an auto-PR workflow for a specific issue
type Workflow struct {
	// Unique identifier for the workflow
	ID int `json:"id,string"`

	// Repository information
	Owner      string `json:"owner"`
	Repo       string `json:"repo"`
	IssueID    int    `json:"issue_id"`
	BaseBranch string `json:"base_branch"`

	// Current state and metadata
	State       WorkflowState `json:"state"`
	BranchName  string        `json:"branch_name,omitempty"`
	PRNumber    *int          `json:"pr_number,omitempty"`
	LastEventTS time.Time     `json:"last_event_ts"`

	// Associated task and workspace
	TaskID      int `json:"task_id,omitempty,string"`
	WorkspaceID int `json:"workspace_id,omitempty,string"`

	// Configuration
	Provider string            `json:"provider"`
	Config   WorkflowConfig    `json:"config"`
	Metadata map[string]string `json:"metadata"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`

	// Error information
	ErrorCount int    `json:"error_count"`
	LastError  string `json:"last_error,omitempty"`

	// Lock information
	LockedBy    string     `json:"locked_by,omitempty"`
	LockedAt    *time.Time `json:"locked_at,omitempty"`
	LockTimeout time.Time  `json:"lock_timeout"`
}

// WorkflowConfig represents configuration for a workflow
type WorkflowConfig struct {
	// Branch naming template (e.g., "feature/{issue-key}-{slug}")
	BranchNamingTemplate string `json:"branch_naming_template" default:"feature/{issue-key}-{slug}"`

	// Required checks that must pass
	RequiredChecks []string `json:"required_checks"`

	// Sync strategy preference
	SyncStrategy string `json:"sync_strategy" default:"rebase"` // "rebase" or "merge"

	// Labels that enable/disable Cowork
	EnableLabels  []string `json:"enable_labels"`  // e.g., ["cowork:on"]
	DisableLabels []string `json:"disable_labels"` // e.g., ["cowork:off"]

	// Timeouts and retry settings
	MaxRetries     int           `json:"max_retries" default:"3"`
	RetryDelay     time.Duration `json:"retry_delay" default:"5m"`
	JobTimeout     time.Duration `json:"job_timeout" default:"2h"`
	CommentTimeout time.Duration `json:"comment_timeout" default:"30s"`

	// Safety settings
	ForcePushDisabled   bool `json:"force_push_disabled" default:"true"`
	OnlyFeatureBranches bool `json:"only_feature_branches" default:"true"`
}

// Validate checks if the workflow config is valid
func (wc *WorkflowConfig) Validate() error {
	if wc.SyncStrategy != "rebase" && wc.SyncStrategy != "merge" {
		return fmt.Errorf("sync strategy must be 'rebase' or 'merge'")
	}

	if wc.MaxRetries < 0 {
		return fmt.Errorf("max retries must be non-negative")
	}

	if wc.RetryDelay < 0 {
		return fmt.Errorf("retry delay must be non-negative")
	}

	if wc.JobTimeout < 0 {
		return fmt.Errorf("job timeout must be non-negative")
	}

	return nil
}

// GetDefaultConfig returns a default workflow configuration
func GetDefaultWorkflowConfig() WorkflowConfig {
	return WorkflowConfig{
		BranchNamingTemplate: "feature/{issue-key}-{slug}",
		RequiredChecks:       []string{"lint", "test"},
		SyncStrategy:         "rebase",
		EnableLabels:         []string{"cowork:on"},
		DisableLabels:        []string{"cowork:off"},
		MaxRetries:           3,
		RetryDelay:           5 * time.Minute,
		JobTimeout:           2 * time.Hour,
		CommentTimeout:       30 * time.Second,
		ForcePushDisabled:    true,
		OnlyFeatureBranches:  true,
	}
}

// WorkflowEvent represents an event that triggers workflow actions
type WorkflowEvent struct {
	// Event information
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Provider  string    `json:"provider"`
	Timestamp time.Time `json:"timestamp"`

	// Repository and issue information
	Owner   string `json:"owner"`
	Repo    string `json:"repo"`
	IssueID int    `json:"issue_id"`

	// Event-specific data
	Data map[string]interface{} `json:"data"`

	// Processing information
	Processed bool   `json:"processed"`
	JobID     string `json:"job_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// CreateWorkflowRequest contains the parameters for creating a new workflow
type CreateWorkflowRequest struct {
	Owner      string         `json:"owner"`
	Repo       string         `json:"repo"`
	IssueID    int            `json:"issue_id"`
	BaseBranch string         `json:"base_branch"`
	Provider   string         `json:"provider"`
	Config     WorkflowConfig `json:"config"`
	TaskID     int            `json:"task_id,omitempty,string"` // Optional: create workflow from existing task
}

// Validate checks if the create workflow request is valid
func (req *CreateWorkflowRequest) Validate() error {
	if req.Owner == "" {
		return fmt.Errorf("owner is required")
	}

	if req.Repo == "" {
		return fmt.Errorf("repo is required")
	}

	if req.IssueID <= 0 {
		return fmt.Errorf("issue ID must be positive")
	}

	if req.BaseBranch == "" {
		return fmt.Errorf("base branch is required")
	}

	if req.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	return req.Config.Validate()
}

// UpdateWorkflowRequest contains the parameters for updating a workflow
type UpdateWorkflowRequest struct {
	WorkflowID  int            `json:"workflow_id,string"`
	State       *WorkflowState `json:"state,omitempty"`
	BranchName  *string        `json:"branch_name,omitempty"`
	PRNumber    *int           `json:"pr_number,omitempty"`
	TaskID      *int           `json:"task_id,omitempty,string"`
	WorkspaceID *int           `json:"workspace_id,omitempty,string"`
	ErrorCount  *int           `json:"error_count,omitempty"`
	LastError   *string        `json:"last_error,omitempty"`
}

// Validate checks if the update workflow request is valid
func (req *UpdateWorkflowRequest) Validate() error {
	if req.WorkflowID == 0 {
		return fmt.Errorf("workflow ID is required")
	}

	if req.State != nil && !req.State.IsValid() {
		return fmt.Errorf("invalid workflow state: %s", *req.State)
	}

	if req.ErrorCount != nil && *req.ErrorCount < 0 {
		return fmt.Errorf("error count must be non-negative")
	}

	return nil
}

// WorkflowLock represents a lock on a workflow
type WorkflowLock struct {
	WorkflowID  int       `json:"workflow_id,string"`
	LockedBy    string    `json:"locked_by"`
	LockedAt    time.Time `json:"locked_at"`
	LockTimeout time.Time `json:"lock_timeout"`
	ProcessID   int       `json:"process_id"`
}

// IsExpired checks if the lock has expired
func (wl *WorkflowLock) IsExpired() bool {
	return time.Now().After(wl.LockTimeout)
}

// IsValid checks if the lock is valid and not expired
func (wl *WorkflowLock) IsValid() bool {
	return !wl.IsExpired()
}
