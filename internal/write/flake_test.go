package write

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

const flakyName = "internal/tui.TestClick"

// openFlakeBugs gives every bug with no fixed_in that lists name, so a test
// can check there is never more than one.
func openFlakeBugs(t *testing.T, cfg config.Config, name string) []*board.Item {
	t.Helper()
	var out []*board.Item
	for _, it := range mustLoad(t, cfg).Items {
		if it.Kind == board.KindBug && it.FixedIn == "" && slices.Contains(it.FlakyTests, name) {
			out = append(out, it)
		}
	}
	return out
}

// flakeRepo is the base repo plus the bug files a case starts from.
func flakeRepo(t *testing.T, bugs map[string]string) config.Config {
	t.Helper()
	files := map[string]string{}
	for p, body := range baseFiles {
		files[p] = body
	}
	for p, body := range bugs {
		files[p] = body
	}
	return repoWith(t, files)
}

func TestRecordFlake(t *testing.T) {
	const planID = "plans/2026-09-25-crash-fix"
	const line = "- 2026-09-26 10:00:00 "

	t.Run("open bug listing the name gets one line", func(t *testing.T) {
		fixNow(t)
		rel := ".acta/bugs/2026-09-20-click.md"
		cfg := flakeRepo(t, map[string]string{rel: "---\nid: BUG-0005\nflaky_tests: [" + flakyName + "]\n---\n# Flaky click\n\n## Symptom\nIt flakes.\n\n## Flakes\n\n- 2026-09-20 09:00:00 no plan\n\n## Notes\nKeep.\n"})
		before := commitCount(t, cfg)
		o, n, err := RecordFlake(cfg, flakyName, planID)
		if err != nil || !o.Committed || n != 2 {
			t.Fatalf("outcome %+v count %d err %v", o, n, err)
		}
		if o.ShortID != "BUG-0005" {
			t.Errorf("short id %q want BUG-0005", o.ShortID)
		}
		want := "## Flakes\n\n- 2026-09-20 09:00:00 no plan\n" + line + planID + "\n\n## Notes\nKeep.\n"
		if got := readFile(t, filepath.Join(cfg.RepoRoot, rel)); !strings.HasSuffix(got, want) {
			t.Errorf("file:\n%s\nwant suffix:\n%s", got, want)
		}
		if got := commitCount(t, cfg); got != before+1 {
			t.Errorf("commits %d want %d", got, before+1)
		}
		if bugs := openFlakeBugs(t, cfg, flakyName); len(bugs) != 1 {
			t.Errorf("open bugs %d want 1", len(bugs))
		}
	})

	t.Run("open bug without a Flakes section gets one at the end", func(t *testing.T) {
		fixNow(t)
		rel := ".acta/bugs/2026-09-20-click.md"
		cfg := flakeRepo(t, map[string]string{rel: "---\nflaky_tests:\n  - other.Test\n  - " + flakyName + "\n---\n# Flaky click\n\n## Symptom\nIt flakes.\n"})
		_, n, err := RecordFlake(cfg, flakyName, "")
		if err != nil || n != 1 {
			t.Fatalf("count %d err %v", n, err)
		}
		want := "## Symptom\nIt flakes.\n\n## Flakes\n\n" + line + "no plan\n"
		if got := readFile(t, filepath.Join(cfg.RepoRoot, rel)); !strings.HasSuffix(got, want) {
			t.Errorf("file:\n%s\nwant suffix:\n%s", got, want)
		}
	})

	t.Run("no bug files a new flake bug in one commit", func(t *testing.T) {
		fixNow(t)
		cfg := flakeRepo(t, nil)
		before := commitCount(t, cfg)
		o, n, err := RecordFlake(cfg, flakyName, planID)
		if err != nil || !o.Committed || n != 1 {
			t.Fatalf("outcome %+v count %d err %v", o, n, err)
		}
		if got := commitCount(t, cfg); got != before+1 {
			t.Errorf("commits %d want %d", got, before+1)
		}
		if s := gitRun(t, cfg.RepoRoot, "status", "--porcelain"); s != "" {
			t.Errorf("tree left dirty: %s", s)
		}
		bugs := openFlakeBugs(t, cfg, flakyName)
		if len(bugs) != 1 {
			t.Fatalf("open bugs %d want 1", len(bugs))
		}
		it := bugs[0]
		if it.Path != o.Path || it.ShortID == "" || it.ShortID != o.ShortID {
			t.Errorf("bug %s %s, outcome %s %s", it.Path, it.ShortID, o.Path, o.ShortID)
		}
		if it.Title != "Flaky test "+flakyName || it.Priority != "medium" {
			t.Errorf("title %q priority %q", it.Title, it.Priority)
		}
		body := readFile(t, it.Path)
		for _, want := range []string{"## Symptom\n", "## Root cause\n\nNot known yet.\n", "## Flakes\n\n" + line + planID + "\n"} {
			if !strings.Contains(body, want) {
				t.Errorf("body lacks %q:\n%s", want, body)
			}
		}
	})

	t.Run("fixed bug stays and a new bug names it", func(t *testing.T) {
		fixNow(t)
		rel := ".acta/bugs/2026-09-20-click.md"
		old := "---\nid: BUG-0005\nfixed_in: abc1234\nflaky_tests: [" + flakyName + "]\n---\n# Flaky click\n\n## Symptom\nIt flakes.\n\n## Flakes\n\n- 2026-09-20 09:00:00 no plan\n"
		cfg := flakeRepo(t, map[string]string{rel: old})
		o, n, err := RecordFlake(cfg, flakyName, "")
		if err != nil || n != 1 {
			t.Fatalf("outcome %+v count %d err %v", o, n, err)
		}
		if got := readFile(t, filepath.Join(cfg.RepoRoot, rel)); got != old {
			t.Errorf("old bug changed:\n%s", got)
		}
		bugs := openFlakeBugs(t, cfg, flakyName)
		if len(bugs) != 1 || bugs[0].Path != o.Path {
			t.Fatalf("open bugs %v want only %s", bugs, o.Path)
		}
		if body := readFile(t, o.Path); !strings.Contains(body, "BUG-0005") {
			t.Errorf("new bug does not name BUG-0005:\n%s", body)
		}
	})

	t.Run("repeat record folds into the last commit", func(t *testing.T) {
		fixNow(t)
		cfg := flakeRepo(t, nil)
		before := commitCount(t, cfg)
		first, _, err := RecordFlake(cfg, flakyName, "")
		if err != nil {
			t.Fatal(err)
		}
		again, n, err := RecordFlake(cfg, flakyName, planID)
		if err != nil || !again.Committed || n != 2 {
			t.Fatalf("outcome %+v count %d err %v", again, n, err)
		}
		if again.Path != first.Path {
			t.Errorf("second record wrote %s, want %s", again.Path, first.Path)
		}
		if got := commitCount(t, cfg); got != before+1 {
			t.Errorf("commits %d want %d", got, before+1)
		}
		if bugs := openFlakeBugs(t, cfg, flakyName); len(bugs) != 1 {
			t.Errorf("open bugs %d want 1", len(bugs))
		}
	})

	t.Run("bad input writes nothing", func(t *testing.T) {
		fixNow(t)
		cfg := flakeRepo(t, nil)
		for _, c := range []struct{ name, plan, want string }{
			{"", "", "flaky test name cannot be empty"},
			{"  ", "", "flaky test name cannot be empty"},
			{"a\nb", "", `flaky test name must be one line, not "a\nb"`},
			{flakyName, "PLN-nope", "unknown id PLN-nope"},
			{flakyName, "bugs/2026-09-24-crash", "bugs/2026-09-24-crash is not a plan"},
		} {
			refuse(t, cfg, c.want, func() error {
				_, _, err := RecordFlake(cfg, c.name, c.plan)
				return err
			})
		}
		if entries, _ := os.ReadDir(filepath.Join(cfg.Root, "bugs")); len(entries) != 1 {
			t.Errorf("bugs folder holds %d files want 1", len(entries))
		}
	})
}
