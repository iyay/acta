---
id: SCR-0022
hash: ip1j9ca
title: omp still looks for house rules under the old pm-board path
status: raw
created: "2026-09-29"
---
User 2026-09-29: "catet temuan yang aneh tadi".

Agent notes: during the PLAN-32 dispatch (omp slug short-ids), omp first tried to read ~/Nayakatara/pm-board/plugin/references/house-rules.md, a stale pre-rename path, then found the right file by search. The repo plugin/ has no pm-board mention. Hits outside the repo: ~/.omp/agent/managed-skills/omp-dispatch-ticket-retry/SKILL.md, plus omp agent.db / history.db (omp memory). Find where omp gets the old path and fix it.
