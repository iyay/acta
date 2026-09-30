---
parent: bugs/2026-09-30-list-shows-slug-when-title-only-in-frontmatter
closes: [SCR-0021]
created: "2026-09-30"
id: PLN-0048
hash: zubk33q
---
# Item title falls back to the frontmatter title Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** An item with no body `# ` heading shows its frontmatter `title:` instead of its slug.

**Architecture:** One fallback in `fileItem` (`internal/board/board.go`): body heading first, then the frontmatter `title:` read through the existing `field` helper, then the slug. No file under `.acta/` changes.

**Tech Stack:** Go, standard `testing`.

**Spec:** `.acta/specs/2026-09-30-title-from-frontmatter-design.md`

**Tests:** fast `scripts/test ./internal/board/`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- The body `# ` heading still wins over the frontmatter, so every file that shows a title today shows the same one.
- No file under `.acta/` is edited to fix the old items.
- Comments are plain English a 10-year-old can read: short words, say why.

## File map

- Modify: `internal/board/board.go` (`fileItem`, the `if it.Title == ""` fallback)
- Test: `internal/board/board_test.go` (new test `TestFileItemTitleFallback`)

## Waves

- Wave 1: Task 1

---

### Task 1: Title falls back to the frontmatter title

**Files:**
- Modify: `internal/board/board.go` (function `fileItem`)
- Test: `internal/board/board_test.go`

**verify:** For every input, the title is the body heading when there is one, else the trimmed frontmatter `title:` when it is not empty, else the slug. No input with a body heading changes its title. List each case checked: title only, heading plus a different title, empty title, spaces-only title, null title, neither, unicode title.

**Interfaces:**
- Consumes: `Parse(src []byte) Doc` (`internal/board/parse.go:59`), `field(front map[string]any, key string) string` (`internal/board/board.go:716`), `fileItem(k Kind, id, path, date, slug string, legacy bool, doc Doc) *Item`.
- Produces: no new names.

- [ ] **Step 1: Write the failing test**

Add to `internal/board/board_test.go`:

```go
func TestFileItemTitleFallback(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, src, want string
	}{
		{"title only", "---\ntitle: Severity or priority\n---\nsome words\n", "Severity or priority"},
		{"heading wins", "---\ntitle: From front\n---\n# From body\n", "From body"},
		{"empty title", "---\ntitle: \"\"\n---\nwords\n", "my-slug"},
		{"spaces title", "---\ntitle: \"   \"\n---\nwords\n", "my-slug"},
		{"null title", "---\ntitle:\n---\nwords\n", "my-slug"},
		{"neither", "---\nstatus: raw\n---\nwords\n", "my-slug"},
		{"unicode title", "---\ntitle: Catat ide 日本語 ✓\n---\nwords\n", "Catat ide 日本語 ✓"},
	}
	for _, c := range cases {
		it := fileItem(KindScratch, "scratch/2026-09-30-my-slug", "scratch/2026-09-30-my-slug.md", "2026-09-30", "my-slug", false, Parse([]byte(c.src)))
		if it.Title != c.want {
			t.Errorf("%s: title = %q, want %q", c.name, it.Title, c.want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/board/ -run TestFileItemTitleFallback -v`
Expected: FAIL on "title only" and "unicode title" with `title = "my-slug"`. The other cases pass already.

- [ ] **Step 3: Write minimal implementation**

In `fileItem` (`internal/board/board.go`), replace:

```go
	if it.Title == "" {
		it.Title = slug
	}
```

with:

```go
	// Old scratch items keep their title only in the frontmatter, with no
	// heading in the body. Use it before falling back to the slug.
	if it.Title == "" {
		it.Title = field(doc.Front, "title")
	}
	if it.Title == "" {
		it.Title = slug
	}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/board/ -run TestFileItemTitleFallback -v`
Expected: PASS

Run: `scripts/test ./internal/board/`
Expected: ok

- [ ] **Step 5: Format, vet, commit**

```bash
gofmt -l internal/board/
go vet ./internal/board/
git add internal/board/board.go internal/board/board_test.go
git commit -m "fix(board): fall back to the frontmatter title before the slug"
```

`gofmt -l` prints nothing.
