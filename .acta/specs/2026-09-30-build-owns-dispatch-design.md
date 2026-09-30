---
created: "2026-09-30"
parent: scratch/2026-09-30-build-owns-dispatch
id: SPC-0049
hash: vwgdcje
started: "2026-09-30"
finished: "2026-09-30"
---
# One implementation skill: build owns dispatch

Status: design approved by the user in chat on 2026-09-30, section by section. Architectural: it removes a skill and changes how build picks and runs its executors.

## Why

1. There are two skills that implement a plan: `acta:build` and `acta:dispatch`. Build already reads the executor from `acta config show` and, for `dispatch`, follows the dispatch skill (`plugin/skills/build/SKILL.md:17,23`).
2. Dispatch is listed as a top-level skill with its own trigger words, so a model can load it without going through build. On 2026-09-30 omp did this right after setup, with no ask and no plan. The skill hit "Step -3 — Refuse inside omp" and told the user the dispatch executor "is not used in omp". That is wrong: build runs `dispatch` as `subagent` on omp.
3. `plugin/skills/dispatch/SKILL.md` (427 lines) holds two kinds of text. One is delivery: the herdr tab, the brief, `/new` and `/goal`, reply-back, the idle watcher. The other is the loop: verify, review, fix rounds, land. The loop is already in build, `acta:review` (`plugin/skills/review/SKILL.md:38-45`) and `acta:land`, so it is written twice with small differences.

## Design

### Files

1. Delete `plugin/skills/dispatch/`.
2. New `plugin/skills/build/dispatch.md`, a reference file that build reads only when the executor is `dispatch`. It keeps the delivery part of the old skill:
   - the herdr tab on top of the build worktree (from "Step -1");
   - the hand-off doc and the brief, which still carries `SKILL: load build`;
   - `/new` then `/goal`, the omp magic keywords, the omp advisor;
   - the comprehension checkpoint and the idle watcher;
   - "Never wait for the recipient" and the two phases;
   - verify the reply from git;
   - sending a fix round to the same agent in its tab;
   - closing the tab after land.
3. Move `plugin/skills/dispatch/herdr-delivery.md` to `plugin/skills/build/herdr-delivery.md` with no text change except links.
4. Drop these sections of the old skill: Entry gate, Autonomous loop, Autonomy, Step -3, Step -2, Review, After a review, Landing, Memory sweep. Before a section is dropped, check each rule in it against build, review and land. A rule with no match there moves into `plugin/skills/build/SKILL.md`; it is never lost.
5. Update every pointer to the old skill: `internal/hook/hook.go:24,31`, `plugin/hooks/default-rules.md:4,11`, `plugin/references/house-rules.md:3`, `plugin/NOTICE`. The `dispatch` line leaves the skill list. The build line becomes: executor from `/build <executor>`, else `acta config show`, else ask.
6. The CLI does not change: `acta dispatch init` and `acta reply-back` stay as they are.

### Build flow

1. **Pick the executor.** The argument of `/build <executor>` (`subagent`, `dispatch` or `inline`) wins. With no argument, use `build_executor` from `acta config show`. With neither, ask. The argument is for that run only and is never saved.
2. **Fallbacks.** On omp, `dispatch` runs as `subagent`. With `dispatch` and no `HERDR_ENV=1` (this session is not inside a herdr pane), it also runs as `subagent`; `herdr` on PATH is not enough. Each fallback is told to the user in one line. Build never refuses because herdr is missing.
3. **Worktree.** Every executor uses the one `## Worktree` section. Dispatch no longer makes its own worktree; `dispatch.md` only adds the tab.
4. **Tasks.** `subagent` and `inline` run the task loop as today. `dispatch` hands the whole plan to one omp agent through `dispatch.md` and ends the turn (phase 1). Phase 2 starts on the reply-back and takes every number again from git, then goes to Close.
5. **Close** is the same for every executor: fast tests, the `progress.done` check, `acta:review`, fix rounds, then `acta:land` without asking when clean. Only the fix-round delivery differs: `subagent` gets a new implementer, `dispatch` prompts the same agent in its tab. With `dispatch`, the tab is closed after land.
6. **Recipient side** does not change. The omp agent in the other tab runs `acta:build` and ends with `acta reply-back` ("Reply back when a dispatch record exists").
7. **Description.** Build's description gains the trigger words "dispatch", "hand to omp" and "another tab or pane", and stays under 1024 characters.

## Testing

1. `internal/plugincheck`:
   - Delete `skill_dispatch_test.go`. Its required phrases move to a new test over `skills/build/dispatch.md`: `SKILL: load build`, `/new`, `/goal`, the idle watcher, never wait.
   - `reply_back_test.go:58,81,96` read `skills/build/dispatch.md` and `skills/build/herdr-delivery.md`.
   - New guard: `skills/dispatch/` does not exist, and no file under `plugin/` says `acta:dispatch`.
   - `skill_build_test.go` drops `acta:dispatch` from its required phrases and adds the executor order (argument, config, ask) and the no-herdr fallback to `subagent`.
   - `models_test.go:13` drops `"dispatch"` from its list.
2. `internal/hook`: the skill list has no `dispatch` line, and `hook.go` still matches `default-rules.md` word for word.
3. No new eval. The failure seen (a model loading dispatch first) cannot happen once the skill is gone. At land, run the six existing eval cases with the branch binary on PATH.
4. Gates: `scripts/test --full`, `go vet ./...`, `gofmt -l .`.

## Out of scope

- Dispatch from omp into another omp tab. On omp, `dispatch` stays `subagent`.
- Any change to `acta dispatch init`, `acta reply-back` or the `.dispatch.json` record.
- A stub `acta:dispatch` skill that points to build. The user chose to delete the skill fully.
