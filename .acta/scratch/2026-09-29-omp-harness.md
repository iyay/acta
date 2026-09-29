---
id: SCR-0015
hash: mgfekh8
title: omp-harness
status: raw
created: "2026-09-29"
---
# omp harness: evals and brainstorm block

Split out of SCRATCH-6 on 2026-09-29. omp does not run Claude Code shell hooks, so the SCRATCH-6 session state, reminder and second-brainstorm block do not work in omp. The claude plugin eval suite also covers Claude Code only.

Idea: give plugin/omp/index.ts a handler for the omp tool-call event (check first that the event exists), and add an omp eval run for the same scenarios.
