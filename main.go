package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sarrrrry/gh-mrepo/internal/app"
	"github.com/sarrrrry/gh-mrepo/internal/config"
	"github.com/sarrrrry/gh-mrepo/internal/executor"
	"github.com/sarrrrry/gh-mrepo/internal/selector"
)

func main() {
	user, args := extractUserFlag(os.Args[1:])

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	configPath := filepath.Join(home, ".config", "gh-mrepo", "config.toml")

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
