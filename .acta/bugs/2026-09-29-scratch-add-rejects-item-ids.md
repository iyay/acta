---
id: BUG-0008
hash: lxqcblj
status: wontfix
finished: "2026-09-29"
---
# acta scratch add does not accept an item id

## Symptom
`acta scratch add SCR-0017 < text.md` fails with `bad input: unknown id SCR-0017`. The same happens with `SCR-17`, `SCRATCH-17` and the hash `s4h6zmf`. Only the path form `scratch/2026-09-29-popup-dim-main-text` works. The acta:scratch skill shows the id form and a `--section` flag, and the flag does not exist either (`flag provided but not defined: -section`).

## Root cause
Not debugged yet. Seen from running the PATH binary at /Users/iyay/.local/bin/acta.

## Repro
1. `echo hi | acta scratch add SCR-0017` gives `unknown id`.
2. `echo hi | acta scratch add SCR-0017 --section words` gives the flag error.
3. `echo hi | acta scratch add scratch/2026-09-29-popup-dim-main-text` works.

## Found in
main at 01cf9a1, while adding user notes to SCR-0017. The PATH binary may be older than main (see the note about rebuilding after merge), so check that first.
