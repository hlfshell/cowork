package container

import (
	"fmt"
	"os/exec"
	"strings"
)

// DetectEngine detects which container engine is available and returns the appropriate manager
func DetectEngine() (ContainerManager, error) {
	// Try Docker first
	if dockerManager := detectDocker(); dockerManager != nil {
		return dockerManager, nil
	}

	// Try Podman if Docker is not available
	if podmanManager := detectPodman(); podmanManager != nil {
		return podmanManager, nil
	}

	return nil, fmt.Errorf("no container engine (docker or podman) found in PATH")
}

// detectDocker checks if Docker is available and returns a Docker manager if it is
func detectDocker() ContainerManager {
	// Check if docker command exists
	if _, err := exec.LookPath("docker"); err != nil {
		return nil
	}

	// Check if docker daemon is running
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	if err := cmd.Run(); err != nil {
		return nil
	}

	return NewDockerManager()
}

// detectPodman checks if Podman is available and returns a Podman manager if it is
func detectPodman() ContainerManager {
	// Check if podman command exists
	if _, err := exec.LookPath("podman"); err != nil {
		return nil
	}

	// Check if podman is working
	cmd := exec.Command("podman", "version", "--format", "{{.Version}}")
	if err := cmd.Run(); err != nil {
		return nil
	}

	return NewPodmanManager()
}

// GetAvailableEngines returns a list of available container engines
func GetAvailableEngines() []ContainerEngine {
	var engines []ContainerEngine

	if detectDocker() != nil {
		engines = append(engines, EngineDocker)
	}

	if detectPodman() != nil {
		engines = append(engines, EnginePodman)
	}

	return engines
}

// IsEngineAvailable checks if a specific engine is available
func IsEngineAvailable(engine ContainerEngine) bool {
	switch engine {
	case EngineDocker:
		return detectDocker() != nil
	case EnginePodman:
		return detectPodman() != nil
	default:
		return false
	}
}

// GetEngineInfo returns information about the detected engine
func GetEngineInfo() (ContainerEngine, string, error) {
	manager, err := DetectEngine()
	if err != nil {
		return EngineUnknown, "", err
	}

	version, err := manager.GetVersion()
	if err != nil {
		return manager.GetEngine(), "", fmt.Errorf("failed to get version: %w", err)
	}

	return manager.GetEngine(), version, nil
}

// ValidateEngine checks if the specified engine is available and working
func ValidateEngine(engine ContainerEngine) error {
	if !IsEngineAvailable(engine) {
		return fmt.Errorf("container engine '%s' is not available", engine)
	}

	var manager ContainerManager
	switch engine {
	case EngineDocker:
		manager = NewDockerManager()
	case EnginePodman:
		manager = NewPodmanManager()
	default:
		return fmt.Errorf("unsupported container engine: %s", engine)
	}

	// Test basic functionality
	if _, err := manager.GetVersion(); err != nil {
		return fmt.Errorf("container engine '%s' is not working properly: %w", engine, err)
	}

	return nil
}

// GetPreferredEngine returns the preferred engine based on availability
// Priority: Docker > Podman
func GetPreferredEngine() (ContainerEngine, error) {
	engines := GetAvailableEngines()
	if len(engines) == 0 {
		return EngineUnknown, fmt.Errorf("no container engine available")
	}

	// Prefer Docker over Podman
	for _, engine := range engines {
		if engine == EngineDocker {
			return EngineDocker, nil
		}
	}

	return engines[0], nil
}

// GetEngineCommand returns the command name for the specified engine
func GetEngineCommand(engine ContainerEngine) string {
	switch engine {
	case EngineDocker:
		return "docker"
	case EnginePodman:
		return "podman"
	default:
		return ""
	}
}

// GetEngineDisplayName returns a human-readable name for the engine
func GetEngineDisplayName(engine ContainerEngine) string {
	switch engine {
	case EngineDocker:
		return "Docker"
	case EnginePodman:
		return "Podman"
	default:
		return "Unknown"
	}
}

// GetEngineCapabilities returns the capabilities of the specified engine
func GetEngineCapabilities(engine ContainerEngine) map[string]bool {
	capabilities := map[string]bool{
		"rootless": false,
		"daemon":   false,
		"build":    true,
		"network":  true,
		"volume":   true,
	}

	switch engine {
	case EngineDocker:
		capabilities["daemon"] = true
		capabilities["rootless"] = false
	case EnginePodman:
		capabilities["daemon"] = false
		capabilities["rootless"] = true
	}

	return capabilities
}

// GetEngineStatus returns detailed status information about the engine
func GetEngineStatus(engine ContainerEngine) (map[string]interface{}, error) {
	if !IsEngineAvailable(engine) {
		return nil, fmt.Errorf("engine %s is not available", engine)
	}

	var manager ContainerManager
	switch engine {
	case EngineDocker:
		manager = NewDockerManager()
	case EnginePodman:
		manager = NewPodmanManager()
	default:
		return nil, fmt.Errorf("unsupported engine: %s", engine)
	}

	version, err := manager.GetVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	status := map[string]interface{}{
		"engine":       engine,
		"version":      version,
		"available":    true,
		"capabilities": GetEngineCapabilities(engine),
	}

	// Get additional status information
	switch engine {
	case EngineDocker:
		// Check Docker daemon status
		cmd := exec.Command("docker", "info", "--format", "{{.ServerVersion}}")
		if output, err := cmd.Output(); err == nil {
			status["daemon_version"] = strings.TrimSpace(string(output))
		}
	case EnginePodman:
		// Check Podman system info
		cmd := exec.Command("podman", "system", "info", "--format", "{{.Version}}")
		if output, err := cmd.Output(); err == nil {
			status["system_version"] = strings.TrimSpace(string(output))
		}
	}

	return status, nil
}
