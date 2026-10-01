package board

import "testing"

// A bug's fixed_in names the commit that fixed it, and that word is written
// only after the merge lands. So a bug that has one is fixed, whatever its
// child plan's boxes say. A status written by hand still wins over it.
func TestBugFixedInMakesItFixed(t *testing.T) {
	t.Parallel()

	const id = "bugs/2026-10-01-crash"

	// The plan hangs under the bug, so the bug counts it as a child of its own.
	plan := func(ticks string) map[string]string {
		return map[string]string{
			"plans/2026-10-01-crash-fix.md": "---\nparent: " + id + "\n---\n# Fix the crash\n\n### Task 1: Write the test\n\n" + ticks + "\n",
		}
	}
	withBug := func(front, ticks string) map[string]string {
		files := map[string]string{id + ".md": "---\n" + front + "---\n# App crashes on start\n"}
		if ticks != "" {
			for k, v := range plan(ticks) {
				files[k] = v
			}
		}
		return files
	}

	cases := []struct {
		name   string
		files  map[string]string
		status string
		source string
	}{
		{
			name:   "fixed_in and no plan",
			files:  withBug("fixed_in: a1b2c3d\n", ""),
			status: "fixed",
			source: "derived",
		},
		{
			name:   "fixed_in wins over a plan with an open task",
			files:  withBug("fixed_in: a1b2c3d\n", "- [ ] write the test"),
			status: "fixed",
			source: "derived",
		},
		{
			name:   "fixed_in with a finished plan",
			files:  withBug("fixed_in: a1b2c3d\n", "- [x] write the test"),
			status: "fixed",
			source: "derived",
		},
		{
			name:   "a written status wins over fixed_in",
			files:  withBug("fixed_in: a1b2c3d\nstatus: wontfix\n", ""),
			status: "wontfix",
			source: "frontmatter",
		},
		{
			name:   "no fixed_in and no plan",
			files:  withBug("", ""),
			status: "open",
			source: "derived",
		},
		{
			name:   "no fixed_in with a finished plan",
			files:  withBug("", "- [x] write the test"),
			status: "fixed",
			source: "derived",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			it := boardWith(t, c.files).Get(id)
			if it == nil {
				t.Fatalf("%s missing from the board", id)
			}
			if it.Status != c.status || it.StatusSource != c.source {
				t.Errorf("status %s (%s), want %s (%s)", it.Status, it.StatusSource, c.status, c.source)
			}
		})
	}
}
