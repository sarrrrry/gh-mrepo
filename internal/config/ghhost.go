package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// HostResolver は app.UserResolver を満たすアダプタ。
type HostResolver struct{}

func NewHostResolver() *HostResolver {
	return &HostResolver{}
}

func (h *HostResolver) ResolveGitHubUser(ghConfigDir string) (string, error) {
	return ResolveGitHubUser(ghConfigDir)
}

// ResolveGitHubUser は ghConfigDir/hosts.yml を読んで github.com の user を返す。
func ResolveGitHubUser(ghConfigDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(ghConfigDir, "hosts.yml"))
	if err != nil {
		return "", fmt.Errorf("failed to read hosts.yml: %w", err)
	}

	var hosts map[string]struct {
		User string `yaml:"user"`
	}
	if err := yaml.Unmarshal(data, &hosts); err != nil {
		return "", fmt.Errorf("failed to parse hosts.yml: %w", err)
	}

	entry, ok := hosts["github.com"]
	if !ok {
		return "", fmt.Errorf("github.com entry not found in hosts.yml")
	}
	if entry.User == "" {
		return "", fmt.Errorf("user is empty for github.com in hosts.yml")
	}
	return entry.User, nil
}
