---
id: PLN-0045
created: "2026-09-30"
hash: mc42ys2
started: "2026-09-30"
finished: "2026-09-30"
---
# TUI Scroll at 60fps Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Make trackpad scrolling in the TUI reach 60fps on an idle machine, and leave a trace switch in the repo that proves it.

**Architecture:** A first wheel notch scrolls at once and later notches gather into a 12 ms frame tick that re-arms while scrolling goes on. The renderer runs at 120fps. `draw` joins the wide layout by plain per-line concat instead of lipgloss joins, so each cell is measured once. A tracer behind `ACTA_TUI_TRACE` logs notches, draws and flushes and writes a summary line on quit.

**Tech Stack:** Go, Bubble Tea v1.3.10, lipgloss.

**Spec:** `.acta/specs/2026-09-30-tui-scroll-60fps-design.md`

**Tests:** fast `scripts/test ./internal/tui/ ./internal/cli/`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- No new dependency. Bubble Tea stays at v1.3.10.
- The trace env var is exactly `ACTA_TUI_TRACE`.
- Log lines are exactly `notch <unix_ns> first`, `notch <unix_ns> gathered`, `view <unix_ns> <duration_ns>`, `flush <unix_ns> <bytes>`.
- The summary line is exactly `summary frames=<n> fps=<n> gap_p95_ms=<n> first_p50_ms=<n> first_p95_ms=<n> view_p95_ms=<n>`.
- A nil tracer does no work. A trace file that cannot be opened prints a warning on stderr and the TUI runs with no trace.
- `wheelFrame` is 12 ms. The program runs with `tea.WithFPS(120)`.
- Old wheel tests that expect the first notch to wait for the tick are rewritten for the new rule, never deleted.
- Every existing view test stays green with the same output.
- Comments are plain English a 10-year-old reads back without stopping: short words, one idea each, say why. No marker tags, no Latin, no emoji.
- Test runs inside a task use the narrowest command, never `./...`.
- The build starts only after PLN-0039 (tui-colors) has landed, since both change `internal/tui/view.go`. Branch from the main that holds it.

## File Map

- `internal/tui/trace.go` (create): the `Tracer`, its log lines, the stdout wrapper, and the summary math.
- `internal/tui/trace_test.go` (create): summary math and log line tests with a fake clock.
- `internal/tui/model.go` (modify): wheel timing (task 2); tracer field, `WithTrace`, notch events (task 4).
- `internal/tui/scroll_test.go` (modify): wheel tests for the new timing.
- `internal/tui/view_test.go` (modify): the frame reuse test that assumed the first notch only gathers.
- `internal/tui/view.go` (modify): per-line concat in `draw` (task 3); view events in `View` (task 4).
- `internal/tui/draw_test.go` (create): frame width and old-join equality tests.
- `internal/tui/view_bench_test.go` (modify): the benchmark scrolls one way.
- `internal/tui/tracewire_test.go` (create): the model sends notch and view events to its tracer.
- `internal/cli/tuitrace.go` (create): reads `ACTA_TUI_TRACE`, opens the trace, gives the program options.
- `internal/cli/tuitrace_test.go` (create): empty env, bad path, good path.
- `internal/cli/cli.go` (modify): `runTUI` uses the options and `tea.WithFPS(120)`.

## Waves

- Wave 1: Task 1 (`trace.go`, `trace_test.go`), Task 2 (`model.go`, `scroll_test.go`, `view_test.go`), Task 3 (`view.go`, `draw_test.go`, `view_bench_test.go`). No file is shared.
- Wave 2: Task 4 (`model.go`, `view.go`, `tracewire_test.go`, `internal/cli/tuitrace.go`, `internal/cli/tuitrace_test.go`, `internal/cli/cli.go`). It needs the `Tracer` from task 1 and touches files from tasks 2 and 3.

---

### Task 1: Tracer and summary

**Files:**
- Create: `internal/tui/trace.go`
- Test: `internal/tui/trace_test.go`

**verify:** Every summary number comes only from the events inside the window the spec names, and no input can make the summary panic or divide by zero. List every input shape checked: no events at all, one notch only, notches with no flush after them, flushes before the first notch, views with no notches, a nil tracer on every method.

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `type Tracer struct` (unexported fields)
  - `func NewTracer(w io.Writer, now func() time.Time) *Tracer`
  - `func (t *Tracer) Notch(first bool)` (nil-safe)
  - `func (t *Tracer) View(start time.Time, d time.Duration)` (nil-safe)
  - `func (t *Tracer) Now() time.Time` (nil-safe; zero time when nil)
  - `func (t *Tracer) Output(f *os.File) io.Writer` returns a `*traceFile` that embeds `*os.File`, so Bubble Tea still sees a terminal (`term.File`: Read, Write, Close, Fd)
  - `func (t *Tracer) Summary() string` (nil-safe; "" when nil) returns the summary line without a newline
  - `func (t *Tracer) Close() error` (nil-safe) writes the summary line plus "\n" to the log

- [x] **Step 1: Write the failing tests**

```go
package tui

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// fakeClock hands out the times a test lists, one per call, so no test sleeps.
type fakeClock struct {
	times []time.Time
	i     int
}

func (c *fakeClock) now() time.Time {
	t := c.times[min(c.i, len(c.times)-1)]
	c.i++
	return t
}

func at(ms float64) time.Time {
	return time.Unix(0, 0).Add(time.Duration(ms * float64(time.Millisecond)))
}

// A steady scroll: a first notch at 0, gathered notches up to 100 ms, and a
// flush every 10 ms. The window is 0 to 100 ms, so 10 or 11 flushes fall in it.
func TestTraceSummaryOfASteadyScroll(t *testing.T) {
	t.Parallel()

	var log bytes.Buffer
	c := &fakeClock{}
	tr := NewTracer(&log, c.now)
	c.times = []time.Time{at(0)}
	tr.Notch(true)
	for ms := 10.0; ms <= 100; ms += 10 {
		c.times, c.i = []time.Time{at(ms - 5)}, 0
		tr.Notch(false)
		c.times, c.i = []time.Time{at(ms)}, 0
		tr.flush(5)
	}
	c.times, c.i = []time.Time{at(100)}, 0
	tr.Notch(false)
	tr.View(at(1), 4*time.Millisecond)
	got := tr.Summary()
	want := "summary frames=10 fps=100.0 gap_p95_ms=10.0 first_p50_ms=10.0 first_p95_ms=10.0 view_p95_ms=4.0"
	if got != want {
		t.Errorf("summary is\n%q\nwant\n%q", got, want)
	}
}

func TestTraceSummaryWithNoEventsIsAllZero(t *testing.T) {
	t.Parallel()

	tr := NewTracer(&bytes.Buffer{}, time.Now)
	want := "summary frames=0 fps=0.0 gap_p95_ms=0.0 first_p50_ms=0.0 first_p95_ms=0.0 view_p95_ms=0.0"
	if got := tr.Summary(); got != want {
		t.Errorf("summary is %q, want %q", got, want)
	}
}

// One notch has no length of time, so there is no rate to give.
func TestTraceSummaryWithOneNotchHasNoRate(t *testing.T) {
	t.Parallel()

	c := &fakeClock{times: []time.Time{at(0), at(5)}}
	tr := NewTracer(&bytes.Buffer{}, c.now)
	tr.Notch(true)
	tr.flush(1)
	got := tr.Summary()
	if !strings.Contains(got, "fps=0.0") || !strings.Contains(got, "first_p50_ms=5.0") {
		t.Errorf("one notch then a flush 5 ms later gave %q", got)
	}
}

// A first notch with no flush after it has no latency to count, and a flush
// before the first notch is not part of the scroll.
func TestTraceSummaryIgnoresFlushesOutsideTheScroll(t *testing.T) {
	t.Parallel()

	c := &fakeClock{times: []time.Time{at(0), at(10), at(20), at(30)}}
	tr := NewTracer(&bytes.Buffer{}, c.now)
	tr.flush(1)     // 0 ms, before any notch
	tr.Notch(true)  // 10 ms
	tr.Notch(false) // 20 ms
	tr.Notch(true)  // 30 ms, no flush after it
	got := tr.Summary()
	want := "summary frames=0 fps=0.0 gap_p95_ms=0.0 first_p50_ms=0.0 first_p95_ms=0.0 view_p95_ms=0.0"
	if got != want {
		t.Errorf("summary is %q, want %q", got, want)
	}
}

func TestTraceWritesOneLinePerEvent(t *testing.T) {
	t.Parallel()

	var log bytes.Buffer
	c := &fakeClock{times: []time.Time{at(1), at(2), at(3)}}
	tr := NewTracer(&log, c.now)
	tr.Notch(true)
	tr.Notch(false)
	tr.View(at(2.5), 1500*time.Microsecond)
	tr.flush(42)
	want := "notch 1000000 first\nnotch 2000000 gathered\nview 2500000 1500000\nflush 3000000 42\n"
	if log.String() != want {
		t.Errorf("log is\n%q\nwant\n%q", log.String(), want)
	}
}

func TestTraceCloseAddsTheSummaryLine(t *testing.T) {
	t.Parallel()

	var log bytes.Buffer
	tr := NewTracer(&log, time.Now)
	if err := tr.Close(); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(log.String(), "summary frames=0 ") || !strings.HasSuffix(log.String(), "\n") {
		t.Errorf("close wrote %q", log.String())
	}
}

// A TUI with no trace holds a nil tracer, and every call on it must be free
// and safe.
func TestNilTracerDoesNothing(t *testing.T) {
	t.Parallel()

	var tr *Tracer
	tr.Notch(true)
	tr.View(time.Now(), time.Millisecond)
	if !tr.Now().IsZero() || tr.Summary() != "" || tr.Close() != nil {
		t.Error("a nil tracer did some work")
	}
}

// Bubble Tea only treats the output as a terminal when it has a file
// descriptor, so the wrapper must keep the file's own.
func TestTraceOutputKeepsTheFileAndCountsWrites(t *testing.T) {
	t.Parallel()

	f, err := os.CreateTemp(t.TempDir(), "out")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var log bytes.Buffer
	c := &fakeClock{times: []time.Time{at(7)}}
	tr := NewTracer(&log, c.now)
	w := tr.Output(f)
	tf, ok := w.(interface{ Fd() uintptr })
	if !ok || tf.Fd() != f.Fd() {
		t.Fatal("the wrapper lost the file descriptor")
	}
	if _, err := w.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if log.String() != "flush 7000000 5\n" {
		t.Errorf("log is %q", log.String())
	}
	got, _ := os.ReadFile(f.Name())
	if string(got) != "hello" {
		t.Errorf("file holds %q", got)
	}
}
```

Add `"os"` to the imports of the test file.

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui/ -run 'Trace' -count=1`
Expected: FAIL to build, `undefined: NewTracer`.

- [x] **Step 3: Write the tracer**

```go
package tui

import (
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"sync"
	"time"
)

// Tracer writes down when wheel notches come, how long each draw takes and
// when the screen is written, so a slow scroll can be measured on a real
// terminal. Turn it on with ACTA_TUI_TRACE. The event loop and the renderer
// both write to it, so a lock guards it. A nil tracer does nothing.
type Tracer struct {
	mu      sync.Mutex
	w       io.Writer
	now     func() time.Time
	notches []traceNotch
	views   []time.Duration
	flushes []time.Time
}

type traceNotch struct {
	at    time.Time
	first bool
}

// NewTracer writes its lines to w and reads the time from now, so a test can
// hand it a fake clock.
func NewTracer(w io.Writer, now func() time.Time) *Tracer {
	return &Tracer{w: w, now: now}
}

// Now is the tracer's clock. The model asks it for the start of a draw.
func (t *Tracer) Now() time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.now()
}

// Notch notes one wheel notch. first says no frame tick was waiting, so this
// notch scrolled at once.
func (t *Tracer) Notch(first bool) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	at := t.now()
	t.notches = append(t.notches, traceNotch{at: at, first: first})
	kind := "gathered"
	if first {
		kind = "first"
	}
	fmt.Fprintf(t.w, "notch %d %s\n", at.UnixNano(), kind)
}

// View notes one full draw of the screen.
func (t *Tracer) View(start time.Time, d time.Duration) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.views = append(t.views, d)
	fmt.Fprintf(t.w, "view %d %d\n", start.UnixNano(), d.Nanoseconds())
}

// flush notes one write of the screen to the terminal.
func (t *Tracer) flush(n int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	at := t.now()
	t.flushes = append(t.flushes, at)
	fmt.Fprintf(t.w, "flush %d %d\n", at.UnixNano(), n)
}

// traceFile is the terminal with a tracer on its writes. It keeps the file
// inside, so Bubble Tea still finds the terminal and its size.
type traceFile struct {
	*os.File
	t *Tracer
}

func (f traceFile) Write(p []byte) (int, error) {
	n, err := f.File.Write(p)
	f.t.flush(n)
	return n, err
}

// Output wraps the terminal so each write the renderer makes is noted.
func (t *Tracer) Output(f *os.File) io.Writer {
	return traceFile{File: f, t: t}
}

// Summary sums up the scroll. The scroll runs from the first notch to the
// last one. frames and fps count the writes inside it, gap is the time
// between two of those writes, and first is how long a notch that came with
// no tick waiting took to reach the screen.
func (t *Tracer) Summary() string {
	if t == nil {
		return ""
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	var frames []time.Time
	var fps float64
	if len(t.notches) > 0 {
		start, end := t.notches[0].at, t.notches[len(t.notches)-1].at
		for _, f := range t.flushes {
			if !f.Before(start) && !f.After(end) {
				frames = append(frames, f)
			}
		}
		if span := end.Sub(start).Seconds(); span > 0 {
			fps = float64(len(frames)) / span
		}
	}
	var gaps []time.Duration
	for i := 1; i < len(frames); i++ {
		gaps = append(gaps, frames[i].Sub(frames[i-1]))
	}
	var first []time.Duration
	for _, n := range t.notches {
		if !n.first {
			continue
		}
		// The first write at or after the notch is the one that shows it.
		i := sort.Search(len(t.flushes), func(i int) bool { return !t.flushes[i].Before(n.at) })
		if i < len(t.flushes) {
			first = append(first, t.flushes[i].Sub(n.at))
		}
	}
	return fmt.Sprintf("summary frames=%d fps=%.1f gap_p95_ms=%.1f first_p50_ms=%.1f first_p95_ms=%.1f view_p95_ms=%.1f",
		len(frames), fps, ms(pct(gaps, 0.95)), ms(pct(first, 0.5)), ms(pct(first, 0.95)), ms(pct(t.views, 0.95)))
}

// Close writes the summary as the last line of the log.
func (t *Tracer) Close() error {
	if t == nil {
		return nil
	}
	_, err := fmt.Fprintln(t.w, t.Summary())
	return err
}

// pct is the nearest-rank percentile: the value that p of the list is at or
// below. An empty list gives zero.
func pct(ds []time.Duration, p float64) time.Duration {
	if len(ds) == 0 {
		return 0
	}
	s := append([]time.Duration(nil), ds...)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	return s[max(0, int(math.Ceil(p*float64(len(s))))-1)]
}

func ms(d time.Duration) float64 { return float64(d) / float64(time.Millisecond) }
```

The first test's flushes sit at 10, 20, ... 100 ms, which all fall inside 0 to 100 ms, so frames is 10 and fps is 10 / 0.1 s = 100.0. The first notch at 0 ms reaches the screen at the 10 ms flush, so `first_p50_ms` is 10.0. If the numbers come out different, fix the test's arithmetic only after checking the rule in the spec's pass mark, never the rule.

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/tui/ -run 'Trace' -count=1 -race`
Expected: PASS.

- [x] **Step 5: Format, vet and commit**

```bash
gofmt -l internal/tui && go vet ./internal/tui/
git add internal/tui/trace.go internal/tui/trace_test.go
git commit -m "tui: add a tracer for notches, draws and screen writes"
```

---

### Task 2: First notch scrolls at once

**Files:**
- Modify: `internal/tui/model.go` (the `wheelFrame` const, the `wheelTickMsg` case in `Update`, the wheel branch of `mouse`)
- Modify: `internal/tui/scroll_test.go`
- Modify: `internal/tui/view_test.go` (`TestViewReusesTheFrameForANotchThatOnlyGathers`)

**verify:** A notch never waits for a tick when no tick is armed, and a notch never scrolls on its own while a tick is armed. Every tick with something pending scrolls and arms exactly one next tick, and every tick with nothing pending arms none. The screen mark still drops gathered notches on every path that changed the screen. List every path checked: first notch, gathered notch, tick with delta, tick with delta and a changed screen, empty tick, notch after an empty tick, notch on another pane, wheel with help, popup, slug or search open.

**Interfaces:**
- Consumes: nothing from task 1.
- Produces: `wheelFrame == 12 * time.Millisecond`. `Update(wheelTickMsg{})` returns a non-nil `tea.Cmd` when it scrolled and nil when it did not. Task 4 adds the tracer call at the spot this task marks `first`.

- [x] **Step 1: Write the failing tests**

Add to `internal/tui/scroll_test.go`:

```go
// Scrolling must start on the notch itself. Waiting for the frame tick made
// every scroll start one frame late.
func TestWheelFirstNotchScrollsAtOnce(t *testing.T) {
	t.Parallel()

	for _, p := range []pane{paneList, paneDone, paneDetail} {
		m := paneModel(t, p)
		b := scrollBox(m, p)
		m = wheelOnly(m, b.x+1, b.y+2, false)
		if want := min(wheelStep, m.lastOff(p)); m.off[p] != want {
			t.Errorf("pane %d: the first notch should scroll %d at once, off is %d", p, want, m.off[p])
		}
	}
}

// While scrolling goes on, the tick must come back each frame. When the wheel
// has stopped, the tick must stop too, so no tick runs for nothing.
func TestWheelTickRearmsOnlyWhileScrolling(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	b := scrollBox(m, paneDetail)
	m = wheelOnly(m, b.x+1, b.y+2, false)
	m = wheelOnly(m, b.x+1, b.y+2, false)
	next, cmd := m.Update(wheelTickMsg{})
	if cmd == nil {
		t.Fatal("a tick that scrolled must arm the next tick")
	}
	m = next.(Model)
	next, cmd = m.Update(wheelTickMsg{})
	if cmd != nil {
		t.Error("a tick with nothing gathered must not arm another tick")
	}
	m = next.(Model)
	before := m.off[paneDetail]
	m = wheelOnly(m, b.x+1, b.y+2, false)
	if m.off[paneDetail] == before {
		t.Error("the first notch after the wheel stopped must scroll at once")
	}
}

// Two ticks inside the time the renderer takes for two writes at 120fps
// keep the gap between writes at 16.7 ms or less.
func TestWheelFrameFitsTwoRendererWrites(t *testing.T) {
	if wheelFrame != 12*time.Millisecond {
		t.Errorf("wheelFrame is %v, the spec says 12ms", wheelFrame)
	}
}
```

Rewrite these existing tests for the new rule. Keep each test's name and its point, and change only what the first notch does:
- `TestWheelNotchesInOneFrameScrollOnce`: after the first `wheelOnly`, `off` is `min(wheelStep, lastOff)`. After the second `wheelOnly`, `off` has not moved. After `wheelTick`, `off` is `min(2*wheelStep, lastOff)`. The empty-tick check stays.
- `TestWheelArmsOneTickPerFrame`: the first notch returns a non-nil command. The second notch returns nil. Keep both checks as they are, and add that the first notch already moved `off`.
- `TestWheelUpAndDownInOneFrameCancel`: the down notch now scrolls at once and the up notch gathers. Change the test so the frame it checks is a gathered one: turn down once and tick (as now), then turn down, down, up inside the next frame, where the first of those three gathers because a tick is armed. Check the pane moved by exactly one step from `start` after the tick.
- Every other test in `scroll_test.go` and `view_test.go` that turns the wheel and then asserts a position: run them after step 3. For each one that fails only because the first notch now moves at once, fix the expected number or add a leading notch plus tick, so the notches the test cares about are gathered ones. Write the reason in the test's comment. Delete none.

In `internal/tui/view_test.go`, change `TestViewReusesTheFrameForANotchThatOnlyGathers` so a first notch scrolls (and draws), then a second notch in the same frame is the one checked with `drewNew`:

```go
func TestViewReusesTheFrameForANotchThatOnlyGathers(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	b := scrollBox(m, paneDetail)
	// The first notch scrolls at once, so it draws. The second one waits for
	// the tick and must not draw.
	m = wheelOnly(m, b.x+1, b.y+2, false)
	m.View()
	m = wheelOnly(m, b.x+1, b.y+2, false)
	if drewNew(m) {
		t.Error("a notch that only adds to the delta drew the screen again")
	}
}
```

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui/ -run 'Wheel|ViewReuses' -count=1`
Expected: FAIL. `TestWheelFirstNotchScrollsAtOnce` says off is 0, `TestWheelTickRearmsOnlyWhileScrolling` says the tick did not arm, and `TestWheelFrameFitsTwoRendererWrites` says 16ms.

- [x] **Step 3: Change the wheel timing**

In `internal/tui/model.go`, change the const and its comment:

```go
// wheelFrame is how long wheel notches are gathered before one scroll. A
// trackpad sends hundreds of notches a second, and drawing the screen after
// each one left the TUI far behind the wheel. It is shorter than two renderer
// writes at 120fps, so a steady scroll never skips a write.
const wheelFrame = 12 * time.Millisecond
```

Give the tick command a name so both places use one:

```go
// wheelTickCmd asks for the end of the frame.
func wheelTickCmd() tea.Cmd {
	return tea.Tick(wheelFrame, func(time.Time) tea.Msg { return wheelTickMsg{} })
}
```

Replace the `case wheelTickMsg:` body in `Update`:

```go
	case wheelTickMsg:
		// Notches belong to the screen they were gathered on, so a reader who
		// moved the item, the tab, the sub-tab or the search before the frame
		// ended never sees them land on the new one.
		if m.wheelDelta != 0 && m.wheelMark == m.mark() {
			m.scrollPane(m.wheelPane, m.wheelDelta)
			m.wheelDelta = 0
			// The wheel is still turning, so the next frame needs its tick.
			return m, wheelTickCmd()
		}
		m.same = true
		m.wheelDelta, m.wheelArmed = 0, false
```

In `mouse`, the wheel branch keeps the focus check, the step sign, the mark reset and the flush to another pane as they are. Replace everything from `// The first notch of a frame writes the screen down.` to the end of the wheel case with:

```go
		// With no tick waiting, the wheel had stopped. Scroll now, so the
		// reader sees the pane move on the notch itself, and start a frame
		// for the notches that follow.
		if !m.wheelArmed {
			m.scrollPane(p, step)
			m.wheelPane = p
			m.wheelArmed = true
			return m, wheelTickCmd()
		}
		// The first gathered notch of a frame writes the screen down. The
		// notches after it share that screen and must not write over it.
		if m.wheelDelta == 0 {
			m.wheelMark = m.mark()
		}
		m.wheelPane = p
		m.wheelDelta += step
		// A flush to another pane moved that pane, so only a plain gather
		// keeps the old frame.
		m.same = !flushed
		return m, nil
```

`scrollPane` already leaves `m.same` false, since only the gather paths set it. Check that `m.same` is false after the first notch; if it is not, set `m.same = false` in that branch.

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/tui/ -run 'Wheel|ViewReuses|ViewDraws' -count=1`
Expected: PASS. Then run the whole package the short way: `scripts/test ./internal/tui/`. Expected: PASS.

- [x] **Step 5: Format, vet and commit**

```bash
gofmt -l internal/tui && go vet ./internal/tui/
git add internal/tui/model.go internal/tui/scroll_test.go internal/tui/view_test.go
git commit -m "tui: scroll on the first wheel notch and keep the frame tick going while the wheel turns"
```

---

### Task 3: Join the wide layout line by line

**Files:**
- Modify: `internal/tui/view.go` (`draw`)
- Create: `internal/tui/draw_test.go`
- Modify: `internal/tui/view_bench_test.go`

**verify:** For every screen size and every board the tests cover, the new `draw` gives the same bytes as the lipgloss join did, and every line of the frame is exactly `m.width` cells wide. List every size and board checked, and name the ones where a box has no room or the columns have different heights.

**Interfaces:**
- Consumes: `Model.geometry()` fields `wide`, `leftW`, `side`, `detail`; `Model.paneView(pane, box) string`; `fit`, `tabRows`.
- Produces: no new names. `draw()` keeps its signature.

- [x] **Step 1: Write the failing tests**

Create `internal/tui/draw_test.go`:

```go
package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

// oldBody is the body the way draw built it before, with lipgloss joining the
// boxes. The new draw must give the same bytes.
func oldBody(m Model) string {
	g := m.geometry()
	var column []string
	for p := range g.side {
		if drawn := m.paneView(pane(p), g.side[p]); drawn != "" {
			column = append(column, drawn)
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left, column...), m.paneView(paneDetail, g.detail))
}

func sized(m Model, w, h int) Model {
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return next.(Model)
}

var drawSizes = [][2]int{{60, 10}, {60, 40}, {80, 24}, {120, 40}, {200, 55}, {61, 5}, {300, 12}}

// The wide layout is joined by hand now. It must be the same screen the
// lipgloss join gave, on every size, with the detail box focused or not.
func TestDrawJoinMatchesTheOldJoin(t *testing.T) {
	t.Parallel()

	for _, p := range []pane{paneList, paneDone, paneDetail} {
		for _, s := range drawSizes {
			m := sized(paneModel(t, p), s[0], s[1])
			if !m.geometry().wide {
				continue
			}
			want := m.frameFrom(strings.Split(oldBody(m), "\n"), false)
			if got := m.draw(); got != want {
				t.Errorf("pane %d, %dx%d: the frame changed", p, s[0], s[1])
			}
		}
	}
}

// Every line must fill the terminal exactly. A short line leaves old text on
// screen, and a long one wraps and breaks the whole layout.
func TestDrawEveryLineIsTheScreenWidth(t *testing.T) {
	t.Parallel()

	uni := press(boardModel(t, "# Plan ✓ ünïcödé 日本語 "+strings.Repeat("wide ", 40)+"\n\n"+strings.Repeat("日本語のテキスト ", 30)+"\n"), tabKey(tabPlans))
	for _, base := range []Model{paneModel(t, paneDetail), uni} {
		for _, s := range append(drawSizes, [2]int{40, 20}, [2]int{59, 20}) {
			m := sized(base, s[0], s[1])
			for i, ln := range strings.Split(m.draw(), "\n") {
				if got := lipgloss.Width(ln); got != s[0] {
					t.Errorf("%dx%d line %d is %d wide", s[0], s[1], i, got)
				}
			}
		}
	}
}
```

`frameFrom(body []string, exact bool)` is the part of `draw` after the body lines are made: add the tab rows, cut to the height, fit, add the status line, paint. `exact` says the body lines are already the screen width and need no fit. Step 3 splits it out of `draw` so the test and `draw` share it and only the join differs.

In `internal/tui/view_bench_test.go`, change the loop so the wheel always turns down, and the pane goes back to the top when it reaches the end:

```go
// BenchmarkWheelFrame is one frame of trackpad scrolling in the detail box:
// ten notches, one tick, one draw. The wheel always turns down, so every frame
// really scrolls; at the end the box goes back to the top.
func BenchmarkWheelFrame(b *testing.B) {
	t := &testing.T{}
	m := paneModel(t, paneDetail)
	box := scrollBox(m, paneDetail)
	for b.Loop() {
		if m.off[paneDetail] >= m.lastOff(paneDetail) {
			m.off[paneDetail] = 0
		}
		for range 10 {
			m = wheelOnly(m, box.x+1, box.y+2, false)
		}
		m = wheelTick(m)
		m.View()
	}
}
```

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui/ -run 'TestDraw' -count=1`
Expected: FAIL to build, `m.frameFrom undefined`.

- [x] **Step 3: Change `draw`**

In `internal/tui/view.go`, `draw` becomes:

```go
// draw paints the whole screen: the boxes, the status line, and any popup on
// top of them.
func (m Model) draw() string {
	g := m.geometry()
	if !g.wide {
		return m.frameFrom(strings.Split(m.paneView(m.focus, g.full), "\n"), false)
	}
	// A box with no room draws nothing, so it takes no line either and the
	// boxes below it keep the place the geometry gave them.
	var left []string
	for p := range g.side {
		if drawn := m.paneView(pane(p), g.side[p]); drawn != "" {
			left = append(left, strings.Split(drawn, "\n")...)
		}
	}
	right := strings.Split(m.paneView(paneDetail, g.detail), "\n")
	// Every box line is already padded to its box width, so the two columns
	// are glued with no measuring. Measuring each cell again was most of the
	// time a frame took. A column that runs out gets blank lines of its width.
	body := make([]string, max(len(left), len(right)))
	for i := range body {
		l, r := strings.Repeat(" ", g.leftW), strings.Repeat(" ", g.detail.w)
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}
		body[i] = l + r
	}
	return m.frameFrom(body, true)
}

// frameFrom puts the tab box on top of the body lines, cuts them to the
// screen, adds the status line and lays any popup over it all. exact says
// the body lines are already the screen width.
func (m Model) frameFrom(body []string, exact bool) string {
	// The status line always has a row of its own, so the tab box and the
	// panes below it share the rest, and a terminal with fewer rows than the
	// box needs simply loses the bottom of the box.
	h := max(1, m.height-1)
	tabs := m.tabRows(m.width)
	lines := append(tabs, body...)
	if len(lines) > h {
		lines = lines[:h]
	}
	for i, ln := range lines {
		// The wide body is glued from boxes that are already the screen
		// width, so only the tab rows and the narrow body need fitting.
		if exact && i >= len(tabs) {
			continue
		}
		lines[i] = fit(ln, m.width)
	}
	out := strings.Join(lines, "\n")
	line := fit(m.statusLine(), m.width)
	if m.popupBox() != "" {
		// The box can be as tall as the screen, so it is laid over the body
		// and the status line together. Cover greys every line it gets and
		// strips the hyperlinks, which are off while a popup is open.
		return m.styles.paintFrame(m.cover(out + "\n" + line))
	}
	return m.styles.paintFrame(out + "\n" + line)
}
```

The two tests are the judges, and they are never loosened:
- If `TestDrawEveryLineIsTheScreenWidth` fails, a box line is not exactly its box width. Find that box in `paneView` and pad the line there. Do not put a `fit` back on the wide body.
- If `TestDrawJoinMatchesTheOldJoin` fails for one size (for example a box with no room, or columns of different heights), change the glue in `draw` until the bytes match. Do not change `oldBody`.

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/tui/ -run 'TestDraw|TestView' -count=1`
Expected: PASS. Then `scripts/test ./internal/tui/`. Expected: PASS. Then run `go test ./internal/tui/ -run '^$' -bench BenchmarkWheelFrame -benchmem -count 3` and put the ns/op numbers from before and after this task in the commit body.

- [x] **Step 5: Format, vet and commit**

```bash
gofmt -l internal/tui && go vet ./internal/tui/
git add internal/tui/view.go internal/tui/draw_test.go internal/tui/view_bench_test.go
git commit -m "tui: glue the wide layout line by line so each cell is measured once"
```

---

### Task 4: Wire the trace into the TUI and the CLI

**Files:**
- Modify: `internal/tui/model.go` (a `trace *Tracer` field, `WithTrace`, a `Notch` call in the wheel branch of `mouse`)
- Modify: `internal/tui/view.go` (`View`)
- Create: `internal/tui/tracewire_test.go`
- Create: `internal/cli/tuitrace.go`
- Create: `internal/cli/tuitrace_test.go`
- Modify: `internal/cli/cli.go` (`runTUI`)

**verify:** With `ACTA_TUI_TRACE` unset or empty, the TUI runs exactly as before: no file, no wrapper, no events. With a path that cannot be opened, the TUI still opens and a warning goes to stderr. With a good path, every notch that reaches the focused pane logs once, with `first` exactly when it scrolled at once, every full draw logs once, a reused frame logs nothing, and the summary line is the last line of the file after the TUI quits. List every path checked.

**Interfaces:**
- Consumes: from task 1, `NewTracer`, `(*Tracer).Notch`, `View`, `Now`, `Output`, `Close`. From task 2, the `if !m.wheelArmed` branch in `mouse`.
- Produces:
  - `func (m Model) WithTrace(t *Tracer) Model`
  - in package `cli`: `func tuiTrace(path string, stdout *os.File, stderr io.Writer) (*tui.Tracer, []tea.ProgramOption, func())`

- [x] **Step 1: Write the failing tests**

Create `internal/tui/tracewire_test.go`:

```go
package tui

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestModelSendsNotchesAndDrawsToItsTracer(t *testing.T) {
	t.Parallel()

	var log bytes.Buffer
	m := paneModel(t, paneDetail).WithTrace(NewTracer(&log, time.Now))
	b := scrollBox(m, paneDetail)
	m = wheelOnly(m, b.x+1, b.y+2, false) // scrolls at once
	m.View()
	m = wheelOnly(m, b.x+1, b.y+2, false) // gathers, keeps the frame
	m.View()
	lb := scrollBox(m, paneList)
	m = wheelOnly(m, lb.x+1, lb.y+2, false) // over a pane without the focus
	var kinds []string
	for _, ln := range strings.Split(strings.TrimSpace(log.String()), "\n") {
		f := strings.Fields(ln)
		kinds = append(kinds, strings.Join(append(f[:1:1], f[2:]...), " "))
	}
	// The view line keeps its duration field, so only its first word counts.
	for i, k := range kinds {
		if strings.HasPrefix(k, "view ") {
			kinds[i] = "view"
		}
	}
	want := []string{"notch first", "view", "notch gathered"}
	if strings.Join(kinds, ",") != strings.Join(want, ",") {
		t.Errorf("events are %v, want %v", kinds, want)
	}
}

func TestModelWithNoTracerLogsNothing(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	b := scrollBox(m, paneDetail)
	m = wheelOnly(m, b.x+1, b.y+2, false)
	if m.View() == "" {
		t.Fatal("no frame")
	}
}
```

Create `internal/cli/tuitrace_test.go`:

```go
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTuiTraceOffWhenThePathIsEmpty(t *testing.T) {
	var errs bytes.Buffer
	tr, opts, done := tuiTrace("", os.Stdout, &errs)
	done()
	if tr != nil || len(opts) != 0 || errs.Len() != 0 {
		t.Errorf("an empty path turned the trace on: %v %d %q", tr, len(opts), errs.String())
	}
}

func TestTuiTraceWarnsAndRunsWhenTheFileCannotOpen(t *testing.T) {
	var errs bytes.Buffer
	bad := filepath.Join(t.TempDir(), "no-such-dir", "f.log")
	tr, opts, done := tuiTrace(bad, os.Stdout, &errs)
	done()
	if tr != nil || len(opts) != 0 {
		t.Error("a trace that could not open still wrapped the output")
	}
	if !strings.Contains(errs.String(), "ACTA_TUI_TRACE") {
		t.Errorf("the warning does not name the switch: %q", errs.String())
	}
}

func TestTuiTraceWritesTheSummaryOnDone(t *testing.T) {
	var errs bytes.Buffer
	path := filepath.Join(t.TempDir(), "f.log")
	tr, opts, done := tuiTrace(path, os.Stdout, &errs)
	if tr == nil || len(opts) != 1 {
		t.Fatalf("a good path gave tracer %v and %d options", tr, len(opts))
	}
	tr.Notch(true)
	done()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(got)), "\n")
	if !strings.HasPrefix(lines[0], "notch ") || !strings.HasPrefix(lines[len(lines)-1], "summary frames=") {
		t.Errorf("log is %q", got)
	}
}
```

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui/ -run 'TestModel(Sends|WithNo)' -count=1 && go test ./internal/cli/ -run 'TuiTrace' -count=1`
Expected: FAIL to build, `m.WithTrace undefined` and `undefined: tuiTrace`.

- [x] **Step 3: Wire it**

In `internal/tui/model.go`, add the field next to `frame`:

```go
	trace    *Tracer      // notes wheel notches and draws; nil when off
```

and the setter next to the other `With` methods:

```go
// WithTrace makes the model note its wheel notches and draws in t.
func (m Model) WithTrace(t *Tracer) Model {
	m.trace = t
	return m
}
```

In `mouse`, after the focus check (so a wheel over another pane logs nothing) and before the mark reset, add:

```go
		m.trace.Notch(!m.wheelArmed)
```

In `internal/tui/view.go`, `View` becomes:

```go
func (m Model) View() string {
	if m.same && m.frame != nil && m.frame.s != "" {
		return m.frame.s
	}
	start := m.trace.Now()
	s := m.draw()
	if m.trace != nil {
		m.trace.View(start, m.trace.Now().Sub(start))
	}
	if m.frame != nil {
		m.frame.s = s
	}
	return s
}
```

Create `internal/cli/tuitrace.go`:

```go
package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/iyay/acta/internal/tui"
)

// tuiTrace turns on the TUI trace when path names a file. It gives the
// tracer, the program option that notes each screen write, and a func that
// writes the summary and closes the file when the TUI is done. A file that
// cannot open only costs the trace: the TUI still runs.
func tuiTrace(path string, stdout *os.File, stderr io.Writer) (*tui.Tracer, []tea.ProgramOption, func()) {
	if path == "" {
		return nil, nil, func() {}
	}
	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(stderr, "ACTA_TUI_TRACE: trace is off: %v\n", err)
		return nil, nil, func() {}
	}
	tr := tui.NewTracer(f, time.Now)
	done := func() {
		if err := tr.Close(); err != nil {
			fmt.Fprintf(stderr, "ACTA_TUI_TRACE: summary not written: %v\n", err)
		}
		f.Close()
	}
	return tr, []tea.ProgramOption{tea.WithOutput(tr.Output(stdout))}, done
}
```

In `internal/cli/cli.go`, `runTUI` builds the program like this:

```go
	m := tui.New(cfg, b, dark).WithTheme(voiceTheme(), dark).WithVersion(version).WithLoad(func() (*board.Board, error) { return trees.Load(cfg) })
	// 120fps halves how long a new frame waits to reach the screen.
	opts := []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseCellMotion(), tea.WithFPS(120)}
	tr, traceOpts, traceDone := tuiTrace(os.Getenv("ACTA_TUI_TRACE"), os.Stdout, stderr)
	defer traceDone()
	m = m.WithTrace(tr)
	p := tea.NewProgram(m, append(opts, traceOpts...)...)
```

`tea.WithFPS(120)` cannot be read back from a `tea.Program`, so no unit test sees it. The live trace in the landing report checks it: `first_p95_ms` above 16.7 with `view_p95_ms` near 4 means the option is missing.

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/tui/ -run 'TestModel(Sends|WithNo)|Trace' -count=1 -race && go test ./internal/cli/ -run 'TuiTrace' -count=1`
Expected: PASS. Then `scripts/test ./internal/tui/ ./internal/cli/`. Expected: PASS.

- [x] **Step 5: Format, vet and commit**

```bash
gofmt -l internal && go vet ./internal/tui/ ./internal/cli/
git add internal/tui/model.go internal/tui/view.go internal/tui/tracewire_test.go internal/cli/tuitrace.go internal/cli/tuitrace_test.go internal/cli/cli.go
git commit -m "tui: log notches, draws and screen writes when ACTA_TUI_TRACE is set, and run the renderer at 120fps"
```

---

## After landing (landing report, not tasks)

1. Rebuild the PATH binary: `go install ./cmd/acta`, then restart the TUI.
2. The user runs `ACTA_TUI_TRACE=/tmp/f.log acta` in another tab, scrolls the detail pane steadily for 5 s on an idle machine, quits with `q`, and runs `tail -1 /tmp/f.log`.
3. Pass: `gap_p95_ms` ≤ 16.7, `fps` ≥ 55, `first_p95_ms` ≤ 16.7. A fail opens the follow-up for a per-pane render cache (spec, Out of scope).
