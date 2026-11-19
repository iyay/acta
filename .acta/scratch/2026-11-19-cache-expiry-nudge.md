---
id: SCR-0097
hash: xfvvkg6
title: Nudge when the prompt cache has likely expired
status: done
created: "2026-11-19 20:48:13"
schema: "1"
finished: "2026-11-20 05:22:21"
---
# Nudge when the prompt cache has likely expired

## Words

### 2026-11-19

boleh. tapi yg higine benchmark kayaknya belum perlu ya?

(Reply to: note the expired prompt cache nudge idea from Chisle into acta scratch.)

## Context

Idea borrowed from Chisle 3.14.2 (github.com/JayPokale/Chisle), hooks/chisle-mode-tracker.js lines 34-62, found by reading the repo on 2026-11-19.
Its UserPromptSubmit hook warns when the session sat idle more than 1 hour with 150k or more context tokens, since the next turn then pays full input price again.
acta could add the same line to the text acta hook builds for UserPromptSubmit (internal/hook).
The user said the benchmark hygiene idea from the same review is not needed yet.

## Log

## Open questions
