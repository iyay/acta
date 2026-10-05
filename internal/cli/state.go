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
	runs := board.Running(cfg)
	if len(args) == 0 {
		for _, r := range runs {
			fmt.Fprintln(stdout, r.Short()[0])
		}
		return exitOK
	}
	for _, r := range runs {
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
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: acta state set <plan> next|findings|rulings < text.md")
		return exitBadInput
	}
	cfg, b, code := loadBoard("", stderr)
	if code != exitOK {
		return code
	}
	body, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	o, err := write.SetState(cfg, b, args[0], args[1], body)
	return report(o, err, stdout, stderr)
}
