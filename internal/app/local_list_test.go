package app_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/sarrrrry/gh-mrepo/internal/app"
	"github.com/sarrrrry/gh-mrepo/internal/domain"
)

// --- mock for DirScanner ---

type mockScanner struct {
	repos map[string][]string // root -> repos
	errs  map[string]error
}

func (m *mockScanner) ScanLocalRepos(root string) ([]string, error) {
	if e, ok := m.errs[root]; ok {
		return nil, e
	}
	return m.repos[root], nil
}

// --- テストケース ---

func TestLocalList_MultipleProfiles(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work", Root: "/home/work"}
	personal := domain.Profile{Name: "personal", GHConfigDir: "/path/personal", Root: "/home/personal"}

	loader := &mockLoader{profiles: []domain.Profile{work, personal}}
	resolver := &mockResolver{
		users: map[string]string{
			"/path/work":     "octocat-work",
			"/path/personal": "octocat",
		},
		err: map[string]error{},
	}
	scanner := &mockScanner{
		repos: map[string][]string{
			"/home/work":     {"octocat-work/project-a", "octocat-work/project-b"},
			"/home/personal": {"octocat/dotfiles"},
		},
		errs: map[string]error{},
	}

	lister := app.NewLocalLister(loader, resolver, scanner)
	var buf bytes.Buffer
	err := lister.ListLocal(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "work") || !strings.Contains(out, "@octocat-work") {
		t.Errorf("output missing work header, got:\n%s", out)
	}
	if !strings.Contains(out, "personal") || !strings.Contains(out, "@octocat") {
		t.Errorf("output missing personal header, got:\n%s", out)
	}
	if !strings.Contains(out, "octocat-work/project-a") {
		t.Errorf("output missing work repos, got:\n%s", out)
	}
	if !strings.Contains(out, "octocat/dotfiles") {
		t.Errorf("output missing personal repos, got:\n%s", out)
	}

	// workがpersonalより先に出力されること(定義順)
	workIdx := strings.Index(out, "work")
	personalIdx := strings.Index(out, "personal")
	if workIdx > personalIdx {
		t.Errorf("work should appear before personal in output")
	}
}

func TestLocalList_EmptyRoot_ShowsError(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work", Root: ""}

	loader := &mockLoader{profiles: []domain.Profile{work}}
	resolver := &mockResolver{
		users: map[string]string{"/path/work": "octocat-work"},
		err:   map[string]error{},
	}
	scanner := &mockScanner{
		repos: map[string][]string{},
		errs:  map[string]error{},
	}

	lister := app.NewLocalLister(loader, resolver, scanner)
	var buf bytes.Buffer
	err := lister.ListLocal(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "root not configured") {
		t.Errorf("output should contain 'root not configured', got:\n%s", out)
	}
}

func TestLocalList_ScannerError_ContinuesOthers(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work", Root: "/home/work"}
	personal := domain.Profile{Name: "personal", GHConfigDir: "/path/personal", Root: "/home/personal"}

	loader := &mockLoader{profiles: []domain.Profile{work, personal}}
	resolver := &mockResolver{
		users: map[string]string{
			"/path/work":     "octocat-work",
			"/path/personal": "octocat",
		},
		err: map[string]error{},
	}
	scanner := &mockScanner{
		repos: map[string][]string{
			"/home/personal": {"octocat/dotfiles"},
		},
		errs: map[string]error{
			"/home/work": errors.New("permission denied"),
		},
	}

	lister := app.NewLocalLister(loader, resolver, scanner)
	var buf bytes.Buffer
	err := lister.ListLocal(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "permission denied") {
		t.Errorf("output should contain scanner error, got:\n%s", out)
	}
	if !strings.Contains(out, "octocat/dotfiles") {
		t.Errorf("output should contain personal repos, got:\n%s", out)
	}
}

func TestLocalList_ResolverError_ContinuesOthers(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work", Root: "/home/work"}
	personal := domain.Profile{Name: "personal", GHConfigDir: "/path/personal", Root: "/home/personal"}

	loader := &mockLoader{profiles: []domain.Profile{work, personal}}
	resolver := &mockResolver{
		users: map[string]string{
			"/path/personal": "octocat",
		},
		err: map[string]error{
			"/path/work": errors.New("hosts.yml not found"),
		},
	}
	scanner := &mockScanner{
		repos: map[string][]string{
			"/home/work":     {"octocat-work/project-a"},
			"/home/personal": {"octocat/dotfiles"},
		},
		errs: map[string]error{},
	}

	lister := app.NewLocalLister(loader, resolver, scanner)
	var buf bytes.Buffer
	err := lister.ListLocal(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	// resolver失敗でもスキャン結果は表示される
	if !strings.Contains(out, "octocat-work/project-a") {
		t.Errorf("output should contain work repos even with resolver error, got:\n%s", out)
	}
	if !strings.Contains(out, "octocat/dotfiles") {
		t.Errorf("output should contain personal repos, got:\n%s", out)
	}
}

func TestLocalList_EmptyDirectory_ShowsNoRepositories(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work", Root: "/home/work"}

	loader := &mockLoader{profiles: []domain.Profile{work}}
	resolver := &mockResolver{
		users: map[string]string{"/path/work": "octocat-work"},
		err:   map[string]error{},
	}
	scanner := &mockScanner{
		repos: map[string][]string{"/home/work": {}},
		errs:  map[string]error{},
	}

	lister := app.NewLocalLister(loader, resolver, scanner)
	var buf bytes.Buffer
	err := lister.ListLocal(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No repositories") {
		t.Errorf("output should contain 'No repositories', got:\n%s", out)
	}
}

func TestLocalList_LoaderError(t *testing.T) {
	loaderErr := errors.New("load failed")
	loader := &mockLoader{err: loaderErr}
	resolver := &mockResolver{}
	scanner := &mockScanner{}

	lister := app.NewLocalLister(loader, resolver, scanner)
	var buf bytes.Buffer
	err := lister.ListLocal(&buf)
	if !errors.Is(err, loaderErr) {
		t.Errorf("err = %v, want %v", err, loaderErr)
	}
}

// --- ListLocalProfile テスト ---

func TestListLocalProfile_Normal(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work", Root: "/home/work"}

	resolver := &mockResolver{
		users: map[string]string{"/path/work": "octocat-work"},
		err:   map[string]error{},
	}
	scanner := &mockScanner{
		repos: map[string][]string{
			"/home/work": {"octocat-work/project-a", "octocat-work/project-b"},
		},
		errs: map[string]error{},
	}

	lister := app.NewLocalLister(nil, resolver, scanner)
	var buf bytes.Buffer
	err := lister.ListLocalProfile(work, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "work") || !strings.Contains(out, "@octocat-work") {
		t.Errorf("output missing header, got:\n%s", out)
	}
	if !strings.Contains(out, "octocat-work/project-a") {
		t.Errorf("output missing repos, got:\n%s", out)
	}
}

func TestListLocalProfile_EmptyRoot(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work", Root: ""}

	resolver := &mockResolver{
		users: map[string]string{"/path/work": "octocat-work"},
		err:   map[string]error{},
	}
	scanner := &mockScanner{}

	lister := app.NewLocalLister(nil, resolver, scanner)
	var buf bytes.Buffer
	err := lister.ListLocalProfile(work, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "root not configured") {
		t.Errorf("output should contain 'root not configured', got:\n%s", out)
	}
}

// --- CollectLocalRepos テスト ---

func TestCollectLocalRepos_MultipleProfiles(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work", Root: "/home/work"}
	personal := domain.Profile{Name: "personal", GHConfigDir: "/path/personal", Root: "/home/personal"}

	scanner := &mockScanner{
		repos: map[string][]string{
			"/home/work":     {"octocat-work/project-a", "octocat-work/project-b"},
			"/home/personal": {"octocat/dotfiles"},
		},
		errs: map[string]error{},
	}

	lister := app.NewLocalLister(nil, nil, scanner)
	repos := lister.CollectLocalRepos([]domain.Profile{work, personal})

	if len(repos) != 3 {
		t.Fatalf("len = %d, want 3", len(repos))
	}

	// プロファイル定義順を維持
	if repos[0].Profile != "work" || repos[0].Owner != "octocat-work" || repos[0].Repo != "project-a" {
		t.Errorf("repos[0] = %+v", repos[0])
	}
	if repos[2].Profile != "personal" || repos[2].Owner != "octocat" || repos[2].Repo != "dotfiles" {
		t.Errorf("repos[2] = %+v", repos[2])
	}
}

func TestCollectLocalRepos_EmptyRoot_Skipped(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work", Root: ""}

	scanner := &mockScanner{repos: map[string][]string{}, errs: map[string]error{}}
	lister := app.NewLocalLister(nil, nil, scanner)
	repos := lister.CollectLocalRepos([]domain.Profile{work})

	if len(repos) != 0 {
		t.Errorf("repos = %v, want empty", repos)
	}
}

func TestCollectLocalRepos_ScannerError_Skipped(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work", Root: "/home/work"}
	personal := domain.Profile{Name: "personal", GHConfigDir: "/path/personal", Root: "/home/personal"}

	scanner := &mockScanner{
		repos: map[string][]string{
			"/home/personal": {"octocat/dotfiles"},
		},
		errs: map[string]error{
			"/home/work": errors.New("permission denied"),
		},
	}

	lister := app.NewLocalLister(nil, nil, scanner)
	repos := lister.CollectLocalRepos([]domain.Profile{work, personal})

	if len(repos) != 1 {
		t.Fatalf("len = %d, want 1", len(repos))
	}
	if repos[0].Profile != "personal" {
		t.Errorf("repos[0].Profile = %q, want %q", repos[0].Profile, "personal")
	}
}
