package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iyay/acta/internal/update"
	"github.com/iyay/acta/plugin"
)

// noticeEnv points the user dir at a temp dir and swaps the spawn and the
// terminal check. It returns the cache path and a count of spawns.
func noticeEnv(t *testing.T, release, tty bool) (cachePath string, spawns *int) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("PM_VOICE_FILE", filepath.Join(dir, "config.yaml"))
	t.Setenv("ACTA_NO_UPDATE_CHECK", "")
	oldRelease, oldSpawn, oldTTY := isRelease, spawnCheck, stderrIsTerminal
	t.Cleanup(func() { isRelease, spawnCheck, stderrIsTerminal = oldRelease, oldSpawn, oldTTY })
	isRelease = func() bool { return release }
	n := 0
	spawnCheck = func() error { n++; return nil }
	stderrIsTerminal = func(any) bool { return tty }
	return filepath.Join(dir, "update.json"), &n
}

func runAll(args ...string) (code int, errOut string) {
	var o, e strings.Builder
	code = Run(args, strings.NewReader(""), false, &o, &e)
	return code, e.String()
}

func makeStale(t *testing.T, path string) {
	t.Helper()
	old := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
}

// Every path through Run that may start a check: it starts one only for a
// release build with a stale cache, with no opt-out, and never for update.
func TestUpdateCheckSpawn(t *testing.T) {
	for _, tc := range []struct {
		name    string
		release bool
		env     string
		config  string
		args    []string
		want    int
	}{
		{"stale release build", true, "", "", []string{"bogus-cmd"}, 1},
		{"hook still checks", true, "", "", []string{"hook", "bogus"}, 1},
		{"source build", false, "", "", []string{"bogus-cmd"}, 0},
		{"env opt-out", true, "1", "", []string{"bogus-cmd"}, 0},
		{"config opt-out", true, "", "update_check: off\n", []string{"bogus-cmd"}, 0},
		{"update itself", true, "", "", []string{"update", "--bogus"}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cache, spawns := noticeEnv(t, tc.release, false)
			t.Setenv("ACTA_NO_UPDATE_CHECK", tc.env)
			if tc.config != "" {
				body := "chat_language: English\nstyle: adhd\nrepo_language: English\n" + tc.config
				if err := os.WriteFile(filepath.Join(filepath.Dir(cache), "config.yaml"), []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			runAll(tc.args...)
			if *spawns != tc.want {
				t.Fatalf("spawns = %d, want %d", *spawns, tc.want)
			}
			if tc.want == 1 {
				if _, err := os.Stat(cache); err != nil {
					t.Fatalf("cache not touched before spawn: %v", err)
				}
				// The touch makes the cache fresh, so the next run waits.
				runAll(tc.args...)
				if *spawns != 1 {
					t.Fatalf("second run spawned again: %d", *spawns)
				}
				makeStale(t, cache)
				runAll(tc.args...)
				if *spawns != 2 {
					t.Fatalf("stale again but spawns = %d, want 2", *spawns)
				}
			}
		})
	}
}

func TestUpdateNoticeLine(t *testing.T) {
	current := "v" + plugin.Version()
	want := "acta v99.0.0 available (you have " + current + "). Run: acta update"
	for _, tc := range []struct {
		name   string
		cached string
		tty    bool
		args   []string
		show   bool
	}{
		{"newer on a terminal", "v99.0.0", true, []string{"bogus-cmd"}, true},
		{"not a terminal", "v99.0.0", false, []string{"bogus-cmd"}, false},
		{"hook", "v99.0.0", true, []string{"hook", "bogus"}, false},
		{"tui with flags", "v99.0.0", true, []string{"--root", "/nonexistent"}, false},
		{"json flag", "v99.0.0", true, []string{"list", "--json", "--root", "/nonexistent"}, false},
		{"same version", current, true, []string{"bogus-cmd"}, false},
		{"older version", "v0.0.1", true, []string{"bogus-cmd"}, false},
		{"no cache", "", true, []string{"bogus-cmd"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cache, _ := noticeEnv(t, true, tc.tty)
			if tc.cached != "" {
				if err := update.WriteCache(cache, tc.cached); err != nil {
					t.Fatal(err)
				}
			}
			_, errOut := runAll(tc.args...)
			if got := strings.Contains(errOut, want); got != tc.show {
				t.Fatalf("notice shown = %v, want %v; stderr %q", got, tc.show, errOut)
			}
			if strings.Count(errOut, "available (you have") > 1 {
				t.Fatalf("notice printed more than once: %q", errOut)
			}
		})
	}
}

// update --check --cache writes the cache on success, and on failure prints
// nothing and leaves the old cache as it was.
func TestUpdateCheckCache(t *testing.T) {
	updateEnv(t, true)
	cache, _ := noticeEnv(t, true, false)
	releaseServer(t, "v99.0.0", false)
	code, out, errOut := runUpdate("--check", "--cache")
	if code != 0 || out != "" || errOut != "" {
		t.Fatalf("code=%d out=%q err=%q, want 0 and silent", code, out, errOut)
	}
	if got := update.ReadCache(cache); got != "v99.0.0" {
		t.Fatalf("cache = %q, want v99.0.0", got)
	}

	t.Setenv("ACTA_DOWNLOAD_URL", "https://127.0.0.1:1/never")
	code, out, errOut = runUpdate("--check", "--cache")
	if out != "" || errOut != "" {
		t.Fatalf("failed check printed out=%q err=%q", out, errOut)
	}
	_ = code
	if got := update.ReadCache(cache); got != "v99.0.0" {
		t.Fatalf("failed check changed cache to %q", got)
	}
}

// A good update writes the new version, so the notice goes away at once.
func TestUpdateWritesCache(t *testing.T) {
	updateEnv(t, true)
	cache, _ := noticeEnv(t, true, false)
	releaseServer(t, "v99.0.0", false)
	if code, _, errOut := runUpdate(); code != 0 {
		t.Fatalf("update failed: %s", errOut)
	}
	if got := update.ReadCache(cache); got != "v99.0.0" {
		t.Fatalf("cache = %q, want v99.0.0", got)
	}
}

// session-start only passes the cached newer version on. One run with a
// newer version cached, one with none.
func TestHookSessionStartUpdateLine(t *testing.T) {
	for _, tc := range []struct {
		cached string
		want   bool
	}{{"v99.0.0", true}, {"", false}} {
		cache, _ := noticeEnv(t, true, false)
		t.Chdir(t.TempDir())
		if tc.cached != "" {
			if err := update.WriteCache(cache, tc.cached); err != nil {
				t.Fatal(err)
			}
		}
		out := mustRun(t, "hook", "session-start")
		line := "acta update available: v99.0.0 (you have v" + plugin.Version() + ")."
		if got := strings.Contains(out, line); got != tc.want {
			t.Errorf("cached %q: line shown = %v, want %v\n%s", tc.cached, got, tc.want, out)
		}
	}
}
