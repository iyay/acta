package write

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// debtBoxRe splits a debt line into its checkbox and the words after it.
var debtBoxRe = regexp.MustCompile(`^(\s*[-*] \[[ xX-]\] )(.*)$`)

// setPriority sets or removes the priority of one bug or debt item and
// commits the file. Every check runs before the file is touched.
func setPriority(cfg config.Config, it *board.Item, id, value string) (Outcome, error) {
	switch {
	case it.Kind != board.KindBug && it.Kind != board.KindDebtItem:
		return Outcome{}, bad("priority is only for bugs and debt items")
	case it.Legacy:
		return Outcome{}, bad("%s is a legacy file; move it into the root folder first", id)
	case it.Worktree != "":
		return Outcome{}, bad("%s is shown from worktree %s; edit it there", id, it.Worktree)
	case value != "none" && !board.ValidPriority(value):
		return Outcome{}, bad("priority must be high, medium, low or none, not %q", value)
	}
	dirty, err := dirtyBefore(cfg, it.Path)
	if err != nil {
		return Outcome{}, err
	}
	if it.Kind == board.KindBug {
		err = setBugPriority(it.Path, value)
	} else {
		err = setLinePriority(it.Path, it.Line, value)
	}
	if err != nil {
		return Outcome{}, bad("%s: %v", id, err)
	}
	return finish(cfg, it.Path, fmt.Sprintf("acta: %s priority %s", id, value), dirty), nil
}

// setBugPriority writes the priority field, or takes it out for none.
func setBugPriority(path, value string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var out []byte
	if value == "none" {
		out, err = RemoveField(src, "priority")
	} else {
		out, err = SetField(src, "priority", value)
	}
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

// setLinePriority swaps the tag on one debt line. An old tag goes first, so
// a line never ends up with two.
func setLinePriority(path string, line int, value string) error {
	unlock, err := lock(path)
	if err != nil {
		return err
	}
	defer unlock()
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(src), "\n")
	if line < 1 || line > len(lines) {
		return fmt.Errorf("line %d is not a checklist box", line)
	}
	m := debtBoxRe.FindStringSubmatch(lines[line-1])
	if m == nil {
		return fmt.Errorf("line %d is not a checklist box", line)
	}
	_, rest := board.SplitPriority(m[2])
	if value != "none" {
		rest = "(" + value + ") " + rest
	}
	lines[line-1] = m[1] + rest
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
