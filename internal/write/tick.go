package write

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
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
	text := string(src)
	nl := "\n"
	if strings.Contains(text, "\r\n") {
		nl = "\r\n"
	}
	lines := strings.Split(text, nl)
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
	return []byte(strings.Join(lines, nl)), done, len(boxes), nil
}

// Tick ticks a box in the plan file at path. A lock file keeps two
// implementers in one worktree from overwriting each other's ticks.
func Tick(path string, headingLine, step int) (int, int, error) {
	unlock, err := lock(path + ".lock")
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

// lock takes a lock file, waiting up to five seconds. A lock older than a
// minute is treated as left behind by a crash and taken over.
func lock(path string) (func(), error) {
	deadline := time.Now().Add(5 * time.Second)
	for {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			f.Close()
			return func() { os.Remove(path) }, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, err
		}
		if st, err := os.Stat(path); err == nil && time.Since(st.ModTime()) > time.Minute {
			os.Remove(path)
			continue
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("plan is locked by %s", path)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
