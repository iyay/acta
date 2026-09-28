package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnsureGitignore adds line to the root folder's .gitignore when it is not
// there yet. acta writes files it must never commit, and a gitignore line is
// what keeps them out of every commit and out of git status. Callers ignore
// the error: a session hook that failed would get in the user's way.
//
// It writes nothing when the root folder is gone or when the folder is not in
// a git repo, because an ignore line outside a repo means nothing.
//
// It also writes nothing when the root folder, after links are followed, is
// outside the repo.
//
// It also writes nothing when .gitignore is a link.
func EnsureGitignore(root, line string) error {
	if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
		return nil
	}
	// A root that is itself a link would lead the walk into the folder it
	// points at, which may hold another repo's .git. Start from its parent
	// so the repo found is the one the link sits in.
	start := root
	if fi, err := os.Lstat(root); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		start = filepath.Dir(root)
	}
	top, ok := gitTop(start)
	if !ok {
		return nil
	}
	// The walk above reads the path as written, so a root that is a link to
	// a folder somewhere else still looks like it sits in this repo. Follow
	// the links on both sides and make sure the real root is under the real
	// repo folder, so the write can never land outside it.
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	realTop, err := filepath.EvalSymlinks(top)
	if err != nil {
		return err
	}
	if rel, err := filepath.Rel(realTop, realRoot); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("%s is outside the repo %s", root, top)
	}
	path := filepath.Join(root, ".gitignore")

	// A .gitignore that is a link would send the write to whatever file it
	// points at, even one outside the repo. Git does not read a linked
	// .gitignore either, so there is nothing to gain by following it.
	if fi, err := os.Lstat(path); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s is a link, not a file", path)
	}
	old, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if hasLine(string(old), line) {
		return nil
	}
	if len(old) > 0 && !strings.HasSuffix(string(old), "\n") {
		old = append(old, '\n')
	}
	return os.WriteFile(path, append(old, []byte(line+"\n")...), 0o644)
}

// hasLine matches the line on its own, so a commented-out ".agents.json" does
// not count as present but a trailing comment does not hide it either.
func hasLine(text, line string) bool {
	for _, l := range strings.Split(text, "\n") {
		if strings.TrimSpace(l) == line {
			return true
		}
	}
	return false
}

// gitTop walks up from dir looking for .git and returns the folder that
// holds it. .git is a file in a linked worktree, so only its presence is
// checked, not that it is a folder.
func gitTop(dir string) (string, bool) {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
