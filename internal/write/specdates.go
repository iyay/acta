package write

import (
	"fmt"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// SpecDates writes the dates of a spec, on the branch a plan just landed on.
// Tick in a worktree leaves a spec alone (see TaskDates), so land calls this
// once per plan. started is the earliest started of the spec's plans here and
// may move earlier on a later land, never later; a plan that started first
// can land last. finished is written once, by the land that makes every plan
// of the spec done. A bug with one plan gets its dates from tick, so this
// writes nothing for it. It gives back the files it changed.
func SpecDates(cfg config.Config, planID string) ([]string, error) {
	b, err := board.Load(cfg)
	if err != nil {
		return nil, err
	}
	plan := b.Get(planID)
	if plan == nil || plan.Kind != board.KindPlan {
		return nil, fmt.Errorf("%s is not a plan", planID)
	}
	spec := b.Get(plan.SpecID)
	if spec == nil {
		return nil, nil
	}
	// Load reads only this checkout, so these are the plans as the parent
	// holds them; a plan still in a worktree shows here unticked.
	var plans []*board.Item
	for _, it := range b.Items {
		if it.Kind == board.KindPlan && it.SpecID == spec.ID {
			plans = append(plans, it)
		}
	}
	if len(plans) < 2 && kindOf(cfg, spec.Path) != "spec" {
		return nil, nil
	}
	earliest, allDone, anyDone := "", true, false
	for _, p := range plans {
		if p.StartedOn != "" && (earliest == "" || p.StartedOn < earliest) {
			earliest = p.StartedOn
		}
		// A dropped plan is closed too; it must not hold finished back.
		switch p.Status {
		case "done":
			anyDone = true
		case "dropped":
		default:
			allDone = false
		}
	}
	var wrote []string
	write := func(set func([]byte) ([]byte, error)) error {
		changed, err := markDate(spec.Path, set)
		if changed && (len(wrote) == 0 || wrote[len(wrote)-1] != spec.Path) {
			wrote = append(wrote, spec.Path)
		}
		return err
	}
	if earliest != "" && (spec.StartedOn == "" || earliest < spec.StartedOn) {
		if err := write(func(src []byte) ([]byte, error) { return SetField(src, "started", earliest) }); err != nil {
			return wrote, err
		}
	}
	if allDone && anyDone {
		if err := write(MarkFinishedOnce); err != nil {
			return wrote, err
		}
	}
	return wrote, nil
}
