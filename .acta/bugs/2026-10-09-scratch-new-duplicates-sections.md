---
id: BUG-0042
hash: y2ys0me
---
# A new scratch item gets its sections twice and then refuses appends

## Symptom
When the body sent to `acta scratch new` already has its own `# title` and `## Context` / `## Log` / `## Open questions` headings, the file ends up with each section twice. After that, `acta scratch add <id> --section log` fails with `bad input: scratch <file>: ## Context is out of order`, and `acta doctor` warns on schema.

## Root cause
`internal/write/scratch.go:52-53`: `NewScratch` always wraps the body as `# <title>`, `## Words`, `### <day>`, the body, then empty `## Context`, `## Log`, `## Open questions`. It never checks whether the body already holds those headings. The schema order check in `internal/board/schema.go:82-84` then sees Context after Open questions and marks the file bad, so `AppendScratch` refuses to write.

## Repro
In a temp git repo with `.acta/scratch/`:
`printf '# U\n\n## Words\n\nhello\n\n## Context\n\n- c\n\n## Log\n\n## Open questions\n\n- q\n' | acta scratch new u --title U`
then `echo x | acta scratch add SCR-0001 --section log` prints `bad input: scratch ...-u.md: ## Context is out of order`.

## Found in
main, while shaping SCR-0059 (2026-10-09): its body was sent with full headings, so the log append failed. Reproduced in a temp repo through acta:debug.
