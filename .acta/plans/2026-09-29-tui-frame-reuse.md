---
id: PLN-0038
created: "2026-09-29"
hash: wso2xr8
---
# TUI Frame Reuse Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `View` returns the last frame without drawing when the message before it surely changed nothing on screen, so wheel notches stop costing a full draw each.

**Architecture:** The model keeps the last frame string behind a pointer that every copy shares, like `dcache`. `Update` clears a `same` flag at the top of every message. Only three paths set it again: a wheel notch that only adds to the pending delta, a mouse event that is ignored, and a wheel tick that scrolls nothing. `View` returns the kept frame when `same` is set and a frame is kept. Otherwise it draws, keeps the frame and returns it.

**Tech Stack:** Go, Bubble Tea v1.3.10 (already in go.mod).

**Spec:** `.acta/specs/2026-09-29-tui-frame-reuse-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- No new dependency in go.mod.
- When in doubt, draw again. A path not named below never sets `same`.
- Do not make the draw itself cheaper, and do not change reload or `board.Load`.
- Comments use plain English that a 10-year-old can read back: short words, and they say why.
- Before the commit, run `gofmt -l internal` (it must print nothing), `go vet ./...` and `go test ./...`.

## File Map

- `internal/tui/model.go`: the `same` field, the `frame` pointer field, `New` makes the frame cache, `WithTheme` makes a fresh one, `Update` clears `same`, and the mouse and tick paths set it.
- `internal/tui/view.go`: `View` checks the flag, and the current body of `View` moves to `draw`.
- `internal/tui/view_test.go`: tests for reusing the frame.

## Waves

- Wave 1: Task 1.

---

### Task 1: Reuse the last frame when the screen did not change

**Files:**
- Modify: `internal/tui/model.go` (the `Model` struct, `New`, `WithTheme`, `Update`, the wheel tick case, `mouse`)
- Modify: `internal/tui/view.go` (`View`)
- Test: `internal/tui/view_test.go`

**verify:** `View` never returns a kept frame after a message that changed what the screen shows, and it never draws for a message that surely changed nothing.

List every path through `Update` and `mouse`, and for each one say whether it sets `same`. Only these three may set it:
- a notch that only adds to the delta, where the same call did not flush and scroll another pane;
- an ignored mouse event: the modal guard, a wheel over a pane without the focus, or a non-left non-wheel event;
- a tick that scrolled nothing.

Show that each path that scrolls or changes state draws again:
- a flush to another pane;
- a tick that scrolls;
- a key;
- a click on a row, a tab or the status line;
- a reload;
- a resize;
- a clock tick;
- the editor returning.

**Interfaces:**
- Produces: Model fields `same bool` and `frame *frameCache`, the type `frameCache struct{ s string }`, and the method `func (m Model) draw() string`, which holds the old body of `View`.

- [ ] **Step 1: Write the failing tests**

Add to `internal/tui/view_test.go`. The helpers `paneModel`, `scrollBox`, `wheelOnly`, `wheelTick`, `click`, `press`, `reloaded` and `sized` already exist in the tests. `drewNew` puts a marker in `m.frame`, so the test sees whether `View` drew.

```go
// drewNew calls View once and says whether it really drew. It puts a marker
// in the kept frame: a reused frame hands the marker back, a new draw does not.
func drewNew(m Model) bool {
	before := m.frame.s
	m.frame.s = "\x00sentinel"
	got := m.View()
	drew := got != "\x00sentinel"
	if !drew {
		m.frame.s = before
	}
	return drew
}

func TestViewReusesTheFrameForANotchThatOnlyGathers(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	b := scrollBox(m, paneDetail)
	m.View()
	m = wheelOnly(m, b.x+1, b.y+2, false)
	if drewNew(m) {
		t.Error("a notch that only adds to the delta drew the screen again")
	}
}

func TestViewReusesTheFrameForIgnoredMouseAndEmptyTick(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	m.View()
	// A wheel over the list, while the detail box has the focus, is ignored.
	lb := scrollBox(m, paneList)
	m = wheelOnly(m, lb.x+1, lb.y+2, false)
	if drewNew(m) {
		t.Error("a wheel over a pane without the focus drew the screen again")
	}
	m = wheelTick(m)
	if drewNew(m) {
		t.Error("a tick with nothing gathered drew the screen again")
	}
	next, _ := m.Update(tea.MouseMsg{X: 1, Y: 1, Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft})
	if drewNew(next.(Model)) {
		t.Error("a mouse release drew the screen again")
	}
}

func TestViewDrawsAgainAfterEveryChange(t *testing.T) {
	t.Parallel()

	base := paneModel(t, paneDetail)
	b := scrollBox(base, paneDetail)
	cases := map[string]func(Model) Model{
		"a tick that scrolls": func(m Model) Model { return wheelTick(wheelOnly(m, b.x+1, b.y+2, false)) },
		"a key":               func(m Model) Model { return press(m, "j") },
		"a click on a row":    func(m Model) Model { lb := scrollBox(m, paneList); return click(m, lb.x+1, lb.y+2) },
		"a reload":            func(m Model) Model { return reloaded(m) },
		"a resize":            func(m Model) Model { return sized(m, m.width-1, m.height) },
		"a clock tick": func(m Model) Model {
			next, _ := m.Update(clockMsg(time.Now()))
			return next.(Model)
		},
		"a flush to another pane": func(m Model) Model {
			m = wheelOnly(m, b.x+1, b.y+2, false)
			m.focus = paneList
			lb := scrollBox(m, paneList)
			return wheelOnly(m, lb.x+1, lb.y+2, false)
		},
	}
	for name, change := range cases {
		m := base
		m.frame = &frameCache{}
		m.View()
		m = wheelOnly(m, b.x+1, b.y+2, false) // leave same set, so the change must clear it
		m = change(m)
		if !drewNew(m) {
			t.Errorf("%s: View gave the old frame back", name)
		}
	}
}
```

If any name above does not exist under that name, look it up with grep in `internal/tui/*_test.go`, use the real name, and say which one you changed in the report. Add `"time"` and the `tea` import to `view_test.go` when they are not there yet.

- [ ] **Step 2: Run them and watch them fail**

Run: `go test ./internal/tui/ -run 'ViewReuses|ViewDrawsAgain'`
Expected: a build failure with `m.frame undefined` and `undefined: frameCache`.

- [ ] **Step 3: Write the implementation**

In `internal/tui/model.go`, next to `detailCache` use:

```go
// frameCache keeps the last frame View drew. A trackpad sends hundreds of
// notches a second, and most of them only add to the pending delta, so the
// screen stays the same. Drawing it again each time kept the TUI busy. The
// model is copied on every update, so the frame sits behind a pointer that
// all the copies share.
type frameCache struct{ s string }
```

Add to `Model`, next to `dcache`:

```go
	frame *frameCache // the last frame View drew, shared by every copy
	same  bool        // true when the last message changed nothing on screen
```

In `New`, next to `dcache: &detailCache{},`, add:

```go
		frame:  &frameCache{},
```

In `WithTheme`, next to `m.dcache = &detailCache{}`, add:

```go
	m.frame = &frameCache{}
```

At the very top of `Update`, before the `switch`, add:

```go
	// Every message may change the screen. The few that surely do not set
	// this back below, so a path nobody thought about always draws again.
	m.same = false
```

In the `wheelTickMsg` case, set `same` when nothing scrolled:

```go
	case wheelTickMsg:
		// Notches belong to the screen they were gathered on, so a reader who
		// moved the item, the tab, the sub-tab or the search before the frame
		// ended never sees them land on the new one.
		if m.wheelDelta != 0 && m.wheelMark == m.mark() {
			m.scrollPane(m.wheelPane, m.wheelDelta)
		} else {
			m.same = true
		}
		m.wheelDelta, m.wheelArmed = 0, false
```

In `mouse`, set `m.same = true` right before each of these returns, and only these:
- the modal guard at the top (`if m.help || m.popup != nil || m.slug != nil || m.searching`);
- `if p != m.focus { return m, nil }` in the wheel case;
- both returns at the end of the wheel case, but only when the flush did not scroll another pane. Track this with a local `flushed := false` that the flush block sets to `true`, then use `m.same = !flushed`;
- the `if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft` return.

Leave the status-line link click, the tab click and the row click as they are, so they always draw.

In `internal/tui/view.go`, rename the current `View` to `draw` and keep its body unchanged. Then add:

```go
// View gives the frame Bubble Tea puts on screen. When the last message
// changed nothing, it hands back the frame it drew last time instead of
// drawing the same screen again.
func (m Model) View() string {
	if m.same && m.frame != nil && m.frame.s != "" {
		return m.frame.s
	}
	s := m.draw()
	if m.frame != nil {
		m.frame.s = s
	}
	return s
}
```

Any existing test that sets model fields by hand after an `Update` that set `same`, and then calls `View`, may now see the old frame. Fix such a test by setting `m.same = false` before its `View` call, and name each one you touched in the report. Never delete or weaken a check.

- [ ] **Step 4: Run the tests and watch them pass**

Run: `go test ./...`, `go vet ./...` and `gofmt -l internal` (the last must print nothing).
Expected: every test passes.

Mutation check: change `View` to always call `draw`. `TestViewReusesTheFrameForANotchThatOnlyGathers` must go red. Then put the change back.

- [ ] **Step 5: Commit**

```bash
git add internal/tui/model.go internal/tui/view.go internal/tui/view_test.go
git commit -m "tui: reuse the last frame when a message changed nothing on screen"
```
