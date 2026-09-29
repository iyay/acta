---
parent: bugs/2026-09-28-session-start-writes-through-gitignore-link
id: PLN-0020
hash: y7ijyab
---

# Refuse a Linked .gitignore Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `hook.EnsureGitignore` never writes through a `.gitignore` that is a symlink, so session start and `acta tick` stop writing into the link target.

**Architecture:** One `os.Lstat` check inside `EnsureGitignore`, before it reads or writes. Every caller (`cli/hook.go`, `cli/tick.go`, `doctor.Fix`) goes through this one function, so the fix sits there and no caller changes.

**Tech Stack:** Go, standard library only.

**Spec:** none (Bounded, approved in chat on 2026-09-28)

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- acta never writes through a `.gitignore` symlink, whether the link is dangling, points inside the repo, or points outside it.
- Callers keep dropping the error; no caller file changes.
- Comments in plain English a 10-year-old can read; say why, not what.

---

## File map

- Modify: `internal/hook/gitignore.go` (`EnsureGitignore`)
- Test: `internal/hook/gitignore_test.go`

## Waves

- Wave 1: Task 1

### Task 1: `EnsureGitignore` refuses a linked `.gitignore`

**Files:**
- Modify: `internal/hook/gitignore.go:16-35`
- Test: `internal/hook/gitignore_test.go`

**verify:** No input makes `EnsureGitignore` change or create any file other than a plain (non-link) `<root>/.gitignore`. List every link shape checked (dangling target, target inside the repo, target outside the repo) and show the target bytes and existence are the same before and after. A plain `.gitignore`, a missing `.gitignore`, and the existing tests behave exactly as before.

**Interfaces:**
- Consumes: nothing new.
- Produces: `EnsureGitignore(root, line string) error` keeps its signature. When `<root>/.gitignore` is a symlink it writes nothing and returns a non-nil error whose text names the path and says it is a link.

- [x] **Step 1: Write the failing test**

Add to `internal/hook/gitignore_test.go` (reuse the file's `gitRoot` and `readFile` helpers):

```go
func TestEnsureGitignoreRefusesALinkedGitignore(t *testing.T) {
	cases := map[string]func(t *testing.T, root, outside string) string{
		"dangling target outside": func(t *testing.T, root, outside string) string {
			return filepath.Join(outside, "new.conf")
		},
		"existing target outside": func(t *testing.T, root, outside string) string {
			p := filepath.Join(outside, "victim.conf")
			if err := os.WriteFile(p, []byte("precious=1\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			return p
		},
		"target inside the repo": func(t *testing.T, root, outside string) string {
			p := filepath.Join(root, "notes.txt")
			if err := os.WriteFile(p, []byte("keep me\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			return p
		},
	}
	for name, target := range cases {
		t.Run(name, func(t *testing.T) {
			root := gitRoot(t)
			outside := t.TempDir()
			dst := target(t, root, outside)
			before, beforeErr := os.ReadFile(dst)
			if err := os.Symlink(dst, filepath.Join(root, ".gitignore")); err != nil {
				t.Fatal(err)
			}
			if err := EnsureGitignore(root, ".agents.json"); err == nil {
				t.Fatal("want an error for a linked .gitignore, got nil")
			}
			after, afterErr := os.ReadFile(dst)
			if os.IsNotExist(beforeErr) != os.IsNotExist(afterErr) || string(before) != string(after) {
				t.Fatalf("link target changed: before %q (%v), after %q (%v)", before, beforeErr, after, afterErr)
			}
		})
	}
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/hook/ -run TestEnsureGitignoreRefusesALinkedGitignore -v`
Expected: FAIL, "want an error for a linked .gitignore, got nil" or "link target changed" in each sub-test.

- [x] **Step 3: Write minimal implementation**

In `internal/hook/gitignore.go`, right after `path := filepath.Join(root, ".gitignore")`, add (and add `"fmt"` to the imports):

```go
	// A .gitignore that is a link would send the write to whatever file it
	// points at, even one outside the repo. Git does not read a linked
	// .gitignore either, so there is nothing to gain by following it.
	if fi, err := os.Lstat(path); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s is a link, not a file", path)
	}
```

Also add one sentence to the doc comment above `EnsureGitignore`: "It also writes nothing when .gitignore is a link."

- [x] **Step 4: Run tests to verify they pass**

Run: `go test ./...`
Expected: PASS, including the four existing `TestEnsureGitignore*` tests and the doctor tests.

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/hook/gitignore.go internal/hook/gitignore_test.go .acta/plans/2026-09-28-gitignore-link.md
git commit -m "fix(hook): refuse a .gitignore that is a symlink"
```
