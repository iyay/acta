package commits

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// git runs one git command in dir with fixed dates, so the tests are stable.
func git(t *testing.T, dir, date string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date,
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func newRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "2026-01-01T00:00:00Z", "init", "-q", "-b", "main")
	git(t, dir, "2026-01-01T00:00:00Z", "config", "user.name", "Tester")
	git(t, dir, "2026-01-01T00:00:00Z", "config", "user.email", "t@example.com")
	return dir
}

// commit makes one commit with msg, touching a new file so it is never empty.
func commit(t *testing.T, dir, date, msg string) string {
	t.Helper()
	name := fmt.Sprintf("f%d.txt", time.Now().UnixNano())
	if err := os.WriteFile(filepath.Join(dir, name), []byte(msg), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, date, "add", name)
	git(t, dir, date, "commit", "-q", "-m", msg)
	return git(t, dir, date, "rev-parse", "HEAD")
}

func shas(cs []Commit) []string {
	var out []string
	for _, c := range cs {
		out = append(out, c.Sha)
	}
	return out
}

func TestTrailers(t *testing.T) {
	cases := []struct {
		name string
		body string
		want []string
	}{
		{"empty", "", nil},
		{"no trailer", "just words\n\nmore words\n", nil},
		{"one", "words\n\nTask: PLN-broaksz#3\n", []string{"broaksz#3"}},
		{"two", "words\n\nTask: PLN-broaksz#3\nTask: PLN-abc1234#10\n", []string{"broaksz#3", "abc1234#10"}},
		{"leading zero", "words\n\nTask: PLN-broaksz#01\n", []string{"broaksz#1"}},
		{"same key twice", "w\n\nTask: PLN-broaksz#1\nTask: PLN-broaksz#01\n", []string{"broaksz#1"}},
		{"middle of body", "Task: PLN-broaksz#1\n\nmore words\n", nil},
		{"middle with later paragraph", "w\n\nTask: PLN-broaksz#1\n\nlast words\n", nil},
		{"trailing blank lines", "w\n\nTask: PLN-broaksz#2\n\n\n", []string{"broaksz#2"}},
		{"no hash mark", "w\n\nTask: PLN-broaksz\n", nil},
		{"not a number", "w\n\nTask: PLN-broaksz#x\n", nil},
		{"empty hash", "w\n\nTask: PLN-#3\n", nil},
		{"upper case hash", "w\n\nTask: PLN-BROAKSZ#3\n", nil},
		{"short hash", "w\n\nTask: PLN-abc#3\n", nil},
		{"bad then good", "w\n\nTask: PLN-broaksz#x\nTask: PLN-broaksz#4\n", []string{"broaksz#4"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Trailers(tc.body); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Trailers(%q) = %v, want %v", tc.body, got, tc.want)
			}
		})
	}
}

func TestFind(t *testing.T) {
	dir := newRepo(t)
	base := commit(t, dir, "2026-01-02T00:00:00Z", "feat: base\n\nno trailer here")
	shared := commit(t, dir, "2026-01-03T00:00:00Z", "feat: shared\n\nTask: PLN-broaksz#1")
	two := commit(t, dir, "2026-01-04T00:00:00Z", "feat: two\n\nbody\n\nTask: PLN-broaksz#2\nTask: PLN-other77#1")
	commit(t, dir, "2026-01-05T00:00:00Z", "feat: middle\n\nTask: PLN-broaksz#9\n\nlast paragraph")
	chore := commit(t, dir, "2026-01-06T00:00:00Z", "chore(plan): tick\n\nTask: PLN-broaksz#1")
	commit(t, dir, "2026-01-07T00:00:00Z", "feat: bad\n\nTask: PLN-broaksz#x")
	git(t, dir, "2026-01-07T00:00:00Z", "checkout", "-q", "-b", "feat")
	onlyFeat := commit(t, dir, "2026-01-08T00:00:00Z", "feat: only on feat\n\nTask: PLN-broaksz#3")
	git(t, dir, "2026-01-08T00:00:00Z", "checkout", "-q", "main")
	_ = base

	// HEAD is main. The branch feat has everything main has plus one more.
	// A key lists its commits oldest first.
	got, err := Find(dir, []Ref{{Name: "HEAD"}, {Name: "feat"}, {Name: "no-such-branch"}})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{
		"broaksz#1": {shared, chore},
		"broaksz#2": {two},
		"other77#1": {two},
		"broaksz#3": {onlyFeat},
	}
	if len(got) != len(want) {
		t.Fatalf("keys = %v, want %d keys", got, len(want))
	}
	for key, w := range want {
		if g := shas(got[key]); !reflect.DeepEqual(g, w) {
			t.Fatalf("%s = %v, want %v", key, g, w)
		}
	}

	// The shared sha is seen from HEAD first, so its Branch is HEAD.
	if b := got["broaksz#1"][0].Branch; b != "HEAD" {
		t.Fatalf("shared commit Branch = %q, want HEAD", b)
	}
	if b := got["broaksz#3"][0].Branch; b != "feat" {
		t.Fatalf("feat-only commit Branch = %q, want feat", b)
	}
	// Chore flag follows the subject only.
	if got["broaksz#1"][0].Chore || !got["broaksz#1"][1].Chore {
		t.Fatalf("Chore flags wrong: %+v", got["broaksz#1"])
	}
	c := got["broaksz#1"][0]
	if c.Subject != "feat: shared" || !c.Date.Equal(time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("commit fields wrong: %+v", c)
	}
	// A commit naming two plans shows once under each key, not twice under one.
	for _, k := range []string{"broaksz#2", "other77#1"} {
		if len(got[k]) != 1 {
			t.Fatalf("%s has %d commits, want 1", k, len(got[k]))
		}
	}

	// The same refs in the other order flip the Branch of shared commits.
	rev, err := Find(dir, []Ref{{Name: "feat"}, {Name: "HEAD"}})
	if err != nil {
		t.Fatal(err)
	}
	if b := rev["broaksz#1"][0].Branch; b != "feat" {
		t.Fatalf("reversed Branch = %q, want feat", b)
	}
}

func TestFindOrdersByAuthorDate(t *testing.T) {
	dir := newRepo(t)
	// The later commit is made first, so git log order differs from date order.
	late := commit(t, dir, "2026-03-01T00:00:00Z", "feat: a\n\nTask: PLN-broaksz#1")
	early := commit(t, dir, "2026-02-01T00:00:00Z", "feat: b\n\nTask: PLN-broaksz#1")
	got, err := Find(dir, []Ref{{Name: "HEAD"}})
	if err != nil {
		t.Fatal(err)
	}
	if g := shas(got["broaksz#1"]); !reflect.DeepEqual(g, []string{early, late}) {
		t.Fatalf("order = %v, want %v", g, []string{early, late})
	}
}

func TestFindErrors(t *testing.T) {
	t.Run("unknown ref only", func(t *testing.T) {
		dir := newRepo(t)
		commit(t, dir, "2026-01-02T00:00:00Z", "feat: a\n\nTask: PLN-broaksz#1")
		got, err := Find(dir, []Ref{{Name: "nope"}})
		if err != nil || len(got) != 0 {
			t.Fatalf("got %v, %v; want empty map and no error", got, err)
		}
	})
	t.Run("empty repo", func(t *testing.T) {
		got, err := Find(newRepo(t), []Ref{{Name: "HEAD"}})
		if err != nil || len(got) != 0 {
			t.Fatalf("got %v, %v; want empty map and no error", got, err)
		}
	})
	t.Run("not a repo", func(t *testing.T) {
		if _, err := Find(t.TempDir(), []Ref{{Name: "HEAD"}}); err == nil {
			t.Fatal("want an error for a folder that is not a repo")
		}
	})
}

func TestDiff(t *testing.T) {
	dir := newRepo(t)
	sha := commit(t, dir, "2026-01-02T00:00:00Z", "feat: a\n\nTask: PLN-broaksz#1")
	out, err := Diff(dir, sha)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{sha, "feat: a", "diff --git", "1 file changed"} {
		if !strings.Contains(out, s) {
			t.Fatalf("Diff output lacks %q:\n%s", s, out)
		}
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatal("Diff output has colour codes")
	}
	if _, err := Diff(dir, "deadbeef"); err == nil {
		t.Fatal("want an error for an unknown sha")
	}
}
