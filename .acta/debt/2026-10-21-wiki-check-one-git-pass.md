---
id: DBT-0137
hash: e4fqd1q
parent: plans/2026-10-21-wiki-check-one-git-pass
---
# Review NOTEs: Wiki check reads git once per run

- [ ] (medium) acta doctor still runs gitc.CheckCommits and gitc.Unseen per wiki page (internal/doctor/wiki.go:129,141), so it costs about 0.6 s per page on a 20k-commit repo; move it to gitc.ReadWikiHistory
