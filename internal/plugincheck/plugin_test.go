package plugincheck

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
	if p.Name != "acta" || p.Version != "0.1.0" || p.License != "MIT" {
		t.Fatalf("plugin.json = %+v", p)
	}
	var m struct {
		Name    string
		Plugins []struct{ Name, Source string }
	}
	if err := json.Unmarshal([]byte(readFile(t, ".claude-plugin", "marketplace.json")), &m); err != nil {
		t.Fatal(err)
	}
	if m.Name != "acta-local" || len(m.Plugins) != 1 || m.Plugins[0].Name != "acta" || m.Plugins[0].Source != "./" {
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
	pre, post := h.Hooks["PreToolUse"], h.Hooks["PostToolUse"]
	if len(pre) != 1 || pre[0].Matcher != "Bash" || pre[0].Hooks[0].Command != `"${CLAUDE_PLUGIN_ROOT}/hooks/pre-tool"` {
		t.Fatalf("PreToolUse = %+v", pre)
	}
	if len(post) != 1 || post[0].Matcher != "Bash" || post[0].Hooks[0].Command != `"${CLAUDE_PLUGIN_ROOT}/hooks/post-tool"` {
		t.Fatalf("PostToolUse = %+v", post)
	}
	for _, s := range []string{"session-start", "prompt-reminder", "pre-tool", "post-tool"} {
		st, err := os.Stat(filepath.Join(pluginRoot(t), "hooks", s))
		if err != nil || st.Mode()&0o111 == 0 {
			t.Errorf("hooks/%s missing or not executable", s)
		}
	}
}

// hooksCopy puts the hook scripts next to fixture files in a temp folder, so
// the test controls what the scripts find.
func hooksCopy(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, s := range []string{"session-start", "prompt-reminder", "pre-tool", "post-tool"} {
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

func fakeActa(t *testing.T, body string) string {
	t.Helper()
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "acta"), []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return bin
}

// runScript runs a hook wrapper with only the given PATH, and feeds stdin to
// it, so the test sees exactly what a real hook would see.
func runScript(t *testing.T, script, path, stdin string) (string, int) {
	t.Helper()
	cmd := exec.Command(script)
	cmd.Env = []string{"PATH=" + path, "HOME=" + t.TempDir()}
	cmd.Stdin = strings.NewReader(stdin)
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

	out, code := runScript(t, script, "/usr/bin:/bin", "")
	if code != 0 || !strings.Contains(out, "DEFAULT RULES") || !strings.Contains(out, "acta binary is not installed or failed") {
		t.Fatalf("no acta: exit %d\n%s", code, out)
	}

	bin := fakeActa(t, `echo "FROM ACTA $*"`)
	out, code = runScript(t, script, bin+":/usr/bin:/bin", "")
	want := "FROM ACTA hook session-start --known " + filepath.Join(dir, "workflow-plugins.txt")
	if code != 0 || strings.TrimSpace(out) != want {
		t.Fatalf("with acta: exit %d %q, want %q", code, out, want)
	}

	bin = fakeActa(t, "exit 1")
	out, code = runScript(t, script, bin+":/usr/bin:/bin", "")
	if code != 0 || !strings.Contains(out, "DEFAULT RULES") {
		t.Fatalf("failing acta must fall back: exit %d\n%s", code, out)
	}
}

func TestPromptReminderScript(t *testing.T) {
	dir := hooksCopy(t)
	script := filepath.Join(dir, "prompt-reminder")
	if out, code := runScript(t, script, "/usr/bin:/bin", ""); code != 0 || out != "" {
		t.Fatalf("no acta: exit %d %q", code, out)
	}
	bin := fakeActa(t, `echo "FROM ACTA $*"`)
	if out, code := runScript(t, script, bin+":/usr/bin:/bin", ""); code != 0 || strings.TrimSpace(out) != "FROM ACTA hook prompt" {
		t.Fatalf("with acta: exit %d %q", code, out)
	}
	bin = fakeActa(t, "exit 3")
	if _, code := runScript(t, script, bin+":/usr/bin:/bin", ""); code != 0 {
		t.Fatalf("failing acta must still exit 0, got %d", code)
	}
}

func TestPreToolScript(t *testing.T) {
	dir := hooksCopy(t)
	script := filepath.Join(dir, "pre-tool")

	if out, code := runScript(t, script, "/usr/bin:/bin", "X"); code != 0 || out != "" {
		t.Fatalf("no acta: exit %d %q", code, out)
	}

	bin := fakeActa(t, `echo "FROM ACTA $*"; cat`)
	out, code := runScript(t, script, bin+":/usr/bin:/bin", "X")
	if code != 0 || !strings.Contains(out, "FROM ACTA hook pre-tool") || !strings.Contains(out, "X") {
		t.Fatalf("with acta: exit %d %q", code, out)
	}

	const messy = "{\n  \"tool_input\": {\"command\": \"acta set scratch/x status brainstorming\"}\n}\n"
	out, code = runScript(t, script, bin+":/usr/bin:/bin", messy)
	if code != 0 || !strings.Contains(out, messy) {
		t.Fatalf("stdin must reach acta whole: exit %d %q", code, out)
	}

	bin = fakeActa(t, "echo BLOCKED >&2; exit 2")
	out, code = runScript(t, script, bin+":/usr/bin:/bin", "")
	if code != 2 || !strings.Contains(out, "BLOCKED") {
		t.Fatalf("blocked: exit %d %q, want exit 2 with the message", code, out)
	}

	for _, c := range []int{1, 3, 127, 130} {
		bin = fakeActa(t, "exit "+strconv.Itoa(c))
		if out, code := runScript(t, script, bin+":/usr/bin:/bin", "X"); code != 0 || out != "" {
			t.Fatalf("acta exit %d must pass silently: exit %d %q", c, code, out)
		}
	}
}

func TestPostToolScript(t *testing.T) {
	dir := hooksCopy(t)
	script := filepath.Join(dir, "post-tool")

	if out, code := runScript(t, script, "/usr/bin:/bin", "X"); code != 0 || out != "" {
		t.Fatalf("no acta: exit %d %q", code, out)
	}

	bin := fakeActa(t, `echo "FROM ACTA $*"; cat`)
	out, code := runScript(t, script, bin+":/usr/bin:/bin", "X")
	if code != 0 || !strings.Contains(out, "FROM ACTA hook post-tool") || !strings.Contains(out, "X") {
		t.Fatalf("with acta: exit %d %q", code, out)
	}

	bin = fakeActa(t, "exit 2")
	if out, code := runScript(t, script, bin+":/usr/bin:/bin", "X"); code != 0 || out != "" {
		t.Fatalf("post-tool must never block: exit %d %q", code, out)
	}

	for _, c := range []int{1, 3, 127, 130} {
		bin = fakeActa(t, "exit "+strconv.Itoa(c))
		if _, code := runScript(t, script, bin+":/usr/bin:/bin", ""); code != 0 {
			t.Fatalf("acta exit %d must still exit 0, got %d", c, code)
		}
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
	for _, want := range []string{"acta:build", "acta:tdd", "acta:debug", "acta:land", "acta bug new", "acta tick", "acta tick plans/<stem>#task-N --step <n>", "PROPERTIES, NOT INSTANCES", "DO NOT REVIEW YOUR OWN WORK", "acta tick plans/<stem>#task-N --all", "progress.done", "before the reply-back", "never write NOTEs to memory", "sorts them", "the review skill's `## After a CLEAN round` sorts them"} {
		if !strings.Contains(txt, want) {
			t.Errorf("house-rules.md missing %q", want)
		}
	}
	for _, bad := range []string{"superpowers:", "/Users/", "CLAUDE.md Core Six", "bugs.md", "per-task reviewer", "acta tick <task-id>", "<its task id>", "files them with acta debt new", "orchestrator sorts them through", "orchestrator sorts it through"} {
		if strings.Contains(txt, bad) {
			t.Errorf("house-rules.md still has %q", bad)
		}
	}
}

func TestNoticeAndReadme(t *testing.T) {
	notice := readFile(t, "NOTICE")
	for _, want := range []string{"Jesse Vincent", "Ayoub Ghriss", "MIT License", "Permission is hereby granted",
		"skills/shape/", "skills/slice/", "skills/build/", "skills/tdd/", "skills/debug/", "skills/review/", "skills/land/"} {
		if !strings.Contains(notice, want) {
			t.Errorf("NOTICE missing %q", want)
		}
	}
	readme := readFile(t, "README.md")
	for _, want := range []string{"## Install", "## First run", "## Voice", "## Other workflow plugins",
		"## Moving rules out of CLAUDE.md", "acta config set", "acta doctor", "acta doctor --fix", "/acta:setup",
		"acta: ", "only between acta markers"} {
		if !strings.Contains(readme, want) {
			t.Errorf("README.md missing %q", want)
		}
	}
	if strings.Contains(strings.ToLower(readme), "never edits your claude.md") {
		t.Error("README.md still promises it never edits CLAUDE.md")
	}
}
