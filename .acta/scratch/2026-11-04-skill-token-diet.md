---
id: SCR-0089
hash: jz2at4q
title: Cut acta skill token cost, cheaper model for Standards reviewer
status: brainstorming
created: "2026-11-04 19:34:33"
schema: "1"
started: "2026-11-04 19:37:27"
---
# Cut acta skill token cost, cheaper model for Standards reviewer

## Words

### 2026-11-04

audit skills acta dong. biar lebih hemat tapi tetep efektif dan powerfull

ya sonnet buat standards, bikin scratch item

## Context

From a /doctor audit on 2026-11-04 (be-pmis session, acta 0.1.52 cache). Sizes are SKILL.md bytes; usage is lifetime skillUsage in ~/.claude.json.

Decided by the user:
- Review Standards-axis reviewer runs on a cheaper model tier than the orchestrator; Spec axis keeps the orchestrator model. Write it harness-neutral (no model names): each harness maps "cheaper tier" to its own model, for example sonnet in Claude Code. /usage showed reviewer subagents at 10% of usage. Note: the user's ~/.claude/CLAUDE.md rule 9 says reviewers use the orchestrator alias, so that rule needs the same change.

Proposed, not yet decided:
1. Split skills by mode, load the unused mode from references/ only when its setting matches:
   - build (12.6k + 14.8k refs): dispatch and subagent executor text when executor is inline.
   - slice (12.7k): full-depth sections (Plan Document Header, Bite-Sized, full Task Structure) when plan depth is minimal.
   - review (13.4k): Receiving findings only when findings exist.
   Est. 3-4k tokens saved per feature cycle.
2. Trim tdd (10.7k, 7 uses): Common Rationalizations, Red Flags, Example: Bug Fix, Verification Checklist overlap the Iron Law. Target ~6k chars.
3. SessionStart hook (~1.15k tokens): dedup the work list (same plan printed twice), cap at 3; drop the Lean coding guide block (same as acta:lean); move core rule 8 (multi-topic brainstorm) into shape. Est. ~600 tokens per session.
4. frame (0 uses) merge into shape; lean (0 uses) served by the hook only.

A full feature cycle loads ~66k chars (~16.5k tokens) of skill bodies; startup is ~2.6k tokens.

## Log

### 2026-11-04

Q5: user gave permission to edit ~/.claude/CLAUDE.md rule 9 directly. Done 2026-11-04: Spec-axis reviewers keep the orchestrator alias, Standards-axis reviewers use sonnet. Out of the spec.

### 2026-11-04

Round 1 (2026-11-04), user took every recommendation: Q1 one spec, plans split per part, Standards reviewer is plan 1. Q2 frame and lean untouched. Q3 evals only for skills whose text is cut or moved (tdd, build, slice, review); hook diet proven by Go tests. Q4 tdd: Rationalizations + Red Flags folded into one ~10 line block, Example and Checklist dropped, target ~6.5k chars.

### 2026-11-04

Round 2 (2026-11-04), user took every recommendation: Q6 skill text names a cheaper tier than the orchestrator with sonnet in Claude Code as the example, no config key. Q7 Lean block stays in the hook; hook diet = dedup work list cap 3 + move rule 8 to shape. Q8 default mode stays in SKILL.md, other modes move to reference files read only when the setting matches. Q9 receiving findings moves to review/receiving.md.

### 2026-11-04

Round 3 (2026-11-04): Q10 (a). Each split skill keeps only text shared by every mode in SKILL.md; each mode gets its own file (slice/full.md, slice/minimal.md, build/subagent.md, build/dispatch.md, build/inline.md) and the agent reads only the one its setting names. Reason: acta's default plan_depth is full, so keeping the default in the body would save nothing for minimal users.

### 2026-11-04

Section 1 approved 2026-11-04: five plans, each lands alone: 1 review (Standards reviewer cheaper tier + review/receiving.md), 2 build (subagent.md/dispatch.md/inline.md), 3 slice (full.md/minimal.md), 4 tdd trim ~6.5k, 5 SessionStart hook (dedup work list cap 3, rule 8 to shape). Each plan fixes plugincheck caps for files it touches.

### 2026-11-04

Section 2 approved 2026-11-04: pointer table per mode in the body (pattern of build's executor table), exactly one mode file read, frontmatter descriptions untouched. Targets: build inline <=7k, slice minimal <=7k, review without findings <=10k, tdd ~6.5k, hook ~400 tokens less per session. Pass = new plugincheck caps green, evals match main baseline, other mode files unread in eval transcripts.

## Open questions
