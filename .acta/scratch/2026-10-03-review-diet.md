---
id: SCR-0038
hash: g9t7khz
title: 'Review diet: fewer tokens per review, same quality'
status: raw
created: "2026-10-03 14:39:50"
schema: "1"
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

## Open questions
