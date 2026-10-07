---
id: DBT-0012
hash: kb3a9cs
parent: plans/2026-09-28-gitignore-link
---
# Review NOTEs: Refuse a Linked .gitignore Implementation Plan

- [x] internal/hook/gitignore_test.go:128 comment says the message is the only thing the user sees, but every caller drops the error or returns before the call.
- [ ] (low) internal/hook/gitignore.go doc says a linked .gitignore "writes nothing" like the nil cases, but this case also returns an error; line 13 "Callers ignore the error" was already wrong for doctor Fix.
- [ ] (medium) Session-start and tick now refuse a linked .gitignore silently, so .acta/.agents.json stays untracked; only acta doctor points it out.
- [ ] (low) A hard-linked .gitignore still gets written through, since Lstat sees a plain file; needs local access.
- [x] Plan Step 1 code block does not show the error-text check that commit de21248 added. (stale)
