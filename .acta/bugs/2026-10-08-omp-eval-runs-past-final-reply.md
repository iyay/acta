---
id: BUG-0036
hash: ddiitc2
fixed_in: ec07b47c982b5c18672c9ba5c4ed6a0cc6153e5b
finished: "2026-10-08 07:03:06"
---
# omp evals time out on cases the agent already finished

## Symptom
`scripts/eval --omp` on main (0.1.38) gives `FAIL side-idea-to-scratch: timeout`, `FAIL state-resume: timeout` and `FAIL wiki-hint: timeout`, while the agent's work in those runs is done well inside the cap.

## Root cause
1. `RunCase` in `internal/evalomp/run.go` waits for the omp process to exit and parses the stream only after it (`ParseStream` in `internal/evalomp/result.go` needs `agent_end`). In print mode omp starts the agent again after a reply that ends with `stopReason: stop` while its todo list has open items, so the run goes on past the answer. Probe: state-resume answered at 283 s, omp sent `agent_start` at 287 s and the agent went on to review and land until the 300 s cap. `max_turns`, which ends the Claude run, is ignored in omp.
2. `RunCase` sets only `PM_VOICE_FILE` for omp and leaves the real `HOME`, so the acta session hook reads the user's own Claude plugin settings and adds the "two plugins that do the same job" notice (`internal/hook/hook.go:142`). In side-idea-to-scratch the agent took that notice for the failing test it was told to fix and spent 384 s on it (cap 240). The Claude runner uses a clean home, so the two harnesses grade different setups.

## Repro
Scaffold a case into an empty folder with a throwaway HOME, then run the same omp command `ompArgs` builds with stdin from /dev/null and a timestamp on each stream line; look for `agent_start` after a `message_end` with `stopReason: stop`, and for the plugin notice in the first reply.

## Found in
main after PLN-0116, debugging the omp eval run of 2026-10-08 (8 pass, 3 fail, 1 skip).
