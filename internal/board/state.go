package board

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/gitc"
)

// roundSubjects are the commit subjects that open a review round, newest
// first. The number after one is the round the build is on.
var roundSubjects = []string{"chore(plan): tick fix round ", "acta: tick fix round "}

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
// that has started, has not finished, and has a worktree of its own where the
// branch committed something on the plan file. A plan in the main tree, or one
// that already landed, is not running.
func Running(cfg config.Config) []RunState {
	return planStates(cfg, func(it *Item) bool { return it.StartedOn != "" && it.Finished == "" })
}

// PlanStates gives the plans that live in a linked worktree and that its
// branch worked on, whatever their stamps say. Land reads a plan after the
// last box is ticked, so the per-plan view asks for these and not for the
// running ones.
func PlanStates(cfg config.Config) []RunState {
	return planStates(cfg, nil)
}

// planStates walks the linked worktrees and gives the plan items keep takes,
// with the facts of each worked out on read. A nil keep takes every plan the
// branch worked on.
func planStates(cfg config.Config, keep func(*Item) bool) []RunState {
	wts, err := gitc.Worktrees(cfg.RepoRoot)
	if err != nil {
		return nil
	}
	var out []RunState
	for i, w := range wts {
		// Git lists the main checkout first, so index 0 is the one folder
		// that is not a worktree of itself. Skipping the folder the command
		// runs in instead would hide the worktree a session was started in.
		if i == 0 || strings.HasPrefix(w.Branch, "(") {
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
		// Git copies every plan file into a new worktree, so a plan that
		// only lives in the main tree shows up here too. Only the files the
		// branch committed say that the work is going on in this worktree.
		worked := branchFiles(cfg.RepoRoot, wts[0].Branch, w.Branch)
		for _, it := range b.Items {
			if it.Kind != KindPlan {
				continue
			}
			if !worked[planName(c, it.Path)] {
				continue
			}
			if keep != nil && !keep(it) {
				continue
			}
			out = append(out, runState(cfg.RepoRoot, b, it, w))
		}
	}
	return out
}

// branchFiles gives the files the commits of branch made after it left main.
// It asks git once for the whole branch. A branch git cannot answer for gives
// nothing, so nothing is called running rather than everything.
func branchFiles(repo, main, branch string) map[string]bool {
	out := map[string]bool{}
	files, err := gitc.Touched(repo, main+".."+branch)
	if err != nil {
		return out
	}
	for _, f := range files {
		out[f] = true
	}
	return out
}

// planName gives the plan file's name from the repo root, the way git names
// it, so it can be looked up in the list of files the branch wrote.
func planName(c config.Config, path string) string {
	rel, err := filepath.Rel(c.RepoRoot, path)
	if err != nil {
		return ""
	}
	return filepath.ToSlash(rel)
}

// runState fills in the facts of one running plan. The current task comes from
// the ticks, the last commit and the review round from the branch, and the
// three subsections from the file.
func runState(repo string, b *Board, it *Item, w gitc.Worktree) RunState {
	r := RunState{Plan: it, Branch: w.Branch, Worktree: w.Path, Task: openTask(b, it)}
	if h, s, err := gitc.NewestCommit(repo, w.Branch, ""); err == nil {
		r.Commit = h + " " + s
	}
	if h, n := newestRound(repo, w.Branch); h != "" {
		r.Round = n
	}
	n, f, o := stateOf(it.Body)
	r.Next, r.Findings, r.Rulings = n, f, o
	return r
}

// newestRound gives the round of the newest fix-round commit on the branch,
// checking both the new and the old subject. The newest commit wins
// whatever subject it uses, so old history keeps working while new
// commits use the new subject.
func newestRound(repo, branch string) (string, string) {
	var found []string
	var rounds []string
	for _, pre := range roundSubjects {
		if h, s, err := gitc.NewestCommit(repo, branch, pre); err == nil && s != "" {
			found = append(found, h)
			rounds = append(rounds, strings.TrimSpace(strings.TrimPrefix(s, pre)))
		}
	}
	if len(found) == 0 {
		return "", ""
	}
	if len(found) == 1 {
		return found[0], rounds[0]
	}
	if first, err := gitc.FirstSeen(repo, found[0]); err == nil {
		if second, err := gitc.FirstSeen(repo, found[1]); err == nil {
			if second > first {
				return found[1], rounds[1]
			}
		}
	}
	return found[0], rounds[0]
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
