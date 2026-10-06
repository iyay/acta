package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/hook"
	"github.com/iyay/acta/internal/write"
)

const tickUsage = "usage: acta tick <id> [--step N | --all | --start | --wontfix | --undo]"

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
	wontfix := fs.Bool("wontfix", false, "mark a debt line wontfix instead of done")
	undo := fs.Bool("undo", false, "put the task or debt line back to open")
	pos, err := parseMixed(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		fmt.Fprintln(stderr, tickUsage)
		return exitBadInput
	}
	// Exactly one action per call: mixing two would lie to the board about
	// which one actually happened, and a bare tick with none is a typo.
	set := 0
	for _, on := range []bool{*step > 0, *all, *start, *wontfix, *undo} {
		if on {
			set++
		}
	}
	if len(pos) != 1 || set != 1 {
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
	case it.Kind == board.KindDebtItem:
		// A debt line is only ever ticked whole or marked wontfix: it has
		// no steps of its own and nothing to start.
		if *step > 0 || *start {
			fmt.Fprintln(stderr, tickUsage)
			return exitSkipped
		}
	case it.Kind != board.KindTask:
		fmt.Fprintf(stderr, "%s is not a task\n", pos[0])
		return exitBadInput
	case it.Legacy:
		fmt.Fprintf(stderr, "%s is in a legacy folder; move it into the root folder first\n", pos[0])
		return exitBadInput
	case *wontfix:
		// wontfix is a debt-only state; a plan task has no such box.
		fmt.Fprintln(stderr, tickUsage)
		return exitSkipped
	}
	switch {
	case *undo:
		// Undo is for a slip of the finger. It clears the boxes and nothing
		// else: dates already written stay, and nobody is recorded.
		if it.Kind == board.KindDebtItem {
			err = write.TickLine(it.Path, it.Line, it.Title, ' ')
			if err == nil {
				fmt.Fprintf(stdout, "%s open\n", it.ID)
			}
		} else {
			var done, total int
			done, total, err = write.Untick(it.Path, it.Line)
			if err == nil {
				fmt.Fprintf(stdout, "%s %d/%d\n", it.ID, done, total)
			}
		}
		if err != nil {
			fmt.Fprintln(stderr, err)
			if errors.Is(err, write.ErrBadInput) {
				return exitBadInput
			}
			return exitOther
		}
		return exitOK
	case it.Kind == board.KindDebtItem:
		state := byte('x')
		if *wontfix {
			state = '-'
		}
		if err := write.TickLine(it.Path, it.Line, it.Title, state); err != nil {
			fmt.Fprintln(stderr, err)
			if errors.Is(err, write.ErrBadInput) {
				return exitBadInput
			}
			return exitOther
		}
		fmt.Fprintf(stdout, "%s ticked\n", it.ID)
	case *start:
		// No box moves; only the started record below is written.
		fmt.Fprintf(stdout, "%s started\n", it.ID)
	default:
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
	// A task tick also dates the plan above it and the spec or bug above
	// that. The dates are a side effect: the tick itself worked, so a
	// failure here only gets printed.
	if it.Kind == board.KindTask {
		if _, err := write.TaskDates(cfg, it); err != nil {
			fmt.Fprintf(stderr, "dates: %v\n", err)
		}
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
