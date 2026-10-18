---
id: SPC-0146
created: "2026-10-18 19:23:25"
hash: jpblcy6
---
# Optional understanding question after deep review

## Problem

Deep review checks the code, but its landing report does not help the user
understand the key behavior and its limits. Sphinx's scenario questions are
useful here, without adding a quiz gate to acta's automatic landing flow.

## Design

This is a Bounded change. The user approved the design in chat.

- Keep the existing review passes, fact-check, fix rounds, polish, gates and
  automatic landing. No question or human answer may delay landing.
- When the selected review depth is `deep`, extend the successful landing
  report with a short explanation of the main behavior, the most important
  risk or boundary, and what the tests do and do not establish. Use a concrete
  example and inspected code as evidence; do not invent a risk to fill space.
- End that report with exactly one optional scenario question. Ask the user
  to predict an outcome and give a short reason. Use the final landed code,
  including fixes, polish and merge changes. Do not test names or syntax.
- The user may answer, say skip, ignore the question or move on. None of
  these affects the completed landing. Do not record passed/skipped status,
  add a setting, or launch an extra reviewer just for this question.
- If the user answers, discuss the prediction and its cause against the
  code. Explain a missed concept and offer a further scenario without
  requiring it. Resolve ambiguity from the code instead of treating a
  short or differently worded answer as wrong.
- `quick` and `standard` retain their existing reports. Applying the deep
  lens to a sensitive path at `standard` does not select this feature.
- Preserve the required landing evidence and report fields. The optional
  question is the report's last item. A failed or blocked landing does not
  produce an after-landing question.

## Scope

Update `plugin/skills/review/SKILL.md`, `plugin/skills/review/depth.md` and
`plugin/skills/land/SKILL.md` to carry the selected depth through the close
and define the report and follow-up. Use existing skill text and review
notes; no new runtime state, provider, Git hook, MCP server or question
engine. Add focused checks in `internal/plugincheck/` and greenfield evals
under `plugin/evals/`. Adjust text budgets only where the change needs it.

## Validation

Start with failing checks. Verify that deep review lands before asking,
uses the final code, and asks one optional behavior question; standard and
quick do not ask it; a deep lens alone does not enable it; skip or no answer
never blocks landing; a follow-up explains the cause without a pass label.
Run plugin checks, the relevant greenfield evals, and the repo's landing
gates. Keep the plugin's file and folder budgets intact.

## Wiki

No new page. Workflow rules live in the skills.
