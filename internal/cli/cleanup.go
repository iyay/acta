package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/cleanup"
	"github.com/iyay/acta/internal/config"
)

// cmdCleanup lists the planning items whose work may be over. It only reads:
// the user decides what to close, and other subcommands write it.
func cmdCleanup(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("cleanup", stderr)
	asJSON := fs.Bool("json", false, "print one JSON object per line")
	fix := fs.Bool("fix", false, "make the changes that need no judgment, in one commit")
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) > 0 {
		fmt.Fprintln(stderr, usageLine("cleanup"))
		return exitBadInput
	}
	cfg, b, code := loadBoard(*root, stderr)
	if code != exitOK {
		return code
	}
	if *fix {
		return runFix(cfg, b, stdout, stderr)
	}
	list, err := cleanup.Candidates(cfg, b, time.Now())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	if *asJSON {
		enc := json.NewEncoder(stdout)
		for _, c := range list {
			if err := enc.Encode(c); err != nil {
				fmt.Fprintln(stderr, err)
				return exitOther
			}
		}
		return exitOK
	}
	kind := ""
	for _, c := range list {
		if c.Kind != kind {
			kind = c.Kind
			fmt.Fprintf(stdout, "%s\n", kind)
		}
		fmt.Fprintf(stdout, "  %4dd  %s  %s  (%s)\n", c.AgeDays, c.ID, c.Title, c.Why)
	}
	changes, err := cleanup.Plan(cfg, b)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	if len(changes) > 0 {
		fmt.Fprintln(stdout, "fix (acta cleanup --fix)")
		printChanges(stdout, changes)
	}
	if len(list) == 0 && len(changes) == 0 {
		fmt.Fprintln(stdout, "nothing to clean up")
	}
	return exitOK
}

// runFix writes the changes that need no judgment and says what it did.
func runFix(cfg config.Config, b *board.Board, stdout, stderr io.Writer) int {
	o, changes, err := cleanup.Fix(cfg, b)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	if len(changes) == 0 {
		fmt.Fprintln(stdout, "nothing to fix")
		return exitOK
	}
	printChanges(stdout, changes)
	return report(o, nil, stdout, stderr)
}

func printChanges(w io.Writer, changes []cleanup.Change) {
	for _, c := range changes {
		fmt.Fprintf(w, "  %s  %s\n", c.ID, c.What)
	}
}

// cmdCleanupApply writes the decisions the user approved, one JSON object per
// line on stdin, as one commit. A batch with any bad line writes nothing.
func cmdCleanupApply(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs, root := flags("cleanup apply", stderr)
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) > 0 {
		fmt.Fprintln(stderr, usageLine("cleanup"))
		return exitBadInput
	}
	ds, err := cleanup.ParseDecisions(stdin)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	cfg, b, code := loadBoard(*root, stderr)
	if code != exitOK {
		return code
	}
	o, err := cleanup.Apply(cfg, b, ds)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	return report(o, nil, stdout, stderr)
}
