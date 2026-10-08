---
name: setup
description: "acta: Use on first run and when the user asks to change setup or one setting (chat language, style, tone, build executor, plan depth): runs acta doctor and writes the optional acta block in CLAUDE.md or AGENTS.md."
---

# Setup

Setup lives in the `acta setup` wizard now. It asks every question itself: chat language, style, tone, build executor, subagent models, plugin installs, and the acta block. It runs `acta doctor` first and writes only with a TTY.

Ask the user to run `! acta setup` in their own terminal. When the harness gives this session a TTY, run it there instead.

Pass `--plugin-dir <path>` when the user fetched this repo somewhere the wizard cannot find, so it knows which plugin to install.

When the user cannot run the wizard at all, fall back to `acta config set` with the flags `acta config show` marks as missing, one question at a time. The acta block is still mandatory there: write it with no yes asked and no skip. Both CLAUDE.md and AGENTS.md exist: write to both. One exists: use it. Neither exists: create a CLAUDE.md holding only the block. Never create an AGENTS.md.

When the user asks to change one setting, run `acta config show --json` first and read `overrides`. No override for that key: run `acta config set` as usual. An override means the repo's `.acta.yaml` sets that key to something other than the user's own value. Ask once and say that `.acta.yaml` is committed and shared with everyone who clones the repo. Offer three ways: keep the repo value, write the new value to the repo with `acta config set --repo --<flag>`, or remove the repo key with `acta config set --repo --unset <key>`. Write nothing until the user picks. After a write, say the file changed and is not committed.

When the user asks why a build went to some executor, answer from the `(repo; yours: ...)` line of `acta config show`.

When the repo has code and `.acta/wiki/` holds no page yet, offer in one line to seed the wiki with `acta:migrate`, and only start on a yes.
