package cli

import (
	"fmt"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/gitc"
)

// building says why a plan cannot be changed from here, or "" when it can.
// While a linked worktree builds a plan, that worktree holds its ticks, so a
// change made on main too would clash at merge. Inside any linked worktree
// every write goes through.
func building(cfg config.Config, it *board.Item) string {
	if it == nil || it.Kind != board.KindPlan || gitc.IsLinkedWorktree(cfg.RepoRoot) {
		return ""
	}
	for _, r := range board.PlanStates(cfg) {
		if r.Plan.ID == it.ID {
			return fmt.Sprintf("%s is building in %s; change it there", planName(it), r.Worktree)
		}
	}
	return ""
}
