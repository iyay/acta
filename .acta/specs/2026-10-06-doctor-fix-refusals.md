---
parent: debt/2026-09-28-skill-rules
closes: [DBT-0011.16, DBT-0011.24, DBT-0011.30, DBT-0011.31]
id: SPC-0093
created: "2026-10-06 08:46:18"
hash: yknw0bb
---
Status: Bounded, approved by the user in chat on 2026-10-06. Written down because `acta doctor --fix` writes files in the user's repo.
Why: `acta doctor --fix` can write into `.git`, or append to another folder's tracked `.gitignore` through a symlinked `.acta`, and still report ok. A broken `.acta` link makes doctor suggest `--fix`, which then exits 3 with a raw error. A git status that cannot be read is treated as clean.

# doctor --fix: one refusal rule, and a report that matches the disk

## Changes

1. **One rule for both sides.** `internal/doctor/doctor.go` gets `refusal(e Env) string`. It returns why `--fix` must not write, or "" when it may. Both `Fix` and `checkRepo` use it, so the write and the report can never disagree. It refuses when:
   - the acta root is outside the repo (the rule that exists today);
   - the acta root is the repo's `.git` folder or anything inside it (DBT-0011.30);
   - any part of the acta root's path below the repo root is a symlink, including a link whose target is missing (DBT-0011.31, DBT-0011.24);
   - the root's `.gitignore` is a symlink (exists today) or exists but is not a regular file, for example a folder (DBT-0011.24).
2. **When it refuses.** `Fix` writes nothing and returns no paths and no error. `checkRepo` reports `fail` with the reason as its message and a manual fix line (for example "fix root in .acta.yaml" or "replace the .acta link with a real folder"), never `acta doctor --fix`. `acta doctor --fix` then exits 1 through the failed report.
3. **Git status that cannot be read.** In `internal/cli/doctor.go`, when `gitc.IsDirty` returns an error, `--fix` still writes the file but does not commit it, and prints `fixed, not committed: cannot read git status: <err>` (DBT-0011.16).

## Out of scope

DBT-0011.02, .03, .10 and .21 (conflicts and voice checks) stay open as debt.

## Testing

Each refusal reason gets a failing test first in `internal/doctor/doctor_test.go`. Each test checks both sides: after `Fix`, no file in the temp repo changed, and `checkRepo` is `fail` with no `acta doctor --fix` in its fix line. The `IsDirty` error case is tested in `internal/cli`. Gate: `scripts/test --full`.

## Close

Last task bumps the patch version to 0.1.25 in `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json` and `plugin/package.json`.
