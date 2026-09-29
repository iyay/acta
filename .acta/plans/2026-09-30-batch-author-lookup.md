---
created: "2026-09-30"
id: PLN-0040
hash: bn76r8e
---
# Batch Author Lookup Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Every board load asks git for authors once per folder, not once per file, so `acta show` drops from about 2.9 s to well under half a second with the same author names.

**Architecture:** `gitc.Author(repo, path)` is replaced by `gitc.Authors(repo, paths)`, which runs one `git log --diff-filter=A --name-only` over all the paths and maps each one to the name on its oldest add. `board.fillAuthors` groups the items on disk by folder, calls `Authors` once per folder, and falls back to `gitUserName(root)` (asked at most once) for any path git did not name.

**Tech Stack:** Go, the git binary through `gitc.run`.

**Spec:** `.acta/specs/2026-09-30-batch-author-lookup-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Author names and fallbacks stay exactly as they are now. Only the number of git calls changes.
- No `acta path` command. No change to `trees`, `show` output or the TUI.
- No new dependency in go.mod.
- Comments use plain English that a 10-year-old can read back: short words, and they say why.
- Before the commit, run `gofmt -l internal` (it must print nothing), `go vet ./...` and `go test ./...`.

## File Map

- `internal/gitc/gitc.go`: `Authors` replaces `Author`.
- `internal/gitc/gitc_test.go`: tests for `Authors`; `TestAuthorIsTheFirstCommit` goes, its cases move to the new tests.
- `internal/board/closed.go`: the `gitAuthors` test variable and the new `fillAuthors`.
- `internal/board/board_test.go`: `TestAuthorIsAskedOncePerFile` becomes `TestAuthorIsAskedOncePerFolder`; one new real-repo test for a folder that mixes committed and new files.

## Waves

- Wave 1: Task 1.

---

### Task 1: One git call per folder for authors

**Files:**
- Modify: `internal/gitc/gitc.go` (`Author`, near line 97)
- Modify: `internal/board/closed.go` (whole file)
- Test: `internal/gitc/gitc_test.go` (`TestAuthorIsTheFirstCommit`, near line 220)
- Test: `internal/board/board_test.go` (`TestAuthorIsAskedOncePerFile`, near line 531)

**verify:** For every item on disk, `Author` is the name the old one-file lookup gave, and git runs at most once per folder in one load, never once per file. List every case checked: first add wins over a later edit by someone else; a file moved inside the folder takes the name of whoever moved it; a file name with non-ASCII letters; a file never committed gets `user.name`, read once per load; a repo with no commit; a folder outside git gives empty with no git run; an empty path list gives an empty map, never the whole history; a plan with many tasks is one path, not many.

**Interfaces:**
- Consumes: `gitc.run(repo string, args ...string) (string, error)`, `gitc.inRepo(dir string) bool`, `gitc.UserName(repo string) string`, all already in `internal/gitc/gitc.go`.
- Produces: `func Authors(repo string, paths []string) map[string]string` in package `gitc`. Keys are `filepath.Join(repo, <name git printed>)`, so a path that sits right inside `repo` comes back as the same string. `gitc.Author` is removed.

- [ ] **Step 1: Write the failing tests in `internal/gitc/gitc_test.go`**

Delete `TestAuthorIsTheFirstCommit` and put these two tests in its place. Add `"reflect"` to the imports.

```go
// Authors asks git once for many files. The author of a file is the person
// whose commit first added it, so a later edit by someone else never changes
// who wrote it. A file moved inside the folder counts as added by whoever
// moved it, which is what a one-file log said too.
func TestAuthorsAsksOnceForManyFiles(t *testing.T) {
	repo := firstCommitRepo(t, "Ana")
	writeFile(t, filepath.Join(repo, "old.md"), "old\n")
	git(t, repo, "add", "old.md")
	git(t, repo, "commit", "-qm", "ana adds old")
	t.Setenv("GIT_AUTHOR_NAME", "Budi")
	t.Setenv("GIT_COMMITTER_NAME", "Budi")
	writeFile(t, filepath.Join(repo, "a.md"), "a again\n")
	writeFile(t, filepath.Join(repo, "catatan-é.md"), "c\n")
	git(t, repo, "mv", "old.md", "moved.md")
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-qm", "budi edits, adds and moves")

	at := func(name string) string { return filepath.Join(repo, name) }
	got := Authors(repo, []string{at("a.md"), at("catatan-é.md"), at("moved.md"), at("new.md")})
	want := map[string]string{at("a.md"): "Ana", at("catatan-é.md"): "Budi", at("moved.md"): "Budi"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Authors = %v, want %v (new.md was never committed, so it has no entry)", got, want)
	}
}

// Some asks have nothing to find. None of them may answer with names.
func TestAuthorsWithNothingToFind(t *testing.T) {
	repo := firstCommitRepo(t, "Ana")
	// No paths at all must not turn into a log of the whole repo.
	if got := Authors(repo, nil); len(got) != 0 {
		t.Errorf("no paths gave %v, want an empty map", got)
	}
	// A repo with a .git but no commit has no author for anything.
	empty := t.TempDir()
	git(t, empty, "init", "-q", "-b", "main")
	if got := Authors(empty, []string{filepath.Join(empty, "a.md")}); len(got) != 0 {
		t.Errorf("a repo with no commit gave %v, want an empty map", got)
	}
	// A folder outside any checkout answers at once, without running git.
	if got := Authors(t.TempDir(), []string{filepath.Join(repo, "a.md")}); len(got) != 0 {
		t.Errorf("a folder with no repo gave %v, want an empty map", got)
	}
}
```

- [ ] **Step 2: Write the failing tests in `internal/board/board_test.go`**

Replace `TestAuthorIsAskedOncePerFile` with this test:

```go
// Git is asked once per folder, not once per file or per item: two specs, a
// plan with three tasks and a debt file with two lines sit in three folders,
// so git gets three questions, and no file is named twice in one question.
func TestAuthorIsAskedOncePerFolder(t *testing.T) {
	var calls [][]string
	gitAuthors = func(_ string, paths []string) map[string]string {
		calls = append(calls, append([]string{}, paths...))
		return map[string]string{}
	}
	names := 0
	gitUserName = func(string) string {
		names++
		return "Sari"
	}
	t.Cleanup(func() { gitAuthors, gitUserName = gitc.Authors, gitc.UserName })

	b := boardWith(t, map[string]string{
		"specs/2026-09-20-a.md":    "---\nid: SPEC-1\n---\n# Spec A\n",
		"specs/2026-09-20-b.md":    "---\nid: SPEC-2\n---\n# Spec B\n",
		"plans/2026-09-21-a.md":    "---\nid: PLAN-1\n---\n# Plan A\n\n### Task 1: One\n- [ ] x\n\n### Task 2: Two\n- [ ] y\n\n### Task 3: Three\n- [ ] z\n",
		"debt/2026-09-24-notes.md": "---\nid: DEBT-1\n---\n# Review NOTEs\n\n- [ ] one\n- [ ] two\n",
	})
	if len(calls) != 3 {
		t.Fatalf("git was asked %d times, want 3 (specs, plans, debt): %v", len(calls), calls)
	}
	for _, paths := range calls {
		seen := map[string]bool{}
		for _, p := range paths {
			if seen[p] {
				t.Errorf("one question named %s twice: %v", p, paths)
			}
			seen[p] = true
		}
		if filepath.Base(filepath.Dir(paths[0])) == "specs" && len(paths) != 2 {
			t.Errorf("the specs question named %d files, want both specs: %v", len(paths), paths)
		}
	}
	if names != 1 {
		t.Errorf("user.name was read %d times in one load, want 1", names)
	}
	for _, id := range []string{"SPC-0001", "SPC-0002", "PLN-0001", "PLN-0001.03", "DBT-0001.01", "DBT-0001.02"} {
		if got := b.Get(id).Author; got != "Sari" {
			t.Errorf("%s author = %q, want the one name the whole load read", id, got)
		}
	}
}

// One folder can hold a committed file and a file nobody committed yet. The
// first keeps its author and the second gets the name this checkout commits
// under.
func TestAuthorMixesCommittedAndNewFilesInOneFolder(t *testing.T) {
	t.Parallel()

	dir := authorRepo(t, "Ana", map[string]string{
		"specs/2026-09-20-a.md": "---\nid: SPEC-1\n---\n# Spec A\n",
	})
	late := filepath.Join(dir, ".acta", "specs", "2026-09-21-b.md")
	if err := os.WriteFile(late, []byte("---\nid: SPEC-2\n---\n# Spec B\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "config", "user.name", "Sari")
	b := loadDir(t, dir)
	if got := b.Get("SPC-0001").Author; got != "Ana" {
		t.Errorf("the committed spec has author %q, want Ana", got)
	}
	if got := b.Get("SPC-0002").Author; got != "Sari" {
		t.Errorf("the new spec has author %q, want the user.name Sari", got)
	}
}
```

- [ ] **Step 3: Run the tests to see them fail**

Run: `go test ./internal/gitc/ ./internal/board/`
Expected: FAIL to build, with `undefined: Authors` in `gitc` and `undefined: gitAuthors` in `board`.

- [ ] **Step 4: Replace `Author` with `Authors` in `internal/gitc/gitc.go`**

Delete `func Author` and its comment. Put this in its place:

```go
// Authors gives, for each path git knows, the name of the person whose commit
// first added it. It asks git once for all the paths, since one call per file
// made every board load take seconds. The paths sit right inside repo. A path
// git never saw has no entry, and a folder outside any checkout gives an
// empty map without running git.
func Authors(repo string, paths []string) map[string]string {
	found := map[string]string{}
	// With no paths, git would log the whole repo. Stop before that.
	if len(paths) == 0 || !inRepo(repo) {
		return found
	}
	// No renames: a file moved inside the folder counts as added where it is
	// now, the same answer a one-file log gives. quotePath off keeps names
	// with non-ASCII letters as they are. Each commit line starts with a NUL
	// byte, so an author line never reads as a file name.
	args := append([]string{"-c", "core.quotePath=false", "log", "--no-renames", "--diff-filter=A",
		"--relative", "--name-only", "--format=%x00%an", "--"}, paths...)
	out, err := run(repo, args...)
	if err != nil {
		return found
	}
	name := ""
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "\x00"):
			name = strings.TrimSpace(line[1:])
		case line != "" && name != "":
			// Git lists the newest commit first, so the last add seen is the first one.
			found[filepath.Join(repo, line)] = name
		}
	}
	return found
}
```

- [ ] **Step 5: Rewrite `internal/board/closed.go`**

```go
package board

import (
	"path/filepath"

	"github.com/iyay/acta/internal/gitc"
)

// The two questions fillAuthors asks git, as variables, so a test can count
// them without a real repo: a board loads a lot and a plan with ten tasks is
// still one file.
var (
	gitAuthors  = gitc.Authors
	gitUserName = gitc.UserName
)

// fillAuthors gives every item the name of the person whose commit first added
// the file it lives in, so a task takes the author of its plan file and a debt
// item the author of its debt file. Git is asked once per folder, since once
// per file made every load take seconds. A file git has no commit for belongs
// to whoever commits here now, and a folder outside git leaves the field
// empty, so the detail leaves the AUTHOR line out there.
func (b *Board) fillAuthors(root string) {
	byDir := map[string][]string{}
	seen := map[string]bool{}
	for _, it := range b.Items {
		if !it.OnDisk || it.Path == "" || seen[it.Path] {
			continue
		}
		seen[it.Path] = true
		dir := filepath.Dir(it.Path)
		byDir[dir] = append(byDir[dir], it.Path)
	}
	found := map[string]string{}
	for dir, paths := range byDir {
		for p, a := range gitAuthors(dir, paths) {
			found[p] = a
		}
	}
	me, asked := "", false
	for _, it := range b.Items {
		if !it.OnDisk || it.Path == "" {
			continue
		}
		a, ok := found[it.Path]
		if !ok {
			if !asked {
				me, asked = gitUserName(root), true
			}
			a = me
		}
		it.Author = a
	}
}
```

- [ ] **Step 6: Run the tests to see them pass**

Run: `go test ./internal/gitc/ ./internal/board/ -run 'Author' -v`
Expected: PASS for `TestAuthorsAsksOnceForManyFiles`, `TestAuthorsWithNothingToFind`, `TestAuthorIsAskedOncePerFolder`, `TestAuthorMixesCommittedAndNewFilesInOneFolder`, `TestAuthorComesFromTheFirstCommit`, `TestDebtItemTakesTheDebtFile`'s author check, and the TUI `TestDetailShowsTheAuthorUnderTheStatus` stays green in the full run below.

- [ ] **Step 7: Run the full gates**

Run: `gofmt -l internal && go vet ./... && go test ./...`
Expected: `gofmt` prints nothing, `go vet` is clean, every package passes.

- [ ] **Step 8: Measure the speed**

Run: `d=$(mktemp -d) && go build -o "$d/acta" ./cmd/acta && time "$d/acta" show SCR-0013`
Expected: the same output as main's `acta show SCR-0013`, in well under 0.5 s total (main takes about 2.9 s). Put the time in the task report.

- [ ] **Step 9: Commit**

```bash
git add internal/gitc/gitc.go internal/gitc/gitc_test.go internal/board/closed.go internal/board/board_test.go
git commit -m "perf(board): ask git for authors once per folder, not once per file"
```
