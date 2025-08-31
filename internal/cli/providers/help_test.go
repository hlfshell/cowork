package providers

import (
	"strings"
	"testing"
)

func TestGetProviderHelp(t *testing.T) {
	tests := []struct {
		name           string
		providerName   string
		expectExists   bool
		expectContains string
	}{
		{
			name:           "GitHub provider",
			providerName:   "github",
			expectExists:   true,
			expectContains: "GitHub Authentication Help",
		},
		{
			name:           "GitLab provider",
			providerName:   "gitlab",
			expectExists:   true,
			expectContains: "GitLab Authentication Help",
		},
		{
			name:           "Bitbucket provider",
			providerName:   "bitbucket",
			expectExists:   true,
			expectContains: "Bitbucket Authentication Help",
		},
		{
			name:         "Invalid provider",
			providerName: "invalid",
			expectExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, exists := GetProviderHelp(tt.providerName)

			if exists != tt.expectExists {
				t.Errorf("GetProviderHelp(%s) exists = %v, want %v", tt.providerName, exists, tt.expectExists)
			}

			if exists && !strings.Contains(content, tt.expectContains) {
				t.Errorf("GetProviderHelp(%s) content does not contain expected text '%s'", tt.providerName, tt.expectContains)
			}

			if !exists && content != "" {
				t.Errorf("GetProviderHelp(%s) returned content when it should not exist", tt.providerName)
			}
		})
	}
}

func TestGetAvailableProviders(t *testing.T) {
	providers := GetAvailableProviders()

	expectedProviders := []string{"github", "gitlab", "bitbucket"}

	if len(providers) != len(expectedProviders) {
		t.Errorf("GetAvailableProviders() returned %d providers, want %d", len(providers), len(expectedProviders))
	}

	for _, expected := range expectedProviders {
		found := false
		for _, provider := range providers {
			if provider == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetAvailableProviders() missing expected provider: %s", expected)
		}
	}
}
