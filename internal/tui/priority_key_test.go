package tui

import (
	"slices"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/write"
)

const prioOptions = "high,medium,low,none"

func TestPriorityPopupOnABugStartsOnItsValue(t *testing.T) {
	t.Parallel()

	// The first Bugs row is BUG-0003, set to high.
	m := press(prioModel(t), tabKey(tabBugs), "p")
	if m.popup == nil || m.popup.field != "priority" || strings.Join(m.popup.options, ",") != prioOptions || m.popup.idx != 0 {
		t.Fatalf("popup on a high bug = %+v", m.popup)
	}
	// The last Bugs row is BUG-0001, with no priority, so the cursor sits on none.
	m = press(prioModel(t), tabKey(tabBugs), "G", "p")
	if m.popup == nil || m.popup.idx != 3 {
		t.Fatalf("popup on an unset bug = %+v", m.popup)
	}
}

func TestPriorityPopupOnADebtLine(t *testing.T) {
	t.Parallel()

	m := press(prioModel(t), tabKey(tabDebt), "p")
	if it := m.Selected(); it == nil || it.Kind != board.KindDebtItem {
		t.Fatalf("selected %+v, want a debt item", it)
	}
	if m.popup == nil || m.popup.field != "priority" || strings.Join(m.popup.options, ",") != prioOptions {
		t.Fatalf("popup on a debt line = %+v", m.popup)
	}
	// The popup offers one level more than the sort order, so the list the
	// sort reads must not grow a none in the middle of it.
	if got := strings.Join(board.Priorities, ","); got != "high,medium,low" {
		t.Fatalf("the popup changed the sort order list: %q", got)
	}
}

func TestPriorityPopupWritesTheChoice(t *testing.T) {
	t.Parallel()

	m := press(prioModel(t), tabKey(tabBugs), "G")
	var got []string
	m.setValue = func(id, field, value string) (write.Outcome, error) {
		got = []string{id, field, value}
		return write.Outcome{Committed: true}, nil
	}
	m = press(m, "p", "k", "k", "k", "enter")
	if strings.Join(got, " ") != "bugs/2026-10-01-a priority high" {
		t.Fatalf("setValue got %v", got)
	}
	if m.popup != nil || !strings.Contains(m.status, "committed") {
		t.Fatalf("popup %v status %q", m.popup, m.status)
	}
}

func TestPriorityKeyRefusesOtherKinds(t *testing.T) {
	t.Parallel()

	const want = "p sets the priority of a bug or a debt line"
	for _, c := range []struct {
		name string
		keys []string
		kind board.Kind
	}{
		{"scratch", []string{tabKey(tabScratchpad)}, board.KindScratch},
		{"spec", []string{tabKey(tabSpecs)}, board.KindStory},
		{"plan", []string{tabKey(tabPlans)}, board.KindPlan},
		{"task", []string{tabKey(tabPlans), " ", "j"}, board.KindTask},
	} {
		m := press(newModel(t), c.keys...)
		if it := m.Selected(); it == nil || it.Kind != c.kind {
			t.Fatalf("%s: selected %+v, want kind %s", c.name, it, c.kind)
		}
		m = press(m, "p")
		if m.popup != nil || m.status != want {
			t.Errorf("%s: popup %+v status %q, want no popup and %q", c.name, m.popup, m.status, want)
		}
		if slices.Contains(m.hints(), "Priority: p") {
			t.Errorf("%s: hints offer Priority: p: %v", c.name, m.hints())
		}
	}
}

func TestPriorityHintOnBugsAndDebtLines(t *testing.T) {
	t.Parallel()

	for _, keys := range [][]string{
		{tabKey(tabBugs)},
		{tabKey(tabBugs), "enter"},
		{tabKey(tabDebt)},
		{tabKey(tabDebt), "enter"},
	} {
		m := press(sized(prioModel(t), 200, 40), keys...)
		if !slices.Contains(m.hints(), "Priority: p") {
			t.Errorf("keys %v: hints %v, want Priority: p", keys, m.hints())
		}
	}
}

func TestHelpListsThePriorityKey(t *testing.T) {
	t.Parallel()

	if !strings.Contains(helpLines, "set the priority of a bug or a debt line: high, medium, low or none") {
		t.Fatalf("help does not list p: %q", helpLines)
	}
}

// The rows that hold no item of their own, or whose file lives somewhere
// else, must refuse p the same way they refuse the other keys.
func TestPriorityKeyRefusesTheRowsWithNoItemOfTheirOwn(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name  string
		start func(*testing.T) Model
		// want is the word the status line must carry, so the refusal comes
		// from the same guard the other keys use.
		want string
	}{
		{"the group row", func(t *testing.T) Model { return press(newModel(t), tabKey(tabSpecs), "G") }, "nothing selected"},
		{"a worktree row", func(t *testing.T) Model { return press(worktreeModel(t), tabKey(tabBugs), "j") }, "worktree feat"},
		{"a legacy bug", legacyBugModel, "legacy file"},
	} {
		m := c.start(t)
		m = press(m, "p")
		if m.popup != nil || !strings.Contains(m.status, c.want) {
			t.Errorf("%s: popup %+v status %q, want no popup and a status with %q", c.name, m.popup, m.status, c.want)
		}
		if slices.Contains(m.hints(), "Priority: p") {
			t.Errorf("%s: hints offer Priority: p: %v", c.name, m.hints())
		}
	}
}

// legacyBugModel stands on a bug that lives in a folder the board calls
// legacy. Such a file is left out of the Bugs tab and sits under the group
// row of the Specs tab, so the cursor walks to it there.
func legacyBugModel(t *testing.T) Model {
	t.Helper()
	cfg := treeCfg(t, map[string]string{
		"docs/superpowers/bugs/2026-10-01-old.md":    "# Old bug\n\n## Symptom\nx\n",
		"docs/superpowers/specs/2026-10-02-older.md": "# Older spec\n\n## Symptom\ny\n",
	})
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	// The group row sits at the end of the Specs list and opens the legacy
	// files under it, so the walk below can reach the bug.
	m = press(m, tabKey(tabSpecs), "G", "enter")
	for m.Selected() == nil || m.Selected().Kind != board.KindBug {
		before := m.Selected()
		m = press(m, "j")
		if m.Selected() == before {
			t.Fatalf("no legacy bug under the group row: %+v", m.Selected())
		}
	}
	return m
}
