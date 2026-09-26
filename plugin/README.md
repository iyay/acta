# pm

A lean workflow plugin for Claude Code and omp. It covers brainstorm, plan,
build, test first, debug, review and land, and writes planning files to `.pm/`
through `pmb`, the pm-board CLI.

## Install

1. Build `pmb` from the pm-board repo and put it on your PATH:
   ```bash
   go build -o "$(go env GOPATH)/bin/pmb" ./cmd/pmb
   ```
2. Claude Code: add this folder as a marketplace and install the plugin:
   ```text
   /plugin marketplace add /path/to/pm-board/plugin
   /plugin install pm@pm-local
   ```
3. omp: `omp plugin link /path/to/pm-board/plugin`; if the extension does not run, see `omp/FACTS.md#installing-pm-in-omp`.

The plugin writes only inside its own folder, `~/.pm/`, and the repo's `.pm/`.
It never edits your CLAUDE.md, AGENTS.md or settings.

## Voice

On the first session, the agent asks which language to chat in, which style
(`adhd` or `plain`) and, if you like, a tone. It saves your answers with:

```bash
pmb voice set --language Korean --style adhd --tone "Casual, short sentences."
```

The setting lives in `~/.pm/voice.yaml` (or `PM_VOICE_FILE`). Change it any
time with `/pm:setup` or `pmb voice set`. Files written to the repo stay in the
repo language (English unless you set `--repo-language`). If your CLAUDE.md
names a language, it wins.

## Other workflow plugins

`pm` is meant to be the only workflow plugin active in a repo. When
superpowers, gstack, Matt Pocock's skills or another plugin listed in
`hooks/workflow-plugins.txt` is enabled, the agent tells you once and shows how
to turn it off for that repo. It never turns anything off itself.

`pmb` itself works with any workflow: list another plugin's docs folder under
`legacy` in `.pm.yaml` and the TUI shows it read-only.

## Moving rules out of CLAUDE.md

Optional, and only by your own hand. Once `pm` works for you, these CLAUDE.md
topics are covered by the plugin and can be removed from your file:

| CLAUDE.md topic | Now in |
|---|---|
| Pipeline: brainstorm, plan, approval gates | `pm:brainstorm`, `pm:plan` |
| Worktree for every change, created without asking | `pm:build` |
| Subagent models and the orchestrator writing no code | `pm:build` |
| TDD rules and the test quality bar | `pm:tdd` |
| Debug phases, read-only until the hypothesis | `pm:debug` |
| Review: two axes, BLOCKER or NOTE, three rounds | `pm:review` |
| Landing: gates, merge --no-ff, no menu, never push | `pm:land` |
| Chat language, style and tone | `~/.pm/voice.yaml` |

Keep anything personal to you: memory rules, your tool setup, your list of
commands that need a warning. Until you trim, the same rules load twice;
nothing breaks, it only costs context.
