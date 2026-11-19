---
parent: scratch/2026-11-19-cache-expiry-nudge
id: SPC-0216
created: "2026-11-19 20:59:45"
hash: x98gt64
started: "2026-11-20 05:13:15"
finished: "2026-11-20 05:22:12"
---
# Nudge when the prompt cache has likely expired

## Why

Claude Code keeps the prompt cache for about one hour. After a longer break, the next turn sends the whole context again at full input price, and every later turn reads that new copy. A session with a big context pays a lot for that one turn. Chisle 3.14.2 warns about this in its UserPromptSubmit hook (`hooks/chisle-mode-tracker.js` lines 34-62). The user asked for the same nudge in acta on WhatsApp (message 3A800FBD4A1BCFEBCFE5, 20:57 on 2026-11-19) and approved this design (message 3A7BBDFCD4A812B49DED, 20:59).

## Design

`acta hook prompt` starts reading its stdin. Claude Code sends a JSON event there with `transcript_path`. The hook reads the last 512 KiB of that file and walks it from the end to the newest `assistant` line that has `message.usage` and a model other than `<synthetic>`. Context = `input_tokens + cache_creation_input_tokens + cache_read_input_tokens`. Idle = now minus that line's `timestamp`.

Three edge rules for the walk:
- A `system` line with `subtype` `compact_boundary` found before any usage line stops the walk with no line. The user just compacted, so the old context is gone.
- When the newest usage line has no `timestamp`, or one that does not parse, the hook adds no line. It does not go on to older lines.
- When the file is bigger than 512 KiB, the first line of the tail is cut in the middle, so the walk ignores it.

When idle is more than 1 hour and context is 150000 tokens or more, the hook adds one more line under the voice line. The line is for the agent, so it reaches a user who reads the chat on a phone:

`acta cache: <gap> idle with ~<N>k tokens of context, so the prompt cache has likely expired and this turn pays for the whole context again. Tell the user in one line and suggest /compact to go on from a summary, or /clear for a new task.`

To build `<gap>`, idle is first rounded to whole minutes `m`. Under 120 it prints `<m> min`. Under 2880 it prints `<h> h`, with h = m/60 rounded. From 2880 up it prints `<d> days`, with d = m/1440 rounded. `<N>` is context/1000 rounded. Rounding is half away from zero (Go `math.Round`).

Both limits are fixed constants. The hook adds nothing when:
- stdin is empty or not JSON, or has no `transcript_path`;
- the file is missing or no line parses;
- the hook runs under Codex (`hook.IsCodex`), whose transcript has another shape.

The exit code stays 0 in every case. The voice line is printed exactly as now. `now` comes in as a value, so tests pin it.

Files: `internal/hook/` (a new `cache.go` with the parse and the line, plus a test) and `internal/cli/hook.go` (the `prompt` case reads stdin and prints the extra line). The shell script `plugin/hooks/prompt-reminder` already passes stdin through, so it stays as it is. No new dependency.

## Verification

Red first: table tests over small fixture transcripts. Cases in `internal/hook`:
- idle 2 h with 200k context gives the line, text checked exactly;
- idle 30 min gives none;
- idle exactly 1 h gives none; context exactly 150000 gives the line;
- 100k context gives none;
- a synthetic last line is skipped for the real one before it;
- a broken last line is skipped;
- a `compact_boundary` line after the last usage line gives none;
- a usage line with a bad timestamp gives none;
- a file over 512 KiB whose cut first line looks like usage gives none from that line;
- a missing file gives none;
- the gap format at 119 min, 2 h, 47 h and 2 days.

Cases in `internal/cli`, each checking the voice line still comes first and the exit code is 0:
- an event on stdin with a stale transcript gives the line;
- empty stdin, non-JSON stdin, and JSON with no `transcript_path` give none;
- the same stale event under Codex (`PLUGIN_ROOT` set with `t.Setenv`) gives none. Then `scripts/test ./internal/hook ./internal/cli` and `go vet`. The full suite runs at land. The diff does not touch `plugin/skills/` or `plugin/hooks/`, so the eval gate does not run.
