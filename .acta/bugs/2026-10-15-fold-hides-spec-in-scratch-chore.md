---
id: BUG-0045
hash: hxrw5it
---
# A new spec folds into a scratch chore commit and the subject hides it

## Symptom
On main, `acta id specs/2026-10-15-strip-release-binary-design` did not make its own commit. The new spec file went into HEAD, a `chore(scratch)` commit. The result is 735cf9b, subject `chore(scratch): new 2026-10-15-optional-tui-install, +3 more`. It holds the scratch file and the new spec, but the subject never names the spec. The body ends with `chore: assign short ids`.

## Root cause
Not proven yet. Likely `internal/gitc/gitc.go:251-254`, `sameFiles`. It checks only that every file of HEAD is in `paths`, not the other way. When `acta id` commits the spec plus its scratch parent (`closeScratchParent` in `internal/write/ids.go`), HEAD holds only the scratch file. So `sameFiles` says yes at `gitc.go:225`, and `tryFold` amends the new spec into the scratch chore.

## Repro
1. On a clean main, run `acta scratch new x` and `acta set scratch/<stem> status raw`, so HEAD is a `chore(scratch)` commit.
2. Write a spec with `parent: scratch/<stem>`.
3. Run `acta id specs/<spec stem>`.
4. `git log -1 --stat`: the spec is inside the `chore(scratch)` commit.

## Found in
main, 2026-10-15, found by reading `git show --stat HEAD` after `acta id` on SPC-0131.

Seen again the same day: SPC-0133 folded into 9adcda7 `chore(scratch): add to 2026-10-15-optional-tui-install`, after only `acta scratch add` calls.
