---
type: Gotcha
title: omp keeps working after its final reply
description: In print mode omp wakes the agent again after a stop reply while todos are open; the eval runner ends at the first one
paths: [internal/evalomp/]
timestamp: 2026-10-07T23:58:09Z
---

In `omp -p --mode json`, the first assistant `message_end` with `stopReason: stop` is not the end of the run. When the agent's todo list still has open items, omp sends `agent_start` again and the agent keeps going, sometimes for minutes. A state-resume run answered at 283 s, then went on to review and land until the 300 s cap. `max_turns` does not exist in omp, so nothing stops it.

So `RunCase` reads the stream while omp runs, ends at the first final reply (or `agent_end` if that comes first), and kills omp's process group. A case that times out under omp is worth a timed look at the stream before anyone blames the skill.

Two more traps when you probe omp by hand:
- omp waits on stdin when it is not a terminal (`Still starting ... phase: readPipedInput`). Give it `</dev/null`. The runner is fine: Go hands a child /dev/null.
- omp needs only `~/.omp` from the real home. The runner gives omp a throwaway `HOME` with `.omp` linked in, so the acta hook does not read the user's own Claude plugin settings.
