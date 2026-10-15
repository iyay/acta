---
id: SPC-0138
created: "2026-10-15 13:26:48"
hash: kf0xpb0
started: "2026-10-15 13:28:06"
finished: "2026-10-15 13:30:12"
---
# TUI filter mode look

## Problem

While the reader types a filter, nothing on screen shows where the typing
goes: the status bar has no cursor, and the box being filtered still looks
active with its full selection color. After enter, the status bar keeps the
same text it showed while typing, so the reader cannot tell typing from a
kept filter.

## Design

Typing a filter (`searching` is true):

- The status bar shows `Filter: <query>` followed by a block cursor (one cell
  in reverse video), so the reader sees where the keys go.
- The focused list box draws as not active: its border takes the style of a
  box without the focus.
- The selected row in that box takes a new `selectedDim` style: the same text,
  a calmer background than `selected`. Its colors keep a contrast ratio of 1.6
  or more (see the wiki page ghostty-minimum-contrast).

After enter (`searching` is false, the query is kept):

- The box is active again and the selected row takes the normal `selected`
  style.
- The status bar shows `Filtered by "<query>" · esc clears`, in place of the
  typing line. No cursor.

Esc while typing or after enter clears the query and returns everything to
normal, as today.

## Testing

- With `searching` true: the status line ends with the cursor cell; the
  focused box border and selected row use the inactive and dim styles.
- After enter: the box and row use the active and `selected` styles, and the
  status line holds `Filtered by "abc" · esc clears` with no cursor.

## Out of scope

Changing which keys start or end a filter.
