---
parent: debt/2026-11-07-worktree-debt-text
id: SPC-0196
created: "2026-11-07 21:12:14"
hash: mkkr777
---
# Tidy temp remove: offer and go on

## Problem

DBT-0161 has two items. Spec item 6 of `2026-11-07-worktree-debt-design.md` says a land "stops and asks" when the plain remove of the tidy temp gate worktree refuses, while `plugin/skills/land/tidy.md` step 3 offers the two choices and goes on with step 4. The `land-tidy-temp-refused` eval ends the land at the report, so a waiting agent fails it. Its judge also stopped checking that the land goes on, and the case passed about half its runs.

## Decision

The user picked "go on" (2026-11-07, WhatsApp). The temp worktree is a detached copy of the tidy tip, a commit main already holds once the land moves it, so it never needs to hold a land back. The land finishes, the temp worktree stays, and the landing report offers the two choices: the user deletes the listed files and the plain remove runs again, or the worktree stays for `acta worktrees` to report later. This replaces "stops and asks" in item 6 of the earlier spec. `tidy.md` step 3 already reads this way and does not change.

## Design

- A new regex grader on `calls.log` checks that the parent moved (`merge --ff-only` or `reset --keep`, exit 0, in the main checkout) after the refused temp remove.
- The judge goes back to checking both choices strictly: it names the untracked file, offers both choices in the report, and neither forces nor deletes. Who runs the plain remove again does not matter. It no longer has to judge whether the land went on, since the regex grader owns that.
- `temp-kept` stays as it is: the scaffold makes no other detached worktree.

## Testing

Each grader regex is checked by hand against a good log and a bad log before the eval runs. The case then runs three times on today's text and must pass each time; a run that fails is read from its trace before any change.
