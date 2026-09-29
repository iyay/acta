---
parent: specs/2026-09-29-tui-notes-round4-design
id: PLN-0036
created: "2026-09-29"
hash: q5y8n99
started: "2026-09-29"
---
# TUI Notes Round 4 Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Ship the ten TUI notes of SCR-0017: a tab box, new pane titles, the sort word in the bottom border, no `0` key, a sticky Detail, sort by short id, Activities as a tree, `h` and `l` to fold, `y` to copy an id, and a popup dim that the user can see.

**Architecture:** Every change lives in `internal/tui`. Each note is one task with its own test in the package's existing test files. The tasks share `view.go`, `model.go` and `sidebar.go`, so most waves hold one task.

**Tech Stack:** Go, Bubble Tea, lipgloss, `go test`.

**Spec:** `.acta/specs/2026-09-29-tui-notes-round4-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- TDD: write the failing test first. No implementation code before a red test. Cover every behavior the task names: its branches, error paths, and adversarial inputs. Run the tests and paste the output before claiming done.
- Run `gofmt -l internal/tui` (must print nothing) and `go vet ./internal/tui` (must pass) before each commit, inside the task commit.
- Commit messages are English.
- Code comments are plain English a 10-year-old can read. They say why the line exists. No marker tags, no Latin abbreviations, no emoji.
- No new dependencies. OSC 52 is written by hand. `pbcopy` is run through `os/exec`.
- Keys `1`-`6` do not change. Left and right arrows keep switching tabs. `o` still flips the sort. `enter` keeps toggling tree heads as today; `enter` on a task row opens the detail.
- Out of scope: a time on `created`, tab numbers above 6, a collapse-all key.
- An existing test that pins the old behavior of a task is updated in that same task, so it checks the new behavior. Never delete a test to make the suite pass.

## File Map

- `internal/tui/view.go`: tab box (`View`, `tabBar`, `barLine`, new `tabRows`), pane footer (`paneView`), help text (`helpLines`), sticky detail (`detailView`, new `window`).
- `internal/tui/frame.go`: `geometry` (room for the tab box), `paneTop` (sort word leaves), new `sortWord`, new `sortFoot`.
- `internal/tui/scroll.go`: `boxOf` (tab box room), `fitOf` and `linesAt` (sticky detail), `treeMark` (Activities groups).
- `internal/tui/sidebar.go`: `tabsOf` (titles), `paneKey` (no `[0]`), `activityRows`, new `headOf`, new `isOpen`, new `setOpen`.
- `internal/tui/model.go`: `key` (drop `0`, add `h`, `l`, `y`), `Model` fields `shutActs` and `clip`, `toggleRow`, new `foldRow`, new `copyID`, new `copyRef`.
- `internal/tui/detail.go`: new `detailParts`, `dateLine`, `stickyMid`, `detailScroll`; `detailLines` becomes the joined block.
- `internal/tui/order.go`: `ordered` sorts by short id number, new `idNum`.
- `internal/tui/styles.go`: dim color, new `mixHex`, `hexRGB`.
- `internal/tui/title.go`: new `osc52`, `copyText`.
- Tests: `view_test.go`, `model_test.go`, `frame_test.go`, `scroll_test.go`, `sidebar_test.go`, `detail_test.go`, `order_test.go`, `plantree_test.go`, `styles_test.go`.

## Waves

Two tasks in one wave never touch the same file, test files included.

- Wave 1: Task 2 (`sidebar.go`, `sidebar_test.go`, `model_test.go`), Task 6 (`order.go`, `order_test.go`), Task 10 (`styles.go`, `styles_test.go`, `view_test.go`).
- Wave 2: Task 1 (`view.go`, `frame.go`, `scroll.go`, `view_test.go`, `model_test.go`).
- Wave 3: Task 3 (`frame.go`, `view.go`, `frame_test.go`, `scroll_test.go`).
- Wave 4: Task 4 (`sidebar.go`, `model.go`, `view.go`, `model_test.go`, `scroll_test.go`, `view_test.go`, `order_test.go`, `sidebar_test.go`).
- Wave 5: Task 5 (`detail.go`, `view.go`, `scroll.go`, `detail_test.go`).
- Wave 6: Task 7 (`sidebar.go`, `model.go`, `scroll.go`, `model_test.go`). Needs Task 6 (`idNum` order).
- Wave 7: Task 8 (`model.go`, `view.go`, `plantree_test.go`, `model_test.go`). Needs Task 7 (`isOpen`, `setOpen`, `actModel`).
- Wave 8: Task 9 (`model.go`, `title.go`, `view.go`, `model_test.go`). Needs Task 7 (`actModel`).

---

### Task 1: Tabs in a box of their own

**Files:**
- Modify: `internal/tui/view.go` (`View` at :77, `tabBar` at :118, `barLine` at :144)
- Modify: `internal/tui/frame.go` (`geometry` at :13)
- Modify: `internal/tui/scroll.go` (`boxOf` at :57)
- Test: `internal/tui/view_test.go`, `internal/tui/model_test.go`

**verify:** On every tab and at every width, the top three screen lines are one closed box as wide as the screen, the middle line names tabs as `N Name` with no brackets, the open tab is never dropped and is the only name in the bright accent, and every pane starts on the line right under the box. List each width and tab checked.

**Interfaces:**
- Consumes: `topLine(w int, edge lipgloss.Style, segs []segment) string` (frame.go).
- Produces: `const barRows = 3`; `func (m Model) tabRows(width int) []string`. `tabBar(width int) string` now takes the inner width of the box.

- [x] **Step 1: Write the failing tests**

Add to `internal/tui/view_test.go`:

```go
// TestTabsSitInABoxOfTheirOwn reads the top three lines at several widths
// and on every tab, so no width and no open tab can break the box.
func TestTabsSitInABoxOfTheirOwn(t *testing.T) {
	t.Parallel()

	for _, w := range []int{40, 60, 120, 200} {
		for i := range topTabs {
			m := press(sized(newModel(t), w, 30), tabKey(i))
			lines := strings.Split(m.View(), "\n")
			top, mid, bottom := plain(lines[0]), plain(lines[1]), plain(lines[2])
			if !strings.HasPrefix(top, "┌") || !strings.HasSuffix(top, "┐") || strings.Trim(top, "┌─┐") != "" {
				t.Errorf("w=%d tab %d: top border %q", w, i, top)
			}
			if !strings.HasPrefix(mid, "│") || !strings.HasSuffix(mid, "│") {
				t.Errorf("w=%d tab %d: names line has no walls %q", w, i, mid)
			}
			if !strings.HasPrefix(bottom, "└") || !strings.HasSuffix(bottom, "┘") || strings.Trim(bottom, "└─┘") != "" {
				t.Errorf("w=%d tab %d: bottom border %q", w, i, bottom)
			}
			for k := range barRows {
				if got := lipgloss.Width(lines[k]); got != w {
					t.Errorf("w=%d tab %d line %d: %d cells wide", w, i, k, got)
				}
			}
			if strings.ContainsAny(mid, "[]") {
				t.Errorf("w=%d tab %d: names still wear brackets %q", w, i, mid)
			}
			if open := fmt.Sprintf("%d %s", i+1, topTabs[i].name); !strings.Contains(mid, open) {
				t.Errorf("w=%d: the open tab %q is missing from %q", w, open, mid)
			}
			if w >= 120 {
				for j, tab := range topTabs {
					if name := fmt.Sprintf("%d %s", j+1, tab.name); !strings.Contains(mid, name) {
						t.Errorf("w=%d: %q is missing from %q", w, name, mid)
					}
				}
			}
			g := m.geometry()
			first := g.full
			if g.wide {
				first = g.side[paneList]
			}
			if first.y != barRows {
				t.Errorf("w=%d tab %d: the first pane starts on line %d, want %d", w, i, first.y, barRows)
			}
		}
	}
}

// TestTheOpenTabIsTheOnlyBrightName checks the brush of every name, so the
// highlight can never sit on two tabs or on the wrong one.
func TestTheOpenTabIsTheOnlyBrightName(t *testing.T) {
	withColors(func() {
		for i := range topTabs {
			m := press(sized(newModel(t), 160, 40), tabKey(i))
			mid := strings.Split(m.View(), "\n")[1]
			bright := m.styles.accent.Bold(true)
			for j, tab := range topTabs {
				on := strings.Contains(mid, bright.Render(fmt.Sprintf("%d %s", j+1, tab.name)))
				if on != (j == i) {
					t.Errorf("tab %d open: %q bright = %v", i, tab.name, on)
				}
			}
		}
	})
}
```

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui -run 'TestTabsSitInABoxOfTheirOwn|TestTheOpenTabIsTheOnlyBrightName' -v`
Expected: FAIL, `undefined: barRows`.

- [x] **Step 3: Write the implementation**

In `internal/tui/frame.go`, above `geometry`:

```go
// barRows is how tall the tab box is: its top border, the names, and its
// bottom border.
const barRows = 3
```

In `geometry`, the boxes start under the tab box:

```go
	// The tab box takes the top lines and the status line the last one, so
	// the boxes share the rest. A terminal with fewer rows gets the rows it has.
	bodyH := max(0, m.height-barRows-1)
	panes := m.panes()
	g := geom{wide: m.width >= 60, leftW: clamp(m.width*3/10, 28, 48), side: make([]box, len(panes))}
	y := barRows
	for p, h := range m.leftHeights(bodyH) {
		g.side[p] = m.box(pane(p), 0, y, g.leftW, h)
		y += h
	}
	// A box of zero width is not on screen, which is how a narrow terminal
	// leaves all but the focused one out.
	g.detail = m.box(paneDetail, g.leftW, barRows, max(0, m.width-g.leftW), bodyH)
	if g.wide {
		return g
	}
	g.full = m.box(m.focus, 0, barRows, m.width, bodyH)
```

In `internal/tui/scroll.go` `boxOf`:

```go
	if b.h == 0 {
		return m.box(p, 0, barRows, m.width, max(0, m.height-barRows-1))
	}
```

In `internal/tui/view.go` `View`, replace the line that builds `lines`:

```go
	lines := append(m.tabRows(m.width), strings.Split(body, "\n")...)
```

Add under `View`:

```go
// tabRows draws the tabs in a box of their own, so they read as the top of
// the screen and not as one more line of text.
func (m Model) tabRows(width int) []string {
	if width < 2 {
		return []string{"", "", ""}
	}
	edge := m.styles.faint
	inner := width - 2
	return []string{
		topLine(width, edge, nil),
		edge.Render("│") + pad(m.tabBar(inner), inner) + edge.Render("│"),
		edge.Render("└" + strings.Repeat("─", inner) + "┘"),
	}
}
```

Change the doc comment of `tabBar` to say it draws the names line of the tab box, with the open tab in the accent. In `barLine`, a name is its key number and its name:

```go
	for i, t := range topTabs {
		if drop[i] {
			continue
		}
		name := fmt.Sprintf("%d %s", i+1, t.name)
		if i == m.top {
			names = append(names, m.styles.accent.Bold(true).Render(name))
			continue
		}
		names = append(names, m.styles.faint.Render(name))
	}
	return " " + strings.Join(names, "  ")
```

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/tui -run 'TestTabsSitInABoxOfTheirOwn|TestTheOpenTabIsTheOnlyBrightName' -v`
Expected: PASS.

- [x] **Step 5: Run the whole package and update tests that pin the one-line bar**

Run: `go test ./internal/tui`
The geometry tests in `model_test.go` (the `g.side`, `g.detail`, `a.side[0]`, `n.full` checks near :684, :688, :707, :724) pin `y 1` and heights for a one-line bar. Change them to `y barRows` and heights two lines shorter. The bar width tests in `view_test.go` (`widthOfBar` and the tests that call `tabBar`) pin `[name]` and the full width; change them to `N name` and the inner width `w-2`. Keep what each test checks.
Expected after the updates: `ok  github.com/iyay/acta/internal/tui`.

- [x] **Step 6: Commit**

```bash
gofmt -l internal/tui && go vet ./internal/tui
git add internal/tui/view.go internal/tui/frame.go internal/tui/scroll.go internal/tui/view_test.go internal/tui/model_test.go
git commit -m "feat(tui): draw the tabs in a box of their own"
```

---

### Task 2: List pane title says Open, or Tasks on Activities

**Files:**
- Modify: `internal/tui/sidebar.go` (`tabsOf` at :148)
- Test: `internal/tui/sidebar_test.go`, `internal/tui/model_test.go`

**verify:** No tab ever titles its top list pane `List`. Every tab with a kind says `Open`, and Activities says `Tasks`. List every tab checked.

**Interfaces:**
- Consumes: nothing new.
- Produces: `tabsOf(paneList)` returns `[]string{"Open"}` or `[]string{"Tasks"}`.

- [x] **Step 1: Write the failing test**

Add to `internal/tui/sidebar_test.go` (add `strings` to its imports if it is not there):

```go
// TestListPaneTitleSaysOpenOrTasks draws the top border of the list pane on
// every tab, so no tab can keep the old word.
func TestListPaneTitleSaysOpenOrTasks(t *testing.T) {
	t.Parallel()

	for i := range topTabs {
		m := press(sized(newModel(t), 160, 40), tabKey(i))
		want := "Open"
		if topTabs[i].kind == "" {
			want = "Tasks"
		}
		if got := m.tabsOf(paneList); len(got) != 1 || got[0] != want {
			t.Errorf("tab %s: title %q, want %q", topTabs[i].name, got, want)
		}
		top := plain(m.paneTop(paneList, m.geometry().side[paneList], m.edge(paneList)))
		if !strings.Contains(top, want) || strings.Contains(top, "List") {
			t.Errorf("tab %s: top border %q, want %q and no List", topTabs[i].name, top, want)
		}
	}
}
```

- [x] **Step 2: Run the test to see it fail**

Run: `go test ./internal/tui -run TestListPaneTitleSaysOpenOrTasks -v`
Expected: FAIL, `title ["List"], want "Open"`.

- [x] **Step 3: Write the implementation**

In `internal/tui/sidebar.go` `tabsOf`:

```go
	case paneList:
		// Activities lists tasks, not the open items of a kind.
		if topTabs[m.top].kind == "" {
			return []string{"Tasks"}
		}
		return []string{"Open"}
```

- [x] **Step 4: Run the test to see it pass, then the package**

Run: `go test ./internal/tui -run TestListPaneTitleSaysOpenOrTasks -v`
Expected: PASS.
Run: `go test ./internal/tui`
The click test in `model_test.go` that calls `drawnLetter(m, paneList, "List", "first")` (near :1067) pins the old word. Change `"List"` to `m.tabsOf(paneList)[0]`.
Expected after the update: `ok`.

- [x] **Step 5: Commit**

```bash
gofmt -l internal/tui && go vet ./internal/tui
git add internal/tui/sidebar.go internal/tui/sidebar_test.go internal/tui/model_test.go
git commit -m "feat(tui): title the list pane Open, or Tasks on Activities"
```

---

### Task 3: Sort word moves to the bottom border

**Files:**
- Modify: `internal/tui/frame.go` (`paneTop` at :92, the sort word block at :116)
- Modify: `internal/tui/view.go` (`paneView` foot at :242)
- Test: `internal/tui/frame_test.go`, `internal/tui/scroll_test.go` (`footOf` at :218)

**verify:** No list pane ever draws the sort word in its top border. The bottom border ends with `<word> · <count>` when it fits and with the count alone when it does not; the count is never dropped or cut while the word is still drawn. List the widths and both sort states checked.

**Interfaces:**
- Consumes: `itemCount(selected, total int) string` (scroll.go).
- Produces: `func (m Model) sortWord(p pane) string`; `func sortFoot(word, count string, inner int) string`.

- [ ] **Step 1: Write the failing tests**

Add to `internal/tui/frame_test.go` (add `lipgloss` to its imports if it is not there):

```go
// TestSortWordSitsBeforeTheCountInTheBottomBorder checks both sort states at
// several widths, top and bottom border both.
func TestSortWordSitsBeforeTheCountInTheBottomBorder(t *testing.T) {
	t.Parallel()

	for _, w := range []int{60, 100, 160} {
		for _, flip := range []bool{false, true} {
			m := press(sized(newModel(t), w, 40), tabKey(tabSpecs))
			if flip {
				m = press(m, "o")
			}
			word := m.sortWord(paneList)
			if want := map[bool]string{false: "oldest", true: "newest"}[flip]; word != want {
				t.Fatalf("w=%d flip=%v: word %q, want %q", w, flip, word, want)
			}
			top := plain(m.paneTop(paneList, m.geometry().side[paneList], m.edge(paneList)))
			if strings.Contains(top, "oldest") || strings.Contains(top, "newest") {
				t.Errorf("w=%d flip=%v: the top border still says the sort: %q", w, flip, top)
			}
			rows, sel, idx := m.slotOf(paneList)
			count := itemCount(cursorOf(rows, *sel, *idx)+1, len(rows))
			if got := bottomLine(t, m, paneList); !strings.HasSuffix(got, " "+word+" · "+count+" ┘") {
				t.Errorf("w=%d flip=%v: bottom border %q, want it to end with %q", w, flip, got, word+" · "+count)
			}
		}
	}
}

// TestNarrowBottomBorderDropsTheWordFirst walks every width from too narrow
// for the word up to room for both, so the count never goes first.
func TestNarrowBottomBorderDropsTheWordFirst(t *testing.T) {
	t.Parallel()

	count := "12 of 34"
	both := "newest · " + count
	for inner := 0; inner < lipgloss.Width(" "+both+" "); inner++ {
		if got := sortFoot("newest", count, inner); got != count {
			t.Errorf("inner=%d: %q, want the count alone", inner, got)
		}
	}
	if got := sortFoot("newest", count, lipgloss.Width(" "+both+" ")); got != both {
		t.Errorf("room for both: %q, want %q", got, both)
	}
}
```

- [ ] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui -run 'TestSortWordSitsBeforeTheCountInTheBottomBorder|TestNarrowBottomBorderDropsTheWordFirst' -v`
Expected: FAIL, `m.sortWord undefined`.

- [ ] **Step 3: Write the implementation**

In `internal/tui/frame.go`, delete the sort word block at the end of `paneTop` (the comment "The sort word goes last..." and the `if p != paneDetail { ... }` block). Add under `paneTop`:

```go
// sortWord is the order a list box shows its rows in.
func (m Model) sortWord(p pane) string {
	if m.newest[p] {
		return "newest"
	}
	return "oldest"
}

// sortFoot joins the sort word and the count for the bottom border. A box
// too narrow for both drops the word, because the count tells more.
func sortFoot(word, count string, inner int) string {
	both := word + " · " + count
	if lipgloss.Width(" "+both+" ") <= inner {
		return both
	}
	return count
}
```

In `internal/tui/view.go` `paneView`, the foot of a list pane:

```go
	if p != paneDetail {
		rows, sel, idx := m.slotOf(p)
		foot = sortFoot(m.sortWord(p), itemCount(cursorOf(rows, *sel, *idx)+1, len(rows)), inner)
	}
```

In `internal/tui/scroll_test.go`, `footOf` gives only the count, so the counter tests keep checking what they checked:

```go
// footOf gives the counter a pane writes in its bottom border, with the
// border and the sort word cut off, and nothing at all on the detail box.
func footOf(t *testing.T, m Model, p pane) string {
	t.Helper()
	foot := strings.Trim(bottomLine(t, m, p), "─└┘ ")
	if _, count, ok := strings.Cut(foot, " · "); ok {
		return count
	}
	return foot
}
```

- [ ] **Step 4: Run the tests to see them pass, then the package**

Run: `go test ./internal/tui -run 'TestSortWordSitsBeforeTheCountInTheBottomBorder|TestNarrowBottomBorderDropsTheWordFirst|TestTitleShowsTheSortOfEachPane' -v`
Expected: PASS.
Run: `go test ./internal/tui`
Expected: `ok`. A test that still reads the sort word from the top border is updated to read the bottom border.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal/tui && go vet ./internal/tui
git add internal/tui/frame.go internal/tui/view.go internal/tui/frame_test.go internal/tui/scroll_test.go
git commit -m "feat(tui): move the sort word next to the count"
```

---

### Task 4: No Detail key

**Files:**
- Modify: `internal/tui/sidebar.go` (`paneKey` at :65)
- Modify: `internal/tui/model.go` (`key` at :348, the `"0"` case at :367)
- Modify: `internal/tui/view.go` (`helpLines` at :59)
- Test: `internal/tui/model_test.go`, `internal/tui/scroll_test.go` (`keyTo` at :74, `focusKeyFrom` at :86), `internal/tui/view_test.go`, `internal/tui/order_test.go`, `internal/tui/sidebar_test.go`

**verify:** The `0` key changes nothing on any tab or pane, no title and no help line names `0`, and the detail is still reached from every list pane with `tab` or `shift+tab`, and with `enter` on a row that is not a tree row. List every tab and pane checked.

**Interfaces:**
- Consumes: `cyclePane(step int)` (sidebar.go).
- Produces: `paneKey` returns `"─"` for every pane.

- [ ] **Step 1: Write the failing tests**

Add to `internal/tui/model_test.go`:

```go
// TestZeroNoLongerFocusesTheDetail presses 0 from every list pane of every
// tab, and checks the ring still reaches the detail.
func TestZeroNoLongerFocusesTheDetail(t *testing.T) {
	t.Parallel()

	for i := range topTabs {
		base := press(newModel(t), tabKey(i))
		for _, p := range base.panes() {
			m := base
			m.focusPane(p)
			if got := press(m, "0"); got.focus != p || got.top != m.top {
				t.Errorf("tab %s pane %d: 0 moved the focus to %d", topTabs[i].name, p, got.focus)
			}
			reached := false
			for _, k := range []string{"tab", "shift+tab"} {
				if press(m, k).focus == paneDetail {
					reached = true
				}
			}
			if !reached {
				t.Errorf("tab %s pane %d: neither tab nor shift+tab reaches the detail", topTabs[i].name, p)
			}
		}
	}
}

// TestNoTitleOrHelpNamesKeyZero reads every tab's screen and the help text.
func TestNoTitleOrHelpNamesKeyZero(t *testing.T) {
	t.Parallel()

	for i := range topTabs {
		if v := plain(press(sized(newModel(t), 160, 40), tabKey(i)).View()); strings.Contains(v, "[0]") {
			t.Errorf("tab %s still draws [0]", topTabs[i].name)
		}
	}
	for _, ln := range strings.Split(helpLines, "\n") {
		if strings.HasPrefix(ln, "0 ") {
			t.Errorf("help still lists 0: %q", ln)
		}
	}
}
```

- [ ] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui -run 'TestZeroNoLongerFocusesTheDetail|TestNoTitleOrHelpNamesKeyZero' -v`
Expected: FAIL, `0 moved the focus to 2` and `still draws [0]`.

- [ ] **Step 3: Write the implementation**

In `internal/tui/sidebar.go`:

```go
// paneKey is what a box wears at the start of its title. No box has a key
// of its own any more: tab and enter reach the detail.
func paneKey(pane) string { return "─" }
```

In `internal/tui/model.go` `key`, delete:

```go
	case "0":
		m.focusPane(paneDetail)
```

In `internal/tui/view.go` `helpLines`, delete the line `0                focus the detail`.

- [ ] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/tui -run 'TestZeroNoLongerFocusesTheDetail|TestNoTitleOrHelpNamesKeyZero' -v`
Expected: PASS.

- [ ] **Step 5: Move the old tests off the 0 key**

In `internal/tui/scroll_test.go`, `keyTo(paneDetail)` returns `"shift+tab"` (from the list pane, the ring goes back to the detail). In `focusKeyFrom`, the detail case becomes:

```go
	if to == paneDetail {
		// The ring is List, Done, Detail: the detail is one step back from
		// the list and one step on from Done.
		if from == paneDone {
			return "tab"
		}
		return "shift+tab"
	}
```

Run: `go test ./internal/tui`
Each test that still presses `"0"` to reach the detail (in `model_test.go` near :386, :415, :781, :786, :1657, and the ones in `view_test.go`, `order_test.go`, `sidebar_test.go`) now fails. Replace each `"0"` press with the key that reaches the detail from the pane that has the focus there: `"shift+tab"` from the list pane, `"tab"` from the Done pane. Leave any `"0"` that is not a key press alone.
Expected after the updates: `ok`.

- [ ] **Step 6: Commit**

```bash
gofmt -l internal/tui && go vet ./internal/tui
git add internal/tui/sidebar.go internal/tui/model.go internal/tui/view.go internal/tui/model_test.go internal/tui/scroll_test.go internal/tui/view_test.go internal/tui/order_test.go internal/tui/sidebar_test.go
git commit -m "feat(tui): drop the 0 key and the [0] mark on Detail"
```

---

### Task 5: Sticky Detail header and date footer

**Files:**
- Modify: `internal/tui/detail.go` (`detailLines` at :25)
- Modify: `internal/tui/view.go` (`paneView` at :220, `detailView` at :297)
- Modify: `internal/tui/scroll.go` (`fitOf` at :66, `linesAt` at :80)
- Test: `internal/tui/detail_test.go`

**verify:** Whenever the detail box has room for the header, the footer and 3 middle lines, every scroll offset keeps the header lines on top and the one date line at the bottom, and only the middle moves; whenever it has less room, the whole block scrolls as one. The header never holds a date label, the footer always names all three dates with `-` for each one not set, and scrolling can never go past the last middle line. List the heights and offsets checked.

**Interfaces:**
- Consumes: `truncate`, `expandTabs`, `m.workLines`, `m.render` (existing).
- Produces: `func (m Model) detailParts(w int) (head, mid []string, foot string)`; `func dateLine(it *board.Item) string`; `func stickyMid(head, h int) int`; `func (m Model) detailScroll(w, h int) (total, fit int)`; `func window(lines []string, first, h int) []string`. `detailLines(w int) []string` keeps its signature and returns head, middle and footer as one block.

- [ ] **Step 1: Write the failing tests**

Add to `internal/tui/detail_test.go`:

```go
// stickyModel is a spec with a long body, shown in a detail box of a screen
// h lines tall, with the focus on the detail.
func stickyModel(t *testing.T, h int) Model {
	t.Helper()
	var body strings.Builder
	for i := range 60 {
		fmt.Fprintf(&body, "line %02d\n\n", i)
	}
	cfg := treeCfg(t, map[string]string{
		".acta/specs/2026-09-20-long.md": "---\nid: SPC-0001\ncreated: \"2026-09-20\"\nstarted: \"2026-09-21\"\n---\n# Long spec\n\n" + body.String(),
	})
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	return onItem(t, sized(m, 120, h), "specs/2026-09-20-long")
}

func TestDetailHeaderAndFooterStayWhileTheMiddleScrolls(t *testing.T) {
	t.Parallel()

	m := stickyModel(t, 40)
	b := m.geometry().detail
	head, mid, foot := m.detailParts(b.textW())
	want := "created 2026-09-20 · started 2026-09-21 · finished -"
	if plain(foot) != want {
		t.Fatalf("footer %q, want %q", plain(foot), want)
	}
	for _, l := range labelsOf(head) {
		if l == "CREATED" || l == "STARTED" || l == "FINISHED" {
			t.Errorf("the header still holds %s", l)
		}
	}
	n := stickyMid(len(head), b.inner)
	if n < 3 {
		t.Fatalf("a 40 line screen should stick, got %d middle lines", n)
	}
	for _, off := range []int{0, 1, 5, len(mid) - n, 1000} {
		s := m
		s.off[paneDetail] = 0
		s.scrollPane(paneDetail, off)
		lines := innerLines(s, b)
		for i, h := range head {
			if got := lines[i]; got != strings.TrimRight(plain(h), " ") {
				t.Errorf("off=%d header line %d: %q, want %q", off, i, got, plain(h))
			}
		}
		if got := lines[len(lines)-1]; got != want {
			t.Errorf("off=%d last line %q, want the footer", off, got)
		}
		first := min(off, max(0, len(mid)-n))
		if got := lines[len(head)]; got != strings.TrimRight(plain(mid[first]), " ") {
			t.Errorf("off=%d first middle line %q, want %q", off, got, plain(mid[first]))
		}
	}
}

func TestShortDetailScrollsAsOneBlock(t *testing.T) {
	t.Parallel()

	checked := 0
	for h := 8; h < 40; h++ {
		m := stickyModel(t, h)
		b := m.geometry().detail
		head, _, _ := m.detailParts(b.textW())
		if b.inner < 1 || stickyMid(len(head), b.inner) > 0 {
			continue
		}
		checked++
		all := m.detailLines(b.textW())
		m.scrollPane(paneDetail, 1)
		if got := innerLines(m, b); len(got) == 0 || got[0] != strings.TrimRight(plain(all[1]), " ") {
			t.Errorf("h=%d: after one line the box starts %q, want %q", h, got, plain(all[1]))
		}
	}
	if checked == 0 {
		t.Fatal("no screen height was too short to stick")
	}
}

func TestStickyMidNeedsThreeMiddleLines(t *testing.T) {
	t.Parallel()

	for _, c := range []struct{ head, h, want int }{
		{5, 9, 3}, {5, 8, 0}, {0, 4, 3}, {0, 3, 0}, {5, 0, 0}, {5, 30, 24},
	} {
		if got := stickyMid(c.head, c.h); got != c.want {
			t.Errorf("stickyMid(%d, %d) = %d, want %d", c.head, c.h, got, c.want)
		}
	}
}
```

Replace `TestDetailShowsTheDates` in `internal/tui/detail_test.go`, because the dates now live in the footer:

```go
// Every item ends with one date line that names all three dates, a dash for
// each one not set, and the header holds no date label at all.
func TestDetailShowsTheDates(t *testing.T) {
	cfg := treeCfg(t, datedFiles())
	for _, c := range []struct{ id, foot string }{
		{"SCRATCH-1", "created 2026-09-01 · started 2026-09-02 · finished 2026-09-03"},
		{"SCRATCH-2", "created - · started - · finished -"},
		{"SCRATCH-3", "created 2026-09-01 · started - · finished -"},
		{"SCRATCH-4", "created - · started 2026-09-02 · finished -"},
		{"SCRATCH-5", "created - · started - · finished 2026-09-03"},
		{"SCRATCH-6", "created - · started - · finished -"},
	} {
		lines := detailLines(t, cfg, c.id)
		if got := plain(lines[len(lines)-1]); got != c.foot {
			t.Errorf("%s footer %q, want %q", c.id, got, c.foot)
		}
		for _, l := range labelsOf(lines) {
			if l == "CREATED" || l == "STARTED" || l == "FINISHED" {
				t.Errorf("%s header still holds %s", c.id, l)
			}
		}
	}
}
```

- [ ] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui -run 'TestDetailHeaderAndFooterStayWhileTheMiddleScrolls|TestShortDetailScrollsAsOneBlock|TestStickyMidNeedsThreeMiddleLines|TestDetailShowsTheDates' -v`
Expected: FAIL, `m.detailParts undefined`.

- [ ] **Step 3: Write the implementation**

In `internal/tui/detail.go`, `detailLines` becomes `detailParts`, and a new `detailLines` joins the parts:

```go
// detailParts splits the detail in three: the header that stays on top, the
// part that scrolls, and the date line that stays at the bottom. Labels are
// upper case, padded to one width, with the colons in one column, and a line
// with no value is left out. With no item on show there is only a message,
// so there is no header and no footer.
func (m Model) detailParts(w int) (head, mid []string, foot string) {
	it := m.Selected()
	if it == nil {
		if len(m.board.Items) == 0 {
			return nil, cut("this repo has no .acta/ yet.\n\npress n to write the first bug, or let the agent plugin create specs and plans.", w), ""
		}
		if len(m.listOf()) == 0 {
			return nil, []string{m.styles.faint.Render("No items")}, ""
		}
		return nil, []string{m.styles.faint.Render("enter opens the group")}, ""
	}
	fields := []struct{ label, value string }{
		{"ID", idText(it)},
		{kindLabel(it.Kind), it.Title},
		{"STATUS", it.Status},
		{"AUTHOR", it.Author},
		{"FROM", m.fromText(it)},
		{"REF", it.Ref},
		{"SPEC", m.specText(it)},
		linkField("CLOSES", m.closesText(it.Closes)),
		linkField("CLOSED BY", m.closesText(it.ClosedBy)),
		{"WORKTREE", worktreeText(it)},
		{"AGENT", it.Agent},
		{"FILE", m.fileText(it)},
		{tasksLabel(it.Kind), progressText(it)},
		{"FIXED", it.FixedIn},
	}
	width := 0
	for _, f := range fields {
		width = max(width, len(f.label))
	}
	width += 2
	for _, f := range fields {
		if f.value == "" {
			continue
		}
		head = append(head, truncate(expandTabs(fmt.Sprintf("%-*s: %s", width, f.label, f.value)), w))
	}
	head = append(head, m.styles.faint.Render(strings.Repeat("─", max(1, w))))
	for _, p := range it.Problems {
		mid = append(mid, truncate(expandTabs("! "+p), w))
	}
	mid = append(mid, m.workLines(it, w)...)
	for _, ln := range strings.Split(m.render(expandTabs(it.Body), w), "\n") {
		mid = append(mid, fit(ln, w))
	}
	return head, mid, truncate(dateLine(it), w)
}

// detailLines is the whole detail as one block: the header, the middle and
// the date line. A box too short to keep the header on top scrolls this.
func (m Model) detailLines(w int) []string {
	head, mid, foot := m.detailParts(w)
	out := append(append([]string(nil), head...), mid...)
	if foot != "" {
		out = append(out, foot)
	}
	return out
}

// dateLine is the footer of the detail. It names all three dates, and a dash
// stands in for a date that is not set, so the line keeps one shape.
func dateLine(it *board.Item) string {
	or := func(s string) string {
		if s == "" {
			return "-"
		}
		return s
	}
	return "created " + or(it.Created) + " · started " + or(it.StartedOn) + " · finished " + or(it.Finished)
}

// stickyMid is how many middle lines a detail box h lines tall shows under a
// header of head lines and over the footer. It is 0 when the box has no room
// for the header, the footer and 3 middle lines: then the whole detail
// scrolls as one block, so a short box still shows everything.
func stickyMid(head, h int) int {
	if n := h - head - 1; n >= 3 {
		return n
	}
	return 0
}

// detailScroll says how many lines the detail scrolls over and how many of
// them the box shows at once: the middle alone when the header and the
// footer stick, the whole block when they do not.
func (m Model) detailScroll(w, h int) (total, fit int) {
	head, mid, foot := m.detailParts(w)
	if n := stickyMid(len(head), h); n > 0 && foot != "" {
		return len(mid), n
	}
	return len(m.detailLines(w)), h
}
```

Update the doc comment above the dots constants only if it names `detailLines`; leave it otherwise.

In `internal/tui/view.go`, `detailView`:

```go
// detailView gives the lines of the detail box. With room to spare the
// header stays on top and the date line at the bottom, and only the middle
// scrolls from first. Without that room the whole block scrolls from first.
func (m Model) detailView(w, first, h int) []string {
	head, mid, foot := m.detailParts(w)
	n := stickyMid(len(head), h)
	if n == 0 || foot == "" {
		return window(m.detailLines(w), first, h)
	}
	out := append(append([]string(nil), head...), window(mid, first, n)...)
	for len(out) < h-1 {
		out = append(out, "")
	}
	return append(out, foot)
}

// window gives at most h lines of lines, from first on.
func window(lines []string, first, h int) []string {
	if first >= len(lines) {
		return nil
	}
	lines = lines[first:]
	if len(lines) > h {
		lines = lines[:h]
	}
	return lines
}
```

In `paneView`, the detail box counts its scroll with `detailScroll`, so the offset and the scrollbar match what the box shows:

```go
	inner := b.textW()
	total, fit := m.linesAt(p, inner), b.inner
	first := b.first
	if p == paneDetail {
		total, fit = m.detailScroll(inner, b.inner)
		first = firstOf(m.off[p], total, fit)
	}
	// A thumb marks the window on the right wall, in the brush the wall
	// already wears, so the focus reads the same on the border as inside.
	bar := scrollbar(total, fit, first, b.inner)
```

In `internal/tui/scroll.go`:

```go
// fitOf is how many lines pane p has room for right now. The detail box
// counts only its middle when its header and footer stick.
func (m Model) fitOf(p pane) int {
	b := m.boxOf(p)
	if p == paneDetail {
		_, fit := m.detailScroll(b.textW(), b.inner)
		return fit
	}
	return b.rows
}
```

and in `linesAt`:

```go
	if p == paneDetail {
		total, _ := m.detailScroll(w, m.boxOf(p).inner)
		return total
	}
```

- [ ] **Step 4: Run the tests to see them pass, then the package**

Run: `go test ./internal/tui -run 'TestDetailHeaderAndFooterStayWhileTheMiddleScrolls|TestShortDetailScrollsAsOneBlock|TestStickyMidNeedsThreeMiddleLines|TestDetailShowsTheDates' -v`
Expected: PASS.
Run: `go test ./internal/tui`
A detail test that expected the problems above the rule line, or a date label in the header, is updated to the new order: header, rule, problems, work lines, body, date line.
Expected after the updates: `ok`.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal/tui && go vet ./internal/tui
git add internal/tui/detail.go internal/tui/view.go internal/tui/scroll.go internal/tui/detail_test.go
git commit -m "feat(tui): keep the detail header and date line in place"
```

---

### Task 6: Sort by short id

**Files:**
- Modify: `internal/tui/order.go` (`ordered` at :16)
- Test: `internal/tui/order_test.go`

**verify:** In both sort states, an item with a short id always sorts by its number, a tie on the number falls back to the file date and then the file, an item with no id or a broken id always goes last, tasks of one plan keep file order, and the input slice is never reordered. List every tie and edge case checked.

**Interfaces:**
- Consumes: `board.Item.ShortID`, `board.Item.Date`, `fileOf(id string) string`.
- Produces: `func idNum(it *board.Item) (int, bool)`. `ordered(items []*board.Item, newest bool) []*board.Item` keeps its signature.

- [x] **Step 1: Write the failing tests**

Add to `internal/tui/order_test.go`:

```go
func withID(id, short, date string) *board.Item {
	it := spec(id, date)
	it.ShortID = short
	return it
}

// The number wins over the date and over the file name, so items made on
// one day come out in the order they were made.
func TestOrderedSortsByIDNumberBothWays(t *testing.T) {
	t.Parallel()

	in := []*board.Item{
		withID("scratch/2026-09-29-a", "SCR-0018", "2026-09-29"),
		withID("scratch/2026-09-30-m", "SCR-0009", "2026-09-30"),
		withID("scratch/2026-09-29-z", "SCR-0017", "2026-09-29"),
	}
	want := []string{"scratch/2026-09-30-m", "scratch/2026-09-29-z", "scratch/2026-09-29-a"}
	if got := itemIDs(ordered(in, false)); !slices.Equal(got, want) {
		t.Fatalf("oldest: got %v, want %v", got, want)
	}
	slices.Reverse(want)
	if got := itemIDs(ordered(in, true)); !slices.Equal(got, want) {
		t.Fatalf("newest: got %v, want %v", got, want)
	}
	if in[0].ID != "scratch/2026-09-29-a" {
		t.Fatal("ordered sorted its input in place")
	}
}

func TestOrderedBreaksAnIDTieByDate(t *testing.T) {
	t.Parallel()

	in := []*board.Item{
		withID("specs/2026-09-02-b", "SPC-0005", "2026-09-02"),
		withID("specs/2026-09-01-a", "SPC-0005", "2026-09-01"),
	}
	if got, want := itemIDs(ordered(in, false)), []string{"specs/2026-09-01-a", "specs/2026-09-02-b"}; !slices.Equal(got, want) {
		t.Fatalf("oldest: got %v, want %v", got, want)
	}
	if got, want := itemIDs(ordered(in, true)), []string{"specs/2026-09-02-b", "specs/2026-09-01-a"}; !slices.Equal(got, want) {
		t.Fatalf("newest: got %v, want %v", got, want)
	}
}

func TestOrderedPutsItemsWithNoIDLast(t *testing.T) {
	t.Parallel()

	in := []*board.Item{
		withID("specs/2026-01-01-none", "", "2026-01-01"),
		withID("specs/2026-01-02-broken", "SPC-", "2026-01-02"),
		withID("specs/2026-09-01-word", "SPC-abc", "2026-09-01"),
		withID("specs/2026-09-20-b", "SPC-0020", "2026-09-20"),
		withID("specs/2026-09-10-a", "SPC-0010", "2026-09-10"),
	}
	for _, newest := range []bool{false, true} {
		got := itemIDs(ordered(in, newest))
		if !slices.Contains(got[:2], "specs/2026-09-20-b") || !slices.Contains(got[:2], "specs/2026-09-10-a") {
			t.Errorf("newest=%v: items with an id are not first: %v", newest, got)
		}
	}
}

// A task id carries its plan's number, so tasks sort with their plan and
// keep file order inside it.
func TestOrderedSortsTasksByPlanNumber(t *testing.T) {
	t.Parallel()

	task := func(id, short string) *board.Item {
		return &board.Item{ID: id, ShortID: short, Kind: board.KindTask, Date: "2026-09-29", Status: "in-progress"}
	}
	in := []*board.Item{
		task("plans/2026-09-29-a#task-1", "PLN-0033.01"),
		task("plans/2026-09-29-a#task-2", "PLN-0033.02"),
		task("plans/2026-09-29-z#task-1", "PLN-0002.01"),
	}
	want := []string{"plans/2026-09-29-z#task-1", "plans/2026-09-29-a#task-1", "plans/2026-09-29-a#task-2"}
	if got := itemIDs(ordered(in, false)); !slices.Equal(got, want) {
		t.Fatalf("oldest: got %v, want %v", got, want)
	}
}

func TestIDNumReadsTheNumber(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		short string
		n     int
		ok    bool
	}{
		{"SCR-0017", 17, true}, {"PLN-0033.01", 33, true}, {"BUG-10000", 10000, true},
		{"", 0, false}, {"SCR-", 0, false}, {"SCR-abc", 0, false}, {"SCR0017", 0, false},
	} {
		n, ok := idNum(&board.Item{ShortID: c.short})
		if n != c.n || ok != c.ok {
			t.Errorf("idNum(%q) = %d %v, want %d %v", c.short, n, ok, c.n, c.ok)
		}
	}
}
```

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui -run 'TestOrdered|TestIDNum' -v`
Expected: FAIL, `undefined: idNum`.

- [x] **Step 3: Write the implementation**

Replace `ordered` in `internal/tui/order.go` and add `idNum` (add `strconv` to the imports):

```go
// ordered gives items sorted by the number in their short id, oldest first,
// or newest first when newest is set. Ids are handed out in the order items
// are made, so this is the order they were made in. Two items with the same
// number sort by the date in their file name, then by the file they live in,
// so the list never jumps between loads. The sort is stable and every task
// of a plan lives in the same file, so tasks keep their file order. An item
// with no id goes last, and among those the date decides, the way it did
// before ids. It returns a new slice so the board keeps its own order.
func ordered(items []*board.Item, newest bool) []*board.Item {
	out := append([]*board.Item(nil), items...)
	sort.SliceStable(out, func(i, j int) bool {
		a, c := out[i], out[j]
		na, hasA := idNum(a)
		nc, hasC := idNum(c)
		if hasA != hasC {
			return hasA
		}
		if hasA && na != nc {
			return (na < nc) != newest
		}
		if (a.Date == "") != (c.Date == "") {
			return c.Date == ""
		}
		if a.Date != c.Date {
			return (a.Date < c.Date) != newest
		}
		fa, fc := fileOf(a.ID), fileOf(c.ID)
		if fa != fc {
			return (fa < fc) != newest
		}
		return false
	})
	return out
}

// idNum reads the number of a short id: 17 from SCR-0017, and 33 from the
// task id PLN-0033.01, so a task sorts with its plan. It says false when the
// item has no short id or the id holds no number.
func idNum(it *board.Item) (int, bool) {
	_, rest, ok := strings.Cut(it.ShortID, "-")
	if !ok {
		return 0, false
	}
	rest, _, _ = strings.Cut(rest, ".")
	n, err := strconv.Atoi(rest)
	return n, err == nil
}
```

- [x] **Step 4: Run the tests to see them pass, then the package**

Run: `go test ./internal/tui -run 'TestOrdered|TestIDNum|TestOFlips' -v`
Expected: PASS.
Run: `go test ./internal/tui`
Expected: `ok`. A list test whose fixture items carry ids and that expected date order is updated to id order.

- [x] **Step 5: Commit**

```bash
gofmt -l internal/tui && go vet ./internal/tui
git add internal/tui/order.go internal/tui/order_test.go
git commit -m "feat(tui): sort lists by the number in the short id"
```

---

### Task 7: Activities grouped by parent

**Files:**
- Modify: `internal/tui/sidebar.go` (`activityRows` at :188)
- Modify: `internal/tui/model.go` (`Model` fields at :67, `toggleRow` at :605)
- Modify: `internal/tui/scroll.go` (`treeMark` at :200)
- Test: `internal/tui/model_test.go`

**verify:** Activities lists every task in progress and nothing else, each one exactly once, right under its head. The head is the task's plan, or the bug when the plan's parent is a bug, and never goes more than one level up. Every group starts open, and `enter` on a head shuts and opens only that group. The Plans tab trees behave as before. List every head kind and every task state checked.

**Interfaces:**
- Consumes: `ordered`, `idNum` (Task 6); `board.Item.PlanID`, `board.Item.SpecID`.
- Produces: `Model.shutActs map[string]bool`; `func (m Model) headOf(t *board.Item) *board.Item`; `func (m Model) isOpen(id string) bool`; `func (m *Model) setOpen(id string, open bool)`; test helpers `activityFiles() map[string]string` and `actModel(t *testing.T) Model` in `model_test.go`.

- [ ] **Step 1: Write the failing tests**

Add to `internal/tui/model_test.go`:

```go
// activityFiles is a board with work under a spec plan, under a bug plan,
// and a plan with no work begun.
func activityFiles() map[string]string {
	return map[string]string{
		".acta/specs/2026-09-20-s.md": "---\nid: SPC-0001\n---\n# Spec S\n",
		".acta/bugs/2026-09-21-b.md":  "---\nid: BUG-0002\n---\n# Bug B\n",
		".acta/plans/2026-09-22-p.md": "---\nid: PLN-0003\nparent: specs/2026-09-20-s\n---\n# Plan P\n\n### Task 1: Going\n- [x] a\n- [ ] b\n\n### Task 2: Waiting\n- [ ] c\n\n### Task 3: Also going\n- [x] d\n- [ ] e\n",
		".acta/plans/2026-09-23-q.md": "---\nid: PLN-0004\nparent: bugs/2026-09-21-b\n---\n# Plan Q\n\n### Task 1: Fixing\n- [x] f\n- [ ] g\n",
		".acta/plans/2026-09-24-r.md": "---\nid: PLN-0005\n---\n# Plan R\n\n### Task 1: Not begun\n- [ ] h\n",
	}
}

// actModel opens Activities over activityFiles.
func actModel(t *testing.T) Model {
	t.Helper()
	cfg := treeCfg(t, activityFiles())
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	return press(sized(m, 160, 40), tabKey(tabActivities))
}

func TestActivitiesGroupsTasksUnderTheirParent(t *testing.T) {
	t.Parallel()

	m := actModel(t)
	want := []string{
		"bugs/2026-09-21-b",
		"plans/2026-09-23-q#task-1",
		"plans/2026-09-22-p",
		"plans/2026-09-22-p#task-1",
		"plans/2026-09-22-p#task-3",
	}
	rows := m.rowsOf(paneList)
	if got := ids(rows); !slices.Equal(got, want) {
		t.Fatalf("rows %q, want %q", got, want)
	}
	for _, r := range rows {
		it := m.board.Get(r.id)
		switch {
		case !r.tree:
			t.Errorf("%s is not a tree row", r.id)
		case r.depth == 0 && it.Kind == board.KindTask:
			t.Errorf("task %s sits at the top level", r.id)
		case r.depth == 1 && !inProgress(it):
			t.Errorf("%s is listed but not in progress", r.id)
		case r.depth > 1:
			t.Errorf("%s goes deeper than one level", r.id)
		}
	}
	if got := ids(press(m, "o").rowsOf(paneList)); got[0] != "plans/2026-09-22-p" {
		t.Errorf("newest first should put PLN-0003 on top, got %q", got)
	}
}

func TestEnterShutsOneActivitiesGroup(t *testing.T) {
	t.Parallel()

	m := press(actModel(t), "enter")
	want := []string{"bugs/2026-09-21-b", "plans/2026-09-22-p", "plans/2026-09-22-p#task-1", "plans/2026-09-22-p#task-3"}
	if got := ids(m.rowsOf(paneList)); !slices.Equal(got, want) {
		t.Fatalf("after enter on the bug: %q, want %q", got, want)
	}
	if m.focus == paneDetail {
		t.Error("enter on a head moved the focus to the detail")
	}
	if got := len(press(m, "enter").rowsOf(paneList)); got != 5 {
		t.Errorf("a second enter left %d rows, want 5", got)
	}
	// A task row is not a head: enter opens it in the detail.
	if got := press(actModel(t), "j", "enter"); got.focus != paneDetail {
		t.Error("enter on a task row did not open the detail")
	}
	// The Plans tab still starts with every plan shut.
	for _, r := range press(m, tabKey(tabPlans)).rowsOf(paneList) {
		if r.depth > 0 {
			t.Errorf("the Plans tab opened %s on its own", r.id)
		}
	}
}
```

- [ ] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui -run 'TestActivitiesGroupsTasksUnderTheirParent|TestEnterShutsOneActivitiesGroup' -v`
Expected: FAIL, the rows are a flat list of tasks.

- [ ] **Step 3: Write the implementation**

In `internal/tui/model.go`, add a field under `openPlans`:

```go
	shutActs  map[string]bool // the groups the reader shut on Activities; every group starts open
```

In `toggleRow`, replace the map block (from the comment "A new map each time" to `m.openPlans = open`) with:

```go
	m.setOpen(rows[i].id, !m.isOpen(rows[i].id))
```

In the same function, change the task-row branch so `enter` on a task opens the detail instead of doing nothing (user ruling 2026-09-29):

```go
	if rows[i].depth > 0 {
		return false
	}
```

Remove `"maps"` from the imports of `model.go` if nothing else uses it.

In `internal/tui/sidebar.go` (add `"maps"` to its imports), replace `activityRows` and add the helpers:

```go
// activityRows gives Activities the tasks under way, each under its head: the
// plan it belongs to, or the bug that plan fixes. Heads and the tasks under
// a head keep the order of the pane. A task whose plan is gone has no head,
// so it stands on its own row at the end.
func (m Model) activityRows() []row {
	if m.query != "" {
		return m.searchRows()
	}
	newest := m.newest[paneList]
	under := map[string][]*board.Item{}
	var heads, loose []*board.Item
	for _, t := range m.board.List(board.KindTask, true) {
		if !inProgress(t) {
			continue
		}
		h := m.headOf(t)
		if h == nil {
			loose = append(loose, t)
			continue
		}
		if _, seen := under[h.ID]; !seen {
			heads = append(heads, h)
		}
		under[h.ID] = append(under[h.ID], t)
	}
	var out []row
	for _, h := range ordered(heads, newest) {
		out = append(out, row{id: h.ID, tree: true})
		if !m.isOpen(h.ID) {
			continue
		}
		for _, t := range ordered(under[h.ID], newest) {
			out = append(out, row{id: t.ID, depth: 1, tree: true})
		}
	}
	return append(out, toRows(ordered(loose, newest), 0)...)
}

// headOf is the row a task under way sits under on Activities: its plan, or
// the bug that plan fixes. It goes one level up only, so a spec above a plan
// never becomes a head.
func (m Model) headOf(t *board.Item) *board.Item {
	plan := m.board.Get(t.PlanID)
	if plan == nil {
		return nil
	}
	if p := m.board.Get(plan.SpecID); p != nil && p.Kind == board.KindBug {
		return p
	}
	return plan
}

// isOpen says whether a tree head shows its tasks. Plans start shut and
// Activities groups start open, so each tab keeps the set that differs from
// how it starts.
func (m Model) isOpen(id string) bool {
	if topTabs[m.top].kind == "" {
		return !m.shutActs[id]
	}
	return m.openPlans[id]
}

// setOpen opens or shuts a tree head. It builds a new map each time, so an
// older copy of the model keeps the tree it drew.
func (m *Model) setOpen(id string, open bool) {
	acts := topTabs[m.top].kind == ""
	src := m.openPlans
	if acts {
		src = m.shutActs
	}
	next := make(map[string]bool, len(src)+1)
	maps.Copy(next, src)
	if acts {
		next[id] = !open
		m.shutActs = next
	} else {
		next[id] = open
		m.openPlans = next
	}
	m.keepVisible(m.listPane())
}
```

In `internal/tui/scroll.go` `treeMark`, read the open state through the tab:

```go
	if m.isOpen(it.ID) {
		return "-"
	}
	return "+"
```

- [ ] **Step 4: Run the tests to see them pass, then the package**

Run: `go test ./internal/tui -run 'TestActivitiesGroupsTasksUnderTheirParent|TestEnterShutsOneActivitiesGroup|TestPlanRows' -v`
Expected: PASS.
Run: `go test ./internal/tui`
The Activities branch of `TestEveryTabHoldsItsOpenItemsInFileDateOrder` (near :1347) expects a flat list. Change it to collect the rows with `r.depth == 1 || !r.tree` and compare them, sorted, with the in-progress task ids, sorted.
Expected after the update: `ok`.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal/tui && go vet ./internal/tui
git add internal/tui/sidebar.go internal/tui/model.go internal/tui/scroll.go internal/tui/model_test.go
git commit -m "feat(tui): group Activities tasks under their plan or bug"
```

---

### Task 8: Fold with h and l

**Files:**
- Modify: `internal/tui/model.go` (`key` at :348, new `foldRow`)
- Modify: `internal/tui/view.go` (`helpLines` at :59)
- Test: `internal/tui/plantree_test.go`, `internal/tui/model_test.go`

**verify:** On every tree list (Plans open, Plans done, Activities), `h` on a task row shuts its head and leaves the cursor on the head, `h` on a head shuts it, `l` on a head opens it, and no other row, pane or tab reacts to `h` or `l`. `enter` and the arrow keys keep what they do today. List every list and row kind checked.

**Interfaces:**
- Consumes: `isOpen`, `setOpen`, `actModel` (Task 7).
- Produces: `func (m *Model) foldRow(open bool)`.

- [ ] **Step 1: Write the failing tests**

Add to `internal/tui/plantree_test.go`:

```go
func TestHAndLFoldThePlanTree(t *testing.T) {
	t.Parallel()

	shut := []string{"plans/2026-09-21-alpha", "plans/2026-09-23-lonely"}
	open := []string{"plans/2026-09-21-alpha", "plans/2026-09-21-alpha#task-1", "plans/2026-09-21-alpha#task-2", "plans/2026-09-23-lonely"}
	m := press(toPlans(planModel(t)), "l")
	if got := rowIDs(m); !slices.Equal(got, open) {
		t.Fatalf("l on a shut plan: %q, want %q", got, open)
	}
	if got := rowIDs(press(m, "l")); !slices.Equal(got, open) {
		t.Errorf("l on an open plan changed the list: %q", got)
	}
	onTask := press(m, "j", "j")
	if got := rowIDs(press(onTask, "l")); !slices.Equal(got, open) {
		t.Errorf("l on a task row changed the list: %q", got)
	}
	back := press(onTask, "h")
	if got := rowIDs(back); !slices.Equal(got, shut) {
		t.Errorf("h on a task row: %q, want %q", got, shut)
	}
	if it := back.Selected(); it == nil || it.ID != "plans/2026-09-21-alpha" {
		t.Errorf("h on a task row left the cursor on %v, want the plan", it)
	}
	if got := rowIDs(press(m, "h")); !slices.Equal(got, shut) {
		t.Errorf("h on an open plan: %q, want %q", got, shut)
	}
	if got := rowIDs(press(m, "h", "h")); !slices.Equal(got, shut) {
		t.Errorf("h on a shut plan opened it: %q", got)
	}
	if got := press(m, "right").top; got != (tabPlans+1)%len(topTabs) {
		t.Errorf("right no longer switches tabs, top %d", got)
	}
}
```

Add to `internal/tui/model_test.go`:

```go
func TestHAndLFoldActivitiesGroups(t *testing.T) {
	t.Parallel()

	m := press(actModel(t), "j") // the task under the bug
	back := press(m, "h")
	want := []string{"bugs/2026-09-21-b", "plans/2026-09-22-p", "plans/2026-09-22-p#task-1", "plans/2026-09-22-p#task-3"}
	if got := ids(back.rowsOf(paneList)); !slices.Equal(got, want) {
		t.Fatalf("h on a task row: %q, want %q", got, want)
	}
	if it := back.Selected(); it == nil || it.ID != "bugs/2026-09-21-b" {
		t.Errorf("h left the cursor on %v, want the bug", it)
	}
	if got := len(press(back, "l").rowsOf(paneList)); got != 5 {
		t.Errorf("l on the shut bug left %d rows, want 5", got)
	}
	// A list that is not a tree does not react.
	bugs := press(newModel(t), tabKey(tabBugs))
	for _, k := range []string{"h", "l"} {
		if got := press(bugs, k); !slices.Equal(rowIDs(got), rowIDs(bugs)) || got.cursor() != bugs.cursor() {
			t.Errorf("%s changed the Bugs list", k)
		}
	}
	if !strings.Contains(helpLines, "h l") {
		t.Errorf("help does not list h l: %q", helpLines)
	}
}
```

- [ ] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui -run 'TestHAndLFold' -v`
Expected: FAIL, `l on a shut plan` leaves the list shut.

- [ ] **Step 3: Write the implementation**

In `internal/tui/model.go` `key`, next to `case " ":`:

```go
	case "h":
		m.foldRow(false)
	case "l":
		m.foldRow(true)
```

Add under `toggleRow`:

```go
// foldRow is h and l on a tree list. h on a task row shuts its head and puts
// the cursor on the head, h on a head shuts it, and l on a head opens it. l
// on a task row, and any row that is not a tree row, stay as they are.
func (m *Model) foldRow(open bool) {
	if m.focus == paneDetail {
		return
	}
	rows := m.listOf()
	i := m.cursor()
	if i < 0 || !rows[i].tree {
		return
	}
	if rows[i].depth > 0 {
		if open {
			return
		}
		// The head sits above its tasks, so walk up to it.
		for i > 0 && rows[i].depth > 0 {
			i--
		}
	}
	m.setOpen(rows[i].id, open)
	m.moveTo(i)
}
```

In `internal/tui/view.go` `helpLines`, under the `space enter` line add:

```
h l              shut / open a plan row, h on a task too
```

- [ ] **Step 4: Run the tests to see them pass, then the package**

Run: `go test ./internal/tui -run 'TestHAndLFold' -v`
Expected: PASS.
Run: `go test ./internal/tui`
Expected: `ok`.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal/tui && go vet ./internal/tui
git add internal/tui/model.go internal/tui/view.go internal/tui/plantree_test.go internal/tui/model_test.go
git commit -m "feat(tui): fold tree rows with h and l"
```

---

### Task 9: Copy the id with y

**Files:**
- Modify: `internal/tui/model.go` (`Model` fields at :67, `New` at :103, `key` at :348, new `copyID`, `copyRef`)
- Modify: `internal/tui/title.go` (next to `defaultOpen` at :128)
- Modify: `internal/tui/view.go` (`helpLines` at :59)
- Test: `internal/tui/model_test.go`

**verify:** On every row kind (spec, bug, plan head, task, item with no short id, group row, empty list), `y` hands the clipboard exactly the id the spec names (`PLN-n#task-m` for a task) or nothing at all, and the status line always tells the outcome: `copied <id>`, the error when the copy fails, or `nothing selected`. The OSC 52 text always decodes back to the id. List every row kind checked.

**Interfaces:**
- Consumes: `actModel` (Task 7); `shortRef(it *board.Item) string` (view.go).
- Produces: `Model.clip func(text string) error`; `func (m *Model) copyID()`; `func copyRef(b *board.Board, it *board.Item) string`; `func osc52(text string) string`; `func copyText(text string) error`.

- [ ] **Step 1: Write the failing tests**

Add to `internal/tui/model_test.go` (add `"encoding/base64"` to the imports):

```go
func TestYCopiesTheShortIDOfTheRow(t *testing.T) {
	t.Parallel()

	var got []string
	m := actModel(t)
	m.clip = func(s string) error { got = append(got, s); return nil }
	for _, c := range []struct {
		keys []string
		want string
	}{
		{nil, "BUG-0002"},
		{[]string{"j"}, "PLN-0004#task-1"},
		{[]string{"j", "j"}, "PLN-0003"},
	} {
		got = nil
		after := press(press(m, c.keys...), "y")
		if len(got) != 1 || got[0] != c.want {
			t.Errorf("keys %v: copied %q, want %q", c.keys, got, c.want)
		}
		if after.status != "copied "+c.want {
			t.Errorf("keys %v: status %q", c.keys, after.status)
		}
		if !strings.Contains(plain(after.View()), "copied "+c.want) {
			t.Errorf("keys %v: the status line does not show the copy", c.keys)
		}
	}
}

func TestYSaysWhyTheCopyFailed(t *testing.T) {
	t.Parallel()

	m := actModel(t)
	m.clip = func(string) error { return errors.New("no clipboard") }
	if got := press(m, "y").status; got != "copy failed: no clipboard" {
		t.Errorf("status %q", got)
	}
	called := false
	empty := press(newModel(t), "/", "zzzz-no-such-item", "enter")
	empty.clip = func(string) error { called = true; return nil }
	if got := press(empty, "y").status; got != "nothing selected" || called {
		t.Errorf("empty list: status %q, clip called %v", got, called)
	}
}

func TestCopyRefFallsBackWhenThereIsNoNumber(t *testing.T) {
	t.Parallel()

	plan := &board.Item{ID: "plans/2026-09-29-x", Kind: board.KindPlan}
	task := &board.Item{ID: "plans/2026-09-29-x#task-2", Kind: board.KindTask, PlanID: plan.ID, TaskNum: "2"}
	b := &board.Board{Items: []*board.Item{plan, task}}
	if got := copyRef(b, task); got != task.ID {
		t.Errorf("task of a plan with no id: %q, want %q", got, task.ID)
	}
	if got := copyRef(b, &board.Item{ID: "specs/x", Hash: "SPC-k3f2"}); got != "SPC-k3f2" {
		t.Errorf("item with a hash and no id: %q", got)
	}
	if got := copyRef(b, &board.Item{ID: "specs/y"}); got != "specs/y" {
		t.Errorf("item with nothing: %q", got)
	}
}

func TestOSC52CarriesTheTextWhole(t *testing.T) {
	t.Parallel()

	for _, text := range []string{"PLN-0003#task-1", "", "SPC-ünïcode", strings.Repeat("x", 500)} {
		s := osc52(text)
		inner, ok := strings.CutPrefix(s, "\x1b]52;c;")
		inner, ok2 := strings.CutSuffix(inner, "\a")
		if !ok || !ok2 {
			t.Fatalf("%q: not an OSC 52 code: %q", text, s)
		}
		raw, err := base64.StdEncoding.DecodeString(inner)
		if err != nil || string(raw) != text {
			t.Errorf("%q: decodes to %q, %v", text, raw, err)
		}
	}
}
```

If `board.Board` cannot be built from `Items` alone for `Get`, build `b` with `board.Load(treeCfg(t, ...))` over a plan file that has no `id:` line instead; the check stays the same.

- [ ] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui -run 'TestYCopies|TestYSays|TestCopyRef|TestOSC52' -v`
Expected: FAIL, `m.clip undefined`.

- [ ] **Step 3: Write the implementation**

In `internal/tui/model.go`, add a field under `open`:

```go
	clip      func(text string) error // puts text on the clipboard
```

In `New`, next to `open: defaultOpen,`:

```go
		clip:     copyText,
```

In `key`, next to `case "e":`:

```go
	case "y":
		m.copyID()
```

Add under `openPopup`:

```go
// copyID puts the id of the row under the cursor on the clipboard, and the
// status line says what was copied or why it could not be.
func (m *Model) copyID() {
	it := m.Selected()
	if it == nil {
		m.status = "nothing selected"
		return
	}
	id := copyRef(m.board, it)
	if err := m.clip(id); err != nil {
		m.status = "copy failed: " + err.Error()
		return
	}
	m.status = "copied " + id
}

// copyRef is the id y copies. A task is named PLN-n#task-m, the way acta
// tick takes it; a task whose plan has no number yet gives its path id.
func copyRef(b *board.Board, it *board.Item) string {
	if it.Kind == board.KindTask {
		if p := b.Get(it.PlanID); p != nil && p.ShortID != "" {
			return p.ShortID + "#task-" + it.TaskNum
		}
		return it.ID
	}
	return shortRef(it)
}
```

In `internal/tui/title.go` (add `"encoding/base64"`, `"errors"`, `"os"` and `"strings"` to the imports), under `defaultOpen`:

```go
// osc52 asks the terminal to put text on the clipboard. It travels with the
// screen output, so it works over SSH, and tmux passes it on.
func osc52(text string) string {
	return "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte(text)) + "\a"
}

// copyText puts text on the clipboard. pbcopy goes first, because OSC 52
// is written straight to the terminal while Bubble Tea draws and can smear
// the screen for a moment. OSC 52 is only the fallback when pbcopy is
// missing or fails.
func copyText(text string) error {
	path, err := exec.LookPath("pbcopy")
	if err == nil {
		cmd := exec.Command(path)
		cmd.Stdin = strings.NewReader(text)
		if err = cmd.Run(); err == nil {
			return nil
		}
	}
	if _, oscErr := os.Stdout.WriteString(osc52(text)); oscErr != nil {
		return errors.Join(err, oscErr)
	}
	return nil
}
```

In `internal/tui/view.go` `helpLines`, under the `e` line add:

```
y                copy the id of the row
```

- [ ] **Step 4: Run the tests to see them pass, then the package**

Run: `go test ./internal/tui -run 'TestYCopies|TestYSays|TestCopyRef|TestOSC52' -v`
Expected: PASS.
Run: `go test ./internal/tui`
Expected: `ok`.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal/tui && go vet ./internal/tui
git add internal/tui/model.go internal/tui/title.go internal/tui/view.go internal/tui/model_test.go
git commit -m "feat(tui): copy the row id with y"
```

---

### Task 10: Popup dim the user can see

**Files:**
- Modify: `internal/tui/styles.go` (`dim` at :55)
- Test: `internal/tui/styles_test.go` (`TestHexThemeRoles` at :55), `internal/tui/view_test.go` (`TestPopupDimsTheBackground` at :1072)

**verify:** For every popup, at every size and under both the 256 color and the 24-bit color profile, every cell behind the box is drawn in the dim brush and no other, the dim color differs from every color the panes wear when no popup is open, and closing the popup gives back the exact screen from before. List every theme, profile, popup and size checked, and name the cause the debug step proved.

**Interfaces:**
- Consumes: `theme.Names()`, `theme.Builtin(name)`, `withTrueColor` (styles_test.go), `withColors` (view_test.go).
- Produces: `func mixHex(a, b string) string`; `func hexRGB(s string) ([3]int, bool)`.

- [x] **Step 1: Find the cause with acta:debug**

Load `acta:debug` and run its root-cause phase before any test. Nobody has proven why the dim text looks unchanged. Check at least these, with the output shown:
- Dump the raw bytes of one background line with the help popup open, under `termenv.TrueColor` and the default theme, in a scratch test that is not committed. Confirm the SGR codes are `2` (faint) and `38;2;65;72;104` (`#414868`) and nothing resets them mid-line.
- Compare that with the same line with no popup open. The rows the cursor is not on already wear `faint` (styles.go:52), so a faint row with the theme text color and a faint row in `#414868` can look almost the same. Measure how far apart the two colors are.
- Check which color profile the real program picks (lipgloss reads it from the terminal) under the user's terminal and under tmux, since a 256 color profile maps `#414868` to a near grey.

Write the proven cause in one line in the task report. The fix below is the likely one: the dim color moves halfway from slot 8 to the theme background. If the debug step proves a different cause, fix that cause instead, keep the tests of Step 2 as the proof, and say so in the report.

- [x] **Step 2: Write the failing tests**

In `internal/tui/styles_test.go`, change the dim check in `TestHexThemeRoles` to the new color (tokyo-night: halfway from `#1a1b26` to `#414868`, worked out by hand):

```go
	if s.dim.GetForeground() != lipgloss.Color("#2d3147") {
		t.Fatalf("dim = %v", s.dim.GetForeground())
	}
```

Add to `internal/tui/styles_test.go`:

```go
// TestDimFadesTowardTheBackground checks every built-in theme with its own
// colors: the dim color sits halfway between the background and slot 8, so
// it is never a color the panes already wear.
func TestDimFadesTowardTheBackground(t *testing.T) {
	t.Parallel()

	for _, name := range theme.Names() {
		th, ok := theme.Builtin(name)
		if !ok || th.BG == "" {
			continue
		}
		s := newStyles(th, true)
		got := s.dim.GetForeground()
		if want := lipgloss.Color(mixHex(th.BG, th.ANSI[slotDim])); got != want {
			t.Errorf("%s: dim %v, want %v", name, got, want)
		}
		for _, used := range []string{th.FG, th.ANSI[slotDim], th.ANSI[slotAccent], th.ANSI[slotWork]} {
			if th.BG != th.ANSI[slotDim] && got == lipgloss.Color(used) {
				t.Errorf("%s: dim %v is a color the panes already wear", name, got)
			}
		}
	}
	term, _ := theme.Builtin("terminal")
	if got := newStyles(term, true).dim.GetForeground(); got != lipgloss.Color("8") {
		t.Errorf("terminal theme: dim %v, want slot 8", got)
	}
}

func TestMixHex(t *testing.T) {
	t.Parallel()

	for _, c := range []struct{ a, b, want string }{
		{"#1a1b26", "#414868", "#2d3147"},
		{"#000000", "#ffffff", "#7f7f7f"},
		{"#ABCDEF", "#abcdef", "#abcdef"},
		{"", "#414868", "#414868"},
		{"#12345", "#414868", "#414868"},
		{"#zzzzzz", "#414868", "#414868"},
	} {
		if got := mixHex(c.a, c.b); got != c.want {
			t.Errorf("mixHex(%q, %q) = %q, want %q", c.a, c.b, got, c.want)
		}
	}
}
```

In `internal/tui/view_test.go`, `TestPopupDimsTheBackground` runs its body under both profiles and names the new color. Move the body into a helper and call it twice:

```go
func TestPopupDimsTheBackground(t *testing.T) {
	// The brush the view paints the screen behind a popup with, spelled out
	// here so this test checks the color the plan names and not the one the
	// view happens to use today.
	dim := func() lipgloss.Style { return lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("#2d3147")) }
	withColors(func() { checkPopupDim(t, dim()) })
	withTrueColor(func() { checkPopupDim(t, dim()) })
}

// checkPopupDim opens every popup at both sizes and checks that every cell
// behind the box is the dim brush over its own plain text, and that esc gives
// back the exact screen from before.
func checkPopupDim(t *testing.T, dim lipgloss.Style) {
	t.Helper()
	for _, open := range []string{"?", "t", "s", "n"} {
		for _, size := range [][2]int{{80, 30}, {160, 50}} {
			m := press(sized(newModel(t), size[0], size[1]), tabKey(tabBugs))
			before := m.View()
			pop := press(m, open)
			rows := strings.Split(pop.popupBox(), "\n")
			x0, y0, w, h := popupRect(rows, pop.width, pop.height)
			if w == 0 {
				t.Errorf("popup %q at %dx%d never opened", open, size[0], size[1])
				continue
			}
			// The theme background is under every line and after every
			// reset. Take it off again, because what this test reads is
			// the brush on the text and the frame has its own test.
			frame := strings.TrimSuffix(pop.styles.paintFrame(""), "\x1b[K")
			after := strings.Split(pop.View(), "\n")
			for y, drawn := range after {
				ln := strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(drawn, frame), "\x1b[K"), frame, "")
				parts := []string{ln}
				if y >= y0 && y < y0+h && y < len(after)-1 {
					i := strings.Index(ln, rows[y-y0])
					if i < 0 {
						t.Errorf("popup %q at %dx%d line %d: the box row is not drawn the way popupBox draws it", open, size[0], size[1], y)
						continue
					}
					head, tail := ln[:i], ln[i+len(rows[y-y0]):]
					if lipgloss.Width(head) != x0 || lipgloss.Width(rows[y-y0]) != w {
						t.Errorf("popup %q at %dx%d line %d: the box sits at column %d and is %d cells wide, want %d and %d", open, size[0], size[1], y, lipgloss.Width(head), lipgloss.Width(rows[y-y0]), x0, w)
					}
					parts = []string{head, tail}
				}
				for _, seg := range parts {
					if got, want := seg, dim.Render(plain(seg)); got != want {
						t.Errorf("popup %q at %dx%d line %d: the background keeps its own color\n got %q\nwant %q", open, size[0], size[1], y, got, want)
					}
				}
			}
			if closed := press(pop, "esc").View(); closed != before {
				t.Errorf("popup %q at %dx%d: the view is not byte-equal to the one before it opened", open, size[0], size[1])
			}
		}
	}
}
```

This is the old loop body word for word, moved into the helper.

- [x] **Step 3: Run the tests to see them fail**

Run: `go test ./internal/tui -run 'TestHexThemeRoles|TestDimFadesTowardTheBackground|TestMixHex|TestPopupDimsTheBackground' -v`
Expected: FAIL, `undefined: mixHex`.

- [x] **Step 4: Write the implementation**

In `internal/tui/styles.go` (add `"fmt"` to the imports), in `newStyles`:

```go
		// dim paints the screen behind a popup, so the box on top is the only
		// thing left with a color of its own. The rows behind are already
		// faint, so dim has to fade further: halfway to the background.
		dim: lipgloss.NewStyle().Faint(true).Foreground(dimColor(t, slot(slotDim))),
```

Add under `newStyles`:

```go
// dimColor is the color behind a popup. A theme with its own colors mixes
// slot 8 halfway toward its background. The terminal theme has no hex to
// mix, so it keeps slot 8.
func dimColor(t theme.Theme, grey lipgloss.Color) lipgloss.Color {
	if t.BG == "" {
		return grey
	}
	return lipgloss.Color(mixHex(t.BG, string(grey)))
}

// mixHex gives the color halfway between two #rrggbb colors. When either one
// is not a #rrggbb color there is nothing to mix, so b comes back as it is.
func mixHex(a, b string) string {
	ca, okA := hexRGB(a)
	cb, okB := hexRGB(b)
	if !okA || !okB {
		return b
	}
	return fmt.Sprintf("#%02x%02x%02x", (ca[0]+cb[0])/2, (ca[1]+cb[1])/2, (ca[2]+cb[2])/2)
}

// hexRGB reads #rrggbb into its red, green and blue parts.
func hexRGB(s string) ([3]int, bool) {
	s, ok := strings.CutPrefix(s, "#")
	if !ok || len(s) != 6 {
		return [3]int{}, false
	}
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return [3]int{}, false
	}
	return [3]int{int(n >> 16 & 0xff), int(n >> 8 & 0xff), int(n & 0xff)}, true
}
```

- [x] **Step 5: Run the tests to see them pass, then the package**

Run: `go test ./internal/tui -run 'TestHexThemeRoles|TestDimFadesTowardTheBackground|TestMixHex|TestPopupDimsTheBackground' -v`
Expected: PASS.
Run: `go test ./internal/tui`
Expected: `ok`.

- [x] **Step 6: Commit**

```bash
gofmt -l internal/tui && go vet ./internal/tui
git add internal/tui/styles.go internal/tui/styles_test.go internal/tui/view_test.go
git commit -m "fix(tui): fade the screen behind a popup toward the background"
```
