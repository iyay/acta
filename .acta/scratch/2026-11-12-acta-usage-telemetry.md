---
id: SCR-0095
hash: y6xxted
title: Anonymous usage telemetry ported from walkie
status: brainstorming
created: "2026-11-12 06:08:02"
schema: "1"
started: "2026-11-12 06:08:09"
finished: "2026-11-12 06:22:58"
---
# Anonymous usage telemetry ported from walkie

## Words

### 2026-11-12

User via WhatsApp 2026-11-12: "tolong cek repo walkie deh.. dia udah implement feedback dan telemetry. boleh porting ke acta juga?" then "maksud gw buat telemetry-nya sih. feedback mah gak usah dulu".

Port walkie's anonymous daily usage telemetry (walkie SPC-0174, .acta/specs/2026-11-10-usage-telemetry-design.md in the walkie repo) to acta. Telemetry only, no feedback.

## Context

## Log

### 2026-11-12

Probe round 1 (WhatsApp). User: "ok semua" to all five recommendations:
- Q1 collector: same walkie Apps Script, new kind acta_telemetry, own ActaTelemetry tab.
- Q2 on by default, one stderr notice per machine, DO_NOT_TRACK / ACTA_TELEMETRY=0 / acta telemetry off; README notice.
- Q3 count: acta command names per day, sessions per harness (session-start hook), acta:* skill loads (pre-tool hook). No error keys yet.
- Q4 send: session-start hook spawns a background acta telemetry flush, at most once an hour, hook never waits.
- Q5 endpoint baked in at release by goreleaser -X from a GitHub Actions secret; go install / go build have none and count nothing.
User then asked: what else could be counted besides those three?

### 2026-11-12

Extra counters offered: workflow (spec/plan/bug created, plans landed), review rounds per plan, pre-tool hook blocks per rule, daily settings snapshot (build executor, plan depth, review depth, chat language), doctor failing check keys.
User: "oke ikut rekomendasi lo" -> add workflow counts, hook blocks per rule, and daily settings snapshot. Review rounds and doctor keys are later.

### 2026-11-12

Probe round 2. User: "ikut semua":
- Q1 notice: stderr once, plus one line in session-start hook text once per machine telling the agent to tell the user, plus README.
- Q2 state lives in ~/.acta/telemetry/, not config.yaml (setup wizard rewrites config.yaml).
- Q3 acta setup shows one telemetry status line and how to turn it off; no new form question.
- Q4 limits same as walkie: 7 unsent days, counts capped at 100000, send at most once an hour.
- Q5 only acta skills from a fixed list are counted; other skills are not recorded at all, not even as other. Chat language maps to a fixed list of common languages, rest other.

### 2026-11-12

Approaches offered: A internal/telemetry package, count in cli.Run and hooks, store per day in ~/.acta/telemetry/, session-start spawns background acta telemetry flush; B same but send inline in session-start with 2s timeout; C send each event live, no local store.
Collector: acta spec writes only the request and receipt contract; collector.gs change is a separate plan in the walkie repo.
User: "setuju sama rekomendasi lo" -> approach A, collector change in walkie.

### 2026-11-12

Design section 1 approved (thumbs up): per local day file in ~/.acta/telemetry/days/ with install_id (32 random hex), acta version, os, arch; counts: commands (top-level, unknown -> other, tui for no args), sessions per harness (claude, codex, omp) at session-start, skills per acta skill at pre-tool (others not recorded), workflow (spec, plan, bug created; plan landed), blocks per rule (go_test, destructive), settings snapshot (executor, plan depth, review depth, chat language fixed list else other). Cap 100000. Lost increment under race OK, corrupt file not; unreadable file dropped unsent. Never sent: paths, repo or branch names, spec/plan text, prompts, shell commands, hostname, username.

### 2026-11-12

Design section 2 approved (thumbs up): on by default only in release binaries with an endpoint; no endpoint means no count, store or send. Notice once per machine on stderr plus one session-start hook line; recorded in ~/.acta/telemetry/state.json; off keeps the notice record. acta telemetry on/off/status; status prints exact request bodies, today's unfinished counts, state and source. DO_NOT_TRACK=1 or ACTA_TELEMETRY=0 beats the setting for counting and flush. Off deletes unsent days and install_id; on again makes a new id. acta setup shows one status line; README gets a Telemetry section.

### 2026-11-12

Design section 3 approved (thumbs up): session-start spawns detached acta telemetry flush when the last attempt is over an hour old; one HTTPS POST per finished day, 20s timeout, 16 KiB receipt cap, walkie redirect rule; receipt handling and 7-day limit as walkie; failures silent and uncounted. Body kind acta_telemetry with install_id, day, acta_version, os, arch, sessions, commands, skills, workflow, blocks, settings. Collector ActaTelemetry tab is a separate plan in the walkie repo. Endpoint by goreleaser -X from GitHub Actions secret ACTA_TELEMETRY_ENDPOINT; tests use a fake HTTP server through a package variable; evals safe since dev builds have no endpoint. Live deploy and first Sheet row are owner actions after review.

### 2026-11-12

Cold read: 5 findings (skill signal missing from hooks, omp undetectable, workflow counters had no trigger, stderr notice unseen from hooks, CLI-only installs never flush). User approved all five fixes (thumbs up); applied to SPC-0201.

## Open questions
