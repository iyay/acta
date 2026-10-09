// Package commits finds the commits that name a plan task in a Task: trailer.
package commits

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Ref is a git ref to search, for example "HEAD" or a branch name.
type Ref struct{ Name string }

// Commit is one commit that names a task.
type Commit struct {
	Sha     string    // full sha
	Date    time.Time // author date
	Subject string
	Branch  string // the Ref.Name it was first found from
	Chore   bool   // subject starts with "chore("
}

// trailerLine matches one Task: line. The hash is the plan's 7 character hash.
var trailerLine = regexp.MustCompile(`^Task: PLN-([a-z0-9]{7})#(\d+)$`)

// Record and field separators. Git never prints them in a subject or body.
// logFormat asks git for them with %x00 and %x1e, because an argument cannot
// hold a raw NUL byte.
const (
	fieldSep  = "\x00"
	recordSep = "\x1e"
	logFormat = "--format=%H%x00%aI%x00%s%x00%b%x1e"
)

// Find returns map key "<plan hash>#<n>" -> commits, for example "broaksz#3".
// Refs are walked in the given order, so a commit seen from two refs keeps
// the Branch of the first one. A ref git cannot resolve is skipped. Each key
// lists its commits oldest first by author date.
func Find(repo string, refs []Ref) (map[string][]Commit, error) {
	out := map[string][]Commit{}
	if _, err := run(repo, "rev-parse", "--git-dir"); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, ref := range refs {
		// A name that starts with "-" would be read as an option. It is no ref.
		if ref.Name == "" || strings.HasPrefix(ref.Name, "-") {
			continue
		}
		// An unknown ref, or HEAD of a repo with no commits, is skipped.
		if _, err := run(repo, "rev-parse", "--verify", "-q", ref.Name+"^{commit}"); err != nil {
			continue
		}
		raw, err := run(repo, "log", "--grep=^Task: ",
			logFormat, ref.Name, "--")
		if err != nil {
			return nil, err
		}
		// git log lists newest first. Reverse it so ties keep the old-to-new order.
		recs := strings.Split(raw, recordSep)
		slices.Reverse(recs)
		for _, rec := range recs {
			c, body, ok := parseRecord(rec, ref.Name)
			if !ok || seen[c.Sha] {
				continue
			}
			seen[c.Sha] = true
			for _, key := range Trailers(body) {
				out[key] = append(out[key], c)
			}
		}
	}
	for _, list := range out {
		slices.SortStableFunc(list, func(a, b Commit) int { return a.Date.Compare(b.Date) })
	}
	return out, nil
}

// parseRecord reads one record of the log format. ok is false for the empty
// piece after the last separator and for a record that is cut short.
func parseRecord(rec, branch string) (Commit, string, bool) {
	rec = strings.TrimLeft(rec, "\n")
	f := strings.SplitN(rec, fieldSep, 4)
	if len(f) != 4 {
		return Commit{}, "", false
	}
	date, err := time.Parse(time.RFC3339, f[1])
	if err != nil {
		return Commit{}, "", false
	}
	return Commit{
		Sha:     f[0],
		Date:    date,
		Subject: f[2],
		Branch:  branch,
		Chore:   strings.HasPrefix(f[2], "chore("),
	}, f[3], true
}

// Trailers returns the task keys ("<plan hash>#<n>") named by Task: trailers
// in a commit body. Git reads trailers from the last paragraph only, so a
// matching line anywhere else does not count. A bad value is skipped.
func Trailers(body string) []string {
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	start := len(lines)
	for start > 0 && strings.TrimSpace(lines[start-1]) != "" {
		start--
	}
	var keys []string
	for _, ln := range lines[start:] {
		m := trailerLine.FindStringSubmatch(strings.TrimRight(ln, " \t"))
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[2])
		if err != nil {
			continue
		}
		key := m[1] + "#" + strconv.Itoa(n)
		if !slices.Contains(keys, key) {
			keys = append(keys, key)
		}
	}
	return keys
}

// Diff returns git show --stat -p --no-color --no-ext-diff <sha> output.
func Diff(repo, sha string) (string, error) {
	return run(repo, "show", "--stat", "-p", "--no-color", "--no-ext-diff", sha, "--")
}

// run runs the git binary in repo, the same way internal/gitc does.
func run(repo string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return out.String(), nil
}
