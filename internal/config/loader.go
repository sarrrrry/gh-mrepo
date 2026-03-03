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
	SSHIdentity    string `toml:"ssh_identity"`
}

type Loader struct {
	path string
}

func NewLoader(path string) *Loader {
	return &Loader{path: path}
}

// loadFile は1つのTOMLファイルから include パスとプロファイルを取り出す
func (l *Loader) loadFile(path string) ([]string, map[string]profileEntry, error) {
	var raw map[string]toml.Primitive
	md, err := toml.DecodeFile(path, &raw)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config %q: %w", path, err)
	}

	var includes []string
	if prim, ok := raw["include"]; ok {
		if err := md.PrimitiveDecode(prim, &includes); err != nil {
			return nil, nil, fmt.Errorf("invalid include value in %q: %w", path, err)
		}
		delete(raw, "include")
	}

	profiles := make(map[string]profileEntry, len(raw))
	for name, prim := range raw {
		var entry profileEntry
		if err := md.PrimitiveDecode(prim, &entry); err != nil {
			return nil, nil, fmt.Errorf("profile %q in %q: %w", name, path, err)
		}
		profiles[name] = entry
	}

	return includes, profiles, nil
}

func (l *Loader) Load() ([]domain.Profile, error) {
	includes, mainProfiles, err := l.loadFile(l.path)
	if err != nil {
		return nil, err
	}

	// includeファイルのプロファイルを収集
	merged := make(map[string]profileEntry)
	for _, inc := range includes {
		resolved, err := l.resolveIncludePath(inc)
		if err != nil {
			return nil, err
		}
		_, fileProfiles, err := l.loadFile(resolved)
		if err != nil {
			return nil, err
		}
		for name, entry := range fileProfiles {
			if _, exists := merged[name]; exists {
				return nil, fmt.Errorf("profile %q: %w", name, domain.ErrDuplicateProfile)
			}
			merged[name] = entry
		}
	}

	// メインファイルのプロファイルをマージ (重複チェック)
	for name, entry := range mainProfiles {
		if _, exists := merged[name]; exists {
			return nil, fmt.Errorf("profile %q: %w", name, domain.ErrDuplicateProfile)
		}
		merged[name] = entry
	}

	if len(merged) == 0 {
		return nil, domain.ErrNoProfiles
	}

	// ソートして安定した順序を保証
	names := make([]string, 0, len(merged))
	for name := range merged {
		names = append(names, name)
	}
	sort.Strings(names)

	profiles := make([]domain.Profile, 0, len(merged))
	for _, name := range names {
		entry := merged[name]
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

		sshIdentity, err := expandTilde(entry.SSHIdentity)
		if err != nil {
			return nil, err
		}
		p.SSHIdentity = sshIdentity

		profiles = append(profiles, p)
	}

	return profiles, nil
}

// resolveIncludePath はチルダ展開と相対パスのconfig基準解決を行う
func (l *Loader) resolveIncludePath(path string) (string, error) {
	expanded, err := expandTilde(path)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(expanded) {
		expanded = filepath.Join(filepath.Dir(l.path), expanded)
	}
	return expanded, nil
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
