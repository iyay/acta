// Package editor opens a file in the user's editor.
package editor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Cmd opens path in $EDITOR, or vi when it is unset. A line above 1 is passed
// as +N, which vi, vim, nvim, nano, emacs, micro and helix all understand.
func Cmd(path string, line int) *exec.Cmd {
	ed := strings.Fields(os.Getenv("EDITOR"))
	if len(ed) == 0 {
		ed = []string{"vi"}
	}
	args := append([]string{}, ed[1:]...)
	if line > 1 {
		args = append(args, fmt.Sprintf("+%d", line))
	}
	args = append(args, path)
	return exec.Command(ed[0], args...)
}
