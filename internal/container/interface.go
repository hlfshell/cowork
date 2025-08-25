package container

import (
	"context"
	"io"
)

// ContainerEngine represents the type of container engine
type ContainerEngine string

const (
	EngineDocker  ContainerEngine = "docker"
	EnginePodman  ContainerEngine = "podman"
	EngineUnknown ContainerEngine = "unknown"
)

// ContainerInfo represents information about a container
type ContainerInfo struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Status  string            `json:"status"`
	Created string            `json:"created"`
	Ports   []string          `json:"ports"`
	Labels  map[string]string `json:"labels"`
}

// ImageInfo represents information about a container image
type ImageInfo struct {
	ID       string   `json:"id"`
	Tags     []string `json:"tags"`
	Size     string   `json:"size"`
	Created  string   `json:"created"`
	Digest   string   `json:"digest"`
	Platform string   `json:"platform"`
}

// RunOptions represents options for running a container
type RunOptions struct {
	Image       string            `json:"image"`
	Name        string            `json:"name"`
	Command     []string          `json:"command"`
	Args        []string          `json:"args"`
	WorkingDir  string            `json:"working_dir"`
	Environment map[string]string `json:"environment"`
	Ports       map[string]string `json:"ports"`
	Volumes     map[string]string `json:"volumes"`
	Network     string            `json:"network"`
	User        string            `json:"user"`
	Privileged  bool              `json:"privileged"`
	Detached    bool              `json:"detached"`
	Remove      bool              `json:"remove"`
	TTY         bool              `json:"tty"`
	Interactive bool              `json:"interactive"`
	Labels      map[string]string `json:"labels"`
}

// BuildOptions represents options for building a container image
type BuildOptions struct {
	Context    string            `json:"context"`
	Dockerfile string            `json:"dockerfile"`
	Tags       []string          `json:"tags"`
	BuildArgs  map[string]string `json:"build_args"`
	Labels     map[string]string `json:"labels"`
	NoCache    bool              `json:"no_cache"`
	Pull       bool              `json:"pull"`
	Platform   string            `json:"platform"`
}

// ContainerManager defines the interface for container operations
type ContainerManager interface {
	// Engine information
	GetEngine() ContainerEngine
	GetVersion() (string, error)
	IsAvailable() bool

	// Container operations
	Run(ctx context.Context, options RunOptions) (string, error)
	Start(ctx context.Context, containerID string) error
	Stop(ctx context.Context, containerID string, timeoutSeconds int) error
	Remove(ctx context.Context, containerID string, force bool) error
	Exec(ctx context.Context, containerID string, command []string, options ExecOptions) error
	Logs(ctx context.Context, containerID string, options LogOptions) (io.ReadCloser, error)
	Inspect(ctx context.Context, containerID string) (*ContainerInfo, error)
	List(ctx context.Context, options ListOptions) ([]ContainerInfo, error)

	// Image operations
	Pull(ctx context.Context, image string) error
	Build(ctx context.Context, options BuildOptions) (string, error)
	RemoveImage(ctx context.Context, imageID string, force bool) error
	ListImages(ctx context.Context) ([]ImageInfo, error)
	InspectImage(ctx context.Context, imageID string) (*ImageInfo, error)

	// Network operations
	CreateNetwork(ctx context.Context, name string, options NetworkOptions) error
	RemoveNetwork(ctx context.Context, name string) error
	ListNetworks(ctx context.Context) ([]NetworkInfo, error)

	// Volume operations
	CreateVolume(ctx context.Context, name string, options VolumeOptions) error
	RemoveVolume(ctx context.Context, name string, force bool) error
	ListVolumes(ctx context.Context) ([]VolumeInfo, error)
}

// ExecOptions represents options for executing commands in a container
type ExecOptions struct {
	User        string            `json:"user"`
	WorkingDir  string            `json:"working_dir"`
	Environment map[string]string `json:"environment"`
	TTY         bool              `json:"tty"`
	Interactive bool              `json:"interactive"`
	Privileged  bool              `json:"privileged"`
}

// LogOptions represents options for retrieving container logs
type LogOptions struct {
	Follow     bool   `json:"follow"`
	Timestamps bool   `json:"timestamps"`
	Tail       int    `json:"tail"`
	Since      string `json:"since"`
	Until      string `json:"until"`
}

// ListOptions represents options for listing containers
type ListOptions struct {
	All     bool              `json:"all"`
	Filters map[string]string `json:"filters"`
	Limit   int               `json:"limit"`
	Size    bool              `json:"size"`
}

// NetworkInfo represents information about a network
type NetworkInfo struct {
	ID         string                      `json:"id"`
	Name       string                      `json:"name"`
	Driver     string                      `json:"driver"`
	Scope      string                      `json:"scope"`
	IPv6       bool                        `json:"ipv6"`
	Internal   bool                        `json:"internal"`
	Attachable bool                        `json:"attachable"`
	Ingress    bool                        `json:"ingress"`
	Labels     map[string]string           `json:"labels"`
	Containers map[string]NetworkContainer `json:"containers"`
}

// NetworkContainer represents a container in a network
type NetworkContainer struct {
	Name        string `json:"name"`
	Endpoint    string `json:"endpoint"`
	IPv4Address string `json:"ipv4_address"`
	IPv6Address string `json:"ipv6_address"`
	MacAddress  string `json:"mac_address"`
}

// NetworkOptions represents options for creating a network
type NetworkOptions struct {
	Driver     string            `json:"driver"`
	Subnet     string            `json:"subnet"`
	Gateway    string            `json:"gateway"`
	IPv6       bool              `json:"ipv6"`
	Internal   bool              `json:"internal"`
	Attachable bool              `json:"attachable"`
	Labels     map[string]string `json:"labels"`
}

// VolumeInfo represents information about a volume
type VolumeInfo struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Mountpoint string            `json:"mountpoint"`
	Created    string            `json:"created"`
	Status     map[string]string `json:"status"`
	Labels     map[string]string `json:"labels"`
	Scope      string            `json:"scope"`
}

// VolumeOptions represents options for creating a volume
type VolumeOptions struct {
	Driver     string            `json:"driver"`
	DriverOpts map[string]string `json:"driver_opts"`
	Labels     map[string]string `json:"labels"`
}
