package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
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
