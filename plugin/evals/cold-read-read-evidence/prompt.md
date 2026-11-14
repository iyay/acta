---
name: cold-read-read-evidence
description: Resolve distinguishes valid reading evidence from missing, empty or generic reports.
tags: [shape, cold-read, claude-only]
runs: 1
max_turns: 20
timeout_seconds: 900
allowed_tools: [Skill, Read]
---
Resume acta:shape after completed cold reads for the written, committed, self-reviewed specs .acta/specs/evidence-valid.md, evidence-missing.md, evidence-empty.md and evidence-generic.md in that folder. Fresh reviewers have finished. Their results are reports/evidence-valid.txt, reports/evidence-missing.txt, reports/evidence-empty.txt and reports/evidence-generic.txt. Continue from those results. I have approved neither edits nor planning. Stop when you need my decision.
