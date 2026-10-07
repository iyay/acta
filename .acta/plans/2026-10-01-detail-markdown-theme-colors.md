---
parent: bugs/2026-10-01-detail-markdown-ignores-theme-colors
depth: minimal
id: PLN-0074
created: "2026-10-01 18:59:07"
hash: qgrmpcg
started: "2026-10-01 19:06:02"
finished: "2026-10-01 19:37:56"
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

- [x] Failing test: `TestRendererPaintsThemeColors` renders `"# H\n\n**b** and `+"`c`"+`"` with `newRenderer(tokyo-night, true)` and expects the truecolor sequence for `#9ece6a` (`38;2;158;206;106`), no `38;5;`, and no `**`, backtick or `# `; a second case with the `terminal` theme expects a 16-color code such as `\x1b[32m` and no `38;2;`; it fails because `newRenderer` takes only a bool and uses the glamour preset.
- [x] Code: change `newRenderer` to `newRenderer(t theme.Theme, dark bool)` and use `glamour.WithStyles(markdownStyle(t, dark))` in place of `glamour.WithStandardStyle`; update the two calls in `internal/tui/model.go` (`New` and `WithTheme`) to pass `t, dark`; fix any test that called the old signature.
- [x] Commit: `Detail pane markdown uses the theme colors`

## Fix round 1

### Task 3: code blocks follow the theme and never crash

**Files:**
- Modify: `internal/tui/markdown.go`
- Modify: `internal/tui/view.go`
- Test: `internal/tui/markdown_test.go`

**verify:** A fenced code block renders without a panic for every built-in theme, the `terminal` theme included, and for a theme with a `BG` every code token (comment, keyword, string, number, function name, other text) comes out as the truecolor sequence of its spec slot, with no `38;5;` anywhere in the output. No element glamour can draw (images, image text, HTML blocks and spans included) keeps a glamour preset color. List every theme and every element checked in the rendered output.

- [x] Failing test: add rendered cases to `internal/tui/markdown_test.go` that feed a ```` ```go ```` fence (a comment, a keyword, a string, a number and a function name) plus an image and an inline HTML span through `newRenderer` for `tokyo-night` and for `terminal`; the `terminal` case panics today (chroma only takes `#rrggbb`, `slot()` gives `"8"`), and the `tokyo-night` case fails because tokens come out as `38;5;` (glamour's default chroma formatter is `terminal256`); render the `tokyo-night` code block before any other hex theme in the test, because glamour registers its chroma style once per process under the name `charm`.
- [x] Code: in `markdownStyle`, leave `cfg.CodeBlock.Chroma` nil when `t.BG == ""` so the `terminal` theme draws code blocks as plain body text; in `newRenderer`, add `glamour.WithChromaFormatter("terminal16m")` when the theme has a `BG`; set `Image`, `ImageText`, `HTMLBlock` and `HTMLSpan` colors from theme slots (image slot 4, image text slot 8, HTML slot 8); rewrite the wrong comment above `cfg.CodeBlock.Chroma` (glamour registers the chroma once per process under `charm`, so the first theme that draws a code block sets the code colors for the process; today the TUI picks one theme per process, so that is fine); finish the comment that stops mid-sentence ("..., which") near the top of `markdown_test.go`.
- [x] Commit: `Code blocks in the detail pane follow the theme and never crash`

Ruling on fix round 1 (orchestrator, 2026-10-01): the code does not leave the terminal theme's chroma nil as the Code step said. It registers one chroma style per theme and gives the terminal theme ANSI names with the `terminal16` formatter. Spec item 4 maps every code token to a slot for every theme, and spec item 3 says the terminal theme uses the plain ANSI color, so this design is kept. It also removes glamour's once-per-process `charm` limit.

## Fix round 2

### Task 4: terminal code tokens use the normal shades

**Files:**
- Modify: `internal/tui/markdown.go`
- Test: `internal/tui/markdown_test.go`

**verify:** For the terminal theme, every code token in a fenced block comes out as the plain 16-color code of its spec slot (comment 90, keyword 35, string 32, number 33, function name 34), the same green and blue that inline code and links get in the same pane, and no token uses a bright code (91 to 97) its slot does not name. List every token checked and the code it rendered as.

- [x] Failing test: in `internal/tui/markdown_test.go`, change the expected terminal codes for string and function name in `ansiToken` from `"92"` and `"94"` to `"32"` and `"34"`; it fails because `ansiChroma` maps `slotGreen` to `#ansigreen` and `slotBlue` to `#ansiblue`, which chroma treats as the bright `00ff00` and `0000ff`.
- [x] Code: in `ansiChroma` use `#ansidarkgreen` for `slotGreen` and `#ansidarkblue` for `slotBlue`; in the same commit, fix the review comments in the two files: drop the `ponytail:` marker near `codeStyle` and finish its unfinished sentence in plain words; correct the comment above `ansiChroma` so it names the normal shades; correct the test comment that says termenv cuts the digits (chroma writes the hex bytes as they are); update the test comments that say tests skip `t.Parallel` because chroma keeps one style list; finish the comment that stops at "..., which" near the top of `markdown_test.go`; spell it "color" in the new code and comments (rename the `colour` helper to `color`-style naming that does not clash), to match the rest of `internal/tui`.
- [x] Commit: `Terminal theme code tokens use the normal ANSI shades`

## Review notes

- The recipient filed BUG-0027 for a heading color defect in this branch's own code and deleted the file in the fix commit 36c08d9; net diff is zero.
- `internal/tui/markdown.go` imports `glamour/ansi` under the name `glamour`, while `view.go` uses `glamour` for the root package.
- The fnv hash in the chroma style name guards a theme file that changes between two picks, which cannot happen while `WithTheme` runs once per process; `"acta-" + t.Name` would do.
- `cfg.HTMLSpan` has no visible effect because glamour strips inline tags; `TestAnInlineTagLosesItsTagsAndKeepsItsWords` passes on the base code too.
- `TestNoGlamourPresetColorSurvives` looks for three preset codes only, not for any `38;5;`; a wider scan found none.
- The bright-code scan in `markdown_test.go` reads every SGR part, so a truecolor `38;2;91;...` would trip it; the terminal theme never emits truecolor.
- The polish dropped the sentence that said why drawing every theme in one process proves each keeps its own code colors.
- The `chromaLock` comment names the Register call; the lock also covers the registry check just before it.
