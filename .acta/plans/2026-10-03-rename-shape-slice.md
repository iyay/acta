---
parent: specs/2026-10-03-rename-shape-slice-design
depth: minimal
id: PLN-0077
created: "2026-10-03 13:10:58"
hash: o1exts7
started: "2026-10-03 13:37:21"
finished: "2026-10-03 14:00:28"
---
# Rename to Shape and Slice, Version Policy, Token Budgets Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** The brainstorm and plan skills load as `acta:shape` and `acta:slice`, the version moves one patch per landed plan starting at 0.1.1, and plugincheck locks today's byte sizes so no skill can quietly grow.

**Spec:** `.acta/specs/2026-10-03-rename-shape-slice-design.md`

**Tests:** fast `scripts/test ./internal/plugincheck` (plus the package a task names), full `scripts/test --full`; land also runs `scripts/eval` with the branch binary on PATH.

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Rename the skills only. The `brainstorming` status, the regex in `internal/hook/session.go`, the TUI, the CLI messages, rule 8's wording, the `plan` file kind (`KindPlan`, `.acta/plans/`, `--plan`) and everything under `.acta/` stay as they are.
- No alias or stub skill for the old names. The `shape` description keeps the word "brainstorm".
- Every run step uses `scripts/test <package> -run <Name>`, never bare `go test` and never `./...`.
- Comments are plain English a 10-year-old can read. They say why, not what.

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2, Task 3 (no shared file; both need the renamed files from Task 1).

## After land (not tasks)

- The orchestrator backs up and updates `~/.claude/CLAUDE.md` as the spec says: `pm:*` to `acta:*`, `acta:shape`, `acta:slice`, `~/.acta/config.yaml`, `.acta/handoff/`.
- The landing report tells the user to run `go install ./cmd/acta` and start a new session, because skill names are cached per session.

### Task 1: Rename the skills to shape and slice

**Files:**
- Move: `plugin/skills/brainstorm/` to `plugin/skills/shape/`, `plugin/skills/plan/` to `plugin/skills/slice/` (`git mv`, then fix `name:`)
- Modify: `plugin/skills/build/SKILL.md`, `plugin/skills/build/dispatch.md`, `plugin/skills/debug/SKILL.md`, `plugin/hooks/default-rules.md`, `plugin/README.md`, `plugin/NOTICE`, `plugin/.claude-plugin/plugin.json` (description), `internal/hook/hook.go` (skill index, rule 7 map), and every other hit of `grep -rnE 'acta:(brainstorm|plan)\b|skills/(brainstorm|plan)\b' --exclude-dir=.acta --exclude-dir=.git .`
- Test: rename `internal/plugincheck/skill_brainstorm_test.go` to `skill_shape_test.go` and `skill_plan_test.go` to `skill_slice_test.go`; modify `internal/plugincheck/skill_build_test.go`, `skill_debug_test.go`, `plugin_test.go` (required skill dirs), `no_old_names_test.go`, `internal/hook/hook_test.go`, `internal/evalomp/run_test.go`, `internal/cli/eval_omp_test.go`

**verify:** No live surface names the old skills in any form: no `acta:brainstorm`, `acta:plan`, `/acta:brainstorm`, `/acta:plan`, `skills/brainstorm` or `skills/plan` outside `.acta/` (paths, ids, slash commands, the rule 7 map, NOTICE paths, eval graders, test fixtures), and the hook text names `acta:shape` and `acta:slice`. Every `brainstorm` left names the activity or the `brainstorming` status, and every `plan` left names the file kind. List every file checked and each hit you kept, with its reason.

- [x] Failing test: point the plugincheck and hook tests at `shape` and `slice`, and add `acta:(brainstorm|plan)` to the old-name check in `no_old_names_test.go`; they fail because the skill folders and the hook text still use the old names.
- [x] Code: `git mv` both folders, set `name: shape` and `name: slice`, and change every live reference listed above.
- [x] Commit: `rename the brainstorm and plan skills to shape and slice`

### Task 2: Byte budgets for skills, descriptions and session start

**Files:**
- Create: `internal/plugincheck/budget_test.go`
- Modify: `internal/plugincheck/final_test.go` (remove `TestTotalSkillSize`)

**verify:** No `.md` file under `plugin/skills/` or `plugin/references/`, no skill `description:` line and no `hook.SessionStart(hook.Input{Voice: config.UserDefault(), VoiceExists: true})` text can grow past its cap without plugincheck going red; no such file can exist without a cap, and no cap can name a file that is gone. Each failure names the file, its bytes, about how many tokens (bytes / 4) and the cap. List every file with its cap, and show one mutation (one byte added to a skill) that turned the test red.

- [x] Failing test: write `budget_test.go` with empty cap tables; it fails and lists every file, description and the session start text that has no cap.
- [x] Code: fill each cap with the size measured after Task 1, and delete `TestTotalSkillSize` from `final_test.go`.
- [x] Commit: `lock byte budgets for skills, descriptions and session start`

### Task 3: Version policy and the first patch bump

**Files:**
- Modify: `CLAUDE.md` (version rule), `internal/plugincheck/plugin_test.go` (drop the `0.1.0` pin), `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** plugincheck goes red whenever the three manifests disagree on the version or any of them is not `x.y.z`, and nothing outside `.acta/` pins `0.1.0` any more. List every place a version string lives.

- [x] Failing test: replace the `0.1.0` pin with a check that the three manifests agree and match `x.y.z`, then set only `plugin.json` to `0.1.1`; it fails because the other two still say `0.1.0`.
- [x] Code: set `marketplace.json` and `package.json` to `0.1.1`, and add the rule to `CLAUDE.md`: until the first release the version stays on 0.1.x, and a plan that changes `plugin/`, `cmd/` or `internal/` ends with a task that adds 1 to the patch in the three manifests; a plan that only changes docs or `.acta/` does not bump.
- [x] Commit: `version policy: one patch per landed plan, now 0.1.1`

## Review notes

- `plugin/skills/shape/SKILL.md:86` "Plan, build, review and land for this brainstorm" names activities, so it stays, but it no longer matches the shape and slice step names in `plugin.json` and `README.md:3`.
- `internal/plugincheck/no_old_names_test.go:91-99` never renders SessionStart with `Herdr: true`, so an old name added only to `herdrExtra` would pass; the `pm:` names had the same coverage.
- `plugin/omp/FACTS.md:157-182` and `plugin/evals/FACTS.md:84` keep the old skill list and version `0.1.0` inside dated transcripts; nothing reads them.
- The version rule in `CLAUDE.md` says nothing about a plan that changes only `scripts/` or `go.mod`; the spec has the same gap, and the user is weighing "every plan bumps".
- `plugin/skills/shape/spec-document-reviewer-prompt.md` has a byte cap but no skill links it (dead since before this plan); delete it in the shape lean-down.
- Putting all three manifests back to `0.1.0` at once keeps `TestManifests` green by design; only the `CLAUDE.md` rule makes the bump happen.
- Until `go install ./cmd/acta` runs after land, the `acta` on PATH still prints `acta:brainstorm` and `acta:plan`.
- `internal/plugincheck/no_old_names_test.go:69-72` doc comment leaves out the `acta voice` check (`oldVoiceRe`, a gap older than this plan) and chains two "or" lists in one sentence.
