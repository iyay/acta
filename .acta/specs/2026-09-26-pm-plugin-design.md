---
id: SPEC-2
hash: i2b1
---
# pm: a lean workflow plugin that replaces superpowers

Date: 2026-09-26. Sub-project 2 of pm-board.
Builds on: the file contract (`.pm/specs/2026-09-26-file-contract-design.md`) and `pmb`
(`.pm/specs/2026-09-26-tui-design.md`, on branch `tui` until it lands).

## Why

The superpowers plugin drives the whole workflow (brainstorm, plan, build, review, land)
but does not fit how this user works. Their `~/.claude/CLAUDE.md` holds nine `OVERRIDE:`
lines and about 150 lines of rules that restate or bend superpowers skills. superpowers
also writes to `docs/superpowers/`, while pm-board wants every planning file under `.pm/`.

`pm` is the user's own plugin: only the skills they use, with their rules built in, writing
the pm-board file contract.

## Goals

- Replace superpowers for daily work in Claude Code and omp.
- Every rule the user now keeps in CLAUDE.md about the workflow lives in a skill instead.
- Planning files land in `.pm/` through the contract, and bugs go through `pmb`.
- Chat language and style are a per-user setting, English and ADHD style by default.
- Never touch the user's own instruction files.

Not goals: support for Codex, Cursor, Gemini or Copilot (later, for open source); automated
behaviour evals; a skill-writing skill.

## 1. Structure and install

```
pm-board/
  plugin/
    .claude-plugin/plugin.json       Claude Code plugin manifest
    .claude-plugin/marketplace.json  local marketplace; omp reads it too
    skills/<name>/SKILL.md           one folder per skill, no nesting
    references/                      house rules, prompt templates, examples
    hooks/hooks.json                 Claude Code hooks
    hooks/session-start              runs `pmb hook session-start`
    hooks/prompt-reminder            runs `pmb hook prompt`
    hooks/default-rules.md           used when pmb is not installed
    omp/                             omp extension (package.json + index.ts)
    NOTICE                           MIT notice for text copied from superpowers
  cmd/pmb, internal/...              pmb (sub-project 1)
```

- The plugin lives in the pm-board repo so skills and `pmb` always ship at matching versions.
- Plugin name `pm`. Skills are called `pm:brainstorm`, `pm:build`, and so on. The prefix
  does not clash with `superpowers:*`, so both can be installed during the switch.
- **Claude Code:** add the repo's `plugin/` folder as a local marketplace, then install `pm`.
- **omp:** omp reads `.claude-plugin/marketplace.json` as a fallback catalog and installs with
  `omp plugin install pm@<marketplace>` from a local directory. Skills are discovered one
  level under `skills/`, which the layout above follows. omp does not run Claude Code's shell
  hooks; its hooks are TS/JS modules. So `plugin/omp/` holds a small extension (declared in
  `package.json` under `omp.extensions`) that listens on `before_agent_start` and
  `session.compacting` and runs `pmb hook session-start` through `pi.exec`. These facts come
  from omp's docs (`docs/skills.md`, `docs/hooks.md`, `docs/marketplace.md` in
  github.com/can1357/oh-my-pi); plan task 1 proves them on this machine before anything
  else is built.

## 2. Skills

Eleven skills: the ten below plus `migrate` (section 9). Text is copied from superpowers (MIT) where a skill has a source, then cut and
merged with the user's CLAUDE.md rules. Line counts are targets for the `SKILL.md` plus its
references.

| Skill | Source (lines) | Target | Cut | Folded in from CLAUDE.md |
|---|---|---|---|---|
| `brainstorm` | brainstorming (598) | 350 | visual companion | Spike / Bounded / Architectural said out loud; a written spec for trust boundary, auth, money or migration even when Bounded; spec to `.pm/specs/`; `CONTEXT.md` and ADR updates |
| `plan` | writing-plans (220) | 200 | nothing | plan to `.pm/plans/`; every task has a property-shaped verify; waves by file ownership; ponytail ladder; tasks only from the user's ask |
| `build` | subagent-driven-development (1044), using-git-worktrees (167), dispatching-parallel-agents (167) | 450 | per-task reviewer, "fix round R of 5" | worktree created without asking; three executors (below); orchestrator writes no code; model rules; format before commit; stage by path |
| `tdd` | test-driven-development (518) | 350 | nothing | cover the plan, not the universe; a test that stays green with the code reverted is not a test; test tooling gets no review rounds |
| `debug` | systematic-debugging (1017) | 400 | long examples move to references | read-only until the hypothesis phase; a fix re-enters `brainstorm`; a confirmed bug is recorded with `pmb bug new` |
| `review` | requesting-code-review (276), receiving-code-review (205) | 300 | Critical / Important / Minor | two parallel reviewers (Spec axis, Standards axis); BLOCKER or NOTE only; review once at close; three rounds at most |
| `land` | verification-before-completion (120), finishing-a-development-branch (225) | 150 | the merge / PR / keep / discard menu | full gates with output shown; contamination check; `merge --no-ff` into the recorded parent; worktree and branch cleanup; never push |
| `bug` | new | 60 | — | bug template from the contract; `pmb bug new`; set `fixed_in` when the fix lands |
| `dispatch` | the user's agent-dispatch skill | 400 | parts that repeat the separate `herdr` skill | the current agent-dispatch loop: brief, `/goal`, reply-back, verify, review, fix round, land |
| `setup` | new | 40 | — | asks for chat language, style and tone; saves them with `pmb voice set` |

Dropped: `writing-skills` (the existing `skill-creator` covers it), `executing-plans`
(replaced by the inline executor), `using-superpowers` (replaced by the session-start hook).

### The three executors of `build`

| Executor | When | Who writes code |
|---|---|---|
| `subagent` (default) | normal work | the current harness's own subagent tool: in Claude Code the Agent tool on `sonnet`; in omp `agent()` with `agent="task"` |
| `dispatch` | the user wants the work in another tab | omp in its own herdr tab, through the `dispatch` skill; refuses without herdr |
| `inline` | the user says "inline" | the orchestrator itself |

All three keep the same gates: an approved spec and plan, a worktree, TDD, one review at the
close, and an automatic merge when the review is clean.

The separate `herdr` skill (raw pane control) stays outside the plugin; it is a tool, not a
workflow step.

## 3. Voice

A per-user file, `~/.pm/voice.yaml`:

```yaml
chat_language: en    # any language name
style: adhd          # adhd (default) or plain
tone: |              # optional free text
  Casual Jakarta office chat. gue/lo, gak, udah. No formal words.
repo_language: en    # for everything written to the repo
```

- **First session with no file:** the session-start hook tells the agent to ask the user,
  once and in English, for their chat language, style and tone, then run
  `pmb voice set --language <x> --style <y> [--tone <text>]`, which writes the file.
- **Every later session:** the session-start hook injects the voice rules on startup,
  resume, clear and after compaction, and a prompt hook adds one short reminder line to every
  user message (for example `Reply in Korean.`), so a long session does not drift back to
  English.
- `/pm:setup` changes the setting at any time. Editing the file by hand also works.
- `style: adhd` carries the rules of the i-have-adhd plugin, so that plugin can be turned off.
- `repo_language` and the plain-English comment rule are separate from the chat language:
  code, comments, commits, specs and plans stay in `repo_language`.
- Destructive-command warnings and security findings are always written in full sentences,
  whatever the style.

## 4. The user's own files

- The plugin never writes to the user's `CLAUDE.md`, `AGENTS.md`, `settings.json` or any
  file outside its own folder, `~/.pm/`, and the repo's `.pm/`.
- When a user rule and a plugin rule conflict, the user's `CLAUDE.md` wins. If CLAUDE.md
  names a chat language, that language is used.
- Trimming CLAUDE.md is an optional, manual step. The plugin ships a migration guide listing
  which CLAUDE.md topics the plugin now covers. Until a user trims, the same rules load twice;
  nothing breaks, it only costs context.

## 5. Session-start hook

`pmb hook session-start` prints what the agent must know every session, under 60 lines:

1. the eleven skills, one line each, with when to use them;
2. a warning when another known workflow plugin is enabled (section 8);
3. the core rules: no action without an ask, no code before an approved spec, TDD, worktree
   for every change, review once at close within the budget;
4. the voice rules from `voice.yaml`, or the first-run question when the file is missing.

`pmb hook prompt` prints the one-line reminder. When `pmb` is not installed,
`hooks/session-start` prints `hooks/default-rules.md` (English, ADHD) plus a line saying
`pmb` is missing, and the prompt hook prints nothing.

## 6. Testing

- `pmb hook` and `pmb voice set` are Go code in the pm-board module, tested with `go test`:
  defaults with no file, a broken file, `style: plain`, a chosen language in the reminder,
  output under 60 lines.
- A structural test over `plugin/` (Go test): every `SKILL.md` has `name` and `description`
  frontmatter; every skill folder sits one level under `skills/`; line counts stay under
  target plus 20%; no `superpowers:` reference is left; every `references/` path a skill
  names exists; copied files are listed in `NOTICE`.
- Skill behaviour (waiting for approval, parallel waves, review rounds) is proven by dogfood,
  not by automated evals.

## 7. Switching from superpowers

1. Install `pm` in Claude Code and omp. Turn superpowers off for the pm-board repo only
   (`enabledPlugins` in its `.claude/settings.local.json`, after asking the user). Other
   repos keep superpowers.
2. Run at least two pm-board plans end to end with `pm:*`, from brainstorm to land, one of
   them through `dispatch`.
3. Then, with the user's yes at each step: propose a CLAUDE.md diff that the user applies;
   uninstall superpowers and turn off i-have-adhd; point `~/.omp/agent/AGENTS.md` at the
   plugin's house rules; move `~/.claude/skills/agent-dispatch` to an archive.
4. Rollback before step 3 is turning superpowers back on; CLAUDE.md is untouched until then.

### Pass criteria before step 3

Step 3 starts only when the two dogfood plans of step 2 show all three:

1. The agent invokes the right `pm:*` skill for each workflow step without the user reminding it.
2. No code is written before the spec is approved.
3. Review rounds per plan are equal to or fewer than comparable superpowers plans.

If (1) fails, strengthen the skill index in the session-start hook (`internal/hook/hook.go`, one file) and run the dogfood again.

## 8. Living next to other workflow plugins

Users may already run superpowers, gstack, Matt Pocock's skills or similar. Two workflow
plugins in one session both inject "use my skills" rules at start, and the agent is pulled
two ways (`superpowers:brainstorming` or `pm:brainstorm`).

- **`pmb` does not depend on any workflow plugin.** It reads `.pm/` plus any folders listed
  under `legacy` in `.pm.yaml`, so a superpowers or gstack user can use the TUI without the
  `pm` plugin.
- **`pm` is meant to be the only workflow plugin active in a repo.** The session-start hook
  reads which plugins are enabled (read only) and compares them with a list of known workflow
  plugins kept in the hook's config. If one is active, the agent tells the user once at the
  start of the session which plugin it is and how to turn it off for this repo
  (`enabledPlugins` in the repo's `.claude/settings.local.json`).
- The plugin never turns another plugin off. It also does not claim priority while both are
  active: a priority claim in the context cannot be relied on.

## 9. Migrating existing docs (optional)

Without migration, old docs already show in `pmb` as read-only legacy items. Migration is
for users who want to classify them and change their status.

### `pmb migrate superpowers`

superpowers already writes close to the contract (`YYYY-MM-DD-<slug>.md` names,
`### Task N`, checkboxes, `**Spec:**`), so migration is a move with no content change.

- `docs/superpowers/specs/*.md` go to the root folder's `specs/`, `plans/*.md` to `plans/`,
  with `git mv` so file history follows.
- Plan-to-spec links keep working, because the contract matches `**Spec:**` by file name.
- Dry run by default: prints an old-path to new-path table and any name clashes. `--apply`
  does the move.
- `--apply` refuses on a dirty working tree or on any name clash, and makes one commit
  (`pm: migrate docs/superpowers into .pm`). Nothing is deleted; `git revert` of that commit
  undoes it. It never pushes.
- After the move it removes `docs/superpowers` from `legacy` in `.pm.yaml`.
- Exit codes follow the CLI table: 1 for a clash or dirty tree, 3 for a git failure.

### `pm:migrate` skill, for any other format

For gstack, a `.scratch/` tracker, or any format `pmb` does not know:

1. The agent reads the source folder and builds a mapping table: each source file, the item
   it becomes (story, plan with tasks, or bug), and the new path.
2. The user approves the table before anything is written.
3. The agent writes new files in the root folder following the contract. A source file with
   no date in its name takes the date of its first git commit.
4. Source files stay where they are. The user decides when to delete them.
5. One commit for the whole migration. Never a push.

The agent route exists because other plugins' formats are not known in advance and change;
the approved mapping table is the safety check.

`pmb migrate` is not part of the pmb plan now running. It gets its own small plan after that
one lands. The skill count becomes eleven with `migrate`.

## 10. Live progress from agents in worktrees

Agents work in worktrees, so the files they change are not in the folder where `pmb` runs.
Two changes let the TUI show what agents are doing while they work.

### pmb reads every worktree of the repo

- `pmb` lists the repo's worktrees with `git worktree list --porcelain` and reads the root
  folder (and legacy folders) of each one, using the same root folder name as the main tree.
  Bare, prunable and missing worktrees are skipped. A worktree that cannot be read is skipped
  and never hides the main board.
- Files are matched by item ID. A file that exists only in a worktree is shown from there.
  A file that exists in both is shown from the worktree only when its plan has more ticked
  boxes than the main copy; otherwise the main copy wins. Status is then derived as usual.
- Items shown from a worktree carry that worktree's branch: in the TUI list after the title
  (`◐ 3 Post reads own type row · tui`), in the detail pane (`worktree <branch>`), and in
  `--json` as a `worktree` field (`""` for the main tree). `enter` opens the worktree's file.
- The TUI does not write to an item shown from a worktree (`t` and `s` say to edit it there).
  `pmb set` keeps working on the main tree's copy only.
- Live reload watches the worktrees' folders too, plus git's worktrees folder, so a new
  worktree shows up without a restart.

### pmb also reads branches that are not checked out

Planning files are committed on branches, and a branch the user switched away from has no
folder on disk. `pmb` reads those branches straight from git, read only, without a checkout.

- Which branches: local branches not merged into HEAD (`git for-each-ref --no-merged HEAD
  refs/heads`), leaving out the current branch and every branch checked out in a worktree
  (those are read from disk). `.pm.yaml` can narrow it: `branches: []` turns branch reading
  off; `branches: ["feat/*", "fix-*"]` keeps only matching names.
- Only the root folder is read from branches, not legacy folders, so old docs are not read
  once per branch.
- Files come from `git ls-tree` and `git cat-file --batch`, one call each per branch.
- Precedence: files on disk in the main tree, then files on disk in worktrees, then branch
  tips. Same rule as worktrees: a file from later in that order is shown only when it exists
  only there or its plan has more ticked boxes.
- Items from a branch carry the branch label like worktree items, and are read only. Their
  detail pane says `branch <name> (not checked out)`; `enter` says to open the branch with
  `git worktree add` instead of opening a file. `--json` shows `worktree: <branch>` and
  `on_disk: false`, and `path` as `<branch>:<repo-relative path>`.
- Live reload also watches git's `refs/heads` folder, so a commit on a branch shows up.
  Branches whose names hold a `/` sit in subfolders there and refresh on the next reload
  or `r`.

### Agents tick their boxes as they go

- New command `pmb tick <task-id> [--step N | --all]`: ticks the Nth checkbox (counted the way
  the board counts them) or every checkbox of one task in its plan file. It edits only that
  line, keeps the file's line endings, takes a lock file so parallel implementers in one
  worktree cannot overwrite each other, and never commits. Ticking a box that is already
  ticked is not an error. Exit codes follow the CLI table.
- The `build` skill and the dispatch house rules tell every implementer to run `pmb tick`
  from the worktree right after each step. The orchestrator commits the plan file once after
  each wave (`pm: tick wave <n>`), because several implementers share it.

## Open questions

None. Plan task 1 is a spike that proves the omp install and hook facts in section 1; if one
fails, the fallback for skills is a symlink from `plugin/skills` into `~/.omp/agent/skills`,
and the plan is revised before other tasks start.
