package container

import (
	"testing"
)

// TestDetectEngine tests container engine detection
func TestDetectEngine(t *testing.T) {
	// Test case: Detecting container engine should return a valid manager or error
	manager, err := DetectEngine()

	// We can't guarantee which engine is available in the test environment,
	// so we just check that we either get a valid manager or a clear error
	if err != nil {
		// If no engine is available, error should be descriptive
		if err.Error() != "no container engine (docker or podman) found in PATH" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	} else {
		// If we got a manager, it should be valid
		if manager == nil {
			t.Error("Expected manager to not be nil when no error")
		}

		// Check that the manager reports the correct engine type
		engine := manager.GetEngine()
		if engine != EngineDocker && engine != EnginePodman {
			t.Errorf("Expected engine to be Docker or Podman, got: %s", engine)
		}

		// Check that the manager is available
		if !manager.IsAvailable() {
			t.Error("Expected manager to be available")
		}
	}
}

// TestGetAvailableEngines tests getting available engines
func TestGetAvailableEngines(t *testing.T) {
	// Test case: Getting available engines should return a list
	engines := GetAvailableEngines()

	// The list should contain valid engine types
	for _, engine := range engines {
		if engine != EngineDocker && engine != EnginePodman {
			t.Errorf("Expected engine to be Docker or Podman, got: %s", engine)
		}
	}
}

// TestIsEngineAvailable tests engine availability checking
func TestIsEngineAvailable(t *testing.T) {
	// Test case: Checking engine availability should return boolean
	tests := []struct {
		name   string
		engine ContainerEngine
	}{
		{"Docker", EngineDocker},
		{"Podman", EnginePodman},
		{"Unknown", EngineUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			available := IsEngineAvailable(tt.engine)

			// Unknown engine should always be false
			if tt.engine == EngineUnknown && available {
				t.Error("Expected unknown engine to be unavailable")
			}

			// For known engines, we can't predict availability in test environment
			// but we can verify the function doesn't panic
			_ = available
		})
	}
}

// TestGetEngineInfo tests getting engine information
func TestGetEngineInfo(t *testing.T) {
	// Test case: Getting engine info should return engine type and version
	engine, version, err := GetEngineInfo()

	if err != nil {
		// If no engine is available, error should be descriptive
		if err.Error() != "no container engine (docker or podman) found in PATH" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	} else {
		// If we got info, it should be valid
		if engine == EngineUnknown {
			t.Error("Expected engine to not be unknown when no error")
		}

		if version == "" {
			t.Error("Expected version to not be empty when no error")
		}
	}
}

// TestValidateEngine tests engine validation
func TestValidateEngine(t *testing.T) {
	// Test case: Validating unknown engine should fail
	err := ValidateEngine(EngineUnknown)
	if err == nil {
		t.Error("Expected error when validating unknown engine")
	}

	// Test case: Validating known engines should either succeed or give clear error
	tests := []ContainerEngine{EngineDocker, EnginePodman}

	for _, engine := range tests {
		err := ValidateEngine(engine)
		if err != nil {
			// Error should be descriptive
			expectedError := "container engine '" + string(engine) + "' is not available"
			if err.Error() != expectedError && !contains(err.Error(), "not working properly") {
				t.Errorf("Expected descriptive error for %s, got: %v", engine, err)
			}
		}
	}
}

// TestGetPreferredEngine tests getting preferred engine
func TestGetPreferredEngine(t *testing.T) {
	// Test case: Getting preferred engine should return available engine
	engine, err := GetPreferredEngine()

	if err != nil {
		// If no engine is available, error should be descriptive
		if err.Error() != "no container engine available" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	} else {
		// If we got an engine, it should be valid
		if engine == EngineUnknown {
			t.Error("Expected engine to not be unknown when no error")
		}
	}
}

// TestGetEngineCommand tests getting engine command name
func TestGetEngineCommand(t *testing.T) {
	// Test case: Getting engine command should return correct command name
	tests := []struct {
		name     string
		engine   ContainerEngine
		expected string
	}{
		{"Docker", EngineDocker, "docker"},
		{"Podman", EnginePodman, "podman"},
		{"Unknown", EngineUnknown, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := GetEngineCommand(tt.engine)
			if command != tt.expected {
				t.Errorf("Expected command '%s' for %s, got '%s'", tt.expected, tt.name, command)
			}
		})
	}
}

// TestGetEngineDisplayName tests getting engine display name
func TestGetEngineDisplayName(t *testing.T) {
	// Test case: Getting engine display name should return human-readable name
	tests := []struct {
		name     string
		engine   ContainerEngine
		expected string
	}{
		{"Docker", EngineDocker, "Docker"},
		{"Podman", EnginePodman, "Podman"},
		{"Unknown", EngineUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			displayName := GetEngineDisplayName(tt.engine)
			if displayName != tt.expected {
				t.Errorf("Expected display name '%s' for %s, got '%s'", tt.expected, tt.name, displayName)
			}
		})
	}
}

// TestGetEngineCapabilities tests getting engine capabilities
func TestGetEngineCapabilities(t *testing.T) {
	// Test case: Getting engine capabilities should return capability map
	tests := []struct {
		name   string
		engine ContainerEngine
	}{
		{"Docker", EngineDocker},
		{"Podman", EnginePodman},
		{"Unknown", EngineUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capabilities := GetEngineCapabilities(tt.engine)

			// Should have basic capabilities
			requiredCapabilities := []string{"build", "network", "volume"}
			for _, capability := range requiredCapabilities {
				if _, exists := capabilities[capability]; !exists {
					t.Errorf("Expected capability '%s' to exist for %s", capability, tt.name)
				}
			}

			// Docker should have daemon capability
			if tt.engine == EngineDocker {
				if !capabilities["daemon"] {
					t.Error("Expected Docker to have daemon capability")
				}
				if capabilities["rootless"] {
					t.Error("Expected Docker to not have rootless capability")
				}
			}

			// Podman should have rootless capability
			if tt.engine == EnginePodman {
				if !capabilities["rootless"] {
					t.Error("Expected Podman to have rootless capability")
				}
				if capabilities["daemon"] {
					t.Error("Expected Podman to not have daemon capability")
				}
			}
		})
	}
}

// TestGetEngineStatus tests getting engine status
func TestGetEngineStatus(t *testing.T) {
	// Test case: Getting engine status should return status information
	tests := []ContainerEngine{EngineDocker, EnginePodman}

	for _, engine := range tests {
		t.Run(string(engine), func(t *testing.T) {
			status, err := GetEngineStatus(engine)

			if err != nil {
				// Error should be descriptive
				expectedError := "engine " + string(engine) + " is not available"
				if err.Error() != expectedError && !contains(err.Error(), "failed to get version") {
					t.Errorf("Expected descriptive error for %s, got: %v", engine, err)
				}
			} else {
				// If we got status, it should be valid
				if status == nil {
					t.Error("Expected status to not be nil when no error")
				}

				// Check required fields
				if status["engine"] != engine {
					t.Errorf("Expected engine field to match, got: %v", status["engine"])
				}

				if status["available"] != true {
					t.Error("Expected available field to be true")
				}

				if status["capabilities"] == nil {
					t.Error("Expected capabilities field to exist")
				}
			}
		})
	}
}

// TestNewDockerManager tests creating Docker manager
func TestNewDockerManager(t *testing.T) {
	// Test case: Creating Docker manager should return valid manager
	manager := NewDockerManager()

	if manager == nil {
		t.Fatal("Expected manager to not be nil")
	}

	// Check engine type
	if manager.GetEngine() != EngineDocker {
		t.Errorf("Expected engine to be Docker, got: %s", manager.GetEngine())
	}
}

// TestNewPodmanManager tests creating Podman manager
func TestNewPodmanManager(t *testing.T) {
	// Test case: Creating Podman manager should return valid manager
	manager := NewPodmanManager()

	if manager == nil {
		t.Fatal("Expected manager to not be nil")
	}

	// Check engine type
	if manager.GetEngine() != EnginePodman {
		t.Errorf("Expected engine to be Podman, got: %s", manager.GetEngine())
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			contains(s[1:len(s)-1], substr))))
}
