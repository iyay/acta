---
type: regex
pattern: '^\W*(sure|certainly|of course|great question|absolutely|happy to|i.d be happy)\b|hope (this|that) helps|let me know if|feel free to'
flags: i
match: not_contains
target: last_message
---
