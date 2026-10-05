// Failing test first: a plan in review must wear its own color on its row
// and name its branch in the detail. Review has no color today, so it fails.
package tui

import (
	"strings"
	"testing"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/theme"
)

func reviewBoard(t *testing.T) Model {
	t.Helper()
	ahead := "---\nid: PLN-0001\nparent: specs/2026-09-20-a\n---\n# Plan A\n\n### Task 1: First\n- [x] a\n- [x] b\n"
	behind := "---\nid: PLN-0001\nparent: specs/2026-09-20-a\n---\n# Plan A\n\n### Task 1: First\n- [x] a\n- [ ] b\n"
	main := treeCfg(t, map[string]string{
		".acta/specs/2026-09-20-a.md": "# Spec A\n",
		".acta/plans/2026-09-21-a.md": behind,
	})
	wt := treeCfg(t, map[string]string{
		".acta/specs/2026-09-20-a.md": "# Spec A\n",
		".acta/plans/2026-09-21-a.md": ahead,
	})
	b, err := board.LoadTrees(main, []board.Tree{{Cfg: wt, Branch: "feat"}})
	if err != nil {
		t.Fatal(err)
	}
	if p := b.Get("plans/2026-09-21-a"); p == nil || p.Status != "review" {
		t.Fatalf("plan = %+v, want review from feat", p)
	}
	m := New(main, b, true)
	m.render = func(md string, _ int) string { return md }
	return sized(m, 160, 40)
}
func TestReviewRowWearsItsOwnColor(t *testing.T) {
	t.Parallel()

	withTrueColor(func() {
		m := reviewBoard(t).WithTheme("tokyo-night", true)
		it := m.board.Get("plans/2026-09-21-a")
		if it == nil || it.Status != "review" {
			t.Fatalf("plan = %+v, want review from feat", it)
		}
		ln := workLine(m.styles, it, false, 80)
		review := sgr.FindString(m.styles.review.Render("x"))
		green := sgr.FindString(m.styles.done.Render("x"))
		// The id keeps its kind color; the title after it wears review.
		id := m.styles.kind(board.KindPlan).Render("PLN-0001")
		tail := ln[strings.Index(ln, id)+len(id):]
		if !strings.Contains(tail, review) {
			t.Errorf("review row title wears no review color: %q", ln)
		}
		if strings.Contains(tail, green) {
			t.Errorf("review row title wears done green: %q", ln)
		}
		if !strings.HasPrefix(ln, m.styles.dot(dotReview).Render(dotReview)) {
			t.Errorf("review row wears no review dot: %q", ln)
		}
	})
}

func TestReviewDetailNamesBranch(t *testing.T) {
	// No t.Parallel: this test and the round test swap the same hook.

	old := rounds
	rounds = func(config.Config) []board.RunState { return nil }
	defer func() { rounds = old }()

	m := reviewBoard(t)
	m.openPlans = everyPlanOpen(m.board)
	m.openTab(tabPlans)
	rows, sel, idx := m.slotOf(paneList)
	for _, r := range rows {
		if r.id == "plans/2026-09-21-a" {
			*sel, *idx = r.id, 0
			lines := plainLines(m.detailLines(100))
			joined := strings.Join(lines, "\n")
			if !strings.Contains(joined, "feat") {
				t.Errorf("review detail names no branch:\n%s", joined)
			}
			if !strings.Contains(joined, "ROUND") {
				t.Errorf("review detail names no round line:\n%s", joined)
			}
			if !strings.Contains(joined, "(none)") {
				t.Errorf("a review plan with no round should say none:\n%s", joined)
			}
			return
		}
	}
	t.Fatal("no row for the review plan, so this test proves nothing")
}

func TestReviewColorDistinctInEveryTheme(t *testing.T) {
	t.Parallel()

	for _, name := range theme.Names() {
		th, _ := theme.Builtin(name)
		for _, dark := range []bool{true, false} {
			s := newStyles(th, dark)
			if s.review.GetForeground() == s.done.GetForeground() {
				t.Errorf("%s dark=%v: review takes the done color", name, dark)
			}
			if s.review.GetForeground() == s.waiting.GetForeground() {
				t.Errorf("%s dark=%v: review takes the waiting color", name, dark)
			}
		}
	}
}

func TestReviewDetailNamesRound(t *testing.T) {
	// No t.Parallel: this test and the branch test swap the same hook.

	old := rounds
	rounds = func(config.Config) []board.RunState {
		return []board.RunState{{Plan: &board.Item{ID: "plans/2026-09-21-a"}, Round: "3"}}
	}
	defer func() { rounds = old }()
	m := reviewBoard(t)
	m.openPlans = everyPlanOpen(m.board)
	m.openTab(tabPlans)
	rows, sel, idx := m.slotOf(paneList)
	for _, r := range rows {
		if r.id == "plans/2026-09-21-a" {
			*sel, *idx = r.id, 0
			lines := plainLines(m.detailLines(100))
			joined := strings.Join(lines, "\n")
			if !strings.Contains(joined, "ROUND") || !strings.Contains(joined, "3") {
				t.Errorf("review detail names no round 3:\n%s", joined)
			}
			return
		}
	}
	t.Fatal("no row for the review plan, so this test proves nothing")
}
