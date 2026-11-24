---
id: BUG-0009
hash: e330k3j
fixed_in: 5af8532
finished: "2026-10-01 16:14:08"
---
# acta set cannot change an item title

## Symptom
`acta set scratch/<stem> title "..."` fails with `bad input: unknown field "title"; use status, type, fixed_in or ref`. The only way to rename an item is to edit the frontmatter by hand.

## Repro
`acta set scratch/2026-09-29-popup-dim-main-text title "x"`

## Found in
main at b0cb7b4, while retitling SCR-0017 at the user's request.
