---
id: PLN-0034
hash: pz2odlu
started: "2026-09-29"
---
# Faster Test Suite Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Cut test wall time: sampled size sweeps under `-short`, and `t.Parallel()` where safe.

**Architecture:** One helper in `internal/tui/view_test.go` picks the full size range or a fixed edge list based on `testing.Short()`. Then top-level tests that touch no process-wide state call `t.Parallel()`.

**Tech Stack:** Go stdlib `testing`.

**Spec:** `.acta/specs/2026-09-29-fast-tests-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Test files only. No production code changes.
- No test changes what it checks. Without `-short`, every loop covers the same sizes as today.
- A test stays serial when it (or a helper it calls) uses `os.Chdir`, `t.Chdir`, `t.Setenv`, `os.Setenv`, writes a package-level var, or changes lipgloss or termenv global state (for example `withColors` in `internal/tui`).
- Comments in plain English, short words, say why.
- Before each commit: `gofmt -l .` prints nothing, `go vet ./...` clean.

## Baseline (measured on main 2026-09-29)

- `go test -count=1 ./...`: 1:41 wall. `internal/tui` 94s, `cmd/acta` 28s, `internal/write` 26s, `internal/cli` 22s.
- `TestViewNeverOverflowsAnyWindow` 43s, `TestNoRoundedCorners` 22s, `TestThumbSitsOnTheBorderNotInside` 17s.

## File map

- `internal/tui/view_test.go`: size helper plus the three slow tests (Task 1).
- `*_test.go` across `cmd/acta` and `internal/*`: `t.Parallel()` lines (Task 2).

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2 (touches `view_test.go` too, so it runs after Task 1).

---

### Task 1: Sampled sizes under -short

**Files:**
- Modify: `internal/tui/view_test.go` (`TestViewNeverOverflowsAnyWindow`, `TestNoRoundedCorners`, `TestThumbSitsOnTheBorderNotInside`)

**verify:** Without `-short`, each of the three tests visits exactly the same (width, height, tab, pane, spot) set as before. With `-short`, each still visits every layout edge: widths below 2 and 6 (frame and view guards), 59/60/61 (the `wide` switch in `frame.go:18`), the left column clamp ends (93/94, 159/160/161 from `clamp(m.width*3/10, 28, 48)`), and the largest size. List each loop and the sizes it visits in both modes.

**Interfaces:**
- Produces: `func sweep(lo, hi int, short []int) []int` in `view_test.go`. Returns every int in `[lo, hi]`, or only the values of `short` inside `[lo, hi]` when `testing.Short()`.

- [x] **Step 1: Write the failing test for the helper**

```go
func TestSweepFullRangeWithoutShort(t *testing.T) {
	if testing.Short() {
		t.Skip("checks the full range")
	}
	got := sweep(3, 6, []int{4})
	if !slices.Equal(got, []int{3, 4, 5, 6}) {
		t.Fatalf("sweep(3, 6) = %v, want every value", got)
	}
}

func TestSweepSamplesInsideRangeUnderShort(t *testing.T) {
	if !testing.Short() {
		t.Skip("checks the short list")
	}
	got := sweep(3, 6, []int{1, 4, 6, 9})
	if !slices.Equal(got, []int{4, 6}) {
		t.Fatalf("sweep(3, 6) = %v, want only listed values inside the range", got)
	}
}
```

- [x] **Step 2: Run both modes, see them fail**

Run: `go test ./internal/tui/ -run TestSweep` and `go test -short ./internal/tui/ -run TestSweep`
Expected: build FAIL, `undefined: sweep`.

- [x] **Step 3: Add the helper and edge lists**

```go
// Edge sizes of the layout. Under -short the size sweeps read only these,
// so a daily run is fast. The full run still walks every size.
var (
	edgeWidths  = []int{1, 2, 5, 6, 7, 29, 30, 31, 59, 60, 61, 93, 94, 159, 160, 161, 200}
	edgeHeights = []int{10, 11, 24, 40, 60}
)

func sweep(lo, hi int, short []int) []int {
	var out []int
	if testing.Short() {
		for _, v := range short {
			if v >= lo && v <= hi {
				out = append(out, v)
			}
		}
		return out
	}
	for v := lo; v <= hi; v++ {
		out = append(out, v)
	}
	return out
}
```

- [x] **Step 4: Use it in the slow tests**

`TestViewNeverOverflowsAnyWindow`: `for _, w := range sweep(30, 200, edgeWidths) {` and `for _, h := range sweep(10, 60, edgeHeights) {`.

`TestNoRoundedCorners`: `for _, w := range sweep(1, 200, edgeWidths) {`. Also build `press(newModel(t), tabKey(i)).panes()` once per tab outside the width loop if it does not depend on width (it uses no size), so the list is not rebuilt 200 times.

`TestThumbSitsOnTheBorderNotInside`: keep both sizes without `-short`; under `-short` use only `{80, 30}`:

```go
sizes := [][2]int{{80, 30}, {160, 50}}
if testing.Short() {
	// One size keeps the check on every tab and pane. The full run adds the wide one.
	sizes = sizes[:1]
}
```

- [x] **Step 5: Run both modes, see them pass and time them**

Run: `go test -count=1 ./internal/tui/` and `go test -count=1 -short ./internal/tui/`
Expected: both `ok`. Paste both times. `-short` should be well under the 94s baseline.

- [x] **Step 6: Commit**

```bash
gofmt -l . && go vet ./internal/tui/
git add internal/tui/view_test.go
git commit -m "Sample edge sizes in slow tui sweeps under -short"
```

### Task 2: Run safe tests in parallel

**Files:**
- Modify: `*_test.go` in `cmd/acta`, `internal/board`, `internal/cli`, `internal/config`, `internal/doctor`, `internal/gitc`, `internal/plugincheck`, `internal/trees`, `internal/tui`, `internal/write` (only files with safe tests)

**verify:** No test marked `t.Parallel()` uses, directly or through a helper, any process-wide state named in Global Constraints. List every helper that touches such state and the tests kept serial because of it. `go test -race -short ./...` passes three runs in a row.

**Interfaces:**
- Consumes: `sweep` from Task 1 (unchanged).

- [x] **Step 1: Find process-wide state**

Run: `grep -rnE 'os\.Chdir|t\.Chdir|t\.Setenv|os\.Setenv|withColors|lipgloss\.Set|termenv' --include='*_test.go' .`
Also grep each package's test helpers for writes to package-level vars. Write the list in the task report.

- [x] **Step 2: Add `t.Parallel()`**

First line of each top-level test that is safe:

```go
func TestSomething(t *testing.T) {
	t.Parallel()
	...
}
```

Start with the slow packages: `cmd/acta`, `internal/write`, `internal/cli`, `internal/tui`, `internal/board`, `internal/gitc`, `internal/trees`. Skip packages under 3s.

- [x] **Step 3: Race check**

Run: `go test -count=1 -race -short ./...` three times.
Expected: all `ok`, no `DATA RACE`. A race or flaky failure means a test shares state: take `t.Parallel()` out of that test and name it in the report.

- [x] **Step 4: Full run and timings**

Run: `time go test -count=1 ./...` and `time go test -count=1 -short ./...`
Expected: both `ok`. Paste both wall times against the 1:41 baseline.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./...
git add -u
git commit -m "Run tests without shared state in parallel"
```
