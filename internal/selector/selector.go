package selector

import (
	"github.com/charmbracelet/huh"

	"github.com/sarrrrry/gh-mrepo/internal/domain"
)

type Selector struct{}

func New() *Selector {
	return &Selector{}
}

func (s *Selector) Select(profiles []domain.Profile) (domain.Profile, error) {
	options := make([]huh.Option[int], len(profiles))
	for i, p := range profiles {
		label := p.Name
		if p.Root != "" {
			label += " (" + p.Root + ")"
		}
		options[i] = huh.NewOption(label, i)
	}

	var selected int
	err := huh.NewSelect[int]().
		Title("Select profile").
		Options(options...).
		Filtering(true).
		Value(&selected).
		Run()
	if err != nil {
		return domain.Profile{}, err
	}

	return profiles[selected], nil
}
