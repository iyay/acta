// Package worktrees finds the git worktrees that agents left behind and no
// open plan owns, so they can be removed without ever forcing a delete.
package worktrees

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Idle is how long a worktree's git files must stay untouched before it can
// count as a stray. A live agent in another tab keeps touching them.
const Idle = 24 * time.Hour

// maxStatus is how many status lines a dirty worktree shows.
const maxStatus = 10

// The states an Entry can have. The spec names them for --json.
const (
	StateProtected = "protected"
	StateClean     = "clean"
	StateDirty     = "dirty"
	StateMissing   = "missing"
	StateRemoved   = "removed"
	StateKept      = "kept"
)

// Plan is what the rules need to know about one plan.
type Plan struct {
	Slug, Status string
}

// Entry is one linked worktree and what the rules make of it.
type Entry struct {
	Path       string   `json:"path"`
	Branch     string   `json:"branch"`
	State      string   `json:"state"`
	Reason     string   `json:"reason"`
	BranchKept bool     `json:"branch_kept"`
	Status     []string `json:"status"`
}

// runGit runs git in dir. Tests swap it to see every call.
var runGit = func(dir string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		return text, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, text)
	}
	return text, nil
}

// porcelain is one block of git worktree list --porcelain.
type porcelain struct {
	path, branch, head string
	locked, missing    bool
}

// List gives every linked worktree of the repo that holds dir, except the
// main checkout and the worktree dir sits in. It only reads.
func List(dir string, plans []Plan, now time.Time) ([]Entry, error) {
	_, res, err := list(dir, plans, now)
	return res, err
}

// Clean removes each clean or missing stray with a plain git worktree
// remove, then deletes its branch with git branch -d. It never forces
// either one, so git itself refuses anything that would lose work. A
// refusal is reported in the entry, never as an error.
func Clean(dir string, plans []Plan, now time.Time) ([]Entry, error) {
	main, res, err := list(dir, plans, now)
	if err != nil {
		return nil, err
	}
	out := make([]Entry, 0, len(res))
	for _, e := range res {
		if e.State == StateClean || e.State == StateMissing {
			e = remove(main, e)
		}
		out = append(out, e)
	}
	return out, nil
}

// remove takes one stray away and gives back its new entry.
func remove(main string, e Entry) Entry {
	if _, err := runGit(main, "worktree", "remove", e.Path); err != nil {
		e.State, e.Reason = StateKept, oneLine(err)
		return e
	}
	e.State = StateRemoved
	if e.Branch == "" {
		return e
	}
	// The -- keeps a branch named like a flag from being read as one.
	if _, err := runGit(main, "branch", "-d", "--", e.Branch); err != nil {
		e.BranchKept, e.Reason = true, oneLine(err)
	}
	return e
}

// oneLine gives an error's text with its line breaks joined. Git adds hint
// lines to its errors, and a reason must stay on one line in the text list.
func oneLine(err error) string {
	return strings.ReplaceAll(err.Error(), "\n", "; ")
}

// list does the work of List and also gives the main checkout's path, which
// every remove runs from.
func list(dir string, plans []Plan, now time.Time) (string, []Entry, error) {
	top, err := runGit(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", nil, err
	}
	out, err := runGit(dir, "worktree", "list", "--porcelain")
	if err != nil {
		return "", nil, err
	}
	blocks := parse(out)
	if len(blocks) == 0 {
		return "", nil, fmt.Errorf("git listed no worktrees in %s", dir)
	}
	main := blocks[0].path
	admins, err := adminDirs(dir)
	if err != nil {
		return "", nil, err
	}
	var res []Entry
	for _, b := range blocks[1:] {
		if same(b.path, top) {
			continue
		}
		res = append(res, judge(main, b, admins[resolve(b.path)], plans, now))
	}
	return main, res, nil
}

// judge applies the rules to one worktree, in the spec's order.
func judge(main string, b porcelain, admin string, plans []Plan, now time.Time) Entry {
	// Status starts empty, not nil, so --json prints [] and never null.
	e := Entry{Path: b.path, Branch: b.branch, Status: []string{}}
	// A lock comes first: people lock a worktree that lives on a drive that
	// is not mounted, so its folder can look gone while it is not.
	if b.locked {
		e.State, e.Reason = StateProtected, "locked"
		return e
	}
	if _, err := os.Stat(b.path); b.missing || err != nil {
		e.State = StateMissing
		// A detached worktree can be the only holder of its commit even
		// with its folder gone. A branch keeps its commits on its own,
		// since git branch -d refuses an unmerged branch. That holds only
		// while the branch ref exists: one deleted by hand holds nothing.
		if reason := missingWork(main, b, admin); reason != "" {
			e.State, e.Reason = StateProtected, reason
		}
		return e
	}
	// Read the file times before any git command runs in the worktree: a
	// plain git status rewrites the index and would reset the clock. With no
	// admin folder, or none of its time files, there is no time to read, and
	// an unknown age must not pass for an idle one: an agent may be working
	// in it right now.
	last := lastActivity(admin)
	if last.IsZero() {
		e.State, e.Reason = StateProtected, "cannot read activity times"
		return e
	}
	if now.Sub(last) < Idle {
		e.State, e.Reason = StateProtected, "active in the last 24 hours"
		return e
	}
	if reason := unmerged(main, b); reason != "" {
		e.State, e.Reason = StateProtected, reason
		return e
	}
	if b.branch != "" {
		for _, p := range plans {
			if p.Slug == b.branch && p.Status != "done" && p.Status != "dropped" {
				e.State, e.Reason = StateProtected, "branch of open plan "+p.Slug
				return e
			}
		}
	}
	out, err := runGit(b.path, "--no-optional-locks", "status", "--porcelain")
	if err != nil {
		e.State, e.Reason = StateProtected, "cannot read git status: "+oneLine(err)
		return e
	}
	if out == "" {
		e.State = StateClean
		return e
	}
	e.State = StateDirty
	lines := strings.Split(out, "\n")
	if len(lines) > maxStatus {
		rest := len(lines) - maxStatus
		lines = append(lines[:maxStatus:maxStatus], fmt.Sprintf("... and %d more", rest))
	}
	e.Status = lines
	return e
}

// unmerged says why a worktree holds commits that removing it could lose, or
// "" when it holds none. A branch is checked against the main checkout's
// HEAD. A detached worktree is checked against every branch, tag and remote,
// since its commits have no branch of their own to stay on. When git cannot
// tell, the answer is to keep it: lost work costs more than junk.
func unmerged(main string, b porcelain) string {
	args := []string{"rev-list", "--count", "HEAD..refs/heads/" + b.branch}
	what := "branch has unmerged commits"
	if b.branch == "" {
		args = []string{"rev-list", "--count", b.head, "--not", "--branches", "--tags", "--remotes"}
		what = "commits no branch, tag or remote holds"
	}
	n, err := runGit(main, args...)
	switch {
	case err != nil:
		return "cannot read its commits: " + oneLine(err)
	case n != "0":
		return what
	}
	return ""
}

// missingWork says why a worktree whose folder is gone still holds commits
// that removing its entry could lose, or "" when it holds none. A detached
// one can be the only holder of its commit. A branch keeps its commits on its
// own, since git branch -d refuses an unmerged branch, but only while the
// branch ref exists. When someone deleted the ref by hand, git shows a zero
// HEAD, so the commit comes from the last line of the admin reflog instead.
func missingWork(main string, b porcelain, admin string) string {
	if b.branch == "" {
		return unmerged(main, b)
	}
	if _, err := runGit(main, "show-ref", "--verify", "--quiet", "refs/heads/"+b.branch); err == nil {
		return ""
	}
	head, err := lastReflogCommit(admin)
	if err != nil {
		return "cannot read its commits: " + oneLine(err)
	}
	return unmerged(main, porcelain{head: head})
}

// lastReflogCommit gives the commit the newest line of the admin reflog moved
// to. Each line starts with the old id and the new id.
func lastReflogCommit(admin string) (string, error) {
	// With no admin folder the path would point into the folder the command
	// runs in, so stop here instead of reading some other reflog.
	if admin == "" {
		return "", fmt.Errorf("no admin folder")
	}
	raw, err := os.ReadFile(filepath.Join(admin, "logs", "HEAD"))
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	fields := strings.Fields(lines[len(lines)-1])
	if len(fields) < 2 {
		return "", fmt.Errorf("no commit in %s", filepath.Join(admin, "logs", "HEAD"))
	}
	return fields[1], nil
}

// parse reads git worktree list --porcelain. The main checkout comes first.
func parse(out string) []porcelain {
	var res []porcelain
	for _, block := range strings.Split(strings.TrimSpace(out), "\n\n") {
		var p porcelain
		bare := false
		for _, ln := range strings.Split(block, "\n") {
			switch {
			case strings.HasPrefix(ln, "worktree "):
				p.path = strings.TrimPrefix(ln, "worktree ")
			case strings.HasPrefix(ln, "HEAD "):
				p.head = strings.TrimPrefix(ln, "HEAD ")
			case strings.HasPrefix(ln, "branch "):
				p.branch = strings.TrimPrefix(strings.TrimPrefix(ln, "branch "), "refs/heads/")
			case ln == "locked" || strings.HasPrefix(ln, "locked "):
				p.locked = true
			case ln == "prunable" || strings.HasPrefix(ln, "prunable "):
				p.missing = true
			case ln == "bare":
				bare = true
			}
		}
		if p.path != "" && !bare {
			res = append(res, p)
		}
	}
	return res
}

// adminDirs maps each linked worktree's folder to its own git folder inside
// the common git dir. Each of those holds a gitdir file that points back at
// the worktree's .git file.
func adminDirs(dir string) (map[string]string, error) {
	common, err := runGit(dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return nil, err
	}
	res := map[string]string{}
	names, err := os.ReadDir(filepath.Join(common, "worktrees"))
	if err != nil {
		// No linked worktree was ever made, so there is nothing to map.
		return res, nil
	}
	for _, n := range names {
		admin := filepath.Join(common, "worktrees", n.Name())
		raw, err := os.ReadFile(filepath.Join(admin, "gitdir"))
		if err != nil {
			continue
		}
		// With worktree.useRelativePaths the file holds a path relative to
		// the admin folder, so join it before comparing.
		gitFile := strings.TrimSpace(string(raw))
		if !filepath.IsAbs(gitFile) {
			gitFile = filepath.Join(admin, gitFile)
		}
		res[resolve(filepath.Dir(gitFile))] = admin
	}
	return res, nil
}

// lastActivity is the newest change time of the files git touches when
// someone works in the worktree. With no admin folder it is the zero time.
func lastActivity(admin string) time.Time {
	var last time.Time
	if admin == "" {
		return last
	}
	for _, f := range []string{"HEAD", "index", filepath.Join("logs", "HEAD")} {
		if st, err := os.Stat(filepath.Join(admin, f)); err == nil && st.ModTime().After(last) {
			last = st.ModTime()
		}
	}
	return last
}

// resolve follows links in a path, so /tmp and /private/tmp on macOS match.
// A path that no longer exists keeps its plain form.
func resolve(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return filepath.Clean(p)
}

func same(a, b string) bool { return resolve(a) == resolve(b) }
