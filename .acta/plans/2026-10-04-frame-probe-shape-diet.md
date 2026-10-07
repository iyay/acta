---
parent: specs/2026-10-04-frame-probe-shape-diet-design
depth: minimal
id: PLN-0080
created: "2026-10-04 06:59:06"
hash: lt31wik
started: "2026-10-04 07:04:13"
finished: "2026-10-04 07:22:25"
---
# Frame, Probe and Shape Diet Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Add the optional `frame` skill, the `probe` question mode of shape with a `questions` user setting, and cut `shape/SKILL.md` from 2696 words to about 1300 with no loss of quality.

**Spec:** `.acta/specs/2026-10-04-frame-probe-shape-diet-design.md`

**Tests:** fast `scripts/test ./internal/<package>` (the package a task names), full `scripts/test --full`; land also runs `scripts/eval` with the branch binary on PATH.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Text adapted from gstack office-hours or mattpocock grilling is rewritten in our own words, never copied 1:1, and is shorter than its source. Sources: `/Users/iyay/Nayakatara/digimov/ai-eceran/.claude/skills/gstack/office-hours/` (SKILL.md, `sections/phase-2a-startup-diagnostic.md`, `sections/phase-2b-builder-brainstorm.md`) and `~/.claude/plugins/marketplaces/mattpocock/skills/productivity/grilling/SKILL.md`.
- Skill text is plain English with short words. No emoji. The word "office hours" never appears in frame.
- `frame` never sets scratch status `brainstorming`.
- Every byte cap change in `internal/plugincheck/budget_test.go` is on purpose: lower caps for files that shrink, add caps for new files, raise only the session-start cap and only by what the new text costs.
- No run of this plan sits beside SPC-0069 plans 2 or 3: all edit `budget_test.go`.

## Waves

- Wave 1: Task 1, Task 2
- Wave 2: Task 3
- Wave 3: Task 4
- Wave 4: Task 5, Task 6

### Task 1: `questions` user setting

**Files:** `internal/config/user.go`, `internal/config/user_test.go`, `internal/cli/config_cmd.go`, `internal/cli/config_repo_test.go`

**verify:** The only values a stored config can hold for `questions` are empty, `one` and `probe`. List every way a value gets in (file load, `acta config set --questions`, `--repo`) and what each does with a bad value or a repo attempt.

- [x] Failing tests: a config with `questions: maybe` is refused like a bad `plan_depth`; `acta config set --questions probe` stores it; `acta config show` prints `questions: one` with the default mark when unset; `--repo --questions` is refused because the key is user only. They fail because the field does not exist.
- [x] Add a `Questions` field (`yaml:"questions,omitempty"`) with its check and trim next to `PlanDepth`, and the `--questions` flag plus the show line in `config_cmd.go`; not in `RepoKeys`.
- [x] Commit: `config: questions setting picks one-at-a-time or probe`

### Task 2: Shape diet and probe mode

**Files:** `plugin/skills/shape/SKILL.md`, create `plugin/skills/shape/probe.md`, delete `plugin/skills/shape/spec-document-reviewer-prompt.md`, `internal/plugincheck/skill_shape_test.go`, `internal/plugincheck/budget_test.go`

**verify:** No rule the old shape enforced is lost: every Must string still holds, the Bounded "changes requested" step can only lead back to the short spec, every `acta scratch add` line in shape and probe has `--section log`, and the dot graph cannot come back. List each removed section and where its rule now lives, or why the model needs no text for it.

- [x] Failing tests: `TestShapeBoundedReviewLoopsToShortSpec` rewritten to read the Bounded checklist (changes requested goes back to the short spec, never to the design doc) and to fail when a `digraph` is present; a new test that shape names `probe.md` and `Questions: probe` and that `probe.md` has "five", "Recommended:" and `--section log`; `MaxLines` and the shape byte and description caps lowered to the new sizes. They fail on today's long file and the missing probe file.
- [x] Rewrite shape per spec section 3 (about 1300 words, description keeps "brainstorm", frame offer line, probe line that fires on "probe", "grill me" or the session note `Questions: probe`, and "one at a time" turns it off); write `probe.md` per spec section 2 (about 250 words); delete the reviewer prompt and its cap line; add the `shape/probe.md` cap.
- [x] Commit: `shape: cut to one flow, add probe mode, drop unused reviewer prompt`

### Task 3: Session note line and setup question

**Files:** `internal/hook/hook.go`, `internal/hook/hook_test.go`, `plugin/skills/setup/SKILL.md`, `internal/plugincheck/budget_test.go`

**verify:** The session-start text has `- Questions: probe.` exactly when the stored value is `probe`, on every voice path (set up, not set up, unreadable file). List each path and what it prints. The default config's session-start size does not change.

- [x] Failing tests: session start with `questions: probe` holds `- Questions: probe.`; with `one`, empty, no voice file and a broken voice file it does not. They fail because the hook does not read the field.
- [x] Write the line in the Voice block of `hook.go` after the Style line; add `questions` to the setup skill's "Change anything already set?" list and a short "Questions" part with `acta config set --questions probe`; raise only the setup cap by what the text costs.
- [x] Commit: `hook: tell shape when the user picked probe; setup asks for it`

### Task 4: The frame skill

**Files:** create `plugin/skills/frame/SKILL.md`, create `plugin/skills/frame/startup.md`, create `internal/plugincheck/skill_frame_test.go`, `internal/hook/hook.go`, `internal/plugincheck/budget_test.go`, `plugin/NOTICE`

**verify:** frame can never use up the session's one brainstorm and can never be hidden from the user: no line in frame sets status `brainstorming`, and its frontmatter never says `user-invocable: false`. List every place frame writes to the scratch item and the command it uses.

- [x] Failing test `skill_frame_test.go`: `CheckSkill` with Must strings for the goal question, `startup.md`, premises, "assignment", `acta scratch add` with `--section context`, `--section log` and `--section questions`, `acta:shape`, and MustNot `status brainstorming`, `office hours`, `user-invocable: false`; `startup.md` names the six forcing questions; the description has `frame` and "worth building". It fails because the skill does not exist.
- [x] Write both files per spec section 1 (about 900 and 700 words), add `frame` to `hook.Skills`, add caps for both files and the frame description, raise `sessionStartCap` by the bytes the new name adds, and credit gstack (Garry Tan) and mattpocock (Matt Pocock) in `plugin/NOTICE` with what each was adapted into.
- [x] Commit: `frame: optional skill for why, for whom and what before shape`

### Task 5: Eval cases for frame and probe

**Files:** create `plugin/evals/frame-no-brainstorming-status/` (case.yaml, prompt.md, scaffold.sh, graders/), create `plugin/evals/probe-round/` (same layout), `internal/plugincheck/evals_test.go`

**verify:** Each new case fails when its rule breaks: list for each grader the bad reply it rejects (frame: status set to brainstorming, a first question that is not the goal question; probe: more than five questions, a question with no recommended answer, any code written).

- [x] Failing test: `evals_test.go` rows for both cases, tied to `skills/frame/SKILL.md` and `skills/shape/probe.md`; it fails because the case folders do not exist.
- [x] Write both cases in the layout of `plugin/evals/answers-appended/`: frame gets a new product idea in a scaffolded repo with a scratch item; probe gets "grill me" on a short design.
- [x] Commit: `evals: frame keeps the brainstorm slot, probe asks one capped round`

### Task 6: Version bump

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The three manifests agree on one version that is one patch above main's at branch time, and no other field changed.

- [x] Failing test: none new; `internal/plugincheck` already fails when the three disagree, so bump one file first and watch it fail.
- [x] Add 1 to the patch in all three files.
- [x] Commit: `plugin: version <new version>`, with the real number (0.1.4 when main is still on 0.1.3).

## Review notes

- R1: TestFrameNeverSetsBrainstormingStatus skips lines with "never" or "no "; the MustNot `status brainstorming` still catches the command form.
- R1: the five-questions-max grader with the m flag fires on any numbered list that reaches 6, not only on a sixth question.
- R1: shape no longer says specced comes from the parent link; harmless, internal/write/ops.go refuses specced as a status.
- R1: `config show --json` prints one for both unset and set-to-one, the same as plan_depth.
- R1: the shape description lost its trigger list (new feature, fix that needs code, behaviour change, decision); the four old shape evals guard triggering at land.
- R1: the early "decompose a multi-subsystem request" guidance is gone; only the self-review scope check covers it, under the spec's drop of general design advice.
- R1: startup.md shares 41 of 740 seven-word runs with gstack and probe.md 27 of 277 with grilling; both are rewritten overall and shorter than their sources.
- R1: Task 2 landed as two commits with the same message (90a9640, b1d3ef2).
- R1: the out-of-plan edits to skill_setup_test.go and plugin/hooks/default-rules.md are needed (a Must string, TestDefaultRulesFile).
- R1: probe.md uses SCR-0001 as its placeholder id while frame uses SCR-xxxx.
- Polish: every-question-has-a-recommendation no longer checks a heading with italic in it (`**Q1. Use *Postgres* here?**`).
- Polish: the looser goal grader also passes a reply where "what" and "goal" come before a first question that is not the goal question.
- Polish: both new max_turns comments still open with "The floor is three turns", which is now stale, and 7 has no measured runs behind it; add turn counts after the first eval runs.
- Polish: without the fence, the probe.md heading and body lines render as one markdown paragraph; still readable.
- Polish: the first `--round polish` send was skipped by omp (plan-title goal looked done); recorded as BUG-0028 on main.
