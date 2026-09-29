---
name: setup
description: "acta: Use on first run and whenever the user asks to change setup: runs acta doctor, then the chat language, style and tone, the default build executor, split subagent models, and the optional acta block in CLAUDE.md or AGENTS.md."
---

# Setup

The chat language, style and tone live in `~/.acta/config.yaml` (or the file `PM_VOICE_FILE` names). An old `voice.yaml` in the same folder moves there on its own. `acta` reads it at the start of every session and before every message.

## First run

1. Run `acta doctor` and show its result. When a repo check failed, offer `acta doctor --fix`.
2. Then ask, one question at a time, and only for what `acta voice show` says is not set yet.

### Voice

Which language should I use when I talk with you? (default: English) Style: `adhd` (the answer or next action first, short numbered steps) or `plain`? (default: adhd) Anything about tone, in your own words? (optional)

Save the answers, writing the language as its full English name (Korean, not ko):

```bash
acta voice set --language Korean --style adhd --tone "Casual, short sentences."
```

From the next message on, talk in the chosen language.

### Default build executor

`subagent` (default) or `inline`. Offer `dispatch` only when herdr is there: `HERDR_ENV=1` in the environment, or `herdr` on PATH. Without herdr, do not offer `dispatch` at all.

```bash
acta voice set --executor subagent
```

### Split subagent models

Ask this one in Claude Code only. Default no. A yes saves `acta voice set --subagent-models split`; a no saves nothing and the user's own config wins.

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

## Change later

A later run asks which part to change, then changes only that part.

Pass only the flags that change: `acta voice set --style plain`, `acta voice set --clear-tone`, `acta voice set --repo-language English`, `acta voice set --executor inline`, `acta voice set --clear-subagent-models`. `acta voice show` prints the current setting.

## Limits

- The user's own CLAUDE.md or AGENTS.md wins when it names a language or style.
- Tone is at most 8 lines and 600 characters.
- This skill edits CLAUDE.md or AGENTS.md only between the acta markers, and only after a yes; the one exception is the new CLAUDE.md that `/init` writes. It never edits settings.
