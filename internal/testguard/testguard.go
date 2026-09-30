// Package testguard stops a test binary whose parent process is gone, so a
// test run nobody waits for any more does not keep burning the CPU.
package testguard

import (
	"fmt"
	"os"
	"time"
)

// stopLine says why the binary gave up before it leaves.
const stopLine = "testguard: parent process is gone, stopping"

// Watch leaves the process when its parent is gone. It notices a change of
// parent, not parent number 1, because a Linux subreaper also reparents a
// living process to number 1 and that one must keep running.
func Watch() {
	startParent := os.Getppid()
	go func() {
		tick := time.NewTicker(time.Second)
		defer tick.Stop()
		for range tick.C {
			if os.Getppid() == startParent {
				continue
			}
			fmt.Fprintln(os.Stderr, stopLine)
			os.Exit(1)
		}
	}()
}
