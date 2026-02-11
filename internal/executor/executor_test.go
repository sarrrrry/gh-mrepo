package executor

import (
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
