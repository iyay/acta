package cleanup

import (
	"fmt"
	"os"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/gitc"
	"github.com/iyay/acta/internal/write"
)

// Change is one write --fix makes. Out holds the new file text, worked out
// before any file is written.
type Change struct {
	ID   string
	Path string
	What string
	out  []byte
}

// Plan works out every change --fix would make, without writing. It fails
// when any change cannot be made, so a fix never lands half done.
func Plan(cfg config.Config, b *board.Board) ([]Change, error) {
	building := map[string]bool{}
	for _, r := range board.PlanStates(cfg) {
		building[r.Plan.ID] = true
	}
	var out []Change
	for _, it := range b.Items {
		if it.Legacy || it.Worktree != "" || building[it.ID] {
			continue
		}
		c, ok, err := fixFor(it)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		dirty, err := gitc.IsDirty(cfg.RepoRoot, it.Path)
		if err != nil {
			return nil, err
		}
		if dirty {
			return nil, fmt.Errorf("%s has changes that are not committed; commit or undo them first", rel(cfg, it.Path))
		}
		out = append(out, c)
	}
	return out, nil
}

// fixFor gives the one change an item needs, if any.
func fixFor(it *board.Item) (Change, bool, error) {
	ignored := false
	for _, p := range it.Problems {
		ignored = ignored || strings.HasPrefix(p, "written status ")
	}
	// Every box is ticked, so the work is over; only a word left in the
	// frontmatter keeps the plan open. A plan with no box proves nothing.
	ticked := it.Kind == board.KindPlan && it.Total > 0 && it.Done == it.Total &&
		it.Status != "done" && it.Status != "review" && it.Status != "dropped"
	if !ignored && !ticked {
		return Change{}, false, nil
	}
	src, err := os.ReadFile(it.Path)
	if err != nil {
		return Change{}, false, err
	}
	c := Change{ID: it.ID, Path: it.Path}
	if ticked {
		c.What = "every task is ticked: status done"
		c.out, err = write.SetField(src, "status", "done")
		if err == nil {
			c.out, err = write.DatesFor(c.out, "done")
		}
	} else {
		c.What = "drop the written status the board ignores"
		c.out, err = write.RemoveField(src, "status")
	}
	if err != nil {
		return Change{}, false, fmt.Errorf("%s: %w", it.ID, err)
	}
	return c, true, nil
}

// Fix writes every change Plan finds and commits them together.
func Fix(cfg config.Config, b *board.Board) (write.Outcome, []Change, error) {
	changes, err := Plan(cfg, b)
	if err != nil {
		return write.Outcome{}, nil, err
	}
	o := write.Outcome{Path: cfg.Root}
	if len(changes) == 0 {
		o.Skipped, o.Reason = true, "nothing to fix"
		return o, nil, nil
	}
	paths := make([]string, 0, len(changes))
	for _, c := range changes {
		if err := os.WriteFile(c.Path, c.out, 0o644); err != nil {
			return write.Outcome{}, nil, err
		}
		paths = append(paths, c.Path)
	}
	return commit(cfg, paths, "cleanup fix", o), changes, nil
}

// commit makes one commit of every path, unless the user turned auto-commit
// off.
func commit(cfg config.Config, paths []string, what string, o write.Outcome) write.Outcome {
	if !cfg.AutoCommit {
		o.Reason = "auto_commit is off"
		return o
	}
	r := gitc.CommitPathsIn(cfg.RepoRoot, cfg.Root, paths, write.Subject(cfg, paths, what))
	o.Committed, o.Reason, o.Skipped = r.Committed, r.Reason, !r.Committed
	return o
}
