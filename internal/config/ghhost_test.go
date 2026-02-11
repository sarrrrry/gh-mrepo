package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sarrrrry/gh-mrepo/internal/config"
)

func TestResolveGitHubUser(t *testing.T) {
	t.Run("正常にuserを取得できる", func(t *testing.T) {
		dir := t.TempDir()
		writeHostsYml(t, dir, "github.com:\n    user: sarrrrry\n")

		got, err := config.ResolveGitHubUser(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "sarrrrry" {
			t.Errorf("got %q, want %q", got, "sarrrrry")
		}
	})

	t.Run("hosts.ymlが存在しない", func(t *testing.T) {
		dir := t.TempDir()

		_, err := config.ResolveGitHubUser(dir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("github.comエントリがない", func(t *testing.T) {
		dir := t.TempDir()
		writeHostsYml(t, dir, "gitlab.com:\n    user: someone\n")

		_, err := config.ResolveGitHubUser(dir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("userが空", func(t *testing.T) {
		dir := t.TempDir()
		writeHostsYml(t, dir, "github.com:\n    user: \"\"\n")

		_, err := config.ResolveGitHubUser(dir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func writeHostsYml(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "hosts.yml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
