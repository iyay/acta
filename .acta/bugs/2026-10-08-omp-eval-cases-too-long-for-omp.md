---
id: BUG-0037
hash: w8v7yjz
priority: low
---
# Two eval cases still time out under omp after the runner stops at the final reply

## Symptom
`scripts/eval --omp` on main at ec07b47 (0.1.39) gives 9 passed, 2 failed, 1 skipped. The two fails are `side-idea-to-scratch: timeout` (cap 240 s) and `state-resume: timeout` (cap 300 s). Both also timed out on the run before ec07b47.

## Root cause
Not proven for the ec07b47 run: the runner still keeps no stream for a run that times out, so that run left no evidence. What the earlier hand probes showed (old runner, real HOME):
- state-resume: the first final reply came at 283 s of a 300 s cap, so a slightly slower run passes the cap before the agent answers.
- side-idea-to-scratch: `plugin/evals/side-idea-to-scratch/prompt.md` asks the agent to fix one failing test, but `scaffold.sh` makes no test at all. The agent searches for it; the first final reply came at 361 s of a 240 s cap. Part of that search chased the plugin notice that ec07b47 removed.
- omp has no turn cap, so `max_turns` (5 and 9 in these cases), which ends the Claude run early, does nothing here (`plugin/evals/FACTS.md`, omp runner notes).
Likely cause: these two cases are too long, or impossible as written, for omp without a turn cap. Confirm with a timed stream from the new runner before changing any case.

## Repro
`go run ./cmd/acta eval-omp --case side-idea-to-scratch plugin` and `go run ./cmd/acta eval-omp --case state-resume plugin` from the repo root, with the acta from main on PATH.

## Found in
main at ec07b47, the omp eval rerun after landing PLN-0117 on 2026-10-08. The user chose to record it rather than probe again now.
