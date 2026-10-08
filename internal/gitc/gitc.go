// Package gitc commits one file after a tool changes it. It runs the git
// binary so repo config and hooks behave the way they do for the user.
package gitc

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Result says whether the commit happened, and why not when it did not.
type Result struct {
	Committed bool
	Folded    bool // the change went into the last commit instead of a new one
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

// CommitOrFold is Commit, except that it amends HEAD instead of adding a new
// commit when HEAD is this tool's own earlier planning commit of the same file
// and nothing else can hold that commit. A run of repeat edits then leaves one
// commit. The amended commit keeps HEAD's subject and adds the new subject as
// a body line, so no subject is lost. Any doubt, or a failed amend, ends in
// the plain Commit.
func CommitOrFold(repo, path, msg string, wasDirty bool) Result {
	if _, err := run(repo, "rev-parse", "--git-dir"); err != nil {
		return Result{Reason: "not a git repo"}
	}
	if wasDirty {
		return Result{Reason: "file had other uncommitted changes"}
	}
	if reason := busy(repo); reason != "" {
		return Result{Reason: reason}
	}
	if canFold(repo, path) {
		if _, err := run(repo, "add", "--", path); err == nil {
			if args, ok := foldArgs(repo, path, msg); ok {
				if _, err := run(repo, args...); err == nil {
					return Result{Committed: true, Folded: true}
				}
			}
		}
	}
	return Commit(repo, path, msg, wasDirty)
}

// trailerLine matches a line like "Co-Authored-By: X". A paragraph made only
// of such lines is a trailer block, the way git reads one.
var trailerLine = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]*:\s+\S`)

// foldArgs builds the git amend call for a fold. HEAD's message stays exactly
// as written. msg is added as one more line when HEAD does not say it yet: it
// goes before the trailer block, so trailers stay trailers, or at the end when
// there is none. It says false when HEAD's message cannot be read.
func foldArgs(repo, path, msg string) ([]string, bool) {
	out, err := run(repo, "log", "-1", "--format=%B")
	if err != nil {
		return nil, false
	}
	lines := strings.Split(strings.TrimRight(out, " \t\r\n"), "\n")
	msg = strings.TrimSpace(msg)
	said := false
	for _, ln := range lines {
		if strings.TrimSpace(ln) == msg {
			said = true
		}
	}
	if !said {
		lines = insertFoldLine(lines, msg)
	}
	args := []string{"commit", "--amend", "--only", "-m", strings.Join(lines, "\n")}
	return append(args, "--", path), true
}

// insertFoldLine puts msg into the message lines: after the last body line
// before a trailer block, else at the end. The subject is never moved.
func insertFoldLine(lines []string, msg string) []string {
	at := len(lines)
	if start := trailerStart(lines); start > 0 {
		at = start
		// Step back over the blank lines, so msg sits with the body text.
		for at > 1 && strings.TrimSpace(lines[at-1]) == "" {
			at--
		}
	}
	add := []string{msg}
	if at == 1 {
		// Right under the subject: keep one blank line between them.
		add = []string{"", msg}
	}
	return slices.Concat(lines[:at], add, lines[at:])
}

// trailerStart is the index of the first line of the last paragraph when
// every line of it is a trailer and it is not the subject. Else it is -1.
func trailerStart(lines []string) int {
	end := len(lines)
	start := end
	for start > 0 && strings.TrimSpace(lines[start-1]) != "" {
		start--
	}
	if start == 0 || start == end {
		return -1
	}
	for _, ln := range lines[start:end] {
		if !trailerLine.MatchString(ln) {
			return -1
		}
	}
	return start
}

// canFold says whether HEAD may be rewritten with a new edit of path. Each
// rule keeps a rewrite from hurting someone else, and any error means no.
func canFold(repo, path string) bool {
	// A file with no change would only rewrite HEAD. Commit says no to it.
	if dirty, err := IsDirty(repo, path); err != nil || !dirty {
		return false
	}
	// A tag marks a release, so its commit must never be rewritten.
	if tags, err := run(repo, "tag", "--points-at", "HEAD"); err != nil || strings.TrimSpace(tags) != "" {
		return false
	}
	// A planning subject (it starts with chore( ), one parent and the same
	// author as this user, so HEAD is our own planning commit and not a merge,
	// a feature commit or a colleague's work.
	out, err := run(repo, "log", "-1", "--format=%s%x00%an%x00%P")
	if err != nil {
		return false
	}
	f := strings.Split(strings.TrimRight(out, "\n"), "\x00")
	if len(f) != 3 || !strings.HasPrefix(f[0], "chore(") || len(strings.Fields(f[2])) != 1 {
		return false
	}
	if me := UserName(repo); me == "" || f[1] != me {
		return false
	}
	// HEAD must hold only this file. Asking with and without the path gives
	// the same list when nothing else is in it.
	all, err1 := run(repo, "diff-tree", "--no-commit-id", "--name-only", "-r", "-z", "HEAD")
	one, err2 := run(repo, "diff-tree", "--no-commit-id", "--name-only", "-r", "-z", "HEAD", "--", path)
	if err1 != nil || err2 != nil || one == "" || all != one {
		return false
	}
	// A commit that was pushed, or can be fetched from a remote, is shared.
	if out, err := run(repo, "branch", "-r", "--contains", "HEAD"); err != nil || strings.TrimSpace(out) != "" {
		return false
	}
	// HEAD must sit on a branch, and no other branch may hold it.
	cur, err := run(repo, "symbolic-ref", "--short", "-q", "HEAD")
	if err != nil {
		return false
	}
	branches, err := run(repo, "for-each-ref", "--contains", "HEAD", "--format=%(refname:short)", "refs/heads")
	if err != nil {
		return false
	}
	for _, b := range strings.Fields(branches) {
		if b != strings.TrimSpace(cur) {
			return false
		}
	}
	return !otherCheckoutHolds(repo)
}

// otherCheckoutHolds says whether another worktree stands on HEAD or on a
// commit built on it, or whether that cannot be told.
func otherCheckoutHolds(repo string) bool {
	top, err := run(repo, "rev-parse", "--show-toplevel")
	if err != nil {
		return true
	}
	head, err := run(repo, "rev-parse", "HEAD")
	if err != nil {
		return true
	}
	trees, err := Worktrees(repo)
	if err != nil {
		return true
	}
	for _, w := range trees {
		if w.Path == strings.TrimSpace(top) {
			continue
		}
		theirs, err := run(w.Path, "rev-parse", "HEAD")
		if err != nil {
			return true
		}
		held, err := IsAncestor(repo, strings.TrimSpace(head), strings.TrimSpace(theirs))
		if err != nil || held {
			return true
		}
	}
	return false
}

// CommitPaths commits only the listed paths in one commit. A dirty file
// outside the list stays out, so one run of a tool is one commit. One path
// is a one-file write, so it may fold like CommitOrFold. More paths never do.
func CommitPaths(repo string, paths []string, msg string) Result {
	if len(paths) == 1 {
		return CommitOrFold(repo, paths[0], msg, false)
	}
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

// IsAncestor says whether commit a comes before commit b on the branch, so
// the newer of two commits wins whatever subject it uses. It runs git
// merge-base --is-ancestor, where exit 1 means no and any other failure is
// an error, so a typo never looks like an order.
func IsAncestor(repo, a, b string) (bool, error) {
	cmd := exec.Command("git", "-C", repo, "merge-base", "--is-ancestor", a, b)
	var errOut bytes.Buffer
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("git merge-base --is-ancestor: %v: %s", err, strings.TrimSpace(errOut.String()))
	}
	return true, nil
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

// Touched lists the files that the commits in rng changed, from the repo root.
// It reads the commits themselves, not the two ends of the range, so a file a
// commit wrote and a later commit put back still comes back. rng is a commit
// range like main..live. Names come back as they are, with no quotes around
// the odd ones.
func Touched(repo, rng string) ([]string, error) {
	// A dash would make git read the range as an option.
	if strings.HasPrefix(rng, "-") {
		return nil, fmt.Errorf("bad range %q: it starts with a dash", rng)
	}
	// The last -- says rng is a revision, never a file name that happens to match.
	out, err := run(repo, "log", "--no-renames", "--name-only", "--format=", "-z", rng, "--")
	if err != nil {
		return nil, err
	}
	var files []string
	seen := map[string]bool{}
	for _, f := range strings.Split(out, "\x00") {
		// A file written by two commits on the branch comes back twice.
		if f != "" && !seen[f] {
			seen[f] = true
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

// NewestCommit gives the short hash and subject of the newest commit on ref.
// When pre is not empty it only looks at commits whose subject starts with
// pre, and a branch with none of those gives two empty strings and no error,
// so a caller can tell "no such commit" from "no branch".
func NewestCommit(repo, ref, pre string) (string, string, error) {
	args := []string{"log", "--format=%h %s"}
	if pre != "" {
		// git greps the whole message, so the lines it gives back are
		// checked here for the subject that really starts with pre.
		args = append(args, "--fixed-strings", "--grep="+pre)
	}
	args = append(args, ref)
	out, err := run(repo, args...)
	if err != nil {
		return "", "", err
	}
	for _, ln := range strings.Split(out, "\n") {
		hash, subject, ok := strings.Cut(strings.TrimSpace(ln), " ")
		if ok && strings.HasPrefix(subject, pre) {
			return hash, subject, nil
		}
	}
	return "", "", nil
}
