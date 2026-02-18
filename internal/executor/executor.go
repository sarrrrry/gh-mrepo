package executor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sarrrrry/gh-mrepo/internal/domain"
)

// ExitError はgh repoの終了コードを伝播するためのエラー型。
type ExitError struct {
	Code   int
	Stderr string
}

func (e *ExitError) Error() string {
	if e.Stderr != "" {
		return strings.TrimRight(e.Stderr, "\n")
	}
	return fmt.Sprintf("exit status %d", e.Code)
}

func (e *ExitError) IsAuthError() bool {
	return strings.Contains(e.Stderr, "HTTP 401") ||
		strings.Contains(e.Stderr, "authentication")
}

type Executor struct{}

func New() *Executor {
	return &Executor{}
}

func (e *Executor) ExecRepo(profile domain.Profile, args []string) error {
	cmd, err := buildRepoCmd(profile, args)
	if err != nil {
		return err
	}

	// clone + root設定時: clone先パスを追加
	if profile.Root != "" && len(args) > 0 && args[0] == "clone" {
		cloneDir := resolveCloneDir(profile.Root, args)
		if cloneDir != "" {
			cmd.Args = append(cmd.Args, cloneDir)
		}
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	var stderrBuf bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	if err := cmd.Run(); err != nil {
		return wrapExitErrorWithStderr(err, stderrBuf.String())
	}
	return nil
}

// resolveCloneDir はclone先ディレクトリを決定する。
// argsは"clone"の後の引数。"owner/repo"形式を探して{root}/owner/repoを返す。
func resolveCloneDir(root string, args []string) string {
	for _, arg := range args[1:] {
		if arg == "--" {
			break
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
		// repo specは必ず"/"か":"を含む (owner/repo, URL形式)
		// フラグの値 (例: --depth "1") を誤認識しないようにする
		if !strings.Contains(arg, "/") && !strings.Contains(arg, ":") {
			continue
		}
		return filepath.Join(root, extractOwnerRepo(arg))
	}
	return ""
}

// extractOwnerRepo は引数からowner/repo部分を抽出する。
// HTTPS URL, SSH URL, owner/repo形式に対応し、.gitサフィックスを除去する。
func extractOwnerRepo(arg string) string {
	// HTTPS URL: https://github.com/owner/repo.git
	if u, err := url.Parse(arg); err == nil && u.Scheme != "" {
		parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
		if len(parts) >= 2 {
			repo := strings.TrimSuffix(parts[1], ".git")
			return parts[0] + "/" + repo
		}
	}

	// SSH URL: git@github.com:owner/repo.git
	if strings.Contains(arg, ":") && !strings.Contains(arg, "://") {
		colonIdx := strings.Index(arg, ":")
		path := arg[colonIdx+1:]
		path = strings.TrimSuffix(path, ".git")
		return path
	}

	// owner/repo or owner/repo.git
	return strings.TrimSuffix(arg, ".git")
}

func (e *Executor) ExecRepoCapture(profile domain.Profile, args []string) (string, error) {
	cmd, err := buildRepoCmd(profile, args)
	if err != nil {
		return "", err
	}

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	out, err := cmd.Output()
	if err != nil {
		return "", wrapExitErrorWithStderr(err, stderrBuf.String())
	}
	return string(out), nil
}

// buildRepoCmd は "gh repo ..." コマンドを構築する。
func buildRepoCmd(profile domain.Profile, args []string) (*exec.Cmd, error) {
	ghArgs := append([]string{"repo"}, args...)
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		return nil, fmt.Errorf("gh command not found: %w", err)
	}
	cmd := exec.Command(ghPath, ghArgs...)
	cmd.Env = appendEnv(os.Environ(), "GH_CONFIG_DIR", profile.GHConfigDir)
	return cmd, nil
}

// wrapExitError は exec.ExitError を executor.ExitError に変換する。
func wrapExitError(err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return &ExitError{Code: exitErr.ExitCode()}
	}
	return err
}

// wrapExitErrorWithStderr は exec.ExitError を stderr 付き executor.ExitError に変換する。
func wrapExitErrorWithStderr(err error, stderr string) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return &ExitError{Code: exitErr.ExitCode(), Stderr: stderr}
	}
	return err
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
