package gitc

import (
	"slices"
	"strings"
)

// WikiHistory is what acta wiki check needs from git, read once for every
// page. On a long history each git log walks every commit, so asking per page
// cost minutes. The same rules as CheckCommits and Unseen apply.
type WikiHistory struct {
	parents map[string][]string // full sha to its parent shas, for HEAD
	commits []histCommit        // newest first, as git log gives them
	checks  map[string][]string // page id to the commits that added its line
	dirty   map[string]bool     // files under the wiki folder with edits not committed
}

// histCommit is one commit that HEAD can reach, with the files it changed.
type histCommit struct {
	sha, short, subject string
	files               []string
}

// ReadWikiHistory reads the wiki history in four git commands, whatever the
// number of pages. wikiDir is the wiki folder from the repo root.
func ReadWikiHistory(repo, wikiDir, checkedFile string) (WikiHistory, error) {
	h := WikiHistory{parents: map[string][]string{}, dirty: map[string]bool{}}
	out, err := run(repo, "rev-list", "--parents", "HEAD")
	if err != nil {
		return h, err
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if f := strings.Fields(line); len(f) > 0 {
			h.parents[f[0]] = f[1:]
		}
	}
	out, err = run(repo, checkedLogArgs(checkedFile)...)
	if err != nil {
		return h, err
	}
	h.checks = checkLines(out)
	// Every commit with the files it changed. No path filter: git compares
	// each file with each path it gets, so a few dozen paths made this one log
	// twice as slow as reading it all. The caller matches the paths itself.
	// No renames, so a file moved out of a path still lists its old name.
	out, err = run(repo, "log", "--no-merges", "--no-renames", "--name-only", "-z",
		"--format=%x01%H%x00%h%x00%s", "HEAD")
	if err != nil {
		return h, err
	}
	for _, chunk := range strings.Split(out, "\x01")[1:] {
		f := strings.Split(chunk, "\x00")
		if len(f) < 3 {
			continue
		}
		c := histCommit{sha: f[0], short: f[1], subject: f[2]}
		for _, name := range f[3:] {
			if name = strings.TrimPrefix(name, "\n"); name != "" {
				c.files = append(c.files, name)
			}
		}
		h.commits = append(h.commits, c)
	}
	// Every untracked file on its own, so a new page in a new folder is not
	// hidden behind one line for the folder.
	out, err = run(repo, "--literal-pathspecs", "status", "--porcelain", "-z", "--untracked-files=all", "--", wikiDir)
	if err != nil {
		return h, err
	}
	// Each entry is "XY name". A rename or copy is followed by its old name
	// alone, which counts as not committed too.
	f := strings.Split(out, "\x00")
	for i := 0; i < len(f); i++ {
		if len(f[i]) < 4 {
			continue
		}
		h.dirty[f[i][3:]] = true
		if f[i][0] == 'R' || f[i][0] == 'C' {
			if i++; i < len(f) {
				h.dirty[f[i]] = true
			}
		}
	}
	return h, nil
}

// Dirty says whether file, from the repo root, has edits not committed.
func (h WikiHistory) Dirty(file string) bool { return h.dirty[file] }

// CheckCommits gives what the function of the same name gives for pageFile
// and id: the commits that changed the page file, newest first, then the
// commits that added a line for id to the checked file.
func (h WikiHistory) CheckCommits(pageFile, id string) []string {
	var shas []string
	for _, c := range h.commits {
		if slices.Contains(c.files, pageFile) {
			shas = append(shas, c.sha)
		}
	}
	for _, sha := range h.checks[id] {
		if !slices.Contains(shas, sha) {
			shas = append(shas, sha)
		}
	}
	return shas
}

// Unseen gives what the function of the same name gives: the newest commit
// with a file that touches says yes to, and that no check commit has behind
// it. Ancestors come from the parent graph, and a check sees itself.
func (h WikiHistory) Unseen(checks []string, touches func(file string) bool) (short, subject string) {
	seen := map[string]bool{}
	todo := slices.Clone(checks)
	for len(todo) > 0 {
		sha := todo[len(todo)-1]
		todo = todo[:len(todo)-1]
		if seen[sha] {
			continue
		}
		seen[sha] = true
		todo = append(todo, h.parents[sha]...)
	}
	for _, c := range h.commits {
		if !seen[c.sha] && slices.ContainsFunc(c.files, touches) {
			return c.short, c.subject
		}
	}
	return "", ""
}
