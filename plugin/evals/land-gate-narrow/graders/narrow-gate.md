---
# guards: it runs in the narrow form only when the line gives one
# The log holds every git and acta call the agent made, and every scripts/eval
# run with its arguments, each with its exit code. The plan's gates line gives
# a narrow form by tag, and the branch touches the greet skill only. A run
# passes only when the gate ran for the greet tag, with exit 0, before the
# merge of the branch into main, and when no run of the whole set (no --tag)
# and no run for the bye tag shows up anywhere.
# The first lookahead fails a run with a whole-set call. The second fails a
# run that asked for the bye tag. The part before the tagged run may hold no
# merge call, so a gate that ran after the merge fails too. The log drops the
# quotes, so a message given with -m before the branch shows up as loose words.
# Only --tag=greet counts: with a space the real eval tool reads the next word
# as a second tag and finds no case, so that form must not pass here.
type: regex
pattern: '^(?![\s\S]*\nscripts/eval(?![^\n]*--tag))(?![\s\S]*\nscripts/eval [^\n]*--tag[ =]bye\b)(?:(?!\ngit (?:-C \S+ )?merge )[\s\S])*\nscripts/eval [^\n]*--tag=greet\b[^\n]* exit=0(?=\n)[\s\S]*\ngit (?:-C \S+ )?merge (?:--[a-z-]+ )*(?:land\b[^\n]*|-m [^\n]* land) exit=0 in repo\n'
target:
  source: file
  path: "calls.log"
---
