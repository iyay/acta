---
id: SCR-0039
hash: q5qg3ap
title: Diet for slice, build, tdd, debug, land and skill descriptions
status: raw
created: "2026-10-03 14:39:51"
schema: "1"
---
# Diet for slice, build, tdd, debug, land and skill descriptions

## Words

### 2026-10-03

Step 5 of the 2026-10-03 roadmap. Goal: acta costs fewer tokens than superpowers, mattpocock/skills and gstack for the same guarantees.

Baseline per load (tokens, bytes/4): slice 3.0k, build 3.8k (plus dispatch.md when dispatching), tdd 2.5k, debug 2.6k, land 1.4k. Skill descriptions average about 78 tokens; target 40 or less, with evals proving triggering still works. Target for one full cycle (design to land): about 11k, from about 19k.

Also close the two debt items from PLN-0077: drop SkillRule.MaxLines so each skill has one size cap, and make byte caps safe on CRLF checkouts (.gitattributes eol=lf).

Optional at the end: a real benchmark, the same small task run with acta, superpowers and mattpocock, comparing total tokens and the result.

Adapted text is rewritten, never copied 1:1, and costs fewer tokens without losing quality; each spec lowers the byte caps of the skills it touches.

## Context

## Log

### 2026-10-03

User ruling 2026-10-03: merged into the SCR-0038 brainstorm (review diet). One spec covers both; this item goes in its closes list. Status stays raw: the pre-tool hook blocks a second brainstorming status in one session.

## Open questions
