// Package cleanup finds planning items whose work may be over and writes the
// closing statuses the user approved.
package cleanup

import (
	"cmp"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// quietDays is how long a brainstorm, a plan or a spec may sit with no work
// before cleanup asks about it. A younger one may still be going.
const quietDays = 14

// Candidate is one item cleanup asks about. Why names the rule that picked it.
type Candidate struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Path    string `json:"path"`
	Status  string `json:"status"`
	AgeDays int    `json:"age_days"`
	Why     string `json:"why"`
}

// kindOrder is the order the groups print in.
var kindOrder = []board.Kind{board.KindDebtItem, board.KindScratch, board.KindBug, board.KindPlan, board.KindStory}

// Candidates lists the items of the main tree that may be finished, grouped
// by kind and oldest first. Items that live only in another worktree are not
// on this board, and plans a linked worktree is building are left out, since
// that work is still going on. It reads git but writes nothing.
func Candidates(cfg config.Config, b *board.Board, now time.Time) ([]Candidate, error) {
	building := map[string]bool{}
	for _, r := range board.PlanStates(cfg) {
		building[r.Plan.ID] = true
	}
	hasPlan := map[string]bool{}
	for _, it := range b.Items {
		if it.Kind == board.KindPlan && it.SpecID != "" {
			hasPlan[it.SpecID] = true
		}
	}
	var out []Candidate
	for _, it := range b.Items {
		if it.Legacy || it.Worktree != "" {
			continue
		}
		why, since, err := rule(cfg, it, building, hasPlan, now)
		if err != nil {
			return nil, err
		}
		if why == "" {
			continue
		}
		out = append(out, Candidate{ID: it.ID, Kind: string(it.Kind), Title: it.Title, Path: rel(cfg, it.Path),
			Status: it.Status, AgeDays: days(since, now), Why: why})
	}
	slices.SortStableFunc(out, func(a, c Candidate) int {
		if d := cmp.Compare(slices.Index(kindOrder, board.Kind(a.Kind)), slices.Index(kindOrder, board.Kind(c.Kind))); d != 0 {
			return d
		}
		return cmp.Compare(c.AgeDays, a.AgeDays)
	})
	return out, nil
}

// rule says why an item is a candidate, or "" when it is not, and the time
// its age counts from.
func rule(cfg config.Config, it *board.Item, building, hasPlan map[string]bool, now time.Time) (string, time.Time, error) {
	born := firstDate(it.Created, it.Date)
	switch {
	case it.Kind == board.KindDebtItem && it.Status == "open":
		return "open debt line", born, nil
	case it.Kind == board.KindScratch && it.Status == "raw":
		return "raw idea with no spec", born, nil
	case it.Kind == board.KindScratch && it.Status == "brainstorming":
		last, err := lastCommit(cfg.RepoRoot, it.Path, "")
		if err != nil {
			return "", time.Time{}, err
		}
		if last.IsZero() {
			last = born
		}
		if !quiet(last, now) {
			return "", time.Time{}, nil
		}
		return fmt.Sprintf("brainstorm with no spec, quiet %d days", days(last, now)), last, nil
	case it.Kind == board.KindBug && it.Status == "open":
		return "open bug", born, nil
	case it.Kind == board.KindPlan && it.Status == "in-progress" && it.Done < it.Total && !building[it.ID]:
		last, err := lastCommit(cfg.RepoRoot, it.Path, it.Hash)
		if err != nil {
			return "", time.Time{}, err
		}
		if last.IsZero() {
			last = firstDate(it.StartedOn, it.Created)
		}
		if !quiet(last, now) {
			return "", time.Time{}, nil
		}
		return fmt.Sprintf("open tasks, no commit for %d days", days(last, now)), last, nil
	case it.Kind == board.KindStory && (it.Status == "draft" || it.Status == "approved") && !hasPlan[it.ID]:
		if !quiet(born, now) {
			return "", time.Time{}, nil
		}
		return it.Status + " spec with no plan", born, nil
	}
	return "", time.Time{}, nil
}

// lastCommit gives the newest author date of a commit that touches path or,
// when hash is set, carries a Task trailer of that plan. Tick commits touch
// the plan file, so they count as work too. Zero means no such commit.
func lastCommit(repo, path, hash string) (time.Time, error) {
	newest := time.Time{}
	queries := [][]string{{"--", path}}
	if hash != "" {
		queries = append(queries, []string{"--grep=^Task: " + hash + "#[0-9]"})
	}
	for _, q := range queries {
		args := append([]string{"-C", repo, "log", "-1", "--format=%aI"}, q...)
		out, err := exec.Command("git", args...).Output()
		if err != nil {
			return time.Time{}, fmt.Errorf("git log for %s: %w", path, err)
		}
		s := strings.TrimSpace(string(out))
		if s == "" {
			continue
		}
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return time.Time{}, fmt.Errorf("git log date %q: %w", s, err)
		}
		if t.After(newest) {
			newest = t
		}
	}
	return newest, nil
}

// firstDate reads the first frontmatter or file-name date that is set. The
// board already checked each one is a real day.
func firstDate(dates ...string) time.Time {
	for _, d := range dates {
		for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02"} {
			if t, err := time.ParseInLocation(layout, d, time.Local); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

// quiet says the time is at least quietDays old. An item with no known date
// counts as quiet, so a file with no dates is still asked about.
func quiet(since, now time.Time) bool {
	return since.IsZero() || days(since, now) >= quietDays
}

func days(since, now time.Time) int {
	if since.IsZero() {
		return 0
	}
	return int(now.Sub(since).Hours() / 24)
}

func rel(cfg config.Config, path string) string {
	if r, ok := strings.CutPrefix(path, cfg.RepoRoot+"/"); ok {
		return r
	}
	return path
}
