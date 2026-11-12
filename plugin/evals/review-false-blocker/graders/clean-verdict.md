---
# guards: Verify each finding against the code before acting
# With its one BLOCKER rejected the round is CLEAN. The prompt asks for the
# verdict as a line of its own, so quoting the other report's CLEAN does not
# count. The reply must also say the BLOCKER was rejected, so a run that
# dropped it without a word does not pass.
type: regex
pattern: '^(?=[\s\S]*\b[Rr]eject)[\s\S]*Round verdict: CLEAN'
target: last_message
---
