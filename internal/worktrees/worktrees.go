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
	Protected = "protected"
	Clean     = "clean"
	Dirty     = "dirty"
	Missing   = "missing"
	Removed   = "removed"
	Kept      = "kept"
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
	Reason     string   `json:"reason,omitempty"`
	BranchKept bool     `json:"branch_kept"`
	Status     []string `json:"status,omitempty"`
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
	path, branch    string
	locked, missing bool
}

// List gives every linked worktree of the repo that holds dir, except the
// main checkout and the worktree dir sits in. It only reads.
func List(dir string, plans []Plan, now time.Time) ([]Entry, error) {
	top, err := runGit(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}
	out, err := runGit(dir, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	blocks := parse(out)
	if len(blocks) == 0 {
		return nil, fmt.Errorf("git listed no worktrees in %s", dir)
	}
	main := blocks[0].path
	admins, err := adminDirs(dir)
	if err != nil {
		return nil, err
	}
	var res []Entry
	for _, b := range blocks[1:] {
		if same(b.path, top) {
			continue
		}
		res = append(res, judge(main, b, admins[clean(b.path)], plans, now))
	}
	return res, nil
}

// judge applies the rules to one worktree, in the spec's order.
func judge(main string, b porcelain, admin string, plans []Plan, now time.Time) Entry {
	e := Entry{Path: b.path, Branch: b.branch}
	if _, err := os.Stat(b.path); b.missing || err != nil {
		e.State = Missing
		return e
	}
	// Read the file times before any git command runs in the worktree: a
	// plain git status rewrites the index and would reset the clock.
	last := lastActivity(admin)
	switch {
	case b.locked:
		e.State, e.Reason = Protected, "locked"
		return e
	case now.Sub(last) < Idle:
		e.State, e.Reason = Protected, "active in the last 24 hours"
		return e
	}
	if b.branch != "" {
		if n, err := runGit(main, "rev-list", "--count", "HEAD..refs/heads/"+b.branch); err != nil || n != "0" {
			// When git cannot tell, keep it: a lost branch costs more than junk.
			e.State, e.Reason = Protected, "branch has unmerged commits"
			return e
		}
		for _, p := range plans {
			if p.Slug == b.branch && p.Status != "done" && p.Status != "dropped" {
				e.State, e.Reason = Protected, "branch of open plan "+p.Slug
				return e
			}
		}
	}
	out, err := runGit(b.path, "--no-optional-locks", "status", "--porcelain")
	if err != nil {
		e.State, e.Reason = Protected, "cannot read git status: "+err.Error()
		return e
	}
	if out == "" {
		e.State = Clean
		return e
	}
	e.State = Dirty
	lines := strings.Split(out, "\n")
	if len(lines) > maxStatus {
		rest := len(lines) - maxStatus
		lines = append(lines[:maxStatus:maxStatus], fmt.Sprintf("... and %d more", rest))
	}
	e.Status = lines
	return e
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
		res[clean(filepath.Dir(strings.TrimSpace(string(raw))))] = admin
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

// clean resolves links in a path, so /tmp and /private/tmp on macOS match.
// A path that no longer exists keeps its plain form.
func clean(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return filepath.Clean(p)
}

func same(a, b string) bool { return clean(a) == clean(b) }
