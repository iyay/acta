package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// runOne calls acta run-one with its lock in a temp cache folder, so a test
// never waits on a real full run of this machine. The cache folder comes
// from HOME on macOS and from XDG_CACHE_HOME on Linux, so both move.
func runOne(t *testing.T, stdin string, args ...string) (int, string, string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	var out, errb bytes.Buffer
	code := Run(append([]string{"run-one"}, args...), strings.NewReader(stdin), false, &out, &errb)
	return code, out.String(), errb.String()
}

func TestRunOnePassesTheExitCodeBack(t *testing.T) {
	for _, want := range []int{0, 1, 7} {
		code, _, _ := runOne(t, "", "--", "sh", "-c", "exit "+strconv.Itoa(want))
		if code != want {
			t.Errorf("exit %d came back as %d", want, code)
		}
	}
}

func TestRunOnePassesOutputAndInputThrough(t *testing.T) {
	code, out, errb := runOne(t, "from stdin\n", "--", "sh", "-c", "cat; echo to-err >&2")
	if code != 0 || out != "from stdin\n" || errb != "to-err\n" {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errb)
	}
}

func TestRunOneRefusesBadUsage(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "ran")
	for _, args := range [][]string{
		nil,
		{"--"},
		{"touch", marker},
		{"-x", "--", "touch", marker},
	} {
		code, _, errb := runOne(t, "", args...)
		if code != exitBadInput || !strings.Contains(errb, "usage: acta run-one -- <command> [args...]") {
			t.Errorf("%q: code %d, stderr %q", args, code, errb)
		}
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("a bad call still ran the command")
	}
}

func TestRunOneCommandThatCannotStart(t *testing.T) {
	code, _, errb := runOne(t, "", "--", filepath.Join(t.TempDir(), "no-such-command"))
	if code != exitOther || errb == "" {
		t.Fatalf("code %d, stderr %q", code, errb)
	}
}
