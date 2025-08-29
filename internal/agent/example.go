package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

// ExampleUsage demonstrates how to use the agent system
func ExampleUsage() {
	// Create an Aider agent
	agent := NewAiderAgent()

	// Create a configuration
	config := &AgentConfig{
		AgentType:  "aider",
		WorkingDir: "/path/to/workspace",
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
	instruction := &AgentInstruction{
		Content:   "Implement OAuth refresh logic to handle expired tokens automatically",
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

// ExampleAdvancedUsage demonstrates advanced usage patterns
func ExampleAdvancedUsage() {
	// Create an Aider agent
	agent := NewAiderAgent()

	// Create a custom agent configuration
	config := &AgentConfig{
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

	// Initialize the agent
	ctx := context.Background()
	err := agent.Initialize(ctx, config)
	if err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}

	// Get agent information
	info := agent.GetInfo()
	fmt.Printf("Agent: %s (v%s)\n", info["name"], info["version"])
	fmt.Printf("Status: %s\n", info["status"])

	// Create multiple instructions
	instructions := []*AgentInstruction{
		{
			Content:   "Add comprehensive error handling to the authentication module",
			TaskID:    124,
			CreatedAt: time.Now(),
			Metadata: map[string]string{
				"priority": "high",
				"module":   "auth",
			},
		},
		{
			Content:   "Implement unit tests for the OAuth refresh functionality",
			TaskID:    125,
			CreatedAt: time.Now(),
			Metadata: map[string]string{
				"priority": "medium",
				"type":     "testing",
			},
		},
	}

	// Execute instructions sequentially
	for i, instruction := range instructions {
		fmt.Printf("Executing instruction %d/%d for task %d\n", i+1, len(instructions), instruction.TaskID)

		err := agent.Execute(ctx, instruction)
		if err != nil {
			log.Printf("Failed to execute instruction %d: %v", i+1, err)
			continue
		}

		fmt.Printf("Instruction %d completed successfully\n", i+1)
	}

	// Get final history
	history := agent.GetHistory()
	fmt.Printf("Agent executed %d events\n", len(history))
}

// ExampleBatchProcessing demonstrates batch processing of multiple tasks
func ExampleBatchProcessing() {
	// Create an Aider agent
	agent := NewAiderAgent()

	// Create a configuration
	config := &AgentConfig{
		AgentType:  "aider",
		WorkingDir: "/path/to/workspace",
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

	// Define tasks to process
	tasks := []struct {
		id      int
		content string
	}{
		{101, "Add input validation to the user registration form"},
		{102, "Implement password strength requirements"},
		{103, "Add rate limiting to the login endpoint"},
		{104, "Create API documentation for the authentication endpoints"},
	}

	// Process each task
	for _, task := range tasks {
		// Generate instructions from task
		instructions, err := agent.GenerateInstructions(task)
		if err != nil {
			log.Printf("Failed to generate instructions for task %d: %v", task.id, err)
			continue
		}

		// Create instruction
		instruction := &AgentInstruction{
			Content:   instructions,
			TaskID:    task.id,
			CreatedAt: time.Now(),
			Metadata: map[string]string{
				"task_id": fmt.Sprintf("%d", task.id),
			},
		}

		// Execute the instruction
		err = agent.Execute(ctx, instruction)
		if err != nil {
			log.Printf("Failed to execute task %d: %v", task.id, err)
			continue
		}

		log.Printf("Task %d completed successfully", task.id)
	}

	// Print summary
	history := agent.GetHistory()
	completedTasks := 0
	for _, entry := range history {
		if entry.EventType == "execution_completed" {
			completedTasks++
		}
	}

	fmt.Printf("Batch processing completed. %d/%d tasks executed successfully.\n", completedTasks, len(tasks))
}
