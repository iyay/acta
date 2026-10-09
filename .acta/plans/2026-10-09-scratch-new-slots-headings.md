---
parent: bugs/2026-10-09-scratch-new-duplicates-sections
depth: minimal
id: PLN-0130
created: "2026-10-09 22:04:43"
hash: kabnr4w
---
# acta scratch new puts a headed body into its sections

**Goal:** `acta scratch new` sorts a body that already has its own headings into the scratch sections, and the schema check reports a section heading that appears twice.

**Spec:** .acta/specs/2026-10-09-scratch-new-slots-headings.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- A body with no `# ` or `## ` heading lines must give a file byte-for-byte the same as before this plan.
- No words of the body are ever lost: every non-heading line of the input ends up in the file.
- Tests use temp repos only. acta write commands auto-commit, so never run them in this repo.
- Each code commit ends with the trailer line `Task: PLN-<hash>#<n>`, using this plan's hash.

## Waves

- Wave 1: Task 01, Task 02
- Wave 2: Task 03

### Task 01: Schema check reports a section that appears twice

**Files:**
- Modify: `internal/board/schema.go`
- Test: `internal/board/schema_test.go`

**verify:** `CheckBody` on a scratch body returns `## <name> appears twice` for every schema section heading that shows up more than once, whatever the order, and keeps every result it gives today for bodies with no repeated heading. List each body shape checked and its result.

- [ ] Failing test: in `schema_test.go`, a scratch body with `## Context` twice (in valid order otherwise) returns a problem `## Context appears twice`; a body with `## Words` twice returns `## Words appears twice`; a valid body still returns nil. Run `scripts/test ./internal/board/ -run CheckBody` and see it fail.
- [ ] Code: in `CheckBody`, when a known section name is already in `seen`, append `"## "+name+" appears twice"`. Plain-English comment saying why.
- [ ] Run `scripts/test ./internal/board/` passes, vet and gofmt clean, commit `fix(board): report a scratch section heading that appears twice`.

### Task 02: scratch new sorts a headed body into its sections

**Files:**
- Modify: `internal/write/scratch.go`
- Test: `internal/write/scratch_test.go`

**verify:** For any body sent to `NewScratch`, the written file has each of `## Words`, `## Context`, `## Log`, `## Open questions` exactly once and passes `board.CheckBody`; a top `# ` title line is dropped; text before the first `## ` and under `## Words` lands in Words below the dated heading; text under Context, Log and Open questions lands in that section, with Log under a dated heading; any other `## ` heading becomes `### ` inside Words; a body with no heading lines gives exactly today's file. List each body shape checked and where each part landed.

- [ ] Failing test: in `scratch_test.go`, `NewScratch` with body `"# T\n\n## Words\n\nhello\n\n## Context\n\n- c\n\n## Log\n\nl\n\n## Open questions\n\n- q\n\n## Extra\n\nx\n"` writes each schema heading once, `hello` and `### Extra` with `x` under Words, `- c` under Context, `l` under a `### <day>` heading in Log, `- q` under Open questions, and `CheckBody` returns nil; then `AppendScratch` with section `log` succeeds. A body `"just words\n"` gives the same bytes as before (assert the full expected file text). Run `scripts/test ./internal/write/ -run NewScratch` and see it fail.
- [ ] Code: in `NewScratch`, split the body into parts by `## ` lines (drop a leading `# ` line), build today's empty skeleton, then fill each part with the existing `putInSection` using the `sections` map (Words and Log dated). Turn an unknown `## X` into `### X` and keep it with the Words text. A body with no headings keeps today's exact path. Plain-English comments saying why.
- [ ] Run `scripts/test ./internal/write/` passes, vet and gofmt clean, commit `fix(write): scratch new puts a headed body into its sections`.

### Task 03: Clean the one old scratch file with doubled sections

**Files:**
- Modify: `.acta/scratch/2026-09-30-faster-agent-test-runs.md`

**verify:** After the change, no scratch file in `.acta/scratch/` fails `acta doctor`'s schema check, and the cleaned file keeps every non-heading line it had before. List the lines compared and the doctor result.

- [ ] Check first: with the branch binary (`go run ./cmd/acta doctor`), see the schema warning name `2026-09-30-faster-agent-test-runs.md`.
- [ ] Edit: merge the two `## Words` sections into one and the two `## Context` sections into one, keeping every line and its order; touch nothing else.
- [ ] Run `go run ./cmd/acta doctor` shows `ok schema`, then commit with `go run ./cmd/acta commit .acta/scratch/2026-09-30-faster-agent-test-runs.md -m "chore(scratch): merge doubled sections in faster-agent-test-runs"`.
