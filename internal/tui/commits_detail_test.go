package tui

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/commits"
	"github.com/iyay/acta/internal/config"
)

// commitBoardFiles is a plan with a hash and four tasks, a bug and a debt
// file. The commits of each task are made up by the test.
func commitBoardFiles() map[string]string {
	return map[string]string{
		".acta/specs/2026-09-20-a.md": "---\nid: SPEC-1\n---\n# Spec A\n",
		".acta/plans/2026-09-21-a.md": "---\nid: PLN-0001\nhash: abcdefg\nparent: specs/2026-09-20-a\n---\n# Plan A\n\n### Task 1: First\n- [x] a\n\n### Task 2: Second\n- [x] b\n\n### Task 3: Third\n- [ ] c\n\n### Task 4: Fourth\n- [ ] d\n",
		".acta/bugs/2026-09-23-b.md":  "---\nid: BUG-1\n---\n# Bug B\n",
		".acta/debt/2026-09-24-d.md":  "---\nid: DEBT-1\n---\n# Review NOTEs\n\n- [ ] first note\n",
	}
}

// fakeCommits makes n commits, oldest first, one hour apart. Their shas and
// subjects carry the prefix, so a test can tell where a row came from.
func fakeCommits(prefix string, n int) []commits.Commit {
	base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	out := make([]commits.Commit, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, commits.Commit{
			Sha:     fmt.Sprintf("%s%02d%s", prefix, i, strings.Repeat("0", 40)),
			Date:    base.Add(time.Duration(i) * time.Hour),
			Subject: fmt.Sprintf("feat: %s change %d", prefix, i),
		})
	}
	return out
}

func choreCommit(sha string, at time.Time) commits.Commit {
	return commits.Commit{Sha: sha + strings.Repeat("0", 40), Date: at, Subject: "chore(plan): tick", Chore: true}
}

// swapFindCommits points the load at a fake and puts the real one back.
func swapFindCommits(t *testing.T, fn func(config.Config) (map[string][]commits.Commit, error)) {
	t.Helper()
	old := findCommits
	findCommits = fn
	t.Cleanup(func() { findCommits = old })
}

// loadedModel builds the board and feeds the model a reloadMsg that carries
// the given commits, the way the real load does.
func loadedModel(t *testing.T, found map[string][]commits.Commit) Model {
	t.Helper()
	cfg := treeCfg(t, commitBoardFiles())
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	m = sized(m, 160, 60)
	next, _ := m.Update(reloadMsg{b: b, commits: found})
	m = next.(Model)
	m.openPlans = everyPlanOpen(b)
	return m
}

// detailText is the plain detail of one item, found on whichever tab lists it.
func detailText(t *testing.T, m Model, id string) string {
	t.Helper()
	it := m.board.Get(id)
	if it == nil {
		t.Fatalf("the board holds no %s", id)
	}
	id = it.ID
	for i := range topTabs {
		m.openTab(i)
		for _, p := range m.panes() {
			for d := range max(1, len(topTabs[i].done)) {
				m.done = d
				m.focus = p
				rows, sel, idx := m.slotOf(p)
				for _, r := range rows {
					if r.id == id {
						*sel, *idx = id, 0
						return strings.Join(plainLines(m.detailLines(100)), "\n")
					}
				}
			}
		}
	}
	t.Fatalf("%s is on no list of the board", id)
	return ""
}

func TestCommitsDetailTaskShowsItsCommits(t *testing.T) {
	m := loadedModel(t, map[string][]commits.Commit{"abcdefg#1": fakeCommits("a", 2)})
	got := detailText(t, m, "PLN-0001.01")
	if !strings.Contains(got, "COMMITS") {
		t.Fatalf("no COMMITS section:\n%s", got)
	}
	// Newest first: the second commit is on top.
	at2 := strings.Index(got, "a020000  feat: a change 2")
	at1 := strings.Index(got, "a010000  feat: a change 1")
	if at2 < 0 || at1 < 0 || at2 > at1 {
		t.Errorf("want 7-char sha rows, newest first:\n%s", got)
	}
	if strings.Contains(got, "no linked commits") || strings.Contains(got, "more") {
		t.Errorf("two commits need no footer:\n%s", got)
	}
}

func TestCommitsDetailTaskWithMoreThanFiveShowsFiveAndCount(t *testing.T) {
	m := loadedModel(t, map[string][]commits.Commit{"abcdefg#2": fakeCommits("b", 7)})
	got := detailText(t, m, "PLN-0001.02")
	rows := strings.Count(got, "feat: b change")
	if rows != 5 {
		t.Errorf("rows = %d, want 5:\n%s", rows, got)
	}
	if !strings.Contains(got, "feat: b change 7") || strings.Contains(got, "feat: b change 2\n") {
		t.Errorf("want the newest five (7 down to 3):\n%s", got)
	}
	if !strings.Contains(got, "+2 more · d to open") {
		t.Errorf("no overflow line:\n%s", got)
	}
}

func TestCommitsDetailPlanShowsEveryTasksCommits(t *testing.T) {
	m := loadedModel(t, map[string][]commits.Commit{
		"abcdefg#1": fakeCommits("a", 2),
		"abcdefg#2": fakeCommits("b", 2),
	})
	got := detailText(t, m, "PLN-0001")
	if !strings.Contains(got, "COMMITS") || strings.Count(got, "feat: ") != 4 {
		t.Errorf("want one row per commit of the plan:\n%s", got)
	}
}

func TestCommitsDetailPlanListsAShaOnceAndNewestFirst(t *testing.T) {
	both := fakeCommits("s", 1)
	// One commit names two tasks, so it sits under two keys.
	m := loadedModel(t, map[string][]commits.Commit{"abcdefg#1": both, "abcdefg#2": both})
	got := detailText(t, m, "PLN-0001")
	if n := strings.Count(got, "feat: s change 1"); n != 1 {
		t.Errorf("a shared commit shows %d times, want 1:\n%s", n, got)
	}
}

func TestCommitsDetailChoreCommitsAreNeverCounted(t *testing.T) {
	at := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	only := []commits.Commit{choreCommit("c1", at), choreCommit("c2", at)}
	m := loadedModel(t, map[string][]commits.Commit{"abcdefg#3": only})
	got := detailText(t, m, "PLN-0001.03")
	if !strings.Contains(got, "no linked commits") || strings.Contains(got, "chore(plan)") {
		t.Errorf("a task with only chore commits has none:\n%s", got)
	}
	// Chore commits do not push a task over the five-row limit either.
	mixed := append(fakeCommits("m", 5), choreCommit("c3", at))
	m = loadedModel(t, map[string][]commits.Commit{"abcdefg#3": mixed})
	got = detailText(t, m, "PLN-0001.03")
	if strings.Contains(got, "more") || strings.Contains(got, "chore(plan)") {
		t.Errorf("a chore commit was counted:\n%s", got)
	}
}

func TestCommitsDetailNoCommitsSaysSo(t *testing.T) {
	m := loadedModel(t, nil)
	for _, id := range []string{"PLN-0001.04", "PLN-0001"} {
		got := detailText(t, m, id)
		if !strings.Contains(got, "COMMITS") || !strings.Contains(got, "no linked commits") {
			t.Errorf("%s: want COMMITS and no linked commits:\n%s", id, got)
		}
	}
}

func TestCommitsDetailOtherKindsShowNoSection(t *testing.T) {
	// Every key is full, so any section on another kind would show rows.
	all := map[string][]commits.Commit{}
	for n := 1; n <= 4; n++ {
		all[fmt.Sprintf("abcdefg#%d", n)] = fakeCommits("x", 3)
	}
	m := loadedModel(t, all)
	for _, id := range []string{"SPEC-1", "BUG-1", "DEBT-1.1"} {
		if got := detailText(t, m, id); strings.Contains(got, "COMMITS") || strings.Contains(got, "no linked commits") {
			t.Errorf("%s shows a COMMITS section:\n%s", id, got)
		}
	}
}

func TestCommitsDetailRowsAreCutToThePaneWidth(t *testing.T) {
	long := fakeCommits("w", 1)
	long[0].Subject = strings.Repeat("long words ", 30)
	m := loadedModel(t, map[string][]commits.Commit{"abcdefg#1": long})
	lines := m.commitLines(m.board.Get("PLN-0001.01"), 40)
	if len(lines) != 2 {
		t.Fatalf("lines = %q, want heading and one row", lines)
	}
	if w := len([]rune(plain(lines[1]))); w > 40 {
		t.Errorf("row is %d wide, want at most 40: %q", w, lines[1])
	}
}

func TestCommitsDetailGitFailureKeepsTheBoard(t *testing.T) {
	swapFindCommits(t, func(config.Config) (map[string][]commits.Commit, error) {
		return map[string][]commits.Commit{"abcdefg#1": fakeCommits("p", 2)}, errors.New("git is gone")
	})
	cfg := treeCfg(t, commitBoardFiles())
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	m = sized(m, 160, 60)
	msg, ok := m.reloadCmd()().(reloadMsg)
	if !ok || msg.err != nil || msg.b == nil {
		t.Fatalf("reload = %+v, want the board and no error", msg)
	}
	if len(msg.commits) != 0 {
		t.Errorf("commits = %v, want an empty map after a git failure", msg.commits)
	}
	next, _ := m.Update(msg)
	m = next.(Model)
	m.openPlans = everyPlanOpen(m.board)
	if got := detailText(t, m, "PLN-0001.01"); !strings.Contains(got, "no linked commits") {
		t.Errorf("want no linked commits after a failed load:\n%s", got)
	}
}

func TestCommitsDetailLoadUsesTheSeam(t *testing.T) {
	swapFindCommits(t, func(config.Config) (map[string][]commits.Commit, error) {
		return map[string][]commits.Commit{"abcdefg#1": fakeCommits("q", 1)}, nil
	})
	cfg := treeCfg(t, commitBoardFiles())
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := sized(New(cfg, b, true), 160, 60)
	m.render = func(md string, _ int) string { return md }
	msg := m.reloadCmd()().(reloadMsg)
	next, _ := m.Update(msg)
	m = next.(Model)
	m.openPlans = everyPlanOpen(m.board)
	if got := detailText(t, m, "PLN-0001.01"); !strings.Contains(got, "feat: q change 1") {
		t.Errorf("the load did not reach the detail:\n%s", got)
	}
}

// firstCol is how many spaces a line starts with, ignoring colors.
func firstCol(ln string) int {
	p := plain(ln)
	return len(p) - len(strings.TrimLeft(p, " "))
}

// The header, the commits and the date line start 2 columns in, where the
// markdown body starts, and the rules stay at column 0. No line is wider than
// the box.
func TestDetailTextStartsAtTheBodyMargin(t *testing.T) {
	m := loadedModel(t, map[string][]commits.Commit{"abcdefg#1": fakeCommits("a", 2)})
	m.render = func(md string, _ int) string { return "  " + md }
	it := m.board.Get("PLN-0001")
	m = onItem(t, m, it.ID)
	for _, w := range []int{20, 40, 100} {
		head, mid, foot := m.buildDetailParts(w)
		for i, ln := range head[:len(head)-1] {
			if c := firstCol(ln); c != 2 {
				t.Errorf("w=%d header line %d starts at %d: %q", w, i, c, plain(ln))
			}
		}
		if c := firstCol(head[len(head)-1]); c != 0 {
			t.Errorf("w=%d rule starts at %d", w, c)
		}
		var sawCommits, sawBody bool
		for _, ln := range mid {
			p := strings.TrimSpace(plain(ln))
			switch {
			case p == "COMMITS" || strings.HasPrefix(p, "a020000"):
				sawCommits = true
				if c := firstCol(ln); c != 2 {
					t.Errorf("w=%d %q starts at %d", w, p, c)
				}
			case strings.HasPrefix(p, "# Plan A"):
				sawBody = true
				if c := firstCol(ln); c != 2 {
					t.Errorf("w=%d body starts at %d", w, c)
				}
			}
		}
		if !sawCommits || !sawBody {
			t.Errorf("w=%d commits=%v body=%v", w, sawCommits, sawBody)
		}
		if c := firstCol(foot); c != 2 {
			t.Errorf("w=%d date line starts at %d: %q", w, c, plain(foot))
		}
		for _, ln := range append(append(head, mid...), foot) {
			if lw := lipgloss.Width(ln); lw > w {
				t.Errorf("w=%d line is %d wide: %q", w, lw, plain(ln))
			}
		}
	}
}

// The detail draws its commits with the same row painter as the Commits
// screen, so the two always look alike.
func TestCommitsDetailRowsMatchTheCommitsScreen(t *testing.T) {
	landed := fakeCommits("l", 1)
	landed[0].Landed = true
	both := fakeCommits("m", 1)
	both[0].Landed = true
	found := map[string][]commits.Commit{
		"abcdefg#1": append(append([]commits.Commit{}, landed...), both...),
		"abcdefg#2": append(fakeCommits("u", 1), both...),
		"abcdefg#3": {choreCommit("c1", time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC))},
	}
	m := loadedModel(t, found)
	const w = 60

	plan := m.board.Get("PLN-0001")
	var want []string
	for _, r := range commitRowsOf(plan, m.taskCommits) {
		if !r.c.Chore {
			want = append(want, m.commitRowLine(r, true, w))
		}
	}
	got := m.commitLines(plan, w)
	if len(want) != 3 || len(got) != 1+len(want) {
		t.Fatalf("got %d lines for %d rows: %q", len(got), len(want), got)
	}
	for i, ln := range want {
		if got[i+1] != ln {
			t.Errorf("plan row %d = %q, want %q", i, got[i+1], ln)
		}
	}
	if !strings.Contains(plain(strings.Join(got, "\n")), "#1,#2") {
		t.Errorf("the two-task commit lost its labels: %q", got)
	}
	if strings.Contains(plain(strings.Join(got, "\n")), "chore(plan)") {
		t.Errorf("a chore row shows: %q", got)
	}

	task := m.board.Get("PLN-0001.01")
	rows := commitRowsOf(task, m.taskCommits)
	tl := m.commitLines(task, w)
	if len(tl) != 1+len(rows) {
		t.Fatalf("task lines = %q", tl)
	}
	for i, r := range rows {
		if tl[i+1] != m.commitRowLine(r, false, w) {
			t.Errorf("task row %d = %q, want %q", i, tl[i+1], m.commitRowLine(r, false, w))
		}
	}
}

func TestCommitsDetailMoreCountSkipsChores(t *testing.T) {
	list := fakeCommits("b", 7)
	list = append(list, choreCommit("c1", time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)))
	m := loadedModel(t, map[string][]commits.Commit{"abcdefg#2": list})
	got := m.commitLines(m.board.Get("PLN-0001.02"), 60)
	if len(got) != 1+commitRows+1 || plain(got[len(got)-1]) != "+2 more · d to open" {
		t.Errorf("lines = %q", got)
	}
}
