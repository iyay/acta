package cli

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iyay/acta/internal/telemetry"
)

// runEnv gives a machine that counts, with a fake terminal check and a fake
// flush starter, and returns the store and the number of flushes started.
func runEnv(t *testing.T, tty bool) (*telemetry.Store, *int) {
	t.Helper()
	s := telemetryHome(t)
	t.Setenv("PM_VOICE_FILE", filepath.Join(t.TempDir(), "config.yaml"))
	t.Setenv("ACTA_NO_UPDATE_CHECK", "1")
	oldSpawn, oldTTY := spawnFlush, stderrIsTerminal
	t.Cleanup(func() { spawnFlush, stderrIsTerminal = oldSpawn, oldTTY })
	n := 0
	spawnFlush = func() error { n++; return nil }
	stderrIsTerminal = func(any) bool { return tty }
	return s, &n
}

func todayCounts(t *testing.T, s *telemetry.Store) telemetry.Day {
	t.Helper()
	days, err := s.Days()
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range days {
		if d.Day == telemetry.Today() {
			return d
		}
	}
	return telemetry.Day{}
}

func TestRunCountsTheCommandNameOnly(t *testing.T) {
	cases := []struct {
		args []string
		key  string
	}{
		{[]string{"list", "--json", "secret-repo-name"}, "list"},
		{[]string{"help"}, "help"},
		{[]string{"help", "show"}, "help"},
		{[]string{"--help"}, "help"},
		{[]string{"--version"}, "version"},
		{[]string{"frobnicate", "/home/me/private"}, "other"},
		{[]string{"hook", "prompt"}, "hook"},
	}
	for _, c := range cases {
		t.Run(strings.Join(c.args, " "), func(t *testing.T) {
			s, _ := runEnv(t, false)
			runAll(c.args...)
			d := todayCounts(t, s)
			if d.Commands[c.key] != 1 || len(d.Commands) != 1 {
				t.Fatalf("commands = %v, want only %s: 1", d.Commands, c.key)
			}
			raw, _ := telemetry.Payload(d)
			for _, arg := range c.args[1:] {
				if strings.Contains(string(raw), arg) && !strings.HasPrefix(arg, "-") && arg != "show" && arg != "prompt" {
					t.Fatalf("payload carries argument %q: %s", arg, raw)
				}
			}
		})
	}
}

func TestRunRecordsTheSettingsSnapshot(t *testing.T) {
	s, _ := runEnv(t, false)
	runAll("help")
	got := todayCounts(t, s).Settings
	if got["chat_language"] != "english" || got["plan_depth"] != "unset" {
		t.Fatalf("settings = %v, want the default config snapshot", got)
	}
}

func TestEveryHelpCommandIsAKnownTelemetryName(t *testing.T) {
	for _, c := range helpTable {
		if !telemetry.KnownCommand(c.name) {
			t.Errorf("command %q counts as other; add it to the telemetry command list", c.name)
		}
	}
}

func TestNoticeOnlyOnATerminalAndOnlyOnce(t *testing.T) {
	s, _ := runEnv(t, false)
	if _, errOut := runAll("help"); strings.Contains(errOut, "telemetry") {
		t.Fatalf("notice printed with no terminal: %q", errOut)
	}
	if s.Noticed() {
		t.Fatal("notice marked shown with no terminal")
	}
	stderrIsTerminal = func(any) bool { return true }
	if _, errOut := runAll("hook", "prompt"); strings.Contains(errOut, "telemetry") || s.Noticed() {
		t.Fatalf("a hook printed or marked the notice: %q", errOut)
	}
	_, errOut := runAll("help")
	if !strings.Contains(errOut, "acta telemetry off") || !s.Noticed() {
		t.Fatalf("first terminal run gave no notice: %q", errOut)
	}
	if _, errOut := runAll("help"); strings.Contains(errOut, "telemetry") {
		t.Fatalf("notice printed twice: %q", errOut)
	}
	if d := todayCounts(t, s); d.Commands["help"] != 3 {
		t.Fatalf("help count = %d, want 3: the notice run counts too", d.Commands["help"])
	}
}

func TestNoticeNeverShowsWhenSwitchedOff(t *testing.T) {
	_, _ = runEnv(t, true)
	t.Setenv("DO_NOT_TRACK", "1")
	if _, errOut := runAll("help"); strings.Contains(errOut, "telemetry") {
		t.Fatalf("notice printed under DO_NOT_TRACK: %q", errOut)
	}
}

func TestRunStartsTheFlushHourly(t *testing.T) {
	s, spawns := runEnv(t, false)
	runAll("help")
	if *spawns != 1 {
		t.Fatalf("flushes started = %d, want 1", *spawns)
	}
	runAll("list", "--json")
	if *spawns != 1 {
		t.Fatalf("a second run inside the hour started another flush (%d)", *spawns)
	}
	if err := s.SetLastFlush(time.Now().Add(-61 * time.Minute)); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"hook", "prompt"}, {"usage"}, {"telemetry", "status"}} {
		runAll(args...)
	}
	if *spawns != 1 {
		t.Fatalf("hook, usage or telemetry started a flush (%d)", *spawns)
	}
	runAll("help")
	if *spawns != 2 {
		t.Fatalf("flushes started = %d after an hour, want 2", *spawns)
	}
}

func TestRunStartsNoFlushWithoutAnEndpoint(t *testing.T) {
	_, spawns := runEnv(t, false)
	telemetry.Endpoint = ""
	runAll("help")
	if *spawns != 0 {
		t.Fatal("a build with no endpoint started a flush")
	}
}
