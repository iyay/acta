---
parent: scratch/2026-09-29-tui-scroll-performance
id: SPC-0028
created: "2026-09-29"
hash: uqx9ohi
---
# TUI scroll performance

Status: design approved by the user in chat on 2026-09-29. Bounded (changes the existing scroll and render flow in `internal/tui`).

## Why

Scrolling feels slow and laggy in every pane, even when no agent runs. A live trace (throwaway build, 141 s run) showed:

- 3970 `tea.MouseMsg`, 222 `tea.KeyMsg`, only 8 reloads. Reload is not the cause.
- Bubble Tea calls `View` after every message: up to 217 calls per second.
- `View` p50 3.0 ms, p90 5.0 ms, max 181 ms.
- The main loop was busy up to 796 ms of every second. Input queues up, so the screen falls behind the wheel.
- The wheel moves one line per event (`internal/tui/model.go:459`), so scrolling is also slow to cover ground.

## Design

1. **Batch wheel events.** A wheel event does not scroll at once. It adds to a pending delta for its pane. The first pending event starts one frame tick (16 ms). When the tick fires, each pane scrolls by its whole delta once, clamped to its content, and the delta resets. Many wheel events in one frame cost one scroll and one render.
2. **Cache detail lines.** `detailLines` is built several times per event (`View`, `scrollPane` through `linesOf`, `clampOff`). Keep the last result, keyed by the selected item, the width and the board pointer. Any change to one of those builds it again.
3. **Wheel step of 3 lines.** One wheel event moves 3 lines, the usual terminal step, so a scroll covers ground.

## Out of scope

- `board.Load` runs one `git` per file for authors (about 1.6 s per reload). This is a separate cost. It is not the cause of the scroll lag.
- Key scrolling (`j`, `k`, arrows) stays as it is. The trace shows keys are few and cheap.

## Testing

- Several wheel events inside one frame end in one scroll with the summed delta. The result stays clamped at both ends of the content.
- Wheel up and down in the same frame cancel out.
- Each wheel event moves 3 lines.
- The detail cache gives new lines after the board, the width or the selected item changes.
- A benchmark for wheel plus `View` on a large board, to see the gain.
