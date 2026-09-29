---
id: PLN-0037
created: "2026-09-29"
hash: z02pmcq
started: "2026-09-29"
---
# TUI Scroll Performance Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Scrolling in every TUI pane keeps up with the wheel: wheel events in one frame cost one scroll and one render, each notch moves 3 lines, and the detail lines are built once per change instead of several times per event.

**Architecture:** The wheel handler in `internal/tui/model.go` stops scrolling at once. It adds to a pending delta on the model and arms one 16 ms tick. The tick message scrolls the pane once by the whole delta, clamped. `detailLines` in `internal/tui/detail.go` keeps its last result in a small cache shared through a pointer, keyed by the selected item, the width and the board.

**Tech Stack:** Go, Bubble Tea v1.3.10, lipgloss (all already in go.mod).

**Spec:** `.acta/specs/2026-09-29-tui-scroll-performance-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- No new dependency in go.mod.
- Key scrolling (`j`, `k`, arrows, `ctrl+d`, `ctrl+u`) does not change.
- `board.Load` and the author lookup do not change.
- Comments are plain English a 10-year-old reads back: short words, say why.
- Before each commit run `gofmt -l internal/tui` (must print nothing) and `go vet ./internal/tui/`.
- Test command: `go test ./internal/tui/`. The full suite `go test ./...` must pass before the last commit.

## File Map

- `internal/tui/model.go`: wheel handler, pending delta fields, tick message, `New` sets the detail cache.
- `internal/tui/model_test.go`: `wheel` test helper delivers the tick.
- `internal/tui/scroll_test.go`: wheel tests move to the 3-line step; new batching tests.
- `internal/tui/detail.go`: detail cache type and cached `detailLines`.
- `internal/tui/detail_test.go`: cache tests.
- `internal/tui/view_bench_test.go` (new): wheel plus `View` benchmark.

## Waves

- Wave 1: Task 1 (model.go, model_test.go, scroll_test.go).
- Wave 2: Task 2 (detail.go, detail_test.go, view_bench_test.go, and small edits in model.go and model_test.go). Task 2 touches the same files as Task 1, so it waits for it.

---

### Task 1: Batch wheel events and scroll 3 lines per notch

**Files:**
- Modify: `internal/tui/model.go` (the `Model` struct, `Update`, the wheel cases in `mouse`)
- Modify: `internal/tui/model_test.go` (the `wheel` helper)
- Modify: `internal/tui/scroll_test.go` (`TestWheelScrollsOnlyTheFocusedPane` and new tests)

**verify:** Every wheel event reaches the screen as part of exactly one scroll, and no scroll path ever leaves an offset outside `0..lastOff(p)`. List every path checked: one notch, many notches in one frame, up and down in one frame, wheel over a pane without focus, wheel while help, popup, slug or search is open, and a tick that arrives with nothing pending.

**Interfaces:**
- Produces: `const wheelStep = 3`, `const wheelFrame = 16 * time.Millisecond`, `type wheelTickMsg struct{}`, Model fields `wheelPane pane`, `wheelDelta int`, `wheelArmed bool`.

- [x] **Step 1: Change the test helper so it delivers the tick**

In `internal/tui/model_test.go`, the helper sends the wheel event, then the tick the real program would send 16 ms later. No sleep.

```go
// wheel turns the scroll wheel over a cell of the screen. up is true for a
// notch away from the user. The real program scrolls on the next frame tick,
// so the helper hands that tick over at once.
func wheel(m Model, x, y int, up bool) Model {
	return wheelTick(wheelOnly(m, x, y, up))
}

// wheelOnly turns the wheel one notch and does not send the frame tick, so a
// test can stack several notches into one frame.
func wheelOnly(m Model, x, y int, up bool) Model {
	button := tea.MouseButtonWheelDown
	if up {
		button = tea.MouseButtonWheelUp
	}
	next, _ := m.Update(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: button})
	return next.(Model)
}

// wheelTick is the frame tick that applies the notches gathered so far.
func wheelTick(m Model) Model {
	next, _ := m.Update(wheelTickMsg{})
	return next.(Model)
}
```

- [x] **Step 2: Move the existing wheel test to the 3-line step**

In `TestWheelScrollsOnlyTheFocusedPane` (`internal/tui/scroll_test.go`), one notch now moves `min(wheelStep, m.lastOff(p))` lines. Replace the two checks after the first `wheel(..., false)`:

```go
		step := min(wheelStep, m.lastOff(p))
		if m.off[p] != step {
			t.Errorf("pane %d: one notch should scroll it by %d, off is %d", p, step, m.off[p])
		}
```

and the `shows(m, p, 1)` check:

```go
		if !shows(m, p, step) {
			t.Errorf("pane %d: after one notch the top of the box should be line %d, it shows %q", p, step, drawnFirst(m, p))
		}
```

The rest of the test (back to the top, stop at both ends) stays as it is.

- [x] **Step 3: Write the new failing tests**

Add to `internal/tui/scroll_test.go`. `paneModel` and `scrollBox` already exist there.

```go
func TestWheelNotchesInOneFrameScrollOnce(t *testing.T) {
	t.Parallel()

	for _, p := range []pane{paneList, paneDone, paneDetail} {
		m := paneModel(t, p)
		b := scrollBox(m, p)
		m = wheelOnly(m, b.x+1, b.y+2, false)
		m = wheelOnly(m, b.x+1, b.y+2, false)
		if m.off[p] != 0 {
			t.Errorf("pane %d: notches must wait for the frame tick, off is %d", p, m.off[p])
		}
		m = wheelTick(m)
		if want := min(2*wheelStep, m.lastOff(p)); m.off[p] != want {
			t.Errorf("pane %d: two notches in one frame should scroll %d, off is %d", p, want, m.off[p])
		}
		// A second tick with nothing gathered must not move the pane again.
		before := m.off[p]
		m = wheelTick(m)
		if m.off[p] != before {
			t.Errorf("pane %d: an empty tick moved the pane from %d to %d", p, before, m.off[p])
		}
	}
}

func TestWheelUpAndDownInOneFrameCancel(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	b := scrollBox(m, paneDetail)
	m = wheel(m, b.x+1, b.y+2, false)
	start := m.off[paneDetail]
	m = wheelOnly(m, b.x+1, b.y+2, false)
	m = wheelOnly(m, b.x+1, b.y+2, true)
	m = wheelTick(m)
	if m.off[paneDetail] != start {
		t.Errorf("down then up in one frame should leave the pane at %d, it is at %d", start, m.off[paneDetail])
	}
}

func TestWheelArmsOneTickPerFrame(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneList)
	b := scrollBox(m, paneList)
	next, first := m.Update(tea.MouseMsg{X: b.x + 1, Y: b.y + 2, Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown})
	if first == nil {
		t.Fatal("the first notch of a frame must start the frame tick")
	}
	_, second := next.Update(tea.MouseMsg{X: b.x + 1, Y: b.y + 2, Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown})
	if second != nil {
		t.Error("a notch in a frame that already has a tick must not start another one")
	}
}

func TestWheelOverclampStaysInsideContent(t *testing.T) {
	t.Parallel()

	for _, p := range []pane{paneList, paneDone, paneDetail} {
		m := paneModel(t, p)
		b := scrollBox(m, p)
		for range 500 {
			m = wheelOnly(m, b.x+1, b.y+2, false)
		}
		m = wheelTick(m)
		if want := m.lastOff(p); m.off[p] != want {
			t.Errorf("pane %d: 500 notches in one frame scrolled to %d, the content allows %d", p, m.off[p], want)
		}
		for range 500 {
			m = wheelOnly(m, b.x+1, b.y+2, true)
		}
		m = wheelTick(m)
		if m.off[p] != 0 {
			t.Errorf("pane %d: 500 notches up in one frame scrolled to %d, the top is 0", p, m.off[p])
		}
	}
}
```

The existing tests that open help, a popup, the slug prompt or search and then turn the wheel keep covering the guard at the top of `mouse`: a notch there must gather nothing. Check that `go test ./internal/tui/ -run Wheel` lists them; if none turns the wheel while help is open, add this one:

```go
func TestWheelWhileHelpIsOpenGathersNothing(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	b := scrollBox(m, paneDetail)
	m.help = true
	m = wheel(m, b.x+1, b.y+2, false)
	if m.off[paneDetail] != 0 || m.wheelDelta != 0 {
		t.Errorf("the wheel under the help popup moved the pane: off %d, pending %d", m.off[paneDetail], m.wheelDelta)
	}
}
```

- [x] **Step 4: Run the tests to watch them fail**

Run: `go test ./internal/tui/ -run 'Wheel'`
Expected: build FAIL with `undefined: wheelTickMsg` and `undefined: wheelStep`.

- [x] **Step 5: Write the implementation**

In `internal/tui/model.go`, next to the other message types:

```go
// wheelStep is how many lines one wheel notch moves, the usual terminal step.
const wheelStep = 3

// wheelFrame is how long wheel notches are gathered before one scroll. A
// trackpad sends hundreds of notches a second, and drawing the screen after
// each one left the TUI far behind the wheel.
const wheelFrame = 16 * time.Millisecond

// wheelTickMsg says the frame is over: scroll by what the wheel gathered.
type wheelTickMsg struct{}
```

Add three fields to `Model`, after `off`:

```go
	wheelPane  pane // the pane the gathered notches scroll
	wheelDelta int  // lines gathered from the wheel, not yet scrolled
	wheelArmed bool // true while a frame tick is on its way
```

In `Update`, add a case before `case tea.KeyMsg:`:

```go
	case wheelTickMsg:
		if m.wheelDelta != 0 {
			m.scrollPane(m.wheelPane, m.wheelDelta)
		}
		m.wheelDelta, m.wheelArmed = 0, false
```

Replace the two wheel cases in `mouse`:

```go
	switch msg.Button {
	case tea.MouseButtonWheelUp, tea.MouseButtonWheelDown:
		// The wheel belongs to the focused pane, the way the keys do, and it
		// never takes the focus. A wheel over another pane is ignored, so
		// scrolling can never move a pane the user is not looking at.
		if p != m.focus {
			return m, nil
		}
		step := wheelStep
		if msg.Button == tea.MouseButtonWheelUp {
			step = -wheelStep
		}
		// Notches for another pane cannot share one delta. Scroll the old
		// pane now so its notches are not lost.
		if m.wheelDelta != 0 && m.wheelPane != p {
			m.scrollPane(m.wheelPane, m.wheelDelta)
			m.wheelDelta = 0
		}
		m.wheelPane = p
		m.wheelDelta += step
		if m.wheelArmed {
			return m, nil
		}
		m.wheelArmed = true
		return m, tea.Tick(wheelFrame, func(time.Time) tea.Msg { return wheelTickMsg{} })
	}
```

- [x] **Step 6: Run the tests to watch them pass**

Run: `go test ./internal/tui/`
Expected: PASS. Any other test that counted on one line per notch fails here; fix its expected number to `wheelStep` the same way as Step 2, never by removing the check.

- [x] **Step 7: Format, vet, commit**

```bash
gofmt -l internal/tui
go vet ./internal/tui/
git add internal/tui/model.go internal/tui/model_test.go internal/tui/scroll_test.go
git commit -m "tui: batch wheel notches per frame and scroll 3 lines each"
```

---

### Task 2: Cache the detail lines

**Files:**
- Modify: `internal/tui/detail.go` (`detailLines`)
- Modify: `internal/tui/model.go` (`Model` struct gets one field, `New` sets it)
- Modify: `internal/tui/model_test.go` (`newModel` starts an empty cache)
- Test: `internal/tui/detail_test.go`
- Create: `internal/tui/view_bench_test.go`

**verify:** The detail box never shows lines built for another item, another width or another board. List every input `detailLines` reads (selected item, width, board, styles, renderer) and say for each how a change to it gets new lines.

**Interfaces:**
- Consumes: `wheel`, `wheelOnly`, `wheelTick` test helpers from Task 1.
- Produces: `type detailCache struct`, Model field `dcache *detailCache`.

- [ ] **Step 1: Write the failing tests**

Add to `internal/tui/detail_test.go`. `newModel` and `sized` already exist in the tests.

```go
func TestDetailLinesAreBuiltOncePerChange(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 160, 50)
	calls := 0
	m.render = func(md string, _ int) string { calls++; return md }
	m.dcache = &detailCache{}
	m.detailLines(80)
	m.detailLines(80)
	m.View()
	if calls != 1 {
		t.Errorf("the same item at the same width rendered its body %d times, want 1", calls)
	}
}

func TestDetailCacheFollowsWidthItemAndBoard(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 160, 50)
	m.dcache = &detailCache{}
	wide := strings.Join(m.detailLines(80), "\n")
	narrow := strings.Join(m.detailLines(30), "\n")
	if wide == narrow {
		t.Error("a new width gave the lines of the old width")
	}
	first := strings.Join(m.detailLines(80), "\n")
	m = press(m, "j")
	if strings.Join(m.detailLines(80), "\n") == first {
		t.Error("a new selected item gave the lines of the old item")
	}
	shown := strings.Join(m.detailLines(80), "\n")
	fresh, err := m.load()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range fresh.Items {
		it.Title = it.Title + " renamed"
	}
	next, _ := m.Update(reloadMsg{b: fresh})
	if got := strings.Join(next.(Model).detailLines(80), "\n"); got == shown || !strings.Contains(got, "renamed") {
		t.Error("a reloaded board gave the lines of the old board")
	}
}

func TestDetailWithoutCacheStillDraws(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 160, 50)
	m.dcache = nil
	if len(m.detailLines(80)) == 0 {
		t.Error("a model with no cache must still draw the detail box")
	}
}
```

Create `internal/tui/view_bench_test.go`:

```go
package tui

import (
	"testing"
)

// BenchmarkWheelFrame is one frame of trackpad scrolling in the detail box:
// ten notches, one tick, one draw.
func BenchmarkWheelFrame(b *testing.B) {
	t := &testing.T{}
	m := paneModel(t, paneDetail)
	box := scrollBox(m, paneDetail)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for range 10 {
			m = wheelOnly(m, box.x+1, box.y+2, i%2 == 0)
		}
		m = wheelTick(m)
		m.View()
	}
}
```

If `paneModel` needs a real `*testing.T` (it calls `t.Helper`, `t.TempDir` or `t.Fatal`), change its parameter to `testing.TB` so the benchmark passes `b`, and pass `b` instead of `t`.

- [ ] **Step 2: Run the tests to watch them fail**

Run: `go test ./internal/tui/ -run 'DetailLines|DetailCache|DetailWithout'`
Expected: build FAIL with `undefined: detailCache` and `m.dcache undefined`.

- [ ] **Step 3: Write the implementation**

In `internal/tui/detail.go`, above `detailLines`:

```go
// detailCache keeps the last lines of the detail box. The box is asked for its
// lines several times on every key and wheel notch, and building them renders
// the whole markdown body each time. The model is copied on every update, so
// the cache sits behind a pointer that all the copies share.
type detailCache struct {
	board *board.Board
	item  *board.Item
	width int
	lines []string
}
```

Rename the current `detailLines` to `buildDetailLines` (body unchanged), and add:

```go
// detailLines gives the lines of the detail box, built again only when the
// item, the width or the board changed. A reload makes a new board, so its
// items are new too and the old lines are never shown for them.
func (m Model) detailLines(w int) []string {
	c := m.dcache
	if c == nil {
		return m.buildDetailLines(w)
	}
	it := m.Selected()
	if c.lines != nil && c.board == m.board && c.item == it && c.width == w {
		return c.lines
	}
	*c = detailCache{board: m.board, item: it, width: w, lines: m.buildDetailLines(w)}
	return c.lines
}
```

In `internal/tui/model.go`, add the field to `Model` next to `render`:

```go
	dcache   *detailCache // the last detail lines, shared by every copy
```

and in `New`, next to `render:`:

```go
		dcache: &detailCache{},
```

`WithTheme` changes the styles and the renderer, so it must start an empty cache. In `WithTheme`, before it returns the model, add:

```go
	m.dcache = &detailCache{}
```

Code that changes `m.render` or `m.styles` on an existing model (tests do this after `New`) must also set `m.dcache = &detailCache{}`; the tests in Step 1 do so. Update `newModel` in `internal/tui/model_test.go` to set `m.dcache = &detailCache{}` after it replaces `m.render`, so no test sees lines built by the glamour renderer.

- [ ] **Step 4: Run the tests to watch them pass**

Run: `go test ./internal/tui/`
Expected: PASS.

Run: `go test ./internal/tui/ -run xxx -bench WheelFrame -benchmem`
Expected: the benchmark runs and prints ns/op. Paste the line in the task report.

- [ ] **Step 5: Full suite, format, vet, commit**

```bash
go test ./...
gofmt -l internal/tui
go vet ./internal/tui/
git add internal/tui/detail.go internal/tui/detail_test.go internal/tui/model.go internal/tui/model_test.go internal/tui/view_bench_test.go
git commit -m "tui: build the detail lines once per item, width and board"
```
