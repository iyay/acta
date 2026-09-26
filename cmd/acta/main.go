// Command acta shows and changes the planning files of the repo it runs in.
package main

import (
	"os"

	"github.com/iyay/acta/internal/cli"
)

func main() {
	st, _ := os.Stdin.Stat()
	tty := st != nil && st.Mode()&os.ModeCharDevice != 0
	os.Exit(cli.Run(os.Args[1:], os.Stdin, tty, os.Stdout, os.Stderr))
}
