---
id: SCR-0082
hash: t5ydsnj
title: Cleanup command for stale planning items
status: brainstorming
created: "2026-10-31 07:10:17"
schema: "1"
started: "2026-10-31 07:10:20"
finished: "2026-10-31 07:24:03"
---
# Cleanup command for stale planning items

## Words

### 2026-10-31

We need a command like `cleanup` whose job is to check, clean and tidy every scratch, bug, debt and spec that is stale, no longer relevant, dropped, closed, already implemented, invalid, and so on.

## Context

## Log

### 2026-10-31

Facts at start (2026-10-31, acta list --all): open bug 1 (BUG-0037), open debt files 73 with 111 open lines, raw scratch 13, plan in-progress 1. 41 scratch files still say status: brainstorming in frontmatter while the board derives specced. Path: Architectural.

### 2026-10-31

Round 1 answers (user: setuju to all): Q1 both a skill acta:cleanup for judgment and an acta cleanup CLI for mechanical work (find candidates, fix frontmatter). Q2 clean means change status (dropped, wontfix, done, fixed) with a short reason; never delete files. Q3 agent proposes a list per kind with evidence (file:line or commit); user says yes, then commit. Q4 first scope: debt, raw scratch, open bugs, stuck plans; specs only when their plan is dead. Q5 manual only for now; no auto trigger at land.

### 2026-10-31

Round 2 answers (user: setuju saran lo): Q1 reason for debt lines is the existing suffix like (stale); scratch, bug and plan get a one-line frontmatter field reason:, settable with acta set reason. Q2 acta cleanup reports candidates by default and writes nothing; --fix only does zero-judgment fixes in one commit: written status that disagrees with the derived one, and plans with every task ticked but status not done. Q3 no age-based auto-drop; age only sorts, the agent checks each item. Q4 the spec closes SCR-0064. Q5 skill sends one message per kind, at most 10 numbered items each with proposal and evidence; user answers ok or the numbers to skip.

### 2026-10-31

Round 3 answers (user: ok). Correction first: the 41 written-brainstorming scratch items are all linked to a spec, so the board already shows specced; --fix only closes plans whose tasks are all ticked. Q1 a raw scratch item built without a spec gets a new written status done plus a reason naming the commit; the TUI gets a Done tab. Q2 a stuck plan is in-progress with open tasks and no commit naming it for 14 days; proposal is dropped plus reason, and its spec is dropped too when no other plan of it is alive. Q3 one read-only subagent per kind runs in parallel and returns proposals with evidence; the main agent re-checks the evidence before showing the user. Q4 an open bug with a fix commit becomes fixed plus fixed_in; one that cannot be reproduced and has no fix commit becomes wontfix plus reason. Q5 a debt line already done becomes [x] with (done in <sha>); a line whose code is gone or no longer relevant becomes [-] with (stale) or a short reason.

### 2026-10-31

Round 4 answers (user: ok): Q1 draft or approved specs with no plan, older than 14 days, are checked; proposal is dropped plus reason, or keep. Q2 brainstorming scratch items with no spec link are checked with the raw ones, by the same 14 days without activity rule. Q3 the wiki decision scratch-dropped-vs-specced gets updated at build: done means built without a spec, and its reason must name the commit. Frontier empty.

### 2026-10-31

Approach A approved (skill plus candidate CLI) and section 1 approved (user thumbs up): acta cleanup prints candidates per kind, oldest first, --json for the skill. Candidates: open debt lines, raw or brainstorming scratch with no spec, open bugs, plans stuck 14 days, specs with no plan for 14 days. --fix in one commit removes written statuses the board already ignores and marks done a plan whose tasks are all ticked but does not show done. Never deletes files.

### 2026-10-31

Section 2 approved (user thumbs up): new one-line frontmatter field reason on scratch, bug, spec and plan, set with acta set <id> reason; acta show and the TUI detail pane show it. Scratch gets a written status done that, like dropped, beats a spec link; the TUI Scratchpad gets a Done tab. acta set on a debt line (DBT-n.mm) accepts status done or wontfix and writes [x] or [-] with the reason as a suffix. All writes auto-commit like acta set today.

### 2026-10-31

Section 3 approved (user thumbs up): skill acta:cleanup runs acta cleanup, shows mechanical fixes, applies --fix only after a yes; then acta cleanup --json and one read-only subagent per kind in parallel, each item gets keep, done, dropped, wontfix or fixed plus evidence and reason; subagents never write. The main agent re-checks evidence (sha exists, file:line holds what is claimed); weak evidence becomes keep. One message per kind, at most 10 numbered items; user answers ok or numbers to skip. acta cleanup apply reads the approved decisions on stdin and writes one commit per approved batch. Items that live only in another worktree are skipped.

### 2026-10-31

Section 4 approved (user thumbs up): Go tests for candidates (14-day rule, worktree skip), --fix, apply one commit per batch, acta set reason, scratch done beating a spec link, debt line [x]/[-] with suffix; TUI Done tab and reason in the detail pane; new skill plugin/skills/cleanup listed in the hook skill list and default-rules.md, passing plugincheck; one eval where the agent proposes with evidence and waits for a yes before writing. Wiki page scratch-dropped-vs-specced updated at build. Spec closes SCR-0064. Likely two plans: CLI and data first, skill and eval second.

## Open questions
