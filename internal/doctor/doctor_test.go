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

	"github.com/iyay/acta/internal/voice"
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
		VoiceExists: true, Voice: voice.Voice{BuildExecutor: "subagent"}}
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

func TestDoctorChecksRunInFixedOrder(t *testing.T) {
	want := []string{"binary", "harness", "stale-links", "conflicts", "repo", "agents-view", "setup"}
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
	t.Run("one dead link", func(t *testing.T) {
		e := env(t)
		link(t, filepath.Join(nodeModules(e), "pm"), filepath.Join(t.TempDir(), "gone"))
		r := byName(Run(e), "stale-links")
		wantLevel(t, r, Warn, "omp plugin unlink pm")
		if !strings.Contains(r.Msg, "pm") {
			t.Fatalf("msg %q does not name pm", r.Msg)
		}
	})
	t.Run("live link and folder beside it", func(t *testing.T) {
		e := env(t)
		link(t, filepath.Join(nodeModules(e), "acta"), t.TempDir())
		if err := os.MkdirAll(filepath.Join(nodeModules(e), "zeta"), 0o755); err != nil {
			t.Fatal(err)
		}
		wantLevel(t, byName(Run(e), "stale-links"), OK, "")
	})
	t.Run("no folder", func(t *testing.T) {
		e := env(t)
		wantLevel(t, byName(Run(e), "stale-links"), OK, "")
	})
	t.Run("two dead links, one fix line each", func(t *testing.T) {
		e := env(t)
		link(t, filepath.Join(nodeModules(e), "pm"), filepath.Join(t.TempDir(), "gone"))
		link(t, filepath.Join(nodeModules(e), "gstack"), filepath.Join(t.TempDir(), "gone2"))
		r := byName(Run(e), "stale-links")
		wantLevel(t, r, Warn, "omp plugin unlink gstack; omp plugin unlink pm")
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
		wantLevel(t, byName(Run(e), "conflicts"), OK, "")
	})
	t.Run("known file missing", func(t *testing.T) {
		e := env(t)
		e.KnownFile = filepath.Join(t.TempDir(), "gone.txt")
		wantLevel(t, byName(Run(e), "conflicts"), OK, "")
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

func TestDoctorSetup(t *testing.T) {
	t.Run("no voice file", func(t *testing.T) {
		e := env(t)
		e.VoiceExists = false
		wantLevel(t, byName(Run(e), "setup"), Warn, "/acta:setup")
	})
	t.Run("no build executor", func(t *testing.T) {
		e := env(t)
		e.Voice = voice.Voice{}
		wantLevel(t, byName(Run(e), "setup"), Warn, "/acta:setup")
	})
	t.Run("all set", func(t *testing.T) {
		e := env(t)
		wantLevel(t, byName(Run(e), "setup"), OK, "")
	})
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
