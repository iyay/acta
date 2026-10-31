package write

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// notSlug matches every run of characters a slug cannot hold.
var notSlug = regexp.MustCompile(`[^a-z0-9]+`)

// RecordFlake writes one flake of the test called name onto the open bug
// that tracks it, and commits that bug file. With no open bug for the name a
// new flake bug is filed. A bug that was fixed is never changed: the new bug
// names it, so the fix history stays true. It gives the bug's outcome and how
// many lines its "## Flakes" section holds now. planID may be empty.
func RecordFlake(cfg config.Config, name, planID string) (Outcome, int, error) {
	if strings.TrimSpace(name) == "" {
		return Outcome{}, 0, bad("flaky test name cannot be empty")
	}
	if strings.ContainsAny(name, "\r\n") {
		return Outcome{}, 0, bad("flaky test name must be one line, not %q", name)
	}
	b, err := board.Load(cfg)
	if err != nil {
		return Outcome{}, 0, err
	}
	plan := "no plan"
	if planID != "" {
		it := b.Get(planID)
		if it == nil {
			return Outcome{}, 0, bad("unknown id %s", planID)
		}
		if it.Kind != board.KindPlan {
			return Outcome{}, 0, bad("%s is not a plan", planID)
		}
		plan = planID
	}
	line := "- " + Now().Format(stampLayout) + " " + plan
	open, fixed := flakeBugs(b, name)
	if open != nil {
		return addFlake(cfg, open, line)
	}
	return newFlakeBug(cfg, name, line, fixed)
}

// flakeBugs finds the open bug that lists name, and the last fixed one that
// does, so a new bug can point back at it.
func flakeBugs(b *board.Board, name string) (open, fixed *board.Item) {
	for _, it := range b.Items {
		if it.Kind != board.KindBug || !slices.Contains(it.FlakyTests, name) {
			continue
		}
		if it.FixedIn == "" {
			return it, nil
		}
		fixed = it
	}
	return nil, fixed
}

// addFlake puts one line in the bug's "## Flakes" section and commits the
// file. A second record of the same file folds into the last commit.
func addFlake(cfg config.Config, it *board.Item, line string) (Outcome, int, error) {
	// The git status has to be read before we write, or our own write makes
	// the file look dirty and the commit is skipped.
	dirty, err := dirtyBefore(cfg, it.Path)
	if err != nil {
		return Outcome{}, 0, err
	}
	src, err := os.ReadFile(it.Path)
	if err != nil {
		return Outcome{}, 0, err
	}
	_, body, _, _, err := frontOf(src)
	if err != nil {
		return Outcome{}, 0, err
	}
	newBody, n := withFlakeLine(body, line)
	out := strings.TrimSuffix(string(src), body) + newBody
	if err := os.WriteFile(it.Path, []byte(out), 0o644); err != nil {
		return Outcome{}, 0, err
	}
	o := finish(cfg, it.Path, Subject(cfg, []string{it.Path}, "add flake to "+stem(it.Path)), dirty)
	o.ShortID = it.ShortID
	return o, n, nil
}

// withFlakeLine adds line at the end of the "## Flakes" list of body, or adds
// the section at the end of the body when it is missing. It also says how
// many lines the list holds after the add.
func withFlakeLine(body, line string) (string, int) {
	lines := strings.Split(strings.TrimRight(body, "\n"), "\n")
	start := slices.IndexFunc(lines, func(ln string) bool { return sectionName(ln) == "Flakes" })
	if start < 0 {
		return strings.TrimRight(body, "\n") + "\n\n## Flakes\n\n" + line + "\n", 1
	}
	end := len(lines)
	for j := start + 1; j < len(lines); j++ {
		if sectionName(lines[j]) != "" {
			end = j
			break
		}
	}
	// The new line goes right under the last line of the list, so no blank
	// line splits it from the lines before it.
	at := end
	for at > start+1 && strings.TrimSpace(lines[at-1]) == "" {
		at--
	}
	n := 1
	for _, ln := range lines[start+1 : at] {
		if strings.HasPrefix(ln, "- ") {
			n++
		}
	}
	ins := []string{line}
	if at == start+1 {
		ins = []string{"", line}
	}
	if at == end && end < len(lines) {
		ins = append(ins, "")
	}
	lines = slices.Insert(lines, at, ins...)
	return strings.Join(lines, "\n") + "\n", n
}

// newFlakeBug files a flake bug for name with line as its first flake. NewBug
// runs with auto-commit off, so flaky_tests is in the file before the one
// commit is made.
func newFlakeBug(cfg config.Config, name, line string, fixed *board.Item) (Outcome, int, error) {
	symptom := name + " fails now and then and passes when run again."
	if fixed != nil {
		symptom += fmt.Sprintf("\n\n%s was fixed in %s, and the test flaked again.", bugName(fixed), fixed.FixedIn)
	}
	body := "## Symptom\n\n" + symptom + "\n\n## Root cause\n\nNot known yet.\n\n## Flakes\n\n" + line + "\n"
	slug := strings.Trim("flaky-"+notSlug.ReplaceAllString(strings.ToLower(name), "-"), "-")
	quiet := cfg
	quiet.AutoCommit = false
	o, err := NewBug(quiet, slug, "Flaky test "+name, "", "medium", []byte(body))
	if err != nil {
		return Outcome{}, 0, err
	}
	src, err := os.ReadFile(o.Path)
	if err != nil {
		return Outcome{}, 0, err
	}
	if src, err = SetList(src, "flaky_tests", []string{name}); err != nil {
		return Outcome{}, 0, err
	}
	if err := os.WriteFile(o.Path, src, 0o644); err != nil {
		return Outcome{}, 0, err
	}
	out := finish(cfg, o.Path, Subject(cfg, []string{o.Path}, "new bug "+stem(o.Path)), false)
	out.ShortID, _ = board.Parse(src).Front["id"].(string)
	return out, 1, nil
}

// bugName is the id a reader knows the bug by, or its file name when the
// bug has no id yet.
func bugName(it *board.Item) string {
	if it.ShortID != "" {
		return it.ShortID
	}
	return stem(it.Path)
}
