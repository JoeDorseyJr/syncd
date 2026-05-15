# syncd

Keep your system synced. Declarative state management for macOS.

## Install

syncd is not yet distributed via Homebrew (planned for v0.5). For now, build from source:

```bash
git clone https://github.com/JoeDorseyJr/syncd.git
cd syncd
make build
# binary at ./bin/syncd
```

## Usage

```bash
syncd plan                        # Show what would change
syncd plan --config ./my.yaml     # Use a custom config path
syncd apply                       # Apply desired state (with confirmation)
syncd apply --yes                 # Skip confirmation prompt
syncd --version                   # Print version
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success / no changes needed |
| 1 | Execution error (one or more operations failed) |
| 2 | Plan has pending package changes |

## What It Manages

- **Homebrew taps, formulae, and casks**
- **System cleanup** (remove undeclared taps/formulae/casks, prune deps, clear caches)

## Config

Default path: `~/.config/syncd/config.yaml` (override with `--config`)

```yaml
taps:
  - nikitabobko/tap

brews:
  - git
  - go
  - neovim

casks:
  - iterm2
  - visual-studio-code
  - firefox

cleanup:
  remove_unlisted: true   # Remove packages not listed above
  clear_cache: true       # Run brew cleanup
  autoremove: true        # Run brew autoremove
```

See [`config.example.yaml`](config.example.yaml) for a starter template.

## How It Works

1. Reads config as the desired state
2. Queries Homebrew for actual state (what's installed)
3. Computes a diff (adds and removals)
4. Applies changes (with confirmation unless `--yes`)

Cleanup mode removes undeclared taps, formulae, and casks (when `cleanup.remove_unlisted` is true), prunes unused dependencies, and clears the Homebrew cache. Cleanup actions are recurring maintenance — they don't affect the plan exit code.

## Built With

- Go (single binary, no runtime dependencies beyond Homebrew)
- [Cobra](https://github.com/spf13/cobra) (CLI framework)
- [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3) (config parsing)

## Project Structure

```
syncd/
├── cmd/syncd/           # CLI entry point
├── internal/
│   ├── config/          # YAML config parser (strict field validation)
│   ├── brew/            # State reader, executor, Homebrew check
│   ├── plan/            # Diff calculator
│   └── cli/             # plan and apply command implementations
├── test/                # Integration tests (fake brew)
├── config.example.yaml  # Starter config template
├── go.mod
├── Makefile
└── README.md
```

## Development

```bash
make build              # Build binary
make test               # Run unit tests
make integration-test   # Run integration tests (uses fake brew)
make lint               # Run go vet
make clean              # Remove build artifacts
```

## Roadmap

See [MILESTONES.md](MILESTONES.md) for planned features: macOS defaults, dotfiles, shell setup, Mac App Store, auto-updates, and more.

## License

MIT

## Status

🚧 In development — v0.1 MVP
