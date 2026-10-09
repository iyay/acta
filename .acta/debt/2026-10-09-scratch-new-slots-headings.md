---
id: DBT-0106
hash: po5f7q3
parent: plans/2026-10-09-scratch-new-slots-headings
---
# Review NOTEs: acta scratch new puts a headed body into its sections

- [ ] (low) internal/write/scratch.go splitHeaded and board.CheckBody read a `## ` line inside a ``` code fence as a heading, so a scratch body with a fenced sample holding `## Foo` gets that line rewritten to `### Foo` and blank lines added inside the fence. No text is lost, but the sample changes.
