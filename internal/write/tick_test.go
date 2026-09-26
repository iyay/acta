package write

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const plan = "# P\n\n### Task 1: One\n- [x] a\n- [ ] b\n```text\n- [ ] in a fence\n```\n- [ ] c\n\n### Task 2: Two\n- [ ] d\n"

func TestTickText(t *testing.T) {
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
	src := strings.ReplaceAll(plan, "\n", "\r\n")
	out, done, _, err := TickText([]byte(src), 3, 2)
	if err != nil || done != 2 || string(out) != strings.Replace(src, "- [ ] b", "- [x] b", 1) {
		t.Fatalf("CRLF: %q %v", out, err)
	}
}

func TestTickTextMixedLineEndings(t *testing.T) {
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

func TestTickFileConcurrent(t *testing.T) {
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
	if _, err := lock(p); err != nil {
		t.Fatal(err)
	}
	os.Stdout.WriteString("held\n")
	select {}
}

func TestTickAfterHolderDies(t *testing.T) {
	p := filepath.Join(t.TempDir(), "plan.md")
	if err := os.WriteFile(p, []byte(plan), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperHoldLock$")
	cmd.Env = append(os.Environ(), "PMB_HOLD_LOCK="+p)
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
	p := filepath.Join(t.TempDir(), "plan.md")
	unlock, err := lock(p)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	if _, err := lock(p); err == nil || !strings.Contains(err.Error(), "locked") {
		t.Fatalf("second lock: err = %v, want a locked error", err)
	}
}
