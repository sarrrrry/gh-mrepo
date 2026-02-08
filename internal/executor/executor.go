package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sarrrrry/gh-mrepo/internal/domain"
)

// ExitError はgh repoの終了コードを伝播するためのエラー型。
type ExitError struct {
	Code int
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.Code)
}

type Executor struct{}

func New() *Executor {
	return &Executor{}
}

func (e *Executor) ExecRepo(profile domain.Profile, args []string) error {
	ghArgs := append([]string{"repo"}, args...)

	// clone + root設定時: clone先パスを追加
	if profile.Root != "" && len(args) > 0 && args[0] == "clone" {
		cloneDir := resolveCloneDir(profile.Root, args)
		if cloneDir != "" {
			ghArgs = append(ghArgs, cloneDir)
		}
	}

	ghPath, err := exec.LookPath("gh")
	if err != nil {
		return fmt.Errorf("gh command not found: %w", err)
	}

	cmd := exec.Command(ghPath, ghArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = appendEnv(os.Environ(), "GH_CONFIG_DIR", profile.GHConfigDir)

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &ExitError{Code: exitErr.ExitCode()}
		}
		return err
	}
	return nil
}

// resolveCloneDir はclone先ディレクトリを決定する。
// argsは"clone"の後の引数。"owner/repo"形式を探して{root}/owner/repoを返す。
func resolveCloneDir(root string, args []string) string {
	// args[0]は"clone"、args[1:]からrepo specを探す
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		// owner/repo形式 (URLでもなくフラグでもない最初の引数)
		return filepath.Join(root, arg)
	}
	return ""
}

func appendEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}
