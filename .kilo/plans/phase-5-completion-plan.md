# Phase 5 Completion Plan

## Goal
Close `docs/ROADMAP.md` Phase 5 end-to-end without prematurely expanding scope into Phase 6 / Milestone M5.

## Current State
- Embedded corpus: `87` recipes across `11` categories.
- Phase 5 exit floor: `100+` validated recipes, richer category depth, locale/theme switching in-app, packaging hardening, broader recipe-field coverage, and aligned validation checks.
- Thin categories: `logs`, `packages`, `processes`, `services` each have `4` recipes.
- Locale/theme defaults already exist in `internal/storage/config.go`, but runtime switching is not wired into the TUI.
- Embedded assets are already loaded through `embed.FS` in `assets.go` and `internal/app/bootstrap.go`.
- Recipe examples are already present in `87/87` recipes; field coverage is the weaker area (`49/87` recipes have fields).
- Tests pass today, but Phase 5-specific thresholds and flows are still incomplete.

## Scope Boundary
This plan finishes **Phase 5 only**.
It does **not** attempt to complete Phase 6 / M5 items such as `150+` recipes, benchmark gates, or `internal/search` restoration unless they become necessary for a Phase 5 change.

## Success Criteria
- Raise the embedded corpus from `87` to at least `100` validated recipes.
- Deepen the thin categories so they no longer feel like starter buckets.
- Implement in-app locale switching with persistence to `~/.config/linux-helper/config.yaml`.
- Implement in-app theme switching with persistence to `~/.config/linux-helper/config.yaml`.
- Keep `en` and `dark` as defaults when config values are absent.
- Improve preview/confirmation/result behavior where the larger corpus exposes rough edges.
- Raise recipe corpus tests to the new thresholds and cover switching/persistence flows.
- Confirm build/test/lint still pass and the binary remains single-file and under the size target.

## Workstreams

### 1. Corpus Growth And Category Fill-Out
Target outcome:
- Add at least `13` recipes, but prefer `16-24` so the corpus lands comfortably above the floor.
- Prioritize `logs`, `packages`, `processes`, and `services` first.

Recommended target distribution:
- `logs`: add `4-6` recipes
- `packages`: add `4-6` recipes
- `processes`: add `4-6` recipes
- `services`: add `4-6` recipes
- Optional top-up in `environment`, `network`, or `system` only if it improves field richness

Recipe quality rules:
- Reuse the existing YAML schema and naming conventions.
- Prefer offline-safe inspection and triage workflows.
- Add fields where parameters materially improve the form/preview flow.
- Keep examples complete for every new recipe.
- Introduce at least a small number of elevated/dangerous recipes only if they exercise the confirmation path meaningfully and remain appropriate for the product.

### 2. Recipe Schema Richness
Target outcome:
- Improve the proportion of recipes that are truly form-driven where it makes sense.
- Focus especially on weak categories like `environment`, `network`, and `system` if low-effort parameterization is available.

Recommended approach:
- Audit zero-field recipes and convert only the ones that naturally benefit from user input.
- Avoid artificial parameters just to increase counts.
- Keep static inspection recipes static when user input would not improve usability.

### 3. Runtime Locale Switching
Target outcome:
- Add a global `l` hotkey that cycles embedded locales and persists the selection.

Implementation outline:
- Extend `internal/app/bootstrap.go` and `internal/app/app.go` so the app model owns enough locale metadata to rotate choices and rebuild screen text.
- Persist locale updates through `internal/storage/config.go`.
- Refresh the active/root screen stack after a locale change so all screen labels and localized recipe text redraw consistently.
- Preserve existing fallback behavior: default `en`, translator fallback to `en` for missing keys.

Likely code areas:
- `internal/app/bootstrap.go`
- `internal/app/app.go`
- `internal/storage/config.go`
- `internal/tui/screens/*.go`
- `README.md`, `docs/ROADMAP.md`, `docs/CHANGELOG.md`

### 4. Runtime Theme Switching
Target outcome:
- Add a global `t` hotkey that cycles embedded themes and persists the selection.

Implementation outline:
- Let the app model keep loaded theme definitions and current theme name.
- Recompute `uitheme.Styles` on switch and propagate refreshed styling through the screen stack.
- Persist the chosen theme through `internal/storage/config.go`.
- Preserve existing default behavior: `dark` when missing.

Likely code areas:
- `internal/app/bootstrap.go`
- `internal/app/app.go`
- `internal/storage/config.go`
- `internal/tui/screens/*.go`
- `internal/tui/theme/*`

### 5. UX Hardening For Larger Corpus
Target outcome:
- Remove obvious friction exposed by the larger recipe set.

Primary checks:
- Catalog help text advertises `l` and `t` once implemented.
- Detail/form/confirm/result screens still read well after locale/theme refresh.
- Result screen remains usable after execution without requiring a manual resize.
- Confirmation flow is still clear if more elevated/dangerous recipes are added.
- Preview text stays readable for richer field-driven recipes.

Guardrail:
- Keep edits minimal; do not redesign the browse-first root flow.

### 6. Test And Validation Alignment
Target outcome:
- Raise automated checks to the new Phase 5 baseline.

Required additions:
- Update `internal/recipes/recipes_test.go` to assert `>=100` recipes and stronger per-category minimums.
- Add app-level tests for locale cycling, theme cycling, and config persistence hooks.
- Add storage tests for saving/loading locale and theme defaults.
- Add TUI regression coverage if locale/theme refresh touches screen behavior.
- Re-run package coverage and close only the gaps directly caused by Phase 5 changes.

Suggested new thresholds after corpus expansion:
- Overall corpus: `>=100`
- `logs`, `packages`, `processes`, `services`: each `>=8`
- Existing strong categories: keep current minimums unless corpus strategy deliberately raises them further

### 7. Packaging And Release Hardening For Phase 5
Target outcome:
- Explicitly verify the existing single-binary path still holds after the larger corpus.

Checks:
- `go build -o bin/linux-helper ./cmd/linux-helper`
- resulting binary remains below `20 MB`
- embedded locales/themes/recipes still load without relying on repo files at runtime
- docs reflect implemented locale/theme behavior instead of planned behavior

## Execution Order
1. Expand recipes in thin categories until the corpus is safely above `100`.
2. Raise corpus tests to reflect the new category floors.
3. Implement locale switching and persistence.
4. Implement theme switching and persistence.
5. Harden the screen refresh path so all active screens pick up locale/theme changes cleanly.
6. Polish preview/confirmation/result flows only where the new corpus reveals real issues.
7. Update docs and changelog to mark Phase 5 complete.
8. Run final verification: `go test ./... -race`, `golangci-lint run`, `go build`.

## Agent Delegation Plan
Use parallel agents during implementation to keep workstreams isolated:

### Agent A: Recipe Expansion
- Add new YAML recipes for `logs`, `packages`, `processes`, `services`.
- Keep schema/examples consistent.
- Return final counts by category and any validation concerns.

### Agent B: Locale/Theme Switching
- Implement app-model support for cycling locale/theme, stack refresh, and persistence.
- Return touched files, hotkey behavior, and any screen-refresh caveats.

### Agent C: Test Alignment
- Update corpus thresholds and add tests for switching/persistence.
- Return coverage deltas and any remaining weak spots.

### Agent D: Docs/Verification
- Update `README.md`, `docs/ROADMAP.md`, and `docs/CHANGELOG.md` after behavior is complete.
- Run final read-only verification commands during planning handoff or full verification during implementation.

Recommended sequence:
- Start Agent A and Agent B in parallel.
- Start Agent C once recipe counts and switching interfaces stabilize.
- Start Agent D last so docs match final behavior.

## Risks And Mitigations
- `internal/app/app.go` will likely grow if switching is added naively.
  - Mitigation: extract small refresh helpers instead of piling logic into `Update`.
- Screen models currently capture `locale` and `styles` at construction time.
  - Mitigation: rebuild or refresh screens centrally from the root app model.
- Thin-category expansion can hit `100+` while still feeling shallow.
  - Mitigation: treat `8` recipes in each thin category as the practical minimum.
- Confirmation flow is lightly exercised because current corpus has no dangerous recipes.
  - Mitigation: either add at least one appropriate dangerous recipe or strengthen confirm tests with synthetic models.

## Out-Of-Scope But Important After Phase 5
These are Phase 6 / M5 follow-ups and should not be silently absorbed into this phase:
- grow corpus to `150+`
- add benchmarks
- reconcile the empty `internal/search` package with AGENTS/README milestone language
- close the broader `80%` coverage targets for `recipes`, `executor`, and `models`

## Verification Checklist
- `go test ./... -race`
- `golangci-lint run`
- `go build -o bin/linux-helper ./cmd/linux-helper`
- confirm corpus count and per-category floors in tests
- confirm locale switch persists across restart
- confirm theme switch persists across restart
- confirm defaults remain `en` and `dark` when config is absent or empty
- confirm the binary size remains under `20 MB`
