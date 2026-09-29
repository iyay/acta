package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/hook"
	"github.com/iyay/acta/internal/write"
)

const tickUsage = "usage: acta tick <id> [--step N | --all | --start | --wontfix]"

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
	for _, on := range []bool{*step > 0, *all, *start, *wontfix} {
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
	case it.Kind == board.KindDebtItem:
		state := byte('x')
		if *wontfix {
			state = '-'
		}
		if err := write.TickLine(it.Path, it.Line, state); err != nil {
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
		if err := markTaskDates(cfg, it); err != nil {
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

// markTaskDates writes the dates a task tick gives its plan and the spec or
// bug above it. Work began the moment any task is ticked, and work ends when
// every task of the plan is done, and the spec a little later: when every
// plan under it is done too. The board is read again because the tick that
// just happened is what the file says.
func markTaskDates(cfg config.Config, it *board.Item) error {
	fresh, err := board.Load(cfg)
	if err != nil {
		return err
	}
	plan := fresh.Get(it.PlanID)
	if err := markDate(plan.Path, write.MarkStarted); err != nil {
		return err
	}
	if plan.Done == plan.Total {
		if err := markDate(plan.Path, write.MarkFinishedOnce); err != nil {
			return err
		}
	}
	spec := fresh.Get(plan.SpecID)
	if spec == nil {
		return nil
	}
	if err := markDate(spec.Path, write.MarkStarted); err != nil {
		return err
	}
	if spec.Status == "done" {
		return markDate(spec.Path, write.MarkFinishedOnce)
	}
	return nil
}

// markDate runs one date writer over a file and writes the result back only
// when the bytes really changed, so a tick never touches a file it has
// nothing new to say about.
func markDate(path string, set func([]byte) ([]byte, error)) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out, err := set(src)
	if err != nil {
		return err
	}
	if bytes.Equal(src, out) {
		return nil
	}
	return os.WriteFile(path, out, 0o644)
}
