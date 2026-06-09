# Phase 1 Completion Plan

## Context

Based on `AGENTS.md`, `docs/ROADMAP.md`, `docs/CHANGELOG.md`, and the current repository state:

- `Phase 1` is marked `in progress`.
- Project scaffolding, core package layout, embedded recipe loading, locale loading, services, executor foundations, and a minimal TUI already exist.
- The main gaps are incomplete theme support, missing execution-capable TUI flow, and insufficient verification depth for Phase 1 exit criteria.

## Current State Summary

Completed:
- Go module, `Makefile`, and `.golangci.yml`
- Core package structure under `cmd/`, `internal/`, and `assets/`
- Embedded recipe loading with parser, validator, and registry
- Locale loading and translator wiring
- Search index and service foundations
- Executor primitives and execution service foundations
- Minimal Bubble Tea flow: search screen, detail screen, navigation stack
- Initial tests across major packages

Incomplete or partially complete:
- Theme loading and selected theme application
- TUI flow for filling fields, previewing commands, risk confirmation, execution, and result display
- Stronger service wiring from bootstrap into the app model for execution lifecycle
- Deeper tests for error paths and integration behavior
- Verification against Phase 1 exit criteria

## Implementation Objectives

1. Complete embedded asset handling so recipes, locales, and themes are all loaded consistently.
2. Extend the TUI from read-only browsing to a minimal executable flow.
3. Keep package boundaries aligned with `AGENTS.md` dependency rules.
4. Add tests that validate both happy paths and critical failure paths.
5. Finish with `go test ./... -race` and `golangci-lint run` passing.

## Execution Plan

### 1. Baseline inspection before edits

- Review the concrete implementations of:
  - `internal/app/bootstrap.go`
  - `internal/app/app.go`
  - `internal/tui/screens/*.go`
  - `internal/services/execution_service.go`
  - `internal/executor/*.go`
  - `internal/storage/config.go`
- Confirm how theme, execution, and recipe field data should enter the root app model.
- Identify any file that risks exceeding the ~400 line limit before adding behavior.

### 2. Finish theme loading

- Add a dedicated theme loader package/file if none exists yet.
- Load theme definitions from embedded `assets/themes/*.yaml` using `embed.FS`.
- Decide on the smallest viable runtime theme representation for the current TUI.
- Wire the configured theme from storage/bootstrap into the app model.
- Apply theme values through Lip Gloss styles instead of hardcoded presentation values.

Definition of done:
- App bootstrap can load the configured theme from embedded assets.
- TUI rendering uses that theme through a central style path.
- Theme loading failure returns contextual errors.

### 3. Wire execution service into the app

- Ensure `bootstrap` constructs and injects the execution-related services needed by the TUI.
- Verify the root app model owns the dependencies required for:
  - recipe selection
  - field value collection
  - risk checking
  - command execution
  - result rendering
- Add or refine app-level messages for async execution updates, following Bubble Tea conventions from `AGENTS.md`.

Definition of done:
- Root app model can initiate execution-related flows without directly touching infrastructure code.

### 4. Extend the TUI flow from detail view to execution

Implement the minimal end-to-end interaction sequence:

- Search list
- Recipe detail
- Field input form for recipe parameters
- Command preview or summary screen
- Risk confirmation for elevated or dangerous commands
- Execution trigger
- Result/output screen with stdout, stderr, and exit status

Implementation notes:
- Keep each screen as its own Bubble Tea model if the current design already trends that way.
- If the existing detail screen can be extended cheaply without becoming too large, prefer the smaller change.
- Run blocking execution only inside `tea.Cmd` functions.
- Never call process execution directly from `Update`.

Definition of done:
- A user can select a recipe, provide required values, confirm risk if needed, execute it, and see the result in the TUI.

### 5. Tighten service and domain validation paths

- Verify that recipe argument interpolation handles required fields and defaults correctly.
- Review risk gating behavior for safe, elevated, and dangerous commands.
- Confirm `ExecutionService` records recent commands only on the intended execution path.
- Check error wrapping and sentinel errors where callers need structured handling.

Definition of done:
- Service behavior is explicit, predictable, and testable across normal and failure paths.

### 6. Expand test coverage where current gaps are most likely

Add or strengthen tests for:
- Theme loading success and failure cases
- Bootstrap integration for recipes, locales, and themes
- Recipe loading failures:
  - invalid YAML
  - duplicate IDs
  - invalid enum values
- Execution service behavior:
  - direct execution path
  - shell execution path
  - risk confirmation requirements
  - runner errors
- TUI navigation for the minimal execution flow
- Result rendering and message transitions

Testing approach:
- Use table-driven tests where multiple cases exist.
- Keep executor tests isolated from real process spawning.
- Use stubs/interfaces at consumer boundaries as required by `AGENTS.md`.

Definition of done:
- Tests cover the newly added paths and the most important regressions for Phase 1.

### 7. Verify Phase 1 exit criteria

Run and fix issues until both succeed:
- `go test ./... -race`
- `golangci-lint run`

Then update docs to reflect the actual project state:
- Update `docs/ROADMAP.md` task checkboxes/status only when exit criteria are truly satisfied.
- Update `docs/CHANGELOG.md` with concrete completed work for the milestone.

Definition of done:
- `Phase 1` can be marked complete only after code, tests, and lint all pass.

## Recommended Implementation Order

1. Inspect bootstrap, app model, and current screens in detail.
2. Implement theme loading and central styling path.
3. Wire execution dependencies into bootstrap/app.
4. Implement the smallest viable field-entry to execution TUI flow.
5. Add and refine tests around new behavior.
6. Run verification and fix breakages.
7. Update `ROADMAP.md` and `CHANGELOG.md` last.

## Risks and Controls

- Risk: TUI files may grow too large.
  - Control: split by screen/state instead of adding more branching into one file.
- Risk: Theme support may sprawl into premature design-system work.
  - Control: implement only the minimum style surface used by the current screens.
- Risk: Execution flow may couple TUI directly to infrastructure.
  - Control: keep orchestration in services and app-level interfaces/messages.
- Risk: Tests may depend on real filesystem or processes.
  - Control: preserve stubbed interfaces and embedded/test FS usage.

## First Implementation Slice

The first practical slice after plan mode should be:

1. Read the current app/bootstrap/screen/execution files in full.
2. Implement theme loading plus app-level style injection.
3. Wire execution service into the root app model.
4. Add the smallest field-input and execution-result flow for one selected recipe path.
5. Backfill tests for those additions before expanding polish.
