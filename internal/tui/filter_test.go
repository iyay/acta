package tui

import (
	"slices"
	"strings"
	"testing"
)

func TestPriorityFilterStepsThroughEveryLevel(t *testing.T) {
	t.Parallel()

	m := press(prioModel(t), tabKey(tabBugs))
	for _, want := range [][]string{
		{"bugs/2026-10-01-c", "bugs/2026-10-01-d"},
		{"bugs/2026-10-01-e"},
		{"bugs/2026-10-01-b"},
		{"bugs/2026-10-01-a"},
		{"bugs/2026-10-01-c", "bugs/2026-10-01-d", "bugs/2026-10-01-e", "bugs/2026-10-01-b", "bugs/2026-10-01-a"},
	} {
		m = press(m, "f")
		if got := rowIDs(m); !slices.Equal(got, want) {
			t.Fatalf("after f the rows are %v, want %v", got, want)
		}
	}
}

func TestPriorityAndStatusFiltersStack(t *testing.T) {
	t.Parallel()

	m := press(prioModel(t), tabKey(tabBugs), "f", "F")
	if got := rowIDs(m); !slices.Equal(got, []string{"bugs/2026-10-01-c", "bugs/2026-10-01-d"}) {
		t.Fatalf("high and open gave %v", got)
	}
	// No high bug is being fixed, so nothing is left. Moving on the empty
	// list must not pick an item or crash.
	m = press(m, "F", "j", "k", "G")
	if got := rowIDs(m); len(got) != 0 {
		t.Fatalf("high and fixing gave %v, want none", got)
	}
	if it := m.Selected(); it != nil {
		t.Fatalf("empty list selected %s", it.ID)
	}
}

func TestStatusFilterOnDebtChangesNothing(t *testing.T) {
	t.Parallel()

	m := press(prioModel(t), tabKey(tabDebt))
	before := rowIDs(m)
	m = press(m, "F")
	if got := rowIDs(m); !slices.Equal(got, before) {
		t.Fatalf("F on Debt changed the rows to %v", got)
	}
	if m.status == "" {
		t.Fatal("F on Debt set no status text")
	}
}

func TestStatusFilterKeepsThePlanOfAMatchingTask(t *testing.T) {
	t.Parallel()

	m := toPlans(planModel(t))
	// f on a tab with no priority does the same as F.
	m = press(m, "f")
	want := []string{"plans/2026-09-23-lonely", "plans/2026-09-23-lonely#task-1"}
	if got := rowIDs(m); !slices.Equal(got, want) {
		t.Fatalf("todo filter gave %v, want %v", got, want)
	}
}

func TestEachTabKeepsItsOwnFilter(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabSpecs), "F")
	specs := rowIDs(m)
	m = press(m, tabKey(tabScratchpad))
	if got := rowIDs(m); len(got) != 2 {
		t.Fatalf("Scratchpad took the Specs filter: %v", got)
	}
	m = press(m, tabKey(tabSpecs))
	if got := rowIDs(m); !slices.Equal(got, specs) {
		t.Fatalf("Specs lost its filter: %v, want %v", got, specs)
	}
}

func TestListTitleNamesTheActiveFilters(t *testing.T) {
	t.Parallel()

	m := press(prioModel(t), tabKey(tabBugs))
	for _, tc := range []struct {
		keys []string
		want string
	}{
		{nil, "Open"},
		{[]string{"f"}, "Open [H]"},
		{[]string{"F"}, "Open [H · open]"},
		{[]string{"f", "f", "f"}, "Open [none · open]"},
		{[]string{"f", "F", "F"}, "Open"},
	} {
		m = press(m, tc.keys...)
		if got := m.tabsOf(paneList); !slices.Equal(got, []string{tc.want}) {
			t.Fatalf("after %v the title is %v, want %q", tc.keys, got, tc.want)
		}
	}
}

func TestEmptyFilteredListSaysNoItemsMatch(t *testing.T) {
	t.Parallel()

	m := sized(press(prioModel(t), tabKey(tabBugs), "f", "F", "F"), 120, 40)
	if v := m.View(); !strings.Contains(v, "no items match") {
		t.Fatalf("empty filtered list view lacks the no items match text:\n%s", v)
	}
}

func TestFilterKeysInHintsAndHelp(t *testing.T) {
	t.Parallel()

	bugs := strings.Join(press(prioModel(t), tabKey(tabBugs)).hints(), " ")
	if !strings.Contains(bugs, "Priority filter: f") || !strings.Contains(bugs, "Status filter: F") {
		t.Fatalf("Bugs hints %q lack the filter keys", bugs)
	}
	specs := strings.Join(press(newModel(t), tabKey(tabSpecs)).hints(), " ")
	if strings.Contains(specs, "Priority filter") || !strings.Contains(specs, "Status filter: F") {
		t.Fatalf("Specs hints %q", specs)
	}
	debt := strings.Join(press(prioModel(t), tabKey(tabDebt)).hints(), " ")
	if strings.Contains(debt, "Status filter") || !strings.Contains(debt, "Priority filter: f") {
		t.Fatalf("Debt hints %q", debt)
	}
	if !strings.Contains(helpLines, "\nf F ") || !strings.Contains(helpLines, "by priority / by status") {
		t.Fatal("the ? help names neither f nor F")
	}
}
