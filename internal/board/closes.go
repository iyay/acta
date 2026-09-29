package board

import (
	"fmt"
	"strings"
)

// listField reads a frontmatter value that may hold one name or several. A
// list gives its items, a plain string gives one item, and anything else is
// not a list of names and gives nothing.
func listField(front map[string]any, key string) []string {
	switch v := front[key].(type) {
	case []any:
		out := make([]string, 0, len(v))
		for _, e := range v {
			if s := strings.TrimSpace(fmt.Sprint(e)); s != "" {
				out = append(out, s)
			}
		}
		return out
	case string:
		if s := strings.TrimSpace(v); s != "" {
			return []string{s}
		}
	}
	return nil
}

// linkCloses reads the closes: list of every spec, plan and bug. A target must
// be a scratch item, a spec or a debt item, since those are the items whose
// status is not decided by their own tasks. An entry that points at nothing,
// or at a kind that cannot be closed, is a problem on the item that wrote it
// and never stops the rest of the board from loading.
func (b *Board) linkCloses() {
	for _, it := range b.Items {
		switch it.Kind {
		case KindStory, KindPlan, KindBug:
		default:
			continue
		}
		for _, entry := range it.fmCloses {
			target := b.Get(entry)
			if target == nil {
				it.Problems = append(it.Problems, "closes "+entry+" not found")
				continue
			}
			switch target.Kind {
			case KindScratch, KindStory, KindDebtItem:
			default:
				it.Problems = append(it.Problems, "closes "+entry+": a "+kindWord(target.Kind)+" cannot be closed")
				continue
			}
			if contains(it.Closes, target.ID) {
				continue
			}
			it.Closes = append(it.Closes, target.ID)
			target.ClosedBy = append(target.ClosedBy, it.ID)
			// A spec closed by a plan carries that plan's work, so it counts
			// the plan and the plan's tasks, the same way a Spec line does.
			// A plan can name the same spec from both sides, so a spec it
			// already counts on keeps one plan and one copy of each task.
			if target.Kind == KindStory && it.Kind == KindPlan && !contains(it.countedOn, target.ID) {
				target.plans++
				target.Children = append(target.Children, it.Children...)
			}
		}
	}
}

// kindWord names a kind the way a person does. A spec file is a story on the
// board, but everyone calls it a spec.
func kindWord(k Kind) string {
	if k == KindStory {
		return "spec"
	}
	return string(k)
}

// closedByStatus gives a spec with no plan of its own the status of the specs
// and bugs that close it: done once every one of them is closed, in progress
// while one of them is being worked on, and approved when none has started. A
// spec with a plan of its own, or a spec whose status was written down, keeps
// the answer its own plans and frontmatter gave, so the bool is false there.
func closedByStatus(b *Board, it *Item) (string, bool) {
	if it.Kind != KindStory || it.plans > 0 || it.fmStatus != "" {
		return "", false
	}
	closers, going, waiting := 0, 0, 0
	for _, id := range it.ClosedBy {
		c := b.byID[id]
		if c == nil || (c.Kind != KindStory && c.Kind != KindBug) {
			continue
		}
		closers++
		switch {
		case c.Status == "in-progress" || c.Status == "fixing":
			going++
		case !Closed(c.Status):
			waiting++
		}
	}
	switch {
	case closers == 0:
		return "", false
	case going > 0:
		return "in-progress", true
	case waiting > 0:
		return "approved", true
	}
	return "done", true
}
