---
id: PLN-0054
created: "2026-09-30"
hash: bq1x3uy
started: "2026-09-30"
finished: "2026-09-30"
---
# Clickable Top Tabs, Wider Sidebar and Pane Key Hints Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Let a click on the top tab bar open a tab, make the left column a third of the screen like lazygit, and show the keys of the focused pane on the status line.

**Architecture:** The tab bar learns the x span of every name it draws, and `mouse` opens the tab under a press on the bar lines. `geometry` sets the left column to `max(width/3, 28)`. A new `hints.go` builds the hint list for the focused pane, and `statusLine` fits it into the room the right side leaves.

**Tech Stack:** Go, Bubble Tea, lipgloss.

**Spec:** `.acta/specs/2026-09-30-tui-tab-click-sidebar-hints-design.md`

**Tests:** `scripts/test ./internal/tui`, `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Start the build only after PLN-0052 (`../acta-tui-tick-items`) has landed on main. It also changes `internal/tui/view.go` and adds the `+` and `-` keys the Tick hints name.
- The left column is `max(width/3, 28)`. There is no top limit. Below 60 columns the narrow layout stays as it is.
- Each hint reads `Label: key`. Hints are joined by ` | `. `Help: ?` is always last and never dropped.
- Comments are plain English a 10-year-old can read: short words, say why. Match the comment style already in `internal/tui`.
- Run `gofmt -l .` and `go vet ./internal/tui` before every commit, inside the task commit.

## File map

- `internal/tui/frame.go`: `geometry` left column width (Task 1).
- `internal/tui/model_test.go`, `internal/tui/scroll_test.go`, `internal/tui/view_test.go`: tests that hardcode the old width (Task 1). `view_test.go` status line tests (Task 3).
- `internal/tui/view.go`: `tabBar` and `barLine` give name spans (Task 2). `hints`, `joinStatus`, `statusLine` (Task 3).
- `internal/tui/model.go`: `mouse` opens the clicked top tab (Task 2).
- `internal/tui/tabbar_test.go` (new): click tests for the bar (Task 2).
- `internal/tui/hints.go` and `internal/tui/hints_test.go` (new): the hint list (Task 3).

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2 (it runs after Task 1 so its tests see the new widths).
- Wave 3: Task 3 (it shares `view.go` with Task 2).

---

### Task 1: The left column is a third of the screen

**Files:**
- Modify: `internal/tui/frame.go` (`geometry`, and the comment above it)
- Modify: `internal/tui/model_test.go` (`TestClickInsideAPaneOnlyFocusesIt` width checks, `clickWidths`)
- Modify: any other test in `internal/tui` that fails only because it hardcodes the old 30% or 48 width

**verify:** At every width of 60 columns or more, the left column is exactly `max(width/3, 28)`, and the detail box starts at that column and fills the rest. No test or comment still says 30% or 48. List every test that changed and why.

**Interfaces:**
- Consumes: nothing.
- Produces: `geom.leftW == max(m.width/3, 28)`.

- [x] **Step 1: Write the failing test**

In `TestClickInsideAPaneOnlyFocusesIt` (`internal/tui/model_test.go`), change the width checks. At 120 columns the left column is now 40:

```go
	if !g.wide || g.leftW != 40 {
		t.Fatalf("wide %v leftW %d", g.wide, g.leftW)
	}
	want := [][4]int{{0, barRows, 40, 18}, {0, barRows + 18, 40, 18}}
```

```go
	if g.detail.x != 40 || g.detail.y != barRows || g.detail.w != 80 || g.detail.h != 36 {
		t.Fatalf("the detail box %+v", g.detail)
	}
```

and the width sweep at the end of the same test:

```go
	// The left column is a third of the width, the way lazygit sizes its
	// side panels, and never under 28 columns.
	for _, w := range []int{60, 80, 100, 120, 150, 200, 240} {
		if got := sized(newModel(t), w, 40).geometry().leftW; got != max(w/3, 28) {
			t.Errorf("at %d columns the left column is %d", w, got)
		}
	}
```

In `clickWidths`, compare with the new rule:

```go
		if w == 60 || max(w/3, 28) != max((w-1)/3, 28) {
```

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/tui -run 'TestClickInsideAPaneOnlyFocusesIt'`
Expected: FAIL with `wide true leftW 36`.

- [x] **Step 3: Write minimal implementation**

In `internal/tui/frame.go`, replace the comment and the width in `geometry`:

```go
// geometry measures the screen. The left column is a third of the width, the
// way lazygit sizes its side panels, and never under 28 columns. Below 60
// columns only the focused box is on screen because two columns of a small
// terminal fit nothing.
```

```go
	g := geom{wide: m.width >= 60, leftW: max(m.width/3, 28), side: make([]box, len(panes))}
```

- [x] **Step 4: Run the package and fix tests that only hardcode the old width**

Run: `scripts/test ./internal/tui`
Expected: PASS. A test that fails only because it expects the old column (36 at 120 columns, 48 on a wide screen, `clamp(w*3/10, 28, 48)`) gets its number changed to the new rule. A test that fails for any other reason is a real break: stop and report it.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./internal/tui
git add internal/tui
git commit -m "tui: left column is a third of the width, like lazygit"
```

### Task 2: A click on a top tab opens it

**Files:**
- Modify: `internal/tui/view.go` (`tabBar`, `barLine`, `tabRows`)
- Modify: `internal/tui/model.go` (`mouse`)
- Create: `internal/tui/tabbar_test.go`

**verify:** At every width from 20 to 200 columns, a left press on any cell of a drawn tab name, on any of the `barRows` top lines, opens that tab, and a press on any other cell of those lines leaves the open tab as it was. The spans come from the same code that draws the names, so they match at every width, including widths that drop names. Help, popup, new-bug and search still swallow the click. List every path checked.

**Interfaces:**
- Consumes: `m.openTab(i int)` in `internal/tui/sidebar.go`, `tabBox{x, w int}` in `internal/tui/view.go`.
- Produces: `func (m Model) barTabs(width int) (string, []tabBox)`. The line is the names line of width `width`. `spans[i]` is where the name of `topTabs[i]` starts, counted from the first inner cell of the bar. A dropped name has `w == 0`.

- [x] **Step 1: Write the failing test**

Create `internal/tui/tabbar_test.go`:

```go
package tui

import (
	"fmt"
	"strings"
	"testing"
)

// TestClickOnATopTabOpensIt reads the drawn bar the way a person does: every
// name it finds opens its tab when clicked, and the cells between names do not.
func TestClickOnATopTabOpensIt(t *testing.T) {
	t.Parallel()

	base := newModel(t)
	for w := 20; w <= 200; w++ {
		m := sized(base, w, 30)
		line := []rune(plain(strings.Split(m.View(), "\n")[1]))
		named := make([]bool, len(line))
		for i, tab := range topTabs {
			name := []rune(fmt.Sprintf("%d %s", i+1, tab.name))
			at := strings.Index(string(line), string(name))
			if at < 0 {
				continue
			}
			x := len([]rune(string(line)[:at]))
			for dx := range name {
				named[x+dx] = true
				for y := 0; y < barRows; y++ {
					if got := click(m, x+dx, y).top; got != i {
						t.Fatalf("at %d columns a click on %q (x %d, y %d) opened tab %d", w, string(name), x+dx, y, got)
					}
				}
			}
		}
		for x := range line {
			if named[x] {
				continue
			}
			if got := click(m, x, 1).top; got != m.top {
				t.Fatalf("at %d columns a click on the blank cell %d opened tab %d", w, x, got)
			}
		}
	}
}

// TestClickOnTheBarIsIgnoredUnderThePopups keeps the bar quiet while help or
// the search box is open, the same as every other click.
func TestClickOnTheBarIsIgnoredUnderThePopups(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 120, 30)
	_, spans := m.barTabs(m.width - 2)
	x := 1 + spans[tabPlans].x
	for _, keys := range [][]string{{"?"}, {"/"}, {"n"}} {
		if got := click(press(m, keys...), x, 1).top; got != m.top {
			t.Errorf("after %v a click on Plans opened tab %d", keys, got)
		}
	}
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/tui -run 'TestClickOnATopTabOpensIt|TestClickOnTheBarIsIgnoredUnderThePopups'`
Expected: FAIL to build with `m.barTabs undefined`.

- [x] **Step 3: Write minimal implementation**

In `internal/tui/view.go`, `barLine` also returns the spans, and `barTabs` does what `tabBar` did:

```go
// barTabs draws the names line of the tab box and says where each name sits,
// so the mouse clicks the same names the screen shows. A name that did not
// fit has a span of zero width.
func (m Model) barTabs(width int) (string, []tabBox) {
	drop := make([]bool, len(topTabs))
	line, spans := m.barLine(drop)
	for lipgloss.Width(line) > width {
		// dropOrder walks the tabs from the right and skips the open one, so
		// the first name at the widest distance is the right one of a tie.
		farthest, at := -1, -1
		for _, j := range dropOrder(len(topTabs), m.top) {
			if drop[j] {
				continue
			}
			if d := max(j-m.top, m.top-j); d > farthest {
				farthest, at = d, j
			}
		}
		if at < 0 {
			break
		}
		drop[at] = true
		line, spans = m.barLine(drop)
	}
	return line, spans
}

// tabBar is the names line alone, for the view.
func (m Model) tabBar(width int) string {
	line, _ := m.barTabs(width)
	return line
}
```

Keep the doc comment that sits above `tabBar` today on `barTabs`, then add the sentence about the spans. In `barLine`, record each name's span as it is added. The line starts with one space and names are joined by two:

```go
func (m Model) barLine(drop []bool) (string, []tabBox) {
	names := make([]string, 0, len(topTabs))
	spans := make([]tabBox, len(topTabs))
	x := 1
	for i, t := range topTabs {
		if drop[i] {
			continue
		}
		name := fmt.Sprintf("%d %s", i+1, t.name)
		spans[i] = tabBox{x: x, w: lipgloss.Width(name)}
		x += lipgloss.Width(name) + 2
		color := m.styles.tabColor(t.kind)
		if i == m.top {
			// The open tab is a band of its own color, so the eye finds it first.
			names = append(names, lipgloss.NewStyle().Bold(true).Foreground(m.styles.bandFG).Background(color).Render(name))
			continue
		}
		names = append(names, lipgloss.NewStyle().Foreground(color).Render(name))
	}
	return " " + strings.Join(names, "  "), spans
}
```

In `internal/tui/model.go`, in `mouse`, right after the block that opens bottom line links, add:

```go
	// The tab box at the top is one more way to open a tab, the same as the
	// number keys.
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft &&
		msg.Y < barRows && m.width >= 2 {
		_, spans := m.barTabs(m.width - 2)
		for i, s := range spans {
			// The bar's left wall takes the first cell, so names start one in.
			if s.w > 0 && msg.X >= 1+s.x && msg.X < 1+s.x+s.w {
				m.openTab(i)
				return m, nil
			}
		}
		m.same = true
		return m, nil
	}
```

`tabRows` keeps calling `m.tabBar(inner)`. Nothing else changes there.

- [x] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/tui`
Expected: PASS, the whole package, so the older click, drag and draw tests still hold.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./internal/tui
git add internal/tui/view.go internal/tui/model.go internal/tui/tabbar_test.go
git commit -m "tui: a click on a top tab opens it"
```

### Task 3: The status line shows the keys of the focused pane

**Files:**
- Create: `internal/tui/hints.go`
- Create: `internal/tui/hints_test.go`
- Modify: `internal/tui/view.go` (`hints` const, `joinStatus`, `statusLine`)
- Modify: `internal/tui/view_test.go` (tests that look for `? help`)

**verify:** For every focused pane, open tab and selected row kind, every hint shown names a key that works there, and no key that is refused there is shown (`s` and `t` on a task or debt line row). At every width, the hints never push the right side off the line, `Help: ?` is always there while the left shows hints, and search text or a status message still takes the place of the hints. List every pane, tab and row kind checked.

**Interfaces:**
- Consumes: `m.Selected() *board.Item`, `m.focus`, `m.top`, `topTabs[i].tree`, `paneList`, `paneDone`, `paneDetail`, `board.KindTask`, `board.KindDebtItem`.
- Produces: `const helpHint = "Help: ?"`, `func (m Model) hints() []string`, `func (m Model) hintLine(w int) string`.

- [x] **Step 1: Write the failing test**

Create `internal/tui/hints_test.go`:

```go
package tui

import (
	"slices"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestHintsFollowTheFocusedPane(t *testing.T) {
	t.Parallel()

	list := press(sized(newModel(t), 200, 40), tabKey(tabSpecs))
	for _, h := range []string{"Detail: enter", "Status: s", "Type: t", "Edit: e", "Copy id: y", "New bug: n", "Sort: o"} {
		if !slices.Contains(list.hints(), h) {
			t.Errorf("the list pane is missing %q: %v", h, list.hints())
		}
	}
	if slices.Contains(list.hints(), "Done tab: [ ]") || slices.Contains(list.hints(), "Fold: space") {
		t.Errorf("the Specs list shows a key it cannot use: %v", list.hints())
	}
	done := press(list, "tab")
	if done.focus != paneDone || !slices.Contains(done.hints(), "Done tab: [ ]") {
		t.Errorf("the Done pane is missing Done tab: %v", done.hints())
	}
	detail := press(list, "enter")
	if detail.focus != paneDetail {
		t.Fatalf("enter gave focus %d", detail.focus)
	}
	for _, h := range []string{"Scroll: j k", "Status: s", "Edit: e", "Copy id: y", "Back: esc"} {
		if !slices.Contains(detail.hints(), h) {
			t.Errorf("the detail pane is missing %q: %v", h, detail.hints())
		}
	}
	if slices.Contains(detail.hints(), "Detail: enter") {
		t.Errorf("the detail pane offers enter: %v", detail.hints())
	}
	plans := press(list, tabKey(tabPlans))
	if !slices.Contains(plans.hints(), "Fold: space") {
		t.Errorf("the Plans list is missing Fold: %v", plans.hints())
	}
}

func TestHintsOnATaskRowOfferTickNotStatus(t *testing.T) {
	t.Parallel()

	m := press(sized(newModel(t), 200, 40), tabKey(tabPlans))
	// Open the first plan and step onto its first task.
	m = press(m, "l", "j")
	it := m.Selected()
	if it == nil || it.Kind != "task" {
		t.Fatalf("the row under the cursor is %+v, want a task", it)
	}
	hs := m.hints()
	if hs[0] != "Tick: +" || hs[1] != "Untick: -" {
		t.Errorf("a task row should start with Tick and Untick: %v", hs)
	}
	if slices.Contains(hs, "Status: s") || slices.Contains(hs, "Type: t") {
		t.Errorf("a task row offers a key the popup refuses: %v", hs)
	}
}

func TestHintLineDropsFromTheRightAndKeepsHelp(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 200, 40)
	full := m.hintLine(1000)
	if !strings.HasSuffix(full, " | "+helpHint) || !strings.HasPrefix(full, m.hints()[0]) {
		t.Fatalf("the full hint line is %q", full)
	}
	for w := 0; w < lipgloss.Width(full); w++ {
		got := m.hintLine(w)
		if !strings.HasSuffix(got, helpHint) {
			t.Fatalf("at %d cells Help is gone: %q", w, got)
		}
		if lipgloss.Width(got) > w && got != helpHint {
			t.Fatalf("at %d cells the line %q does not fit", w, got)
		}
	}
	for w := 30; w <= 200; w++ {
		last := plain(lastLine(sized(m, w, 20).View()))
		if lipgloss.Width(last) != w {
			t.Fatalf("at %d columns the status line is %d cells", w, lipgloss.Width(last))
		}
	}
}
```

In `internal/tui/view_test.go`, the old `? help` checks become `Help: ?`: `TestViewStatusLineShowsHelpAndClock` checks `strings.Contains(plain(last), helpHint)` in place of the `? help` prefix, and the width sweep that sets `sawLeftGone` checks `strings.Contains(line, helpHint)`.

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/tui -run 'TestHints|TestHintLine|TestViewStatusLine'`
Expected: FAIL to build with `m.hints undefined` and `undefined: helpHint`.

- [x] **Step 3: Write minimal implementation**

Create `internal/tui/hints.go`:

```go
package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/iyay/acta/internal/board"
)

// helpHint is the one hint the status line never drops, so the full key map
// is always one key away.
const helpHint = "Help: ?"

// hints lists the keys the focused pane can use right now, most useful first,
// the way lazygit shows them on its bottom line.
func (m Model) hints() []string {
	it := m.Selected()
	// The status popup refuses tasks and debt lines, so those rows offer the
	// tick keys in its place.
	tick := it != nil && (it.Kind == board.KindTask || it.Kind == board.KindDebtItem)
	var out []string
	if tick {
		out = append(out, "Tick: +", "Untick: -")
	}
	if m.focus == paneDetail {
		out = append(out, "Scroll: j k")
		if !tick {
			out = append(out, "Status: s")
		}
		return append(out, "Edit: e", "Copy id: y", "Back: esc")
	}
	out = append(out, "Detail: enter")
	if !tick {
		out = append(out, "Status: s", "Type: t")
	}
	out = append(out, "Edit: e", "Copy id: y", "New bug: n", "Sort: o")
	if topTabs[m.top].tree {
		out = append(out, "Fold: space")
	}
	if m.focus == paneDone {
		out = append(out, "Done tab: [ ]")
	}
	return out
}

// hintLine joins the hints that fit in w cells. It drops them from the right,
// but Help always stays.
func (m Model) hintLine(w int) string {
	hs := m.hints()
	for n := len(hs); n > 0; n-- {
		line := strings.Join(append(hs[:n:n], helpHint), " | ")
		if lipgloss.Width(line) <= w {
			return line
		}
	}
	return helpHint
}
```


In `internal/tui/view.go`:
- Delete the `hints` const and its comment.
- In `joinStatus`, measure `helpHint` in place of `hints`, so the right side keeps room for the one hint that never drops.
- In `statusLine`, start with `left, faintLeft := helpHint, true`. Right after `right := statusText(pieces)`, add:

```go
	// The hints take what the right side leaves, less one cell so the two
	// sides never touch.
	if faintLeft {
		left = m.hintLine(m.width - lipgloss.Width(right) - 1)
	}
```

The rest of `statusLine` stays as it is.

- [x] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/tui`
Expected: PASS, the whole package, so the older status line and link click tests still hold.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./internal/tui
git add internal/tui/hints.go internal/tui/hints_test.go internal/tui/view.go internal/tui/view_test.go
git commit -m "tui: status line shows the keys of the focused pane"
```

## Fix round 1

Review round 1 over `b0aa5c9..7e14629` found two BLOCKERs. Both go in one task and land as one commit.

### Task 4: Fix the stale drag highlight on a bar miss and the hidden Edit and Copy hints

**Files:**
- Modify: `internal/tui/model.go` (`mouse`, the tab bar branch)
- Modify: `internal/tui/hints.go` (`hints`)
- Test: `internal/tui/tabbar_test.go`, `internal/tui/hints_test.go`

**verify:** (1) No left press on any cell of the `barRows` top lines ever leaves a drag highlight on screen after the press cleared the drag. A press on a tab name, a press on a blank cell, and a press while a drag highlight is showing all redraw the screen whenever `m.drag` changed. List every branch of the bar click path and what it sets `m.same` to. (2) Every hint the status line shows names a key that works on that row, and every key that works on the row is shown when it fits. In particular: `Edit: e` shows on every row whose item is on disk, including worktree and legacy rows. `Copy id: y` shows on every row that has an item. `Status: s`, `Type: t`, `Tick: +` and `Untick: -` stay hidden where their key refuses the row. List every row kind checked (group row, worktree row, legacy row, row not on disk, task, debt line, plain item) and the hints each one gets.

**Interfaces:**
- Consumes: `m.drag.on`, `m.anchorAt`, `m.Selected() *board.Item`, `it.OnDisk`, `it.Worktree`, `it.Legacy`, `worktreeModel(t)` in `internal/tui/trees_test.go`.
- Produces: nothing new.

Findings:
1. `internal/tui/model.go`, bar miss branch in `mouse`. The press has already run `m.drag = m.anchorAt(msg.X, msg.Y)`, which clears an old drag because the bar holds no words. The miss branch then sets `m.same = true`, so `View()` reuses the old frame and the old highlight stays on screen with nothing selected under it. Before this plan the same press fell through to `focusPane` and redrew. Expected: the screen redraws when the press cleared a drag. Take the pattern the wheel branch uses: remember `cleared := m.drag.on` before the press handling, then set `m.same = !cleared` on a miss.
2. `internal/tui/hints.go`, the `own` gate in `hints`. It hides `Edit: e` and `Copy id: y` on worktree rows, legacy rows and rows not on disk. But `edit()` opens any item that is on disk, and `copyID()` only refuses a nil item. Repro: `worktreeModel`, Bugs tab, `j` onto the worktree bug (`Worktree == "feat"`, `OnDisk == true`). The hints are `[Detail: enter New bug: n Sort: o]`, but `y` copies and `e` opens the editor. Expected: `Edit: e` when `it != nil && it.OnDisk`, and `Copy id: y` when `it != nil`. `Status: s`, `Type: t` and the tick keys keep the `own` gate, because their code refuses those rows.

- [x] **Step 1: Write the failing tests**

In `internal/tui/tabbar_test.go`, add a test. Start a drag across detail words the way `internal/tui/drag_test.go` does, until `m.drag.on` is true and the view shows the highlight. Then left-press a blank cell of bar line 1. The test asserts that `m.same` is false and that `m.View()` equals the view a fresh model in the same state draws, with no highlight.

In `internal/tui/hints_test.go`, add a test. Use `worktreeModel(t)`, open the Bugs tab and press `j` onto the worktree bug. Assert that `Edit: e` and `Copy id: y` are in `hints()` and that `Status: s` and `Type: t` are not. Also assert that a row with an item that is not on disk shows `Copy id: y` and does not show `Edit: e`.

- [x] **Step 2: Run the tests to verify they fail**

Run: `scripts/test ./internal/tui -run 'TestClickOnTheBar|TestHints'`
Expected: FAIL on the new tests only.

- [x] **Step 3: Fix both findings as the Findings list says**

- [x] **Step 4: Run the package**

Run: `scripts/test ./internal/tui`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./internal/tui
git add internal/tui/model.go internal/tui/hints.go internal/tui/tabbar_test.go internal/tui/hints_test.go
git commit -m "tui: bar miss redraws a cleared drag; Edit and Copy hints follow their keys"
```
