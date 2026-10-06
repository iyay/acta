package write

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/iyay/acta/internal/board"
)

// These match the board parser, so tick counts the same boxes the board shows.
var (
	tickTaskRe = regexp.MustCompile(`^### Task \S`)
	tickBoxRe  = regexp.MustCompile(`^(\s*[-*] \[)([ xX])(\])`)
	// A debt line also allows the wontfix state, so this one accepts a
	// box already in any of the three states, not just open or done.
	tickAnyBoxRe = regexp.MustCompile(`^(\s*[-*] \[)([ xX-])(\])`)
)

// TickText ticks box number step (1-based) of the task whose heading is on
// headingLine (1-based), or every box of that task when step is 0. Only the
// ticked lines change.
func TickText(src []byte, headingLine, step int) ([]byte, int, int, error) {
	return boxText(src, headingLine, step, 'x')
}

// UntickText clears every box of the task whose heading is on headingLine,
// so a task marked done by mistake can go back to open.
func UntickText(src []byte, headingLine int) ([]byte, int, int, error) {
	return boxText(src, headingLine, 0, ' ')
}

// boxText writes state ('x' or ' ') into the ticked lines of one task and
// leaves every other line byte for byte as it was.
func boxText(src []byte, headingLine, step int, state byte) ([]byte, int, int, error) {
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
			lines[i] = tickBoxRe.ReplaceAllString(lines[i], "${1}"+string(state)+"${3}")
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
	return rewriteTask(path, headingLine, func(src []byte) ([]byte, int, int, error) {
		return TickText(src, headingLine, step)
	})
}

// Untick opens every box of the task on headingLine (1-based) in the plan
// file at path, and gives back the boxes of that task after the change.
func Untick(path string, headingLine int) (int, int, error) {
	return rewriteTask(path, headingLine, func(src []byte) ([]byte, int, int, error) {
		return UntickText(src, headingLine)
	})
}

// rewriteTask reads the file, hands it to edit and writes the result back
// through a temp file, so a reader never sees half a plan.
func rewriteTask(path string, headingLine int, edit func([]byte) ([]byte, int, int, error)) (int, int, error) {
	unlock, err := lock(path)
	if err != nil {
		return 0, 0, err
	}
	defer unlock()
	src, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}
	out, done, total, err := edit(src)
	if err != nil {
		return 0, 0, err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return 0, 0, err
	}
	return done, total, os.Rename(tmp, path)
}

// TickLine sets the single checklist box whose text (any priority tag
// removed) is text to state ('x' or '-'). It is for a debt file, where each
// checklist line is its own item instead of one line among a task's boxes.
// The caller passes the line the board read, which may be stale when a hand
// edit added a line above meanwhile. That line wins when its text matches;
// otherwise the file is searched for the one line with this text, and none
// or more than one is an error naming the text. Nothing is written then.
func TickLine(path string, line int, text string, state byte) error {
	unlock, err := lock(path)
	if err != nil {
		return err
	}
	defer unlock()
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	orig := string(src)
	lines := strings.Split(orig, "\n")
	at := -1
	if line >= 1 && line <= len(lines) && tickAnyBoxRe.MatchString(lines[line-1]) && lineText(lines[line-1]) == text {
		at = line - 1
	} else {
		at = findLine(lines, text)
		if at < 0 {
			return bad("no checklist line with text %q", text)
		}
		if at == -2 {
			return bad("more than one checklist line with text %q", text)
		}
	}
	lines[at] = tickAnyBoxRe.ReplaceAllString(lines[at], "${1}"+string(state)+"${3}")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// lineText is what a debt checklist line says, without its box and without
// any priority tag, the same text the board keeps as the item's title.
func lineText(ln string) string {
	text := strings.TrimSpace(ln[strings.Index(ln, "]")+1:])
	_, rest := board.SplitPriority(text)
	return rest
}

// findLine gives the one checklist line whose text is text: its index, -1
// when none matches, -2 when more than one does.
// ponytail: two sentinels instead of a count; only caller reads them.
func findLine(lines []string, text string) int {
	at := -1
	for i, ln := range lines {
		if !tickAnyBoxRe.MatchString(ln) || lineText(ln) != text {
			continue
		}
		if at >= 0 {
			return -2
		}
		at = i
	}
	return at
}

// lockRoot points the lock folder somewhere else. Only a test sets it, to a
// temp folder, so no test ever writes into the real cache folder. Empty
// means the user's own cache folder.
var lockRoot string

// lockDir gives the one folder every lock of this user lives in. The cache
// folder sits inside the user's own home, so no other user can create it
// first, and it is the same for every process whatever their TMPDIR says.
// There is no /tmp fallback: a shared folder there can be taken over by
// someone else, and a lock nobody can trust blocks every tick forever.
// A cache folder that is not an absolute path is refused too, because a
// relative one is read against the working folder, and anyone who can write
// there could put their own folder in the way first.
func lockDir() (string, error) {
	dir := lockRoot
	if dir == "" {
		cache, err := os.UserCacheDir()
		if err != nil {
			return "", fmt.Errorf("cannot find a cache folder for the lock: %w", err)
		}
		dir = filepath.Join(cache, "acta", "locks")
	}
	if !filepath.IsAbs(dir) {
		return "", fmt.Errorf("cannot find a cache folder for the lock: %s is not an absolute path", dir)
	}
	return dir, nil
}

// lockPath gives the lock file for a plan, named from the plan's real path
// so two names of one plan still share one lock.
func lockPath(plan string) (string, error) {
	real, err := filepath.Abs(plan)
	if err != nil {
		return "", err
	}
	if r, err := filepath.EvalSymlinks(real); err == nil {
		real = r
	}
	dir, err := lockDir()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(real))
	return filepath.Join(dir, "acta-"+hex.EncodeToString(sum[:8])+".lock"), nil
}

// safeDir makes the lock folder, or checks the one already there. Someone
// else could have made it first to steal or block our locks, so it must be
// a real folder, ours, and closed to everyone else. Each refusal says which
// of those three failed, so the fix is obvious.
func safeDir(dir string) error {
	if err := os.Mkdir(dir, 0o700); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}
	fi, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("lock folder %s is not a folder", dir)
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok || int(st.Uid) != os.Getuid() {
		return fmt.Errorf("lock folder %s is not owned by you", dir)
	}
	if fi.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("lock folder %s is open to other users", dir)
	}
	return nil
}

// lock takes an OS lock (flock) on a file in the user's cache folder, named
// from the plan's real path. The kernel drops the lock when the process
// ends, even in a crash, so no lock is ever left behind to take over, and
// two processes can never both hold it. Keeping the file out of the repo
// means no stray file shows up in git status.
func lock(plan string) (func(), error) {
	path, err := lockPath(plan)
	if err != nil {
		return nil, err
	}
	// The cache folder and the acta folder under it are ours and are not
	// there yet on a first run, so make them closed to everyone else.
	if err := os.MkdirAll(filepath.Dir(filepath.Dir(path)), 0o700); err != nil {
		return nil, err
	}
	if err := safeDir(filepath.Dir(path)); err != nil {
		return nil, err
	}
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
			return nil, fmt.Errorf("%s is locked by another acta command", plan)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
