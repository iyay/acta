// Package gitc commits one file after a tool changes it. It runs the git
// binary so repo config and hooks behave the way they do for the user.
package gitc

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Result says whether the commit happened, and why not when it did not.
type Result struct {
	Committed bool
	Reason    string
}

// IsDirty says whether path has changes git has not committed.
func IsDirty(repo, path string) (bool, error) {
	out, err := run(repo, "status", "--porcelain", "--", path)
	return strings.TrimSpace(out) != "", err
}

// Commit commits only path. wasDirty is whether path had changes before the
// tool wrote to it; those edits are not ours to commit.
func Commit(repo, path, msg string, wasDirty bool) Result {
	if _, err := run(repo, "rev-parse", "--git-dir"); err != nil {
		return Result{Reason: "not a git repo"}
	}
	if wasDirty {
		return Result{Reason: "file had other uncommitted changes"}
	}
	if reason := busy(repo); reason != "" {
		return Result{Reason: reason}
	}
	if _, err := run(repo, "add", "--", path); err != nil {
		return Result{Reason: err.Error()}
	}
	if _, err := run(repo, "commit", "--only", "-m", msg, "--", path); err != nil {
		return Result{Reason: "commit failed: " + err.Error()}
	}
	return Result{Committed: true}
}

// busy says why the repo cannot take a commit now, or "" when it can.
func busy(repo string) string {
	states := []struct{ file, what string }{
		{"MERGE_HEAD", "merge"},
		{"CHERRY_PICK_HEAD", "cherry-pick"},
		{"REVERT_HEAD", "revert"},
		{"rebase-merge", "rebase"},
		{"rebase-apply", "rebase"},
	}
	for _, s := range states {
		p, err := run(repo, "rev-parse", "--git-path", s.file)
		if err != nil {
			continue
		}
		p = strings.TrimSpace(p)
		if !filepath.IsAbs(p) {
			p = filepath.Join(repo, p)
		}
		if _, err := os.Stat(p); err == nil {
			return "repo is in the middle of a " + s.what
		}
	}
	if _, err := run(repo, "symbolic-ref", "-q", "HEAD"); err != nil {
		return "HEAD is detached"
	}
	return ""
}

func run(repo string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return out.String(), nil
}
