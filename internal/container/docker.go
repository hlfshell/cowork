package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// DockerManager implements ContainerManager for Docker
type DockerManager struct {
	command string
}

// NewDockerManager creates a new Docker manager
func NewDockerManager() ContainerManager {
	return &DockerManager{
		command: "docker",
	}
}

// GetEngine returns the container engine type
func (d *DockerManager) GetEngine() ContainerEngine {
	return EngineDocker
}

// GetVersion returns the Docker version
func (d *DockerManager) GetVersion() (string, error) {
	cmd := exec.Command(d.command, "version", "--format", "{{.Server.Version}}")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Docker version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// IsAvailable checks if Docker is available and working
func (d *DockerManager) IsAvailable() bool {
	_, err := d.GetVersion()
	return err == nil
}

// Run runs a container with the specified options
func (d *DockerManager) Run(ctx context.Context, options RunOptions) (string, error) {
	args := []string{"run"}

	// Add name if specified
	if options.Name != "" {
		args = append(args, "--name", options.Name)
	}

	// Add working directory
	if options.WorkingDir != "" {
		args = append(args, "-w", options.WorkingDir)
	}

	// Add environment variables
	for key, value := range options.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Add port mappings
	for hostPort, containerPort := range options.Ports {
		args = append(args, "-p", fmt.Sprintf("%s:%s", hostPort, containerPort))
	}

	// Add volume mounts
	for hostPath, containerPath := range options.Volumes {
		args = append(args, "-v", fmt.Sprintf("%s:%s", hostPath, containerPath))
	}

	// Add network
	if options.Network != "" {
		args = append(args, "--network", options.Network)
	}

	// Add user
	if options.User != "" {
		args = append(args, "--user", options.User)
	}

	// Add privileged flag
	if options.Privileged {
		args = append(args, "--privileged")
	}

	// Add detached flag
	if options.Detached {
		args = append(args, "-d")
	}

	// Add remove flag
	if options.Remove {
		args = append(args, "--rm")
	}

	// Add TTY flag
	if options.TTY {
		args = append(args, "-t")
	}

	// Add interactive flag
	if options.Interactive {
		args = append(args, "-i")
	}

	// Add labels
	for key, value := range options.Labels {
		args = append(args, "-l", fmt.Sprintf("%s=%s", key, value))
	}

	// Add image
	args = append(args, options.Image)

	// Add command and arguments
	if len(options.Command) > 0 {
		args = append(args, options.Command...)
	}
	if len(options.Args) > 0 {
		args = append(args, options.Args...)
	}

	cmd := exec.CommandContext(ctx, d.command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run container: %w", err)
	}

	containerID := strings.TrimSpace(string(output))
	return containerID, nil
}

// Start starts a stopped container
func (d *DockerManager) Start(ctx context.Context, containerID string) error {
	cmd := exec.CommandContext(ctx, d.command, "start", containerID)
	return cmd.Run()
}

// Stop stops a running container
func (d *DockerManager) Stop(ctx context.Context, containerID string, timeoutSeconds int) error {
	args := []string{"stop"}
	if timeoutSeconds > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", timeoutSeconds))
	}
	args = append(args, containerID)

	cmd := exec.CommandContext(ctx, d.command, args...)
	return cmd.Run()
}

// Remove removes a container
func (d *DockerManager) Remove(ctx context.Context, containerID string, force bool) error {
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, containerID)

	cmd := exec.CommandContext(ctx, d.command, args...)
	return cmd.Run()
}

// Exec executes a command in a running container
func (d *DockerManager) Exec(ctx context.Context, containerID string, command []string, options ExecOptions) error {
	args := []string{"exec"}

	// Add user
	if options.User != "" {
		args = append(args, "--user", options.User)
	}

	// Add working directory
	if options.WorkingDir != "" {
		args = append(args, "-w", options.WorkingDir)
	}

	// Add environment variables
	for key, value := range options.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Add TTY flag
	if options.TTY {
		args = append(args, "-t")
	}

	// Add interactive flag
	if options.Interactive {
		args = append(args, "-i")
	}

	// Add privileged flag
	if options.Privileged {
		args = append(args, "--privileged")
	}

	// Add container ID and command
	args = append(args, containerID)
	args = append(args, command...)

	cmd := exec.CommandContext(ctx, d.command, args...)
	return cmd.Run()
}

// Logs retrieves logs from a container
func (d *DockerManager) Logs(ctx context.Context, containerID string, options LogOptions) (io.ReadCloser, error) {
	args := []string{"logs"}

	if options.Follow {
		args = append(args, "-f")
	}
	if options.Timestamps {
		args = append(args, "-t")
	}
	if options.Tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", options.Tail))
	}
	if options.Since != "" {
		args = append(args, "--since", options.Since)
	}
	if options.Until != "" {
		args = append(args, "--until", options.Until)
	}

	args = append(args, containerID)

	cmd := exec.CommandContext(ctx, d.command, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start logs command: %w", err)
	}

	return stdout, nil
}

// Inspect inspects a container and returns detailed information
func (d *DockerManager) Inspect(ctx context.Context, containerID string) (*ContainerInfo, error) {
	cmd := exec.CommandContext(ctx, d.command, "inspect", "--format", "{{json .}}", containerID)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	var inspectData map[string]interface{}
	if err := json.Unmarshal(output, &inspectData); err != nil {
		return nil, fmt.Errorf("failed to parse inspect output: %w", err)
	}

	// Extract container information
	info := &ContainerInfo{
		ID:      containerID,
		Name:    extractString(inspectData, "Name"),
		Image:   extractString(inspectData, "Image"),
		Status:  extractString(inspectData, "State.Status"),
		Created: extractString(inspectData, "Created"),
		Labels:  extractLabels(inspectData),
	}

	// Extract ports
	if ports, ok := inspectData["NetworkSettings"].(map[string]interface{}); ok {
		if portBindings, ok := ports["Ports"].(map[string]interface{}); ok {
			for port := range portBindings {
				info.Ports = append(info.Ports, port)
			}
		}
	}

	return info, nil
}

// List lists containers
func (d *DockerManager) List(ctx context.Context, options ListOptions) ([]ContainerInfo, error) {
	args := []string{"ps", "--format", "{{json .}}"}

	if options.All {
		args = append(args, "-a")
	}
	if options.Size {
		args = append(args, "-s")
	}

	cmd := exec.CommandContext(ctx, d.command, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var containers []ContainerInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var containerData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &containerData); err != nil {
			continue // Skip invalid entries
		}

		container := ContainerInfo{
			ID:      extractString(containerData, "ID"),
			Name:    extractString(containerData, "Names"),
			Image:   extractString(containerData, "Image"),
			Status:  extractString(containerData, "Status"),
			Created: extractString(containerData, "CreatedAt"),
		}

		containers = append(containers, container)
	}

	return containers, nil
}

// Pull pulls an image
func (d *DockerManager) Pull(ctx context.Context, image string) error {
	cmd := exec.CommandContext(ctx, d.command, "pull", image)
	return cmd.Run()
}

// Build builds an image
func (d *DockerManager) Build(ctx context.Context, options BuildOptions) (string, error) {
	args := []string{"build"}

	// Add Dockerfile path
	if options.Dockerfile != "" {
		args = append(args, "-f", options.Dockerfile)
	}

	// Add tags
	for _, tag := range options.Tags {
		args = append(args, "-t", tag)
	}

	// Add build arguments
	for key, value := range options.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}

	// Add labels
	for key, value := range options.Labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", key, value))
	}

	// Add no-cache flag
	if options.NoCache {
		args = append(args, "--no-cache")
	}

	// Add pull flag
	if options.Pull {
		args = append(args, "--pull")
	}

	// Add platform
	if options.Platform != "" {
		args = append(args, "--platform", options.Platform)
	}

	// Add build context
	args = append(args, options.Context)

	cmd := exec.CommandContext(ctx, d.command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to build image: %w", err)
	}

	// Extract image ID from output
	lines := strings.Split(string(output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "Successfully built ") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return parts[2], nil
			}
		}
	}

	return "", fmt.Errorf("failed to extract image ID from build output")
}

// RemoveImage removes an image
func (d *DockerManager) RemoveImage(ctx context.Context, imageID string, force bool) error {
	args := []string{"rmi"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, imageID)

	cmd := exec.CommandContext(ctx, d.command, args...)
	return cmd.Run()
}

// ListImages lists images
func (d *DockerManager) ListImages(ctx context.Context) ([]ImageInfo, error) {
	cmd := exec.CommandContext(ctx, d.command, "images", "--format", "{{json .}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	var images []ImageInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var imageData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &imageData); err != nil {
			continue // Skip invalid entries
		}

		image := ImageInfo{
			ID:      extractString(imageData, "ID"),
			Tags:    extractStringSlice(imageData, "Tag"),
			Size:    extractString(imageData, "Size"),
			Created: extractString(imageData, "CreatedAt"),
		}

		images = append(images, image)
	}

	return images, nil
}

// InspectImage inspects an image
func (d *DockerManager) InspectImage(ctx context.Context, imageID string) (*ImageInfo, error) {
	cmd := exec.CommandContext(ctx, d.command, "inspect", "--format", "{{json .}}", imageID)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect image: %w", err)
	}

	var inspectData map[string]interface{}
	if err := json.Unmarshal(output, &inspectData); err != nil {
		return nil, fmt.Errorf("failed to parse inspect output: %w", err)
	}

	info := &ImageInfo{
		ID:       imageID,
		Tags:     extractStringSlice(inspectData, "RepoTags"),
		Size:     extractString(inspectData, "Size"),
		Created:  extractString(inspectData, "Created"),
		Digest:   extractString(inspectData, "Id"),
		Platform: extractString(inspectData, "Architecture"),
	}

	return info, nil
}

// CreateNetwork creates a network
func (d *DockerManager) CreateNetwork(ctx context.Context, name string, options NetworkOptions) error {
	args := []string{"network", "create"}

	// Add driver
	if options.Driver != "" {
		args = append(args, "--driver", options.Driver)
	}

	// Add subnet
	if options.Subnet != "" {
		args = append(args, "--subnet", options.Subnet)
	}

	// Add gateway
	if options.Gateway != "" {
		args = append(args, "--gateway", options.Gateway)
	}

	// Add IPv6 flag
	if options.IPv6 {
		args = append(args, "--ipv6")
	}

	// Add internal flag
	if options.Internal {
		args = append(args, "--internal")
	}

	// Add attachable flag
	if options.Attachable {
		args = append(args, "--attachable")
	}

	// Add labels
	for key, value := range options.Labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", key, value))
	}

	args = append(args, name)

	cmd := exec.CommandContext(ctx, d.command, args...)
	return cmd.Run()
}

// RemoveNetwork removes a network
func (d *DockerManager) RemoveNetwork(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, d.command, "network", "rm", name)
	return cmd.Run()
}

// ListNetworks lists networks
func (d *DockerManager) ListNetworks(ctx context.Context) ([]NetworkInfo, error) {
	cmd := exec.CommandContext(ctx, d.command, "network", "ls", "--format", "{{json .}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	var networks []NetworkInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var networkData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &networkData); err != nil {
			continue // Skip invalid entries
		}

		network := NetworkInfo{
			ID:     extractString(networkData, "ID"),
			Name:   extractString(networkData, "Name"),
			Driver: extractString(networkData, "Driver"),
			Scope:  extractString(networkData, "Scope"),
		}

		networks = append(networks, network)
	}

	return networks, nil
}

// CreateVolume creates a volume
func (d *DockerManager) CreateVolume(ctx context.Context, name string, options VolumeOptions) error {
	args := []string{"volume", "create"}

	// Add driver
	if options.Driver != "" {
		args = append(args, "--driver", options.Driver)
	}

	// Add driver options
	for key, value := range options.DriverOpts {
		args = append(args, "--opt", fmt.Sprintf("%s=%s", key, value))
	}

	// Add labels
	for key, value := range options.Labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", key, value))
	}

	args = append(args, name)

	cmd := exec.CommandContext(ctx, d.command, args...)
	return cmd.Run()
}

// RemoveVolume removes a volume
func (d *DockerManager) RemoveVolume(ctx context.Context, name string, force bool) error {
	args := []string{"volume", "rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, name)

	cmd := exec.CommandContext(ctx, d.command, args...)
	return cmd.Run()
}

// ListVolumes lists volumes
func (d *DockerManager) ListVolumes(ctx context.Context) ([]VolumeInfo, error) {
	cmd := exec.CommandContext(ctx, d.command, "volume", "ls", "--format", "{{json .}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	var volumes []VolumeInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var volumeData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &volumeData); err != nil {
			continue // Skip invalid entries
		}

		volume := VolumeInfo{
			Name:   extractString(volumeData, "Name"),
			Driver: extractString(volumeData, "Driver"),
		}

		volumes = append(volumes, volume)
	}

	return volumes, nil
}

// Helper functions for extracting data from JSON
func extractString(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func extractStringSlice(data map[string]interface{}, key string) []string {
	if value, ok := data[key]; ok {
		if slice, ok := value.([]interface{}); ok {
			var result []string
			for _, item := range slice {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return []string{}
}

func extractLabels(data map[string]interface{}) map[string]string {
	labels := make(map[string]string)
	if config, ok := data["Config"].(map[string]interface{}); ok {
		if labelsData, ok := config["Labels"].(map[string]interface{}); ok {
			for key, value := range labelsData {
				if str, ok := value.(string); ok {
					labels[key] = str
				}
			}
		}
	}
	return labels
}
