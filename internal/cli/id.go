package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/iyay/acta/internal/write"
)

// cmdID gives short number and hash IDs to items that lack them, or repairs
// duplicate numbers after a merge. It prints one line per change and stays
// silent when there is nothing to do.
func cmdID(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("id", stderr)
	fix := fs.Bool("fix-duplicates", false, "give the later of two items with one number the next free number")
	pos, err := parseMixed(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitBadInput
	}
	cfg, b, code := loadAllTrees(*root, stderr)
	if code != exitOK {
		return code
	}
	var changes []string
	var o write.Outcome
	if *fix {
		if len(pos) > 0 {
			fmt.Fprintln(stderr, "usage: acta id [--fix-duplicates] [<path-id>...]")
			return exitBadInput
		}
		changes, o, err = write.FixDuplicates(cfg, b)
	} else {
		changes, o, err = write.AssignIDs(cfg, b, pos)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		if errors.Is(err, write.ErrBadInput) {
			return exitBadInput
		}
		return exitOther
	}
	for _, c := range changes {
		fmt.Fprintln(stdout, c)
	}
	for _, s := range o.Skips {
		fmt.Fprintln(stderr, s)
	}
	if o.Skipped {
		fmt.Fprintln(stderr, "written, not committed:", o.Reason)
		return exitSkipped
	}
	return exitOK
}
