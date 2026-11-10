---
id: SCR-0094
hash: mwzy5gc
title: 'Build executor: default inline, ask subagent vs inline when plans can run in parallel'
status: brainstorming
created: "2026-11-10 16:53:08"
schema: "1"
started: "2026-11-10 16:53:12"
finished: "2026-11-10 18:02:54"
---
# Build executor: default inline, ask subagent vs inline when plans can run in parallel

## Words

### 2026-11-10

help me to iterate the acta build settings

gw kefikiran buat ganti opsi inline sama subagent. by default semua agent ngerjainnya inline, kalo confignya gak diganti sama user sebagai sub-agent, acta harus proactive nanya ke user mau pake subagent atau tetep inline kalo ada multiple plan dan gak saling nunggu (deps) satu sama lain

## Context

## Log

### 2026-11-10

Round 1 asked Q1-Q5 (when the ask fires, dispatch option, where it asks, who builds on subagent, save the answer).
User added Q6: ask each run up to 3 times; on the 4th time ask "you used subagent 3x for independent plans, want it set as the default config?"

### 2026-11-10

User: the Q6 text was the answer to Q5. Rest okay (recommendations taken).
- Q1: ask whenever build_executor is not subagent (unset or explicit inline).
- Q2: dispatch stays, but is not offered in this question; only config or /build dispatch.
- Q3: ask at the end of slice, when 2+ plans are ready now.
- Q4: subagent pick = this session orchestrates: each ready plan gets its own worktree and subagents, in parallel; review and land per plan, one at a time.
- Q5: ask each run; on the 4th time offer to set build_executor: subagent.
- Q6a: count total subagent picks, not in a row.
- Q6b: on no, never offer again; the per-run ask stays.
- Q6c: count lives in ~/.acta/config.yaml, written by an acta command.

### 2026-11-10

Round 3 all recommendations taken:
- Q7: inline with 2+ ready plans keeps today's flow (print build lines for other sessions).
- Q8: subagent mode goes on to after: plans once the plans they wait on land.
- Q9: a stuck plan stops in its worktree and is reported; others go on; plans that wait on it stop too.

### 2026-11-10

User confirmed the summary implicitly and pushed back on approach A: it burns the orchestrator's context.

### 2026-11-10

User asked which option is cheapest in tokens: inline, then B (one worker per plan), then A and C (one worker per task).
User asked why workers are needed at all. Answer: only for parallel work in one session and a small orchestrator context. Today's printed build lines already give parallel work across sessions without workers, at the lowest token cost.
Offered: X keep the subagent option (B); Y no workers, ask to open one session per plan where the harness can; Z drop the feature.

### 2026-11-10

User asked what a worker is; it is a subagent, same thing. User asked for the recommendation: X (subagent option, one subagent per plan). Reasons: Y needs a harness that opens sessions; inline stays the cheap default.
Design section 1 sent: ask at end of slice split; inline not counted; subagent pick runs acta config pick subagent (adds subagent_picks); after 3 picks the 4th time offers build_executor: subagent; no stores subagent_offer: declined; acta config show prints both.

### 2026-11-10

Design section 1 approved.

### 2026-11-10

Design section 2 approved.

### 2026-11-10

Spec written: SPC-0197 .acta/specs/2026-11-10-parallel-plans-subagent-ask-design.md. Asked for cold read and spec review.

### 2026-11-10

Cold read on SPC-0197 found 5 issues: F1 ask before plan approval and clash with slice never builds; F2 the default offer sets build_executor: subagent, which means one subagent per task everywhere; F3 many.md has no entry in the build executor table and the brief drops implementer-prompt rules; F4 land order clashes with the Build brief; F5 resuming a dead parallel run is not defined.
Proposed: F2 option a, a new setting parallel_plans: subagent; F1, F3, F4, F5 fixed as proposed. Waiting for the user.

### 2026-11-10

User took F2 option a (new setting parallel_plans: subagent) and fixes F1, F3, F4, F5. Spec SPC-0197 updated.

### 2026-11-10

Spec SPC-0197 approved by the user (2026-11-10 18:14). Next: acta:slice.

## Open questions
