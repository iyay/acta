package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/iyay/acta/internal/telemetry"
)

// countedSession runs the session-start hook in an empty folder, so it never
// touches the real repo, and gives back what the agent would read.
func countedSession(t *testing.T) string {
	t.Helper()
	t.Chdir(t.TempDir())
	var out, errOut strings.Builder
	if code := Run([]string{"hook", "session-start"}, strings.NewReader(""), false, &out, &errOut); code != exitOK {
		t.Fatalf("session-start exit %d: %s", code, errOut.String())
	}
	if strings.Contains(errOut.String(), "telemetry") {
		t.Fatalf("session-start printed telemetry text on stderr: %q", errOut.String())
	}
	return out.String()
}

func TestSessionStartCountsTheHarness(t *testing.T) {
	cases := []struct {
		name, harness, pluginRoot, want string
	}{
		{"claude code", "", "", "claude"},
		{"omp", "omp", "", "omp"},
		{"codex", "", "/codex/plugin", "codex"},
		{"omp wins over codex", "omp", "/codex/plugin", "omp"},
		{"unknown harness", "vim", "", "other"},
		{"a known name that only omp may set", "claude", "", "other"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, _ := runEnv(t, false)
			t.Setenv("ACTA_HARNESS", c.harness)
			t.Setenv("PLUGIN_ROOT", c.pluginRoot)
			countedSession(t)
			got := todayCounts(t, s).Sessions
			if got[c.want] != 1 || len(got) != 1 {
				t.Fatalf("sessions = %v, want only %s: 1", got, c.want)
			}
		})
	}
}

func TestSessionStartShowsTheNoticeOnce(t *testing.T) {
	s, _ := runEnv(t, false)
	t.Setenv("ACTA_HARNESS", "")
	t.Setenv("PLUGIN_ROOT", "")
	if out := countedSession(t); !strings.Contains(out, telemetryNotice) || !s.Noticed() {
		t.Fatalf("first session gave no notice or did not mark it:\n%s", out)
	}
	if out := countedSession(t); strings.Contains(out, "acta telemetry off") {
		t.Fatalf("second session showed the notice again:\n%s", out)
	}
}

func TestSessionStartShowsNoNoticeWhenTelemetryIsOff(t *testing.T) {
	for name, off := range map[string]func(t *testing.T){
		"do not track": func(t *testing.T) { t.Setenv("DO_NOT_TRACK", "1") },
		"no endpoint":  func(t *testing.T) { telemetry.Endpoint = "" },
	} {
		t.Run(name, func(t *testing.T) {
			s, spawns := runEnv(t, false)
			off(t)
			if out := countedSession(t); strings.Contains(out, "acta telemetry off") {
				t.Fatalf("notice shown with telemetry off:\n%s", out)
			}
			// Turn it back on, so the store can read what was written.
			telemetry.Endpoint = "https://script.google.com/macros/s/test/exec"
			t.Setenv("DO_NOT_TRACK", "")
			if s.Noticed() || *spawns != 0 {
				t.Fatalf("telemetry off still marked the notice (%v) or started a flush (%d)", s.Noticed(), *spawns)
			}
		})
	}
}

func TestSessionStartStartsTheFlushHourly(t *testing.T) {
	s, spawns := runEnv(t, false)
	countedSession(t)
	if *spawns != 1 {
		t.Fatalf("flushes started = %d, want 1", *spawns)
	}
	countedSession(t)
	if *spawns != 1 {
		t.Fatalf("a second session inside the hour started another flush (%d)", *spawns)
	}
	if err := s.SetLastFlush(time.Now().Add(-61 * time.Minute)); err != nil {
		t.Fatal(err)
	}
	countedSession(t)
	if *spawns != 2 {
		t.Fatalf("flushes started = %d after an hour, want 2", *spawns)
	}
}
