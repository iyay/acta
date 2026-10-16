---
id: SCR-0073
hash: qe5825q
title: acta id should not make its own commit; the id lands in the commit that adds the file
status: raw
created: "2026-10-16 08:10:15"
schema: "1"
finished: "2026-10-16 08:12:14"
---
# acta id should not make its own commit; the id lands in the commit that adds the file

## Words

### 2026-10-16

Seen in sales-agent on 2026-10-16: every new spec or plan leaves a lone "chore(spec): assign short ids" or "chore(plan): assign short ids" commit right after the commit that added the file. In 124 unpushed commits there were about 22 of them, often as a plan+spec pair next to each other. The kinds come from SPC-0130, so same-kind ones fold, but a plan+spec pair never folds, and the id commit never joins the commit that created the file.

The id commit holds only the frontmatter id and hash. It cannot be reviewed or reverted on its own, so it should not exist as a commit.

Want: when acta id gives an id to a file that is new or uncommitted, the id goes into that file's own commit (acta commit, or the shape/slice commit), not a commit of its own. When the file is already committed and unpushed, amend or fold the id into the commit that added it. Only a file that is already pushed keeps a separate id commit.

User words: "chore: assign short ids itu perlu satu commit sendiri nggak ya sebenernya" ... "ini harus dibenerin".
Related: SCR-0056, SCR-0067, SPC-0130.

## Context

## Log

## Open questions
