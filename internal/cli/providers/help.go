package providers

import (
	_ "embed"
)

//go:embed github.md
var githubHelp string

//go:embed gitlab.md
var gitlabHelp string

//go:embed bitbucket.md
var bitbucketHelp string

// GetProviderHelp returns the help content for the specified provider
func GetProviderHelp(providerName string) (string, bool) {
	switch providerName {
	case "github":
		return githubHelp, true
	case "gitlab":
		return gitlabHelp, true
	case "bitbucket":
		return bitbucketHelp, true
	default:
		return "", false
	}
}

// GetAvailableProviders returns a list of available provider names
func GetAvailableProviders() []string {
	return []string{"github", "gitlab", "bitbucket"}
}
