package types

// GitAuthMethod represents the git authentication method
type GitAuthMethod string

const (
	GitAuthMethodNone  GitAuthMethod = "none"
	GitAuthMethodSSH   GitAuthMethod = "ssh"
	GitAuthMethodHTTPS GitAuthMethod = "https"
)

// GitAuthInfo contains authentication information for git operations
type GitAuthInfo struct {
	Method     GitAuthMethod
	SSHKeyPath string
	Username   string
	Password   string
	Token      string
}
