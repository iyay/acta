---
created: "2026-10-01 04:43:52"
id: SPC-0053
hash: g0ov0fp
started: "2026-10-01 04:46:48"
finished: "2026-10-01 04:48:54"
---
# Debt note in the detail looks like a spec or plan body

Status: design approved by the user in chat on 2026-10-01. Bounded: the debt item branch of `buildDetailParts` (`internal/tui/detail.go`) already draws the note (PLN-0061).

## Why

Since PLN-0061 the debt item detail shows its note as plain wrapped text that runs flush against the pane walls. A spec or plan detail draws its body through the markdown renderer (`m.render`), which leaves one blank line above, two columns on the left and right, and colors `` `code` `` spans. The user wants the note to look the same, gaps included.

The renderer drops text that looks like an HTML tag: `/tmp/pmb-<uid>:` comes out as `/tmp/pmb-:`. Real notes hold text like that, so the note cannot go through the renderer as it is.

## Design

1. **Same renderer as the body.** The debt item branch draws the note with `m.render(text, w)`, each line through `fit`, the same way the body of other kinds is drawn. The `xansi.Wordwrap` step goes away, since the renderer wraps. The renderer can still leave a line wider than the pane (a word longer than the pane, or a `-` after a space), so its output goes through `xansi.Hardwrap(out, w, true)`. Lines that already fit are left as they are.
2. **Escape before rendering.** Before the note goes to the renderer, `&`, `<` and `>` become `&amp;`, `&lt;` and `&gt;`, so every character the note holds is drawn as written.
3. Nothing else changes: the header, the footer, other kinds and the debt file detail stay as they are.

## Out of scope

- Other markdown in a note. A note that starts with `- ` or `#` renders as a list or a heading, and `*word*` or backticks turn into styling. Notes are one line, so this is rare and stays that way.

## Testing

A failing test comes first:
- With the real renderer, for widths 10 to 160, every non-space character of a note is drawn, except the backticks and `*` that markdown turns into styling. The notes tested include `<uid>`, `&`, a ` -` after a space, a `` `code` `` span, and a word longer than the pane.
- No drawn line is wider than the pane.
- The note starts in the same column as a spec body, and has the same blank line above it.
