# leet-tui

A terminal-based LeetCode practice tool with spaced repetition and study plans.

## Features

- **Problem Browser** — Browse, search, and filter LeetCode problems by topic and difficulty
- **Study Plans** — Built-in plans: Blind 75, NeetCode 150, and 21-Day Challenge
- **Spaced Repetition (FSRS-5)** — Intelligent review scheduling based on your recall performance
- **Code Editor Integration** — Open problems in your preferred editor directly from the TUI
- **Progress Tracking** — Dashboard with completion stats and review history

## Installation

```bash
go install github.com/tzy0608/leet-tui/cmd/leet-tui@latest
```

## Quick Start

```bash
# 1. Login — reads session from your browser cookies
leet-tui login

# 2. Sync problems from LeetCode
leet-tui --sync

# 3. Launch the TUI
leet-tui
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `tab` | Switch between pages |
| `/` | Search problems |
| `enter` | Open problem detail |
| `e` | Open in editor |
| `r` | Rate recall (review mode) |
| `?` | Toggle help |
| `q` | Quit |

## Build

```bash
make build    # Build binary to bin/leet-tui
make run      # Run directly
make test     # Run tests
make lint     # Run linter
```

## Configuration

Config file is stored at `~/.config/leet-tui/config.toml`.

## License

MIT
