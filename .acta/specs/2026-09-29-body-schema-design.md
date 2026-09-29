---
parent: scratch/2026-09-29-scratch-skill-adds-context
id: SPC-0022
hash: j95c98o
status: approved
---
# One Body Schema for Every Kind

Status: design approved by the user on 2026-09-29 (Architectural). The rulings are in SCRATCH-19.

## Why

- Scratch bodies have no shape. Some start with a `# Title`, some start with bullets. Agents paste `---` lines and "Agent notes" into the text by hand, so long items grow a pile of separators (the themes item has about 20). The scratch skill only asks for the user's words verbatim. Context is optional, and the skill never says what the context should hold.
- The other kinds each have a shape, but in most of them only the skill asks for it. Only bug has code that checks it: `BugTemplate` and `HasSymptom` in `internal/write/template.go`. Spec and plan follow the skill text only.
- The user wants one rule for all five kinds (scratch, bug, debt, spec, plan). The CLI writes the rule and the CLI checks it, so agents and people read every file the same way.

## Design

### Base rule

- The body starts with `# <title>`, which matches the frontmatter `title`.
- Sections are `##` headings. They follow the schema order.
- A required section must be there. An optional section may be missing.
- Extra sections that are not in the schema may come after the schema ones.
- `## Context` is the one section every kind shares. Agents put their notes there.

### Schema

A single table in code lists each kind's sections in order. Required ones are marked with *.

| kind | sections |
|---|---|
| scratch | Words*, Context, Log, Open questions |
| bug | Symptom*, Root cause, Repro, Found in, Context |
| debt | Notes*, Context |
| spec | Why*, Design*, Testing*, Context |
| plan | Global Constraints*, File Map*, Waves*, Context |

Before the debt plan starts, check the debt sections against real debt files. Change the table here if they do not match.

### Old files and when checks run

- New files get `schema: 1` in their frontmatter. `acta <kind> new` writes it, and so does `acta id` when a file gets its first id.
- Only files with `schema: 1` are checked. Files without it count as old and are skipped. The existing 20 specs and 29 plans mostly do not fit the schema, and old scratch items stay as they are (user ruling).
- Checks run in the write commands (`new`, `add`), in `acta id`, and in `acta doctor`. In `acta doctor`, a failing file is a `warn`.
- A missing required section gives exit 1 with a message that names the file and the section, like `spec 2026-09-29-x-design.md: missing ## Testing`. Nothing is written.
- In `acta id`, a failing file gets no id. The other files still get theirs.

### Scratch commands

- `acta scratch new <slug> [--title T] < words` writes the whole skeleton. Stdin goes into `## Words` under a `### YYYY-MM-DD` heading. The other sections start out empty.
- `acta scratch add SCRATCH-n [--section words|context|log|questions]` has `words` as its default.
  - For `words` and `log`, the text gets a `### YYYY-MM-DD` heading and goes at the end of its section.
  - For `context` and `questions`, the text goes at the end of its section.
- On an old item (no `schema: 1`), `add` with no `--section` appends at the end of the file, the same as today. `add` with `--section` gives exit 1 with `SCRATCH-n is an old item with no sections`.

### Skills

- `acta:scratch`: right after `new`, the agent runs `add --section context`. The context says what work was going on, what already exists, the file:line spots, and where the facts came from (reading, debug, or a run). Open questions go to `--section questions`. Agents stop writing `---` or "Agent notes" by hand.
- `acta:brainstorm`: user answers and approved design sections go to `--section log`.
- `plugincheck` guards the scratch skill rule.

### Date meta

Every kind gets three dates in its frontmatter, written as `YYYY-MM-DD`. Most statuses are worked out from other data rather than stored: a plan's status comes from its task ticks, and a scratch item is `specced` once a spec links to it. So each date is written by a command that already writes the file, never by the agent.

| field | written by |
|---|---|
| `created` | `acta <kind> new`, or `acta id` when a file gets its first id |
| `started` | the first `acta tick --start` on a task, which also fills the plan and spec above it when they have none; or `acta set status` to a working status (`in-progress`, `fixing`, `brainstorming`) |
| `finished` | the tick that closes a plan's last task, which also fills the spec once all its plans are done; or `acta set status` to a closed status; or `acta set ... fixed_in` for a bug; or `acta id` on a child spec, for a scratch item that becomes `specced` |

- `started` is written once. Later ticks leave it alone.
- When a status goes back to a working one, `finished` is removed.
- Old files get no backfill.

### TUI

The detail view already shows the body markdown as it is, so the schema needs no TUI change. The detail meta line shows `created`, `started` and `finished` when a file has them.

### Plans

This spec gets one plan per kind:

1. The base rule (schema table, `schema: 1`, the checks in write commands, `acta id` and `acta doctor`), the date meta and its TUI line, plus scratch.
2. Bug.
3. Debt.
4. Spec.
5. Plan.

Plan 1 lands first. The others build on its table.

## Testing

- Schema table: each kind's required sections, in order. A file that has every required section passes. A file that misses one fails, and the error names the file and the section. Extra sections after the schema ones pass. A section out of order fails.
- `schema: 1`: a file without it is skipped in every check. `new` writes it. `acta id` writes it only on a file's first id, and never on an old file that already has an id.
- `acta id`: a failing file gets no id, and the passing files next to it still get theirs.
- `acta doctor`: a `schema: 1` file with a missing section shows as `warn`. A clean repo shows `ok`.
- `scratch new`: the whole skeleton, with stdin under `## Words` / `### <date>`. Empty stdin still fails, like today.
- `scratch add`: each `--section` value lands text in the right place. The default is `words`. A bad `--section` value gives exit 1. Words and log get a dated heading, and context and questions do not. On an old item, no flag appends at the end, and `--section` gives exit 1 with the exact message.
- plugincheck: fails when the scratch skill loses the `--section context` step.
- Date meta: each event in the table writes its field. A second `tick --start` does not overwrite `started`. The closing tick fills `finished` on the plan, and on the spec only when every plan under it is done. Reopening removes `finished`. A file that has no dates stays without dates. The TUI meta line shows the dates when they are there, and adds nothing when they are not.

## Context

- Found while running `/acta:setup` on 2026-09-29. SCRATCH-17 and SCRATCH-18 held only the user's sentence until context was added by hand.
- `AppendScratch` (`internal/write/scratch.go:62`) adds only one blank line and the text. The `---` lines come from agents.
- Rulings, in order: the CLI writes the skeleton; old items stay; all five kinds share one rule; one spec with one plan per kind; schema table plus checks (approach 1); the date meta goes into this spec and into plan 1.
