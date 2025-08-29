package agent

import (
	"context"
	"fmt"
	"time"
)

// AgentStatus represents the current state of an AI agent
type AgentStatus string

const (
	// AgentStatusIdle indicates the agent is ready but not currently working
	AgentStatusIdle AgentStatus = "idle"

	// AgentStatusWorking indicates the agent is currently processing instructions
	AgentStatusWorking AgentStatus = "working"

	// AgentStatusCompleted indicates the agent has finished its work
	AgentStatusCompleted AgentStatus = "completed"

	// AgentStatusFailed indicates the agent encountered an error
	AgentStatusFailed AgentStatus = "failed"

	// AgentStatusStopped indicates the agent was stopped
	AgentStatusStopped AgentStatus = "stopped"
)

// String returns the string representation of the agent status
func (as AgentStatus) String() string {
	return string(as)
}

// IsValid checks if the agent status is valid
func (as AgentStatus) IsValid() bool {
	switch as {
	case AgentStatusIdle, AgentStatusWorking, AgentStatusCompleted, AgentStatusFailed, AgentStatusStopped:
		return true
	default:
		return false
	}
}

// IsTerminal checks if the agent status is a terminal state
func (as AgentStatus) IsTerminal() bool {
	switch as {
	case AgentStatusCompleted, AgentStatusFailed, AgentStatusStopped:
		return true
	default:
		return false
	}
}

// InstructionType represents the type of instruction given to the agent
type InstructionType string

const (
	// InstructionTypeTask indicates the instruction is for a specific task
	InstructionTypeTask InstructionType = "task"

	// InstructionTypePRReview indicates the instruction is for PR review/feedback
	InstructionTypePRReview InstructionType = "pr_review"

	// InstructionTypeUserMessage indicates the instruction is a direct user message
	InstructionTypeUserMessage InstructionType = "user_message"

	// InstructionTypeIssue indicates the instruction is based on an issue
	InstructionTypeIssue InstructionType = "issue"
)

// String returns the string representation of the instruction type
func (it InstructionType) String() string {
	return string(it)
}

// IsValid checks if the instruction type is valid
func (it InstructionType) IsValid() bool {
	switch it {
	case InstructionTypeTask, InstructionTypePRReview, InstructionTypeUserMessage, InstructionTypeIssue:
		return true
	default:
		return false
	}
}

// AgentInstruction contains the instructions for the AI agent
type AgentInstruction struct {
	// Content of the instruction
	Content string `json:"content"`

	// Associated task ID
	TaskID int `json:"task_id"`

	// Additional metadata for the instruction
	Metadata map[string]string `json:"metadata,omitempty"`

	// When the instruction was created
	CreatedAt time.Time `json:"created_at"`
}

// Validate checks if the agent instruction is valid
func (ai *AgentInstruction) Validate() error {
	if ai.Content == "" {
		return fmt.Errorf("instruction content is required")
	}

	if ai.TaskID <= 0 {
		return fmt.Errorf("task ID must be positive")
	}

	return nil
}

// AgentConfig contains configuration for an AI agent
type AgentConfig struct {
	// Type of agent (e.g., "aider", "copilot", etc.)
	AgentType string `json:"agent_type"`

	// Working directory for the agent
	WorkingDir string `json:"working_dir"`

	// Command to run the agent
	Command []string `json:"command"`

	// Arguments for the agent command
	Args []string `json:"args,omitempty"`

	// Environment variables for the agent
	Environment map[string]string `json:"environment,omitempty"`

	// Timeout for agent execution
	Timeout time.Duration `json:"timeout" default:"30m"`

	// Maximum number of retries
	MaxRetries int `json:"max_retries" default:"3"`

	// Whether to enable verbose logging
	Verbose bool `json:"verbose" default:"false"`

	// Agent-specific configuration
	AgentSpecific map[string]interface{} `json:"agent_specific,omitempty"`
}

// Validate checks if the agent configuration is valid
func (ac *AgentConfig) Validate() error {
	if ac.AgentType == "" {
		return fmt.Errorf("agent type is required")
	}

	if ac.WorkingDir == "" {
		return fmt.Errorf("working directory is required")
	}

	if len(ac.Command) == 0 {
		return fmt.Errorf("command is required")
	}

	if ac.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if ac.MaxRetries < 0 {
		return fmt.Errorf("max retries must be non-negative")
	}

	return nil
}

// HistoryEntry represents a single entry in the agent's history
type HistoryEntry struct {
	// Timestamp of the event
	Timestamp time.Time `json:"timestamp"`

	// Type of event (instruction, status_change, error, etc.)
	EventType string `json:"event_type"`

	// Description of what happened
	Description string `json:"description"`

	// Additional data for the event
	Data map[string]interface{} `json:"data,omitempty"`
}

// Agent defines the core interface for AI coding agents
// This interface should be implemented by all AI agents (Aider, Copilot, etc.)
type Agent interface {
	// Initialize sets up the agent with the given configuration
	Initialize(ctx context.Context, config *AgentConfig) error

	// Execute runs the agent with the given instruction
	Execute(ctx context.Context, instruction *AgentInstruction) error

	// GetStatus returns the current status of the agent
	GetStatus() AgentStatus

	// Stop stops the agent if it's currently running
	Stop(ctx context.Context) error

	// Cleanup cleans up any resources used by the agent
	Cleanup(ctx context.Context) error

	// GetInfo returns information about the agent
	GetInfo() map[string]interface{}

	// GetName returns the name of this agent
	GetName() string

	// GetVersion returns the version of this agent
	GetVersion() string

	// GenerateInstructions generates instructions from a task
	GenerateInstructions(task interface{}) (string, error)

	// GetHistory returns the agent's execution history
	GetHistory() []HistoryEntry
}
