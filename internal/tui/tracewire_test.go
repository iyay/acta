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
	// A notch takes the focus the way a click does, so this one belongs to
	// the list box. The notches gathered for the detail box land with it, and
	// the notch itself only gathers, so the frame stays and the draw after it
	// is not logged again.
	m = wheelOnly(m, lb.x+1, lb.y+2, false)
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
	want := []string{"notch first", "view", "notch gathered", "notch gathered"}
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

// The help, a popup, a slug and a search all sit over the panes and take the
// wheel themselves, so a notch there never reaches a pane. A notch that goes
// nowhere is not a notch of the scroll, so it must not be in the log either.
func TestModelLogsNothingWhenTheWheelIsBlocked(t *testing.T) {
	t.Parallel()

	slug := "bug-"
	for _, c := range []struct {
		name string
		open func(m Model) Model
	}{
		{
			name: "the help",
			open: func(m Model) Model { m.help = true; return m },
		},
		{
			name: "a popup",
			open: func(m Model) Model {
				m.popup = &popup{field: "status", options: []string{"open", "done"}, idx: 0}
				return m
			},
		},
		{
			name: "a slug",
			open: func(m Model) Model { m.slug = &slug; return m },
		},
		{
			name: "a search",
			open: func(m Model) Model { m.searching = true; return m },
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			var log bytes.Buffer
			m := c.open(paneModel(t, paneDetail).WithTrace(NewTracer(&log, time.Now)))
			b := scrollBox(m, paneDetail)
			m = wheelOnly(m, b.x+1, b.y+2, false)
			m = wheelTick(m)
			if log.Len() != 0 {
				t.Errorf("the wheel under %s wrote to the trace: %q", c.name, log.String())
			}
		})
	}
}

// A notch that only gathers keeps the frame the reader already has, so the
// screen is not drawn again and the draw must not be logged a second time.
func TestModelLogsNoDrawWhileTheFrameIsReused(t *testing.T) {
	t.Parallel()

	var log bytes.Buffer
	m := paneModel(t, paneDetail).WithTrace(NewTracer(&log, time.Now))
	b := scrollBox(m, paneDetail)
	m = wheelOnly(m, b.x+1, b.y+2, false) // scrolls at once
	first := m.View()
	m = wheelOnly(m, b.x+1, b.y+2, false) // gathers, so the frame stays
	if second := m.View(); second != first {
		t.Fatalf("the gathered notch drew again:\n%s", second)
	}
	if got := strings.Count(log.String(), "view "); got != 1 {
		t.Errorf("the log has %d draws, want 1: %q", got, log.String())
	}
}

// A nil tracer is what every run without ACTA_TUI_TRACE has, so the wheel and
// the draw must ask it for the time and take its answers without a word.
func TestModelWithANilTracerWheelsAndDraws(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail).WithTrace(nil)
	b := scrollBox(m, paneDetail)
	m = wheelOnly(m, b.x+1, b.y+2, false)
	first := m.View()
	if first == "" {
		t.Fatal("no frame")
	}
	m = wheelOnly(m, b.x+1, b.y+2, false)
	if second := m.View(); second != first {
		t.Errorf("the gathered notch drew again:\n%s", second)
	}
}
