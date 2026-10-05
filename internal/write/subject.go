package write

import (
	"path/filepath"
	"strings"

	"github.com/iyay/acta/internal/config"
)

// Subject builds the commit subject for the files a write touched. One kind
// of file gives a scope for it, more than one kind gives none, so the log
// says what the commit holds without naming the tool.
func Subject(cfg config.Config, paths []string, what string) string {
	kinds := map[string]bool{}
	for _, p := range paths {
		kinds[kindOf(cfg, p)] = true
	}
	if len(kinds) == 1 {
		for k := range kinds {
			if k != "" {
				return "chore(" + k + "): " + what
			}
		}
	}
	return "chore: " + what
}

// kindOf gives the scope for a planning file from the folder it sits in.
// Anything else gives "", so it never claims a scope of its own.
func kindOf(cfg config.Config, path string) string {
	rel, err := filepath.Rel(cfg.Root, path)
	if err != nil {
		return ""
	}
	first, _, _ := strings.Cut(filepath.ToSlash(rel), "/")
	switch first {
	case cfg.Dirs.Scratch:
		return "scratch"
	case cfg.Dirs.Specs:
		return "spec"
	case cfg.Dirs.Plans:
		return "plan"
	case cfg.Dirs.Bugs:
		return "bug"
	case cfg.Dirs.Debt:
		return "debt"
	case "wiki":
		return "wiki"
	}
	return ""
}

// subjectWhat keeps the old words after the id, so a set or mark commit
// still says what changed and only the prefix is new.
func subjectWhat(id, field, value string) string {
	return id + " " + field + " " + value
}
