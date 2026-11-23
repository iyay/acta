package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/gitc"
	"github.com/iyay/acta/internal/telemetry"
	"github.com/iyay/acta/internal/write"
)

// cmdLandNotes writes the notes that close a plan, on the branch it just
// merged into, and commits them as one commit. It is the full land mode's
// twin of acta tidy --plan: both run write.LandNotes, so both write the same.
func cmdLandNotes(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("land-notes", stderr)
	pos, err := parseMixed(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitBadInput
	}
	if len(pos) != 1 {
		fmt.Fprintln(stderr, usageLine("land-notes"))
		return exitBadInput
	}
	// The hash goes into the commit subject, so only the full one will do.
	hash := strings.TrimPrefix(pos[0], "PLN-")
	if !board.IsHash(hash) {
		fmt.Fprintf(stderr, "%q is not a 7 character plan hash\n", pos[0])
		return exitBadInput
	}
	cfg, code := loadConfig(*root, stderr)
	if code != exitOK {
		return code
	}
	wrote, err := write.LandNotes(cfg, "PLN-"+hash)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	if len(wrote) > 0 {
		res := gitc.CommitPathsIn(cfg.RepoRoot, cfg.Root, wrote, "chore(acta): land of PLN-"+hash)
		if !res.Committed {
			fmt.Fprintln(stderr, "written, not committed:", res.Reason)
			return exitOther
		}
		for _, path := range wrote {
			fmt.Fprintln(stdout, path)
		}
	}
	// Full mode runs this once per land, so a run that works is one landed
	// plan, even when the plan had no notes to write.
	_ = telemetry.Default().Add("workflow", "plan_landed")
	return exitOK
}
