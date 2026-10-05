package doctor

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
)

// env gives every check a temp home and a fresh git repo, so no test can
// read or write the real user folders.
func env(t *testing.T) Env {
	t.Helper()
	home := t.TempDir()
	repo := t.TempDir()
	if out, err := exec.Command("git", "-C", repo, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	return Env{Home: home, ClaudeDir: filepath.Join(home, ".claude"), RepoRoot: repo,
		ActaRoot: filepath.Join(repo, ".acta"), Binary: "/bin/acta", Version: "test",
		VoiceExists: true, Voice: config.User{BuildExecutor: "subagent"}}
}

func byName(rs []Result, name string) Result {
	for _, r := range rs {
		if r.Name == name {
			return r
		}
	}
	return Result{}
}

func names(rs []Result) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, r.Name)
	}
	return out
}

// write puts a file under a temp root, making the folders it needs.
func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// link points a symlink at a target, live or already deleted.
func link(t *testing.T, path, target string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
}

func nodeModules(e Env) string { return filepath.Join(e.Home, ".omp", "plugins", "node_modules") }

// wantLevel fails unless r carries this level, and when fix is not empty
// unless r.Fix holds it.
func wantLevel(t *testing.T, r Result, level Level, fix string) {
	t.Helper()
	if r.Level != level {
		t.Fatalf("level=%q want %q (msg %q fix %q)", r.Level, level, r.Msg, r.Fix)
	}
	if fix != "" && !strings.Contains(r.Fix, fix) {
		t.Fatalf("fix %q does not hold %q", r.Fix, fix)
	}
}

// TestCheckTheme covers every level the check can report. A theme that does
// not load is a warning, never a failure: the TUI still opens on the default,
// so a broken name must not cost the user their board.
func TestCheckTheme(t *testing.T) {
	cases := []struct {
		name  string
		env   Env
		level Level
		msg   string
	}{
		{"empty", Env{}, OK, "tokyo-night"},
		{"chosen", Env{Voice: config.User{Theme: "dracula"}}, OK, "dracula"},
		{"broken", Env{Voice: config.User{Theme: "nope"}, ThemeErr: errors.New(`theme "nope": unknown theme`)}, Warn, `theme "nope"`},
		{"bad name", Env{Voice: config.User{Theme: "NOPE!"}, ThemeErr: errors.New(`theme "NOPE!": name may only use a-z, 0-9 and -`)}, Warn, "NOPE!"},
		{"broken file", Env{Voice: config.User{Theme: "mine"}, ThemeErr: errors.New(`theme "mine": ansi needs 16 colors, has 3`)}, Warn, "mine"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := checkTheme(c.env)
			fix := ""
			if c.level != OK {
				fix = "acta config set --theme"
			}
			wantLevel(t, r, c.level, fix)
			if !strings.Contains(r.Msg, c.msg) {
				t.Fatalf("msg %q does not hold %q", r.Msg, c.msg)
			}
			if c.level == OK && r.Fix != "" {
				t.Fatalf("a good theme has the fix %q", r.Fix)
			}
		})
	}
}

func TestRunIncludesTheme(t *testing.T) {
	if byName(Run(env(t)), "theme").Name != "theme" {
		t.Fatal("Run has no theme check")
	}
}

// A broken theme must not change the exit code, because the TUI opens
// anyway.
func TestBrokenThemeDoesNotFailTheRun(t *testing.T) {
	good := env(t)
	broken := env(t)
	broken.Voice.Theme, broken.ThemeErr = "nope", errors.New(`theme "nope": unknown theme`)
	if Failed(Run(broken)) != Failed(Run(good)) {
		t.Fatal("a broken theme failed a run that was not failing before")
	}
}

func TestDoctorChecksRunInFixedOrder(t *testing.T) {
	want := []string{"binary", "harness", "stale-links", "conflicts", "repo", "schema", "files", "agents-view", "setup", "theme"}
	if got := names(Run(env(t))); !reflect.DeepEqual(got, want) {
		t.Fatalf("order %v want %v", got, want)
	}
}

func TestDoctorBinary(t *testing.T) {
	t.Run("path and version", func(t *testing.T) {
		r := byName(Run(env(t)), "binary")
		wantLevel(t, r, OK, "")
		if r.Msg != "/bin/acta test" {
			t.Fatalf("msg %q", r.Msg)
		}
	})
	t.Run("no path", func(t *testing.T) {
		e := env(t)
		e.Binary = ""
		wantLevel(t, byName(Run(e), "binary"), Fail, "")
	})
}

func TestDoctorHarness(t *testing.T) {
	t.Run("claude plugin enabled", func(t *testing.T) {
		e := env(t)
		write(t, filepath.Join(e.ClaudeDir, "settings.json"), `{"enabledPlugins":{"acta@local":true}}`)
		wantLevel(t, byName(Run(e), "harness"), OK, "")
	})
	t.Run("claude plugin enabled without marketplace", func(t *testing.T) {
		e := env(t)
		write(t, filepath.Join(e.ClaudeDir, "settings.json"), `{"enabledPlugins":{"acta":true}}`)
		wantLevel(t, byName(Run(e), "harness"), OK, "")
	})
	t.Run("nothing installed", func(t *testing.T) {
		e := env(t)
		wantLevel(t, byName(Run(e), "harness"), Fail, "## Install")
	})
	t.Run("omp link alive", func(t *testing.T) {
		e := env(t)
		target := t.TempDir()
		link(t, filepath.Join(nodeModules(e), "acta"), target)
		wantLevel(t, byName(Run(e), "harness"), OK, "")
	})
	t.Run("omp link target gone", func(t *testing.T) {
		e := env(t)
		gone := filepath.Join(t.TempDir(), "deleted")
		link(t, filepath.Join(nodeModules(e), "acta"), gone)
		r := byName(Run(e), "harness")
		wantLevel(t, r, Fail, "omp plugin link "+gone)
	})
	t.Run("broken claude settings", func(t *testing.T) {
		e := env(t)
		write(t, filepath.Join(e.ClaudeDir, "settings.json"), `{not json`)
		wantLevel(t, byName(Run(e), "harness"), Fail, "## Install")
	})
}

func TestDoctorStaleLinks(t *testing.T) {
	// omp has no "unlink" action, so a hint naming one is a dead end. The
	// dead link is just a file left in node_modules, so the fix removes it.
	noUnlink := func(t *testing.T, rs []Result) {
		t.Helper()
		for _, r := range rs {
			if strings.Contains(r.Fix, "omp plugin unlink") {
				t.Fatalf("%s fix %q names an action omp does not have", r.Name, r.Fix)
			}
		}
	}
	t.Run("one dead link", func(t *testing.T) {
		e := env(t)
		link(t, filepath.Join(nodeModules(e), "pm"), filepath.Join(t.TempDir(), "gone"))
		rs := Run(e)
		r := byName(rs, "stale-links")
		wantLevel(t, r, Warn, "rm "+filepath.Join(nodeModules(e), "pm"))
		if r.Fix != "rm "+filepath.Join(nodeModules(e), "pm") {
			t.Fatalf("fix %q wants only the one rm line", r.Fix)
		}
		if !strings.Contains(r.Msg, "pm") {
			t.Fatalf("msg %q does not name pm", r.Msg)
		}
		noUnlink(t, rs)
	})
	t.Run("live link and folder beside it", func(t *testing.T) {
		e := env(t)
		link(t, filepath.Join(nodeModules(e), "acta"), t.TempDir())
		if err := os.MkdirAll(filepath.Join(nodeModules(e), "zeta"), 0o755); err != nil {
			t.Fatal(err)
		}
		rs := Run(e)
		wantLevel(t, byName(rs, "stale-links"), OK, "")
		noUnlink(t, rs)
	})
	t.Run("no folder", func(t *testing.T) {
		e := env(t)
		rs := Run(e)
		wantLevel(t, byName(rs, "stale-links"), OK, "")
		noUnlink(t, rs)
	})
	t.Run("two dead links, one fix line each", func(t *testing.T) {
		e := env(t)
		link(t, filepath.Join(nodeModules(e), "pm"), filepath.Join(t.TempDir(), "gone"))
		link(t, filepath.Join(nodeModules(e), "gstack"), filepath.Join(t.TempDir(), "gone2"))
		rs := Run(e)
		wantLevel(t, byName(rs, "stale-links"), Warn,
			"rm "+filepath.Join(nodeModules(e), "gstack")+"; rm "+filepath.Join(nodeModules(e), "pm"))
		noUnlink(t, rs)
	})
	t.Run("unreadable folder", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("root reads everything")
		}
		e := env(t)
		dir := nodeModules(e)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(dir, 0o000); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
		wantLevel(t, byName(Run(e), "stale-links"), OK, "")
	})
}

func TestDoctorConflicts(t *testing.T) {
	t.Run("clashing plugin enabled", func(t *testing.T) {
		e := env(t)
		known := filepath.Join(t.TempDir(), "workflow-plugins.txt")
		write(t, known, "# clashers\nsuperpowers\n")
		e.KnownFile = known
		write(t, filepath.Join(e.ClaudeDir, "settings.json"), `{"enabledPlugins":{"superpowers@x":true}}`)
		r := byName(Run(e), "conflicts")
		wantLevel(t, r, Warn, "settings.local.json")
		if !strings.Contains(r.Msg, "superpowers@x") {
			t.Fatalf("msg %q does not name the plugin", r.Msg)
		}
	})
	t.Run("no known file", func(t *testing.T) {
		e := env(t)
		write(t, filepath.Join(e.ClaudeDir, "settings.json"), `{"enabledPlugins":{"superpowers@x":true}}`)
		r := byName(Run(e), "conflicts")
		wantLevel(t, r, OK, "")
		if want := "no list of plugins that overlap acta given"; r.Msg != want {
			t.Fatalf("msg %q, want %q", r.Msg, want)
		}
	})
	t.Run("known file missing", func(t *testing.T) {
		e := env(t)
		e.KnownFile = filepath.Join(t.TempDir(), "gone.txt")
		wantLevel(t, byName(Run(e), "conflicts"), OK, "")
	})

	// These read the real list that ships in the plugin folder, so a name
	// that drops out of it fails here.
	realList := filepath.Join("..", "..", "plugin", "hooks", "workflow-plugins.txt")
	msgHas := func(t *testing.T, r Result, wants ...string) {
		t.Helper()
		for _, want := range wants {
			if !strings.Contains(r.Msg, want) {
				t.Fatalf("msg %q does not hold %q", r.Msg, want)
			}
		}
		if strings.Contains(r.Msg, "workflow") {
			t.Fatalf("msg %q still says workflow", r.Msg)
		}
	}
	t.Run("caveman and ponytail enabled", func(t *testing.T) {
		e := env(t)
		e.KnownFile = realList
		write(t, filepath.Join(e.ClaudeDir, "settings.json"), `{"enabledPlugins":{"caveman@caveman":true,"ponytail@ponytail":true}}`)
		r := byName(Run(e), "conflicts")
		wantLevel(t, r, Warn, "settings.local.json")
		msgHas(t, r, "caveman@caveman", "ponytail@ponytail", "overlaps acta")
	})
	t.Run("matched by plugin or marketplace name in any case", func(t *testing.T) {
		e := env(t)
		e.KnownFile = realList
		write(t, filepath.Join(e.ClaudeDir, "settings.json"), `{"enabledPlugins":{"Caveman@some-market":true,"some-tool@PONYTAIL":true}}`)
		r := byName(Run(e), "conflicts")
		wantLevel(t, r, Warn, "settings.local.json")
		msgHas(t, r, "Caveman@some-market", "some-tool@PONYTAIL", "overlaps acta")
	})
	t.Run("nothing that overlaps acta enabled", func(t *testing.T) {
		e := env(t)
		e.KnownFile = realList
		write(t, filepath.Join(e.ClaudeDir, "settings.json"), `{"enabledPlugins":{"acta@acta-local":true,"other@market":true}}`)
		r := byName(Run(e), "conflicts")
		wantLevel(t, r, OK, "")
		msgHas(t, r, "overlaps acta")
	})
}

func TestDoctorRepo(t *testing.T) {
	t.Run("no acta folder", func(t *testing.T) {
		e := env(t)
		wantLevel(t, byName(Run(e), "repo"), Fail, "acta doctor --fix")
	})
	t.Run("acta folder without gitignore", func(t *testing.T) {
		e := env(t)
		if err := os.MkdirAll(e.ActaRoot, 0o755); err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(e.ActaRoot, ".gitignore"), "notes.txt\n")
		wantLevel(t, byName(Run(e), "repo"), Fail, "acta doctor --fix")
	})
	t.Run("line present", func(t *testing.T) {
		e := env(t)
		write(t, filepath.Join(e.ActaRoot, ".gitignore"), "notes.txt\n.agents.json\n")
		wantLevel(t, byName(Run(e), "repo"), OK, "")
	})
	t.Run("line commented out is not there", func(t *testing.T) {
		e := env(t)
		write(t, filepath.Join(e.ActaRoot, ".gitignore"), "# .agents.json\n")
		wantLevel(t, byName(Run(e), "repo"), Fail, "acta doctor --fix")
	})
	t.Run("after fix", func(t *testing.T) {
		e := env(t)
		if _, err := Fix(e); err != nil {
			t.Fatal(err)
		}
		wantLevel(t, byName(Run(e), "repo"), OK, "")
	})
	t.Run("outside a git repo", func(t *testing.T) {
		e := env(t)
		e.RepoRoot = ""
		e.ActaRoot = filepath.Join(e.Home, ".acta")
		r := byName(Run(e), "repo")
		wantLevel(t, r, OK, "")
		if !strings.Contains(r.Msg, "skipped") {
			t.Fatalf("msg %q does not say skipped", r.Msg)
		}
	})
	t.Run("acta root is a file", func(t *testing.T) {
		e := env(t)
		write(t, e.ActaRoot, "not a folder")
		wantLevel(t, byName(Run(e), "repo"), Fail, "acta doctor --fix")
	})
}

// The schema check is pure: the CLI hands it the problems it already found,
// so the doctor package reads no board file itself.
func TestCheckSchema(t *testing.T) {
	t.Run("no problem", func(t *testing.T) {
		r := checkSchema(Env{})
		wantLevel(t, r, OK, "")
		if r.Msg != "every schema file has its sections" {
			t.Fatalf("msg %q", r.Msg)
		}
	})
	t.Run("two problems, one line, sorted by the caller", func(t *testing.T) {
		r := checkSchema(Env{SchemaProblems: []string{
			"scratch a.md: missing ## Words",
			"scratch b.md: missing ## Words",
		}})
		wantLevel(t, r, Warn, "acta scratch add --section")
		if r.Msg != "scratch a.md: missing ## Words; scratch b.md: missing ## Words" {
			t.Fatalf("msg %q", r.Msg)
		}
		if !strings.Contains(Format([]Result{r}), "warn schema: ") {
			t.Fatal("the report line has no level and name")
		}
	})
	t.Run("a file without schema: 1 is never named", func(t *testing.T) {
		// Env carries only what the CLI found, so a file the CLI skipped
		// cannot appear here.
		r := checkSchema(Env{SchemaProblems: []string{"scratch old.md: missing ## Words"}})
		if strings.Contains(r.Msg, "no-schema.md") {
			t.Fatalf("msg %q", r.Msg)
		}
	})
}

// The files check names planning files acta never stamped: a spec or plan
// with no id, and a plan file with no task heading. It reads only the
// checkout's own folders, so a file on another branch or worktree is never
// named.
func TestDoctorFiles(t *testing.T) {
	good := "---\nid: PLN-0086\nhash: kit90tv\n---\n# Good\n\n### Task 1: One\n\n- [ ] a\n"
	t.Run("stamped plan with tasks is ok", func(t *testing.T) {
		e := env(t)
		write(t, filepath.Join(e.ActaRoot, "plans", "2026-10-05-good.md"), good)
		r := byName(Run(e), "files")
		wantLevel(t, r, OK, "")
	})
	t.Run("plan with no id", func(t *testing.T) {
		e := env(t)
		write(t, filepath.Join(e.ActaRoot, "plans", "2026-10-05-noid.md"), "---\nstatus: draft\n---\n# No id\n\n### Task 1: One\n\n- [ ] a\n")
		r := byName(Run(e), "files")
		wantLevel(t, r, Warn, "acta id")
		if !strings.Contains(r.Msg, "plans/2026-10-05-noid.md") {
			t.Fatalf("msg %q does not name the file", r.Msg)
		}
	})
	t.Run("spec with no id", func(t *testing.T) {
		e := env(t)
		write(t, filepath.Join(e.ActaRoot, "specs", "2026-10-05-noid.md"), "# No id at all\n")
		r := byName(Run(e), "files")
		wantLevel(t, r, Warn, "acta id")
		if !strings.Contains(r.Msg, "specs/2026-10-05-noid.md") {
			t.Fatalf("msg %q does not name the file", r.Msg)
		}
	})
	t.Run("plans file with no task heading", func(t *testing.T) {
		e := env(t)
		write(t, filepath.Join(e.ActaRoot, "plans", "2026-10-05-notask.md"), "---\nid: PLN-0087\nhash: a1b2c3d\n---\n# Notes\n\nJust words.\n")
		r := byName(Run(e), "files")
		wantLevel(t, r, Warn, "move it out of plans/")
		if !strings.Contains(r.Msg, "plans/2026-10-05-notask.md") {
			t.Fatalf("msg %q does not name the file", r.Msg)
		}
	})
	t.Run("every bad file named, stamped file not", func(t *testing.T) {
		e := env(t)
		write(t, filepath.Join(e.ActaRoot, "plans", "2026-10-05-good.md"), good)
		write(t, filepath.Join(e.ActaRoot, "plans", "2026-10-05-noid.md"), "# No id\n\n### Task 1: One\n\n- [ ] a\n")
		write(t, filepath.Join(e.ActaRoot, "specs", "2026-10-05-noid.md"), "# No id\n")
		write(t, filepath.Join(e.ActaRoot, "plans", "2026-10-05-notask.md"), "---\nid: PLN-0087\nhash: a1b2c3d\n---\n# Notes\n")
		r := byName(Run(e), "files")
		wantLevel(t, r, Warn, "acta id")
		for _, name := range []string{"plans/2026-10-05-noid.md", "specs/2026-10-05-noid.md", "plans/2026-10-05-notask.md"} {
			if !strings.Contains(r.Msg, name) {
				t.Fatalf("msg %q does not name %s", r.Msg, name)
			}
		}
		if strings.Contains(r.Msg, "good.md") {
			t.Fatalf("msg %q names the stamped file", r.Msg)
		}
		if !strings.Contains(r.Fix, "move it out of plans/") {
			t.Fatalf("fix %q misses the plans fix", r.Fix)
		}
	})
	t.Run("a file outside the acta folder is never read", func(t *testing.T) {
		e := env(t)
		write(t, filepath.Join(e.ActaRoot, "plans", "2026-10-05-good.md"), good)
		write(t, filepath.Join(e.RepoRoot, "plans", "stray.md"), "# No id\n")
		r := byName(Run(e), "files")
		wantLevel(t, r, OK, "")
	})
	t.Run("outside a git repo", func(t *testing.T) {
		e := env(t)
		e.RepoRoot = ""
		r := byName(Run(e), "files")
		wantLevel(t, r, OK, "")
		if !strings.Contains(r.Msg, "skipped") {
			t.Fatalf("msg %q does not say skipped", r.Msg)
		}
	})
}

func TestDoctorAgentsView(t *testing.T) {
	cases := []struct {
		name string
		body string
		want Level
		fix  string
	}{
		{"no file", "", OK, ""},
		{"empty object", `{}`, OK, ""},
		{"true", `{"leftArrowOpensAgents":true}`, OK, ""},
		{"false", `{"leftArrowOpensAgents":false}`, Warn, "/config"},
		{"broken json", `{bad json`, Warn, ""},
		{"other keys only", `{"theme":"dark"}`, OK, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := env(t)
			if c.body != "" {
				write(t, filepath.Join(e.Home, ".claude.json"), c.body)
			}
			wantLevel(t, byName(Run(e), "agents-view"), c.want, c.fix)
		})
	}
	t.Run("broken json keeps the parse error", func(t *testing.T) {
		e := env(t)
		write(t, filepath.Join(e.Home, ".claude.json"), `{bad json`)
		if msg := byName(Run(e), "agents-view").Msg; !strings.Contains(msg, "json") {
			t.Fatalf("msg %q does not hold the parse error", msg)
		}
	})
}

// The setup check must tell a healthy setup from a missing one, so a user
// who skipped picking an executor knows that is fine and not broken.
func TestDoctorSetup(t *testing.T) {
	tests := []struct {
		name        string
		voiceExists bool
		executor    string
		level       Level
		msg         string
		fix         string
	}{
		{"no voice file", false, "", Warn, "no voice file yet", "/acta:setup"},
		{"voice with no executor", true, "", OK, "voice is set; build executor not set, build asks each time", ""},
		{"voice with executor", true, "subagent", OK, "voice and build executor are set", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := env(t)
			e.VoiceExists = tt.voiceExists
			e.Voice = config.User{BuildExecutor: tt.executor}
			r := byName(Run(e), "setup")
			if r.Level != tt.level {
				t.Fatalf("level=%q want %q (msg %q fix %q)", r.Level, tt.level, r.Msg, r.Fix)
			}
			if r.Msg != tt.msg {
				t.Fatalf("msg=%q want %q", r.Msg, tt.msg)
			}
			if r.Fix != tt.fix {
				t.Fatalf("fix=%q want %q", r.Fix, tt.fix)
			}
		})
	}
}

// TestDoctorFixTwice is the point of --fix: it must be safe to run again.
func TestDoctorFixTwice(t *testing.T) {
	e := env(t)
	first, err := Fix(e)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) == 0 {
		t.Fatalf("first fix changed nothing: %v", first)
	}
	second, err := Fix(e)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 0 {
		t.Fatalf("second fix changed %v", second)
	}
	raw, err := os.ReadFile(filepath.Join(e.ActaRoot, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(raw), ".agents.json"); n != 1 {
		t.Fatalf(".gitignore holds .agents.json %d times: %q", n, raw)
	}
	wantLevel(t, byName(Run(e), "repo"), OK, "")
}

func TestDoctorFixKeepsOtherGitignoreLines(t *testing.T) {
	e := env(t)
	write(t, filepath.Join(e.ActaRoot, ".gitignore"), "notes.txt")
	paths, err := Fix(e)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 {
		t.Fatalf("paths %v", paths)
	}
	raw, err := os.ReadFile(filepath.Join(e.ActaRoot, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "notes.txt") || !strings.Contains(string(raw), ".agents.json") {
		t.Fatalf("gitignore %q", raw)
	}
	if _, err := Fix(e); err != nil {
		t.Fatal(err)
	}
	again, err := os.ReadFile(filepath.Join(e.ActaRoot, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(raw) {
		t.Fatalf("second fix rewrote the file: %q", again)
	}
}

func TestDoctorFixOutsideGitRepoWritesNothing(t *testing.T) {
	e := env(t)
	e.RepoRoot = ""
	e.ActaRoot = filepath.Join(e.Home, ".acta")
	paths, err := Fix(e)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 0 {
		t.Fatalf("paths %v", paths)
	}
	if _, err := os.Stat(e.ActaRoot); !os.IsNotExist(err) {
		t.Fatalf("Fix made %s outside a git repo", e.ActaRoot)
	}
}

// A root that points out of the repo is the one case where --fix has no
// work it may do. The check names the root, because --fix cannot move it.
func TestDoctorFixWritesNothingWhenRootLeavesTheRepo(t *testing.T) {
	cases := []struct {
		name string
		root func(t *testing.T, e Env) string
	}{
		{"parent folder", func(_ *testing.T, e Env) string { return filepath.Join(e.RepoRoot, "..", "escaped") }},
		{"absolute path", func(t *testing.T, _ Env) string { return filepath.Join(t.TempDir(), "elsewhere") }},
		{"sibling with the repo name as prefix", func(_ *testing.T, e Env) string { return e.RepoRoot + "-other" }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := env(t)
			e.ActaRoot = c.root(t, e)
			paths, err := Fix(e)
			if err != nil {
				t.Fatal(err)
			}
			if len(paths) != 0 {
				t.Fatalf("paths %v", paths)
			}
			if _, err := os.Stat(e.ActaRoot); !os.IsNotExist(err) {
				t.Fatalf("Fix made %s outside the repo", e.ActaRoot)
			}
			r := byName(Run(e), "repo")
			wantLevel(t, r, Fail, "root")
			if strings.Contains(r.Fix, "acta doctor --fix") {
				t.Fatalf("fix %q sends the user to --fix, which cannot move root", r.Fix)
			}
		})
	}
}

// A .acta folder that is a link out of the repo is a way to write anywhere
// on the disk, so --fix must write nothing and the check must say so.
func TestDoctorFixWritesNothingWhenActaRootIsALinkOut(t *testing.T) {
	e := env(t)
	outside := filepath.Join(filepath.Dir(e.RepoRoot), "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	link(t, e.ActaRoot, filepath.Join("..", "outside"))
	paths, err := Fix(e)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 0 {
		t.Fatalf("paths %v", paths)
	}
	// snapshot lists the folder itself as ./, so any other entry is a write.
	if got := snapshot(t, outside); len(got) > 1 {
		t.Fatalf("Fix wrote outside the repo: %v", got)
	}
	wantLevel(t, byName(Run(e), "repo"), Fail, "root")
}

// A .gitignore that is a link to a file that is not there yet still makes
// the append land on the target, creating a file outside the repo. So --fix
// must write nothing, the target must stay missing, and the check must say
// the link is the problem instead of sending the user back to --fix.
func TestDoctorFixWritesNothingWhenGitignoreIsADanglingLink(t *testing.T) {
	e := env(t)
	outside := filepath.Join(filepath.Dir(e.RepoRoot), "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	victim := filepath.Join(outside, "new.conf")
	if err := os.MkdirAll(e.ActaRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	link(t, filepath.Join(e.ActaRoot, ".gitignore"), filepath.Join("..", "..", "outside", "new.conf"))
	paths, err := Fix(e)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 0 {
		t.Fatalf("paths %v", paths)
	}
	if _, err := os.Stat(victim); !os.IsNotExist(err) {
		t.Fatalf("Fix created the link target %s", victim)
	}
	r := byName(Run(e), "repo")
	wantLevel(t, r, Fail, "replace the .gitignore link")
	if strings.Contains(r.Fix, "acta doctor --fix") {
		t.Fatalf("fix %q sends the user to --fix, which would write through the link", r.Fix)
	}
}

// A .gitignore that is a link to a file of the same repo looks fine to a
// bounds check, because the target really is in the repo. But the append
// lands on that other file, and the check then reads the line back through
// the link and says ok. So the target must keep its bytes and the check
// must fail.
func TestDoctorFixWritesNothingWhenGitignoreLinksInsideTheRepo(t *testing.T) {
	e := env(t)
	victim := filepath.Join(e.RepoRoot, "docs", "notes.txt")
	write(t, victim, "release notes\n")
	link(t, filepath.Join(e.ActaRoot, ".gitignore"), filepath.Join("..", "docs", "notes.txt"))
	paths, err := Fix(e)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 0 {
		t.Fatalf("paths %v", paths)
	}
	raw, err := os.ReadFile(victim)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "release notes\n" {
		t.Fatalf("Fix wrote through the link: %q", raw)
	}
	wantLevel(t, byName(Run(e), "repo"), Fail, "replace the .gitignore link")
}

// A config file that does not parse is a broken setup, not a skipped
// check, so the repo line has to carry the error and fail.
func TestDoctorRepoFailsOnBrokenConfig(t *testing.T) {
	e := env(t)
	e.ConfigErr = errors.New(".acta.yaml: yaml: line 1: did not find expected node content")
	r := byName(Run(e), "repo")
	wantLevel(t, r, Fail, "")
	if !strings.Contains(r.Msg, ".acta.yaml") {
		t.Fatalf("msg %q does not carry the parse error", r.Msg)
	}
}

// TestDoctorFixNeverTouchesHome walks the whole home folder, so a new write
// anywhere under it shows up, not just the one path we thought of.
func TestDoctorFixNeverTouchesHome(t *testing.T) {
	e := env(t)
	write(t, filepath.Join(e.Home, ".claude.json"), `{"leftArrowOpensAgents":false}`)
	write(t, filepath.Join(e.Home, ".claude", "settings.json"), `{"enabledPlugins":{"acta@local":true}}`)
	link(t, filepath.Join(nodeModules(e), "pm"), filepath.Join(t.TempDir(), "gone"))
	before := snapshot(t, e.Home)
	if _, err := Fix(e); err != nil {
		t.Fatal(err)
	}
	after := snapshot(t, e.Home)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("home changed:\n%v\n%v", before, after)
	}
}

func TestDoctorFailed(t *testing.T) {
	if Failed([]Result{{Level: OK}, {Level: Warn}}) {
		t.Fatal("warn is not fail")
	}
	if !Failed([]Result{{Level: OK}, {Level: Fail}}) {
		t.Fatal("fail must fail")
	}
	if Failed(nil) {
		t.Fatal("no result cannot fail")
	}
}

func TestDoctorFormat(t *testing.T) {
	rs := []Result{
		{Name: "binary", Level: OK, Msg: "/bin/acta test"},
		{Name: "repo", Level: Fail, Msg: "no .acta folder", Fix: "acta doctor --fix"},
		{Name: "setup", Level: Warn, Msg: "no build executor", Fix: "/acta:setup"},
	}
	want := "ok binary: /bin/acta test\nfail repo: no .acta folder\nfix: acta doctor --fix\nwarn setup: no build executor\nfix: /acta:setup\n"
	if got := Format(rs); got != want {
		t.Fatalf("format %q want %q", got, want)
	}
}

// snapshot lists every path under root with its content or link target, so a
// second run can prove nothing was added, removed or rewritten.
func snapshot(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		fi, err := os.Lstat(p)
		if err != nil {
			return err
		}
		switch {
		case fi.Mode()&os.ModeSymlink != 0:
			target, err := os.Readlink(p)
			if err != nil {
				return err
			}
			out = append(out, rel+" -> "+target)
		case fi.IsDir():
			out = append(out, rel+"/")
		default:
			raw, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			out = append(out, rel+" "+string(raw))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
