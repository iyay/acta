package board

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/gitc"
)

// roundSubject is the commit subject that opens a review round. The number
// after it is the round the build is on.
const roundSubject = "acta: tick fix round "

// stateHeads are the three subsections a State section holds, in the order the
// file fixes. The writer in internal/write keeps the same names.
var stateHeads = []string{"Next", "Findings", "Open rulings"}

// RunState is one running plan with what the work is standing on. Every field
// but the three subsections is worked out on read, from the plan ticks or from
// git, so none of it can go stale.
type RunState struct {
	Plan     *Item    // the plan item
	Task     *Item    // first task with an open box, nil when every box is ticked
	Branch   string   // branch the plan's worktree is on
	Worktree string   // where that worktree is on disk
	Commit   string   // short hash and subject of the branch's newest commit
	Round    string   // the newest review round, "" when the branch has none
	Next     []string // lines under ### Next
	Findings []string // lines under ### Findings
	Rulings  []string // lines under ### Open rulings
}

// Running gives the plans of this repo that work is going on right now: one
// that has started, has not finished, and has a worktree of its own. A plan
// in the main tree, or one that already landed, is not running.
func Running(cfg config.Config) []RunState {
	wts, err := gitc.Worktrees(cfg.RepoRoot)
	if err != nil {
		return nil
	}
	var out []RunState
	for _, w := range wts {
		if sameDir(w.Path, cfg.RepoRoot) || strings.HasPrefix(w.Branch, "(") {
			continue
		}
		c, err := config.Load(w.Path, "")
		if err != nil {
			continue
		}
		b, err := Load(c)
		if err != nil {
			continue
		}
		for _, it := range b.Items {
			if it.Kind != KindPlan || it.StartedOn == "" || it.Finished != "" {
				continue
			}
			out = append(out, runState(cfg.RepoRoot, b, it, w))
		}
	}
	return out
}

// runState fills in the facts of one running plan. The current task comes from
// the ticks, the last commit and the review round from the branch, and the
// three subsections from the file.
func runState(repo string, b *Board, it *Item, w gitc.Worktree) RunState {
	r := RunState{Plan: it, Branch: w.Branch, Worktree: w.Path, Task: openTask(b, it)}
	if h, s, err := gitc.NewestCommit(repo, w.Branch, ""); err == nil {
		r.Commit = h + " " + s
	}
	if _, s, err := gitc.NewestCommit(repo, w.Branch, roundSubject); err == nil {
		r.Round = strings.TrimSpace(strings.TrimPrefix(s, roundSubject))
	}
	n, f, o := stateOf(it.Body)
	r.Next, r.Findings, r.Rulings = n, f, o
	return r
}

// openTask gives the first task of the plan with a box still open, or nil when
// every box is ticked.
func openTask(b *Board, plan *Item) *Item {
	for _, id := range plan.Children {
		t := b.Get(id)
		if t != nil && t.Done < t.Total {
			return t
		}
	}
	return nil
}

// stateOf reads the three State subsections out of a plan body. A subsection
// the file does not hold comes back empty, so nothing is printed for it.
func stateOf(body string) (next, findings, rulings []string) {
	var out [3][]string
	in, at := false, -1
	for _, ln := range strings.Split(body, "\n") {
		s := strings.TrimRight(ln, " \t\r")
		switch {
		case s == "## State":
			in, at = true, -1
		case strings.HasPrefix(s, "## "):
			in, at = false, -1
		case !in:
		case strings.HasPrefix(s, "### "):
			at = -1
			for i, h := range stateHeads {
				if s == "### "+h {
					at = i
				}
			}
		case at >= 0 && strings.TrimSpace(ln) != "":
			out[at] = append(out[at], s)
		}
	}
	return out[0], out[1], out[2]
}

// Full gives the whole view: the plan, the four facts worked out on read, and
// then the State subsections as the file holds them.
func (r RunState) Full() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s  %s\n", r.Plan.ID, r.Plan.Title)
	if r.Task != nil {
		fmt.Fprintf(&b, "task: %s  %s\n", r.Task.ID, r.Task.Title)
	} else {
		b.WriteString("task: none, every box is ticked\n")
	}
	fmt.Fprintf(&b, "last commit: %s\n", r.Commit)
	fmt.Fprintf(&b, "worktree: %s  branch: %s\n", r.Worktree, r.Branch)
	if r.Round != "" {
		fmt.Fprintf(&b, "round: %s\n", r.Round)
	} else {
		b.WriteString("round: none\n")
	}
	for i, h := range stateHeads {
		lines := [...][]string{r.Next, r.Findings, r.Rulings}[i]
		if len(lines) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n### %s\n\n%s\n", h, strings.Join(lines, "\n"))
	}
	return b.String()
}

// Short gives at most three lines for a session that only needs to know what
// is going on: which plan and task, where it is, and the first Next line.
func (r RunState) Short() []string {
	task := "no open task"
	if r.Task != nil {
		task = r.Task.ID
	}
	out := []string{
		fmt.Sprintf("%s  %s  task: %s", r.Plan.ID, r.Plan.Title, task),
		fmt.Sprintf("worktree: %s  last commit: %s", r.Worktree, r.Commit),
	}
	if len(r.Next) > 0 {
		out = append(out, "next: "+r.Next[0])
	}
	return out
}

// sameDir says whether two paths name the same folder, following symlinks so
// a temp dir and its real path still count as one.
func sameDir(a, b string) bool {
	ra, err1 := filepath.EvalSymlinks(a)
	rb, err2 := filepath.EvalSymlinks(b)
	if err1 != nil || err2 != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}
	return ra == rb
}
