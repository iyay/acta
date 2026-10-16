---
id: DBT-0125
hash: szvf8xo
parent: plans/2026-10-16-minimal-plan-task-boxes
---
# Review NOTEs: Minimal plan tasks always count toward progress Implementation Plan

- [ ] (medium) Started minimal plans with box-less tasks, like PLN-0145, never get finished: tick refuses box-less tasks and TaskDates needs every box ticked, so they stay open until set done by hand.
- [ ] (medium) Reopening a plan to approved or draft keeps its old finished date (DatesFor clears it only for working statuses), so its box-less tasks read done again; clear finished on any non-closed status.
