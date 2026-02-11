package selector

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/sarrrrry/gh-mrepo/internal/domain"
)

var activeLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓ active")

type Selector struct{}

func New() *Selector {
	return &Selector{}
}

func (s *Selector) Select(profiles []domain.Profile) (domain.Profile, error) {
	return s.selectWithOptions(profiles, "Select profile", -1)
}

// SelectForSwitch はアクティブプロファイルを強調して表示する。
// activeIdx が 0 以上の場合、そのインデックスのプロファイルに (active) を付与する。
func (s *Selector) SelectForSwitch(profiles []domain.Profile, activeIdx int) (domain.Profile, error) {
	return s.selectWithOptions(profiles, "Switch account", activeIdx)
}

func (s *Selector) selectWithOptions(profiles []domain.Profile, title string, activeIdx int) (domain.Profile, error) {
	options := make([]huh.Option[int], len(profiles))
	for i, p := range profiles {
		label := p.Name
		if p.Root != "" {
			label += " (" + p.Root + ")"
		}
		if i == activeIdx {
			label += " " + activeLabel
		}
		options[i] = huh.NewOption(label, i)
	}

	var selected int
	err := huh.NewSelect[int]().
		Title(title).
		Options(options...).
		Filtering(true).
		Value(&selected).
		Run()
	if err != nil {
		return domain.Profile{}, err
	}

	return profiles[selected], nil
}
