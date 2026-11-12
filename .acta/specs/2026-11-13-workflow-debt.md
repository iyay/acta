---
id: SPC-0208
created: "2026-11-13 05:47:43"
hash: f7mtrw9
---
# Close the five workflow debt groups

## Why

DBT-0164 through DBT-0168 hold 14 open items from the five plans landed by session #6. The user asked to continue this work on WhatsApp on 2026-11-13, after the proposed design named all five groups. This is a Bounded change to existing skills and their tests. Build runs inline, in five serial plans, one per debt file.

## Design

1. **Land, DBT-0164.01–.02.** Use `<gate>:<case>` in the baseline rule and its test. The second-run rules in land and tidy apply the same gate selection as the first run: forcing paths also trigger a run; no run paths means a whole run; missing gate metadata or doubt also means a whole run. A parent that brings only a forcing path cannot reuse that gate.
2. **Review evidence, DBT-0165.01–.03.** A cut BLOCKER list or missing Coverage keeps the round BLOCKED even after every listed BLOCKER is rejected. With no fixable BLOCKER, repeat the incomplete range in the next round. With fixes too, the next review covers both the fixes and every incomplete range. Round 3 still incomplete means stop and ask; there is no round 4. Build hands review the merge-base range. Eval cases must exercise a report without Coverage and a false BLOCKER whose repro actually runs. A missing Coverage report returns to the same reviewer once. Verdict grading accepts capitalization and Markdown emphasis without accepting a quoted verdict from another report.
3. **Review auto, DBT-0166.01–.02.** State that the safe-case plan was built inline. Count distinct agents assigned a reviewer role, not all Agent calls: returning to the same reviewer for Coverage is allowed, and a helper alone is insufficient. Set an explicit standard depth in the polish case that expects two reviewers; grade their assigned roles, rather than incidental words in a combined brief.
4. **Debug, DBT-0167.01–.03.** Capture the absolute root of the checkout under investigation before changing folders. Copy from that root, never `$OLDPWD`; test with a worktree that differs from main. Until proof, tests, polluter hunts and profilers that write run in the temp copy with output and cache paths there. File-tool writes also use verified absolute temp paths. Bug records are allowed after proof; the existing known-flake recording path remains an explicit metadata-only exception. Neither allows source edits. Probe folders are private and newly created: restrictive permissions, exclusive creation, no fixed shared `/tmp` directory and no reuse after a collision. Use a creation method the eval sandbox supports; do not weaken isolation when creation fails.
5. **Cold read and evals, DBT-0168.01–.04.** Remove the answer-revealing requests to relay all findings from the two planted-defect prompts. Add a case with six findings that checks every returned finding reaches the user and any cut-count line is preserved. The contradiction grader requires recognition of the conflict and accepts both `30 days`/`7 days` and `30-day`/`7-day`. Require `### Read`, a nonempty list of spec sections, and a list of opened reference paths or an explicit `none`. Missing or empty evidence follows the existing incomplete-report choice; a bare “Read: all of it” is insufficient. Correct the seven comments about `input_match` to describe JavaScript regex matching, including both `review-depth-floor` grader comments.

Use existing grader types and test helpers. Update required-string tests with the behavior they guard. Raise a text cap only when the full instruction needs it; do not trim safety checks to fit. Update affected wiki facts during build. No CLI feature, new configuration, dependency, review depth or work on SCR-0096 belongs here.

## Testing

Each task starts with a focused failing regression test. Grader changes include valid and invalid sample reports, including Coverage retries, helpers, quoted verdicts, Markdown emphasis, hyphenated durations, incomplete reports and mixed fixed/incomplete rounds. Check grader behavior against those samples before using its scores as evidence.

Measure each skill-text group on relevant behavioral cases with at least 10 runs per arm, old text against new text under the same corrected graders and scaffolds. Keep raw counts and separate grader corrections from skill effects. If a group regresses on a case, isolate its changes and leave out the regressing change; its debt stays open. Do not claim an item closed merely because a required-string test passes.

Use `scripts/test ./internal/plugincheck` for focused checks. The land gate runs `scripts/test --full` and the eval tags selected by each plan's changed paths. One `--case` per call; tags use `--tag=<name>`. Builds and reviews use the branch binary on PATH where hooks need it.

## Context

Plans set `parent: debt/<stem>` and list their exact item IDs in `closes:`. Run the five plans serially because they share plugincheck and eval files. Required wiki context: `skill-escape-lines-skip-proof`, `eval-gate-by-tag`, `eval-one-case-per-call`, `reviewer-temp-clone`, `plugincheck-caps-per-folder`, and `plugin-loads-from-repo`. No push is authorized.
