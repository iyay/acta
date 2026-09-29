---
id: SCR-0019
hash: bhqgz92
title: Scratch skill must add context, not only verbatim words
status: brainstorming
created: "2026-09-29"
---
hmm.. rasanya skill scratch sekarang gak terlalu detail coba buat ngasih context ya. verbatim nyatet doang. betul?

catet, harus diperbaiki skill-nya

---
Agent notes:
- plugin/skills/scratch/SKILL.md only says "You may add your own lines below theirs, marked as yours". That is optional, and it does not say which context to add. So SCRATCH-17 and SCRATCH-18 hold only the user's sentence.
- Context an agent can gather while filing: what work was going on when the idea came up, what already exists (for example `acta set scratch/<stem> status dropped` exists for scratch only), files or symbols likely touched, open questions.
- Keep the user's words verbatim at the top. The agent's context goes below, marked as the agent's.


terus format scratch-nya beda2 juga. harus disamain, biar agent dan human juga gampang ngerti

---
Agent notes (2026-09-29, from a quick look at .acta/scratch/):
- Bodies have no fixed shape. Some start with a `# Title` line (2026-09-28-drag-select-copy.md), some start straight with bullets (2026-09-28-dispatch-brief-loads-build.md). Every `acta scratch add` puts in another `---` line, so long items pile up many of them (the themes item has around 20). Agent notes are sometimes marked and sometimes not.
- A fixed body shape could look like this: `# Title`, then `## Words` (the user's text, verbatim, one dated entry per add), then `## Context` (the agent's notes), then `## Open questions`. Whether `acta scratch new/add` should write that shape itself, or only the skill should ask for it, is open.
- Old items: migrate them or leave them as they are? Open.


## Brainstorm 2026-09-29
- Classified Architectural: the change can touch the scratch file contract, `acta scratch new/add` in internal/write/scratch.go, and the TUI detail view.
- Fact check: `AppendScratch` (internal/write/scratch.go:62) only adds one blank line and the text. It never writes `---`. The many `---` lines come from agents who put them in the text they send. The earlier note in this item was wrong about that.

- Q1 who enforces the shape: user picked 2. The CLI writes the skeleton. `acta scratch new` makes the sections, and `add` puts text into the right section.

- Q2 old items: user picked 1. Leave them as they are. The new shape is for new items only, and the TUI keeps reading old bodies.

- Q3 skeleton sections: user said "disamain dengan spec, plan, bug, debt juga aja" (make it the same as spec, plan, bug and debt too). The meaning still needs a follow-up question.

- Q3b meaning of "disamain": user picked 2. All five kinds (scratch, spec, plan, bug, debt) follow one rule. Each has a fixed skeleton, the CLI writes it and checks it, and some base sections are shared. Whether this is one spec or split per kind is still open.

- Q4 split: user picked 2. One spec holds the base rule for all kinds. Plans are split per kind. The first plan is the base rule plus scratch, then bug, debt, spec and plan get one plan each.

- Approach: user picked 1. A schema table per kind lists the title and the required and optional sections in order. `acta <kind> new` writes the skeleton from it. Write commands and `acta id` refuse a file that lacks a required section. Scratch `add` gets `--section`. This builds on the BugTemplate and HasSymptom pattern in internal/write/template.go.

- Design section 1 APPROVED (base rule + schema). The body starts with `# <title>`, which matches the frontmatter title. Sections are `##` only and follow the schema order. A required section must be there. Optional sections may be missing. Extra sections may come after the schema ones. `## Context` is the one section every kind shares, for agent notes. Schema (required marked *): scratch: Words*, Context, Log, Open questions. bug: Symptom*, Root cause, Repro, Found in, Context. debt: Notes*, Context. spec: Why*, Design*, Testing*, Context. plan: Global Constraints*, File Map*, Waves*, Context. Check the debt sections against the real files before the debt plan.

- Correction: 19 of 20 old specs and 27 of 29 old plans lack the schema's required sections, so old files would fail the check.
- Design section 2 APPROVED (old files and when checks run). New files get `schema: 1` in frontmatter, written by `acta <kind> new`, or by `acta id` when a file gets its first id. Only files with `schema: 1` are checked. Files without it count as old and are skipped. Checks run in write commands (new, add), in `acta id`, and in `acta doctor` (a warn). A missing required section gives exit 1 with a message like `spec <file>: missing ## Testing`, and nothing is written. In `acta id`, a failing file gets no id and the other files still go through.

- Design section 3 APPROVED (scratch commands, skills, TUI). `new` puts stdin into `## Words` under `### YYYY-MM-DD` and writes every section. `add --section words|context|log|questions` defaults to words. Words and log entries get a dated `###` heading and go at the end of their section. Context and questions text goes at the end of its section. On an old item, `add` with no flag appends at the end like today, and `--section` gives exit 1 `SCRATCH-n is an old item with no sections`. The scratch skill must run `add --section context` after `new` (the work going on, what exists, file:line, and where the facts came from). Open questions go to `--section questions`. Agents stop writing `---` or "Agent notes". The brainstorm skill logs answers and approved sections with `--section log`. The TUI does not change. plugincheck guards the scratch skill rule.

- User, after the spec was written: "ooh.. harus nambah meta juga. kapan dikerjain dan kapan selesai dikerjain" (add meta too: when work started and when it finished). Only scratch has a date today (`created`, internal/write/scratch.go:41).

- Meta goes into SPEC-22 (user picked 1).
- Design section 4 APPROVED (date meta). Most statuses are derived, not stored, so dates are written at the events that write files. `created`: `new`, or `acta id` on a first id. `started`: the first `acta tick --start` on a task, which also fills the plan and spec above it if they have none, or `acta set status` to a working status. `finished`: the last task tick of a plan, which also fills the spec once all its plans are done, or `acta set status` to a closed status, or `acta set fixed_in` for a bug, or `acta id` on a child spec for a scratch item that becomes specced. Dates are YYYY-MM-DD. `started` is written once and never overwritten. Reopening removes `finished`. Old files get no backfill. The TUI shows the dates in the detail meta line.

- SPEC-22 approved by the user on 2026-09-29 (includes the date meta). Next is acta:plan for plan 1.
