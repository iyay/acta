package write

import (
	"path/filepath"
	"strings"
	"testing"
)

// addPlan is a plan with two tasks, an earlier fix round and a State section,
// the shape a plan has when a review round adds to it.
const addPlan = "---\nid: PLN-0190\nhash: td8w75r\n---\n# Plan writes\n\n" +
	"### Task 1: first\n\n- [x] code\n\n" +
	"## Fix round 1\n\n### Task 2: fix it\n\n- [x] fix\n\n" +
	"## State\n\n### Next\n\nrun the gate\n"

const addID = "plans/2026-10-05-live-work-state"

// addOK runs AddPlanSection on a fresh repo holding plan, and gives the file
// it left behind and the line the command prints.
func addOK(t *testing.T, plan, section, title, verify, body string) (string, string) {
	t.Helper()
	cfg, path := stateRepo(t, plan)
	o, line, err := AddPlanSection(cfg, mustLoad(t, cfg), addID, section, title, verify, []byte(body))
	if err != nil || !o.Committed {
		t.Fatalf("%s: outcome %+v err %v", section, o, err)
	}
	return stateBody(t, path), line
}

// Each section adds only its own lines, in the place the spec fixes, and
// every other byte of the plan stays where it was.
func TestAddPlanSectionWritesEachSection(t *testing.T) {
	head := "---\nid: PLN-0190\nhash: td8w75r\n---\n# Plan writes\n\n" +
		"### Task 1: first\n\n- [x] code\n\n" +
		"## Fix round 1\n\n### Task 2: fix it\n\n- [x] fix\n\n"
	state := "## State\n\n### Next\n\nrun the gate\n"
	for _, c := range []struct {
		name, plan, section, title, verify, body, want, line string
	}{
		{
			name: "fix round before State", plan: addPlan, section: "fix-round",
			title: "Guard empty input", verify: "empty input is refused",
			body: "- [ ] BLOCKER: empty input crashes\n[fix] rename x\n",
			want: head + "## Fix round 2\n\n### Task 3: Guard empty input\n\n**verify:** empty input is refused\n\n" +
				"- [ ] BLOCKER: empty input crashes\n- [ ] [fix] rename x\n- [ ] Commit with the trailer `Task: PLN-td8w75r#3`\n\n" + state,
			line: addID + "#task-3 added",
		},
		{
			name: "polish before State", plan: addPlan, section: "polish",
			body: "- rename x\nadd a test\n",
			want: head + "## Polish\n\n### Task 3: Review polish\n\n**verify:** every NOTE below is applied, and nothing else changes.\n\n" +
				"- [ ] rename x\n- [ ] add a test\n- [ ] Commit: `polish: review notes for PLN-0190`\n\n" + state,
			line: addID + "#task-3 added",
		},
		{
			name: "review notes created before State", plan: addPlan, section: "review-notes",
			body: "Deep review, round 1 CLEAN.\n",
			want: head + "## Review notes\n\n- Deep review, round 1 CLEAN.\n\n" + state,
			line: "PLN-0190",
		},
		{
			name: "review notes add to what is there", section: "review-notes",
			plan: head + "## Review notes\n\nfirst line\n\n" + state,
			body: "- second line\n",
			want: head + "## Review notes\n\nfirst line\n- second line\n\n" + state,
			line: "PLN-0190",
		},
		{
			name: "polish before Review notes", section: "polish",
			plan: head + "## Review notes\n\nfirst line\n",
			body: "rename x\n",
			want: head + "## Polish\n\n### Task 3: Review polish\n\n**verify:** every NOTE below is applied, and nothing else changes.\n\n" +
				"- [ ] rename x\n- [ ] Commit: `polish: review notes for PLN-0190`\n\n## Review notes\n\nfirst line\n",
			line: addID + "#task-3 added",
		},
		{
			name: "fix round at the end", section: "fix-round", title: "T", verify: "V",
			plan: "---\nid: PLN-0190\nhash: td8w75r\n---\n# Plan writes\n\n### Task 01: first\n\n- [ ] code\n",
			body: "fix\n",
			want: "---\nid: PLN-0190\nhash: td8w75r\n---\n# Plan writes\n\n### Task 01: first\n\n- [ ] code\n\n" +
				"## Fix round 1\n\n### Task 2: T\n\n**verify:** V\n\n- [ ] fix\n- [ ] Commit with the trailer `Task: PLN-td8w75r#2`\n",
			line: addID + "#task-2 added",
		},
		{
			name: "review notes at the end", section: "review-notes",
			plan: "---\nid: PLN-0190\n---\n# Plan writes\n\n### Task 1: first\n\n- [ ] code\n",
			body: "one\n\ntwo\n",
			want: "---\nid: PLN-0190\n---\n# Plan writes\n\n### Task 1: first\n\n- [ ] code\n\n## Review notes\n\n- one\n- two\n",
			line: "PLN-0190",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, line := addOK(t, c.plan, c.section, c.title, c.verify, c.body)
			if got != c.want {
				t.Fatalf("file = %q\nwant %q", got, c.want)
			}
			if line != c.line {
				t.Fatalf("line = %q, want %q", line, c.line)
			}
		})
	}
}

// The new task is a real task: the board finds it under the number the
// command printed, and tick can tick every box of it.
func TestAddPlanSectionTaskCanBeTicked(t *testing.T) {
	for _, section := range []string{"fix-round", "polish"} {
		cfg, path := stateRepo(t, addPlan)
		if _, _, err := AddPlanSection(cfg, mustLoad(t, cfg), addID, section, "T", "V", []byte("a\nb\n")); err != nil {
			t.Fatal(err)
		}
		it := mustLoad(t, cfg).Get(addID + "#task-3")
		if it == nil {
			t.Fatalf("%s: board has no task-3", section)
		}
		done, total, err := Tick(path, it.Line, 0)
		if err != nil || done != 3 || total != 3 {
			t.Fatalf("%s: tick gave %d/%d err %v, want 3/3", section, done, total, err)
		}
	}
}

// The commit folds the change in the way acta state set does, and it names
// the plan.
func TestAddPlanSectionCommits(t *testing.T) {
	cfg, _ := stateRepo(t, addPlan)
	if _, _, err := AddPlanSection(cfg, mustLoad(t, cfg), addID, "review-notes", "", "", []byte("x\n")); err != nil {
		t.Fatal(err)
	}
	if got := gitRun(t, cfg.RepoRoot, "show", "--name-only", "--format=", "HEAD"); got != ".acta/plans/2026-10-05-live-work-state.md" {
		t.Fatalf("commit holds %q", got)
	}
	if got := gitRun(t, cfg.RepoRoot, "log", "-1", "--format=%s"); !strings.Contains(got, "PLN-0190") {
		t.Fatalf("commit subject %q does not name the plan", got)
	}
}

// Every bad request is refused with no file change and no commit.
func TestAddPlanSectionRefuses(t *testing.T) {
	withPolish := strings.Replace(addPlan, "## State", "## Polish\n\n### Task 3: Review polish\n\n- [ ] x\n\n## State", 1)
	for _, c := range []struct {
		name, plan, id, section, title, verify, body, want string
	}{
		{name: "unknown section", section: "notes", body: "x\n", want: "fix-round, polish or review-notes"},
		{name: "empty stdin", section: "review-notes", body: "", want: "nothing on stdin"},
		{name: "blank stdin", section: "polish", body: "\n  \n\n", want: "nothing on stdin"},
		{name: "fix round without title", section: "fix-round", verify: "V", body: "x\n", want: "--title"},
		{name: "fix round without verify", section: "fix-round", title: "T", body: "x\n", want: "--verify"},
		{name: "two line title", section: "fix-round", title: "T\nU", verify: "V", body: "x\n", want: "one line"},
		{name: "second polish", plan: withPolish, section: "polish", body: "x\n", want: "already has ## Polish"},
		{name: "not a plan", id: "BUG-1", section: "review-notes", body: "x\n", want: "is not a plan"},
		{name: "unknown id", id: "plans/2026-10-06-nope", section: "review-notes", body: "x\n", want: "unknown id"},
	} {
		t.Run(c.name, func(t *testing.T) {
			plan := c.plan
			if plan == "" {
				plan = addPlan
			}
			id := c.id
			if id == "" {
				id = addID
			}
			cfg := repoWith(t, map[string]string{
				".acta/plans/2026-10-05-live-work-state.md": plan,
				".acta/bugs/2026-10-05-crash.md":            "---\nid: BUG-1\n---\n# Crash\n",
			})
			path := filepath.Join(cfg.RepoRoot, ".acta", "plans", "2026-10-05-live-work-state.md")
			head := gitRun(t, cfg.RepoRoot, "rev-parse", "HEAD")
			_, _, err := AddPlanSection(cfg, mustLoad(t, cfg), id, c.section, c.title, c.verify, []byte(c.body))
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want one naming %q", err, c.want)
			}
			if got := stateBody(t, path); got != plan {
				t.Fatalf("file changed: %q", got)
			}
			if gitRun(t, cfg.RepoRoot, "rev-parse", "HEAD") != head {
				t.Fatal("a refused call made a commit")
			}
		})
	}
}

// A heading inside a code block is an example, not a section, so it is not
// counted and the new section does not land inside the block.
func TestAddPlanSectionSkipsCodeBlocks(t *testing.T) {
	plan := "---\nid: PLN-0190\nhash: td8w75r\n---\n# Plan writes\n\n### Task 1: first\n\n- [ ] code\n\n" +
		"```\n## Fix round 7\n### Task 9: x\n## State\n```\n"
	got, line := addOK(t, plan, "fix-round", "T", "V", "fix\n")
	if line != addID+"#task-2 added" {
		t.Fatalf("line = %q", line)
	}
	if !strings.HasSuffix(got, "```\n\n## Fix round 1\n\n### Task 2: T\n\n**verify:** V\n\n- [ ] fix\n- [ ] Commit with the trailer `Task: PLN-td8w75r#2`\n") {
		t.Fatalf("file = %q", got)
	}
}
