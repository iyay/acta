package write

import (
	"bytes"
	"fmt"
	"os"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/gitc"
)

// TaskDates writes the dates a task tick gives its plan and the spec or bug
// above it. Work began the moment any task is ticked, and work ends when
// every task of the plan is done, and the spec a little later: when every
// plan under it is done too. The board is read again because the tick that
// just happened is what the file says. It gives back the files it changed.
func TaskDates(cfg config.Config, it *board.Item) ([]string, error) {
	fresh, err := board.Load(cfg)
	if err != nil {
		return nil, err
	}
	var wrote []string
	mark := func(path string, set func([]byte) ([]byte, error)) error {
		changed, err := markDate(path, set)
		if changed {
			wrote = append(wrote, path)
		}
		return err
	}
	plan := fresh.Get(it.PlanID)
	if plan == nil {
		// A task whose plan the fresh board cannot see has no dates to
		// write, and a press must never end in a crash.
		return wrote, nil
	}
	if err := mark(plan.Path, MarkStarted); err != nil {
		return wrote, err
	}
	if plan.Done == plan.Total {
		if err := mark(plan.Path, MarkFinishedOnce); err != nil {
			return wrote, err
		}
	}
	spec := fresh.Get(plan.SpecID)
	if spec == nil {
		return wrote, nil
	}
	if err := mark(spec.Path, MarkStarted); err != nil {
		return wrote, err
	}
	if spec.Status == "done" {
		return wrote, mark(spec.Path, MarkFinishedOnce)
	}
	return wrote, nil
}

// markDate runs one date writer over a file and writes the result back only
// when the bytes really changed, so a tick never touches a file it has
// nothing new to say about.
func markDate(path string, set func([]byte) ([]byte, error)) (bool, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	out, err := set(src)
	if err != nil {
		return false, err
	}
	if bytes.Equal(src, out) {
		return false, nil
	}
	return true, os.WriteFile(path, out, 0o644)
}

// MarkItem sets every box of a task, or the one box of a debt line, to done
// or back to open. It is what the TUI + and - keys call, so a person gets one
// commit per press, the same way the status popup works.
func MarkItem(cfg config.Config, b *board.Board, id string, done bool) (Outcome, error) {
	it := b.Get(id)
	switch {
	case it == nil:
		return Outcome{}, bad("unknown id %s", id)
	case it.Legacy:
		return Outcome{}, bad("%s is a legacy file; move it into the root folder first", id)
	case it.Worktree != "":
		return Outcome{}, bad("%s is shown from worktree %s; edit it there", id, it.Worktree)
	case it.Kind != board.KindTask && it.Kind != board.KindDebtItem:
		return Outcome{}, bad("%s is not a task or a debt line", id)
	}
	// Check every file this press can touch before writing any of them.
	// Changes a person already had in one of them are not ours to commit.
	paths := []string{it.Path}
	if it.Kind == board.KindTask && done {
		if plan := b.Get(it.PlanID); plan != nil {
			if spec := b.Get(plan.SpecID); spec != nil {
				paths = append(paths, spec.Path)
			}
		}
	}
	dirty := false
	for _, p := range paths {
		d, err := dirtyBefore(cfg, p)
		if err != nil {
			return Outcome{}, err
		}
		dirty = dirty || d
	}
	state, word := byte(' '), "open"
	if done {
		state, word = 'x', "done"
	}
	wrote := []string{it.Path}
	switch {
	case it.Kind == board.KindDebtItem:
		if err := TickLine(it.Path, it.Line, state); err != nil {
			return Outcome{}, err
		}
	case done:
		if _, _, err := Tick(it.Path, it.Line, 0); err != nil {
			return Outcome{}, err
		}
		dated, err := TaskDates(cfg, it)
		wrote = append(wrote, dated...)
		if err != nil {
			return Outcome{Path: it.Path}, err
		}
	default:
		if _, _, err := Untick(it.Path, it.Line); err != nil {
			return Outcome{}, err
		}
	}
	msg := fmt.Sprintf("acta: %s %s", id, word)
	o := Outcome{Path: it.Path}
	switch {
	case !cfg.AutoCommit:
		o.Reason = "auto_commit is off"
	case dirty:
		o.Skipped, o.Reason = true, "file had other uncommitted changes"
	default:
		r := gitc.CommitPaths(cfg.RepoRoot, unique(wrote), msg)
		o.Committed, o.Reason, o.Skipped = r.Committed, r.Reason, !r.Committed
	}
	return o, nil
}

// unique drops repeats, since the plan file is both ticked and dated.
func unique(paths []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range paths {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}
