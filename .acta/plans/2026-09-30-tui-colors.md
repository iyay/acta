---
created: "2026-09-30"
id: PLN-0039
hash: eg4p4yy
started: "2026-09-30"
---
# TUI Colors Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** The TUI paints with the theme's colors: each kind has its own color on ids and tabs, rows are no longer faint, the detail has colored labels, dots and problems, the detail footer gets a rule above it, and the `copied` toast hides by itself.

**Architecture:** `newStyles` in `internal/tui/styles.go` grows a few new brushes and a kind-to-slot map. Every new color is a slot number read through the existing `slot` func, so hex themes give hex and the `terminal` theme gives plain ANSI numbers. The list, the detail, the tab bar and the status line then pick these brushes instead of `faint`. The toast is one `tea.Tick` message that clears the status only when the status still holds the text it was sent for.

**Tech Stack:** Go, Bubble Tea v1.3.10, lipgloss (both already in go.mod).

**Spec:** `.acta/specs/2026-09-30-tui-colors-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Every new color is a slot number, never a new hex. The theme gives the hex for that slot, and the `terminal` theme gives the plain ANSI number.
- Slots: scratch green (2), bug red (1), debt yellow (3), spec magenta (5), plan blue (4), task blue (4), Activities tab cyan (6), labels cyan (6), footer labels magenta (5), done dot green (2), waiting dot grey (8), problem lines red (1), `live` green (2).
- Unchanged: the selection band colors, the focus border, the popup dim, and the theme file format.
- No new dependency in go.mod.
- Tests that pin the old look (faint rows, a one-line footer, a faint status line) are updated in the task that changes that look. Never delete a test, never loosen an assert about something else. List every test you changed in the task report.
- Tests use no sleep. A tick is tested by sending its message to `Update`.
- Comments use plain English that a 10-year-old can read back: short words, and they say why.
- Before each commit, run `gofmt -l internal` (it must print nothing), `go vet ./...` and `go test ./...`.

## File Map

- `internal/tui/styles.go`: slot constants, `kindSlots`, new brushes (`label`, `footLabel`, `done`, `waiting`, `problem`, `live`), `kinds`, `cyan`, `bandFG`, `kind()`, `tabColor()`; `work` loses `Faint`.
- `internal/tui/styles_test.go`: role and slot tests, and the `sgrHas` helper later tasks use.
- `internal/tui/scroll.go`: `listView` stops using `faint` for rows; new `paintID`.
- `internal/tui/scroll_test.go`: list row color tests.
- `internal/tui/detail.go`: colored labels, id value, dots, problems, footer labels; `stickyMid` counts two footer lines; new `rule`, `dot`, `paintDates`, `footLines`.
- `internal/tui/view.go`: `detailView` draws the footer rule (Task 3); `barLine` and `statusLine` colors (Task 4).
- `internal/tui/detail_test.go`: detail color and footer tests.
- `internal/tui/view_test.go`: tab bar and status line tests.
- `internal/tui/model.go`: `clearStatusMsg`, `toastFor`, `clearStatusAfter`, `copyID` returns a `tea.Cmd`, `Update` handles the message.
- `internal/tui/model_test.go`: toast tests.

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2, Task 5.
- Wave 3: Task 3.
- Wave 4: Task 4.

Task 3 and Task 4 both touch `view.go`, and Task 2 and Task 3 may both need to update old tests in `view_test.go`, so they sit in different waves.

---

### Task 1: Brushes for kinds and roles

**Files:**
- Modify: `internal/tui/styles.go` (`styles`, the slot constants, `newStyles`)
- Test: `internal/tui/styles_test.go`

**verify:** Every role and every kind named in Global Constraints maps to its slot in every theme: hex themes give `t.ANSI[slot]`, the `terminal` theme gives the ANSI number as text. A kind with no color gives a plain brush, never a panic. `work` is never faint. List every role and kind the test checks, for both themes.

**Interfaces:**
- Consumes: nothing new.
- Produces (all in package `tui`):
  - `const slotRed, slotGreen, slotYellow, slotBlue, slotMagenta, slotCyan = 1, 2, 3, 4, 5, 6`
  - `var kindSlots map[board.Kind]int`
  - `styles` fields: `label, footLabel, done, waiting, problem, live lipgloss.Style`, `kinds map[board.Kind]lipgloss.Color`, `cyan lipgloss.Color`, `bandFG lipgloss.Color`
  - `func (s styles) kind(k board.Kind) lipgloss.Style`
  - `func (s styles) tabColor(k board.Kind) lipgloss.Color`
  - test helper `func sgrHas(s, code string) bool` in `styles_test.go`

- [x] **Step 1: Write the failing test**

Add to `internal/tui/styles_test.go` (add `regexp`, `strconv` and `github.com/iyay/acta/internal/board` to the imports):

```go
// sgr finds each color code a style writes.
var sgr = regexp.MustCompile(`\x1b\[([0-9;]*)m`)

// sgrHas says whether s turns on the plain code, like "1" for bold or "2"
// for faint. The numbers inside a color, like the 2 in "38;2;r;g;b", are
// skipped, so they never pass for faint.
func sgrHas(s, code string) bool {
	for _, m := range sgr.FindAllStringSubmatch(s, -1) {
		ps := strings.Split(m[1], ";")
		for i := 0; i < len(ps); i++ {
			switch ps[i] {
			case "38", "48":
				if i+1 < len(ps) && ps[i+1] == "5" {
					i += 2
				} else {
					i += 4
				}
				continue
			}
			if ps[i] == code {
				return true
			}
		}
	}
	return false
}

func TestKindAndRoleSlots(t *testing.T) {
	t.Parallel()

	hex, _ := theme.Builtin("tokyo-night")
	term, _ := theme.Builtin("terminal")
	for _, c := range []struct {
		name string
		th   theme.Theme
		at   func(int) lipgloss.Color
	}{
		{"tokyo-night", hex, func(i int) lipgloss.Color { return lipgloss.Color(hex.ANSI[i]) }},
		{"terminal", term, func(i int) lipgloss.Color { return lipgloss.Color(strconv.Itoa(i)) }},
	} {
		s := newStyles(c.th, true)
		for k, slot := range map[board.Kind]int{
			board.KindScratch: 2, board.KindBug: 1, board.KindDebt: 3, board.KindDebtItem: 3,
			board.KindStory: 5, board.KindPlan: 4, board.KindTask: 4,
		} {
			if got := s.kind(k).GetForeground(); got != c.at(slot) {
				t.Errorf("%s: kind %s = %v, want slot %d", c.name, k, got, slot)
			}
			if got := s.tabColor(k); got != c.at(slot) {
				t.Errorf("%s: tab %s = %v, want slot %d", c.name, k, got, slot)
			}
		}
		if got := s.tabColor(""); got != c.at(6) {
			t.Errorf("%s: Activities tab = %v, want slot 6", c.name, got)
		}
		for name, r := range map[string]struct {
			brush lipgloss.Style
			slot  int
		}{
			"label": {s.label, 6}, "footLabel": {s.footLabel, 5}, "done": {s.done, 2},
			"waiting": {s.waiting, 8}, "problem": {s.problem, 1}, "live": {s.live, 2},
		} {
			if got := r.brush.GetForeground(); got != c.at(r.slot) {
				t.Errorf("%s: %s = %v, want slot %d", c.name, name, got, r.slot)
			}
		}
		if s.work.GetFaint() {
			t.Errorf("%s: work is still faint", c.name)
		}
	}
}

func TestKindWithNoColorIsPlain(t *testing.T) {
	t.Parallel()

	th, _ := theme.Builtin("tokyo-night")
	s := newStyles(th, true)
	if _, ok := s.kind("").GetForeground().(lipgloss.NoColor); !ok {
		t.Fatalf("empty kind has color %v", s.kind("").GetForeground())
	}
	if _, ok := s.kind("nope").GetForeground().(lipgloss.NoColor); !ok {
		t.Fatalf("unknown kind has color %v", s.kind("nope").GetForeground())
	}
}

func TestBandTextUsesTheBackground(t *testing.T) {
	t.Parallel()

	hex, _ := theme.Builtin("tokyo-night")
	if got := newStyles(hex, true).bandFG; got != lipgloss.Color(hex.BG) {
		t.Fatalf("tokyo-night band text = %v, want %v", got, hex.BG)
	}
	term, _ := theme.Builtin("terminal")
	if got := newStyles(term, true).bandFG; got != lipgloss.Color("0") {
		t.Fatalf("terminal band text = %v, want slot 0", got)
	}
}

func TestSgrHasSkipsColorNumbers(t *testing.T) {
	t.Parallel()

	if sgrHas("\x1b[38;2;2;2;2mx\x1b[0m", "2") {
		t.Fatal("a 24-bit color passed for faint")
	}
	if !sgrHas("\x1b[2;38;2;1;1;1mx\x1b[0m", "2") {
		t.Fatal("faint next to a color was missed")
	}
	if !sgrHas("\x1b[1mx\x1b[0m", "1") {
		t.Fatal("plain bold was missed")
	}
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run 'TestKindAndRoleSlots|TestKindWithNoColorIsPlain|TestBandTextUsesTheBackground|TestSgrHasSkipsColorNumbers'`
Expected: FAIL to build with `s.kind undefined` (and the other new names).

- [x] **Step 3: Write minimal implementation**

In `internal/tui/styles.go`, import `github.com/iyay/acta/internal/board`, then:

```go
// styles are the brushes the TUI paints with. They come from one theme, so
// changing the theme changes every color at once.
type styles struct {
	accent, work, faint, dim, selected           lipgloss.Style
	label, footLabel, done, waiting, problem, live lipgloss.Style
	accentColor                                  lipgloss.TerminalColor
	kinds                                        map[board.Kind]lipgloss.Color
	cyan                                         lipgloss.Color // the Activities tab, which holds no one kind
	bandFG                                       lipgloss.Color // text on a colored band
	bg, fg                                       string         // empty for the terminal theme
}

// Each role always takes the same slot, so any theme with 16 colors works.
const (
	slotAccent  = 12
	slotWork    = 4
	slotDim     = 8
	slotRed     = 1
	slotGreen   = 2
	slotYellow  = 3
	slotBlue    = 4
	slotMagenta = 5
	slotCyan    = 6
)

// kindSlots gives each kind a color of its own, so an id tells what it is
// before anyone reads it. A task is part of a plan, so it wears the plan color.
var kindSlots = map[board.Kind]int{
	board.KindScratch: slotGreen, board.KindBug: slotRed,
	board.KindDebt: slotYellow, board.KindDebtItem: slotYellow,
	board.KindStory: slotMagenta, board.KindPlan: slotBlue, board.KindTask: slotBlue,
}
```

In `newStyles`, before the `return`:

```go
	kinds := make(map[board.Kind]lipgloss.Color, len(kindSlots))
	for k, i := range kindSlots {
		kinds[k] = slot(i)
	}
	// The terminal theme has no background color, so text on a band takes
	// slot 0, the dark end of its colors.
	bandFG := slot(0)
	if t.BG != "" {
		bandFG = lipgloss.Color(t.BG)
	}
```

In the returned struct: drop `.Faint(true)` from `work` and change its comment to `// work marks a row whose work has begun.`, then add:

```go
		label:     lipgloss.NewStyle().Foreground(slot(slotCyan)),
		footLabel: lipgloss.NewStyle().Foreground(slot(slotMagenta)),
		done:      lipgloss.NewStyle().Foreground(slot(slotGreen)),
		waiting:   lipgloss.NewStyle().Foreground(slot(slotDim)),
		problem:   lipgloss.NewStyle().Foreground(slot(slotRed)),
		live:      lipgloss.NewStyle().Foreground(slot(slotGreen)),
		kinds:     kinds,
		cyan:      slot(slotCyan),
		bandFG:    bandFG,
```

After `newStyles`:

```go
// kind is the brush of an id of kind k. A kind with no color of its own
// gets a plain brush, so its id reads in the theme foreground.
func (s styles) kind(k board.Kind) lipgloss.Style {
	c, ok := s.kinds[k]
	if !ok {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Foreground(c)
}

// tabColor is the color of a top tab. The Activities tab has no kind, so it
// takes cyan.
func (s styles) tabColor(k board.Kind) lipgloss.Color {
	if c, ok := s.kinds[k]; ok {
		return c
	}
	return s.cyan
}
```

If `TestHexThemeRoles` or another old test checks that `work` is faint, update that one assert to check it is not faint.

- [x] **Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l internal && go vet ./... && go test ./...
git add internal/tui/styles.go internal/tui/styles_test.go
git commit -m "tui: add kind and role brushes from theme slots"
```

---

### Task 2: List rows at full brightness with colored ids

**Files:**
- Modify: `internal/tui/scroll.go` (`listView`)
- Test: `internal/tui/scroll_test.go`

**verify:** No list row is faint on any path: a plain row, a row with work under way, a tree row, a task row, a group row. The id of every row that is not under the cursor is drawn in its kind color whenever the id is on screen whole. The row under the cursor is the selection band, bold. List every row type the test checks.

**Interfaces:**
- Consumes: `styles.kind(board.Kind) lipgloss.Style`, `styles.work`, `styles.selected` (Task 1); test helper `sgrHas` (Task 1); `actModel`, `press`, `tabKey`, `plain` (existing test helpers).
- Produces: `func (m Model) paintID(text string, it *board.Item, base lipgloss.Style) string`.

- [ ] **Step 1: Write the failing test**

Add to `internal/tui/scroll_test.go`. `actModel(t)` opens on the Activities tab with `BUG-0002` under the cursor, then the task `PLN-0004.01` (work under way), then the plan `PLN-0003` (see `TestYCopiesTheIDOfTheRow` in `model_test.go`). Check `rowIDs` first if the order differs, and pick the rows by id, not by index.

```go
// listLines draws the list box of the open tab the way the screen does.
func listLines(m Model) []string {
	p := m.listPane()
	b := m.geometry().at(p)
	return m.listView(p, b.textW(), b)
}

func TestListRowsAreNotFaintAndWearKindColors(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true)
		lines := listLines(m)
		byID := map[string]string{}
		for _, ln := range lines {
			f := strings.Fields(plain(ln))
			if len(f) > 0 {
				byID[strings.TrimLeft(f[0], "+-●○✓")] = ln
				if len(f) > 1 {
					byID[f[1]] = ln
				}
			}
		}
		for _, c := range []struct {
			id   string
			kind board.Kind
		}{{"PLN-0004.01", board.KindTask}, {"PLN-0003", board.KindPlan}} {
			ln, ok := byID[c.id]
			if !ok {
				t.Fatalf("no row for %s in %q", c.id, plainLines(lines))
			}
			if sgrHas(ln, "2") {
				t.Errorf("%s row is faint: %q", c.id, ln)
			}
			if !strings.Contains(ln, m.styles.kind(c.kind).Render(c.id)) {
				t.Errorf("%s id is not in its kind color: %q", c.id, ln)
			}
		}
		sel := byID["BUG-0002"]
		if !sgrHas(sel, "1") {
			t.Errorf("selected row is not bold: %q", sel)
		}
		if sgrHas(sel, "2") {
			t.Errorf("selected row is faint: %q", sel)
		}
	})
}

func TestListRowWithWorkKeepsTheWorkColor(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true)
		want := m.styles.work.Render("  ") // the gap after the id is in the work brush
		for _, ln := range listLines(m) {
			if strings.Contains(plain(ln), "PLN-0004.01") {
				if !strings.Contains(ln, strings.TrimSuffix(want, "\x1b[0m")) {
					t.Fatalf("work row lost the work color: %q", ln)
				}
				return
			}
		}
		t.Fatal("no row for PLN-0004.01")
	})
}

func TestPaintIDLeavesACutIDPlain(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true)
		it := &board.Item{ShortID: "PLN-0003", Kind: board.KindPlan}
		base := lipgloss.NewStyle()
		if got := m.paintID("PLN-0", it, base); got != base.Render("PLN-0") {
			t.Fatalf("cut id got a color: %q", got)
		}
		if got := m.paintID("x", nil, base); got != base.Render("x") {
			t.Fatalf("row with no item got a color: %q", got)
		}
	})
}
```

Add imports `github.com/charmbracelet/lipgloss` and `github.com/iyay/acta/internal/board` if the file lacks them.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run 'TestListRowsAreNotFaintAndWearKindColors|TestListRowWithWorkKeepsTheWorkColor|TestPaintIDLeavesACutIDPlain'`
Expected: FAIL to build with `m.paintID undefined`.

- [ ] **Step 3: Write minimal implementation**

In `listView` (`internal/tui/scroll.go`), replace the brush switch and the `append` with:

```go
		it := m.board.Get(rows[n].id)
		text := pad(m.rowText(rows[n], it, w), w)
		if n == cur {
			out = append(out, m.styles.selected.Render(text))
			continue
		}
		base := lipgloss.NewStyle()
		if inProgress(it) {
			base = m.styles.work
		}
		out = append(out, m.paintID(text, it, base))
```

Add below `listView`:

```go
// paintID draws a row with its id in the color of its kind and the rest in
// base. A row with no item, or whose id was cut off, is all base.
func (m Model) paintID(text string, it *board.Item, base lipgloss.Style) string {
	if it == nil {
		return base.Render(text)
	}
	name := it.ShortID
	if name == "" {
		name = it.ID
	}
	i := strings.Index(text, name)
	if i < 0 {
		return base.Render(text)
	}
	return base.Render(text[:i]) + m.styles.kind(it.Kind).Render(name) + base.Render(text[i+len(name):])
}
```

Leave `nothing here` faint: it is a hint, not a row. Update any old test that checks rows are faint.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal && go vet ./... && go test ./...
git add internal/tui/scroll.go internal/tui/scroll_test.go
git commit -m "tui: draw list rows at full brightness with kind-colored ids"
```

(Add any old test file you updated to the `git add`.)

---

### Task 3: Detail colors and the footer rule

**Files:**
- Modify: `internal/tui/detail.go` (`buildDetailParts`, `detailLines`, `stickyMid`, `stepLines`, `workLine`)
- Modify: `internal/tui/view.go` (`detailView` only)
- Test: `internal/tui/detail_test.go`

**verify:** In the detail, no header label, work line, step line, problem line or footer line is faint. Every label is cyan, the id on the ID line is in its kind color, every dot is in the color of its state (done green, under way the accent, waiting grey), every `!` line is red, the footer labels are magenta. The footer is always two lines, the rule first, on both the sticky path and the one-block path. A pane too short for the header, both footer lines and 3 middle lines always falls back to one scrolling block. List each path the tests check.

**Interfaces:**
- Consumes: `styles.label`, `footLabel`, `done`, `waiting`, `problem`, `kind()`, `accent`, `work`, `faint` (Task 1); `sgrHas` (Task 1); `paintID` (Task 2, `scroll.go`).
- Produces: `const footLines = 2`; `func (m Model) rule(w int) string`; `func (s styles) dot(mark string) lipgloss.Style`; `func paintDates(s styles, line string) string`.

- [ ] **Step 1: Write the failing test**

Add to `internal/tui/detail_test.go`:

```go
func TestDetailLabelsAndIDWearColors(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true) // BUG-0002 under the cursor
		head, _, _ := m.detailParts(100)
		var id, status string
		for _, ln := range head {
			switch {
			case strings.HasPrefix(plain(ln), "ID"):
				id = ln
			case strings.HasPrefix(plain(ln), "STATUS"):
				status = ln
			}
		}
		if !strings.HasPrefix(status, m.styles.label.Render("STATUS")) {
			t.Errorf("STATUS label is not cyan: %q", status)
		}
		if !strings.Contains(id, m.styles.kind(board.KindBug).Render("BUG-0002")) {
			t.Errorf("id is not in the bug color: %q", id)
		}
		for _, ln := range head {
			if sgrHas(ln, "2") && !strings.Contains(plain(ln), "──") {
				t.Errorf("header line is faint: %q", ln)
			}
		}
	})
}

func TestDetailDotsProblemsAndWorkLines(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true)
		s := m.styles
		if s.dot(dotDone).GetForeground() != s.done.GetForeground() ||
			s.dot(dotGoing).GetForeground() != s.accent.GetForeground() ||
			s.dot(dotWaiting).GetForeground() != s.waiting.GetForeground() {
			t.Fatal("a dot does not wear the color of its state")
		}
		it := &board.Item{ShortID: "PLN-0003", Kind: board.KindPlan, Title: "Plan"}
		off := workLine(s, it, false, 60)
		if sgrHas(off, "2") {
			t.Errorf("work line is faint: %q", off)
		}
		if !strings.HasPrefix(off, s.dot(dotWaiting).Render(dotWaiting)) {
			t.Errorf("waiting dot is not grey: %q", off)
		}
		if on := workLine(s, it, true, 60); !sgrHas(on, "1") {
			t.Errorf("the line the reader is on is not bold: %q", on)
		}
		bug := m.Selected()
		bug.Problems = []string{"broken"}
		m.dcache = &detailCache{}
		_, mid, _ := m.detailParts(60)
		if !strings.Contains(strings.Join(mid, "\n"), s.problem.Render("! broken")) {
			t.Errorf("problem line is not red: %q", mid)
		}
	})
}

func TestDetailFooterHasARuleAndColoredLabels(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true)
		for _, h := range []int{40, 6} { // sticky, then one block
			lines := m.detailView(60, 0, h)
			if h == 6 {
				lines = m.detailLines(60)
			}
			n := len(lines)
			for n > 0 && strings.TrimSpace(plain(lines[n-1])) == "" {
				n--
			}
			foot, rule := lines[n-1], plain(lines[n-2])
			if strings.Trim(rule, "─") != "" || rule == "" {
				t.Errorf("h=%d: line above the footer is %q, want a rule", h, rule)
			}
			if !strings.Contains(foot, m.styles.footLabel.Render("created")) {
				t.Errorf("h=%d: footer label is not magenta: %q", h, foot)
			}
		}
	})
}

func TestStickyMidCountsBothFooterLines(t *testing.T) {
	t.Parallel()

	for _, c := range []struct{ head, h, want int }{
		{5, 10, 3}, // 5 head + 2 foot + 3 middle fits exactly
		{5, 9, 0},  // one line short: the whole detail scrolls
		{0, 5, 3},
		{0, 4, 0},
	} {
		if got := stickyMid(c.head, c.h); got != c.want {
			t.Errorf("stickyMid(%d, %d) = %d, want %d", c.head, c.h, got, c.want)
		}
	}
}
```

If `actModel`'s bug has no `created` date, the footer still says `created -`, so the assert holds. If the test finds `detailCache` keyed so a changed `Problems` is not seen, the fresh `&detailCache{}` above handles it.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run 'TestDetailLabelsAndIDWearColors|TestDetailDotsProblemsAndWorkLines|TestDetailFooterHasARuleAndColoredLabels|TestStickyMidCountsBothFooterLines'`
Expected: FAIL to build with `s.dot undefined`.

- [ ] **Step 3: Write minimal implementation**

In `internal/tui/detail.go`:

```go
// footLines is how tall the detail footer is: the rule and the date line.
const footLines = 2

// rule is the line that sets the header and the footer off from the middle.
func (m Model) rule(w int) string {
	return m.styles.faint.Render(strings.Repeat("─", max(1, w)))
}

// dot is the brush of a status dot: green when done, the accent while the
// work is under way, grey while it waits.
func (s styles) dot(mark string) lipgloss.Style {
	switch mark {
	case dotDone:
		return s.done
	case dotGoing:
		return s.accent
	}
	return s.waiting
}

// paintDates colors the names in the date line, so the dates stand out from
// the words around them. Each part is "name date", split by " · ".
func paintDates(s styles, line string) string {
	parts := strings.Split(line, " · ")
	for i, p := range parts {
		if name, rest, ok := strings.Cut(p, " "); ok {
			parts[i] = s.footLabel.Render(name) + " " + rest
		}
	}
	return strings.Join(parts, " · ")
}
```

In `buildDetailParts`, the header loop becomes:

```go
	for _, f := range fields {
		if f.value == "" {
			continue
		}
		line := truncate(expandTabs(fmt.Sprintf("%-*s: %s", width, f.label, f.value)), w)
		if f.label == "ID" {
			line = m.paintID(line, it, lipgloss.NewStyle())
		}
		// A label cut by a narrow pane is left as it is.
		if f.label != "" && strings.HasPrefix(line, f.label) {
			line = m.styles.label.Render(f.label) + line[len(f.label):]
		}
		head = append(head, line)
	}
	head = append(head, m.rule(w))
```

`paintID` comes from Task 2 (`scroll.go`). Paint the id first, then the label, because `paintID` looks for the id text and the label is plain ASCII. When the label is "ID", the id text never starts at 0, so the two do not overlap.

The problem line becomes `mid = append(mid, m.styles.problem.Render(truncate(expandTabs("! "+p), w)))`, and the return is `return head, mid, paintDates(m.styles, dateLine(it, w))`.

`detailLines`:

```go
	if foot != "" {
		out = append(out, m.rule(w), foot)
	}
```

`stickyMid`: change `h - head - 1` to `h - head - footLines`, and change "the footer" in its comment to "the two footer lines".

`stepLines`: the start brush `m.styles.faint` becomes `lipgloss.NewStyle()`. Paint the dot on its own:

```go
			line := truncate(expandTabs(mark+" "+s.Text), w)
			if strings.HasPrefix(line, mark) {
				line = m.styles.dot(mark).Render(mark) + brush.Render(line[len(mark):])
			} else {
				line = brush.Render(line)
			}
			out = append(out, line)
```

`workLine`: the start brush `s.faint` becomes `lipgloss.NewStyle()`, and `on` sets `brush = lipgloss.NewStyle().Bold(true)`. The return becomes the same dot split as in `stepLines`, with `s.dot(mark)`. Change its comment: "A line the reader is on is bold; the rest are plain, and the dot wears the color of its state."

In `internal/tui/view.go`, `detailView`:

```go
	for len(out) < h-footLines {
		out = append(out, "")
	}
	return append(out, m.rule(w), foot)
```

Update old tests that pin a one-line footer or the old `stickyMid` numbers.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal && go vet ./... && go test ./...
git add internal/tui/detail.go internal/tui/view.go internal/tui/detail_test.go
git commit -m "tui: color detail labels, dots and problems, add a rule over the footer"
```

---

### Task 4: Colored top tabs and status line

**Files:**
- Modify: `internal/tui/view.go` (`barLine`, `statusLine`)
- Test: `internal/tui/view_test.go`

**verify:** No tab name is faint. The open tab is always a band in its own kind color with the band text color, bold, and every other tab is text in its kind color, on every tab and in both themes. On the status line the project name is in the accent and `live` is green whenever they are on screen, `paused` and the rest stay faint, and every link can still be clicked (`statusLineBoxes` finds every link). Tab clicks land on the same names as before. List the tabs and status pieces checked.

**Interfaces:**
- Consumes: `styles.tabColor(board.Kind) lipgloss.Color`, `styles.bandFG`, `styles.live`, `styles.accent` (Task 1); `topTabs` (`sidebar.go`); `sgrHas` (Task 1).
- Produces: `func (m Model) paintRight(pieces []statusPiece, right string) string`.

- [ ] **Step 1: Write the failing test**

Add to `internal/tui/view_test.go`:

```go
func TestTabsWearTheirKindColors(t *testing.T) {
	withTrueColor(func() {
		for _, name := range []string{"tokyo-night", "terminal"} {
			for open := range topTabs {
				m := press(actModel(t).WithTheme(name, true), tabKey(open))
				line := m.barLine(make([]bool, len(topTabs)))
				for i, tb := range topTabs {
					label := fmt.Sprintf("%d %s", i+1, tb.name)
					want := lipgloss.NewStyle().Foreground(m.styles.tabColor(tb.kind)).Render(label)
					if i == open {
						want = lipgloss.NewStyle().Bold(true).Foreground(m.styles.bandFG).
							Background(m.styles.tabColor(tb.kind)).Render(label)
					}
					if !strings.Contains(line, want) {
						t.Errorf("%s, open %d: tab %q is not drawn as %q in %q", name, open, label, want, line)
					}
				}
				if sgrHas(line, "2") {
					t.Errorf("%s, open %d: a tab is faint: %q", name, open, line)
				}
			}
		}
	})
}

func TestStatusLineColorsProjectAndLive(t *testing.T) {
	withTrueColor(func() {
		m := sized(actModel(t).WithTheme("tokyo-night", true), 200, 40)
		line := m.statusLine()
		project := filepath.Base(m.cfg.RepoRoot)
		if !strings.Contains(line, m.styles.accent.Render(project)) {
			t.Errorf("project %q is not in the accent: %q", project, line)
		}
		if !strings.Contains(line, m.styles.live.Render("live")) {
			t.Errorf("live is not green: %q", line)
		}
		m.manual = true
		if strings.Contains(m.statusLine(), m.styles.live.Render("paused")) {
			t.Error("paused is green")
		}
	})
}
```

Add `fmt` and `path/filepath` to the imports if missing. The existing `statusLineBoxes` tests stay as they are and must still pass.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run 'TestTabsWearTheirKindColors|TestStatusLineColorsProjectAndLive'`
Expected: FAIL: tabs are drawn faint and in the accent, and the project is faint.

- [ ] **Step 3: Write minimal implementation**

`barLine` (`internal/tui/view.go`):

```go
		name := fmt.Sprintf("%d %s", i+1, t.name)
		color := m.styles.tabColor(t.kind)
		if i == m.top {
			// The open tab is a band of its own color, so the eye finds it first.
			names = append(names, lipgloss.NewStyle().Bold(true).Foreground(m.styles.bandFG).Background(color).Render(name))
			continue
		}
		names = append(names, lipgloss.NewStyle().Foreground(color).Render(name))
```

Update the `tabBar` comment: "the open one as a band of its color" in place of "the open one in the accent".

`statusLine`: the last line becomes `return left + gap + m.paintRight(pieces, right)`. The narrow path that returns `fit(right, m.width)` stays as it is. Add:

```go
// paintRight colors the words on the right that tell something at a glance:
// the project in the accent and live in green. The rest stays faint. It finds
// each word in order, so a word cut off by a narrow window keeps the faint
// brush, and links keep their wrappers so a click still finds them.
func (m Model) paintRight(pieces []statusPiece, right string) string {
	var b strings.Builder
	at := 0
	for i, p := range pieces {
		var brush lipgloss.Style
		switch {
		case p.url != "":
			continue
		case i == 0 && p.text == filepath.Base(m.cfg.RepoRoot):
			brush = m.styles.accent
		case p.text == "live":
			brush = m.styles.live
		default:
			continue
		}
		j := strings.Index(right[at:], p.text)
		if j < 0 {
			continue
		}
		if j > 0 {
			b.WriteString(m.styles.faint.Render(right[at : at+j]))
		}
		b.WriteString(brush.Render(p.text))
		at += j + len(p.text)
	}
	if at < len(right) {
		b.WriteString(m.styles.faint.Render(right[at:]))
	}
	return b.String()
}
```

Update old tests that check the open tab is in the accent or that the right side is one faint span.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/`
Expected: PASS, including every existing `statusLineBoxes` and tab-click test.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal && go vet ./... && go test ./...
git add internal/tui/view.go internal/tui/view_test.go
git commit -m "tui: color top tabs by kind and the project and live words"
```

---

### Task 5: The copy toast hides by itself

**Files:**
- Modify: `internal/tui/model.go` (`copyID`, the `"y"` key case, `Update`)
- Test: `internal/tui/model_test.go`

**verify:** Only a successful copy ever starts the timer. The clear message empties the status only when the status is still the exact text it was sent for; any newer status, any error and `nothing selected` stay. List every status path the tests check.

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: `type clearStatusMsg struct{ text string }`; `const toastFor = 2 * time.Second`; `func clearStatusAfter(d time.Duration, text string) tea.Cmd`; `func (m *Model) copyID() tea.Cmd`.

- [ ] **Step 1: Write the failing test**

Add to `internal/tui/model_test.go`:

```go
func TestCopyToastHidesByItself(t *testing.T) {
	t.Parallel()

	m := actModel(t)
	m.clip = func(string) error { return nil }
	next, cmd := m.Update(key("y"))
	m = next.(Model)
	if cmd == nil {
		t.Fatal("a copy started no timer, so the toast would stay")
	}
	if m.status != "copied BUG-0002" {
		t.Fatalf("status %q", m.status)
	}
	next, _ = m.Update(clearStatusMsg{text: "copied BUG-0002"})
	if got := next.(Model).status; got != "" {
		t.Fatalf("toast did not hide: %q", got)
	}
}

func TestCopyToastKeepsANewerMessage(t *testing.T) {
	t.Parallel()

	m := actModel(t)
	m.status = "reload failed: disk"
	next, _ := m.Update(clearStatusMsg{text: "copied BUG-0002"})
	if got := next.(Model).status; got != "reload failed: disk" {
		t.Fatalf("a newer message was cleared: %q", got)
	}
}

func TestCopyErrorsDoNotHide(t *testing.T) {
	t.Parallel()

	m := actModel(t)
	m.clip = func(string) error { return errors.New("no clipboard") }
	next, cmd := m.Update(key("y"))
	if cmd != nil {
		t.Error("a failed copy started a timer")
	}
	if got := next.(Model).status; !strings.HasPrefix(got, "copy failed") {
		t.Errorf("status %q", got)
	}

	cfg := treeCfg(t, map[string]string{})
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	next, cmd = New(cfg, b, true).Update(key("y"))
	if cmd != nil {
		t.Error("nothing selected started a timer")
	}
	if got := next.(Model).status; got != "nothing selected" {
		t.Errorf("status %q", got)
	}
}

func TestClearStatusAfterSendsItsText(t *testing.T) {
	t.Parallel()

	if got := clearStatusAfter(0, "copied X")(); got != (clearStatusMsg{text: "copied X"}) {
		t.Fatalf("got %#v", got)
	}
}
```

Add `errors` to the imports if missing.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run 'TestCopyToast|TestCopyErrorsDoNotHide|TestClearStatusAfterSendsItsText'`
Expected: FAIL to build with `undefined: clearStatusMsg`.

- [ ] **Step 3: Write minimal implementation**

In `internal/tui/model.go`, next to the other message types:

```go
// clearStatusMsg hides a toast. It carries the text it was sent for, so a
// newer message that took the line in the meantime stays.
type clearStatusMsg struct{ text string }

// toastFor is how long a copy message stays on the status line.
const toastFor = 2 * time.Second

// clearStatusAfter sends the clear message for text once d has passed.
func clearStatusAfter(d time.Duration, text string) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return clearStatusMsg{text: text} })
}
```

In `Update`, next to `case clockMsg:`:

```go
	case clearStatusMsg:
		if m.status == msg.text {
			m.status = ""
		}
```

`copyID` returns a `tea.Cmd`: `return nil` on both error paths, and the end becomes:

```go
	m.status = "copied " + id
	return clearStatusAfter(toastFor, m.status)
```

Change its comment to add: "A good copy hides itself after a short while; an error stays until the next message."

The `"y"` key case. Call first, then return, because Go does not promise that `m` is read after the call in `return m, m.copyID()`:

```go
	case "y":
		cmd := m.copyID()
		return m, cmd
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/`
Expected: PASS, including `TestYCopiesTheIDOfTheRow` and `TestYCopiesTheIDOfARowWithNoNumber`.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal && go vet ./... && go test ./...
git add internal/tui/model.go internal/tui/model_test.go
git commit -m "tui: hide the copy toast after two seconds"
```
