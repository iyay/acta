package write

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/gitc"
)

// dropLowReason goes in the commit, so git log says why the lines closed.
const dropLowReason = "low, under new debt bar"

// DropLowDebt marks every open debt line that is not medium or high as
// wontfix, across all debt files, in one commit. Low notes piled up as debt
// nobody fixed, so they are closed once here. Closed lines and medium or
// high lines are left alone. With nothing to close it writes and commits
// nothing.
func DropLowDebt(cfg config.Config) (Outcome, error) {
	dir := filepath.Join(cfg.Root, cfg.Dirs.Debt)
	files, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return Outcome{}, err
	}
	var changed []string
	for _, path := range files {
		did, err := dropLowIn(path)
		if err != nil {
			return Outcome{}, err
		}
		if did {
			changed = append(changed, path)
		}
	}
	o := Outcome{Path: dir}
	if len(changed) == 0 {
		o.Skipped, o.Reason = true, "no low debt is open"
		return o, nil
	}
	if !cfg.AutoCommit {
		o.Reason = "auto_commit is off"
		return o, nil
	}
	r := gitc.CommitPathsIn(cfg.RepoRoot, cfg.Root, changed, Subject(cfg, changed, "drop low debt: "+dropLowReason))
	o.Committed, o.Reason = r.Committed, r.Reason
	o.Skipped = !r.Committed
	return o, nil
}

// dropLowIn closes the open low and untagged lines of one debt file under
// its lock, so a tick running at the same time is not lost.
func dropLowIn(path string) (bool, error) {
	unlock, err := lock(path)
	if err != nil {
		return false, err
	}
	defer unlock()
	src, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	lines := strings.Split(string(src), "\n")
	did := false
	for i, ln := range lines {
		m := tickAnyBoxRe.FindStringSubmatch(ln)
		if m == nil || m[2] != " " {
			continue
		}
		// lineText drops the tag, so read it from the text right after the box.
		if level, _ := board.SplitPriority(strings.TrimPrefix(ln, m[0]+" ")); level == "medium" || level == "high" {
			continue
		}
		lines[i] = tickAnyBoxRe.ReplaceAllString(ln, "${1}-${3}")
		did = true
	}
	if !did {
		return false, nil
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return false, err
	}
	return true, os.Rename(tmp, path)
}
