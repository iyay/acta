---
parent: bugs/2026-09-28-session-start-writes-through-acta-folder-link
id: PLAN-21
hash: vb12
---

# Keep EnsureGitignore Inside the Repo Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `hook.EnsureGitignore` never writes a `.gitignore` in a folder that sits outside the repo, even when the root folder (`.acta`) is a symlink.

**Architecture:** `inGitRepo` becomes `gitTop`, which returns the folder that holds `.git`. `EnsureGitignore` then resolves symlinks on both the root and that folder and refuses when the real root is not under the real repo folder. Every caller goes through this one function, so no caller changes.

**Tech Stack:** Go, standard library only.

**Spec:** none (Bounded, approved in chat on 2026-09-28)

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- `EnsureGitignore` never writes outside the repo that holds the root folder, whatever links sit in the root path.
- A root that is a link to a folder inside the same repo, and a repo reached through a linked parent folder (macOS `/var` -> `/private/var`), keep working as today.
- Callers keep dropping the error; no caller file changes.
- Out of scope: a `root:` in the user's own `.acta.yaml` that passes through a linked middle folder into another git repo. `acta doctor` already reports that case; this plan covers a committed `.acta` link.
- Comments in plain English a 10-year-old can read; say why, not what.

---

## File map

- Modify: `internal/hook/gitignore.go` (`EnsureGitignore`, `inGitRepo` renamed to `gitTop`)
- Test: `internal/hook/gitignore_test.go`

## Waves

- Wave 1: Task 1

### Task 1: `EnsureGitignore` refuses a root that resolves outside the repo

**Files:**
- Modify: `internal/hook/gitignore.go`
- Test: `internal/hook/gitignore_test.go`

**verify:** No root path makes `EnsureGitignore` create or change a file outside the repo whose `.git` it found. Enumerate every root shape checked (plain folder, link to a folder outside the repo, link to a folder in another git repo, link to a folder inside the same repo, repo reached through a linked parent folder) and every caller of `EnsureGitignore`, and report both lists. The first two outside shapes write nothing and return an error; the in-repo shapes write the line as before; the five existing `TestEnsureGitignore*` tests stay green.

**Interfaces:**
- Consumes: nothing new.
- Produces: `EnsureGitignore(root, line string) error` keeps its signature. `gitTop(dir string) (string, bool)` replaces `inGitRepo` inside package `hook` (unexported, no other users).

- [x] **Step 1: Write the failing test**

Add to `internal/hook/gitignore_test.go` (reuse the file's `gitRoot` and `readFile` helpers):

```go
func TestEnsureGitignoreRefusesARootLinkedOutTheRepo(t *testing.T) {
	cases := map[string]func(t *testing.T) string{
		"plain folder outside": func(t *testing.T) string { return t.TempDir() },
		"folder in another repo": func(t *testing.T) string { return gitRoot(t) },
	}
	for name, outside := range cases {
		t.Run(name, func(t *testing.T) {
			repo := gitRoot(t)
			out := outside(t)
			root := filepath.Join(repo, ".acta")
			if err := os.Symlink(out, root); err != nil {
				t.Fatal(err)
			}
			if err := EnsureGitignore(root, ".agents.json"); err == nil {
				t.Fatal("want an error for a root linked out of the repo, got nil")
			}
			if _, err := os.Stat(filepath.Join(out, ".gitignore")); !os.IsNotExist(err) {
				t.Fatalf("wrote a .gitignore outside the repo (stat err %v)", err)
			}
		})
	}
}

func TestEnsureGitignoreFollowsARootLinkedInsideTheRepo(t *testing.T) {
	repo := gitRoot(t)
	real := filepath.Join(repo, "docs", "planning")
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(repo, ".acta")
	if err := os.Symlink(real, root); err != nil {
		t.Fatal(err)
	}
	if err := EnsureGitignore(root, ".agents.json"); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(real, ".gitignore")); got != ".agents.json\n" {
		t.Fatalf("gitignore = %q, want %q", got, ".agents.json\n")
	}
}

func TestEnsureGitignoreWorksThroughALinkedParentFolder(t *testing.T) {
	repo := gitRoot(t)
	link := filepath.Join(t.TempDir(), "via")
	if err := os.Symlink(repo, link); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(link, ".acta")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := EnsureGitignore(root, ".agents.json"); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(repo, ".acta", ".gitignore")); got != ".agents.json\n" {
		t.Fatalf("gitignore = %q, want %q", got, ".agents.json\n")
	}
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/hook/ -run 'TestEnsureGitignore(RefusesARootLinkedOutTheRepo|FollowsARootLinkedInsideTheRepo|WorksThroughALinkedParentFolder)' -v`
Expected: `TestEnsureGitignoreRefusesARootLinkedOutTheRepo` FAILs in both sub-tests ("want an error ... got nil" and a `.gitignore` written outside). The other two already pass; they guard the in-repo shapes against the fix.

- [x] **Step 3: Write minimal implementation**

In `internal/hook/gitignore.go`, replace `inGitRepo` with `gitTop`:

```go
// gitTop walks up from dir looking for .git and returns the folder that
// holds it. .git is a file in a linked worktree, so only its presence is
// checked, not that it is a folder.
func gitTop(dir string) (string, bool) {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
```

In `EnsureGitignore`, replace the `if !inGitRepo(root) { return nil }` block with:

```go
	// A root that is itself a link would lead the walk into the folder it
	// points at, which may hold another repo's .git. Start from its parent
	// so the repo found is the one the link sits in.
	start := root
	if fi, err := os.Lstat(root); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		start = filepath.Dir(root)
	}
	top, ok := gitTop(start)
	if !ok {
		return nil
	}
	// The walk above reads the path as written, so a root that is a link to
	// a folder somewhere else still looks like it sits in this repo. Follow
	// the links on both sides and make sure the real root is under the real
	// repo folder, so the write can never land outside it.
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	realTop, err := filepath.EvalSymlinks(top)
	if err != nil {
		return err
	}
	if rel, err := filepath.Rel(realTop, realRoot); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("%s is outside the repo %s", root, top)
	}
```

Add one sentence to the doc comment above `EnsureGitignore`: "It also writes nothing when the root folder, after links are followed, is outside the repo."

- [x] **Step 4: Run tests to verify they pass**

Run: `go test ./...`
Expected: PASS, including the old `TestEnsureGitignore*` tests and the doctor tests.

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/hook/gitignore.go internal/hook/gitignore_test.go .acta/plans/2026-09-28-acta-folder-link.md
git commit -m "fix(hook): keep EnsureGitignore inside the repo when the root is a link"
```
