---
parent: scratch/2026-11-24-repo-init
id: SPC-0220
created: "2026-11-24 18:46:27"
hash: z9wh0pe
started: "2026-11-24 18:59:29"
finished: "2026-11-24 19:36:13"
---
# acta init sets up a new repo

## Why

In a new repo, for example claver, `acta doctor --fix` makes `.acta/` and leaves CLAUDE.md alone, by design. The acta plugin already runs in every repo, since it is turned on at user level. What a new repo still lacks is the acta block in CLAUDE.md or AGENTS.md. Today only the `acta setup` wizard or the `acta:setup` skill writes that block, and both also ask every setting again. The user wants one repo-level step that takes the settings from the global config and asks nothing.

## Design

A new command, `acta init`, run anywhere inside a git repo.

1. Outside a git repo it writes nothing and exits with an error that says acta needs a git repo, since every acta write commits.
2. It runs `doctor.Fix`: the `.acta` folder, its `.gitignore` line and the wiki check record.
3. It writes the acta block with `setup.WriteBlock`, picking files by the rule `appendBlocks` in `internal/setup/plan.go` already uses: both CLAUDE.md and AGENTS.md exist, write both; one exists, write that one; neither exists, create CLAUDE.md. The pick moves into one exported helper that both setup and init call, so the two never drift apart. Before step 2 writes anything, init checks every picked file for a broken marker pair (half a pair, the end before the begin, or a repeated marker, the cases where `WriteBlock` returns an error). One broken file means init writes nothing, prints which file is broken and exits non-zero.
4. It commits every changed path in one new commit, `chore: acta init`. It never folds into HEAD, even when only one path changed, so a user's own commit is never rewritten. `gitc.CommitPaths` folds a single path, so init must not use it for that case. Before writing, init checks whether any of these files is already dirty: the picked block files and the three files `acta doctor --fix` checks (`.acta/.gitignore`, `.gitattributes`, the wiki check record). Any dirty file means init still writes everything but commits nothing, and prints `written, not committed:` with the reason, the same way `acta doctor --fix` skips its whole commit today.
5. It prints the `acta doctor` report. The exit code is 0 when the writes and the commit (or the dirty skip) went fine, whatever the report says. Only an init failure (outside git, a broken marker, a failed write or commit) exits non-zero.

A repo is already set up when no path init would write has different bytes after the writes than before them. `WriteBlock` rewrites its file every time and does not say whether anything changed, so init reads each path before and after and compares the bytes itself. When nothing differs, init makes no commit and says the repo is already set up. When something differs, for example an old block that was refreshed, init commits the paths that differ. `WriteBlock` only swaps the bytes between its markers, so a re-run never touches other text.

The block is fixed text. The settings it points at (chat language, style, build executor and the rest) are read by the hooks from `~/.acta/config.yaml` at session start, so init asks no questions and writes no `.acta.yaml`.

A thin skill, `acta:init`, runs `acta init` and reports what it printed. It is for typing the step from chat.

Also: a `help.go` entry and a case in `cli.go` for `init`, and one eval case tagged `init`.

Out of scope: writing `.claude/settings.json` so people who clone the repo get acta. That needs a github marketplace entry and the binary on their PATH, and lives in SCR-0102.

Adding the skill folder also touches three files `internal/plugincheck` reads: `internal/hook/hook.go` gains `init` in its skill list (the session-start text grows by one word and must stay under its cap), and `internal/plugincheck/budget_test.go` and `internal/plugincheck/evals_test.go` get rows for the new skill file and its eval case. The eval case guards the line in the skill that says to run `acta init` and report its output.

Files: `internal/cli/init.go` (new), `internal/cli/init_test.go` (new), `internal/cli/cli.go`, `internal/cli/help.go`, `internal/setup/plan.go`, `internal/hook/hook.go`, `internal/plugincheck/budget_test.go`, `internal/plugincheck/evals_test.go`, `plugin/skills/init/SKILL.md` (new), one case under `plugin/evals/`.

## Verification

Red first, in a temp git repo:
- A repo with neither file: init makes `.acta/`, creates CLAUDE.md holding the block, and makes one commit.
- A repo with both files: both get the block, other text is kept.
- A second run: every file has the same bytes, `git rev-parse HEAD` is unchanged, and the output says the repo is already set up.
- CLAUDE.md holds an old block between the markers: init refreshes it and commits it.
- A file dirty before init: every file is written, no commit is made, and the output says so.
- Only CLAUDE.md changes, and HEAD is a user commit that touched only CLAUDE.md: init makes a new commit and HEAD~1 is still the user's commit, unchanged.
- One picked file has a broken marker pair: nothing is written, the error names the file, exit is non-zero.
- The doctor report has a failing check after a good init: exit is 0.
- A folder outside git: error, nothing written.

Then `scripts/test ./internal/cli ./internal/setup ./internal/plugincheck`. The full suite and `scripts/eval --tag=init` run at land, since the diff touches `plugin/skills/`.
