package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hlfshell/cowork/internal/types"
)

// ExampleAiderUsage demonstrates how to use the AiderAgent with container-based execution
func ExampleAiderUsage() {
	// Create an Aider agent
	agent := NewAiderAgent()

	// Create a configuration for the agent
	config := &AgentConfig{
		AgentType:  "aider",
		WorkingDir: "/path/to/your/workspace", // Replace with actual workspace path
		Command:    []string{"aider"},
		Environment: map[string]string{
			"OPENAI_API_KEY": os.Getenv("OPENAI_API_KEY"), // Make sure this is set
		},
		Timeout:    30 * time.Minute,
		MaxRetries: 3,
		Verbose:    true,
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
		Description: "Add OAuth token refresh logic to handle expired tokens automatically. Include proper error handling and retry logic.",
		Metadata: map[string]string{
			"priority":  "high",
			"component": "authentication",
			"files":     "auth/oauth.go, auth/refresh.go",
		},
		Tags: []string{"authentication", "oauth", "security"},
	}

	// Generate instructions from the task
	instructions, err := agent.GenerateInstructions(task)
	if err != nil {
		log.Fatalf("Failed to generate instructions: %v", err)
	}

	fmt.Printf("Generated instructions:\n%s\n", instructions)

	// Create an agent instruction
	instruction := &AgentInstruction{
		Content:   instructions,
		TaskID:    123,
		CreatedAt: time.Now(),
		Metadata: map[string]string{
			"task_name": task.Name,
			"priority":  task.Metadata["priority"],
		},
	}

	// Execute the instruction
	fmt.Println("Starting Aider execution...")
	err = agent.Execute(ctx, instruction)
	if err != nil {
		log.Fatalf("Failed to execute instruction: %v", err)
	}

	// Get the result
	result := agent.GetLastResult()
	if result != nil {
		fmt.Printf("Execution completed:\n")
		fmt.Printf("  Success: %t\n", result.Success)
		fmt.Printf("  Summary: %s\n", result.Summary)
		fmt.Printf("  Duration: %s\n", result.Duration)
		if !result.Success {
			fmt.Printf("  Error: %s\n", result.Error)
		}
		fmt.Printf("  Output: %s\n", result.Output)
	}

	// Get agent history
	history := agent.GetHistory()
	fmt.Printf("\nAgent history (%d entries):\n", len(history))
	for i, entry := range history {
		fmt.Printf("  %d. [%s] %s: %s\n", i+1, entry.Timestamp.Format(time.RFC3339), entry.EventType, entry.Description)
	}

	// Cleanup
	err = agent.Cleanup(ctx)
	if err != nil {
		log.Printf("Warning: cleanup failed: %v", err)
	}
}

// ExampleAiderWithWorkspace demonstrates using AiderAgent with a workspace
func ExampleAiderWithWorkspace(workspacePath string, apiKey string) error {
	// Create an Aider agent
	agent := NewAiderAgent()

	// Create a configuration for the workspace
	config := &AgentConfig{
		AgentType:  "aider",
		WorkingDir: workspacePath,
		Command:    []string{"aider"},
		Environment: map[string]string{
			"OPENAI_API_KEY": apiKey,
		},
		Timeout:    30 * time.Minute,
		MaxRetries: 3,
		Verbose:    true,
	}

	// Initialize the agent
	ctx := context.Background()
	err := agent.Initialize(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to initialize agent: %w", err)
	}

	// Create a task for implementing a feature
	task := &types.CreateTaskRequest{
		Name:        "Add user profile endpoint",
		Description: "Create a new API endpoint to retrieve user profile information. Include proper validation, error handling, and tests.",
		Metadata: map[string]string{
			"priority":  "medium",
			"component": "api",
			"endpoint":  "/api/v1/users/{id}/profile",
		},
		Tags: []string{"api", "user", "profile"},
	}

	// Generate instructions
	instructions, err := agent.GenerateInstructions(task)
	if err != nil {
		return fmt.Errorf("failed to generate instructions: %w", err)
	}

	// Create instruction
	instruction := &AgentInstruction{
		Content:   instructions,
		TaskID:    456,
		CreatedAt: time.Now(),
		Metadata: map[string]string{
			"task_name": task.Name,
			"workspace": workspacePath,
		},
	}

	// Execute
	fmt.Printf("Executing task: %s\n", task.Name)
	err = agent.Execute(ctx, instruction)
	if err != nil {
		return fmt.Errorf("failed to execute instruction: %w", err)
	}

	// Check result
	result := agent.GetLastResult()
	if result == nil {
		return fmt.Errorf("no result available")
	}

	if result.Success {
		fmt.Printf("✅ Task completed successfully!\n")
		fmt.Printf("   Summary: %s\n", result.Summary)
		fmt.Printf("   Duration: %s\n", result.Duration)
	} else {
		fmt.Printf("❌ Task failed: %s\n", result.Error)
		fmt.Printf("   Output: %s\n", result.Output)
	}

	// Cleanup
	return agent.Cleanup(ctx)
}

// ExampleAiderWithCustomInstructions demonstrates using custom instructions
func ExampleAiderWithCustomInstructions(workspacePath string, apiKey string, customInstructions string) error {
	// Create an Aider agent
	agent := NewAiderAgent()

	// Create a configuration
	config := &AgentConfig{
		AgentType:  "aider",
		WorkingDir: workspacePath,
		Command:    []string{"aider"},
		Environment: map[string]string{
			"OPENAI_API_KEY": apiKey,
		},
		Timeout:    30 * time.Minute,
		MaxRetries: 3,
	}

	// Initialize the agent
	ctx := context.Background()
	err := agent.Initialize(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to initialize agent: %w", err)
	}

	// Create instruction with custom content
	instruction := &AgentInstruction{
		Content:   customInstructions,
		TaskID:    789,
		CreatedAt: time.Now(),
		Metadata: map[string]string{
			"custom": "true",
		},
	}

	// Execute
	fmt.Println("Executing custom instructions...")
	err = agent.Execute(ctx, instruction)
	if err != nil {
		return fmt.Errorf("failed to execute instruction: %w", err)
	}

	// Get result
	result := agent.GetLastResult()
	if result != nil {
		if result.Success {
			fmt.Printf("✅ Custom task completed successfully!\n")
		} else {
			fmt.Printf("❌ Custom task failed: %s\n", result.Error)
		}
	}

	// Cleanup
	return agent.Cleanup(ctx)
}
