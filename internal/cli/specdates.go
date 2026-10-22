package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/iyay/acta/internal/gitc"
	"github.com/iyay/acta/internal/write"
)

// cmdSpecDates writes the dates of a plan's spec, on the branch the plan
// just landed on, and commits them right away. Land runs it
// before it moves refs/acta/last-land, so the parent is never left with an
// uncommitted spec that would block the next land. It never folds into an
// older commit: the commits tidy just built stay as they are.
func cmdSpecDates(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("spec-dates", stderr)
	pos, err := parseMixed(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitBadInput
	}
	if len(pos) != 1 {
		fmt.Fprintln(stderr, usageLine("spec-dates"))
		return exitBadInput
	}
	cfg, code := loadConfig(*root, stderr)
	if code != exitOK {
		return code
	}
	wrote, err := write.SpecDates(cfg, pos[0])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	for _, path := range wrote {
		res := gitc.Commit(cfg.RepoRoot, path, "chore(spec): dates of "+gitc.Stem(path), false)
		if !res.Committed {
			fmt.Fprintln(stderr, "written, not committed:", res.Reason)
			return exitOther
		}
		fmt.Fprintln(stdout, path)
	}
	return exitOK
}
