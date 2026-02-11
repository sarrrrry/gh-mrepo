package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/sarrrrry/gh-mrepo/internal/app"
	"github.com/sarrrrry/gh-mrepo/internal/config"
	"github.com/sarrrrry/gh-mrepo/internal/domain"
	"github.com/sarrrrry/gh-mrepo/internal/executor"
	"github.com/sarrrrry/gh-mrepo/internal/selector"
)

func main() {
	user, args := extractUserFlag(os.Args[1:])
	if user == "" {
		user = os.Getenv("GH_MREPO_PROFILE")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	configPath := filepath.Join(home, ".config", "gh-mrepo", "config.toml")

	if len(args) > 0 && args[0] == "init" {
		if err := config.NewInitializer().Init(configPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("config.toml created: %s\n", configPath)
		return
	}

	if len(args) > 0 && args[0] == "ls" {
		loader := config.NewLoader(configPath)
		exec := executor.New()
		resolver := config.NewHostResolver()
		lister := app.NewLister(loader, exec, resolver)
		if err := lister.List(args[1:], os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if len(args) > 0 && args[0] == "switch" {
		loader := config.NewLoader(configPath)
		profiles, err := loader.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		wd, _ := os.Getwd()
		p, err := domain.FindByDirectory(profiles, wd)
		if err != nil {
			activeUser := resolveActiveUser()
			activeIdx := -1
			for i, prof := range profiles {
				u, e := config.ResolveGitHubUser(prof.GHConfigDir)
				if e == nil && u == activeUser {
					activeIdx = i
					break
				}
			}
			sel := selector.New()
			p, err = sel.SelectForSwitch(profiles, activeIdx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
		username, err := config.ResolveGitHubUser(p.GHConfigDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		cmd := exec.Command("gh", "auth", "switch", "--user", username)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	loader := config.NewLoader(configPath)
	sel := selector.New()
	exec := executor.New()

	a := app.New(loader, sel, exec)
	if err := a.Run(user, args); err != nil {
		if exitErr, ok := err.(*executor.ExitError); ok {
			os.Exit(exitErr.Code)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// extractUserFlag は引数から --user <value> を抽出し、残りの引数を返す。
func extractUserFlag(args []string) (string, []string) {
	var user string
	var rest []string

	for i := 0; i < len(args); i++ {
		if args[i] == "--user" && i+1 < len(args) {
			user = args[i+1]
			i++ // skip value
			continue
		}
		rest = append(rest, args[i])
	}
	return user, rest
}

var activeAccountRe = regexp.MustCompile(`account (\S+)`)

// resolveActiveUser は gh auth status --active の出力からアクティブユーザー名を返す。
func resolveActiveUser() string {
	out, err := exec.Command("gh", "auth", "status", "--active").CombinedOutput()
	if err != nil {
		return ""
	}
	m := activeAccountRe.FindSubmatch(out)
	if m == nil {
		return ""
	}
	return string(m[1])
}
