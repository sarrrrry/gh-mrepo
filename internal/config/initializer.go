package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sarrrrry/gh-mrepo/internal/domain"
)

const templateConfig = `# gh-mrepo configuration
# See: gh mrepo --help

[default]
gh_config_dir = "~/.config/gh"
# root = "~/repos"
`

type Initializer struct{}

func NewInitializer() *Initializer {
	return &Initializer{}
}

func (i *Initializer) Init(configPath string) error {
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("%w: %s", domain.ErrConfigAlreadyExists, configPath)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(configPath, []byte(templateConfig), 0o644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
