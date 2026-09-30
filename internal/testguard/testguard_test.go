package testguard

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	Watch()
	os.Exit(m.Run())
}

// childReady says the child is up and Watch is already watching, so the test
// knows the parent can be killed now.
const childReady = "child-ready"

// TestChild only burns time when a test starts it as an orphan, so a plain
// run of this package stays fast.
func TestChild(t *testing.T) {
	if os.Getenv("TESTGUARD_CHILD") != "1" {
		t.Skip("only does work when a test starts it as an orphan")
	}
	fmt.Fprintln(os.Stderr, childReady)
	time.Sleep(30 * time.Second)
}

func TestWatchStopsOrphan(t *testing.T) {
	if os.Getenv("TESTGUARD_CHILD") == "1" {
		t.Skip("the child run only runs TestChild")
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatalf("cannot find this test binary: %v", err)
	}
	stderrFile := filepath.Join(t.TempDir(), "child-stderr")

	// sh holds the child so it starts with a living parent, the way a test
	// binary starts under go test. When sh dies the child is an orphan and
	// its parent changes, which is the case Watch has to catch.
	script := fmt.Sprintf("TESTGUARD_CHILD=1 %q -test.run TestChild 2>%q & echo $!; wait", binary, stderrFile)
	sh := exec.Command("sh", "-c", script)
	stdout, err := sh.StdoutPipe()
	if err != nil {
		t.Fatalf("cannot pipe sh output: %v", err)
	}
	if err := sh.Start(); err != nil {
		t.Fatalf("cannot start sh: %v", err)
	}
	defer sh.Wait()

	pid, err := strconv.Atoi(strings.TrimSpace(readLine(t, stdout)))
	if err != nil {
		t.Fatalf("cannot read the child pid: %v", err)
	}

	myParent := os.Getppid()
	// Killing sh before the child is up would be a race: the child could save
	// the new parent and never see it change.
	waitForLine(t, stderrFile, childReady)

	if err := sh.Process.Kill(); err != nil {
		t.Fatalf("cannot kill sh: %v", err)
	}

	// Case one: the orphaned test binary must stop on its own within 5 seconds.
	deadline := time.Now().Add(5 * time.Second)
	for running(pid) && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if running(pid) {
		syscall.Kill(pid, syscall.SIGKILL)
		t.Fatalf("the orphaned test binary was still alive after 5 seconds")
	}
	if said, _ := os.ReadFile(stderrFile); !strings.Contains(string(said), stopLine) {
		t.Errorf("the orphaned test binary did not print %q, got %q", stopLine, string(said))
	}

	// Case two: this test is a parent whose own parent stays alive, so Watch
	// must leave it alone and the test keeps running to the end.
	if got := os.Getppid(); got != myParent {
		t.Fatalf("this test process was reparented from %d to %d", myParent, got)
	}
	if err := syscall.Kill(os.Getpid(), 0); err != nil {
		t.Fatalf("this test process is gone: %v", err)
	}
}

func readLine(t *testing.T, stdout interface{ Read([]byte) (int, error) }) string {
	t.Helper()
	line, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil && line == "" {
		t.Fatalf("cannot read the child pid: %v", err)
	}
	return line
}

func waitForLine(t *testing.T, path, want string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if said, _ := os.ReadFile(path); strings.Contains(string(said), want) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("the child test binary never said it was up")
}

func running(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}
