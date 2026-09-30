package write

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const plan = "# P\n\n### Task 1: One\n- [x] a\n- [ ] b\n```text\n- [ ] in a fence\n```\n- [ ] c\n\n### Task 2: Two\n- [ ] d\n"

func TestTickText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		line, step  int
		want        string
		done, total int
	}{
		{"second box", 3, 2, strings.Replace(plan, "- [ ] b", "- [x] b", 1), 2, 3},
		{"third box skips the fence", 3, 3, strings.Replace(plan, "- [ ] c", "- [x] c", 1), 2, 3},
		{"already ticked", 3, 1, plan, 1, 3},
		{"all boxes of task 1 only", 3, 0, strings.Replace(strings.Replace(plan, "- [ ] b", "- [x] b", 1), "- [ ] c", "- [x] c", 1), 3, 3},
		{"task 2", 11, 1, strings.Replace(plan, "- [ ] d", "- [x] d", 1), 1, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, done, total, err := TickText([]byte(plan), c.line, c.step)
			if err != nil || string(out) != c.want || done != c.done || total != c.total {
				t.Fatalf("got %q %d/%d %v", out, done, total, err)
			}
		})
	}
}

func TestTickTextCRLF(t *testing.T) {
	t.Parallel()

	src := strings.ReplaceAll(plan, "\n", "\r\n")
	out, done, _, err := TickText([]byte(src), 3, 2)
	if err != nil || done != 2 || string(out) != strings.Replace(src, "- [ ] b", "- [x] b", 1) {
		t.Fatalf("CRLF: %q %v", out, err)
	}
}

func TestTickTextMixedLineEndings(t *testing.T) {
	t.Parallel()

	// Every layout the board parser accepts. The board turns \r\n into \n
	// and splits on \n, so task 1's heading is line 3 and its boxes are
	// a1 then a2 whichever endings each line keeps.
	body := []string{"# Plan", "", "### Task 1: a", "- [ ] a1", "- [ ] a2", "", "### Task 2: b", "- [ ] b1", ""}
	layouts := []struct {
		name string
		join func([]string) string
	}{
		{"all LF", func(ls []string) string { return strings.Join(ls, "\n") }},
		{"all CRLF", func(ls []string) string { return strings.Join(ls, "\r\n") }},
		{"LF frontmatter over CRLF body", func(ls []string) string {
			return "---\nstatus: open\n---\n" + strings.Join(ls, "\r\n")
		}},
		{"CRLF frontmatter over LF body", func(ls []string) string {
			return "---\r\nstatus: open\r\n---\r\n" + strings.Join(ls, "\n")
		}},
		{"alternating lines", func(ls []string) string {
			out := ""
			for i, l := range ls {
				if i > 0 {
					out += "\n"
				}
				out += l
				if i%2 == 0 {
					out += "\r"
				}
			}
			return out
		}},
		{"no trailing newline", func(ls []string) string {
			return strings.TrimSuffix(strings.Join(ls, "\n"), "\n")
		}},
	}
	for _, l := range layouts {
		t.Run(l.name, func(t *testing.T) {
			src := l.join(body)
			heading := strings.Count(src[:strings.Index(src, "### Task 1: a")], "\n") + 1
			out, done, total, err := TickText([]byte(src), heading, 2)
			if err != nil {
				t.Fatalf("TickText: %v", err)
			}
			if done != 1 || total != 2 {
				t.Fatalf("progress = %d/%d, want 1/2", done, total)
			}
			if !strings.Contains(string(out), "- [x] a2") || strings.Contains(string(out), "- [x] a1") || strings.Contains(string(out), "- [x] b1") {
				t.Fatalf("wrong box ticked:\n%s", out)
			}
			// Only the ticked line may change: every line keeps its own ending.
			want := strings.Replace(src, "- [ ] a2", "- [x] a2", 1)
			if string(out) != want {
				t.Fatalf("got %q want %q", out, want)
			}
		})
	}
}

func TestTickTextBadInput(t *testing.T) {
	t.Parallel()

	for name, c := range map[string]struct{ line, step int }{
		"step too big":       {3, 4},
		"negative step":      {3, -1},
		"line is not a task": {1, 1},
		"line past the end":  {99, 1},
	} {
		if _, _, _, err := TickText([]byte(plan), c.line, c.step); !errors.Is(err, ErrBadInput) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
	if _, _, _, err := TickText([]byte("# P\n\n### Task 1: Empty\ntext\n"), 3, 1); !errors.Is(err, ErrBadInput) {
		t.Errorf("a task with no boxes: err = %v", err)
	}
}

func TestUntickTextClearsEveryBoxOfTheTaskOnly(t *testing.T) {
	t.Parallel()

	src := []byte("# P\n\n### Task 1: A\n- [x] a\n- [x] b\n\n### Task 2: B\n- [x] c\n")
	out, done, total, err := UntickText(src, 3)
	if err != nil {
		t.Fatal(err)
	}
	want := "# P\n\n### Task 1: A\n- [ ] a\n- [ ] b\n\n### Task 2: B\n- [x] c\n"
	if string(out) != want || done != 0 || total != 2 {
		t.Fatalf("got %q %d/%d", out, done, total)
	}
}

// The plan constant hides a box inside a code fence and puts a second task
// after it. Untick clears the boxes of the one task, so the fenced line and
// the next task keep their ticks.
func TestUntickTextLeavesTheFenceAndTheNextTaskAlone(t *testing.T) {
	t.Parallel()

	out, done, total, err := UntickText([]byte(plan), 3)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Replace(plan, "- [x] a", "- [ ] a", 1)
	if string(out) != want || done != 0 || total != 3 {
		t.Fatalf("got %q %d/%d want %q", out, done, total, want)
	}
}

func TestTickLineSetsOnlyThatBox(t *testing.T) {
	useLockBase(t)
	path := filepath.Join(t.TempDir(), "d.md")
	os.WriteFile(path, []byte("# R\n\n- [ ] a\n- [ ] b\n- [x] c\n"), 0o644)
	if err := TickLine(path, 4, '-'); err != nil {
		t.Fatal(err)
	}
	if err := TickLine(path, 5, 'x'); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "# R\n\n- [ ] a\n- [-] b\n- [x] c\n" {
		t.Fatalf("got %q", got)
	}
}

func TestTickLineRejectsNonBox(t *testing.T) {
	useLockBase(t)
	path := filepath.Join(t.TempDir(), "d.md")
	os.WriteFile(path, []byte("# R\n\ntext\n"), 0o644)
	if err := TickLine(path, 3, 'x'); err == nil {
		t.Fatal("want error on a line with no box")
	}
	got, _ := os.ReadFile(path)
	if string(got) != "# R\n\ntext\n" {
		t.Fatalf("file changed: %q", got)
	}
}

func TestTickFileConcurrent(t *testing.T) {
	useLockBase(t)
	p := filepath.Join(t.TempDir(), "plan.md")
	if err := os.WriteFile(p, []byte(plan), 0o644); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for _, step := range []int{2, 3} {
		wg.Add(1)
		go func(step int) {
			defer wg.Done()
			if _, _, err := Tick(p, 3, step); err != nil {
				t.Error(err)
			}
		}(step)
	}
	wg.Wait()
	b, _ := os.ReadFile(p)
	if strings.Count(string(b), "- [x]") != 3 {
		t.Fatalf("a tick was lost:\n%s", b)
	}
	if ents, _ := os.ReadDir(filepath.Dir(p)); len(ents) != 1 {
		t.Fatalf("tick left files next to the plan: %v", ents)
	}
}

// TestHelperHoldLock is not a real test. TestTickAfterHolderDies runs it
// in a child process to take the lock and then get killed.
func TestHelperHoldLock(t *testing.T) {
	p := os.Getenv("PMB_HOLD_LOCK")
	if p == "" {
		t.Skip("helper for TestTickAfterHolderDies")
	}
	if root := os.Getenv("PMB_LOCK_ROOT"); root != "" {
		lockRoot = root
	}
	if _, err := lock(p); err != nil {
		t.Fatal(err)
	}
	os.Stdout.WriteString("held\n")
	// Wait to be killed. A bare select{} trips Go's deadlock check and the
	// child would crash before the parent kills it.
	time.Sleep(time.Minute)
}

func TestTickAfterHolderDies(t *testing.T) {
	useLockBase(t)
	p := filepath.Join(t.TempDir(), "plan.md")
	if err := os.WriteFile(p, []byte(plan), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperHoldLock$")
	cmd.Env = append(os.Environ(), "PMB_HOLD_LOCK="+p, "PMB_LOCK_ROOT="+lockRoot)
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 5)
	if _, err := io.ReadFull(out, buf); err != nil || string(buf) != "held\n" {
		t.Fatalf("helper did not take the lock: %q %v", buf, err)
	}
	cmd.Process.Kill()
	cmd.Wait()

	start := time.Now()
	if _, _, err := Tick(p, 3, 2); err != nil {
		t.Fatalf("a dead holder blocked the tick: %v", err)
	}
	if d := time.Since(start); d > time.Second {
		t.Fatalf("tick waited %v after the holder died", d)
	}
}

func TestLockOneHolder(t *testing.T) {
	useLockBase(t)
	p := filepath.Join(t.TempDir(), "plan.md")
	if err := os.WriteFile(p, []byte(plan), 0o644); err != nil {
		t.Fatal(err)
	}
	// A lock file left by the old version, old enough to look stale.
	// It must not let anyone skip the lock.
	old := time.Now().Add(-2 * time.Minute)
	os.WriteFile(p+".lock", nil, 0o644)
	os.Chtimes(p+".lock", old, old)

	var holders, most atomic.Int32
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unlock, err := lock(p)
			if err != nil {
				t.Error(err)
				return
			}
			n := holders.Add(1)
			for {
				m := most.Load()
				if n <= m || most.CompareAndSwap(m, n) {
					break
				}
			}
			// Hold the lock a moment, so a lock that does nothing lets
			// several holders overlap and the test goes red.
			time.Sleep(5 * time.Millisecond)
			holders.Add(-1)
			unlock()
		}()
	}
	wg.Wait()
	if most.Load() != 1 {
		t.Fatalf("%d holders at once, want 1", most.Load())
	}
}

func TestLockTimeout(t *testing.T) {
	useLockBase(t)
	p := filepath.Join(t.TempDir(), "plan.md")
	unlock, err := lock(p)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	if _, err := lock(p); err == nil || !strings.Contains(err.Error(), p+" is locked by another acta command") {
		t.Fatalf("second lock: err = %v, want the locked message for %s", err, p)
	}
}

func TestLockPathIgnoresTMPDIR(t *testing.T) {
	want := useLockBase(t)
	p := filepath.Join(t.TempDir(), "plan.md")
	t.Setenv("TMPDIR", t.TempDir())
	a, err := lockPath(p)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", t.TempDir())
	b, err := lockPath(p)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("lock path follows TMPDIR: %s vs %s", a, b)
	}
	if filepath.Dir(a) != want {
		t.Fatalf("lock folder = %s, want %s", filepath.Dir(a), want)
	}
}

// useLockBase points the lock folder at a fresh temp folder for one test,
// so no test ever writes a lock into the real cache folder.
func useLockBase(t *testing.T) string {
	old := lockRoot
	lockRoot = filepath.Join(t.TempDir(), "locks")
	t.Cleanup(func() { lockRoot = old })
	return lockRoot
}

// useRealCache sends the lock folder back to the user's own cache folder
// for one test, because those tests only read where that folder would be.
func useRealCache(t *testing.T) {
	old := lockRoot
	lockRoot = ""
	t.Cleanup(func() { lockRoot = old })
}

// lockFiles counts the lock files under root, so a test can prove a refused
// folder never got one.
func lockFiles(t *testing.T, root string) int {
	t.Helper()
	n := 0
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".lock") {
			n++
		}
		return nil
	})
	return n
}

var lockNameRe = regexp.MustCompile(`^acta-[0-9a-f]{16}\.lock$`)

func TestLockDirIsInTheCacheFolder(t *testing.T) {
	useRealCache(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, "cache"))
	dir, err := lockDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(dir, filepath.Join("acta", "locks")) {
		t.Fatalf("lock folder = %s, want it to end in acta/locks", dir)
	}
	if !strings.HasPrefix(dir, home+string(os.PathSeparator)) {
		t.Fatalf("lock folder = %s, want it inside the user's own home %s", dir, home)
	}
	// The first run finds no cache folder at all, so lock has to make the
	// folders itself, closed to everyone else.
	unlock, err := lock(filepath.Join(t.TempDir(), "plan.md"))
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	fi, err := os.Lstat(filepath.Dir(dir))
	if err != nil {
		t.Fatal(err)
	}
	if !fi.IsDir() || fi.Mode().Perm() != 0o700 {
		t.Fatalf("acta folder mode = %v, want a 0700 folder", fi.Mode())
	}
	if lockFiles(t, dir) != 1 {
		t.Fatalf("the lock file is not inside %s", dir)
	}
}

func TestLockFailsWithNoCacheFolder(t *testing.T) {
	useRealCache(t)
	temp := t.TempDir()
	t.Setenv("HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")
	_, err := lock(filepath.Join(t.TempDir(), "plan.md"))
	if err == nil || !strings.HasPrefix(err.Error(), "cannot find a cache folder for the lock: ") {
		t.Fatalf("err = %v, want the cache folder message", err)
	}
	// Nothing at all may be created when the cache folder cannot be found.
	filepath.WalkDir(temp, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			t.Errorf("no cache folder: %s was created", p)
		}
		return nil
	})
}

// A cache folder that is not a real absolute path is refused, because a
// relative one is resolved against the working folder, and any user who can
// write in the working folder can put a folder of their own there first.
func TestLockDirRefusesARelativeCache(t *testing.T) {
	// macOS reads only HOME, so the XDG cases can only run where the
	// system really uses XDG_CACHE_HOME.
	xdg := t.TempDir()
	xdgUsed := func() bool {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("XDG_CACHE_HOME", xdg)
		got, err := os.UserCacheDir()
		return err == nil && got == xdg
	}()
	for name, c := range map[string]struct{ home, xdg string }{
		"relative home":     {"rel", ""},
		"dot home":          {".", ""},
		"empty home":        {"", ""},
		"relative cache":    {"", "rel"},
		"dot cache":         {"", "."},
		"cache under dot":   {".", "./rel"},
		"empty cache alone": {"", ""},
	} {
		t.Run(name, func(t *testing.T) {
			useRealCache(t)
			// A good absolute home for the cases that only move the
			// cache variable, so the refusal cannot come from a
			// missing home. Both empty is its own case.
			home := c.home
			if home == "" && c.xdg != "" {
				home = t.TempDir()
			}
			t.Setenv("HOME", home)
			t.Setenv("XDG_CACHE_HOME", c.xdg)
			dir, err := lockDir()
			if c.xdg != "" && !xdgUsed {
				// This system builds the cache folder from HOME
				// only, so the rule to prove here is the same one
				// everywhere: the answer is absolute or refused.
				if err == nil && !filepath.IsAbs(dir) {
					t.Fatalf("lock folder = %s, want an absolute path or a refusal", dir)
				}
				return
			}
			if err == nil {
				t.Fatalf("lock folder = %s, want a refusal", dir)
			}
			if !strings.HasPrefix(err.Error(), "cannot find a cache folder for the lock: ") {
				t.Fatalf("err = %v, want the cache folder message", err)
			}
			if dir != "" {
				t.Fatalf("lock folder = %s with an error, want nothing", dir)
			}
			// A refusal writes nothing, and above all not into the
			// working folder a relative path would point into.
			work := t.TempDir()
			t.Chdir(work)
			if _, err := lock(filepath.Join(work, "plan.md")); err == nil {
				t.Fatal("lock took a relative cache folder")
			}
			entries, err := os.ReadDir(work)
			if err != nil {
				t.Fatal(err)
			}
			for _, e := range entries {
				t.Errorf("a refused lock still created %s", e.Name())
			}
		})
	}
}

func TestSafeDirNamesEachRefusal(t *testing.T) {
	for name, c := range map[string]struct {
		setup func(dir string)
		want  string
	}{
		"plain file":       {func(dir string) { os.WriteFile(dir, nil, 0o600) }, "is not a folder"},
		"link to a folder": {func(dir string) { os.Symlink(t.TempDir(), dir) }, "is not a folder"},
		"folder others can read": {func(dir string) {
			os.Mkdir(dir, 0o755)
			os.Chmod(dir, 0o755)
		}, "is open to other users"},
	} {
		t.Run(name, func(t *testing.T) {
			dir := useLockBase(t)
			c.setup(dir)
			unlock, err := lock(filepath.Join(t.TempDir(), "plan.md"))
			if err == nil {
				unlock()
				t.Fatal("lock used an unsafe folder")
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want it to say %q", err, c.want)
			}
			if n := lockFiles(t, filepath.Dir(dir)); n != 0 {
				t.Fatalf("%d lock files were taken in a refused folder", n)
			}
		})
	}
}

func TestLockFileIsNamedActa(t *testing.T) {
	dir := useLockBase(t)
	unlock, err := lock(filepath.Join(t.TempDir(), "plan.md"))
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ents) != 1 {
		t.Fatalf("lock folder holds %d files, want 1", len(ents))
	}
	if !lockNameRe.MatchString(ents[0].Name()) {
		t.Fatalf("lock file = %s, want acta-<16 hex>.lock", ents[0].Name())
	}
}

func TestLockMakesPrivateFolder(t *testing.T) {
	dir := useLockBase(t)
	unlock, err := lock(filepath.Join(t.TempDir(), "plan.md"))
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	fi, err := os.Lstat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !fi.IsDir() || fi.Mode().Perm() != 0o700 {
		t.Fatalf("lock folder mode = %v, want a 0700 folder", fi.Mode())
	}
}

func TestLockRejectsUnsafeFolder(t *testing.T) {
	for name, setup := range map[string]func(dir string){
		"symlink": func(dir string) { os.Symlink(t.TempDir(), dir) },
		"others can write": func(dir string) {
			os.Mkdir(dir, 0o700)
			os.Chmod(dir, 0o777)
		},
		"plain file": func(dir string) { os.WriteFile(dir, nil, 0o600) },
	} {
		t.Run(name, func(t *testing.T) {
			dir := useLockBase(t)
			setup(dir)
			if unlock, err := lock(filepath.Join(t.TempDir(), "plan.md")); err == nil {
				unlock()
				t.Fatal("lock used an unsafe folder")
			}
		})
	}
}
