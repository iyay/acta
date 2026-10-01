---
id: SPC-0065
created: "2026-10-01 18:56:11"
hash: u6a7iw6
---
Status: approved by the user on 2026-10-01 (Bounded).

# Paint detail pane markdown with the theme colors

Fixes `bugs/2026-10-01-detail-markdown-ignores-theme-colors`.

## Problem

`newRenderer` (`internal/tui/view.go:900`) gets only a dark or light flag and uses glamour's built-in `dark` or `light` style. Those styles have fixed 256-color codes: blue headings, a yellow H1 on a purple bar, red inline code on a gray box. The theme never reaches the markdown, so the detail pane body does not match the rest of the TUI.

## Design

1. `newRenderer` takes the `theme.Theme` and the dark flag. The two callers in `internal/tui/model.go` (`New` and `WithTheme`) pass the theme they already have.
2. A new file `internal/tui/markdown.go` builds an `ansi.StyleConfig` for glamour. It starts from a copy of glamour's dark or light style, so margins, bullets and list layout stay as they are. Only the colors and the markers change.
3. Colors use the same rule as `slot` in `styles.go`: a theme with a background gives hex, the `terminal` theme gives the plain ANSI number so the terminal picks the color.
4. The mapping, in the user's chosen "render" look (markers are hidden, color shows the meaning):
   - body text: theme FG (no color for the `terminal` theme)
   - headings H1 to H6: slot 3 (yellow), bold, no background, no `#` prefix
   - bold: slot 1 (red), bold; italic stays italic in the body color
   - inline code: slot 2 (green), no background, no padding spaces
   - links: slot 4 (blue); block quotes and rules: slot 8 (dim)
   - code blocks: comment slot 8, keyword slot 5, string slot 2, number slot 3, function name slot 4, the rest theme FG
5. The render cache does not change. `WithTheme` already makes a new renderer.

## Testing

Write the failing tests first, in `internal/tui/markdown_test.go`:

- Render `# H`, `**b**` and `` `c` `` with `tokyo-night`. The output holds the green `#9ece6a` as a truecolor sequence, holds no 256-color sequence (`38;5;`), and holds no `**`, no backtick and no `# `.
- Render the same with the `terminal` theme. The output uses 16-color ANSI codes and no hex truecolor sequence.

The test sets the color profile on purpose, because a test run has no real terminal and glamour would drop all color.

## Out of scope

The `source` look (markers kept, as in an editor's source view). Changing the slot roles in `styles.go`.
