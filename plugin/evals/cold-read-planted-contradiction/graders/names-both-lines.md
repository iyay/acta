---
# guards: contradictions
# Require conflict recognition and both durations, in either order or form.
type: regex
pattern: '^(?![\s\S]*\b(?:no|not a) (?:conflict|contradiction)\b)(?=[\s\S]*\b30[ -]days?\b)(?=[\s\S]*\b7[ -]days?\b)(?=[\s\S]*\b(?:contradict\w*|conflict\w*|inconsisten\w*|incompatib\w*|cannot both|at odds|two retention periods|breaks? (?:the )?(?:30[ -]day )?promise|both (?:cannot|can\x27t) hold)\b)'
flags: i
target: last_message
---
