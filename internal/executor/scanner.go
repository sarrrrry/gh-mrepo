package executor

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FsScanner struct{}

func NewFsScanner() *FsScanner {
	return &FsScanner{}
}

func (s *FsScanner) ScanLocalRepos(root string) ([]string, error) {
	owners, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var repos []string
	for _, owner := range owners {
		if !owner.IsDir() || strings.HasPrefix(owner.Name(), ".") {
			continue
		}

		repoEntries, err := os.ReadDir(filepath.Join(root, owner.Name()))
		if err != nil {
			continue
		}

		for _, repo := range repoEntries {
			if !repo.IsDir() || strings.HasPrefix(repo.Name(), ".") {
				continue
			}
			repos = append(repos, owner.Name()+"/"+repo.Name())
		}
	}

	sort.Strings(repos)
	return repos, nil
}
