# All Mode Category-First Design

## Objective

Replace the current large mixed `All` recipe list with a category-first view that scales as the embedded corpus grows.

## Problem

The current `All` mode renders recipes as one long grouped list.
This works for a small corpus, but it will become noisy and hard to scan when more categories and recipes are added.

## Decision

In `All` mode, the search screen will show categories first instead of recipes.

Each category row will show:

- category name
- recipe count for the current query
- optional visual marker when the row is selected

Pressing `enter` on a category row will switch the active filter from `All` to that category.
After that transition, the screen will show recipe rows for the selected category, using the existing recipe list interaction.

## Interaction Model

### All mode

- `All` is the default search mode.
- The result pane shows category rows, not recipe rows.
- Only categories that have at least one matching recipe for the current query are shown.
- The query still applies immediately while typing.

### Category mode

- `left/right` changes the active category filter.
- Entering a category from `All` behaves the same as switching to that category filter.
- Inside a concrete category, the result pane shows recipe rows for that category only.

### Search behavior

- The query continues to match recipes, not category names alone.
- Category counts in `All` are computed from the filtered recipe set.
- Empty state in `All` appears when no category has matching recipes.

## Rendering

### All mode rendering

- show category filter chips at the top, as today
- show the search input below them
- render one row per matching category
- row format should stay compact and text-first

Recommended row shape:

`Filesystem (6)`

### Category mode rendering

- keep the current recipe list behavior
- keep favorites, selection, and recent-commands block unchanged

## State changes

The search model needs two result views derived from the same filtered corpus:

- filtered categories for `All`
- filtered recipes for a concrete category

The selected row should reset safely when moving between `All` and a concrete category.

## Error handling

- if search fails, preserve the current behavior and show the existing error state
- if a category becomes empty because of a query change, it should disappear from `All`
- if the current selection becomes invalid after filtering, clamp selection to the last valid row

## Testing

Add coverage for:

- `All` rendering category rows instead of recipe rows
- counts updating with the current query
- `enter` on a category row switching into that category
- `left/right` preserving category navigation
- empty state when no categories match the query

## Scope boundaries

This change does not add:

- nested tree navigation
- collapsible category sections inside `All`
- separate screens for categories and recipes
- new recipe categories

Those can be considered later if the corpus keeps growing.
