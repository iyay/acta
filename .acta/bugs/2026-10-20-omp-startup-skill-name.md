---
id: BUG-0052
hash: g8qnh99
started: "2026-10-20 08:02:07"
fixed_in: 31ed038
finished: "2026-10-20 08:17:44"
---
# OMP starts by reading an unknown acta-prefixed skill

## Symptom

At the start of an OMP session, the agent reads `skill://acta-build` and gets
`Unknown skill: acta-build`. The available skills include `build`.

## Root cause

`plugin/skills/build/SKILL.md:2` registers `name: build`.
`plugin/omp/index.ts:122-139` forwards the shared startup rules without an
OMP-specific skill-loading note. The shared text asks for the Skill tool,
while OMP loads skills with Read and bare-name `skill://` URIs. The startup
context does not explain how references such as `acta:build` map to `build`.
The user's transcript shows the agent guessing the invalid `acta-build` name.

## Repro

The user's OMP transcript reports `Unknown skill: acta-build` while listing
`build`. A read-only Bun probe importing `createState` and forwarding the
fallback startup rules reports registeredName=build, resolves=false for
acta-build, bareNameResolves=true, startupMentionsReadSkillURI=false,
and startupRequestsSkillTool=true. The corrected bare name matches the
registered frontmatter name. A live model run after the fix remains unverified.

## Found in

Main checkout, debug following the user's startup error. No source edits made.
