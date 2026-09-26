package board

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/iyay/acta/internal/config"
)

// closedSpec writes one finished spec file under dir and gives its path back.
func closedSpec(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, ".acta", "specs", name+".md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\ntitle: " + name + "\nstatus: done\n---\n\nbody\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// commitFile commits one file at the moment the test picks, so which of two
// files was touched last is the test's choice and not the clock's.
func commitFile(t *testing.T, dir, when, path string) {
	t.Helper()
	gitRun(t, dir, "add", path)
	cmd := exec.Command("git", "-C", dir, "commit", "-q", "-m", "add "+filepath.Base(path))
	cmd.Env = gitEnv(when)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v %s", err, out)
	}
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = gitEnv("")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
}

// gitEnv is the environment git needs to commit here, with the date the test
// wants for the commit itself.
func gitEnv(when string) []string {
	env := append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
	if when != "" {
		env = append(env, "GIT_AUTHOR_DATE="+when, "GIT_COMMITTER_DATE="+when)
	}
	return env
}

func loadDir(t *testing.T, dir string) *Board {
	t.Helper()
	b, err := Load(config.Default(dir))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// day is the second the date in a file name stands for.
func day(t *testing.T, date string) int64 {
	t.Helper()
	ts, err := time.Parse("2006-01-02", date)
	if err != nil {
		t.Fatal(err)
	}
	return ts.Unix()
}

// moment is the second a commit made at that time stands for.
func moment(t *testing.T, s string) int64 {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	return ts.Unix()
}

// TestClosedItemsSortByTheLastCommit is the rule of pane [2]: the newest commit
// on the file comes first, and a file git has no commit for falls back to the
// date in its name. Nothing here errors when there is no repo at all.
func TestClosedItemsSortByTheLastCommit(t *testing.T) {
	t.Run("the last commit beats the file name", func(t *testing.T) {
		dir := t.TempDir()
		gitRun(t, dir, "init", "-q", "-b", "main")
		early := closedSpec(t, dir, "2026-09-01-early")
		late := closedSpec(t, dir, "2026-09-20-late")
		commitFile(t, dir, "2026-09-20T10:00:00+07:00", late)
		commitFile(t, dir, "2026-09-27T10:00:00+07:00", early)

		b := loadDir(t, dir)
		got, other := b.Get("specs/2026-09-01-early"), b.Get("specs/2026-09-20-late")
		if want := moment(t, "2026-09-27T10:00:00+07:00"); got.SortTime() != want {
			t.Fatalf("the 09-01 file was last committed at %d, want %d", got.SortTime(), want)
		}
		if got.SortTime() <= other.SortTime() {
			t.Fatalf("the 09-01 file (%d) should come before the 09-20 one (%d)", got.SortTime(), other.SortTime())
		}
	})

	t.Run("the file name decides when git has no commit", func(t *testing.T) {
		dir := t.TempDir()
		gitRun(t, dir, "init", "-q", "-b", "main")
		closedSpec(t, dir, "2026-09-01-early")
		closedSpec(t, dir, "2026-09-20-late")

		b := loadDir(t, dir)
		got := b.Get("specs/2026-09-20-late")
		if want := day(t, "2026-09-20"); got.SortTime() != want {
			t.Fatalf("a file with no commit was placed at %d, want its file-name date %d", got.SortTime(), want)
		}
		if other := b.Get("specs/2026-09-01-early"); other.SortTime() != day(t, "2026-09-01") {
			t.Fatalf("the other file was placed at %d", other.SortTime())
		}
	})

	t.Run("a commit and a file name compare on the same line", func(t *testing.T) {
		dir := t.TempDir()
		gitRun(t, dir, "init", "-q", "-b", "main")
		closedSpec(t, dir, "2026-09-20-late")
		touched := closedSpec(t, dir, "2026-09-01-early")
		commitFile(t, dir, "2026-09-27T10:00:00+07:00", touched)

		b := loadDir(t, dir)
		if b.Get("specs/2026-09-01-early").SortTime() <= b.Get("specs/2026-09-20-late").SortTime() {
			t.Fatalf("a commit from 09-27 should beat a file name of 09-20")
		}

		old := closedSpec(t, dir, "2026-09-05-old")
		commitFile(t, dir, "2026-09-10T10:00:00+07:00", old)
		b = loadDir(t, dir)
		if b.Get("specs/2026-09-20-late").SortTime() <= b.Get("specs/2026-09-05-old").SortTime() {
			t.Fatalf("a file name of 09-20 should beat a commit from 09-10")
		}
	})

	t.Run("a folder with no git falls back to the file name", func(t *testing.T) {
		dir := t.TempDir()
		closedSpec(t, dir, "2026-09-20-late")
		closedSpec(t, dir, "2026-09-01-early")

		b := loadDir(t, dir)
		got := b.Get("specs/2026-09-20-late")
		if got == nil || got.SortTime() != day(t, "2026-09-20") {
			t.Fatalf("a board outside git did not fall back to the file name: %v", got)
		}
	})
}
