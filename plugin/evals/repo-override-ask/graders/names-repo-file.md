---
# guards: is committed and shared
# Rejects a reply that never says the repo file is the reason. A run that
# changes nothing and says nothing, or that fails quietly, has no mention of it.
type: regex
pattern: "\\.acta\\.yaml"
target: last_message
---
