package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/worktrees"
)

// cmdWorktrees lists the linked worktrees and what the stray rules make of
// each. With --clean it removes the clean strays, never by force.
func cmdWorktrees(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("worktrees", stderr)
	asJSON := fs.Bool("json", false, "print one JSON object per line")
	clean := fs.Bool("clean", false, "remove the clean strays and delete their merged branches")
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) > 0 {
		fmt.Fprintln(stderr, usageLine("worktrees"))
		return exitBadInput
	}
	cfg, b, code := loadBoard(*root, stderr)
	if code != exitOK {
		return code
	}
	run := worktrees.List
	if *clean {
		run = worktrees.Clean
	}
	// Like every acta command, the repo comes from the folder it runs in;
	// --root only moves the planning folder.
	list, err := run(cfg.RepoRoot, planList(b), time.Now())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	if *asJSON {
		enc := json.NewEncoder(stdout)
		for _, e := range list {
			if err := enc.Encode(e); err != nil {
				fmt.Fprintln(stderr, err)
				return exitOther
			}
		}
		return exitOK
	}
	if len(list) == 0 {
		fmt.Fprintln(stdout, "no linked worktrees")
	}
	for _, e := range list {
		branch := e.Branch
		if branch == "" {
			branch = "detached"
		}
		fmt.Fprintf(stdout, "%-9s  %s  (%s)", e.State, e.Path, branch)
		if e.BranchKept {
			fmt.Fprint(stdout, "  branch kept")
		}
		if e.Reason != "" {
			fmt.Fprintf(stdout, "  %s", e.Reason)
		}
		fmt.Fprintln(stdout)
		for _, ln := range e.Status {
			fmt.Fprintf(stdout, "           %s\n", ln)
		}
	}
	return exitOK
}

// planList gives the slug and status of every plan, closed ones too, since
// the stray rules need to know which plans are done or dropped.
func planList(b *board.Board) []worktrees.Plan {
	var out []worktrees.Plan
	for _, it := range b.List(board.KindPlan, true) {
		out = append(out, worktrees.Plan{Slug: it.Slug, Status: it.Status})
	}
	return out
}
