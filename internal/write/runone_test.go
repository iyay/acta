package write

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestHelperHoldRunOne is not a real test. The tests below run it in a
// child process to hold the run-one lock and then get killed.
func TestHelperHoldRunOne(t *testing.T) {
	if os.Getenv("ACTA_HOLD_RUN_ONE") == "" {
		t.Skip("helper for the run-one tests")
	}
	lockRoot = os.Getenv("ACTA_LOCK_ROOT")
	if _, err := HoldRunOne(func() {}); err != nil {
		t.Fatal(err)
	}
	os.Stdout.WriteString("held\n")
	// Wait to be killed. A bare select{} trips Go's deadlock check.
	time.Sleep(time.Minute)
}

// holdInChild takes the run-one lock in a child process and returns once
// the child holds it. The caller kills the child to let go.
func holdInChild(t *testing.T) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperHoldRunOne$")
	cmd.Env = append(os.Environ(), "ACTA_HOLD_RUN_ONE=1", "ACTA_LOCK_ROOT="+lockRoot)
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cmd.Process.Kill(); cmd.Wait() })
	buf := make([]byte, 5)
	if _, err := io.ReadFull(out, buf); err != nil || string(buf) != "held\n" {
		t.Fatalf("the child did not take the lock: %q %v", buf, err)
	}
	return cmd
}

func TestHoldRunOneWaitsForTheHolder(t *testing.T) {
	useLockBase(t)
	child := holdInChild(t)
	waited := 0
	got := make(chan error, 1)
	go func() {
		release, err := HoldRunOne(func() { waited++ })
		if err == nil {
			release()
		}
		got <- err
	}()
	select {
	case err := <-got:
		t.Fatalf("took the lock while the child held it: %v", err)
	case <-time.After(300 * time.Millisecond):
	}
	child.Process.Kill()
	child.Wait()
	select {
	case err := <-got:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("still waiting after the holder died")
	}
	if waited != 1 {
		t.Fatalf("waiting ran %d times, want 1", waited)
	}
}

func TestHoldRunOneFreeLockDoesNotWait(t *testing.T) {
	useLockBase(t)
	waited := false
	release, err := HoldRunOne(func() { waited = true })
	if err != nil {
		t.Fatal(err)
	}
	release()
	if waited {
		t.Fatal("said it was waiting on a free lock")
	}
	// A released lock can be taken again at once.
	release, err = HoldRunOne(func() { t.Fatal("waited on a released lock") })
	if err != nil {
		t.Fatal(err)
	}
	release()
}

func TestHoldRunOneRefusesAnOpenFolder(t *testing.T) {
	dir := useLockBase(t)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := HoldRunOne(func() {}); err == nil {
		t.Fatal("took a lock in a folder open to other users")
	}
	if _, err := os.Stat(filepath.Join(dir, "run-one.lock")); !os.IsNotExist(err) {
		t.Fatalf("made a lock file in a folder it refused: %v", err)
	}
}
