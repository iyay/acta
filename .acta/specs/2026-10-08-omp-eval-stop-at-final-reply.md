---
parent: bugs/2026-10-08-omp-eval-runs-past-final-reply
id: SPC-0108
created: "2026-10-08 06:48:36"
hash: irpap2i
started: "2026-10-08 06:53:46"
finished: "2026-10-08 06:57:32"
---
# The omp eval runner stops at the first final reply and runs omp in a clean home

Status: Bounded, approved by the user in chat on 2026-10-08. Fixes the bug in `.acta/bugs/2026-10-08-omp-eval-runs-past-final-reply.md`.

Why: `RunCase` (`internal/evalomp/run.go`) waits for omp to exit, and omp in print mode wakes the agent again after its final reply while todos are open, so finished cases hit `timeout_seconds`. omp also runs with the user's real `HOME`, so the acta session hook reads the user's own Claude plugin settings and steers the agent; the Claude runner uses a clean home.

Design:
- `RunCase` reads omp's JSON stream while omp runs. The run ends at the first assistant `message_end` whose `stopReason` is `stop`: that message's text is the reply, the tool calls up to it are the calls, and the runner then kills omp's process group the same way the timeout does. `agent_end` still ends the run when it comes first, as today.
- A stream that ends with neither a final reply nor `agent_end` stays an error, and a run past `timeout_seconds` stays `ErrTimeout`.
- omp runs with `HOME` set to the case's throwaway home. Before omp starts, the runner links `<throwaway home>/.omp` to the user's real `~/.omp`, so omp keeps its login and model settings and nothing else from the real home is visible. `PM_VOICE_FILE` stays as it is.
- The judge (`OmpJudge`) is not changed.
- The last task adds 1 to the patch version in the three plugin files.

Out of scope: changing any case, prompt or grader (side-idea-to-scratch included); turning omp's own continue behaviour off; a turn cap.

Tests: with a fake omp script that prints a final reply, then sleeps longer than the case timeout: the run returns that reply with no `ErrTimeout`, and the fake process is gone after. Tool calls after the first final reply are not counted. A fake omp that prints `HOME` and lists `$HOME/.omp` shows the throwaway home and the link to the real `~/.omp` (the test points the real home at a temp dir). A stream with no final reply and no `agent_end` still errors.
