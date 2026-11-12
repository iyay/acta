---
# guards: ## Review depth
# An inline review must still show what stands behind its verdict, so the
# last reply holds a Coverage heading and the verdict CLEAN.
type: regex
pattern: '^(?=[\s\S]*### Coverage)[\s\S]*\bCLEAN\b'
target: last_message
---
