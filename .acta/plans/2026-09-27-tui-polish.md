---
id: PLAN-16
hash: o8nd
---
# TUI Polish Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** One-line list rows, task dots and AUTHOR in the detail pane, per-pane scroll with scrollbars, square corners, a popup that covers only its box, frames that never break, and a land gate for hanging tasks.

**Architecture:** All TUI work stays in `internal/tui`. `view.go` is 900 lines, so new code goes into new files: `frame.go` (geometry and pane borders), `scroll.go` (per-pane offsets and scrollbar), `detail.go` (detail header, dot list, body). A small `gitc.Author` helper gives AUTHOR, cached once per board load. The land gate is skill text guarded by `internal/plugincheck`.

**Tech Stack:** Go 1.27, Bubble Tea, lipgloss, glamour.

**Spec:** `.acta/specs/2026-09-27-tui-polish-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- No new keys, no new tabs, no new data files, no frontmatter field (spec Non-goals).
- In-progress means `inProgress(it)` in `internal/tui/model.go` (status doing, in-progress, fixing, or started). Reuse it everywhere.
- Dots: `●` in progress (accent), `○` not started (dim), `✓` done or wontfix (dim).
- Square corners only: `┌ ┐ └ ┘`. No `╭ ╮ ╰ ╯` anywhere.
- No file in `internal/tui` grows past 800 lines; `view.go` must shrink, not grow.
- Comments: plain English a 10-year-old reads, say why. No marker tags.
- Gates before every commit: `gofmt -l .` prints nothing, `go vet ./...` clean, `go test ./...` passes.
- TDD: write the failing test first, watch it fail, then the minimum code.
- Tests run the acta binary only inside `t.TempDir()` repos; never run `acta` write commands (bug new, debt new, set, id) in the worktree itself, except the `acta tick` lines this plan names.

## File map

| File | Tasks |
|---|---|
| `internal/tui/frame.go` (new: `geometry`, `split`, `topLine`, `paneBottom`, `paneTop` moved from `view.go`), `internal/tui/model.go` (WindowSizeMsg), `internal/tui/frame_test.go` (new) | 1 |
| `plugin/skills/land/SKILL.md`, `plugin/skills/plan/SKILL.md`, `internal/plugincheck/skill_land_test.go`, `internal/plugincheck/skill_plan_test.go` | 2 |
| `internal/tui/frame.go` (corners), `internal/tui/view.go` (`cover`, `popupBox`), `internal/tui/view_test.go` | 3 |
| `internal/tui/view.go` (`listView`, `rowText`, styles), `internal/tui/model.go` (`rowLines`, divider row), `internal/tui/model_test.go`, `internal/tui/view_test.go` | 4 |
| `internal/tui/scroll.go` (new), `internal/tui/model.go` (scroll state, `step`/`top`/`end`, wheel), `internal/tui/view.go` (`box`, `detailView` footer), `internal/tui/scroll_test.go` (new) | 5 |
| `internal/tui/detail.go` (new: `detailLines` moved from `view.go`, dot list, debt body), `internal/gitc/gitc.go` (`Author`, `UserName`), `internal/board/board.go` (author cache, debt item body), tests | 6 |

## Waves

- Wave 1: Task 1, Task 2 (no shared files)
- Wave 2: Task 3
- Wave 3: Task 4
- Wave 4: Task 5
- Wave 5: Task 6
- Wave 6: Task 7 (touches board.go after Task 6)

Tasks 3 to 6 all touch `view.go` or `model.go`, so they run one at a time.

---

### Task 1: Frames never break

**Files:**
- Create: `internal/tui/frame.go` (move `geometry`, `split`, `paneTop`, `topLine`, `paneBottom` out of `view.go` unchanged first, then fix)
- Modify: `internal/tui/model.go` (the `tea.WindowSizeMsg` case in `Update`, today at about line 176)
- Create: `internal/tui/frame_test.go`

**verify:** No terminal size and no scroll position ever gives a frame with rows or borders out of place. List every size range checked (heights 3 to 60, widths 20 to 200) with the three checks (left column height equals detail height; `View` line count at most the height; every line fits the width), and the resize and scroll sequence checked at full size, with what each gave. If the full-size repro shows a cause other than stale rows after a size change, stop and report it instead of fixing.

**Interfaces:**
- Produces: `split(h int) (top, bottom int)` with `top+bottom == h` for every `h >= 2`; `Update` returns `tea.ClearScreen` on every `tea.WindowSizeMsg` whose size differs from the last one.

- [x] **Step 1: Write the failing tests**

```go
func TestSplitAlwaysFillsTheHeight(t *testing.T) {
	for h := 2; h <= 80; h++ {
		top, bottom := split(h)
		if top+bottom != h || top < 1 || bottom < 1 {
			t.Errorf("split(%d) = %d, %d", h, top, bottom)
		}
	}
}

func TestViewFitsEveryTerminalSize(t *testing.T) {
	m := newModel(t)
	for h := 3; h <= 60; h++ {
		for _, w := range []int{20, 40, 60, 80, 120, 200} {
			v := sized(m, w, h).View()
			lines := strings.Split(v, "\n")
			if len(lines) > h {
				t.Fatalf("%dx%d: %d lines", w, h, len(lines))
			}
			for i, ln := range lines {
				if lipgloss.Width(ln) > w {
					t.Fatalf("%dx%d line %d is %d wide", w, h, i, lipgloss.Width(ln))
				}
			}
			g := sized(m, w, h).geometry()
			if g.wide && g.open.h+g.done.h != g.detail.h {
				t.Fatalf("%dx%d: left %d+%d, detail %d", w, h, g.open.h, g.done.h, g.detail.h)
			}
		}
	}
}

func TestResizeClearsTheScreen(t *testing.T) {
	m := sized(newModel(t), 120, 40)
	_, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if cmd == nil || fmt.Sprint(cmd()) != fmt.Sprint(tea.ClearScreen()) {
		t.Fatal("a size change must clear the screen so old rows do not stay")
	}
}
```

Then reproduce the user's full-size breakage by hand in a real terminal: run `go run ./cmd/acta` in a scratch clone (`git clone . /tmp/tui-repro`), resize the window several times and scroll the detail pane over a long plan. Write down what you saw in your report. If rows from an earlier frame stay on screen, the `ClearScreen` test above covers the cause.

- [x] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui/ -run 'Split|FitsEvery|ResizeClears'`
Expected: FAIL (`split(12) = 8, 5`, left column taller than detail, no clear command).

- [x] **Step 3: Write minimal implementation**

`split`: `top := h * 2 / 3`, clamp `top` to `[1, h-1]`, return `top, h - top`. In `Update`, when the new size differs from `m.width`/`m.height`, set them and return `tea.ClearScreen`. Keep the line cut in `View` as a last guard.

- [x] **Step 4: Run tests to verify they pass**

Run: `go test ./...`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/tui/frame.go internal/tui/frame_test.go internal/tui/view.go internal/tui/model.go
git commit -m "fix(tui): panes always fill the height and resize clears old rows"
```

### Task 2: Land refuses hanging tasks

**Files:**
- Modify: `plugin/skills/land/SKILL.md` (a new precondition right after step 1 of "Landing")
- Modify: `plugin/skills/plan/SKILL.md` (one line in "Task Right-Sizing")
- Test: `internal/plugincheck/skill_land_test.go`, `internal/plugincheck/skill_plan_test.go`

**verify:** No landing can finish while its plan shows an open box, and no plan skill text lets a step wait for after the merge. List each skill file checked and each guard string added, and confirm the old step list still passes its existing guards.

- [x] **Step 1: Write the failing tests**

In `skill_land_test.go` add to `Must`: `"acta show <plan id>"` and `"done` is less than `total"` (match the exact words you will write). In `skill_plan_test.go` add to `Must`: `"never holds a step that can only happen after landing"`.

- [x] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/plugincheck/`
Expected: FAIL, missing strings.

- [x] **Step 3: Write the text**

land, after step 1: "Run `acta show <plan id>`. When `done` is less than `total`, do not merge: tick each box whose work is really done, and report the rest to the user." Renumber the steps after it.
plan, in "Task Right-Sizing": "A plan never holds a step that can only happen after landing; that work goes in the landing report as the next action."

- [x] **Step 4: Run tests to verify they pass**

Run: `go test ./...`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add plugin/skills/land/SKILL.md plugin/skills/plan/SKILL.md internal/plugincheck/skill_land_test.go internal/plugincheck/skill_plan_test.go
git commit -m "docs(plugin): land checks plan progress, plans hold no after-land steps"
```

### Task 3: Square corners and a popup that covers only its box

**Files:**
- Modify: `internal/tui/frame.go` (`topLine`, `paneBottom` corner literals)
- Modify: `internal/tui/view.go` (`cover`, `popupBox`)
- Test: `internal/tui/view_test.go`

**verify:** No frame ever shows a rounded corner, and no cell outside the popup box ever changes when the popup opens. List each frame checked (every tab, focus on each pane, popup open, widths 60/120/200) and, for the popup, the count of cells outside the box that differ (must be 0).

- [x] **Step 1: Write the failing tests**

```go
func TestNoRoundedCorners(t *testing.T) {
	for _, w := range []int{60, 120, 200} {
		m := sized(newModel(t), w, 40)
		for _, v := range []string{m.View(), press(m, "?").View()} {
			if strings.ContainsAny(v, "╭╮╰╯") {
				t.Fatalf("width %d: rounded corner in frame", w)
			}
		}
	}
}

func TestPopupChangesOnlyItsBox(t *testing.T) {
	m := sized(newModel(t), 120, 40)
	before := strings.Split(ansi.Strip(m.View()), "\n")
	after := strings.Split(ansi.Strip(press(m, "?").View()), "\n")
	x0, y0, w, h := popupRect(press(m, "?"))
	for y := range before {
		b, a := []rune(before[y]), []rune(after[y])
		for x := 0; x < len(b) && x < len(a); x++ {
			inside := x >= x0 && x < x0+w && y >= y0 && y < y0+h
			if !inside && b[x] != a[x] {
				t.Fatalf("cell %d,%d changed outside the popup", x, y)
			}
		}
	}
}
```

`popupRect` is a small helper you add next to `popupBox` that returns where the box is drawn; `View` uses the same helper so both agree. Use the ANSI strip helper the package already imports (or `github.com/charmbracelet/x/ansi` if it is already in `go.mod`).

- [x] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui/ -run 'Rounded|PopupChanges'`
Expected: FAIL.

- [x] **Step 3: Write minimal implementation**

Swap the corner literals to `┌ ┐ └ ┘`. Make `cover` splice the popup's cells into each covered line only between `x0` and `x0+w` (cut the line at display width with the existing `truncate`/width helpers, keep the left and right parts), instead of replacing whole lines.

- [x] **Step 4: Run tests to verify they pass**

Run: `go test ./...`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/tui/frame.go internal/tui/view.go internal/tui/view_test.go
git commit -m "fix(tui): square corners and a popup that covers only its box"
```

### Task 4: One-line list rows

**Files:**
- Modify: `internal/tui/model.go` (`rowLines` from 3 to 1, divider row)
- Modify: `internal/tui/view.go` (`listView`, `rowText`, the `selected` style)
- Test: `internal/tui/model_test.go`, `internal/tui/view_test.go`

**verify:** Every list row on every tab is exactly one line; only in-progress rows carry `n/m · agent`, and no row ever shows a status word; the selected row always has the subtle dark background (never reverse video) and bright text. List each tab and pane checked, the text of one in-progress row and one other row, the selected row's style, the divider line, and a click on the first and last visible row landing on the right item.

**Interfaces:**
- Produces: `rowLines = 1`; divider drawn as `strings.Repeat("─", w)` in the dim style.

- [x] **Step 1: Write the failing tests**

- For every tab in pane [1] and [2]: each item takes one line; the line after an item is the next item or the divider.
- An in-progress item's line ends with `2/5 · claude` (use a fixture task started with an agent, or set `Started` and the agent record the way existing agent tests do); a not-started item's line holds only the short id and title, and no line contains `todo`, `doing`, `open`, `done`, `approved` as a status word.
- A 200-character title is cut with `…` and the `n/m · agent` tail is still whole.
- The selected row's rendered string contains no reverse-video code (`\x1b[7m`), carries the background color 236 (dark) across the full inner width, and is not faint; unselected rows carry no background and are rendered faint.
- The divider line is `─` repeated across the inner width; hidden when one group is empty.
- Clicks: a click on each visible row selects that row (`rowAt` with one-line rows).

- [x] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui/`
Expected: FAIL.

- [x] **Step 3: Write minimal implementation**

Set `rowLines = 1`. `rowText` returns one line: short id, two spaces, title, and for `inProgress(it)` a right-aligned `done/total` plus ` · agent` when an agent is known; cut the title first so the tail fits. Replace `selected = Reverse(true)` with bright bold text on a subtle dark background `lipgloss.AdaptiveColor{Light: "254", Dark: "236"}`, padded to the full row width; render unselected rows faint; keep the accent color on in-progress titles. Draw the divider as a full-width `─` line. Drop the meta line code.

- [x] **Step 4: Run tests to verify they pass**

Run: `go test ./...`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/tui/model.go internal/tui/view.go internal/tui/model_test.go internal/tui/view_test.go
git commit -m "feat(tui): one-line rows, progress only on work in progress"
```

### Task 5: Each pane scrolls on its own, with a scrollbar

**Files:**
- Create: `internal/tui/scroll.go` (offset per pane, clamp, scrollbar column, `n/m` footer)
- Modify: `internal/tui/model.go` (replace the single `scroll int` with one offset per pane; `step`/`top`/`end`; wheel in `mouse`)
- Modify: `internal/tui/view.go` (`box`, `listView`, `detailView` use the offsets and scrollbar)
- Create: `internal/tui/scroll_test.go`

**verify:** Scrolling one pane never moves another, and the scrollbar and `n/m` always match what is on screen. List each pane checked (open, done, detail) with: wheel over it while focused (only it moves), wheel over it while unfocused (nothing moves, focus stays), keys on the focused pane, selection kept visible in lists, detail back to top on a new item, scrollbar shown only on overflow, thumb position at top, middle and end, and the `n/m` numbers at each.

**Interfaces:**
- Produces: `Model.off [3]int` (index by pane); `scrollbar(total, visible, first, h int) []string` returning one cell per inner row (`█` or `░`), empty when `total <= visible`.

- [x] **Step 1: Write the failing tests**

```go
func TestScrollbarThumbFollowsTheOffset(t *testing.T) {
	cases := []struct{ total, vis, first int; wantTop, wantBottom bool }{
		{100, 10, 0, true, false},
		{100, 10, 90, false, true},
	}
	for _, c := range cases {
		bar := scrollbar(c.total, c.vis, c.first, c.vis)
		if len(bar) != c.vis {
			t.Fatalf("bar has %d cells, want %d", len(bar), c.vis)
		}
		if (bar[0] == "█") != c.wantTop || (bar[c.vis-1] == "█") != c.wantBottom {
			t.Errorf("%+v: bar %v", c, bar)
		}
	}
	if len(scrollbar(5, 10, 0, 10)) != 0 {
		t.Error("no scrollbar when the content fits")
	}
}
```

Also: with a long plan selected, a wheel over the detail pane moves only the detail offset, the list selection and pane [2] do not change; a wheel over pane [2] while pane [1] is focused changes nothing (no scroll, focus stays on [1]); after a click on pane [2], a wheel over it scrolls pane [2] only; the detail pane's bottom border shows `1/N` at the top and `N-h+1/N`-style numbers after `G`; selecting another item puts the detail offset back to 0; a list longer than its pane keeps the selected row visible while moving with `j`.

- [x] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui/`
Expected: FAIL, `undefined: scrollbar`.

- [x] **Step 3: Write minimal implementation**

Keep one offset per pane in the model; clamp it every frame to `[0, total-visible]`. Lists move the offset only to keep the selection inside. The wheel scrolls by 1 only when the pane under the pointer is the focused pane, without moving the list selection when that pane is a list; over any other pane the wheel is ignored (no scroll, no focus change). When content overflows, reserve the last inner column for `scrollbar(...)` and write `first+1/total` into the bottom border through `paneBottom`, for every pane, not only detail.

- [x] **Step 4: Run tests to verify they pass**

Run: `go test ./...`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/tui/scroll.go internal/tui/scroll_test.go internal/tui/model.go internal/tui/view.go
git commit -m "feat(tui): per-pane scroll with a scrollbar and position"
```

### Task 6: Detail pane: AUTHOR, SUBTASKS, task dots, debt body

**Files:**
- Create: `internal/tui/detail.go` (move `detailLines`, `kindLabel`, `fromText` out of `view.go`, then change)
- Modify: `internal/gitc/gitc.go` (add `Author`, `UserName`)
- Modify: `internal/board/board.go` (fill `Item.Author` once per load with a cache per file; give debt items a `Body`)
- Test: `internal/tui/detail_test.go` (new), `internal/gitc/gitc_test.go`, `internal/board/board_test.go`

**verify:** Every detail kind shows the right header and the right task lines with the right dots, and no detail pane is ever empty when its file has content. List each kind checked (spec, plan, task, bug, debt item, item with no tasks) with its label (TASKS or SUBTASKS), its AUTHOR value and source (first commit, plan's author, user.name, left out), and its lines with dots; list the debt body shown.

**Interfaces:**
- Produces: `gitc.Author(repo, path string) (string, bool)` = author name of the commit that first added the file (`git log --diff-filter=A --format=%an -- <path>`, last line); `gitc.UserName(repo string) string` = `git config user.name`; `board.Item.Author string`.

- [x] **Step 1: Write the failing tests**

- `gitc`: in a temp repo with `GIT_AUTHOR_NAME=Ana` for the first commit and `Budi` for a later edit, `Author` returns `Ana`; an untracked file returns `"", false`; `UserName` returns the configured name.
- `board`: a plan's `Author` is its first-commit author; its tasks carry the same; a debt item's `Author` is its debt file's; an uncommitted file uses `UserName`; with no git at all the field is empty. Git runs once per file, not once per item (count calls through a package-level hook, or assert on a board with 3 tasks in one plan).
- `board`: a debt item's `Body` holds the debt file's text outside the checklist.
- `tui` detail header: a task shows `SUBTASKS`, a plan shows `TASKS`; `AUTHOR` sits right under `STATUS` and is left out when empty.
- `tui` dot list, one test per kind: plan lists its tasks in file order with `●`/`○`/`✓`; task lists its steps; spec lists its plans' tasks grouped under one line per plan; bug lists the tasks of plans whose `parent` is the bug; debt item lists every line of its file, the selected one bright; an item with no tasks shows no list and no empty line. In-progress lines end with `n/m · agent`.
- `tui` debt detail: the non-checklist text of the debt file renders under the list.

- [x] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/gitc/ ./internal/board/ ./internal/tui/`
Expected: FAIL.

- [x] **Step 3: Write minimal implementation**

Add the two `gitc` helpers. In `board.Load`, fill `Author` for file items with a `map[path]string` cache (one `gitc.Author` call per file, one `UserName` call per load); copy the file's author to its tasks and debt items. Give debt items the file's body minus checklist lines. In `detail.go`: label `SUBTASKS` for `KindTask`; add `AUTHOR` under `STATUS`; after the header rule, write the dot lines built from `Children` (and for a task, its step boxes from `board.Parse` of its section), then the markdown body.

- [x] **Step 4: Run tests to verify they pass**

Run: `go test ./...`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/tui/detail.go internal/tui/detail_test.go internal/tui/view.go internal/gitc internal/board
git commit -m "feat(tui): detail shows author, subtasks and task dots; debt detail has its text"
```

### Task 7: Status word `in-progress` instead of `doing`

Added by the user on 2026-09-27.

**Files:**
- Modify: `internal/board/board.go` (`taskStatuses`, the task status derivation that returns `"doing"`)
- Modify: `internal/tui/model.go` (`inProgress` checks `"doing"`)
- Modify: `plugin/skills/brainstorm/SKILL.md` (mentions `doing`)
- Test: `internal/board/board_test.go`, `internal/tui/model_test.go`, `cmd/acta/*_test.go` or `internal/cli/*_test.go` where they assert `doing`

**verify:** The word `doing` can never come back as a status in any form: no board item, `acta show`, `acta list --json` output or TUI frame shows it, and every task that is started or partly ticked reads `in-progress`. List every place checked (each status path in the derivation, `Allowed(KindTask)`, JSON output, the TUI in-progress check and row text, skill text) and what each shows now.

- [x] **Step 1: Write the failing tests**

- Board: a task with 1 of 3 boxes ticked, and a task started with `acta tick --start` and no ticks, both have `Status == "in-progress"`; `Allowed(KindTask)` is `todo, in-progress, done`.
- Guard: load the test fixture board and assert no item's `Status` is `doing`.
- TUI: a started task still counts as in progress (accent row, `n/m · agent`).
- CLI: `acta show` on a half-ticked task prints `status: in-progress`.
- Change every existing assertion that expects `doing` to expect `in-progress`.

- [x] **Step 2: Run tests to verify they fail**

Run: `go test ./...`
Expected: FAIL where `doing` is still produced.

- [x] **Step 3: Write minimal implementation**

Replace `"doing"` with `"in-progress"` in `taskStatuses` and in the task status derivation; in `inProgress` drop `"doing"` (keep `"in-progress"` and `"fixing"`). Reword the brainstorm skill line. Grep the repo once more for `"doing"` and `doing ·` and fix any hit that is a status.

- [x] **Step 4: Run tests to verify they pass**

Run: `go test ./...`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/board internal/tui/model.go internal/tui/model_test.go plugin/skills/brainstorm/SKILL.md internal/cli cmd/acta
git commit -m "feat: tasks under way read in-progress, not doing"
```
