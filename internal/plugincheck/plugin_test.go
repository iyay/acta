package plugincheck

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func readFile(t *testing.T, parts ...string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(append([]string{pluginRoot(t)}, parts...)...))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestManifests(t *testing.T) {
	var p struct{ Name, Version, License string }
	if err := json.Unmarshal([]byte(readFile(t, ".claude-plugin", "plugin.json")), &p); err != nil {
		t.Fatal(err)
	}
	if p.Name != "pm" || p.Version != "0.1.0" || p.License != "MIT" {
		t.Fatalf("plugin.json = %+v", p)
	}
	var m struct {
		Name    string
		Plugins []struct{ Name, Source string }
	}
	if err := json.Unmarshal([]byte(readFile(t, ".claude-plugin", "marketplace.json")), &m); err != nil {
		t.Fatal(err)
	}
	if m.Name != "pm-local" || len(m.Plugins) != 1 || m.Plugins[0].Name != "pm" || m.Plugins[0].Source != "./" {
		t.Fatalf("marketplace.json = %+v", m)
	}
}

func TestHooksJSON(t *testing.T) {
	var h struct {
		Hooks map[string][]struct {
			Matcher string
			Hooks   []struct{ Type, Command string }
		}
	}
	if err := json.Unmarshal([]byte(readFile(t, "hooks", "hooks.json")), &h); err != nil {
		t.Fatal(err)
	}
	ss, up := h.Hooks["SessionStart"], h.Hooks["UserPromptSubmit"]
	if len(ss) != 1 || ss[0].Matcher != "startup|resume|clear|compact" || ss[0].Hooks[0].Command != `"${CLAUDE_PLUGIN_ROOT}/hooks/session-start"` {
		t.Fatalf("SessionStart = %+v", ss)
	}
	if len(up) != 1 || up[0].Hooks[0].Command != `"${CLAUDE_PLUGIN_ROOT}/hooks/prompt-reminder"` {
		t.Fatalf("UserPromptSubmit = %+v", up)
	}
	for _, s := range []string{"session-start", "prompt-reminder"} {
		st, err := os.Stat(filepath.Join(pluginRoot(t), "hooks", s))
		if err != nil || st.Mode()&0o111 == 0 {
			t.Errorf("hooks/%s missing or not executable", s)
		}
	}
}

// hooksCopy puts the two scripts next to fixture files in a temp folder, so the
// test controls what the scripts find.
func hooksCopy(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, s := range []string{"session-start", "prompt-reminder"} {
		b := readFile(t, "hooks", s)
		if err := os.WriteFile(filepath.Join(dir, s), []byte(b), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "default-rules.md"), []byte("DEFAULT RULES\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow-plugins.txt"), []byte("superpowers\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func fakePmb(t *testing.T, body string) string {
	t.Helper()
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "pmb"), []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return bin
}

func runScript(t *testing.T, script, path string) (string, int) {
	t.Helper()
	cmd := exec.Command(script)
	cmd.Env = []string{"PATH=" + path, "HOME=" + t.TempDir()}
	out, err := cmd.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return string(out), code
}

func TestSessionStartScript(t *testing.T) {
	dir := hooksCopy(t)
	script := filepath.Join(dir, "session-start")

	out, code := runScript(t, script, "/usr/bin:/bin")
	if code != 0 || !strings.Contains(out, "DEFAULT RULES") || !strings.Contains(out, "pmb binary is not installed or failed") {
		t.Fatalf("no pmb: exit %d\n%s", code, out)
	}

	bin := fakePmb(t, `echo "FROM PMB $*"`)
	out, code = runScript(t, script, bin+":/usr/bin:/bin")
	want := "FROM PMB hook session-start --known " + filepath.Join(dir, "workflow-plugins.txt")
	if code != 0 || strings.TrimSpace(out) != want {
		t.Fatalf("with pmb: exit %d %q, want %q", code, out, want)
	}

	bin = fakePmb(t, "exit 1")
	out, code = runScript(t, script, bin+":/usr/bin:/bin")
	if code != 0 || !strings.Contains(out, "DEFAULT RULES") {
		t.Fatalf("failing pmb must fall back: exit %d\n%s", code, out)
	}
}

func TestPromptReminderScript(t *testing.T) {
	dir := hooksCopy(t)
	script := filepath.Join(dir, "prompt-reminder")
	if out, code := runScript(t, script, "/usr/bin:/bin"); code != 0 || out != "" {
		t.Fatalf("no pmb: exit %d %q", code, out)
	}
	bin := fakePmb(t, `echo "FROM PMB $*"`)
	if out, code := runScript(t, script, bin+":/usr/bin:/bin"); code != 0 || strings.TrimSpace(out) != "FROM PMB hook prompt" {
		t.Fatalf("with pmb: exit %d %q", code, out)
	}
	bin = fakePmb(t, "exit 3")
	if _, code := runScript(t, script, bin+":/usr/bin:/bin"); code != 0 {
		t.Fatalf("failing pmb must still exit 0, got %d", code)
	}
}

func TestWorkflowPluginsList(t *testing.T) {
	txt := readFile(t, "hooks", "workflow-plugins.txt")
	for _, want := range []string{"\nsuperpowers\n", "\ngstack\n", "\nmattpocock\n"} {
		if !strings.Contains(txt, want) {
			t.Errorf("workflow-plugins.txt missing %q", strings.TrimSpace(want))
		}
	}
}

func TestHouseRules(t *testing.T) {
	txt := readFile(t, "references", "house-rules.md")
	for _, want := range []string{"pm:build", "pm:tdd", "pm:debug", "pm:land", "pmb bug new", "pmb tick", "pmb tick plans/<stem>#task-N --step <n>", "PROPERTIES, NOT INSTANCES", "DO NOT REVIEW YOUR OWN WORK"} {
		if !strings.Contains(txt, want) {
			t.Errorf("house-rules.md missing %q", want)
		}
	}
	for _, bad := range []string{"superpowers:", "/Users/", "CLAUDE.md Core Six", "bugs.md", "per-task reviewer", "pmb tick <task-id>"} {
		if strings.Contains(txt, bad) {
			t.Errorf("house-rules.md still has %q", bad)
		}
	}
}

func TestNoticeAndReadme(t *testing.T) {
	notice := readFile(t, "NOTICE")
	for _, want := range []string{"Jesse Vincent", "Ayoub Ghriss", "MIT License", "Permission is hereby granted",
		"skills/brainstorm/", "skills/plan/", "skills/build/", "skills/tdd/", "skills/debug/", "skills/review/", "skills/land/"} {
		if !strings.Contains(notice, want) {
			t.Errorf("NOTICE missing %q", want)
		}
	}
	readme := readFile(t, "README.md")
	for _, want := range []string{"## Install", "## Voice", "## Other workflow plugins", "## Moving rules out of CLAUDE.md", "pmb voice set"} {
		if !strings.Contains(readme, want) {
			t.Errorf("README.md missing %q", want)
		}
	}
}
