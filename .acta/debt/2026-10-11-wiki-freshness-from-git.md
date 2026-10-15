---
id: DBT-0109
hash: c1pswss
parent: plans/2026-10-11-wiki-freshness-from-git
---
# Review NOTEs: Wiki freshness from git

- [ ] (medium) acta tidy can move a Wiki-Checked trailer onto a later commit: when `acta wiki ok` runs before the branch's first kept commit, plan() carries the trailer forward onto the next code commit, so after tidy the page reads fresh against a change it never saw.
- [-] (low) gitc.CheckCommits runs a full-history `git log --grep` once per page, and gitc.Unseen passes every check sha on the command line; a page with thousands of edits makes the argument list grow (feed the shas with --stdin).
