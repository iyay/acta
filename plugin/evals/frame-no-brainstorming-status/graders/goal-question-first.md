---
type: regex
pattern: "^[^?]*what[^?]*goal[^?]*\\?"
flags: i
target: last_message
---

Rejects a reply whose first question is not the goal question. The pattern
takes any wording, so it passes on a reply that files the item, says what it
did, then asks what's your goal with this. It fails on a reply that opens with
"who is this for?", "what problem does it solve?" or any other question mark
before the goal one, because no question mark may come first.
