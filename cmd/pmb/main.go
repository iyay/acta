// Command pmb is the old name of acta: it warns once, then runs acta.
package main

import (
	"fmt"
	"os"

	"github.com/iyay/acta/internal/cli"
)

func main() {
	// Warn before anything else, so scripts see the rename even on failure.
	fmt.Fprintln(os.Stderr, "pmb is now acta; this name goes away in a later version")
	st, _ := os.Stdin.Stat()
	tty := st != nil && st.Mode()&os.ModeCharDevice != 0
	os.Exit(cli.Run(os.Args[1:], os.Stdin, tty, os.Stdout, os.Stderr))
}
