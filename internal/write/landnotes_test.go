package write

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// landPlan is a done plan with hash abcd123, a start stamp, the given parent
// and the given closes list.
func landPlan(parent, closes string) string {
	front := "id: PLN-0001\nhash: abcd123\nstarted: \"2026-09-20 09:00:00\"\n"
	if parent != "" {
		front += "parent: " + parent + "\n"
	}
	if closes != "" {
		front += "closes: [" + closes + "]\n"
	}
	return "---\n" + front + "---\n# Plan\n\n### Task 1: Do it\n- [x] a\n"
}

// LandNotes writes every note that closes a plan, and only those: fixed_in on
// its bugs, ticks on its debt lines, and its spec dates. A closes id with no
// file stops it before any file changes.
func TestLandNotes(t *testing.T) {
	const (
		plan  = ".acta/plans/2026-09-20-fix.md"
		spec  = ".acta/specs/2026-09-19-s.md"
		bug   = ".acta/bugs/2026-09-18-crash.md"
		bug2  = ".acta/bugs/2026-09-18-hang.md"
		debt  = ".acta/debt/2026-09-17-a.md"
		debt2 = ".acta/debt/2026-09-17-b.md"
	)
	bugs := map[string]string{
		bug:  "---\nid: BUG-0001\n---\n# Crash\n",
		bug2: "---\nid: BUG-0002\n---\n# Hang\n",
	}
	debts := map[string]string{
		debt:  "---\nid: DBT-0001\n---\n# A\n\n- [ ] (low) one\n- [-] two\n- [ ] three\n",
		debt2: "---\nid: DBT-0002\n---\n# B\n\n- [ ] four\n- [ ] five\n",
	}
	cases := []struct {
		name    string
		files   map[string]string
		wantErr string
		// want holds a text each changed file must hold; every other file
		// stays as it was.
		want map[string][]string
	}{
		{
			name:  "bug parent",
			files: map[string]string{plan: landPlan("bugs/2026-09-18-crash", "")},
			want:  map[string][]string{bug: {"fixed_in: PLN-abcd123", "finished: \"2026-09-26 10:00:00\""}},
		},
		{
			name:  "bug in closes",
			files: map[string]string{plan: landPlan("", "BUG-0002"), spec: "---\nid: SPC-0001\n---\n# S\n"},
			want:  map[string][]string{bug2: {"fixed_in: PLN-abcd123"}},
		},
		{
			name:  "debt in closes",
			files: map[string]string{plan: landPlan("", "DBT-0001, DBT-0002.02")},
			want: map[string][]string{
				debt:  {"- [x] (low) one\n- [-] two\n- [x] three\n"},
				debt2: {"- [ ] four\n- [x] five\n"},
			},
		},
		{
			name: "sibling plan already done",
			files: map[string]string{
				plan:                            landPlan("specs/2026-09-19-s", ""),
				".acta/plans/2026-09-19-one.md": "---\nid: PLN-0002\nparent: specs/2026-09-19-s\nstarted: \"2026-09-19 08:00:00\"\n---\n# One\n\n### Task 1: A\n- [x] a\n",
				spec:                            "---\nid: SPC-0001\n---\n# S\n",
			},
			want: map[string][]string{spec: {"started: \"2026-09-19 08:00:00\"", "finished: \"2026-09-26 10:00:00\""}},
		},
		{
			name:  "plan with no spec",
			files: map[string]string{plan: landPlan("", "")},
		},
		{
			name:  "plan with nothing to close",
			files: map[string]string{plan: landPlan("specs/2026-09-19-s", ""), spec: "---\nid: SPC-0001\n---\n# S\n"},
			want:  map[string][]string{spec: {"started: \"2026-09-20 09:00:00\"", "finished: \"2026-09-26 10:00:00\""}},
		},
		{
			name:    "missing id",
			files:   map[string]string{plan: landPlan("bugs/2026-09-18-crash", "BUG-0001, DBT-0009")},
			wantErr: "DBT-0009",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fixNow(t)
			files := map[string]string{}
			for _, m := range []map[string]string{bugs, debts, c.files} {
				for k, v := range m {
					files[k] = v
				}
			}
			cfg := repoWith(t, files)
			paths, err := LandNotes(cfg, "abcd123")
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("err = %v, want one naming %s", err, c.wantErr)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			var got, want []string
			for _, p := range paths {
				rel, _ := filepath.Rel(cfg.RepoRoot, p)
				got = append(got, rel)
			}
			for rel, texts := range c.want {
				want = append(want, rel)
				body, err := os.ReadFile(filepath.Join(cfg.RepoRoot, rel))
				if err != nil {
					t.Fatal(err)
				}
				for _, text := range texts {
					if !strings.Contains(string(body), text) {
						t.Errorf("%s lacks %q:\n%s", rel, text, body)
					}
				}
			}
			slices.Sort(got)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Errorf("changed paths = %v, want %v", got, want)
			}
			// The disk agrees with the list: no file changed that it leaves out.
			var dirty []string
			for _, ln := range strings.Split(strings.TrimSpace(gitRun(t, cfg.RepoRoot, "status", "--porcelain")), "\n") {
				if ln != "" {
					dirty = append(dirty, strings.TrimSpace(ln[2:]))
				}
			}
			slices.Sort(dirty)
			if !slices.Equal(dirty, want) {
				t.Errorf("files on disk changed = %v, want %v", dirty, want)
			}
			// LandNotes never commits.
			if n := gitRun(t, cfg.RepoRoot, "rev-list", "--count", "HEAD"); strings.TrimSpace(n) != "1" {
				t.Errorf("commit count = %s, want 1", n)
			}
		})
	}
}
