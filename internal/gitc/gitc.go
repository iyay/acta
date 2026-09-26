// Package gitc commits one file after a tool changes it. It runs the git
// binary so repo config and hooks behave the way they do for the user.
package gitc

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// Worktree is one checkout of the repo: the main one or a linked one.
type Worktree struct {
	Path   string
	Branch string // short branch name, or "(detached)"
}

// Worktrees lists the repo's checkouts, the main one first, leaving out bare,
// prunable and missing ones.
func Worktrees(repo string) ([]Worktree, error) {
	out, err := run(repo, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseWorktrees(out), nil
}

func parseWorktrees(out string) []Worktree {
	var res []Worktree
	for _, block := range strings.Split(strings.TrimSpace(out), "\n\n") {
		var w Worktree
		skip := false
		for _, ln := range strings.Split(block, "\n") {
			switch {
			case strings.HasPrefix(ln, "worktree "):
				w.Path = strings.TrimPrefix(ln, "worktree ")
			case strings.HasPrefix(ln, "branch "):
				w.Branch = strings.TrimPrefix(strings.TrimPrefix(ln, "branch "), "refs/heads/")
			case ln == "detached":
				w.Branch = "(detached)"
			case ln == "bare" || strings.HasPrefix(ln, "prunable"):
				skip = true
			}
		}
		if w.Path == "" || skip {
			continue
		}
		if _, err := os.Stat(w.Path); err != nil {
			continue
		}
		res = append(res, w)
	}
	return res
}

// UnmergedBranches lists local branches not merged into HEAD.
func UnmergedBranches(repo string) ([]string, error) {
	out, err := run(repo, "for-each-ref", "--no-merged", "HEAD", "--format=%(refname:short)", "refs/heads")
	if err != nil {
		return nil, err
	}
	var res []string
	for _, ln := range strings.Split(strings.TrimSpace(out), "\n") {
		if ln = strings.TrimSpace(ln); ln != "" {
			res = append(res, ln)
		}
	}
	return res, nil
}

// BranchFiles reads every file under dir (repo-relative, slash separated) at
// the tip of branch, without a checkout. A branch with no such folder gives an
// empty map.
func BranchFiles(repo, branch, dir string) (map[string][]byte, error) {
	ref := "refs/heads/" + branch
	list, err := run(repo, "ls-tree", "-r", "-z", "--full-tree", ref, "--", dir)
	if err != nil {
		return nil, err
	}
	var names, objects []string
	for _, entry := range strings.Split(list, "\x00") {
		meta, name, ok := strings.Cut(entry, "\t")
		if !ok {
			continue
		}
		f := strings.Fields(meta) // mode, type, object
		if len(f) == 3 && f[1] == "blob" {
			names = append(names, name)
			objects = append(objects, f[2])
		}
	}
	files := map[string][]byte{}
	if len(objects) == 0 {
		return files, nil
	}
	cmd := exec.Command("git", "-C", repo, "cat-file", "--batch")
	cmd.Stdin = strings.NewReader(strings.Join(objects, "\n") + "\n")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git cat-file: %w", err)
	}
	// Each answer is "<object> blob <size>\n<content>\n".
	for i := range objects {
		head, rest, ok := bytes.Cut(out, []byte("\n"))
		if !ok {
			return nil, fmt.Errorf("git cat-file: short output")
		}
		parts := strings.Fields(string(head))
		if len(parts) != 3 {
			return nil, fmt.Errorf("git cat-file: bad header %q", head)
		}
		size, err := strconv.Atoi(parts[2])
		if err != nil || size > len(rest) {
			return nil, fmt.Errorf("git cat-file: bad size %q", head)
		}
		files[names[i]] = rest[:size]
		out = rest[size:]
		if len(out) > 0 && out[0] == '\n' {
			out = out[1:]
		}
	}
	return files, nil
}

// CommonDir is the .git folder shared by every worktree of the repo.
func CommonDir(repo string) (string, error) {
	out, err := run(repo, "rev-parse", "--path-format=absolute", "--git-common-dir")
	return strings.TrimSpace(out), err
}
