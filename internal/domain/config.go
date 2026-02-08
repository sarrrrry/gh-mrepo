package domain

import "errors"

var (
	ErrNoProfiles      = errors.New("no profiles found in config")
	ErrEmptyGHConfigDir = errors.New("gh_config_dir must not be empty")
	ErrEmptyName       = errors.New("profile name must not be empty")
)
