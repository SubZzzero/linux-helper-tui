## Phase 1 - Foundation

**Status**: [x] Done

### Tasks
- [x] Create Go module, Makefile, and lint configuration
- [x] Create core package structure under `cmd/`, `internal/`, and `assets/`
- [x] Implement embedded recipe, locale, and theme loading
- [x] Implement services, execution primitives, and minimal TUI flow
- [x] Add unit tests and verification commands

### Exit criteria
- All tasks checked
- `go test ./... -race` passes
- `golangci-lint run` clean

## Phase 2 - Productivity

**Status**: [x] Done

### Tasks
- [x] Surface persisted favorites in the TUI and keep them across launches
- [x] Surface recent commands in the TUI
- [x] Expand keyboard navigation for productivity actions
- [x] Add tests that cover persistence-backed TUI state

### Exit criteria
- Favorites and recent commands are visible in the TUI
- `go test ./... -race` passes
- `golangci-lint run` clean

## Phase 3 - Recipe Expansion

**Status**: [x] Done

### Current UI constraint
- The root TUI entry screen is a browse-only recipe catalog.
- Do not reintroduce free-text search in the root flow unless a new design document explicitly replaces the current catalog model.

### Tasks
- [x] Add the first expanded embedded recipe batch for `filesystem` and `system`
- [x] Add a corpus test that validates all embedded recipes
- [x] Add category-aware catalog UI for the current embedded categories
- [x] Expand supported categories beyond `filesystem` and `system`
- [x] Grow the embedded corpus toward milestone-scale coverage with broader workflow coverage across the six active categories

### Exit criteria
- Embedded recipes cover multiple real Linux workflows across all six active categories
- Category-aware catalog navigation works for embedded recipe groups
- The full embedded corpus loads and validates in tests
- `go test ./... -race` passes
- `golangci-lint run` clean

## Strategic Alignment

`docs/ROADMAP.md` is the primary planning document for ongoing work.
`AGENTS.md` defines the project constraints, milestone targets, and quality gates that this roadmap must mirror.

### Current baseline
- `Phase 1` through `Phase 6` are complete.
- The current root TUI flow is a browse-only recipe catalog.
- The embedded corpus currently covers eleven active categories: `filesystem`, `environment`, `logs`, `network`, `packages`, `processes`, `services`, `system`, `text`, `troubleshooting`, and `users`.
- The embedded corpus currently includes `151` validated recipes.
- Locale can now be switched in-app with `ctrl+l` and persists to `~/.config/linux-helper/config.yaml`, with `en` as the default locale.
- Theme can now be switched in-app with `ctrl+t` and persists to `~/.config/linux-helper/config.yaml`, with `dark` as the default theme.
- Do not reintroduce single-letter or destructive text-editing hotkeys for navigation or confirmation on screens that may receive typed input. Keep conflict-prone actions on non-text keys such as `esc`, `enter`, arrows, `tab`, and explicit `ctrl+...` bindings.
- Persisted invalid theme names now fall back to the default embedded theme instead of blocking startup.

### Ongoing quality gates
- Keep the application offline-first and single-binary.
- Keep all assets embedded at build time.
- Run `go test ./... -race` before closing a phase.
- Keep `golangci-lint run` clean before closing a phase.
- Maintain at least `80%` coverage for `recipes/`, `app/`, `executor/`, and `models/`.

### Milestone targets carried forward from `AGENTS.md`
- Reach `150+` bundled recipes.
- Keep the built binary below `20 MB`.
- Measure startup below `200 ms` for `200` recipes.
- Measure catalog discovery operations below `50 ms` for `200` recipes.
- Keep valid recipe parse success above `99%`.

## Phase 4 - Category Expansion

**Status**: [x] Done

### Tasks
- [x] Define the next category set beyond the current six active groups.
- [x] Update `internal/models/category.go`, `assets/recipes/`, and the catalog flow to support the expanded category set.
- [x] Add new recipe categories that cover missing Linux, DevOps, and SRE workflow areas.
- [x] Add an initial high-value recipe batch for each new category.
- [x] Keep validation and catalog tests aligned with the expanded category matrix.

### Completed category expansion
- Add category groups for workflows that were missing from the catalog.
- Expand the corpus with `logs`, `packages`, `processes`, and `services` as the first high-value operational categories beyond the original six.
- Keep category additions grounded in offline-safe Linux workflows that fit the project scope.

### Exit criteria
- The project supports more than the current six active categories.
- Each new category has enough recipes to be immediately useful, not just present in the UI.
- Category coverage is broad enough to represent more Linux and DevOps workflows than the current six-category baseline.
- `go test ./... -race` passes.
- `golangci-lint run` clean.

## Phase 5 - Category Fill-Out And Library Growth

**Status**: [x] Done

### Tasks
- [x] Add the `troubleshooting` category with an initial triage-focused recipe batch.
- [x] Expand the newly added categories with deeper recipe coverage.
- [x] Continue filling the original six categories only after the wider category map is in place.
- [x] Grow the embedded recipe library from the expanded category base toward `100+` recipes.
- [x] Harden the single-binary packaging path and verify all runtime assets stay embedded.
- [x] Improve execution preview, confirmation, and result presentation where the larger corpus exposes UX gaps.
- [x] Add in-app locale switching with a dedicated hotkey and persist the selected locale to `config.yaml`.
- [x] Keep `en` as the default locale when no explicit locale is configured.
- [x] Add in-app theme switching with a dedicated hotkey and persist the selected theme to `config.yaml`.
- [x] Expand recipe examples and field coverage for the broader corpus.
- [x] Keep recipe validation, category coverage checks, and examples aligned with the larger library.

### Exit criteria
- The embedded recipe library reaches at least `100` validated recipes.
- New categories are no longer thin starter buckets and have meaningful workflow depth.
- Locale can be switched in the TUI and the selection persists across launches.
- Theme can be switched in the TUI and the selection persists across launches.
- `go build` produces one binary smaller than `20 MB`.
- `go test ./... -race` passes.
- `golangci-lint run` clean.

## Phase 6 - Milestone M5 And Release Hardening

**Status**: [x] Done

### Tasks
- [x] Grow the embedded recipe library to `150+` bundled recipes from the now-expanded category set.
- [x] Add benchmarks that track startup, catalog discovery, and corpus loading at milestone scale.
- [x] Close coverage gaps in the core packages called out by `AGENTS.md`.
- [x] Run a full TUI smoke pass across catalog, forms, confirmations, execution, and recent commands.
- [x] Smoke-test locale and theme switching hotkeys across the main TUI flow.
- [x] Stabilize the user experience for a larger corpus and broader category set.
- [x] Prepare the project for a release-quality milestone with consistent docs and changelog updates.
- [x] Keep keyboard UX safe for text entry and avoid reintroducing conflicting single-letter shortcuts such as `q`, `r`, `f`, `k`, `j`, `y`, `n`, or `backspace` as action triggers.
- [x] Keep `favorites` as a supported product feature after the final cleanup pass.

### Exit criteria
- The embedded recipe library reaches at least `150` validated recipes.
- Benchmarks exist for the critical performance paths and run cleanly.
- Startup stays below `200 ms` for `200` recipes.
- Catalog discovery operations stay below `50 ms` for `200` recipes.
- Valid recipe parse success stays above `99%`.
- `go build` produces one binary smaller than `20 MB`.
- Coverage targets are met for the required core packages.
- The main TUI flows are verified against the expanded corpus.
- Locale and theme switching remain stable across the main TUI flows.
- `go test ./... -race` passes.
- `golangci-lint run` clean.
