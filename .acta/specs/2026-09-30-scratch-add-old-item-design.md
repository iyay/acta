---
parent: bugs/2026-09-30-scratch-add-words-refused-on-old-item
created: "2026-09-30"
id: SPC-0036
hash: kuloco4
---
# Scratch add takes any section on an old item

Status: design approved by the user in chat on 2026-09-30. Bounded (one existing function, `AppendScratch` in `internal/write/scratch.go`, plus its test).

## Why

`acta scratch add <id> --section <name>` refuses an item with no body schema: "bad input: SCR-0001 is an old item with no sections" (`internal/write/scratch.go:105-107`). The scratch skill calls `--section words` the default, and the brainstorm skill appends answers with `--section log`, so an agent that follows the skills loses the user's words on every old item. This repo holds 24 old scratch items. The eval case `answers-appended` is red on main for this reason.

## Design

1. On an item with no schema, every known section (`words`, `context`, `log`, `questions`) does the plain append at the end of the file, the same as a call with no `--section`. The refusal is removed.
2. An unknown section name is still refused, for old and new items alike. That check already runs first (`scratch.go:93`), before the old-item branch.
3. Items with a schema behave as today.
4. The comment above `AppendScratch` says that an old item has no parts, so any section lands as the plain append.
5. The scratch skill text and the CLI usage line stay as they are.

## Testing

`TestAppendScratchOldItem` in `internal/write/scratch_test.go`: each known section on an old item leaves the file exactly as a no-flag append would, with no heading added; an unknown section on an old item is still refused and leaves the file unchanged. The old assertion that `--section context` is refused goes. The test goes red before the code changes. After the build, `scripts/eval --case answers-appended` passes.

## Out of scope

Moving old items to the schema. Any change to items that already have a schema.
