---
parent: specs/2026-10-03-output-style-and-lean-design
depth: minimal
id: PLN-0078
created: "2026-10-03 15:13:53"
hash: ooo3dhj
---
# Output Style, Lean Coding Guide, Session Start Diet Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** acta forces one output style that holds every chat rule, ships an `acta:lean` coding guide that a `coding_guide` key can turn off, and prints about 2508 bytes at session start instead of 3591.

**Spec:** `.acta/specs/2026-10-03-output-style-and-lean-design.md`

**Tests:** fast `scripts/test ./internal/<package> -run <Name>` for the package a task names, omp `(cd plugin && bun test omp/index.test.ts)`, full `scripts/test --full`; land also runs `go vet ./...`, `gofmt -l .`, `(cd plugin && bun test)` and `scripts/eval` with the branch binary on PATH.

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Every text taken from caveman, ponytail or i-have-adhd is rewritten in new words, never copied 1:1. Claude Code's Concise style gives ideas only. The style file is the spec appendix, word for word.
- Byte caps in `internal/plugincheck/budget_test.go` move in the same commit as their file: a new file gets a cap at its real size, a file that shrinks lowers its cap, and a cap goes up only for a growth this plan names.
- Rules 1-6 and rule 8 of the session text stay byte for byte. No eval grader gets looser.
- Run steps use `scripts/test ./internal/<package> -run <Name>`, never bare `go test` and never `./...`.
- Tests that touch config set HOME (or `PM_VOICE_FILE`) to a temp dir and `t.Chdir` into a temp repo, so they never read or write the real `~/.acta/config.yaml` or this repo's `.acta.yaml`.
- An eval costs quota: run only the case the task names, with the branch binary first on PATH: `go build -o "$TMPDIR/acta-bin/acta" ./cmd/acta && PATH="$TMPDIR/acta-bin:$PATH" scripts/eval --case <name>`. Nothing under `plugin/` may hold an absolute user path (`TestNoSuperpowersPrefixOrUserPaths`).
- Comments are plain English a 10-year-old can read. They say why, not what. No `ponytail:` or other marker comments.

## Waves

- Wave 1: Task 1, Task 2, Task 3, Task 4 (no shared file).
- Wave 2: Task 5 (needs `User.CodingGuide` from Task 2 and `style_test.go` from Task 1; shares `budget_test.go` with Task 1).
- Wave 3: Task 6 (adds `lean` to the names-only `Skills` from Task 5; shares `hook.go`, `default-rules.md` and `budget_test.go` with Task 5).
- Wave 4: Task 7, Task 8 (no shared file; Task 7 needs the FACTS.md entry from Task 1 and the style line from Task 5).

## After land (not tasks)

- The landing report tells the user to run `go install ./cmd/acta`, restart Claude Code (styles and skills are read at start), and turn off the caveman and ponytail plugins.

### Task 1: Output style file and the eval sandbox probe

**Files:**
- Create: `plugin/output-styles/acta.md`, `internal/plugincheck/style_test.go`
- Modify: `internal/plugincheck/budget_test.go` (`TestBudgetFiles` also walks `output-styles/`; cap for `output-styles/acta.md`), `plugin/evals/FACTS.md`
- Throwaway, never committed: `plugin/evals/zz-style-probe/`

**verify:** The style file holds the approved text and nothing else: `name: acta`, `keep-coding-instructions: true`, `force-for-plugin: true`, both headings exactly as the spec appendix writes them, and a body of at most 300 words. The FACTS.md entry decides load or no load by a phrase only the style file has (`Open with the answer`), never by one the old hook text also prints (`ADHD reader`). List each key, each heading, and the probe command with its answer.

- [ ] Failing test: `style_test.go` checks the three keys, both headings and the 300-word body cap, and the budget walk covers `output-styles/`; it fails because the file does not exist.
- [ ] Code: write the file from the spec appendix with its cap; add a throwaway `zz-style-probe/prompt.md` (runs 1, max_turns 1, timeout_seconds 120, allowed_tools []) asking the agent to run no setup and no tool and to quote word for word any instruction under a heading named "Every reply", or say NONE; run `scripts/eval --case zz-style-probe`, write the date, the command and the answer to `plugin/evals/FACTS.md`, then delete the probe folder.
- [ ] Commit: `add the acta output style and record whether evals load it`

### Task 2: Config key coding_guide

**Files:**
- Modify: `internal/config/user.go` (`User.CodingGuide`, yaml `coding_guide,omitempty`; `Validate` takes "", `lean`, `off`), `internal/config/repo_user.go` (`RepoKeys` gains `coding_guide`; a `MergeRepo` case), `internal/cli/config_cmd.go` (`--coding-guide lean|off` on `config set` and `config set --repo`; the `setRepo` check case; `config show` prints `coding_guide: lean (default)` when unset and `(repo)` when the repo set it; JSON key `coding_guide`; `configUsage` and the `--repo takes only` message name it)
- Test: `internal/config/user_test.go`, `internal/config/repo_user_test.go`, `internal/cli/config_repo_test.go`

**verify:** No path stores or accepts a `coding_guide` other than `lean` or `off`: the user file, `.acta.yaml`, `config set`, `config set --repo`. An unset value reads as lean everywhere it is shown, and a repo value beats the user value and shows `(repo)`. List every read and write path checked.

- [ ] Failing test: tests for `Validate`, `MergeRepo`, both `set` forms and `config show` in text and JSON; they fail because the key does not exist.
- [ ] Code: add the field, the check, the repo key and the flag, following `PlanDepth` and `--plan-depth` line by line.
- [ ] Commit: `add the coding_guide config key`

### Task 3: omp adds the style to the session rules

**Files:**
- Modify: `plugin/omp/index.ts` (`createState(run, readDefaults, readStyle)`; the default `readStyle` reads `output-styles/acta.md` under `pluginRoot`), `plugin/omp/index.test.ts`

**verify:** Every session-start text omp sends, from acta or from the default rules, ends with the style body without its frontmatter, and no failure to read the style file stops the session or changes the rest of the text. The reminder never carries the style. List each path: acta ok, acta failed, reader throws, file with no frontmatter.

- [ ] Failing test: every `createState` call in the tests passes a `readStyle` stub (none reads the real file, since Task 1 writes it in the same wave); new tests expect the body after RULES and after DEFAULTS, the frontmatter gone, and the old text when the stub throws; they fail because `index.ts` never reads the style.
- [ ] Code: call `readStyle` inside a try/catch, drop a leading `---` to `---` block, and append the body after a blank line to the rules part only.
- [ ] Commit: `omp: add the acta style to the session rules`

### Task 4: caveman and ponytail overlap acta

**Files:**
- Modify: `plugin/hooks/workflow-plugins.txt` (add `caveman` and `ponytail`; the top comment says plugins that overlap acta and no longer names pm), `internal/doctor/doctor.go` (`checkConflicts` says a plugin that overlaps acta is enabled), `internal/doctor/doctor_test.go`, `internal/plugincheck/plugin_test.go` (`TestWorkflowPluginsList` wants both names), `internal/cli/doctor.go` and `internal/cli/hook.go` (the comment and the `--known` flag help say plugins that overlap acta; `internal/cli/hook.go` still names pm)

**verify:** caveman and ponytail are flagged wherever doctor reads the list, matched by plugin or marketplace name in any case, and no doctor message or list comment still calls them workflow plugins or names pm. List each message and comment checked. The session start note changes in Task 5.

- [ ] Failing test: a doctor test with `caveman@caveman` and `ponytail@ponytail` enabled, reading the real `plugin/hooks/workflow-plugins.txt`, expects a warn that names both and says "overlaps acta", and `TestWorkflowPluginsList` wants both names; they fail because the list lacks them.
- [ ] Code: add the two names and the new wording.
- [ ] Commit: `flag caveman and ponytail as plugins that overlap acta`

### Task 5: Session start text

**Files:**
- Modify: `internal/hook/hook.go` (`Skills` becomes names only, no `When`; header names the skills in one line; rule 7 keeps the same map in fewer bytes; Voice drops the destructive-warning line and adds `- Style: <style>.`; first-run and broken-config text say `Style: adhd`; `adhdRules` goes; the lean summary from spec section 2 prints unless `Voice.CodingGuide` is `off`; the conflicts note and the comments on `Input.Conflicts` and `LoadKnown` say a plugin that overlaps acta), `internal/hook/hook_test.go`, `cmd/acta/hook_test.go` (its check that the output has no "Another workflow plugin" follows the new wording, so it still means something), `plugin/hooks/default-rules.md` (regenerate: `scripts/test ./internal/hook -run TestDefaultRulesFile -update`), `internal/plugincheck/final_test.go` (`TestSkillFoldersMatchHookIndex` reads names), `internal/plugincheck/style_test.go` (the hook line matches the style heading), `internal/plugincheck/budget_test.go` (`sessionStartCap` to the new size)

**verify:** For every input shape (first run, broken config, adhd, plain, coding_guide unset, lean and off, herdr, conflicts) the text names every skill in `Skills`, holds exactly one style line, `Style: adhd` for adhd, first run and broken config and `Style: plain` for plain, has no adhd block and no destructive-warning line, keeps rules 1-6 and 8 byte for byte and all ten superpowers names in rule 7, prints the lean summary unless coding_guide is off, and fits the new cap. List each shape and what it printed.

- [ ] Failing test: rewrite the `hook_test.go` checks for the new text (style line, no `Style (ADHD reader):`, lean on and off, names-only index, each superpowers pair, "overlaps acta") and add the hook-to-heading match in `style_test.go`; they fail because the hook still prints the old text.
- [ ] Code: change `SessionStart` and `Skills`, regenerate `default-rules.md`, and set `sessionStartCap` to the new size.
- [ ] Commit: `cut the session start text and move the chat rules to the style`

### Task 6: Lean skill replaces the ponytail lines

**Files:**
- Create: `plugin/skills/lean/SKILL.md`, `internal/plugincheck/skill_lean_test.go`
- Modify: `plugin/skills/slice/SKILL.md` (the `minimal` depth bullet says "the lean line"; the `## Keep it small` line becomes: unless `acta config show` says `coding_guide: off`, every plan's Global Constraints carry "Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility."), `plugin/references/house-rules.md` (delete the PONYTAIL line), `plugin/skills/build/dispatch.md` (drop ponytail from "(TDD, ponytail, COMMENTS, PARALLEL)"), `plugin/NOTICE` (the ponytail and caveman entries from spec section 7; i-have-adhd points at the ADHD block of `output-styles/acta.md`), `internal/hook/hook.go` (`Skills` gains `lean`), `plugin/hooks/default-rules.md` (regenerate), `internal/plugincheck/skill_slice_test.go` ("ponytail-lazy" becomes "Follow acta:lean"), `internal/plugincheck/plugin_test.go` (`TestNoticeAndReadme` wants `DietrichGebert`, `Julius Brussee`, `skills/lean/`, `output-styles/acta.md`), `internal/plugincheck/budget_test.go` (caps for the lean file and description at real size; slice, house-rules and dispatch at their new sizes; `sessionStartCap` up by the bytes `lean` adds to the index)

**verify:** "ponytail" cannot come back, in any case, anywhere under `plugin/` except `NOTICE` and `hooks/workflow-plugins.txt`; the lean skill is at most 400 words, keeps every never-cut item from spec section 4, and has no level and no `ponytail:` marker; a plan gets the lean line unless coding_guide is off. List every ponytail hit removed and every file the scan covers.

- [ ] Failing test: `skill_lean_test.go` uses `CheckSkill` with the ladder and never-cut words as `Must`, checks the 400-word cap, and scans every file under `plugin/` (no `node_modules`) for "ponytail" outside the two allowed files; with the slice and NOTICE pins it fails because the skill is missing and four ponytail lines remain.
- [ ] Code: write the skill from spec section 4 in new words (description starts with `acta: ` and stays short), change the four lines, NOTICE and `Skills`, regenerate `default-rules.md`, and set the caps.
- [ ] Commit: `add the lean skill and replace the ponytail lines`

### Task 7: Style eval case

Skip this task when the Task 1 FACTS.md entry says the eval sandbox does not load plugin output styles: tick it with that reason and change nothing.

**Files:**
- Create: `plugin/evals/style-short-answer/` (`prompt.md`, `case.yaml`, `scaffold.sh` that writes an English, adhd `$HOME/.acta/config.yaml` only when missing, the way `note-to-scratch/scaffold.sh` does, and two regex graders on `last_message`: the first line answers the question, and no opener and no closing pleasantry, with `match: not_contains` and `flags: i`)
- Modify: `internal/plugincheck/evals_test.go` (`evalCases` names a plugin-relative file in place of a skill folder, so a case can guard `output-styles/acta.md`; the new row guards `Open with the answer`)

**verify:** The case can fail: a reply with an opener, a closing pleasantry, or a first line that does not answer breaks a grader, and a direct reply passes. It adds no llm grader (`TestEvalLLMGradersAreOnlyTheTwoJudgementCalls` stays as it is), sets max_turns and timeout_seconds, and guards a phrase the style file has. Show each regex against one bad and one good sample reply, and the case run with the branch binary on PATH.

- [ ] Failing test: add the `evalCases` row with the file field; `TestEvalCases` fails because the case folder is missing.
- [ ] Code: write the case with a simple question that has one short right answer, then run `scripts/eval --case style-short-answer` as the Global Constraints say.
- [ ] Commit: `add an eval case for the acta style`

### Task 8: README, version rule and the 0.1.2 bump

**Files:**
- Modify: `plugin/README.md` (the forced style replaces the user's `outputStyle` and needs a restart; `acta config set --coding-guide off` and its `--repo` form; the lean skill; caveman and ponytail named in `## Other workflow plugins`, heading unchanged), `CLAUDE.md` (`## Version`: every plan that lands ends with a task that adds 1 to the patch in the three manifests, no exceptions, no folder list), `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The three manifests agree on `0.1.2`; the CLAUDE.md version rule has no exception and no folder list left; the README tells a user every visible change of this plan (forced style and restart, the coding_guide flag in both forms, the lean skill, the two overlapping plugins) and no README line says the adhd rules come from the hook. List each README section checked.

- [ ] Failing test: bump `plugin.json` alone; `TestManifests` fails because the other two still say `0.1.1`.
- [ ] Code: bump the other two, rewrite the CLAUDE.md version rule, and add the README lines.
- [ ] Commit: `document the style and lean guide, bump every landed plan, 0.1.2`
