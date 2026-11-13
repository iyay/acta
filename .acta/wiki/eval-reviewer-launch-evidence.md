---
type: Gotcha
title: Reviewer counts need actual launch inputs
description: Trace-focused judge input omitted Agent launches; count declared reviewer roles from actual tool inputs
paths: [plugin/evals/review-auto-safe-one-reviewer/, plugin/evals/polish-full-review/]
---

For reviewer-count evidence, compare the judge's actual input with the raw trace. In PLN-0239, `focus: trace` omitted Agent launches that the raw trace contained, including on passing scores. Workspace PostToolUse capture did not run, and `tool_calls` was not a supported focus source. The two reviewer-role cases now use declared prompt-prefix axes and existing `tool_used` checks over actual fresh Agent inputs; nonempty resumes reuse an identity. These declarations prove assignments, not review quality. Validate loaded graders with positive and negative call samples before native runs.
