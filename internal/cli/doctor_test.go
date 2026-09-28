package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// doctorHome points HOME, CLAUDE_CONFIG_DIR and the voice file at a temp dir,
// so the command reads nothing of the real user and writes nothing there.
func doctorHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(home, ".claude"))
	t.Setenv("PM_VOICE_FILE", filepath.Join(home, ".acta", "voice.yaml"))
	return home
}

// doctorRepo is a git repo with one commit, ready for a commit from --fix.
func doctorRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("GIT_AUTHOR_NAME", "test")
	t.Setenv("GIT_COMMITTER_NAME", "test")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")
	run := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	run("init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("repo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-q", "-m", "init")
	return dir
}

// ompActa links a live plugin folder into the temp home, so the harness
// check passes and only the repo line can fail.
func ompActa(t *testing.T, home string) {
	t.Helper()
	plugin := filepath.Join(t.TempDir(), "plugin")
	if err := os.MkdirAll(plugin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".omp", "plugins", "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(plugin, filepath.Join(home, ".omp", "plugins", "node_modules", "acta")); err != nil {
		t.Fatal(err)
	}
}

func TestDoctorCLIFailsWhenRepoIsBroken(t *testing.T) {
	home := doctorHome(t)
	dir := doctorRepo(t)
	// The harness check needs a live plugin, so a real user only sees the
	// repo line failing.
	ompActa(t, home)
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		if code := Run([]string{"doctor"}, strings.NewReader(""), false, &stdout, &stderr); code != 1 {
			t.Fatalf("exit %d want 1, stdout %q stderr %q", code, stdout.String(), stderr.String())
		}
	})
	if !strings.Contains(stdout.String(), "fail repo:") {
		t.Fatalf("stdout %q has no fail repo line", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".acta")); !os.IsNotExist(err) {
		t.Fatal("a plain run wrote .acta")
	}
}

func TestDoctorCLIFixThenPlainRunIsClean(t *testing.T) {
	home := doctorHome(t)
	dir := doctorRepo(t)
	ompActa(t, home)
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		if code := Run([]string{"doctor", "--fix"}, strings.NewReader(""), false, &stdout, &stderr); code != 0 {
			t.Fatalf("--fix exit %d want 0, stdout %q stderr %q", code, stdout.String(), stderr.String())
		}
	})
	if !strings.Contains(stdout.String(), "fixed repo:") {
		t.Fatalf("stdout %q does not name the fix", stdout.String())
	}
	// A second run has nothing left to do, so it prints no fix line and
	// exits 0.
	var second, errOut strings.Builder
	inDir(t, dir, func() {
		if code := Run([]string{"doctor"}, strings.NewReader(""), false, &second, &errOut); code != 0 {
			t.Fatalf("second run exit %d want 0, stdout %q stderr %q", code, second.String(), errOut.String())
		}
	})
	if !strings.Contains(second.String(), "ok repo:") {
		t.Fatalf("stdout %q has no ok repo line", second.String())
	}
	if strings.Contains(second.String(), "fail ") {
		t.Fatalf("stdout %q still fails after --fix", second.String())
	}
	raw, err := os.ReadFile(filepath.Join(dir, ".acta", ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(raw), ".agents.json"); n != 1 {
		t.Fatalf(".gitignore holds .agents.json %d times: %q", n, raw)
	}
}

func TestDoctorCLIRejectsExtraArgs(t *testing.T) {
	doctorHome(t)
	var stdout, stderr strings.Builder
	if code := Run([]string{"doctor", "extra"}, strings.NewReader(""), false, &stdout, &stderr); code != exitBadInput {
		t.Fatalf("exit %d want %d", code, exitBadInput)
	}
	if !strings.Contains(stderr.String(), doctorUsage) {
		t.Fatalf("stderr %q has no usage", stderr.String())
	}
}

func TestDoctorCLIRejectsUnknownFlag(t *testing.T) {
	doctorHome(t)
	var stdout, stderr strings.Builder
	if code := Run([]string{"doctor", "--nope"}, strings.NewReader(""), false, &stdout, &stderr); code != exitBadInput {
		t.Fatalf("exit %d want %d", code, exitBadInput)
	}
}

// A run outside a git repo must not fail: the repo check skips, so the exit
// code follows the other checks.
func TestDoctorCLIOutsideGitRepo(t *testing.T) {
	home := doctorHome(t)
	ompActa(t, home)
	dir := t.TempDir()
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		Run([]string{"doctor", "--fix"}, strings.NewReader(""), false, &stdout, &stderr)
	})
	if !strings.Contains(stdout.String(), "ok repo:") || !strings.Contains(stdout.String(), "skipped") {
		t.Fatalf("stdout %q does not skip the repo check", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".acta")); !os.IsNotExist(err) {
		t.Fatal("--fix wrote under the home folder")
	}
}

// --fix must not commit anything it did not write, so a dirty file stays.
func TestDoctorCLIFixCommitsOnlyItsOwnPaths(t *testing.T) {
	home := doctorHome(t)
	dir := doctorRepo(t)
	ompActa(t, home)
	dirty := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(dirty, []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		Run([]string{"doctor", "--fix"}, strings.NewReader(""), false, &stdout, &stderr)
	})
	out, err := exec.Command("git", "-C", dir, "status", "--porcelain").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "notes.txt") {
		t.Fatalf("git status %q: notes.txt was committed", out)
	}
}

// A plugin folder linked into omp carries the clashes list, so the check
// works without a flag. A dead link gives nothing, and the check skips.
func TestDoctorCLIFindsKnownFileInLinkedPlugin(t *testing.T) {
	home := doctorHome(t)
	plugin := filepath.Join(t.TempDir(), "plugin")
	if err := os.MkdirAll(filepath.Join(plugin, "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plugin, "hooks", "workflow-plugins.txt"), []byte("superpowers\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".omp", "plugins", "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(plugin, filepath.Join(home, ".omp", "plugins", "node_modules", "acta")); err != nil {
		t.Fatal(err)
	}
	if got, want := linkedKnownFile(home), filepath.Join(plugin, "hooks", "workflow-plugins.txt"); got != want {
		t.Fatalf("known file %q want %q", got, want)
	}
	// With a clashing plugin enabled the report must warn, not skip.
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	settings := filepath.Join(home, ".claude", "settings.json")
	if err := os.WriteFile(settings, []byte(`{"enabledPlugins":{"superpowers@x":true}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	Run([]string{"doctor"}, strings.NewReader(""), false, &stdout, &stderr)
	if !strings.Contains(stdout.String(), "warn conflicts:") {
		t.Fatalf("stdout %q has no conflicts warning", stdout.String())
	}
}

func TestDoctorCLIKnownFileMissingSkipsConflicts(t *testing.T) {
	home := doctorHome(t)
	ompActa(t, home)
	if got := linkedKnownFile(home); got != "" {
		t.Fatalf("known file %q, a plugin folder without the list must give none", got)
	}
	var stdout, stderr strings.Builder
	Run([]string{"doctor"}, strings.NewReader(""), false, &stdout, &stderr)
	if !strings.Contains(stdout.String(), "ok conflicts:") {
		t.Fatalf("stdout %q does not skip the conflicts check", stdout.String())
	}
}
