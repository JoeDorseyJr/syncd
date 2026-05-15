# syncd

Keep your system synced. Declarative state management for macOS.

## Install

syncd is not yet distributed via Homebrew (planned for v0.5). For now, build from source:

```bash
git clone https://github.com/joedorseyjr/syncd.git
cd syncd
make build
# binary at ./bin/syncd
```

## Usage (v0.1)

```bash
syncd plan      # Show what would change
syncd apply     # Apply desired state
```

## What It Manages (v0.1)

- **Homebrew taps, formulae, and casks**
- **System cleanup** (remove undeclared taps/formulae/casks, prune deps, clear caches)

## Config

Lives at `~/.config/syncd/config.yaml`:

```yaml
taps: []

brews:
  - git
  - go
  - neovim

casks:
  - iterm2
  - discord
  - visual-studio-code

cleanup:
  remove_unlisted: true
  clear_cache: true
  autoremove: true
```

## How It Works

1. Reads `config.yaml` as the desired state
2. Queries the system for actual state (what's installed)
3. Computes a diff
4. Applies changes (with confirmation unless `--yes`)

Cleanup mode removes undeclared taps, formulae, and casks (when `cleanup.remove_unlisted` is true), prunes unused dependencies, and clears the Homebrew cache.

## Built With

- Go (single binary; no runtime dependencies beyond Homebrew CLI)
- Homebrew (package management backend)

## Project Structure

```
syncd/
├── cmd/syncd/           # CLI entry point
├── internal/
│   ├── config/          # YAML config parser
│   ├── brew/            # Tap, formula, cask management
│   ├── plan/            # Diff calculator
│   └── cli/             # Command implementations
├── config.yaml          # User config
├── go.mod
├── Makefile
└── README.md
```

## Roadmap

See [MILESTONES.md](MILESTONES.md) for planned features: macOS defaults, dotfiles, shell setup, Mac App Store, auto-updates, and more.

## Status

🚧 In development — v0.1 MVP
