---
parent: debt/2026-09-30-config-file-name
closes: [DBT-0032.01]
created: "2026-09-30"
id: SPC-0034
hash: v09ti5l
---
# Build skill stops calling subagent the default

Status: design approved by the user in chat on 2026-09-30. Bounded (text in one existing skill, plus its text test).

## Why

PLN-0041 made `acta:plan` and the session index take the build executor from `acta voice show`. `plugin/skills/build/SKILL.md` still says "subagent (default, ...)" in its `description` and "`subagent` (default)" in the Executors table. The body already reads the config first, so behaviour is right, but the skill list the harness shows still tells agents subagent is the default. Review NOTE DBT-0032.01.

## Design

1. The `description` in `plugin/skills/build/SKILL.md` says the executor is the one `acta voice show` names, else ask, then lists subagent, dispatch and inline with their short meanings. No word "default".
2. The Executors table row reads `` `subagent` `` with no "(default)".
3. The "Before you pick one, run `acta voice show`" paragraph stays as it is.
4. `plugin/skills/setup/SKILL.md:29` is not touched: there "(default)" is the default answer to the setup question, which is still true.

## Testing

`TestSkillBuild` in `internal/plugincheck/skill_build_test.go` gains the new description phrase in `Must` and both old phrases, `subagent (default` and `` `subagent` (default) ``, in `MustNot`. It goes red before the skill text changes.

## Out of scope

Any other DBT-0032 item.
