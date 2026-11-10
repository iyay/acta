---
parent: specs/2026-11-10-parallel-plans-subagent-ask-design
depth: minimal
id: PLN-0223
created: "2026-11-10 18:16:09"
hash: t0ce23m
---
# Parallel plans: user config and CLI

**Goal:** The user config holds `parallel_plans`, `subagent_picks` and `subagent_offer`, and `acta config` can set, count and show them.

**Spec:** `.acta/specs/2026-11-10-parallel-plans-subagent-ask-design.md` (SPC-0197), sections Settings and Part 1 steps 4 to 6.

**Tests:** fast `scripts/test ./internal/config ./internal/cli`; full `scripts/test --full`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- The three new keys live in the user config only. `.acta.yaml` never holds them, and `build_executor` keeps its meaning.
- Tests pin HOME and `PM_VOICE_FILE` to a temp dir, so no test writes the real `~/.acta/config.yaml`.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2
- Wave 3: Task 3

### Task 1: User config holds the three keys

**Files:** `internal/config/user.go`, `internal/config/user_test.go`
**Seam:** `config.User` YAML round trip and `User.Validate`
**verify:** No value outside the allowed set survives a save: `parallel_plans` is unset or `subagent`, `subagent_offer` is unset or `declined`, `subagent_picks` is 0 or more. List each path checked: each allowed value saves and loads back; a wrong word, a negative count, and a value with spaces around it (trimmed, then checked); a file written before this change loads with all three unset.
- [ ] Red: a round-trip test fails because `config.User` has no `ParallelPlans`, `SubagentPicks` or `SubagentOffer` field.
- [ ] Change: add the three fields with their YAML names, trim and validate them next to `PlanSplit`.
- [ ] Commit: feat(config): add parallel_plans, subagent_picks and subagent_offer

### Task 2: config set and show know the keys

**Files:** `internal/cli/config_cmd.go`, `internal/cli/config_parallel_test.go`
**Seam:** `cli.Run` with `config set` and `config show`
**verify:** The keys are set only through the user file and shown on both outputs. List each path checked: `--parallel-plans subagent` and `--subagent-offer declined` save; a bad value for either exits with the bad-input code and leaves the file as it was; either flag with `--repo` is refused and `.acta.yaml` is not written; `config show` text prints `parallel_plans`, `subagent_picks` and `subagent_offer` with a `(default)` mark when unset; `--json` prints the same three keys without the mark.
- [ ] Red: `config set --parallel-plans subagent` fails because the flag does not exist.
- [ ] Change: add the two flags to `config set`, keep them out of the `--repo` list, and print the three keys in `config show`.
- [ ] Commit: feat(cli): set and show the parallel plan keys

### Task 3: config pick counts subagent picks

**Files:** `internal/cli/config_cmd.go`, `internal/cli/config_parallel_test.go`
**Seam:** `cli.Run` with `config pick`
**verify:** Only `acta config pick subagent` changes the count, by exactly 1, in the user file. List each path checked: first call prints 1 and saves 1; second call prints 2; other user keys stay as they were; no argument, `inline`, or two arguments exit with the bad-input code and leave the count as it was; inside a repo with `.acta.yaml`, that file does not change.
- [ ] Red: `config pick subagent` fails because `pick` is not a config subcommand.
- [ ] Change: add `pick` to `cmdConfig`: load the user file, add 1 to `SubagentPicks`, save, print the count; name it in the config usage line.
- [ ] Commit: feat(cli): add acta config pick subagent
