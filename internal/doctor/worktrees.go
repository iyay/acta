package doctor

import (
	"fmt"
	"strings"
)

// checkWorktrees warns about linked worktrees that no open plan owns. It
// only warns: junk on disk never breaks the setup. Clean and missing strays
// are counted apart from dirty ones, since acta worktrees --clean removes
// the first kind and never the second, so the fix must not promise more.
func checkWorktrees(e Env) Result {
	r := Result{Name: "worktrees", Level: OK, Msg: "no stray worktrees"}
	var parts []string
	if n := len(e.CleanStrays); n > 0 {
		parts = append(parts, fmt.Sprintf("%d stray %s --clean can remove: %s", n, worktreeNoun(n), strings.Join(e.CleanStrays, ", ")))
	}
	if n := len(e.DirtyStrays); n > 0 {
		parts = append(parts, fmt.Sprintf("%d dirty stray %s: %s", n, worktreeNoun(n), strings.Join(e.DirtyStrays, ", ")))
	}
	if len(parts) == 0 {
		return r
	}
	r.Level, r.Msg = Warn, strings.Join(parts, "; ")
	const dirtyFix = "dirty worktrees hold files only you can keep or delete"
	switch {
	case len(e.CleanStrays) > 0 && len(e.DirtyStrays) > 0:
		r.Fix = "acta worktrees --clean removes the clean ones; " + dirtyFix
	case len(e.CleanStrays) > 0:
		r.Fix = "acta worktrees --clean"
	default:
		r.Fix = dirtyFix
	}
	return r
}

func worktreeNoun(n int) string {
	if n == 1 {
		return "worktree"
	}
	return "worktrees"
}
