---
parent: scratch/2026-10-28-offline-session-usage
id: SPC-0182
created: "2026-10-28 23:01:08"
hash: sogaz0l
---
# Offline Codex session usage

## Problem and evidence

acta has no command for measuring the model usage of its sessions. Raw input
totals combine fresh and cached tokens, and a root thread alone omits child-agent
usage. This makes raw token counts a poor basis for evaluating cost controls.

A read-only spike inspected 61 historical Codex logs for this repository from
2026-10-15 through 2026-10-26. They contained 22 root sessions and 2,318 unique
responses. Recorded input was 227,518,220 tokens, including 221,152,000 cached
tokens. Output was 684,695 tokens. Per-response sums matched final thread counters
in every inspected log. These are recorded usage counts, not billed amounts or
evidence of savings. The current research session was excluded from the spike.

The user approved Codex-only support, token counts without prices, and a stateless
CLI that reads existing transcripts. Existing hooks and workflow gates remain as
they are. The feature makes no model calls and requires no new dependency.

## Command and scope

Add `acta usage [--json]`. Run it inside a Git repository, including from a
subdirectory or linked worktree. No planning files or acta setup are required.
Reject positional arguments and unknown flags with the existing bad-input exit
code. Outside Git, report the missing repository and return that same code.

Read regular `.jsonl` files recursively beneath `$CODEX_HOME/sessions`, or
`~/.codex/sessions` when `CODEX_HOME` is unset or empty. Resolve a relative
`CODEX_HOME` against the command's working directory. Do not follow symlinked
files or directories while scanning. Do not read archived or compressed logs.

Use existing Git helpers to identify the current checkout and its registered,
existing worktrees. Compare canonical absolute paths. A session belongs to this
repository when its starting `session_meta.cwd` is inside one of those checkouts;
an existing nested Git repository belongs only to its own repository. Pass paths
as arguments to Git; never execute shell text obtained from a transcript.

Session attribution uses where a session started, not every directory it later
visited. Child usage belongs to its recorded root session, even when the child
started elsewhere. If root metadata is absent, include a child only when its own
starting directory establishes repository membership, and mark its group partial.
Do not guess membership from a folder name or a remote URL.

Historical sessions whose only checkout was deleted or moved may be impossible
to associate with this repository. State this limitation in command help and in
the report's scope description. The report covers discovered logs, not complete
repository history. Read all available history; this increment adds no date,
session, repository, pricing, or agent-selection flags.

## Records and counting

Support the `session_meta`, `turn_context`, and `token_usage_record` shapes
observed by the spike. Read files as streams; retain metadata and usage records,
not conversation text or tool output. Unknown event types do not fail a file.

### Supported event shapes

These synthetic records show the required nesting. Each object occupies one line
in the transcript. Extra fields are allowed and do not change these field paths.

```jsonl
{"timestamp":"2026-10-28T10:00:00Z","type":"session_meta","payload":{"id":"root-1","cwd":"/example/repo"}}
{"timestamp":"2026-10-28T10:00:01Z","type":"turn_context","payload":{"model":"example-model"}}
{"timestamp":"2026-10-28T10:00:02.125Z","type":"token_usage_record","payload":{"thread_id":"root-1","session_id":"root-1","response_id":"response-1","usage":{"input_tokens":100,"cached_input_tokens":60,"output_tokens":10,"reasoning_output_tokens":4},"turn_token_usage":{"input_tokens":100,"cached_input_tokens":60,"output_tokens":10},"thread_token_usage":{"input_tokens":100,"cached_input_tokens":60,"output_tokens":10}}}
```

Read thread identity and starting directory from `session_meta.payload.id` and
`session_meta.payload.cwd`. Read the current model from
`turn_context.payload.model`; a missing, empty, or non-string model resets it to
`unknown`. Read response identity, thread identity, root identity, and usage from
`token_usage_record.payload.response_id`, `.thread_id`, `.session_id`, and `.usage`.
The response timestamp is the event's top-level `timestamp`, not a payload field.

Accept RFC 3339 timestamps with an explicit timezone, with or without fractional
seconds, as parsed by Go's `time.RFC3339Nano`. Normalize output to UTC using that
same format. Missing or invalid response timestamps use the partial-data behavior
below. Metadata and context timestamps do not determine usage ranges.

A child file has its own `session_meta.payload.id`, such as `child-1`; its usage
records have `payload.thread_id: "child-1"` and `payload.session_id: "root-1"`.
The root file and child file are separate streams. An event-level `session_id`
outside `payload`, or a context `session_id`, is not a substitute for these paths.

### Counting rules

The first usable `session_meta` supplies the file's thread identity and starting
directory. Later metadata cannot change that identity.
Only usage records whose `thread_id` matches that identity belong to that file;
inherited records from another thread do not count again. A usage record's
`session_id` supplies the root-session identity. Do not infer root identity from
file names, timestamps, or matching directories. Thread, root-session and
response identities must be nonempty strings; otherwise exclude the affected
record and report incomplete coverage.

Use each record's `usage`, not its cumulative `turn_token_usage` or
`thread_token_usage`. Ignore the cumulative `event_msg.token_count` stream for
counting, since adding it would count the same requests twice. A file with no
supported usage records has unavailable usage, not a measured zero.

Count each `response_id` once across all discovered files. Identical duplicates
with the same owner, model, and usage are harmless. Conflicting copies are
excluded in full and reported; scan order must not decide which copy wins.

Associate each response with the model in its preceding `turn_context`. Keep
responses with no known model in an explicit `unknown` model group. Do not
attribute them to the machine's current default model or discard their tokens.

For each supported response, report:

- fresh input: `input_tokens - cached_input_tokens`;
- cached input: `cached_input_tokens`;
- output: `output_tokens`;
- response count: one.

These three token fields must be present, nonnegative integers, and fit signed
64-bit integers. Cached input must not exceed input.
Missing or invalid fields make that response unavailable and the report partial.
Reject arithmetic overflow instead of wrapping a total. Reasoning tokens are
already part of output and are never added again. Optional fields, including
cache-write and cumulative totals, do not add tokens to these categories.

## Output and coverage

Default output has one table grouped by root session and model, followed by a
per-model summary. Show root identity, the first and last included response
timestamps in UTC, fresh input, cached input, output, and response count. Order
session rows by first included timestamp, then root identity and model; order
model summaries by model name. Missing or invalid timestamps do not discard valid
usage. Mark coverage partial, omit those timestamps from the UTC range, use null
for an unavailable range endpoint in JSON, and sort undated rows last. Human
output labels an unavailable endpoint `unknown`. Escape control characters from
transcript strings before rendering them in a terminal.

JSON contains `schema_version: 1`, repository and checkout scope, source directory,
the included UTC range, session/model rows, per-model totals, and coverage. Output
stable field names and ordering of arrays. Token values are integers. Do not emit
prompts, commands, file contents, credentials, or full response identifiers.

### JSON version 1

The following example is the complete report for the synthetic response above
when its file also ends with one pending line. Its 4 reasoning tokens are already
within the 10 output tokens. The pending line makes the report partial and the
command exits with the skipped exit code.

```json
{
  "schema_version": 1,
  "status": "partial",
  "scope": {
    "repository": "/example/repo",
    "checkouts": ["/example/repo"],
    "source_directory": "/example/codex/sessions",
    "history_coverage": "discovered_logs_only",
    "limitations": ["Sessions whose only checkout was deleted or moved may be absent."]
  },
  "range": {
    "first_response_at": "2026-10-28T10:00:02.125Z",
    "last_response_at": "2026-10-28T10:00:02.125Z"
  },
  "sessions": [
    {
      "session_id": "root-1",
      "model": "example-model",
      "range": {
        "first_response_at": "2026-10-28T10:00:02.125Z",
        "last_response_at": "2026-10-28T10:00:02.125Z"
      },
      "fresh_input_tokens": 40,
      "cached_input_tokens": 60,
      "output_tokens": 10,
      "response_count": 1
    }
  ],
  "models": [
    {
      "model": "example-model",
      "fresh_input_tokens": 40,
      "cached_input_tokens": 60,
      "output_tokens": 10,
      "response_count": 1
    }
  ],
  "coverage": {
    "partial": true,
    "matched_root_sessions": 1,
    "included_thread_logs": 1,
    "counted_responses": 1,
    "excluded_invalid_records": 0,
    "conflicting_responses": 0,
    "logs_without_supported_usage": 0,
    "unclassified_logs": 0,
    "pending_lines": 1,
    "undated_responses": 0
  },
  "diagnostics": [
    {
      "code": "pending_line",
      "file": "example.jsonl",
      "line": 4,
      "message": "The final line is not terminated and was not read."
    }
  ]
}
```

All fields in this example are required. `sessions` has one row per root-session
and model pair; `models` has one summary per model. Both use the same token and
response-count field names. `range` has this same shape at report and row level;
each endpoint may be null. Sort `scope.checkouts` lexically. Other array orders
follow the rules above; sort diagnostics by file, line, then code, with nulls last.

`status` is `partial` whenever `coverage.partial` is true, including when no
responses could be counted. Otherwise it is `complete` when at least one
response was counted, or `no_data` when none were counted. A counted response
whose token values are all zero still makes a complete report, not no-data.

For a readable source with no matching sessions, the full no-data shape is:

```json
{
  "schema_version": 1,
  "status": "no_data",
  "scope": {
    "repository": "/example/repo",
    "checkouts": ["/example/repo"],
    "source_directory": "/example/codex/sessions",
    "history_coverage": "discovered_logs_only",
    "limitations": ["Sessions whose only checkout was deleted or moved may be absent."]
  },
  "range": {"first_response_at": null, "last_response_at": null},
  "sessions": [],
  "models": [],
  "coverage": {
    "partial": false,
    "matched_root_sessions": 0,
    "included_thread_logs": 0,
    "counted_responses": 0,
    "excluded_invalid_records": 0,
    "conflicting_responses": 0,
    "logs_without_supported_usage": 0,
    "unclassified_logs": 0,
    "pending_lines": 0,
    "undated_responses": 0
  },
  "diagnostics": []
}
```

The missing-source-directory result uses that same no-data shape and succeeds.
Human output explicitly says no data; empty arrays do not represent measured
zero repository usage. Fatal failures have no report on stdout and return the
operational-error code with a diagnostic on stderr, even under `--json`.

Each diagnostic contains `code`, `file`, `line`, and `message`. `file` is relative
to `source_directory`, or null when no file can be named. `line` is one-based,
or null for a file-level error. `message` is human-readable text, not a stable
machine identifier. Codes are `pending_line`, `invalid_json`, `unreadable_log`,
`missing_metadata`, `unsupported_usage`, `missing_identity`, `invalid_usage`,
`conflicting_response`, and `invalid_timestamp`. They describe the recoverable
conditions already listed in this spec. Unknown event types need no diagnostic.

### Coverage counters

Coverage records whether the report is partial, the number of matched root
sessions, included thread logs, unique counted responses, excluded invalid or
conflicting responses, matched logs without supported usage, and unclassified
logs whose unreadable or missing metadata prevents repository attribution.
Keep unrelated, successfully classified logs out of repository coverage counts.
An unclassified log cannot be called unrelated; count it separately and explain
that its effect on repository coverage is unknown.

`matched_root_sessions` counts distinct root identities established by matched
metadata or included child records, even if their usage is unavailable.
`included_thread_logs` counts distinct thread identities attributed to this
repository, so copied files do not increase it. `counted_responses` counts unique
valid responses after conflict exclusion. `excluded_invalid_records` counts
invalid owned usage records before response deduplication; a record without a
response identity is still a record. `conflicting_responses` counts distinct
response identities with conflicting copies, not the number of copies.
`logs_without_supported_usage` counts distinct matched thread identities whose
files contain no supported owned usage records. `unclassified_logs` counts
physical files without usable membership metadata. `pending_lines` counts skipped
unterminated final lines. `undated_responses` counts included responses with
missing or invalid timestamps. Identical copies of a counted response do not
increase the response or token totals.

No matching sessions produces empty rows and an explicit no-data result. It does
not claim that this repository used zero tokens. Diagnostics are structured in
JSON and go to stderr for human output. A valid `--json` response stays parseable.

## Errors and offline behavior

A missing sessions directory produces no data and succeeds. An unreadable source
directory, failed Git discovery, or overflow that prevents safe reporting returns
the existing operational-error exit code. Print no apparently complete totals on
those fatal paths.

Recoverable file or record errors retain valid records, report the affected
counts, mark the result partial, and return the existing skipped exit code.
Malformed complete lines, unreadable files, missing identity, unsupported usage,
and conflicting duplicates are recoverable. Unknown events alone are not errors.

Snapshot each file's size before reading and read no bytes beyond that snapshot.
Ignore an unterminated final line as a pending write, count it in coverage, and
mark the result partial. A later invocation can include it once complete. This
does not make reading several live files an atomic session snapshot; the report
describes the records observed during this invocation.

The command performs no network requests, starts no background update probe, and
writes no transcripts, configuration, planning files, database, or usage cache.
Its invocation must bypass the CLI's automatic update check. Local read-only Git
queries are allowed. Diagnostics name files and counts without reproducing their
conversation content.

## Implementation boundaries and verification

Keep transcript reading and aggregation in `internal/usage`, with command wiring
and rendering in `internal/cli`. Reuse repository/worktree discovery, JSON output,
exit codes, and help/man generation where they fit. Update the command help table.
Do not add a generic agent-adapter framework for this single reader.

Use synthetic fixtures in temporary directories, never copy real conversations
into the repository. Focused tests must establish that root and child usage is
counted once, models remain separate, fresh plus cached equals input, reasoning
is not added again, duplicates are independent of scan order, unrelated repos
are excluded, nested repos remain separate, and worktrees group with their parent.
Cover malformed and unfinished lines, absent counters, unknown models, unsafe
numbers, missing sources, and partial-result exit behavior. Verify that command
execution starts no update process and modifies no input files.

Before landing, compare the new reader with an independent offline sum on the
historical spike files when those files are still available. Do not pin a test
to the historical counts because local logs can change. The normal full landing
gates remain required.

## Deferred work

This increment does not estimate money, infer successful tasks, assign workflow
phases, measure savings, support other agents, add TUI screens, learn facts,
prune tools, or intervene in loops. Those require separate evidence and design.
