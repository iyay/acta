---
parent: bugs/2026-10-19-unslop-rule-28-loses-dense-sentence-guidance
id: SPC-0162
created: "2026-10-19 19:14:40"
hash: dezmmzx
started: "2026-10-19 19:17:07"
finished: "2026-10-19 19:17:58"
---
# Unslop rule 28 keeps its dense-sentence meaning

## Problem

BUG-0049. Rule 28 in `plugin/references/unslop.md` says "Sentences that fit
any project". The user's rule 28 asks to shorten or split dense sentences.
That text is gone, so agents lose it. The "fits any project" check came from
rule 27.

## Design

Bounded change to one reference file and one test.

- Rule 28 becomes the user's text, word for word: "Shorten or split dense
  sentences. If the reader has to backtrack to parse a sentence, break it in
  two or drop clauses. One idea per sentence."
- Rule 27 gains one last sentence: "If the sentence could appear unchanged in
  another project's docs, make it specific or cut it."
- Rule numbers 29 to 33 stay as they are.
- A test in `internal/plugincheck` checks that rule 28 says "dense
  sentences" and rule 27 says "another project". A shifted rule then fails.

## Files

- `plugin/references/unslop.md`
- `internal/plugincheck` (the test file that already checks unslop.md)

## Testing

- The new test fails on main, then passes.
- `scripts/test ./internal/plugincheck ./plugin` passes.
