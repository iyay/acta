---
id: BUG-0051
hash: de6chi6
fixed_in: ca16f38b0b4ce49af5d8075b7737192464f664ea
finished: "2026-10-19 21:09:06"
---
# Tidy squashes a docs-only branch under a stranger's commit message

## Symptom
Landing PLN-0176 printed `tidy: 10 commits -> 1`. The one commit (2009408)
carries the subject `docs(bug): correct missing task checkboxes diagnosis`,
a bug-file edit, not the branch's work. The subjects `docs(skills)`,
`docs(setup)`, `polish` and every `Task: PLN-fd6jcne#n` trailer are gone, so
the Commits screen cannot tie the change to its tasks.

## Root cause
internal/tidy/tidy.go:36 puts `docs` and `polish` in `foldTypes`, so a branch
whose real work is skill text keeps no commit. That sends it to the
`len(nodes) == 0` path at internal/tidy/tidy.go:206, which takes the message of
the newest commit whose type is not `chore` (tidy.go:211). That check ignores
`foldTypes`, so any late `docs(...)` commit wins, and the other messages and
trailers are dropped.

## Repro
On a branch from main, commit `docs(skills): a` with a Task trailer touching
plugin/skills/x/SKILL.md, then `polish: b` with a trailer, then
`docs(bug): c` touching only .acta/bugs/. Run `acta tidy main <branch>`.
The result is one commit with subject `docs(bug): c` and no trailers.

## Found in
main after landing PLN-0176 (commit 2009408), read while explaining the tidy line.
