---
parent: scratch/2026-09-29-brainstorm-always-writes-spec
closes: [SCR-0014]
id: SPC-0015
hash: u8739wj
---

# Bounded Brainstorms Write a Spec, Dispatch Lands Through acta:land

Status: approved by the user on 2026-09-29. Covers SCRATCH-12 and SCRATCH-14.

## Why

- SCRATCH-12: the user expects every brainstorm they ask for to end in a spec file. The Bounded path says "No spec file, no implementation plan document", so a Bounded design lived only in chat and the scratch item. A scratch item shows as specced only through a spec's parent link, and a plan needs a spec to point at.
- SCRATCH-14: `acta:land` already runs `acta id --fix-duplicates` right after the merge (`plugin/skills/land/SKILL.md` step 5). The Landing section in `plugin/skills/dispatch/SKILL.md` keeps its own merge steps and lacks that line. Dispatch landings followed it, so duplicate ids landed twice on 2026-09-29 (DEBT-14, DEBT-15). The bug came from the landing rules living in two places.

## Part 1: Brainstorm paths

File: `plugin/skills/brainstorm/SKILL.md`.

- Spike: unchanged. No spec; the output is an answer.
- Bounded: after the clarifying questions, present one recommended design in chat. On the user's yes, write a short spec (about half a page) to `.acta/specs/YYYY-MM-DD-<topic>-design.md` in the worktree, run `acta id`, commit, and ask the user to review the file. On approval, invoke `acta:plan`. No 2-3 approaches, no per-section approval, and no one-per-session limit.
- Architectural: unchanged.
- Every place that still says Bounded has no spec or no plan changes to match: the Three Paths entry, the Bounded checklist, the flow graph (the "Implement via normal workflow (no plan doc)" node becomes a short-spec node that leads to `acta:plan`), the terminal-states paragraph, and the "short in-chat design is the whole process" line.
- The Red Flags row "I'll call it bounded and skip the spec" stays; it now matches the rule.

## Part 2: Dispatch landing

File: `plugin/skills/dispatch/SKILL.md`.

- The section "Landing — from the main checkout" shrinks to a pointer: clean and complete means run `acta:land` in the same turn, with no ask.
- It keeps only the steps `acta:land` does not have:
  - close the recipient's pane by slug (`herdr pane close`), after the worktree is removed;
  - the `Bugs found by recipient:` line in the landing report;
  - the omp memory harvest, which already lives in "Memory sweep before done".
- Steps that repeat `acta:land` go: preconditions, stray-file check, parent check, the merge command, post-merge gates, worktree and branch cleanup.
- Other mentions of landing in the file (the loop at a glance, the autonomous loop) point at `acta:land` and do not restate the merge.

## Testing

In `internal/plugincheck`, each guard written red first:

- Brainstorm: the Bounded text must name `.acta/specs/` and `acta:plan`; the old line "No spec file, no implementation plan document" must not return.
- Dispatch: the Landing section must name `acta:land`; `merge --no-ff` must not appear anywhere in `plugin/skills/dispatch/`. Lower the dispatch MaxLines cap to the new size.
- `go test ./...`, `gofmt -l .` and `go vet ./...` stay green.

## Out of scope

- No change to `plugin/skills/land/SKILL.md`; it already holds the rule.
- No Go code change to `acta id` or `AssignIDs`.
