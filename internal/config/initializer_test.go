package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sarrrrry/gh-mrepo/internal/config"
	"github.com/sarrrrry/gh-mrepo/internal/domain"
)

func TestInit_CreatesTemplateConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	init := config.NewInitializer()
	if err := init.Init(configPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	want := `# gh-mrepo configuration
# See: gh mrepo --help

[default]
gh_config_dir = "~/.config/gh"
# root = "~/repos"
`
	if string(data) != want {
		t.Errorf("content mismatch\ngot:\n%s\nwant:\n%s", string(data), want)
	}
}

func TestInit_CreatesParentDirectories(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "nested", "deep", "config.toml")

	init := config.NewInitializer()
	if err := init.Init(configPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("config file not found: %v", err)
	}
}

func TestInit_ErrorWhenConfigAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	if err := os.WriteFile(configPath, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	init := config.NewInitializer()
	err := init.Init(configPath)
	if !errors.Is(err, domain.ErrConfigAlreadyExists) {
		t.Errorf("err = %v, want %v", err, domain.ErrConfigAlreadyExists)
	}
}
