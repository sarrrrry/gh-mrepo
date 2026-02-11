package app

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/sarrrrry/gh-mrepo/internal/domain"
)

type Lister struct {
	loader   ConfigLoader
	executor GHExecutor
	resolver UserResolver
}

func NewLister(loader ConfigLoader, executor GHExecutor, resolver UserResolver) *Lister {
	return &Lister{
		loader:   loader,
		executor: executor,
		resolver: resolver,
	}
}

type ProfileResult struct {
	Profile  domain.Profile
	Username string
	Output   string
	Err      error
}

func (l *Lister) List(args []string, w io.Writer) error {
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
			r := ProfileResult{Profile: prof}

			username, err := l.resolver.ResolveGitHubUser(prof.GHConfigDir)
			if err != nil {
				r.Err = err
				results[idx] = r
				return
			}
			r.Username = username

			repoArgs := append([]string{"list"}, args...)
			output, err := l.executor.ExecRepoCapture(prof, repoArgs)
			if err != nil {
				r.Err = err
				results[idx] = r
				return
			}
			r.Output = output
			results[idx] = r
		}(i, p)
	}
	wg.Wait()

	FormatResults(results, w)
	return nil
}

func FormatResults(results []ProfileResult, w io.Writer) {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

	separator := separatorStyle.Render(strings.Repeat("\u2500", 40))

	for i, r := range results {
		if i > 0 {
			fmt.Fprintln(w)
		}

		headerText := r.Profile.Name
		if r.Username != "" {
			headerText += fmt.Sprintf(" (@%s)", r.Username)
		}
		fmt.Fprintln(w, headerStyle.Render(headerText))
		fmt.Fprintln(w, separator)

		if r.Err != nil {
			fmt.Fprintln(w, errorStyle.Render(r.Err.Error()))
			continue
		}

		output := strings.TrimRight(r.Output, "\n")
		if output == "" {
			fmt.Fprintln(w, "No repositories")
			continue
		}

		fmt.Fprintln(w, output)
	}
}
