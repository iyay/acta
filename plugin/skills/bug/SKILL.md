---
name: bug
description: Use when a bug is confirmed (root cause proven with file:line and a repro, a review BLOCKER, or a bug found while reading and reported to the user), when the user asks to note something as a bug, or when a recorded bug's fix has landed. Writes the bug file through pmb.
---

# Bug

Bugs live as files under `.pm/bugs/`, one per bug, in the shape the pm-board file contract gives. `pmb` writes and commits them; do not hand-write the file.

## When

- `pm:debug` proved a root cause: `file:line` plus a repro.
- `pm:review` returned a BLOCKER.
- You found a bug while reading and told the user.
- The user asks to note something as a bug, in any language.

Not a bug file: a guess, a review NOTE, a problem in test tooling or CI glue, anything not reproduced yet.

## Record

Pick a slug: lower-case words joined by `-`, naming the symptom. Then:

```bash
pmb bug new <slug> --ref <ticket code, if there is one> <<'EOF'
# <Short symptom, the way a user would say it>

## Symptom
<What a user or a test sees.>

## Root cause
<file:line, and why.>

## Repro
<Exact steps or command.>

## Found in
<Branch or worktree, and how it was found: debug, review round N, reading.>
EOF
```

`## Symptom` is required; the other sections may wait until they are known. Tell the user the path `pmb` printed. In a worktree, the bug file commits on that branch and lands with it.

Exit codes: 0 written and committed; 1 bad input (fix the slug or the body); 2 written but not committed (tell the user why; `pmb` says it on stderr); 3 something else failed (report it).

## Fix

A bug is fixed through a normal plan whose frontmatter says `parent: bugs/<file name without .md>`. That plan's tasks become the bug's children, and its status follows them. When the fix has landed, record the merge commit:

```bash
pmb set bugs/<file name without .md> fixed_in <merge sha>
```

Never push.
