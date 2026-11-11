package cli

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/telemetry"
)

// spawnFlush is a variable so a test can count flushes without starting one.
var spawnFlush = execFlush

// telemetryNotice is shown once per machine. It says what is counted, what is
// never sent, and how to stop it.
const telemetryNotice = "acta counts anonymous daily usage (command names, harness, acta skill names, settings) to see how it is used. No text, paths, repo or user names are sent. Turn it off with: acta telemetry off (or DO_NOT_TRACK=1). Details: acta telemetry status"

// countRun counts this call, shows the notice on a terminal the first time,
// and starts a background flush once an hour. Nothing here may fail the
// command, so every error is dropped.
func countRun(args []string, stderr io.Writer) {
	s := telemetry.Default()
	if !s.Enabled() {
		return
	}
	name := commandName(args)
	_ = s.Add("commands", name)
	if v, _, err := config.ResolveUser(); err == nil {
		_ = s.SetSettings(map[string]string{
			"build_executor": v.BuildExecutor,
			"plan_depth":     v.PlanDepth,
			"review_depth":   v.ReviewDepth,
			"chat_language":  v.ChatLanguage,
		})
	}
	// Hooks speak to an agent, not a person, so they never print the notice
	// or start a flush here. The session-start hook does both on its own.
	if name == "hook" {
		return
	}
	if stderrIsTerminal(stderr) {
		if first, err := s.MarkNoticed(); err == nil && first {
			fmt.Fprintln(stderr, telemetryNotice)
		}
	}
	// usage promises to stay offline, and telemetry would start itself again.
	if name == "usage" || name == "telemetry" {
		return
	}
	startFlush(s)
}

// startFlush starts one background flush when the last try is over an hour
// old. It writes the time first, so commands right after this one do not
// start a second flush.
func startFlush(s *telemetry.Store) {
	if time.Since(s.LastFlush()) < time.Hour {
		return
	}
	if err := s.SetLastFlush(time.Now()); err != nil {
		return
	}
	_ = spawnFlush()
}

// commandName is the one word a call is counted under. Arguments and flags are
// never part of it.
func commandName(args []string) string {
	if len(args) == 0 {
		return "tui"
	}
	if args[0] == "help" {
		return "help"
	}
	if strings.HasPrefix(args[0], "-") {
		for _, a := range args {
			switch {
			case a == "--":
				return "tui"
			case isHelpFlag(a):
				return "help"
			case a == "--version":
				return "version"
			}
		}
		return "tui"
	}
	return args[0]
}

// execFlush runs acta telemetry flush in its own session, so it outlives this
// command and its output goes nowhere.
func execFlush() error {
	exe, err := exePath()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "telemetry", "flush")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}
