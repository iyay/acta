---
type: regex
pattern: "\\*\\*Q\\d+\\.[^\\n*]*\\*\\*(?:(?!Recommended:)[\\s\\S])*?(?=\\*\\*Q\\d+\\.|$)"
match: not_contains
target: last_message
---

Rejects a round where any question carries no recommended answer. The pattern
opens a **Qn. heading and runs to the next heading or the end of the reply, so
it only fires when a question block never says Recommended: before the block
ends. It passes on a round where every question ends with a recommendation.
