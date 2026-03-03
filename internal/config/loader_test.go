package config_test

import (
	"errors"
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
	if err := os.WriteFile(tomlPath, []byte(content), 0o644); err != nil {
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

func TestLoad_GitConfigFields(t *testing.T) {
	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "config.toml")
	content := `
[work]
gh_config_dir = "/home/user/.config/gh-work"
root = "/home/user/repos"
git_config_name = "Work User"
git_config_email = "work@example.com"

[personal]
gh_config_dir = "/home/user/.config/gh-personal"
`
	if err := os.WriteFile(tomlPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(tomlPath)
	profiles, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := make(map[string]domain.Profile)
	for _, p := range profiles {
		m[p.Name] = p
	}

	work := m["work"]
	if work.GitConfigName != "Work User" {
		t.Errorf("work.GitConfigName = %q, want %q", work.GitConfigName, "Work User")
	}
	if work.GitConfigEmail != "work@example.com" {
		t.Errorf("work.GitConfigEmail = %q, want %q", work.GitConfigEmail, "work@example.com")
	}

	personal := m["personal"]
	if personal.GitConfigName != "" {
		t.Errorf("personal.GitConfigName = %q, want empty", personal.GitConfigName)
	}
	if personal.GitConfigEmail != "" {
		t.Errorf("personal.GitConfigEmail = %q, want empty", personal.GitConfigEmail)
	}
}

func TestLoad_SSHIdentity(t *testing.T) {
	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "config.toml")
	content := `
[work]
gh_config_dir = "/home/user/.config/gh-work"
root = "/home/user/repos"
ssh_identity = "/home/user/.ssh/id_ed25519_work"

[personal]
gh_config_dir = "/home/user/.config/gh-personal"
`
	if err := os.WriteFile(tomlPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(tomlPath)
	profiles, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := make(map[string]domain.Profile)
	for _, p := range profiles {
		m[p.Name] = p
	}

	work := m["work"]
	if work.SSHIdentity != "/home/user/.ssh/id_ed25519_work" {
		t.Errorf("work.SSHIdentity = %q, want %q", work.SSHIdentity, "/home/user/.ssh/id_ed25519_work")
	}

	personal := m["personal"]
	if personal.SSHIdentity != "" {
		t.Errorf("personal.SSHIdentity = %q, want empty", personal.SSHIdentity)
	}
}

func TestLoad_SSHIdentityTildeExpansion(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home dir")
	}

	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "config.toml")
	content := `
[work]
gh_config_dir = "/home/user/.config/gh-work"
ssh_identity = "~/.ssh/id_ed25519_work"
`
	if err := os.WriteFile(tomlPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(tomlPath)
	profiles, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join(home, ".ssh/id_ed25519_work")
	if profiles[0].SSHIdentity != want {
		t.Errorf("SSHIdentity = %q, want %q", profiles[0].SSHIdentity, want)
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
	if err := os.WriteFile(tomlPath, []byte(content), 0o644); err != nil {
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
	if err := os.WriteFile(tomlPath, []byte(content), 0o644); err != nil {
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
	if err := os.WriteFile(tomlPath, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(tomlPath)
	_, err := loader.Load()
	if err != domain.ErrNoProfiles {
		t.Errorf("err = %v, want %v", err, domain.ErrNoProfiles)
	}
}

// --- include機能のテスト ---

func TestLoad_Include(t *testing.T) {
	dir := t.TempDir()

	// includeされるファイル
	workPath := filepath.Join(dir, "work.toml")
	workContent := `
[work]
gh_config_dir = "/home/user/.config/gh-work"
root = "/home/user/repos/work"
`
	if err := os.WriteFile(workPath, []byte(workContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// メインconfig
	mainPath := filepath.Join(dir, "config.toml")
	mainContent := `
include = ["` + workPath + `"]

[personal]
gh_config_dir = "/home/user/.config/gh-personal"
root = "/home/user/repos/personal"
`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(mainPath)
	profiles, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("len(profiles) = %d, want 2", len(profiles))
	}

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
	if work.Root != "/home/user/repos/work" {
		t.Errorf("work.Root = %q", work.Root)
	}

	personal, ok := m["personal"]
	if !ok {
		t.Fatal("profile 'personal' not found")
	}
	if personal.GHConfigDir != "/home/user/.config/gh-personal" {
		t.Errorf("personal.GHConfigDir = %q", personal.GHConfigDir)
	}
}

func TestLoad_IncludeMultipleFiles(t *testing.T) {
	dir := t.TempDir()

	// work.toml
	workPath := filepath.Join(dir, "work.toml")
	workContent := `
[work]
gh_config_dir = "/home/user/.config/gh-work"
root = "/home/user/repos/work"
`
	if err := os.WriteFile(workPath, []byte(workContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// personal.toml
	personalPath := filepath.Join(dir, "personal.toml")
	personalContent := `
[personal]
gh_config_dir = "/home/user/.config/gh-personal"
root = "/home/user/repos/personal"
`
	if err := os.WriteFile(personalPath, []byte(personalContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// メインconfig (プロファイルなし、includeのみ)
	mainPath := filepath.Join(dir, "config.toml")
	mainContent := `include = ["` + workPath + `", "` + personalPath + `"]
`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(mainPath)
	profiles, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("len(profiles) = %d, want 2", len(profiles))
	}

	m := make(map[string]domain.Profile)
	for _, p := range profiles {
		m[p.Name] = p
	}

	if _, ok := m["work"]; !ok {
		t.Error("profile 'work' not found")
	}
	if _, ok := m["personal"]; !ok {
		t.Error("profile 'personal' not found")
	}
}

func TestLoad_IncludeRelativePath(t *testing.T) {
	dir := t.TempDir()

	// サブディレクトリにincludeファイルを配置
	subdir := filepath.Join(dir, "includes")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}

	workPath := filepath.Join(subdir, "work.toml")
	workContent := `
[work]
gh_config_dir = "/home/user/.config/gh-work"
root = "/home/user/repos/work"
`
	if err := os.WriteFile(workPath, []byte(workContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// 相対パスでinclude
	mainPath := filepath.Join(dir, "config.toml")
	mainContent := `
include = ["includes/work.toml"]

[personal]
gh_config_dir = "/home/user/.config/gh-personal"
`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(mainPath)
	profiles, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("len(profiles) = %d, want 2", len(profiles))
	}

	m := make(map[string]domain.Profile)
	for _, p := range profiles {
		m[p.Name] = p
	}

	if _, ok := m["work"]; !ok {
		t.Error("profile 'work' not found")
	}
	if _, ok := m["personal"]; !ok {
		t.Error("profile 'personal' not found")
	}
}

func TestLoad_IncludeTildePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home dir")
	}

	// ホームディレクトリ配下にテスト用ファイルを作成
	testDir := filepath.Join(home, ".gh-mrepo-test-include")
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(testDir) })

	includePath := filepath.Join(testDir, "extra.toml")
	includeContent := `
[extra]
gh_config_dir = "/home/user/.config/gh-extra"
`
	if err := os.WriteFile(includePath, []byte(includeContent), 0o644); err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	mainPath := filepath.Join(dir, "config.toml")
	mainContent := `
include = ["~/.gh-mrepo-test-include/extra.toml"]

[main]
gh_config_dir = "/home/user/.config/gh-main"
`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(mainPath)
	profiles, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("len(profiles) = %d, want 2", len(profiles))
	}

	m := make(map[string]domain.Profile)
	for _, p := range profiles {
		m[p.Name] = p
	}

	if _, ok := m["extra"]; !ok {
		t.Error("profile 'extra' not found")
	}
	if _, ok := m["main"]; !ok {
		t.Error("profile 'main' not found")
	}
}

func TestLoad_IncludeDuplicateProfile(t *testing.T) {
	dir := t.TempDir()

	// includeファイルにも"work"プロファイルがある
	incPath := filepath.Join(dir, "extra.toml")
	incContent := `
[work]
gh_config_dir = "/home/user/.config/gh-work-other"
`
	if err := os.WriteFile(incPath, []byte(incContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// メインにも"work"プロファイルがある → 重複エラー
	mainPath := filepath.Join(dir, "config.toml")
	mainContent := `
include = ["` + incPath + `"]

[work]
gh_config_dir = "/home/user/.config/gh-work"
`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(mainPath)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrDuplicateProfile) {
		t.Errorf("err = %v, want %v", err, domain.ErrDuplicateProfile)
	}
}

func TestLoad_IncludeDuplicateProfileBetweenFiles(t *testing.T) {
	dir := t.TempDir()

	// 2つのincludeファイルに同名プロファイル
	inc1Path := filepath.Join(dir, "a.toml")
	inc1Content := `
[shared]
gh_config_dir = "/home/user/.config/gh-a"
`
	if err := os.WriteFile(inc1Path, []byte(inc1Content), 0o644); err != nil {
		t.Fatal(err)
	}

	inc2Path := filepath.Join(dir, "b.toml")
	inc2Content := `
[shared]
gh_config_dir = "/home/user/.config/gh-b"
`
	if err := os.WriteFile(inc2Path, []byte(inc2Content), 0o644); err != nil {
		t.Fatal(err)
	}

	mainPath := filepath.Join(dir, "config.toml")
	mainContent := `include = ["` + inc1Path + `", "` + inc2Path + `"]
`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(mainPath)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrDuplicateProfile) {
		t.Errorf("err = %v, want %v", err, domain.ErrDuplicateProfile)
	}
}

func TestLoad_IncludeFileNotFound(t *testing.T) {
	dir := t.TempDir()

	mainPath := filepath.Join(dir, "config.toml")
	mainContent := `
include = ["/nonexistent/path/extra.toml"]

[work]
gh_config_dir = "/home/user/.config/gh-work"
`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(mainPath)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoad_IncludeEmpty(t *testing.T) {
	dir := t.TempDir()

	mainPath := filepath.Join(dir, "config.toml")
	mainContent := `
include = []

[work]
gh_config_dir = "/home/user/.config/gh-work"
root = "/home/user/repos"
`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(mainPath)
	profiles, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("len(profiles) = %d, want 1", len(profiles))
	}
	if profiles[0].Name != "work" {
		t.Errorf("profiles[0].Name = %q, want %q", profiles[0].Name, "work")
	}
}

func TestLoad_NoInclude(t *testing.T) {
	// includeキーなしで既存動作が壊れないことを確認(後方互換)
	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "config.toml")
	content := `
[work]
gh_config_dir = "/home/user/.config/gh-work"
root = "/home/user/repos"
git_config_name = "Work User"
git_config_email = "work@example.com"
ssh_identity = "/home/user/.ssh/id_ed25519_work"

[personal]
gh_config_dir = "/home/user/.config/gh-personal"
`
	if err := os.WriteFile(tomlPath, []byte(content), 0o644); err != nil {
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

	m := make(map[string]domain.Profile)
	for _, p := range profiles {
		m[p.Name] = p
	}

	work := m["work"]
	if work.GHConfigDir != "/home/user/.config/gh-work" {
		t.Errorf("work.GHConfigDir = %q", work.GHConfigDir)
	}
	if work.Root != "/home/user/repos" {
		t.Errorf("work.Root = %q", work.Root)
	}
	if work.GitConfigName != "Work User" {
		t.Errorf("work.GitConfigName = %q", work.GitConfigName)
	}
	if work.GitConfigEmail != "work@example.com" {
		t.Errorf("work.GitConfigEmail = %q", work.GitConfigEmail)
	}
	if work.SSHIdentity != "/home/user/.ssh/id_ed25519_work" {
		t.Errorf("work.SSHIdentity = %q", work.SSHIdentity)
	}

	personal := m["personal"]
	if personal.GHConfigDir != "/home/user/.config/gh-personal" {
		t.Errorf("personal.GHConfigDir = %q", personal.GHConfigDir)
	}
}
