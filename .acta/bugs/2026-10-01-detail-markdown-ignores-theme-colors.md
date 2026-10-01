---
id: BUG-0026
hash: fujb7ep
started: "2026-10-01 19:06:02"
---
# Detail pane markdown does not use the theme colors

## Symptom
The markdown body in the detail pane is painted with colors that are not in the
picked theme (tokyo-night by default). Inline code is red on a gray box, headings
are bright blue, and the H1 is yellow text on a purple bar. The rest of the TUI
uses the theme palette, so the body looks like it came from another app. The
user wants it to look like the markdown view of an editor theme: colors taken
from the theme palette, no foreign boxes.

## Root cause
internal/tui/view.go:900 newRenderer takes only a dark/light bool and calls
glamour.WithStandardStyle("dark" or "light") at view.go:912. Those glamour
presets hard-code 256-color codes (glamour v1.0.0 styles/styles.go:157 heading
39, :165-166 H1 fg 228 on bg 63, plus the code span colors). The Theme is never
passed in (internal/tui/model.go:232 and :247 pass only t.Dark(dark)), so no
theme color can reach the markdown.

## Repro
1. go install ./cmd/acta
2. Run acta in this repo with the default theme.
3. Open any spec, for example SPC-0062, and look at the detail pane body.
4. Headings, the H1 bar and inline code show the glamour preset colors, not the
   tokyo-night palette.

## Found in
main at 84a02f1, debug phase 1 from a user screenshot.
