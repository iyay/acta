---
parent: scratch/2026-09-28-skill-rules-spec
id: SPEC-11
hash: yk3t
---
# Skill rules: scratch, one brainstorm per session, first-run setup

Status: design approved by the user in chat on 2026-09-28, section by section. The brainstorm log is SCRATCH-5 (`.acta/scratch/2026-09-28-skill-rules-spec.md`, rulings 1-15). This is spec 2 of 3. Spec 1 is SPEC-10 (Scratchpad kind, five-pane sidebar). Spec 3 is SCRATCH-6 (harness: hook state, eval suite) and is out of scope.

## Why

- Raw ideas still land in agent memory. The Scratchpad kind exists now, but no skill tells the agent to use it.
- A brainstorm can die with its session. Its answers live only in chat.
- Two big brainstorms in one session eat the context window.
- There is no first-run command. `/acta:setup` only sets the voice. Users cannot see what is broken (stale omp links, clashing plugins, a missing `.gitignore` line).
- In omp, skill names have no plugin prefix, so acta's `brainstorm` mixes with other plugins' skills.
- The build executor is asked every time. The user wants a saved default.

## Non-goals

- No hook state that counts brainstorms per session, and no eval suite (spec 3, SCRATCH-6). Here the rule lives in skill text only.
- No skill rename and no generated skill copies for omp (see section 3).
- No per-repo override of the build executor.
- No writes to `~/.claude.json`, `~/.omp` or other user-scope files.
- No change to `tdd`, `land`, `bug` and `migrate` skill bodies beyond the description prefix.

## 1. `acta doctor`

New command `acta doctor [--fix]`. Logic in a new package `internal/doctor`; the command in `internal/cli/doctor.go`.

Output: one line per check, `ok|warn|fail <name>: <message>`. A check that is not ok adds a second line `fix: <exact command or step>`.

Checks, in this order:

| # | Name | Level when bad | What it looks at | `--fix` |
|---|------|----------------|------------------|---------|
| 1 | `binary` | fail | Path and version of the running `acta` (shows a stale hand-built binary) | no |
| 2 | `harness` | fail | acta plugin enabled in Claude Code (`hook.EnabledPlugins`); if `~/.omp/plugins/node_modules/acta` exists, its target folder exists | no |
| 3 | `stale-links` | warn | Every link in `~/.omp/plugins/node_modules/` whose target is gone. Fix line: `omp plugin unlink <name>` | no |
| 4 | `conflicts` | warn | Clashing workflow plugins (`hook.Conflicts` with `plugin/hooks/workflow-plugins.txt`) | no |
| 5 | `repo` | fail | `.acta/` exists; `.agents.json` is in its `.gitignore`. Skipped outside a git repo | yes (create `.acta/`, reuse `hook.EnsureGitignore`) |
| 6 | `agents-view` | warn | `leftArrowOpensAgents` in `~/.claude.json` is `false`. Unset counts as true. Fix line: `Open /config, turn on '← opens agents'` | no |
| 7 | `setup` | warn | Voice file missing, or `build_executor` unset. Fix line: `/acta:setup` | no |

- Exit code 0 when no check is `fail`; 1 when any is.
- A broken or missing `~/.claude.json` or `~/.omp` is read as "not there", never a crash. For check 6 it gives `warn` with the parse error.
- `--fix` only touches the repo. Fixed items print `fixed <name>: <what>` and are committed the way other acta write commands commit.
- `~/.claude.json` key `leftArrowOpensAgents` is undocumented. It was found in the Claude Code binary (2026-09-28). Claude Code rewrites that file while it runs, so acta never writes it.

## 2. Voice file fields

Two new optional fields in `~/.acta/voice.yaml` (`internal/voice`). The file keeps its name.

- `build_executor`: `subagent`, `dispatch` or `inline`. Set with `acta voice set --executor <x>`; any other value is rejected with exit code for bad input.
- `subagent_models`: `split` or unset. Set with `acta voice set --subagent-models split`, removed with `acta voice set --clear-subagent-models`; any other value is rejected.
- `acta voice show` prints both fields when set.

## 3. Skill names

- Every `plugin/skills/*/SKILL.md` keeps its `name:`.
- Every `description:` starts with `acta: `. `internal/plugincheck/check.go` reports a skill without it. The 1024-character limit stays.
- Why not rename at load in omp: omp's `resources_discover` event only returns `skillPaths` (omp 18.3.5, `src/extensibility/extensions/types.ts:742`), so a rename needs generated skill copies. That stays the upgrade path if a real name clash shows up.

## 4. Skill text

### 4.1 New skill `acta:scratch` (`plugin/skills/scratch/SKILL.md`)

- Explicit words ("catet", "nanti", "kepikiran", "note this", "later") file at once: `acta scratch new <slug> [--title T] < body.md`, then one line to the user: `Filed SCRATCH-n <title>`.
- An idea guessed from context (a side idea during a build) is asked first: "File this in Scratchpad?"
- The body is the user's words verbatim. Images are written as their paths.
- Scratch items never go to agent memory and never need a new session.
- Scratch writes (`acta scratch new`, `acta scratch add`, `acta set` on a scratch item) may commit on the main branch. They are data, not code.
- Dropping: the user says drop, the agent runs `acta set scratch/<stem> status dropped`.

### 4.2 `acta:brainstorm`

These rules apply to the Architectural path only. Spike and Bounded need no scratch item and are free in number.

- Step 0: find the scratch item the request names. None: file the user's words verbatim with `acta scratch new`. Then `acta set scratch/<stem> status brainstorming`.
- Each answer and each approved design section is appended as it happens with `acta scratch add SCRATCH-n`, so a dead session loses nothing.
- The spec's frontmatter carries `parent: scratch/<stem>`. `specced` then comes from that link.
- One Architectural brainstorm per session. When a second one comes up, the agent files it as a scratch item and asks the user to pick one of three:
  - (a) Background agent: run `claude --bg 'brainstorm SCRATCH-n'` from the repo, then tell the user to open agents view (press ←, or `claude agents`).
  - (b) New herdr tab. Offered only when `HERDR_ENV=1`.
  - (c) Manual new session: copy the prompt `brainstorm SCRATCH-n` to the clipboard (`pbcopy` on macOS, `wl-copy` or `xclip` on Linux, OSC 52 when none work) and also print it.
- Plan, build, review and land for this brainstorm may continue in the same session.

### 4.3 `acta:setup`

- First step: run `acta doctor` and show its result. Offer `acta doctor --fix` when a repo check failed.
- Then ask, one at a time, only what is not set yet on the first run:
  1. Voice (language, style, tone), as today.
  2. Default build executor. `dispatch` is offered only when herdr is present (`HERDR_ENV=1` or `herdr` on PATH). Saved with `acta voice set --executor`.
  3. Claude Code only: split subagent models? Default no. Saved with `acta voice set --subagent-models split`.
  4. Add the acta block to CLAUDE.md / AGENTS.md? Show the exact block first. Write only after a yes, only to files that already exist, and ask which one when both exist.
- A later run asks which part to change.
- The block, between markers; a re-run replaces only the text inside them:

```markdown
<!-- acta:begin -->
## acta
This repo uses the acta plugin. Before each workflow step, load the matching acta skill and follow it.
Specs, plans, bugs, debt and scratch items live in `.acta/`.
Raw ideas go to Scratchpad with `acta scratch new`, not to agent memory.
<!-- acta:end -->
```

- The "Limits" section changes from "never edits CLAUDE.md, AGENTS.md or settings" to "edits CLAUDE.md or AGENTS.md only between acta markers, only after a yes; never edits settings".

### 4.4 `acta:build`

- Read `build_executor` from `acta voice show`. When set, use it without asking. When unset, ask as today.

### 4.5 Subagent models (plan, brainstorm, debug, build, dispatch, review)

Each of these skills carries the same sentence: when `acta voice show` lists `subagent_models: split` and the harness is Claude Code, subagents that write code use `model: "sonnet"`, all other subagents (mapping, explore, planning help, debug investigation, spikes) use `model: "opus"`, and reviewers use the orchestrator's alias. Otherwise the skill names no model and the user's own config wins.

## 5. Session rules, hook and README

- `plugin/hooks/default-rules.md` and the text from `hook.SessionStart` stay the same text:
  - New index line: `acta:scratch: raw ideas ("catet", "nanti", side ideas); file with acta scratch new, never memory`.
  - `acta:setup` line becomes: `first-run setup and later changes: doctor, voice, build executor, subagent models, CLAUDE.md block`.
  - New core rule 8: `One Architectural brainstorm per session; a second one becomes a scratch item and the user picks how to open it.`
- When the voice is not set, the session-start text points to `/acta:setup` instead of listing the three voice questions.
- `plugin/README.md`:
  - "It never edits your CLAUDE.md, AGENTS.md or settings" becomes "It edits CLAUDE.md or AGENTS.md only between acta markers, and only after your yes in /acta:setup. It never edits settings."
  - New section "First run": install, `/acta:setup`, `acta doctor`.
  - omp note: skill names have no prefix there; each description starts with `acta:`.

## 6. Testing

TDD, red first, for every item:

- `internal/doctor`: table tests with a temp home and a temp repo, one case per check and level. Adversarial cases: stale omp link, `acta` link pointing nowhere, broken `~/.claude.json`, `leftArrowOpensAgents` false and unset, repo with no `.gitignore`, run outside git, `--fix` twice (second run changes nothing).
- `internal/cli`: `acta doctor` exit codes; `acta voice set --executor` with each good value, a bad value and an empty value; `--subagent-models` with `split`, a bad value and `--clear-subagent-models`.
- `internal/plugincheck`: description prefix check (good, missing, prefix without a space); one text guard per rule in section 4 (for example `skill_scratch_test.go` checks `acta scratch new` and "verbatim"; `skill_brainstorm_test.go` checks `claude --bg`, `HERDR_ENV`, `pbcopy`, `acta scratch add`, `parent: scratch/`; `skill_setup_test.go` checks `acta doctor`, `acta:begin`, `--executor`); README no longer contains "never edits your CLAUDE.md".
- `internal/hook`: session-start text has the scratch line and rule 8; unset voice points to `/acta:setup`; `default-rules.md` matches the hook text (add the parity test if none exists).

## 7. Plan shape

One plan, three waves:

1. Go: `internal/doctor` + `acta doctor`; voice fields `build_executor` and `subagent_models`.
2. Skill text: description prefix, `acta:scratch`, brainstorm, setup, build, the model sentence in six skills.
3. `default-rules.md`, hook text, README.
