---
parent: specs/2026-10-18-split-spec-into-plans-design
depth: minimal
id: PLN-0159
created: "2026-10-18 20:21:42"
hash: jw8iwox
started: "2026-10-18 20:23:18"
---
# Split Spec Into Plans Plan

**Goal:** With `plan_split` on, slice cuts a spec into 2 to 4 plans that several sessions build at once, guarded by `acta ready` and with spec dates owned by land.

**Spec:** `.acta/specs/2026-10-18-split-spec-into-plans-design.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- `plan_split` defaults to `off`; with it off, every output (setup, session note, slice, build, land, tick) behaves exactly as today, and the session start byte cap does not change.
- `acta ready` exit codes: 0 ready, 1 waiting, 2 bad `after:` (unknown id, non-plan id, self-reference, cycle with its chain).

## Waves

- Wave 1: Task 1, Task 4, Task 6
- Wave 2: Task 2, Task 3, Task 5
- Wave 3: Task 7
- Wave 4: Task 8, Task 9

### Task 1: plan_split setting

- [x] Red: config tests fail because `plan_split` is not a known field, `config set --plan-split` does not exist, and `config show` prints no `plan_split` line.
- [x] Change: add `PlanSplit` (`plan_split`, `on|off`) to the user config and repo `.acta.yaml` merge, validate it, add `--plan-split` to `config set`, and show it in text (`off (default)` when unset) and JSON. Files: `internal/config/user.go`, `internal/config/repo_user.go`, `internal/cli/config_cmd.go` and their tests. Seam: `config.User` validation, `acta config set|show`. verify: no value but `on`, `off` or empty is ever saved or read back, from either file; unset always shows as `off (default)` in text and `off` in JSON.
- [x] Commit: `feat(config): plan_split setting`

### Task 2: Setup step

- [ ] Red: form tests fail because there is no "Plan split" step after "Plan detail".
- [ ] Change: add the "Plan split" select right after "Plan detail" in `internal/setup/form.go`, options `off: one plan per spec` (first, default) and `on: big specs become 2 to 4 plans, one session each`, carried into the saved user, answered line and text form; shift the index-based tests in `internal/setup/*_test.go`. Seam: `FormDefaults`, `AnsweredLines`, the built user. verify: every setup path (form, text mode, answered line, saved config) shows and saves the pick, starts on `off` when unset, and never replaces a saved value.
- [ ] Commit: `feat(setup): plan split step`

### Task 3: Session note line

- [ ] Red: hook test fails because a user with `plan_split: on` gets no `- Plan split: on.` line.
- [ ] Change: in `settingsBlock` (`internal/hook/hook.go`) add the line only for `on`; tests in `internal/hook/hook_test.go`. Seam: `SessionStart`. verify: the line appears for `on` and in no other case; `default-rules.md` and the session start cap stay byte for byte the same.
- [ ] Commit: `feat(hook): session note names plan split`

### Task 4: Board reads after and marks waiting

- [x] Red: board test fails because a plan's `after:` list is not read and a plan waiting on an open plan has no `waiting` mark.
- [x] Change: read `after:` into the plan item in `internal/board/board.go` (same list parse as `closes:`), and mark a plan `waiting` while any listed plan is not done; draw the mark where the board shows plan status. Seam: `board.Load`, item fields. verify: a plan is waiting exactly while one of its `after:` plans is not done; a plan with no `after:` never shows the mark.
- [x] Commit: `feat(board): plans wait on their after list`

### Task 5: acta ready

- [ ] Red: `acta ready` tests fail because the command does not exist.
- [ ] Change: new `internal/cli/ready.go` (dispatch in `cli.go`, entry in `help.go`): `acta ready <plan> --parent <ref> [--head <rev>]`. Read each `after:` plan with `git show <ref>:<path>`; landed means the file is there with every box ticked; with `--head`, the first commit on the ref where that plan is fully ticked must be an ancestor of the head. Exit 2 for unknown id, non-plan id, self-reference or a cycle (print the chain `A -> B -> A`). Tests in `internal/cli/ready_test.go` over a temp git repo. Seam: `acta ready`. verify: list every path that can say ready: only one where each `after:` plan is fully ticked in the ref tree and, with `--head`, its landing commit is an ancestor; a plan done only in a worktree or landed on another branch is never ready; unknown, non-plan, self and 2- and 3-plan cycles each exit 2 with the cause named.
- [ ] Commit: `feat(cli): acta ready checks after plans on the parent`

### Task 6: Tick leaves split spec dates alone

- [x] Red: tick test fails because a tick on one of two plans of a spec writes `started:` into the spec.
- [x] Change: in `TaskDates` (`internal/write/mark.go`) skip the spec writes when the spec has more than one plan; tests in `internal/write/`. Seam: `TaskDates`. verify: with two or more plans no tick ever writes the spec file; with one plan the spec gets `started:` and `finished:` exactly as today.
- [x] Commit: `feat(tick): split spec dates are left to land`

### Task 7: acta spec-dates for land

- [ ] Red: `acta spec-dates` tests fail because the command does not exist.
- [ ] Change: new `acta spec-dates <plan>` (`internal/write/specdates.go`, `internal/cli/specdates.go`, `cli.go`, `help.go`): on the checked-out parent, set the spec's `started:` to the earliest plan `started:` (move earlier, never later) and `finished:` when every plan of the spec is done, then commit it with the acta commit path; no-op for a one-plan spec. Seam: `acta spec-dates`. verify: in both land orders (first-started plan lands first, and last) the spec ends with the earliest `started:` and one `finished:` written only by the land that closes the last plan; after each run the checkout has no tracked change.
- [ ] Commit: `feat(cli): acta spec-dates writes split spec dates at land`

### Task 8: Skill text

- [ ] Red: plugincheck tests fail because slice has no split rules, build does not run `acta ready`, land's tidy flow has no spec-dates step, and setup's fallback does not name plan split.
- [ ] Change: `plugin/skills/slice/SKILL.md` (split rules, `after:`, 2 to 4 plans, ready lines per plan, `/slice split`, run the ready check over written plans), `plugin/skills/build/SKILL.md` (pick parent, run `acta ready --parent`, `--head` on a reused worktree, stop on exit 1 or 2), `plugin/skills/land/tidy.md` (after moving the parent: `acta spec-dates`, then last-land, then clean-check), `plugin/skills/setup/SKILL.md`; tests in `internal/plugincheck/`. Seam: plugincheck required-text checks. verify: no skill still tells an agent to build a plan without the ready check, or to land a split plan without spec-dates before last-land; every skill stays inside its size cap.
- [ ] Commit: `feat(skills): slice splits, build checks ready, land writes spec dates`

### Task 9: Two worktrees end to end

- [ ] Red: an end-to-end test fails until tasks 5 to 7 are wired together.
- [ ] Change: `internal/cli/split_e2e_test.go`: one spec, two plans, two real worktrees started at different times, both landed through `acta tidy` plus `acta spec-dates`, in both land orders; plus a worktree made before its `after:` plan landed. Seam: the CLI. verify: wiring only: both lands succeed with no clash, the spec ends with the earliest `started:` and one `finished:`, the main checkout is clean after each land, and the stale worktree gets `acta ready` exit 1 until it takes in the parent.
- [ ] Commit: `test(cli): split plans land from two worktrees`
