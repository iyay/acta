package cleanup

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/gitc"
	"github.com/iyay/acta/internal/write"
)

// Decision is one approved close, read from one JSON line of stdin.
type Decision struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	FixedIn string `json:"fixed_in,omitempty"`
	line    int
}

// closes lists the statuses apply may write for each kind. Keep is no write,
// so it is on no list.
var closes = map[board.Kind][]string{
	board.KindDebtItem: {"done", "wontfix"},
	board.KindScratch:  {"done", "dropped"},
	board.KindBug:      {"fixed", "wontfix"},
	board.KindPlan:     {"dropped"},
	board.KindStory:    {"dropped"},
}

var shaRe = regexp.MustCompile(`^[0-9a-f]{7,40}$`)

// ParseDecisions reads one JSON object per line. Blank lines are skipped. A
// line that does not parse, or names a field apply does not know, is an error
// that names the line.
func ParseDecisions(r io.Reader) ([]Decision, error) {
	var out []Decision
	var errs []error
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for n := 1; sc.Scan(); n++ {
		text := strings.TrimSpace(sc.Text())
		if text == "" {
			continue
		}
		dec := json.NewDecoder(strings.NewReader(text))
		dec.DisallowUnknownFields()
		var d Decision
		if err := dec.Decode(&d); err != nil {
			errs = append(errs, fmt.Errorf("line %d: %v", n, err))
			continue
		}
		// Decode stops after the first object. Anything after it would be
		// dropped without a word, so it is an error.
		if int(dec.InputOffset()) != len(text) {
			errs = append(errs, fmt.Errorf("line %d: text after the JSON object", n))
			continue
		}
		d.line = n
		out = append(out, d)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, errors.Join(errs...)
}

// Apply checks every decision, then writes them all and commits them as one
// commit. One failed check stops it before the first write. Dropping a plan
// also drops its spec when no other plan of that spec is still open.
func Apply(cfg config.Config, b *board.Board, ds []Decision) (write.Outcome, error) {
	if len(ds) == 0 {
		return write.Outcome{}, fmt.Errorf("no decisions on stdin")
	}
	building := map[string]bool{}
	for _, r := range board.PlanStates(cfg) {
		building[r.Plan.ID] = true
	}
	var errs []error
	seen := map[string]bool{}
	for _, d := range ds {
		it := b.Get(d.ID)
		if err := check(it, d, building); err != nil {
			errs = append(errs, fmt.Errorf("line %d: %s: %v", d.line, d.ID, err))
			continue
		}
		if seen[it.ID] {
			errs = append(errs, fmt.Errorf("line %d: %s is in the batch twice", d.line, d.ID))
		}
		seen[it.ID] = true
	}
	if err := errors.Join(errs...); err != nil {
		return write.Outcome{}, err
	}
	all := withSpecs(b, ds)
	files, err := render(cfg, b, all)
	if err != nil {
		return write.Outcome{}, err
	}
	paths := make([]string, 0, len(files))
	for path, out := range files {
		if err := os.WriteFile(path, out, 0o644); err != nil {
			return write.Outcome{}, err
		}
		paths = append(paths, path)
	}
	slices.Sort(paths)
	return commit(cfg, paths, fmt.Sprintf("cleanup apply, %d items", len(all)), write.Outcome{Path: cfg.Root}), nil
}

// check says why one decision cannot be written, or nil.
func check(it *board.Item, d Decision, building map[string]bool) error {
	switch {
	case it == nil:
		return fmt.Errorf("unknown id")
	case it.Legacy || it.Worktree != "":
		return fmt.Errorf("lives outside the main planning folder")
	case !slices.Contains(closes[it.Kind], d.Status):
		return fmt.Errorf("status %q is not a close for a %s (%s)", d.Status, it.Kind, strings.Join(closes[it.Kind], ", "))
	case building[it.ID]:
		return fmt.Errorf("a linked worktree is building this plan; change it there")
	case d.Status == "fixed" && !shaRe.MatchString(d.FixedIn):
		return fmt.Errorf("fixed needs fixed_in, a commit sha of 7 to 40 lower-case hex characters")
	case d.Status != "fixed" && d.FixedIn != "":
		return fmt.Errorf("fixed_in goes only with fixed")
	}
	if strings.Trim(d.Reason, "() \t") == "" {
		return fmt.Errorf("reason is required")
	}
	return write.ValidReason(d.Reason)
}

// withSpecs adds a drop for the spec of each dropped plan, unless the spec
// is already in the batch or another plan of it is still open.
func withSpecs(b *board.Board, ds []Decision) []Decision {
	dropped := map[string]bool{}
	named := map[string]bool{}
	for _, d := range ds {
		it := b.Get(d.ID)
		named[it.ID] = true
		if it.Kind == board.KindPlan {
			dropped[it.ID] = true
		}
	}
	out := slices.Clone(ds)
	for _, d := range ds {
		plan := b.Get(d.ID)
		spec := b.Get(plan.SpecID)
		if plan.Kind != board.KindPlan || spec == nil || spec.Kind != board.KindStory || named[spec.ID] || board.Closed(spec.Status) {
			continue
		}
		// A legacy or worktree spec is not ours to write from here, the same
		// as check refuses it when it is named.
		if spec.Legacy || spec.Worktree != "" {
			continue
		}
		if otherOpenPlan(b, spec.ID, dropped) {
			continue
		}
		named[spec.ID] = true
		out = append(out, Decision{ID: spec.ID, Status: "dropped", Reason: "plan " + shortName(plan) + " dropped: " + d.Reason, line: d.line})
	}
	return out
}

func otherOpenPlan(b *board.Board, specID string, dropped map[string]bool) bool {
	for _, it := range b.Items {
		if it.Kind == board.KindPlan && it.SpecID == specID && !dropped[it.ID] && it.Status != "done" && it.Status != "dropped" {
			return true
		}
	}
	return false
}

func shortName(it *board.Item) string {
	if it.ShortID != "" {
		return it.ShortID
	}
	return it.ID
}

// render works out the new text of every file the batch touches. Several
// debt lines of one file are closed in the same text, so it is saved once.
func render(cfg config.Config, b *board.Board, ds []Decision) (map[string][]byte, error) {
	files := map[string][]byte{}
	for _, d := range ds {
		it := b.Get(d.ID)
		src, ok := files[it.Path]
		if !ok {
			if err := cleanFile(cfg, it.Path); err != nil {
				return nil, err
			}
			var err error
			if src, err = os.ReadFile(it.Path); err != nil {
				return nil, err
			}
		}
		out, err := closeItem(src, it, d)
		if err != nil {
			return nil, fmt.Errorf("line %d: %s: %v", d.line, d.ID, err)
		}
		files[it.Path] = out
	}
	return files, nil
}

// cleanFile refuses a file with changes the user has not committed, because
// the batch commit would sweep them in.
func cleanFile(cfg config.Config, path string) error {
	dirty, err := gitc.IsDirty(cfg.RepoRoot, path)
	if err != nil {
		return err
	}
	if dirty {
		return fmt.Errorf("%s has changes that are not committed; commit or undo them first", rel(cfg, path))
	}
	return nil
}

// closeItem writes one decision into the text of the item's file.
func closeItem(src []byte, it *board.Item, d Decision) ([]byte, error) {
	if it.Kind == board.KindDebtItem {
		return write.CloseDebtLineIn(src, it.Line, d.Status, d.Reason)
	}
	out, err := write.SetField(src, "status", d.Status)
	if err == nil {
		out, err = write.SetField(out, "reason", strings.TrimSpace(d.Reason))
	}
	if err == nil && d.FixedIn != "" {
		out, err = write.SetField(out, "fixed_in", d.FixedIn)
	}
	if err == nil {
		out, err = write.DatesFor(out, d.Status)
	}
	return out, err
}
