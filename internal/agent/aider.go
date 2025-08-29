package agent

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// AiderAgent implements the Agent interface for the Aider AI coding agent
type AiderAgent struct {
	// Configuration for the agent
	config *AgentConfig

	// Current status of the agent
	status AgentStatus

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Process for the running agent
	process *exec.Cmd

	// Context for cancellation
	cancel context.CancelFunc

	// Agent information
	info map[string]interface{}

	// Agent history
	history []HistoryEntry
}

// NewAiderAgent creates a new Aider agent instance
func NewAiderAgent() *AiderAgent {
	return &AiderAgent{
		status: AgentStatusIdle,
		info: map[string]interface{}{
			"name":        "Aider AI Coding Agent",
			"version":     "latest",
			"description": "Open-source AI pair programming tool",
			"website":     "https://github.com/Aider-AI/aider",
		},
		history: make([]HistoryEntry, 0),
	}
}

// Initialize sets up the Aider agent with the given configuration
func (a *AiderAgent) Initialize(ctx context.Context, config *AgentConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Set default timeout if not specified
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Minute
	}

	// Set default max retries if not specified
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	// Validate configuration after setting defaults
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	a.config = config
	a.status = AgentStatusIdle

	return nil
}

// Execute runs the Aider agent with the given instruction
func (a *AiderAgent) Execute(ctx context.Context, instruction *AgentInstruction) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Add to history
	a.addHistoryEntry("instruction_received", "Received instruction for execution", map[string]interface{}{
		"task_id": instruction.TaskID,
		"content": instruction.Content,
	})

	// Validate instruction
	if err := instruction.Validate(); err != nil {
		a.addHistoryEntry("validation_error", "Instruction validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("invalid instruction: %w", err)
	}

	// Check if agent is ready
	if a.status != AgentStatusIdle {
		a.addHistoryEntry("status_error", "Agent is not ready for execution", map[string]interface{}{
			"current_status": a.status.String(),
		})
		return fmt.Errorf("agent is not ready (current status: %s)", a.status)
	}

	// Create instruction file
	instructionFile, err := a.createInstructionFile(instruction)
	if err != nil {
		a.addHistoryEntry("file_creation_error", "Failed to create instruction file", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create instruction file: %w", err)
	}
	defer os.Remove(instructionFile)

	// Set status to working
	a.status = AgentStatusWorking
	a.addHistoryEntry("status_change", "Agent status changed to working", map[string]interface{}{
		"new_status": a.status.String(),
	})

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, a.config.Timeout)
	defer cancel()

	// Execute Aider
	err = a.executeAider(execCtx, instructionFile)
	if err != nil {
		a.status = AgentStatusFailed
		a.addHistoryEntry("execution_error", "Aider execution failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to execute Aider: %w", err)
	}

	// Set status to completed
	a.status = AgentStatusCompleted
	a.addHistoryEntry("execution_completed", "Aider execution completed successfully", map[string]interface{}{
		"task_id": instruction.TaskID,
	})

	return nil
}

// GetStatus returns the current status of the agent
func (a *AiderAgent) GetStatus() AgentStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.status
}

// Stop stops the agent if it's currently running
func (a *AiderAgent) Stop(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.status != AgentStatusWorking {
		return fmt.Errorf("agent is not currently working")
	}

	if a.cancel != nil {
		a.cancel()
	}

	if a.process != nil && a.process.Process != nil {
		if err := a.process.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}

	a.status = AgentStatusStopped
	return nil
}

// Cleanup cleans up any resources used by the agent
func (a *AiderAgent) Cleanup(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Stop the agent if it's running
	if a.status == AgentStatusWorking {
		if a.cancel != nil {
			a.cancel()
		}
		if a.process != nil && a.process.Process != nil {
			a.process.Process.Kill()
		}
	}

	a.status = AgentStatusIdle
	return nil
}

// GetInfo returns information about the agent
func (a *AiderAgent) GetInfo() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	info := make(map[string]interface{})
	for k, v := range a.info {
		info[k] = v
	}

	// Add current status
	info["status"] = a.status.String()

	// Add working directory if config is available
	if a.config != nil {
		info["working_dir"] = a.config.WorkingDir
	} else {
		info["working_dir"] = "not configured"
	}

	return info
}

// GetName returns the name of this agent
func (a *AiderAgent) GetName() string {
	return "aider"
}

// GetVersion returns the version of this agent
func (a *AiderAgent) GetVersion() string {
	return "latest"
}

// GenerateInstructions generates instructions from a task
func (a *AiderAgent) GenerateInstructions(task interface{}) (string, error) {
	// For now, return a simple instruction based on task
	// This can be enhanced to generate more sophisticated instructions
	return "Please implement the requested feature based on the task requirements.", nil
}

// GetHistory returns the agent's execution history
func (a *AiderAgent) GetHistory() []HistoryEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Return a copy of the history
	history := make([]HistoryEntry, len(a.history))
	copy(history, a.history)
	return history
}

// addHistoryEntry adds an entry to the agent's history
func (a *AiderAgent) addHistoryEntry(eventType, description string, data map[string]interface{}) {
	entry := HistoryEntry{
		Timestamp:   time.Now(),
		EventType:   eventType,
		Description: description,
		Data:        data,
	}
	a.history = append(a.history, entry)
}

// createInstructionFile creates a temporary file with the instruction content
func (a *AiderAgent) createInstructionFile(instruction *AgentInstruction) (string, error) {
	// Check if config is available
	if a.config == nil {
		return "", fmt.Errorf("agent not initialized")
	}

	// Create a temporary file in the working directory
	tempFile := filepath.Join(a.config.WorkingDir, fmt.Sprintf("instruction_%d.md", time.Now().Unix()))

	// Prepare the instruction content
	content := fmt.Sprintf(`# AI Agent Instruction

**Task ID:** %d
**Created:** %s

## Content

%s

`, instruction.TaskID, instruction.CreatedAt.Format(time.RFC3339), instruction.Content)

	// Add metadata if available
	if len(instruction.Metadata) > 0 {
		content += "\n## Metadata\n\n"
		for k, v := range instruction.Metadata {
			content += fmt.Sprintf("- **%s:** %s\n", k, v)
		}
	}

	// Write the content to the file
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write instruction file: %w", err)
	}

	return tempFile, nil
}

// executeAider runs the Aider command with the instruction file
func (a *AiderAgent) executeAider(ctx context.Context, instructionFile string) error {
	startTime := time.Now()

	// Prepare the command
	cmd := exec.CommandContext(ctx, a.config.Command[0], a.config.Command[1:]...)
	cmd.Dir = a.config.WorkingDir

	// Set environment variables
	if a.config.Environment != nil {
		for k, v := range a.config.Environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Add instruction file to arguments
	cmd.Args = append(cmd.Args, instructionFile)

	// Capture output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Aider: %w", err)
	}

	// Store the process for potential cancellation
	a.process = cmd

	// Read output
	stdoutBytes, err := io.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf("failed to read stdout: %w", err)
	}

	stderrBytes, err := io.ReadAll(stderr)
	if err != nil {
		return fmt.Errorf("failed to read stderr: %w", err)
	}

	// Wait for the command to complete
	err = cmd.Wait()
	duration := time.Since(startTime)

	// Add execution details to history
	a.addHistoryEntry("execution_details", "Aider execution completed", map[string]interface{}{
		"duration": duration.String(),
		"stdout":   string(stdoutBytes),
		"stderr":   string(stderrBytes),
	})

	// Return error if command failed
	if err != nil {
		return fmt.Errorf("aider command failed: %w", err)
	}

	return nil
}
