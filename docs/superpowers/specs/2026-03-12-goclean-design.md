# Goclean — Design Spec

## Overview

Goclean is a TUI-first CLI tool in Go that scans the filesystem for dependency folders (node_modules, vendor, .venv, etc.) in stale projects and lets the user bulk-delete them to reclaim disk space. It tracks lifetime statistics and persists user preferences in a config file.

## Architecture

```
goclean/
├── cmd/goclean/          # Entry point, flag parsing
├── internal/
│   ├── config/           # Load/save ~/.goclean.yaml + first-run wizard
│   ├── scanner/          # Recursive filesystem walk with goroutines + channels
│   ├── filter/           # Staleness detection via config file mod time
│   ├── cleaner/          # Directory deletion (respects dry-run)
│   ├── stats/            # Persistent statistics (~/.goclean_stats.json)
│   └── tui/              # Bubbletea model, views, key bindings
├── pkg/
│   └── targets/          # Target registry (name, folder, config file)
└── go.mod
```

## Target Registry

Each target is a struct: `{Name, DependencyDir, ConfigFile}`. Examples:

| Name        | DependencyDir  | ConfigFile       |
|-------------|---------------|------------------|
| Node.js     | node_modules  | package.json     |
| PHP         | vendor        | composer.json    |
| Python      | .venv / venv  | requirements.txt |
| Rust        | target        | Cargo.toml       |
| Java Maven  | target        | pom.xml          |
| Go          | vendor        | go.mod           |
| Dart/Flutter| .dart_tool    | pubspec.yaml     |
| CocoaPods   | Pods          | Podfile          |
| Gradle      | build         | build.gradle     |

Targets are opt-in: the user selects which to monitor during the first-run wizard. Stored in `~/.goclean.yaml`.

## Config — ~/.goclean.yaml

```yaml
days: 30
targets:
  - node_modules
  - vendor
  - .venv
excluded_paths:
  - /Users/me/important-project
```

Overridable via flags: `--days`, `--path`.

## Scanner

- Uses `filepath.WalkDir` for the recursive walk.
- Skips automatically: dot-directories (`.git`, `.config`, etc.), system paths (`/System`, `/Library`, `/usr`, `/bin`, `/proc`, `/sys`), and user-excluded paths from config.
- When a target dependency dir is found, sends it to a channel.
- A pool of goroutines calculates directory sizes in parallel.
- Once inside a dependency dir, does not recurse further (prunes the walk).

## Filter

A project is "stale" if the config file (e.g. `package.json` next to `node_modules`) has a modification time older than `--days` (default 30). If the config file is missing, the directory is still reported but flagged.

## TUI (Bubbletea + Lipgloss)

### First-run wizard
Multi-select list of available targets. Saved to config on confirm.

### Main view
- Scrollable list: path, target type, size (MB/GB), last modified date
- `Space` — toggle selection
- `a` — select all, `n` — deselect all
- `Enter` — confirm deletion (shows confirmation screen)
- `q` / `Esc` — quit
- Footer: real-time total of selected size + lifetime stats summary

### Confirmation screen
Recap of selected items and total size. `y` to proceed, `n` to go back.

### Deletion progress
Progress bar/spinner while deleting. Summary at the end.

## Statistics — ~/.goclean_stats.json

```json
{
  "total_cleaned_bytes": 15032385536,
  "total_deletions": 47,
  "first_run": "2026-01-15T10:00:00Z",
  "last_run": "2026-03-12T14:30:00Z",
  "history": [
    { "date": "2026-03-12T14:30:00Z", "freed_bytes": 2147483648, "count": 5 }
  ]
}
```

Accessible via `goclean --stats` or in the TUI footer.

## Flags

| Flag             | Default | Description                        |
|------------------|---------|------------------------------------|
| `--path <dir>`   | `~`     | Root directory to scan             |
| `--days <n>`     | `30`    | Staleness threshold in days        |
| `--dry-run`      | `false` | Simulate without deleting          |
| `--stats`        | `false` | Show lifetime statistics and exit  |
| `--reset-config` | `false` | Re-run the first-run wizard        |

## Safety

- Automatic skip of dot-directories and system paths.
- Dry-run mode simulates everything without deletion.
- Confirmation screen before any deletion.
- No deletion of the config file itself, only the dependency directory.

## Dependencies

- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — TUI styling
- `github.com/charmbracelet/bubbles` — TUI components (spinner, progress, etc.)
- `gopkg.in/yaml.v3` — YAML config parsing

## Testing

- Unit tests for `filter` (staleness logic with mocked file times)
- Unit tests for `targets` registry
- Unit tests for `scanner` skip logic (dot-dirs, system paths, excluded paths)
- Unit tests for `stats` (accumulation, serialization)
