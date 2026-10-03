---
id: SCR-0037
hash: gvr3inh
title: Optional frame skill and probe mode in shape
status: brainstorming
created: "2026-10-03 14:39:49"
schema: "1"
started: "2026-10-04 06:42:05"
finished: "2026-10-04 06:51:08"
---
# Optional frame skill and probe mode in shape

## Words

### 2026-10-03

Step 3 of the 2026-10-03 roadmap. User rulings so far:

1. `frame`: a new optional skill for new products, adapted from gstack /office-hours (MIT) and rewritten. Keep: the goal question that picks startup or builder mode, the six forcing questions (startup mode only, in a separate file loaded only then), the builder questions, the premise challenge, two or three product directions, and one real assignment at the end. Drop: the shared preamble, telemetry, gbrain, browser search, the codex second opinion, the founder profile and mockups. gstack's office-hours costs about 21k tokens per load; target about 900 words plus the startup file.
2. frame's output goes into a scratch item (Context, Log and Open questions sections through acta scratch add), then shape continues from it. frame must not set status brainstorming, because the hook counts that as the session's one brainstorm.
3. `probe`: a mode of shape, from mattpocock's grilling skill (MIT), rewritten. A decision tree asked in rounds: every question that can be asked now goes in one round, each with a recommended answer, five at most per round. Facts are looked up by subagents; decisions belong to the user. The text lives in its own file, loaded only when probe is used. Trigger: "probe" or "grill me".
4. Boundary: frame answers why, for whom and what; shape answers how.
5. Diet shape in the same spec (about 4.4k tokens now, target about 2k). Delete plugin/skills/shape/spec-document-reviewer-prompt.md, which no skill links.
6. Adapted text is rewritten, never copied 1:1, and costs fewer tokens without losing quality.

## Context

Sources found on this machine (2026-10-04):
- gstack office-hours (MIT, Garry Tan): /Users/iyay/Nayakatara/digimov/ai-eceran/.claude/skills/gstack/office-hours/ (gstack 06ed920). SKILL.md 11871 words; sections/phase-2a-startup-diagnostic.md 2135 words (six forcing questions: demand reality, status quo, desperate specificity, narrowest wedge, observation and surprise, future-fit); sections/phase-2b-builder-brainstorm.md 524 words; sections/design-and-handoff.md 5437 words (assignment lives here). Goal question at SKILL.md:598; premise challenge :751; alternatives :926.
- mattpocock grilling (MIT, Matt Pocock): ~/.claude/plugins/marketplaces/mattpocock/skills/productivity/grilling/SKILL.md, 319 words.
- shape today: SKILL.md 2696 words / 17458 bytes cap; spec-document-reviewer-prompt.md 235 words, linked by nothing.
- Guards that a shape diet must keep or change on purpose: internal/plugincheck/skill_shape_test.go (Must list, description word "brainstorm", every scratch add line has --section log, Bounded graph edges), evals_test.go (4 shape eval cases), budget_test.go caps.
- A new skill must join hook.Skills in internal/hook/hook.go:19 and NOTICE.

## Log

### 2026-10-04

Q1 when frame runs: A. frame is its own skill, runs when the user asks (frame X, office hours, is this worth building). shape only offers it in one line when the work is a new product and the scratch item has no Context. Never forced.

### 2026-10-04

Q2 after frame: A. frame asks "continue with shape SCR-xxxx now?"; on yes shape runs in the same session (frame does not use the one-brainstorm slot). The assignment may need real-world work first, so the user picks when.

### 2026-10-04

Q3 probe scope: A. probe works on Bounded and Architectural; when on, it replaces shape's one-question-at-a-time step; done when the frontier is empty, then shape goes on to approaches or design. Architectural appends each round's answers to Log. Grilling an existing spec or plan outside shape is out of scope.

### 2026-10-04

Approach for the shape diet: A. One file, one flow: the per-path checklist is the only flow; the dot graph, Red Flags, Anti-Pattern and generic design advice go; TestShapeBoundedReviewLoopsToShortSpec is rewritten to check the Bounded checklist loops changes back to the short spec. About 1300 words. Rejected: B keep graph (about 1650 words), C split architectural.md (agent may skip the second file that evals depend on).

### 2026-10-04

Design section 1 (frame skill) approved:
- Files: plugin/skills/frame/SKILL.md (about 900 words) and plugin/skills/frame/startup.md (six forcing questions and pushback patterns, read only in startup mode, about 700 words).
- Triggers (user ruling: no "office hours"): frame, frame this idea, is this worth building, should I build this, new product idea, validate an idea, who is this for. A feature in an existing product stays with shape.
- Flow: scratch item (named one, else acta scratch new; never status brainstorming); goal question picks startup or builder mode (switch to startup when customers or revenue show up); questions one at a time because pushback builds on the last answer, skip what the prompt answered, "just do it" or a full plan jumps to premises; premise challenge as numbered premises to agree or disagree; 2-3 product directions (narrowest, ideal, lateral), recommend one, user picks, no tech approaches; one real-world assignment; write Context, Log, Open questions with acta scratch add; ask "continue with shape SCR-xxxx now?".
- Dropped from gstack: preamble, telemetry, gbrain, web search, codex second opinion, founder profile, mockups, design doc and handoff, AskUserQuestion format rules.
- Also: frame joins hook.Skills (internal/hook/hook.go:19), gstack and mattpocock credits in plugin/NOTICE, new byte caps in budget_test.go.

### 2026-10-04

Design section 2 (probe mode) approved:
- File plugin/skills/shape/probe.md (about 250 words, from grilling 319). shape SKILL.md gets one line: user says probe or grill me, read probe.md and use it in place of one-question-at-a-time.
- Decision tree; frontier = decisions whose prerequisites are settled. One round = the frontier, at most 5; questions that unblock the most go first, the rest wait. A question that depends on one still open waits.
- Format: **Q1. title**, body, then "Recommended: ...". No emoji.
- Facts come from subagents under shape's subagent model rule; do not block, only dependent questions wait. Decisions belong to the user.
- Architectural appends each round's answers with acta scratch add SCR-0001 --section log.
- Done when the frontier is empty and the user confirms shared understanding; then shape goes on to approaches or design.
- Dropped from grilling: emoji format, rules between questions. Text rewritten, not copied.

### 2026-10-04

Design section 3 (shape diet and proof) approved: new shape SKILL.md about 1300 words (from 2696) keeps description word brainstorm, short HARD-GATE, three paths with heavier-path, ratchet and sensitive-change rules, a frame offer line and a probe line, per-path checklists as the only flow (Bounded changes requested loops to the short spec), Architectural extras (step 0, Log appends, one per session with a/b/c, subagent models), spec rules. Dropped: dot graph, Red Flags, Anti-Pattern, generic design advice, spec-document-reviewer-prompt.md. Tests: Must list kept, MaxLines lowered, graph test rewritten against the checklist, byte caps lowered and added for frame and probe. Proof: word counts before and after; 4 old shape evals green; new evals frame-no-brainstorming-status and probe-round; scripts/test --full and scripts/eval with the branch binary; patch version bump.

### 2026-10-04

Spec review changes (user, 2026-10-04): 1. frame callable as /acta:frame (omp: frame); plugincheck keeps frame free of user-invocable: false. 2. probe as a user setting: key questions (one default, probe) in ~/.acta/config.yaml, user only; acta config show/set --questions; hook Voice line '- Questions: probe.' only when probe; setup asks it; in-session probe/grill me and one at a time still switch.

## Open questions
