package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

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

	if len(args) > 0 && args[0] == "lls" {
		allFlag, jsonFlag := extractLlsFlags(args[1:])
		loader := config.NewLoader(configPath)
		resolver := config.NewHostResolver()
		scanner := executor.NewFsScanner()
		localLister := app.NewLocalLister(loader, resolver, scanner)

		profiles, err := loader.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var selected []domain.Profile
		if allFlag {
			selected = profiles
		} else {
			sel := selector.New()
			p, err := sel.Select(profiles)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			selected = []domain.Profile{p}
		}

		if jsonFlag {
			repos := localLister.CollectLocalRepos(selected)
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(repos); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}

		var buf bytes.Buffer
		if allFlag {
			if err := localLister.ListLocal(&buf); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := localLister.ListLocalProfile(selected[0], &buf); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
		if err := viewInPager(buf.Bytes()); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if len(args) > 0 && args[0] == "ls" {
		lsArgs, allFlag := extractAllFlag(args[1:])
		if allFlag {
			loader := config.NewLoader(configPath)
			e := executor.New()
			resolver := config.NewHostResolver()
			lister := app.NewLister(loader, e, resolver)
			var buf bytes.Buffer
			if err := lister.List(lsArgs, &buf); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if err := viewInPager(buf.Bytes()); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
		args = append([]string{"list"}, lsArgs...)
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

// extractLlsFlags は引数から -a/--all と -j/--json を検出する。-aj等の結合フラグにも対応。
func extractLlsFlags(args []string) (allFlag, jsonFlag bool) {
	for _, a := range args {
		switch a {
		case "--all":
			allFlag = true
		case "--json":
			jsonFlag = true
		default:
			if strings.HasPrefix(a, "-") && !strings.HasPrefix(a, "--") {
				for _, c := range a[1:] {
					switch c {
					case 'a':
						allFlag = true
					case 'j':
						jsonFlag = true
					}
				}
			}
		}
	}
	return
}

// extractAllFlag は引数から --all/-a を検出・除去し、残りの引数とフラグの有無を返す。
func extractAllFlag(args []string) ([]string, bool) {
	var rest []string
	all := false
	for _, a := range args {
		if a == "--all" || a == "-a" {
			all = true
			continue
		}
		rest = append(rest, a)
	}
	return rest, all
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

// viewInPager は内容をページャ経由で表示する。
func viewInPager(content []byte) error {
	pager := os.Getenv("GH_PAGER")
	if pager == "" {
		pager = os.Getenv("PAGER")
	}
	if pager == "" {
		pager = "less -R"
	}
	cmd := exec.Command("sh", "-c", pager)
	cmd.Stdin = bytes.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

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
