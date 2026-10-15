//go:build notui

package cli

import (
	"fmt"
	"io"

	"github.com/iyay/acta/internal/config"
)

// Variant names this build. This one leaves the TUI board out, so the
// binary is about 7M smaller.
const Variant = "cli"

// runTUI has no board to open in this build. It shows the help and says how
// to get the board, so bare acta still answers with something useful.
var runTUI = func(cfg config.Config, stdout, stderr io.Writer) int {
	fmt.Fprint(stdout, renderHelp())
	fmt.Fprintln(stdout, "This acta has no TUI board. To get it, run the installer again with ACTA_TUI=1.")
	return exitOK
}
