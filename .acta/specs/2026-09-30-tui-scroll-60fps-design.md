---
parent: scratch/2026-09-30-scroll-60fps
created: "2026-09-30"
id: SPC-0035
hash: w97o6ps
started: "2026-09-30"
finished: "2026-09-30"
---
# TUI scroll at 60fps

Status: design approved by the user in chat on 2026-09-30. Architectural (changes the wheel timing and how `draw` puts the panes together, and adds a trace switch).

## Why

The user wants trackpad scrolling at 60fps. A benchmark on the real board (200x55, idle machine, main 0cfe59a) showed:

- The wheel path runs in series. The first notch arms a 16 ms tick (`internal/tui/model.go:562`). The tick scrolls (`model.go:232`). Then `View` runs. Then the Bubble Tea v1 renderer writes on its own 60fps ticker. One scroll step costs about 16 + 7.5 + up to 16.6 ms, so the best case is near 30 fps.
- `View` takes 7.5 ms with the detail pane focused and 2.7 ms with the list pane focused.
- Most of `View` is `ansi.StringWidth`. Each cell is measured about three times: in `pad`, in `lipgloss.JoinHorizontal` and `JoinVertical`, and in `fit` inside `draw`.
- A notch costs 38 to 90 us and a tick about 150 us, so `Update` is not the problem.

## Pass mark

Measured with the trace below, on an idle machine, over 5 s of steady trackpad scrolling in the detail pane:

1. The gap between two flushes is at most 16.7 ms at p95, and there are at least 55 flushes per second. The rate is the number of flushes between the first and the last notch, divided by that time.
2. From a first notch to the next flush is at most 16.7 ms at p95. A first notch is one that comes when no wheel tick is armed.

The time from each notch to its flush is not part of the mark. At about 220 notches a second, a notch that comes in the middle of a frame must wait for the next one. Meeting 16.7 ms for every notch would mean one `View` per notch, which is more than the loop can do.

The user runs the trace in another tab. The TUI cannot run through `!` because there is no `/dev/tty`.

## Design

### 1. Trace switch

1. A new file, `internal/tui/trace.go`, holds a tracer. Two goroutines write to it: the event loop (notches and `View`) and the renderer (flushes), so a mutex guards it. A nil tracer does no work at all.
2. In `internal/cli/cli.go`, when `ACTA_TUI_TRACE` names a file, the TUI opens it and wraps stdout with `tea.WithOutput`. Each `Write` to stdout counts as one flush. If the file cannot be opened, the TUI prints a warning on stderr and runs with no trace.
3. The log has one line per event:
   - `notch <unix_ns> first` or `notch <unix_ns> gathered`
   - `view <unix_ns> <duration_ns>`
   - `flush <unix_ns> <bytes>`
4. When the TUI quits, it adds one last line: `summary frames=<n> fps=<n> gap_p95_ms=<n> first_p50_ms=<n> first_p95_ms=<n> view_p95_ms=<n>`. A log with no notches still gets a summary line, with zero counts.

### 2. Wheel timing

1. A notch that comes when no tick is armed scrolls at once and arms a tick.
2. A notch that comes while a tick is armed only adds to the pending delta, as it does now, and keeps the last frame (`m.same = true`).
3. A tick with a pending delta scrolls and arms the next tick, so steady scrolling moves once per frame.
4. A tick with nothing pending stops. The next notch is a first notch again.
5. The mark rule (drop the delta when the screen changed) and the flush to another pane stay as they are. A first notch has no pending delta, so it needs no mark.
6. `wheelFrame` goes from 16 ms to 12 ms. The renderer writes every 8.3 ms at 120fps. A 16 ms step plus a little work would sometimes miss two flush ticks and leave a 25 ms gap. With 12 ms, the gaps are 8.3 or 16.7 ms.
7. `cli.go` starts the program with `tea.WithFPS(120)`, so a frame waits at most 8.3 ms to be written.

### 3. Measure each line once

1. In the wide layout, `draw` (`internal/tui/view.go:98-120`) stops using `lipgloss.JoinVertical` and `lipgloss.JoinHorizontal`. It builds each line as `column[i] + detail[i]`. `paneView` already pads each line to its box width, so nothing needs to be measured again.
2. The joined lines are exactly `m.width` wide, so `draw` does not call `fit` on them. The narrow layout and the tab rows still go through `fit`.

## Out of scope

- A cache of each pane's drawn lines. It becomes a follow-up plan only if the trace still fails the pass mark.
- Row sort cost (`ordered`, `idNum`). It is 2.5% of `View`, and it only matters on the notch and tick path, which is already under 0.2 ms.
- DBT-0030.07 (`detailParts` built several times per frame). `detailParts` has been cached since PLN-0037 (`internal/tui/detail.go:43`), so the item looks stale. Check it during triage; this spec does not close it.
- Lag while agents use the CPU. The pass mark is for an idle machine only.

## Testing

- Trace: the summary numbers come from fake events with set clock times, with no sleeps. There are also tests for an empty `ACTA_TUI_TRACE`, a file that cannot be opened, and a log with no notches.
- Wheel:
  - a first notch moves the offset without a tick;
  - gathered notches move only on the tick;
  - a tick with a delta returns a command (the next tick) and a tick with nothing pending returns nil;
  - a notch after an empty tick scrolls at once again;
  - a screen change between the first notch and the tick drops the gathered delta.
- Old wheel tests that expect the first notch to wait for the tick are rewritten for the new rule, not deleted.
- Draw:
  - every existing view test stays green with the same output;
  - a new test checks that every line of the frame is exactly `m.width` cells wide, in the wide and the narrow layout, with unicode text and with cut lines.
- Benchmark: `BenchmarkWheelFrame` now turns the wheel the same way every notch, so it really scrolls on the long board.

## Order

PLN-0039 (tui-colors) is being built in `../acta-tui-colors` and changes `internal/tui/view.go`. The build of this spec waits until PLN-0039 has landed.
