---
id: SPC-0149
created: "2026-10-18 21:20:30"
hash: wsu2o97
started: "2026-10-18 21:22:56"
finished: "2026-10-18 21:23:31"
---
# Review understanding in setup

## Problem

The Sphinx-inspired `review_understanding` setting can only be changed through
`acta config`. The setup wizard preserves its current value but offers no choice.
The user approved adding a separate question after Review depth.

## Design

This is a Bounded change to the existing setup flow.

- Add a Review understanding question immediately after Review depth, before
  any plugin installation question. Use the existing wizard controls and look.
- Offer off and on. Explain that on adds an explanation and one optional
  scenario question after successful landing; the answer never blocks landing.
- Start with the saved value. An unset value starts off, matching the existing
  effective default. Save the selected value as an explicit boolean.
- Keep review depth and review understanding independent. Changing either
  answer must not change the other.
- Show the selected value in the answered-question summary. Update question
  counts to include the new step.

## Scope and validation

Update the existing setup defaults, form, answer conversion and focused tests.
Replace the test that requires setup not to ask about understanding. Verify
question order, labels and counts, unset/true/false defaults, changing either
boolean value, saved answers, summary text and independence from review depth.
Reuse the existing config key and controls; no new flag or dependency.

This supersedes the setup-preservation-only requirement in SPC-0146. Other
review and landing behavior remains governed by that design.
