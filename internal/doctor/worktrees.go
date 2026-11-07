package doctor

import (
	"fmt"
	"strings"
)

// checkWorktrees warns about linked worktrees that no open plan owns. It
// only warns: junk on disk never breaks the setup.
func checkWorktrees(e Env) Result {
	r := Result{Name: "worktrees", Level: OK, Msg: "no stray worktrees"}
	if len(e.StrayWorktrees) > 0 {
		r.Level, r.Fix = Warn, "acta worktrees --clean"
		r.Msg = fmt.Sprintf("%d stray worktrees: %s", len(e.StrayWorktrees), strings.Join(e.StrayWorktrees, ", "))
	}
	return r
}
