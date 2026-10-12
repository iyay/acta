---
id: SCR-0063
hash: k0lrow8
title: Bad-call usage output is one long line
status: raw
created: "2026-10-12 10:18:48"
schema: "1"
---
# Bad-call usage output is one long line

## Words

### 2026-10-12

catet dulu aja, ini perlu dirapihin outputnya

```
❯ acta hook
usage: acta hook session-start [--known <file>] | acta hook prompt | acta hook pre-tool | acta hook post-tool

❯ acta config
usage: acta config show [--json] | acta config set [--language L] [--style adhd|plain] [--tone T] [--clear-tone] [--repo-language L] [--executor subagent|dispatch|inline] [--plan-depth minimal|full] [--review-depth quick|standard|deep] [--commit-history tidy|full] [--questions one|probe] [--coding-guide lean|off] [--subagent-models split|default] [--clear-subagent-models] [--theme NAME] [--clear-theme] | acta config set --repo [--repo-language L] [--executor E] [--plan-depth D] [--review-depth R] [--commit-history H] [--coding-guide G] [--unset KEY]...
```

## Context

Came right after PLN-0138 (acta help, version and man page) landed on main (a22dc5d). The user ran the release binary on PATH (~/.local/bin/acta 0.1.45), which predates PLN-0138, so this shows the old bad-call output.

On main now: a bad call still prints one `usage:` line built by `usageLine` in internal/cli/help.go, which joins every usage line of the command (and subcommand) with " | ". So `acta config` and `acta hook` with no subcommand still print one long wrapped line. `acta help config` / `acta config --help` print the multi-line table entry (usage lines, flags, example) instead.

Likely shape of the fix: a bad call with a missing or wrong subcommand prints the multi-line usage block (one usage line per row, maybe the subcommand list) plus a hint to `acta help <cmd>`, instead of one joined line. Existing tests assert some `usage: acta ...` substrings.

## Log

## Open questions
