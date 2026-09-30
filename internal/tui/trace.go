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
	if t == nil {
		return
	}
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
