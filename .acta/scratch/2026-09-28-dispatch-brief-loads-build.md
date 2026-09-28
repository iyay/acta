---
id: SCRATCH-7
hash: wo3i
title: Dispatch brief must load the build skill, not whitelist acta commands
status: brainstorming
created: "2026-09-28"
---
User 2026-09-28, during the PLAN-18 dispatch: "harusnya bukan di brief gak sih? harusnya omp kan jalanin skill juga buat ngerjain tasknya".

What happened: the Active pane stayed empty while omp worked, because omp never ran `acta tick <task> --start`. The rule already lives in acta:build (plugin/skills/build/SKILL.md:138), but the brief never told omp to load the build skill, and its "ACTA CLI: only allowed ..." line listed only the after-commit tick, which overrode the skill.

Proposal:
- The brief template in plugin/skills/dispatch/SKILL.md gets a required line: SKILL: load build (acta:build; in omp the name is `build`) before the todo list, and follow it for every task.
- Briefs never whitelist acta commands. They may only forbid acta write commands the build skill does not call.
- herdr-delivery.md:151 then stops restating the --start rule.
- plugincheck guards for both.

Maybe part of spec 3 (SCRATCH-6, harness).


User 2026-09-29, verbatim: "ooh sama satu lagi, ini mungkin harus update skill dispatch biar bikin agent lebih disiplin dan respect terhadap instruksi. tadi omp berkali2 gak langsung reply back. kaddang diem, kadang malah nanya perlu reply back kemana"

Meaning: the dispatch skill should make the recipient agent follow its instructions more strictly. During PLAN-18/20/21 omp several times did not fire the reply-back right after the last commit: sometimes it went idle and said nothing, sometimes it asked where to reply, although the brief and the /goal both carry the literal command with the pane id filled in.

To look at in the brainstorm: why the literal command was skipped or not trusted (placeholder `<new-head-sha>` it must fill, `$HERDR_PANE_ID` escaping, compaction dropping the /goal tail, the build skill having no reply-back step); whether the reply-back should become a build-skill step or a hook instead of brief text; an orchestrator-side fallback when the agent goes idle without replying. Same brainstorm as the brief/build-skill point above.


Brainstorm 2026-09-29 Q1 (idle without reply-back): user picked B = make the recipient stricter (reply-back becomes a required last step of acta:build, command needs no hand-filled values) PLUS an orchestrator fallback: herdr tells the orchestrator pane when the agent goes idle; the orchestrator checks git and starts review itself when every task is committed.


Q1 finding 2026-09-29: herdr has no push-to-pane on idle (herdr notification only shows a popup). Fallback = orchestrator runs herdr agent wait <slug> --until idle --until done as a Claude Code background job (run_in_background), so its exit arrives as a task notification without blocking the turn. idle can blink between subagents, so the orchestrator always checks git first.
Q2 2026-09-29 (idle but tasks not all committed): user picked B = nudge omp once with a short /goal (finish the open tasks, then reply-back), re-arm the watcher; still idle with no new commit = tell the user.
