# acta

A lean workflow plugin for Claude Code and omp. It covers shape, slice,
build, test first, debug, review and land, and writes planning files to `.acta/`
through `acta`, the acta CLI.

## Install

1. Install `acta` and put it on your PATH:
   ```bash
   go install github.com/iyay/acta/cmd/acta@latest
   ```
2. Claude Code: add this folder as a marketplace and install the plugin:
   ```text
   /plugin marketplace add /path/to/acta/plugin
   /plugin install acta@acta-local
   ```
3. omp: `omp plugin link /path/to/acta/plugin`; if the extension does not run, see `omp/FACTS.md#installing-acta-in-omp`.

The `pmb` command still runs as an alias of `acta` and prints `pmb is now acta; this name goes away in a later version` to stderr first.

The plugin writes only inside its own folder, `~/.acta/`, and the repo's `.acta/`.
It edits CLAUDE.md or AGENTS.md only between acta markers, and only after your
yes in `/acta:setup`. It never edits settings.

## First run

After the install steps above, run `/acta:setup` once. It runs `acta doctor`
first, then asks for your chat language, style and, if you want, a tone, and
saves them.

`acta doctor` checks the install on its own and exits non-zero when something
is missing. `acta doctor --fix` repairs the repo-side items it can, and
re-running it changes nothing when there is nothing left to fix.

In omp, skill names carry no prefix: the skill is `setup`, not `acta:setup`.
Every acta skill description starts with `acta: ` so the index is readable.

## Voice

On the first session, the agent asks which language to chat in, which style
(`adhd` or `plain`) and, if you like, a tone. It saves your answers with:

```bash
acta config set --language Korean --style adhd --tone "Casual, short sentences."
```

The setting lives in `~/.acta/config.yaml` (or `PM_VOICE_FILE`). Change it any
time with `/acta:setup` or `acta config set`. Files written to the repo stay in the
repo language (English unless you set `--repo-language`). If your CLAUDE.md
names a language, it wins.

The reply rules behind `adhd` and `plain` live in an output style that ships
with the plugin, also named `acta`. Its core rules keep every reply short and
direct, and its ADHD block applies only when your style is `adhd`. In Claude
Code the style is on whenever the plugin is on, and it replaces the
`outputStyle` you picked yourself. Claude Code reads styles at start, so a
change needs a restart. omp has no output styles, so the extension adds the
same text to the session rules.

## Lean coding guide

`acta:lean` is a short coding guide: understand the code first, then make the
smallest change that is still right, and never cut the checks that protect users
or data. It is on by default. The session start text carries a short form of it,
and `acta:slice` adds a lean line to every new plan.

Turn it off for all your repos, or for one repo only:

```bash
acta config set --coding-guide off
acta config set --repo --coding-guide off
```

`--repo` saves the choice in the repo's `.acta.yaml`. With the guide off, the
short form and the plan line go away, and the skill stays installed.
`acta config show` lists the value in force.

## Other workflow plugins

`acta` is meant to be the only workflow plugin active in a repo. When
superpowers, gstack, Matt Pocock's skills or another plugin listed in
`hooks/workflow-plugins.txt` is enabled, the agent tells you once and shows how
to turn it off for that repo. It never turns anything off itself.

That file also lists two plugins that are not workflow plugins but overlap
acta: caveman, which the `acta` output style replaces, and a coding-guide
plugin, which `acta:lean` replaces. The agent names them once in the same way.
Run next to acta, they load the same rules twice.

`acta` itself works with any workflow: list another plugin's docs folder under
`legacy` in `.acta.yaml` and the TUI shows it read-only.

## Moving rules out of CLAUDE.md

Optional, and only by your own hand. Once `acta` works for you, these CLAUDE.md
topics are covered by the plugin and can be removed from your file:

| CLAUDE.md topic | Now in |
|---|---|
| Pipeline: brainstorm, plan, approval gates | `acta:shape`, `acta:slice` |
| Worktree for every change, created without asking | `acta:build` |
| Subagent models and the orchestrator writing no code | `acta:build` |
| TDD rules and the test quality bar | `acta:tdd` |
| Debug phases, read-only until the hypothesis | `acta:debug` |
| Review: two axes, BLOCKER or NOTE, three rounds | `acta:review` |
| Landing: gates, merge --no-ff, no menu, never push | `acta:land` |
| Chat language, style and tone | `~/.acta/config.yaml` |

Keep anything personal to you: memory rules, your tool setup, your list of
commands that need a warning. Until you trim, the same rules load twice;
nothing breaks, it only costs context.
