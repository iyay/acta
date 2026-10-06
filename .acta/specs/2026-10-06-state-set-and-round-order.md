---
parent: debt/2026-10-06-setup-block-refresh
closes: [DBT-0082.01, DBT-0077.01, DBT-0077.02]
id: SPC-0096
created: "2026-10-06 11:57:55"
hash: tsp3r5x
---
Status: Bounded, approved by the user in chat on 2026-10-06.
Why: `acta state set` with nothing on stdin empties that State part and commits, so an agent that forgets the pipe wipes its own Next. `acta state` picks the newest review round by passing commit hashes to `gitc.FirstSeen`, which wants a file path, so the order check never works and an older round can show as the current one.

# acta state: no silent clear, and the right newest round

## Changes

1. **No clear without asking (DBT-0082.01).** `acta state set <plan> <part>` refuses when stdin is empty or holds only blank lines, unless `--clear` is given. The refusal says `nothing on stdin; pipe the lines in, or use --clear to empty <part>`, writes nothing and commits nothing. `--clear` empties the part, as an empty stdin does today. `--clear` together with non-blank stdin is refused too. The usage line in `internal/cli/state.go` names `[--clear]`. The rule lives in `internal/write/state.go` so every caller gets it. `runningRule` in `internal/hook/hook.go` says "run acta state set <plan> next with the lines on stdin".
2. **Newest round by git order (DBT-0077.01, DBT-0077.02).** `newestRound` in `internal/board/state.go` stops using `gitc.FirstSeen`. A new `gitc.IsAncestor(repo, a, b string) (bool, error)` runs `git merge-base --is-ancestor a b`; when the first found round commit is an ancestor of the second, the second is newer. Any git error keeps today's fallback (the first found). The doc comment says this. `TestRunningRoundPrefersNewSubjectWithoutOrder` is rewritten to expect the commit that is really newer.

## Out of scope

DBT-0077.03 (TUI status line text) and DBT-0082.02 (test that CLAUDE.md's acta block matches the setup template) stay open as debt.

## Testing

Failing test first for each: empty stdin and blank-only stdin are refused with the file unchanged and no commit; `--clear` empties the part; `--clear` with text is refused; a branch where the old-subject round commit is newer than the new-subject one shows the old-subject round. Gate: `scripts/test --full`.

## Close

Last task bumps the patch version to 0.1.28 in `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json` and `plugin/package.json`.
