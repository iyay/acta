package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iyay/acta/internal/telemetry"
)

// telemetryHome points HOME at a fresh folder and turns on an endpoint, so the
// telemetry command works on a machine of its own.
func telemetryHome(t *testing.T) *telemetry.Store {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", filepath.Join(home, ".acta", "config.yaml"))
	t.Setenv("DO_NOT_TRACK", "")
	t.Setenv("ACTA_TELEMETRY", "")
	old := telemetry.Endpoint
	telemetry.Endpoint = "https://script.google.com/macros/s/test/exec"
	t.Cleanup(func() { telemetry.Endpoint = old })
	return telemetry.Default()
}

func runTelemetry(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code := Run(append([]string{"telemetry"}, args...), strings.NewReader(""), false, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestTelemetryOnOff(t *testing.T) {
	s := telemetryHome(t)
	if code, _, errOut := runTelemetry(t, "off"); code != exitOK {
		t.Fatalf("off exit %d: %s", code, errOut)
	}
	if s.Enabled() {
		t.Fatal("still enabled after off")
	}
	if code, _, errOut := runTelemetry(t, "on"); code != exitOK {
		t.Fatalf("on exit %d: %s", code, errOut)
	}
	if !s.Enabled() {
		t.Fatal("not enabled after on")
	}
}

func TestTelemetryStatusShowsBodiesAndSource(t *testing.T) {
	s := telemetryHome(t)
	if err := s.Add("commands", "show"); err != nil {
		t.Fatal(err)
	}
	code, out, errOut := runTelemetry(t, "status")
	if code != exitOK {
		t.Fatalf("status exit %d: %s", code, errOut)
	}
	// The status run counts itself before it prints, so read the day after it.
	days, _ := s.Days()
	body, _ := telemetry.Payload(days[0])
	for _, want := range []string{"Telemetry: on", "Source: setting", "not finished", string(body)} {
		if !strings.Contains(out, want) {
			t.Fatalf("status misses %q:\n%s", want, out)
		}
	}
	t.Setenv("DO_NOT_TRACK", "1")
	_, out, _ = runTelemetry(t, "status")
	if !strings.Contains(out, "Telemetry: off") || !strings.Contains(out, "Source: environment") {
		t.Fatalf("status under DO_NOT_TRACK:\n%s", out)
	}
	telemetry.Endpoint = ""
	_, out, _ = runTelemetry(t, "status")
	if !strings.Contains(out, "Source: no endpoint in this build") {
		t.Fatalf("status with no endpoint:\n%s", out)
	}
}

func TestTelemetryFlushPrintsNothing(t *testing.T) {
	s := telemetryHome(t)
	telemetry.Endpoint = "https://127.0.0.1:1/unreachable"
	if err := s.Add("commands", "show"); err != nil {
		t.Fatal(err)
	}
	code, out, errOut := runTelemetry(t, "flush")
	if code != exitOK || out != "" || errOut != "" {
		t.Fatalf("flush = %d, %q, %q; want 0 and no output", code, out, errOut)
	}
	if s.LastFlush().Before(time.Now().Add(-time.Minute)) {
		t.Fatal("flush did not run")
	}
}

func TestTelemetryBadSubcommand(t *testing.T) {
	telemetryHome(t)
	for _, args := range [][]string{nil, {"sideways"}, {"on", "extra"}} {
		code, _, errOut := runTelemetry(t, args...)
		if code != exitBadInput || !strings.Contains(errOut, "acta telemetry") {
			t.Fatalf("%v = %d, %q; want bad input and the usage line", args, code, errOut)
		}
	}
}
