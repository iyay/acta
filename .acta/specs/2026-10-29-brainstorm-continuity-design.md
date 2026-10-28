---
id: SPC-0183
created: "2026-10-29 06:30:10"
hash: gs4lpcc
started: "2026-10-29 06:37:25"
finished: "2026-10-29 07:28:06"
---
# Brainstorm continuity by scenario

## Problem

Shape limits each session to one Architectural brainstorm. Hook guards and
session instructions enforce that limit even when the user wants to continue.
The count does not establish whether earlier context helps or harms the next
discussion. Moving sessions can also lose useful decisions and assumptions.

## Scope

This Bounded change replaces the existing session limit with guidance based on
the relationship between topics and the state of the discussion. It applies
across harnesses and communication channels. Walkie needs no special behavior.

## Behavior

Continue in the current session by default. Handle these cases explicitly:

- For a related topic, identify the earlier decisions that still apply. Keep
  each separate Architectural topic attached to its own scratch item.
- For an unrelated topic, establish its scope and assumptions separately from
  earlier work. Offer a new session when earlier context could interfere, with
  a concrete reason rather than a count of past brainstorms.
- When resuming an earlier topic, read its scratch item and recorded decisions.
  Confirm assumptions that may have changed before relying on them.
- When topics or assumptions become mixed, restate the active scope, assumptions,
  and decisions first. Offer a new session if that would help. The user chooses
  whether to continue here or move; continuing remains allowed.

Do not launch another session, tab, or agent without the user's choice. When the
user chooses to move, preserve the current discussion in its scratch item and
give the destination the item's actual identifier. Offer only handoff methods
available in the current harness; Herdr remains conditional on its session hint.

Architectural scratch creation, brainstorming status, answer logs, classification,
and design and plan approvals remain required. Spike and Bounded keep their
existing paths. Frame keeps new items raw because framing precedes design.

## Implementation scope

Update shape and frame instructions, shared and generated session rules, and
harness guidance that describes the limit. Remove the hook block, its reminder,
and session bookkeeping used only to enforce the limit. Keep shared event parsing
and unrelated guards. Old brainstorm state files need no migration or deletion;
they must no longer block or warn about a later topic.

Update affected hook, CLI, and plugincheck tests and behavior evals. Replace
mandatory second-brainstorm handoff expectations with the scenarios above.
Historical completed specs and plans are records, not active instructions.

## Verification

Focused tests must show that two different Architectural topics can start in one
session, including when old brainstorm state exists. Existing test-runner guards,
wiki hints, and patch handling must retain their behavior. Behavior evals cover
related and unrelated topics, resuming a topic, mixed context, and user choices
to stay or move. Check scratch and approval requirements remain intact and skill
files stay within their budgets.

No new command, setting, dependency, or automatic session routing is needed.
