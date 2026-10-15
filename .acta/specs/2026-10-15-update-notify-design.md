---
parent: scratch/2026-10-15-update-notify
id: SPC-0132
created: "2026-10-15 09:32:35"
hash: b5r37ho
started: "2026-10-15 09:36:30"
finished: "2026-10-15 09:41:22"
---
# Update notice: tell the user when a newer acta exists

Status: Architectural, approved by the user in chat on 2026-10-15, section by section. Full answers are in the Log of SCR-0069.

## Goal

Every run of acta (CLI, hooks, TUI) learns when a newer release exists and tells the user to run `acta update`. It never installs anything by itself. It never makes a run slower.

## Part 1: check and cache

- Cache file: `update.json` in the acta user dir (`~/.acta/`, next to `config.yaml`). It holds `{"latest": "v0.1.47", "checked": "<RFC 3339 time>"}`.
- Before command dispatch in `cli.Run`, acta decides whether to start a check. It does nothing when any of these hold:
  - the binary was built from source (`update.IsRelease()` is false);
  - `ACTA_NO_UPDATE_CHECK=1` is set;
  - `update_check: off` is in `~/.acta/config.yaml` (new field on `config.User`, default on);
  - the command is `update` itself.
- When the cache is missing, or its file mtime is older than 24 hours, acta first touches the file (so a parallel run sees it fresh and does not start a second check), then starts `acta update --check --cache` as a detached process: stdin, stdout and stderr go to the null device, and it runs in its own session. The parent does not wait.
- `--cache` is a hidden flag, like `--refresh-plugin`. With it, `update --check` calls `update.Latest` with a 10 second timeout and writes the cache atomically (temp file plus rename). On any error it prints nothing and exits; the next try comes 24 hours later.
- After `acta update` swaps the binary, it writes the cache with the new version, so the notice goes away at once.

## Part 2: where the notice shows

A notice shows only when the cached `latest` is a higher semver than the running version. Equal, lower or unparsable means no notice.

- CLI: after the command finishes, one line on stderr: `acta v0.1.47 available (you have v0.1.46). Run: acta update`. Not shown for `acta hook ...`, for a command run with `--json`, or when stderr is not a terminal.
- TUI: a small badge in the title bar (`internal/tui/title.go`): `↑ v0.1.47`. Read from the cache when the TUI opens; not refreshed while it runs.
- Hook: one more line in `SessionStart` (`internal/hook/hook.go`), only when an update exists: `acta update available: v0.1.47 (you have v0.1.46). Tell the user once, in a single line, to run acta update.` Not added to the prompt hook, so it is not repeated on every prompt.

## Testing

- Unit tests for the decision: stale vs fresh, each skip reason, and the semver compare.
- The spawn goes through a variable a test can swap (same pattern as `update_cmd.go`); a test checks it runs at most once and only when stale.
- `--cache` writes the file on success and leaves it alone on failure (fake server, as in `internal/update` tests).
- Hook output test for the new `SessionStart` line, present and absent.
- TUI title render test with and without the badge.

## Out of scope

- Installing the update by itself.
- Release notes or a changelog in the notice.
- Windows (SCR-0058).
