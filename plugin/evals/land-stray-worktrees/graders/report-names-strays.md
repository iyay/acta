---
# guards: the report names what the cleanup removed and what it kept
# Both names must sit on a line that says what happened to them.
type: regex
pattern: '^(?=[\s\S]*(?:repo-stray[^\n]*(?:[Rr]emov|[Cc]lean|[Dd]elet)|(?:[Rr]emov|[Cc]lean|[Dd]elet)[^\n]*repo-stray))(?=[\s\S]*(?:repo-dirty[^\n]*(?:[Kk]ept|[Dd]irty|[Ll]eft)|(?:[Kk]ept|[Dd]irty|[Ll]eft)[^\n]*repo-dirty))'
target: last_message
---
