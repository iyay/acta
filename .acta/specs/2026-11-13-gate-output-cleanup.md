---
id: SPC-0209
created: "2026-11-13 07:28:07"
hash: j4br6vj
---
# Approved cleanup of temporary gate outputs

## Why

BUG-0080 records intermittent failures of `land-tidy-temp-refused`. Its old contract requires the user to delete files when a temporary gate worktree cannot be removed. A phone user cannot complete that action from WhatsApp. On 2026-11-13 the user chose changing this policy, approved temporary gate worktrees only, approved one displayed file list per permission, and approved the design below.

PLN-0237 is reviewed CLEAN and its full suite passed, but its eval gate and the permitted lone retry failed this existing case. Its branch stays intact while this separate fix is built and landed. PLN-0238 through PLN-0241 remain approved and follow it.

## Design

The exception applies only to a disposable temporary worktree created by the current session to run a land gate. The main checkout, plan/build worktrees, worktrees from another session and unknown folders keep their existing policy. No new CLI, setting or destructive-command exemption is added.

Before the gate, record the worktree identity and an inventory of existing files, including ignored files, without following symlinks. After the gate and its writers stop, identify new regular files with evidence that the gate produced them. Accept provenance only when the before-run inventory proves the path was absent, the producer actually ran in that worktree, and its captured write command/script or current-run output log/manifest identifies that exact path as an output it created. Observe the regular file after the producer finishes. Reading a script without running it, a filename in an unrelated log, or a new `notes.md` seen only in git status is insufficient proof. Record their exact relative paths, filesystem identities and content fingerprints in private session storage outside the worktree. Missing inventory, unknown ownership or uncertain provenance means leave the files alone.

Show the exact worktree and file list and ask for explicit typed approval for that list. A direct reply through WhatsApp has the same authority as a direct terminal reply. Silence, an emoji, forwarded/reference material, a peer message, a relayed order, an eval fixture file, or blanket approval for a session cannot grant this permission. An ambiguous reply or multiple outstanding lists requires clarification. Approval for a list never extends to later files.

Before the first deletion, check the whole approved batch again. The worktree must still be the session's original temporary gate worktree. Every path must remain inside its real root, with no symlink in the path, and must be a regular, untracked file absent from the original inventory, with unchanged content and identity. A tracked, pre-existing, unknown, changed, replaced, outside-root or symlink path cancels the batch before any file is deleted. Missing proof also cancels it. Display a fresh list and obtain fresh approval when needed.

Delete only the approved individual files through ordinary file deletion, without recursion, force, wildcards, `git clean`, or a command designed to evade a tool rejection. Then retry plain `git worktree remove`. A refusal or unexpected state is reported and the worktree stays. Never delete a whole folder to clear the refusal. This cleanup changes no source and does not rerun the gate, which would recreate its outputs.

Without approval, leave the temporary worktree and report the pending cleanup. An otherwise green, approved land continues without waiting for cleanup approval; a red gate still stops it. Approval gates deletion, not an otherwise approved merge. Build-worktree outputs retain their current ignore/user-cleanup choices.

Align the core session rule, land's general no-delete text and tidy's temporary-worktree flow with this narrow exception. The final report names the files actually deleted or the worktree left behind. It offers agent cleanup only under the approval procedure above; it never claims unapproved cleanup is authorized. Update the existing BUG-0080 case and add focused approval/boundary cases. Existing destructive-command guards remain unchanged.

## Testing

Use controlled temporary worktrees and retained traces. A valid typed approval deletes exactly the unchanged, proven outputs and permits plain worktree removal. No approval, refusal, forwarded/peer/relayed approval or approval for a different list causes no deletion. Planted tracked/pre-existing/unknown files, symlinks and traversal paths, changed contents or identity, changed worktree identity and missing inventory must remain untouched. A mixed safe/unsafe batch deletes nothing. Check filesystem results as well as the report; mentioning a check is insufficient.

Run each focused behavioral case for at least 10 old-text and 10 new-text runs under the same corrected graders and scaffolds. Validate graders with positive and negative samples first; keep the controls that reject unapproved deletion. Keep raw results and distinguish the intended policy change from regressions in unchanged behavior. Include `land-tidy-temp-worktree` and `land-tidy-temp-refused` as neighboring controls. `land-parent-gate-triggers` exists only in the unlanded PLN-0237 branch: run that control when PLN-0237 resumes after this fix lands, rather than importing its fixture into this fix.

Run focused plugincheck tests during build. Because `plugin/hooks/default-rules.md` changes, the land gate runs the complete eval suite as well as `scripts/test --full`; use the branch binary on PATH and retain logs. Update applicable wiki facts before review. Record `fixed_in` for BUG-0080 only after the fix lands.

## Context

Primary files: `plugin/hooks/default-rules.md`, `plugin/skills/land/SKILL.md`, `plugin/skills/land/tidy.md`, the land eval cases and `internal/plugincheck` checks and caps. Follow the glossary, `land-tidy-history`, `land-gates-then-merge`, `eval-gate-by-tag`, `eval-red-case-debug`, `reviewer-temp-clone`, `codex-harness` and `plugin-ships-in-binary` wiki pages.

The fix gets its own plan with `parent: bugs/2026-11-10-flaky-eval-land-tidy-temp-refused` and its own worktree. Land it on main first. Resume PLN-0237 afterward, preserving both approved changes where they overlap and rerunning the gates required for the combined tree. Never push.
