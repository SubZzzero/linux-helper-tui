# linux-helper

`linux-helper` is an offline terminal UI for Linux, DevOps, and SRE workflows. It lets you browse command recipes, fill required fields, preview the final command, and execute it locally without depending on external services.

The project is implemented in Go and uses Bubble Tea for the TUI layer. Recipes, locales, and themes are embedded into the binary at build time.

## Current status

The repository currently includes:

- embedded recipe loading with optional user overrides
- fuzzy search over available recipes
- multi-screen Bubble Tea flow: search, detail, form, confirmation, result
- direct and shell-based execution modes
- risk confirmation for dangerous commands
- embedded locales: `en`, `ua`, `ru`
- embedded themes: `dark`, `light`
- storage for config, favorites, recent commands, and logs

The current bundled recipes include:

- filesystem: `find-file`, `list-directory`, `show-file-head`, `show-file-tail`, `count-files`, `find-large-files`
- system: `disk-usage`, `memory-usage`, `system-uptime`, `kernel-info`, `process-list`, `top-cpu-processes`

## Requirements

- Linux
- Go `1.22+`
- `golangci-lint` for linting

## Quick start

Build the binary:

```bash
make build
```

Run the application:

```bash
./bin/linux-helper
```

Run tests:

```bash
make test
```

Run linter:

```bash
make lint
```

Generate coverage report:

```bash
make cover
```

## Development commands

The repository exposes these Make targets:

- `make build` builds `bin/linux-helper`
- `make test` runs `go test ./... -race -count=1`
- `make lint` runs `golangci-lint run`
- `make cover` writes `coverage.out` and opens an HTML coverage report
- `make clean` removes build and coverage artifacts

## User data paths

The application uses XDG-style paths under the current user's home directory:

- config: `~/.config/linux-helper/config.yaml`
- favorites: `~/.config/linux-helper/favorites.yaml`
- recent commands: `~/.config/linux-helper/recent.yaml`
- recipe overrides: `~/.config/linux-helper/recipes/`
- log file: `~/.local/share/linux-helper/app.log`

If `config.yaml` does not exist, the application falls back to:

- locale: `en`
- theme: `dark`

## Recipe model

Recipes are YAML files grouped by category under `assets/recipes/`. Each recipe defines:

- metadata such as `id`, `category`, and `risk`
- execution mode: `direct` or `shell`
- command arguments or shell command template
- localized title and description
- input fields used by the TUI form
- examples and tags

At startup, the application loads embedded recipes first and then overlays user-provided recipes from `~/.config/linux-helper/recipes/` when present.

## Project layout

```text
cmd/linux-helper/        application entry point
internal/app/           bootstrap and root Bubble Tea model
internal/tui/           screens, navigation, and theme styling
internal/recipes/       recipe loading, parsing, validation, registry
internal/search/        fuzzy search index and ranking
internal/executor/      command execution and risk handling
internal/services/      orchestration layer used by the TUI
internal/storage/       config, favorites, and recent command persistence
internal/i18n/          locale loading and translation
internal/logger/        file-backed slog logger
internal/models/        domain models
assets/                 embedded recipes, locales, and themes
docs/                   roadmap and changelog
```

## TUI flow

The current application flow is:

1. Search for a recipe.
2. Open the recipe detail screen.
3. Fill recipe fields.
4. Preview the resolved command.
5. Confirm execution for dangerous recipes.
6. Execute and inspect the result screen.

## Keyboard shortcuts

- Search: `enter` open recipe, `f` toggle favorite, `j`/`k` or arrows move, `g`/`G` jump to first or last result, `q` quit
- Detail: `enter` or `r` continue to the form, `f` toggle favorite, `esc` or `q` go back
- Form: `tab`, arrows, or `j`/`k` move between fields, `enter` or `ctrl+s` submit, `esc` or `q` go back
- Confirm: `enter` or `y` approve, `esc`, `q`, or `n` cancel
- Result: `enter`, `esc`, or `q` return to the previous screen after execution finishes

## Architecture notes

- The app is a single binary.
- Embedded assets are exposed through `embed.FS`.
- Blocking work stays out of the Bubble Tea update loop.
- Logging goes to a file only, never to standard output.
- Tests cover the core internal packages and TUI transitions.

## Documentation

Additional project context lives in:

- `docs/ROADMAP.md`
- `docs/CHANGELOG.md`
- `AGENTS.md`
