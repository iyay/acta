---
parent: scratch/2026-10-03-rename-shape-slice
id: SPC-0067
created: "2026-10-03 13:04:49"
hash: x6gwy94
started: "2026-10-03 13:37:21"
finished: "2026-10-03 14:00:28"
---
# Rename brainstorm and plan to shape and slice, version policy, token budgets

Status: design approved by the user in chat on 2026-10-03, section by section. Architectural: skill names are an interface that users, hooks, omp and evals depend on.

Step 1 of five. The goal of all five: acta costs fewer tokens than superpowers, mattpocock/skills and gstack for the same guarantees. This step renames two skills, sets how the version moves, and locks today's sizes so later steps can only shrink them.

## 1. Rename the skills

- `plugin/skills/brainstorm/` becomes `plugin/skills/shape/`, and `plugin/skills/plan/` becomes `plugin/skills/slice/`. The frontmatter `name:` changes to match, so the skills load as `acta:shape` and `acta:slice` (bare `shape` and `slice` in omp).
- Every live reference to the old skill names moves to the new ones: the other skills, `internal/hook/hook.go` (skill index and rule 7), `plugin/hooks/default-rules.md`, `plugin/README.md`, the paths in `plugin/NOTICE`, the eval cases and graders, the omp extension and its tests, and the plugincheck tests. The `plugin.json` description lists the workflow by the new names.
- Rule 7's map reads `brainstorming→acta:shape, writing-plans→acta:slice`.
- These stay as they are:
  - "brainstorm" as the name of the activity: rule 8, the scratch status `brainstorming`, the regex in `internal/hook/session.go`, the TUI and the CLI messages.
  - "plan" as the name of the file kind: `.acta/plans/`, the Plans pane, plan ids, `parent: plans/...`.
  - Everything under `.acta/`, which is history.
- The `shape` description keeps the word "brainstorm", so "brainstorm X" still reaches it.
- No alias skills for the old names: each alias would add a description that loads on every turn. `/acta:brainstorm` and `/acta:plan` stop working.
- `internal/plugincheck/no_old_names_test.go` fails when `acta:brainstorm` or `acta:plan` shows up again in the plugin files or in the text the hooks print, the same places it already checks for the old `pm:` names.

The user's own `~/.claude/CLAUDE.md` still names the old `pm:*` skills (`pm:brainstorm`, `pm:plan`, `pm:build`, `pm:tdd`, `pm:review`), `~/.pm/voice.yaml` and `.pm/handoff/`. With the user's permission (2026-10-03), the orchestrator updates that file right after this plan lands, not before, so it never names a skill that is not installed yet: `pm:*` becomes `acta:*`, brainstorm becomes `acta:shape`, plan becomes `acta:slice`, `~/.pm/voice.yaml` becomes `~/.acta/config.yaml`, `.pm/handoff/` becomes `.acta/handoff/`. It takes a backup first. `~/.claude/AGENTS.md` is an older copy and is not touched.

## 2. Version policy

- The version stays on the 0.1.x line until the first release.
- New rule in this repo's `CLAUDE.md`: a plan that changes `plugin/`, `cmd/` or `internal/` ends with one task that adds 1 to the patch in `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json` and `plugin/package.json`. A plan that only changes docs or `.acta/` does not bump.
- The implementer does the bump as that last task. The orchestrator writes no code, so it never bumps at land.
- `internal/plugincheck/plugin_test.go` stops pinning `0.1.0`. It checks instead that the three files carry the same version, in the form `x.y.z`.
- This plan is the first bump: `0.1.0` becomes `0.1.1`.

Rejected: an `acta version bump` command (new code for one line); a bump at land (the orchestrator writes no code).

## 3. Token budgets

A new plugincheck test locks today's sizes. Sizes are in bytes: about 4 bytes make a token, and a byte count gives the same result on every run with no API.

1. Every `.md` file under `plugin/skills/` and `plugin/references/` has a byte cap in one table in the test. A file with no entry fails, so a new file gets a cap on purpose. An entry for a file that no longer exists fails too, so the table stays true.
2. Every skill's `description:` line has a byte cap.
3. The text `hook.SessionStart` returns for a fixed input in the test (default user config, no herdr) has a byte cap.

- Each cap starts at the file's size after this plan's rename edits.
- A later spec that makes a skill leaner lowers that skill's caps in the same commit. A file that grows past its cap turns the test red, so a gain cannot be lost quietly.
- A failure names the file, its size in bytes, about how many tokens that is, and the cap.
- The total cap of 4240 lines (`TestTotalSkillSize` in `internal/plugincheck/final_test.go`) goes away. The caps per file are stricter.

Rejected: word counts, because they do not track tokens well; a real tokenizer, because it needs an API and the test would not be stable.

## Testing

- `scripts/test --full`, including the plugincheck tests above: renamed skill tests, old names, budgets, manifests.
- `scripts/eval` with the branch binary on PATH, because the skill names changed and the eval graders check which skill ran. It uses quota.
