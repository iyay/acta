---
id: SPC-0080
created: "2026-10-05 15:18:54"
hash: raxfale
started: "2026-10-05 15:24:25"
finished: "2026-10-05 15:43:17"
---
# Planning commits use chore(<kind>): instead of acta:

Status: Bounded, approved by the user in chat on 2026-10-05.

Why: every commit acta makes for planning files starts with `acta: `. The user wants the conventional commit form, with the kind of file as the scope and no tool name.

Design:
- Format: `chore(<kind>): <what>`. `<kind>` is the planning folder of the files in the commit: `scratch`, `spec`, `plan`, `bug`, `debt`, `wiki`.
- A commit that touches more than one kind (for example `acta id` stamping specs and plans) uses `chore: <what>` with no scope. So do `acta doctor --fix` and the root folder migrate.
- The `<what>` part keeps today's words, for example `chore(plan): tick wave 2`, `chore(debt): new 2026-10-05-live-work-state`, `chore: assign short ids`.
- Every auto-commit message in `internal/write`, `internal/cli` (doctor, migrate) and anything else that writes `acta: ` into a commit subject changes to this form.
- Skill text that tells an agent to commit with `acta: ...` (land, migrate, build, and the review and fix-round steps) uses the new form.
- `internal/board/state.go` finds the last review round from `chore(plan): tick fix round N`, and still reads the old `acta: tick fix round N`, so history keeps working.
- Code commits written by implementers (`hook: ...`, `tui: ...`) do not change. The skill description prefix `acta: ` that plugincheck checks is not a commit message and stays.
- The last task adds 1 to the patch version in the three plugin files.

Tests: each write command's commit subject has the right scope; a multi-kind commit has no scope; the review round is found from both the new and the old subject; plugincheck passes; no skill line still tells an agent to commit with `acta: `.
