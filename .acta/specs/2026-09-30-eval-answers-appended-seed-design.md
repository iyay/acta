---
parent: bugs/2026-09-30-eval-answers-appended-seed-has-no-sections
created: "2026-09-30"
id: SPC-0037
hash: sby57ob
status: dropped
finished: "2026-09-30"
---
# The answers-appended eval seeds a current-format scratch item

Status: design approved by the user in chat on 2026-09-30. Bounded: it fixes one eval fixture and adds one check.

## Why

The `answers-appended` eval scores 0.50 on every run, on main as well. `plugin/evals/answers-appended/scaffold.sh` seeds `.acta/scratch/eval-answer.md` in the old format, with no `schema: "1"` and no sections. The scratch skill tells the agent to use `acta scratch add --section words`. On an old item, `acta scratch add` refuses any `--section` on purpose (`internal/write/scratch.go:107`, body schema spec line 55). So the agent stops and the answer never lands. The seed was written before the body schema landed (dc983b7). In a copy of the plugin with a current-format seed, the case passed 3 of 3 runs.

## Design

1. **Seed in the current format.** `scaffold.sh` writes the item the way `acta scratch new` writes it today:
   - frontmatter `id: SCR-0001`, a 7-character `hash: evalans`, `title`, `status: raw`, `created: "2026-01-01"`, `schema: "1"`
   - `# Download button fires twice`
   - `## Words`, then `### 2026-01-01`, then the line "The download button fires twice when it is clicked fast. Which browsers?"
   - empty `## Context`, `## Log` and `## Open questions`
   The graders, the prompt and `case.yaml` stay as they are.
2. **A cheap check that fails first.** A new test in `internal/plugincheck/evals_test.go`:
   - It builds `acta` from `cmd/acta` into a temp folder and puts it on PATH.
   - It runs `answers-appended/scaffold.sh` in an empty temp folder, with a temp HOME.
   - It runs `./bin/acta scratch add SCR-0001 --section words` there with a line of text on stdin.
   - It checks the exit code is 0 and the text is in `.acta/scratch/eval-answer.md`.
   Today it fails with "SCR-0001 is an old item with no sections". So if the item format moves again, `go test` catches it without a paid eval run.

## Out of scope

- The CLI refusal itself: the body schema spec chose it.
- Other eval cases, and the scratch skill text.

## Verify

- `go test ./internal/plugincheck/` is red before the seed change and green after.
- `scripts/eval --case answers-appended` scores 1.00 in 3 of 3 runs, with the branch's `acta` first on PATH.
- The land gate eval runs, because the diff touches `plugin/`. Only `plugin/evals/` changes, not `plugin/skills/` or `plugin/hooks/`, so strictly the gate does not require it; run it anyway, because this change is about the eval.
