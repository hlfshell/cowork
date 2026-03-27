package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hlfshell/cowork/internal/agent"
	"github.com/hlfshell/cowork/internal/types"
)

func main() {
	// Get the test repository path
	testRepoPath, err := filepath.Abs("test-git-dir")
	if err != nil {
		log.Fatalf("Failed to get test repository path: %v", err)
	}

	// Check if test repository exists
	if _, err := os.Stat(testRepoPath); os.IsNotExist(err) {
		log.Fatalf("Test repository not found at: %s", testRepoPath)
	}

	fmt.Printf("🧪 Testing AiderAgent with repository: %s\n", testRepoPath)

	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create an Aider agent
	aiderAgent := agent.NewAiderAgent()

	// Create configuration for the test repository
	config := &agent.AgentConfig{
		AgentType:  "aider",
		WorkingDir: testRepoPath,
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
	fmt.Println("🔧 Initializing AiderAgent...")
	err = aiderAgent.Initialize(ctx, config)
	if err != nil {
		log.Fatalf("Failed to initialize AiderAgent: %v", err)
	}
	fmt.Println("✅ AiderAgent initialized successfully")

	// Create a test task
	testTask := &types.CreateTaskRequest{
		Name:        "Add a simple calculator function",
		Description: "Create a simple calculator function that can perform basic arithmetic operations (add, subtract, multiply, divide). Include proper error handling for division by zero and invalid operations. Add unit tests for the function.",
		Metadata: map[string]string{
			"priority":  "medium",
			"component": "math",
			"language":  "go",
		},
		Tags: []string{"calculator", "math", "go", "testing"},
	}

	// Generate instructions from the task
	fmt.Println("📝 Generating instructions...")
	instructions, err := aiderAgent.GenerateInstructions(testTask)
	if err != nil {
		log.Fatalf("Failed to generate instructions: %v", err)
	}

	fmt.Println("📋 Generated Instructions:")
	fmt.Println("==========================")
	fmt.Println(instructions)

	// Create an agent instruction
	instruction := &agent.AgentInstruction{
		Content:   instructions,
		TaskID:    1,
		CreatedAt: time.Now(),
		Metadata: map[string]string{
			"task_name": testTask.Name,
			"priority":  testTask.Metadata["priority"],
			"repo_path": testRepoPath,
		},
	}

	// Execute the instruction
	fmt.Println("\n🚀 Starting AiderAgent execution...")
	fmt.Println("This will run Aider in a Docker container to implement the calculator function.")
	fmt.Println("The process may take several minutes...")

	startTime := time.Now()
	err = aiderAgent.Execute(ctx, instruction)
	executionTime := time.Since(startTime)

	if err != nil {
		log.Fatalf("❌ AiderAgent execution failed: %v", err)
	}

	// Get the result
	result := aiderAgent.GetLastResult()
	if result == nil {
		log.Fatal("❌ No result available from AiderAgent")
	}

	fmt.Printf("\n📊 Execution Results:\n")
	fmt.Printf("====================\n")
	fmt.Printf("✅ Success: %t\n", result.Success)
	fmt.Printf("📝 Summary: %s\n", result.Summary)
	fmt.Printf("⏱️  Duration: %s\n", result.Duration)
	fmt.Printf("🕐 Total time: %s\n", executionTime)

	if result.Success {
		fmt.Printf("📁 Modified files: %v\n", result.ModifiedFiles)
		fmt.Printf("🆕 Created files: %v\n", result.CreatedFiles)
		fmt.Printf("🗑️  Deleted files: %v\n", result.DeletedFiles)
	} else {
		fmt.Printf("❌ Error: %s\n", result.Error)
	}

	// Show agent history
	fmt.Printf("\n📜 Agent History:\n")
	fmt.Printf("================\n")
	history := aiderAgent.GetHistory()
	for i, entry := range history {
		fmt.Printf("%d. [%s] %s: %s\n", i+1, entry.Timestamp.Format("15:04:05"), entry.EventType, entry.Description)
	}

	// Show container output if available
	if result.Output != "" {
		fmt.Printf("\n📋 Container Output:\n")
		fmt.Printf("===================\n")
		fmt.Println(result.Output)
	}

	// Cleanup
	fmt.Println("\n🧹 Cleaning up...")
	err = aiderAgent.Cleanup(ctx)
	if err != nil {
		log.Printf("⚠️  Warning: cleanup failed: %v", err)
	}

	if result.Success {
		fmt.Println("\n🎉 Test completed successfully!")
		fmt.Println("The AiderAgent has implemented the calculator function in the test repository.")
		fmt.Printf("Check the repository at: %s\n", testRepoPath)
	} else {
		fmt.Println("\n❌ Test failed!")
		fmt.Println("The AiderAgent encountered an error during execution.")
	}

	// Show final status
	fmt.Printf("\n📈 Final Status:\n")
	fmt.Printf("===============\n")
	fmt.Printf("Agent Status: %s\n", aiderAgent.GetStatus())

	info := aiderAgent.GetInfo()
	fmt.Printf("Container Engine: %s\n", info["container_engine"])
	fmt.Printf("Container Available: %t\n", info["container_available"])
}
