---
parent: scratch/2026-10-05-dispatch-false-drift
id: SPC-0082
created: "2026-10-05 19:15:46"
hash: sid8mui
started: "2026-10-05 19:28:54"
finished: "2026-10-05 19:33:53"
---
# Dispatch checkpoint: a todo list with no task ids is unconfirmed, not drift

Status: Bounded, approved by the user in chat on 2026-10-05.

Why: `acta dispatch send` read the pane once, saw omp's first todo card (generic steps like "Read brief", "Scope", "Build"), found none of the plan's task ids on it, and exited 4 with `drift: missing ...`. That happened four times on 2026-10-05; each time omp went on to do the right work. Exit 4 then asks the orchestrator to judge by hand every time.

Design (`checkpoint` in `internal/cli/dispatch_herdr.go`):
- A todo card that names none of the plan's task ids gives `unconfirmed` (exit 0), the same as no card yet.
- A card that names some task ids but misses others gives `drift` with the missing ids, as today.
- A card that names every task id gives `ok`, as today.
- `plugin/skills/build/dispatch.md` exit 4 lines say drift now means a card that names some task ids and skips others.
- The last task adds 1 to the patch version in the three plugin files.

Tests: generic card with no ids gives unconfirmed; partial card gives drift with the missing ids; full card gives ok; no card gives unconfirmed.
