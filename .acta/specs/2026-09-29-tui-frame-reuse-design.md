---
parent: scratch/2026-09-29-tui-lag-while-agents-work
id: SPC-0029
created: "2026-09-29"
hash: nla4kqe
---
# TUI frame reuse

Status: design approved by the user in chat on 2026-09-29. Bounded (changes the existing render flow in `internal/tui`).

## Why

After PLN-0037, scrolling is much better but still lags a little, and it hangs for a while when an agent is working. A live trace on main 527ddd3 (24 s, an agent working) showed:

- No watcher events and no reloads, so reload chains were not the cause.
- 451 `MouseMsg` and 154 wheel ticks, but `View` still ran 637 times. About 450 of those renders changed nothing on screen, because the notch only added to the pending delta.
- `View` took 7.8 ms at p50, 14.6 ms at p90, and 292 ms at worst. That is twice the 3.0 ms from the first trace, likely because the agent's tests and builds take CPU.
- `Update` is cheap: 0.09 ms per mouse message. All the cost is in `View`.
- The main loop was busy up to 965 ms of one second.

## Design

1. **Keep the last frame.** The model keeps the last string `View` returned. It sits behind a pointer that every copy of the model shares, the same way `dcache` does.
2. **A "screen unchanged" flag.** Only the paths that surely change nothing on screen set it:
   - a wheel notch that only adds to the pending delta,
   - a mouse event that is ignored (a wheel over a pane without the focus, or anything that is not a left click or a wheel),
   - a wheel tick with nothing pending.

   Every other message clears the flag. When in doubt, the screen is drawn again.
3. **`View`:** when the flag is set and a frame is kept, return the kept frame. Otherwise draw the full frame, keep it, and return it.

## Out of scope

- Making `View` itself cheaper (7.8 ms per frame). Measure with the same trace after this lands, then decide.
- Reload cost (`board.Load` spawning one `git` per file).

## Testing

- A notch that only adds to the delta returns the exact same frame without drawing. Count the draws.
- Every message that changes the screen draws a new frame: a tick that scrolls, a key, a click, a reload, a resize, and a clock tick.
- An ignored mouse event and an empty tick draw nothing.
- Turning the flag off makes the draw-count tests go red.
