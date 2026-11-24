package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/setup"
)

// runInit runs acta init inside dir and gives back its exit code and output.
func runInit(t *testing.T, dir string) (int, string, string) {
	t.Helper()
	var stdout, stderr strings.Builder
	code := -1
	inDir(t, dir, func() {
		code = Run([]string{"init"}, strings.NewReader(""), false, &stdout, &stderr)
	})
	return code, stdout.String(), stderr.String()
}

// commitFile writes one file and commits it as the user would.
func commitFile(t *testing.T, dir, name, body, msg string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOut(t, dir, "add", name)
	gitOut(t, dir, "commit", "-q", "-m", msg)
}

// A fresh repo gets the .acta folder and a new CLAUDE.md in one commit. The
// temp home has no plugin installed, so the doctor report fails, and init
// still exits 0 because its own work went fine.
func TestInitFreshRepo(t *testing.T) {
	doctorHome(t)
	dir := doctorRepo(t)
	before := gitOut(t, dir, "rev-parse", "HEAD")
	code, out, errOut := runInit(t, dir)
	if code != 0 {
		t.Fatalf("exit %d want 0, stdout %q stderr %q", code, out, errOut)
	}
	if !strings.Contains(out, "fail ") {
		t.Fatalf("stdout %q has no failing doctor line; the exit-0 case is not tested", out)
	}
	if got, _ := readOrMissing(t, filepath.Join(dir, "CLAUDE.md")); got != setup.Block {
		t.Fatalf("CLAUDE.md = %q, want only the block", got)
	}
	if _, ok := readOrMissing(t, filepath.Join(dir, ".acta", ".gitignore")); !ok {
		t.Fatal(".acta/.gitignore was not made")
	}
	if got := gitOut(t, dir, "rev-parse", "HEAD~1"); got != before {
		t.Fatalf("HEAD~1 = %s, want the old HEAD %s: init made more or less than one commit", got, before)
	}
	if got := gitOut(t, dir, "log", "-1", "--format=%s"); got != "chore: acta init" {
		t.Fatalf("subject %q", got)
	}
	if st := gitOut(t, dir, "status", "--porcelain"); st != "" {
		t.Fatalf("init left files out of its commit: %q", st)
	}
}

// Both files get the block and keep their own text. A second run finds the
// same bytes, makes no commit and says the repo is already set up.
func TestInitBothFilesThenRerun(t *testing.T) {
	doctorHome(t)
	dir := doctorRepo(t)
	commitFile(t, dir, "CLAUDE.md", "claude notes\n", "docs: claude")
	commitFile(t, dir, "AGENTS.md", "agent notes\n", "docs: agents")
	if code, out, errOut := runInit(t, dir); code != 0 {
		t.Fatalf("exit %d, stdout %q stderr %q", code, out, errOut)
	}
	for name, own := range map[string]string{"CLAUDE.md": "claude notes\n", "AGENTS.md": "agent notes\n"} {
		got, _ := readOrMissing(t, filepath.Join(dir, name))
		if !strings.HasPrefix(got, own) || !strings.Contains(got, setup.Block) {
			t.Fatalf("%s = %q, want its own text then the block", name, got)
		}
	}
	head := gitOut(t, dir, "rev-parse", "HEAD")
	claude, _ := readOrMissing(t, filepath.Join(dir, "CLAUDE.md"))
	code, out, errOut := runInit(t, dir)
	if code != 0 {
		t.Fatalf("rerun exit %d, stdout %q stderr %q", code, out, errOut)
	}
	if !strings.Contains(out, "already set up") {
		t.Fatalf("rerun stdout %q does not say already set up", out)
	}
	if got := gitOut(t, dir, "rev-parse", "HEAD"); got != head {
		t.Fatalf("rerun moved HEAD from %s to %s", head, got)
	}
	if got, _ := readOrMissing(t, filepath.Join(dir, "CLAUDE.md")); got != claude {
		t.Fatalf("rerun changed CLAUDE.md: %q", got)
	}
}

// An old block is refreshed and committed in a new commit, even when HEAD
// only added CLAUDE.md, the kind of commit a fold would rewrite.
func TestInitRefreshesOldBlockWithoutFold(t *testing.T) {
	doctorHome(t)
	dir := doctorRepo(t)
	if code, out, errOut := runInit(t, dir); code != 0 {
		t.Fatalf("first exit %d, stdout %q stderr %q", code, out, errOut)
	}
	gitOut(t, dir, "rm", "-q", "CLAUDE.md")
	gitOut(t, dir, "commit", "-q", "-m", "docs: drop claude md")
	commitFile(t, dir, "CLAUDE.md", "<!-- acta:begin -->\nold block\n<!-- acta:end -->\n", "docs: claude md")
	user := gitOut(t, dir, "rev-parse", "HEAD")
	if code, out, errOut := runInit(t, dir); code != 0 {
		t.Fatalf("exit %d, stdout %q stderr %q", code, out, errOut)
	}
	if got, _ := readOrMissing(t, filepath.Join(dir, "CLAUDE.md")); got != setup.Block {
		t.Fatalf("CLAUDE.md = %q, want the new block", got)
	}
	if got := gitOut(t, dir, "rev-parse", "HEAD~1"); got != user {
		t.Fatalf("HEAD~1 = %s, want the user's commit %s untouched", got, user)
	}
	if got := gitOut(t, dir, "log", "-1", "--format=%s"); got != "chore: acta init" {
		t.Fatalf("subject %q", got)
	}
}

// A file the user had dirty is still written, but nothing is committed.
func TestInitDirtyFileWritesWithoutCommit(t *testing.T) {
	doctorHome(t)
	dir := doctorRepo(t)
	commitFile(t, dir, "CLAUDE.md", "claude notes\n", "docs: claude")
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("claude notes, edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	head := gitOut(t, dir, "rev-parse", "HEAD")
	code, out, errOut := runInit(t, dir)
	if code != 0 {
		t.Fatalf("exit %d, stdout %q stderr %q", code, out, errOut)
	}
	if !strings.Contains(out+errOut, "written, not committed:") {
		t.Fatalf("output %q %q does not say it skipped the commit", out, errOut)
	}
	if got := gitOut(t, dir, "rev-parse", "HEAD"); got != head {
		t.Fatalf("init committed over a dirty file: HEAD %s, want %s", got, head)
	}
	got, _ := readOrMissing(t, filepath.Join(dir, "CLAUDE.md"))
	if !strings.HasPrefix(got, "claude notes, edited\n") || !strings.Contains(got, setup.Block) {
		t.Fatalf("CLAUDE.md = %q, want the user's edit then the block", got)
	}
}

// A dirty file that doctor --fix owns skips the commit too, even when the
// block itself has a change to commit.
func TestInitDirtyDoctorFileWritesWithoutCommit(t *testing.T) {
	doctorHome(t)
	dir := doctorRepo(t)
	if code, out, errOut := runInit(t, dir); code != 0 {
		t.Fatalf("first exit %d, stdout %q stderr %q", code, out, errOut)
	}
	commitFile(t, dir, "CLAUDE.md", "<!-- acta:begin -->\nold block\n<!-- acta:end -->\n", "docs: old block")
	gi := filepath.Join(dir, ".acta", ".gitignore")
	old, _ := readOrMissing(t, gi)
	if err := os.WriteFile(gi, []byte(old+"mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	head := gitOut(t, dir, "rev-parse", "HEAD")
	code, out, errOut := runInit(t, dir)
	if code != 0 {
		t.Fatalf("exit %d, stdout %q stderr %q", code, out, errOut)
	}
	if !strings.Contains(out+errOut, "written, not committed:") {
		t.Fatalf("output %q %q does not say it skipped the commit", out, errOut)
	}
	if got := gitOut(t, dir, "rev-parse", "HEAD"); got != head {
		t.Fatalf("init committed over a dirty .gitignore: HEAD %s, want %s", got, head)
	}
	if got, _ := readOrMissing(t, filepath.Join(dir, "CLAUDE.md")); got != setup.Block {
		t.Fatalf("CLAUDE.md = %q, want the new block written", got)
	}
}

// A broken .acta.yaml inside git names the config, not a missing git repo.
func TestInitBrokenConfigNamesIt(t *testing.T) {
	doctorHome(t)
	dir := doctorRepo(t)
	if err := os.WriteFile(filepath.Join(dir, ".acta.yaml"), []byte("root: [oops\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, errOut := runInit(t, dir)
	if code == 0 || strings.Contains(errOut, "git init") {
		t.Fatalf("exit %d, stderr %q: want a non-zero exit that names the config", code, errOut)
	}
}

// A broken marker pair stops init before it writes anything at all.
func TestInitBrokenMarkerWritesNothing(t *testing.T) {
	doctorHome(t)
	dir := doctorRepo(t)
	commitFile(t, dir, "CLAUDE.md", "notes\n", "docs: claude")
	commitFile(t, dir, "AGENTS.md", "<!-- acta:begin -->\nhalf a pair\n", "docs: agents")
	code, out, errOut := runInit(t, dir)
	if code == 0 {
		t.Fatalf("exit 0 on a broken marker, stdout %q", out)
	}
	if !strings.Contains(errOut, "AGENTS.md") {
		t.Fatalf("stderr %q does not name the broken file", errOut)
	}
	if _, ok := readOrMissing(t, filepath.Join(dir, ".acta")); ok {
		t.Fatal(".acta was made before the marker check")
	}
	if got, _ := readOrMissing(t, filepath.Join(dir, "CLAUDE.md")); got != "notes\n" {
		t.Fatalf("CLAUDE.md changed to %q", got)
	}
}

// Outside git init writes nothing, since every acta write commits.
func TestInitOutsideGit(t *testing.T) {
	doctorHome(t)
	dir := t.TempDir()
	code, out, errOut := runInit(t, dir)
	if code == 0 {
		t.Fatalf("exit 0 outside git, stdout %q", out)
	}
	if !strings.Contains(errOut, "git") {
		t.Fatalf("stderr %q does not say git is needed", errOut)
	}
	for _, name := range []string{".acta", "CLAUDE.md"} {
		if _, ok := readOrMissing(t, filepath.Join(dir, name)); ok {
			t.Fatalf("%s was written outside git", name)
		}
	}
}
