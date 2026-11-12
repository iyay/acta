---
parent: scratch/2026-11-12-acta-usage-telemetry
id: SPC-0201
created: "2026-11-12 06:22:58"
hash: w12yomt
status: approved
started: "2026-11-12 06:35:55"
---
# Anonymous daily usage telemetry

Architectural. Worked out with the user on WhatsApp on 2026-11-12, with cold
read fixes F1 to F5. It ports the
telemetry of walkie (walkie SPC-0174) to acta. Feedback is not part of this work.

## Outcome

The maintainer can see how many installs use acta each week, which harnesses
they run, which commands and acta skills they use, whether work gets from spec
to land, which hook rules block most, and which settings are popular. The data
goes to the owner's Google Sheet through the Apps Script collector that walkie
already uses, at no cost.

## What is counted

Counting happens on the user's machine, per local calendar day (`YYYY-MM-DD`).
Each day is one file under `~/.acta/telemetry/days/`.

- `install_id`: 32 lowercase hex characters from 16 random bytes, made on the
  first counted event. It is not derived from any path, host name, user name or
  git identity.
- `acta_version`, `os` and `arch`. `os` and `arch` come from fixed lists of the
  values release builds target; anything else is `other`.
- `commands`: count per top-level acta command, using the names `cli.Run`
  already dispatches (`list`, `show`, `set`, `hook`, `tidy` and the rest).
  Running `acta` with no arguments counts as `tui`. An unknown name counts as
  `other`. Arguments and flags are never recorded.
- `sessions`: count per harness each time the session-start hook runs. The
  harness is `omp` when `ACTA_HARNESS=omp` is set, `codex` when `hook.IsCodex`
  says so, else `claude`. The omp extension in `plugin/omp/` sets
  `ACTA_HARNESS=omp` on every hook call it makes, since nothing else tells omp
  apart from Claude Code. Any other `ACTA_HARNESS` value is `other`.
- `skills`: count per acta skill (`shape`, `slice`, `build`, `tdd`, `review`,
  `land` and the other folders in `plugin/skills/`). A skill outside that fixed
  list is not recorded at all, not even as `other`, because user skill names
  can be private. The signal differs per harness:
  - Claude Code: `Skill` joins the pre-tool matcher in `plugin/hooks/hooks.json`,
    and the hook reads the skill name from `tool_input.skill`, with the
    `acta:` prefix removed.
  - omp: the extension forwards reads of `skill://<name>` to the pre-tool hook.
  - Codex: no skill signal. Codex days send an empty `skills` object.
- `workflow`: four counters, each bumped once per id:
  - `spec_created` and `plan_created`: when `acta id` gives a new SPC or PLN
    number.
  - `bug_created`: when `acta bug new` writes a bug.
  - `plan_landed`: when `acta spec-dates <plan hash>` runs, which land does once
    per plan.
- `blocks`: count per pre-tool block rule: `go_test` and `destructive`.
- `settings`: one snapshot per day of `build_executor`, `plan_depth`,
  `review_depth` and `chat_language`. The first three are already fixed lists.
  `chat_language` maps to a fixed list of common languages; anything else is
  `other`. A missing value is `unset`.
- Every count stops at 100000, so a loop cannot grow a file without end.

Never counted or sent: paths, repo names, branch names, spec, plan or bug text,
prompts, shell commands, host names, user names, git identities.

Several acta processes can count at the same moment. Day files use a lock and
an atomic write with owner-only permissions. A lost increment under a race is
fine; a corrupt or half-written file is not. A day file that fails to parse is
dropped, never sent.

## Consent and notice

Telemetry is on by default, but only in a release binary that has an endpoint.
A binary with no endpoint (every `go install` and `go build`) counts nothing,
stores nothing and sends nothing.

- A notice shows once per machine, on one of two paths only:
  - on stderr, for a command that is not `acta hook ...` and whose stderr is a
    terminal;
  - as one line in the session-start hook text that tells the agent to tell the
    user.
  No other hook ever prints the notice, because nobody reads hook stderr and a
  pre-tool block message must stay clean. The notice says what is counted, that
  no text, paths or names are sent, and how to turn it off. Only those two paths
  mark it shown. The mark lives in `~/.acta/telemetry/state.json`, not under the
  install ID. `off` keeps the mark, so turning telemetry on again shows no
  second notice. Counting does not wait for the notice; the command that prints
  it is counted too.
- `acta telemetry on`, `acta telemetry off` and `acta telemetry status`. Status
  prints the exact request body for each unsent day, today's counts marked as
  not finished, the current state, and its source (setting, environment, or no
  endpoint in this build).
- `DO_NOT_TRACK=1` or `ACTA_TELEMETRY=0` (also `false`, `off`, `no`) beats the
  stored setting. That process counts nothing, and a flush under it sends
  nothing.
- `off` stops counting at once and deletes unsent days and the install ID.
  Turning it on again makes a new install ID, so old and new rows cannot be
  linked.
- The on/off setting lives in `~/.acta/telemetry/`, never in
  `~/.acta/config.yaml`, because the setup wizard rewrites that file.
- `acta setup` prints one line with the telemetry state and how to turn it off.
  It adds no question to the form.
- The README gets a Telemetry section with the same facts.

## Sending

- The session-start hook, and every command that is not `acta hook ...` (the
  TUI included), starts `acta telemetry flush` as a detached process with its
  output thrown away, only when the last flush attempt is more than one hour
  old. So an install that uses only the CLI or the TUI still sends. The time of the last attempt lives in `state.json`. A flush that
  finds another flush running stops at once. The hook never waits for it.
- A flush sends each finished day (never today) as one HTTPS POST. Each request
  has a 20-second timeout and reads at most 16 KiB of the response. Redirects
  follow the walkie rule: only the expected HTTPS Google content redirect, as a
  GET.
- A receipt with `ok: true` and the same `install_id` and `day` marks the day
  sent and deletes it. A receipt with `ok: false` and `retryable: false` deletes
  the day unsent. Anything else, a missing or malformed receipt included, keeps
  the day for the next flush.
- At most 7 unsent days are kept. Older ones are deleted unsent.
- A failure never prints to the user and never counts as anything.

## Request contract

```json
{
  "kind": "acta_telemetry",
  "install_id": "0f3a9c1e5b7d2f4a6c8e0b1d3f5a7c9e",
  "day": "2026-11-12",
  "acta_version": "0.1.60",
  "os": "darwin",
  "arch": "arm64",
  "sessions": {"claude": 3, "codex": 1},
  "commands": {"hook": 120, "show": 14, "tui": 2},
  "skills": {"shape": 2, "slice": 1, "build": 1},
  "workflow": {"spec_created": 1, "plan_created": 1, "bug_created": 0, "plan_landed": 1},
  "blocks": {"go_test": 2, "destructive": 0},
  "settings": {"build_executor": "inline", "plan_depth": "minimal", "review_depth": "deep", "chat_language": "indonesian"}
}
```

Receipts use the walkie shape and echo `install_id` and `day`:

```json
{"ok": true, "install_id": "0f3a9c1e5b7d2f4a6c8e0b1d3f5a7c9e", "day": "2026-11-12", "duplicate": false}
{"ok": false, "error": "invalid_day", "retryable": false}
```

The collector side is a separate plan in the walkie repo: `collector.gs` routes
`kind: "acta_telemetry"` to its own `ActaTelemetry` tab, validates the same way
as walkie telemetry, keeps one row per install ID and day, and returns the
receipts above. This spec only fixes the contract.

## Build

- The endpoint is set at release time with an `-X` line in both goreleaser
  builds, read from a GitHub Actions secret `ACTA_TELEMETRY_ENDPOINT`. When the
  secret is empty the release has no endpoint.
- Tests swap the endpoint through a package variable. No flag or environment
  variable changes the endpoint in a release build.
- Evals run the dev binary on PATH, which has no endpoint, so they never send.

## Validation

Automated tests cover: the notice shown once on both paths; on, off and status;
both environment switches beating the setting; off deleting the install ID and
unsent days; a no-endpoint build counting nothing; unknown commands and harnesses
becoming `other`; `ACTA_HARNESS=omp` counted as `omp`; a Claude Code `Skill`
call and an omp `skill://` read counted; a non-acta skill never recorded; each
workflow counter bumped once per id; no hook other than session-start printing
the notice; a non-hook command starting the gated flush; the 100000 cap; concurrent
increments leaving a valid file; the 7-day limit; the hourly flush gate; success,
duplicate, malformed, retryable and terminal receipts; a receipt for the wrong
install ID or day; and no forbidden data in any payload.

Deploying the collector, setting the secret and seeing the first real row in the
Sheet are owner actions after review. Live telemetry is not claimed to work until
a real row appears.

## Out of scope

Feedback, error keys, review rounds per plan, doctor check keys, a skill signal for Codex, PostHog or any
other service, dashboards, per-event data, opt-in prompts, Windows.
