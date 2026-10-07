---
id: SCR-0055
hash: q80ev4r
title: Repo overrides are invisible to setup
status: raw
created: "2026-10-07 19:33:50"
schema: "1"
---
# Repo overrides are invisible to setup

## Words

### 2026-10-07

A user who only talks to the agent cannot see or change a repo override.

Seen 2026-10-07: the user set build_executor to subagent through setup, but this repo's .acta.yaml says dispatch, and `acta config show` marks it "(repo)". The build went to omp and the user did not know why.

Gaps:
- The acta:setup skill only points to the `acta setup` wizard, which writes the user file. It never shows or asks about .acta.yaml.
- There is no way to remove a key from .acta.yaml (only set it).
- Asking the agent "use subagent from now on" has no route that touches the repo file.

Idea: setup (wizard and skill) lists each repo override next to the user value and asks keep / change / remove; the skill handles "change my executor" asks by calling `acta config set --repo` or a new clear flag.

## Context

## Log

## Open questions
