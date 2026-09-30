---
parent: scratch/2026-10-01-plan-depth-config
id: SPC-0056
created: "2026-10-01 05:27:02"
hash: f8z5kqo
started: "2026-10-01 05:57:35"
---
# Plan depth setting and repo-level config

Status: design approved by the user in chat on 2026-10-01, section by section. Architectural: it changes the shape of plans that build, dispatch and review read, and how the user config is found.

## Why

Every plan today holds real code in its steps (see PLN-0062). For small work the user wants a short plan, like PLN-0063 but tighter, and wants build to start without a second yes. The user also wants some settings to differ per repo: a small repo can run `minimal` and `inline`, a big one `full` and `dispatch`. Today the only config is the global `~/.acta/config.yaml`.

## Design

### 1. Two config layers

- The global `~/.acta/config.yaml` (`internal/config/user.go`) keeps every key and gains `plan_depth`: `minimal` or `full`. Empty means `full`. Any other value is an error, the same way a bad `build_executor` is.
- The repo `.acta.yaml` (`fileConfig` in `internal/config/config.go`) gains three optional keys: `repo_language`, `build_executor`, `plan_depth`. A value set there beats the global one. Personal keys (`chat_language`, `style`, `tone`, `theme`, `subagent_models`) are global only; one of them in `.acta.yaml` is an error that names the key.
- One helper in `internal/config` merges the user config with the repo overrides and says which keys the repo set. `acta config show` and `internal/hook` both read through it, so the two never disagree.
- `acta config show` marks each value `.acta.yaml` set, for example `plan_depth: minimal (repo)`, and always prints `plan_depth` (`full` when unset).
- `acta config set --repo` with `--repo-language`, `--executor` or `--plan-depth` writes `.acta.yaml`; any other flag with `--repo` is refused. Without `--repo` it writes the global file, as today. A new `--plan-depth minimal|full` flag sets the depth.
- A bad value in `.acta.yaml` fails with the same error as the same bad value in the global file.

### 2. Picking the depth

- The plan skill picks the depth in this order and does not ask when one answers: the argument of `/plan minimal|full` (this run only, never saved), then `plan_depth` from `acta config show`, then `full`.
- The build skill picks the executor in the same order: `/build <executor>`, then `acta config show` (now merged), then ask.
- The setup skill asks for the depth after the executor, then asks once whether to save both for every repo or for this repo only.

### 3. The minimal plan shape

- Frontmatter: `parent:` and `depth: minimal`. A `full` plan writes no `depth` field.
- Header: `**Goal:**` (one sentence), `**Spec:**`, `**Tests:**` (fast and full commands), `## Global Constraints` (the ponytail-lazy line plus only the rules this plan needs), `## Waves`. No Architecture, Tech Stack, File map or Interfaces; the spec holds those.
- Each task: a title, `**Files:**`, a property-shaped `**verify:**`, and three one-sentence boxes: the failing test and why it fails, the code change, the commit message. The boxes stay because `acta tick --step` and the board count them.
- The code-block rule and No Placeholders apply to `full` only. A `minimal` plan holds no code blocks.
- Self-review checks two things: every part of the spec is covered, and every verify line is a property, not one case.
- Target: about 6 lines per task; a 2-task plan under 30 lines.

### 4. Approval

- `full`: unchanged. The plan waits for the user's yes.
- `minimal`: the plan skill writes the plan, runs `acta id`, commits on main, tells the user the path in one line, and invokes `acta:build` in the same turn.
- The build skill refuses to start unless the spec is approved (or, for Bounded work with no spec file, approved in chat) and the plan is approved or has `depth: minimal` in its frontmatter.
- Core rule 2 in `internal/hook/hook.go` and `plugin/hooks/default-rules.md` gains: "a `depth: minimal` plan needs no plan yes". `plugin/omp` holds no copy of this rule today.

## Testing

- Merge helper: repo beats global for each of the three keys; a personal key in `.acta.yaml` is an error; a bad value in either file is an error; the set of keys the repo set is exact.
- A bad `.acta.yaml` never changes the chat language or style in the session hook; the hook names the file and keeps the global values.
- `acta config set --repo`: writes `.acta.yaml`, refuses a personal key, leaves the global file alone.
- `acta config show`: prints the merged value and its source.
- Hook: a `repo_language` set in `.acta.yaml` shows in the session text.
- plugincheck: the plan skill names `depth: minimal` and `/plan minimal`; the build skill carries the new refuse rule; the hook test holds the new rule 2 text.
- Every test that touches config sets `HOME` and `TMPDIR` to temp dirs, so the real `~/.acta/config.yaml` is never written.

## Out of scope

- A new eval case for the minimal plan; add one later if the skill drifts.
- The user's own `~/.claude/CLAUDE.md` core rule 3 still asks for a plan yes. The user edits that file, not an agent.
- A personal, uncommitted per-repo config file. `.acta.yaml` is committed, so its overrides reach everyone who clones the repo.
