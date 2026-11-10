---
parent: specs/2026-11-10-cold-read-offer-by-weight-design
depth: minimal
id: PLN-0226
created: "2026-11-10 19:32:35"
hash: hsgbnmd
started: "2026-11-10 19:34:23"
finished: "2026-11-10 19:36:29"
---
# Offer the cold read only when the spec is heavy

**Goal:** Shape offers the cold read for every Architectural spec, and for a Bounded spec only when it shows a concrete signal; a Bounded spec with no signal goes straight to the user review gate without a word.

**Spec:** `.acta/specs/2026-11-10-cold-read-offer-by-weight-design.md` (SPC-0198)

**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- `plugin/skills/shape/cold-read.md` and `plugin/evals/cold-read-skip/` do not change.
- The `## Cold read` section keeps the phrases "Want a cold read before planning?" and "Skip goes to final spec approval", which existing evals anchor on.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: Cold read offer follows the signals

**Files:** `plugin/skills/shape/SKILL.md`, `internal/plugincheck/skill_shape_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** `TestShapeColdRead` over the text of `plugin/skills/shape/SKILL.md`
**verify:** No reader of the `## Cold read` section can offer a cold read on a Bounded spec with no signal, or skip it on an Architectural spec. List each rule checked in the section: Spike never; Architectural always; Bounded only on one of the three signals (trust boundary, auth, money or data migration; a changed interface, file format or command others depend on; files in more than one area); the silent skip with no hint line; signals judged once at step 7; a cold read the user asks for runs whatever the path. Step 7 still sits between self-review and user review in both checklists, and the file stays under its byte cap.
- [x] Red: `TestShapeColdRead` fails because the `## Cold read` section does not name the Bounded signals, the silent skip or the once-at-step-7 rule.
- [x] Change: rewrite the `## Cold read` section and Bounded step 7 in shape `SKILL.md`, and raise the `skills/shape/SKILL.md` cap in `budget_test.go` by about 400 bytes.
- [x] Commit: feat(shape): offer the cold read only for heavy specs

### Task 2: Eval for a Bounded spec with no signal

**Files:** `plugin/evals/cold-read-no-signal/` (new: `case.yaml`, `prompt.md`, `scaffold.sh`, `graders/behavior.md`), `internal/plugincheck/evals_test.go`, `plugin/evals/cold-read-offer/graders/behavior.md`
**Seam:** the `evalCases` table and the LLM-grader allowlist in `evals_test.go`, and the eval case itself
**verify:** The no-signal case passes only when the agent, resuming shape after self-review on a one-area Bounded spec with no sensitive or interface change, asks for spec review without any cold read offer or hint and never loads `cold-read.md`. The `cold-read-offer` grader says its export spec carries a sensitive signal, so the offer must still fire there. `evals_test.go` anchors the new case on a phrase that exists in shape `SKILL.md`.
- [x] Red: `evals_test.go` fails because its new `cold-read-no-signal` row points at a case folder that does not exist yet.
- [x] Change: add the case folder, modelled on `cold-read-offer`, with a spec that changes one help string in one Go file, and add the signal line to the `cold-read-offer` grader.
- [x] Commit: test(evals): cover the cold read skip on a spec with no signal

## Polish

### Task 3: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] [fix] plugin/skills/shape/SKILL.md Bounded step 7 wording ("only on a signal below") has no test; add one check in TestShapeColdRead.
- [x] Commit: `polish: review notes for PLN-0226`

## Review notes

- Review level deep (from review_depth), no floor raised it. Round 1 CLEAN: BLOCKER 0, NOTE 9, pre-existing 0 (a git templates warning in the scaffold was judged harmless, not a defect). One [fix] went to Polish task 3; its review was CLEAN. Open [note]s: cap 8270 has no slack; "more than one area" is loose; the no-signal eval -h string could read as a command; bare substring checks; long Cold read paragraph; Architectural step 7 has no matching wording check.

## State

### Next

All tasks committed; next is the close: wiki check, then acta:review over main..HEAD.
