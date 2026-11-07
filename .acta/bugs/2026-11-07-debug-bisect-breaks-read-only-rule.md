---
id: BUG-0074
hash: qipr4wo
---
# Debug skill asks for git bisect but also forbids changing the repo

## Symptom
In Phase 2 the debug skill tells agents to compare with an older commit using `git bisect`, while its ground rule says Phases 1 to 3 change no file in the repo. Bisect checks out other commits in the working tree, so following both is not possible. The probe copy from `probe-copy.sh` leaves out `.git`, so bisect cannot run there either. The skill does not say where bisect should run.

## Root cause
`plugin/skills/debug/SKILL.md:36` (Phase 2 step 2) names `git bisect`; `plugin/skills/debug/SKILL.md:12` forbids repo changes in Phases 1 to 3. Neither line says how to bisect without touching the repo (for example `git worktree add` to a temp path).

## Repro
Read both lines of the skill. Then in a probe copy made with `bash plugin/skills/debug/probe-copy.sh`, run `git bisect start`: it prints "not a git repository".

## Found in
Branch `debug-temp-copy-check`, review round 2 of PLN-0217 (Spec axis, pre-existing).
