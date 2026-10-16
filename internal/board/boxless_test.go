package board

import "testing"

// Old minimal plans can hold tasks with no box at all. Once the plan is
// finished such a task is done, or the plan would look open forever. Before
// that, a task with no box keeps the status it always had.
func TestBoxlessTaskOfFinishedPlan(t *testing.T) {
	t.Parallel()

	const id = "plans/2026-10-01-old"

	plan := func(front, body string) map[string]string {
		return map[string]string{id + ".md": "---\n" + front + "---\n# Old plan\n\n### Task 1: Do it\n\n" + body + "\n"}
	}

	cases := []struct {
		name  string
		files map[string]string
		task  string
		plan  string
	}{
		{"finished plan, no box", plan("started: \"2026-10-01 10:00:00\"\nfinished: \"2026-10-01 11:00:00\"\n", "**verify:** it works"), "done", "done"},
		{"started plan, no box", plan("started: \"2026-10-01 10:00:00\"\n", "**verify:** it works"), "todo", "approved"},
		{"new plan, no box", plan("", "**verify:** it works"), "todo", "approved"},
		// Dropping a plan stamps finished too, but its work was never done.
		{"dropped plan, no box", plan("status: dropped\nfinished: \"2026-10-01 11:00:00\"\n", "**verify:** it works"), "todo", "dropped"},
		{"finished plan, open box", plan("finished: \"2026-10-01 11:00:00\"\n", "- [ ] Red: it fails"), "todo", "approved"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := boardWith(t, c.files)
			p := b.Get(id)
			if p == nil || len(p.Children) != 1 {
				t.Fatalf("%s missing or without one task: %+v", id, p)
			}
			task := b.Get(p.Children[0])
			if task.Status != c.task {
				t.Errorf("task status %s, want %s", task.Status, c.task)
			}
			if p.Status != c.plan {
				t.Errorf("plan status %s, want %s", p.Status, c.plan)
			}
		})
	}
}
