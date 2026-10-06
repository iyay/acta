---
parent: debt/2026-10-03-dispatch-send
closes: [DBT-0071.01, DBT-0071.02, DBT-0071.03, DBT-0078.01]
id: SPC-0094
created: "2026-10-06 09:24:49"
hash: bbe08fm
---
Status: Bounded, approved by the user in chat on 2026-10-06.
Why: `acta dispatch send` reads omp's todo card once and says ok, unconfirmed or drift. Today it says drift when omp folds a long list to 8 rows, when an old round's card is still on screen, or when a line like "Task 1 of 3" sits near the card. It never sees the card at all with the ascii symbol preset, because omp frames that header. A false drift tells the orchestrator to interrupt an agent that is working correctly.

# dispatch checkpoint: a verdict that matches the card

## Changes

All in `checkpoint` and `todoCardHeader`, `internal/cli/dispatch_herdr.go`.

1. **Only the newest round (DBT-0071.03).** `checkpoint` keeps only the pane text after the last line that holds `goalMark`. No goal mark in the text: `unconfirmed`.
2. **Framed header (DBT-0071.02).** `todoCardHeader` matches both the unframed header and the framed ascii one, `+--- [x] Todo 3 tasks ---+`, starting from the regex the debt item gives: `^[^\p{L}\p{N}\n]*(?:\[x\][^\p{L}\p{N}\n]*)?Todo(?:\s|$)`. When the header holds `<N> tasks`, N is read.
3. **Folded list (DBT-0071.01).** With N read from the header:
   - N less than the plan's task count: `drift` (omp listed fewer tasks than the plan has);
   - N at least the task count and every id seen: `ok`;
   - N at least the task count and some ids not seen: `unconfirmed` (they may be folded), never `drift`.
   With no N in the header, today's rule stays: some ids seen and some missing is `drift`.
4. **Generic task text (DBT-0078.01).** Ids are looked for only on lines after the card header, and a `task <n> of <m>` phrase never counts as id n.

## Out of scope

DBT-0071.04 (old goal mark in deliverGoal), .05 (brief file not git-ignored) and .06 (tab reuse by name) stay open as debt.

## Testing

Failing test first for each case in `internal/cli/dispatch_herdr_test.go`, with the fake herdr those tests already use: a 12-task plan whose card says `Todo 12 tasks` and shows 8 rows (unconfirmed, not drift); a card that says `Todo 3 tasks` for a 5-task plan (drift); the framed ascii header (card found); an old round's card above the newest goal mark (only the new part read); a "Task 1 of 3" line (not id 1). Gate: `scripts/test --full`.

## Close

Last task bumps the patch version to 0.1.26 in `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json` and `plugin/package.json`.
