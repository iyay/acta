---
id: SCR-0038
hash: g9t7khz
title: 'Review diet: fewer tokens per review, same quality'
status: brainstorming
created: "2026-10-03 14:39:50"
schema: "1"
started: "2026-10-03 20:33:23"
finished: "2026-10-03 20:58:15"
---
# Review diet: fewer tokens per review, same quality

## Words

### 2026-10-03

Step 4 of the 2026-10-03 roadmap. Measured on PLN-0077: subagents used about 1.07M tokens; review took 54%; the polish of four tiny fixes took about 308k (29%).

Ideas from anthropics/claude-code plugins/code-review. That repo is "All rights reserved", so take ideas only, no copied text:
1. A high-signal bar: BLOCKERs, plus NOTEs that cite a written rule or a concrete failure. No taste NOTEs; at most five NOTEs per reviewer.
2. An explicit false-positive list: issues already on the parent branch, what a linter catches, nitpicks, issues silenced in code, wishes for more coverage.
3. Diff first: open other files only to confirm a finding; check callers only when a signature or behaviour changed.
4. A validator subagent per BLOCKER before a fix round starts, since a fix round is the most expensive step.
5. Trim code-reviewer.md (the generic checklist, "acknowledge strengths") and move "Receiving findings" (about 460 words) to a reference file.

User rulings that stay: reviewers run on the orchestrator's model (sonnet only for implementers), a polish always gets two reviewers, three rounds at most.

## Context

## Log

### 2026-10-03

User ruling 2026-10-03: SCR-0039 (skill diet) merges into this brainstorm. One spec with parent scratch/2026-10-03-review-diet and SCR-0039 in closes. The spec may still split into more than one plan.

### 2026-10-03

Q1 proof of same quality: C. Gate at land = old eval cases green + new eval cases for new behaviour + plugincheck + word counts before/after. One replay at the end on the PLN-0077 range (64b934f..c8f5e24), old vs new review, for the token number. All of this runs in the acta repo only; nothing new ships to user repos except leaner skill text and the new review rules. Cross-plugin benchmark stays optional, outside this spec.

### 2026-10-03

Q2 BLOCKER validator: B. One validator subagent per round, on the orchestrator's model, only when a round has BLOCKERs. It merges duplicates from the two reviewers, re-runs each BLOCKER's concrete input, and marks each valid or dropped before the fix round starts.

### 2026-10-03

Q3 polish brief: A. Polish reviewers get only the polish range and the [fix] NOTE list, no spec or plan. Spec axis checks each NOTE was done as written; Standards axis checks nothing broke on the changed lines. Polish NOTEs still sort into [debt] or [note] only. Two reviewers stay.

### 2026-10-03

Q4 Receiving findings: A. Rewrite to about five lines in review/SKILL.md (verify each finding against the code; push back with technical reasons; ask about unclear items before doing any; no praise or thanks, state the fix; one item at a time, test each). No new file. Keep the '## Receiving findings' heading the test pins.

### 2026-10-03

Q5 scope: B. SKILL.md of slice, build, tdd, debug, land, review; all 12 descriptions; plus files read every cycle: build/implementer-prompt.md, review/code-reviewer.md, build/dispatch.md, build/herdr-delivery.md. On-demand reference files (tdd/writing-good-tests.md, debug refs) stay out. Shape stays out (SCR-0037). User concern: dispatch is the biggest cost; wants it much faster and leaner.

### 2026-10-03

Q6 dispatch: B. Merge dispatch.md and herdr-delivery.md into one rewritten file, and add a Go command acta dispatch send --plan <path> that runs the whole herdr sequence (find or create tab, launch omp, wait ready, /new, confirm, rename, dispatch init, /goal, confirm goal, checkpoint read). Gotchas live in code with tests against a fake herdr. Skill target about 1.5-2k tokens; orchestrator steps from about 10 turns to 1. Its own plan inside this spec.

### 2026-10-03

Q7 brief: A. acta dispatch send writes the brief itself from the plan (tasks with verify lines and waves, spec path) and the dispatch record (worktree, branch, parent, base), adds absolute paths of house-rules.md and both MEMORY.md, the gate command and the REPLY-BACK line, and puts 'ultrathink orchestrate' plus REPLY-BACK in the /goal. Orchestrator may add job facts with --note. Fix rounds and polish use the same command with --round (no /new; TICKETS = the tasks of the plan's '## Fix round <n>' section). The command refuses a plan task with no verify line.

### 2026-10-03

Q8 checkpoint: A. acta dispatch send waits about 20s after /goal, reads the pane once, checks every task id shows. Prints one line: checkpoint ok / checkpoint drift: missing <ids> / checkpoint unconfirmed. Pane text printed only on drift. Wave check dropped: the brief now carries the plan's waves.

### 2026-10-03

Q9 launch: A (user idea: no /new). First dispatch launches a fresh omp with no argument (new process = new session, so no /new, no 'New session started' check, no re-rename), then /goal via herdr agent prompt once omp is ready; goal mode stays. omp '/goal x' as a launch argument does not start goal mode (omp v18.4.4), so the goal cannot ride the launch. Fix rounds: omp still runs in the tab, /goal via herdr prompt.

### 2026-10-03

Approach: A. One spec, three serial plans (all touch budget_test.go): 1) acta dispatch send + close and one merged dispatch file; 2) review diet; 3) slice, build SKILL.md, tdd, debug, land, implementer-prompt, descriptions, PLN-0077 debt. Also in the design with no objection: diff-first reviewers, false-positive list, at most 5 NOTEs per reviewer citing a written rule or concrete failure, generic checklist dropped, descriptions <= about 160 bytes with trigger evals green, MaxLines dropped, .gitattributes eol=lf, acta dispatch close by slug.

### 2026-10-03

Design section 1 approved (acta dispatch send + close): send --plan [--round] [--note-file|-] in the worktree; reads tasks/verify/waves (round: last '## Fix round <n>'; polish: [fix] NOTEs from the note); refuses a task with no verify; writes record + brief .claude/dispatch/<slug>-brief.md; tab by slug reuse or create (--workspace, --no-focus, fresh omp, rename); waits ready max 60s; clears held goal; sends /goal (summary, ultrathink orchestrate, brief path, REPLY-BACK), waits for goal mark, one retry; 20s then one read, checkpoint line; prints about 6 lines incl. the watcher command. Exit 0 ok/unconfirmed, 1 drift (pane text), 2 delivery failed. close: pane from slug, herdr pane close. dispatch.md + herdr-delivery.md merge into one build/dispatch.md about 1.5-2k tokens. Tests: Go with scripted fakeHerdr; no eval (no herdr in sandbox).

### 2026-10-03

Design section 2 approved (review diet): code-reviewer.md rewritten to about 0.5k tokens (axis/range/plan, diff first, do-not-report list, max 5 NOTEs citing a rule or concrete failure, no subagents, output model/BLOCKERs/NOTEs with bucket tags/CLEAN or BLOCKED; drop generic checklist, strengths, Answers) with a polish block and a validator block. review/SKILL.md to about 1.6k: validator before a fix round on the orchestrator model, not a round; all dropped = CLEAN; dropped-but-real becomes a NOTE; Receiving findings 5 lines; polish dispatch via acta dispatch send --round polish; other rules kept, wording tightened. Proof: old evals green; new eval cases: BLOCKED report leads to validator before fix round, parent-branch finding goes to acta:bug; replay on PLN-0077 range after plan 2 lands, numbers in plan 2 Review notes.

### 2026-10-03

Design section 3 approved (other skills): rewrite targets slice 3.0k->1.8k, build SKILL.md 3.8k->2.0k, tdd 2.5k->1.5k, debug 2.6k->1.6k, land 1.4k->0.9k, implementer-prompt 1.9k->1.0k (tokens). Cycle about 18k -> about 12.2k (shape 4.4k left for SCR-0037). Every plugincheck Must rule survives; wording changes only with the test phrase in the same commit; no grader loosened. budget_test.go caps lowered in the same commit. Word counts before/after in Review notes. Descriptions <= about 160 bytes, descriptionCaps lowered, existing evals green, no new trigger evals. Debt: drop SkillRule.MaxLines; .gitattributes '*.md text eol=lf'. Patch bump per plan, 0.1.5 after three.

### 2026-10-03

Spec written: SPC-0069 .acta/specs/2026-10-03-review-and-skill-diet-design.md (af19620). Waiting for user review.

### 2026-10-03

User approved spec SPC-0069 on 2026-10-03. Next: acta:slice for plan 1 (dispatch send/close).

### 2026-10-03

Plan 1 written: PLN-0079 .acta/plans/2026-10-03-dispatch-send.md (45fa750), 5 tasks in 3 waves. Plan-time rulings: --rules flag carries the absolute house-rules.md path (the Go binary cannot find the plugin dir); exit codes 0 ok/unconfirmed, 1 bad input, 2 delivery failed, 4 drift (spec fixed 834bdde); waits are package variables so tests run in ms; agent found but working = refuse, nothing sent. Waiting for user yes (user CLAUDE.md rule 3 overrides minimal no-wait).

### 2026-10-03

User approved PLN-0079 on 2026-10-03. Build starts with executor dispatch.

## Open questions
