package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/telemetry"
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
			fmt.Fprintln(stderr, usageLine("id"))
			return exitBadInput
		}
		changes, o, err = write.FixDuplicates(cfg, b)
	} else {
		had := shortIDs(b)
		changes, o, err = write.AssignIDs(cfg, b, pos)
		if err == nil {
			countNewIDs(had, changes)
		}
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

// shortIDs lists every number the board holds before acta id runs, so a
// number that shows up only after the run is known to be new.
func shortIDs(b *board.Board) map[string]bool {
	had := map[string]bool{}
	for _, it := range b.Items {
		had[it.ShortID] = true
	}
	return had
}

// countNewIDs counts each spec and plan that got its first number. A change
// line starts with the id; one the board had already only got a hash or a
// new form, so it is no new item.
func countNewIDs(had map[string]bool, changes []string) {
	keys := map[string]string{"SPC": "spec_created", "PLN": "plan_created"}
	s := telemetry.Default()
	for _, c := range changes {
		id, _, _ := strings.Cut(c, " ")
		prefix, _, _ := strings.Cut(id, "-")
		if key, ok := keys[prefix]; ok && !had[id] {
			_ = s.Add("workflow", key)
		}
	}
}
