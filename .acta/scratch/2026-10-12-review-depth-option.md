---
id: SCR-0061
hash: vhp9n9u
title: Review depth option, ideas from alibaba/open-code-review
status: brainstorming
created: "2026-10-12 06:53:28"
schema: "1"
started: "2026-10-12 06:53:38"
finished: "2026-10-12 07:01:50"
---
# Review depth option, ideas from alibaba/open-code-review

## Words

### 2026-10-12

User ask (2026-10-12): analyse https://github.com/alibaba/open-code-review and bring its ideas into acta. Add a new option so the user can pick the code-review depth / method.

## Context

## Log

### 2026-10-12

Analysis of alibaba/open-code-review (HEAD 357c3f0, Apache-2.0, Go). Its only depth knob is --effort low|medium|high = 1/2/3 review rounds (default medium). Other ideas: a fact-check filter pass that may only drop findings the diff proves wrong (Ground A: code not in diff; Ground B: one diff line contradicts it; protected subjects never dropped); prior findings fed into later rounds with "do not repeat"; early stop when a round adds nothing; churn thresholds (50 lines/file, 100/group) gate a plan phase; per-file-type rule checklists matched by glob; every changed file gets its own pass.
Existing acta setting pattern to follow: plan_depth (user.go, repo_user.go, config_cmd.go, setup form, skill reads acta config show, per-run override like /slice minimal).

### 2026-10-12

Probe round 1, user agreed to all recommendations: (Q1) one key review_depth with levels quick|standard|deep, each level changes which passes run; (Q2) bring in fact-check filter (deep only), carry earlier findings into rounds 2-3 and per-file coverage rule (all levels); skip per-file-type checklists and churn thresholds; (Q3) sensitive paths (security, auth, migration, money, delete) floor at standard plus deep lens, setting can only raise depth; (Q4) user config, repo .acta.yaml and per-run /review <level> override, plus a setup form question, same as plan_depth; (Q5) default standard, three-round cap unchanged at every level.

### 2026-10-12

Design approved (three sections): (1) review_depth key quick|standard|deep, empty = standard, user + repo config, repo wins, setup form question, pick order /acta:review <level> arg then acta config show then standard; (2) levels live in review/depth.md: quick = light tier inline as today, logic gets one combined Spec+Standards reviewer; standard = today; deep = two reviewers plus a read-only fact-check subagent over BLOCKERs only (review/fact-check.md, may only demote with Ground A/B, doubt keeps it, protected subjects never demoted); all levels: per-file coverage line in code-reviewer.md and earlier-round findings passed in with "do not repeat"; (3) floors: sensitive paths, fix rounds and dispatch-executor work are at least standard; three-round cap unchanged; plugincheck caps raised; eval: quick on an auth path still runs two reviewers.

## Open questions
