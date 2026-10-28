---
parent: bugs/2026-10-28-tidy-branch-delete-asks-confirm
id: SPC-0181
created: "2026-10-28 08:24:12"
hash: yvof5ai
started: "2026-10-28 08:30:25"
finished: "2026-10-28 09:23:45"
---
# Exempt post-tidy branch delete from the destructive rule

## Scope

Bounded fix for the tidy branch delete confirmation bug, approved in chat
on 2026-10-28.

Every land with `commit_history: tidy` asks "Boleh dihapus?" before
deleting the branch. The worktree is removed, then the branch delete
stalls on a confirmation that should never come: land runs without
asking when clean.

## Design

Two rules conflict. Core rule 5 in `internal/hook/hook.go` says never
run a destructive command without a full-sentence warning and a yes.
The tidy cleanup in `plugin/skills/land/tidy.md` requires
`git branch -D`, which the agent reads as destructive. Tidy is the
default, so every land hits this.

Change rule 5 to name the exception:

```
5. Never push. Never run a destructive command without a full-sentence
warning and a yes. Only the `git branch -D` in tidy.md step 7, after
its final plan check passes, is not destructive: tidy proved the tip
equals the merge, so delete without asking. A stopped or red tidy land
still asks.
```

The exception covers only that one `-D`. Tidy.md step 5
`git reset --keep` and step 7 `git update-ref -d` stay under rule 5:
`reset --keep` mutates the parent, and neither is what the user
reported. If they prompt too, that is a separate bug.

Regenerate `plugin/hooks/default-rules.md` with
`go test ./internal/hook -run TestDefaultRulesFile -update`, since that
file must match the default session note.

Append to tidy.md step 7, after its proof clause (which stays):
"; per core rule 5 this delete needs no confirmation".

Bump both size caps in the same commit to the new measured sizes:
`sessionStartCap` and the `skills/land/tidy.md` entry in
`internal/plugincheck/budget_test.go`. Measure with the failing test
output, never guess.

## Verification

Add a hook test in the shape of `TestSessionStartRuleSixSplit`:
assert the new rule 5 exception text is in both `coreRules` and
`default-rules.md`. The old bare rule is a verbatim prefix of the new
one, so the negative check must match the exact old line including its
trailing newline, never a bare `strings.Contains` of the sentence:
+
old: "5. Never push. Never run a destructive command without a full-sentence warning and a yes.\n"
+
Run the focused hook and plugincheck tests, then `scripts/test`. No new
dependency, public API, or file format is needed.
