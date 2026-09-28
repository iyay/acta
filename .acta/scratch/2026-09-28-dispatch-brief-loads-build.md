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


Q3 2026-09-29 (reply-back shape): user picked B = new CLI command acta reply-back. At dispatch the orchestrator writes a small record in the worktree (orchestrator pane, base sha, plan path). acta:build last step: when a dispatch record exists, run acta reply-back; the command builds and sends /acta:review <base>..<HEAD> itself. No placeholders or escaping left for the recipient.


Q4 2026-09-29 (reply-back with open tasks): user picked B = acta reply-back refuses with exit 1 and lists the open tasks; acta reply-back --blocked "<reason>" sends the reason to the orchestrator as prose instead of a review request.


Design section 1 (components) approved 2026-09-29: (1) record <worktree>/.acta/.dispatch.json, git-ignored via EnsureGitignore, holds orchestrator pane, base sha, plan path, round slug; (2) acta dispatch init --pane <id> --plan <path> [--round <slug>] run by the orchestrator in the worktree before /goal, base = HEAD then, re-run each fix round; (3) acta reply-back [--blocked "reason"]: reads the record, checks plan progress with the acta show engine, done<total = exit 1 listing open tasks, complete = herdr agent prompt <pane> "/acta:review <base>..<HEAD> - plan <path>, round <slug>, pane <own pane>", --blocked sends prose, no record = exit 1, herdr failure = exit 3; (4) acta:build last step: when .acta/.dispatch.json exists run acta reply-back; brief and /goal only say load skill build; (5) orchestrator watcher: herdr agent wait <slug> --until idle --until done as a background job; on notify check git; complete = review; not complete = nudge once and re-arm; still idle = tell the user.


Design section 2 (skills and brief) approved 2026-09-29: (1) dispatch/SKILL.md brief template gets a required line SKILL: load build (omp: build) and tdd before the todo list; briefs never whitelist acta commands, may only forbid acta write commands build does not call; the REPLY-BACK line and /goal tail become one sentence: after the last commit the build skill runs acta reply-back; (2) herdr-delivery.md drops its --start/tick restates, adds acta dispatch init before /goal and the watcher step after the checkpoint; (3) dispatch "Never wait" gets one exception: a background herdr agent wait job; foreground wait, polling and sleep stay banned; (4) build/SKILL.md last step: when .acta/.dispatch.json exists run acta reply-back, exit 1 = finish open tasks, real blocker = --blocked; a dispatch recipient never stops before acta reply-back exits 0; (5) plugincheck text guards for the SKILL line, no acta whitelist, and the build reply-back step.


Design section 3 (errors and tests) approved 2026-09-29: dispatch init exit 1 outside git, empty --pane/--plan, unknown plan; overwrite on re-run. reply-back exit 1 no/broken record, open tasks (lists ids), empty --blocked; exit 3 herdr missing/fails (herdr stderr shown). Record is untrusted: pane must match ^[A-Za-z0-9]+:[A-Za-z0-9]+$, base must be a valid hex sha, plan must be a path inside the acta root; herdr runs through exec args, never a shell. Tests red first: init writes HEAD and the gitignore line, second init moves base; reply-back open/complete (fake herdr on PATH gets exact args)/blocked/missing/broken/bad pane/bad sha send nothing; three plugincheck text guards. Watcher not auto-tested; checked on the next real dispatch.
