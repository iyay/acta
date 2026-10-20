---
id: BUG-0053
hash: obk4y61
priority: high
---
# acta wiki check takes minutes on a repo with long history

## Symptom
`acta wiki check` in be-pmis (20,399 commits, 173 wiki pages) runs past 120 s and looks hung. A SIGQUIT dump shows it waiting on a child `git` process.

## Root cause
Each page runs about 4 git commands: `gitc.IsDirty`, two `git log --full-history` in `gitc.CheckCommits` (`internal/gitc/gitc.go:592`), and `gitc.Unseen`, all called from `staleMsgs` (`internal/wiki/check.go:131`). One full-history log costs about 0.66 s on that repo, so 173 pages take about 8 minutes. The `log -p` over the checked file gives the same output for every page, but it runs again for each one.

## Repro
cd ~/Nayakatara/PMIS/codes/febe/be-pmis && time acta wiki check

## Found in
main, by reading a SIGQUIT goroutine dump of a stuck run.
