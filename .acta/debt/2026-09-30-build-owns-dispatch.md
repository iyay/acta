---
id: DBT-0051
hash: mo71at6
parent: plans/2026-09-30-build-owns-dispatch
started: "2026-09-30 22:51:02"
---
# Review NOTEs: Build Owns Dispatch Implementation Plan

- [ ] Old dispatch "delivery stuck after ~2 attempts, stop" rule has no home; herdr-delivery.md Failure handling sets no retry limit.
- [ ] Old "drift at the checkpoint: correct, re-dispatch, then stop" stop rule is not stated as a stop in build/dispatch.md.
- [ ] build/dispatch.md Stopping rule still says a non-blocking item is logged as one NOTE (memory); acta:review files NOTEs with acta debt new.
- [ ] plugin/omp/FACTS.md lines 169, 179, 182 still list dispatch in the omp --skills example.
- [ ] internal/cli/hook_test.go comment near line 12 still talks about the dispatch skill line in the index.
- [ ] Cheap stale-text MustNot guards from the old dispatch test were dropped (Contamination check, WORKTREE LANDING, bugs.md, Core Six, Important/Minor, <fixed-from>, <new-head>, docs/superpowers, herdr-pane-moves).
- [ ] TestNoDispatchSkill matches only the exact string acta:dispatch; a bare "the dispatch skill" would pass.
- [ ] Old Step -1 rule "inherited in-tree worktree: git worktree move" has no home.
- [ ] Old "escalate before the cap only for a human decision" is not in acta:review word for word.
- [ ] Old landing stop "a parent the user has called protected" is not in acta:land.
- [ ] build/SKILL.md Step 2 runs npm install while line 47 and house-rules say link dependency folders, no npm install (older clash, dispatch now goes through it).
- [ ] herdr-delivery.md says "four steps (same as dispatch.md)" but dispatch.md lists five (older mismatch).
- [ ] herdr-delivery.md says check the harness lists acta:build; in omp the skill shows as bare build.
- [ ] build/dispatch.md still says tickets are mirrored in the tracker, but the rule to mirror the fix task was dropped.
- [x] plugin/skills/setup/SKILL.md still offers dispatch when herdr is on PATH without HERDR_ENV=1; build then falls back to subagent. Setup should key on HERDR_ENV=1 alone.
- [ ] Plan Task 3 Files list omits dispatch.md, herdr-delivery.md and build_dispatch_test.go, which the fix commit also changed.
- [ ] herdr-delivery.md lead line states HERDR_ENV=1 as a fact and does not tell a reader who arrives without it to stop.
- [ ] build/SKILL.md Close says dispatch sends the fix round to its tab and closes the tab, without saying only when dispatch really ran in a pane.
