---
id: SCR-0008
hash: mdj2pr8
title: Spec frontmatter status pins the status and hides a finished spec
status: brainstorming
created: "2026-09-28"
---
User 2026-09-28: "btw SPEC-6 kenapa gantung ya?" SPEC-6 showed `approved (frontmatter)` with 8/8 tasks done and landed (deaeeb9), so it never moved to Done. Fixed by hand with `acta set specs/2026-09-26-review-debt-design status done`.

Cause: a written `status:` in a spec's frontmatter wins over the status derived from its plans. Five old specs carry one (file-contract, pm-plugin, review-debt, tui, tui-panes); land set four to done and missed SPEC-6. New specs (SPEC-10, SPEC-11) have no status line and derive it.

Proposal:
- Remove the `status:` line from the five old specs so status is always derived.
- Maybe a board problem: "frontmatter status approved but every task done" (or refuse a written approved/in-progress/done on specs that have plans).

2026-09-29, brainstorm started. User: "semuanya aja lah sekalian" (do SCRATCH-6, 8 and 14 together). Scope for this brainstorm: SCRATCH-8 + SCRATCH-14 + a new finding; SCRATCH-6 runs as its own Architectural brainstorm elsewhere.

New finding: SPEC-16 shows draft (derived) while PLAN-25 that built it is done and landed (bb71f0f). PLAN-25 has parent: debt/2026-09-29-skill-paths, so no plan links to SPEC-16. A plan has one parent only.

SCRATCH-14 shows raw although PLAN-24 fixed it: SPEC-15 has one parent (SCRATCH-12), and scratch statuses are only raw, brainstorming, dropped.

Common root: acta links each item to one parent, while real work often closes several items at once.

Q1 (how one piece of work closes several items): user picked option 2. parent stays one and sets the place in the tree; a new frontmatter field closes: lists the items the work also finishes (for example closes: [SCRATCH-14, DEBT-17.1]); a plan's **Spec:** line always links the plan to its spec, even when parent is a debt. Rejected: parent as a list (item shows in several tree places), and a manual scratch done status only.

Q2 (how a closed item's status changes): user picked mixed. Scratch and spec items named in closes: derive their status from the closer (no file write). DEBT items named in closes: are ticked at land by acta:land reading the field, because debt items are checkboxes in a file.

Q3 (written status on specs): user picked derived wins. A spec with any plan (through Spec:, parent or closes) ignores a written status and the board shows a problem "written status X ignored, derived Y". A spec with no plan may still carry a written draft or approved.

Section 1 approved (closes field): closes: is a YAML list in the frontmatter of specs, plans and bugs; entries are short ids or path ids resolved like acta show; targets may be scratch, spec or debt items only; an unknown id or a wrong kind is a board problem on the item that wrote closes; parent: stays one and sets the tree place; a plan's **Spec:** line always links the plan to its spec even when parent is set, so the spec counts the plan and its tasks.

Section 2 approved (status and display): a scratch named in any closes shows specced (derived), written dropped still wins, no new status; a spec named in a plan's closes counts that plan like a Spec: link; a spec named in a spec's or bug's closes follows the closer's status; debt items are not derived and change only when acta:land ticks them; a spec with any plan ignores a written status and shows the problem "written status X ignored, derived Y"; the TUI detail and acta show get CLOSES and CLOSED BY meta lines; acta list --json gets closes and closed_by.

Section 3 approved (writers, skills, migration): agents write closes: by hand in frontmatter, no new command and no acta set closes for now; acta:plan Debt items section moves DEBT ids into closes:; acta:brainstorm puts the first scratch in parent and the rest in closes:; acta:land step 9 ticks each debt item in the plan's closes:; migration on the same branch: SPEC-15 gets closes: [SCRATCH-14], PLAN-25 gets closes of its four DEBT ids, SPEC-16 is fixed by the Spec: link change, the five old specs lose their status: done line; the ignored-status problem shows only when written and derived differ.
