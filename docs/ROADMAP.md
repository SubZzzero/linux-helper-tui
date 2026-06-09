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
