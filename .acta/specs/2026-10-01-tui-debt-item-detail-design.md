---
created: "2026-10-01 00:10:02"
id: SPC-0052
hash: vek4kv8
---
# Debt item detail shows the full note, not a copy of the list

Status: design approved by the user in chat on 2026-10-01. Bounded: the detail box (`internal/tui/detail.go`) already draws debt items.

## Why

The detail of a debt item adds almost nothing to the Open pane next to it:

1. The note itself sits in the header line `DEBT :`, and every header line is cut to the pane width (`buildDetailParts`, `detail.go:134`). A long note is cut off, and the note is the one thing the reader came for.
2. The middle lists every line of the debt file (`debtLines`, `detail.go:296`). That is the same list the Open and Done panes already show.
3. The body is only the file heading `# Review NOTEs: <plan title>`, which says again what the `FROM` line says.

## Design

This changes a debt item (`board.KindDebtItem`) only. A debt file (`board.KindDebt`) keeps its list of lines.

1. **Header.** ID, STATUS, AUTHOR, FROM, FILE, the same as now, but with no `DEBT :` line.
2. **Middle.** The full note text, wrapped to the pane width so no word is lost. Problems (`! ...` lines) still come first, the same as for other kinds.
3. **No sibling list and no file heading.** The middle shows the note and nothing else.
4. **Footer.** The date line stays as it is.

## Out of scope

- A place marker like "note 2 of 4". The FILE line already says where the note lives.
- Any change to the debt file detail, the list panes, or other kinds.

## Testing

A failing test comes first: a debt item whose note is longer than the pane shows every word of it in the detail; the detail has no `DEBT` header line, no sibling note lines and no `Review NOTEs` heading; a debt file's detail still lists its lines.
