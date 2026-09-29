---
id: SCR-0005
hash: ogxewwz
title: skill-rules-spec
status: brainstorming
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


## Brainstorm 2026-09-28 (Architectural)

Rulings so far:
1. Skill names: keep frontmatter `name:`; prefix every `description:` with "acta: ". Rename at load in omp rejected: `resources_discover` only returns skillPaths (omp 18.3.5 types.ts:742), so it needs generated copies. Upgrade path if a name collision shows up.
2. First-run setup: new CLI `acta doctor` runs the mechanical checks (acta on PATH, harness, stale links, conflicting plugins, .acta/ init, .gitignore for .agents.json). `/acta:setup` runs it, then asks the questions (voice, default executor, CLAUDE.md/AGENTS.md block).
3. `acta doctor` reports every problem with the exact fix command. `--fix` fixes repo-scope items only (.acta/ init, .gitignore). User-scope items (omp links, other plugins) are reported with the command, never fixed.
4. CLAUDE.md/AGENTS.md block: short pointer (about 5 lines) between `<!-- acta:begin -->` and `<!-- acta:end -->`. A re-run replaces only the text inside the markers. Setup shows the exact block and asks yes first. It writes only files that exist and asks which one when both exist. README "never edits" becomes "only with your yes, between markers".
5. acta:scratch trigger: explicit words ("catet", "nanti", "kepikiran") file at once and say one line "Filed SCRATCH-n <title>". An idea guessed from context (side idea mid-build) is asked first. Body = the user's words verbatim, images as paths.
6. Scratch writes (`acta scratch new/add`, `acta set` on scratch) may commit on main during a brainstorm, before the worktree exists. They are data, not code. The spec is still committed in its worktree.

7. One brainstorm per session counts Architectural brainstorms only. Spike and Bounded are free, any number per session. Only Architectural brainstorms must start from a scratch item; Spike and Bounded do not need one.

8. Default build executor lives in the user file ~/.acta/voice.yaml as a new field `build_executor` (subagent | dispatch | inline), set with `acta voice set --executor <x>`. Same in every repo. File keeps its name. No per-repo override (YAGNI).

9. Model rule, Claude Code only: subagents that write code use model "sonnet"; mapping, explore, planning helpers, debug investigation and spikes use model "opus"; reviewers keep the orchestrator's alias. Other harnesses: no model rule. Effort is a note only (the skill cannot set it). Skills touched: plan, brainstorm, debug, build, dispatch, review.

10. Agents view key: `acta doctor` checks `leftArrowOpensAgents` in ~/.claude.json (undocumented key, default true when unset). When false it reports "Open /config, turn on '← opens agents'"; it never writes ~/.claude.json (Claude Code rewrites it live, user-scope, key may change). On handoff choice (a) the agent says "press ← to open agents view".

11. Approach A: one spec, one plan, 3 waves - (1) Go: `acta doctor` + `build_executor` voice field; (2) skill text; (3) README + hooks/default-rules.md.

12. Design section 1 approved: `acta doctor [--fix]` in internal/doctor + internal/cli/doctor.go; lines `ok|warn|fail <name>: <msg>` plus `fix: <cmd>`; checks: binary path+version, harness (Claude plugin enabled, omp acta link target exists), stale omp links (report `omp plugin unlink <name>`), conflicting plugins (hook.Conflicts), repo .acta/ + .agents.json in .gitignore (--fix; reuse hook.EnsureGitignore; skip outside git), leftArrowOpensAgents false = warn, voice/executor unset = warn. Exit 1 only on fail (checks 1, 2, 5). `build_executor` in internal/voice, `acta voice set --executor subagent|dispatch|inline`, bad value rejected, shown in voice show. Table tests with temp home/repo incl. stale link, broken ~/.claude.json, no .gitignore, bad executor.

13. Ruling 9 REPLACED: model rule is optional, Claude Code only. New voice field `subagent_models` (`split` or unset). `split`: coding subagents sonnet, non-coding opus, reviewers orchestrator alias. Unset (default): skills name no model, the user's own config wins. Set with `acta voice set --subagent-models split`, clear with `--clear-subagent-models`; other values rejected. /acta:setup asks only in Claude Code, default no. Six skills carry one sentence gated on it.
14. Design section 2 (skill text) approved with 13: "acta: " description prefix enforced in plugincheck check.go; new acta:scratch skill; brainstorm step 0 scratch item + appends + parent link + one Architectural per session + 3-choice handoff; setup runs doctor then voice, executor, subagent_models, marker block; build reads build_executor; text guard test per rule, red first.

15. Design section 3 approved: default-rules.md + hook.SessionStart get an acta:scratch line, a new setup line, core rule 8 (one Architectural brainstorm per session); unset voice points to /acta:setup; README line "never edits" becomes marker opt-in, new "First run" section, omp naming note; tests for hook text, default-rules vs hook parity, README phrase gone. Hook state and evals stay in SCRATCH-6.
