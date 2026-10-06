---
id: DBT-0039
hash: hyhg747
parent: plans/2026-09-30-tui-scroll-60fps
---
# Review NOTEs: TUI Scroll at 60fps Implementation Plan

- [ ] (low) internal/tui/model.go wheelMoved field is not in the plan; it only serves the wheel() test helper (Update, tick, no View), and costs one extra draw at the end of each wheel burst in the real program; its comment "has not been written down yet" is wrong for the real loop.
- [ ] (low) internal/tui/model.go a tick that drops a gathered delta after the mark changed now draws again, where base kept the frame (speed only).
- [ ] (low) internal/tui/model.go a first notch on a pane already at its end moves nothing but leaves same false, so a wheel at the bottom costs two useless draws where main drew none.
- [ ] (low) internal/tui/view.go draw filler branch for columns of different heights never runs on screen (only at height <= 3, where frameFrom cuts the body); no test covers it.
- [ ] (low) internal/tui/draw_test.go both draw tests compare against the old output, so they stay green if draw is reverted; the benchmark is the only measure of the gain.
- [x] Task 3 benchmark went from about 1.58 ms to 1.44 ms per frame (about 8%), far less than the spec's estimate; the live trace after landing decides whether the per-pane render cache follow-up is needed. (stale)
- [ ] (low) internal/tui/trace.go traceFile embeds *os.File, so the promoted WriteString (used by the renderer's execute for control sequences) is not traced; only frame writes count, which differs from the spec's "each Write counts as a flush" wording.
- [ ] (low) internal/tui/trace.go Close writes the summary without holding the lock; safe today because runTUI calls it after p.Run returns.
- [ ] (low) internal/tui/trace.go Summary uses sort.Search and assumes flushes are in time order; true with one renderer goroutine, but no test pins it.
- [x] internal/tui/trace.go flush got a nil guard and internal/cli/tuitrace.go done() is safe to call twice; both go beyond the plan's code, both harmless. (stale)
- [ ] (low) Every Write to stdout counts as a flush, so non-frame escape writes also count; they mostly land before the first notch, outside the window.
- [ ] (low) internal/tui/view_bench_test.go resets the pane by writing m.off[paneDetail] = 0 directly.
