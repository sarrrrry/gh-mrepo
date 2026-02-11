package app

import (
	"errors"
	"io"
	"strings"
	"sync"

	"github.com/sarrrrry/gh-mrepo/internal/domain"
)

type LocalRepo struct {
	Profile string `json:"profile"`
	Owner   string `json:"owner"`
	Repo    string `json:"repo"`
}

type LocalLister struct {
	loader   ConfigLoader
	resolver UserResolver
	scanner  DirScanner
}

func NewLocalLister(loader ConfigLoader, resolver UserResolver, scanner DirScanner) *LocalLister {
	return &LocalLister{
		loader:   loader,
		resolver: resolver,
		scanner:  scanner,
	}
}

func (l *LocalLister) ListLocal(w io.Writer) error {
	profiles, err := l.loader.Load()
	if err != nil {
		return err
	}

	results := make([]ProfileResult, len(profiles))

	var wg sync.WaitGroup
	for i, p := range profiles {
		wg.Add(1)
		go func(idx int, prof domain.Profile) {
			defer wg.Done()
			results[idx] = l.scanProfile(prof)
		}(i, p)
	}
	wg.Wait()

	FormatResults(results, w)
	return nil
}

func (l *LocalLister) ListLocalProfile(prof domain.Profile, w io.Writer) error {
	r := l.scanProfile(prof)
	FormatResults([]ProfileResult{r}, w)
	return nil
}

func (l *LocalLister) CollectLocalRepos(profiles []domain.Profile) []LocalRepo {
	reposByIdx := make([][]LocalRepo, len(profiles))

	var wg sync.WaitGroup
	for i, p := range profiles {
		wg.Add(1)
		go func(idx int, prof domain.Profile) {
			defer wg.Done()
			if prof.Root == "" {
				return
			}
			repos, err := l.scanner.ScanLocalRepos(prof.Root)
			if err != nil {
				return
			}
			for _, r := range repos {
				parts := strings.SplitN(r, "/", 2)
				if len(parts) == 2 {
					reposByIdx[idx] = append(reposByIdx[idx], LocalRepo{
						Profile: prof.Name,
						Owner:   parts[0],
						Repo:    parts[1],
					})
				}
			}
		}(i, p)
	}
	wg.Wait()

	var all []LocalRepo
	for _, repos := range reposByIdx {
		all = append(all, repos...)
	}
	return all
}

func (l *LocalLister) scanProfile(prof domain.Profile) ProfileResult {
	r := ProfileResult{Profile: prof}

	username, err := l.resolver.ResolveGitHubUser(prof.GHConfigDir)
	if err == nil {
		r.Username = username
	}

	if prof.Root == "" {
		r.Err = errors.New("root not configured")
		return r
	}

	repos, err := l.scanner.ScanLocalRepos(prof.Root)
	if err != nil {
		r.Err = err
		return r
	}

	r.Output = strings.Join(repos, "\n")
	return r
}
