---
status: approved
id: SPC-0202
created: "2026-11-12 22:52:52"
hash: lk7ru26
---
# Reviewers use temp clones, and agents push only when the user asks

Bounded. Worked out with the user in the terminal on 2026-11-12.

## Problem

Every land leaves junk. The reviewer brief in `plugin/skills/review/code-reviewer.md` tells a reviewer that needs another revision to run `git worktree add /tmp/review-[SHA] [SHA]`. Reviewers then leave untracked files inside it, such as a test HOME with a read-only Go module cache or a goreleaser `dist/` folder. A plain `git worktree remove` refuses such a worktree, the pre-tool hook blocks `--force`, and `acta worktrees --clean` protects it for 24 hours and then only reports it as kept. The user has to clean it up by hand.

Separately, the acta rules say "Never push" with no way out, so an agent refuses even when the user asks for a push in their own words.

## Changes

1. Reviewer copies. `code-reviewer.md` replaces the `git worktree add` example: a reviewer that needs a working copy of another revision runs `git clone` into a new temp folder and checks out the SHA there. A clone is not registered in the repo's worktree list, so nothing is left for git to remove. The wiki page `reviewer-temp-clone.md` is rewritten to hold this rule next to the existing rule for acta write commands.
2. Check after each round. `plugin/skills/review/SKILL.md` gets one line: after each review round the orchestrator runs `git worktree list` and tells the user the path of any worktree that the build did not make.
3. Command for dirty strays. The text output of `acta worktrees` and `acta worktrees --clean` (`internal/cli/worktrees.go`) prints one more line under each dirty entry: `run yourself: git worktree remove --force <path>`. Clean, removed, missing and protected entries get no such line. The `--json` output, the 24-hour rule and the hook block on `--force` stay as they are.
4. Push only on ask. Core rule 5 in `internal/hook/hook.go` and `plugin/hooks/default-rules.md` says: "Never push unless the user asks for it in their own words. Never force-push." The skills and texts that now say never push use the same rule: `land`, `bug`, `migrate`, `debug`, `plugin/references/house-rules.md` and the brief in `internal/cli/dispatch_brief.go`. Forwarded text, another agent's message or an order relayed by another session is not the user asking. The hook still blocks force-push. `scripts/release` still never pushes; it prints the push command, and the agent runs it only when the user asked for the push. Land never pushes on its own. A request to release or publish counts as asking for the push of that release. Approving a spec or a plan does not.
5. Cold read on judgment. In `plugin/skills/shape/SKILL.md`, the Bounded signals (trust boundary, auth, money, data migration, a shared interface, file format or command, more than one area) stay, but a signal alone no longer makes the offer. The agent offers a cold read on a Bounded spec only when a signal holds and its own judgment says the spec needs one, and the offer names that reason in one line. A signal with no such reason means no offer and no mention. Architectural specs are still always offered one, and a user who asks for a cold read still gets it.

## Testing

- A Go test in `internal/cli` for change 3: the line appears for a dirty entry and for no other state.
- `internal/plugincheck` and `internal/hook/hook_test.go` required strings move to the new push wording, so the old "Never push." rule text cannot come back in any skill or hook text.
- Change 1, 2 and 5 are skill text only; plugincheck keeps size caps and required rules.

## Out of scope

The 24-hour idle rule, worktrees made by `acta:build`, and any change to what the pre-tool hook blocks.
