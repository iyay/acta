---
id: DBT-0055
hash: hygcwjj
parent: plans/2026-10-01-tui-debt-note-markdown
---
# Review NOTEs: Debt Note Rendered Like a Spec Body Implementation Plan

- [ ] A backslash inside a code span in a debt note is drawn twice (run `a\b` shows a\\b), because markdown keeps backslashes inside backticks; escape only outside code spans when a real note needs it.
- [ ] Hardwrap splits a glamour styled run and leaves its color open at the line end, so the right pane wall on that row picks up glamour colors; 7 of 515 real notes at width 40, none at 60 or more.
- [ ] A line broken by Hardwrap starts at column 0, so it loses the 2-column gap and the code span style.
- [ ] A single * in a note can pair with another and turn text between them italic ('*acta/locks*' shows as acta/locks), which the spec leaves out of scope.
- [ ] A note that starts with 1. renders with two blank lines above it, and __init__ renders bold; both are markdown the spec leaves out of scope.
- [ ] Glamour render of a debt note costs about 130-200 microseconds cold at width 100 against about 50 before; warm and scroll paths are cached and unchanged.
- [ ] The newRenderer cache never shrinks and now also holds one entry per debt note per width.
- [ ] TestDetailDebtItemNoteHasTheSpecBodyGaps checks widths 40, 80 and 120 only.
- [ ] TestDetailDebtItemNoteLosesNothingAtAnyWidth strips backticks and * from the wanted text, so it cannot catch a doubled backslash inside a code span or a note with a single unpaired *.
- [ ] Spec and plan bodies on main have the same wrap loss: glamour can leave a line wider than the pane (long word, or a - after a space) and fit cuts the rest.
