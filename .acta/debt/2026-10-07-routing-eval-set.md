---
id: DBT-0090
hash: uowam0l
parent: plans/2026-10-07-routing-eval-set
started: "2026-10-07 08:32:45"
finished: "2026-10-07 09:34:39"
---
# Review NOTEs: Routing eval set Implementation Plan

- [x] (high) Light routing cases have no positive grader, so an agent that only answers in text (the chat route) still passes; the fix needs the target files (src/greet.py and the others) in the scaffold so an Edit min 1 grader can work.
- [x] (high) The chat no-skill grader can never fail: every routing case sets allowed_tools: [], so Skill is never callable; decide one Skill grant for all routing cases at once, so no route is favoured.
- [x] (medium) Indonesian routing prompts run with chat_language: English, because every case copies the note-to-scratch scaffold; Indonesian cases need a scaffold that writes chat_language: Indonesian to match a real user.
