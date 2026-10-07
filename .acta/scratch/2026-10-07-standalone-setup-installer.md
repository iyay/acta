---
id: SCR-0051
hash: ymeenom
title: Setup wizard outside Claude Code at install time
status: brainstorming
created: "2026-10-07 09:28:20"
schema: "1"
started: "2026-10-07 09:29:01"
finished: "2026-10-07 09:33:10"
---
# Setup wizard outside Claude Code at install time

## Words

### 2026-10-07

# Setup wizard outside Claude Code at install time

User idea, 2026-10-07: setup should run outside Claude Code. When a user installs acta from the command line (for example with bun or cargo), the install drops straight into setup, like installing a skill with npx: an interactive wizard that asks its questions one at a time.

Today setup lives in the acta:setup skill and needs an agent session (acta doctor, chat language, style, tone, build executor, subagent models, the optional acta block in CLAUDE.md or AGENTS.md).

Open questions for the brainstorm:
- Which install channels to support (bun/npm, cargo, go install, brew, a curl script), and which one runs a post-install step at all.
- Whether the wizard is a new `acta setup` command in the Go CLI that the skill then calls, so both paths write the same ~/.acta/config.yaml.
- How the plugin itself gets installed into each harness (Claude Code, omp) from that wizard.
- What runs when there is no TTY (CI), and how to re-run setup later.

## Context

## Log

### 2026-10-07

Q1 install channels: user picked go install + curl script for v1. `acta setup` is the core command; curl script downloads the binary then execs `acta setup`. npm/bun and brew come later.

### 2026-10-07

Q2 plugin install: option 3. Wizard detects claude/omp on PATH, asks per harness [Y/n], runs the install command; on failure prints the command for the user to run by hand.

### 2026-10-07

Q3 no TTY / re-run: option 3. No TTY: exit with a message pointing to `acta config set`. Re-run: current config values are the defaults for each question.

### 2026-10-07

Q4 skill + repo block: option 3. Wizard owns everything; inside a git repo it also offers the CLAUDE.md/AGENTS.md acta block. acta:setup skill becomes thin and calls `acta setup`.

### 2026-10-07

Approach: charmbracelet/huh forms over a pure plan function (answers -> config + actions) tested without a TTY.

### 2026-10-07

Curl script + release pipeline split to its own scratch item (no remote, releases or goreleaser yet). This spec: `acta setup` + go install only.
Section 1 approved: flow = TTY check, doctor, huh questions with current config as defaults, harness detect + ask + install or print command, repo block offer inside a git repo, summary. Decisions in pure plan(answers, env) -> []action.

### 2026-10-07

Section 2 approved: internal/setup (plan.go pure, form.go huh), cli case "setup", acta block text + marker writing moved to Go, acta:setup skill thin, plugincheck skill_setup_test adjusted, fake runner for harness installs, no /init (wizard creates block-only CLAUDE.md), version bump task.

## Open questions
