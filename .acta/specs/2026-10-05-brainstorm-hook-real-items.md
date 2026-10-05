---
parent: bugs/2026-10-05-failed-brainstorm-set-blocks-the-real-one
id: SPC-0079
created: "2026-10-05 14:54:48"
hash: dhsg2v1
---
# The brainstorm hook counts only real scratch items

Status: Bounded, approved by the user in chat on 2026-10-05. Fixes BUG-0031.

Why: the post-tool hook records the stem of any `acta set scratch/<stem> status brainstorming` command, even when acta refused it as an unknown id. The next command with the right stem then counts as a second brainstorm and the pre-tool hook blocks it.

Design (`internal/hook/session.go`):
- `RecordBrainstorm` records a stem only when `<acta root>/scratch/<stem>.md` exists.
- `PreTool` blocks only when the new stem names an existing scratch file and differs from the one recorded. A stem with no file passes; acta refuses it itself.
- A real second brainstorm (two different existing items) is still blocked, as today.
- The last task adds 1 to the patch version in the three plugin files.

Tests: a mistyped stem then the real stem is not blocked; two different real items are blocked; the same stem twice is not blocked.

Out of scope: `acta set SCR-0001 status brainstorming` (an id instead of a path) is not seen by the hook at all.
