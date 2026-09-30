---
id: BUG-0014
hash: lpjj005
---
# A frame tick scrolls the pane under an open help

## Symptom

The reader turns the wheel over a pane, so one notch scrolls at once and a
second notch waits for the frame tick. The reader presses `?` before the tick
arrives, so the help sits over the panes. The tick then scrolls the pane that
the help covers, and the pane moves under a popup nobody is reading.

To see it: focus the detail box of a plan with a long body, turn the wheel
twice, press `?`, then wait one frame. The detail box has moved under the help.

The same happens with a popup and with a slug. Only the search is part of the
screen mark the tick checks.

## Cause

`internal/tui/model.go` refuses a notch while the help, a popup, a slug or a
search is open, so no delta can start there. The notches that came before the
help opened are still pending, and the tick only drops them when the item, the
tab, the sub-tab or the search changed. The help, the popup and the slug are not
part of that mark, so the tick scrolls the pane the reader cannot see.

## Where it was seen

Found while building PLN-0045 task 2, which changed the wheel timing. The old
rule and the new rule both scroll the pane: the `wheelTickMsg` case in
`internal/tui/model.go` applied `wheelDelta` whatever the help was doing. PLN-0045
leaves the screen mark rule as it is on purpose, so that ticket did not fix it.

## Fix sketch

Add the help, the popup and the slug to the screen mark, or clear the pending
delta when one of them opens, so a tick never moves a pane under a popup.
