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

func TestPreToolCountsActaSkillLoads(t *testing.T) {
	long := strings.Repeat("a", 10*1024)
	cases := []struct {
		name, tool, field, value, want string
	}{
		{"claude code acta skill", "Skill", "skill", "acta:shape", "shape"},
		{"omp skill read", "read", "file_path", "skill://tdd", "tdd"},
		{"claude code skill of another plugin", "Skill", "skill", "other:shape", ""},
		{"claude code bare name", "Skill", "skill", "shape", ""},
		{"a path out of the prefix", "Skill", "skill", "acta:../x", ""},
		{"an empty name", "Skill", "skill", "acta:", ""},
		{"a 10 KB name", "Skill", "skill", "acta:" + long, ""},
		{"mixed case", "Skill", "skill", "acta:Shape", ""},
		{"mixed case prefix", "Skill", "skill", "ACTA:shape", ""},
		{"omp skill not on the list", "read", "file_path", "skill://my-private-skill", ""},
		{"omp skill with an acta prefix", "read", "file_path", "skill://acta:shape", ""},
		{"omp a 10 KB name", "read", "file_path", "skill://" + long, ""},
		{"a plain file read", "read", "file_path", "shape", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := hookRepo(t)
			s, _ := runEnv(t, false)
			t.Chdir(dir)
			code, out, errb := runHook(t, "pre-tool", toolEvent(t, "s1", "", c.tool, map[string]string{c.field: c.value}))
			if code != exitOK || errb != "" {
				t.Fatalf("exit %d, stderr %q; want a quiet pass", code, errb)
			}
			if c.tool == "Skill" && out != "" {
				t.Fatalf("a Skill call printed %q; want no hint", out)
			}
			got := todayCounts(t, s).Skills
			if c.want == "" && len(got) != 0 {
				t.Fatalf("skills = %v, want nothing recorded", got)
			}
			if c.want != "" && (got[c.want] != 1 || len(got) != 1) {
				t.Fatalf("skills = %v, want only %s: 1", got, c.want)
			}
		})
	}
}

func TestPreToolCountsEachBlockAndKeepsItsText(t *testing.T) {
	cases := []struct {
		name, rule, command string
		repo                func(t *testing.T) string
	}{
		{"bare go test", "go_test", "go test ./...", goTestRepoCLI},
		{"destructive", "destructive", "rm -rf x", hookRepo},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := c.repo(t)
			s, _ := runEnv(t, false)
			t.Chdir(dir)
			ev := toolEvent(t, "s1", "", "Bash", map[string]string{"command": c.command})
			// The same call with telemetry off gives the text the block must keep.
			t.Setenv("DO_NOT_TRACK", "1")
			_, _, want := runHook(t, "pre-tool", ev)
			t.Setenv("DO_NOT_TRACK", "")
			code, out, errb := runHook(t, "pre-tool", ev)
			if code != exitBlock || out != "" || errb != want || want == "" {
				t.Fatalf("exit %d, stdout %q, stderr %q; want exit %d and stderr %q", code, out, errb, exitBlock, want)
			}
			if got := todayCounts(t, s).Blocks; got[c.rule] != 1 || len(got) != 1 {
				t.Fatalf("blocks = %v, want only %s: 1", got, c.rule)
			}
		})
	}
}
