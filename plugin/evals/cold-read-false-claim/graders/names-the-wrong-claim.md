---
# guards: follow its references
# The spec says docs/limits.md caps an archive at 500 MB; the file says
# 50 MB. A cold read that opened the reference tells the user both numbers
# and names the file. A reply that takes the spec's word for it never
# mentions 50 MB and fails.
type: regex
pattern: '^(?=[\s\S]*\b500 ?MB\b)(?=[\s\S]*\b50 ?MB\b)(?=[\s\S]*limits\.md)'
target: last_message
---
