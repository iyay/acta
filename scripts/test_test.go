package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runScript runs scripts/test with a fake go in place of the real one, so the
// test sees what the go call would do without running any tests.
func runScript(t *testing.T, fakeGo string, args ...string) string {
	t.Helper()
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(fakeGo), 0o755); err != nil {
		t.Fatal(err)
	}
	script, err := filepath.Abs("test")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(script, args...)
	cmd.Dir = t.TempDir()
	// Git config values and the PATH this machine already has are dropped, so
	// the child sees only the fake go. A repeated key keeps its last value,
	// so the fake PATH has to go last.
	env := make([]string, 0, len(os.Environ())+1)
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "GIT_CONFIG_") && !strings.HasPrefix(e, "PATH=") {
			env = append(env, e)
		}
	}
	env = append(env, "PATH="+bin+":/usr/bin:/bin")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// callScript runs scripts/test with a fake go that prints its folder and its
// args, so the test sees the exact go call without running any tests.
func callScript(t *testing.T, args ...string) string {
	t.Helper()
	return runScript(t, "#!/bin/sh\necho \"$(pwd) $*\"\n", args...)
}

// callScriptGitEnv runs scripts/test with a fake go that prints the three git
// config values, so the test sees the environment the tests would get.
func callScriptGitEnv(t *testing.T, args ...string) string {
	t.Helper()
	return runScript(t, "#!/bin/sh\necho \"$GIT_CONFIG_COUNT $GIT_CONFIG_KEY_0 $GIT_CONFIG_VALUE_0\"\n", args...)
}

// Git starts an automatic maintenance run after every commit, and that run
// holds a lock inside .git/objects while a test tries to delete its temp
// folder. Turning it off here keeps both paths, short and --full, free of it.
func TestScriptTurnsOffGitAutoMaintenance(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{
		nil,
		{"./internal/tui/"},
		{"--full"},
		{"--full", "./internal/tui/"},
	} {
		if got := callScriptGitEnv(t, args...); got != "1 maintenance.auto false" {
			t.Errorf("%q: got %q, want %q", args, got, "1 maintenance.auto false")
		}
	}
}

func TestScriptMapsEachCallToOneGoCall(t *testing.T) {
	t.Parallel()

	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		args []string
		want string
	}{
		{nil, "test -short ./..."},
		{[]string{"./internal/tui/"}, "test -short ./internal/tui/"},
		{[]string{"-count=1"}, "test -short -count=1 ./..."},
		{[]string{"-run", "TestX", "./internal/tui/"}, "test -short -run TestX ./internal/tui/"},
		{[]string{"--full"}, "run ./cmd/acta run-one -- go test ./..."},
		{[]string{"--full", "-count=1"}, "run ./cmd/acta run-one -- go test -count=1 ./..."},
		{[]string{"--full", "./internal/cli/"}, "run ./cmd/acta run-one -- go test ./internal/cli/"},
	} {
		if got := callScript(t, tc.args...); got != root+" "+tc.want {
			t.Errorf("%q: got %q, want %q", tc.args, got, root+" "+tc.want)
		}
	}
}
