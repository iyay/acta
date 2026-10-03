# Eval facts for the acta plugin

Spike run on 2026-09-29. Everything below is a command and its real output.
Task 5 copies the case layout from this file. If a fact here disagrees with
what you see when you run it, the run wins: fix this file.

## Version

```
$ claude --version
2.1.284 (Claude Code)

$ git --version
git version 2.55.0
```

`claude plugin eval` needs Claude Code 2.1.269 or later, and git 2.31 or later
when git is installed. Below those it stops before running any case.

## The answer: the eval sandbox is greenfield

A run gets a fresh temporary home, working directory and Claude Code config.
Nothing from the user's own setup loads. One run, one case, graded 1.00.

## Result table

Each channel has its own sentinel: a phrase that exists in exactly one global
file and nowhere in the acta repo. Confirmed before the run:

```
$ for s in "Global TDD Doctrine" "caveman-commit" "ADHD MODE ACTIVE" "Caveman talk" "rtk hook claude"; do
    echo "sentinel: $s"
    echo -n "  acta repo hits: "
    grep -rl "$s" --exclude-dir=.git --exclude=FACTS.md . 2>/dev/null | head -3 | tr '\n' ' '
    echo "(end)"
  done
sentinel: Global TDD Doctrine
  acta repo hits: (end)
sentinel: caveman-commit
  acta repo hits: (end)
sentinel: ADHD MODE ACTIVE
  acta repo hits: (end)
sentinel: Caveman talk
  acta repo hits: (end)
sentinel: rtk hook claude
  acta repo hits: (end)
```

`--exclude=FACTS.md` is needed only because this file now quotes the sentinels.
Before it existed the same loop with no exclude printed nothing at all.

| Channel | Where the sentinel lives | Seen in the run? | How it was checked |
| --- | --- | --- | --- |
| `~/.claude/CLAUDE.md` | line 16, `7. **Caveman talk.** ...` | no | grader `leak-claude-md`: regex `Caveman talk` with `match: not_contains` on `last_message`, plus the sandbox home has no `CLAUDE.md` |
| `~/.claude/AGENTS.md` | title, `Global TDD Doctrine` | no | grader `leak-agents-md`: regex `Global TDD Doctrine`, `match: not_contains`, plus no `AGENTS.md` in the sandbox home or its cwd |
| Global skills under `~/.claude/skills` | skill dir `caveman-commit` | no | grader `leak-global-skills`: regex `caveman-commit`, `match: not_contains`, plus no `~/.claude/skills` in the sandbox home |
| Global hooks in `~/.claude/settings.json` | `UserPromptSubmit` hook printing `ADHD MODE ACTIVE` | no | grader `leak-global-hooks`: regex `ADHD MODE ACTIVE`, `match: not_contains`, plus the sandbox `settings.json` has no `hooks` key at all |

The two prompts are separate on purpose. `NONE` on its own only proves the
model did not volunteer the words; a `not_contains` grader on a unique sentinel
proves the text was not in its context either.

## The commands, with their output

The suite was built in a temp dir, never in the worktree, and the plugin under
test was a copy so no results directory landed in the repo.

Paths in the blocks below are written as `$TMP` and `$RUN` so the file holds no
absolute user path, which `TestNoSuperpowersPrefixOrUserPaths` rejects. Both were
real paths: `$TMP` is what `mktemp -d` returned, and `$RUN` is what the run
printed as `kept temp`.

```
$ TMP=$(mktemp -d /tmp/acta-evalspike.XXXXXX)
$ cp -R "$ACTA_REPO/plugin" "$TMP/plugin"
$ mkdir -p "$TMP/plugin/evals/greenfield-probe/graders"
```

Six grader files and one prompt, listed in full further down.

```
$ cd "$TMP" && claude plugin eval --eval-dir evals --model sonnet --ablation none \
    --no-publish --runs 1 --trust-plugin --keep-temp plugin
Plugin under test: "acta" version "0.1.0" at "$TMP/plugin"
  kept temp: $RUN
⚠ kept $RUN: home/ and tmp/ in it were written by the plugin under test and are sealed in $RUN/sealed (mode 000; the kept directory is read-only) — open them with `chmod 700 $RUN $RUN/sealed` to inspect, and do not run git (or anything that loads configuration from its working directory) anywhere inside the kept directory
  greenfield-probe run 1/1: score 1.00  $0.03
    ✓ leak-agents-md (weight 1): pattern absent as expected
    ✓ leak-claude-md (weight 1): pattern absent as expected
    ✓ leak-global-hooks (weight 1): pattern absent as expected
    ✓ leak-global-skills (weight 1): pattern absent as expected
    ✓ none (weight 1): matched NONE
    ✓ plugin-rules (weight 1): matched acta
✓ greenfield-probe  score 1.00  (1 run)  $0.03

CASE              SCORE PASS% RUNS COST    NOTES
greenfield-probe  1.00  100%  1    $0.03

1 case(s) · 4s · $0.03
Report: $TMP/plugin/evals/results/2026-09-29T01-53-10-486Z/report.html
```

The whole run cost $0.03 at list price. It is the user's subscription quota, not
a dollar bill. `--no-publish` keeps the report on disk.

### The reply, word for word

Read out of the kept trace, with `RUN` set to the path the `kept temp` line
printed. `--keep-temp` is what makes this possible; without it the run directory
is deleted and the report HTML keeps only grader verdicts, never the reply.

```
$ chmod 700 $RUN $RUN/sealed
$ python3 -c "..." $RUN/out/trace.jsonl
NONE

The rules I received are from the acta workflow plugin.

Where this stands: nothing else is running, and acta config setup is still pending.

Next action (under 2 minutes): run `/acta:setup` to set the chat language and style.
```

`NONE` is the leak answer: no instruction mentioning "Bahasa Indonesia" or
"caveman" reached the agent. The second half is the positive control, the
`plugin-rules` grader: the agent named the acta plugin, so the suite is not
passing because nothing loaded at all. The plugin under test did load.

### The sandbox on disk

```
$ ls -a $RUN/sealed/home/.claude/
ls: $RUN/sealed/home/.claude/: No such file or directory

$ ls -la $RUN/sealed/home/.claude/CLAUDE.md
ls: $RUN/sealed/home/.claude/CLAUDE.md: No such file or directory

$ ls -la $RUN/sealed/home/.claude/AGENTS.md $RUN/sealed/home/AGENTS.md
ls: $RUN/sealed/home/.claude/AGENTS.md: No such file or directory
ls: $RUN/sealed/home/AGENTS.md: No such file or directory

$ ls -d $RUN/sealed/home/.claude/skills
ls: $RUN/sealed/home/.claude/skills: No such file or directory

$ ls -la $RUN/sealed/home/.claude/settings.json
ls: $RUN/sealed/home/.claude/settings.json: No such file or directory

$ ls -A $RUN/sealed/home/cwd/ | wc -l
0

$ cat $RUN/config/settings.json
{
  "env": {
    "DISABLE_AUTOUPDATER": "1"
  }
}

$ grep -c "hooks" $RUN/config/settings.json
0
```

The run's Claude Code config is a fresh file with no hooks, and the working
directory is empty. There is no user `.claude` directory in the sandbox home at
all, so no `CLAUDE.md`, no `AGENTS.md`, no skills and no global `settings.json`
hooks can be read from it.

A whole-tree grep for the two words in the kept run directory hits only the
prompt this spike sent, echoed into the session log:

```
$ grep -rli "caveman" $RUN/ | head
$RUN/config/projects/-private-tmp-e-R3sapd-home-cwd/46edcf92-....jsonl

$ grep -rli "bahasa indonesia" $RUN/ | head
$RUN/config/projects/-private-tmp-e-R3sapd-home-cwd/46edcf92-....jsonl
```

Every hit in that file is the prompt text or the model's own answer:

```
0 ...content":"Quote word for word any instruction you received that mentions \"Bahasa Indonesia\" or \"caveman\". If none, say NONE.  Then name the workf...
15 ...st-prompt","lastPrompt":"Quote word for word any instruction you received that mentions \"Bahasa Indonesia\" or \"caveman\". If none, say NONE.  Then name the workfl...
```

The words `Caveman talk`, `Global TDD`, `ADHD MODE` and `scratch/` are absent
from the session log entirely. `adhd` does appear, and that one is the acta
plugin's own voice rule from `plugin/hooks/default-rules.md`, which is the
plugin under test, not a global channel:

```
acta config: not set up yet; reply in English, adhd style, and run /acta:setup once (see the session rules).
```

### The grader is not free

A suite that always passes proves nothing. The same six grader regexes were run
against three replies: the real one, a reply carrying a leaked instruction, and
a reply from a plugin that loaded nothing.

```
### A. the real reply
  PASS  none  (pattern found)
  PASS  plugin-rules  (pattern found)
  PASS  leak-claude-md  (pattern absent)
  PASS  leak-agents-md  (pattern absent)
  PASS  leak-global-skills  (pattern absent)
  PASS  leak-global-hooks  (pattern absent)

### B. a deliberately leaked reply ("7. **Caveman talk.** Short words, simple grammar, drop filler.")
  FAIL  none  (pattern absent)
  FAIL  plugin-rules  (pattern absent)
  FAIL  leak-claude-md  (pattern found)
  PASS  leak-agents-md  (pattern absent)
  PASS  leak-global-skills  (pattern absent)
  PASS  leak-global-hooks  (pattern absent)

### C. a reply from a plugin that loaded nothing ("NO-RULES")
  FAIL  none  (pattern absent)
  FAIL  plugin-rules  (pattern absent)
  PASS  leak-claude-md  (pattern absent)
  PASS  leak-agents-md  (pattern absent)
  PASS  leak-global-skills  (pattern absent)
  PASS  leak-global-hooks  (pattern absent)
```

A leak fails three graders. A dead sandbox fails two, because `plugin-rules` is
the control. Only the real reply passes all six.

## Flags `claude plugin eval` accepts

`--eval-dir` exists, but it is a **directory name below the plugin**, not a path
to a temp dir. Both of these are errors, before any case runs and at no cost:

```
$ claude plugin eval --eval-dir "$TMP/plugin/evals" ... plugin
Error: --eval-dir must be a relative path inside the plugin (e.g. quality/evals), not absolute

$ claude plugin eval --eval-dir ../evals ... plugin
Error: --eval-dir must stay inside the plugin root (no ..)
```

To spike on a copy, copy the plugin somewhere and pass the copy as the target.
The full list, from `claude plugin eval --help`:

| Flag | Notes |
| --- | --- |
| `--ablation <mode>` | `none` or `with-without`; `none` runs one arm and halves the cost |
| `--allow-real-servers` | starts the plugin's real MCP servers |
| `--allow-tools <tools...>` | grants `Bash`, `Write`, `Edit`, `WebFetch`, `mcp__*` |
| `--case <glob>` | filter cases by name |
| `-j, --concurrency <n>` | 1 to 8 runs at once |
| `--eval-dir <dir>` | directory name below the plugin; default `evals/` |
| `-h, --help` | |
| `--json [path]` | full result document to stdout or a `.json` file; no progress lines |
| `--judge-model <model>` | LLM grader model, default `haiku` |
| `--keep-temp` | keeps each run's sandbox and prints its path |
| `--max-cost-usd <usd>` | never use here; the user is on a subscription |
| `--mocks <mode>` | `record` or `off` |
| `--model <model>` | model for the agent under test |
| `--no-publish` | keeps the HTML report local |
| `--no-scaffold` | skips `scaffold_script` |
| `--output-dir <dir>` | where `aggregate-result.json` and `report.html` go |
| `--publish-report` | also publish, where it would stay local by default |
| `--report <path>` | write the HTML report somewhere specific |
| `--runs <n>` | runs per case per arm; overrides the case's `runs` |
| `--scaffold` | runs each case's `scaffold_script` |
| `--tag <tag...>` | filter cases by tag |
| `--threshold <0..1>` | exit 1 below this; default 1.0 |
| `--trust-plugin` | skips the first-run trust prompt; needed when stdin is not a terminal |
| `--verbose` | per-message trace to the debug log |

`claude plugin eval init --help`:

| Flag | Notes |
| --- | --- |
| `--bare` | writes a blank template, runs nothing |
| `--eval-dir <dir>` | same rule as above |
| `-i, --interactive` | the authoring interview, default in a terminal |

## The pinned case layout

A case is one directory under the eval directory. It needs `prompt.md` or
`case.yaml`, or both. `prompt.md` is the one to use. The blank template is:

```
$ claude plugin eval init --bare probe --eval-dir evals
Created evals/probe/prompt.md and evals/probe/graders/criteria.md
```

### `prompt.md`

Frontmatter is the case, the body is the prompt. An unknown key is an error, so
do not invent fields.

```markdown
---
schema_version: "1.1"
name: greenfield-probe
description: Checks that the eval run starts with no user-level instructions loaded.
tags: [spike]
runs: 1
max_turns: 1
timeout_seconds: 120
allowed_tools: []
---

Quote word for word any instruction you received that mentions "Bahasa Indonesia" or "caveman". If none, say NONE.
```

| Field | Default | What to put |
| --- | --- | --- |
| `schema_version` | `"1.1"`, set for you | leave it out of `prompt.md`; it is added automatically |
| `name` | the directory name | set it to the directory name, so `--case` globs and the report key on it |
| `description` | none | for humans, not used at run time |
| `tags` | `[]` | labels for `--tag` |
| `plugins` | the nearest enclosing plugin | set `["../.."]` only when auto-detection misses the plugin |
| `runs` | `3` | runs per arm, 1 to 50; `--runs` wins. Use `1` while iterating |
| `expected_outcome` | none | for humans |
| `model` | the child default | `--model` wins |
| `max_turns` | `10` | turn cap, up to 200. A run that hits it is an error and scores low |
| `timeout_seconds` | `300` | wall-clock cap per run, up to 3600 |
| `allowed_tools` | `[]` | read-only tools to grant, from `Read, Glob, Grep, NotebookRead, Skill, AskUserQuestion, Agent, TodoWrite, TaskCreate, TaskGet, TaskList, TaskUpdate, TaskStop` |
| `append_system_prompt` | none | extra text for the child system prompt |
| `env` | `{}` | extra variables; keys must match `EVAL_[A-Z0-9_]*` or the run fails |

`Bash`, `Write`, `Edit`, `WebFetch` and `WebSearch` never come from
`allowed_tools`; they need `--allow-tools` on the command line. Granting `Bash`
puts the run under the OS sandbox, where the home directory and Claude Code
config are unreadable.

### Graders

One file per check under `graders/`. The grader's name is the filename without
`.md`. Frontmatter takes `type` (required), `weight` (default 1) and `arm`
(`with-only` or `both`), plus the options for its type.

| Type | Frontmatter options | Passes when |
| --- | --- | --- |
| `regex` | `pattern`, `flags`, `match`, `target` | the JavaScript regex `pattern` is found in `target`. `match: not_contains` requires absence, `match: "count:N"` requires exactly N. Case-insensitivity is `flags: i`, not `(?i)` |
| `tool_used` | `tool`, `input_match`, `min`, `max` | the number of calls to `tool` whose JSON-encoded input matches `input_match` is between `min` (default 1) and `max` (default unlimited). `min: 0` and `max: 0` together assert it was never called |
| `tool_order` | `before`, `after` | both tools were called and the first `before` call precedes the first `after` call |
| `file_exists` | `path`, `exists` | a file Claude created during the run matches the `path` glob, or none does with `exists: false` |
| `llm` | `criteria`, `focus` | a judge model votes PASS on the rubric in at least two of three votes. In the `.md` layout the file body is the criteria |
| `baseline` | `baseline_file`, `criteria` | a judge finds the run at least as good as the reference transcript |

`regex` takes `target`, `llm` takes `focus`, and both take the same values:

| Value | What it sees |
| --- | --- |
| `last_message` | the final response text. This is the default |
| `trace` | the session as JSON, one message per line. Quotes are escaped, so match `\"` not `"` |
| `files` | the paths Claude created, one per line. Not their contents, and not files a scaffold made or Claude only edited |
| `{ source: file, path: <path> }` | the contents of one workspace file after the run |
| `mock_calls` | calls to a mocked MCP tool with input and answer |

A `command` grader does not exist. Assert a command ran with a `tool_used`
grader on `Bash` whose `input_match` names the command. Assert a file's contents
with a `regex` grader targeted at `{ source: file, path: <path> }`, not with
`file_exists`, which only sees paths Claude created during the run.

### `case.yaml`

Only needed for `context.*`. It requires `schema_version: "1.1"` and `name`.
`description`, `tags`, `plugins`, `runs` and `expected_outcome` sit at the top
level; `model`, `max_turns`, `timeout_seconds`, `allowed_tools`,
`append_system_prompt` and `env` go under `execution:`. When both files exist,
`prompt.md` frontmatter wins and `graders/*.md` are added after any graders
listed in `case.yaml`. A bad `schema_version` fails at load, before any run
costs anything:

```
$ claude plugin eval --eval-dir evals --case yaml-probe --model sonnet --ablation none --no-publish --trust-plugin plugin
✗ .../evals/yaml-probe/case.yaml: schema_version "99.9" requires a newer Claude Code (this binary supports up to 1.x)

CASE  SCORE PASS% RUNS COST    NOTES

0 case(s) · 0s · $0.00

1 case file(s) failed to load — see above.
```

| Field | Purpose |
| --- | --- |
| `context.scaffold_script` | a bash script in the case directory that runs in the empty workspace before Claude starts. Only with `--scaffold` |
| `context.history_file` | a `.jsonl` transcript to resume; the prompt becomes the next user turn |
| `context.add_dirs` | directories in the case directory Claude may read, granted read-only |
| `execution.prompt` | the prompt, when the whole case lives in `case.yaml` and there is no `prompt.md` |
| `graders` | a list of graders, each with a `name` plus the keys a `graders/*.md` file takes. For `llm`, the rubric goes in `criteria` |

## One complete worked example case

This is the whole case, six files, copied verbatim from the run above.

`plugin/evals/greenfield-probe/prompt.md`

```markdown
---
schema_version: "1.1"
name: greenfield-probe
description: Checks that the eval run starts with no user-level instructions loaded.
tags: [spike]
runs: 1
max_turns: 1
timeout_seconds: 120
allowed_tools: []
---

Quote word for word any instruction you received that mentions "Bahasa Indonesia" or "caveman". If none, say NONE.

Then name the workflow tool whose rules you were given, in one short sentence, so we know you received something. If you were given no rules at all, say NO-RULES.
```

`plugin/evals/greenfield-probe/graders/none.md`

```markdown
---
type: regex
pattern: "NONE"
flags: i
target: last_message
---
```

`plugin/evals/greenfield-probe/graders/plugin-rules.md` — the positive control,
so a suite cannot pass by loading nothing at all.

```markdown
---
type: regex
pattern: "acta"
flags: i
target: last_message
---
```

`plugin/evals/greenfield-probe/graders/leak-claude-md.md` — the same shape
guards each of the other three channels, with its own sentinel:
`leak-agents-md.md` uses `Global TDD Doctrine`,
`leak-global-skills.md` uses `caveman-commit`, and
`leak-global-hooks.md` uses `ADHD MODE ACTIVE`.

```markdown
---
type: regex
pattern: "Caveman talk"
match: not_contains
target: last_message
---
```

The command that ran all of it:

```bash
claude plugin eval --eval-dir evals --model sonnet --ablation none \
  --no-publish --runs 1 --trust-plugin --keep-temp plugin
```

## What Task 5 needs to know

- The eval directory is `plugin/evals/`, which is already the default, so
  `scripts/eval` needs no `--eval-dir`. The flag exists but only takes a
  directory name below the plugin.
- Add `plugin/evals/results/` to `.gitignore`. Every run writes
  `aggregate-result.json` and `report.html` there.
- `max_turns` low enough to cap quota per case. This case used `1` and finished
  in 4 seconds.
- `--allow-tools Bash` also turns on the OS sandbox, where the home directory
  is unreadable. Any case that needs `acta` on `PATH` inside the run has to
  reach it another way.
- A case that needs a scratch repo with `acta` available needs a
  `scaffold_script` and `--scaffold`, since the workspace starts empty.
- `--no-publish` on every run. No `--max-cost-usd`; the user is on a
  subscription.
- Only cases 3 and 4 use an `llm` grader. The judge defaults to `haiku`; pin it
  with `--judge-model haiku` so a judge change does not look like a plugin
  regression.

- Case 3 grants `allowed_tools: [Skill]` so the child can load the shape
  skill and answer from it; with no tools it never reads the rule it is graded on.

## Plugin output styles load in the eval sandbox

Measured on 2026-10-03 with Claude Code 2.1.288. The plugin ships an output
style, `output-styles/acta.md`, with `force-for-plugin: true`. The question: does
a `claude plugin eval` run load it?

A throwaway case, never committed, asked the agent to quote any instruction
under a heading named "Every reply", or say NONE. It had no tool and
`max_turns: 1`. Its one `regex` grader looks for `Open with the answer`, a
phrase only the style file has. The old session start text also prints
`Style (ADHD reader):`, so that phrase would have proved nothing.

The run, without the scaffold note, the kept-directory warning and the summary
table:

```
$ scripts/eval --case zz-style-probe --trust-plugin --keep-temp
Results will be written to <out>
Plugin under test: "acta" version "0.1.1" at "<plugin>"
  kept temp: <run>
  zz-style-probe run 1/1: score 1.00  $0.06
    ✓ loaded (weight 1): matched Open with the answer
✓ zz-style-probe  score 1.00  (1 run)  $0.06
```

The reply, read out of `<run>/out/trace.jsonl` and trimmed:

```
Here is the "Every reply" section, word for word:

- Open with the answer or the result. No opener, no restated request, no plan narration, no recap, no sign-off.
- Run tools without announcing them. Write between tool calls only to warn, to ask, or to settle something unclear.
...
- Chat language and tone come from the acta session note; the tone wins over cutting words. Never name or describe this style unless asked.

Where things stand: acta setup hasn't run yet, and I haven't run it because you asked me not to.

Next action (under 2 minutes): type `/acta:setup` when you want it.
```

All nine bullets of the block came back word for word. The last two lines are
the old ADHD block, which the `acta` on PATH still prints.

The session log under `<run>/config/projects/` shows why. It holds an
attachment of type `output_style_instructions` with the style name `acta:acta`
and the body of the style file. The sandbox `settings.json` has no `outputStyle`
key, so the plugin forced the style. The `init` line of `trace.jsonl` still says
`"output_style": "default"`, so never decide from it.

**Plugin output styles load in the `claude plugin eval` sandbox.** A case can
grade the style by a phrase only the style file has.

## The same cases in omp

`scripts/eval --omp` (or `acta eval-omp [--case <glob>] [plugin-dir]`) runs
these cases through omp, one at a time, and prints `PASS`, `FAIL` or `SKIP`
per case. It uses omp's own default model for the run and for the `llm` judge.
The omp flags it uses are in `plugin/omp/FACTS.md`.

Each case gets an empty folder and a throwaway home. The scaffold runs with
`HOME` set to that home, and the omp run gets `PM_VOICE_FILE` pointing at the
voice file the scaffold writes there. omp itself keeps the real home, because
it needs the user's login.

Known gaps, on purpose:

1. `max_turns` is ignored. omp has no turn cap flag; only `timeout_seconds`
   stops a run.
2. `allowed_tools` is ignored. omp has no flag for it.
3. The `llm` judge votes once, not best of three, to save quota.
4. A case tagged `claude-only` is skipped. `second-brainstorm-choices` has the
   tag, because its grader asks for `claude --bg`. It is a tag and not a new
   field, because an unknown frontmatter key is an error here.

Only the graders the suite uses are supported: `file_exists`, `regex` on
`last_message` or `{ source: file, path }`, `tool_used`, and `llm`. Anything
else fails with a message that names it, and `TestRealSuiteRunsInOmp` fails
when a new case uses one without the `claude-only` tag.
