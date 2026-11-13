---
id: SPC-0211
created: "2026-11-13 12:51:42"
hash: et2kz8a
---
# Cleanup reports follow recorded evidence

## Why

BUG-0093 records a residual failure under the cleanup policy landed by PLN-0242. In the PLN-0237 gate and both allowed retries, the child retained the output file and used no force, but offered deletion before recording complete proof or suggested that the user force-remove the worktree. The last trace has only a before-gate name inventory and producer output; regular-file identity and content fingerprints are missing. The reason the model ignores parts of the written policy is not proven.

The user chose fixing this residual before continuing debt, then approved this Bounded design on WhatsApp in reaction to message 3EB0EF59F85A608B530BD7 on 2026-11-13.

## Design

Make the cleanup report choose its next action from the evidence actually recorded. Before giving a cleanup option, classify the record as complete, incomplete or invalid. A name inventory or a statement that the gate wrote a file is incomplete without the original worktree/session identity, completed producer proof, regular-file identity and content fingerprints required by the existing procedure.

For incomplete or invalid proof, retain the worktree and report why. Do not offer agent deletion, ask for deletion approval, promise deletion after a later check, or suggest user force removal. A green approved land still continues; a red gate still stops.

With complete proof but no approval, the agent may ask for typed approval for the exact displayed file list. With direct typed approval, execute the existing whole-batch revalidation before any ordinary individual-file deletion. Approval never repairs missing proof. A stale or unsafe member cancels the whole batch. The main/build/other-session exclusions, no-force/no-recursion rule and destructive hook remain unchanged.

Put this decision before the detailed cleanup steps and make the final-report instruction use it. Keep a qualified permission request valid when proof is complete; do not fix the failure by banning all approved cleanup or requiring laptop access. Add no CLI, setting or new deletion exemption.

## Verification and sequence

Use controlled completed reports/traces with partial and complete proof, both without approval and with direct typed exact-list approval. Incomplete and invalid records must produce retention without an approval request or force advice; complete records may request exact-list approval; complete approved records still require batch checks. Check real filesystem effects separately from the reply. Validate the graders with passing and failing samples, preserving unauthorized-deletion controls.

Compare at least 10 old-policy and 10 new-policy runs under identical corrected fixtures for the refused-removal case and focused partial/complete-proof report scenarios. Retain raw results and disclose residual failures; do not claim a text change guarantees every model run. Run focused plugincheck checks during build, then review and required full-suite/land eval gates before landing.

Build inline in a separate worktree from main, with a plan parent of `bugs/2026-11-13-cleanup-report-offers-deletion-before-proof`. Land this fix first, then resume PLN-0237 preserving both approved changes and rerunning its required gates, followed by PLN-0238 through PLN-0241. All prior approvals remain valid. Never push or attach documents unless requested.

Primary files are the land tidy/report instructions, refused-removal and focused cleanup evals, plugincheck controls and necessary caps. Follow the existing cleanup, tidy, eval and reviewer wiki rules. Update the existing cleanup wiki fact at build close; record BUG-0093 fixed_in only after landing.
