package executor_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/sarrrrry/gh-mrepo/internal/executor"
)

func TestScanLocalRepos_Normal(t *testing.T) {
	root := t.TempDir()
	mkDir(t, root, "octocat/project-a")
	mkDir(t, root, "octocat/project-b")
	mkDir(t, root, "org/repo-x")

	scanner := executor.NewFsScanner()
	repos, err := scanner.ScanLocalRepos(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"octocat/project-a", "octocat/project-b", "org/repo-x"}
	if !equal(repos, want) {
		t.Errorf("repos = %v, want %v", repos, want)
	}
}

func TestScanLocalRepos_SkipsDotDirs(t *testing.T) {
	root := t.TempDir()
	mkDir(t, root, ".hidden/repo")
	mkDir(t, root, "owner/.git")
	mkDir(t, root, "owner/visible-repo")

	scanner := executor.NewFsScanner()
	repos, err := scanner.ScanLocalRepos(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"owner/visible-repo"}
	if !equal(repos, want) {
		t.Errorf("repos = %v, want %v", repos, want)
	}
}

func TestScanLocalRepos_SkipsFiles(t *testing.T) {
	root := t.TempDir()
	mkDir(t, root, "owner/repo")
	// ownerレベルにファイルを作成
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	// repoレベルにファイルを作成
	if err := os.WriteFile(filepath.Join(root, "owner", "notes.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	scanner := executor.NewFsScanner()
	repos, err := scanner.ScanLocalRepos(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"owner/repo"}
	if !equal(repos, want) {
		t.Errorf("repos = %v, want %v", repos, want)
	}
}

func TestScanLocalRepos_EmptyRoot(t *testing.T) {
	root := t.TempDir()

	scanner := executor.NewFsScanner()
	repos, err := scanner.ScanLocalRepos(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repos) != 0 {
		t.Errorf("repos = %v, want empty", repos)
	}
}

func TestScanLocalRepos_NonexistentRoot(t *testing.T) {
	scanner := executor.NewFsScanner()
	_, err := scanner.ScanLocalRepos("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for nonexistent root, got nil")
	}
}

func TestScanLocalRepos_SortedOutput(t *testing.T) {
	root := t.TempDir()
	mkDir(t, root, "zeta/repo")
	mkDir(t, root, "alpha/repo")
	mkDir(t, root, "middle/aaa")
	mkDir(t, root, "middle/zzz")

	scanner := executor.NewFsScanner()
	repos, err := scanner.ScanLocalRepos(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !sort.StringsAreSorted(repos) {
		t.Errorf("repos not sorted: %v", repos)
	}

	want := []string{"alpha/repo", "middle/aaa", "middle/zzz", "zeta/repo"}
	if !equal(repos, want) {
		t.Errorf("repos = %v, want %v", repos, want)
	}
}

// --- helpers ---

func mkDir(t *testing.T, base string, rel string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(base, rel), 0o755); err != nil {
		t.Fatal(err)
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
