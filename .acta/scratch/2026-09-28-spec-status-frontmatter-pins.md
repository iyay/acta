---
id: SCRATCH-8
hash: mdj2
title: Spec frontmatter status pins the status and hides a finished spec
status: raw
created: "2026-09-28"
---
User 2026-09-28: "btw SPEC-6 kenapa gantung ya?" SPEC-6 showed `approved (frontmatter)` with 8/8 tasks done and landed (deaeeb9), so it never moved to Done. Fixed by hand with `acta set specs/2026-09-26-review-debt-design status done`.

Cause: a written `status:` in a spec's frontmatter wins over the status derived from its plans. Five old specs carry one (file-contract, pm-plugin, review-debt, tui, tui-panes); land set four to done and missed SPEC-6. New specs (SPEC-10, SPEC-11) have no status line and derive it.

Proposal:
- Remove the `status:` line from the five old specs so status is always derived.
- Maybe a board problem: "frontmatter status approved but every task done" (or refuse a written approved/in-progress/done on specs that have plans).
