---
id: SCR-0033
hash: gauqdwi
title: Config for how deep a plan goes
status: brainstorming
created: "2026-10-01 05:15:07"
schema: "1"
started: "2026-10-01 05:16:08"
finished: "2026-10-01 05:27:02"
---
# Config for how deep a plan goes

## Words

### 2026-10-01

User 2026-10-01: "menarik.. kita kayaknya perlu config buat nentuin planing, mau seberapa dalem planing-nya. minimalm hanya instruksi kecil kayak barusan (harusnya lebih Concise), atau default which termasuk nulis real code di step planing."

## Context

Came up while building SPC-0054. The user said "skip plan, langsung build", but acta:build refuses to run without a plan file (tick, reply-back and dispatch read it). So the orchestrator wrote PLN-0063 (.acta/plans/2026-10-01-show-path-lookup.md): 2 tasks, verify lines and short steps, no real code. PLN-0062 is an example of the default depth, with full test and implementation code in its steps.

Setting would live next to build_executor in ~/.acta/config.yaml (internal/config/user.go), read by acta:plan. Two levels named so far: minimal (short instructions, tighter than PLN-0063) and default (real code in the steps).

## Log

### 2026-10-01

Q1 levels: user picked (a) two levels, `minimal` (tasks, files, verify line, one-sentence steps, no code) and `full` (today: real code in steps).

### 2026-10-01

Q2 default: user picked `full` when unset (same as today). Per-run override `/plan minimal|full`, like `/build <executor>`, for that run only, never saved: yes.

### 2026-10-01

Q3 approval: user picked (b) for `minimal`: plan is written, committed, and build starts with no plan yes. The spec yes still gates. `full` keeps the plan yes.

### 2026-10-01

Approach: user picked 1, a `plan_depth` field in the user config (~/.acta/config.yaml, internal/config/user.go), next to build_executor; plan writes `depth: minimal` in plan frontmatter so build knows no plan yes is needed. Rejected: per-repo .acta.yaml (taste is personal), override only (user wants a saved default). Noted: user CLAUDE.md core rule 3 still demands a plan yes; user edits it, not the agent.

### 2026-10-01

Scope widened (user: "gabung"): config becomes two layers. Global ~/.acta/config.yaml stays for personal keys (chat_language, style, tone, theme, subagent_models). Repo .acta.yaml may override repo_language, build_executor, plan_depth. Order: /plan or /build arg, then .acta.yaml, then global, then default. `acta config set --repo <key> <value>` writes .acta.yaml.

### 2026-10-01

Section 1 approved: global ~/.acta/config.yaml gains plan_depth (minimal|full, empty = full). .acta.yaml fileConfig gains optional repo_language, build_executor, plan_depth that beat global; personal keys never per repo. One merge helper in internal/config used by acta config show and internal/hook. config show prints the source, like `plan_depth: minimal (repo)`. `acta config set --repo` writes .acta.yaml, only for the three keys. Bad values in .acta.yaml fail like global ones. Plan depth order: /plan arg, merged config, full; build executor same order. Setup asks depth and executor, then "all repos or this repo only".

### 2026-10-01

Section 2 approved: minimal plan = frontmatter parent + `depth: minimal` (full plans write no depth); header Goal (one sentence), Spec, Tests, Global Constraints (ponytail line + plan-specific only); no Architecture/Tech Stack/File map/Interfaces; Waves kept. Task = title, Files, property verify, three one-sentence boxes (failing test and why, code change, commit message). Code-block and No Placeholders rules apply to full only. Self-review checks spec coverage and property-shaped verify only. Target about 6 lines per task, under 30 lines for 2 tasks (PLN-0063 was 71).

### 2026-10-01

Section 3 approved: plan skill minimal = write, acta id, commit on main, one-line path, invoke acta:build same turn; full waits for yes. Build refuses unless the plan is approved or has `depth: minimal`; spec approval still required. Core rule 2 (internal/hook/hook.go, plugin/hooks/default-rules.md, any plugin/omp copy) gains "a `depth: minimal` plan needs no plan yes". Tests: merge helper, config set --repo, hook output, plugincheck plan/build strings; HOME temp in tests. Out of scope: new eval, user's own ~/.claude/CLAUDE.md rule 3.

## Open questions
