---
parent: specs/2026-10-01-tui-debt-item-detail-design
id: PLN-0061
created: "2026-10-01 00:12:04"
hash: h77lzid
started: "2026-10-01 00:16:59"
finished: "2026-10-01 00:21:31"
---
# Debt Item Detail Shows the Full Note Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** The detail of a debt item shows the whole note, wrapped to the pane, and drops the `DEBT :` header line, the list of the other notes and the text of the debt file.

**Architecture:** One change in `buildDetailParts` (`internal/tui/detail.go`). For a `board.KindDebtItem` the kind line gets no value, so the header leaves it out, and the middle is the problems plus `it.Title` wrapped with `xansi.Wrap`. The body and `workLines` are skipped for that kind. `debtLines` keeps working for a debt file, so its branch for a debt item goes away.

**Tech Stack:** Go, Bubble Tea, lipgloss, `github.com/charmbracelet/x/ansi` (already imported in `view.go` as `xansi`).

**Spec:** `.acta/specs/2026-10-01-tui-debt-item-detail-design.md`

**Tests:** fast `scripts/test ./internal/tui`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Only a debt item (`board.KindDebtItem`) changes. A debt file (`board.KindDebt`) and every other kind keep the detail they have now.
- The header keeps ID, STATUS, AUTHOR, FROM and FILE for a debt item. The footer date line stays as it is.
- No "note N of M" marker. No change to `internal/board`.
- The note is plain text, not markdown: it goes through `xansi.Wrap`, never through `m.render`, so text like `<uid>` stays as written.
- Comments are plain English a 10-year-old can read. They say why, not what.

## File map

- Modify: `internal/tui/detail.go` (`buildDetailParts`, `debtLines`)
- Modify: `internal/tui/detail_test.go` (replace the two debt item tests, add the long note test and the debt file test)
- Modify: `internal/tui/view_test.go` (the header check that wants `DEBT\s+: a$`)

## Waves

- Wave 1: Task 1.

---

### Task 1: Debt item detail is the note alone

**Files:**
- Modify: `internal/tui/detail.go` (`buildDetailParts`, `debtLines`)
- Test: `internal/tui/detail_test.go`, `internal/tui/view_test.go`

**verify:** For every debt item at every pane width, the detail holds every word of the note, no line wider than the pane, no `DEBT` header line, no line of any other note in the same file, and no text of the debt file outside the note. A debt file and every other kind draw the same detail as before. List each kind and width checked.

**Interfaces:**
- Consumes: `Model.buildDetailParts(w int) (head, mid []string, foot string)`, `Model.debtLines(it *board.Item, w int) []string`, `xansi.Wrap(s string, limit int, breakpoints string) string`, test helpers `detailLines(t, cfg, id)`, `detailOf(t, id)`, `plainLines`, `ruleLine`, `treeCfg`, `boardFiles`, `detailModel`.
- Produces: nothing new. `debtLines` keeps its signature.

- [x] **Step 1: Write the failing tests**

In `internal/tui/detail_test.go`, delete `TestDetailDebtItemListsEveryLineOfItsFile` and `TestDetailDebtItemShowsTheTextAroundItsLine`, and add these in their place:

```go
// A debt item is one note, and the list pane already names the other notes,
// so its detail shows the whole note and nothing else of the file.
func TestDetailDebtItemShowsOnlyItsNote(t *testing.T) {
	t.Parallel()

	lines := plainLines(detailOf(t, "DEBT-1.1"))
	head := strings.Join(lines[:ruleLine(t, lines)], "\n")
	below := strings.Join(lines[ruleLine(t, lines):], "\n")
	if strings.Contains(head, "DEBT ") {
		t.Errorf("the header still has a DEBT line:\n%s", head)
	}
	if n := strings.Count(below, "first note"); n != 1 {
		t.Errorf("the note shows %d times under the header, want once:\n%s", n, below)
	}
	for _, gone := range []string{"second note", "Prose about the review.", "Review NOTEs"} {
		if strings.Contains(below, gone) {
			t.Errorf("the debt item detail still shows %q:\n%s", gone, below)
		}
	}
}

// A note longer than the pane is wrapped, so every word of it can be read and
// no line spills over the wall.
func TestDetailDebtItemWrapsALongNote(t *testing.T) {
	t.Parallel()

	note := strings.TrimSpace(strings.Repeat("a long note word ", 20)) + " last<uid>"
	cfg := treeCfg(t, map[string]string{
		".acta/debt/2026-09-24-d.md": "---\nid: DEBT-1\n---\n# Review NOTEs\n\n- [ ] " + note + "\n",
	})
	lines := plainLines(detailLines(t, cfg, "DEBT-1.1"))
	below := lines[ruleLine(t, lines):]
	if got := strings.Join(strings.Fields(strings.Join(below, " ")), " "); !strings.Contains(got, note) {
		t.Errorf("the detail does not hold the whole note %q:\n%s", note, strings.Join(below, "\n"))
	}
	for i, ln := range lines {
		if w := lipgloss.Width(ln); w > 100 {
			t.Errorf("line %d is %d cells wide: %q", i, w, ln)
		}
	}
}

// A debt file still lists every one of its lines, because only the debt
// item's detail changed.
func TestDetailDebtFileListsEveryLine(t *testing.T) {
	t.Parallel()

	m := detailModel(t, treeCfg(t, boardFiles()))
	lines := plainLines(m.debtLines(m.board.Get("DEBT-1"), 100))
	wantInOrder(t, lines, "○ DBT-0001.01  first note", "✓ DBT-0001.02  second note")
}
```

In `internal/tui/view_test.go`, in the loop that checks the detail header (the one with `` `DEBT\s+: a$` ``), drop the `DEBT` entry and add a check that it is gone:

```go
	for _, want := range []string{`STATUS\s+: open$`, `FROM\s+: PLN-0003 . Short IDs$`} {
		if ok, _ := regexp.MatchString("(?m)"+want, detail); !ok {
			t.Errorf("detail header is missing %q, got:\n%s", want, detail)
		}
	}
	if ok, _ := regexp.MatchString(`(?m)^DEBT\s+:`, detail); ok {
		t.Errorf("detail header still has a DEBT line:\n%s", detail)
	}
```

- [x] **Step 2: Run the tests to see them fail**

Run: `scripts/test ./internal/tui -run 'TestDetailDebt|TestView'`
Expected: FAIL. `TestDetailDebtItemShowsOnlyItsNote` finds the DEBT header line and "second note"; `TestDetailDebtItemWrapsALongNote` finds the note cut off; the view test finds the DEBT line. `TestDetailDebtFileListsEveryLine` may already pass.

- [x] **Step 3: Write the minimal code**

In `internal/tui/detail.go`, add `xansi "github.com/charmbracelet/x/ansi"` to the imports.

In `buildDetailParts`, before the `fields` list:

```go
	title := it.Title
	if it.Kind == board.KindDebtItem {
		// A note is too long for one header line, so it goes in the middle.
		title = ""
	}
```

and use `{kindLabel(it.Kind), title},` in the list. After the loop that adds `it.Problems` to `mid`, replace the rest of the function with:

```go
	if it.Kind == board.KindDebtItem {
		// The list pane already names the other notes, so the middle is this
		// note alone, wrapped so every word can be read.
		for _, ln := range strings.Split(xansi.Wrap(expandTabs(it.Title), w, ""), "\n") {
			mid = append(mid, fit(ln, w))
		}
		return head, mid, paintDates(m.styles, dateLine(it, w))
	}
	mid = append(mid, m.workLines(it, w)...)
	for _, ln := range strings.Split(m.render(expandTabs(it.Body), w), "\n") {
		mid = append(mid, fit(ln, w))
	}
	return head, mid, paintDates(m.styles, dateLine(it, w))
```

`debtLines` now only gets a debt file. Drop its debt item branch and the bold line, and fix its comment:

```go
// debtLines are the checklist lines of a debt file, one line each.
func (m Model) debtLines(file *board.Item, w int) []string {
	var out []string
	for _, id := range file.Children {
		if line := m.board.Get(id); line != nil {
			out = append(out, workLine(m.styles, line, false, w))
		}
	}
	return out
}
```

In `workLines`, change `case board.KindDebt, board.KindDebtItem:` to `case board.KindDebt:`, because a debt item no longer reaches it, and in its doc comment change "a debt file or debt item the lines of the debt file" to "a debt file its lines".

- [x] **Step 4: Run the tests to see them pass**

Run: `scripts/test ./internal/tui`
Expected: PASS, including `TestDetailIsNeverEmpty` and the tab and width test that walks `DEBT-1.1`.

Then run `go vet ./internal/tui && gofmt -l internal/tui` and expect no output.

- [x] **Step 5: Commit**

```bash
git add internal/tui/detail.go internal/tui/detail_test.go internal/tui/view_test.go
git commit -m "Debt item detail shows the full note alone (PLN-0061)"
```

## Fix round 1

### Task 2: A wrapped note never loses text

Review round 1 found that `xansi.Wrap(s, w, "")` can return a line wider than `w` when a `-` follows a space. `fit` then cuts the end of that line. Real note DBT-0046.01 at w=52 wraps to a 54-cell line `...already done (or -`, and the detail drops the ` -`.

**Files:**
- Modify: `internal/tui/detail.go` (the debt item branch of `buildDetailParts`)
- Test: `internal/tui/detail_test.go`

**verify:** For every debt item note and every pane width from 10 to 160, every wrapped line is at most w cells wide before `fit`, so `fit` never cuts anything, and joining the drawn middle lines gives back every non-space character of the note in order. The test sweeps the widths over notes that hold ` -` after a space, a word longer than the pane, and wide runes. List the notes and widths checked.

**Interfaces:**
- Consumes: `xansi.Wordwrap(s string, limit int, breakpoints string) string`, `xansi.Hardwrap(s string, limit int, preserveSpace bool) string`, test helpers `treeCfg`, `detailModel`, `onItem`, `plain`, `ruleLine`.
- Produces: nothing new.

- [x] **Step 1: Write the failing test**

In `internal/tui/detail_test.go`, add:

```go
// A note is wrapped to the pane at every width, and no wrapped line is cut,
// so the reader never loses a word or a sign like a lone "-".
func TestDetailDebtItemNoteLosesNothingAtAnyWidth(t *testing.T) {
	t.Parallel()

	notes := []string{
		"internal/write/mark.go: + on a task already done (or - on one already open) changes nothing, and more",
		`a checklist line gives "- [ ] - [ ] text".`,
		"averyveryveryverylongwordwithnospacesatallthatrunsonandon then short",
		"日本語のメモ and ünïcode words in one note",
	}
	for _, note := range notes {
		cfg := treeCfg(t, map[string]string{
			".acta/debt/2026-09-24-d.md": "---\nid: DEBT-1\n---\n# Review NOTEs\n\n- [ ] " + note + "\n",
		})
		m := onItem(t, detailModel(t, cfg), "DEBT-1.1")
		want := strings.Join(strings.Fields(note), "")
		for w := 10; w <= 160; w++ {
			_, mid, _ := m.buildDetailParts(w)
			var got strings.Builder
			for _, ln := range mid {
				got.WriteString(strings.Join(strings.Fields(plain(ln)), ""))
			}
			if got.String() != want {
				t.Errorf("at %d wide the note reads %q, want %q", w, got.String(), want)
			}
		}
	}
}
```

- [x] **Step 2: Run the test to see it fail**

Run: `scripts/test ./internal/tui -run TestDetailDebtItemNoteLosesNothingAtAnyWidth`
Expected: FAIL at width 52 for the first note (the ` -` is gone) and at width 90 or so for the second.

- [x] **Step 3: Write the minimal code**

In `buildDetailParts`, in the debt item branch, change the wrap call to:

```go
		// Wordwrap can leave a line wider than the pane when a "-" follows a
		// space, and fit would then cut words off. Hardwrap breaks it again.
		note := xansi.Hardwrap(xansi.Wordwrap(expandTabs(it.Title), w, ""), w, true)
		for _, ln := range strings.Split(note, "\n") {
			mid = append(mid, fit(ln, w))
		}
```

- [x] **Step 4: Run the tests to see them pass**

Run: `scripts/test ./internal/tui`
Expected: PASS. Then `go vet ./internal/tui && gofmt -l internal/tui` prints nothing.

- [x] **Step 5: Commit**

```bash
git add internal/tui/detail.go internal/tui/detail_test.go
git commit -m "Wrapped debt note never loses text at any width (PLN-0061 fix round 1)"
```
