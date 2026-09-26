package tui

import (
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"pm-board/internal/board"
	"pm-board/internal/config"
)

func sized(m Model, w, h int) Model {
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return next.(Model)
}

// ansi strips the color codes a style leaves behind, so a test can read the
// screen the way a person does.
var ansi = regexp.MustCompile("\x1b\\[[0-9;]*[A-Za-z]")

func plain(s string) string { return ansi.ReplaceAllString(s, "") }

// body cuts the pane walls off a screen line and drops the padding on the
// right, so a test sees only what the pane holds.
func body(line string) string {
	b := plain(line)
	b = strings.TrimPrefix(b, "│")
	b = strings.TrimSuffix(b, "│")
	return strings.TrimRight(b, " ")
}

// column cuts one pane out of the whole screen, counted in cells.
func column(v string, x, w int) []string {
	var out []string
	for _, ln := range strings.Split(v, "\n") {
		r := []rune(plain(ln))
		if x >= len(r) {
			continue
		}
		out = append(out, string(r[x:min(len(r), x+w)]))
	}
	return out
}

// clocked pins the clock of the model, so a test can read the time it draws.
func clocked(m Model, hh, mm int) Model {
	m.now = time.Date(2026, 9, 27, hh, mm, 0, 0, time.UTC)
	return m
}

// linesFor finds the screen line that holds a piece of text.
func lineFor(t *testing.T, v, want string) ([]string, int) {
	t.Helper()
	lines := strings.Split(v, "\n")
	for i, ln := range lines {
		if strings.Contains(ln, want) {
			return lines, i
		}
	}
	t.Fatalf("no line holds %q", want)
	return nil, -1
}

func TestViewShowsTheThreePanes(t *testing.T) {
	m := sized(press(clocked(newModel(t), 20, 46), "j"), 200, 40)
	v := m.View()
	for _, want := range []string{
		"[1]─Specs", "─ Plans", "─ Tasks", "─ Bugs",
		"[2]─Done", "─ Dropped",
		"[3]─Detail",
		"pmb · basic · live · 20:46",
		"specs/2026-09-20-alpha  Alpha story",
		"in-progress · 1/2",
		"ID        : specs/2026-09-20-alpha",
		"FILE      : .pm/specs/2026-09-20-alpha.md",
		"Some text.",
	} {
		if !strings.Contains(v, want) {
			t.Errorf("view is missing %q", want)
		}
	}
	for _, line := range strings.Split(v, "\n") {
		if w := lipgloss.Width(line); w > 200 {
			t.Fatalf("line wider than the terminal (%d): %q", w, line)
		}
	}
}

func TestViewNeverOverflowsAnyWindow(t *testing.T) {
	for w := 30; w <= 200; w++ {
		for h := 10; h <= 60; h++ {
			base := sized(newModel(t), w, h)
			for name, m := range map[string]Model{
				"open":   base,
				"detail": press(base, "3"),
				"help":   press(base, "?"),
			} {
				for _, line := range strings.Split(sized(m, w, h).View(), "\n") {
					if got := lipgloss.Width(line); got > w {
						t.Fatalf("%dx%d %s: line is %d cells wide, the window is %d: %q", w, h, name, got, w, line)
					}
				}
			}
		}
	}
}

func TestViewNarrowShowsOnlyTheFocusedPane(t *testing.T) {
	for _, tc := range []struct {
		keys []string
		want string
		gone string
	}{
		{nil, "[1]─Specs", "[2]─"},
		{[]string{"2"}, "[2]─Done", "[1]─"},
		{[]string{"3"}, "[3]─Detail", "[1]─"},
	} {
		m := sized(press(newModel(t), tc.keys...), 40, 20)
		v := m.View()
		if !strings.Contains(v, tc.want) {
			t.Errorf("after %v the view is missing %q", tc.keys, tc.want)
		}
		if strings.Contains(v, tc.gone) {
			t.Errorf("after %v the view still shows %q", tc.keys, tc.gone)
		}
	}
}

func TestViewRowsTakeTwoLinesAndABlank(t *testing.T) {
	m := sized(press(newModel(t), "j"), 200, 40)
	v := m.View()
	g := m.geometry()
	col := column(v, g.open.x, g.open.w)
	i := -1
	for n, ln := range col {
		if strings.Contains(ln, "specs/2026-09-20-alpha  Alpha story") {
			i = n
			break
		}
	}
	if i < 0 {
		t.Fatalf("no row holds the alpha story:\n%s", strings.Join(col, "\n"))
	}
	if got := body(col[i]); !strings.HasPrefix(got, "specs/2026-09-20-alpha  Alpha story") {
		t.Errorf("a row should start with the short ID, got %q", got)
	}
	if got := body(col[i+1]); got != "  in-progress · 1/2" {
		t.Errorf("the meta line is %q", got)
	}
	if got := body(col[i+2]); got != "" {
		t.Errorf("a row should end with a blank line, got %q", got)
	}
	// The row above shows its status without progress, because it has no tasks.
	if got := body(col[i-2]); got != "  draft" {
		t.Errorf("the meta line of the row above is %q", got)
	}
}

func TestViewRowsClampLongTitles(t *testing.T) {
	m := sized(press(newModel(t), "j"), 60, 20)
	v := m.View()
	if !strings.Contains(v, "specs/2026-09-22-beta  Be…") {
		t.Error("a row too long for the pane should be clamped with …")
	}
	lines, i := lineFor(t, v, "…")
	if got := body(lines[i]); !strings.HasPrefix(got, "specs/2026-09-22-beta") {
		t.Errorf("the clamp should keep the short ID first, got %q", got)
	}
}

func TestViewMetaLineShowsTheAgentWhenThereIsOne(t *testing.T) {
	m := press(newModel(t), "j")
	for _, it := range m.board.Items {
		if it.ID == "specs/2026-09-20-alpha" {
			it.Agent = "omp"
		}
	}
	v := sized(m, 120, 40).View()
	if !strings.Contains(v, "in-progress · 1/2 · omp") {
		t.Error("the meta line should hold the agent")
	}
}

func TestViewDetailHeaderAlignsItsColons(t *testing.T) {
	var colon = regexp.MustCompile(`^([A-Z]+) +: `)
	m := sized(press(clocked(newModel(t), 20, 46), "j"), 120, 40)
	g := m.geometry()
	var at []int
	first := ""
	for _, seg := range column(m.View(), g.detail.x, g.detail.w) {
		b := strings.TrimRight(strings.TrimPrefix(strings.TrimSuffix(seg, "│"), "│"), " ")
		if !colon.MatchString(b) {
			continue
		}
		if first == "" {
			first = b
		}
		at = append(at, strings.Index(b, ":"))
	}
	if len(at) < 3 {
		t.Fatalf("the detail header should hold several labels, found %d", len(at))
	}
	for _, c := range at {
		if c != at[0] {
			t.Fatalf("the colons do not line up: %v", at)
		}
	}
	if !strings.HasPrefix(first, "ID") {
		t.Errorf("the first label should be ID, got %q", first)
	}
}

func TestViewDetailLeavesEmptyLinesOut(t *testing.T) {
	m := sized(press(newModel(t), "]"), 120, 40)
	g := m.geometry()
	detail := strings.Join(column(m.View(), g.detail.x, g.detail.w), "\n")
	if strings.Contains(detail, "SPEC") {
		t.Error("a plan with no spec should not show a SPEC line")
	}
	if strings.Contains(detail, "WORKTREE") || strings.Contains(detail, "AGENT") {
		t.Error("a line with no value should be left out")
	}
	// The plan next to it does link a spec, so its line has to be there.
	m = sized(press(m, "j"), 120, 40)
	g = m.geometry()
	detail = strings.Join(column(m.View(), g.detail.x, g.detail.w), "\n")
	if !strings.Contains(detail, "SPEC") {
		t.Error("a plan with a spec should show its SPEC line")
	}
}

func TestViewDetailShowsTheSectionOfItsOwnItem(t *testing.T) {
	m := sized(press(newModel(t), "]", "]", "j"), 120, 40)
	v := plain(m.View())
	if !strings.Contains(v, "plans/2026-09-21-alpha#task-2") {
		t.Error("the task row is missing")
	}
	if !strings.Contains(v, "TASK      : Second step") {
		t.Error("the detail should name the kind and the title")
	}
	if strings.Contains(v, "First step") {
		t.Error("the detail of a task should not hold another section")
	}
}

func TestViewShowsProblemsOfTheSelectedItem(t *testing.T) {
	// The third row of the Specs tab is the story with a status nobody knows.
	m := sized(press(newModel(t), "j", "j"), 120, 60)
	if !strings.Contains(plain(m.View()), "! ") {
		t.Error("the problems of the selected item are missing")
	}
	if !strings.Contains(plain(m.View()), "untyped (1)") {
		t.Error("the row that holds the files outside .pm/ is missing")
	}
}

func TestViewStatusLineShowsKeysAndClock(t *testing.T) {
	m := sized(clocked(newModel(t), 20, 46), 100, 30)
	last := lastLine(m.View())
	if got := lipgloss.Width(last); got != 100 {
		t.Errorf("the status line is %d cells wide, want 100", got)
	}
	if !strings.HasPrefix(plain(last), "1 2 3") {
		t.Errorf("the key hints should sit on the left, got %q", plain(last))
	}
	if !strings.Contains(last, "pmb · basic · live · 20:46") {
		t.Errorf("the clock or the mode is missing: %q", plain(last))
	}
	if !strings.HasSuffix(plain(last), "20:46") {
		t.Error("the clock should sit on the right")
	}
	m.manual = true
	if !strings.Contains(m.View(), "pmb · basic · paused · 20:46") {
		t.Error("watching off should read paused")
	}
	m.status = "pm: x status done ✓ committed"
	m = sized(m, 100, 30)
	if !strings.Contains(m.View(), "✓ committed") {
		t.Error("the last message should take the place of the hints")
	}
	m.searching, m.query = true, "alpha"
	if !strings.Contains(plain(m.View()), "/alpha") {
		t.Error("the search box should take the place of the hints")
	}
	// A narrow line gives up the hints, never the clock or the mode.
	for _, w := range []int{59, 40, 30} {
		last := lastLine(sized(clocked(newModel(t), 20, 46), w, 20).View())
		if got := lipgloss.Width(last); got != w {
			t.Errorf("at %d columns the status line is %d cells wide", w, got)
		}
		if !strings.HasSuffix(plain(last), "pmb · basic · live · 20:46") {
			t.Errorf("at %d columns the clock is missing: %q", w, plain(last))
		}
	}
}

func TestViewHelpPopupCoversThePanes(t *testing.T) {
	m := sized(press(clocked(newModel(t), 20, 46), "?"), 100, 30)
	v := m.View()
	if !strings.Contains(v, "Keys") || !strings.Contains(v, "new bug") {
		t.Error("the help popup is missing")
	}
	if !strings.Contains(v, "[1]─Specs") || !strings.Contains(v, "[3]─Detail") {
		t.Error("the popup should cover the panes, not replace them")
	}
	if !strings.HasSuffix(plain(lastLine(v)), "20:46") {
		t.Error("the popup should not hide the status line")
	}
	if strings.Contains(press(m, "?").View(), "Keys") {
		t.Error("? should close the help popup")
	}
}

func TestViewValuePopupAndSlug(t *testing.T) {
	m := sized(press(newModel(t), "]", "]", "]", "s"), 100, 30)
	if !strings.Contains(m.View(), "wontfix") {
		t.Error("the value popup options are not shown")
	}
	m = sized(press(newModel(t), "n", "a", "b"), 100, 30)
	if !strings.Contains(m.View(), "new bug slug: ab") {
		t.Error("the slug prompt is not shown")
	}
}

func TestViewFocusedPaneWearsTheAccent(t *testing.T) {
	m := newModel(t)
	if m.edge(paneOpen).GetForeground() != accentColor {
		t.Error("the pane with the focus should draw its border in the accent color")
	}
	if m.edge(paneDone).GetForeground() == accentColor {
		t.Error("a pane without the focus should not wear the accent color")
	}
	m = press(m, "2")
	if m.edge(paneDone).GetForeground() != accentColor {
		t.Error("the focus moved, so the accent color should move with it")
	}
	if m.edge(paneOpen).GetForeground() == accentColor {
		t.Error("the pane that lost the focus should be dim")
	}
}

func TestViewFinishedTabsFollowTheOpenTab(t *testing.T) {
	for _, tc := range []struct {
		keys []string
		want string
	}{
		{nil, "[2]─Done ─ Dropped"},
		{[]string{"]", "]", "2"}, "[2]─Done──"},
		{[]string{"]", "]", "]", "2"}, "[2]─Fixed ─ Wontfix"},
	} {
		v := sized(press(newModel(t), tc.keys...), 120, 40).View()
		if !strings.Contains(v, tc.want) {
			t.Errorf("after %v pane [2] is missing %q", tc.keys, tc.want)
		}
	}
}

func TestViewEmptyRepo(t *testing.T) {
	b, err := board.Load(config.Default(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	m := sized(New(config.Default(t.TempDir()), b, true), 100, 30)
	if !strings.Contains(m.View(), "no .pm/ yet") {
		t.Error("empty repo message missing")
	}
}

func lastLine(v string) string {
	lines := strings.Split(v, "\n")
	return lines[len(lines)-1]
}
