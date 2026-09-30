package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// evalPlugin builds a tiny plugin with one case whose regex grader looks
// for want, and a fake omp on PATH that prints a stream saying said.
func evalPlugin(t *testing.T, want, said string) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"skills/plan/SKILL.md":      "x",
		"evals/one/prompt.md":       "---\ntimeout_seconds: 30\n---\nHi",
		"evals/one/graders/said.md": "---\ntype: regex\npattern: \"" + want + "\"\n---\n",
		"bin/omp":                   "#!/bin/sh\necho '{\"type\":\"agent_end\",\"messages\":[{\"role\":\"assistant\",\"content\":[{\"type\":\"text\",\"text\":\"" + said + "\"}]}]}'\n",
	}
	for p, text := range files {
		full := filepath.Join(dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", filepath.Join(dir, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TMPDIR", t.TempDir())
	return dir
}

func TestEvalOmpExitCodes(t *testing.T) {
	var out, errb bytes.Buffer
	dir := evalPlugin(t, "hello", "hello")
	if code := cmdEvalOmp([]string{dir}, &out, &errb); code != exitOK || !strings.Contains(out.String(), "PASS one") {
		t.Errorf("passing suite: code %d, out %q, err %q", code, out.String(), errb.String())
	}

	// A glob that matches nothing runs nothing, so the suite is green and
	// prints no case line.
	out.Reset()
	dir = evalPlugin(t, "hello", "hello")
	if code := cmdEvalOmp([]string{dir, "--case", "nothing-matches"}, &out, &errb); code != exitOK || strings.Contains(out.String(), "PASS") {
		t.Errorf("empty glob: code %d, out %q", code, out.String())
	}

	out.Reset()
	dir = evalPlugin(t, "hello", "bye")
	if code := cmdEvalOmp([]string{"--case", "one", dir}, &out, &errb); code != exitCaseFailed || !strings.Contains(out.String(), "FAIL one") {
		t.Errorf("failing suite: code %d, out %q", code, out.String())
	}

	if code := cmdEvalOmp([]string{filepath.Join(t.TempDir(), "nope")}, &out, &errb); code != exitOther {
		t.Errorf("missing plugin dir: code %d", code)
	}
	if code := cmdEvalOmp([]string{"--bogus"}, &out, &errb); code != exitBadInput {
		t.Errorf("bad flag: code %d", code)
	}
	if code := cmdEvalOmp([]string{"a", "b"}, &out, &errb); code != exitBadInput {
		t.Errorf("two plugin dirs: code %d", code)
	}

	// Bad usage has to say what the right call looks like, not just fail.
	errb.Reset()
	if code := cmdEvalOmp([]string{"--bogus"}, &out, &errb); code != exitBadInput || !strings.Contains(errb.String(), "usage: acta eval-omp [--case <glob>] [plugin-dir]") {
		t.Errorf("bad flag usage line: code %d, err %q", code, errb.String())
	}
	errb.Reset()
	if code := cmdEvalOmp([]string{"a", "b"}, &out, &errb); code != exitBadInput || !strings.Contains(errb.String(), "usage: acta eval-omp [--case <glob>] [plugin-dir]") {
		t.Errorf("two dirs usage line: code %d, err %q", code, errb.String())
	}
}

// A plugin folder whose evals will not load is a different failure from a
// folder that is not a plugin at all, and both mean exit 3.
func TestEvalOmpBrokenEvalDir(t *testing.T) {
	dir := evalPlugin(t, "hello", "hello")
	if err := os.WriteFile(filepath.Join(dir, "evals", "one", "prompt.md"), []byte("Hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if code := cmdEvalOmp([]string{dir}, &out, &errb); code != exitOther {
		t.Errorf("unloadable eval dir: code %d, err %q", code, errb.String())
	}

	out.Reset()
	errb.Reset()
	if code := cmdEvalOmp([]string{filepath.Join(t.TempDir(), "nope")}, &out, &errb); code != exitOther || !strings.Contains(errb.String(), "is not a plugin folder") {
		t.Errorf("missing plugin dir: code %d, err %q", code, errb.String())
	}
}

func TestEvalOmpNoOmp(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	var out, errb bytes.Buffer
	if code := cmdEvalOmp(nil, &out, &errb); code != exitOther || !strings.Contains(errb.String(), "omp is not on PATH") {
		t.Errorf("code %d, err %q", code, errb.String())
	}
}

// With no positional argument the runner looks for a plugin folder named
// plugin, so running it from a checkout works with no argument at all.
func TestEvalOmpDefaultsToPluginDir(t *testing.T) {
	dir := evalPlugin(t, "hello", "hello")
	root := t.TempDir()
	if err := os.Symlink(dir, filepath.Join(root, "plugin")); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	var out, errb bytes.Buffer
	if code := cmdEvalOmp(nil, &out, &errb); code != exitOK || !strings.Contains(out.String(), "PASS one") {
		t.Errorf("default plugin dir: code %d, out %q, err %q", code, out.String(), errb.String())
	}
}

func TestEvalOmpIsACommand(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	var out, errb bytes.Buffer
	Run([]string{"eval-omp"}, strings.NewReader(""), false, &out, &errb)
	if strings.Contains(errb.String(), "unknown command") {
		t.Errorf("eval-omp is not wired into Run: %q", errb.String())
	}
}
