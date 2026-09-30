---
id: PLN-0057
created: "2026-09-30"
hash: b7a6i3b
started: "2026-09-30"
finished: "2026-09-30"
---
# Times in the Detail Dates and Self-Hiding Toasts Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Write and show the time, down to the second, in the `created`, `started` and `finished` dates, and make every status line toast hide itself after `toastFor`.

**Architecture:** One layout constant in `internal/write` carries `2006-01-02 15:04:05`, and every writer of the three fields uses it. `board.dateField` reads a bare day or a day with a time. The detail footer prints what is there. In `internal/tui`, `Update` wraps the old body and starts the clear timer whenever the status changed to a new, non-empty text.

**Tech Stack:** Go, Bubble Tea, lipgloss, yaml.v3.

**Spec:** `.acta/specs/2026-09-30-tui-date-times-and-toasts-design.md`

**Tests:** `scripts/test ./internal/write ./internal/board ./internal/tui`, `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Dates are written as `YYYY-MM-DD HH:MM:SS` in local time (`Now()` as it comes), quoted, for example `"2026-09-30 16:14:05"`.
- `dateField` accepts exactly two shapes: `YYYY-MM-DD` and `YYYY-MM-DD HH:MM:SS`. Nothing else, and no migration of old files.
- File names, the day heading in a new scratch body, and the status line clock keep their current shape.
- Every status toast, errors included, lasts `toastFor` (2 seconds). The search box is not a toast.
- Comments are plain English a 10-year-old can read: short words, say why. Match the comment style already in each package.
- Run `gofmt -l .` and `go vet` on the touched packages before every commit, inside the task commit.

## File map

- `internal/write/dates.go`, `internal/write/ids.go`, `internal/write/scratch.go`: the three writers (Task 1).
- `internal/write/*_test.go`: tests that expect a bare day in these fields (Task 1).
- `internal/board/board.go`: `dateField`, `isDate`, and the `Created`, `StartedOn` and `Finished` comments (Task 1).
- `internal/tui/detail.go` and `internal/tui/detail_test.go`: the footer and its tests (Task 1).
- `internal/tui/model.go`, `internal/tui/select.go`, `internal/tui/model_test.go`, `internal/tui/select_test.go`: the toast timer (Task 2).

## Waves

- Wave 1: Task 1 and Task 2 in parallel. They share no file.

---

### Task 1: The three dates carry the time

**Files:**
- Modify: `internal/write/dates.go` (`MarkStarted`, `MarkFinished`)
- Modify: `internal/write/ids.go` (the `created` write for a new id)
- Modify: `internal/write/scratch.go` (the `created` write for a new scratch item only; the `day` heading stays a day)
- Modify: `internal/board/board.go` (`dateField`, `isDate`, the three field comments on `Item`)
- Modify: `internal/tui/detail.go` (only the `dateLine` comment if it names the shape)
- Test: `internal/write/dates_test.go`, `internal/write/ids_test.go`, `internal/write/scratch_test.go`, any other `internal/write` test that expects a bare day in `created`, `started` or `finished`, and `internal/tui/detail_test.go`

**verify:** No path that writes `created`, `started` or `finished` writes a bare day any more, and no path writes a time for anything else (file names, the scratch day heading). List every write of those three fields you found, with grep for `"created"`, `"started"`, `"finished"` and `2006-01-02` over `internal`, and what each one writes now. `dateField` returns a value only for a real day or a real day with a real time down to the second. List the inputs you tried and what each one gave. The footer keeps every date readable at every width from 5 to 130 when the dates carry times.

**Interfaces:**
- Consumes: `Now` in `internal/write/ops.go`, `SetField`, `hasField`.
- Produces: `const stampLayout = "2006-01-02 15:04:05"` in `internal/write/dates.go`, used by every writer of the three fields. `board.dateField` returns either shape unchanged.

- [x] **Step 1: Write the failing tests**

In `internal/write/dates_test.go`, add:

```go
func TestDatesCarryTheTime(t *testing.T) {
	old := Now
	Now = func() time.Time { return time.Date(2026, 9, 30, 16, 14, 5, 0, time.Local) }
	t.Cleanup(func() { Now = old })

	src := []byte("---\nstatus: draft\n---\n# x\n")
	out, err := MarkStarted(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `started: "2026-09-30 16:14:05"`) {
		t.Fatalf("started has no time:\n%s", out)
	}
	out, err = MarkFinished(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `finished: "2026-09-30 16:14:05"`) {
		t.Fatalf("finished has no time:\n%s", out)
	}
}
```

Match the quoting `SetField` really writes. Read one existing test in `dates_test.go` first and copy its assertion style. Add the `strings` and `time` imports only if the file lacks them.

In `internal/write/ids_test.go` and `internal/write/scratch_test.go`, find the tests that check the `created` value after `fixNow(t)`. `fixNow` pins `2026-09-26 10:00:00 UTC`, so the expected value becomes `2026-09-26 10:00:00`. Every other `internal/write` test that pins a bare day in one of the three fields gets the time too.

In `internal/tui/detail_test.go`, add one item to `datedFiles()` whose frontmatter has `created: "2026-09-01 08:00:00"`, `started: "2026-09-02 09:30:15"` and `finished: "2026-09-03 17:45:59"`. Add it to the case lists of `TestItemDatesFromFrontmatter`, `TestDetailShowsTheDates` and `TestDetailFooterNamesEveryDateAtEveryWidth`, with the long form `created 2026-09-01 08:00:00 · started 2026-09-02 09:30:15 · finished 2026-09-03 17:45:59` and the short form `c 2026-09-01 08:00:00 · s 2026-09-02 09:30:15 · f 2026-09-03 17:45:59`. Also add items whose values must read as "": `"2026-02-30 10:00:00"`, `"2026-09-30 25:00:00"`, `"2026-09-30 16:14"` and `"yesterday"`.

- [x] **Step 2: Run the tests to verify they fail**

Run: `scripts/test ./internal/write ./internal/tui -run 'TestDatesCarryTheTime|TestItemDates|TestDetailShowsTheDates|TestDetailFooterNamesEveryDate|Created|Scratch|ID'`
Expected: FAIL. The writers still write a bare day, and the timed item reads as "".

- [x] **Step 3: Write minimal implementation**

In `internal/write/dates.go`:

```go
// stampLayout is how a date field is written: the day and the time to the
// second, so the reader can tell two changes on one day apart.
const stampLayout = "2006-01-02 15:04:05"
```

In `MarkStarted` and `MarkFinished`, use `Now().Format(stampLayout)`. In `ids.go`, the `created` write uses `Now().Format(stampLayout)`. In `scratch.go`, only the `{"created", ...}` entry uses `stampLayout`. The `day` heading and the file names keep `2006-01-02`.

In `internal/board/board.go`:

```go
// dateField is a date from the frontmatter, or "". It is a bare day
// (YYYY-MM-DD) in old files and a day with a time (YYYY-MM-DD HH:MM:SS) in
// new ones. A bare date comes back from yaml as a time and a quoted one as a
// string, so both are read; anything that is not a real day or time is
// dropped.
```

```go
// isDate says whether the text is a real day, or a real day and time, so a
// date that does not exist never reaches the reader.
func isDate(s string) bool {
	for _, layout := range []string{"2006-01-02", "2006-01-02 15:04:05"} {
		if _, err := time.Parse(layout, s); err == nil {
			return true
		}
	}
	return false
}
```

The `time.Time` case in `dateField` stays as it is. Update the `Created`, `StartedOn` and `Finished` comments on `Item` to say `YYYY-MM-DD or YYYY-MM-DD HH:MM:SS`.

- [x] **Step 4: Run the tests to verify they pass**

Run: `scripts/test ./internal/write ./internal/board ./internal/tui`
Expected: PASS for all three packages.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./internal/write ./internal/board ./internal/tui
git add internal/write internal/board/board.go internal/tui/detail.go internal/tui/detail_test.go
git commit -m "write, board: date fields carry the time to the second"
```

### Task 2: Every toast hides itself

**Files:**
- Modify: `internal/tui/model.go` (`Update`, `copyID`, the `toastFor` comment)
- Modify: `internal/tui/select.go` (`copyPicked`)
- Test: `internal/tui/model_test.go`, `internal/tui/select_test.go`

**verify:** No message that leaves a new, non-empty `m.status` returns without a `clearStatusMsg` timer for that exact text, and no toast ever gets two timers. List every place in `internal/tui` that sets `m.status` (grep `m.status =`) and show that each one goes through `Update`. A toast set after an older one is never cleared by the older timer. Search text never starts a timer.

**Interfaces:**
- Consumes: `clearStatusAfter(d time.Duration, text string) tea.Cmd`, `clearStatusMsg{text string}`, `toastFor`.
- Produces: `Update` is a thin wrapper over the old body, now named `func (m Model) update(msg tea.Msg) (tea.Model, tea.Cmd)`.

- [x] **Step 1: Write the failing tests**

In `internal/tui/model_test.go`, replace `TestCopyErrorsDoNotHide` with a test that checks the opposite, because the user ruled that errors hide too:

```go
// Every message on the status line hides by itself, errors too, so the key
// hints always come back.
func TestEveryToastHidesByItself(t *testing.T) {
	t.Parallel()

	m := actModel(t)
	m.clip = func(string) error { return errors.New("no clipboard") }
	cases := []struct {
		name string
		run  func(Model) (tea.Model, tea.Cmd)
	}{
		{"copy failed", func(m Model) (tea.Model, tea.Cmd) { return m.Update(key("y")) }},
		{"editor error", func(m Model) (tea.Model, tea.Cmd) { return m.Update(editorDoneMsg{err: errors.New("boom")}) }},
		{"tasks refuse the popup", func(m Model) (tea.Model, tea.Cmd) {
			return press(m, tabKey(tabPlans), "l", "j").Update(key("s"))
		}},
	}
	for _, c := range cases {
		next, cmd := c.run(m)
		got := next.(Model)
		if got.status == "" {
			t.Fatalf("%s: no message on the line", c.name)
		}
		if cmd == nil {
			t.Fatalf("%s: %q started no timer, so it would stay", c.name, got.status)
		}
		cleared, _ := got.Update(clearStatusMsg{text: got.status})
		if s := cleared.(Model).status; s != "" {
			t.Fatalf("%s: the toast did not hide: %q", c.name, s)
		}
	}
}
```

Before you write the test, check that `editorDoneMsg` has a field named `err` and that `press(..., tabKey(tabPlans), "l", "j")` stands on a task. `TestHintsOnATaskRowOfferTickNotStatus` in `hints_test.go` uses the same keys. If a name differs, use the real one.

The "nothing selected" half of the old test becomes a check that it also starts a timer.

Add one test for the timer command itself. Run the `cmd` returned for a toast and check that it yields `clearStatusMsg{text: <that status>}`, through `tea.Batch` if needed. The simplest way: call `cmd()`. When the result is a `tea.BatchMsg`, run each inner command and look for the `clearStatusMsg`.

`TestCopyToastHidesByItself`, `TestCopyToastKeepsANewerMessage` and the drag copy test in `select_test.go` stay and must still pass.

- [x] **Step 2: Run the tests to verify they fail**

Run: `scripts/test ./internal/tui -run 'TestEveryToastHidesByItself|TestCopyToast|TestDrag'`
Expected: FAIL on `TestEveryToastHidesByItself`, because no timer is started for errors and refusals.

- [x] **Step 3: Write minimal implementation**

In `internal/tui/model.go`, rename the current `Update` body to `update`, and add:

```go
// Update runs one message. When the message leaves a new text on the status
// line, a timer is started for it here, in one place, so every toast hides
// by itself and the key hints come back.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	before := m.status
	next, cmd := m.update(msg)
	after := next.(Model).status
	if after == "" || after == before {
		return next, cmd
	}
	return next, tea.Batch(cmd, clearStatusAfter(toastFor, after))
}
```

The comment that sits above the old `Update` moves onto `update`. In `copyID`, `return clearStatusAfter(toastFor, m.status)` becomes `return nil`. In `copyPicked` in `select.go`, do the same. Change the `toastFor` comment to say it is how long any status message stays.

Check every `return` in `update` that returns a `Model` value (not a pointer or another type). The wrapper's type assertion `next.(Model)` must hold on every path.

- [x] **Step 4: Run the tests to verify they pass**

Run: `scripts/test ./internal/tui`
Expected: PASS for the whole package.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./internal/tui
git add internal/tui/model.go internal/tui/select.go internal/tui/model_test.go internal/tui/select_test.go
git commit -m "tui: every status toast hides itself after toastFor"
```
