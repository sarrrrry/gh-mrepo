package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sarrrrry/gh-mrepo/internal/config"
	"github.com/sarrrrry/gh-mrepo/internal/domain"
)

func TestLoad_Success(t *testing.T) {
	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "config.toml")
	content := `
[work]
gh_config_dir = "/home/user/.config/gh-work"
root = "/home/user/repos"

[personal]
gh_config_dir = "/home/user/.config/gh-personal"
`
	if err := os.WriteFile(tomlPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(tomlPath)
	profiles, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("len(profiles) = %d, want 2", len(profiles))
	}

	// プロファイルをマップに変換して順序非依存にテスト
	m := make(map[string]domain.Profile)
	for _, p := range profiles {
		m[p.Name] = p
	}

	work, ok := m["work"]
	if !ok {
		t.Fatal("profile 'work' not found")
	}
	if work.GHConfigDir != "/home/user/.config/gh-work" {
		t.Errorf("work.GHConfigDir = %q", work.GHConfigDir)
	}
	if work.Root != "/home/user/repos" {
		t.Errorf("work.Root = %q", work.Root)
	}

	personal, ok := m["personal"]
	if !ok {
		t.Fatal("profile 'personal' not found")
	}
	if personal.GHConfigDir != "/home/user/.config/gh-personal" {
		t.Errorf("personal.GHConfigDir = %q", personal.GHConfigDir)
	}
	if personal.Root != "" {
		t.Errorf("personal.Root = %q, want empty", personal.Root)
	}
}

func TestLoad_TildeExpansion(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home dir")
	}

	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "config.toml")
	content := `
[work]
gh_config_dir = "~/.config/gh-work"
root = "~/repos"
`
	if err := os.WriteFile(tomlPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(tomlPath)
	profiles, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("len(profiles) = %d, want 1", len(profiles))
	}

	p := profiles[0]
	want := filepath.Join(home, ".config/gh-work")
	if p.GHConfigDir != want {
		t.Errorf("GHConfigDir = %q, want %q", p.GHConfigDir, want)
	}
	wantRoot := filepath.Join(home, "repos")
	if p.Root != wantRoot {
		t.Errorf("Root = %q, want %q", p.Root, wantRoot)
	}
}

func TestLoad_EmptyGHConfigDir(t *testing.T) {
	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "config.toml")
	content := `
[work]
gh_config_dir = ""
`
	if err := os.WriteFile(tomlPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(tomlPath)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	loader := config.NewLoader("/nonexistent/path/config.toml")
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoad_NoProfiles(t *testing.T) {
	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "config.toml")
	// 空のTOML
	if err := os.WriteFile(tomlPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(tomlPath)
	_, err := loader.Load()
	if err != domain.ErrNoProfiles {
		t.Errorf("err = %v, want %v", err, domain.ErrNoProfiles)
	}
}
