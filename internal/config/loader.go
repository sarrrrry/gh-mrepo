package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/sarrrrry/gh-mrepo/internal/domain"
)

type profileEntry struct {
	GHConfigDir    string `toml:"gh_config_dir"`
	Root           string `toml:"root"`
	GitConfigName  string `toml:"git_config_name"`
	GitConfigEmail string `toml:"git_config_email"`
}

type Loader struct {
	path string
}

func NewLoader(path string) *Loader {
	return &Loader{path: path}
}

func (l *Loader) Load() ([]domain.Profile, error) {
	var raw map[string]profileEntry
	if _, err := toml.DecodeFile(l.path, &raw); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if len(raw) == 0 {
		return nil, domain.ErrNoProfiles
	}

	// ソートして安定した順序を保証
	names := make([]string, 0, len(raw))
	for name := range raw {
		names = append(names, name)
	}
	sort.Strings(names)

	profiles := make([]domain.Profile, 0, len(raw))
	for _, name := range names {
		entry := raw[name]
		ghConfigDir, err := expandTilde(entry.GHConfigDir)
		if err != nil {
			return nil, err
		}
		root, err := expandTilde(entry.Root)
		if err != nil {
			return nil, err
		}

		p, err := domain.NewProfile(name, ghConfigDir, root)
		if err != nil {
			return nil, fmt.Errorf("profile %q: %w", name, err)
		}
		p.GitConfigName = entry.GitConfigName
		p.GitConfigEmail = entry.GitConfigEmail
		profiles = append(profiles, p)
	}

	return profiles, nil
}

func expandTilde(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to expand ~: %w", err)
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}
