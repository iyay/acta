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
// It also writes nothing when .gitignore is a link.
func EnsureGitignore(root, line string) error {
	if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
		return nil
	}
	if !inGitRepo(root) {
		return nil
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

// inGitRepo walks up from dir looking for .git. It is a file in a linked
// worktree, so only its presence is checked, not that it is a folder.
func inGitRepo(dir string) bool {
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
