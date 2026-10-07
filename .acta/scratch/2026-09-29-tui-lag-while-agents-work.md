---
id: SCR-0027
hash: fgyni8z
title: TUI still lags a bit, and hangs while an agent works
status: raw
created: "2026-09-29"
schema: "1"
finished: "2026-09-29"
---
# TUI still lags a bit, and hangs while an agent works

## Words

### 2026-09-29

soo much better, mesikupun masih sedikit lagging
dan kadang lag-nya lama sampe hang ketika ada agent yang lagi kerja

## Context

Came up right after PLN-0037 (SCR-0010, wheel batching) landed at 7de14e7 on 2026-09-29. The user confirmed scrolling is much better but still lags a little.
Suspects, not measured yet:
- Bubble Tea still runs a full View (3-5 ms) after every mouse message, even a notch that only adds to the pending delta.
- Reload chaining while agents write files (.agents.json, acta tick, worktree commits). One trees.Load took about 1.9 s on 572 items and spawns one git process per file (fillAuthors, internal/board/closed.go:22). WatchDirs also includes worktrees.
- Motion events from tea.WithMouseCellMotion (internal/cli/cli.go:54).
A throwaway trace build is ready at the session scratchpad (acta-trace2, writes trace2.log). It logs Update and View times, every watcher event and every reload with its duration. Run it while an agent works to measure before designing.

## Log

### 2026-09-29

2026-09-29 trace on main 63f99d0 (throwaway build, 24 s run, the user saw lag while an agent worked):
- 451 MouseMsg, 154 wheelTickMsg, 25 KeyMsg.
- 0 watcher file events and 0 reloads. Reload chaining was not the cause in this run.
- View ran 637 times: p50 7.8 ms, p90 14.6 ms, p99 35 ms, max 292 ms. That is about twice the p50 of 3.0 ms from the first trace, likely because the agent's work (tests, builds) competes for CPU.
- The main loop was busy up to 965 ms of one second.
- Bubble Tea still runs View after every MouseMsg, including a notch that only adds to the pending delta. About 450 of the 637 renders changed nothing on screen.
- Update is cheap: MouseMsg averages 0.09 ms and the tick 0.15 ms. The cost is all in View.
- WindowSizeMsg took 62-75 ms.
Next idea: keep the last frame string and return it when nothing on screen changed, then cut the cost of View itself.

## Open questions
