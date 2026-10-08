---
id: SCR-0055
hash: q80ev4r
title: Repo overrides are invisible to setup
status: brainstorming
created: "2026-10-07 19:33:50"
schema: "1"
started: "2026-10-08 07:36:45"
finished: "2026-10-08 07:41:02"
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

### 2026-10-08

2026-10-08 probe round 1, all five recommendations accepted by the user:
- Q1 surfaces: wizard, setup skill, and one session-note line only when a repo value differs from the user's own value.
- Q2 removal: one repeatable flag, acta config set --repo --unset <key>.
- Q3 agent ask with a repo override present: ask once, say .acta.yaml is committed and shared, offer change repo value / remove repo key / keep.
- Q4 no auto-commit for .acta.yaml; the agent says the file changed and is not committed.
- Q5 wizard asks keep / change / remove only for keys whose repo value differs from the user's value.

### 2026-10-08

2026-10-08 probe round 2, both recommendations accepted; frontier empty, shared summary confirmed:
- Q6 acta config show prints "<repo value> (repo; yours: <user value>)" when they differ, and --json carries the user's value too; same values print as today.
- Q7 --unset works only with --repo.

### 2026-10-08

2026-10-08 approach A chosen: one Overrides(user, repoRoot) function in internal/config, next to MergeRepo, returning {key, repo value, user value} for keys that differ; config show, the session hook and the wizard call it; the skill reads acta config show --json.

### 2026-10-08

2026-10-08 section 1 (CLI) approved:
- internal/config: Overrides(user, repoRoot) ([]Override{Key, Repo, Yours}, error), only differing keys; Yours is the user's effective value with defaults; same source and errors as MergeRepo.
- UnsetRepoUser(repoRoot, keys): removes keys, keeps other lines and order, leaves an empty file, missing key is no error, non-repo key refused naming the five.
- acta config set --repo --unset <key>, repeatable, can mix with sets, same key set and unset is refused, refused without --repo, prints the path, no auto-commit.
- acta config show: "<repo> (repo; yours: <user>)" when different, "(repo)" when same; --json adds "overrides": [{key, repo, yours}] and keeps from_repo.
- Tests in temp repos only.

### 2026-10-08

2026-10-08 section 2 (session note) approved: internal/hook input gets Overrides []config.Override filled from config.Overrides where RepoErr is filled; when non-empty the note adds one short English block naming each key, repo value and user value, and telling the agent to tell the user once and name acta config set --repo --unset <key>; nothing when empty; shown wherever the plugin-conflict block shows (session start and per prompt), omp included; no override block when RepoErr is set; tests for both and for the RepoErr exclusion, plus any hook text size cap.

### 2026-10-08

2026-10-08 section 3 (wizard) approved: after the normal answers the wizard runs config.Overrides with the values just chosen; one select per differing key (keep / yours: write the user's value into .acta.yaml / remove), default keep; no question when nothing differs or outside a git repo; two new plan actions (set repo key, unset repo key) shown in the summary before writing, calling SaveRepoUser / UnsetRepoUser; no auto-commit, one closing line says .acta.yaml changed; a broken .acta.yaml skips the override step with one line naming the error; tests for no-diff, each choice, default unchanged, outside a repo.

### 2026-10-08

2026-10-08 section 4 (setup skill) approved, eval case included: on a change-setting ask the agent runs acta config show --json, checks overrides; no override means acta config set to the user file as today; an override means one question naming .acta.yaml as committed and shared, with keep / write (acta config set --repo --<flag>) / remove (acta config set --repo --unset <key>), then says the file changed and is not committed; "why did the build go to X" is answered from the (repo; yours: ...) line. Raise MaxLines and the byte cap in the same plan if needed, add --unset and overrides to the required strings. New eval case repo-override-ask: scaffold with .acta.yaml build_executor: dispatch, user asks for subagent from now on; graders: reply names .acta.yaml, the agent asks before writing, no config set --repo runs in that turn.

## Open questions
