package tui

import (
	"bytes"
	"math/rand/v2"
	"os"
	"strings"
	"sync"
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

// A nil tracer has no tracer inside the terminal, so writing to the terminal
// it handed out must still work and must write nothing down.
func TestNilTracerOutputStillWrites(t *testing.T) {
	t.Parallel()

	f, err := os.CreateTemp(t.TempDir(), "out")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var tr *Tracer
	w := tr.Output(f)
	tf, ok := w.(interface{ Fd() uintptr })
	if !ok || tf.Fd() != f.Fd() {
		t.Fatal("the wrapper lost the file descriptor")
	}
	if _, err := w.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Errorf("file holds %q", got)
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

// A scroll can draw many times and never write to the terminal. Every number
// must still come out, and none of them may need a rate.
func TestTraceSummaryWithViewsButNoFlush(t *testing.T) {
	t.Parallel()

	c := &fakeClock{times: []time.Time{at(0), at(10), at(20), at(100)}}
	tr := NewTracer(&bytes.Buffer{}, c.now)
	tr.Notch(true) // 0 ms
	for ms := 1.0; ms <= 5; ms++ {
		c.times, c.i = []time.Time{at(ms * 20)}, 0
		tr.View(at(ms), 2*time.Millisecond)
	}
	tr.View(at(50), 8*time.Millisecond)
	got := tr.Summary()
	want := "summary frames=0 fps=0.0 gap_p95_ms=0.0 first_p50_ms=0.0 first_p95_ms=0.0 view_p95_ms=8.0"
	if got != want {
		t.Errorf("summary is %q, want %q", got, want)
	}
}

// The first and the last notch count as inside the window, so a flush that
// lands on either edge is a frame.
func TestTraceSummaryCountsFlushesOnTheNotchEdges(t *testing.T) {
	t.Parallel()

	c := &fakeClock{times: []time.Time{at(0), at(50), at(100), at(0), at(50), at(100)}}
	tr := NewTracer(&bytes.Buffer{}, c.now)
	tr.Notch(true)  // 0 ms
	tr.Notch(false) // 50 ms
	tr.Notch(false) // 100 ms
	tr.flush(1)     // 0 ms, on the first notch
	tr.flush(1)     // 50 ms
	tr.flush(1)     // 100 ms, on the last notch
	got := tr.Summary()
	want := "summary frames=3 fps=30.0 gap_p95_ms=50.0 first_p50_ms=0.0 first_p95_ms=0.0 view_p95_ms=0.0"
	if got != want {
		t.Errorf("summary is %q, want %q", got, want)
	}
}

// Draws can be handed a start that sits after the draw, and one that sits
// very far from it. The list still has to come out sorted, and the worst
// draw is what the p95 gives.
func TestTraceSummaryWithAHugeDraw(t *testing.T) {
	t.Parallel()

	c := &fakeClock{times: []time.Time{at(0)}}
	tr := NewTracer(&bytes.Buffer{}, c.now)
	tr.Notch(true)
	tr.View(at(10), -3*time.Millisecond)
	tr.View(at(20), 10*time.Hour)
	tr.View(at(30), 1*time.Millisecond)
	got := tr.Summary()
	want := "summary frames=0 fps=0.0 gap_p95_ms=0.0 first_p50_ms=0.0 first_p95_ms=0.0 view_p95_ms=36000000.0"
	if got != want {
		t.Errorf("summary is %q, want %q", got, want)
	}
}

// A clock that went backwards gives a draw a negative length. It sorts below
// the others, and the p95 still has to name a number.
func TestTraceSummaryWithOneNegativeDraw(t *testing.T) {
	t.Parallel()

	c := &fakeClock{times: []time.Time{at(0)}}
	tr := NewTracer(&bytes.Buffer{}, c.now)
	tr.Notch(true)
	tr.View(at(10), -3*time.Millisecond)
	got := tr.Summary()
	want := "summary frames=0 fps=0.0 gap_p95_ms=0.0 first_p50_ms=0.0 first_p95_ms=0.0 view_p95_ms=-3.0"
	if got != want {
		t.Errorf("summary is %q, want %q", got, want)
	}
}

// Every notch lands on the same instant, so the scroll has no length and the
// rate has nothing to divide by. The first notches still wait for the writes
// that come after them.
func TestTraceSummaryWithNoTimeBetweenTheNotches(t *testing.T) {
	t.Parallel()

	c := &fakeClock{times: []time.Time{at(0), at(0), at(2), at(4), at(6)}}
	tr := NewTracer(&bytes.Buffer{}, c.now)
	tr.Notch(true) // 0 ms
	tr.Notch(true) // 0 ms again, no time passed
	tr.flush(1)    // 2 ms, after the whole scroll
	tr.flush(1)    // 4 ms
	tr.flush(1)    // 6 ms
	got := tr.Summary()
	want := "summary frames=0 fps=0.0 gap_p95_ms=0.0 first_p50_ms=2.0 first_p95_ms=2.0 view_p95_ms=0.0"
	if got != want {
		t.Errorf("summary is %q, want %q", got, want)
	}
}

// A first notch that comes with the writes already done has to wait for the
// next one, and the last notch in the log has none left to wait for.
func TestTraceSummaryWithFirstNotchesAtBothEnds(t *testing.T) {
	t.Parallel()

	c := &fakeClock{times: []time.Time{at(0), at(5), at(10), at(20), at(30)}}
	tr := NewTracer(&bytes.Buffer{}, c.now)
	tr.Notch(true) // 0 ms, no write yet
	tr.flush(1)    // 5 ms
	tr.Notch(true) // 10 ms
	tr.flush(1)    // 20 ms
	tr.Notch(true) // 30 ms, no write after it
	got := tr.Summary()
	want := "summary frames=2 fps=66.7 gap_p95_ms=15.0 first_p50_ms=5.0 first_p95_ms=10.0 view_p95_ms=0.0"
	if got != want {
		t.Errorf("summary is %q, want %q", got, want)
	}
}

// A scroll with one long stall among short gaps. The median gap stays small
// while the p95 names the stall, so the two must not be the same number.
func TestTraceSummaryGapP95IsNotTheMedian(t *testing.T) {
	t.Parallel()

	c := &fakeClock{times: []time.Time{at(0), at(10), at(20), at(30), at(40), at(200), at(210), at(220)}}
	tr := NewTracer(&bytes.Buffer{}, c.now)
	tr.Notch(false) // 0 ms, a gathered notch, so the first gap is not a wait
	tr.flush(1)     // 10 ms
	tr.flush(1)     // 20 ms
	tr.flush(1)     // 30 ms
	tr.flush(1)     // 40 ms
	tr.flush(1)     // 200 ms, after a 160 ms stall
	tr.flush(1)     // 210 ms
	tr.flush(1)     // 220 ms
	tr.Notch(false) // 220 ms, the last notch closes the window
	got := tr.Summary()
	want := "summary frames=7 fps=31.8 gap_p95_ms=160.0 first_p50_ms=0.0 first_p95_ms=0.0 view_p95_ms=0.0"
	if got != want {
		t.Errorf("summary is %q, want %q", got, want)
	}
}

// The event loop and the renderer write to the tracer at the same time, so
// the lock has to hold. The race detector is the check.
func TestTraceTakesWritesFromTwoGoroutines(t *testing.T) {
	t.Parallel()

	var log syncBuffer
	var clock sync.Mutex
	n := 0
	tr := NewTracer(&log, func() time.Time {
		clock.Lock()
		defer clock.Unlock()
		n++
		return at(float64(n) / 2)
	})

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		for range 500 {
			tr.Notch(true)
			tr.View(at(1), time.Millisecond)
		}
	}()
	go func() {
		defer wg.Done()
		for range 500 {
			tr.flush(3)
		}
	}()
	go func() {
		defer wg.Done()
		for range 500 {
			// Reading the summary while the others write is a normal thing
			// for the TUI to do, so the lock has to cover it too.
			if !strings.HasPrefix(tr.Summary(), "summary frames=") {
				t.Errorf("summary is %q", tr.Summary())
				return
			}
		}
	}()
	wg.Wait()

	got := tr.Summary()
	if !strings.HasPrefix(got, "summary frames=") {
		t.Errorf("summary is %q", got)
	}
	if lines := strings.Count(log.String(), "\n"); lines != 1500 {
		t.Errorf("log has %d lines, want 1500", lines)
	}
}

// Every kind of scroll has to end in a summary line. This walks shapes a
// person would not list by hand, and the run is the check: a panic or a
// divide by zero fails the test.
func TestTraceSummarySurvivesAnyShape(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewPCG(7, 11))
	for shape := range 400 {
		var log bytes.Buffer
		clock := 0.0
		tr := NewTracer(&log, func() time.Time {
			// Time can jump backwards, stand still, or leap far ahead.
			clock += r.NormFloat64() * 7
			if r.IntN(20) == 0 {
				clock = -clock
			}
			return at(clock)
		})
		for range r.IntN(6) {
			tr.Notch(r.IntN(2) == 0)
		}
		for range r.IntN(6) {
			tr.View(at(clock), time.Duration(r.IntN(2000)-1000)*time.Millisecond)
		}
		for range r.IntN(6) {
			tr.flush(r.IntN(4096))
		}
		got := tr.Summary()
		if strings.Contains(got, "NaN") || strings.Contains(got, "Inf") {
			t.Fatalf("shape %d gave %q", shape, got)
		}
		if !strings.HasPrefix(got, "summary frames=") {
			t.Fatalf("shape %d gave %q", shape, got)
		}
	}
}

// syncBuffer is a log that grows while two goroutines write to it. The tracer
// holds its lock while it writes, so the log needs no lock of its own.
type syncBuffer struct{ b bytes.Buffer }

func (s *syncBuffer) Write(p []byte) (int, error) { return s.b.Write(p) }

func (s *syncBuffer) String() string { return s.b.String() }
