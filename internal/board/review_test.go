package board

import (
	"testing"
)

// A plan with every box ticked still waits for its branch to merge, so the
// board reads it as in review and never as done.
func TestAllTickedPlanInWorktreeReadsReview(t *testing.T) {
	t.Parallel()

	main := tree(t, map[string]string{
		".acta/specs/2026-09-20-a.md": specA,
		".acta/plans/2026-09-21-a.md": planBehind,
	})
	wt := tree(t, map[string]string{
		".acta/specs/2026-09-20-a.md": specA,
		".acta/plans/2026-09-21-a.md": planAhead,
	})
	b, err := LoadTrees(main, []Tree{{Cfg: wt, Branch: "feat"}})
	if err != nil {
		t.Fatal(err)
	}
	plan := b.Get("plans/2026-09-21-a")
	if plan == nil || plan.Status != "review" || plan.Worktree != "feat" {
		t.Fatalf("plan = %+v, want review from feat", plan)
	}
	spec := b.Get("specs/2026-09-20-a")
	if spec == nil || spec.Status != "in-progress" {
		t.Fatalf("spec = %+v, want in-progress while its only plan waits", spec)
	}
	if Closed("review") {
		t.Error("review must stay off the closed list")
	}
}

// The same boxes merged to the main tree mean the work is done.
func TestAllTickedPlanInMainTreeStaysDone(t *testing.T) {
	t.Parallel()

	b := boardWith(t, map[string]string{
		"specs/2026-09-20-a.md": specA,
		"plans/2026-09-21-a.md": planAhead,
	})
	plan := b.Get("plans/2026-09-21-a")
	if plan == nil || plan.Status != "done" || plan.Worktree != "" {
		t.Fatalf("plan = %+v, want done from the main tree", plan)
	}
	spec := b.Get("specs/2026-09-20-a")
	if spec == nil || spec.Status != "done" {
		t.Fatalf("spec = %+v, want done once its plan is merged", spec)
	}
}

// A box still open on the branch means there is work left, not work waiting.
func TestHalfTickedPlanInWorktreeStaysInProgress(t *testing.T) {
	t.Parallel()

	main := tree(t, map[string]string{
		".acta/specs/2026-09-20-a.md": specA,
	})
	wt := tree(t, map[string]string{
		".acta/specs/2026-09-20-a.md": specA,
		".acta/plans/2026-09-21-a.md": planBehind,
	})
	b, err := LoadTrees(main, []Tree{{Cfg: wt, Branch: "feat"}})
	if err != nil {
		t.Fatal(err)
	}
	plan := b.Get("plans/2026-09-21-a")
	if plan == nil || plan.Status != "in-progress" || plan.Worktree != "feat" {
		t.Fatalf("plan = %+v, want in-progress from feat with a box still open", plan)
	}
}

// A plan with no boxes keeps the word it has today.
func TestPlanWithNoTasksKeepsTodayStatus(t *testing.T) {
	t.Parallel()

	b := boardWith(t, map[string]string{
		"plans/2026-09-23-lonely.md": "# Lonely plan\n\n### Task 1: Only step\n\nJust do it.\n",
	})
	plan := b.Get("plans/2026-09-23-lonely")
	if plan == nil || plan.Status != "approved" {
		t.Fatalf("plan = %+v, want approved the way a plan with no boxes reads today", plan)
	}
}

// A bug whose only plan waits for merge is still being fixed, never fixed.
func TestBugWithOnlyPlanInReviewReadsFixing(t *testing.T) {
	t.Parallel()

	main := tree(t, map[string]string{
		".acta/bugs/2026-09-24-crash.md": "# Crash\n",
	})
	wt := tree(t, map[string]string{
		".acta/bugs/2026-09-24-crash.md":      "# Crash\n",
		".acta/plans/2026-09-25-crash-fix.md": "---\nparent: bugs/2026-09-24-crash\n---\n# Fix\n\n### Task F1: One\n- [x] a\n",
	})
	b, err := LoadTrees(main, []Tree{{Cfg: wt, Branch: "feat"}})
	if err != nil {
		t.Fatal(err)
	}
	plan := b.Get("plans/2026-09-25-crash-fix")
	if plan == nil || plan.Status != "review" {
		t.Fatalf("plan = %+v, want review from feat", plan)
	}
	bug := b.Get("bugs/2026-09-24-crash")
	if bug == nil || bug.Status != "fixing" {
		t.Fatalf("bug = %+v, want fixing while its only plan waits", bug)
	}
}
