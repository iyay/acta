---
id: BUG-0019
hash: j1b4znk
started: "2026-09-30"
---
# omp runs the dispatch executor and opens another omp tab

## Symptom
With `build_executor: dispatch` in `~/.acta/config.yaml`, an omp session that runs `acta:build` follows `acta:dispatch` and hands the plan to a new omp agent in a new herdr tab. This also hits an omp agent that is itself a dispatch recipient. Dispatch is meant for harnesses other than omp; inside omp, build should fan out to its own `agent("task")` subagents.

## Root cause
- `internal/voice/voice.go:23`: `build_executor` is one global value with no per-harness choice.
- `plugin/skills/build/SKILL.md:22`: when `acta voice show` prints `build_executor`, the skill says use it and do not ask. Nothing checks which harness is running.
- `plugin/omp/index.ts`: the omp extension passes the hook text through and never says the session is omp.

## Repro
1. Set `build_executor: dispatch` (`acta voice show` prints it).
2. In omp, with herdr, run `acta:build` on an approved plan (or receive a dispatch brief).
3. The agent follows `acta:dispatch` and creates another omp tab instead of running subagents.

## Found in
main at a4b3c22, found by reading (acta:debug) after a user report.
