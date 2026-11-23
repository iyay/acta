package write

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// LandNotes writes the notes that close a plan when it lands: fixed_in on
// each bug it fixes, a tick on each debt line it closes, and its spec dates.
// Tidy runs it on the merged tree and acta land-notes runs it on main, so
// both land modes write the same thing. It writes files and never commits,
// and gives back the paths it changed. Every id is checked before the first
// write, so an id with no file leaves every file as it was.
func LandNotes(cfg config.Config, planHash string) ([]string, error) {
	b, err := board.Load(cfg)
	if err != nil {
		return nil, err
	}
	plan := b.Get(planHash)
	if plan == nil || plan.Kind != board.KindPlan {
		return nil, fmt.Errorf("%s is not a plan", planHash)
	}
	// The hash never changes on a renumber, so the bug keeps pointing at
	// the right plan. An old plan with no full hash has nothing to write.
	fixedIn := plan.Hash
	if !validFixedIn(fixedIn) {
		return nil, fmt.Errorf("%s has no 7 character plan hash; run acta id first", plan.ID)
	}
	bugs, lines, err := landTargets(b, plan)
	if err != nil {
		return nil, err
	}
	var wrote []string
	for _, bug := range bugs {
		changed, err := markDate(bug.Path, func(src []byte) ([]byte, error) {
			out, err := SetField(src, "fixed_in", fixedIn)
			if err != nil {
				return nil, err
			}
			// A tick may have closed the bug already; that day stays.
			return MarkFinishedOnce(out)
		})
		if err != nil {
			return wrote, err
		}
		if changed {
			wrote = append(wrote, bug.Path)
		}
	}
	for _, ln := range lines {
		if err := TickLine(ln.Path, ln.Line, ln.Title, 'x'); err != nil {
			return wrote, err
		}
		wrote = append(wrote, ln.Path)
	}
	dates, err := specDates(cfg, b, plan.ID)
	return unique(append(wrote, dates...)), err
}

// landTargets finds the bugs a plan fixes and the open debt lines it closes.
// The bug is the plan's own parent, plus each bug in closes. A debt file in
// closes gives all its open lines, and a debt line gives itself while open,
// so a line marked wontfix stays wontfix. Any closes id the board does not
// hold is an error that names it.
func landTargets(b *board.Board, plan *board.Item) (bugs, lines []*board.Item, err error) {
	if p := b.Get(plan.SpecID); p != nil && p.Kind == board.KindBug {
		bugs = append(bugs, p)
	}
	ids, err := closesOf(plan.Path)
	if err != nil {
		return nil, nil, err
	}
	for _, id := range ids {
		it := b.Get(id)
		if it == nil {
			return nil, nil, fmt.Errorf("%s closes %s, but no file holds it", plan.ID, id)
		}
		switch it.Kind {
		case board.KindBug:
			// The parent bug named again here writes nothing new the
			// second time, so it needs no check.
			bugs = append(bugs, it)
		case board.KindDebt:
			for _, child := range it.Children {
				if ln := b.Get(child); ln != nil && ln.Status == "open" {
					lines = append(lines, ln)
				}
			}
		case board.KindDebtItem:
			if it.Status == "open" {
				lines = append(lines, it)
			}
		}
	}
	return bugs, lines, nil
}

// closesOf reads the closes list as the plan file writes it. The board keeps
// only the ids it can close by status, and a bug or a whole debt file is not
// one of those, so its own list would leave them out.
func closesOf(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ids []string
	switch v := board.Parse(raw).Front["closes"].(type) {
	case []any:
		for _, e := range v {
			ids = append(ids, strings.TrimSpace(fmt.Sprint(e)))
		}
	case string:
		ids = []string{strings.TrimSpace(v)}
	}
	return slices.DeleteFunc(ids, func(id string) bool { return id == "" }), nil
}
