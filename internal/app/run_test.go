package app_test

import (
	"errors"
	"testing"

	"github.com/sarrrrry/gh-mrepo/internal/app"
	"github.com/sarrrrry/gh-mrepo/internal/domain"
)

// --- 手書きモック ---

type mockLoader struct {
	profiles []domain.Profile
	err      error
}

func (m *mockLoader) Load() ([]domain.Profile, error) {
	return m.profiles, m.err
}

type mockSelector struct {
	selected domain.Profile
	err      error
	called   bool
}

func (m *mockSelector) Select(profiles []domain.Profile) (domain.Profile, error) {
	m.called = true
	return m.selected, m.err
}

type mockExecutor struct {
	profile domain.Profile
	args    []string
	err     error
	called  bool
}

func (m *mockExecutor) ExecRepo(profile domain.Profile, args []string) error {
	m.called = true
	m.profile = profile
	m.args = args
	return m.err
}

func (m *mockExecutor) ExecRepoCapture(_ domain.Profile, _ []string) (string, error) {
	return "", nil
}

// --- テストケース ---

func TestRun_MultipleProfiles_SelectorCalled(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work"}
	personal := domain.Profile{Name: "personal", GHConfigDir: "/path/personal"}

	loader := &mockLoader{profiles: []domain.Profile{work, personal}}
	selector := &mockSelector{selected: work}
	executor := &mockExecutor{}

	a := app.New(loader, selector, executor)
	err := a.Run("", []string{"clone", "owner/repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !selector.called {
		t.Error("selector.Select was not called")
	}
	if !executor.called {
		t.Error("executor.ExecRepo was not called")
	}
	if executor.profile.Name != "work" {
		t.Errorf("executor.profile.Name = %q, want %q", executor.profile.Name, "work")
	}
}

func TestRun_SingleProfile_SelectorSkipped(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work"}

	loader := &mockLoader{profiles: []domain.Profile{work}}
	selector := &mockSelector{}
	executor := &mockExecutor{}

	a := app.New(loader, selector, executor)
	err := a.Run("", []string{"clone", "owner/repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selector.called {
		t.Error("selector.Select should not be called for single profile")
	}
	if executor.profile.Name != "work" {
		t.Errorf("executor.profile.Name = %q, want %q", executor.profile.Name, "work")
	}
}

func TestRun_UserFlag_SpecifiedProfile(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work"}
	personal := domain.Profile{Name: "personal", GHConfigDir: "/path/personal"}

	loader := &mockLoader{profiles: []domain.Profile{work, personal}}
	selector := &mockSelector{}
	executor := &mockExecutor{}

	a := app.New(loader, selector, executor)
	err := a.Run("work", []string{"clone", "owner/repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selector.called {
		t.Error("selector should not be called when --user is specified")
	}
	if executor.profile.Name != "work" {
		t.Errorf("executor.profile.Name = %q, want %q", executor.profile.Name, "work")
	}
}

func TestRun_UserFlag_UnknownProfile(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work"}

	loader := &mockLoader{profiles: []domain.Profile{work}}
	selector := &mockSelector{}
	executor := &mockExecutor{}

	a := app.New(loader, selector, executor)
	err := a.Run("unknown", []string{"clone", "owner/repo"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRun_LoaderError(t *testing.T) {
	loaderErr := errors.New("load failed")
	loader := &mockLoader{err: loaderErr}
	selector := &mockSelector{}
	executor := &mockExecutor{}

	a := app.New(loader, selector, executor)
	err := a.Run("", nil)
	if !errors.Is(err, loaderErr) {
		t.Errorf("err = %v, want %v", err, loaderErr)
	}
}

func TestRun_SelectorError(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work"}
	personal := domain.Profile{Name: "personal", GHConfigDir: "/path/personal"}

	selectorErr := errors.New("select failed")
	loader := &mockLoader{profiles: []domain.Profile{work, personal}}
	selector := &mockSelector{err: selectorErr}
	executor := &mockExecutor{}

	a := app.New(loader, selector, executor)
	err := a.Run("", nil)
	if !errors.Is(err, selectorErr) {
		t.Errorf("err = %v, want %v", err, selectorErr)
	}
}

func TestRun_ExecutorError(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work"}

	executorErr := errors.New("exec failed")
	loader := &mockLoader{profiles: []domain.Profile{work}}
	selector := &mockSelector{}
	executor := &mockExecutor{err: executorErr}

	a := app.New(loader, selector, executor)
	err := a.Run("", []string{"clone", "owner/repo"})
	if !errors.Is(err, executorErr) {
		t.Errorf("err = %v, want %v", err, executorErr)
	}
}

func TestRun_ReturnsProfileError(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work"}

	executorErr := errors.New("exec failed")
	loader := &mockLoader{profiles: []domain.Profile{work}}
	selector := &mockSelector{}
	executor := &mockExecutor{err: executorErr}

	a := app.New(loader, selector, executor)
	err := a.Run("", []string{"clone", "owner/repo"})

	var profileErr *app.ProfileError
	if !errors.As(err, &profileErr) {
		t.Fatalf("err should be *ProfileError, got %T", err)
	}
	if profileErr.Profile.Name != "work" {
		t.Errorf("ProfileError.Profile.Name = %q, want %q", profileErr.Profile.Name, "work")
	}
	if !errors.Is(profileErr.Err, executorErr) {
		t.Errorf("ProfileError.Err = %v, want %v", profileErr.Err, executorErr)
	}
}

func TestRun_ArgsPassedToExecutor(t *testing.T) {
	work := domain.Profile{Name: "work", GHConfigDir: "/path/work"}

	loader := &mockLoader{profiles: []domain.Profile{work}}
	selector := &mockSelector{}
	executor := &mockExecutor{}

	a := app.New(loader, selector, executor)
	args := []string{"clone", "owner/repo", "--depth", "1"}
	err := a.Run("", args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(executor.args) != len(args) {
		t.Fatalf("executor.args len = %d, want %d", len(executor.args), len(args))
	}
	for i, a := range args {
		if executor.args[i] != a {
			t.Errorf("executor.args[%d] = %q, want %q", i, executor.args[i], a)
		}
	}
}
