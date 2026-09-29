---
parent: scratch/2026-09-28-spec-status-frontmatter-pins
closes: [SCR-0014]
id: SPC-0017
hash: sh2epvu
---

# One Piece of Work Closes Several Items

Status: approved by the user on 2026-09-29 (Architectural, section by section). Built from SCRATCH-8; also closes SCRATCH-14 and the SPEC-16 finding.

## Why

acta links each item to one parent, while real work often finishes several items at once. Three symptoms today:

- SPEC-16 shows `draft` although PLAN-25, which built it, is done and landed. PLAN-25 has `parent: debt/2026-09-29-skill-paths`, and `linkPlan` (`internal/board/board.go`) skips the plan's `**Spec:**` line whenever `parent:` is set.
- SCRATCH-14 shows `raw` although PLAN-24 fixed it. SPEC-15 could name only one scratch item as its parent, and scratch statuses are only `raw`, `brainstorming` and `dropped`.
- SCRATCH-8: a written `status:` in a spec's frontmatter wins over the status derived from its plans, so SPEC-6 stayed `approved` with every task done.

## 1. The closes field

- `closes:` is a YAML list in the frontmatter of a spec, a plan or a bug, for example `closes: [SCRATCH-14, DEBT-17.1]`.
- Each entry is a short id (`SCRATCH-14`, `DEBT-17.1`) or a path id (`scratch/2026-09-29-land-fix-duplicate-ids`, `debt/<stem>#item-1`), resolved the same way `acta show` resolves ids.
- A target must be a scratch item, a spec or a debt item. A plan, a task or a bug cannot be closed this way, since their status comes from their own tasks.
- An entry that does not resolve, or points at a kind that cannot be closed, is a board problem on the item that wrote `closes:`, for example `closes SCRATCH-99 not found` or `closes PLAN-3: a plan cannot be closed`.
- `parent:` stays one value. It only sets where the item sits in the tree.
- A plan's `**Spec:**` line always links the plan to its spec, also when the plan has a `parent:`. The plan still sits under its parent in the tree, and the spec counts the plan and its tasks.

## 2. Status and display

- A scratch item named in any `closes:` shows `specced` (derived), the same as a parent link. A written `dropped` still wins. No new scratch status.
- A spec named in a plan's `closes:` counts that plan the same way as a `**Spec:**` link. A spec named in a spec's or a bug's `closes:` follows the status of the item that closes it.
- Debt items are not derived. Their box in the debt file stays the source, and it changes only when `acta:land` ticks it.
- A spec that has any plan, through `**Spec:**`, `parent:` or `closes:`, ignores a written `status:` and uses the derived one. When the written value differs from the derived one, the board shows the problem `written status <X> ignored, derived <Y>`. A spec with no plan may still carry a written `draft` or `approved`.
- The TUI detail and `acta show` gain two meta lines: `CLOSES` (the ids this item closes) and `CLOSED BY` (the ids of the items that close it). `acta list --json` gains `closes` and `closed_by`.

## 3. Writers, skills and migration

- Agents write `closes:` by hand in the frontmatter, the same way they write `parent:`. No new command, and no `acta set ... closes` for now.
- `plugin/skills/plan/SKILL.md`, "Debt items": a plan that works debt keeps `parent: debt/<stem>` and lists the DEBT ids in `closes:`. A plan that also finishes other scratch items or specs lists them in `closes:` too.
- `plugin/skills/brainstorm/SKILL.md`, "Scratch items": a spec built from several scratch items sets the first one as `parent:` and lists the rest in `closes:`.
- `plugin/skills/land/SKILL.md`, the debt step: run `acta tick <DEBT-n.m> --all` for each debt item in the plan's `closes:`, instead of reading ids from the plan text.
- Migration on the same branch:
  - SPEC-15 (`.acta/specs/2026-09-29-skill-paths-design.md`) gets `closes: [SCRATCH-14]`.
  - PLAN-25 (`.acta/plans/2026-09-29-debt-c.md`) gets `closes: [DEBT-17.1, DEBT-17.2, DEBT-16.1, DEBT-16.2]` as a record; those items are already ticked.
  - SPEC-16 needs no edit; the `**Spec:**` link change fixes it.
  - The five old specs (file-contract, pm-plugin, review-debt, tui, tui-panes) lose their `status: done` line.

## Testing

Each behaviour red first.

- Board: `closes:` resolves short ids and path ids; an unknown id and a closed plan, task or bug each give the problem text above; a plan with both `parent:` and `**Spec:**` sits under the parent and counts on the spec; a scratch named in `closes:` shows `specced`, and a written `dropped` still wins; a spec closed by a spec or a bug follows the closer's status; a spec with a plan ignores a written status, and the problem shows only when the values differ; a spec with no plan keeps its written status.
- `acta show` and the TUI detail print `CLOSES` and `CLOSED BY`; `acta list --json` carries `closes` and `closed_by`.
- plugincheck: plan, brainstorm and land name `closes:` in the places above; land no longer asks to read DEBT ids from plan text.
- After migration on the branch: SPEC-16 and SPEC-15 show `done`, SCRATCH-14 shows `specced`, and the five old specs still show `done` (derived).
- `gofmt -l .` empty, `go vet ./...` ok, `go test ./...` green.

## Out of scope

- A `closes` value on scratch or debt files.
- An `acta set ... closes` command.
- SCRATCH-6 (harness), which has its own brainstorm.
