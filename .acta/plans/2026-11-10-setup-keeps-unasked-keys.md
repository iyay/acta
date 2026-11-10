---
parent: bugs/2026-11-10-setup-wizard-wipes-unasked-keys
depth: minimal
id: PLN-0225
created: "2026-11-10 19:30:22"
hash: jhy2g35
---
# Setup keeps the keys it does not ask about

**Goal:** `AnswersUser` starts from the current config and sets only the ten fields the form asks, so `acta setup` stops wiping the others.

**Spec:** `.acta/specs/2026-11-10-setup-keeps-unasked-keys-design.md` (SPC-0199)

**Tests:** fast `scripts/test ./internal/setup`; full `scripts/test --full`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- A rejected config file and changes made while the wizard is open stay out of scope.

## Waves

- Wave 1: Task 1

### Task 1: AnswersUser keeps unasked fields

**Files:** `internal/setup/form.go`, `internal/setup/form_test.go`
**Seam:** `setup.AnswersUser(current, answers)`
**verify:** No field of `config.User` is lost or left stale by setup. List each path checked, walking every field by reflection: with every field of `current` set, each field outside the ten asked ones comes back equal to `current`; each of the ten asked fields equals the form answer, not the old value; an empty tone answer over a saved tone gives an empty tone.
- [ ] Red: the reflection test fails because `theme`, `coding_guide`, `update_check`, `parallel_plans`, `subagent_picks` and `subagent_offer` come back empty.
- [ ] Change: `AnswersUser` copies `current`, then sets the ten asked fields from the answers.
- [ ] Commit: fix(setup): keep the config keys the form does not ask about
