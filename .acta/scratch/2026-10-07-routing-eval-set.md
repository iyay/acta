---
id: SCR-0050
hash: gyxidsy
title: Eval set for ambiguous routing
status: brainstorming
created: "2026-10-07 06:37:43"
schema: "1"
started: "2026-10-07 06:38:41"
finished: "2026-10-07 06:40:28"
---
# Eval set for ambiguous routing

## Words

### 2026-10-07

# Eval set for ambiguous routing

Idea, 2026-10-07: build an eval set that measures how well the skill rules route an unclear prompt. Routes: scratch, Architectural brainstorm, bounded or light fix, plain chat, and second-brainstorm detection.

Why: SCR-0004 (Jev routing spike) wants to benchmark rules against Jev on the existing eval scenarios. Those cannot serve as the benchmark:
- The five routing cases (note-to-scratch, side-idea-to-scratch, brainstorm-files-scratch-first, one-file-fix-no-brainstorm, second-brainstorm-choices) name the route in the prompt. They test whether the skill follows a clear route, not whether it picks one from an unclear prompt.
- All prompts are English. Plugin users write in other languages too, and Jev says it handles languages other than English less well.
- No case covers plain chat.
- Each case has runs: 1, which gives no accuracy number.

Rough shape: 15 to 20 prompts that do not say which route they want, with more than one language, every route covered, and several runs each. The output is a baseline accuracy for the rules alone.

This item blocks SCR-0004. If the baseline is already high, SCR-0004 can be dropped with "rules are enough".

## Context

## Log

### 2026-10-07

2026-10-07 probe round 1, all accepted: (1) each prompt is a normal case in plugin/evals with a route grader, accuracy from pass rate; (2) graders check action traces, not the reply's claim; (3) about half English, half other languages (Indonesian first) plus 2-3 code-switched; (4) 16 prompts, 4 each for scratch, Architectural, light fix, chat, plus 2 second-brainstorm, runs: 3, tag routing; (5) rules are enough when overall accuracy >= 90% and no route is below 75%; fixed in the spec before the first run.

## Open questions
