---
id: SCR-0032
hash: vcbj3s9
title: 'One implementation skill: build owns dispatch'
status: brainstorming
created: "2026-09-30"
schema: "1"
started: "2026-09-30"
finished: "2026-09-30"
---
# One implementation skill: build owns dispatch

## Words

### 2026-09-30

User words (2026-09-30): "kenapa butuh 2 skill buat implementasi? kenapa gak pake build aja? baca config, kalo default-nya dispatch ya dispatch.. atau `/build dispatch|subagent|inline`"

What happened: after setup in omp saved build_executor dispatch, omp loaded acta:dispatch on its own (no ask, no plan). The dispatch skill hit Step -3 "Refuse inside omp" and told the user the dispatch executor "is not used in omp", which is wrong: acta:build runs dispatch as subagent on omp (plugin/skills/build/SKILL.md:23).

Idea: build is the only implementation skill. Dispatch stops being a top-level skill and becomes reference files that build reads only when the executor is dispatch (like build/implementer-prompt.md). `/build dispatch|subagent|inline` overrides the config for one run.

## Context

## Log

### 2026-09-30

Q1 scope: user picked B (move + dedupe): build/dispatch.md keeps only delivery (herdr, tab, worktree, brief, /new + /goal, reply-back, idle watcher); verify, review, fix rounds, land live only in build for every executor.

### 2026-09-30

Q2 dispatch without herdr: user picked A: build runs as subagent and says so in one line, same as on omp; build never refuses for a missing herdr. Also settled: /build <executor> arg beats config, config beats asking; the arg is for that run only, never saved.

### 2026-09-30

Q3 acta:dispatch skill: user picked A: delete it fully; build's description gains the trigger words (dispatch, hand to omp, another tab) so such asks load build.

### 2026-09-30

Design section 1 approved (file shape): delete plugin/skills/dispatch/; new build/dispatch.md holds delivery only (worktree+tab, hand-off doc+brief with SKILL: load build, /new then /goal, omp keywords+advisor, checkpoint+idle watcher, never-wait two phases, verify reply from git, fix round to same agent, close tab after land); herdr-delivery.md moves to build/ unchanged; loop sections dropped after each rule is checked against build/review/land (missing ones move into build/SKILL.md); hook.go:24,31, default-rules.md:4,11, house-rules.md:3, NOTICE and plugincheck tests follow; acta dispatch init and acta reply-back CLI unchanged.

### 2026-09-30

Design section 2 approved (build flow): executor = /build arg, else config, else ask; omp or no herdr turns dispatch into subagent with a one-line note; one ## Worktree for all executors; dispatch hands the whole plan to one omp agent via dispatch.md and ends the turn, phase 2 on reply-back re-derives from git; Close same for all (fast tests, progress check, acta:review, fix round, acta:land), only fix-round delivery differs (new implementer vs prompt same agent), dispatch closes the tab after land; recipient side unchanged; build description gains dispatch trigger words under 1024 chars.

### 2026-09-30

Design section 3 approved (tests): skill_dispatch_test.go deleted, its Must phrases move to a test over build/dispatch.md; reply_back_test.go:58,81,96 point at skills/build/; new guard: no skills/dispatch/ and no acta:dispatch anywhere in plugin/; skill_build_test drops acta:dispatch, adds executor order and no-herdr fallback phrases; models_test drops dispatch; hook test: no dispatch line, hook.go matches default-rules.md; no new eval, run the 6 existing at land with branch binary; gates scripts/test --full, go vet, gofmt.

## Open questions
