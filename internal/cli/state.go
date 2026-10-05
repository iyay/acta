package cli

import (
	"fmt"
	"io"

	"github.com/iyay/acta/internal/write"
)

// cmdState writes one part of the State section of a plan, or prints the
// running plans. The body comes from stdin, so an agent pipes in the lines it
// wants a session to find later.
func cmdState(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "set" {
		return cmdStateSet(args[1:], stdin, stdout, stderr)
	}
	// The bare view is still being built, so it says so rather than guessing.
	fmt.Fprintln(stderr, "acta state <plan> is not built yet; run: acta state set <plan> next|findings|rulings < text.md")
	return exitBadInput
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
