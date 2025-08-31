# AI Agent System

This package provides a generic interface for AI coding agents that can be run in workspace containers. It includes a complete implementation for the Aider AI coding agent and provides a framework for integrating other AI agents.

## Overview

The agent system is designed to:

1. **Run AI agents in workspace containers** - Agents operate within isolated workspace environments using Docker/Podman
2. **Process different types of instructions** - Support for tasks, PR reviews, issues, and user messages
3. **Provide a unified interface** - Generic interface that can work with different AI agents
4. **Manage agent lifecycle** - Creation, execution, monitoring, and cleanup of agents
5. **Generate comprehensive instructions** - Automatically create detailed instructions from task descriptions
6. **Handle Git workflow** - Automatically commit changes and push to remote repositories

## Architecture

### Core Components

- **`Agent`** - Generic interface for AI agents (implements duck typing)
- **`AiderAgent`** - Implementation for the Aider AI coding agent with container-based execution
- **`HistoryEntry`** - Represents events in the agent's execution history
- **`AgentResult`** - Contains the result of agent execution including modified files and metadata

### Key Types

- **`AgentStatus`** - Represents the current state of an agent (idle, working, completed, failed, stopped)
- **`AgentInstruction`** - Contains the instruction content and metadata
- **`AgentConfig`** - Configuration for agent execution
- **`AgentResult`** - Contains execution results, modified files, and metadata
- **`HistoryEntry`** - Represents a single event in the agent's history

## AiderAgent Features

### Container-Based Execution
The AiderAgent runs Aider in a Docker container using the official `paulgauthier/aider` image, following the pattern described in the aider.md documentation:

- **Isolated execution** - Each task runs in its own container
- **Automatic cleanup** - Containers are removed after execution
- **File mounting** - Workspace is mounted into the container
- **Environment isolation** - API keys are securely passed via .env files

### Intelligent Instruction Generation
The agent automatically generates comprehensive instructions from task descriptions:

- **Task context** - Includes task name, description, and metadata
- **Git workflow** - Instructions for committing changes and pushing to remote
- **Code quality guidelines** - Best practices for implementation
- **Completion criteria** - Clear success metrics

### Git Integration
The generated instructions include specific Git workflow guidance:

1. **Group files into commits based on functionality**
2. **Write meaningful commit messages** using conventional commit format
3. **Push the branch remotely** after completion

## Installation and Setup

### Prerequisites

1. **Container Engine** - Docker or Podman must be installed and available
2. **Aider Image** - The `paulgauthier/aider` image will be pulled automatically
3. **OpenAI API Key** - Set your OpenAI API key in the environment

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/hlfshell/cowork/internal/agent"
    "github.com/hlfshell/cowork/internal/types"
)

func main() {
    // Create an Aider agent
    agent := agent.NewAiderAgent()

    // Create a configuration
    config := &agent.AgentConfig{
        AgentType:  "aider",
        WorkingDir: "/path/to/workspace",
        Command:    []string{"aider"},
        Environment: map[string]string{
            "OPENAI_API_KEY": "your-api-key-here",
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

    // Create a task
    task := &types.CreateTaskRequest{
        Name:        "Implement OAuth refresh",
        Description: "Add OAuth token refresh logic to handle expired tokens automatically",
        Metadata: map[string]string{
            "priority":  "high",
            "component": "auth",
        },
        Tags: []string{"authentication", "oauth", "security"},
    }

    // Generate instructions from the task
    instructions, err := agent.GenerateInstructions(task)
    if err != nil {
        log.Fatalf("Failed to generate instructions: %v", err)
    }

    // Create an agent instruction
    instruction := &agent.AgentInstruction{
        Content:   instructions,
        TaskID:    123,
        CreatedAt: time.Now(),
    }

    // Execute the instruction
    err = agent.Execute(ctx, instruction)
    if err != nil {
        log.Fatalf("Failed to execute instruction: %v", err)
    }

    // Get the result
    result := agent.GetLastResult()
    if result != nil {
        if result.Success {
            log.Printf("✅ Task completed successfully: %s", result.Summary)
        } else {
            log.Printf("❌ Task failed: %s", result.Error)
        }
    }

    // Cleanup
    agent.Cleanup(ctx)
}
```

## Configuration Options

### AgentConfig Fields

- **`AgentType`** - Type of agent (e.g., "aider", "copilot")
- **`WorkingDir`** - Working directory for the agent (workspace path)
- **`Command`** - Command to run the agent (usually `["aider"]`)
- **`Args`** - Arguments for the agent command
- **`Environment`** - Environment variables for the agent (API keys, etc.)
- **`Timeout`** - Timeout for agent execution (default: 30 minutes)
- **`MaxRetries`** - Maximum number of retries (default: 3)
- **`Verbose`** - Whether to enable verbose logging
- **`AgentSpecific`** - Agent-specific configuration options

### Aider-Specific Configuration

```go
config := &agent.AgentConfig{
    AgentType: "aider",
    WorkingDir: "/path/to/workspace",
    Command:   []string{"aider"},
    Environment: map[string]string{
        "OPENAI_API_KEY": apiKey,
        // Optional: Add other API keys
        "ANTHROPIC_API_KEY": anthropicKey,
        "GEMINI_API_KEY": geminiKey,
    },
    Timeout:    30 * time.Minute,
    MaxRetries: 3,
    Verbose:    true,
}
```

## Container Execution Details

### Container Configuration
The AiderAgent automatically configures the container with:

- **Image**: `paulgauthier/aider`
- **Working Directory**: `/app` (mounted from workspace)
- **User**: Current user ID (for proper file permissions)
- **Environment**: API keys and other environment variables
- **Volumes**: Workspace directory and .env file
- **Auto-removal**: Container is removed after execution

### Aider Command Line
The agent runs Aider with these flags:

```bash
aider --model gpt-4o \
      --message-file /app/instructions.md \
      --yes-always \
      --auto-commits \
      --notifications \
      --timeout 900
```

### Generated Instructions Format
The agent generates comprehensive instructions that include:

```markdown
# Task: [Task Name]

## Description
[Task description]

## Requirements
- Implement the requested feature based on the task description
- Follow best practices for code quality and maintainability
- Ensure proper error handling and edge case coverage
- Write clear, readable code with appropriate comments

## Git Workflow
When you are done with the implementation:

1. **Group files into commits based on functionality**
2. **Write meaningful commit messages** using conventional commit format
3. **Push the branch remotely** using `git push origin HEAD`

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
```

## Error Handling

The agent system provides comprehensive error handling:

```go
result, err := agent.Execute(ctx, instruction)
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
log.Printf("Modified files: %v", result.ModifiedFiles)
log.Printf("Created files: %v", result.CreatedFiles)
```

## Testing

The package includes comprehensive tests:

```bash
# Run all agent tests
go test ./internal/agent/... -v

# Run specific test files
go test ./internal/agent/aider_test.go -v
go test ./internal/agent/interface_test.go -v
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

// Create agent for the workspace
agent := agent.NewAiderAgent()
config := &agent.AgentConfig{
    AgentType:  "aider",
    WorkingDir: workspace.Path,
    Environment: map[string]string{
        "OPENAI_API_KEY": apiKey,
    },
    Timeout: 30 * time.Minute,
}

// Initialize and execute
agent.Initialize(ctx, config)

task := &types.CreateTaskRequest{
    Name:        "Implement OAuth refresh",
    Description: "Add OAuth token refresh logic",
}
instructions, _ := agent.GenerateInstructions(task)

instruction := &agent.AgentInstruction{
    Content:   instructions,
    TaskID:    workspace.ID,
    CreatedAt: time.Now(),
}

err = agent.Execute(ctx, instruction)
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

func (a *MyAgent) Execute(ctx context.Context, instruction *AgentInstruction) error {
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

func (a *MyAgent) GetName() string {
    return "my-agent"
}

func (a *MyAgent) GetVersion() string {
    return "1.0.0"
}

func (a *MyAgent) GenerateInstructions(task interface{}) (string, error) {
    // Generate instructions from task
}

func (a *MyAgent) GetHistory() []HistoryEntry {
    // Return agent history
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

## Best Practices

1. **Always set timeouts** - Prevent agents from running indefinitely
2. **Handle errors gracefully** - Check both execution errors and result success
3. **Clean up resources** - Call `Cleanup()` on agents when done
4. **Use appropriate instruction types** - Choose the right type for your use case
5. **Monitor agent status** - Check agent status during long-running operations
6. **Validate instructions** - Ensure instructions are valid before execution
7. **Use metadata** - Include relevant context in instruction metadata
8. **Set proper file permissions** - Ensure the agent can read/write workspace files
9. **Use environment variables** - Pass API keys securely via environment
10. **Monitor container resources** - Ensure sufficient memory and CPU for agent execution

## Troubleshooting

### Common Issues

1. **Container engine not found** - Ensure Docker or Podman is installed and in PATH
2. **Permission denied** - Check file permissions in the workspace directory
3. **Timeout errors** - Increase timeout for complex tasks
4. **API key issues** - Verify OpenAI API key is set correctly
5. **Workspace not found** - Ensure workspace path exists and is accessible
6. **Container pull failed** - Check network connectivity and Docker registry access
7. **File mounting issues** - Ensure workspace path is absolute and accessible

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

### Container Logs

Access container logs for debugging:

```go
result := agent.GetLastResult()
if result != nil {
    fmt.Printf("Container output: %s\n", result.Output)
}
```

## Contributing

When contributing to the agent system:

1. **Follow Go conventions** - Use proper naming and documentation
2. **Add tests** - Include tests for new functionality
3. **Update documentation** - Keep README and examples current
4. **Handle errors properly** - Provide meaningful error messages
5. **Use descriptive variable names** - Follow the project's naming conventions
6. **Add godoc comments** - Document all exported functions and types
