---
parent: specs/2026-10-01-tui-debt-note-markdown-design
id: PLN-0062
created: "2026-10-01 04:46:10"
hash: v5ya0xy
started: "2026-10-01 04:46:48"
finished: "2026-10-01 04:48:54"
---
# Debt Note Rendered Like a Spec Body Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** The debt item detail draws its note through the same markdown renderer as a spec or plan body, with the same gaps and code colors, and never loses a character of the note.

**Architecture:** In the debt item branch of `buildDetailParts` (`internal/tui/detail.go`), escape `&`, `<`, `>` in the note, render it with `m.render(text, w)`, pass the result through `xansi.Hardwrap(out, w, true)` so no line is wider than the pane, then `fit` each line. The `xansi.Wordwrap` step goes away.

**Tech Stack:** Go, glamour (through `m.render`), `github.com/charmbracelet/x/ansi` (already imported in `detail.go` as `xansi`).

**Spec:** `.acta/specs/2026-10-01-tui-debt-note-markdown-design.md`

**Tests:** fast `scripts/test ./internal/tui`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Only the debt item branch of `buildDetailParts` changes. The header, the footer, other kinds and the debt file detail stay as they are.
- Escape exactly `&`, `<` and `>` (to `&amp;`, `&lt;`, `&gt;`). No other markdown escaping: backticks and `*` still turn into styling.
- Comments are plain English a 10-year-old can read. They say why, not what.

## File map

- Modify: `internal/tui/detail.go` (debt item branch of `buildDetailParts`)
- Modify: `internal/tui/detail_test.go`

## Waves

- Wave 1: Task 1.

---

### Task 1: Debt note goes through the markdown renderer

**Files:**
- Modify: `internal/tui/detail.go` (debt item branch of `buildDetailParts`)
- Test: `internal/tui/detail_test.go`

**verify:** With the real renderer, for every pane width from 10 to 160 and every test note, the drawn middle holds every non-space character of the note in order, except the backticks and `*` that markdown turns into styling. No drawn line is wider than the pane. The note begins after the same blank line and in the same column as a spec body drawn at the same width. The test notes hold `<uid>`, `&`, a ` -` after a space, a `` `code` `` span, a word longer than the pane, and wide runes. List the notes and widths checked.

**Interfaces:**
- Consumes: `Model.buildDetailParts(w int) (head, mid []string, foot string)`, `m.render func(string, int) string`, `xansi.Hardwrap(s string, limit int, preserveSpace bool) string`, test helpers `treeCfg`, `detailModel`, `onItem`, `plain`, `fit`.
- Produces: nothing new.

- [x] **Step 1: Write the failing tests**

In `internal/tui/detail_test.go`, replace the body of `TestDetailDebtItemNoteLosesNothingAtAnyWidth` so it covers the new notes and skips the characters markdown styles away:

```go
// A note is drawn by the markdown renderer at every width, and no line is
// cut, so the reader never loses a word, a lone "-" or a tag like <uid>.
func TestDetailDebtItemNoteLosesNothingAtAnyWidth(t *testing.T) {
	t.Parallel()

	notes := []string{
		"internal/write/mark.go: + on a task already done (or - on one already open) changes nothing, and more",
		`a checklist line gives "- [ ] - [ ] text".`,
		"averyveryveryverylongwordwithnospacesatallthatrunsonandon then short",
		"日本語のメモ and ünïcode words in one note",
		"Another local user can pre-create /tmp/pmb-<uid>: a & b",
		"view.go still runs `acta list --json` with *no* agent field",
	}
	// Markdown turns these into styling, so they are not drawn as text.
	styled := strings.NewReplacer("`", "", "*", "")
	for _, note := range notes {
		cfg := treeCfg(t, map[string]string{
			".acta/debt/2026-09-24-d.md": "---\nid: DEBT-1\n---\n# Review NOTEs\n\n- [ ] " + note + "\n",
		})
		m := onItem(t, detailModel(t, cfg), "DEBT-1.1")
		want := strings.Join(strings.Fields(styled.Replace(note)), "")
		for w := 10; w <= 160; w++ {
			_, mid, _ := m.buildDetailParts(w)
			var got strings.Builder
			for i, ln := range mid {
				if lw := lipgloss.Width(ln); lw > w {
					t.Errorf("at %d wide line %d is %d cells: %q", w, i, lw, plain(ln))
				}
				got.WriteString(strings.Join(strings.Fields(plain(ln)), ""))
			}
			if got.String() != want {
				t.Errorf("at %d wide the note reads %q, want %q", w, got.String(), want)
			}
		}
	}
}

// A note sits in the detail the way a spec body does: the same blank line
// above it and the same gap on the left, so the two kinds look alike.
func TestDetailDebtItemNoteHasTheSpecBodyGaps(t *testing.T) {
	t.Parallel()

	const text = "Tests that call lock without the base still leave lock files"
	cfg := treeCfg(t, map[string]string{
		".acta/debt/2026-09-24-d.md":  "---\nid: DEBT-1\n---\n# Review NOTEs\n\n- [ ] " + text + "\n",
		".acta/specs/2026-09-20-a.md": "---\nid: SPEC-1\n---\n" + text + "\n",
	})
	firstText := func(mid []string) (row, col int) {
		for i, ln := range mid {
			p := plain(ln)
			if c := strings.Index(p, "Tests"); c >= 0 {
				return i, c
			}
		}
		t.Fatalf("no line holds the text:\n%s", strings.Join(mid, "\n"))
		return 0, 0
	}
	for _, w := range []int{40, 80, 120} {
		_, debtMid, _ := onItem(t, detailModel(t, cfg), "DEBT-1.1").buildDetailParts(w)
		_, specMid, _ := onItem(t, detailModel(t, cfg), "SPEC-1").buildDetailParts(w)
		dr, dc := firstText(debtMid)
		_, sc := firstText(specMid)
		if dc != sc {
			t.Errorf("at %d wide the note starts at column %d, a spec body at %d", w, dc, sc)
		}
		if dr == 0 || strings.TrimSpace(plain(debtMid[dr-1])) != "" {
			t.Errorf("at %d wide the note has no blank line above it:\n%s", w, strings.Join(debtMid, "\n"))
		}
	}
}
```

If `SPEC-1` with a body and no title has a different first line than expected, give it a `# Title` line above the text. The test compares only where the text starts.

- [x] **Step 2: Run the tests to see them fail**

Run: `scripts/test ./internal/tui -run 'TestDetailDebtItemNote'`
Expected: FAIL. `TestDetailDebtItemNoteHasTheSpecBodyGaps` finds the note at column 0 with no blank line above it. `TestDetailDebtItemNoteLosesNothingAtAnyWidth` may already pass, because the plain wrap keeps every character.

- [x] **Step 3: Write the minimal code**

In `buildDetailParts`, replace the debt item branch's wrap with:

```go
	if it.Kind == board.KindDebtItem {
		// The list pane already names the other notes, so the middle is this
		// note alone, drawn like the body of a spec so both look the same.
		// The renderer drops text that looks like an HTML tag, such as
		// <uid>, so those signs are escaped first. It can also leave a line
		// wider than the pane, and fit would cut it, so Hardwrap breaks it.
		note := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(expandTabs(it.Title))
		for _, ln := range strings.Split(xansi.Hardwrap(m.render(note, w), w, true), "\n") {
			mid = append(mid, fit(ln, w))
		}
		return head, mid, paintDates(m.styles, dateLine(it, w))
	}
```

- [x] **Step 4: Run the tests to see them pass**

Run: `scripts/test ./internal/tui`
Expected: PASS, including the older debt item tests. They use a render stub that returns the text as it is, and `Hardwrap` still wraps it to the pane. Then run `go vet ./internal/tui && gofmt -l internal/tui` and expect no output.

- [x] **Step 5: Commit**

```bash
git add internal/tui/detail.go internal/tui/detail_test.go
git commit -m "Debt note renders like a spec body, with escaped tags (PLN-0062)"
```

## Fix round 1

### Task 2: A backslash in a note is drawn as written

Review round 1 found that markdown reads `\|` as an escaped pipe. The real note in `.acta/debt/2026-09-29-harness.md` that holds `(?:^|&&|;|\|\|)` draws as `(?:^|&&|;|||)`. The spec says every character the note holds is drawn as written.

**Files:**
- Modify: `internal/tui/detail.go` (the escape in the debt item branch of `buildDetailParts`)
- Test: `internal/tui/detail_test.go`

**verify:** For every pane width from 10 to 160, a note that holds backslashes before punctuation draws every backslash, and no other test note changes. List the notes and widths checked.

**Interfaces:**
- Consumes: the Task 1 test `TestDetailDebtItemNoteLosesNothingAtAnyWidth`.
- Produces: nothing new.

- [x] **Step 1: Write the failing test**

Add this note to the `notes` list of `TestDetailDebtItemNoteLosesNothingAtAnyWidth`:

```go
		`brainCmd: the (?:^|&&|;|\|\|) alternation is dead`,
```

- [x] **Step 2: Run the test to see it fail**

Run: `scripts/test ./internal/tui -run TestDetailDebtItemNoteLosesNothingAtAnyWidth`
Expected: FAIL. The note reads `(?:^|&&|;|||)`.

- [x] **Step 3: Write the minimal code**

Put the backslash first in the replacer, and name it in the comment:

```go
		// The renderer drops text that looks like an HTML tag, such as
		// <uid>, and eats a backslash before a sign, so those are escaped
		// first. It can also leave a line wider than the pane, and fit
		// would cut it, so Hardwrap breaks it.
		note := strings.NewReplacer(`\`, `\\`, "&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(expandTabs(it.Title))
```

- [x] **Step 4: Run the tests to see them pass**

Run: `scripts/test ./internal/tui`
Expected: PASS. Then `go vet ./internal/tui && gofmt -l internal/tui` prints nothing.

- [x] **Step 5: Commit**

```bash
git add internal/tui/detail.go internal/tui/detail_test.go
git commit -m "Backslash in a debt note is drawn as written (PLN-0062 fix round 1)"
```
