# Plan: troubleshooting category and Phase 5 continuation

## Current state

- `docs/ROADMAP.md` shows `Phase 5 - Category Fill-Out And Library Growth` as the next unfinished phase.
- The embedded corpus currently has `76` recipes across `10` active categories: `filesystem`, `environment`, `logs`, `network`, `packages`, `processes`, `services`, `system`, `text`, `users`.
- Category support is currently hard-coded in `internal/models/category.go`, `internal/tui/screens/catalog.go`, and category assertions in tests.
- The next highest-value step is to expand the corpus with a new high-signal operational category instead of only deepening existing thin buckets.

## Goal

Add a new `troubleshooting` category that is immediately useful for Linux / DevOps / SRE triage, wire it through the catalog and validation path, and keep the repo green with full test verification.

## Scope

1. Add `troubleshooting` as a first-class category in models and catalog UI.
2. Add a non-trivial initial recipe batch for the category.
3. Update corpus and UI tests to recognize the new category.
4. Update roadmap / changelog / README so the docs match the shipped corpus.
5. Run `go test ./... -race` and fix any fallout.

## Files likely to change

- `internal/models/category.go`
- `internal/models/models_test.go`
- `internal/tui/screens/catalog.go`
- `internal/tui/screens/screens_test.go`
- `internal/recipes/recipes_test.go`
- `assets/recipes/troubleshooting/*.yaml`
- `docs/ROADMAP.md`
- `docs/CHANGELOG.md`
- `README.md`

## Implementation plan

### 1. Add category plumbing

Update `internal/models/category.go`:
- Add `CategoryTroubleshooting Category = "troubleshooting"`.
- Include it in `Valid()`.
- Return `"Troubleshooting"` from `DisplayName()`.

Update tests in `internal/models/models_test.go`:
- Add `ParseCategory("troubleshooting")` coverage.
- Add `DisplayName()` coverage for the new category.

### 2. Expose the category in the catalog

Update `internal/tui/screens/catalog.go`:
- Add localized descriptions for `troubleshooting` in `en`, `ru`, and `ua`.
- Keep the wording workflow-oriented, for example:
  - `en`: `Failure triage, diagnostics, and root-cause checks`
  - `ru`: `Разбор сбоев, диагностика и поиск первопричины`
  - `ua`: `Розбір збоїв, діагностика та пошук першопричини`

Update `internal/tui/screens/screens_test.go`:
- Add a sample `troubleshooting` recipe to the catalog fixture, or add a focused test that asserts the category row renders correctly.
- Keep the alignment assertions stable after the new row is introduced.

### 3. Add an initial high-value troubleshooting corpus

Create `assets/recipes/troubleshooting/` and add a strong starter batch. Target `8-10` recipes so the category is useful on day one rather than a thin placeholder.

Recommended recipe set:
- `journal-priority-errors` — `journalctl` view of recent warnings/errors.
- `boot-errors` — current-boot error view via `journalctl -b`.
- `kernel-warnings` — kernel warnings/errors via `dmesg`.
- `deleted-open-files` — leaked disk space / rotated logs via `lsof +L1`.
- `port-owner` — identify which process owns a listening TCP port.
- `tcp-connectivity-check` — validate host:port reachability with timeout.
- `dns-host-records` — inspect resolver output for a hostname.
- `inode-usage` — detect inode exhaustion on a filesystem.
- `systemd-critical-chain` — inspect boot/service dependency latency.
- `failed-unit-logs` or `recent-service-failures` — triage failing units from logs.

Recipe selection rules:
- Prefer offline-safe commands already common on Linux systems.
- Prefer direct execution when possible; use shell execution only when the workflow truly needs piping / filtering.
- Avoid near-duplicates of existing `logs`, `services`, `network`, and `processes` recipes; each troubleshooting recipe should answer an operator question, not merely restate a command family.
- Keep fields concrete and reusable: `unit`, `host`, `port`, `lines`, `path`, `timeout`.
- Add at least one example per recipe.

### 4. Raise corpus expectations in tests

Update `internal/recipes/recipes_test.go`:
- Raise total recipe floor above `76` to match the final added batch.
- Increase expected category count from `10` to `11`.
- Add representative `troubleshooting` recipe IDs to the corpus assertions.
- Add a per-category minimum for `troubleshooting` based on the final batch size.

Use the final exact minimum only after the recipe count is known. If the batch lands at `8`, assert `>= 8`; if it lands at `10`, assert `>= 10`.

### 5. Align documentation with the shipped state

Update `docs/ROADMAP.md`:
- Mark `Phase 5` as started if the implementation meaningfully advances category fill-out and library growth.
- Replace references to `10 active categories` with `11 active categories`.
- Mention `troubleshooting` in the baseline category list.

Update `docs/CHANGELOG.md`:
- Add unreleased entries for the new category and recipe corpus growth.

Update `README.md`:
- Refresh the current bundled category summary.
- Add `troubleshooting` to the catalog/category description section.
- Keep the README concise; it does not need a full recipe inventory, only representative coverage.

## Verification plan

1. Run `go test ./... -race`.
2. If available in the environment, run `golangci-lint run` because roadmap quality gates still require it.
3. If a catalog rendering test becomes brittle after the new category row, adjust the fixture rather than weakening the assertion.
4. Confirm the embedded corpus loads from `assets/recipes/troubleshooting/` without special loader changes.

## Notes and risks

- Some troubleshooting commands have distro or package variability (`nc`, `lsof`, `systemd-analyze`). Prefer commands already represented elsewhere in the corpus when possible, and keep descriptions clear so the operator understands what the recipe assumes.
- `journalctl` and `dmesg` recipes may need `risk: elevated`; match the existing risk conventions used in `logs` and `services` recipes.
- The new category is cross-cutting by design. Keep its contents symptom-oriented so it does not become a duplicate taxonomy of the existing command-family categories.

## Definition of done

- `troubleshooting` is recognized by models, catalog, and tests.
- The repo contains a useful starter set of troubleshooting recipes, not fewer than `8`.
- Corpus tests reflect `11` categories and the new total floor.
- Docs match the new shipped category map.
- `go test ./... -race` passes.
