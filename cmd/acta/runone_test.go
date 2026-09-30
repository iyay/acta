package main

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// A stop sent to acta has to reach the command. Otherwise acta dies, the
// lock goes free, and the command keeps eating the machine with no lock.
func TestRunOneStopReachesTheCommand(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGTERM, syscall.SIGINT} {
		home := t.TempDir()
		cmd := exec.Command(actaBin, "run-one", "--", "sh", "-c", "echo $$; exec sleep 30")
		cmd.Env = append(os.Environ(), "HOME="+home, "XDG_CACHE_HOME="+filepath.Join(home, ".cache"))
		out, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		line, err := bufio.NewReader(out).ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		pid, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil {
			t.Fatal(err)
		}
		cmd.Process.Signal(sig)
		done := make(chan struct{})
		go func() { cmd.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			cmd.Process.Kill()
			t.Fatalf("%v: acta did not stop", sig)
		}
		if err := syscall.Kill(pid, 0); err == nil {
			syscall.Kill(pid, syscall.SIGKILL)
			t.Fatalf("%v: the command outlived acta", sig)
		}
	}
}
