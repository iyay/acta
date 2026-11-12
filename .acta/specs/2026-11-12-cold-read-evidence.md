---
id: SPC-0205
created: "2026-11-12 23:06:39"
hash: q1p72lb
---
# Cold read findings carry their evidence

Bounded. Worked out with the user in the terminal on 2026-11-12. The user approved the design in chat. The review skill and the debug skill have their own specs, SPC-0203 and SPC-0204.

## Problem

The cold read in `plugin/skills/shape/cold-read.md` guards its gates well: a fresh context, no silent fallback, one pass, no approval of spec or plan. It does not guard the quality of what the reviewer returns.

- A reviewer that answers "no consequential findings" does not have to say which parts of the spec it read or which references it opened. A pass that read nothing looks the same as a careful one.
- The brief says the reviewer follows the spec's references "as needed to check a claim". A spec that states something wrong about a file, a command or how the code behaves today can pass, though a new reader is the one best placed to catch it.
- The orchestrator is told to "present the findings briefly". It wrote the spec, and nothing stops it from shortening, softening or leaving out a finding.
- The brief looks for missing decisions, contradictions and ambiguity. It does not look for a requirement that nobody could test.
- The reviewer returns five findings at most and says nothing about the ones it cut.
- `plugin/evals/` has six cold read cases. They cover the offer, skip, accept, the unavailable path and the context that must not leak. None plants a defect in a spec and checks that the reviewer finds it.

## Changes

1. Eval cases first. Two new cases in `plugin/evals/`, written and run against today's text before `cold-read.md` changes, so each has a baseline:
   - `cold-read-planted-contradiction`: two sections of the spec say opposite things. The findings must name both lines.
   - `cold-read-false-claim`: the spec says a file in the scaffold holds something it does not hold. The findings must name the spec line and the file.
2. More for the reviewer to check. The brief adds two duties: check every claim the spec makes about files, commands or behaviour that exist today against the repo, and report each one that is wrong; and report each requirement that has no way to be tested. Each finding quotes the spec line it is about.
3. A report that shows what was read. The reviewer's reply gets a `Read` part: the sections of the spec it read and the references it opened. It is there when the answer is "no consequential findings" too. A reviewer that cut findings at the cap of five adds `N more findings not listed`.
4. Nothing dropped on the way to the user. `## Resolve` says the orchestrator presents every finding the reviewer returned, each with its quoted spec line. It may add one line under a finding saying whether it agrees and why. It never leaves a finding out or merges two into one. A reply with no `Read` part is told to the user as an incomplete cold read, the same way a failed one is: ask whether to retry or go on without it, then wait.

## Rule for keeping a text change

Each numbered change to `cold-read.md` is judged on the two new cases and on `cold-read-accept`, 10 runs per arm, old text against new text. A change that scores below the old text on any of those cases is left out, and the plan's `## Review notes` say which. This is the rule the wiki page `skill-escape-lines-skip-proof.md` records.

## Files

- `plugin/skills/shape/cold-read.md`
- Two new case folders in `plugin/evals/`
- `internal/plugincheck/skill_shape_test.go` for the required strings, `internal/plugincheck/evals_test.go` for the case list, and `internal/plugincheck/budget_test.go` for the byte cap. The cap for `cold-read.md` is 2400 bytes, so it goes up. No words an agent acts on are trimmed to fit.

## Testing

- The two new cases and `cold-read-accept`, 10 runs per arm. That is about 60 runs, about 10 minutes at `scripts/eval -j 8`, and it costs quota.
- `internal/plugincheck` keeps the required strings and the size caps.

## Order

The plans of SPC-0203 and SPC-0204 (PLN-0232 and PLN-0233) change `evals_test.go` and `budget_test.go` too. The build of this spec starts after PLN-0233 lands.

## Out of scope

When a cold read is offered (PLN-0231 changes that), a cold read of a plan, a second pass, and a second reviewer.
