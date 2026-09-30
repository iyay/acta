package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/iyay/acta/internal/write"
)

// cmdRunOne runs one command while it holds the machine-wide run-one lock,
// so two full test runs from two worktrees never fight over the CPU. The
// command's own output and exit code come back unchanged.
func cmdRunOne(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) < 2 || args[0] != "--" {
		fmt.Fprintln(stderr, "usage: acta run-one -- <command> [args...]")
		return exitBadInput
	}
	release, err := write.HoldRunOne(func() {
		fmt.Fprintln(stderr, "waiting for another run-one to finish")
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	defer release()
	cmd := exec.Command(args[1], args[2:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdin, stdout, stderr
	// Catch a stop only now, after the wait, so a stop while waiting still
	// ends acta at once. From here on the stop goes to the command, or acta
	// would die and leave the command running with no lock.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	go func() {
		for s := range sig {
			cmd.Process.Signal(s)
		}
	}()
	err = cmd.Wait()
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() >= 0 {
		return exit.ExitCode()
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	return exitOK
}
