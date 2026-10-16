---
id: SPC-0144
created: "2026-10-18 18:46:22"
hash: de9eixc
---
# Inline is the default build executor

## Why

Inline builds proved cheaper than subagent builds: one agent holds the
context once instead of handing it to a new helper for every task.
Today setup picks `subagent` first, and a build with no executor set
stops to ask. Both should land on inline.

## Design

1. `acta setup` preselects inline. `FormDefaults` in
   `internal/setup/form.go` fills an empty executor with `"inline"`, and
   the inline option is listed first.
2. A build with no executor set uses inline and does not ask. The
   executor order in `plugin/skills/build/SKILL.md` (and its
   description) becomes: `/build <executor>`, else the one
   `acta config show` names, else inline.
3. `acta doctor` says the build uses inline when no executor is set,
   in place of "build asks each time" (`checkSetup` in
   `internal/doctor/doctor.go`).
4. If `acta config show` or the session note print nothing for an
   empty executor, they show `inline (default)`. The plan checks this
   and drops the step when neither path prints executor text.

A config that already says `subagent` or `dispatch` keeps working as it
does today.

## Testing

- Setup form: an empty Env gives `BuildExecutor == "inline"`.
- Doctor: the new message for an empty executor.
- `internal/plugincheck` stays green after the skill edit.
