# Goclean

A fast, interactive TUI tool that scans your filesystem for stale dependency folders (`node_modules`, `vendor`, `.venv`, etc.) and lets you bulk-delete them to reclaim disk space.

Built in Go with parallel scanning and a [Bubbletea](https://github.com/charmbracelet/bubbletea) terminal interface.

## Features

- **Smart detection** — only flags projects whose config file (`package.json`, `composer.json`, etc.) hasn't been modified in N days
- **Parallel scanning** — uses a bounded goroutine pool to scan large filesystems fast
- **Interactive TUI** — browse results, select/deselect with spacebar, see total space to free in real-time
- **First-run wizard** — choose which dependency types to monitor on first launch
- **Dry-run mode** — simulate deletion without removing anything
- **Lifetime statistics** — tracks how much space you've freed over time
- **Configurable** — threshold days, excluded paths, and target selection saved to `~/.goclean.yaml`
- **Safe** — automatically skips dot-directories (`.git`, etc.) and system paths (`/System`, `/Library`, etc.)

## Supported Targets

| Ecosystem    | Dependency Folder | Config File        |
|-------------|------------------|--------------------|
| Node.js     | `node_modules`   | `package.json`     |
| PHP         | `vendor`         | `composer.json`    |
| Python      | `.venv` / `venv` | `requirements.txt` |
| Rust        | `target`         | `Cargo.toml`       |
| Java Maven  | `target`         | `pom.xml`          |
| Go          | `vendor`         | `go.mod`           |
| Dart/Flutter| `.dart_tool`     | `pubspec.yaml`     |
| CocoaPods   | `Pods`           | `Podfile`          |
| Gradle      | `build`          | `build.gradle`     |

Targets are opt-in — you choose which ones to monitor during the setup wizard.

## Requirements

- **Go 1.22+** — install with `brew install go` (macOS) or see [go.dev/dl](https://go.dev/dl)
- After installing Go, add `~/go/bin` to your PATH if not already there:
  ```bash
  echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
  source ~/.zshrc
  ```

## Installation

### From source

```bash
git clone https://github.com/corgab/goclean.git
cd goclean
go install ./cmd/goclean
```

Now `goclean` is available globally from any directory.

### Build locally (without global install)

```bash
go build -o goclean ./cmd/goclean
./goclean
```

## Usage

### Basic

```bash
# Scan home directory with default settings (30 days threshold)
goclean
```

On first run, the setup wizard will ask you which dependency types to monitor.

### Flags

| Flag             | Default   | Description                                    |
|------------------|-----------|------------------------------------------------|
| `--path <dir>`   | `~`       | Root directory to scan                         |
| `--days <n>`     | `30`      | Show projects inactive for more than N days    |
| `--dry-run`      | `false`   | Simulate deletion without removing anything    |
| `--stats`        | `false`   | Show lifetime statistics and exit              |
| `--reset-config` | `false`   | Re-run the setup wizard                        |

### Examples

```bash
# Scan only your Dev folder
goclean --path ~/Dev

# Show everything inactive for more than 7 days
goclean --days 7

# Simulate without deleting
goclean --dry-run

# Check how much space you've freed over time
goclean --stats

# Re-select which targets to monitor
goclean --reset-config
```

### TUI Controls

**Setup Wizard:**

| Key       | Action              |
|-----------|---------------------|
| `Space`   | Toggle selection    |
| `a`       | Toggle all          |
| `j` / `k` | Navigate down/up   |
| `Enter`   | Confirm             |
| `q` / `Esc` | Quit              |

**Main View:**

| Key       | Action                  |
|-----------|-------------------------|
| `Space`   | Toggle selection        |
| `a`       | Select all              |
| `n`       | Deselect all            |
| `j` / `k` | Navigate down/up       |
| `Enter`   | Proceed to delete       |
| `q` / `Esc` | Quit                  |

**Confirmation:**

| Key | Action         |
|-----|----------------|
| `y` | Confirm delete |
| `n` | Go back        |

## Configuration

Goclean stores its config at `~/.goclean.yaml`:

```yaml
days: 30
targets:
  - node_modules
  - vendor
excluded_paths:
  - /Users/me/important-project
```

- **days** — default staleness threshold (overridable with `--days`)
- **targets** — dependency folder names to scan for (set during wizard)
- **excluded_paths** — directories to skip entirely during scan

Edit this file directly or use `--reset-config` to re-run the wizard.

## How It Works

1. **Walks** the filesystem from the root directory (default `~`)
2. **Skips** dot-directories, system paths, and excluded paths
3. **Finds** directories matching your selected targets (`node_modules`, `vendor`, etc.)
4. **Checks** the config file in the parent directory (`package.json`, `composer.json`, etc.)
5. **Filters** — only shows projects where the config file is older than the threshold
6. **Disambiguates** shared directory names (e.g. `vendor` → checks for `composer.json` vs `go.mod`)
7. **Calculates** sizes in parallel using a bounded goroutine pool
8. **Displays** results in an interactive TUI for selection and deletion

Only the dependency folder is deleted (e.g. `node_modules/`), never the project itself.

## Statistics

Goclean tracks lifetime cleaning stats in `~/.goclean_stats.json`:

```bash
$ goclean --stats
Goclean has freed 14.3 GB in 47 operations since 15 Jan 2026
```

## Project Structure

```
goclean/
├── cmd/goclean/          # Entry point, flag parsing
├── internal/
│   ├── config/           # YAML config load/save
│   ├── scanner/          # Parallel filesystem scanner
│   ├── filter/           # Staleness detection
│   ├── cleaner/          # Directory deletion with dry-run
│   ├── stats/            # Lifetime statistics (JSON)
│   ├── fsutil/           # Shared utilities (DirSize, FormatBytes)
│   └── tui/              # Bubbletea TUI (wizard + main view)
├── pkg/
│   └── targets/          # Target registry
├── go.mod
└── go.sum
```

## Running Tests

```bash
go test ./... -v
```

## License

MIT
