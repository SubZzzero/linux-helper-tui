# Phase 6 Completion Plan

## Goal
Close `docs/ROADMAP.md` Phase 6 end-to-end and ship a truthful final milestone for the current catalog-first application.

This plan assumes the approved product decision:
- do not restore the old root search flow
- reinterpret the stale search-specific Phase 6 expectations around the current catalog-first discovery architecture

## Final Phase Definition
`Phase 6 - Milestone M5 And Release Hardening` will be considered complete when all of the following are true:
- the embedded recipe corpus reaches at least `150` validated recipes
- benchmark coverage exists for the current catalog-first architecture and replaces stale search-specific expectations
- required coverage gates are met for the packages that actually back the shipped product
- main TUI flows are covered by automated smoke-style tests, including dangerous confirmation paths
- locale and theme switching are stable across the main flow
- release docs reflect the final shipped architecture and quality gates
- `go test ./... -race`, `golangci-lint run`, and `go build` pass

## Key Product Decisions

### 1. Search milestone realignment
Approved direction:
- keep the root UX browse-only
- do not rebuild `internal/search` just to satisfy stale milestone wording
- update `AGENTS.md`, `docs/ROADMAP.md`, and `README.md` so Phase 6 measures catalog/discovery behavior instead of a removed search subsystem

Resulting metric replacement:
- keep `startup < 200 ms for 200 recipes`
- replace `search < 50 ms` with a catalog/discovery benchmark target such as category filtering or recipe-open preparation under the same `200 recipe` scale
- keep `parse success > 99%`

### 2. Favorites disposition
Keep favorites.

Reason:
- favorites are already integrated across app, storage, UI, tests, and docs
- the roadmap note about maybe removing them is stale and should be removed during Phase 6 closeout

### 3. Phase 6 should include code, tests, and docs
This is not a docs-only closeout.
The verified implementation gap is still substantial.

## Verified Current Baseline
Read-only verification on the current repository shows:
- `go test ./... -race` passes
- `golangci-lint run` is currently clean
- embedded recipe corpus: `103` recipes across `11` categories
- recipe gap to phase target: `47`
- no Go benchmarks exist yet
- package coverage baseline:
  - `internal/recipes`: `55.7%`
  - `internal/executor`: `40.7%`
  - `internal/models`: `58.2%`
  - `internal/app`: `79.8%`
  - `internal/tui/screens`: `71.5%`
- dangerous-flow TUI smoke coverage is incomplete
- persisted invalid theme handling appears brittle at bootstrap time

## Scope
In scope:
- add at least `47` bundled recipes
- raise corpus tests and validation checks to the new thresholds
- add catalog-first performance benchmarks
- raise coverage in the required core packages
- harden locale/theme and dangerous-confirm flows
- update roadmap, changelog, README, and AGENTS milestone language to match the shipped architecture
- verify build, size, tests, and lint before phase closeout

Out of scope:
- restoring the old root search UI
- removing favorites
- broad redesign of the catalog-first interaction model
- adding new external dependencies unless necessary for a benchmark or product-aligned implementation

## Workstreams

### Workstream A. Corpus Growth To `150+`
Target:
- raise corpus from `103` to `150` or more validated recipes

Recommended recipe distribution for exactly `150`:
- `logs`: add `5` recipes, from `8` to `13`
- `packages`: add `5` recipes, from `8` to `13`
- `processes`: add `5` recipes, from `8` to `13`
- `services`: add `5` recipes, from `8` to `13`
- `environment`: add `4` recipes, from `10` to `14`
- `filesystem`: add `4` recipes, from `10` to `14`
- `network`: add `4` recipes, from `10` to `14`
- `system`: add `4` recipes, from `10` to `14`
- `text`: add `4` recipes, from `10` to `14`
- `users`: add `4` recipes, from `10` to `14`
- `troubleshooting`: add `3` recipes, from `11` to `14`

Quality rules:
- stay within offline-safe Linux, DevOps, and SRE workflows
- preserve current YAML schema and naming conventions
- include examples for every new recipe
- add fields only where they improve the form/preview flow
- include a few elevated or dangerous recipes only when they genuinely exercise confirmation behavior

Expected files:
- `assets/recipes/**/*.yaml`
- `internal/recipes/recipes_test.go`
- possibly localized docs that mention corpus counts

### Workstream B. Corpus Validation And Package Coverage
Target:
- meet or exceed `80%` coverage for the required product-relevant core packages

Direct coverage priorities from the measured baseline:
- `internal/executor`
  - add tests for `process.go`: `Run`, `RunShell`, `joinArgs`, `exitCode`
- `internal/models`
  - add tests for `ParseExecutionType`, `ParseRiskLevel`, `Field.Resolve`, `Field.Valid`, `Recipe.Validate` edge cases
- `internal/recipes`
  - add tests for `Loader.Load` override/error branches
  - add tests for `Registry.All`
  - expand parser/validator edge coverage

Search-package decision under the approved catalog-first direction:
- remove `internal/search` from active quality-gate language rather than reviving an empty package
- move the coverage gate to packages that actually underpin discovery today, likely `internal/app` and `internal/tui/screens`, if milestone language needs a fourth package set

Expected files:
- `internal/executor/executor_test.go`
- `internal/models/models_test.go`
- `internal/recipes/recipes_test.go`
- possibly new focused `_test.go` files in those packages
- `AGENTS.md`
- `docs/ROADMAP.md`

### Workstream C. Benchmarks For Catalog-First Discovery
Target:
- add measurable benchmarks that satisfy the hardening intent of Phase 6 without reviving removed search architecture

Recommended benchmark set:
- recipe corpus load + registry creation benchmark around `services.NewRecipeService` and the embedded loader
- catalog initialization benchmark around `screens.NewCatalogModel` using a `150-200` recipe corpus
- category-switch/filter benchmark around `CatalogModel.SetSelectedCategory` / category filtering path
- optional bootstrap-scope benchmark if it can be isolated without flaky filesystem effects

Benchmark files likely needed:
- `internal/recipes/recipes_benchmark_test.go`
- `internal/services/recipe_service_benchmark_test.go`
- `internal/tui/screens/catalog_benchmark_test.go`

Phase metric rewrite:
- benchmark and document startup for `200` recipes
- benchmark and document catalog/discovery responsiveness for `200` recipes
- retain parse-success measurement from corpus tests

### Workstream D. TUI Smoke Hardening
Target:
- close the remaining app-level and screen-level flow gaps

Known gaps to address:
- dangerous recipe `form -> confirm -> result` smoke path at app level
- confirm-screen locale/theme refresh behavior
- confirm-screen keyboard behavior regression tests
- invalid persisted theme fallback during bootstrap
- final keyboard-safety verification so single-letter conflicting actions do not return

Expected files:
- `internal/app/app_test.go`
- `internal/app/bootstrap_test.go`
- `internal/tui/screens/screens_test.go`
- `internal/tui/screens/presentation_test.go`
- `internal/app/bootstrap.go`
- `internal/storage/config.go`
- `internal/tui/theme/theme.go`

### Workstream E. Release Alignment And Phase Closeout
Target:
- make the docs and phase history truthful and final

Required updates:
- `docs/ROADMAP.md`
  - complete Phase 6 task list
  - rewrite search-specific wording into catalog/discovery wording
  - remove the stale “maybe remove favorites” item
  - update current baseline counts and final status
- `AGENTS.md`
  - align milestone M5 and quality-gate wording with the shipped catalog-first architecture
  - stop describing `internal/search` as an active required subsystem if it remains absent
- `README.md`
  - update recipe counts
  - remove stale single-letter shortcut docs
  - remove stale search-architecture wording
  - document the final browse-first workflow and active controls
- `docs/CHANGELOG.md`
  - record the new recipes, benchmarks, hardening work, coverage closeout, and Phase 6 completion

## Execution Sequence
1. Expand the corpus to `150+` first so all later tests and benchmarks use the intended scale.
2. Raise corpus validation tests to the new thresholds and category floors.
3. Add low-level coverage tests for `executor`, `models`, and `recipes` until they clear `80%`.
4. Add catalog-first benchmarks at the same `150-200` recipe scale.
5. Harden app/TUI flows for dangerous confirm, locale/theme refresh, and invalid persisted theme fallback.
6. Update milestone language in `AGENTS.md`, `docs/ROADMAP.md`, `README.md`, and `docs/CHANGELOG.md` to match the shipped architecture.
7. Run full verification and close the phase.

## Delegation Plan
Implementation should be split across parallel agents.

### Agent A. Recipe Expansion
Owns:
- `assets/recipes/**/*.yaml`
- related corpus tests

Deliverables:
- at least `47` new valid recipes
- updated corpus count/category-floor assertions
- report of final category counts

### Agent B. Core Coverage
Owns:
- `internal/executor`
- `internal/models`
- `internal/recipes`

Deliverables:
- new focused tests for all current zero-coverage and low-coverage functions
- package coverage report showing `>=80%` for the required packages

### Agent C. Benchmarks And Discovery Metrics
Owns:
- benchmark files in `internal/recipes`, `internal/services`, and `internal/tui/screens`
- associated metric wording in docs if needed

Deliverables:
- reproducible benchmark suite for startup/corpus/discovery paths
- clear mapping from old search metric intent to current catalog-first benchmarks

### Agent D. TUI Hardening
Owns:
- `internal/app`
- `internal/tui/screens`
- bootstrap/theme fallback paths

Deliverables:
- dangerous confirm smoke coverage
- locale/theme confirm-flow coverage
- invalid-theme fallback fix and tests
- keyboard-safety regression coverage

### Agent E. Docs And Release Closeout
Owns:
- `AGENTS.md`
- `docs/ROADMAP.md`
- `README.md`
- `docs/CHANGELOG.md`

Deliverables:
- final architecture-aligned docs
- completed Phase 6 documentation
- no stale search-first or single-letter hotkey guidance

Recommended parallel order:
- start Agent A, Agent B, and Agent D first
- start Agent C once corpus size and discovery surfaces stabilize
- start Agent E last so docs reflect final measurements and shipped behavior

## Risks And Mitigations

### Risk 1. Recipe expansion becomes shallow filler
Mitigation:
- use category-specific workflow themes and require examples/fields where they improve usability
- prefer balanced category deepening over dumping many recipes into one area

### Risk 2. Coverage climbs slowly in `executor`
Mitigation:
- target `process.go` directly first because it currently contributes the largest zero-coverage surface

### Risk 3. Benchmark design drifts from real product behavior
Mitigation:
- benchmark actual catalog/discovery construction and category filtering paths, not synthetic dead code

### Risk 4. Search wording remains inconsistent after Phase 6
Mitigation:
- treat `AGENTS.md`, `docs/ROADMAP.md`, and `README.md` as a single closeout set and update them together

### Risk 5. Invalid persisted theme breaks startup
Mitigation:
- add a defensive fallback path and cover it in bootstrap tests before final verification

## Verification Checklist
- `go test ./... -race`
- `golangci-lint run`
- `go build -o bin/linux-helper ./cmd/linux-helper`
- `go test ./... -bench . -run ^$`
- coverage check confirms `>=80%` for the required core packages after the search-gate realignment
- embedded corpus count is `>=150`
- per-category counts match or exceed the planned floor
- binary size remains `<20 MB`
- roadmap and AGENTS milestone wording matches the shipped catalog-first architecture
- README keyboard shortcuts match current code behavior

## Success Criteria
- Phase 6 is marked complete without any false claim about restored root search.
- The application remains browse-first and single-binary.
- The corpus reaches at least `150` validated recipes.
- Benchmarks exist and are tied to real catalog-first discovery behavior.
- Coverage gates are satisfied for the final required packages.
- Dangerous confirm, locale/theme switching, and bootstrap fallback flows are hardened and tested.
- Final docs, changelog, and milestone language are internally consistent.
