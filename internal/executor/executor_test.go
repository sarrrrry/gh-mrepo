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
