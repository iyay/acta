---
id: BUG-0043
hash: uus6jd5
fixed_in: 96fbdbe8b7c85a99c52f2acb005d98d40c4fdda3
finished: "2026-10-12 09:03:06"
---
# acta tidy names the folded commit after the last chore commit

## Symptom
When no commit of a branch is kept as its own, `acta tidy` folds the whole range into one commit that carries the message of the last commit in the range. When that last commit is a planning chore, git log shows only `chore(spec): mark started` with an empty body and no `Task:` trailer, while the tree holds the real change. A reader of git log misses the change, and the board loses the task trailer.

## Root cause
`internal/tidy/tidy.go:36`: `docs`, `wiki`, `polish` and `chore` subjects are all fold types, so a branch whose only real commit is `docs(wiki): ...` keeps no node. `internal/tidy/tidy.go:205-208`: with no kept node, the fallback builds one commit with `src: lastC`, the newest commit of the range, and `commitNode` (`tidy.go:356`) copies only that commit's message. Nothing prefers a non-chore commit, or a commit with a `Task:` trailer, as the source.

## Repro
Seen in a user repo on 2026-10-11. Branch off the parent with three commits: `docs(wiki): narrow paths that were broad by accident` (trailer `Task: PLN-<hash>#1`), then `chore(plan): tick and review notes`, then `chore(spec): mark started`. The parent has two unpushed chore commits after its upstream (`chore(spec): assign short ids`, `chore(plan): assign short ids`). Run `acta tidy <parent> <branch> --onto <upstream>`. It prints `tidy: 3 commits -> 1, tree ok, parent <sha>, folded 2`, and the one commit is `chore(spec): mark started` with no body and no `Task:` trailer.

Expected: the folded commit keeps the subject, body and `Task:` trailer of the one non-chore commit (the docs commit), since that is the real change.

Workaround: rebuild the commit with `git commit-tree` on the same tree and parent with the docs message and trailer, then `git reset --keep` to it.

## Found in
Reported by the user from a land in another repo; root cause read in main at 660b741.
