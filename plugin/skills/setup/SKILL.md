---
name: setup
description: "acta: Use on first run and whenever the user asks to change setup: runs acta doctor, then the chat language, style and tone, the default build executor, split subagent models, and the optional acta block in CLAUDE.md or AGENTS.md."
---

# Setup

Setup lives in the `acta setup` wizard now. It asks every question itself: chat language, style, tone, build executor, subagent models, plugin installs, and the acta block. It runs `acta doctor` first and writes only with a TTY.

Ask the user to run `! acta setup` in their own terminal. When the harness gives this session a TTY, run it there instead.

Pass `--plugin-dir <path>` when the user fetched this repo somewhere the wizard cannot find, so it knows which plugin to install.

When the user cannot run the wizard at all, fall back to `acta config set` with the flags `acta config show` marks as missing, one question at a time.

When the repo has code and `.acta/wiki/` holds no page yet, offer in one line to seed the wiki with `acta:migrate`, and only start on a yes.
