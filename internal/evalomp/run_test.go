package evalomp

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeOmp writes a stand-in omp that records its arguments, its folder and
// PM_VOICE_FILE next to itself, then prints the fixture stream. Tests never
// run the real omp, so they cost no model quota.
func fakeOmp(t *testing.T, body string) (omp, logDir string) {
	t.Helper()
	logDir = t.TempDir()
	omp = filepath.Join(logDir, "omp")
	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$@\" > \"" + logDir + "/args\"\n" +
		"pwd > \"" + logDir + "/pwd\"\n" +
		"ls -A > \"" + logDir + "/ls\"\n" +
		"printf '%s' \"$PM_VOICE_FILE\" > \"" + logDir + "/voice\"\n" +
		body + "\n"
	if err := os.WriteFile(omp, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return omp, logDir
}

func fixture(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs("testdata/stream.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	return "cat \"" + abs + "\""
}

func read(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestSkillNames(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "skills/plan/SKILL.md"), "x")
	write(t, filepath.Join(dir, "skills/brainstorm/SKILL.md"), "x")
	write(t, filepath.Join(dir, "skills/README.md"), "x")
	got, err := SkillNames(dir)
	if err != nil || strings.Join(got, ",") != "brainstorm,plan" {
		t.Errorf("SkillNames = %v, %v", got, err)
	}
}

// A plugin without a skills folder is a broken checkout, not an empty run.
func TestSkillNamesMissingFolder(t *testing.T) {
	if _, err := SkillNames(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Error("want an error when the plugin has no skills folder")
	}
}

// withEnv must drop the old value, or the child may read the user's HOME
// instead of the throwaway one.
func TestWithEnv(t *testing.T) {
	got := withEnv([]string{"PATH=/bin", "HOME=/real"}, "HOME", "/throwaway")
	want := map[string]bool{"PATH=/bin": true, "HOME=/throwaway": true}
	if len(got) != len(want) {
		t.Fatalf("env = %v, want only %v", got, want)
	}
	for _, kv := range got {
		if !want[kv] {
			t.Errorf("env = %v, has %q that the child must not see", got, kv)
		}
	}
}

func TestRunCaseCommandLine(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
	omp, logs := fakeOmp(t, fixture(t))
	o := Options{Omp: omp, PluginDir: "/p/plugin", Skills: []string{"brainstorm", "plan"}}
	w, err := RunCase(Case{Name: "c", Prompt: "Do it.", TimeoutSeconds: 30}, o)
	if err != nil {
		t.Fatal(err)
	}
	want := "-p\n--mode\njson\n--no-session\n--no-extensions\n--no-rules\n--skills=brainstorm,plan\n" +
		"--plugin-dir\n/p/plugin\n-e\n/p/plugin/omp/index.ts\nDo it.\n"
	if got := read(t, filepath.Join(logs, "args")); got != want {
		t.Errorf("args =\n%s\nwant\n%s", got, want)
	}
	if strings.TrimSpace(read(t, filepath.Join(logs, "ls"))) != "" {
		t.Error("omp must start in an empty folder when there is no scaffold")
	}
	pwd, _ := filepath.EvalSymlinks(strings.TrimSpace(read(t, filepath.Join(logs, "pwd"))))
	dir, _ := filepath.EvalSymlinks(w.Dir)
	if pwd != dir {
		t.Errorf("omp ran in %s, workspace is %s", pwd, dir)
	}
	voice := read(t, filepath.Join(logs, "voice"))
	if voice != filepath.Join(filepath.Dir(w.Dir), "home", ".acta", "config.yaml") {
		t.Errorf("PM_VOICE_FILE = %q", voice)
	}
	if w.Result.Reply == "" || len(w.Result.Calls) != 2 {
		t.Errorf("result = %+v", w.Result)
	}
}

// A case with no skills must not ask omp to load some other skill set.
func TestRunCaseNoSkills(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
	omp, logs := fakeOmp(t, fixture(t))
	if _, err := RunCase(Case{Prompt: "x", TimeoutSeconds: 30}, Options{Omp: omp, PluginDir: "/p"}); err != nil {
		t.Fatal(err)
	}
	want := "-p\n--mode\njson\n--no-session\n--no-extensions\n--no-rules\n--skills=\n" +
		"--plugin-dir\n/p\n-e\n/p/omp/index.ts\nx\n"
	if got := read(t, filepath.Join(logs, "args")); got != want {
		t.Errorf("args =\n%s\nwant\n%s", got, want)
	}
}

func TestRunCaseScaffold(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
	omp, logs := fakeOmp(t, fixture(t))
	dir := t.TempDir()
	scaffold := filepath.Join(dir, "scaffold.sh")
	write(t, scaffold, "set -e\n[ -z \"$(ls -A)\" ]\ntouch pre.txt\nmkdir -p \"$HOME/.acta\"\necho 'chat_language: English' > \"$HOME/.acta/config.yaml\"\n")
	w, err := RunCase(Case{Name: "c", Prompt: "x", TimeoutSeconds: 30, Scaffold: scaffold}, Options{Omp: omp, PluginDir: "/p"})
	if err != nil {
		t.Fatal(err)
	}
	if !w.Before["pre.txt"] {
		t.Errorf("before = %v, want pre.txt", w.Before)
	}
	if got := read(t, read(t, filepath.Join(logs, "voice"))); !strings.Contains(got, "English") {
		t.Errorf("voice file = %q, want what the scaffold wrote in the throwaway home", got)
	}
}

func TestRunCaseFailures(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
	broken := filepath.Join(t.TempDir(), "bad.sh")
	write(t, broken, "exit 4\n")
	cases := map[string]struct {
		body string
		c    Case
	}{
		"scaffold fails": {fixture(t), Case{Prompt: "x", TimeoutSeconds: 30, Scaffold: broken}},
		"omp fails":      {"echo nope >&2; exit 2", Case{Prompt: "x", TimeoutSeconds: 30}},
		"broken stream":  {"echo '{\"type\":\"agent_start\"}'", Case{Prompt: "x", TimeoutSeconds: 30}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			omp, _ := fakeOmp(t, tc.body)
			if _, err := RunCase(tc.c, Options{Omp: omp, PluginDir: "/p"}); err == nil {
				t.Error("want an error")
			}
		})
	}
	t.Run("timeout", func(t *testing.T) {
		omp, _ := fakeOmp(t, "exec sleep 20")
		_, err := RunCase(Case{Prompt: "x", TimeoutSeconds: 1}, Options{Omp: omp, PluginDir: "/p"})
		if !errors.Is(err, ErrTimeout) {
			t.Errorf("err = %v, want ErrTimeout", err)
		}
	})
}

// A failure with nothing to say is still a failure, so the message has to name
// the step that broke and what the step printed.
func TestRunCaseErrorText(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
	bad := filepath.Join(t.TempDir(), "bad.sh")
	write(t, bad, "echo scaffold-said >&2\nexit 4\n")
	t.Run("scaffold", func(t *testing.T) {
		omp, _ := fakeOmp(t, fixture(t))
		_, err := RunCase(Case{Prompt: "x", TimeoutSeconds: 30, Scaffold: bad}, Options{Omp: omp, PluginDir: "/p"})
		if err == nil {
			t.Fatal("want an error")
		}
		if msg := err.Error(); !strings.Contains(msg, "scaffold") || !strings.Contains(msg, "scaffold-said") {
			t.Errorf("err = %q, want it to name the scaffold and its output", msg)
		}
	})
	t.Run("omp", func(t *testing.T) {
		omp, _ := fakeOmp(t, "echo boom >&2; exit 2")
		_, err := RunCase(Case{Prompt: "x", TimeoutSeconds: 30}, Options{Omp: omp, PluginDir: "/p"})
		if err == nil {
			t.Fatal("want an error")
		}
		if msg := err.Error(); !strings.Contains(msg, "omp") || !strings.Contains(msg, "boom") {
			t.Errorf("err = %q, want it to name omp and its stderr", msg)
		}
	})
}

// The judge reads one reply with a bare omp: no plugin, no skills, no rules,
// or it would grade the plugin instead of the reply.
func TestOmpJudge(t *testing.T) {
	omp, logs := fakeOmp(t, "echo graded")
	got, err := OmpJudge(omp)("does it pass?")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(got) != "graded" {
		t.Errorf("judge reply = %q", got)
	}
	want := "-p\n--no-session\n--no-extensions\n--no-rules\n--no-skills\ndoes it pass?\n"
	if args := read(t, filepath.Join(logs, "args")); args != want {
		t.Errorf("judge args =\n%s\nwant\n%s", args, want)
	}
}

func TestRunAll(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	omp, _ := fakeOmp(t, fixture(t))
	cases := []Case{
		{Name: "good", Prompt: "x", TimeoutSeconds: 30, Graders: []Grader{{Name: "said", Type: "regex", Pattern: "SCR-0001"}}},
		{Name: "bad", Prompt: "x", TimeoutSeconds: 30, Graders: []Grader{{Name: "missing", Type: "regex", Pattern: "NOPE"}}},
		{Name: "claude", Tags: []string{"claude-only"}, Prompt: "x"},
	}
	var out bytes.Buffer
	failed := RunAll(cases, Options{Omp: omp, PluginDir: "/p"}, "", nil, &out)
	got := out.String()
	for _, line := range []string{"PASS good", "FAIL bad: missing (", "SKIP claude (claude-only)", "1 passed, 1 failed, 1 skipped"} {
		if !strings.Contains(got, line) {
			t.Errorf("output lacks %q:\n%s", line, got)
		}
	}
	if !failed {
		t.Error("a failed case must make RunAll report failure")
	}
	left, _ := filepath.Glob(filepath.Join(tmp, "acta-eval-omp-*"))
	if len(left) != 0 {
		t.Errorf("throwaway folders left behind: %v", left)
	}

	out.Reset()
	if RunAll(cases, Options{Omp: omp, PluginDir: "/p"}, "go*", nil, &out) {
		t.Errorf("only the good case ran, so nothing failed:\n%s", out.String())
	}
	if strings.Contains(out.String(), "bad") {
		t.Errorf("--case filter let another case run:\n%s", out.String())
	}
}

// Every case ends as one line and the total, never two lines and never none.
func TestRunAllOneLinePerCase(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
	omp, _ := fakeOmp(t, fixture(t))
	cases := []Case{
		{Name: "good", Prompt: "x", TimeoutSeconds: 30, Graders: []Grader{{Name: "said", Type: "regex", Pattern: "SCR-0001"}}},
		{Name: "bad", Prompt: "x", TimeoutSeconds: 30, Graders: []Grader{{Name: "missing", Type: "regex", Pattern: "NOPE"}}},
		{Name: "claude", Tags: []string{"claude-only"}, Prompt: "x"},
	}
	var out bytes.Buffer
	RunAll(cases, Options{Omp: omp, PluginDir: "/p"}, "", nil, &out)
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("got %d lines, want one per case and a total:\n%s", len(lines), out.String())
	}
	if !strings.HasPrefix(lines[0], "PASS good") {
		t.Errorf("line 1 = %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "FAIL bad: missing (") {
		t.Errorf("line 2 = %q", lines[1])
	}
	if lines[2] != "SKIP claude (claude-only)" {
		t.Errorf("line 3 = %q", lines[2])
	}
	if lines[3] != "1 passed, 1 failed, 1 skipped" {
		t.Errorf("line 4 = %q", lines[3])
	}
}

// A case that dies mid-run still prints its one line, and still cleans up.
func TestRunAllFailurePaths(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	broken := filepath.Join(t.TempDir(), "bad.sh")
	write(t, broken, "exit 4\n")
	for name, tc := range map[string]struct {
		body string
		c    Case
	}{
		"scaffold fails": {fixture(t), Case{Name: "s", Prompt: "x", TimeoutSeconds: 30, Scaffold: broken}},
		"omp fails":      {"echo nope >&2; exit 2", Case{Name: "s", Prompt: "x", TimeoutSeconds: 30}},
		"broken stream":  {"echo '{\"type\":\"agent_start\"}'", Case{Name: "s", Prompt: "x", TimeoutSeconds: 30}},
		"timeout":        {"exec sleep 20", Case{Name: "s", Prompt: "x", TimeoutSeconds: 1}},
	} {
		t.Run(name, func(t *testing.T) {
			omp, _ := fakeOmp(t, tc.body)
			var out bytes.Buffer
			failed := RunAll([]Case{tc.c}, Options{Omp: omp, PluginDir: "/p"}, "", nil, &out)
			if !failed {
				t.Errorf("a case that broke must make RunAll report failure:\n%s", out.String())
			}
			lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
			if len(lines) != 2 {
				t.Fatalf("got %d lines, want one FAIL line and a total:\n%s", len(lines), out.String())
			}
			if !strings.HasPrefix(lines[0], "FAIL s: ") {
				t.Errorf("line 1 = %q", lines[0])
			}
			if lines[1] != "0 passed, 1 failed, 0 skipped" {
				t.Errorf("line 2 = %q", lines[1])
			}
			if left, _ := filepath.Glob(filepath.Join(tmp, "acta-eval-omp-*")); len(left) != 0 {
				t.Errorf("throwaway folders left behind: %v", left)
			}
		})
	}
}
