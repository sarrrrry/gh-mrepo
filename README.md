![header](https://capsule-render.vercel.app/api?type=waving&height=300&color=gradient&text=gh-mrepo)

[![Go](https://img.shields.io/github/go-mod/go-version/sarrrrry/gh-mrepo)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/sarrrrry/gh-mrepo)](https://github.com/sarrrrry/gh-mrepo/releases/latest)
[![CI](https://img.shields.io/github/actions/workflow/status/sarrrrry/gh-mrepo/release.yml?label=release)](https://github.com/sarrrrry/gh-mrepo/actions/workflows/release.yml)

A [gh](https://cli.github.com/) extension to run `gh` commands with multiple GitHub accounts.

## Installation

```bash
gh extension install sarrrrry/gh-mrepo
```

To build from source:

```bash
git clone https://github.com/sarrrrry/gh-mrepo.git
cd gh-mrepo
go build -o gh-mrepo .
```

## Setup

### Initialize config

Run `init` to generate a config file on first use.

```bash
gh mrepo init
```

A template will be created at `~/.config/gh-mrepo/config.toml`.
If the file already exists, the command will exit with an error.

### config.toml

`config.toml` defines profiles (GitHub accounts).

```toml
# gh-mrepo configuration
# See: gh mrepo --help

[default]
gh_config_dir = "~/.config/gh"
# root = "~/repos"
```

| Key | Required | Description |
|-----|----------|-------------|
| `gh_config_dir` | Yes | Path to the `gh` config directory. Specify a separate directory for each account. |
| `root` | No | Root directory for cloning repositories. |

The section name (`[default]`) becomes the profile name.
Add more sections to use multiple accounts.

```toml
[work]
gh_config_dir = "~/.config/gh-work"
root = "~/repos/work"

[personal]
gh_config_dir = "~/.config/gh-personal"
root = "~/repos/personal"
```

## Usage

`gh mrepo` wraps `gh repo` commands with profile-aware `GH_CONFIG_DIR`.

```bash
# Select a profile interactively and run a gh repo command
gh mrepo repo list

# Specify a profile directly
gh mrepo --user work repo list
```

You can also set the profile via the `GH_MREPO_PROFILE` environment variable:

```bash
export GH_MREPO_PROFILE=work
gh mrepo repo clone owner/repo
```

### Clone with auto-routing

When `root` is set, `gh mrepo repo clone` automatically routes the clone destination under the profile's root directory:

```bash
gh mrepo --user work repo clone owner/repo
# => cloned to ~/repos/work/owner/repo
```

### Switch account

`gh mrepo switch` switches the active `gh` account (`gh auth switch`) based on the profile configuration.

```bash
gh mrepo switch
```

- If the current directory is under a profile's `root`, the account is switched automatically.
- Otherwise, an interactive selector is displayed. The currently active account is highlighted with a green `✓ active` label.

```
Switch account
> personal (~/repos/personal/)
  work (~/repos/work/) ✓ active
```

The GitHub username is resolved from `hosts.yml` in each profile's `gh_config_dir`, so the profile name (TOML section name) does not need to match the GitHub username.
