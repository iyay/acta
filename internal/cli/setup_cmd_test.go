package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/setup"
)

// TestSetupNoTTY checks the paths where acta setup must write nothing and
// run no harness command: stdin not a TTY (piped input), and stdout not a
// TTY (a strings.Builder is never *os.File). Both exit non-zero and name
// acta config set, while the home and the working directory gain no file.
func TestSetupNoTTY(t *testing.T) {
	// stdin piped: Run gets stdinIsTTY=false.
	t.Run("stdin piped", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("PM_VOICE_FILE", filepath.Join(home, "config.yaml"))
		t.Setenv("TMPDIR", home)
		dir := t.TempDir()
		var stdout, stderr strings.Builder
		var code int
		inDir(t, dir, func() {
			code = Run([]string{"setup"}, strings.NewReader(""), false, &stdout, &stderr)
		})
		if code == 0 {
			t.Fatal("exit 0, want non-zero")
		}
		if !strings.Contains(stderr.String(), "acta config set") {
			t.Fatalf("stderr %q names no acta config set", stderr.String())
		}
		if entries, _ := os.ReadDir(home); len(entries) != 0 {
			t.Fatalf("home gained files: %v", entries)
		}
		if entries, _ := os.ReadDir(dir); len(entries) != 0 {
			t.Fatalf("repo dir gained files: %v", entries)
		}
	})

	// stdout piped: stdin is a TTY but stdoutIsTTY sees a Builder.
	t.Run("stdout piped", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("PM_VOICE_FILE", filepath.Join(home, "config.yaml"))
		t.Setenv("TMPDIR", home)
		dir := t.TempDir()
		var stdout, stderr strings.Builder
		var code int
		inDir(t, dir, func() {
			code = Run([]string{"setup"}, strings.NewReader(""), true, &stdout, &stderr)
		})
		if code == 0 {
			t.Fatal("exit 0, want non-zero")
		}
		if !strings.Contains(stderr.String(), "acta config set") {
			t.Fatalf("stderr %q names no acta config set", stderr.String())
		}
		if entries, _ := os.ReadDir(home); len(entries) != 0 {
			t.Fatalf("home gained files: %v", entries)
		}
	})

	// Unknown flags still fail as bad input, not as no-TTY.
	t.Run("bad flag", func(t *testing.T) {
		var stdout, stderr strings.Builder
		if code := Run([]string{"setup", "--nope"}, strings.NewReader(""), false, &stdout, &stderr); code != exitBadInput {
			t.Fatalf("exit %d, want %d", code, exitBadInput)
		}
		if !strings.Contains(stderr.String(), setupUsage) {
			t.Fatalf("stderr %q misses the usage", stderr.String())
		}
	})
}

// TestSetupEnvFindings checks the Env the command builds: harnesses come
// from PATH, the git root and the CLAUDE.md/AGENTS.md flags come from the
// working directory, and the plugin dir comes from the flag.
func TestSetupEnvFindings(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if out, err := gitIn(dir, "init", "-q", "-b", "main"); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	for _, name := range []string{"CLAUDE.md", "AGENTS.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	bin := t.TempDir()
	for _, name := range []string{"claude", "omp"} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	inDir(t, dir, func() {
		got := setupEnv("/p")
		if len(got.Harnesses) != 2 || got.Harnesses[0] != "claude" || got.Harnesses[1] != "omp" {
			t.Fatalf("harnesses = %v, want [claude omp]", got.Harnesses)
		}
		if got.PluginDir != "/p" {
			t.Fatalf("plugin dir = %q, want /p", got.PluginDir)
		}
		if got.RepoRoot != dir {
			t.Fatalf("repo root = %q, want %q", got.RepoRoot, dir)
		}
		if !got.HasClaudeMD || !got.HasAgentsMD {
			t.Fatalf("file flags = %+v, want both true", got)
		}
	})
	plain := t.TempDir()
	inDir(t, plain, func() {
		got := setupEnv("")
		if got.RepoRoot != "" {
			t.Fatalf("repo root = %q, want empty", got.RepoRoot)
		}
	})
}

// TestSetupEnvCurrent checks setupEnv fills Env.Current from the real user
// config: the file values when one exists, else the built-in defaults the
// form falls back to.
func TestSetupEnvCurrent(t *testing.T) {
	t.Run("from config file", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		cfg := filepath.Join(home, "config.yaml")
		t.Setenv("PM_VOICE_FILE", cfg)
		t.Setenv("TMPDIR", home)
		if err := os.WriteFile(cfg, []byte("chat_language: Korean\nstyle: plain\nrepo_language: English\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		got := setupEnv("")
		if got.Current.ChatLanguage != "Korean" || got.Current.Style != "plain" {
			t.Fatalf("current = %+v, want Korean/plain", got.Current)
		}
	})
	t.Run("no config file", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("PM_VOICE_FILE", filepath.Join(home, "config.yaml"))
		t.Setenv("TMPDIR", home)
		got := setupEnv("")
		if got.Current != config.UserDefault() {
			t.Fatalf("current = %+v, want defaults %+v", got.Current, config.UserDefault())
		}
	})
}

// TestBlockNoteShort checks the note cmdSetup prints before Apply: it names
// each block file from the live plan actions on its own short line and never
// prints the block itself.
func TestBlockNoteShort(t *testing.T) {
	dir := t.TempDir()
	claude := filepath.Join(dir, "CLAUDE.md")
	agents := filepath.Join(dir, "AGENTS.md")
	for _, c := range []struct {
		name string
		env  setup.Env
		want string
	}{
		{"both files", setup.Env{RepoRoot: dir, HasClaudeMD: true, HasAgentsMD: true},
			"│  acta block → " + claude + "\n│  acta block → " + agents + "\n"},
		{"agents only", setup.Env{RepoRoot: dir, HasAgentsMD: true},
			"│  acta block → " + agents + "\n"},
		{"no file yet", setup.Env{RepoRoot: dir},
			"│  acta block → " + claude + "\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			answers := setup.Answers{Install: map[string]bool{}}
			env := c.env
			env.TTY = true
			got := blockNote(setup.Plan(answers, env))
			if got != c.want {
				t.Fatalf("block note = %q, want %q", got, c.want)
			}
			if strings.Contains(got, setup.Block) {
				t.Fatalf("block note prints the block text")
			}
		})
	}
}

// TestSetupSummaryBox checks the closing box names the voice file and the
// TUI hint, and no longer lists installs, block files or a next line.
func TestSetupSummaryBox(t *testing.T) {
	home := t.TempDir()
	voice := filepath.Join(home, "config.yaml")
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", voice)
	t.Setenv("TMPDIR", home)
	got := setupSummary()
	for _, want := range []string{"config: " + voice, "└  Run acta in a repo to browse specs, plans and bugs.\n"} {
		if !strings.Contains(got, want) {
			t.Fatalf("summary %q misses %q", got, want)
		}
	}
	for _, bad := range []string{"installed:", "block:", "next:"} {
		if strings.Contains(got, bad) {
			t.Fatalf("summary %q holds %q", got, bad)
		}
	}
}

// TestApplySetupHardError keeps Apply's rule: a bad config value stops the
// run and comes back as an error.
func TestApplySetupHardError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", filepath.Join(home, "config.yaml"))
	t.Setenv("TMPDIR", home)
	bad := config.User{Style: "nope"}
	var out strings.Builder
	if err := setup.Apply([]setup.Action{{Kind: setup.ActionConfig, User: bad}}, setupRunner{}, &out); err == nil {
		t.Fatal("want an error for an invalid config value")
	}
}

// TestSetupRunnerNoStdin checks an installer that asks a question does not
// hang the wizard. With no stdin it reads end of input, exits 1, and the run
// is reported as failed.
func TestSetupRunnerNoStdin(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", filepath.Join(home, "voice.yaml"))
	t.Setenv("TMPDIR", home)
	bin := t.TempDir()
	script := "#!/bin/sh\nread x || exit 1\nexit 0\n"
	if err := os.WriteFile(filepath.Join(bin, "fake-installer"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	// Give the process a stdin that never ends, so an inherited stdin hangs.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	hold := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = hold; r.Close() }()

	done := make(chan error, 1)
	go func() { done <- setupRunner{}.Run([]string{"fake-installer"}) }()
	select {
	case err := <-done:
		if err == nil {
			t.Error("installer that wanted input must be reported as failed")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("installer blocked on stdin")
	}
}
