package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/iyay/acta/internal/cleanup"
)

// cmdCleanup lists the planning items whose work may be over. It only reads:
// the user decides what to close, and other subcommands write it.
func cmdCleanup(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("cleanup", stderr)
	asJSON := fs.Bool("json", false, "print one JSON object per line")
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) > 0 {
		fmt.Fprintln(stderr, usageLine("cleanup"))
		return exitBadInput
	}
	cfg, b, code := loadBoard(*root, stderr)
	if code != exitOK {
		return code
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
	if len(list) == 0 {
		fmt.Fprintln(stdout, "nothing to clean up")
	}
	return exitOK
}
