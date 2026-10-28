package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/usage"
)

// usageRepo makes a repository with one commit, so usage has a scope.
func usageRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	run("init", "-q", "-b", "main")
	run("config", "user.name", "test")
	run("config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "README.md")
	run("commit", "-q", "-m", "init")
	return dir
}

// usageHome makes a CODEX_HOME holding an empty sessions folder.
func usageHome(t *testing.T) (home, sessions string) {
	t.Helper()
	home = t.TempDir()
	sessions = filepath.Join(home, "sessions")
	if err := os.MkdirAll(sessions, 0o755); err != nil {
		t.Fatal(err)
	}
	return home, sessions
}

func usageLog(t *testing.T, sessions, rel string, lines ...string) {
	t.Helper()
	path := filepath.Join(sessions, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// jsonString quotes a value as JSON, so a fixture can hold a raw control
// character the way a transcript would.
func jsonString(t *testing.T, s string) string {
	t.Helper()
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func usageMeta(t *testing.T, id, cwd string) string {
	t.Helper()
	return fmt.Sprintf(`{"timestamp":"2026-10-28T10:00:00Z","type":"session_meta","payload":{"id":%s,"cwd":%s}}`, jsonString(t, id), jsonString(t, cwd))
}

func usageModel(t *testing.T, model string) string {
	t.Helper()
	return fmt.Sprintf(`{"type":"turn_context","payload":{"model":%s}}`, jsonString(t, model))
}

func usageRecord(t *testing.T, at, thread, root, response string, input, cached, output int64) string {
	t.Helper()
	return fmt.Sprintf(`{"timestamp":%s,"type":"token_usage_record","payload":{"thread_id":%s,"session_id":%s,"response_id":%s,"usage":{"input_tokens":%d,"cached_input_tokens":%d,"output_tokens":%d}}}`,
		jsonString(t, at), jsonString(t, thread), jsonString(t, root), jsonString(t, response), input, cached, output)
}

// usageRun runs one command from dir and gives back the code and streams.
func usageRun(t *testing.T, dir string, args ...string) (int, string, string) {
	t.Helper()
	var out, errw strings.Builder
	code := exitOK
	inDir(t, dir, func() {
		code = Run(args, strings.NewReader(""), false, &out, &errw)
	})
	return code, out.String(), errw.String()
}

// failWriter fails every write, the way a full disk does.
type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

// One complete command run shows the session table, the model summary and a
// clean stderr, and a partial run keeps the table and explains itself on
// stderr with the skipped exit code.
func TestUsageCommandTables(t *testing.T) {
	repo := usageRepo(t)

	t.Run("complete", func(t *testing.T) {
		home, sessions := usageHome(t)
		usageLog(t, sessions, "root.jsonl",
			usageMeta(t, "root-1", repo),
			usageModel(t, "example-model"),
			usageRecord(t, "2026-10-28T10:00:02.125Z", "root-1", "root-1", "response-1", 100, 60, 10),
		)
		t.Setenv("CODEX_HOME", home)
		code, out, errw := usageRun(t, repo, "usage")
		if code != exitOK || errw != "" {
			t.Fatalf("exit %d stderr %q", code, errw)
		}
		for _, want := range []string{"root-1", "example-model", "FRESH", "CACHED", "OUTPUT", "RESPONSES", "2026-10-28T10:00:02.125Z"} {
			if !strings.Contains(out, want) {
				t.Errorf("output misses %q:\n%s", want, out)
			}
		}
	})

	t.Run("partial", func(t *testing.T) {
		home, sessions := usageHome(t)
		usageLog(t, sessions, "root.jsonl",
			usageMeta(t, "root-1", repo),
			usageRecord(t, "2026-10-28T10:00:02Z", "root-1", "root-1", "response-1", 10, 0, 1),
			`{"type":"token_usage_record", broken`,
		)
		t.Setenv("CODEX_HOME", home)
		code, out, errw := usageRun(t, repo, "usage")
		if code != exitSkipped {
			t.Fatalf("exit %d, want %d", code, exitSkipped)
		}
		if !strings.Contains(out, "root-1") {
			t.Errorf("valid rows must survive a partial run:\n%s", out)
		}
		if !strings.Contains(errw, "invalid_json") || !strings.Contains(errw, "root.jsonl:3") {
			t.Errorf("stderr = %q, want the diagnostic and its place", errw)
		}
	})

	t.Run("no data", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("CODEX_HOME", home)
		code, out, errw := usageRun(t, repo, "usage")
		if code != exitOK || errw != "" {
			t.Fatalf("exit %d stderr %q", code, errw)
		}
		if !strings.Contains(strings.ToLower(out), "no data") {
			t.Fatalf("output = %q, want an explicit no data line", out)
		}
	})
}

// --json prints one parseable report with structured diagnostics, and never
// names a response.
func TestUsageCommandJSON(t *testing.T) {
	repo := usageRepo(t)
	home, sessions := usageHome(t)
	usageLog(t, sessions, "root.jsonl",
		usageMeta(t, "root-1", repo),
		usageModel(t, "example-model"),
		usageRecord(t, "2026-10-28T10:00:02Z", "root-1", "root-1", "response-1", 100, 60, 10),
		`{"type":"token_usage_record", broken`,
	)
	t.Setenv("CODEX_HOME", home)
	code, out, errw := usageRun(t, repo, "usage", "--json")
	if code != exitSkipped {
		t.Fatalf("exit %d, want %d", code, exitSkipped)
	}
	if strings.Contains(out, "response-1") {
		t.Fatalf("JSON shows a response identifier:\n%s", out)
	}
	var report usage.Report
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("stdout is not one JSON report: %v\n%s", err, out)
	}
	if report.SchemaVersion != 1 || report.Status != usage.StatusPartial {
		t.Fatalf("schema = %d status = %q", report.SchemaVersion, report.Status)
	}
	if len(report.Diagnostics) != 1 || report.Diagnostics[0].Code != "invalid_json" {
		t.Fatalf("diagnostics = %+v", report.Diagnostics)
	}
	if errw != "" {
		t.Fatalf("stderr = %q, want empty under --json", errw)
	}
}

// Every bad call is refused before any reading, and the exit codes separate
// bad input, an operational failure and a partial report.
func TestUsageCommandErrors(t *testing.T) {
	repo := usageRepo(t)

	t.Run("outside git", func(t *testing.T) {
		home, _ := usageHome(t)
		t.Setenv("CODEX_HOME", home)
		code, out, errw := usageRun(t, t.TempDir(), "usage")
		if code != exitBadInput || out != "" {
			t.Fatalf("exit %d stdout %q", code, out)
		}
		if !strings.Contains(errw, "repository") {
			t.Fatalf("stderr = %q, want the missing repository named", errw)
		}
	})

	t.Run("unknown flag", func(t *testing.T) {
		code, out, errw := usageRun(t, repo, "usage", "--bogus")
		if code != exitBadInput || out != "" || !strings.Contains(errw, "usage: acta usage") {
			t.Fatalf("exit %d stdout %q stderr %q", code, out, errw)
		}
	})

	t.Run("positional", func(t *testing.T) {
		code, _, errw := usageRun(t, repo, "usage", "extra")
		if code != exitBadInput || !strings.Contains(errw, "usage: acta usage") {
			t.Fatalf("exit %d stderr %q", code, errw)
		}
	})

	t.Run("overflow is operational", func(t *testing.T) {
		home, sessions := usageHome(t)
		usageLog(t, sessions, "root.jsonl",
			usageMeta(t, "root-1", repo),
			usageModel(t, "m"),
			usageRecord(t, "2026-10-28T10:00:02Z", "root-1", "root-1", "r1", math.MaxInt64, 0, 0),
			usageRecord(t, "2026-10-28T10:00:03Z", "root-1", "root-1", "r2", math.MaxInt64, 0, 0),
		)
		t.Setenv("CODEX_HOME", home)
		code, out, errw := usageRun(t, repo, "usage")
		if code != exitOther || out != "" {
			t.Fatalf("exit %d stdout %q, want no report on a fatal path", code, out)
		}
		if errw == "" {
			t.Fatal("stderr is empty, want the overflow named")
		}
	})

	t.Run("broken git is operational", func(t *testing.T) {
		// A repo of its own: the shared one must keep working for the later
		// subtests.
		broken := usageRepo(t)
		home, _ := usageHome(t)
		t.Setenv("CODEX_HOME", home)
		cfg, err := os.OpenFile(filepath.Join(broken, ".git", "config"), os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fmt.Fprint(cfg, "this line has no equals sign\n"); err != nil {
			t.Fatal(err)
		}
		if err := cfg.Close(); err != nil {
			t.Fatal(err)
		}
		code, out, errw := usageRun(t, broken, "usage")
		if code != exitOther || out != "" {
			t.Fatalf("exit %d stdout %q, want the operational code and no report", code, out)
		}
		if !strings.Contains(errw, "bad config") {
			t.Fatalf("stderr = %q, want git's own reason", errw)
		}
	})

	t.Run("no home folder", func(t *testing.T) {
		t.Setenv("CODEX_HOME", "")
		t.Setenv("HOME", "")
		code, out, errw := usageRun(t, repo, "usage")
		if code != exitOther || out != "" {
			t.Fatalf("exit %d stdout %q, want the operational code", code, out)
		}
		if errw == "" {
			t.Fatal("stderr is empty, want the missing home named")
		}
	})

	t.Run("output writer failure", func(t *testing.T) {
		home, sessions := usageHome(t)
		usageLog(t, sessions, "root.jsonl",
			usageMeta(t, "root-1", repo),
			usageModel(t, "m"),
			usageRecord(t, "2026-10-28T10:00:02Z", "root-1", "root-1", "r1", 10, 0, 1),
		)
		t.Setenv("CODEX_HOME", home)
		var errw strings.Builder
		code := exitOK
		inDir(t, repo, func() {
			code = Run([]string{"usage"}, strings.NewReader(""), false, failWriter{}, &errw)
		})
		if code != exitOther {
			t.Fatalf("exit %d, want %d for a failed write", code, exitOther)
		}
	})
}

// The sessions folder follows CODEX_HOME: unset or empty means the home
// default, and a relative value is read from the working directory.
func TestUsageCommandCodexHome(t *testing.T) {
	repo := usageRepo(t)
	log := []string{
		usageMeta(t, "root-1", repo),
		usageModel(t, "m"),
		usageRecord(t, "2026-10-28T10:00:02Z", "root-1", "root-1", "r1", 10, 0, 1),
	}

	t.Run("empty means the home default", func(t *testing.T) {
		home := t.TempDir()
		sessions := filepath.Join(home, ".codex", "sessions")
		if err := os.MkdirAll(sessions, 0o755); err != nil {
			t.Fatal(err)
		}
		usageLog(t, sessions, "root.jsonl", log...)
		t.Setenv("HOME", home)
		t.Setenv("CODEX_HOME", "")
		code, out, errw := usageRun(t, repo, "usage")
		if code != exitOK || !strings.Contains(out, "root-1") {
			t.Fatalf("exit %d stderr %q stdout %q", code, errw, out)
		}
	})

	t.Run("relative is read from the working directory", func(t *testing.T) {
		usageLog(t, filepath.Join(repo, "cx", "sessions"), "root.jsonl", log...)
		t.Setenv("CODEX_HOME", "cx")
		code, out, errw := usageRun(t, repo, "usage")
		if code != exitOK || !strings.Contains(out, "root-1") {
			t.Fatalf("exit %d stderr %q stdout %q", code, errw, out)
		}
	})
}

// A session identity from a transcript cannot paint the terminal.
func TestUsageCommandEscapesTerminalText(t *testing.T) {
	repo := usageRepo(t)
	home, sessions := usageHome(t)
	usageLog(t, sessions, "root.jsonl",
		usageMeta(t, "root\x1b[31mevil", repo),
		usageModel(t, "m\x07odel"),
		usageRecord(t, "2026-10-28T10:00:02Z", "root\x1b[31mevil", "root\x1b[31mevil", "r1", 10, 0, 1),
	)
	t.Setenv("CODEX_HOME", home)
	code, out, errw := usageRun(t, repo, "usage")
	if code != exitOK {
		t.Fatalf("exit %d stderr %q", code, errw)
	}
	if strings.ContainsRune(out, 0x1b) || strings.ContainsRune(out, 0x07) {
		t.Fatalf("output kept raw control characters:\n%q", out)
	}
	if !strings.Contains(out, `\u001b[31mevil`) || !strings.Contains(out, `m\u0007odel`) {
		t.Fatalf("output does not show the escaped names:\n%s", out)
	}
}

// The human renderer labels a missing range end unknown, prints the model
// summary, and reports a write failure.
func TestUsageRender(t *testing.T) {
	r := usage.Report{
		SchemaVersion: 1,
		Status:        usage.StatusComplete,
		Scope:         usage.Scope{Repository: "/example/repo", Checkouts: []string{"/example/repo"}, SourceDirectory: "/example/codex/sessions", Limitations: []string{usage.HistoryLimitation}},
		Range:         usage.Span{},
		Sessions: []usage.SessionRow{{
			SessionID:     "root-1",
			Model:         "m1",
			Range:         usage.Span{},
			ResponseCount: 1,
		}},
		Models: []usage.ModelRow{{Model: "m1", ResponseCount: 1}},
	}
	var out strings.Builder
	if err := renderUsage(&out, r); err != nil {
		t.Fatalf("renderUsage: %v", err)
	}
	text := out.String()
	for _, want := range []string{"SESSION", "MODEL", "unknown", "m1", "deleted or moved"} {
		if !strings.Contains(text, want) {
			t.Errorf("output misses %q:\n%s", want, text)
		}
	}
	if err := renderUsage(failWriter{}, r); err == nil {
		t.Fatal("renderUsage: got no error from a failed write")
	}

	noData := usage.Report{Status: usage.StatusNoData, Sessions: []usage.SessionRow{}, Models: []usage.ModelRow{},
		Scope: usage.Scope{SourceDirectory: "/example/codex/sessions", Limitations: []string{usage.HistoryLimitation}}}
	out.Reset()
	if err := renderUsage(&out, noData); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(out.String()), "no data") {
		t.Fatalf("no data output = %q", out.String())
	}
	if !strings.Contains(out.String(), "deleted or moved") {
		t.Fatalf("no data output drops the scope limitation: %q", out.String())
	}
}
