---
ref: SPC-0161
id: BUG-0049
hash: m097j1d
fixed_in: a999a1e71aa1f64b056e2df055a099c413cba835
finished: "2026-10-19 19:22:00"
---
# Unslop rule 28 no longer covers dense sentences

## Symptom

The shared Unslop reference preserves rule number 28 but changes its meaning.
The supplied rule requires shortening or splitting dense sentences, with one
idea per sentence. The shipped rule instead discusses sentences that could
fit any project, which was part of supplied rule 27. Models following this
reference therefore lose the explicit dense-sentence instruction.

## Root cause

plugin/references/unslop.md:59 replaces the supplied rule 28 with generic-copy
guidance. The static checks verify that numbered rules exist, but do not
establish that each number retains its supplied meaning.

## Repro

1. Read rule 28 in plugin/references/unslop.md at commit b6ac311.
2. Compare it with the user's supplied rule 28: "Shorten or split dense sentences. If the reader has to backtrack to parse a sentence, break it in two or drop clauses. One idea per sentence."
3. Observe that the shipped rule 28 says "Sentences that fit any project" and contains no instruction to shorten or split dense sentences.
4. Run scripts/test ./internal/plugincheck ./plugin; both packages pass despite this mismatch.

## Found in

main at b6ac311, while testing the landed change at the user's request.
