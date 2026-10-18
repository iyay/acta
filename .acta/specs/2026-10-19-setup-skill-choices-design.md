---
id: SPC-0150
created: "2026-10-19 05:32:05"
hash: isfhmzd
started: "2026-10-19 05:38:11"
finished: "2026-10-19 05:42:07"
---
# Setup questions as choices in the setup skill

## Problem

`acta:setup` sends the user to the `acta setup` wizard in a terminal. When the
wizard cannot run, the skill asks each setting as a plain chat question. In
Claude Code the user has to leave the chat, or type every answer by hand.

## Design

This is a Bounded change to the existing setup skill. No Go code changes.

- When the harness has `AskUserQuestion`, the skill is the main path. It runs
  `acta doctor` and `acta config show --json`, then asks the settings as
  choices, at most 4 questions per call and at most 4 options per question.
- First run (no config file yet): ask every core setting, grouped into calls
  that belong together (chat language, style, tone, repo language; then
  executor, plan depth, plan split, review depth; then the rest).
- Run again: ask only the settings the plain `acta config show` (not
  `--json`) marks as missing or `(default)`. The JSON output carries no
  default marker, so a value like `plan_depth: full` there cannot tell a
  default from a real choice. `--json` stays for reading `overrides`.
  Changing one setting keeps the existing override flow.
- Free-text settings (chat language, repo language, tone) offer common,
  neutral options plus the built-in "Other". Tone offers a skip. Options never
  follow one user's own setup.
- Write every answer with one `acta config set ...` call, then write the acta
  block with the existing CLAUDE.md / AGENTS.md rules.
- Without `AskUserQuestion` (omp, other harnesses), keep today's flow: the
  wizard, then the one-question-at-a-time chat fallback.
- Plugin installs stay in the wizard only. The skill reports `acta doctor`
  results and installs nothing.

## Files

- `plugin/skills/setup/SKILL.md`, plus a companion `choices.md` when the text
  does not fit. The line cap counts every Markdown file in the skill folder,
  so a companion does not dodge it: raise `MaxLines` (now 20) to fit.
- `internal/plugincheck/skill_setup_test.go`: new `MaxLines`, new `Must`
  phrases (`AskUserQuestion`, `acta config show`), and a check of the
  `MustNot` flag list, which today bans `--language`, `--style` and others
  from the skill text. Keep the ban; the skill names settings, not flags.
- `internal/plugincheck/budget_test.go`: new byte and line budget for
  `SKILL.md`, and a budget entry for `choices.md` if it is added.
- one new eval case in `plugin/evals/` whose `allowed_tools` includes
  `AskUserQuestion`

## Testing

- `scripts/test ./internal/plugincheck` passes (frontmatter, caps, budgets,
  eval layout). This checks structure only, not behaviour.
- The eval sandbox never hands the child `AskUserQuestion`, even when
  `allowed_tools` names it (measured 2026-10-19). So the eval case
  `setup-fallback-no-choices` checks the fallback path: setup reads
  `acta config show`, writes no setting, and points to `acta setup`. Run it
  alone with `scripts/eval --case setup-fallback-no-choices`.
- The choices path is checked by hand: run `/acta:setup` in a normal Claude
  Code session with no config file and see the questions come as choices.
