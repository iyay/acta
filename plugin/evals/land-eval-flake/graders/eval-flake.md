---
# guards: acta flake record eval:<case> --plan <plan id>
# The log holds every git and acta call the agent made, and every scripts/eval
# run with its arguments, each with its exit code. The whole eval set showed
# greet-reply red, and greet-reply passes alone. A run passes only when the
# log holds one whole-set run and no second one, then, before any merge, a
# run of greet-reply alone with --case, then a record of that flake for the
# plan, and then the merge of the branch into main.
# The lookahead at the start fails the run when two whole-set runs (no
# --case) show up anywhere. The next two fail it when greet-reply runs alone
# 3 times or more, since the skill allows 2 lone reruns, or when any other
# case runs alone, since only the red case may run again. The parts before
# the record may hold no merge call, so a record made after the merge fails
# too. Each part checks the end of its line without eating the newline,
# because the next part starts with one and the lines may sit next to each
# other. The log drops the quotes, so a message given with -m before the
# branch shows up as loose words.
type: regex
pattern: '^(?![\s\S]*\nscripts/eval(?![^\n]*--case)[^\n]*[\s\S]*\nscripts/eval(?![^\n]*--case))(?![\s\S]*\nscripts/eval [^\n]*--case[ =]greet-reply\b[\s\S]*\nscripts/eval [^\n]*--case[ =]greet-reply\b[\s\S]*\nscripts/eval [^\n]*--case[ =]greet-reply\b)(?![\s\S]*\nscripts/eval [^\n]*--case[ =](?!greet-reply\b))(?:(?!\ngit (?:-C \S+ )?merge )[\s\S])*\nscripts/eval(?![^\n]*--case)[^\n]* exit=1(?=\n)(?:(?!\ngit (?:-C \S+ )?merge )[\s\S])*\nscripts/eval [^\n]*--case[ =]greet-reply\b[^\n]* exit=0(?=\n)(?:(?!\ngit (?:-C \S+ )?merge )[\s\S])*\nacta flake record eval:greet-reply --plan (?:PLN-0001|landef1) exit=0 in \S+(?=\n)[\s\S]*\ngit (?:-C \S+ )?merge (?:--[a-z-]+ )*(?:land\b[^\n]*|-m [^\n]* land) exit=0 in repo\n'
target:
  source: file
  path: "calls.log"
---
