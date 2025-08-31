package config

import (
	"fmt"
	"strconv"
	"strings"
)

// GetContainerRuntime returns the configured container runtime (docker/podman)
func (m *Manager) GetContainerRuntime() string {
	if m.config == nil {
		return "docker" // default
	}
	return m.config.Container.Engine
}

// SetContainerRuntime sets the container runtime (docker/podman)
func (m *Manager) SetContainerRuntime(runtime string) error {
	if m.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	if runtime != "docker" && runtime != "podman" {
		return fmt.Errorf("invalid container runtime: %s. Valid options: docker, podman", runtime)
	}

	m.config.Container.Engine = runtime
	return m.SaveProject(m.config)
}

// SetRegistryAuth sets authentication for a container registry
func (m *Manager) SetRegistryAuth(registry, username, password string) error {
	if m.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	if m.config.Auth.Container.Registries == nil {
		m.config.Auth.Container.Registries = make(map[string]RegistryConfig)
	}

	m.config.Auth.Container.Registries[registry] = RegistryConfig{
		URL:        registry,
		Username:   username,
		Password:   password,
		AuthMethod: "basic",
	}

	return m.SaveProject(m.config)
}

// RemoveRegistryAuth removes authentication for a container registry
func (m *Manager) RemoveRegistryAuth(registry string) error {
	if m.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	if m.config.Auth.Container.Registries != nil {
		delete(m.config.Auth.Container.Registries, registry)
	}

	return m.SaveProject(m.config)
}

// GetMemoryLimit returns the container memory limit
func (m *Manager) GetMemoryLimit() string {
	if m.config == nil {
		return "2048m" // default
	}

	memoryMB := m.config.Container.Resources.MemoryMB
	if memoryMB == 0 {
		return "2048m"
	}

	if memoryMB >= 1024 {
		gb := float64(memoryMB) / 1024.0
		if gb == float64(int(gb)) {
			return fmt.Sprintf("%.0fg", gb)
		}
		return fmt.Sprintf("%.1fg", gb)
	}

	return fmt.Sprintf("%dm", memoryMB)
}

// SetMemoryLimit sets the container memory limit
func (m *Manager) SetMemoryLimit(limit string) error {
	if m.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	memoryMB, err := parseMemoryLimit(limit)
	if err != nil {
		return fmt.Errorf("invalid memory limit: %w", err)
	}

	m.config.Container.Resources.MemoryMB = memoryMB
	return m.SaveProject(m.config)
}

// GetCPULimit returns the container CPU limit
func (m *Manager) GetCPULimit() string {
	if m.config == nil {
		return "0" // no limit
	}

	cpuLimit := m.config.Container.Resources.CPULimit
	if cpuLimit == 0 {
		return "0" // no limit
	}

	if cpuLimit == float64(int(cpuLimit)) {
		return fmt.Sprintf("%.0f", cpuLimit)
	}

	return fmt.Sprintf("%.1f", cpuLimit)
}

// SetCPULimit sets the container CPU limit
func (m *Manager) SetCPULimit(limit string) error {
	if m.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	cpuLimit, err := strconv.ParseFloat(limit, 64)
	if err != nil {
		return fmt.Errorf("invalid CPU limit: %s", limit)
	}

	if cpuLimit < 0 {
		return fmt.Errorf("CPU limit cannot be negative")
	}

	m.config.Container.Resources.CPULimit = cpuLimit
	return m.SaveProject(m.config)
}

// GetContainerNetwork returns the container network configuration
func (m *Manager) GetContainerNetwork() string {
	if m.config == nil {
		return "bridge" // default
	}
	return m.config.Container.Network.Mode
}

// SetContainerNetwork sets the container network configuration
func (m *Manager) SetContainerNetwork(network string) error {
	if m.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	validNetworks := []string{"bridge", "host", "none"}
	valid := false
	for _, validNetwork := range validNetworks {
		if network == validNetwork {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid network mode: %s. Valid options: %s", network, strings.Join(validNetworks, ", "))
	}

	m.config.Container.Network.Mode = network
	return m.SaveProject(m.config)
}

// parseMemoryLimit parses a memory limit string like "2g", "512m", "1024"
func parseMemoryLimit(limit string) (int, error) {
	limit = strings.ToLower(strings.TrimSpace(limit))

	if limit == "" {
		return 0, fmt.Errorf("empty memory limit")
	}

	// Check for unit suffix
	var multiplier int
	var numberStr string

	if strings.HasSuffix(limit, "g") || strings.HasSuffix(limit, "gb") {
		multiplier = 1024
		if strings.HasSuffix(limit, "gb") {
			numberStr = limit[:len(limit)-2]
		} else {
			numberStr = limit[:len(limit)-1]
		}
	} else if strings.HasSuffix(limit, "m") || strings.HasSuffix(limit, "mb") {
		multiplier = 1
		if strings.HasSuffix(limit, "mb") {
			numberStr = limit[:len(limit)-2]
		} else {
			numberStr = limit[:len(limit)-1]
		}
	} else {
		// No suffix, assume MB
		multiplier = 1
		numberStr = limit
	}

	number, err := strconv.ParseFloat(numberStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number in memory limit: %s", numberStr)
	}

	if number <= 0 {
		return 0, fmt.Errorf("memory limit must be positive")
	}

	return int(number * float64(multiplier)), nil
}
