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

**Status**: [~] In progress

### Tasks
- [x] Add the first expanded embedded recipe batch for `filesystem` and `system`
- [x] Add a corpus test that validates all embedded recipes
- [x] Add category-aware search UI for the current embedded categories
- [ ] Expand supported categories beyond `filesystem` and `system`
- [ ] Grow the embedded corpus toward milestone-scale coverage

### Exit criteria
- Embedded recipes cover multiple real Linux workflows
- Category-aware navigation works for embedded recipe groups
- The full embedded corpus loads and validates in tests
- `go test ./... -race` passes
- `golangci-lint run` clean
