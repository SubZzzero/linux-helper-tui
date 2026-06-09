# Project Initialization Plan

## Goal

Initialize `linux-helper` as a new Go 1.22+ project that matches `AGENTS.md`: single-binary Bubble Tea TUI, embedded assets, layered internal packages, tests, lint/build tooling, and seed data required for the application to compile and boot.

## Current State

- Workspace currently contains only `AGENTS.md` and `.kilo/` service files.
- No Go module, source tree, assets, docs, or build tooling exist yet.
- Initialization therefore needs to create the full repository skeleton from scratch.

## Constraints From AGENTS.md

- Use Go 1.22+ with a toolchain directive.
- Use only the listed dependencies unless explicitly discussed.
- Keep files focused and under roughly 300 lines.
- Add brief comments for exported symbols, functions, and non-trivial blocks.
- Follow dependency direction: `tui -> services -> infrastructure -> models`.
- Embed all bundled assets from `assets/`.
- Add tests for internal packages and run `go test ./... -race` before considering the task complete.
- Provide `Makefile`, `.golangci.yml`, `docs/ROADMAP.md`, and `docs/CHANGELOG.md`.

## Implementation Plan

### 1. Bootstrap repository metadata and tooling

Create the repository foundation:

- `go.mod` with module path, Go version, toolchain directive, and required dependencies.
- `Makefile` with `build`, `test`, `lint`, `cover`, and `clean` targets.
- `.golangci.yml` configured for a small, practical starter lint set.
- `.gitignore` for Go build outputs, coverage, and local config/log files.

### 2. Create application entrypoint and bootstrap layer

Add the minimum executable path:

- `cmd/linux-helper/main.go` to create the app and start Bubble Tea.
- `internal/app/bootstrap.go` to construct dependencies, load assets, and return the root model.
- `internal/app/app.go` for the top-level Bubble Tea model and initial screen stack.

Outcome: the binary can start and render a minimal initial screen even before all features are complete.

### 3. Define core domain models

Create `internal/models/` with the base types required by the rest of the code:

- `recipe.go`
- `category.go`
- `field.go`
- `execution.go`

Include enums and helpers for categories, risk levels, and execution type so loaders, search, and execution code can share one source of truth.

### 4. Seed embedded assets

Create the minimum embedded data needed for a working application:

- `assets/recipes/` with a small starter corpus of valid YAML recipes.
- `assets/locales/en.json`, `ua.json`, and `ru.json`.
- `assets/themes/light.yaml` and `dark.yaml`.

Use a small but representative recipe set rather than aiming for milestone M5 volume during initialization.

### 5. Implement recipe loading pipeline

Create `internal/recipes/`:

- `parser.go` to decode YAML into domain models.
- `validator.go` to validate schema rules and enums.
- `registry.go` to index recipes by ID.
- `loader.go` to read embedded assets and optional user overrides.

Outcome: application startup can load recipes deterministically from embedded assets.

### 6. Implement search indexing

Create `internal/search/`:

- `index.go` to build searchable text entries.
- `fuzzy.go` to wrap `sahilm/fuzzy`.
- `ranking.go` for stable ordering and tie-breaking.

Outcome: TUI and services can search the starter recipe set from memory.

### 7. Implement execution primitives

Create `internal/executor/`:

- `direct.go`
- `shell.go`
- `process.go`
- `risk.go`

Keep execution behind interfaces so tests do not spawn real processes. Add risk gating and result capture structures now, even if preview UX remains minimal.

### 8. Implement infrastructure support packages

Create:

- `internal/i18n/loader.go` and `translator.go`
- `internal/storage/config.go`, `favorites.go`, and `recent.go`
- `internal/logger/logger.go`

Keep these minimal but real so the bootstrap path reflects the intended architecture from the start.

### 9. Implement services layer

Create `internal/services/`:

- `recipe_service.go`
- `search_service.go`
- `execution_service.go`

Services should consume small interfaces and expose application-facing methods for the TUI.

### 10. Implement first TUI screen set

Create a minimal but compilable TUI structure:

- `internal/tui/navigation/` for screen stack helpers.
- `internal/tui/screens/` with at least an initial search/list screen and a detail/placeholder screen.
- Optional small reusable widgets/layout helpers only when needed.

Target for initialization: working navigation shell, recipe list, search input, and basic detail display.

### 11. Add documentation

Create:

- `docs/ROADMAP.md` with Phase 1 marked in progress and explicit tasks.
- `docs/CHANGELOG.md` with an initial unreleased section or first milestone note.

This keeps future work aligned with the documented milestone structure in `AGENTS.md`.

### 12. Add tests package by package

Start with the packages that have the clearest pure logic:

- `internal/models`
- `internal/recipes`
- `internal/search`
- `internal/executor`

Prefer table-driven tests and fakes/stubs for filesystem or process interactions.

### 13. Validate the initialized project

Run and fix issues until these pass:

- `go test ./... -race`
- `golangci-lint run`
- `go build ./cmd/linux-helper`

If lint or tests reveal file-size or architecture issues, split files before considering initialization complete.

## Proposed Delivery Scope

Initialization should produce a compileable, testable starter application with:

- complete repository structure
- seed embedded data
- minimal but real TUI flow
- working recipe load/search path
- execution abstractions
- documentation and tooling

It should not try to fully complete all roadmap milestones in one pass.

## Open Decisions

Default assumptions for implementation unless the user changes them:

- Use a minimal starter recipe corpus, not 150+ recipes.
- Implement a minimal initial TUI with search and recipe detail only.
- Keep themes/config/favorites/recent as simple first-pass implementations.
- Use local filesystem paths under XDG-style config/data directories for config and logs.

## Execution Order

1. Tooling and module files.
2. Models and assets.
3. Recipe/search infrastructure.
4. Services and bootstrap.
5. TUI shell.
6. Executor and storage support.
7. Tests, lint, and build verification.

## Risks To Watch

- Overbuilding the initial TUI before core loaders/services are stable.
- Letting files exceed the project size guideline.
- Introducing dependency direction violations while wiring packages.
- Making storage or execution code difficult to test due to concrete OS/process coupling.
- Spending too much time on seed content instead of a stable compileable scaffold.
