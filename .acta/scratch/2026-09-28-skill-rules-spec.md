---
id: SCRATCH-5
hash: ogxe
title: skill-rules-spec
status: raw
created: "2026-09-28"
---
# Spec 2: skill rules for scratch, brainstorm and build

Second of three specs (spec 1 = SPEC-10, scratch kind and five-pane sidebar).

Scope agreed 2026-09-27 and 2026-09-28:
- New skill acta:scratch. Raw ideas the user dumps ("catet aja", "nanti", "kepikiran", side ideas mid-build) go to Scratchpad through acta scratch new, never agent memory. Scratch items never need a new session.
- Every brainstorm starts from a scratch item. No item yet: the agent first files the user's words verbatim (images as paths) with acta scratch new, sets it to brainstorming, then brainstorms from it. Each answer and approved section is appended with acta scratch add as it goes, so a dead session loses nothing. The spec links the item with parent: scratch/<stem>.
- At most one brainstorm per session. On a second brainstorm the agent files the idea as a scratch item and asks the user, three choices: (a) background agent: claude --bg 'brainstorm SCRATCH-n', chat in claude agents; (b) new herdr tab, only offered when HERDR_ENV=1; (c) manual new session: the agent copies the prefilled prompt to the clipboard (pbcopy on macOS, wl-copy/xclip on Linux, OSC 52 fallback) and also prints it. Build, review and land may continue in the brainstorm's session.
- Default build executor: acta:setup asks the default acta:build executor (subagent, dispatch, inline); dispatch only offered when herdr is present (HERDR_ENV=1 or herdr on PATH). Changeable later with /acta:setup. acta:build reads the saved default and asks only when none is set.
- Model rule (user 2026-09-28): sonnet subagents only for writing code. Planning, code mapping and exploring subagents run on opus at medium effort or more. acta:plan (and any skill that dispatches non-coding subagents) should say so.
- First-run setup and skill names in omp (user 2026-09-28, tried omp in be-pmis): there is no single first-run command that sets everything up (/acta:setup only sets the voice). In omp, skill names have no plugin prefix, so acta's brainstorm, plan, build and the rest mix with other plugins' skills and the user cannot tell which one is acta's. Wants a clear way to see a skill is acta's in both harnesses (for example an "acta" prefix in the name or description) and a first-run setup command.
- First-run setup should also offer to add acta instructions to the repo's CLAUDE.md / AGENTS.md (user 2026-09-28), plus the other setup steps. Tension to settle in the brainstorm: the README promises the plugin "never edits your CLAUDE.md, AGENTS.md or settings", so this has to be an opt-in step the user says yes to, showing the exact block first. Candidate setup steps: acta CLI on PATH, harness check (Claude Code plugin installed; omp plugin linked and extension running), voice, default build executor, repo .acta/ init and .gitignore for .agents.json, conflicting workflow plugins, stale links (like the old pm link in ~/.omp/plugins), and the opt-in CLAUDE.md / AGENTS.md block.
