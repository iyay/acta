---
parent: specs/2026-11-19-cache-expiry-nudge
depth: minimal
id: PLN-0249
created: "2026-11-20 04:21:12"
hash: a4fi1iu
started: "2026-11-20 05:13:15"
finished: "2026-11-20 05:16:01"
---
# Cache expiry nudge in the prompt hook

**Goal:** `acta hook prompt` adds one line for the agent when the session sat idle more than 1 hour with 150k or more tokens of context, and stays quiet and exit 0 in every other case.

**Spec:** `.acta/specs/2026-11-19-cache-expiry-nudge.md` (SPC-0216)
**Tests:** fast `scripts/test ./internal/hook ./internal/cli`; full `scripts/test --full`.
**Gates:** eval: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched; whole when the diff touches `plugin/hooks/` or `plugin/references/`. This plan touches neither, so the gate does not run.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Stdlib only. Limits (1 hour, 150000 tokens, 512 KiB) are fixed constants.
- The voice line from `hook.Prompt` prints first and unchanged; the hook never exits non-zero.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: Read the transcript tail and build the cache line

**Files:** `internal/hook/cache.go` (new), `internal/hook/cache_test.go` (new)
**Seam:** `hook.CacheNudge(transcriptPath string, now time.Time) string` (empty string means no line)
**verify:** The line appears only when the newest real usage line is more than 1 hour old and its context is 150000 or more, and never after a compact boundary. Every other path gives an empty string and no panic. List each path tried: limits exactly at 1 h and 150000, synthetic last line, broken last line, compact boundary after usage, bad or missing timestamp, cut first line of a tail over 512 KiB, missing file, and the exact text with gap at 119 min, 2 h, 47 h and 2 days.
- [x] Red: table tests over fixture transcripts written in `t.TempDir()` fail because `CacheNudge` does not exist.
- [x] Change: add `CacheNudge` per the spec's Design: read the last 512 KiB, walk lines from the end, stop on `compact_boundary`, take the newest non-synthetic assistant usage, apply the limits and the gap format.
- [x] Commit: feat(hook): build a nudge line when the prompt cache has likely expired

### Task 2: Wire the nudge into acta hook prompt

**Files:** `internal/cli/hook.go`, `internal/cli/hook_session_test.go`
**Seam:** `cmdHook([]string{"prompt"}, stdin, stdout, stderr)`
**verify:** Wiring only. A stale transcript event prints the voice line, then the cache line. Under Codex (`PLUGIN_ROOT` set), and for empty stdin, non-JSON stdin, or JSON with no `transcript_path`, only the voice line prints. Exit code is 0 on every path listed.
- [x] Red: a CLI test feeding a stale event on stdin fails because the `prompt` case never reads stdin.
- [x] Change: in the `prompt` case, read stdin, take `transcript_path` from the JSON, skip when `hook.IsCodex()`, and print `hook.CacheNudge(path, time.Now())` under the voice line when it is not empty.
- [x] Commit: feat(cli): add the cache nudge to the prompt hook

## Review notes

- Review depth auto, picked standard (two reviewers): risk signal is hook text the agent reads. Round 1 range 97f980c..10aaf44 CLEAN on both axes. BLOCKER 0, NOTE 8, pre-existing 0. Notes: exact-text test uses Contains; io.ReadAll waits on a TTY stdin (real callers close it); FIFO transcript path would block; a single line over 512 KiB gives no nudge; gap cases check only the prefix; Codex stdin not drained; hook prompt help does not mention stdin; first tail line dropped even when cut lands on a line start.

## State

### Next

All tasks committed; next is the wiki check, then acta:review over the branch.
