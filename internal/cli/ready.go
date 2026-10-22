package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/iyay/acta/internal/board"
)

// The exit codes of acta ready. Waiting is not an error: the plan is fine,
// it just cannot start yet. A bad after list can never become ready.
const (
	readyWaiting = 1
	readyBad     = 2
)

// cmdReady says whether a plan may start: every plan in its after list must
// have landed on the parent branch. It reads the parent's committed tree,
// never the board, because the board also counts work still in worktrees.
func cmdReady(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("ready", stderr)
	parent := fs.String("parent", "", "the branch the build will land on")
	head := fs.String("head", "", "the worktree head that must already hold every landed after plan")
	pos, err := parseMixed(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitOther
	}
	if len(pos) != 1 || *parent == "" {
		fmt.Fprintln(stderr, usageLine("ready"))
		return exitOther
	}
	// Usage and load errors exit 3: exit 2 means the after list itself is bad.
	cfg, b, code := loadBoard(*root, stderr)
	if code != exitOK {
		return exitOther
	}
	if _, err := readyGit(cfg.RepoRoot, "rev-parse", "--verify", "--quiet", *parent+"^{commit}"); err != nil {
		fmt.Fprintf(stderr, "%s is not a branch or commit here\n", *parent)
		return readyBad
	}
	plan := b.Get(pos[0])
	if plan == nil || plan.Kind != board.KindPlan {
		fmt.Fprintf(stderr, "%s is not a plan\n", pos[0])
		return readyBad
	}
	if msg := badAfter(b, plan); msg != "" {
		fmt.Fprintln(stderr, msg)
		return readyBad
	}
	var waits []string
	for _, id := range plan.After {
		dep := b.Get(id)
		rel, err := repoRel(cfg.RepoRoot, dep.Path)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return readyBad
		}
		landed := landingCommit(cfg.RepoRoot, *parent, rel)
		switch {
		case landed == "":
			waits = append(waits, fmt.Sprintf("%s has not landed on %s", id, *parent))
		case *head != "" && !isAncestor(cfg.RepoRoot, landed, *head):
			waits = append(waits, fmt.Sprintf("%s landed on %s, but %s does not hold it yet; take in %s first", id, *parent, *head, *parent))
		}
	}
	if len(waits) > 0 {
		fmt.Fprintf(stdout, "%s waits:\n  %s\n", planName(plan), strings.Join(waits, "\n  "))
		return readyWaiting
	}
	fmt.Fprintf(stdout, "%s is ready\n", planName(plan))
	return exitOK
}

// badAfter checks the after lists reachable from plan and says what is wrong
// with them, or "" when nothing is. A cycle is told as its chain, so the
// user sees which plan to change.
func badAfter(b *board.Board, plan *board.Item) string {
	var walk func(it *board.Item, path []string) string
	walk = func(it *board.Item, path []string) string {
		for _, id := range it.After {
			dep := b.Get(id)
			switch {
			case dep == nil:
				return fmt.Sprintf("%s lists %s in after, and no item has that id", planName(it), id)
			case dep.Kind != board.KindPlan:
				return fmt.Sprintf("%s lists %s in after, but %s is not a plan", planName(it), id, id)
			// By id, not by pointer: acta set checks a copy of the plan
			// that holds the new list.
			case dep.ID == it.ID:
				return fmt.Sprintf("%s waits on itself", planName(it))
			case dep.Status == "dropped":
				return fmt.Sprintf("%s was dropped and will never land; take it out of the after list of %s", id, planName(it))
			}
			name := planName(dep)
			for i, seen := range path {
				if seen == name {
					return "after lists form a cycle: " + strings.Join(append(path[i:], name), " -> ")
				}
			}
			if msg := walk(dep, append(path, name)); msg != "" {
				return msg
			}
		}
		return ""
	}
	return walk(plan, []string{planName(plan)})
}

// landingCommit is the first commit on ref whose copy of the plan at rel has
// every box ticked, or "" when ref never held it so. That commit is where
// the plan landed.
func landingCommit(repo, ref, rel string) string {
	out, err := readyGit(repo, "log", "--reverse", "--format=%H", ref, "--", rel)
	if err != nil {
		return ""
	}
	for _, sha := range strings.Fields(out) {
		body, err := readyGit(repo, "show", sha+":"+filepath.ToSlash(rel))
		if err == nil && allTicked(board.Parse([]byte(body))) {
			return sha
		}
	}
	return ""
}

func isAncestor(repo, commit, head string) bool {
	return exec.Command("git", "-C", repo, "merge-base", "--is-ancestor", commit, head).Run() == nil
}

func readyGit(repo string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", repo}, args...)...).Output()
	return string(out), err
}

// planName is the number id a person reads, or the path id when there is none.
func planName(it *board.Item) string {
	if it.ShortID != "" {
		return it.ShortID
	}
	return it.ID
}

// repoRel is path relative to the repo root, as git show wants it. Both sides
// go through EvalSymlinks first: a temp folder on macOS can be reached as
// /var or /private/var, and a plain Rel of the two climbs out with "..".
func repoRel(repo, path string) (string, error) {
	r, err := filepath.EvalSymlinks(repo)
	if err != nil {
		return "", err
	}
	p, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	return filepath.Rel(r, p)
}

// allTicked reads a plan the way the board does: only boxes inside task
// sections count, so an example box in a code fence or a note never holds
// the plan back. A task with no box counts as done once the plan has a
// finished date, as on the board.
func allTicked(d board.Doc) bool {
	if len(d.Tasks) == 0 {
		return false
	}
	// The key alone is not enough: an empty, null or broken finished is no
	// date, and the board does not count it either.
	finished := board.DateField(d.Front, "finished") != ""
	for _, task := range d.Tasks {
		if task.Total == 0 && !finished || task.Done < task.Total {
			return false
		}
	}
	return true
}
