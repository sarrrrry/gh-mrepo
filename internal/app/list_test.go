package app_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/sarrrrry/gh-mrepo/internal/app"
	"github.com/sarrrrry/gh-mrepo/internal/domain"
)

// --- mock for UserResolver ---

type mockResolver struct {
	users map[string]string // ghConfigDir -> username
	err   map[string]error
}

func (m *mockResolver) ResolveGitHubUser(ghConfigDir string) (string, error) {
	if e, ok := m.err[ghConfigDir]; ok {
		return "", e
	}
	return m.users[ghConfigDir], nil
}

// --- mock for GHExecutor (capture) ---

type mockCaptureExecutor struct {
	outputs map[string]string // ghConfigDir -> output
	errs    map[string]error
}

func (m *mockCaptureExecutor) ExecRepo(_ domain.Profile, _ []string) error {
	return nil
}

func (m *mockCaptureExecutor) ExecRepoCapture(profile domain.Profile, _ []string) (string, error) {
	if e, ok := m.errs[profile.GHConfigDir]; ok {
		return "", e
	}
	return m.outputs[profile.GHConfigDir], nil
}

// --- テストケース ---

func TestList_MultipleProfiles(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work"}
	personal := domain.Profile{Name: "personal", GHConfigDir: "/path/personal"}

	loader := &mockLoader{profiles: []domain.Profile{work, personal}}
	executor := &mockCaptureExecutor{
		outputs: map[string]string{
			"/path/work":     "octocat-work/project-a\tpublic\t2026-02-10T12:00:00Z\noctocat-work/project-b\tBackend\tprivate\t2026-02-09T08:00:00Z\n",
			"/path/personal": "octocat/dotfiles\tpublic\t2026-01-30T07:52:58Z\n",
		},
		errs: map[string]error{},
	}
	resolver := &mockResolver{
		users: map[string]string{
			"/path/work":     "octocat-work",
			"/path/personal": "octocat",
		},
		err: map[string]error{},
	}

	lister := app.NewLister(loader, executor, resolver)
	var buf bytes.Buffer
	err := lister.List(nil, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	// プロファイル名とユーザー名がヘッダーに含まれること
	if !strings.Contains(out, "work") || !strings.Contains(out, "@octocat-work") {
		t.Errorf("output missing work header, got:\n%s", out)
	}
	if !strings.Contains(out, "personal") || !strings.Contains(out, "@octocat") {
		t.Errorf("output missing personal header, got:\n%s", out)
	}

	// リポジトリ出力が含まれること
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

func TestList_ProfileError_ContinuesOthers(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work"}
	personal := domain.Profile{Name: "personal", GHConfigDir: "/path/personal"}

	loader := &mockLoader{profiles: []domain.Profile{work, personal}}
	executor := &mockCaptureExecutor{
		outputs: map[string]string{
			"/path/personal": "octocat/dotfiles\tpublic\t2026-01-30T07:52:58Z\n",
		},
		errs: map[string]error{
			"/path/work": errors.New("gh command failed"),
		},
	}
	resolver := &mockResolver{
		users: map[string]string{
			"/path/work":     "octocat-work",
			"/path/personal": "octocat",
		},
		err: map[string]error{},
	}

	lister := app.NewLister(loader, executor, resolver)
	var buf bytes.Buffer
	err := lister.List(nil, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	// エラープロファイルのヘッダーは出力される
	if !strings.Contains(out, "work") {
		t.Errorf("output should contain work header, got:\n%s", out)
	}
	// エラーメッセージが出力される
	if !strings.Contains(out, "gh command failed") {
		t.Errorf("output should contain error message, got:\n%s", out)
	}
	// 他のプロファイルは正常に出力される
	if !strings.Contains(out, "octocat/dotfiles") {
		t.Errorf("output should contain personal repos, got:\n%s", out)
	}
}

func TestList_ResolverError_ContinuesOthers(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work"}
	personal := domain.Profile{Name: "personal", GHConfigDir: "/path/personal"}

	loader := &mockLoader{profiles: []domain.Profile{work, personal}}
	executor := &mockCaptureExecutor{
		outputs: map[string]string{
			"/path/personal": "octocat/dotfiles\tpublic\t2026-01-30T07:52:58Z\n",
		},
		errs: map[string]error{},
	}
	resolver := &mockResolver{
		users: map[string]string{
			"/path/personal": "octocat",
		},
		err: map[string]error{
			"/path/work": errors.New("hosts.yml not found"),
		},
	}

	lister := app.NewLister(loader, executor, resolver)
	var buf bytes.Buffer
	err := lister.List(nil, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	// エラーメッセージが出力される
	if !strings.Contains(out, "hosts.yml not found") {
		t.Errorf("output should contain resolver error, got:\n%s", out)
	}
	// 他のプロファイルは正常に出力される
	if !strings.Contains(out, "octocat/dotfiles") {
		t.Errorf("output should contain personal repos, got:\n%s", out)
	}
}

func TestList_EmptyOutput_ShowsNoRepositories(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work"}

	loader := &mockLoader{profiles: []domain.Profile{work}}
	executor := &mockCaptureExecutor{
		outputs: map[string]string{"/path/work": ""},
		errs:    map[string]error{},
	}
	resolver := &mockResolver{
		users: map[string]string{"/path/work": "octocat-work"},
		err:   map[string]error{},
	}

	lister := app.NewLister(loader, executor, resolver)
	var buf bytes.Buffer
	err := lister.List(nil, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No repositories") {
		t.Errorf("output should contain 'No repositories', got:\n%s", out)
	}
}

func TestList_LoaderError(t *testing.T) {
	loaderErr := errors.New("load failed")
	loader := &mockLoader{err: loaderErr}
	executor := &mockCaptureExecutor{}
	resolver := &mockResolver{}

	lister := app.NewLister(loader, executor, resolver)
	var buf bytes.Buffer
	err := lister.List(nil, &buf)
	if !errors.Is(err, loaderErr) {
		t.Errorf("err = %v, want %v", err, loaderErr)
	}
}

func TestList_ArgsPassedToExecutor(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work"}

	loader := &mockLoader{profiles: []domain.Profile{work}}

	var capturedArgs []string
	executor := &argsCapturingExecutor{
		output:       "octocat-work/repo\tpublic\t2026-01-01T00:00:00Z\n",
		capturedArgs: &capturedArgs,
	}
	resolver := &mockResolver{
		users: map[string]string{"/path/work": "octocat-work"},
		err:   map[string]error{},
	}

	lister := app.NewLister(loader, executor, resolver)
	var buf bytes.Buffer
	args := []string{"--limit", "5"}
	err := lister.List(args, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// argsが "list" + 渡した引数として転送されること
	if len(capturedArgs) < 1 || capturedArgs[0] != "list" {
		t.Errorf("first arg should be 'list', got: %v", capturedArgs)
	}
	if len(capturedArgs) < 3 || capturedArgs[1] != "--limit" || capturedArgs[2] != "5" {
		t.Errorf("args should contain --limit 5, got: %v", capturedArgs)
	}
}

// --- 認証エラーを模倣するモック ---

type mockAuthError struct {
	msg string
}

func (e *mockAuthError) Error() string    { return e.msg }
func (e *mockAuthError) IsAuthError() bool { return true }

func TestFormatResults_AuthErrorShowsHint(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/home/user/.config/gh-work"}
	results := []app.ProfileResult{
		{
			Profile:  work,
			Username: "octocat-work",
			Err:      &mockAuthError{msg: "HTTP 401: Bad credentials"},
		},
	}

	var buf bytes.Buffer
	app.FormatResults(results, &buf)
	out := buf.String()

	if !strings.Contains(out, "HTTP 401: Bad credentials") {
		t.Errorf("output should contain error message, got:\n%s", out)
	}
	if !strings.Contains(out, "hint: GH_CONFIG_DIR=/home/user/.config/gh-work gh auth login") {
		t.Errorf("output should contain auth hint, got:\n%s", out)
	}
}

func TestFormatResults_NonAuthErrorNoHint(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/home/user/.config/gh-work"}
	results := []app.ProfileResult{
		{
			Profile:  work,
			Username: "octocat-work",
			Err:      errors.New("repository not found"),
		},
	}

	var buf bytes.Buffer
	app.FormatResults(results, &buf)
	out := buf.String()

	if strings.Contains(out, "hint:") {
		t.Errorf("output should not contain hint for non-auth error, got:\n%s", out)
	}
}

type argsCapturingExecutor struct {
	output       string
	capturedArgs *[]string
}

func (m *argsCapturingExecutor) ExecRepo(_ domain.Profile, _ []string) error {
	return nil
}

func (m *argsCapturingExecutor) ExecRepoCapture(_ domain.Profile, args []string) (string, error) {
	*m.capturedArgs = args
	return m.output, nil
}
