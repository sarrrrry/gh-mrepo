package domain

import (
	"fmt"
	"strings"
)

// Profile はGitHubアカウントの設定プロファイルを表す値オブジェクト。
type Profile struct {
	Name        string // TOMLセクション名
	GHConfigDir string // 展開済み絶対パス
	Root        string // clone先ルート (空の場合はデフォルト動作)
}

func NewProfile(name, ghConfigDir, root string) (Profile, error) {
	if name == "" {
		return Profile{}, ErrEmptyName
	}
	if ghConfigDir == "" {
		return Profile{}, ErrEmptyGHConfigDir
	}
	return Profile{
		Name:        name,
		GHConfigDir: ghConfigDir,
		Root:        root,
	}, nil
}

// FindByDirectory は指定ディレクトリに一致するプロファイルを返す。
func FindByDirectory(profiles []Profile, dir string) (Profile, error) {
	for _, p := range profiles {
		if p.Root != "" && strings.HasPrefix(dir, p.Root) {
			return p, nil
		}
	}
	return Profile{}, fmt.Errorf("no profile found for directory %q", dir)
}
