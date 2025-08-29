# AI Agent System

This package provides a generic interface for AI coding agents that can be run in workspace containers. It includes a complete implementation for the Aider AI coding agent and provides a framework for integrating other AI agents.

## Overview

The agent system is designed to:

1. **Run AI agents in workspace containers** - Agents operate within isolated workspace environments
2. **Process different types of instructions** - Support for tasks, PR reviews, issues, and user messages
3. **Provide a unified interface** - Generic interface that can work with different AI agents
4. **Manage agent lifecycle** - Creation, execution, monitoring, and cleanup of agents

## Architecture

### Core Components

- **`Agent`** - Generic interface for AI agents (implements duck typing)
- **`AiderAgent`** - Implementation for the Aider AI coding agent
- **`HistoryEntry`** - Represents events in the agent's execution history

### Key Types

- **`AgentStatus`** - Represents the current state of an agent (idle, working, completed, failed, stopped)
- **`AgentInstruction`** - Contains the instruction content and metadata
- **`AgentConfig`** - Configuration for agent execution
- **`HistoryEntry`** - Represents a single event in the agent's history

## Installation and Setup

### Prerequisites

1. **Aider Installation** - Install Aider in your workspace container:
   ```bash
   # Create and activate a virtual environment
   python -m venv aider-env
   source aider-env/bin/activate

   # Install Aider
   pip install aider-chat
   ```

2. **OpenAI API Key** - Set your OpenAI API key:
   ```bash
   export OPENAI_API_KEY="your-api-key-here"
   ```

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/hlfshell/cowork/internal/agent"
)

func main() {
    // Create an Aider agent
    agent := agent.NewAiderAgent()
    
    // Create a configuration
    config := &agent.AgentConfig{
        AgentType:  "aider",
        WorkingDir: "/path/to/your/workspace",
        Command:    []string{"aider"},
        Args:       []string{"--yes"},
        Environment: map[string]string{
            "OPENAI_API_KEY": os.Getenv("OPENAI_API_KEY"),
        },
        Timeout:    30 * time.Minute,
        MaxRetries: 3,
    }
    
    // Initialize the agent
    ctx := context.Background()
    err := agent.Initialize(ctx, config)
    if err != nil {
        log.Fatalf("Failed to initialize agent: %v", err)
    }
    
    // Create an instruction
    instruction := &agent.AgentInstruction{
        Content:   "Add OAuth token refresh logic to handle expired tokens",
        TaskID:    123,
        CreatedAt: time.Now(),
        Metadata: map[string]string{
            "task_name": "Implement OAuth refresh",
        },
    }
    
    // Execute the instruction
    err = agent.Execute(ctx, instruction)
    if err != nil {
        log.Fatalf("Failed to execute instruction: %v", err)
    }
    
    log.Printf("Execution completed successfully")
    
    // Get agent history
    history := agent.GetHistory()
    for _, entry := range history {
        log.Printf("[%s] %s: %s", entry.Timestamp.Format(time.RFC3339), entry.EventType, entry.Description)
    }
}
```

## Advanced Usage

### Custom Agent Configuration

```go
// Create a custom agent configuration
config := &agent.AgentConfig{
    AgentType:  "aider",
    WorkingDir: "/path/to/workspace",
    Command:    []string{"aider"},
    Args:       []string{"--yes", "--model", "gpt-4"},
    Environment: map[string]string{
        "OPENAI_API_KEY": os.Getenv("OPENAI_API_KEY"),
        "AIDER_VERBOSE":  "true",
    },
    Timeout:    45 * time.Minute,
    MaxRetries: 5,
    Verbose:    true,
    AgentSpecific: map[string]interface{}{
        "auto_commit": true,
        "model":       "gpt-4",
        "temperature": 0.1,
    },
}

// Create and use the agent
baseManager := agent.NewManager()
agent, err := baseManager.CreateAgent("aider", config)
if err != nil {
    log.Fatalf("Failed to create agent: %v", err)
}

// Execute instruction
instruction := &agent.AgentInstruction{
    Type:      agent.InstructionTypeTask,
    Content:   "Implement a REST API endpoint for user profile management",
    TaskID:    456,
    CreatedAt: time.Now(),
    Priority:  1,
    Metadata: map[string]string{
        "framework":     "gin",
        "database":      "postgresql",
        "auth_method":   "jwt",
        "test_framework": "testify",
    },
}

result, err := agent.Execute(ctx, instruction)
if err != nil {
    log.Fatalf("Failed to execute instruction: %v", err)
}

// Clean up
if err := agent.Cleanup(ctx); err != nil {
    log.Printf("Failed to cleanup agent: %v", err)
}
```

### Batch Processing

```go
// Create multiple instructions
instructions := []*agent.AgentInstruction{
    manager.CreateInstructionFromTask("Fix authentication bug", "The login endpoint is not validating passwords correctly", 1),
    manager.CreateInstructionFromPRReview("PR-123", "Please add input validation for the email field"),
    manager.CreateInstructionFromIssue("ISSUE-456", "Add unit tests", "The codebase lacks comprehensive unit test coverage"),
    manager.CreateInstructionFromUserMessage("Please add API documentation using Swagger"),
}

// Process instructions sequentially
for i, instruction := range instructions {
    log.Printf("Processing instruction %d/%d: %s", i+1, len(instructions), instruction.Type)
    
    result, err := manager.ExecuteWithAider(ctx, instruction, apiKey)
    if err != nil {
        log.Printf("Failed to execute instruction %d: %v", i+1, err)
        continue
    }
    
    log.Printf("Instruction %d result: %s (Success: %t)", i+1, result.Summary, result.Success)
}
```

## Instruction Types

### Task Instructions
Created from development tasks with task ID and description:
```go
instruction := manager.CreateInstructionFromTask(
    "Implement user authentication",
    "Add JWT-based authentication with refresh tokens",
    123,
)
```

### PR Review Instructions
Created from pull request review comments:
```go
instruction := manager.CreateInstructionFromPRReview(
    "PR-456",
    "Please add error handling and include unit tests for the new feature",
)
```

### Issue Instructions
Created from GitHub/GitLab issues:
```go
instruction := manager.CreateInstructionFromIssue(
    "ISSUE-789",
    "Add comprehensive logging",
    "The authentication module needs better logging for debugging",
)
```

### User Message Instructions
Created from direct user messages:
```go
instruction := manager.CreateInstructionFromUserMessage(
    "Please refactor the user authentication code to use dependency injection",
)
```

## Agent Results

The `AgentResult` contains comprehensive information about the execution:

```go
type AgentResult struct {
    Success       bool              // Whether the agent completed successfully
    Summary       string            // Summary of what was accomplished
    Output        string            // Detailed output from the agent
    ModifiedFiles []string          // List of files that were modified
    CreatedFiles  []string          // List of files that were created
    DeletedFiles  []string          // List of files that were deleted
    Error         string            // Error message if the agent failed
    Metadata      map[string]string // Metadata about the execution
    CompletedAt   time.Time         // Timestamp when the result was generated
    Duration      time.Duration     // Duration of the agent's execution
}
```

## Configuration Options

### AgentConfig Fields

- **`AgentType`** - Type of agent (e.g., "aider", "copilot")
- **`WorkingDir`** - Working directory for the agent
- **`Command`** - Command to run the agent
- **`Args`** - Arguments for the agent command
- **`Environment`** - Environment variables for the agent
- **`Timeout`** - Timeout for agent execution (default: 30 minutes)
- **`MaxRetries`** - Maximum number of retries (default: 3)
- **`Verbose`** - Whether to enable verbose logging
- **`AgentSpecific`** - Agent-specific configuration options

### Aider-Specific Configuration

```go
config := &agent.AgentConfig{
    AgentType: "aider",
    Command:   []string{"aider"},
    Args:      []string{"--yes", "--model", "gpt-4"},
    Environment: map[string]string{
        "OPENAI_API_KEY": apiKey,
    },
    AgentSpecific: map[string]interface{}{
        "auto_commit": true,
        "model":       "gpt-4",
        "temperature": 0.1,
    },
}
```

## Error Handling

The agent system provides comprehensive error handling:

```go
result, err := manager.ExecuteWithAider(ctx, instruction, apiKey)
if err != nil {
    // Handle execution errors
    log.Printf("Execution failed: %v", err)
    return
}

if !result.Success {
    // Handle agent failures
    log.Printf("Agent failed: %s", result.Error)
    return
}

// Process successful results
log.Printf("Success: %s", result.Summary)
```

## Testing

The package includes comprehensive tests:

```bash
# Run all agent tests
go test ./internal/agent/... -v

# Run specific test files
go test ./internal/agent/interface_test.go -v
go test ./internal/agent/aider_test.go -v
go test ./internal/agent/manager_test.go -v
```

## Integration with Workspace System

The agent system integrates with the cowork workspace system:

```go
// Create a workspace
workspaceManager := workspace.NewManager(gitTimeoutSeconds)
workspace, err := workspaceManager.CreateWorkspace(&types.CreateWorkspaceRequest{
    TaskName:    "oauth-refresh",
    Description: "Implement OAuth token refresh",
    SourceRepo:  "https://github.com/user/repo",
    BaseBranch:  "main",
})

// Create agent manager for the workspace
agentManager := agent.NewWorkspaceAgentManager(workspace.Path)

// Execute agent in the workspace
instruction := agentManager.CreateInstructionFromTask(
    "Implement OAuth refresh",
    "Add OAuth token refresh logic",
    workspace.ID,
)

result, err := agentManager.ExecuteWithAider(ctx, instruction, apiKey)
```

## Extending the System

### Adding New Agent Types

The system is designed to be easily extensible through duck typing. To add support for a new AI agent:

1. **Implement the Agent Interface**:
```go
type MyAgent struct {
    // Your agent implementation
}

func (a *MyAgent) Initialize(ctx context.Context, config *AgentConfig) error {
    // Initialize your agent
}

func (a *MyAgent) Execute(ctx context.Context, instruction *AgentInstruction) (*AgentResult, error) {
    // Execute the instruction
}

func (a *MyAgent) GetStatus() AgentStatus {
    // Return current status
}

func (a *MyAgent) Stop(ctx context.Context) error {
    // Stop the agent
}

func (a *MyAgent) Cleanup(ctx context.Context) error {
    // Clean up resources
}

func (a *MyAgent) GetInfo() map[string]interface{} {
    // Return agent information
}

func (a *MyAgent) GetType() string {
    return "my-agent"
}

func (a *MyAgent) GetVersion() string {
    return "1.0.0"
}
```

2. **Create an Agent Provider**:
```go
type MyAgentProvider struct{}

func (p *MyAgentProvider) CreateAgent(config *AgentConfig) (Agent, error) {
    agent := &MyAgent{}
    ctx := context.Background()
    if err := agent.Initialize(ctx, config); err != nil {
        return nil, err
    }
    return agent, nil
}

func (p *MyAgentProvider) GetCapabilities() *AgentCapabilities {
    return &AgentCapabilities{
        SupportedInstructionTypes: []InstructionType{
            InstructionTypeTask,
            InstructionTypeUserMessage,
        },
        SupportedLanguages: []string{"python", "javascript"},
        CanCreateFiles:     true,
        CanModifyFiles:     true,
        CanDeleteFiles:     false,
        CanRunTests:        true,
        CanCommitChanges:   false,
        RequiresAPIKey:     true,
        MaxFileSize:        5 * 1024 * 1024, // 5MB
        MaxTokens:          8000,
    }
}

func (p *MyAgentProvider) GetSupportedTypes() []string {
    return []string{"my-agent"}
}

func (p *MyAgentProvider) ValidateConfig(config *AgentConfig) error {
    if config.AgentType != "my-agent" {
        return fmt.Errorf("invalid agent type for MyAgent provider: %s", config.AgentType)
    }
    return config.Validate()
}
```

3. **Register the Provider**:
```go
manager := agent.NewManager()
manager.RegisterProvider("my-agent", &MyAgentProvider{})
```

4. **Use the new agent**:
```go
config := &agent.AgentConfig{
    AgentType:  "my-agent",
    WorkingDir: "/path/to/workspace",
    Command:    []string{"my-agent"},
    // ... other configuration
}

agent, err := manager.CreateAgent("my-agent", config)
```

### Agent Capabilities

The `AgentCapabilities` struct allows you to define what your agent can do:

- **SupportedInstructionTypes**: What types of instructions your agent can handle
- **SupportedLanguages**: Programming languages your agent supports
- **CanCreateFiles**: Whether your agent can create new files
- **CanModifyFiles**: Whether your agent can modify existing files
- **CanDeleteFiles**: Whether your agent can delete files
- **CanRunTests**: Whether your agent can run tests
- **CanCommitChanges**: Whether your agent can commit changes to version control
- **RequiresAPIKey**: Whether your agent requires an API key
- **MaxFileSize**: Maximum file size your agent can handle
- **MaxTokens**: Maximum tokens your agent can process

## Best Practices

1. **Always set timeouts** - Prevent agents from running indefinitely
2. **Handle errors gracefully** - Check both execution errors and result success
3. **Clean up resources** - Call `Cleanup()` on agents when done
4. **Use appropriate instruction types** - Choose the right type for your use case
5. **Monitor agent status** - Check agent status during long-running operations
6. **Validate instructions** - Ensure instructions are valid before execution
7. **Use metadata** - Include relevant context in instruction metadata

## Troubleshooting

### Common Issues

1. **Agent not found** - Ensure the agent is installed and in PATH
2. **Permission denied** - Check file permissions in the workspace directory
3. **Timeout errors** - Increase timeout for complex tasks
4. **API key issues** - Verify OpenAI API key is set correctly
5. **Workspace not found** - Ensure workspace path exists and is accessible

### Debug Mode

Enable verbose logging for debugging:

```go
config := &agent.AgentConfig{
    // ... other config
    Verbose: true,
    Environment: map[string]string{
        "AIDER_VERBOSE": "true",
    },
}
```

## Contributing

When contributing to the agent system:

1. **Follow Go conventions** - Use proper naming and documentation
2. **Add tests** - Include tests for new functionality
3. **Update documentation** - Keep README and examples current
4. **Handle errors** - Provide meaningful error messages
5. **Validate inputs** - Validate all configuration and instruction inputs
