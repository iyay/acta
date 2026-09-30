package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// callScript runs scripts/test with a fake go that prints its folder and its
// args, so the test sees the exact go call without running any tests.
func callScript(t *testing.T, args ...string) string {
	t.Helper()
	bin := t.TempDir()
	fake := "#!/bin/sh\necho \"$(pwd) $*\"\n"
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(fake), 0o755); err != nil {
		t.Fatal(err)
	}
	script, err := filepath.Abs("test")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(script, args...)
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "PATH="+bin+":/usr/bin:/bin")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
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
