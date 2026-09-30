---
parent: bugs/2026-09-30-scratch-add-words-refused-on-old-item
created: "2026-09-30"
id: PLN-0044
hash: ergzrsb
started: "2026-09-30"
finished: "2026-09-30"
---
# Scratch Add Old Item Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `acta scratch add <id> --section <known name>` on an item with no body schema appends the text at the end of the file instead of refusing, so no answer is lost.

**Architecture:** In `AppendScratch` (`internal/write/scratch.go`), the old-item branch stops refusing a section and always does the plain append. The unknown-section check that runs before it stays.

**Tech Stack:** Go, standard library only.

**Spec:** `.acta/specs/2026-09-30-scratch-add-old-item-design.md`

**Tests:** fast `scripts/test`, full `scripts/test --full`.

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Items that have a schema behave exactly as today.
- An unknown section name is still refused, for old and new items alike.
- The scratch skill text and the CLI usage line are not changed.
- No new dependency in go.mod.
- Comments use plain English that a 10-year-old can read back: short words, and they say why.
- Before the commit, run `gofmt -l cmd internal` (it must print nothing), `go vet ./internal/write/` and `go test ./internal/write/ ./internal/cli/`.

## File Map

- `internal/write/scratch.go`: `AppendScratch` comment and the old-item branch.
- `internal/write/scratch_test.go`: `TestAppendScratchOldItem`.

## Waves

- Wave 1: Task 1.

---

### Task 1: Any known section on an old item does the plain append

**Files:**
- Modify: `internal/write/scratch.go` (the comment above `AppendScratch`, and the `if !board.HasSchema(...)` branch inside it)
- Test: `internal/write/scratch_test.go` (`TestAppendScratchOldItem`)

**verify:** No call to `AppendScratch` on an item without a schema loses the text when the section name is known. List every section name (`""`, `words`, `context`, `log`, `questions`, an unknown name) and what each does to an old item and to an item with a schema. An unknown name is refused on both and leaves the file unchanged. Items with a schema keep today's output for every section.

**Interfaces:**
- Consumes: `AppendScratch(cfg config.Config, b *board.Board, id, section string, text []byte) (Outcome, error)`; test helpers `fixNow(t)`, `repoWith(t, files)`, `mustLoad(t, cfg)`, `refuse(t, cfg, msg, fn)` already in `internal/write`.
- Produces: nothing new; the signature stays.

- [x] **Step 1: Write the failing test**

Replace `TestAppendScratchOldItem` in `internal/write/scratch_test.go` with:

```go
func TestAppendScratchOldItem(t *testing.T) {
	// An old item has no schema field, so it has no sections to fill. Every
	// known section, and no section at all, adds the text at the end of the
	// file, so an answer is never lost. An unknown name is still refused.
	fixNow(t)
	const old = "---\nid: SCR-0001\nhash: aaaaaaa\ntitle: old\nstatus: raw\n---\nfree text\n"
	const want = old + "\nmore\n"
	for _, section := range []string{"", "words", "context", "log", "questions"} {
		t.Run("section "+section, func(t *testing.T) {
			cfg := repoWith(t, map[string]string{".acta/scratch/2026-09-01-old.md": old})
			path := filepath.Join(cfg.Root, "scratch", "2026-09-01-old.md")
			o, err := AppendScratch(cfg, mustLoad(t, cfg), "SCRATCH-1", section, []byte("more\n"))
			if err != nil {
				t.Fatal(err)
			}
			if o.Path != path {
				t.Errorf("path %s want %s", o.Path, path)
			}
			if src, _ := os.ReadFile(path); string(src) != want {
				t.Errorf("the old item was reshaped:\ngot  %q\nwant %q", src, want)
			}
		})
	}
	t.Run("unknown section", func(t *testing.T) {
		cfg := repoWith(t, map[string]string{".acta/scratch/2026-09-01-old.md": old})
		path := filepath.Join(cfg.Root, "scratch", "2026-09-01-old.md")
		refuse(t, cfg, `unknown section "bogus"; use words, context, log or questions`, func() error {
			_, err := AppendScratch(cfg, mustLoad(t, cfg), "SCRATCH-1", "bogus", []byte("y\n"))
			return err
		})
		if after, _ := os.ReadFile(path); string(after) != old {
			t.Errorf("the refused write changed the file: %q", after)
		}
	})
}
```

If `repoWith` cannot be called twice in one test (for example, it pins something global), stop and report NEEDS_CONTEXT instead of changing the helper.

- [x] **Step 2: Run the test to see it fail**

Run: `go test ./internal/write/ -run TestAppendScratchOldItem -v`
Expected: FAIL in the `words`, `context`, `log` and `questions` subtests, each with "SCR-0001 is an old item with no sections"; the `""` and `unknown section` subtests pass.

- [x] **Step 3: Write the code**

In `internal/write/scratch.go`, the comment above `AppendScratch` becomes:

```go
// AppendScratch puts text in one part of a scratch body, one blank line
// below what that part already holds. An old item has no parts, so any
// section lands as the plain append at the end of the file, and the text
// is never lost. Any status takes more text: an idea that was specced or
// dropped can still collect an answer.
```

Inside `AppendScratch`, the old-item branch becomes:

```go
	if !board.HasSchema(board.Parse(src).Front) {
		if !strings.HasSuffix(out, "\n") {
			out += "\n"
		}
		out += "\n" + string(text)
	} else {
```

(The removed lines are the `if section != "" { return Outcome{}, bad("%s is an old item with no sections", it.ShortID) }` block. Leave the rest of the function as it is.)

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/write/ ./internal/cli/`
Expected: PASS.

Run: `grep -rn 'old item with no sections' internal cmd`
Expected: no output.

- [x] **Step 5: Commit**

```bash
gofmt -l cmd internal && go vet ./internal/write/ && go test ./internal/write/ ./internal/cli/
git add internal/write/scratch.go internal/write/scratch_test.go
git commit -m "Scratch add takes any known section on an old item as a plain append"
```
