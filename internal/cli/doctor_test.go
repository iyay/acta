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

// readOrMissing gives a file's bytes and whether it is there at all, so a
// file that was never made is not confused with an empty one.
func readOrMissing(t *testing.T, path string) (string, bool) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", false
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(raw), true
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

// writeTheme puts a voice file and, when a theme file is given, a user theme
// in the temp home, so one doctor run sees the whole setup.
func writeTheme(t *testing.T, home, name, themeFile string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(home, ".acta", "themes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if themeFile != "" {
		if err := os.WriteFile(filepath.Join(home, ".acta", "themes", name+".yaml"), []byte(themeFile), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	body := ""
	if name != "" {
		body = "theme: " + name + "\n"
	}
	if err := os.WriteFile(filepath.Join(home, ".acta", "voice.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// themeLine runs acta doctor over a temp home and returns the theme line it
// printed, plus the exit code.
func themeLine(t *testing.T, name, themeFile string) (string, int) {
	t.Helper()
	home := doctorHome(t)
	ompActa(t, home)
	writeTheme(t, home, name, themeFile)
	var stdout, stderr strings.Builder
	code := 0
	inDir(t, t.TempDir(), func() {
		code = Run([]string{"doctor"}, strings.NewReader(""), false, &stdout, &stderr)
	})
	for _, line := range strings.Split(stdout.String(), "\n") {
		if strings.HasPrefix(line, "ok theme:") || strings.HasPrefix(line, "warn theme:") || strings.HasPrefix(line, "fail theme:") {
			return line, code
		}
	}
	t.Fatalf("no theme line in %q (stderr %q)", stdout.String(), stderr.String())
	return "", code
}

// A theme must never make the command fail: the TUI falls back to the
// default, so the user keeps their board. This walks every way a theme can go
// wrong, from no theme at all to a user file that does not parse.
func TestDoctorCLIPrintsTheme(t *testing.T) {
	ansi := "ansi:\n"
	for range 16 {
		ansi += "  - '#0a0a0a'\n"
	}
	cases := []struct {
		name, theme, file, want string
	}{
		{"no theme set", "", "", "ok theme: theme tokyo-night loads"},
		{"built in", "dracula", "", "ok theme: theme dracula loads"},
		{"user file", "mine", "bg: '#111111'\nfg: '#eeeeee'\n" + ansi, "ok theme: theme mine loads"},
		{"unknown name", "nope", "", `warn theme: theme "nope": unknown theme`},
		{"bad name", "Mine!", "", "name may only use a-z, 0-9 and -"},
		{"file does not parse", "mine", "bg: [oops\n", "yaml"},
		{"bad hex", "mine", "bg: 'red'\nfg: '#eeeeee'\n" + ansi, `bg "red" is not #rrggbb`},
		{"wrong color count", "mine", "bg: '#111111'\nfg: '#eeeeee'\nansi: ['#0a0a0a']\n", "ansi needs 16 colors, has 1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			line, code := themeLine(t, c.theme, c.file)
			if code != 0 {
				t.Fatalf("exit %d, want 0: %s", code, line)
			}
			if !strings.Contains(line, c.want) {
				t.Fatalf("line %q does not hold %q", line, c.want)
			}
		})
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

// commitCount is how many commits the branch holds, so a test can tell a
// new commit from the ones that were already there.
func commitCount(t *testing.T, dir string) int {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "rev-list", "--count", "HEAD").CombinedOutput()
	if err != nil {
		t.Fatalf("rev-list: %v %s", err, out)
	}
	n := 0
	for _, c := range strings.TrimSpace(string(out)) {
		n = n*10 + int(c-'0')
	}
	return n
}

// A dirty .acta/.gitignore holds the user's own uncommitted lines. --fix
// still adds its own line on disk, but those lines must not ride along in a
// commit, and the run has to say why it skipped the commit.
func TestDoctorCLIFixDoesNotCommitADirtyGitignore(t *testing.T) {
	home := doctorHome(t)
	dir := doctorRepo(t)
	ompActa(t, home)
	gi := filepath.Join(dir, ".acta", ".gitignore")
	if err := os.MkdirAll(filepath.Dir(gi), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gi, []byte("mysecret.txt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := commitCount(t, dir)
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		Run([]string{"doctor", "--fix"}, strings.NewReader(""), false, &stdout, &stderr)
	})
	if got := commitCount(t, dir); got != before {
		t.Fatalf("commits %d, want %d: the dirty file was committed", got, before)
	}
	if !strings.Contains(stderr.String(), "not committed") {
		t.Fatalf("stderr %q does not say why the commit was skipped", stderr.String())
	}
	raw, err := os.ReadFile(gi)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), ".agents.json") {
		t.Fatalf("gitignore %q was not fixed on disk", raw)
	}
}

// auto_commit: false in .acta.yaml means the user wants nothing committed
// without asking. --fix still fixes the file, and says it did not commit.
func TestDoctorCLIFixMakesNoCommitWhenAutoCommitIsOff(t *testing.T) {
	home := doctorHome(t)
	dir := doctorRepo(t)
	ompActa(t, home)
	if err := os.WriteFile(filepath.Join(dir, ".acta.yaml"), []byte("auto_commit: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := commitCount(t, dir)
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		Run([]string{"doctor", "--fix"}, strings.NewReader(""), false, &stdout, &stderr)
	})
	if got := commitCount(t, dir); got != before {
		t.Fatalf("commits %d, want %d: auto_commit is off", got, before)
	}
	if !strings.Contains(stderr.String(), "auto_commit is off") {
		t.Fatalf("stderr %q does not name auto_commit", stderr.String())
	}
	raw, err := os.ReadFile(filepath.Join(dir, ".acta", ".gitignore"))
	if err != nil {
		t.Fatalf("the file was not fixed on disk: %v", err)
	}
	if !strings.Contains(string(raw), ".agents.json") {
		t.Fatalf("gitignore %q was not fixed on disk", raw)
	}
}

// root: in .acta.yaml can point out of the repo. --fix must write nothing
// there, and the repo check has to point at root, not at --fix.
func TestDoctorCLIFixWritesNothingWhenRootLeavesTheRepo(t *testing.T) {
	home := doctorHome(t)
	dir := doctorRepo(t)
	ompActa(t, home)
	if err := os.WriteFile(filepath.Join(dir, ".acta.yaml"), []byte("root: ../escaped\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	escaped := filepath.Join(filepath.Dir(dir), "escaped")
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		if code := Run([]string{"doctor", "--fix"}, strings.NewReader(""), false, &stdout, &stderr); code != 1 {
			t.Fatalf("exit %d want 1, stdout %q stderr %q", code, stdout.String(), stderr.String())
		}
	})
	if _, err := os.Stat(escaped); !os.IsNotExist(err) {
		t.Fatalf("--fix wrote %s outside the repo", escaped)
	}
	if _, err := os.Stat(filepath.Join(dir, ".acta")); !os.IsNotExist(err) {
		t.Fatal("--fix wrote .acta in the repo while root points out of it")
	}
	if !strings.Contains(stdout.String(), "fail repo:") {
		t.Fatalf("stdout %q has no fail repo line", stdout.String())
	}
	if strings.Contains(stdout.String(), "fix: acta doctor --fix") {
		t.Fatalf("stdout %q sends the user to --fix, which cannot fix root", stdout.String())
	}
	if !strings.Contains(stdout.String(), "root") {
		t.Fatalf("stdout %q does not name root", stdout.String())
	}
}

// A broken .acta.yaml is a real problem, not a skipped check: the parse
// error has to reach the report and the exit code.
func TestDoctorCLIFailsOnBrokenActaYaml(t *testing.T) {
	home := doctorHome(t)
	dir := doctorRepo(t)
	ompActa(t, home)
	if err := os.WriteFile(filepath.Join(dir, ".acta.yaml"), []byte("root: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		if code := Run([]string{"doctor"}, strings.NewReader(""), false, &stdout, &stderr); code != 1 {
			t.Fatalf("exit %d want 1, stdout %q stderr %q", code, stdout.String(), stderr.String())
		}
	})
	if !strings.Contains(stdout.String(), "fail repo:") {
		t.Fatalf("stdout %q has no fail repo line", stdout.String())
	}
	if !strings.Contains(stdout.String(), ".acta.yaml") {
		t.Fatalf("stdout %q does not carry the parse error", stdout.String())
	}
	if strings.Contains(stdout.String(), "skipped") {
		t.Fatalf("stdout %q skipped the repo check", stdout.String())
	}
}

// A clean repo has no reason to hold back, so --fix must leave one commit
// behind with its own message. Without this, dropping the commit entirely
// would still pass every other test.
func TestDoctorCLIFixCommitsOnACleanRepo(t *testing.T) {
	home := doctorHome(t)
	dir := doctorRepo(t)
	ompActa(t, home)
	before := commitCount(t, dir)
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		if code := Run([]string{"doctor", "--fix"}, strings.NewReader(""), false, &stdout, &stderr); code != 0 {
			t.Fatalf("exit %d want 0, stdout %q stderr %q", code, stdout.String(), stderr.String())
		}
	})
	if got := commitCount(t, dir); got != before+1 {
		t.Fatalf("commits %d, want %d: --fix made no commit", got, before+1)
	}
	out, err := exec.Command("git", "-C", dir, "log", "-1", "--pretty=%s").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if msg := strings.TrimSpace(string(out)); msg != "acta: doctor fix" {
		t.Fatalf("commit message %q want %q", msg, "acta: doctor fix")
	}
	if strings.Contains(stderr.String(), "not committed") {
		t.Fatalf("stderr %q says the commit was skipped", stderr.String())
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

// root: in .acta.yaml can name a link that points out of the repo. --fix
// must write nothing through it, and the report must fail.
func TestDoctorCLIFixWritesNothingWhenRootIsALinkOut(t *testing.T) {
	home := doctorHome(t)
	dir := doctorRepo(t)
	ompActa(t, home)
	outside := filepath.Join(filepath.Dir(dir), "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join("..", "outside"), filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".acta.yaml"), []byte("root: link\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		if code := Run([]string{"doctor", "--fix"}, strings.NewReader(""), false, &stdout, &stderr); code != 1 {
			t.Fatalf("exit %d want 1, stdout %q stderr %q", code, stdout.String(), stderr.String())
		}
	})
	if got, err := os.ReadDir(outside); err != nil {
		t.Fatal(err)
	} else if len(got) != 0 {
		t.Fatalf("--fix wrote %v outside the repo", got)
	}
	if !strings.Contains(stdout.String(), "fail repo:") {
		t.Fatalf("stdout %q has no fail repo line", stdout.String())
	}
}

// A .acta/.gitignore that is a link to a file outside the repo must not be
// read as the repo's own ignore file, and --fix must leave that file with the
// bytes it had.
func TestDoctorCLIFixLeavesAGitignoreLinkAlone(t *testing.T) {
	home := doctorHome(t)
	dir := doctorRepo(t)
	ompActa(t, home)
	victim := filepath.Join(t.TempDir(), "victim.conf")
	if err := os.WriteFile(victim, []byte("notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".acta"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(victim, filepath.Join(dir, ".acta", ".gitignore")); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		if code := Run([]string{"doctor"}, strings.NewReader(""), false, &stdout, &stderr); code != 1 {
			t.Fatalf("exit %d want 1, stdout %q stderr %q", code, stdout.String(), stderr.String())
		}
	})
	if !strings.Contains(stdout.String(), "fail repo:") {
		t.Fatalf("stdout %q has no fail repo line", stdout.String())
	}
	if strings.Contains(stdout.String(), "fix: acta doctor --fix") {
		t.Fatalf("stdout %q sends the user to --fix, which cannot follow the link", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	inDir(t, dir, func() {
		if code := Run([]string{"doctor", "--fix"}, strings.NewReader(""), false, &stdout, &stderr); code != 1 {
			t.Fatalf("exit %d want 1, stdout %q stderr %q", code, stdout.String(), stderr.String())
		}
	})
	raw, err := os.ReadFile(victim)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "notes\n" {
		t.Fatalf("--fix wrote through the link: %q", raw)
	}
}

// A .acta/.gitignore that is a link must not be written through, whichever
// way it points: a target that is not there yet would be created outside the
// repo, and a target inside the repo would still be someone else's file. The
// report has to say fail and exit 1, never ok, and --fix must commit nothing.
func TestDoctorCLIFixRefusesAGitignoreLink(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, dir string) string
	}{
		{"link to a file that is not there yet", func(t *testing.T, dir string) string {
			outside := filepath.Join(filepath.Dir(dir), "outside")
			if err := os.MkdirAll(outside, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(dir, ".acta"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join("..", "..", "outside", "new.conf"), filepath.Join(dir, ".acta", ".gitignore")); err != nil {
				t.Fatal(err)
			}
			return filepath.Join(outside, "new.conf")
		}},
		{"link to a file of the same repo", func(t *testing.T, dir string) string {
			victim := filepath.Join(dir, "docs", "notes.txt")
			if err := os.MkdirAll(filepath.Join(dir, "docs"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(victim, []byte("release notes\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(dir, ".acta"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join("..", "docs", "notes.txt"), filepath.Join(dir, ".acta", ".gitignore")); err != nil {
				t.Fatal(err)
			}
			return victim
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			home := doctorHome(t)
			dir := doctorRepo(t)
			ompActa(t, home)
			victim := c.setup(t, dir)
			before, existed := readOrMissing(t, victim)
			commits := commitCount(t, dir)
			var stdout, stderr strings.Builder
			inDir(t, dir, func() {
				if code := Run([]string{"doctor", "--fix"}, strings.NewReader(""), false, &stdout, &stderr); code != 1 {
					t.Fatalf("exit %d want 1, stdout %q stderr %q", code, stdout.String(), stderr.String())
				}
			})
			out := stdout.String()
			if !strings.Contains(out, "fail repo:") {
				t.Fatalf("stdout %q has no fail repo line", out)
			}
			if strings.Contains(out, "ok repo:") {
				t.Fatalf("stdout %q calls a linked .gitignore ok", out)
			}
			if got := commitCount(t, dir); got != commits {
				t.Fatalf("commits %d, want %d: --fix made a commit", got, commits)
			}
			if got, now := readOrMissing(t, victim); now != existed || got != before {
				t.Fatalf("%s went from %q (%t) to %q (%t)", victim, before, existed, got, now)
			}
		})
	}
}

// A repo reached through a symlinked path is still the same repo, so the
// root spelled with the link must pass the check and --fix must commit.
func TestDoctorCLIFixWorksThroughASymlinkedRepoPath(t *testing.T) {
	home := doctorHome(t)
	dir := doctorRepo(t)
	ompActa(t, home)
	via := filepath.Join(t.TempDir(), "repo-link")
	if err := os.Symlink(dir, via); err != nil {
		t.Fatal(err)
	}
	// git answers with the real path, the root keeps the link, so the two
	// spell the same folder in two ways.
	t.Setenv("ACTA_ROOT", filepath.Join(via, ".acta"))
	before := commitCount(t, dir)
	var stdout, stderr strings.Builder
	inDir(t, via, func() {
		if code := Run([]string{"doctor", "--fix"}, strings.NewReader(""), false, &stdout, &stderr); code != 0 {
			t.Fatalf("exit %d want 0, stdout %q stderr %q", code, stdout.String(), stderr.String())
		}
	})
	if !strings.Contains(stdout.String(), "ok repo:") {
		t.Fatalf("stdout %q does not pass the repo check", stdout.String())
	}
	if got := commitCount(t, dir); got != before+1 {
		t.Fatalf("commits %d, want %d: --fix made no commit", got, before+1)
	}
}

// scratchFile writes one scratch file under .acta/scratch, so the schema
// check has a board to read. opt is the "schema: 1" line, or "" for an old
// file the check has to skip. sections is what follows the title.
func scratchFile(t *testing.T, dir, name, opt, sections string) {
	t.Helper()
	slug := strings.ToUpper(strings.TrimSuffix(strings.TrimPrefix(name, "2026-09-29-"), ".md"))
	body := "---\nid: SCRATCH-" + slug + "\ntitle: " + slug + "\nstatus: raw\n" + opt +
		"---\n# " + slug + "\n" + sections
	path := filepath.Join(dir, ".acta", "scratch", name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// badScratch is a schema file that lost its required "## Words" section.
func badScratch(t *testing.T, dir, name string) {
	t.Helper()
	scratchFile(t, dir, name, "schema: 1\n", "")
}

// doctorRun runs the doctor in dir and gives back the report.
func doctorRun(t *testing.T, dir string) string {
	t.Helper()
	var stdout, stderr strings.Builder
	inDir(t, dir, func() {
		Run([]string{"doctor"}, strings.NewReader(""), false, &stdout, &stderr)
	})
	return stdout.String()
}

func TestDoctorCLISchemaCheck(t *testing.T) {
	t.Run("no acta folder", func(t *testing.T) {
		home := doctorHome(t)
		dir := doctorRepo(t)
		ompActa(t, home)
		if got := doctorRun(t, dir); !strings.Contains(got, "ok schema:") {
			t.Fatalf("stdout %q has no ok schema line", got)
		}
	})
	t.Run("clean schema file", func(t *testing.T) {
		home := doctorHome(t)
		dir := doctorRepo(t)
		ompActa(t, home)
		scratchFile(t, dir, "2026-09-29-good.md", "schema: 1\n", "\n## Words\n\nhello\n")
		if got := doctorRun(t, dir); !strings.Contains(got, "ok schema:") {
			t.Fatalf("stdout %q has no ok schema line", got)
		}
	})
	t.Run("one bad schema file, an old one ignored", func(t *testing.T) {
		home := doctorHome(t)
		dir := doctorRepo(t)
		ompActa(t, home)
		badScratch(t, dir, "2026-09-29-bad.md")
		scratchFile(t, dir, "2026-09-29-old.md", "", "")
		got := doctorRun(t, dir)
		if !strings.Contains(got, "warn schema: scratch 2026-09-29-bad.md: missing ## Words") {
			t.Fatalf("stdout %q has no warn schema line", got)
		}
		if strings.Contains(got, "2026-09-29-old.md") {
			t.Fatalf("stdout %q names a file without schema: 1", got)
		}
	})
	t.Run("two bad files are both named, sorted", func(t *testing.T) {
		home := doctorHome(t)
		dir := doctorRepo(t)
		ompActa(t, home)
		badScratch(t, dir, "2026-09-29-zzz.md")
		badScratch(t, dir, "2026-09-29-aaa.md")
		got := doctorRun(t, dir)
		want := "warn schema: scratch 2026-09-29-aaa.md: missing ## Words; scratch 2026-09-29-zzz.md: missing ## Words"
		if !strings.Contains(got, want) {
			t.Fatalf("stdout %q has no %q", got, want)
		}
	})
}
