package cli

import (
	"fmt"
	"io"

	"github.com/iyay/acta/internal/write"
)

// cmdPlan adds a fix round, the polish task or review notes to a plan. The
// items come from stdin, one per line, so an agent never reads the plan to
// add them.
func cmdPlan(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "add" {
		fmt.Fprintln(stderr, usageLine("plan", "add"))
		return exitBadInput
	}
	fs, root := flags("plan add", stderr)
	title := fs.String("title", "", "task title (fix-round only)")
	verify := fs.String("verify", "", "verify line (fix-round only)")
	pos, err := parseMixed(fs, args[1:])
	if err != nil || len(pos) != 2 {
		fmt.Fprintln(stderr, usageLine("plan", "add"))
		return exitBadInput
	}
	cfg, b, code := loadBoard(*root, stderr)
	if code != exitOK {
		return code
	}
	if msg := building(cfg, b.Get(pos[0])); msg != "" {
		fmt.Fprintln(stderr, msg)
		return exitBadInput
	}
	body, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	o, line, err := write.AddPlanSection(cfg, b, pos[0], pos[1], *title, *verify, body)
	if err != nil {
		return report(o, err, stdout, stderr)
	}
	fmt.Fprintln(stdout, line)
	if o.Skipped {
		fmt.Fprintln(stderr, "written, not committed:", o.Reason)
		return exitSkipped
	}
	return exitOK
}
