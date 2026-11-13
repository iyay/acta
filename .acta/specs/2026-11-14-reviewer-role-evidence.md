---
parent: debt/2026-11-13-review-depth-auto
id: SPC-0212
created: "2026-11-14 05:36:37"
hash: o00a18d
---
# Count reviewer launches from actual tool inputs

## Why

This Bounded addendum replaces the evidence method for SPC-0208 section 3 and PLN-0239. The native trace-focused judge input omitted Agent launches present in the raw trace. Its scores cannot prove reviewer identity or count. The user approved this revised design through WhatsApp reaction 3AD28D3753B833E48F9E on question 3EB0B33466D07531FC1FB7 at 05:33 on 2026-11-14.

## Design

Both eval prompts require each actual reviewer Agent prompt to start with `ACTA_REVIEW_AXIS=BOTH`, `ACTA_REVIEW_AXIS=SPEC`, or `ACTA_REVIEW_AXIS=STANDARDS`, followed by the normal review brief. The declaration must match the assigned work. Helpers receive no reviewer declaration. These declarations describe assignments; they do not prove review quality.

Use existing `tool_used` graders over actual Agent inputs. Count fresh declared reviewer launches and exclude calls with a nonempty resume ID. Resuming the same reviewer for Coverage and sending it a message do not create another reviewer. Empty or null resume fields still mean a fresh launch. Incidental role words elsewhere in a prompt do not count.

The safe auto case records an inline build. Its existing `one-reviewer.md` grader requires exactly one fresh declared reviewer launch. A new `combined-role.md` grader requires exactly one fresh BOTH launch. Helpers alone, a Spec-only reviewer, and two fresh reviewers fail. A combined reviewer with helpers or resumed requests passes these role checks.

The polish case explicitly sets standard depth. Its existing Spec and Standards graders each require exactly one fresh launch of their own axis. A new `reviewer-count.md` grader requires exactly two fresh declared reviewer launches in total. A single combined reviewer, a missing axis, an extra reviewer, and a resumed reviewer presented as the other axis fail. Helpers and resumes do not inflate the count.

Keep this work within the two eval cases and plugincheck tests. Restore the separate Standards grader and remove the superseded LLM judge allowlist entries. No product skill, hook, CLI, dependency, configuration option, or other debt item changes.

## Verification

Start with failing samples through the real loaded graders using `evalomp.LoadCases` and `evalomp.Grade`. Cover empty traces, helpers, incidental words, combined assignments, separate axes, missing axes, extra launches, nonempty resumes, and empty/null resumes. Check the full set of role graders for each sample, including quoted marker text that is not the start of the actual Agent prompt.

Run focused plugincheck checks and the package suite. Then run each named case separately for 10 native runs with the matching branch binary. Preserve the earlier raw traces and scores as evidence of the rejected method, separately from the corrected results. These are grader and fixture changes, so they make no claim of improved skill behavior. The full repository suite runs at land. Review and land follow the existing plan settings; no push is authorized.
