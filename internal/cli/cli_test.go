package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"io"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/write"
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
	run("init", "-q", "-b", "main")
	// Name the author inside the repo, not in the env, so tests that make
	// commits can still run side by side.
	run("config", "user.name", "test")
	run("config", "user.email", "test@example.com")
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
// The config command needs no repo, so no inDir here.
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
	fixDay(t, "2026-09-26")
	dir := debtRepo(t)
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		code := Run([]string{"scratch", "new", "newest-first"}, strings.NewReader("sort the list ✓\n"), false, &stdout, &stderr)
		if code != exitOK {
			t.Fatalf("exit %d stderr %q", code, stderr.String())
		}
	})
	if !strings.Contains(stdout.String(), "SCR-0001") {
		t.Fatalf("stdout %q lacks the short id", stdout.String())
	}
	files, err := filepath.Glob(filepath.Join(dir, ".pm", "scratch", "*-newest-first.md"))
	if err != nil || len(files) != 1 {
		t.Fatalf("scratch files %v err %v", files, err)
	}
	if !strings.Contains(stdout.String(), "scratch/") {
		t.Fatalf("stdout %q lacks the path", stdout.String())
	}
	src, _ := os.ReadFile(files[0])
	doc := board.Parse(src)
	if !strings.Contains(string(src), "status: raw") || !board.HasSchema(doc.Front) || doc.Front["created"] != "2026-09-26 00:00:00" {
		t.Fatalf("file = %q", src)
	}
	if want := "# newest-first\n\n## Words\n\n### 2026-09-26\n\nsort the list ✓\n\n## Context\n\n## Log\n\n## Open questions\n"; doc.Body != want {
		t.Fatalf("body %q want %q", doc.Body, want)
	}
	if out, err := exec.Command("git", "-C", dir, "log", "-1", "--format=%s").CombinedOutput(); err != nil ||
		!strings.HasPrefix(strings.TrimSpace(string(out)), "acta: new scratch ") {
		t.Fatalf("commit message %q err %v", out, err)
	}
}

func TestScratchAddAppendsAndCommits(t *testing.T) {
	fixDay(t, "2026-09-26")
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
	if !strings.Contains(out.String(), "SCR-0001") {
		t.Fatalf("stdout %q lacks the short id", out.String())
	}
	files, _ := filepath.Glob(filepath.Join(dir, ".pm", "scratch", "*-idea.md"))
	src, _ := os.ReadFile(files[0])
	if want := "# idea\n\n## Words\n\n### 2026-09-26\n\nfirst note\n\n### 2026-09-26\n\nanswer ✓\n\n## Context\n\n## Log\n\n## Open questions\n"; board.Parse(src).Body != want {
		t.Fatalf("body %q want %q", src, want)
	}
	if msg, _ := exec.Command("git", "-C", dir, "log", "-1", "--format=%s").CombinedOutput(); !strings.HasPrefix(strings.TrimSpace(string(msg)), "acta: add to scratch ") {
		t.Fatalf("commit message %q", msg)
	}
}

// fixDay pins the clock so the dated headings in a file can be compared
// byte for byte.
func fixDay(t *testing.T, day string) {
	t.Helper()
	old := write.Now
	when, err := time.Parse("2006-01-02", day)
	if err != nil {
		t.Fatal(err)
	}
	write.Now = func() time.Time { return when }
	t.Cleanup(func() { write.Now = old })
}

func TestScratchAddSectionPutsTextInThatSection(t *testing.T) {
	fixDay(t, "2026-09-26")
	dir := debtRepo(t)
	var out, errOut strings.Builder
	inDir(t, dir, func() {
		if code := Run([]string{"scratch", "new", "idea", "--title", "Idea"}, strings.NewReader("first note\n"), false, &out, &errOut); code != exitOK {
			t.Fatalf("scratch new exit %d stderr %q", code, errOut.String())
		}
		out.Reset()
		code := Run([]string{"scratch", "add", "SCRATCH-1", "--section", "context"}, strings.NewReader("more ✓\n"), false, &out, &errOut)
		if code != exitOK {
			t.Fatalf("exit %d stderr %q", code, errOut.String())
		}
	})
	files, _ := filepath.Glob(filepath.Join(dir, ".pm", "scratch", "*-idea.md"))
	src, _ := os.ReadFile(files[0])
	want := "# Idea\n\n## Words\n\n### 2026-09-26\n\nfirst note\n\n## Context\n\nmore ✓\n\n## Log\n\n## Open questions\n"
	if body := board.Parse(src).Body; body != want {
		t.Fatalf("body %q want %q", body, want)
	}
}

func TestScratchAddBadSectionExitsBadInput(t *testing.T) {
	fixDay(t, "2026-09-26")
	dir := debtRepo(t)
	var out, errOut strings.Builder
	inDir(t, dir, func() {
		if code := Run([]string{"scratch", "new", "idea"}, strings.NewReader("first note\n"), false, &out, &errOut); code != exitOK {
			t.Fatalf("scratch new exit %d stderr %q", code, errOut.String())
		}
		out.Reset()
		code := Run([]string{"scratch", "add", "SCRATCH-1", "--section", "bogus"}, strings.NewReader("more\n"), false, &out, &errOut)
		if code != exitBadInput {
			t.Fatalf("exit %d want %d", code, exitBadInput)
		}
	})
	want := `unknown section "bogus"; use words, context, log or questions`
	if !strings.Contains(errOut.String(), want) {
		t.Fatalf("stderr %q lacks %q", errOut.String(), want)
	}
	files, _ := filepath.Glob(filepath.Join(dir, ".pm", "scratch", "*-idea.md"))
	src, _ := os.ReadFile(files[0])
	if body := board.Parse(src).Body; strings.Contains(body, "more") {
		t.Fatalf("the refused write changed the body: %q", body)
	}
	if n, _ := exec.Command("git", "-C", dir, "log", "--format=%s").Output(); !strings.HasPrefix(strings.TrimSpace(string(n)), "acta: new scratch") {
		t.Fatalf("the refused write made a commit: %q", n)
	}
}

func TestScratchAddHelpNamesTheSectionFlag(t *testing.T) {
	dir := debtRepo(t)
	var out, errOut strings.Builder
	inDir(t, dir, func() {
		if code := Run([]string{"scratch", "add", "-h"}, strings.NewReader(""), false, &out, &errOut); code != exitBadInput {
			t.Fatalf("exit %d stderr %q", code, errOut.String())
		}
		for _, want := range []string{
			"usage: acta scratch add <SCRATCH-n> [--section words|context|log|questions] < text.md",
			"words, context, log or questions (default: words)",
		} {
			if !strings.Contains(errOut.String(), want) {
				t.Fatalf("stderr %q lacks %q", errOut.String(), want)
			}
		}
	})
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
		for _, want := range []string{"usage: acta scratch new", "scratch add <SCRATCH-n> [--section words|context|log|questions] < text.md"} {
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
	// The acta repo has its own .acta.yaml. Run outside it so its values
	// do not show up here.
	t.Chdir(t.TempDir())
	t.Setenv("PM_VOICE_FILE", filepath.Join(t.TempDir(), "voice.yaml"))
	mustRun(t, "config", "set", "--language", "Korean", "--tone", "short")
	if out := mustRun(t, "config", "show"); strings.Contains(out, "build_executor") {
		t.Errorf("show printed an unset build_executor:\n%s", out)
	}
	mustRun(t, "config", "set", "--executor", "dispatch")
	mustRun(t, "config", "set", "--subagent-models", "split")
	out := mustRun(t, "config", "show")
	for _, want := range []string{"chat_language: Korean", "tone: short", "build_executor: dispatch", "subagent_models: split"} {
		if !strings.Contains(out, want) {
			t.Errorf("show missing %q:\n%s", want, out)
		}
	}
	// The JSON is indented, so compare without spaces.
	asJSON := strings.ReplaceAll(mustRun(t, "config", "show", "--json"), " ", "")
	for _, key := range []string{`"build_executor":"dispatch"`, `"subagent_models":"split"`} {
		if !strings.Contains(asJSON, key) {
			t.Errorf("json missing %q:\n%s", key, asJSON)
		}
	}
	// Clearing the models must leave every other field alone.
	mustRun(t, "config", "set", "--clear-subagent-models")
	after := mustRun(t, "config", "show")
	if strings.Contains(after, "subagent_models") {
		t.Errorf("clear left subagent_models:\n%s", after)
	}
	for _, want := range []string{"chat_language: Korean", "tone: short", "build_executor: dispatch"} {
		if !strings.Contains(after, want) {
			t.Errorf("clear dropped %q:\n%s", want, after)
		}
	}
	// Clear then set in one call: the set wins, so the user ends up with split.
	mustRun(t, "config", "set", "--clear-subagent-models", "--subagent-models", "split")
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "subagent_models: split") {
		t.Errorf("clear-then-set did not keep split:\n%s", out)
	}
}

// A no in setup is saved as default, so the next setup run sees it as set.
func TestConfigSetSubagentModelsDefault(t *testing.T) {
	t.Setenv("PM_VOICE_FILE", filepath.Join(t.TempDir(), "config.yaml"))
	mustRun(t, "config", "set", "--subagent-models", "default")
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "subagent_models: default") {
		t.Errorf("show lacks subagent_models: default:\n%s", out)
	}
	// The JSON is indented, so compare without spaces.
	if out := strings.ReplaceAll(mustRun(t, "config", "show", "--json"), " ", ""); !strings.Contains(out, `"subagent_models":"default"`) {
		t.Errorf("json lacks subagent_models default:\n%s", out)
	}
	mustRun(t, "config", "set", "--clear-subagent-models")
	if out := mustRun(t, "config", "show"); strings.Contains(out, "subagent_models") {
		t.Errorf("clear left subagent_models:\n%s", out)
	}
	// The usage must name default too, or a user cannot find the value that stops the question.
	var out, errs strings.Builder
	Run([]string{"config", "set"}, strings.NewReader(""), false, &out, &errs)
	if !strings.Contains(errs.String(), "--subagent-models split|default") {
		t.Errorf("usage lacks --subagent-models split|default:\n%s", errs.String())
	}
}

func TestVoiceSetBadExecutor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "voice.yaml")
	t.Setenv("PM_VOICE_FILE", path)
	for _, args := range [][]string{
		{"config", "set", "--executor", "robot"},
		{"config", "set", "--subagent-models", "all"},
		{"config", "set", "--language", "Korean", "--executor", "omp"},
	} {
		if code := runCode(t, args...); code != exitBadInput {
			t.Errorf("%v: exit %d, want %d", args, code, exitBadInput)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("%v: a bad value was written to the voice file", args)
		}
	}
}

// voiceHome gives a test its own HOME and voice file, so neither the theme
// files nor the voice file of the machine the test runs on leak in.
func voiceHome(t *testing.T) (home, path string) {
	t.Helper()
	home = t.TempDir()
	t.Setenv("HOME", home)
	path = filepath.Join(home, "voice.yaml")
	t.Setenv("PM_VOICE_FILE", path)
	return home, path
}

// writeUserTheme puts a theme file where theme.Load looks for one.
func writeUserTheme(t *testing.T, home, name, body string) {
	t.Helper()
	dir := filepath.Join(home, ".acta", "themes")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

const goodTheme = `bg: "#101010"
fg: "#eeeeee"
ansi: ["#000000","#111111","#222222","#333333","#444444","#555555","#666666","#777777",
       "#888888","#999999","#aaaaaa","#bbbbbb","#cccccc","#dddddd","#eeeeee","#ffffff"]
`

func TestVoiceSetThemeShowsIt(t *testing.T) {
	voiceHome(t)
	mustRun(t, "config", "set", "--theme", "dracula")
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "theme: dracula") {
		t.Errorf("show is missing the theme line:\n%s", out)
	}
	asJSON := strings.ReplaceAll(mustRun(t, "config", "show", "--json"), " ", "")
	if !strings.Contains(asJSON, `"theme":"dracula"`) {
		t.Errorf("json is missing the theme:\n%s", asJSON)
	}
}

// A theme file the user wrote is a real theme, so --theme has to take it.
func TestVoiceSetThemeTakesAUserThemeFile(t *testing.T) {
	home, _ := voiceHome(t)
	writeUserTheme(t, home, "mine", goodTheme)
	mustRun(t, "config", "set", "--theme", "mine")
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "theme: mine") {
		t.Errorf("show is missing the theme line:\n%s", out)
	}
}

// A name theme.Load refuses must never reach the file, because the TUI would
// then fall back to another theme with nothing said about it.
func TestVoiceSetThemeRefusesWhatThemeLoadRefuses(t *testing.T) {
	home, path := voiceHome(t)
	mustRun(t, "config", "set", "--theme", "dracula")
	writeUserTheme(t, home, "broken", "bg: [")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"nope", "../x", "not a theme", "broken"} {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr strings.Builder
			code := Run([]string{"config", "set", "--theme", name}, strings.NewReader(""), false, &stdout, &stderr)
			if code != exitBadInput {
				t.Errorf("--theme %s: exit %d, want %d", name, code, exitBadInput)
			}
			if !strings.Contains(stderr.String(), name) {
				t.Errorf("--theme %s: the error does not name it: %s", name, stderr.String())
			}
			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Errorf("--theme %s changed the file:\n%s", name, after)
			}
		})
	}
}

// A bad theme refuses the whole run, so the other flag in the same call is
// not saved either. Half a change would leave the user guessing what stuck.
func TestVoiceRefusedThemeSavesNoOtherFlag(t *testing.T) {
	_, path := voiceHome(t)
	var stdout, stderr strings.Builder
	code := Run([]string{"config", "set", "--language", "Korean", "--theme", "nope"},
		strings.NewReader(""), false, &stdout, &stderr)
	if code != exitBadInput {
		t.Fatalf("exit %d, want %d", code, exitBadInput)
	}
	written, err := os.ReadFile(path)
	if !os.IsNotExist(err) {
		t.Fatalf("the run wrote a voice file anyway:\n%s", written)
	}
}

// The TUI reads its theme from the voice file. A file that does not parse
// still gives the default voice, so the board opens either way; only the
// name is left empty and the TUI paints the default.
func TestVoiceThemeFallsBackOnABrokenVoiceFile(t *testing.T) {
	home, path := voiceHome(t)
	if got := voiceTheme(); got != "" {
		t.Errorf("no voice file yet: theme = %q, want the empty default", got)
	}
	mustRun(t, "config", "set", "--theme", "dracula")
	if got := voiceTheme(); got != "dracula" {
		t.Errorf("theme = %q want dracula", got)
	}
	if err := os.WriteFile(path, []byte("language: [oops\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := voiceTheme(); got != "" {
		t.Errorf("broken voice file: theme = %q, want the empty default", got)
	}
	if home == "" {
		t.Fatal("voiceHome gave no home")
	}
}

// An empty --theme names no theme, so it is not a run that sets anything and
// the theme already saved stays as it is.
func TestVoiceSetEmptyThemeLeavesTheSavedOne(t *testing.T) {
	_, path := voiceHome(t)
	mustRun(t, "config", "set", "--theme", "dracula")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if code := runCode(t, "config", "set", "--theme", ""); code != exitBadInput {
		t.Errorf("exit %d, want %d", code, exitBadInput)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("the file changed:\n%s", after)
	}
}

func TestVoiceClearTheme(t *testing.T) {
	voiceHome(t)
	mustRun(t, "config", "set", "--theme", "dracula")
	mustRun(t, "config", "set", "--clear-theme")
	if out := mustRun(t, "config", "show"); strings.Contains(out, "theme:") {
		t.Errorf("clear left the theme:\n%s", out)
	}
	if asJSON := strings.ReplaceAll(mustRun(t, "config", "show", "--json"), " ", ""); !strings.Contains(asJSON, `"theme":""`) {
		t.Errorf("json theme is not empty:\n%s", asJSON)
	}
	// Clear then set in one call: the set wins, like the other clear pairs.
	mustRun(t, "config", "set", "--clear-theme", "--theme", "dracula")
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "theme: dracula") {
		t.Errorf("clear-then-set did not keep the theme:\n%s", out)
	}
}

// closesRepo makes a repo with a spec that closes a scratch item, that scratch
// item, and a spec that closes nothing, so a test can look at an item with
// closes, one with closed by and one with neither.
func closesRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		".acta/specs/2026-09-29-a.md":   "---\nid: SPEC-1\ncloses: [SCRATCH-1, scratch/2026-09-28-j]\n---\n# Spec A\n",
		".acta/scratch/2026-09-28-i.md": "---\nid: SCRATCH-1\n---\n# Idea\n",
		// No short id in the frontmatter, so a reader sees this one by its
		// path, the same way it is named in the closes list.
		".acta/scratch/2026-09-28-j.md": "# Another idea\n",
		".acta/specs/2026-09-29-b.md":   "---\nid: SPEC-2\n---\n# Spec B\n",
	}
	for p, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// showLine returns the line of a show that begins with the prefix, so a test
// can name the exact line a reader sees.
func showLine(t *testing.T, out, prefix string) (string, bool) {
	t.Helper()
	for _, ln := range strings.Split(out, "\n") {
		if strings.HasPrefix(ln, prefix) {
			return ln, true
		}
	}
	return "", false
}

func TestShowPrintsClosesAndClosedBy(t *testing.T) {
	dir := closesRepo(t)
	t.Chdir(dir)
	for _, c := range []struct {
		id, line string
		absent   []string
	}{
		{"SPEC-1", "closes: SCR-0001, scratch/2026-09-28-j", []string{"closed by:"}},
		{"SCRATCH-1", "closed by: SPC-0001", []string{"closes:"}},
		// This scratch file carries no short id, so the line names it by its
		// path, the same way the closes list wrote it.
		{"scratch/2026-09-28-j", "closed by: SPC-0001", []string{"closes:"}},
		{"SPEC-2", "", []string{"closes:", "closed by:"}},
	} {
		out := mustRun(t, "show", c.id)
		if c.line == "" {
			for _, a := range c.absent {
				if got, ok := showLine(t, out, a); ok {
					t.Errorf("show %s printed %q, want no such line:\n%s", c.id, got, out)
				}
			}
			continue
		}
		got, ok := showLine(t, out, c.line)
		if !ok {
			t.Errorf("show %s printed no %q:\n%s", c.id, c.line, out)
			continue
		}
		if got != c.line {
			t.Errorf("show %s printed %q, want %q", c.id, got, c.line)
		}
		for _, a := range c.absent {
			if _, ok := showLine(t, out, a); ok {
				t.Errorf("show %s printed a %q line it has no link for:\n%s", c.id, a, out)
			}
		}
	}
}

// The JSON always carries a list, so a reader never has to tell an empty link
// from a field that is missing.
func TestShowJSONCarriesClosesAndClosedByAsLists(t *testing.T) {
	dir := closesRepo(t)
	t.Chdir(dir)
	for _, c := range []struct {
		id      string
		closes  []string
		closedB []string
	}{
		{"SPEC-1", []string{"scratch/2026-09-28-i", "scratch/2026-09-28-j"}, []string{}},
		{"SCRATCH-1", []string{}, []string{"specs/2026-09-29-a"}},
		{"scratch/2026-09-28-j", []string{}, []string{"specs/2026-09-29-a"}},
		{"SPEC-2", []string{}, []string{}},
	} {
		var j struct {
			Closes   []string `json:"closes"`
			ClosedBy []string `json:"closed_by"`
		}
		if err := json.Unmarshal([]byte(mustRun(t, "show", c.id, "--json")), &j); err != nil {
			t.Fatalf("show %s --json: %v", c.id, err)
		}
		if j.Closes == nil || j.ClosedBy == nil {
			t.Errorf("show %s --json wrote a null link: closes %v closed_by %v", c.id, j.Closes, j.ClosedBy)
			continue
		}
		if !slices.Equal(j.Closes, c.closes) {
			t.Errorf("show %s --json closes = %v, want %v", c.id, j.Closes, c.closes)
		}
		if !slices.Equal(j.ClosedBy, c.closedB) {
			t.Errorf("show %s --json closed_by = %v, want %v", c.id, j.ClosedBy, c.closedB)
		}
	}
}

// pathRepo makes a repo whose planning files cover every way of naming an
// item: a new id, a hash, a sub-item number, an old id only the board alias
// knows, and an id written in quotes. root picks the folder name, so a moved
// root gets its own repo.
func pathRepo(t *testing.T, root string) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"specs/2026-10-01-a.md":   "---\nid: SPC-0054\nhash: az0vfpu\n---\n# Design\n",
		"plans/2026-10-01-b.md":   "---\nid: PLN-0001\nhash: bk4n2qp\n---\n# Plan\n\n### Task 1: Do it\n\n- [ ] **Step 1: Write the failing test**\n",
		"plans/2026-10-01-c.md":   "---\nid: PLN-0003\n---\n# Old plan\n",
		"scratch/2026-10-01-d.md": "---\nid: \"SCR-0028\"\n---\n# Idea\n",
	}
	for p, body := range files {
		full := filepath.Join(dir, root, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// The path a reader gets from --path must be the same one the plain show
// prints on its path line, so a reader can paste either one.
func TestShowPathPrintsTheFileForEveryInputForm(t *testing.T) {
	dir := pathRepo(t, ".acta")
	t.Chdir(dir)
	for _, c := range []struct {
		name, id, want string
	}{
		{"new id", "SPC-0054", ".acta/specs/2026-10-01-a.md"},
		{"full hash", "az0vfpu", ".acta/specs/2026-10-01-a.md"},
		{"quoted id", "SCR-0028", ".acta/scratch/2026-10-01-d.md"},
		{"sub item takes the plan file", "PLN-0001.01", ".acta/plans/2026-10-01-b.md"},
		{"hash prefix falls back", "az0v", ".acta/specs/2026-10-01-a.md"},
		{"prefixed hash falls back", "SPC-az0vfpu", ".acta/specs/2026-10-01-a.md"},
		{"old id falls back", "PLAN-3", ".acta/plans/2026-10-01-c.md"},
	} {
		t.Run(c.name, func(t *testing.T) {
			out := mustRun(t, "show", c.id, "--path")
			if out != c.want+"\n" {
				t.Errorf("show %s --path = %q, want one line %q", c.id, out, c.want)
			}
			line, ok := showLine(t, mustRun(t, "show", c.id), "path: ")
			if !ok || line != "path: "+c.want {
				t.Errorf("show %s path line = %q, want %q", c.id, line, "path: "+c.want)
			}
		})
	}
}

// The root can move, so the scan has to read where the config says the files
// are, not where it expects them.
func TestShowPathFollowsAMovedRoot(t *testing.T) {
	dir := pathRepo(t, ".pm")
	t.Chdir(dir)
	want := ".pm/specs/2026-10-01-a.md\n"
	t.Setenv("ACTA_ROOT", "")
	if out := mustRun(t, "show", "SPC-0054", "--root", ".pm", "--path"); out != want {
		t.Errorf("--root gave %q, want %q", out, want)
	}
	t.Setenv("ACTA_ROOT", ".pm")
	if out := mustRun(t, "show", "SPC-0054", "--path"); out != want {
		t.Errorf("ACTA_ROOT gave %q, want %q", out, want)
	}
}

// A hit reads the folder, so it must work where there is no git to call.
func TestShowPathFindsANewIDAndHashWithoutGit(t *testing.T) {
	dir := pathRepo(t, ".acta")
	t.Chdir(dir)
	t.Setenv("PATH", t.TempDir())
	for id, want := range map[string]string{
		"SPC-0054": ".acta/specs/2026-10-01-a.md\n",
		"az0vfpu":  ".acta/specs/2026-10-01-a.md\n",
	} {
		if out := mustRun(t, "show", id, "--path"); out != want {
			t.Errorf("show %s --path with no git = %q, want %q", id, out, want)
		}
	}
}

// The board answers every form, so only calling the scan tells us the fast
// path is what runs. A hash prefix and an old id are the board's job: the scan
// matches whole values only, so it says nothing for them.
func TestScanPathMatchesWholeValuesOnly(t *testing.T) {
	dir := pathRepo(t, ".acta")
	t.Chdir(dir)
	cfg, code := loadConfig("", io.Discard)
	if code != exitOK {
		t.Fatalf("loadConfig exit %d", code)
	}
	for _, c := range []struct {
		id, want string
	}{
		{"SPC-0054", ".acta/specs/2026-10-01-a.md"},
		{"az0vfpu", ".acta/specs/2026-10-01-a.md"},
		{"SCR-0028", ".acta/scratch/2026-10-01-d.md"},
		{"PLN-0001.01", ".acta/plans/2026-10-01-b.md"},
		{"az0v", ""},
		{"SPC-az0vfpu", ""},
		{"PLAN-3", ""},
		{"PLN-0001.1", ".acta/plans/2026-10-01-b.md"},
	} {
		if got := scanPath(cfg, c.id); got != c.want {
			t.Errorf("scanPath(%q) = %q, want %q", c.id, got, c.want)
		}
	}
}

func TestShowPathUnknownIDIsBadInput(t *testing.T) {
	dir := pathRepo(t, ".acta")
	t.Chdir(dir)
	var stdout, stderr strings.Builder
	code := Run([]string{"show", "SPC-9999", "--path"}, strings.NewReader(""), false, &stdout, &stderr)
	if code != exitBadInput {
		t.Fatalf("exit %d, want %d", code, exitBadInput)
	}
	if stderr.String() != "unknown id SPC-9999\n" {
		t.Errorf("stderr %q, want %q", stderr.String(), "unknown id SPC-9999\n")
	}
	if stdout.String() != "" {
		t.Errorf("stdout %q, want nothing", stdout.String())
	}
}

// --path asks for the path and nothing else, so it wins over --json.
func TestShowPathBeatsJSON(t *testing.T) {
	dir := pathRepo(t, ".acta")
	t.Chdir(dir)
	if out := mustRun(t, "show", "SPC-0054", "--json", "--path"); out != ".acta/specs/2026-10-01-a.md\n" {
		t.Errorf("show --json --path = %q, want one path line", out)
	}
}

// list --json draws the same items through the same struct, so a link that
// show prints has to be there too.
func TestListJSONCarriesClosesAndClosedBy(t *testing.T) {
	dir := closesRepo(t)
	t.Chdir(dir)
	var items []struct {
		ID       string   `json:"id"`
		Closes   []string `json:"closes"`
		ClosedBy []string `json:"closed_by"`
	}
	if err := json.Unmarshal([]byte(mustRun(t, "list", "--json", "--all")), &items); err != nil {
		t.Fatal(err)
	}
	want := map[string][2][]string{
		"specs/2026-09-29-a":   {[]string{"scratch/2026-09-28-i", "scratch/2026-09-28-j"}, []string{}},
		"scratch/2026-09-28-i": {[]string{}, []string{"specs/2026-09-29-a"}},
		"scratch/2026-09-28-j": {[]string{}, []string{"specs/2026-09-29-a"}},
		"specs/2026-09-29-b":   {[]string{}, []string{}},
	}
	seen := 0
	for _, it := range items {
		w, ok := want[it.ID]
		if !ok {
			continue
		}
		seen++
		if it.Closes == nil || it.ClosedBy == nil {
			t.Errorf("%s in list --json wrote a null link: closes %v closed_by %v", it.ID, it.Closes, it.ClosedBy)
			continue
		}
		if !slices.Equal(it.Closes, w[0]) || !slices.Equal(it.ClosedBy, w[1]) {
			t.Errorf("%s in list --json: closes %v closed_by %v, want closes %v closed_by %v", it.ID, it.Closes, it.ClosedBy, w[0], w[1])
		}
	}
	if seen != len(want) {
		t.Errorf("list --json showed %d of the %d items under test", seen, len(want))
	}
}

// TestConfigShowNamesOldFile covers the user who moved config.yaml away: show
// once named the missing file and printed the values of ~/.pm/voice.yaml.
func TestConfigShowNamesOldFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", "")
	old := filepath.Join(home, ".pm", "voice.yaml")
	if err := os.MkdirAll(filepath.Dir(old), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(old, []byte("chat_language: Korean\nstyle: plain\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writes := filepath.Join(home, ".acta", "config.yaml")
	out := mustRun(t, "config", "show")
	want := "file: " + old + " (exists: true, old file; config set writes " + writes + ")"
	if !strings.Contains(out, want) {
		t.Errorf("show:\n%s\nwant line %q", out, want)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(mustRun(t, "config", "show", "--json")), &got); err != nil {
		t.Fatal(err)
	}
	if got["path"] != old || got["writes"] != writes {
		t.Errorf("json path %v writes %v, want %q and %q", got["path"], got["writes"], old, writes)
	}
}

// TestConfigSetNamesBrokenOldFile covers a broken ~/.pm/voice.yaml: set once
// told the user to fix ~/.acta/config.yaml, a file that was not there.
func TestConfigSetNamesBrokenOldFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", "")
	old := filepath.Join(home, ".pm", "voice.yaml")
	if err := os.MkdirAll(filepath.Dir(old), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(old, []byte("chat_language: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	newFile := filepath.Join(home, ".acta", "config.yaml")
	var stdout, stderr strings.Builder
	code := Run([]string{"config", "set", "--language", "Korean"}, strings.NewReader(""), false, &stdout, &stderr)
	if code != exitBadInput {
		t.Fatalf("exit %d, want %d; stderr %q", code, exitBadInput, stderr.String())
	}
	if want := "fix or delete " + old + " first"; !strings.Contains(stderr.String(), want) {
		t.Errorf("stderr %q lacks %q", stderr.String(), want)
	}
	if strings.Contains(stderr.String(), newFile) {
		t.Errorf("stderr %q names %s, which does not exist", stderr.String(), newFile)
	}
	if _, err := os.Stat(newFile); !os.IsNotExist(err) {
		t.Errorf("set wrote %s after a read error: %v", newFile, err)
	}
}

// brokenVoice is content no voice file can parse.
const brokenVoice = "chat_language: [\n"

// setWithBrokenFile makes a temp HOME, writes brokenVoice to rel inside it and
// hands back a run func for config set, so a test can change the home before
// the run. When pmRel is not empty, PM_VOICE_FILE points at that file, which is
// then both the file set reads and the file it would write.
func setWithBrokenFile(t *testing.T, rel, pmRel string) (home, broken string, run func() string) {
	t.Helper()
	home = t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", "")
	broken = filepath.Join(home, rel)
	if err := os.MkdirAll(filepath.Dir(broken), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(broken, []byte(brokenVoice), 0o644); err != nil {
		t.Fatal(err)
	}
	if pmRel != "" {
		t.Setenv("PM_VOICE_FILE", filepath.Join(home, pmRel))
	}
	return home, broken, func() string {
		t.Helper()
		var stdout, stderr strings.Builder
		code := Run([]string{"config", "set", "--language", "Korean"}, strings.NewReader(""), false, &stdout, &stderr)
		if code != exitBadInput {
			t.Fatalf("exit %d, want %d; stderr %q", code, exitBadInput, stderr.String())
		}
		return stderr.String()
	}
}

// wantFixLine checks the fix line names exactly broken, and that set left that
// file the way it found it.
func wantFixLine(t *testing.T, stderr, broken string) {
	t.Helper()
	if want := "fix or delete " + broken + " first"; !strings.Contains(stderr, want) {
		t.Errorf("stderr %q lacks %q", stderr, want)
	}
	if raw, err := os.ReadFile(broken); err != nil || string(raw) != brokenVoice {
		t.Errorf("set changed the broken file %s: %q %v", broken, raw, err)
	}
}

// TestConfigSetNamesBrokenConfigFile covers a broken ~/.acta/config.yaml: the
// fix line must name that same file.
func TestConfigSetNamesBrokenConfigFile(t *testing.T) {
	_, broken, run := setWithBrokenFile(t, filepath.Join(".acta", "config.yaml"), "")
	wantFixLine(t, run(), broken)
}

// TestConfigSetNamesBrokenPMVoiceFile covers a broken PM_VOICE_FILE: the fix
// line must name that path, not the default one.
func TestConfigSetNamesBrokenPMVoiceFile(t *testing.T) {
	_, broken, run := setWithBrokenFile(t, filepath.Join("elsewhere", "voice.yaml"), filepath.Join("elsewhere", "voice.yaml"))
	wantFixLine(t, run(), broken)
}

// TestConfigSetNamesBrokenUnmovableVoiceFile covers a broken ~/.acta/voice.yaml
// that could not be moved to config.yaml: the fix line must name voice.yaml.
func TestConfigSetNamesBrokenUnmovableVoiceFile(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can rename in a read-only folder")
	}
	home, broken, run := setWithBrokenFile(t, filepath.Join(".acta", "voice.yaml"), "")
	// A folder we cannot write to makes the move fail, so the broken file
	// stays where it is and is the one that failed to be read.
	dir := filepath.Join(home, ".acta")
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	wantFixLine(t, run(), broken)
	if _, err := os.Stat(filepath.Join(home, ".acta", "config.yaml")); !os.IsNotExist(err) {
		t.Errorf("set wrote config.yaml after a read error: %v", err)
	}
}
