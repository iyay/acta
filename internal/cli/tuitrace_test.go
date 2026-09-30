package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTuiTraceOffWhenThePathIsEmpty(t *testing.T) {
	var errs bytes.Buffer
	tr, opts, done := tuiTrace("", os.Stdout, &errs)
	done()
	if tr != nil || len(opts) != 0 || errs.Len() != 0 {
		t.Errorf("an empty path turned the trace on: %v %d %q", tr, len(opts), errs.String())
	}
}

// The switch is set but the file cannot open. The TUI must still run, so the
// trace is the only thing lost, and the reader has to hear why.
func TestTuiTraceWarnsAndRunsWhenTheFileCannotOpen(t *testing.T) {
	var errs bytes.Buffer
	bad := filepath.Join(t.TempDir(), "no-such-dir", "f.log")
	tr, opts, done := tuiTrace(bad, os.Stdout, &errs)
	done()
	if tr != nil || len(opts) != 0 {
		t.Error("a trace that could not open still wrapped the output")
	}
	if !strings.Contains(errs.String(), "ACTA_TUI_TRACE") {
		t.Errorf("the warning does not name the switch: %q", errs.String())
	}
}

func TestTuiTraceWritesTheSummaryOnDone(t *testing.T) {
	var errs bytes.Buffer
	path := filepath.Join(t.TempDir(), "f.log")
	tr, opts, done := tuiTrace(path, os.Stdout, &errs)
	if tr == nil || len(opts) != 1 {
		t.Fatalf("a good path gave tracer %v and %d options", tr, len(opts))
	}
	tr.Notch(true)
	done()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(got)), "\n")
	if !strings.HasPrefix(lines[0], "notch ") || !strings.HasPrefix(lines[len(lines)-1], "summary frames=") {
		t.Errorf("log is %q", got)
	}
}

// The TUI ends once, but nothing stops a caller from saying it ended twice.
// A second call must not take the program down or take the writer with it.
func TestTuiTraceDoneTwiceIsSafe(t *testing.T) {
	var errs bytes.Buffer
	path := filepath.Join(t.TempDir(), "f.log")
	tr, _, done := tuiTrace(path, os.Stdout, &errs)
	tr.Notch(true)
	done()
	done()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(got)), "\n")
	if last := lines[len(lines)-1]; !strings.HasPrefix(last, "summary frames=") {
		t.Errorf("the last line is %q, want the summary", last)
	}
	if errs.Len() != 0 {
		t.Errorf("the second call complained: %q", errs.String())
	}
}
