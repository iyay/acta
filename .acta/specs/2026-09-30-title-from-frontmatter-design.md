---
parent: bugs/2026-09-30-list-shows-slug-when-title-only-in-frontmatter
closes: [SCR-0021]
created: "2026-09-30"
id: SPC-0040
hash: c8lc8cp
---
# Item title falls back to the frontmatter title

Status: design approved by the user in chat on 2026-09-30. Bounded (one existing function, `fileItem` in `internal/board/board.go`, plus its test).

## Why

The board takes an item's title only from the first `# ` heading in the body (`internal/board/parse.go:99`). With no heading it shows the slug (`internal/board/board.go:392`). Scratch items made before commit 3116f9f carry their title only in the frontmatter `title:` field, so 16 files show their slug in the TUI list and in `acta show` (for example SCR-0011 shows `bug-debt-severity`).

## Design

1. In `fileItem`, the title comes from the first place that has one: the body `# ` heading, then the frontmatter `title:`, then the slug.
2. The frontmatter value is read with the existing `field` helper (`board.go:716`). It trims spaces and gives `""` for a missing or null value, so an empty `title:` falls through to the slug.
3. The heading still wins over the frontmatter, so every file that shows a title today shows the same one.
4. This covers every kind that goes through `fileItem`, not only scratch.
5. No file under `.acta/` is changed. The 16 old files show their title once the board reads them again.

## Testing

A test in `internal/board/board_test.go`, red before the code changes, with one case each:

- `title:` and no heading: the title is the frontmatter value.
- A heading and a different `title:`: the title is the heading.
- `title:` empty or only spaces: the title is the slug.
- Neither: the title is the slug.
- `title:` with unicode text: the title is kept whole.

After the build, `acta show SCR-0011` prints `Severity or priority for bugs and debt`.
