---
parent: scratch/2026-11-19-cache-expiry-nudge
id: SPC-0216
created: "2026-11-19 20:59:45"
hash: x98gt64
---
# Nudge when the prompt cache has likely expired

## Why

Claude Code keeps the prompt cache for about one hour. After a longer break, the next turn sends the whole context again at full input price, and every later turn reads that new copy. A session with a big context pays a lot for that one turn. Chisle 3.14.2 warns about this in its UserPromptSubmit hook (`hooks/chisle-mode-tracker.js` lines 34-62). The user asked for the same nudge in acta on WhatsApp (message 3A800FBD4A1BCFEBCFE5, 20:57 on 2026-11-19) and approved this design (message 3A7BBDFCD4A812B49DED, 20:59).

## Design

`acta hook prompt` starts reading its stdin. Claude Code sends a JSON event there with `transcript_path`. The hook reads the last 512 KiB of that file and walks it from the end to the newest `assistant` line that has `message.usage` and a model other than `<synthetic>`. Context = `input_tokens + cache_creation_input_tokens + cache_read_input_tokens`. Idle = now minus that line's `timestamp`.

When idle is more than 1 hour and context is 150000 tokens or more, the hook adds one more line under the voice line. The line is for the agent, so it reaches a user who reads the chat on a phone:

`acta cache: <gap> idle with ~<N>k tokens of context, so the prompt cache has likely expired and this turn pays for the whole context again. Tell the user in one line and suggest /compact to go on from a summary, or /clear for a new task.`

`<gap>` reads as minutes under 2 hours, hours under 2 days, days after that.

Both limits are fixed constants. The hook adds nothing when:
- stdin is empty or not JSON, or has no `transcript_path`;
- the file is missing or no line parses;
- the hook runs under Codex (`hook.IsCodex`), whose transcript has another shape.

The exit code stays 0 in every case. The voice line is printed exactly as now. `now` comes in as a value, so tests pin it.

Files: `internal/hook/` (a new `cache.go` with the parse and the line, plus a test) and `internal/cli/hook.go` (the `prompt` case reads stdin and prints the extra line). The shell script `plugin/hooks/prompt-reminder` already passes stdin through, so it stays as it is. No new dependency.

## Verification

Red first: table tests over small fixture transcripts. Cases: idle 2 h with 200k context gives the line; idle 30 min gives none; 100k context gives none; a synthetic last line is skipped for the real one before it; a broken last line is skipped; a missing file gives none. One CLI test feeds an event on stdin and checks the voice line still comes first. Then `scripts/test ./internal/hook ./internal/cli` and `go vet`. The full suite runs at land. The diff does not touch `plugin/skills/` or `plugin/hooks/`, so the eval gate does not run.
