package app

import (
	"fmt"

	"github.com/sarrrrry/gh-mrepo/internal/domain"
)

// ProfileError はどのプロファイルでエラーが発生したかを示すエラー型。
type ProfileError struct {
	Profile domain.Profile
	Err     error
}

func (e *ProfileError) Error() string { return e.Err.Error() }
func (e *ProfileError) Unwrap() error { return e.Err }

type App struct {
	loader   ConfigLoader
	selector ProfileSelector
	executor GHExecutor
}

func New(loader ConfigLoader, selector ProfileSelector, executor GHExecutor) *App {
	return &App{
		loader:   loader,
		selector: selector,
		executor: executor,
	}
}

func (a *App) Run(user string, args []string) error {
	profiles, err := a.loader.Load()
	if err != nil {
		return err
	}

	var selected domain.Profile

	switch {
	case user != "":
		p, err := findProfile(profiles, user)
		if err != nil {
			return err
		}
		selected = p
	case len(profiles) == 1:
		selected = profiles[0]
	default:
		p, err := a.selector.Select(profiles)
		if err != nil {
			return err
		}
		selected = p
	}

	if err := a.executor.ExecRepo(selected, args); err != nil {
		return &ProfileError{Profile: selected, Err: err}
	}
	return nil
}

func findProfile(profiles []domain.Profile, name string) (domain.Profile, error) {
	for _, p := range profiles {
		if p.Name == name {
			return p, nil
		}
	}
	return domain.Profile{}, fmt.Errorf("profile %q not found", name)
}
