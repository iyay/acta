---
name: setup
description: "acta: Use on first run and whenever the user asks to change setup: runs acta doctor, then the chat language, style and tone, the default build executor, split subagent models, and the optional acta block in CLAUDE.md or AGENTS.md."
---

# Setup

The chat language, style and tone live in `~/.acta/config.yaml` (or the file `PM_VOICE_FILE` names). An old `voice.yaml` in the same folder moves there on its own. `acta` reads it at the start of every session and before every message.

## First run

1. Run `acta doctor` and show its result. When a repo check failed, offer `acta doctor --fix`.
2. Run `acta config show` and read what it prints. Nothing is set (no config file, or every value marked `(default)`) means a first run: ask every question below, one at a time, in order. A value marked `(default)` is not set yet, so ask for it. Any other run: show the current setting first. Then:
   - When the user already said which part to change, change only that part and stop there.
   - Otherwise ask only the parts not set yet, then ask once: "Change anything already set? (voice, executor, plan depth, subagent models, acta block)". A yes means the user names the parts to change, so change only those.

Pass only the flags that change: `acta config set --style plain`, `acta config set --clear-tone`, `acta config set --repo-language English`, `acta config set --executor inline`, `acta config set --clear-subagent-models`. `acta config show` prints the current setting.

### Voice

Which language should I use when I talk with you? (default: English) Style: `adhd` (the answer or next action first, short numbered steps) or `plain`? (default: adhd) Anything about tone, in your own words? (optional)

Save the answers, writing the language as its full English name (Korean, not ko):

```bash
acta config set --language Korean --style adhd --tone "Casual, short sentences."
```

From the next message on, talk in the chosen language.

### Default build executor

`subagent` (default) or `inline`. Offer `dispatch` only when `HERDR_ENV=1` is in the environment, which means this session runs inside a herdr pane. `herdr` on PATH is not enough: dispatch needs this session's own pane. Without `HERDR_ENV=1`, do not offer `dispatch` at all.

```bash
acta config set --executor subagent
```

### Plan depth

`full` (default: real code in every plan step, and the plan waits for a yes) or `minimal` (short steps, no code, build starts right away).

```bash
acta config set --plan-depth full
```

Then ask once whether to save the executor and the depth for every repo or this repo only. This repo only: `acta config set --repo --executor inline --plan-depth minimal` writes `.acta.yaml`, which is committed, so it reaches everyone who clones the repo. Use `--plan-depth minimal` the same way in the global command.

### Split subagent models

Ask this one in Claude Code only, and only while `acta config show` has no `subagent_models` line. Default no. A yes saves `acta config set --subagent-models split`. A no saves `acta config set --subagent-models default`, so the question is not asked again; the user's own config wins.

### The acta block

Ask: add the acta block to CLAUDE.md / AGENTS.md? When the repo has neither, ask to create a CLAUDE.md for it. Show this exact block first, before any yes:

```markdown
<!-- acta:begin -->
## acta
This repo uses the acta plugin. Before each workflow step, load the matching acta skill and follow it.
Specs, plans, bugs, debt and scratch items live in `.acta/`.
Raw ideas go to Scratchpad with `acta scratch new`, not to agent memory.
<!-- acta:end -->
```

Write only after a yes. When both files exist, ask which one. When only CLAUDE.md exists, write the block there. When only AGENTS.md exists, write the block there and make no CLAUDE.md. When neither exists: in Claude Code, run `/init` first, then add the block to the CLAUDE.md it made; in any other harness, create a CLAUDE.md that holds only the block. Write only between the two markers; a re-run replaces the text inside them and leaves the rest of the file alone.

## Limits

- The user's own CLAUDE.md or AGENTS.md wins when it names a language or style.
- Tone is at most 8 lines and 600 characters.
- This skill edits CLAUDE.md or AGENTS.md only between the acta markers, and only after a yes; the one exception is the new CLAUDE.md that `/init` writes. It never edits settings.
