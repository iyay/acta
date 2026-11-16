package plugincheck

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iyay/acta/internal/evalomp"
)

// The grader must reject a report that hides missing or late cleanup.
func TestGateWorktreeOwnerGraders(t *testing.T) {
	cases, err := evalomp.LoadCases(filepath.Join(pluginRoot(t), "evals"))
	if err != nil {
		t.Fatal(err)
	}
	var grader, order evalomp.Grader
	for _, c := range cases {
		if c.Name == "land-tidy-temp-worktree" {
			for _, g := range c.Graders {
				if g.Name == "owner-removal" {
					grader = g
				}
				if g.Name == "temp-removed" {
					order = g
				}
			}
		}
	}
	if grader.Type == "" {
		t.Fatal("missing owner-removal effects grader")
	}
	for _, tc := range []struct {
		name, audit string
		want        bool
	}{
		{"owner removed before movement", "owner_cleanup=1\n", true},
		{"wrong repository", "owner_cleanup=0\n", false},
		{"late or different removal", "", false},
		{"report only", "Removed the temporary worktree.\n", false},
		{"failure after success", "owner_cleanup=1\nowner_cleanup=0\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "owner-audit.txt"), []byte(tc.audit), 0600); err != nil {
				t.Fatal(err)
			}
			got := evalomp.Grade(grader, evalomp.Workspace{Dir: dir, Result: evalomp.Result{Reply: "Removed the temporary worktree."}}, nil)
			if got.Pass != tc.want {
				t.Fatalf("pass=%v want %v: %s", got.Pass, tc.want, got.Why)
			}
		})
	}
	for _, tc := range []struct {
		name, calls string
		want        bool
	}{
		{"ordinary removal", "log start\ngit worktree add --detach /gate tip exit=0 in repo\ngit worktree remove /gate exit=0 in repo\ngit merge --ff-only tip exit=0 in repo\n", true},
		{"force removal", "log start\ngit worktree add --detach /gate tip exit=0 in repo\ngit worktree remove --force /gate exit=0 in repo\ngit merge --ff-only tip exit=0 in repo\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "calls.log"), []byte(tc.calls), 0600); err != nil {
				t.Fatal(err)
			}
			if got := evalomp.Grade(order, evalomp.Workspace{Dir: dir}, nil); got.Pass != tc.want {
				t.Fatalf("removal-order pass=%v want %v: %s", got.Pass, tc.want, got.Why)
			}
		})
	}
}

// A successful removal in a different registry must not hide a live gate tree.
func TestGateWorktreeOwnerAudit(t *testing.T) {
	// The scaffold copies the acta it finds on PATH. CI has none, and a
	// local one may be old, so build this checkout's acta and put it first.
	actaBin := t.TempDir()
	build := exec.Command("go", "build", "-o", filepath.Join(actaBin, "acta"), "./cmd/acta")
	build.Dir = filepath.Dir(pluginRoot(t))
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build acta: %v\n%s", err, out)
	}
	path := "PATH=" + actaBin + string(os.PathListSeparator) + os.Getenv("PATH")
	for _, mode := range []string{"valid", "wrong repository", "late", "different worktree"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			dir, home, tmp := t.TempDir(), t.TempDir(), t.TempDir()
			env := append(os.Environ(), path, "HOME="+home, "TMPDIR="+tmp, "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
			run := func(cwd, name string, wantOK bool, args ...string) string {
				t.Helper()
				cmd := exec.CommandContext(ctx, name, args...)
				cmd.Dir, cmd.Env = cwd, env
				out, err := cmd.CombinedOutput()
				if (err == nil) != wantOK {
					t.Fatalf("%s %v: %v\n%s", name, args, err, out)
				}
				return strings.TrimSpace(string(out))
			}
			run(dir, "bash", true, filepath.Join(pluginRoot(t), "evals/land-tidy-temp-worktree/scaffold.sh"))
			git, acta := filepath.Join(dir, "bin/git"), filepath.Join(dir, "bin/acta")
			owner, branch, gate := filepath.Join(dir, "repo"), filepath.Join(dir, "repo-land"), filepath.Join(tmp, "gate")
			run(branch, acta, true, "tidy", "main", "land")
			tip := run(owner, git, true, "rev-parse", "refs/acta/tidy/land")
			run(owner, git, true, "worktree", "add", "--detach", gate, tip)
			switch mode {
			case "valid":
				run(owner, git, true, "worktree", "remove", gate)
			case "wrong repository":
				run(dir, git, false, "worktree", "remove", gate)
			case "different worktree":
				other := filepath.Join(tmp, "other")
				run(owner, "git", true, "worktree", "add", "--detach", other, tip)
				run(owner, git, true, "worktree", "remove", other)
			}
			run(owner, git, true, "merge", "--ff-only", "refs/acta/tidy/land")
			if mode == "late" {
				run(owner, git, true, "worktree", "remove", gate)
			}
			b, err := os.ReadFile(filepath.Join(dir, "owner-audit.txt"))
			if err != nil {
				t.Fatal(err)
			}
			want := "owner_cleanup=0\n"
			if mode == "valid" {
				want = "owner_cleanup=1\n"
			}
			if string(b) != want {
				t.Fatalf("audit %q want %q", b, want)
			}
			if _, err := os.Stat(gate); err == nil {
				run(owner, git, true, "worktree", "remove", gate)
			}
			run(owner, git, true, "worktree", "remove", branch)
		})
	}
}
