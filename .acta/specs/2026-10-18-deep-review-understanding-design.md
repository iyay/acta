---
id: SPC-0146
created: "2026-10-18 19:23:25"
hash: jpblcy6
---
# Optional understanding question after landing

## Problem

Review checks the code, but its landing report does not help the user
understand the key behavior and its limits. Sphinx's scenario questions are
useful here, without adding a quiz gate to acta's automatic landing flow.

## Design

This is a Bounded change. The user approved the design in chat.

- Keep the existing review passes, fact-check, fix rounds, polish, gates and
  automatic landing. No question or human answer may delay landing.
- When `review_understanding` is effectively `true`, extend the successful landing
  report with a short explanation of the main behavior, the most important
  risk or boundary, and what the tests do and do not establish. Use a concrete
  example and inspected code as evidence; do not invent a risk to fill space.
- End that report with exactly one optional scenario question. Ask the user
  to predict an outcome and give a short reason. Use the final landed code,
  including fixes, polish and merge changes. Do not test names or syntax.
- The user may answer, say skip, ignore the question or move on. None of
  these affects the completed landing. Do not record passed/skipped status
  or launch an extra reviewer just for this question.
- If the user answers, discuss the prediction and its cause against the
  code. Explain a missed concept and offer a further scenario without
  requiring it. Resolve ambiguity from the code instead of treating a
  short or differently worded answer as wrong.
- Review depth and understanding are independent. The feature works with
  `quick`, `standard` or `deep` when enabled, and none enables it implicitly.
  Applying the deep lens to a sensitive path does not enable it either.
- Preserve the required landing evidence and report fields. The optional
  question is the report's last item. A failed or blocked landing does not
  produce an after-landing question.

## Setting

- Add `review_understanding`, a YAML boolean, in `~/.acta/config.yaml` and
  `.acta.yaml`. Missing means `false`; an explicit repo value wins over the
  user's value in either direction. Preserve explicit `false` through reads,
  writes and setup; an omitted repo key inherits the user value.
- `acta config set --review-understanding true|false` writes the user value;
  `--repo` writes the repo value. `acta config set --repo --unset
  review_understanding` removes that override. Invalid values are rejected
  before writing, with the key named in the error.
- `acta config show` prints the effective value and existing default/repo
  marks; JSON uses a boolean and the existing provenance fields. Include the
  effective value in the session settings so agents honor it in every harness.
- When `false`, use the ordinary landing report at every depth: omit both
  the added explanation and question, and do not offer the check again.
- Keep the existing setup questions; preserve this setting during setup.
  No per-run flag or separate quiz mode is added.

## Scope

Update the existing config model and repo merge, CLI set/show/help, session
settings and setup preservation, with their tests. Update
`plugin/skills/review/SKILL.md`, `plugin/skills/review/depth.md` and
`plugin/skills/land/SKILL.md` to define the independent switch, report and
follow-up. Use existing skill text and the effective boolean; no new
checkpoint state, provider, Git hook, MCP server or question engine.
Add focused checks in `internal/plugincheck/` and greenfield evals
under `plugin/evals/`. Adjust text budgets only where the change needs it.

## Validation

Start with failing checks. Verify both switch values at all three review
depths: enabled lands before asking, uses the final code and asks one
optional behavior question; disabled neither asks nor offers a check.
A deep lens alone does not enable it; skip or no answer never blocks
landing; a follow-up explains the cause without a pass label.
Verify the default, explicit false, both directions of repo override, unset,
boolean YAML/JSON round-trips, rejected invalid input without file changes,
setup preservation, and effective config in CLI and session settings. Verify
that choosing a review depth never changes the understanding setting.
Run plugin checks, the relevant greenfield evals, and the repo's landing
gates. Keep the plugin's file and folder budgets intact.

## Wiki

No new page. Workflow rules live in the skills.
