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
	"time"
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

// CommitPaths commits only the listed paths in one commit. A dirty file
// outside the list stays out, so one run of a tool is one commit.
func CommitPaths(repo string, paths []string, msg string) Result {
	if _, err := run(repo, "rev-parse", "--git-dir"); err != nil {
		return Result{Reason: "not a git repo"}
	}
	if reason := busy(repo); reason != "" {
		return Result{Reason: reason}
	}
	args := append([]string{"add", "--"}, paths...)
	if _, err := run(repo, args...); err != nil {
		return Result{Reason: err.Error()}
	}
	args = append([]string{"commit", "-m", msg, "--"}, paths...)
	if _, err := run(repo, args...); err != nil {
		return Result{Reason: "commit failed: " + err.Error()}
	}
	return Result{Committed: true}
}

// FirstSeen is where path first shows up along the first-parent history of
// HEAD: its place in git rev-list --first-parent --reverse HEAD, counting
// from zero. A file the branch only got through a merge is seen at that
// merge, and a file never committed comes last. Commit order, not clock
// time, so two commits in the same second cannot swap the order.
func FirstSeen(repo, path string) (int, error) {
	out, err := run(repo, "rev-list", "--first-parent", "--reverse", "HEAD")
	if err != nil {
		return 0, err
	}
	var commits []string
	for _, h := range strings.Split(strings.TrimSpace(out), "\n") {
		if h != "" {
			commits = append(commits, h)
		}
	}
	touched, err := run(repo, "log", "--first-parent", "--format=%H", "--reverse", "--", path)
	if err != nil {
		return 0, err
	}
	first, _, _ := strings.Cut(strings.TrimSpace(touched), "\n")
	for i, h := range commits {
		if h == first {
			return i, nil
		}
	}
	return len(commits), nil
}

// LastChange is the commit time of the newest commit on the way back from HEAD
// that touched one of the paths, or the zero time when none did. Each path is a
// plain name from the repo root, and a folder ends in a slash. The paths are not
// globs and not git path magic, since the wiki lists names and matches them by
// prefix. git stops at the first commit it finds, which is the newest unless a
// clock was set wrong.
func LastChange(repo string, paths []string) (time.Time, error) {
	// With no paths, git would answer with the newest commit of the whole repo.
	if len(paths) == 0 {
		return time.Time{}, nil
	}
	args := append([]string{"--literal-pathspecs", "log", "-1", "--format=%cI", "--"}, paths...)
	out, err := run(repo, args...)
	if err != nil {
		return time.Time{}, err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, out)
}

// Changed lists the files that the commits in rng changed, from the repo root.
// rng is a range like main..HEAD. A move shows as a file that left and a file
// that came, so the folder a file moved out of counts as changed too. Names come
// back as they are, with no quotes around the odd ones.
func Changed(repo, rng string) ([]string, error) {
	// A dash would make git read the range as an option, and some options write files.
	if strings.HasPrefix(rng, "-") {
		return nil, fmt.Errorf("bad range %q: it starts with a dash", rng)
	}
	// git diff a..b sets the two tips side by side. If a has moved on since b
	// split from it, a's own changes would show up as b's. Three dots start at
	// the point where they split, which is what a commit range means.
	if left, right, ok := strings.Cut(rng, ".."); ok && !strings.HasPrefix(right, ".") {
		rng = left + "..." + right
	}
	// The last -- says rng is a revision, never a file name that happens to match.
	out, err := run(repo, "diff", "--no-renames", "--name-only", "-z", rng, "--")
	if err != nil {
		return nil, err
	}
	var files []string
	for _, f := range strings.Split(out, "\x00") {
		if f != "" {
			files = append(files, f)
		}
	}
	return files, nil
}

// Authors gives, for each path git knows, the name of the person whose commit
// first added it. It asks git once for all the paths, since one call per file
// made every board load take seconds. The paths sit right inside repo. A path
// git never saw has no entry, and a folder outside any checkout gives an
// empty map without running git.
func Authors(repo string, paths []string) map[string]string {
	found := map[string]string{}
	// With no paths, git would log the whole repo. Stop before that.
	if len(paths) == 0 || !inRepo(repo) {
		return found
	}
	// No renames: a file moved inside the folder counts as added where it is
	// now, the same answer a one-file log gives. quotePath off keeps names
	// with non-ASCII letters as they are. Each commit line starts with a NUL
	// byte, so an author line never reads as a file name.
	args := append([]string{"-c", "core.quotePath=false", "log", "--no-renames", "--diff-filter=A",
		"--relative", "--name-only", "--format=%x00%an", "--"}, paths...)
	out, err := run(repo, args...)
	if err != nil {
		return found
	}
	name := ""
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "\x00"):
			name = strings.TrimSpace(line[1:])
		case line != "" && name != "":
			// Git lists the newest commit first, so the last add seen is the first one.
			found[filepath.Join(repo, line)] = name
		}
	}
	return found
}

// UserName gives the name this checkout commits under, or "" when git has no
// name to give. A folder outside any checkout answers at once, without running
// git.
func UserName(repo string) string {
	if !inRepo(repo) {
		return ""
	}
	out, err := run(repo, "config", "user.name")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// inRepo says whether dir sits inside a checkout, which the .git entry at the
// top of it tells: a folder or a file in a linked worktree, both work.
func inRepo(dir string) bool {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return false
		}
		dir = parent
	}
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
	out, err := run(repo, "for-each-ref", "--no-merged", "HEAD", "--format=%(refname:lstrip=2)", "refs/heads")
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
