// Command flakerun runs go test and runs each failed test again on its own,
// so a flaky test does not fail a full run. scripts/test --full calls it.
package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"github.com/iyay/acta/internal/flakerun"
)

func main() {
	os.Exit(flakerun.Run(os.Args[1:], run, os.Stdout))
}

// run starts go test -json with args and hands back all it printed. The
// error is set when go could not start or a test failed.
func run(args []string) (io.Reader, error) {
	cmd := exec.Command("go", append([]string{"test", "-json"}, args...)...)
	b, err := cmd.CombinedOutput()
	return bytes.NewReader(b), err
}
