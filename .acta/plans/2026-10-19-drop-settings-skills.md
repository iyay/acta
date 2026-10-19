---
parent: specs/2026-10-19-drop-repo-language-and-subagent-models-design
depth: minimal
id: PLN-0176
created: "2026-10-19 20:02:37"
hash: fd6jcne
started: "2026-10-19 20:11:23"
---

# Drop repo language and subagent models: plugin side

**Goal:** Skills, README and eval scaffolds stop naming either setting, and every subagent uses the harness default model.

**Spec:** .acta/specs/2026-10-19-drop-repo-language-and-subagent-models-design.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- New model rule text, the same in every skill: name no model; the harness default picks it. The first line of a subagent report still names its model.
- Skill edits must stay inside `internal/plugincheck` size caps.

## Waves

- Wave 1: Task 1, Task 2

### Task 1: One model rule in every skill

**Files:** `plugin/skills/build/SKILL.md`, `plugin/skills/build/implementer-prompt.md`, `plugin/skills/debug/SKILL.md`, `plugin/skills/review/SKILL.md`, `plugin/skills/review/depth.md`, `plugin/skills/shape/SKILL.md`, `plugin/skills/shape/probe.md`, `plugin/skills/slice/SKILL.md`, `plugin/README.md`, `internal/plugincheck/models_test.go`, `internal/plugincheck/skill_build_test.go`
**Seam:** `internal/plugincheck` tests over skill text
**verify:** No skill or README path picks a subagent model: list every subagent kind checked (implementer, worker, spike, debug helper, quick reviewer, both deep reviewers, fact-check, omp `agent="task"`). The phrases `Subagent models` line, `subagent_models`, `own model`, `never lower`, `reviewer model` and `sonnet` cannot come back in any skill file; the test guards each.
**Commit:** `docs(skills): subagents use the harness default model`

### Task 2: Setup skill and eval scaffolds drop both settings

**Files:** `plugin/skills/setup/SKILL.md`, `internal/plugincheck/skill_setup_test.go`, every `plugin/evals/*/scaffold.sh` and `plugin/evals-routing/*/scaffold.sh` that writes `repo_language`
**Seam:** `internal/plugincheck` setup skill test; scaffold text
**verify:** No setup step asks for repo language or subagent models, and no scaffold writes `repo_language` or `subagent_models`. List the setup questions and every scaffold checked.
**Commit:** `docs(setup): drop repo language and subagent models questions`

## Polish

### Task 3: Review polish

**verify:** Each NOTE below applied, and nothing else changes.

- [x] README no longer documents `--repo-language`; repo files are English
- [x] build skill and implementer prompt say English, not "the repo language"
- [x] build skill drops "or wrong" from the report model-line check
- [x] Commit: `polish: review notes PLN-0176`

## Review notes

Deep review, round 1 CLEAN on both axes; polish review CLEAN. Notes kept as notes: README line 62 wraps long; hook.go Subagent models line and setup/form.go "repo language" text stay until the Go-side plan lands; banned-phrase test does not scan scaffolds, hooks or omp.
