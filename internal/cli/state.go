package cli

import (
	"fmt"
	"io"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/write"
)

// cmdState writes one part of the State section of a plan, or prints the
// running plans. The body comes from stdin, so an agent pipes in the lines it
// wants a session to find later.
func cmdState(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	switch {
	case len(args) > 0 && args[0] == "set":
		return cmdStateSet(args[1:], stdin, stdout, stderr)
	case len(args) > 1:
		fmt.Fprintln(stderr, "usage: acta state [plan]")
		return exitBadInput
	}
	cfg, code := loadConfig("", stderr)
	if code != exitOK {
		return code
	}
	if len(args) == 0 {
		for _, r := range board.Running(cfg) {
			fmt.Fprintln(stdout, r.Short()[0])
		}
		return exitOK
	}
	// A plan named on the command line gets its view whether or not work is
	// still going on: land reads a plan after the last box is ticked.
	for _, r := range board.PlanStates(cfg) {
		if namesPlan(r.Plan, args[0]) {
			fmt.Fprint(stdout, r.Full())
			return exitOK
		}
	}
	fmt.Fprintf(stderr, "no running plan %s; run: acta state\n", args[0])
	return exitBadInput
}

// namesPlan says whether id is this plan in any of the forms the board takes:
// the file path, the number id or the hash.
func namesPlan(it *board.Item, id string) bool {
	want := board.Canon(id)
	for _, k := range []string{it.ID, it.ShortID, it.Hash} {
		if k != "" && board.Canon(k) == want {
			return true
		}
	}
	return false
}

func cmdStateSet(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs, root := flags("state set", stderr)
	clear := fs.Bool("clear", false, "empty the State part instead of writing lines")
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) != 2 {
		fmt.Fprintln(stderr, "usage: acta state set <plan> next|findings|rulings [--clear] < text.md")
		return exitBadInput
	}
	cfg, b, code := loadBoard(*root, stderr)
	if code != exitOK {
		return code
	}
	body, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	o, err := write.SetState(cfg, b, pos[0], pos[1], body, *clear)
	return report(o, err, stdout, stderr)
}
