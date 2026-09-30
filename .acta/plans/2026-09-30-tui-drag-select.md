---
created: "2026-09-30"
parent: specs/2026-09-30-tui-drag-select-design
id: PLN-0050
hash: k0qzc6d
---
# TUI Drag Select and Pink Scratch Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** A left-button drag inside any pane selects text like a terminal and copies it on release, with the same toast as `y`, and scratch ids and the Scratches tab turn pink.

**Architecture:** A new `drag` value in `internal/tui/select.go` holds the anchor and end cell in screen cells, clamped to the text area of the pane the press landed in. On release, the screen is drawn again with no highlight and the selected cells are cut out of it with `xansi.Cut` and `xansi.Strip`. `draw()` lays the highlight over the frame the same way. Scratch takes a fixed pink per theme kind instead of slot 2.

**Tech Stack:** Go, Bubble Tea v1.3.10, lipgloss, `github.com/charmbracelet/x/ansi` v0.10.2 (already imported as `xansi` in `internal/tui`).

**Spec:** `.acta/specs/2026-09-30-tui-drag-select-design.md`

**Tests:** fast `scripts/test ./internal/tui/`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Import `github.com/charmbracelet/x/ansi` as `xansi`, the way `view.go` and `frame.go` do. The test files of this package declare a package var named `ansi`, so a bare import does not compile.
- The selection is a stream: first row from the start cell to the right edge of the text area, middle rows whole, last row from the left edge to the end cell, both ends included. Anchor and end are swapped when the end comes first in row order.
- The text area of a box is `x+1 .. x+w-2` by `y+1 .. y+h-2`: never a wall, the title line, the bottom line or the scrollbar wall.
- No auto-scroll. The end cell is clamped to the text area of the anchor pane.
- The copied text is what the screen shows, colors stripped, trailing spaces trimmed per row, rows joined with `\n`.
- The toast is the `y` toast: `m.status` plus `clearStatusAfter(toastFor, m.status)`. Good copy: `copied <text>`, newlines shown as spaces, cut with `…` to fit. Failed copy: `copy failed: <error>`, no clear command. Text that is only spaces copies nothing and shows nothing.
- The highlight clears on the next left press, any key, a wheel turn and a resize. It is off while a popup, the help, search or the slug input is open.
- Scratch pink: `#ff79c6` for a dark hex theme, `#c2185b` for a light hex theme, `212` for the `terminal` theme. Green stays for done, live and the pulse.
- A click with no drag keeps today's behaviour and copies nothing. A mouse release with no selection must still reuse the frame (`m.same`), as `TestViewReusesTheFrameForIgnoredMouseAndEmptyTick` checks today.
- Old tests that pin scratch as slot 2 are updated in Task 1. Never delete a test about anything else, never loosen an assert about anything else, and list every test you changed in the task report.
- Tests replace `m.clip` with a func that records the text. Tests never touch the real clipboard and never sleep.
- Comments use plain English that a 10-year-old can read back: short words, and they say why.
- Before each commit, run `gofmt -l internal` (it must print nothing), `go vet ./internal/tui/` and `scripts/test ./internal/tui/`.

## File Map

- `internal/tui/styles.go`: Task 1 (`scratchColor`, scratch leaves `kindSlots`).
- `internal/tui/styles_test.go`: Task 1.
- `internal/tui/select.go` (new): Task 2 (`drag`, `rect`, `span`, `box.textArea`, `dragText`, `paintDrag`, `Model.anchorAt`, `Model.copyDrag`).
- `internal/tui/select_test.go` (new): Task 2.
- `internal/tui/model.go`: Task 2 (the `drag` field only), Task 3 (mouse, key and resize wiring).
- `go.mod`: Task 2 (`go mod tidy` drops the `// indirect` on `x/ansi`).
- `internal/tui/view.go`: Task 3 (`draw()` paints the highlight).
- `internal/tui/drag_test.go` (new): Task 3.

## Waves

- Wave 1: Task 1, Task 2 (no shared file).
- Wave 2: Task 3 (needs Task 2's `drag`, `anchorAt`, `copyDrag`, `paintDrag`).

---

### Task 1: Scratch is pink

**Files:**
- Modify: `internal/tui/styles.go` (`kindSlots`, `newStyles`)
- Test: `internal/tui/styles_test.go` (`TestKindAndRoleSlots`, new `TestScratchIsPink`)

**verify:** In every built-in theme and the `terminal` theme, the scratch id brush and the Scratches tab color are the same pink the spec names for that theme kind, and no other kind, role brush or pulse frame changed color. List every built-in theme checked.

**Interfaces:**
- Consumes: nothing new.
- Produces: `func scratchColor(t theme.Theme, dark bool) lipgloss.Color`.

- [ ] **Step 1: Write the failing test**

In `styles_test.go`, drop `board.KindScratch: 2,` from the slot map in `TestKindAndRoleSlots` (scratch has no slot now; the rest of that test stays). Add:

```go
func TestScratchIsPink(t *testing.T) {
	t.Parallel()

	for _, name := range theme.Names() {
		th, _ := theme.Builtin(name)
		for _, dark := range []bool{true, false} {
			want := lipgloss.Color("#c2185b")
			switch {
			case th.BG == "":
				want = "212"
			case th.Dark(dark):
				want = "#ff79c6"
			}
			s := newStyles(th, dark)
			if got := s.kind(board.KindScratch).GetForeground(); got != want {
				t.Errorf("%s dark=%v: scratch id = %v, want %v", name, dark, got, want)
			}
			if got := s.tabColor(board.KindScratch); got != want {
				t.Errorf("%s dark=%v: Scratches tab = %v, want %v", name, dark, got, want)
			}
			if got := s.done.GetForeground(); got == want {
				t.Errorf("%s dark=%v: done took the scratch pink", name, dark)
			}
		}
	}
}
```

`theme.Names()` lists every built-in theme, `terminal` included.

- [ ] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/tui/ -run 'TestScratchIsPink|TestKindAndRoleSlots'`
Expected: FAIL, `scratch id = ... want #ff79c6` (it is slot 2 green now).

- [ ] **Step 3: Write minimal implementation**

In `styles.go`, remove `board.KindScratch: slotGreen,` from `kindSlots`. In `newStyles`, after the loop that fills `kinds`:

```go
	kinds[board.KindScratch] = scratchColor(t, dark)
```

Add below `newStyles`:

```go
// scratchColor is the pink of a scratch id and the Scratches tab. Green means
// done, and every hue of the 16 slots is taken, so scratch gets a pink of its
// own. A light background needs a deeper pink to stay readable. The terminal
// theme has no hex, so it asks for pink from the 256 colors.
func scratchColor(t theme.Theme, dark bool) lipgloss.Color {
	switch {
	case t.BG == "":
		return "212"
	case t.Dark(dark):
		return "#ff79c6"
	}
	return "#c2185b"
}
```

Update the `kindSlots` comment only if it now says something false.

- [ ] **Step 4: Run tests to verify they pass**

Run: `scripts/test ./internal/tui/`
Expected: PASS. Any other test that pinned scratch as green gets its expected color changed to the pink, and is listed in the report.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal && go vet ./internal/tui/
git add internal/tui/styles.go internal/tui/styles_test.go
git commit -m "Paint scratch pink so it no longer looks like done"
```

---

### Task 2: Selection core

**Files:**
- Create: `internal/tui/select.go`
- Create: `internal/tui/select_test.go`
- Modify: `internal/tui/model.go` (one field in `Model`)
- Modify: `go.mod`

**verify:** For every anchor and end pair inside one text area, `dragText` returns exactly the cells of the stream the Global Constraints define, in row order, with no wall, title, bottom line or scrollbar cell, no color codes and no trailing spaces; the same pair in the other order gives the same text; and `paintDrag` changes no cell outside the spans. List the cases checked: one row, many rows, reversed, one cell, a wide character on the edge.

**Interfaces:**
- Consumes: `box` (`x, y, w, h`), `Model.geometry()`, `geom.at`, `geom.full`, `geom.wide`, `Model.hit`, `Model.draw`, `Model.clip`, `clearStatusAfter`, `toastFor`, `clamp`.
- Produces:
  - `type drag struct { on, held bool; area rect; ax, ay, ex, ey int }`
  - `type rect struct{ x0, y0, x1, y1 int }` with `func (r rect) has(x, y int) bool`
  - `type span struct{ y, x0, x1 int }`
  - `func (b box) textArea() (rect, bool)`
  - `func (d drag) shown() bool`
  - `func (d drag) to(x, y int) drag`
  - `func (d drag) spans() []span`
  - `func dragText(frame string, d drag) string`
  - `func paintDrag(frame string, d drag, brush lipgloss.Style) string`
  - `func (m Model) anchorAt(x, y int) drag`
  - `func (m *Model) copyPicked(text string) tea.Cmd`
  - `func (m *Model) copyDrag() tea.Cmd`
  - Model field `drag drag`

- [ ] **Step 1: Write the failing tests**

`internal/tui/select_test.go`:

```go
package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// area5 is a text area of five cells by three rows at the top left, inside a
// one-cell wall, so the frame below has a wall on every side of it.
var area5 = rect{x0: 1, y0: 1, x1: 5, y1: 3}

const frame5 = "┌─────┐\n" +
	"│abcde│\n" +
	"│fg  h│\n" +
	"│ijklm│\n" +
	"└─────┘"

func sel(ax, ay, ex, ey int) drag {
	return drag{on: true, area: area5, ax: ax, ay: ay, ex: ex, ey: ey}
}

func TestDragTextIsAStream(t *testing.T) {
	t.Parallel()

	for name, c := range map[string]struct {
		d    drag
		want string
	}{
		"one row":        {sel(2, 1, 4, 1), "bcd"},
		"one cell apart": {sel(1, 1, 2, 1), "ab"},
		"many rows":      {sel(3, 1, 2, 3), "cde\nfg  h\nij"},
		"reversed":       {sel(2, 3, 3, 1), "cde\nfg  h\nij"},
		"trailing space": {sel(1, 2, 4, 2), "fg"},
		"whole area":     {sel(1, 1, 5, 3), "abcde\nfg  h\nijklm"},
	} {
		if got := dragText(frame5, c.d); got != c.want {
			t.Errorf("%s: got %q, want %q", name, got, c.want)
		}
	}
}

func TestDragWithNoMoveSelectsNothing(t *testing.T) {
	t.Parallel()

	if d := sel(2, 1, 2, 1); d.shown() || dragText(frame5, d) != "" {
		t.Errorf("a press with no drag selected %q", dragText(frame5, d))
	}
}

func TestDragToStaysInsideTheArea(t *testing.T) {
	t.Parallel()

	d := sel(2, 2, 2, 2).to(40, -3)
	if d.ex != area5.x1 || d.ey != area5.y0 {
		t.Fatalf("end = (%d,%d), want (%d,%d)", d.ex, d.ey, area5.x1, area5.y0)
	}
	if got := dragText(frame5, d); strings.ContainsAny(got, "│─┌┐└┘") {
		t.Errorf("a drag past the edge copied a wall: %q", got)
	}
}

func TestDragTextKeepsWideCharactersWhole(t *testing.T) {
	t.Parallel()

	// 日 and 本 take two cells each: cells 1-2 and 3-4.
	frame := "│日本x│"
	area := rect{x0: 1, y0: 0, x1: 5, y1: 0}
	for x0 := 1; x0 <= 5; x0++ {
		for x1 := x0 + 1; x1 <= 5; x1++ {
			got := dragText(frame, drag{on: true, area: area, ax: x0, ay: 0, ex: x1, ey: 0})
			if !strings.Contains("日本x", got) {
				t.Errorf("cells %d-%d gave %q, not a whole piece of the row", x0, x1, got)
			}
		}
	}
}

func TestPaintDragOnlyTouchesTheSpans(t *testing.T) {
	t.Parallel()

	withColor(func() {
		brush := lipgloss.NewStyle().Reverse(true)
		got := paintDrag(frame5, sel(3, 1, 2, 3), brush)
		if plain(got) != frame5 {
			t.Fatalf("the words changed under the band:\n%s", plain(got))
		}
		lines := strings.Split(got, "\n")
		for _, i := range []int{0, 4} {
			if lines[i] != strings.Split(frame5, "\n")[i] {
				t.Errorf("line %d is outside the selection but changed: %q", i, lines[i])
			}
		}
		if !strings.Contains(lines[1], brush.Render("cde")) {
			t.Errorf("row 1 does not carry the band over cde: %q", lines[1])
		}
	})
}

func TestTextAreaLeavesOutWallsAndTitle(t *testing.T) {
	t.Parallel()

	a, ok := box{x: 10, y: 4, w: 6, h: 5}.textArea()
	if !ok || a != (rect{x0: 11, y0: 5, x1: 14, y1: 7}) {
		t.Fatalf("text area = %+v %v", a, ok)
	}
	if _, ok := (box{x: 0, y: 0, w: 2, h: 2}).textArea(); ok {
		t.Error("a box with no room for words has a text area")
	}
}

func TestAnchorOnlyInATextArea(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 120, 40)
	b := m.geometry().at(paneList)
	if d := m.anchorAt(b.x+1, b.y+1); !d.on || !d.held {
		t.Error("a press on the first text cell set no anchor")
	}
	for name, c := range map[string][2]int{
		"left wall":   {b.x, b.y + 1},
		"right wall":  {b.x + b.w - 1, b.y + 1},
		"title line":  {b.x + 1, b.y},
		"bottom line": {b.x + 1, b.y + b.h - 1},
		"status line": {1, m.height - 1},
	} {
		if d := m.anchorAt(c[0], c[1]); d.on {
			t.Errorf("a press on the %s set an anchor", name)
		}
	}
}

func TestCopyPickedToast(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 120, 40)
	var got string
	m.clip = func(s string) error { got = s; return nil }
	cmd := m.copyPicked("abc\ndef")
	if got != "abc\ndef" {
		t.Fatalf("clipboard = %q", got)
	}
	if m.status != "copied abc def" || cmd == nil {
		t.Errorf("status = %q, cmd nil = %v", m.status, cmd == nil)
	}
	// The toast clears the same way the y toast does.
	next, _ := m.Update(clearStatusMsg{text: m.status})
	if s := next.(Model).status; s != "" {
		t.Errorf("the toast stayed: %q", s)
	}

	m.width = 20
	m.copyPicked(strings.Repeat("x", 50))
	if !strings.HasSuffix(m.status, "…") || lipgloss.Width(m.status) > m.width {
		t.Errorf("a long copy was not cut to fit: %q", m.status)
	}

	m.clip = func(string) error { return errors.New("no clipboard") }
	if cmd := m.copyPicked("abc"); cmd != nil || m.status != "copy failed: no clipboard" {
		t.Errorf("failed copy: status %q, cmd nil = %v", m.status, cmd == nil)
	}
}

func TestCopyPickedOfSpacesCopiesNothing(t *testing.T) {
	t.Parallel()

	m := newModel(t)
	called := false
	m.clip = func(string) error { called = true; return nil }
	m.status = "before"
	for _, text := range []string{"", "   ", " \n  "} {
		if cmd := m.copyPicked(text); cmd != nil || called || m.status != "before" {
			t.Errorf("%q: copied %v, status %q, cmd nil = %v", text, called, m.status, cmd == nil)
		}
	}
}

func TestCopyDragCopiesTheScreenWithNoBand(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 120, 40)
	a, _ := m.geometry().at(paneList).textArea()
	var got string
	m.clip = func(s string) error { got = s; return nil }
	m.drag = drag{on: true, area: a, ax: a.x0, ay: a.y0, ex: a.x1, ey: a.y0 + 1}
	bare := m
	bare.drag = drag{}
	want := dragText(bare.draw(), m.drag)
	m.copyDrag()
	if got == "" || got != want {
		t.Errorf("clipboard = %q, want %q", got, want)
	}
}
```

The `withColor` and `plain` helpers already live in `view_test.go`.

- [ ] **Step 2: Run tests to verify they fail**

Run: `scripts/test ./internal/tui/ -run 'Drag|TextArea|Anchor|CopyPicked'`
Expected: FAIL to compile: `undefined: drag`, `undefined: rect`, `undefined: dragText`.

- [ ] **Step 3: Write minimal implementation**

Add to `Model` in `model.go`, next to the wheel fields:

```go
	drag       drag  // the text a mouse drag selected; empty when none
```

`internal/tui/select.go`:

```go
package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

// drag is text picked with the mouse. It lives in screen cells, so it only
// means something until the screen moves.
type drag struct {
	on     bool // from a press on words until the highlight clears
	held   bool // the button is still down, so the end follows the pointer
	area   rect // the words of the pane the press landed in
	ax, ay int  // the cell the press landed on
	ex, ey int  // the cell the pointer is on now
}

// rect is a block of screen cells, both ends included.
type rect struct{ x0, y0, x1, y1 int }

func (r rect) has(x, y int) bool { return x >= r.x0 && x <= r.x1 && y >= r.y0 && y <= r.y1 }

// span is the cells of one screen row a drag covers, both ends included.
type span struct{ y, x0, x1 int }

// textArea is where a box keeps its words: inside the walls, below the title
// line and above the bottom line. The scrollbar is the right wall, so it is
// left out too. A box with no room for words gives false.
func (b box) textArea() (rect, bool) {
	r := rect{b.x + 1, b.y + 1, b.x + b.w - 2, b.y + b.h - 2}
	return r, r.x1 >= r.x0 && r.y1 >= r.y0
}

// shown says there is something to paint or copy: the end left the anchor.
func (d drag) shown() bool { return d.on && (d.ax != d.ex || d.ay != d.ey) }

// to moves the end to the pointer. It stays inside the pane the press was
// in, so a drag never picks up another pane or a wall.
func (d drag) to(x, y int) drag {
	d.ex = clamp(x, d.area.x0, d.area.x1)
	d.ey = clamp(y, d.area.y0, d.area.y1)
	return d
}

// spans gives the rows a drag covers, top to bottom, the way a terminal
// picks text: the first row from the start on, the middle rows whole, and
// the last row up to the end. A drag up or left works the same as down.
func (d drag) spans() []span {
	if !d.shown() {
		return nil
	}
	sx, sy, ex, ey := d.ax, d.ay, d.ex, d.ey
	if ey < sy || (ey == sy && ex < sx) {
		sx, sy, ex, ey = ex, ey, sx, sy
	}
	out := make([]span, 0, ey-sy+1)
	for y := sy; y <= ey; y++ {
		x0, x1 := d.area.x0, d.area.x1
		if y == sy {
			x0 = sx
		}
		if y == ey {
			x1 = ex
		}
		out = append(out, span{y, x0, x1})
	}
	return out
}

// dragText is the picked text the way the screen shows it: no colors, and
// no spaces at the end of a row.
func dragText(frame string, d drag) string {
	lines := strings.Split(frame, "\n")
	var rows []string
	for _, s := range d.spans() {
		if s.y >= len(lines) {
			break
		}
		cut := xansi.Strip(xansi.Cut(lines[s.y], s.x0, s.x1+1))
		rows = append(rows, strings.TrimRight(cut, " "))
	}
	return strings.Join(rows, "\n")
}

// paintDrag lays the band over the picked cells. The words under it drop
// their own colors, so the band reads the same wherever it falls.
func paintDrag(frame string, d drag, brush lipgloss.Style) string {
	sp := d.spans()
	if len(sp) == 0 {
		return frame
	}
	lines := strings.Split(frame, "\n")
	for _, s := range sp {
		if s.y >= len(lines) {
			break
		}
		ln := lines[s.y]
		mid := xansi.Strip(xansi.Cut(ln, s.x0, s.x1+1))
		lines[s.y] = xansi.Cut(ln, 0, s.x0) + brush.Render(mid) + xansi.Cut(ln, s.x1+1, xansi.StringWidth(ln))
	}
	return strings.Join(lines, "\n")
}

// anchorAt starts a drag on the cell of a press. A press on a wall, a title
// line, a bottom line or the status line holds no words, so it starts none.
func (m Model) anchorAt(x, y int) drag {
	g := m.geometry()
	p, _, _ := m.hit(x, y)
	b := g.full
	if g.wide {
		b = g.at(p)
	}
	a, ok := b.textArea()
	if !ok || !a.has(x, y) {
		return drag{}
	}
	return drag{on: true, held: true, area: a, ax: x, ay: y, ex: x, ey: y}
}

// copyDrag copies what the drag picked. The text comes from a screen drawn
// with no band, so the band never ends up in it.
func (m *Model) copyDrag() tea.Cmd {
	bare := *m
	bare.drag = drag{}
	return m.copyPicked(dragText(bare.draw(), m.drag))
}

// copyPicked puts text on the clipboard and says so the way y does. Text that
// is only spaces copies nothing and says nothing.
func (m *Model) copyPicked(text string) tea.Cmd {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	if err := m.clip(text); err != nil {
		m.status = "copy failed: " + err.Error()
		return nil
	}
	const lead = "copied "
	m.status = lead + xansi.Truncate(strings.ReplaceAll(text, "\n", " "), max(1, m.width-len(lead)), "…")
	return clearStatusAfter(toastFor, m.status)
}
```

Then run `go mod tidy`. It drops `// indirect` from `github.com/charmbracelet/x/ansi` in `go.mod`, since `internal/tui` imports it straight. Check `git diff go.mod go.sum` shows only that change.

- [ ] **Step 4: Run tests to verify they pass**

Run: `scripts/test ./internal/tui/ -run 'Drag|TextArea|Anchor|CopyPicked'`
Expected: PASS.

Then: `scripts/test ./internal/tui/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal && go vet ./internal/tui/
git add internal/tui/select.go internal/tui/select_test.go internal/tui/model.go go.mod go.sum
git commit -m "Add the text selection a mouse drag makes, and its copy"
```

---

### Task 3: Wire the drag into the mouse, keys and screen

**Files:**
- Modify: `internal/tui/model.go` (`Update` resize case, `key`, `mouse`)
- Modify: `internal/tui/view.go` (`draw`)
- Test: `internal/tui/drag_test.go` (new)

**verify:** Through `Update` alone, every way the highlight can end (left press, any key, wheel turn, resize) leaves no band on screen, no path with a popup, help, search or slug input open ever starts a drag, a click with no drag keeps today's cursor move and copies nothing, and a drag in a sidebar pane and in the detail pane copies the screen text of that pane only. List every path checked.

**Interfaces:**
- Consumes (Task 2): `drag`, `drag.held`, `drag.on`, `drag.shown()`, `drag.to(x, y int) drag`, `drag.spans()`, `Model.anchorAt(x, y int) drag`, `Model.copyDrag() tea.Cmd`, `paintDrag(frame string, d drag, brush lipgloss.Style) string`, `box.textArea() (rect, bool)`, Model field `drag`.
- Produces: nothing new for later tasks.

- [ ] **Step 1: Write the failing tests**

`internal/tui/drag_test.go`:

```go
package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func mouseAt(m Model, x, y int, a tea.MouseAction) Model {
	next, _ := m.Update(tea.MouseMsg{X: x, Y: y, Action: a, Button: tea.MouseButtonLeft})
	return next.(Model)
}

// dragged presses at one cell, moves to another and lets go. It gives back
// the model and what reached the clipboard.
func dragged(m Model, x0, y0, x1, y1 int) (Model, string) {
	var got string
	m.clip = func(s string) error { got = s; return nil }
	m = mouseAt(m, x0, y0, tea.MouseActionPress)
	m = mouseAt(m, x1, y1, tea.MouseActionMotion)
	m = mouseAt(m, x1, y1, tea.MouseActionRelease)
	return m, got
}

// screenText cuts the cells a drag should copy straight out of the plain
// screen, rune by rune, so the test does not lean on the code it checks.
func screenText(m Model, x0, y0, x1, y1 int, a rect) string {
	lines := strings.Split(plain(m.View()), "\n")
	var rows []string
	for y := y0; y <= y1; y++ {
		r := []rune(lines[y])
		from, to := a.x0, a.x1
		if y == y0 {
			from = x0
		}
		if y == y1 {
			to = x1
		}
		rows = append(rows, strings.TrimRight(string(r[from:to+1]), " "))
	}
	return strings.Join(rows, "\n")
}

func TestDragCopiesThePaneText(t *testing.T) {
	t.Parallel()

	for _, p := range []pane{paneList, paneDetail} {
		m := paneModel(t, p)
		a, _ := scrollBox(m, p).textArea()
		want := screenText(m, a.x0+1, a.y0, a.x0+6, a.y0+1, a)
		m, got := dragged(m, a.x0+1, a.y0, a.x0+6, a.y0+1)
		if got == "" || got != want {
			t.Errorf("pane %d: copied %q, want %q", p, got, want)
		}
		if strings.ContainsAny(got, "│─╭╮╰╯┌┐└┘") {
			t.Errorf("pane %d: copy holds a wall: %q", p, got)
		}
		if !strings.HasPrefix(m.status, "copied ") {
			t.Errorf("pane %d: status = %q", p, m.status)
		}
	}
}

func TestDragBackwardsCopiesTheSame(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	a, _ := scrollBox(m, paneDetail).textArea()
	_, down := dragged(m, a.x0+2, a.y0, a.x0+5, a.y0+2)
	_, up := dragged(m, a.x0+5, a.y0+2, a.x0+2, a.y0)
	if down == "" || down != up {
		t.Errorf("down %q, up %q", down, up)
	}
}

func TestDragPastTheEdgeStaysInThePane(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneList)
	a, _ := scrollBox(m, paneList).textArea()
	m, got := dragged(m, a.x0, a.y0, m.width-1, m.height-1)
	if got != screenText(m, a.x0, a.y0, a.x1, a.y1, a) {
		t.Errorf("a drag off the pane copied %q", got)
	}
}

func TestClickWithNoDragCopiesNothing(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneList)
	b := scrollBox(m, paneList)
	want := click(paneModel(t, paneList), b.x+1, b.y+2).Selected()
	called := false
	m.clip = func(string) error { called = true; return nil }
	m = mouseAt(m, b.x+1, b.y+2, tea.MouseActionPress)
	m = mouseAt(m, b.x+1, b.y+2, tea.MouseActionRelease)
	if called {
		t.Error("a plain click copied")
	}
	if m.Selected() != want {
		t.Error("a plain click no longer moves the cursor")
	}
}

func TestHighlightClears(t *testing.T) {
	t.Parallel()

	for name, end := range map[string]func(Model, rect) Model{
		"a key":    func(m Model, _ rect) Model { return press(m, "j") },
		"a press":  func(m Model, a rect) Model { return mouseAt(m, a.x0, a.y0, tea.MouseActionPress) },
		"a wheel":  func(m Model, a rect) Model { return wheel(m, a.x0, a.y0, false) },
		"a resize": func(m Model, _ rect) Model { return sized(m, m.width-1, m.height) },
	} {
		m := paneModel(t, paneDetail)
		a, _ := scrollBox(m, paneDetail).textArea()
		m, _ = dragged(m, a.x0, a.y0, a.x0+4, a.y0)
		if !m.drag.shown() {
			t.Fatalf("%s: the drag left no highlight to clear", name)
		}
		if m = end(m, a); m.drag.shown() {
			t.Errorf("%s: the highlight stayed", name)
		}
	}
}

func TestHighlightIsDrawnAndCleared(t *testing.T) {
	t.Parallel()

	withColor(func() {
		m := paneModel(t, paneDetail)
		a, _ := scrollBox(m, paneDetail).textArea()
		before := m.View()
		m, _ = dragged(m, a.x0, a.y0, a.x0+4, a.y0)
		shown := m.View()
		if shown == before || plain(shown) != plain(before) {
			t.Error("the band did not draw, or it changed the words")
		}
		// With the pick gone, the screen is the one from before the drag.
		cleared := m
		cleared.drag = drag{}
		cleared.same = false
		if cleared.View() != before {
			t.Error("the band is still drawn with no pick")
		}
	})
}

func TestNoDragWhileSomethingIsOpen(t *testing.T) {
	t.Parallel()

	for name, open := range map[string]func(Model) Model{
		"help":   func(m Model) Model { return press(m, "?") },
		"search": func(m Model) Model { return press(m, "/") },
	} {
		m := open(paneModel(t, paneDetail))
		a, _ := scrollBox(m, paneDetail).textArea()
		m, got := dragged(m, a.x0, a.y0, a.x0+4, a.y0)
		if got != "" || m.drag.on {
			t.Errorf("%s open: a drag picked %q", name, got)
		}
	}
}

func TestReleaseWithNoDragReusesTheFrame(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	m.View()
	if next := mouseAt(m, 1, 1, tea.MouseActionRelease); !next.same {
		t.Error("a release with no drag drew the screen again")
	}
}
```

`?` opens the help and `/` opens search in `model.go`'s `key`. The slug input and the popups have their own open keys; add them to `TestNoDragWhileSomethingIsOpen` in the same form, reading `key` for how they open. `paneList`, `paneDetail`, `paneModel`, `scrollBox`, `press`, `click`, `wheel`, `sized`, `plain` and `withColor` already exist in this package's tests. A blank pick is covered by Task 2's `TestCopyPickedOfSpacesCopiesNothing`.

- [ ] **Step 2: Run tests to verify they fail**

Run: `scripts/test ./internal/tui/ -run 'Drag|Highlight|Release|Click'`
Expected: FAIL, `copied ""` (no drag wiring yet).

- [ ] **Step 3: Write minimal implementation**

In `model.go`, `Update`, in the `tea.WindowSizeMsg` case right after `m.width, m.height = msg.Width, msg.Height`:

```go
		// The picked cells were on the old screen, so they point at nothing now.
		m.drag = drag{}
```

In `key`, as its first line, before the `m.help` check:

```go
	// Any key may move the screen, so the picked cells stop meaning anything.
	m.drag = drag{}
```

In `mouse`, right after the guard that returns when `m.help || m.popup != nil || m.slug != nil || m.searching`:

```go
	switch {
	case msg.Action == tea.MouseActionMotion && m.drag.held:
		next := m.drag.to(msg.X, msg.Y)
		m.same = next == m.drag
		m.drag = next
		return m, nil
	case msg.Action == tea.MouseActionRelease && m.drag.held:
		m.drag.held = false
		if !m.drag.shown() {
			m.drag = drag{}
			m.same = true
			return m, nil
		}
		return m, m.copyDrag()
	case msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft:
		// A new press ends the old pick, and starts a new one when it lands on words.
		m.drag = m.anchorAt(msg.X, msg.Y)
	}
```

In the wheel case of `mouse`, as its first lines:

```go
		// The wheel moves the words, so the picked cells stop meaning anything.
		cleared := m.drag.on
		m.drag = drag{}
```

and in the two wheel paths that set `m.same`, keep a redraw when the band just went: the ignored wheel becomes `m.same = !cleared`, and the gather path becomes `m.same = !flushed && !cleared`.

In `view.go`, `draw`, paint the band on both return paths. Change the narrow return to:

```go
		return m.withDrag(m.frameFrom(strings.Split(m.paneView(m.focus, g.full), "\n"), false))
```

and the last line to:

```go
	return m.withDrag(m.frameFrom(body, true))
```

and add below `draw`:

```go
// withDrag lays the band of a mouse pick over the frame. It uses the colors
// the theme picks for selected text, with no bold, so the words keep their
// width and look.
func (m Model) withDrag(frame string) string {
	if !m.drag.shown() {
		return frame
	}
	return paintDrag(frame, m.drag, m.styles.selected.UnsetBold())
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `scripts/test ./internal/tui/ -run 'Drag|Highlight|Release|Click'`
Expected: PASS.

Then: `scripts/test ./internal/tui/`
Expected: PASS, `TestViewReusesTheFrameForIgnoredMouseAndEmptyTick` and `TestViewDrawsAgainAfterEveryChange` included.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal && go vet ./internal/tui/
git add internal/tui/model.go internal/tui/view.go internal/tui/drag_test.go
git commit -m "Select text with a mouse drag in any pane and copy it on release"
```
