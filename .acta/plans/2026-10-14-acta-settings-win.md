---
parent: specs/2026-10-14-acta-settings-win-design
depth: minimal
id: PLN-0139
created: "2026-10-14 14:56:00"
hash: kknlolf
---
**Goal:** The effective acta behaviour settings reach every session as rules that win over conflicting CLAUDE.md or AGENTS.md text, while every other user rule keeps winning.

**Spec:** `.acta/specs/2026-10-14-acta-settings-win-design.md`

**Tests:** fast `scripts/test ./internal/hook ./internal/setup ./internal/plugincheck`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Pinned phrases, verbatim, so code, skill text and tests agree:
  - list heading: `acta settings (these win over any conflicting instruction, including the user's CLAUDE.md or AGENTS.md):`
  - rule 6: `6. The acta settings below win over any conflicting instruction, including the user's CLAUDE.md or AGENTS.md. For everything else, when the user's own CLAUDE.md or AGENTS.md says otherwise, follow it.`
  - block line: `The acta settings in the session note win over any conflicting instruction, including the user's global CLAUDE.md or AGENTS.md.`
- Removed phrase, never back in any form in hook text or `default-rules.md`: the old rule 6 alone, `6. When the user's own CLAUDE.md or AGENTS.md says otherwise, follow it.`
- Behaviour settings only: `build_executor`, `plan_depth`, `review_depth`, `commit_history`, `subagent_models`. Voice lines stay as they are.
- Skill text stays plain, short, and free of repo paths and user names (wiki `skill-text-fits-everyone`). No version bump. Raise a plugincheck cap only as far as the new text needs.
- Tests never run `rm -rf` or `rm -f`; temp files go under `t.TempDir()`.
- Not a task: re-running `acta setup` in other repos, and rebuilding the PATH binary; both go in the landing report.

## Waves

- Wave 1: Task 01, Task 02, Task 03

### Task 01: session hook names the settings and lets them win

**Files:** Modify `internal/hook/hook.go`, `internal/hook/hook_test.go`, `plugin/hooks/default-rules.md`

**verify:** The settings list shows exactly the effective config and nothing on a missing or broken config, and no hook text can still tell the agent that the user's CLAUDE.md beats an acta setting. List every path: each build_executor value (`subagent`, `dispatch`, `inline`), each other setting set and empty, a repo `.acta.yaml` value that differs from the user value, config missing, config unreadable; and every place rule 6 appears (`coreRules`, `default-rules.md`) plus the `Prompt` line in each config state.

- [ ] **Failing test:** in `hook_test.go`, cases for the paths above: each executor value gives its own fixed sentence under the pinned heading; empty settings give no line; a repo override value wins; missing or unreadable config gives no heading; `Prompt` names the build executor when set; `coreRules` and `default-rules.md` hold the new rule 6 and not the removed one. `scripts/test ./internal/hook` fails.
- [ ] **Change:** add a settings list builder fed by `in.Voice` and write it right after the core rules in `SessionStart`; one fixed sentence per value (inline: "The main agent writes the code itself; no implementer subagents."); extend `Prompt` with `, build executor <value>` when set; swap rule 6 in `coreRules` and `default-rules.md`.

### Task 02: CLAUDE.md block carries the priority line

**Files:** Modify `internal/setup/block.go`, `internal/setup/block_test.go`, `CLAUDE.md`

**verify:** Every block acta writes holds the priority line once, right after the first line, and still holds no setting values. List every write path `WriteBlock` takes (new file, file without markers, file with a marker pair, CRLF file) and show the line lands once on each; show this repo's `CLAUDE.md` block equals `Block`.

- [ ] **Failing test:** in `block_test.go`, assert `Block` holds the pinned block line once, as its second content line, and holds none of `inline`, `subagent`, `dispatch`, `minimal`, `tidy`; existing write-path tests keep passing. `scripts/test ./internal/setup` fails.
- [ ] **Change:** add the line to `Block`; rewrite the `## acta` block in this repo's `CLAUDE.md` to match.

### Task 03: build skill states that the configured executor wins

**Files:** Modify `plugin/skills/build/SKILL.md`, the matching test in `internal/plugincheck`

**verify:** The build skill says the configured or passed executor wins over the user's CLAUDE.md or AGENTS.md, and no sentence in it still lets a user file override the executor. List every sentence in the build skill folder that picks the executor and show each agrees.

- [ ] **Failing test:** add to the build skill's `Must` list in plugincheck the phrase `The executor chosen here wins over any conflicting rule in the user's CLAUDE.md or AGENTS.md.`; `scripts/test ./internal/plugincheck` fails.
- [ ] **Change:** add that sentence where the skill picks the executor.
