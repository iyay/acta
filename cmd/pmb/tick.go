package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"pm-board/internal/board"
	"pm-board/internal/hook"
	"pm-board/internal/write"
)

const tickUsage = "usage: pmb tick plans/<stem>#task-N [--step N | --all | --start]"

// cmdTick ticks checkboxes in a plan so the board shows progress while an
// agent works. It never commits: the plan file is shared, and the
// orchestrator commits it once per wave.
func cmdTick(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("tick", stderr)
	// Show the id shape and the flags on -h/--help so agents copy it right.
	fs.Usage = func() {
		fmt.Fprintln(stderr, tickUsage)
		fs.PrintDefaults()
	}
	step := fs.Int("step", 0, "tick this checkbox, counting from 1")
	all := fs.Bool("all", false, "tick every checkbox of the task")
	agent := fs.String("agent", "", "name the agent running this tick (default: the AI_AGENT variable)")
	start := fs.Bool("start", false, "mark the task started without ticking a box")
	pos, err := parseMixed(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		fmt.Fprintln(stderr, tickUsage)
		return exitBadInput
	}
	// Starting and ticking are different actions; mixing them is a typo,
	// and writing half of each would lie to the board. A bare tick with
	// no flag ticks every box, so --start only conflicts with the flags.
	if len(pos) != 1 || *start && (*all || *step > 0) || !*start && ((*step > 0) == *all) {
		fmt.Fprintln(stderr, tickUsage)
		return exitBadInput
	}
	cfg, b, code := loadBoard(*root, stderr)
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
	if *start {
		// No box moves; only the started record below is written.
		fmt.Fprintf(stdout, "%s started\n", it.ID)
	} else {
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
	}
	// The board shows who works on what, so a successful tick says who it was.
	// The record is a side effect: the tick itself worked, so a failure here
	// only gets printed.
	_ = hook.EnsureGitignore(cfg.Root, ".agents.json")
	if err := write.RecordAgent(cfg.Root, it.ID, write.AgentName(*agent, os.Getenv), time.Now(), *start); err != nil {
		fmt.Fprintf(stderr, "agent record: %v\n", err)
	}
	return exitOK
}
