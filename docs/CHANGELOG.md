## Unreleased

- Initialize the `linux-helper` repository structure.
- Add a minimal Bubble Tea application scaffold with embedded assets.
- Add loader, search, storage, executor, and service foundations.
- Add embedded theme loading and centralized TUI styling.
- Extend the TUI flow to collect recipe fields, confirm dangerous commands, execute recipes, and display results.
- Add tests for theme loading, screen transitions, execution flow, and race-safe verification.
- Add persisted favorites with TUI toggle support, favorite-aware search ordering, and coverage for the new flow.
- Add recent command history to the search screen and refresh it after command execution.
- Add productive keyboard shortcuts across search, detail, form, confirmation, and result screens.
- Mark `Phase 2` complete after the lint gate passes.
- Start `Phase 3` with eight new embedded recipes and a corpus test for all bundled recipe files.
- Add category filters and grouped search results for the embedded recipe categories.
- Expand category support to `network` and `text` and add eight embedded recipes for those workflows.
