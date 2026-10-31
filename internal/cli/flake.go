package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/write"
)

// cmdFlakeRecord writes one flake of a test onto the bug that tracks it and
// prints that bug on one line, so the caller sees which bug grew and by how
// much.
func cmdFlakeRecord(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("flake record", stderr)
	plan := fs.String("plan", "", "plan id the flake showed up in")
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) != 1 {
		fmt.Fprintln(stderr, usageLine("flake"))
		return exitBadInput
	}
	cfg, code := loadConfig(*root, stderr)
	if code != exitOK {
		return code
	}
	o, n, err := write.RecordFlake(cfg, pos[0], *plan)
	switch {
	case errors.Is(err, write.ErrBadInput):
		fmt.Fprintln(stderr, err)
		return exitBadInput
	case err != nil:
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	// RecordFlake does not give the title back. A new bug is always called
	// this, and an old bug keeps its own title, which the board holds.
	title := "Flaky test " + pos[0]
	if b, err := board.Load(cfg); err == nil {
		if it := b.Get(o.ShortID); it != nil {
			title = it.Title
		}
	}
	fmt.Fprintf(stdout, "%s %s (%d flakes)\n", o.ShortID, title, n)
	if o.Skipped {
		fmt.Fprintln(stderr, "written, not committed:", o.Reason)
		return exitSkipped
	}
	return exitOK
}
