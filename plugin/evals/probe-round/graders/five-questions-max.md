---
# guards: five questions at most
type: regex
pattern: "(?:^\\*\\*Q6\\.|^ {0,3}6\\.\\s)"
flags: im
match: not_contains
target: last_message
---

Rejects a round that asks a sixth question, in the probe format **Q6. or in a
plain numbered list. It passes on a round of one to five questions, so it fails
only when the round is over the cap.
