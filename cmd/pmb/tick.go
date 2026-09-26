package main

import (
	"errors"
	"fmt"
	"io"

	"pm-board/internal/board"
	"pm-board/internal/write"
)

const tickUsage = "usage: pmb tick <task-id> [--step N | --all]"

// cmdTick ticks checkboxes in a plan so the board shows progress while an
// agent works. It never commits: the plan file is shared, and the
// orchestrator commits it once per wave.
func cmdTick(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("tick", stderr)
	step := fs.Int("step", 0, "tick this checkbox, counting from 1")
	all := fs.Bool("all", false, "tick every checkbox of the task")
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) != 1 || (*step > 0) == *all {
		fmt.Fprintln(stderr, tickUsage)
		return exitBadInput
	}
	_, b, code := loadBoard(*root, stderr)
	if code != exitOK {
		return code
	}
	it := b.Get(pos[0])
	switch {
	case it == nil:
		fmt.Fprintf(stderr, "unknown id %s\n", pos[0])
		return exitBadInput
	case it.Kind != board.KindTask:
		fmt.Fprintf(stderr, "%s is not a task\n", pos[0])
		return exitBadInput
	case it.Legacy:
		fmt.Fprintf(stderr, "%s is in a legacy folder; move it into the root folder first\n", pos[0])
		return exitBadInput
	}
	n := *step
	if *all {
		n = 0
	}
	done, total, err := write.Tick(it.Path, it.Line, n)
	if err != nil {
		fmt.Fprintln(stderr, err)
		if errors.Is(err, write.ErrBadInput) {
			return exitBadInput
		}
		return exitOther
	}
	fmt.Fprintf(stdout, "%s %d/%d\n", it.ID, done, total)
	return exitOK
}
