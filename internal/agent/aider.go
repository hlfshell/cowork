package agent

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hlfshell/cowork/internal/container"
	"github.com/hlfshell/cowork/internal/types"
)

// AiderAgent implements the Agent interface for the Aider AI coding agent
type AiderAgent struct {
	// Configuration for the agent
	config *AgentConfig

	// Current status of the agent
	status AgentStatus

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Container manager for running aider in containers
	containerManager container.ContainerManager

	// Agent information
	info map[string]interface{}

	// Agent history
	history []HistoryEntry

	// Last execution result
	lastResult *AgentResult
}

// AgentResult represents the result of an agent execution
type AgentResult struct {
	Success       bool              `json:"success"`
	Summary       string            `json:"summary"`
	Output        string            `json:"output"`
	ModifiedFiles []string          `json:"modified_files"`
	CreatedFiles  []string          `json:"created_files"`
	DeletedFiles  []string          `json:"deleted_files"`
	Error         string            `json:"error,omitempty"`
	Metadata      map[string]string `json:"metadata"`
	CompletedAt   time.Time         `json:"completed_at"`
	Duration      time.Duration     `json:"duration"`
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
			"container":   "paulgauthier/aider",
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

	// Initialize container manager
	containerManager, err := container.DetectEngine()
	if err != nil {
		return fmt.Errorf("failed to detect container engine: %w", err)
	}

	a.config = config
	a.containerManager = containerManager
	a.status = AgentStatusIdle

	a.addHistoryEntry("initialized", "Aider agent initialized successfully", map[string]interface{}{
		"container_engine": string(containerManager.GetEngine()),
		"working_dir":      config.WorkingDir,
		"timeout":          config.Timeout.String(),
	})

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

	// Set status to working
	a.status = AgentStatusWorking
	a.addHistoryEntry("status_change", "Agent status changed to working", map[string]interface{}{
		"new_status": a.status.String(),
	})

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, a.config.Timeout)
	defer cancel()

	// Execute Aider in container
	result, err := a.executeAiderInContainer(execCtx, instruction)
	if err != nil {
		a.status = AgentStatusFailed
		a.addHistoryEntry("execution_error", "Aider execution failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to execute Aider: %w", err)
	}

	// Store the result
	a.lastResult = result

	// Set status based on result
	if result.Success {
		a.status = AgentStatusCompleted
		a.addHistoryEntry("execution_completed", "Aider execution completed successfully", map[string]interface{}{
			"task_id":  instruction.TaskID,
			"summary":  result.Summary,
			"duration": result.Duration.String(),
		})
	} else {
		a.status = AgentStatusFailed
		a.addHistoryEntry("execution_failed", "Aider execution failed", map[string]interface{}{
			"task_id": instruction.TaskID,
			"error":   result.Error,
		})
		return fmt.Errorf("aider execution failed: %s", result.Error)
	}

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

	a.status = AgentStatusStopped
	a.addHistoryEntry("stopped", "Agent was stopped by user", nil)
	return nil
}

// Cleanup cleans up any resources used by the agent
func (a *AiderAgent) Cleanup(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

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

	// Add container engine info
	if a.containerManager != nil {
		info["container_engine"] = string(a.containerManager.GetEngine())
		info["container_available"] = a.containerManager.IsAvailable()
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
	// Type assertion to get task details
	taskReq, ok := task.(*types.CreateTaskRequest)
	if !ok {
		return "", fmt.Errorf("invalid task type, expected *types.CreateTaskRequest")
	}

	// Generate comprehensive instructions based on the task
	instructions := fmt.Sprintf(`# Task: %s

## Description
%s

## Requirements
- Implement the requested feature based on the task description
- Follow best practices for code quality and maintainability
- Ensure proper error handling and edge case coverage
- Write clear, readable code with appropriate comments

## Git Workflow
When you are done with the implementation:

1. **Group files into commits based on functionality**:
   - Create separate commits for different logical components
   - Group related changes together (e.g., all authentication changes in one commit)
   - Keep commits focused and atomic

2. **Write meaningful commit messages**:
   - Use conventional commit format: type(scope): description
   - Examples:
     - feat(auth): implement OAuth token refresh
     - fix(api): resolve race condition in user creation
     - docs(readme): update installation instructions
     - test(auth): add unit tests for token validation

3. **Push the branch remotely**:
   - After creating all commits, push the current branch to the remote repository
   - Use: git push origin HEAD

## Code Quality Guidelines
- Write tests for new functionality
- Follow existing code style and patterns
- Add appropriate error handling
- Include documentation where needed
- Run any existing tests to ensure nothing is broken

## Completion Criteria
- All requested functionality is implemented
- Code is properly tested
- Changes are committed with meaningful messages
- Branch is pushed to remote repository
- No obvious bugs or issues remain

Please proceed with the implementation.`, taskReq.Name, taskReq.Description)

	// Add metadata if available
	if len(taskReq.Metadata) > 0 {
		instructions += "\n\n## Additional Context\n"
		for k, v := range taskReq.Metadata {
			instructions += fmt.Sprintf("- **%s:** %s\n", k, v)
		}
	}

	// Add tags if available
	if len(taskReq.Tags) > 0 {
		instructions += fmt.Sprintf("\n**Tags:** %s\n", strings.Join(taskReq.Tags, ", "))
	}

	return instructions, nil
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

// GetLastResult returns the result of the last execution
func (a *AiderAgent) GetLastResult() *AgentResult {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastResult
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

// executeAiderInContainer runs Aider in a Docker container following the aider.md pattern
func (a *AiderAgent) executeAiderInContainer(ctx context.Context, instruction *AgentInstruction) (*AgentResult, error) {
	startTime := time.Now()

	// Create instruction file
	instructionFile, err := a.createInstructionFile(instruction)
	if err != nil {
		return nil, fmt.Errorf("failed to create instruction file: %w", err)
	}
	defer os.Remove(instructionFile)

	// Create .env file for API keys
	envFile, err := a.createEnvFile()
	if err != nil {
		return nil, fmt.Errorf("failed to create .env file: %w", err)
	}
	defer os.Remove(envFile)

	// Prepare container run options
	runOptions := container.RunOptions{
		Image:      "paulgauthier/aider",
		Name:       fmt.Sprintf("aider-task-%d", instruction.TaskID),
		WorkingDir: "/app",
		Environment: map[string]string{
			"OPENAI_API_KEY": a.config.Environment["OPENAI_API_KEY"],
		},
		Volumes: map[string]string{
			a.config.WorkingDir: "/app",
			envFile:             "/app/.env",
		},
		User:        fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
		Remove:      true,
		TTY:         false,
		Interactive: false,
		Command: []string{
			"--model", "gpt-4o",
			"--message-file", "/app/instructions.md",
			"--yes-always",
			"--auto-commits",
			"--notifications",
			"--timeout", "900",
		},
	}

	// Copy instruction file to workspace
	workspaceInstructionFile := filepath.Join(a.config.WorkingDir, "instructions.md")
	if err := copyFile(instructionFile, workspaceInstructionFile); err != nil {
		return nil, fmt.Errorf("failed to copy instruction file to workspace: %w", err)
	}
	defer os.Remove(workspaceInstructionFile)

	// Run the container
	containerID, err := a.containerManager.Run(ctx, runOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to run aider container: %w", err)
	}

	// Get container logs
	logs, err := a.containerManager.Logs(ctx, containerID, container.LogOptions{
		Follow:     false,
		Timestamps: true,
		Tail:       0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}

	// Read logs
	logsBytes, err := io.ReadAll(logs)
	if err != nil {
		return nil, fmt.Errorf("failed to read container logs: %w", err)
	}
	logs.Close()

	// Check if container completed successfully
	containerInfo, err := a.containerManager.Inspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	duration := time.Since(startTime)

	// Determine success based on container exit code and logs
	success := containerInfo.Status == "exited" && strings.Contains(string(logsBytes), "success")

	result := &AgentResult{
		Success:     success,
		Output:      string(logsBytes),
		CompletedAt: time.Now(),
		Duration:    duration,
		Metadata: map[string]string{
			"container_id": containerID,
			"task_id":      fmt.Sprintf("%d", instruction.TaskID),
		},
	}

	if success {
		result.Summary = "Aider successfully completed the task and committed changes"
		// TODO: Parse modified/created files from git status
	} else {
		result.Error = "Aider execution failed or timed out"
		result.Summary = "Aider failed to complete the task"
	}

	return result, nil
}

// createInstructionFile creates a temporary file with the instruction content
func (a *AiderAgent) createInstructionFile(instruction *AgentInstruction) (string, error) {
	// Create a temporary file
	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("aider_instruction_%d_%d.md", instruction.TaskID, time.Now().Unix()))

	// Write the content to the file
	if err := os.WriteFile(tempFile, []byte(instruction.Content), 0644); err != nil {
		return "", fmt.Errorf("failed to write instruction file: %w", err)
	}

	return tempFile, nil
}

// createEnvFile creates a temporary .env file with API keys
func (a *AiderAgent) createEnvFile() (string, error) {
	// Create a temporary .env file
	envFile := filepath.Join(os.TempDir(), fmt.Sprintf("aider_env_%d.env", time.Now().Unix()))

	// Build environment content
	var envContent strings.Builder

	// Add OpenAI API key
	if apiKey, ok := a.config.Environment["OPENAI_API_KEY"]; ok {
		envContent.WriteString(fmt.Sprintf("OPENAI_API_KEY=%s\n", apiKey))
	}

	// Add Anthropic API key if available
	if apiKey, ok := a.config.Environment["ANTHROPIC_API_KEY"]; ok {
		envContent.WriteString(fmt.Sprintf("ANTHROPIC_API_KEY=%s\n", apiKey))
	}

	// Add other API keys
	for key, value := range a.config.Environment {
		if strings.HasSuffix(key, "_API_KEY") && key != "OPENAI_API_KEY" && key != "ANTHROPIC_API_KEY" {
			envContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		}
	}

	// Write the content to the file
	if err := os.WriteFile(envFile, []byte(envContent.String()), 0600); err != nil {
		return "", fmt.Errorf("failed to write .env file: %w", err)
	}

	return envFile, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
