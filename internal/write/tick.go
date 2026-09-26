package write

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

// These match the board parser, so tick counts the same boxes the board shows.
var (
	tickTaskRe = regexp.MustCompile(`^### Task \S`)
	tickBoxRe  = regexp.MustCompile(`^(\s*[-*] \[)([ xX])(\])`)
)

// TickText ticks box number step (1-based) of the task whose heading is on
// headingLine (1-based), or every box of that task when step is 0. Only the
// ticked lines change.
func TickText(src []byte, headingLine, step int) ([]byte, int, int, error) {
	// Split on \n like the board parser so a 1-based Line means the same
	// line here. Each line keeps its own trailing \r; the heading, fence
	// and box matches are all prefix checks, so a trailing \r is harmless.
	lines := strings.Split(string(src), "\n")
	if headingLine < 1 || headingLine > len(lines) || !tickTaskRe.MatchString(lines[headingLine-1]) {
		return nil, 0, 0, bad("line %d is not a ### Task heading", headingLine)
	}
	if step < 0 {
		return nil, 0, 0, bad("step must be 1 or more")
	}
	var boxes []int
	inFence := false
	for i := headingLine; i < len(lines); i++ {
		ln := lines[i]
		if strings.HasPrefix(strings.TrimSpace(ln), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if strings.HasPrefix(ln, "### ") || strings.HasPrefix(ln, "## ") {
			break
		}
		if tickBoxRe.MatchString(ln) {
			boxes = append(boxes, i)
		}
	}
	if len(boxes) == 0 {
		return nil, 0, 0, bad("the task has no checkboxes")
	}
	if step > len(boxes) {
		return nil, 0, 0, bad("step %d, but the task has %d checkboxes", step, len(boxes))
	}
	for n, i := range boxes {
		if step == 0 || n == step-1 {
			lines[i] = tickBoxRe.ReplaceAllString(lines[i], "${1}x${3}")
		}
	}
	done := 0
	for _, i := range boxes {
		if m := tickBoxRe.FindStringSubmatch(lines[i]); m != nil && m[2] != " " {
			done++
		}
	}
	return []byte(strings.Join(lines, "\n")), done, len(boxes), nil
}

// Tick ticks a box in the plan file at path. An OS lock keeps two
// implementers in one worktree from overwriting each other's ticks.
func Tick(path string, headingLine, step int) (int, int, error) {
	unlock, err := lock(path)
	if err != nil {
		return 0, 0, err
	}
	defer unlock()
	src, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}
	out, done, total, err := TickText(src, headingLine, step)
	if err != nil {
		return 0, 0, err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return 0, 0, err
	}
	return done, total, os.Rename(tmp, path)
}

// lock takes an OS lock (flock) on a file in the temp folder, named from
// the plan's real path. The kernel drops the lock when the process ends,
// even in a crash, so no lock is ever left behind to take over, and two
// processes can never both hold it. Keeping the file out of the repo means
// no stray file shows up in git status.
func lock(plan string) (func(), error) {
	real, err := filepath.Abs(plan)
	if err != nil {
		return nil, err
	}
	if r, err := filepath.EvalSymlinks(real); err == nil {
		real = r
	}
	sum := sha256.Sum256([]byte(real))
	path := filepath.Join(os.TempDir(), "pmb-"+hex.EncodeToString(sum[:8])+".lock")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return func() {
				syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
				f.Close()
			}, nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) {
			f.Close()
			return nil, err
		}
		if time.Now().After(deadline) {
			f.Close()
			return nil, fmt.Errorf("plan %s is locked by another pmb tick", plan)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
