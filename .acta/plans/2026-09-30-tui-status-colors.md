---
created: "2026-09-30"
id: PLN-0047
hash: ylyfrwe
started: "2026-09-30"
finished: "2026-09-30"
---
# TUI Status Colors Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Done work shows green, the dot of work under way pulses green, the text of work under way is the foreground color, and the `?` popup shows its keys in cyan in a right-aligned column.

**Architecture:** `newStyles` builds eight pulse brushes per theme, each keeping a contrast of at least 1.6 against the background. Every `●` is first drawn with frame 0. `frameFrom` then swaps it for the current frame and notes in the frame cache whether one was on screen. A `pulseMsg` tick, armed by a resize or a reload while the board holds work under way, moves the frame on. It draws again only when a dot was on screen. Rows and lines for done work use the existing `done` brush (green, slot 2). The `work` brush goes.

**Tech Stack:** Go, Bubble Tea v1.3.10, lipgloss (both already in go.mod).

**Spec:** `.acta/specs/2026-09-30-tui-status-colors-design.md`

**Tests:** fast `scripts/test ./internal/tui/`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Every color is a theme slot or a mix of the theme's own hex. There is no new hex.
- The pulse has exactly 8 frames. Frame 0 is green (slot 2). The frames go toward the background and back: steps 0,1,2,3,4,3,2,1. The pulse moves one frame every 120 ms.
- In every built-in hex theme, no pulse frame has a contrast below 1.6 against the theme background. The `terminal` theme uses frames 0-3 = slot 2 and frames 4-7 = slot 10.
- The tick never runs while the board has no work under way. A tick with no dot on screen changes nothing and sets `m.same`.
- The row under the cursor keeps its selection band, and its dot does not pulse.
- Done means `dotOf(it) == dotDone` for rows and work lines, and a ticked box for step lines. A done title and mark are green (`styles.done`). The id keeps its kind color.
- No new dependency in go.mod.
- Old tests that pin the old look are updated in the task that changes that look: the `work` color, the accent dot, the accent step line, the old help layout. Never delete a test about anything else, never loosen an assert about anything else, and list every test you changed in the task report.
- Tests use no sleep. A tick is tested by sending its message to `Update`. Never call a `tea.Tick` command with a real delay.
- Comments use plain English that a 10-year-old can read back: short words, and they say why.
- Before each commit, run `gofmt -l internal` (it must print nothing), `go vet ./internal/tui/` and `scripts/test ./internal/tui/`.

## File Map

- `internal/tui/styles.go`:
  - Task 1: `pulse`, `goingDot`, `pulseBrushes`, `mixHex`, `hexRGB`, `contrastRatio`, `farthestMix`, `slotBrightGreen`.
  - Task 6: `work` and `slotWork` removed.
- `internal/tui/detail.go`:
  - Task 1: `styles.dot` gives `pulse[0]` for `dotGoing`.
  - Task 3: `workLine` and `stepLines` use green for done and the foreground for under way.
- `internal/tui/scroll.go` (Task 2): `listView` paints done rows green and swaps the tree dot for `goingDot`.
- `internal/tui/model.go` (Task 4): `pulseMsg`, `pulseStep`, `pulseAfter`, `hasWork`, `armPulse`, the `pulse`/`pulsing` fields, `frameCache.dots`, and the resize and reload arming.
- `internal/tui/view.go`:
  - Task 4: `pulseDots`, and `frameFrom` calls it.
  - Task 5: `helpText` lays out the key column.
- Tests:
  - `internal/tui/styles_test.go`: Tasks 1 and 6.
  - `internal/tui/detail_test.go`: Tasks 1 and 3.
  - `internal/tui/scroll_test.go`: Task 2.
  - `internal/tui/model_test.go`: Task 4.
  - `internal/tui/view_test.go`: Tasks 4 and 5.

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2, Task 3 and Task 4, in one message. Their files do not overlap:
  - Task 2: `scroll.go`, `scroll_test.go`.
  - Task 3: `detail.go`, `detail_test.go`.
  - Task 4: `model.go`, `view.go`, `model_test.go`, `view_test.go`.
- Wave 3: Task 5 and Task 6, in one message. Task 5 has `view.go` and `view_test.go`. Task 6 has `styles.go` and `styles_test.go`. Task 6 must come after Tasks 2 and 3, because they stop using `work`.

---

### Task 1: Pulse brushes

**Files:**
- Modify: `internal/tui/styles.go` (`styles`, the slot constants, `newStyles`, new helpers)
- Modify: `internal/tui/detail.go` (`styles.dot` only)
- Test: `internal/tui/styles_test.go`, `internal/tui/detail_test.go` (the old accent-dot assert only)

**verify:** In every built-in theme there are exactly 8 pulse brushes. Frame 0 is slot 2. Frames `i` and `8-i` match for i = 1..3. In hex themes the contrast against the background never rises from frame 0 to frame 4, and no frame is below 1.6. In the `terminal` theme, frames 0-3 are ANSI `2` and frames 4-7 are ANSI `10`, with no 24-bit code. `styles.dot(dotGoing)` is frame 0, and `goingDot` is frame 0 drawing `●`. A theme whose green is already under 1.6 gets 8 green frames. List every theme checked.

**Interfaces:**
- Consumes: `slot` (inside `newStyles`), `slotGreen`, `dotGoing` (`detail.go`), and the test helper `contrast(t, a, b)` in `styles_test.go`.
- Produces:
  - `const pulseSteps = 8`
  - `const minPulseContrast = 1.6`
  - `const slotBrightGreen = 10`
  - `styles` fields `pulse []lipgloss.Style` and `goingDot string`
  - `func pulseBrushes(t theme.Theme, slot func(int) lipgloss.Color) []lipgloss.Style`
  - `func mixHex(a, b string, f float64) string`
  - `func hexRGB(s string) ([3]int, bool)`
  - `func contrastRatio(a, b string) float64`
  - `func farthestMix(fg, bg string, floor float64) float64`

- [x] **Step 1: Write the failing test**

Add to `internal/tui/styles_test.go`:

```go
// TestPulseFramesStayReadable walks every built-in theme. A frame too close
// to the background would trip the terminal's minimum contrast and flash
// white, so none may drop under 1.6.
func TestPulseFramesStayReadable(t *testing.T) {
	t.Parallel()

	for _, name := range theme.Names() {
		th, _ := theme.Builtin(name)
		s := newStyles(th, true)
		if len(s.pulse) != pulseSteps {
			t.Fatalf("%s: %d pulse frames, want %d", name, len(s.pulse), pulseSteps)
		}
		fg := func(i int) string { return string(s.pulse[i].GetForeground().(lipgloss.Color)) }
		if th.BG == "" {
			for i := range s.pulse {
				want := "2"
				if i >= pulseSteps/2 {
					want = "10"
				}
				if fg(i) != want {
					t.Errorf("%s frame %d = %s, want %s", name, i, fg(i), want)
				}
			}
			continue
		}
		if fg(0) != th.ANSI[slotGreen] {
			t.Errorf("%s frame 0 = %s, want green %s", name, fg(0), th.ANSI[slotGreen])
		}
		for i := 1; i < pulseSteps/2; i++ {
			if fg(i) != fg(pulseSteps-i) {
				t.Errorf("%s: frame %d %s and frame %d %s differ, the pulse is not symmetric", name, i, fg(i), pulseSteps-i, fg(pulseSteps-i))
			}
		}
		prev := contrast(t, fg(0), th.BG)
		for i := range s.pulse {
			c := contrast(t, fg(i), th.BG)
			if c < minPulseContrast {
				t.Errorf("%s frame %d %s has contrast %.2f, want at least %.1f", name, i, fg(i), c, minPulseContrast)
			}
			if i > 0 && i <= pulseSteps/2 && c > prev+1e-9 {
				t.Errorf("%s frame %d gets brighter on the way down: %.3f after %.3f", name, i, c, prev)
			}
			prev = c
		}
	}
}

func TestTokyoNightPulseMoves(t *testing.T) {
	t.Parallel()

	th, _ := theme.Builtin("tokyo-night")
	s := newStyles(th, true)
	if s.pulse[0].GetForeground() == s.pulse[pulseSteps/2].GetForeground() {
		t.Fatal("frame 4 is still green, so the dot would not pulse")
	}
}

func TestGoingDotIsFrameZero(t *testing.T) {
	withTrueColor(func() {
		for _, name := range []string{"tokyo-night", "terminal"} {
			th, _ := theme.Builtin(name)
			s := newStyles(th, true)
			if s.dot(dotGoing).GetForeground() != s.pulse[0].GetForeground() {
				t.Errorf("%s: dot of work under way is %v, want frame 0 %v", name, s.dot(dotGoing).GetForeground(), s.pulse[0].GetForeground())
			}
			if s.goingDot != s.pulse[0].Render(dotGoing) {
				t.Errorf("%s: goingDot %q, want %q", name, s.goingDot, s.pulse[0].Render(dotGoing))
			}
		}
	})
}

func TestFarthestMixKeepsTheFloor(t *testing.T) {
	t.Parallel()

	if f := farthestMix("#1a1b26", "#1a1b26", 1.6); f != 0 {
		t.Errorf("a green already on the floor mixed %.2f, want 0", f)
	}
	f := farthestMix("#9ece6a", "#1a1b26", 1.6)
	if f <= 0 || f >= 1 {
		t.Fatalf("farthestMix = %.2f, want between 0 and 1", f)
	}
	if c := contrastRatio(mixHex("#9ece6a", "#1a1b26", f), "#1a1b26"); c < 1.6 {
		t.Errorf("the farthest mix has contrast %.2f, want at least 1.6", c)
	}
	if got := mixHex("#000000", "#ffffff", 0.5); got != "#808080" && got != "#7f7f7f" {
		t.Errorf("mixHex halfway = %s", got)
	}
	if got := mixHex("nope", "#ffffff", 0.5); got != "nope" {
		t.Errorf("mixHex on a bad color = %s, want it back as it came", got)
	}
}
```

In `internal/tui/detail_test.go`, change the assert in `TestDetailDotsProblemsAndWorkLines` that expects `s.dot(dotGoing)` to be the accent, so that it expects `s.pulse[0]`.

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/tui/ -run 'TestPulseFramesStayReadable|TestTokyoNightPulseMoves|TestGoingDotIsFrameZero|TestFarthestMixKeepsTheFloor|TestDetailDotsProblemsAndWorkLines'`
Expected: FAIL to build with `s.pulse undefined`.

- [x] **Step 3: Write minimal implementation**

In `internal/tui/styles.go`, add `math` to the imports and add below the slot constants:

```go
	slotBrightGreen = 10
)

// The pulse of a dot whose work is under way: eight frames, from green toward
// the background and back. Its darkest frame still keeps this much contrast,
// or a terminal with a minimum contrast would draw it white.
const (
	pulseSteps       = 8
	minPulseContrast = 1.6
)
```

The existing `const (` block closes with `slotCyan = 6`. Put `slotBrightGreen = 10` inside that block, then open the new block.

Add to the `styles` struct:

```go
	pulse    []lipgloss.Style // the frames of the dot of work under way
	goingDot string           // that dot drawn in frame 0, as every view draws it
```

In `newStyles`, before the `return`:

```go
	pulse := pulseBrushes(t, slot)
```

In the returned struct:

```go
		pulse:    pulse,
		goingDot: pulse[0].Render(dotGoing),
```

Add after `newStyles`:

```go
// pulseBrushes gives the frames of the pulse. A theme with its own colors
// fades green toward its background, but only as far as keeps it readable.
// The terminal theme has no hex to fade, so it swaps green for bright green.
func pulseBrushes(t theme.Theme, slot func(int) lipgloss.Color) []lipgloss.Style {
	out := make([]lipgloss.Style, pulseSteps)
	if t.BG == "" {
		for i := range out {
			c := slotGreen
			if i >= pulseSteps/2 {
				c = slotBrightGreen
			}
			out[i] = lipgloss.NewStyle().Foreground(slot(c))
		}
		return out
	}
	green := t.ANSI[slotGreen]
	far := farthestMix(green, t.BG, minPulseContrast)
	half := pulseSteps / 2
	for i := range out {
		step := i
		if step > half {
			step = pulseSteps - i
		}
		out[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(mixHex(green, t.BG, far*float64(step)/float64(half))))
	}
	return out
}

// farthestMix is how far fg can move toward bg, from 0 to 1, and still keep
// the floor contrast. A color already under the floor cannot move at all.
func farthestMix(fg, bg string, floor float64) float64 {
	best := 0.0
	for f := 0.01; f <= 1; f += 0.01 {
		if contrastRatio(mixHex(fg, bg, f), bg) < floor {
			break
		}
		best = f
	}
	return best
}

// mixHex moves the #rrggbb color a the part f of the way toward b. When
// either one is not a #rrggbb color there is nothing to mix, so a comes back.
func mixHex(a, b string, f float64) string {
	ca, okA := hexRGB(a)
	cb, okB := hexRGB(b)
	if !okA || !okB {
		return a
	}
	var out [3]int
	for i := range out {
		out[i] = int(math.Round(float64(ca[i]) + (float64(cb[i])-float64(ca[i]))*f))
	}
	return fmt.Sprintf("#%02x%02x%02x", out[0], out[1], out[2])
}

// hexRGB reads #rrggbb into its red, green and blue parts.
func hexRGB(s string) ([3]int, bool) {
	s, ok := strings.CutPrefix(s, "#")
	if !ok || len(s) != 6 {
		return [3]int{}, false
	}
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return [3]int{}, false
	}
	return [3]int{int(n >> 16 & 0xff), int(n >> 8 & 0xff), int(n & 0xff)}, true
}

// contrastRatio is the WCAG contrast of two #rrggbb colors, the same measure
// terminals use for their minimum contrast. A color that cannot be read
// gives 1, the lowest there is.
func contrastRatio(a, b string) float64 {
	lum := func(s string) (float64, bool) {
		c, ok := hexRGB(s)
		if !ok {
			return 0, false
		}
		var out float64
		for i, w := range []float64{0.2126, 0.7152, 0.0722} {
			v := float64(c[i]) / 255
			if v <= 0.03928 {
				v /= 12.92
			} else {
				v = math.Pow((v+0.055)/1.055, 2.4)
			}
			out += w * v
		}
		return out, true
	}
	la, okA := lum(a)
	lb, okB := lum(b)
	if !okA || !okB {
		return 1
	}
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}
```

Add `fmt` to the imports if it is gone. `strconv` and `strings` are already there.

In `internal/tui/detail.go`, `styles.dot`, the `dotGoing` case becomes `return s.pulse[0]`. Change its comment: "green when done, the first frame of the pulse while the work is under way, grey while it waits."

- [x] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/tui/`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l internal && go vet ./internal/tui/ && scripts/test ./internal/tui/
git add internal/tui/styles.go internal/tui/detail.go internal/tui/styles_test.go internal/tui/detail_test.go
git commit -m "tui: add pulse frames that stay readable in every theme"
```

---

### Task 2: List rows show done in green and pulse the tree dot

**Files:**
- Modify: `internal/tui/scroll.go` (`listView`)
- Test: `internal/tui/scroll_test.go`

**verify:** In every list pane:
- A row whose item is done (`dotOf(it) == dotDone`) has its title and mark in green, and its id in its kind color.
- A row whose work is under way is never blue and never faint: its text is the foreground color.
- A tree task row with work under way draws its `●` as `goingDot`.
- A row that is not done carries no green except in its id.
- The row under the cursor is the selection band, with no `goingDot` in it.
- A group row and a row with no item still draw without a panic.

List every row type checked: plain, tree plan, tree task, Done pane, group, cursor.

**Interfaces:**
- Consumes: `styles.done`, `styles.goingDot` (Task 1); `dotOf`, `dotDone`, `dotGoing` (`detail.go`); test helpers `listRow`, `listRows`, `sgr`, `sgrHas`, `withTrueColor`, `plain`, `treeCfg`, `actModel`, `press`, `tabKey`, `sized`.
- Produces: nothing new.

- [x] **Step 1: Write the failing test**

Replace `TestListRowWithWorkKeepsTheWorkColor` in `internal/tui/scroll_test.go` with the tests below. The old test pins the `work` color, which this task removes.

```go
// A row whose work is under way reads in the plain foreground, and its tree
// dot is the pulse dot, so the view can make it pulse.
func TestListRowWithWorkIsPlainWithAPulseDot(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true)
		ln := listRow(t, m, "plans/2026-09-23-q#task-1") // PLN-0004.01, under way, not the cursor row
		if blue := sgr.FindString(m.styles.accent.Render("x")); strings.Contains(ln, blue) {
			t.Errorf("work row is still blue: %q", ln)
		}
		if sgrHas(ln, "2") {
			t.Errorf("work row is faint: %q", ln)
		}
		if !strings.Contains(ln, m.styles.goingDot) {
			t.Errorf("work row has no pulse dot: %q", ln)
		}
		if want := m.styles.kind(board.KindTask).Render("PLN-0004.01"); !strings.Contains(ln, want) {
			t.Errorf("work row id is not in its kind color: %q", ln)
		}
	})
}

func TestDoneListRowsAreGreen(t *testing.T) {
	withTrueColor(func() {
		cfg := treeCfg(t, map[string]string{
			".acta/plans/2026-09-25-d.md": "---\nid: PLN-0009\n---\n# Plan D\n\n### Task 1: Finished\n- [x] a\n\n### Task 2: Open\n- [ ] b\n",
		})
		b, err := board.Load(cfg)
		if err != nil {
			t.Fatal(err)
		}
		m := press(sized(New(cfg, b, true).WithTheme("tokyo-night", true), 160, 40), tabKey(tabPlans))
		m.setOpen("plans/2026-09-25-d", true)
		rows := listRows(m)
		done, ok := rows["plans/2026-09-25-d#task-1"]
		if !ok {
			t.Fatalf("no row for the done task in %q", plainLines(listLines(m)))
		}
		green := sgr.FindString(m.styles.done.Render("x"))
		if !strings.Contains(done, green+"✓") && !strings.Contains(done, m.styles.done.Render("✓")) {
			t.Errorf("done tick is not green: %q", done)
		}
		if !strings.Contains(done, green) || !strings.Contains(plain(done), "Finished") {
			t.Errorf("done title is not green: %q", done)
		}
		if want := m.styles.kind(board.KindTask).Render("PLN-0009.01"); !strings.Contains(done, want) {
			t.Errorf("done row id is not in its kind color: %q", done)
		}
		open := rows["plans/2026-09-25-d#task-2"]
		id := m.styles.kind(board.KindTask).Render("PLN-0009.02")
		if strings.Contains(strings.Replace(open, id, "", 1), green) {
			t.Errorf("a row that is not done wears green: %q", open)
		}
	})
}

func TestCursorRowKeepsItsBandWithoutAPulseDot(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true)
		const work = "plans/2026-09-23-q#task-1" // PLN-0004.01, under way
		for i := 0; i < 10 && (m.Selected() == nil || m.Selected().ID != work); i++ {
			m = press(m, "j")
		}
		if m.Selected() == nil || m.Selected().ID != work {
			t.Fatalf("the cursor never reached %s; rows are %q", work, rowIDs(m))
		}
		ln := listRow(t, m, work)
		if strings.Contains(ln, m.styles.goingDot) {
			t.Errorf("the cursor row pulses its dot inside the band: %q", ln)
		}
		if !sgrHas(ln, "1") {
			t.Errorf("the cursor row is not bold: %q", ln)
		}
	})
}
```

If the fixture paths or ids above differ from what `board.Load` gives, print `rowIDs(m)` and adjust the keys in the test, not the fixture meaning.

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/tui/ -run 'TestListRowWithWorkIsPlainWithAPulseDot|TestDoneListRowsAreGreen|TestCursorRowKeepsItsBandWithoutAPulseDot'`
Expected: FAIL. The work row is blue and has no pulse dot, and the done row is not green.

- [x] **Step 3: Write minimal implementation**

In `listView` (`internal/tui/scroll.go`), replace the part after the selected-row `continue`:

```go
		base := lipgloss.NewStyle()
		if it != nil && dotOf(it) == dotDone {
			// Done work reads green, so the eye skips it.
			base = m.styles.done
		}
		line := m.paintID(text, it, base)
		if r := rows[n]; r.tree && r.depth > 0 && it != nil && dotOf(it) == dotGoing {
			// The tree dot of work under way is the pulse dot, so the view can
			// swap it for the frame of the moment. The mark comes before the
			// title, so the first dot on the line is the mark.
			line = strings.Replace(line, dotGoing, m.styles.goingDot, 1)
		}
		out = append(out, line)
```

Remove the old `inProgress(it)` → `m.styles.work` branch. Update the comment on `listView` if it names the work color.

- [x] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/tui/`
Expected: PASS. Old list tests that pinned the work color are updated here and listed in the report.

- [x] **Step 5: Commit**

```bash
gofmt -l internal && go vet ./internal/tui/ && scripts/test ./internal/tui/
git add internal/tui/scroll.go internal/tui/scroll_test.go
git commit -m "tui: list rows show done in green and pulse the dot of work under way"
```

---

### Task 3: Detail lines show done in green and pulse the dot

**Files:**
- Modify: `internal/tui/detail.go` (`workLine`, `stepLines`)
- Test: `internal/tui/detail_test.go`

**verify:** On every detail work line and step line:
- A done item or a ticked step has its title and mark in green, and a work line keeps its id in its kind color.
- Work under way has its text in the foreground color, never the accent and never blue, and its dot drawn as `goingDot`.
- A waiting line stays grey-dotted and plain.
- The line the reader is on is bold, and still green when it is done.

This holds for every caller: `scratchLines`, `taskLines` (the plan tasks, and the tasks under a spec or a bug through `planLines`), `debtLines`, and `stepLines`. List every caller checked.

**Interfaces:**
- Consumes: `styles.done`, `styles.goingDot`, `styles.dot` (Task 1), `styles.paintID`, `sgr`, `sgrHas`, `withTrueColor`, `actModel`.
- Produces: nothing new.

- [x] **Step 1: Write the failing test**

Add to `internal/tui/detail_test.go`:

```go
func TestDetailWorkLinesShowDoneAndGoing(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true)
		s := m.styles
		green := sgr.FindString(s.done.Render("x"))
		done := &board.Item{ShortID: "PLN-0003.01", Kind: board.KindTask, Title: "Finished", Status: "done"}
		ln := workLine(s, done, false, 80)
		if !strings.HasPrefix(ln, s.done.Render(dotDone)) {
			t.Errorf("done mark is not green: %q", ln)
		}
		if !strings.Contains(ln, s.kind(board.KindTask).Render("PLN-0003.01")) {
			t.Errorf("done id lost its kind color: %q", ln)
		}
		if !strings.Contains(ln[strings.Index(ln, "PLN-0003.01"):], green) {
			t.Errorf("done title is not green: %q", ln)
		}
		if on := workLine(s, done, true, 80); !sgrHas(on, "1") || !strings.Contains(on, green) {
			t.Errorf("the done line the reader is on is not bold green: %q", on)
		}

		going := m.board.Get("plans/2026-09-23-q#task-1") // PLN-0004.01, under way
		gl := workLine(s, going, false, 80)
		if !strings.HasPrefix(gl, s.goingDot) {
			t.Errorf("work under way has no pulse dot: %q", gl)
		}
		rest := strings.TrimPrefix(gl, s.goingDot)
		if strings.Contains(rest, sgr.FindString(s.accent.Render("x"))) {
			t.Errorf("work under way text is still blue: %q", gl)
		}
		if strings.Contains(strings.Replace(rest, s.kind(board.KindTask).Render(going.ShortID), "", 1), green) {
			t.Errorf("work under way text is green: %q", gl)
		}
	})
}

func TestDetailStepLinesShowDoneAndGoing(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true)
		s := m.styles
		task := m.board.Get("plans/2026-09-23-q#task-1") // steps: [x] f, [ ] g, under way
		lines := m.stepLines(task, 60)
		if len(lines) != 2 {
			t.Fatalf("got %d step lines, want 2: %q", len(lines), lines)
		}
		if want := s.done.Render(dotDone); !strings.HasPrefix(lines[0], want) {
			t.Errorf("ticked step mark is not green: %q", lines[0])
		}
		if !strings.Contains(lines[0], sgr.FindString(s.done.Render("x"))+" f") && !strings.Contains(lines[0], s.done.Render(" f")) {
			t.Errorf("ticked step text is not green: %q", lines[0])
		}
		if !strings.HasPrefix(lines[1], s.goingDot) {
			t.Errorf("the step under way has no pulse dot: %q", lines[1])
		}
		if strings.Contains(lines[1], sgr.FindString(s.accent.Render("x"))+" g") {
			t.Errorf("the step under way is still in the accent: %q", lines[1])
		}
	})
}
```

If the fixture ids or step texts differ from what `actModel` loads, print them and fix the test keys, not the fixture.

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/tui/ -run 'TestDetailWorkLinesShowDoneAndGoing|TestDetailStepLinesShowDoneAndGoing'`
Expected: FAIL. Done text is plain, and work under way uses the `work` or accent color.

- [x] **Step 3: Write minimal implementation**

In `workLine` (`internal/tui/detail.go`), the start becomes:

```go
	mark, brush := dotOf(it), lipgloss.NewStyle()
	if mark == dotDone {
		// Done work reads green, so the eye skips it.
		brush = s.done
	}
	if on {
		brush = brush.Bold(true)
	}
```

Remove the `if mark == dotGoing { brush = s.work }` branch. The end already draws the dot with `s.dot(mark)`, which gives `goingDot` for work under way (Task 1). Update the comment above `workLine`: done lines are green, the line the reader is on is bold, and the dot wears its state.

In `stepLines`, the switch becomes:

```go
			mark, brush := dotWaiting, lipgloss.NewStyle()
			switch {
			case s.State != ' ':
				mark, brush = dotDone, m.styles.done
			case going:
				mark, going = dotGoing, false
			}
```

- [x] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/tui/`
Expected: PASS. Old detail tests that pinned the accent step or the work color are updated here and listed in the report.

- [x] **Step 5: Commit**

```bash
gofmt -l internal && go vet ./internal/tui/ && scripts/test ./internal/tui/
git add internal/tui/detail.go internal/tui/detail_test.go
git commit -m "tui: detail lines show done in green and pulse the dot of work under way"
```

---

### Task 4: The pulse tick

**Files:**
- Modify: `internal/tui/model.go` (`frameCache`, `Model`, `Update`: the `tea.WindowSizeMsg` and `reloadMsg` cases and a new `pulseMsg` case)
- Modify: `internal/tui/view.go` (`frameFrom`, new `pulseDots`)
- Test: `internal/tui/model_test.go`, `internal/tui/view_test.go`

**verify:**
- A pulse tick moves the frame on and draws again only when the board has work under way and the last frame showed a pulse dot.
- With work but no dot on screen, it re-arms and sets `m.same`, so nothing is drawn.
- With no work, it arms nothing and the chain stops.
- A resize or a reload arms the tick once when there is work and no tick is in flight, never twice, and never when there is no work. A resize still clears the screen.
- The drawn frame shows the current pulse frame on every pulse dot and records `frame.dots`.
- With a popup open, no dot is recorded.

List every message path checked.

**Interfaces:**
- Consumes: `styles.pulse`, `styles.goingDot`, `pulseSteps` (Task 1); `inProgress`, `dotGoing`.
- Produces:
  - `type pulseMsg struct{}`
  - `const pulseStep = 120 * time.Millisecond`
  - `func pulseAfter(d time.Duration) tea.Cmd`
  - `func (m Model) hasWork() bool`
  - `func (m *Model) armPulse() tea.Cmd`
  - `func (m Model) pulseDots(s string) string`
  - `Model` fields `pulse int` and `pulsing bool`
  - `frameCache` field `dots bool`

- [x] **Step 1: Write the failing test**

Add to `internal/tui/model_test.go`:

```go
// quietModel is a board with no work under way, so the pulse never starts.
func quietModel(t *testing.T) Model {
	t.Helper()
	cfg := treeCfg(t, map[string]string{".acta/bugs/2026-09-21-q.md": "---\nid: BUG-0009\n---\n# Quiet\n"})
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return New(cfg, b, true)
}

func TestPulseMovesOnlyWithADotOnScreen(t *testing.T) {
	t.Parallel()

	m := actModel(t)
	m.frame = &frameCache{dots: true}
	next, cmd := m.Update(pulseMsg{})
	got := next.(Model)
	if cmd == nil || !got.pulsing {
		t.Fatal("a pulse with work under way did not arm the next one")
	}
	if got.pulse != 1 || got.same {
		t.Fatalf("pulse %d same %v, want frame 1 drawn again", got.pulse, got.same)
	}
	got.frame = &frameCache{dots: false}
	next, cmd = got.Update(pulseMsg{})
	quiet := next.(Model)
	if cmd == nil {
		t.Fatal("work under way but no dot on screen stopped the pulse chain")
	}
	if quiet.pulse != 1 || !quiet.same {
		t.Fatalf("with no dot on screen the pulse moved to %d or drew again (same %v)", quiet.pulse, quiet.same)
	}
}

func TestPulseStopsWithNoWork(t *testing.T) {
	t.Parallel()

	m := quietModel(t)
	m.frame = &frameCache{dots: true}
	next, cmd := m.Update(pulseMsg{})
	if cmd != nil || next.(Model).pulsing {
		t.Fatal("a board with no work under way kept pulsing")
	}
}

func TestResizeAndReloadArmThePulseOnce(t *testing.T) {
	t.Parallel()

	m := actModel(t)
	m.pulsing = false
	next, cmd := m.Update(tea.WindowSizeMsg{Width: m.width + 1, Height: m.height})
	armed := next.(Model)
	if cmd == nil || !armed.pulsing {
		t.Fatal("a resize with work under way did not arm the pulse")
	}
	next, cmd = armed.Update(tea.WindowSizeMsg{Width: armed.width + 1, Height: armed.height})
	if cmd == nil || fmt.Sprint(cmd()) != fmt.Sprint(tea.ClearScreen()) {
		t.Fatal("a second resize while pulsing must only clear the screen")
	}
	quiet := quietModel(t)
	if _, cmd := quiet.Update(tea.WindowSizeMsg{Width: 99, Height: 30}); cmd == nil || fmt.Sprint(cmd()) != fmt.Sprint(tea.ClearScreen()) {
		t.Fatal("a resize with no work must only clear the screen")
	}
	idle := actModel(t)
	idle.pulsing = false
	next, cmd = idle.Update(reloadMsg{b: idle.board})
	if cmd == nil || !next.(Model).pulsing {
		t.Fatal("a reload with work under way did not arm the pulse")
	}
}

func TestPulseAfterSendsAPulse(t *testing.T) {
	t.Parallel()

	if got := pulseAfter(0)(); got != (pulseMsg{}) {
		t.Fatalf("got %#v", got)
	}
}
```

If `actModel` already comes back with `pulsing` set (it is sized, and a resize arms the tick), the explicit `m.pulsing = false` lines above keep each test starting from a known state.

Add to `internal/tui/view_test.go`:

```go
func TestFrameShowsTheCurrentPulseFrame(t *testing.T) {
	withTrueColor(func() {
		m := sized(actModel(t).WithTheme("tokyo-night", true), 160, 40)
		m.frame = &frameCache{}
		out := m.View()
		if !m.frame.dots {
			t.Fatal("a screen with work under way recorded no pulse dot")
		}
		if !strings.Contains(out, m.styles.goingDot) {
			t.Fatalf("frame 0 shows no pulse dot")
		}
		m.pulse = 3
		m.same = false
		out = m.View()
		if strings.Contains(out, m.styles.goingDot) {
			t.Error("frame 3 still shows the frame 0 dot")
		}
		if !strings.Contains(out, m.styles.pulse[3].Render(dotGoing)) {
			t.Error("frame 3 is not on screen")
		}
		pop := press(m, "?")
		pop.frame = &frameCache{}
		pop.View()
		if pop.frame.dots {
			t.Error("a popup is open, yet a pulse dot was recorded, so the tick keeps drawing behind it")
		}
	})
}
```

Add `fmt` to the `model_test.go` imports if it is missing.

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/tui/ -run 'TestPulseMovesOnlyWithADotOnScreen|TestPulseStopsWithNoWork|TestResizeAndReloadArmThePulseOnce|TestPulseAfterSendsAPulse|TestFrameShowsTheCurrentPulseFrame'`
Expected: FAIL to build with `undefined: pulseMsg`.

- [x] **Step 3: Write minimal implementation**

In `internal/tui/model.go`, change `frameCache` and its comment:

```go
// frameCache keeps the last frame View drew, and whether a pulse dot was on
// it, so the pulse tick knows if drawing again would change anything. ...
type frameCache struct {
	s    string
	dots bool
}
```

Keep the rest of the old comment as it is.

Next to the other message types:

```go
// pulseMsg moves the pulse of the dots of work under way one frame on.
type pulseMsg struct{}

// pulseStep is the time one pulse frame stays on screen, so eight frames make
// one breath of about a second.
const pulseStep = 120 * time.Millisecond

// pulseAfter sends the next pulse once d has passed.
func pulseAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return pulseMsg{} })
}

// hasWork says if any item on the board is under way, the only time a dot
// can pulse at all.
func (m Model) hasWork() bool {
	if m.board == nil {
		return false
	}
	for _, it := range m.board.Items {
		if inProgress(it) {
			return true
		}
	}
	return false
}

// armPulse starts the pulse chain when there is work under way and no pulse
// is on its way already, so two chains never run at once.
func (m *Model) armPulse() tea.Cmd {
	if m.pulsing || !m.hasWork() {
		return nil
	}
	m.pulsing = true
	return pulseAfter(pulseStep)
}
```

Add to `Model`, next to `same`:

```go
	pulse    int  // the pulse frame the dots of work under way wear now
	pulsing  bool // true while a pulse message is on its way
```

In `Update`, the `tea.WindowSizeMsg` case ends:

```go
		if c := m.armPulse(); c != nil {
			return m, tea.Batch(tea.ClearScreen, c)
		}
		return m, tea.ClearScreen
```

In the `reloadMsg` case, after `m.moveTo(m.cursor())`, add `return m, m.armPulse()`.

Add a case:

```go
	case pulseMsg:
		m.pulsing = false
		if !m.hasWork() {
			m.same = true
			return m, nil
		}
		m.pulsing = true
		if m.frame == nil || !m.frame.dots {
			// No dot on screen: nothing to draw, but work may show again.
			m.same = true
			return m, pulseAfter(pulseStep)
		}
		m.pulse = (m.pulse + 1) % pulseSteps
		return m, pulseAfter(pulseStep)
```

In `internal/tui/view.go`, add `pulseDots` and use it in `frameFrom`:

```go
// pulseDots swaps every dot of work under way for the frame of the moment, and
// notes whether there was one, so the pulse tick knows if it has to draw.
func (m Model) pulseDots(s string) string {
	dots := m.styles.goingDot != "" && strings.Contains(s, m.styles.goingDot)
	if m.frame != nil {
		m.frame.dots = dots
	}
	if !dots || m.pulse == 0 {
		return s
	}
	return strings.ReplaceAll(s, m.styles.goingDot, m.styles.pulse[m.pulse].Render(dotGoing))
}
```

In `frameFrom`, the tail becomes:

```go
	line := fit(m.statusLine(), m.width)
	if m.popupBox() != "" {
		// The box can be as tall as the screen, so it is laid over the body
		// and the status line together. Cover greys every line it gets and
		// strips the hyperlinks, which are off while a popup is open. No dot
		// can pulse behind it, so the tick has nothing to draw.
		if m.frame != nil {
			m.frame.dots = false
		}
		return m.styles.paintFrame(m.cover(out+"\n"+line), m.width)
	}
	return m.styles.paintFrame(m.pulseDots(out+"\n"+line), m.width)
```

Keep every other line of `frameFrom` as it is.

- [x] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/tui/`
Expected: PASS, including `TestResizeClearsTheScreen`. That test is sized from `newModel`: its first resize may arm the pulse, and its second resize, while pulsing, returns `tea.ClearScreen` alone. If it fails because `newModel` holds work and is not pulsing at the second resize, set `m.pulsing = true` before the second resize in that test, and list the change.

- [x] **Step 5: Commit**

```bash
gofmt -l internal && go vet ./internal/tui/ && scripts/test ./internal/tui/
git add internal/tui/model.go internal/tui/view.go internal/tui/model_test.go internal/tui/view_test.go
git commit -m "tui: pulse the dots of work under way while any is on screen"
```

---

### Task 5: Help keys in a right-aligned cyan column

**Files:**
- Modify: `internal/tui/view.go` (`helpText`)
- Test: `internal/tui/view_test.go` (`TestHelpKeysWearTheAccent` is replaced)

**verify:** On every line of the `?` popup:
- The key is drawn in cyan (`styles.label`, slot 6), never in the accent.
- The key cells end at the same column on every line, the width of the widest key.
- The description starts exactly one cell after that column, in the plain foreground.
- The words of the popup equal the words of `helpLines`.
- Each popup row has the same width as before, at widths 60, 80, 120 and 200, in both the hex and the `terminal` theme.

The value picker and the new-bug prompt stay free of accent inside the box. List every line and width checked.

**Interfaces:**
- Consumes: `helpLines`, `styles.label`, `boxView`, `popupBox`; test helpers `withTrueColor`, `plain`, `newModel`, `sized`, `press`.
- Produces: `func (m Model) helpText() string` (same name, new layout).

- [x] **Step 1: Write the failing test**

In `internal/tui/view_test.go`, replace `TestHelpKeysWearTheAccent` with:

```go
// TestHelpKeysSitInARightAlignedColumn reads every line of the ? popup. The
// keys are cyan so they differ from the border, and they end at one column,
// with what each does starting right after.
func TestHelpKeysSitInARightAlignedColumn(t *testing.T) {
	withTrueColor(func() {
		for _, name := range []string{"tokyo-night", "terminal"} {
			m := sized(newModel(t).WithTheme(name, true), 120, 40)
			text := m.helpText()
			if got, want := strings.Fields(plain(text)), strings.Fields(helpLines); strings.Join(got, " ") != strings.Join(want, " ") {
				t.Fatalf("%s: help words changed:\n got %q\nwant %q", name, got, want)
			}
			keyW := 0
			type pair struct{ key, what string }
			var want []pair
			for _, ln := range strings.Split(helpLines, "\n") {
				k, w, _ := strings.Cut(ln, "  ")
				p := pair{strings.TrimSpace(k), strings.TrimSpace(w)}
				want = append(want, p)
				keyW = max(keyW, lipgloss.Width(p.key))
			}
			accent := strings.TrimSuffix(m.styles.accent.Render("x"), "x\x1b[0m")
			for i, ln := range strings.Split(text, "\n") {
				p := want[i]
				wantLine := strings.Repeat(" ", keyW-lipgloss.Width(p.key)) + m.styles.label.Render(p.key) + " " + p.what
				if ln != wantLine {
					t.Errorf("%s line %d:\n got %q\nwant %q", name, i, ln, wantLine)
				}
				if accent != "" && strings.Contains(ln, accent) {
					t.Errorf("%s line %d: a key is still in the accent: %q", name, i, ln)
				}
			}
			for _, w := range []int{60, 80, 120, 200} {
				mw := sized(newModel(t).WithTheme(name, true), w, 40)
				before := strings.Split(mw.boxView("Keys", helpLines), "\n")
				after := strings.Split(press(mw, "?").popupBox(), "\n")
				if len(before) != len(after) {
					t.Fatalf("%s at %d: box has %d rows, want %d", name, w, len(after), len(before))
				}
				for i := range before {
					if lipgloss.Width(after[i]) != lipgloss.Width(before[i]) {
						t.Errorf("%s at %d row %d: width %d, want %d", name, w, i, lipgloss.Width(after[i]), lipgloss.Width(before[i]))
					}
				}
			}
		}
	})
}
```

In the `terminal` theme the accent and label codes are ANSI numbers, so `accent` is still a real code there. Keep `TestOtherPopupsHaveNoAccentInside` as it is.

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/tui/ -run 'TestHelpKeysSitInARightAlignedColumn|TestOtherPopupsHaveNoAccentInside'`
Expected: FAIL. The keys are in the accent and left-aligned.

- [x] **Step 3: Write minimal implementation**

In `internal/tui/view.go`, replace the body and comment of `helpText`:

```go
// helpText lays the key map out the way lazygit does: the keys in cyan, so
// they differ from the border, in a column that ends at one place, and what
// each key does right after it. A line of helpLines splits at its first run
// of two spaces, the gap its key column is written with.
func (m Model) helpText() string {
	type pair struct{ key, what string }
	var pairs []pair
	keyW := 0
	for _, ln := range strings.Split(helpLines, "\n") {
		k, w, _ := strings.Cut(ln, "  ")
		p := pair{strings.TrimSpace(k), strings.TrimSpace(w)}
		pairs = append(pairs, p)
		keyW = max(keyW, lipgloss.Width(p.key))
	}
	lines := make([]string, len(pairs))
	for i, p := range pairs {
		lines[i] = strings.Repeat(" ", keyW-lipgloss.Width(p.key)) + m.styles.label.Render(p.key) + " " + p.what
	}
	return strings.Join(lines, "\n")
}
```

- [x] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/tui/`
Expected: PASS, including the help tests in `model_test.go` and `frame_test.go` that read `helpLines`.

- [x] **Step 5: Commit**

```bash
gofmt -l internal && go vet ./internal/tui/ && scripts/test ./internal/tui/
git add internal/tui/view.go internal/tui/view_test.go
git commit -m "tui: lay the help keys out in a right-aligned cyan column"
```

---

### Task 6: Remove the work brush

**Files:**
- Modify: `internal/tui/styles.go` (`styles.work`, `slotWork`)
- Test: `internal/tui/styles_test.go`

**verify:**
- No file in `internal/` still names `styles.work`, `.work` on a styles value, or `slotWork`.
- The package builds and every test passes.
- The asserts that pinned `work` are replaced by asserts about what took its place: work under way is plain, and its dot is the pulse.

List every place that named them.

**Interfaces:**
- Consumes: Tasks 2 and 3 have stopped using `work`.
- Produces: nothing.

- [x] **Step 1: Write the failing test**

In `internal/tui/styles_test.go`, add:

```go
func TestNoWorkBrushIsLeft(t *testing.T) {
	t.Parallel()

	if _, ok := reflect.TypeOf(styles{}).FieldByName("work"); ok {
		t.Error("styles still has a work brush; work under way is plain now")
	}
}
```

Add `reflect` to the imports if it is missing. `slotWork` is a constant, so the test cannot see it; the grep in Step 3 covers it. Then update the old asserts that name `s.work`:
- In `TestTerminalThemeUsesANSISlots`, drop `+ s.work.Render("x")` from the line that joins the rendered brushes.
- In `TestHexThemeRoles`, replace the `s.work` foreground assert with `if s.pulse[0].GetForeground() != lipgloss.Color(th.ANSI[slotGreen]) { t.Fatalf(...) }`.
- In the test that checks `s.work.GetFaint()`, drop that check.

Run `grep -rn "\.work\b\|slotWork" internal/` to find every other place.

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/tui/ -run 'TestNoWorkBrushIsLeft'`
Expected: FAIL. `styles` still has a `work` field.

- [x] **Step 3: Write minimal implementation**

In `internal/tui/styles.go`:
- Remove `work` from the `styles` fields.
- Remove `slotWork = 4` from the slot constants.
- Remove the `work:` entry and its comment from `newStyles`.

Then run `grep -rn "\.work\b\|slotWork" internal/`. It must print nothing.

- [x] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/tui/`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l internal && go vet ./internal/tui/ && scripts/test ./internal/tui/
git add internal/tui/styles.go internal/tui/styles_test.go
git commit -m "tui: remove the work brush now that work under way is plain"
```
