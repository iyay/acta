---
type: regex
pattern: "\\*\\*Q1\\."
match: contains
target: last_message
---

Rejects a reply that skips the probe format and never opens a numbered
question. It fails when the round has no **Q1. heading anywhere.
