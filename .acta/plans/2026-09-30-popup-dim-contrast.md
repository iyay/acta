---
parent: bugs/2026-09-30-popup-dim-turns-bright-in-ghostty
created: "2026-09-30"
id: PLN-0046
hash: d9ep3ly
---
# Popup Dim Contrast Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** The screen behind a popup stays dim in terminals that enforce a minimum contrast, and the key column of the `?` help popup is drawn in the theme accent. The right wall of every pane shows again (BUG-0015).

**Architecture:** `dim` in `internal/tui/styles.go` becomes plain slot 8, with no faint and no mix toward the background. `dimColor`, `mixHex` and `hexRGB` lose their only use and are removed. The `?` popup builds its content from `helpLines`, with each line's key part painted in `styles.accent` before `boxView` lays it out. `paintFrame` takes the window width and adds `\x1b[K` only to a line narrower than it, so a full line keeps its last cell.

**Also fixes:** BUG-0015 (`bugs/2026-09-30-right-wall-cut-by-erase-line`). A plan has one parent and `closes:` cannot name a bug, so the lander records `fixed_in` on BUG-0015 by hand, next to BUG-0014.

**Tech Stack:** Go, lipgloss (already in go.mod).

**Spec:** `.acta/specs/2026-09-30-popup-dim-contrast-design.md`

**Tests:** fast `scripts/test ./internal/tui/`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- `dim` is slot 8: `t.ANSI[8]` for a hex theme, ANSI `8` for the `terminal` theme. It is never faint, and never mixed toward the background.
- For every built-in theme with its own colors, the WCAG contrast ratio of `dim` against the theme background is at least 1.6.
- `cover`, the popup box, its border, its size and its place stay as they are. The value picker and the new bug prompt do not change.
- `paintFrame` adds `\x1b[K` only to a line whose width is less than the window width. A line as wide as the window never ends in `\x1b[K`.
- No new dependency in go.mod.
- Tests that pin the old dim (faint, the mixed color, `mixHex`) are replaced in Task 1. Never delete a test about anything else, never loosen an assert about anything else, and list every test you changed in the task report.
- Comments use plain English that a 10-year-old can read back: short words, and they say why.
- Before each commit, run `gofmt -l internal` (it must print nothing), `go vet ./internal/tui/` and `scripts/test ./internal/tui/`.

## File Map

- `internal/tui/styles.go`: `dim` becomes `Foreground(slot(slotDim))`; `dimColor`, `mixHex` and `hexRGB` are removed. `fmt` may drop from the imports.
- `internal/tui/styles_test.go`: `TestDimFadesTowardTheBackground` and `TestMixHex` are replaced by the contrast test and the plain-slot test; the `#2d3147` assert in `TestHexThemeRoles` changes to `#414868`.
- `internal/tui/styles.go` (Task 3): `paintFrame(out string, width int)`.
- `internal/tui/styles_test.go` (Task 3): the two old `paintFrame` tests move to the new signature; a new test for short and full lines.
- `internal/tui/view.go`: new `helpText` builds the `?` popup content; `popupBox` calls it (Task 2). Both `paintFrame` calls in `draw` pass `m.width` (Task 3).
- `internal/tui/view_test.go`: the dim brush in `TestPopupDimsTheBackground` (Task 1); the new help key test (Task 2); the `paintFrame` call in `checkPopupDim` and the new right wall test (Task 3).

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2.
- Wave 3: Task 3.

Every task touches `internal/tui/view_test.go`, and Tasks 1 and 3 both touch `internal/tui/styles.go` and `internal/tui/styles_test.go`, so they run one after the other.

---

### Task 1: Dim is plain slot 8

**Files:**
- Modify: `internal/tui/styles.go` (`newStyles`, `dimColor`, `mixHex`, `hexRGB`)
- Test: `internal/tui/styles_test.go`, `internal/tui/view_test.go` (`TestPopupDimsTheBackground`)

**verify:** For every built-in theme, `dim` is slot 8 of that theme and is never faint. In every theme with its own colors, its contrast against the theme background is at least 1.6. The `terminal` theme gives ANSI `8` with no 24-bit code. Every popup (`?`, `t`, `s`, `n`), at both sizes and under both color profiles, still paints every cell behind the box with exactly that brush over the cell's plain text, and `esc` gives back the screen from before. No code path still builds a mixed or faint dim. List every theme and every popup the tests check.

**Interfaces:**
- Consumes: `slot`, `slotDim`, `theme.Names`, `theme.Builtin` (existing).
- Produces: nothing new. `styles.dim` keeps its name and type.

- [ ] **Step 1: Write the failing test**

In `internal/tui/styles_test.go`, replace `TestDimFadesTowardTheBackground` and `TestMixHex` with the tests below. Add `math` and `strconv` to the imports if they are missing.

```go
// contrast is the WCAG contrast ratio of two #rrggbb colors. Terminals like
// Ghostty swap a foreground for white or black when this is under their
// floor, so a dim color that sits too close to the background turns bright.
func contrast(t *testing.T, a, b string) float64 {
	t.Helper()
	lum := func(h string) float64 {
		n, err := strconv.ParseUint(strings.TrimPrefix(h, "#"), 16, 32)
		if err != nil || len(h) != 7 {
			t.Fatalf("not a #rrggbb color: %q", h)
		}
		var out float64
		for i, w := range []float64{0.2126, 0.7152, 0.0722} {
			c := float64(n>>(16-8*i)&0xff) / 255
			if c <= 0.03928 {
				c /= 12.92
			} else {
				c = math.Pow((c+0.055)/1.055, 2.4)
			}
			out += w * c
		}
		return out
	}
	la, lb := lum(a), lum(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// TestDimKeepsContrastOverTheBackground checks every built-in theme with its
// own colors. The dim color must stay far enough from the background that a
// terminal with a minimum contrast of 1.5 leaves it alone.
func TestDimKeepsContrastOverTheBackground(t *testing.T) {
	t.Parallel()

	for _, name := range theme.Names() {
		th, ok := theme.Builtin(name)
		if !ok || th.BG == "" {
			continue
		}
		s := newStyles(th, true)
		got := s.dim.GetForeground()
		if got != lipgloss.Color(th.ANSI[slotDim]) {
			t.Errorf("%s: dim %v, want slot 8 %s", name, got, th.ANSI[slotDim])
			continue
		}
		if r := contrast(t, string(got.(lipgloss.Color)), th.BG); r < 1.6 {
			t.Errorf("%s: dim %v on %s has contrast %.2f, want at least 1.6", name, got, th.BG, r)
		}
	}
}

func TestDimIsNeverFaint(t *testing.T) {
	t.Parallel()

	for _, name := range theme.Names() {
		th, _ := theme.Builtin(name)
		if newStyles(th, true).dim.GetFaint() {
			t.Errorf("%s: dim is faint, which lowers its contrast again", name)
		}
	}
	term, _ := theme.Builtin("terminal")
	if got := newStyles(term, true).dim.GetForeground(); got != lipgloss.Color("8") {
		t.Errorf("terminal theme: dim %v, want slot 8", got)
	}
}
```

In `TestHexThemeRoles`, change the dim assert from `#2d3147` to `#414868` (tokyo-night slot 8).

In `internal/tui/view_test.go`, `TestPopupDimsTheBackground`, change the brush and its comment:

```go
	// The brush the view paints the screen behind a popup with, spelled out
	// here so this test checks the color the plan names and not the one the
	// view happens to use today: tokyo-night slot 8, never faint.
	dim := func() lipgloss.Style { return lipgloss.NewStyle().Foreground(lipgloss.Color("#414868")) }
```

- [ ] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/tui/ -run 'TestDimKeepsContrastOverTheBackground|TestDimIsNeverFaint|TestHexThemeRoles|TestPopupDimsTheBackground'`
Expected: FAIL. On today's code, tokyo-night dim is `#2d3147`, not slot 8, and dim is faint.

- [ ] **Step 3: Write minimal implementation**

In `newStyles` (`internal/tui/styles.go`), the `dim` entry and its comment become:

```go
		// dim paints the screen behind a popup, so the box on top is the only
		// thing left with a color of its own. It is slot 8 as it is, with no
		// faint: a color closer to the background trips the minimum contrast
		// some terminals keep, and they then draw it bright.
		dim: lipgloss.NewStyle().Foreground(slot(slotDim)),
```

Delete `dimColor`, `mixHex` and `hexRGB` together with their comments. Remove `fmt` from the imports if nothing else in the file uses it. Before you delete them, run `grep -rn "mixHex\|hexRGB\|dimColor" internal/` and check that no other file uses them.

- [ ] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/tui/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal && go vet ./internal/tui/ && scripts/test ./internal/tui/
git add internal/tui/styles.go internal/tui/styles_test.go internal/tui/view_test.go
git commit -m "tui: dim the screen behind a popup with plain slot 8"
```

---

### Task 2: Help keys in the accent

**Files:**
- Modify: `internal/tui/view.go` (`popupBox`, new `helpText`)
- Test: `internal/tui/view_test.go`

**verify:** On every line of the `?` popup, the key part (the text before the first run of two or more spaces) is drawn in the accent, and the description is not. This holds for every line of `helpLines`, in the hex and the `terminal` theme. The plain text of the box and its width are exactly as before. The value picker and the new bug prompt draw no accent inside the box. List every line checked.

**Interfaces:**
- Consumes: `helpLines`, `styles.accent`, `boxView` (existing); `withTrueColor`, `plain`, `newModel`, `sized`, `press` (existing test helpers).
- Produces: `func (m Model) helpText() string`.

- [ ] **Step 1: Write the failing test**

Add to `internal/tui/view_test.go`:

```go
// TestHelpKeysWearTheAccent reads every line of the ? popup. The keys must be
// in the accent so they stand out from what they do, and nothing else on the
// screen may move.
func TestHelpKeysWearTheAccent(t *testing.T) {
	withTrueColor(func() {
		for _, name := range []string{"tokyo-night", "terminal"} {
			m := sized(newModel(t).WithTheme(name, true), 120, 40)
			text := m.helpText()
			if plain(text) != helpLines {
				t.Fatalf("%s: help text changed its words:\n%q", name, plain(text))
			}
			got := strings.Split(text, "\n")
			for i, ln := range strings.Split(helpLines, "\n") {
				key := ln
				if at := strings.Index(ln, "  "); at > 0 {
					key = ln[:at]
				}
				want := m.styles.accent.Render(key)
				if !strings.HasPrefix(got[i], want) {
					t.Errorf("%s line %d: key %q is not in the accent: %q", name, i, key, got[i])
				}
				if rest := strings.TrimPrefix(got[i], want); strings.Contains(rest, "\x1b[") {
					t.Errorf("%s line %d: the description has a color: %q", name, i, rest)
				}
			}
			before := strings.Split(m.boxView("Keys", helpLines), "\n")
			after := strings.Split(press(m, "?").popupBox(), "\n")
			if len(before) != len(after) {
				t.Fatalf("%s: box has %d rows, want %d", name, len(after), len(before))
			}
			for i := range before {
				if plain(after[i]) != plain(before[i]) {
					t.Errorf("%s row %d: box text moved:\n got %q\nwant %q", name, i, plain(after[i]), plain(before[i]))
				}
			}
		}
	})
}

func TestOtherPopupsHaveNoAccentInside(t *testing.T) {
	withTrueColor(func() {
		for _, open := range []string{"t", "n"} {
			m := press(sized(newModel(t), 120, 40), tabKey(tabBugs), open)
			rows := strings.Split(m.popupBox(), "\n")
			accent := strings.TrimSuffix(m.styles.accent.Render("x"), "x\x1b[0m")
			for i, r := range rows[1 : len(rows)-1] {
				inner := strings.TrimSuffix(strings.TrimPrefix(r, m.styles.accent.Render("│")), m.styles.accent.Render("│"))
				if strings.Contains(inner, accent) {
					t.Errorf("popup %q row %d: accent inside the box: %q", open, i+1, inner)
				}
			}
		}
	})
}
```

If `newModel(t)` opens on a tab where `t` opens no popup, pick the tab the way `checkPopupDim` does (`tabKey(tabBugs)`). The `s` popup is left out because it picks from the same list shape as `t`.

- [ ] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/tui/ -run 'TestHelpKeysWearTheAccent|TestOtherPopupsHaveNoAccentInside'`
Expected: FAIL to build with `m.helpText undefined`.

- [ ] **Step 3: Write minimal implementation**

In `internal/tui/view.go`, add below `helpLines`:

```go
// helpText is the key map with each key in the accent, so the eye finds the
// key first and reads what it does after. A line is split at its first run of
// two spaces, the gap the key column is laid out with.
func (m Model) helpText() string {
	lines := strings.Split(helpLines, "\n")
	for i, ln := range lines {
		if at := strings.Index(ln, "  "); at > 0 {
			lines[i] = m.styles.accent.Render(ln[:at]) + ln[at:]
		}
	}
	return strings.Join(lines, "\n")
}
```

In `popupBox`, change the help case to:

```go
	case m.help:
		return m.boxView("Keys", m.helpText())
```

- [ ] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/tui/`
Expected: PASS, including `TestPopupDimsTheBackground` and the help tests in `model_test.go` and `frame_test.go`, which read `helpLines` itself.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal && go vet ./internal/tui/ && scripts/test ./internal/tui/
git add internal/tui/view.go internal/tui/view_test.go
git commit -m "tui: draw the help keys in the theme accent"
```

---

### Task 3: Full lines keep their last cell

**Files:**
- Modify: `internal/tui/styles.go` (`paintFrame`)
- Modify: `internal/tui/view.go` (`draw`, both `paintFrame` calls)
- Test: `internal/tui/styles_test.go`, `internal/tui/view_test.go`

**verify:** No drawn line as wide as the window ever ends in `\x1b[K`, on every tab, at every size tested, with a popup open and without one. Every such line ends in the character the view drew there, never a blank. A line narrower than the window still ends in `\x1b[K`, so the theme background reaches the right edge. The `terminal` theme, which has no background, is still left untouched. List every tab, size and popup state checked, and every caller of `paintFrame`.

**Interfaces:**
- Consumes: `styles.bg`, `styles.fg`, `withTrueColor`, `plain`, `newModel`, `sized`, `press`, `tabKey`, `topTabs` (existing).
- Produces: `func (s styles) paintFrame(out string, width int) string`.

- [ ] **Step 1: Write the failing test**

Add to `internal/tui/styles_test.go`:

```go
// TestPaintFrameErasesOnlyShortLines guards the right wall. Erase-line right
// after a full line clears its last cell, because the cursor is still sitting
// in the last column. So only a short line may get it.
func TestPaintFrameErasesOnlyShortLines(t *testing.T) {
	th, _ := theme.Builtin("tokyo-night")
	s := newStyles(th, true)
	withTrueColor(func() {
		out := strings.Split(s.paintFrame("abcde\nab\n"+s.accent.Render("abcd")+"┐", 5), "\n")
		if strings.HasSuffix(out[0], "\x1b[K") {
			t.Errorf("full line got an erase: %q", out[0])
		}
		if !strings.HasSuffix(out[1], "\x1b[K") {
			t.Errorf("short line lost its erase, so the background stops short: %q", out[1])
		}
		if strings.HasSuffix(out[2], "\x1b[K") {
			t.Errorf("full line with colors and a wide rune edge got an erase: %q", out[2])
		}
	})
}
```

Change the two old calls in `styles_test.go` to the new signature, with the same asserts: `s.paintFrame("a\nb", 1)` in `TestTerminalThemeUsesANSISlots`, and `s.paintFrame(s.accent.Render("a")+"b\n"+"c", 10)` in `TestPaintFrameKeepsBackgroundAfterResets`. Both lines there are shorter than 10, so they keep their erase, and that test does not change otherwise.

Add to `internal/tui/view_test.go`:

```go
// TestFullLinesKeepTheirLastCell draws every tab at several sizes, with and
// without a popup. A line as wide as the window must not end in erase-line,
// or the terminal blanks the right wall.
func TestFullLinesKeepTheirLastCell(t *testing.T) {
	withTrueColor(func() {
		for _, size := range [][2]int{{80, 30}, {120, 40}, {204, 58}} {
			for tab := range topTabs {
				for _, open := range []string{"", "?"} {
					m := press(sized(newModel(t), size[0], size[1]), tabKey(tab))
					if open != "" {
						m = press(m, open)
					}
					for y, ln := range strings.Split(m.View(), "\n") {
						txt := plain(ln)
						if lipgloss.Width(txt) != m.width {
							continue
						}
						if strings.HasSuffix(ln, "\x1b[K") {
							t.Errorf("%dx%d tab %d popup %q line %d: full line ends in erase-line", size[0], size[1], tab, open, y)
						}
						if strings.HasSuffix(txt, " ") && y < size[1]-1 {
							t.Errorf("%dx%d tab %d popup %q line %d: full line ends in a blank, not the wall: %q", size[0], size[1], tab, open, y, txt)
						}
					}
				}
			}
		}
	})
}
```

The last line of the screen is the status line; it may end in a blank when its right side is short, so the blank check skips it. In `checkPopupDim`, change `pop.styles.paintFrame("")` to `pop.styles.paintFrame("", pop.width)`. An empty line is short, so it still ends in `\x1b[K` and the `TrimSuffix` there works as before.

If a pane line in some tab ends in a blank on purpose (a pane with no right wall at that size), look at what the view draws there before you change the assert, and report it.

- [ ] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/tui/ -run 'TestPaintFrameErasesOnlyShortLines|TestFullLinesKeepTheirLastCell|TestPaintFrameKeepsBackgroundAfterResets|TestTerminalThemeUsesANSISlots|TestPopupDimsTheBackground'`
Expected: FAIL to build with `too many arguments in call to s.paintFrame`.

- [ ] **Step 3: Write minimal implementation**

In `internal/tui/styles.go`:

```go
// paintFrame lays the theme background under the whole screen. Every style
// ends with a reset that drops the background, so the theme colors go back
// on right after each reset and at the start of each line. A short line gets
// an erase to the end, so the background reaches the right edge. A full line
// gets none: the cursor still sits in its last column, and an erase there
// would clear the last cell, which is the right wall.
func (s styles) paintFrame(out string, width int) string {
	if s.bg == "" {
		return out
	}
	p := termenv.TrueColor
	set := termenv.CSI + p.Color(s.bg).Sequence(true) + "m" + termenv.CSI + p.Color(s.fg).Sequence(false) + "m"
	lines := strings.Split(out, "\n")
	for i, ln := range lines {
		painted := set + strings.ReplaceAll(ln, "\x1b[0m", "\x1b[0m"+set)
		if lipgloss.Width(ln) < width {
			painted += "\x1b[K"
		}
		lines[i] = painted
	}
	return strings.Join(lines, "\n")
}
```

In `draw` (`internal/tui/view.go`), both calls pass the width:

```go
		return m.styles.paintFrame(m.cover(body+"\n"+line), m.width)
	}
	return m.styles.paintFrame(body+"\n"+line, m.width)
```

Before you finish, run `grep -rn "paintFrame(" internal/` and check that every caller passes a width.

- [ ] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/tui/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal && go vet ./internal/tui/ && scripts/test ./internal/tui/
git add internal/tui/styles.go internal/tui/view.go internal/tui/styles_test.go internal/tui/view_test.go
git commit -m "tui: erase only short lines so the right wall stays"
```
