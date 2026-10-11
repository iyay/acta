package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Every test repo has two dates. Its files are committed on wikiLongAgo. A page
// carries a timestamp between the two dates, so it is fresh until a commit made
// on wikiLater touches its paths.
const (
	wikiLongAgo = "2020-01-01T00:00:00Z"
	wikiLater   = "2030-01-01T00:00:00Z"
)

// The lines the usual pages print, with the line break at the end.
const (
	glossaryLine = ".acta/wiki/glossary.md: Words this repo uses\n"
	lockLine     = ".acta/wiki/test-lock.md: Run tests with scripts/test\n"
	wrapLine     = ".acta/wiki/tui-wrap.md: Use Hardwrap(Wordwrap(...)) for wide lines\n"
)

// wikiPage is a page that keeps every rule.
func wikiPage(typ, description, paths string) string {
	return "---\ntype: " + typ + "\ntitle: A page\ndescription: " + description +
		"\npaths: [" + paths + "]\ntimestamp: 2026-01-01T00:00:00Z\n---\nbody\n"
}

// wikiFiles are the usual pages and the files they cover. Every page is fresh.
func wikiFiles() map[string]string {
	return map[string]string{
		".acta/wiki/glossary.md":  wikiPage("Glossary", "Words this repo uses", ""),
		".acta/wiki/test-lock.md": wikiPage("Decision", "Run tests with scripts/test", "scripts/test"),
		".acta/wiki/tui-wrap.md":  wikiPage("Gotcha", "Use Hardwrap(Wordwrap(...)) for wide lines", "internal/tui/"),
		"internal/tui/model.go":   "package tui\n",
		"scripts/test":            "#!/bin/sh\n",
	}
}

// wikiGit runs git in repo and gives back what it printed. A date makes both
// git dates that time, so a commit lands on the day a test wants. No date
// leaves them alone.
func wikiGit(t *testing.T, repo, when string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	if when != "" {
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+when, "GIT_COMMITTER_DATE="+when)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// wikiWrite writes files into repo, making folders as needed. A file that is
// there already is replaced.
func wikiWrite(t *testing.T, repo string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		file := filepath.Join(repo, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// wikiRepo makes a git repo that holds files, committed on wikiLongAgo, and goes
// into it. HOME is a temp folder and the root variables are empty, so nothing
// of the real user and no root from the shell leaks in.
func wikiRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ACTA_ROOT", "")
	t.Setenv("PM_ROOT", "")
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	wikiGit(t, repo, "", "init", "-q", "-b", "main")
	wikiGit(t, repo, "", "config", "user.name", "test")
	wikiGit(t, repo, "", "config", "user.email", "test@example.com")
	wikiWrite(t, repo, files)
	wikiGit(t, repo, wikiLongAgo, "add", ".")
	wikiGit(t, repo, wikiLongAgo, "commit", "-q", "-m", "init")
	t.Chdir(repo)
	return repo
}

// wikiSnapshot is the HEAD commit plus every file git sees as changed, new or
// ignored. A command that wrote a file or made a commit would change it.
func wikiSnapshot(t *testing.T, repo string) string {
	t.Helper()
	return wikiGit(t, repo, "", "rev-parse", "HEAD") + "\n" + wikiGit(t, repo, "", "status", "--porcelain", "--ignored")
}

// runWiki runs one acta wiki command. It also holds the command to its promise
// to only read: the snapshot must be the same after it as before.
func runWiki(t *testing.T, repo string, args ...string) (code int, out, errOut string) {
	t.Helper()
	before := wikiSnapshot(t, repo)
	var o, e strings.Builder
	code = cmdWiki(args, &o, &e)
	if after := wikiSnapshot(t, repo); after != before {
		t.Errorf("acta wiki %v changed the repo\nbefore: %q\nafter:  %q", args, before, after)
	}
	return code, o.String(), e.String()
}

// wikiWants runs one command and checks its exit code and its output. Nothing
// may reach stderr.
func wikiWants(t *testing.T, repo string, args []string, wantCode int, wantOut string) {
	t.Helper()
	code, out, errOut := runWiki(t, repo, args...)
	if code != wantCode || out != wantOut || errOut != "" {
		t.Errorf("acta wiki %v: exit %d, stdout %q, stderr %q; want exit %d, stdout %q, stderr empty",
			args, code, out, errOut, wantCode, wantOut)
	}
}

// outLines splits what a command printed into its lines.
func outLines(out string) []string {
	if out == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(out, "\n"), "\n")
}

func TestWikiLs(t *testing.T) {
	repo := wikiRepo(t, wikiFiles())
	for _, c := range []struct {
		name string
		args []string
		want string
	}{
		{"every page, by name", []string{"ls"}, glossaryLine + lockLine + wrapLine},
		{"one type", []string{"ls", "--type", "Gotcha"}, wrapLine},
		{"the flag written with an equals sign", []string{"ls", "--type=Decision"}, lockLine},
		{"a type that has no page", []string{"ls", "--type", "Runbook"}, ""},
	} {
		t.Run(c.name, func(t *testing.T) { wikiWants(t, repo, c.args, exitOK, c.want) })
	}
}

// A description with a line break must still be one line, since a line is a page.
func TestWikiLsKeepsOneLinePerPage(t *testing.T) {
	files := wikiFiles()
	files[".acta/wiki/multi.md"] = "---\ntype: Gotcha\ntitle: A page\ndescription: |\n  first part\n  second part\n" +
		"paths: []\ntimestamp: 2026-01-01T00:00:00Z\n---\nbody\n"
	repo := wikiRepo(t, files)
	want := glossaryLine + ".acta/wiki/multi.md: first part second part\n" + lockLine + wrapLine
	wikiWants(t, repo, []string{"ls"}, exitOK, want)
}

func TestWikiMatch(t *testing.T) {
	repo := wikiRepo(t, wikiFiles())
	for _, c := range []struct {
		name string
		args []string
		want string
	}{
		{"one file", []string{"match", "internal/tui/model.go"}, wrapLine},
		{"a file written the way a shell word is", []string{"match", "./internal/tui/model.go"}, wrapLine},
		{"two pages, in page order", []string{"match", "scripts/test", "internal/tui/model.go"}, lockLine + wrapLine},
		{"the same two files in the other order", []string{"match", "internal/tui/model.go", "scripts/test"}, lockLine + wrapLine},
		{"two files under one page print it once", []string{"match", "internal/tui/model.go", "internal/tui/view.go"}, wrapLine},
		{"a file no page covers", []string{"match", "docs/none.md"}, ""},
		{"one file covered and one not", []string{"match", "docs/none.md", "scripts/test"}, lockLine},
	} {
		t.Run(c.name, func(t *testing.T) { wikiWants(t, repo, c.args, exitOK, c.want) })
	}
}

func TestWikiCheckCleanWiki(t *testing.T) {
	repo := wikiRepo(t, wikiFiles())
	wikiWants(t, repo, []string{"check"}, exitOK, "")
}

// A problem prints as its page and its message, and any problem makes exit 1.
func TestWikiCheckPrintsProblems(t *testing.T) {
	files := wikiFiles()
	files[".acta/wiki/bad.md"] = "---\ntype: Gotcha\ntitle: A page\npaths: []\ntimestamp: 2026-01-01T00:00:00Z\n---\nbody\n"
	repo := wikiRepo(t, files)
	wikiWants(t, repo, []string{"check"}, exitBadInput, ".acta/wiki/bad.md: description is missing\n")
}

// A page that cannot load is a problem too: it is printed on one line with the
// file named, it makes exit 1, and ls and match leave it out.
func TestWikiLoadErrors(t *testing.T) {
	files := wikiFiles()
	files[".acta/wiki/broken.md"] = "no frontmatter here\n"
	// paths is a text here, not a list. The YAML error for it runs over two lines.
	files[".acta/wiki/not-a-list.md"] = "---\ntype: Gotcha\ntitle: A page\ndescription: x\npaths: internal/tui/\n" +
		"timestamp: 2026-01-01T00:00:00Z\n---\nbody\n"
	repo := wikiRepo(t, files)

	code, out, errOut := runWiki(t, repo, "check")
	lines := outLines(out)
	if code != exitBadInput || errOut != "" || len(lines) != 2 {
		t.Fatalf("check: exit %d, stdout %q, stderr %q; want exit %d and two lines", code, out, errOut, exitBadInput)
	}
	// The page is named from the repo root, the way a problem names it, so an
	// agent can open it from any checkout.
	if want := ".acta/wiki/broken.md: no frontmatter between two --- lines"; lines[0] != want {
		t.Errorf("first line %q, want %q", lines[0], want)
	}
	if want := ".acta/wiki/not-a-list.md: "; !strings.HasPrefix(lines[1], want) {
		t.Errorf("second line %q, want it to start with %q", lines[1], want)
	}
	if strings.Contains(out, repo) {
		t.Errorf("check names the full path of the checkout: %q", out)
	}

	wikiWants(t, repo, []string{"ls"}, exitOK, glossaryLine+lockLine+wrapLine)
	wikiWants(t, repo, []string{"match", "internal/tui/model.go"}, exitOK, wrapLine)
}

// Every way a page can fail to load names the page from the repo root: a file
// that cannot be read, and a folder or a wiki folder that cannot be walked. The
// files in a subfolder too.
func TestWikiLoadErrorsNameThePageFromTheRepoRoot(t *testing.T) {
	files := wikiFiles()
	files[".acta/wiki/sub/deep/broken.md"] = "no frontmatter here\n"
	// paths is a text here, not a list, so the frontmatter does not parse.
	files[".acta/wiki/bad-paths.md"] = "---\ntype: Gotcha\ntitle: A page\ndescription: x\npaths: internal/tui/\n---\nbody\n"
	repo := wikiRepo(t, files)
	// A link that points nowhere cannot be read.
	if err := os.Symlink("nowhere", filepath.Join(repo, ".acta", "wiki", "dead.md")); err != nil {
		t.Fatal(err)
	}

	code, out, errOut := runWiki(t, repo, "check")
	want := []string{
		".acta/wiki/bad-paths.md: ",
		".acta/wiki/dead.md: ",
		".acta/wiki/sub/deep/broken.md: no frontmatter between two --- lines",
	}
	lines := outLines(out)
	if code != exitBadInput || errOut != "" || len(lines) != len(want) {
		t.Fatalf("check: exit %d, stdout %q, stderr %q; want exit %d and %d lines", code, out, errOut, exitBadInput, len(want))
	}
	for i, prefix := range want {
		if !strings.HasPrefix(lines[i], prefix) {
			t.Errorf("line %d is %q, want it to start with %q", i, lines[i], prefix)
		}
	}
	if strings.Contains(out, repo) {
		t.Errorf("check names the full path of the checkout: %q", out)
	}

	// A wiki that is a file, not a folder, cannot be walked. The line names the wiki.
	fileWiki := wikiRepo(t, map[string]string{".acta/wiki": "not a folder\n"})
	wikiWants(t, fileWiki, []string{"check"}, exitBadInput, ".acta/wiki: not a directory\n")
}

// The land gate: a branch that changes a covered file leaves its page stale, the
// range limits the check to the pages the branch touched, and a new timestamp
// makes the check pass.
func TestWikiCheckRange(t *testing.T) {
	files := wikiFiles()
	// This page is wrong, but nothing in the branch comes near it.
	files[".acta/wiki/bad.md"] = "---\ntype: Gotcha\ntitle: A page\npaths: [docs/guide.md]\ntimestamp: 2026-01-01T00:00:00Z\n---\nbody\n"
	files["docs/guide.md"] = "guide\n"
	repo := wikiRepo(t, files)
	wikiGit(t, repo, "", "switch", "-q", "-c", "feature")
	wikiWrite(t, repo, map[string]string{"internal/tui/model.go": "package tui // changed\n"})
	wikiGit(t, repo, wikiLater, "add", ".")
	wikiGit(t, repo, wikiLater, "commit", "-q", "-m", "change the model")
	const bad = ".acta/wiki/bad.md: description is missing"
	const stale = ".acta/wiki/tui-wrap.md: stale: "

	code, out, errOut := runWiki(t, repo, "check")
	if lines := outLines(out); code != exitBadInput || errOut != "" || len(lines) != 2 || lines[0] != bad || !strings.HasPrefix(lines[1], stale) {
		t.Errorf("check with no range: exit %d, stdout %q, stderr %q; want the bad page and the stale page", code, out, errOut)
	}

	code, out, errOut = runWiki(t, repo, "check", "main..HEAD")
	if lines := outLines(out); code != exitBadInput || errOut != "" || len(lines) != 1 || !strings.HasPrefix(lines[0], stale) {
		t.Errorf("check main..HEAD: exit %d, stdout %q, stderr %q; want the stale page only", code, out, errOut)
	}

	wikiWrite(t, repo, map[string]string{
		".acta/wiki/tui-wrap.md": strings.Replace(files[".acta/wiki/tui-wrap.md"], "2026-01-01", "2031-01-01", 1),
	})
	wikiGit(t, repo, wikiLater, "add", ".acta/wiki/tui-wrap.md")
	wikiGit(t, repo, wikiLater, "commit", "-q", "-m", "bump the timestamp")
	wikiWants(t, repo, []string{"check", "main..HEAD"}, exitOK, "")

	// A range git cannot read is a problem, not a pass. It names no page.
	code, out, errOut = runWiki(t, repo, "check", "nope..HEAD")
	if lines := outLines(out); code != exitBadInput || errOut != "" || len(lines) != 1 || !strings.HasPrefix(lines[0], `cannot read the range "nope..HEAD"`) {
		t.Errorf("check nope..HEAD: exit %d, stdout %q, stderr %q; want one line about the range", code, out, errOut)
	}
}

// With no wiki folder there are no pages, so every command prints nothing and
// exits 0, even outside a git repo.
func TestWikiWithoutAWikiFolder(t *testing.T) {
	t.Run("a repo with no wiki folder", func(t *testing.T) {
		repo := wikiRepo(t, map[string]string{"main.go": "package main\n"})
		for _, args := range [][]string{{"ls"}, {"ls", "--type", "Gotcha"}, {"match", "main.go"}, {"check"}, {"check", "main..HEAD"}} {
			wikiWants(t, repo, args, exitOK, "")
		}
	})
	t.Run("a folder that is not a repo", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("ACTA_ROOT", "")
		t.Setenv("PM_ROOT", "")
		t.Chdir(t.TempDir())
		for _, args := range [][]string{{"ls"}, {"match", "main.go"}, {"check"}, {"check", "main..HEAD"}} {
			var out, errOut strings.Builder
			if code := cmdWiki(args, &out, &errOut); code != exitOK || out.String() != "" || errOut.String() != "" {
				t.Errorf("acta wiki %v: exit %d, stdout %q, stderr %q; want exit 0 and nothing printed", args, code, out.String(), errOut.String())
			}
		}
	})
}

// Bad arguments print the usage on stderr, nothing on stdout, and exit with the
// usage code. The repo has pages, so an argument taken by mistake would print.
func TestWikiBadArguments(t *testing.T) {
	repo := wikiRepo(t, wikiFiles())
	for _, c := range []struct {
		name string
		args []string
	}{
		{"no command", nil},
		{"an unknown command", []string{"frobnicate"}},
		{"ls with a word", []string{"ls", "extra"}},
		{"ls with a flag it does not have", []string{"ls", "--bogus"}},
		{"ls --type with no type", []string{"ls", "--type"}},
		{"match with no file", []string{"match"}},
		{"match with --type", []string{"match", "--type", "Gotcha", "scripts/test"}},
		{"check with two ranges", []string{"check", "main..HEAD", "main..HEAD"}},
		{"check with --type", []string{"check", "--type", "Gotcha"}},
		{"ok with no id", []string{"ok"}},
		{"ok with --type", []string{"ok", "--type", "Gotcha", "glossary"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			code, out, errOut := runWiki(t, repo, c.args...)
			if code != exitBadInput || out != "" || !strings.Contains(errOut, "usage: acta wiki") {
				t.Errorf("acta wiki %v: exit %d, stdout %q, stderr %q; want exit %d, no stdout and the usage on stderr",
					c.args, code, out, errOut, exitBadInput)
			}
		})
	}
}

// The real command line reaches the wiki commands, and its list of commands
// names them.
func TestRunKnowsTheWikiCommands(t *testing.T) {
	wikiRepo(t, wikiFiles())
	var out, errOut strings.Builder
	if code := Run([]string{"wiki", "ls", "--type", "Gotcha"}, strings.NewReader(""), false, &out, &errOut); code != exitOK || out.String() != wrapLine {
		t.Errorf("acta wiki ls --type Gotcha: exit %d, stdout %q, stderr %q", code, out.String(), errOut.String())
	}
	errOut.Reset()
	if code := Run([]string{"nonsense"}, strings.NewReader(""), false, &out, &errOut); code != exitBadInput || !strings.Contains(errOut.String(), "wiki ls, wiki match, wiki check, wiki ok") {
		t.Errorf("an unknown command: exit %d, stderr %q; want it to list wiki ls, wiki match, wiki check, wiki ok", code, errOut.String())
	}
}

// okRepo is a repo with two pages, a and sub/b, so ids with a slash get tested.
func okRepo(t *testing.T) string {
	t.Helper()
	return wikiRepo(t, map[string]string{
		".acta/wiki/a.md":     wikiPage("Gotcha", "Page a", ""),
		".acta/wiki/sub/b.md": wikiPage("Gotcha", "Page b", ""),
	})
}

// runOk runs acta wiki ok and gives back the exit code, stdout and stderr.
func runOk(t *testing.T, ids ...string) (code int, out, errOut string) {
	t.Helper()
	var o, e strings.Builder
	code = cmdWiki(append([]string{"ok"}, ids...), &o, &e)
	return code, o.String(), e.String()
}

// okFails runs acta wiki ok and holds it to the failure promise: a non-zero
// exit, a reason on stderr, nothing on stdout, and HEAD where it was.
func okFails(t *testing.T, repo string, wantCode int, ids ...string) string {
	t.Helper()
	head := wikiGit(t, repo, "", "rev-parse", "HEAD")
	code, out, errOut := runOk(t, ids...)
	if code != wantCode || out != "" || strings.TrimSpace(errOut) == "" {
		t.Errorf("acta wiki ok %v: exit %d, stdout %q, stderr %q; want exit %d, no stdout and a reason on stderr",
			ids, code, out, errOut, wantCode)
	}
	if after := wikiGit(t, repo, "", "rev-parse", "HEAD"); after != head {
		t.Errorf("acta wiki ok %v moved HEAD from %s to %s", ids, head, after)
	}
	return errOut
}

// A good re-check is one commit with the same files as its parent, a subject
// that lists the ids in the order given, and one trailer per id.
func TestWikiOkRecordsACheckWithoutChangingAFile(t *testing.T) {
	repo := okRepo(t)
	parent := wikiGit(t, repo, "", "rev-parse", "HEAD")
	code, out, errOut := runOk(t, "a", "sub/b")
	if code != exitOK || errOut != "" {
		t.Fatalf("acta wiki ok a sub/b: exit %d, stderr %q", code, errOut)
	}
	if want := wikiGit(t, repo, "", "rev-parse", "--short", "HEAD") + "\n"; out != want {
		t.Errorf("stdout %q; want the short sha %q", out, want)
	}
	if wikiGit(t, repo, "", "rev-parse", "HEAD^") != parent {
		t.Error("the check is not a single commit on top of the old HEAD")
	}
	if wikiGit(t, repo, "", "rev-parse", "HEAD^{tree}") != wikiGit(t, repo, "", "rev-parse", parent+"^{tree}") {
		t.Error("the check commit changed a file")
	}
	want := "chore(wiki): re-check a, sub/b\n\nWiki-Checked: a\nWiki-Checked: sub/b"
	if msg := wikiGit(t, repo, "", "log", "-1", "--format=%B"); msg != want {
		t.Errorf("message %q; want %q", msg, want)
	}
	if st := wikiGit(t, repo, "", "status", "--porcelain"); st != "" {
		t.Errorf("the repo is dirty after the check: %q", st)
	}
}

// One bad id among good ones refuses the whole run, and names the bad one.
func TestWikiOkRefusesAnUnknownId(t *testing.T) {
	repo := okRepo(t)
	if errOut := okFails(t, repo, exitBadInput, "a", "nope"); !strings.Contains(errOut, "nope") {
		t.Errorf("stderr %q does not name the unknown id", errOut)
	}
}

// With no ids there is nothing to record, so it is a usage error.
func TestWikiOkWithNoIdsIsAUsageError(t *testing.T) {
	repo := okRepo(t)
	if errOut := okFails(t, repo, exitBadInput); !strings.Contains(errOut, "usage: acta wiki") {
		t.Errorf("stderr %q is not the usage", errOut)
	}
}

// The same id twice is one id, in the place it came first.
func TestWikiOkDropsDuplicateIds(t *testing.T) {
	repo := okRepo(t)
	if code, _, errOut := runOk(t, "a", "sub/b", "a"); code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, errOut)
	}
	want := "chore(wiki): re-check a, sub/b\n\nWiki-Checked: a\nWiki-Checked: sub/b"
	if msg := wikiGit(t, repo, "", "log", "-1", "--format=%B"); msg != want {
		t.Errorf("message %q; want %q", msg, want)
	}
}

// A file the user staged is not part of the check commit and is still staged.
func TestWikiOkLeavesStagedFilesAlone(t *testing.T) {
	repo := okRepo(t)
	wikiWrite(t, repo, map[string]string{"notes.txt": "half done\n"})
	wikiGit(t, repo, "", "add", "notes.txt")
	parent := wikiGit(t, repo, "", "rev-parse", "HEAD")
	if code, _, errOut := runOk(t, "a"); code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, errOut)
	}
	if wikiGit(t, repo, "", "rev-parse", "HEAD^{tree}") != wikiGit(t, repo, "", "rev-parse", parent+"^{tree}") {
		t.Error("the staged file went into the check commit")
	}
	if st := wikiGit(t, repo, "", "status", "--porcelain"); st != "A  notes.txt" {
		t.Errorf("status %q; want the file still staged", st)
	}
}

// Outside a git repo the command stops with a reason and makes nothing.
func TestWikiOkOutsideAGitRepo(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ACTA_ROOT", "")
	t.Setenv("PM_ROOT", "")
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	wikiWrite(t, dir, map[string]string{".acta/wiki/a.md": wikiPage("Gotcha", "Page a", "")})
	t.Chdir(dir)
	code, out, errOut := runOk(t, "a")
	if code == exitOK || out != "" || strings.TrimSpace(errOut) == "" {
		t.Errorf("exit %d, stdout %q, stderr %q; want a failure with a reason", code, out, errOut)
	}
}

// When git refuses the commit, the command fails with git's reason and HEAD stays.
func TestWikiOkWhenGitRefusesTheCommit(t *testing.T) {
	repo := okRepo(t)
	hook := filepath.Join(repo, ".git", "hooks", "commit-msg")
	if err := os.WriteFile(hook, []byte("#!/bin/sh\necho no commits today >&2\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if errOut := okFails(t, repo, exitOther, "a"); !strings.Contains(errOut, "no commits today") {
		t.Errorf("stderr %q does not carry git's reason", errOut)
	}
}

// Two branches that both re-checked the same page merge with no conflict,
// because the check touches no file.
func TestWikiOkBranchesMergeCleanly(t *testing.T) {
	repo := okRepo(t)
	wikiGit(t, repo, "", "checkout", "-q", "-b", "one")
	if code, _, errOut := runOk(t, "a"); code != exitOK {
		t.Fatalf("branch one: exit %d, stderr %q", code, errOut)
	}
	wikiGit(t, repo, "", "checkout", "-q", "main")
	wikiGit(t, repo, "", "checkout", "-q", "-b", "two")
	if code, _, errOut := runOk(t, "a"); code != exitOK {
		t.Fatalf("branch two: exit %d, stderr %q", code, errOut)
	}
	wikiGit(t, repo, "", "checkout", "-q", "main")
	wikiGit(t, repo, "", "merge", "--no-ff", "-q", "-m", "merge one", "one")
	wikiGit(t, repo, "", "merge", "--no-ff", "-q", "-m", "merge two", "two")
}
