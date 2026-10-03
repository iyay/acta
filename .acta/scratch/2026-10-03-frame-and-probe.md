---
id: SCR-0037
hash: gvr3inh
title: Optional frame skill and probe mode in shape
status: raw
created: "2026-10-03 14:39:49"
schema: "1"
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

## Log

## Open questions
