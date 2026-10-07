package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/iyay/acta/internal/tidy"
)

// cmdTidy folds a branch into one commit per code task and prints the
// count. Every failure exits non-zero and leaves all refs as they were,
// because tidy.Run writes its ref only after the tree proof holds.
func cmdTidy(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("tidy", stderr)
	onto := fs.String("onto", "", "ref whose chore commits fold forward into the first new commit")
	pos, err := parseMixed(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitBadInput
	}
	if len(pos) != 2 {
		fmt.Fprintln(stderr, "usage: acta tidy <base> <branch> [--onto <ref>]")
		return exitBadInput
	}
	cfg, code := loadConfig(*root, stderr)
	if code != exitOK {
		return code
	}
	// Run wants the planning root relative to the repo root.
	rel, err := filepath.Rel(cfg.RepoRoot, cfg.Root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	res, err := tidy.Run(cfg.RepoRoot, tidy.Options{Base: pos[0], Branch: pos[1], Onto: *onto, PlanningRoot: rel})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	parent := res.Parent
	if len(parent) > 7 {
		parent = parent[:7]
	}
	fmt.Fprintf(stdout, "tidy: %d commits -> %d, tree ok, parent %s, folded %d\n", res.OldCount, res.NewCount, parent, res.Folded)
	return exitOK
}
