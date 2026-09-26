# acta

A lean workflow plugin for Claude Code and omp. It covers brainstorm, plan,
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
It never edits your CLAUDE.md, AGENTS.md or settings.

## Voice

On the first session, the agent asks which language to chat in, which style
(`adhd` or `plain`) and, if you like, a tone. It saves your answers with:

```bash
acta voice set --language Korean --style adhd --tone "Casual, short sentences."
```

The setting lives in `~/.acta/voice.yaml` (or `PM_VOICE_FILE`). Change it any
time with `/acta:setup` or `acta voice set`. Files written to the repo stay in the
repo language (English unless you set `--repo-language`). If your CLAUDE.md
names a language, it wins.

## Other workflow plugins

`acta` is meant to be the only workflow plugin active in a repo. When
superpowers, gstack, Matt Pocock's skills or another plugin listed in
`hooks/workflow-plugins.txt` is enabled, the agent tells you once and shows how
to turn it off for that repo. It never turns anything off itself.

`acta` itself works with any workflow: list another plugin's docs folder under
`legacy` in `.acta.yaml` and the TUI shows it read-only.

## Moving rules out of CLAUDE.md

Optional, and only by your own hand. Once `acta` works for you, these CLAUDE.md
topics are covered by the plugin and can be removed from your file:

| CLAUDE.md topic | Now in |
|---|---|
| Pipeline: brainstorm, plan, approval gates | `acta:brainstorm`, `acta:plan` |
| Worktree for every change, created without asking | `acta:build` |
| Subagent models and the orchestrator writing no code | `acta:build` |
| TDD rules and the test quality bar | `acta:tdd` |
| Debug phases, read-only until the hypothesis | `acta:debug` |
| Review: two axes, BLOCKER or NOTE, three rounds | `acta:review` |
| Landing: gates, merge --no-ff, no menu, never push | `acta:land` |
| Chat language, style and tone | `~/.acta/voice.yaml` |

Keep anything personal to you: memory rules, your tool setup, your list of
commands that need a warning. Until you trim, the same rules load twice;
nothing breaks, it only costs context.
