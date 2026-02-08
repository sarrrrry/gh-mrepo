# gh-mrepo

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

```bash
# Select a profile interactively and run a gh command
gh mrepo repo list

# Specify a profile directly
gh mrepo --user work repo list
```
