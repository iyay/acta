package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var actaBin string
var pmbBin string

func TestMain(m *testing.M) {
	// An inherited PM_ROOT would point every test at the wrong board. Clear
	// it once here, so no helper has to set env and block parallel tests.
	os.Unsetenv("PM_ROOT")
	dir, err := os.MkdirTemp("", "acta-bin")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	actaBin = filepath.Join(dir, "acta")
	if out, err := exec.Command("go", "build", "-o", actaBin, ".").CombinedOutput(); err != nil {
		fmt.Println(string(out))
		os.Exit(1)
	}
	pmbBin = filepath.Join(dir, "pmb")
	if out, err := exec.Command("go", "build", "-o", pmbBin, "../pmb").CombinedOutput(); err != nil {
		fmt.Println(string(out))
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// fixtureRepo copies the board fixture into a fresh git repo.
func fixtureRepo(t *testing.T) string {
	t.Helper()
	dst, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join("..", "..", "internal", "board", "testdata", "basic")
	err = filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dst, rel), b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	// Name the author inside the repo, not in the env, so tests that make
	// commits can still run side by side.
	for _, args := range [][]string{{"init", "-q", "-b", "main"}, {"config", "user.name", "test"}, {"config", "user.email", "test@example.com"}, {"add", "."}, {"commit", "-q", "-m", "init"}} {
		if out, err := exec.Command("git", append([]string{"-C", dst}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	return dst
}

func acta(t *testing.T, dir, stdin string, args ...string) (string, string, int) {
	t.Helper()
	return runBin(t, actaBin, dir, stdin, args...)
}

func pmb(t *testing.T, dir, stdin string, args ...string) (string, string, int) {
	t.Helper()
	return runBin(t, pmbBin, dir, stdin, args...)
}

func runBin(t *testing.T, bin, dir, stdin string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	// The child must never put a lock in the real cache folder of whoever
	// runs the suite, so it gets a home of its own. macOS reads HOME for the
	// cache folder, Linux reads XDG_CACHE_HOME, so we move both.
	home := t.TempDir()
	cmd.Env = append(os.Environ(), "HOME="+home, "XDG_CACHE_HOME="+home)
	cmd.Stdin = strings.NewReader(stdin)
	var out, errOut strings.Builder
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return out.String(), errOut.String(), code
}

type jsonItem struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Title        string   `json:"title"`
	Status       string   `json:"status"`
	StatusSource string   `json:"status_source"`
	Ref          string   `json:"ref"`
	Parent       string   `json:"parent"`
	Children     []string `json:"children"`
	Progress     struct {
		Done  int `json:"done"`
		Total int `json:"total"`
	} `json:"progress"`
	Path     string   `json:"path"`
	Legacy   bool     `json:"legacy"`
	Problems []string `json:"problems"`
}

func decode(t *testing.T, s string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(s), v); err != nil {
		t.Fatalf("bad json %v: %s", err, s)
	}
}

func TestListJSON(t *testing.T) {
	t.Parallel()

	dir := fixtureRepo(t)
	out, _, code := acta(t, dir, "", "list", "--type", "story", "--json")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	var items []jsonItem
	decode(t, out, &items)
	var got []string
	for _, it := range items {
		got = append(got, it.ID)
	}
	want := "specs/2026-09-28-from-scratch-design specs/2026-09-22-beta specs/2026-09-20-alpha specs/2026-09-18-weird specs/2026-09-17-broken"
	if strings.Join(got, " ") != want {
		t.Fatalf("ids %v", got)
	}
	if items[0].Children == nil || items[0].Problems == nil {
		t.Fatal("children and problems must be [] not null")
	}
}

// A plan is its own item, so a list line says so instead of calling it a story.
func TestListShowsPlansAsPlans(t *testing.T) {
	t.Parallel()

	dir := fixtureRepo(t)
	out, _, code := acta(t, dir, "", "list", "--all")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	var found int
	for _, line := range lines {
		fields := strings.Fields(line)
		// A plan item's ID is the plan path; a task's ID adds "#task-n".
		if len(fields) < 4 || !strings.HasPrefix(fields[3], "plans/") || strings.Contains(fields[3], "#") {
			continue
		}
		found++
		if fields[1] != "plan" {
			t.Errorf("plan line has no plan type: %q", line)
		}
	}
	if found == 0 {
		t.Fatalf("no plan in the list: %s", out)
	}
}

func TestListAllIncludesLegacyTasks(t *testing.T) {
	t.Parallel()

	dir := fixtureRepo(t)
	out, _, _ := acta(t, dir, "", "list", "--type", "task", "--all", "--json")
	if !strings.Contains(out, `"docs/superpowers/plans/2026-01-02-old#task-1"`) {
		t.Fatalf("legacy task missing: %s", out)
	}
	out, _, _ = acta(t, dir, "", "list", "--type", "task", "--json")
	if strings.Contains(out, "2026-01-02-old") {
		t.Fatal("legacy task shown without --all")
	}
}

func TestShow(t *testing.T) {
	t.Parallel()

	dir := fixtureRepo(t)
	out, _, code := acta(t, dir, "", "show", "specs/2026-09-20-alpha", "--json")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	var it jsonItem
	decode(t, out, &it)
	if it.Status != "in-progress" || it.Progress.Done != 1 || it.Progress.Total != 2 || len(it.Children) != 2 ||
		it.Path != ".acta/specs/2026-09-20-alpha.md" || it.Ref != "A-1" || it.StatusSource != "derived" {
		t.Fatalf("got %+v", it)
	}
	if _, _, code := acta(t, dir, "", "show", "specs/nope"); code != 1 {
		t.Fatalf("unknown id exit %d, want 1", code)
	}
}

// A task under way reads in-progress on the show line and in the JSON list,
// and the old word never turns up in either.
func TestAStartedOrHalfTickedTaskReadsInProgress(t *testing.T) {
	t.Parallel()

	dir := fixtureRepo(t)
	out, errOut, code := acta(t, dir, "", "show", "plans/2026-09-21-alpha#task-2")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	if !strings.Contains(out, "status: in-progress (derived)") || strings.Contains(out, "doing") {
		t.Fatalf("show says %q, want the status in-progress", out)
	}
	list, _, code := acta(t, dir, "", "list", "--all", "--json")
	if code != 0 {
		t.Fatalf("list exit %d", code)
	}
	var items []jsonItem
	decode(t, list, &items)
	if len(items) == 0 {
		t.Fatal("the fixture board has items")
	}
	for _, it := range items {
		if it.Status == "doing" {
			t.Errorf("%s reads doing in the JSON list, want in-progress", it.ID)
		}
	}
}

func TestSet(t *testing.T) {
	t.Parallel()

	dir := fixtureRepo(t)
	if _, errOut, code := acta(t, dir, "", "set", "bugs/2026-09-26-open", "status", "fixing"); code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".acta/bugs/2026-09-26-open.md"))
	// Setting a working status writes the day the work began too. The date is
	// written as text, so the file holds it in quotes.
	want := "---\nstatus: fixing\nstarted: \"" + time.Now().Format("2006-01-02") + "\"\n---\n"
	if !strings.HasPrefix(string(b), want) {
		t.Fatalf("file %q, want it to start with %q", b, want)
	}
	for _, args := range [][]string{
		{"set", "docs/superpowers/specs/2026-01-01-old", "status", "done"},
		{"set", "bugs/2026-09-26-open", "status", "in-progress"},
		{"set", "bugs/2026-09-26-open", "status"},
		{"set", "--nope"},
		{"frobnicate"},
	} {
		if _, _, code := acta(t, dir, "", args...); code != 1 {
			t.Errorf("%v: exit %d, want 1", args, code)
		}
	}
}

func TestSetDirtyFileExits2(t *testing.T) {
	t.Parallel()

	dir := fixtureRepo(t)
	p := filepath.Join(dir, ".acta/bugs/2026-09-26-open.md")
	if err := os.WriteFile(p, []byte("# Button does nothing\n\n## Symptom\nEdited.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, errOut, code := acta(t, dir, "", "set", "bugs/2026-09-26-open", "status", "fixing")
	if code != 2 || !strings.Contains(errOut, "other uncommitted changes") {
		t.Fatalf("exit %d stderr %q", code, errOut)
	}
}

func TestBugNewFromStdin(t *testing.T) {
	t.Parallel()

	dir := fixtureRepo(t)
	out, errOut, code := acta(t, dir, "## Symptom\nTwo ACKs.\n", "bug", "new", "ack-dup", "--ref", "New-261")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	path := strings.TrimSpace(out)
	b, err := os.ReadFile(filepath.Join(dir, path))
	if err != nil || !strings.Contains(string(b), "ref: New-261") || !strings.Contains(string(b), "Two ACKs.") {
		t.Fatalf("path %q file %q err %v", path, b, err)
	}
	if _, _, code := acta(t, dir, "## Repro\nx\n", "bug", "new", "no-symptom"); code != 1 {
		t.Fatalf("no symptom exit %d, want 1", code)
	}
}

const pmbWarning = "pmb is now acta; this name goes away in a later version"

// TestPmbAliasWarns needs the thin pmb wrapper: warn on stderr first,
// then match the acta run on stdout and exit code.
func TestPmbAliasWarns(t *testing.T) {
	t.Parallel()

	dir := fixtureRepo(t)
	wantOut, _, wantCode := acta(t, dir, "", "list", "--all")
	gotOut, gotErr, gotCode := pmb(t, dir, "", "list", "--all")
	if !strings.HasPrefix(gotErr, pmbWarning+"\n") {
		t.Fatalf("stderr %q, want it to start with %q", gotErr, pmbWarning)
	}
	if gotOut != wantOut {
		t.Fatalf("stdout mismatch:\nacta: %q\npmb: %q", wantOut, gotOut)
	}
	if gotCode != wantCode {
		t.Fatalf("exit %d, want %d", gotCode, wantCode)
	}
}

// TestUsageSaysActa needs every help line to name acta, never pmb.
func TestUsageSaysActa(t *testing.T) {
	t.Parallel()

	dir := fixtureRepo(t)
	_, noArgsErr, _ := acta(t, dir, "", "-h")
	if !strings.Contains(noArgsErr, "acta") || strings.Contains(noArgsErr, "pmb") {
		t.Fatalf("acta -h stderr %q, want acta without pmb", noArgsErr)
	}
	_, tickErr, tickCode := acta(t, dir, "", "tick", "-h")
	if tickCode != 0 {
		t.Fatalf("acta tick -h exit %d", tickCode)
	}
	if !strings.Contains(tickErr, "acta") || strings.Contains(tickErr, "pmb") {
		t.Fatalf("acta tick -h stderr %q, want acta without pmb", tickErr)
	}
}
