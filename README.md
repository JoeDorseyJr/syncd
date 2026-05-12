# syncd

Keep your system synced. Declarative state management for macOS.

## Install

```bash
brew tap joedorseyjr/syncd
brew install syncd
```

## Usage

```bash
syncd init      # Snapshot current system into config.yaml
syncd plan      # Show what would change
syncd apply     # Apply desired state
syncd update    # Upgrade all packages to latest
syncd status    # Show drift from declared state
syncd schedule  # Install/remove auto-update launchd job
```

## What It Manages

- **Homebrew taps, formulae, and casks**
- **Mac App Store apps** (via mas)
- **macOS system defaults** (dock, finder, trackpad, etc.)
- **Dotfiles** (symlinked from a central location)
- **Shell setup** (oh-my-zsh, plugins, theme)
- **Fonts**
- **System cleanup** (remove undeclared packages, prune deps, clear caches)
- **Auto-updates** (scheduled via launchd)

## Config

Lives at `~/.config/syncd/config.yaml`:

```yaml
taps:
  - buo/cask-upgrade

brews:
  - git
  - go
  - neovim

casks:
  - iterm2
  - discord
  - visual-studio-code

mas:
  - { name: "Xcode", id: 497799835 }

dotfiles:
  .gitconfig: dotfiles/.gitconfig
  .zshrc: dotfiles/.zshrc
  .ssh/config: dotfiles/ssh/config

defaults:
  com.apple.dock:
    autohide: true
    tilesize: 38
  com.apple.finder:
    ShowPathbar: true

cleanup:
  remove_unlisted: true
  clear_cache: true
  autoremove: true

auto_update:
  enabled: true
  interval: daily
  time: "03:00"
```

## How It Works

1. Reads `config.yaml` as the desired state
2. Queries the system for actual state (what's installed, current defaults, etc.)
3. Computes a diff
4. Applies changes (with confirmation unless `--yes`)

Cleanup mode removes anything installed that isn't declared in the config — keeping your system exactly as specified.

## Built With

- Go (single binary, no runtime dependencies)
- Homebrew (package management backend)
- launchd (scheduling)

## Project Structure

```
syncd/
├── cmd/syncd/           # CLI entry point
├── internal/
│   ├── config/          # YAML config parser
│   ├── brew/            # Tap, formula, cask management
│   ├── mas/             # Mac App Store
│   ├── defaults/        # macOS defaults read/write
│   ├── dotfiles/        # Symlink management
│   ├── shell/           # Oh-my-zsh, plugins, theme
│   ├── cleanup/         # Remove unlisted, clear cache
│   └── scheduler/       # Launchd plist generation
├── config.yaml          # Example config
├── dotfiles/            # User's dotfiles
├── go.mod
├── Makefile
└── README.md
```

## Status

🚧 In development
