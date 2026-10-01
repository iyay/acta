---
parent: bugs/2026-10-01-detail-markdown-ignores-theme-colors
depth: minimal
id: PLN-0074
created: "2026-10-01 18:59:07"
hash: qgrmpcg
started: "2026-10-01 19:06:02"
---
# Detail Markdown Theme Colors Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Paint the detail pane markdown with the picked theme's colors instead of glamour's fixed 256-color presets (fixes `bugs/2026-10-01-detail-markdown-ignores-theme-colors`).

**Spec:** `.acta/specs/2026-10-01-detail-markdown-theme-colors-design.md`

**Tests:** fast `scripts/test ./internal/tui`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- No new dependency. Use `github.com/charmbracelet/glamour/ansi` and `github.com/charmbracelet/glamour/styles` (glamour v1.0.0, already in go.mod).
- Colors follow the rule of `slot` in `internal/tui/styles.go`: a theme with a `BG` gives its hex, the `terminal` theme (empty `BG`) gives the ANSI number as a string. Reuse the `slot*` constants; add none.
- Never write through a pointer taken from `styles.DarkStyleConfig` or `styles.LightStyleConfig`: copy the struct, then set new pointers, so the glamour globals stay unchanged.
- glamour's renderer already defaults to `termenv.TrueColor`, so tests need no profile setup for glamour output.
- Every run step uses `scripts/test ./internal/tui -run <Name>`, never bare `go test` and never `./...`.
- Comments are plain English a 10-year-old can read. They say why, not what.

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2 (needs `markdownStyle` from Task 1, and both touch the test file).

### Task 1: markdown style from the theme

**Files:**
- Create: `internal/tui/markdown.go`
- Test: `internal/tui/markdown_test.go`

**verify:** Every element the spec maps (body, H1 to H6, bold, italic, inline code, links, block quote, rule, the code block tokens comment, keyword, string, number, function name, other text) gets its color only from the theme, for both a hex theme and the `terminal` theme, and no heading keeps a background, a `#` prefix, or the glamour preset color. List every element checked and the slot each one got.

- [x] Failing test: `TestMarkdownStyleUsesThemeSlots` builds `markdownStyle(theme tokyo-night, true)` and `markdownStyle(theme terminal, true)` and checks each mapped field against the spec slot (hex for tokyo-night, `"2"` style numbers for terminal), H1 `BackgroundColor` nil and every heading `Prefix` empty, inline code `BackgroundColor` nil with empty `Prefix`/`Suffix`; it fails because `markdownStyle` does not exist.
- [x] Code: add `func markdownStyle(t theme.Theme, dark bool) ansi.StyleConfig` in `internal/tui/markdown.go`; copy `styles.DarkStyleConfig` or `styles.LightStyleConfig` by `t.Dark(dark)`, then set the colors and markers from the spec mapping, with a fresh `&ansi.Chroma{...}` for code blocks.
- [x] Commit: `Build the detail markdown style from the theme palette`

### Task 2: renderer uses the theme

**Files:**
- Modify: `internal/tui/view.go`
- Modify: `internal/tui/model.go`
- Test: `internal/tui/markdown_test.go`

**verify:** No path that builds a renderer (`New`, `WithTheme`, and the fallback to the default theme in `WithTheme`) can paint markdown with a color outside the active theme, and rendered output never shows the `**`, backtick or `# ` markers. List every caller of `newRenderer` checked and the theme each one passes.

- [ ] Failing test: `TestRendererPaintsThemeColors` renders `"# H\n\n**b** and `+"`c`"+`"` with `newRenderer(tokyo-night, true)` and expects the truecolor sequence for `#9ece6a` (`38;2;158;206;106`), no `38;5;`, and no `**`, backtick or `# `; a second case with the `terminal` theme expects a 16-color code such as `\x1b[32m` and no `38;2;`; it fails because `newRenderer` takes only a bool and uses the glamour preset.
- [ ] Code: change `newRenderer` to `newRenderer(t theme.Theme, dark bool)` and use `glamour.WithStyles(markdownStyle(t, dark))` in place of `glamour.WithStandardStyle`; update the two calls in `internal/tui/model.go` (`New` and `WithTheme`) to pass `t, dark`; fix any test that called the old signature.
- [ ] Commit: `Detail pane markdown uses the theme colors`
