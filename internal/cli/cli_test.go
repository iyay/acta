package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// debtRepo makes a git repo with .pm/plans/2026-09-26-short-ids.md, a plan
// with an id and hash already set, so acta debt new has a plan to attach to.
func debtRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	t.Setenv("GIT_AUTHOR_NAME", "test")
	t.Setenv("GIT_COMMITTER_NAME", "test")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")
	run("init", "-q", "-b", "main")
	plan := filepath.Join(dir, ".pm", "plans", "2026-09-26-short-ids.md")
	if err := os.MkdirAll(filepath.Dir(plan), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plan, []byte("---\nid: PLAN-3\nhash: k3f2\n---\n# Short IDs\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-q", "-m", "init")
	return dir
}

// inDir runs fn with the process cwd set to dir, then restores it. Run
// resolves the repo from os.Getwd, so an end-to-end CLI test needs this.
func inDir(t *testing.T, dir string, fn func()) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	fn()
}

// runCode runs the CLI and returns its exit code, ignoring the output.
// The voice command needs no repo, so no inDir here.
func runCode(t *testing.T, args ...string) int {
	t.Helper()
	var stdout, stderr strings.Builder
	return Run(args, strings.NewReader(""), false, &stdout, &stderr)
}

// mustRun runs the CLI, fails unless it exits OK, and returns stdout.
func mustRun(t *testing.T, args ...string) string {
	t.Helper()
	var stdout, stderr strings.Builder
	if code := Run(args, strings.NewReader(""), false, &stdout, &stderr); code != exitOK {
		t.Fatalf("%v: exit %d stderr %q", args, code, stderr.String())
	}
	return stdout.String()
}

func TestDebtNewWritesFileAndCommits(t *testing.T) {
	dir := debtRepo(t)
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		code := Run([]string{"debt", "new", "PLAN-3"}, strings.NewReader("x\n"), false, &stdout, &stderr)
		if code != exitOK {
			t.Fatalf("exit %d stderr %q", code, stderr.String())
		}
	})
	entries, err := os.ReadDir(filepath.Join(dir, ".pm", "debt"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("debt dir entries=%v err=%v", entries, err)
	}
	body, _ := os.ReadFile(filepath.Join(dir, ".pm", "debt", entries[0].Name()))
	if !strings.Contains(string(body), "- [ ] x\n") {
		t.Fatalf("file lost its note: %q", body)
	}
	if out, err := exec.Command("git", "-C", dir, "log", "-1", "--format=%s").CombinedOutput(); err != nil || !strings.HasPrefix(strings.TrimSpace(string(out)), "acta: new debt ") {
		t.Fatalf("commit message %q err %v", out, err)
	}
}

func TestDebtNewEmptyStdinExitsBadInput(t *testing.T) {
	dir := debtRepo(t)
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		code := Run([]string{"debt", "new", "PLAN-3"}, strings.NewReader(""), false, &stdout, &stderr)
		if code != exitBadInput {
			t.Fatalf("exit %d stderr %q", code, stderr.String())
		}
	})
	if _, err := os.Stat(filepath.Join(dir, ".pm", "debt")); !os.IsNotExist(err) {
		t.Fatal("debt folder must not exist after rejected input")
	}
}

func TestDebtNewTTYRefusesWithoutReadingStdin(t *testing.T) {
	dir := debtRepo(t)
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		code := Run([]string{"debt", "new", "PLAN-3"}, strings.NewReader("x\n"), true, &stdout, &stderr)
		if code != exitBadInput {
			t.Fatalf("exit %d stderr %q", code, stderr.String())
		}
		if !strings.Contains(stderr.String(), "pipe NOTEs on stdin") {
			t.Fatalf("stderr %q", stderr.String())
		}
	})
}

func TestDebtNoSubcommandExitsUsage(t *testing.T) {
	dir := debtRepo(t)
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		code := Run([]string{"debt"}, strings.NewReader(""), false, &stdout, &stderr)
		if code != exitBadInput {
			t.Fatalf("exit %d stderr %q", code, stderr.String())
		}
		if !strings.Contains(stderr.String(), "usage: acta debt new") {
			t.Fatalf("stderr %q", stderr.String())
		}
	})
}

func TestScratchNewWritesFileAndCommits(t *testing.T) {
	dir := debtRepo(t)
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		code := Run([]string{"scratch", "new", "newest-first"}, strings.NewReader("sort the list ✓\n"), false, &stdout, &stderr)
		if code != exitOK {
			t.Fatalf("exit %d stderr %q", code, stderr.String())
		}
	})
	if !strings.Contains(stdout.String(), "SCRATCH-1") {
		t.Fatalf("stdout %q lacks the short id", stdout.String())
	}
	files, err := filepath.Glob(filepath.Join(dir, ".pm", "scratch", "*-newest-first.md"))
	if err != nil || len(files) != 1 {
		t.Fatalf("scratch files %v err %v", files, err)
	}
	if !strings.Contains(stdout.String(), "scratch/") {
		t.Fatalf("stdout %q lacks the path", stdout.String())
	}
	body, _ := os.ReadFile(files[0])
	if !strings.Contains(string(body), "status: raw") || !strings.HasSuffix(string(body), "sort the list ✓\n") {
		t.Fatalf("file = %q", body)
	}
	if out, err := exec.Command("git", "-C", dir, "log", "-1", "--format=%s").CombinedOutput(); err != nil ||
		!strings.HasPrefix(strings.TrimSpace(string(out)), "acta: new scratch ") {
		t.Fatalf("commit message %q err %v", out, err)
	}
}

func TestScratchAddAppendsAndCommits(t *testing.T) {
	dir := debtRepo(t)
	var out, errOut strings.Builder
	inDir(t, dir, func() {
		if code := Run([]string{"scratch", "new", "idea"}, strings.NewReader("first note\n"), false, &out, &errOut); code != exitOK {
			t.Fatalf("scratch new exit %d stderr %q", code, errOut.String())
		}
		out.Reset()
		code := Run([]string{"scratch", "add", "SCRATCH-1"}, strings.NewReader("answer ✓\n"), false, &out, &errOut)
		if code != exitOK {
			t.Fatalf("exit %d stderr %q", code, errOut.String())
		}
	})
	if !strings.Contains(out.String(), "SCRATCH-1") {
		t.Fatalf("stdout %q lacks the short id", out.String())
	}
	files, _ := filepath.Glob(filepath.Join(dir, ".pm", "scratch", "*-idea.md"))
	body, _ := os.ReadFile(files[0])
	if !strings.HasSuffix(string(body), "first note\n\nanswer ✓\n") {
		t.Fatalf("file = %q", body)
	}
	if msg, _ := exec.Command("git", "-C", dir, "log", "-1", "--format=%s").CombinedOutput(); !strings.HasPrefix(strings.TrimSpace(string(msg)), "acta: add to scratch ") {
		t.Fatalf("commit message %q", msg)
	}
}

func TestScratchNewBadSlugExitsBadInput(t *testing.T) {
	dir := debtRepo(t)
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		code := Run([]string{"scratch", "new", "Bad Slug"}, strings.NewReader("x\n"), false, &stdout, &stderr)
		if code != exitBadInput {
			t.Fatalf("exit %d stderr %q", code, stderr.String())
		}
	})
	if _, err := os.Stat(filepath.Join(dir, ".pm", "scratch")); !os.IsNotExist(err) {
		t.Fatal("the scratch folder must not exist after a bad slug")
	}
}

func TestScratchNoSubcommandExitsUsage(t *testing.T) {
	dir := debtRepo(t)
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		code := Run([]string{"scratch"}, strings.NewReader(""), false, &stdout, &stderr)
		if code != exitBadInput {
			t.Fatalf("exit %d stderr %q", code, stderr.String())
		}
		for _, want := range []string{"usage: acta scratch new", "scratch add"} {
			if !strings.Contains(stderr.String(), want) {
				t.Fatalf("stderr %q lacks %q", stderr.String(), want)
			}
		}
	})
}

func TestScratchUnknownSubcommandNamesBothCommands(t *testing.T) {
	dir := debtRepo(t)
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		code := Run([]string{"scratch", "drop"}, strings.NewReader(""), false, &stdout, &stderr)
		if code != exitBadInput {
			t.Fatalf("exit %d stderr %q", code, stderr.String())
		}
		for _, want := range []string{"scratch new", "scratch add"} {
			if !strings.Contains(stderr.String(), want) {
				t.Fatalf("stderr %q lacks %q", stderr.String(), want)
			}
		}
	})
}

func TestVoiceSetExecutorKeepsOtherFields(t *testing.T) {
	t.Setenv("PM_VOICE_FILE", filepath.Join(t.TempDir(), "voice.yaml"))
	mustRun(t, "voice", "set", "--language", "Korean", "--tone", "short")
	if out := mustRun(t, "voice", "show"); strings.Contains(out, "build_executor") {
		t.Errorf("show printed an unset build_executor:\n%s", out)
	}
	mustRun(t, "voice", "set", "--executor", "dispatch")
	mustRun(t, "voice", "set", "--subagent-models", "split")
	out := mustRun(t, "voice", "show")
	for _, want := range []string{"chat_language: Korean", "tone: short", "build_executor: dispatch", "subagent_models: split"} {
		if !strings.Contains(out, want) {
			t.Errorf("show missing %q:\n%s", want, out)
		}
	}
	// The JSON is indented, so compare without spaces.
	asJSON := strings.ReplaceAll(mustRun(t, "voice", "show", "--json"), " ", "")
	for _, key := range []string{`"build_executor":"dispatch"`, `"subagent_models":"split"`} {
		if !strings.Contains(asJSON, key) {
			t.Errorf("json missing %q:\n%s", key, asJSON)
		}
	}
	// Clearing the models must leave every other field alone.
	mustRun(t, "voice", "set", "--clear-subagent-models")
	after := mustRun(t, "voice", "show")
	if strings.Contains(after, "subagent_models") {
		t.Errorf("clear left subagent_models:\n%s", after)
	}
	for _, want := range []string{"chat_language: Korean", "tone: short", "build_executor: dispatch"} {
		if !strings.Contains(after, want) {
			t.Errorf("clear dropped %q:\n%s", want, after)
		}
	}
	// Clear then set in one call: the set wins, so the user ends up with split.
	mustRun(t, "voice", "set", "--clear-subagent-models", "--subagent-models", "split")
	if out := mustRun(t, "voice", "show"); !strings.Contains(out, "subagent_models: split") {
		t.Errorf("clear-then-set did not keep split:\n%s", out)
	}
}

func TestVoiceSetBadExecutor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "voice.yaml")
	t.Setenv("PM_VOICE_FILE", path)
	for _, args := range [][]string{
		{"voice", "set", "--executor", "robot"},
		{"voice", "set", "--subagent-models", "all"},
		{"voice", "set", "--language", "Korean", "--executor", "omp"},
	} {
		if code := runCode(t, args...); code != exitBadInput {
			t.Errorf("%v: exit %d, want %d", args, code, exitBadInput)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("%v: a bad value was written to the voice file", args)
		}
	}
}
