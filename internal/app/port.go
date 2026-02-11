package app

import "github.com/sarrrrry/gh-mrepo/internal/domain"

type ConfigLoader interface {
	Load() ([]domain.Profile, error)
}

type ProfileSelector interface {
	Select(profiles []domain.Profile) (domain.Profile, error)
}

type GHExecutor interface {
	ExecRepo(profile domain.Profile, args []string) error
	ExecRepoCapture(profile domain.Profile, args []string) (string, error)
}

type UserResolver interface {
	ResolveGitHubUser(ghConfigDir string) (string, error)
}

type DirScanner interface {
	ScanLocalRepos(root string) ([]string, error)
}
