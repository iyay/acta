---
name: setup
description: "acta: Use on first run and when the user asks to change setup or one setting (chat language, style, tone, build executor, plan depth): runs acta doctor and writes the optional acta block in CLAUDE.md or AGENTS.md."
---

# Setup

The `acta setup` wizard asks every question itself in a terminal: chat language, style, tone, build executor, subagent models, plugin installs, and the acta block. It runs `acta doctor` first and writes only with a TTY.

## Ask in the chat when AskUserQuestion exists

When this harness has the `AskUserQuestion` tool, you ask the questions yourself instead of the wizard.

1. Run `acta doctor` and report what it finds. Install nothing. When it reports a plugin missing, point to `! acta setup` for the install only.
2. Run the plain `acta config show`, not `--json`: only the plain form marks a value `(default)`. First run (the `file:` line ends in `(exists: false)`): ask every setting it lists, plus tone and subagent models, which it prints only once set. Later runs: ask only the settings it marks `(default)`, plus subagent models when it does not print them.
3. Ask with `AskUserQuestion`, at most 4 questions per call and 4 options per question. Group settings that belong together: the chat ones (chat language, style, tone, repo language), then the planning ones (build executor, plan depth, plan split, review depth), then the rest. Put the recommended option first.
4. Chat language, repo language and tone take free text. Offer two or three common options and let the built-in Other take the rest. Tone also offers a skip. Keep options neutral: never shape them on one user's setup.
5. Write every answer with one `acta config set` call, then write the acta block (see Every path).

## No AskUserQuestion

Run `acta doctor` and the plain `acta config show` and report what they show; ask nothing here. Then ask the user to run `! acta setup` in their own terminal. When the harness gives this session a TTY, run it there instead.

Pass `--plugin-dir <path>` when the user fetched this repo somewhere the wizard cannot find, so it knows which plugin to install.

When the user cannot run the wizard at all, fall back to `acta config set` for the settings step 2 above would ask, plus `--plan-split` when it shows `plan_split: off (default)`, one question at a time.

## Every path

On a full setup run outside the wizard the acta block is mandatory: write it with no yes asked and no skip. Both CLAUDE.md and AGENTS.md exist: write to both. One exists: use it. Neither exists: create a CLAUDE.md holding only the block. Never create an AGENTS.md.

When the user asks to change one setting, run `acta config show --json` first and read `overrides`. No override for that key: run `acta config set` as usual. An override means the repo's `.acta.yaml` sets that key to something other than the user's own value. Ask once and say that `.acta.yaml` is committed and shared with everyone who clones the repo. Offer three ways: keep the repo value, write the new value to the repo with `acta config set --repo --<flag>`, or remove the repo key with `acta config set --repo --unset <key>`. Write nothing until the user picks. After a write, say the file changed and is not committed.

When the user asks why a build went to some executor, answer from the `(repo; yours: ...)` line of `acta config show`.

When the repo has code and `.acta/wiki/` holds no page yet, offer in one line to seed the wiki with `acta:migrate`, and only start on a yes.
