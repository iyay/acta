---
parent: specs/2026-10-01-tui-priority-key-design
id: PLN-0066
created: "2026-10-01 06:17:35"
hash: t2qbabk
---
# TUI Key p Sets Priority Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** In the TUI, `p` opens a priority popup on a bug or a debt line, and the hint `Priority: p` shows only on those rows.

**Architecture:** Reuse the status popup. `openPopup` gets a `p` case that builds `popup{field: "priority", options: high, medium, low, none}`; `popupKey` already writes through `m.setValue(id, field, value)`, which is `write.SetValue`, which already handles `priority` (PLN-0064). `hints()` adds one hint; `helpLines` adds one line.

**Tech Stack:** Go, Bubble Tea.

**Spec:** `.acta/specs/2026-10-01-tui-priority-key-design.md`

**Tests:** fast `scripts/test ./internal/tui`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Options are exactly `high`, `medium`, `low`, `none`, in that order. The cursor starts on the current value, or on `none` when unset.
- `p` works only on a bug or a debt item. A wrong kind sets the status line to exactly `p sets the priority of a bug or a debt line` and opens no popup. A row from another worktree and a legacy file are refused with the same messages `s` uses today.
- The hint text is exactly `Priority: p`. The help line is exactly `p                set the priority of a bug or a debt line: high, medium, low or none` (key column padded like its neighbours).
- Comments are plain English a 10-year-old can read. They say why, not what.

## File map

- `internal/tui/model.go`: `p` in the key switch; priority case in `openPopup`.
- `internal/tui/hints.go`: `Priority: p` in both the list and the detail branch.
- `internal/tui/view.go`: one line in `helpLines`, after the `t` line.
- `internal/tui/priority_key_test.go` (new).

## Waves

- Wave 1: Task 1.

---

### Task 1: Key p, its popup, its hint and its help line

**Files:**
- Modify: `internal/tui/model.go` (the `case "t", "s":` key line; `openPopup`)
- Modify: `internal/tui/hints.go` (`hints`)
- Modify: `internal/tui/view.go` (`helpLines`)
- Test: `internal/tui/priority_key_test.go`

**verify:** `p` opens the popup on every bug and debt item of this checkout and on nothing else, and the hint `Priority: p` shows on exactly the rows where `p` opens the popup, in the list panes and the detail pane. No other key, popup or hint changes. List every row kind you checked (scratch, bug, debt item, spec, plan, task, group row, worktree row, legacy row) and what `p` and the hints do on each.

**Interfaces:**
- Consumes: `prioModel(t)` and `prioCfg(t)` in `internal/tui/priority_test.go` (bugs BUG-0001..0007, debt DBT-0001 items; Bugs open list order c, d, e, b, a); `newModel(t)`, `press`, `tabKey`, `sized` from the existing tests; `board.Item.Priority`.
- Produces: nothing other tasks use.

- [ ] **Step 1: Write the failing test**

Create `internal/tui/priority_key_test.go`:

```go
package tui

import (
	"slices"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/write"
)

const prioOptions = "high,medium,low,none"

func TestPriorityPopupOnABugStartsOnItsValue(t *testing.T) {
	t.Parallel()

	// The first Bugs row is BUG-0003, set to high.
	m := press(prioModel(t), tabKey(tabBugs), "p")
	if m.popup == nil || m.popup.field != "priority" || strings.Join(m.popup.options, ",") != prioOptions || m.popup.idx != 0 {
		t.Fatalf("popup on a high bug = %+v", m.popup)
	}
	// The last Bugs row is BUG-0001, with no priority, so the cursor sits on none.
	m = press(prioModel(t), tabKey(tabBugs), "G", "p")
	if m.popup == nil || m.popup.idx != 3 {
		t.Fatalf("popup on an unset bug = %+v", m.popup)
	}
}

func TestPriorityPopupOnADebtLine(t *testing.T) {
	t.Parallel()

	m := press(prioModel(t), tabKey(tabDebts), "p")
	if it := m.Selected(); it == nil || it.Kind != board.KindDebtItem {
		t.Fatalf("selected %+v, want a debt item", it)
	}
	if m.popup == nil || m.popup.field != "priority" || strings.Join(m.popup.options, ",") != prioOptions {
		t.Fatalf("popup on a debt line = %+v", m.popup)
	}
}

func TestPriorityPopupWritesTheChoice(t *testing.T) {
	t.Parallel()

	m := press(prioModel(t), tabKey(tabBugs), "G")
	var got []string
	m.setValue = func(id, field, value string) (write.Outcome, error) {
		got = []string{id, field, value}
		return write.Outcome{Committed: true}, nil
	}
	m = press(m, "p", "k", "k", "k", "enter")
	if strings.Join(got, " ") != "bugs/2026-10-01-a priority high" {
		t.Fatalf("setValue got %v", got)
	}
	if m.popup != nil || !strings.Contains(m.status, "committed") {
		t.Fatalf("popup %v status %q", m.popup, m.status)
	}
}

func TestPriorityKeyRefusesOtherKinds(t *testing.T) {
	t.Parallel()

	const want = "p sets the priority of a bug or a debt line"
	for _, c := range []struct {
		name string
		keys []string
		kind board.Kind
	}{
		{"scratch", []string{tabKey(tabScratches)}, board.KindScratch},
		{"spec", []string{tabKey(tabSpecs)}, board.KindStory},
		{"plan", []string{tabKey(tabPlans)}, board.KindPlan},
		{"task", []string{tabKey(tabPlans), " ", "j"}, board.KindTask},
	} {
		m := press(newModel(t), c.keys...)
		if it := m.Selected(); it == nil || it.Kind != c.kind {
			t.Fatalf("%s: selected %+v, want kind %s", c.name, it, c.kind)
		}
		m = press(m, "p")
		if m.popup != nil || m.status != want {
			t.Errorf("%s: popup %+v status %q, want no popup and %q", c.name, m.popup, m.status, want)
		}
		if slices.Contains(m.hints(), "Priority: p") {
			t.Errorf("%s: hints offer Priority: p: %v", c.name, m.hints())
		}
	}
}

func TestPriorityHintOnBugsAndDebtLines(t *testing.T) {
	t.Parallel()

	for _, keys := range [][]string{
		{tabKey(tabBugs)},
		{tabKey(tabBugs), "enter"},
		{tabKey(tabDebts)},
		{tabKey(tabDebts), "enter"},
	} {
		m := press(sized(prioModel(t), 200, 40), keys...)
		if !slices.Contains(m.hints(), "Priority: p") {
			t.Errorf("keys %v: hints %v, want Priority: p", keys, m.hints())
		}
	}
}

func TestHelpListsThePriorityKey(t *testing.T) {
	t.Parallel()

	if !strings.Contains(helpLines, "set the priority of a bug or a debt line: high, medium, low or none") {
		t.Fatalf("help does not list p: %q", helpLines)
	}
}
```

If a fixture row in `TestPriorityKeyRefusesOtherKinds` is not the kind named (for example the Scratches tab of `newModel` is empty), move the cursor with `j` to a row of that kind; never drop the case.

- [ ] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/tui -run 'Priority(Popup|Key|Hint)|HelpListsThePriorityKey'`
Expected: FAIL: no popup on `p`, no hint, no help line.

- [ ] **Step 3: Write minimal implementation**

`internal/tui/model.go`, key switch: change `case "t", "s":` to `case "t", "s", "p":`.

`internal/tui/model.go`, `openPopup`: add a first case to the refusal `switch`, right after `case it == nil:` and its return:

```go
	case key == "p" && it.Kind != board.KindBug && it.Kind != board.KindDebtItem:
		m.status = "p sets the priority of a bug or a debt line"
		return
```

This case must come before `case it.Kind == board.KindTask:`, so a task gets the priority message and not the checkbox one. After the `if key == "t" { ... }` block:

```go
	if key == "p" {
		p = popup{field: "priority", options: append(append([]string(nil), board.Priorities...), "none")}
		current = it.Priority
		if current == "" {
			current = "none"
		}
	}
```

`internal/tui/hints.go`, `hints`: after the `own` line, add:

```go
	// Only bugs and debt lines carry a priority, so only they offer p.
	prio := own && (it.Kind == board.KindBug || it.Kind == board.KindDebtItem)
```

In the detail branch, right after the `if own && !tick { ... "Status: s" ... }` block:

```go
		if prio {
			out = append(out, "Priority: p")
		}
```

In the list branch, right after `if own && !tick { out = append(out, "Status: s", "Type: t") }`:

```go
	if prio {
		out = append(out, "Priority: p")
	}
```

`internal/tui/view.go`, `helpLines`: right after the `t                set the type: spec or bug` line:

```
p                set the priority of a bug or a debt line: high, medium, low or none
```

- [ ] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/tui`
Expected: PASS (whole package, so the existing hint, help layout and popup tests still hold).

- [ ] **Step 5: Commit**

```bash
gofmt -l internal/tui && go vet ./internal/tui
git add internal/tui/model.go internal/tui/hints.go internal/tui/view.go internal/tui/priority_key_test.go
git commit -m "TUI key p sets the priority of a bug or a debt line"
```
