package executor

import (
	"errors"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestExtractOwnerRepo(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{
			name: "owner/repo形式",
			arg:  "sarrrrry/init-setup",
			want: "sarrrrry/init-setup",
		},
		{
			name: "owner/repo.git形式",
			arg:  "sarrrrry/init-setup.git",
			want: "sarrrrry/init-setup",
		},
		{
			name: "HTTPS URL (.gitあり)",
			arg:  "https://github.com/sarrrrry/init-setup.git",
			want: "sarrrrry/init-setup",
		},
		{
			name: "HTTPS URL (.gitなし)",
			arg:  "https://github.com/sarrrrry/init-setup",
			want: "sarrrrry/init-setup",
		},
		{
			name: "SSH URL (.gitあり)",
			arg:  "git@github.com:sarrrrry/init-setup.git",
			want: "sarrrrry/init-setup",
		},
		{
			name: "SSH URL (.gitなし)",
			arg:  "git@github.com:sarrrrry/init-setup",
			want: "sarrrrry/init-setup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractOwnerRepo(tt.arg)
			if got != tt.want {
				t.Errorf("extractOwnerRepo(%q) = %q, want %q", tt.arg, got, tt.want)
			}
		})
	}
}

func TestExitError_Error(t *testing.T) {
	t.Run("Stderrがある場合はStderrを返す", func(t *testing.T) {
		e := &ExitError{Code: 1, Stderr: "HTTP 401: Bad credentials\n"}
		got := e.Error()
		want := "HTTP 401: Bad credentials"
		if got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Stderrが空の場合はexit statusを返す", func(t *testing.T) {
		e := &ExitError{Code: 42}
		got := e.Error()
		want := "exit status 42"
		if got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})
}

func TestExitError_IsAuthError(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		want   bool
	}{
		{"HTTP 401を含む", "HTTP 401: Bad credentials", true},
		{"authenticationを含む", "authentication required", true},
		{"無関係なエラー", "repository not found", false},
		{"空文字列", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ExitError{Code: 1, Stderr: tt.stderr}
			if got := e.IsAuthError(); got != tt.want {
				t.Errorf("IsAuthError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrapExitErrorWithStderr(t *testing.T) {
	t.Run("exec.ExitErrorをstderr付きExitErrorに変換", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", "exit 1")
		origErr := cmd.Run()

		got := wrapExitErrorWithStderr(origErr, "HTTP 401: Bad credentials")

		var exitErr *ExitError
		if !errors.As(got, &exitErr) {
			t.Fatalf("should return *ExitError, got %T", got)
		}
		if exitErr.Code != 1 {
			t.Errorf("Code = %d, want 1", exitErr.Code)
		}
		if exitErr.Stderr != "HTTP 401: Bad credentials" {
			t.Errorf("Stderr = %q, want %q", exitErr.Stderr, "HTTP 401: Bad credentials")
		}
	})

	t.Run("exec.ExitError以外はそのまま返す", func(t *testing.T) {
		origErr := errors.New("some error")
		got := wrapExitErrorWithStderr(origErr, "stderr content")
		if got != origErr {
			t.Errorf("should return original error, got %v", got)
		}
	})
}

func TestWrapExitError(t *testing.T) {
	t.Run("exec.ExitErrorをexecutor.ExitErrorに変換", func(t *testing.T) {
		// 存在しないコマンドを実行してExitErrorを生成
		cmd := exec.Command("sh", "-c", "exit 42")
		origErr := cmd.Run()

		got := wrapExitError(origErr)

		var exitErr *ExitError
		if !errors.As(got, &exitErr) {
			t.Fatalf("wrapExitError() should return *ExitError, got %T", got)
		}
		if exitErr.Code != 42 {
			t.Errorf("ExitError.Code = %d, want 42", exitErr.Code)
		}
	})

	t.Run("exec.ExitError以外はそのまま返す", func(t *testing.T) {
		origErr := errors.New("some error")
		got := wrapExitError(origErr)
		if got != origErr {
			t.Errorf("wrapExitError() should return original error, got %v", got)
		}
	})
}

func TestResolveCloneDir(t *testing.T) {
	root := "/repos"
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "owner/repo形式",
			args: []string{"clone", "sarrrrry/init-setup"},
			want: filepath.Join(root, "sarrrrry/init-setup"),
		},
		{
			name: "HTTPS URL (.gitあり)",
			args: []string{"clone", "https://github.com/sarrrrry/init-setup.git"},
			want: filepath.Join(root, "sarrrrry/init-setup"),
		},
		{
			name: "SSH URL",
			args: []string{"clone", "git@github.com:sarrrrry/init-setup.git"},
			want: filepath.Join(root, "sarrrrry/init-setup"),
		},
		{
			name: "フラグ値がrepo specとして誤認識されない",
			args: []string{"clone", "-u", "upstream", "sarrrrry/init-setup"},
			want: filepath.Join(root, "sarrrrry/init-setup"),
		},
		{
			name: "-- 以降のgitフラグは無視される",
			args: []string{"clone", "sarrrrry/init-setup", "--", "--depth", "1"},
			want: filepath.Join(root, "sarrrrry/init-setup"),
		},
		{
			name: "引数なし",
			args: []string{"clone"},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveCloneDir(root, tt.args)
			if got != tt.want {
				t.Errorf("resolveCloneDir(%q, %v) = %q, want %q", root, tt.args, got, tt.want)
			}
		})
	}
}
