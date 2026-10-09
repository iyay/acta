---
parent: bugs/2026-10-09-scratch-new-duplicates-sections
id: SPC-0121
created: "2026-10-09 22:03:24"
hash: tiosvo4
---
# acta scratch new puts a headed body into its sections

Status: Bounded, approved by the user in chat on 2026-10-09. Fixes BUG-0042.

## Problem

`NewScratch` (`internal/write/scratch.go:52-53`) always wraps the body in `# <title>`, `## Words`, `### <day>`, then empty `## Context`, `## Log` and `## Open questions`. A body that already has those headings ends up with every section twice. The schema check then calls the file out of order, and `acta scratch add` refuses to write to it.

## Design

1. `NewScratch` sorts a headed body into the slots:
   - A `# ` title line at the top of the body is dropped. The `--title` flag or the slug still names the item.
   - Text before the first `## ` heading, and the text under `## Words`, goes under `## Words`, below the `### <day>` heading.
   - Text under `## Context`, `## Log` and `## Open questions` goes into that section. Log text gets a `### <day>` heading, the same as `acta scratch add --section log`.
   - A `## ` heading that is not one of these four becomes `### ` inside Words, so no words are lost and the section order stays valid.
   - A body with no `#` or `## ` headings gives exactly the same file as today.
2. `CheckBody` (`internal/board/schema.go`) reports `## <name> appears twice` when a section heading shows up more than once. The schema applies to scratch only, so specs and plans are not touched.
3. `.acta/scratch/2026-09-30-faster-agent-test-runs.md` has `## Words` and `## Context` twice. Clean it by hand, keeping every line, and commit it with `acta commit`. Without that, the new check marks it bad and `acta scratch add` refuses it.

## Tests

- `NewScratch` with a headed body: every section heading appears once, each part lands in its slot, and `AppendScratch` with `--section log` then succeeds.
- `NewScratch` with a plain body: the file is byte-for-byte the same as before the change.
- `CheckBody` on a scratch body with `## Context` twice returns `## Context appears twice`.
- Gate: `scripts/test --full`, `go vet ./...`, `gofmt -l .` empty.

## Out of scope

- A repair command for old files.
- Changing the scratch skill text.
