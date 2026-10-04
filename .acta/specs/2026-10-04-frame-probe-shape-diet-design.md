---
parent: scratch/2026-10-03-frame-and-probe
id: SPC-0070
created: "2026-10-04 06:51:08"
hash: wz6nf5o
started: "2026-10-04 07:04:13"
finished: "2026-10-04 07:22:25"
---
# Frame skill, probe mode and shape diet

Status: design approved by the user in chat on 2026-10-04, section by section. Architectural: it adds a skill, adds a mode to shape, and rewrites the skill every design starts with. Rulings and answers live in the SCR-0037 Log.

Step 3 of the 2026-10-03 roadmap. Boundary: `frame` answers why, for whom and what. `shape` answers how.

Every adapted text is rewritten, never copied 1:1, and must cost fewer tokens with no loss of quality. Sources, both MIT:

- gstack `office-hours` (Garry Tan), gstack commit 06ed920: SKILL.md 11871 words, `sections/phase-2a-startup-diagnostic.md` 2135 words, `sections/phase-2b-builder-brainstorm.md` 524 words, `sections/design-and-handoff.md` 5437 words.
- mattpocock `grilling` (Matt Pocock): SKILL.md 319 words.

## 1. The `frame` skill

Optional. It runs only when the user asks for it. `shape` offers it in one line when the work is a new product and the scratch item has no Context. It is never forced.

Files:

- `plugin/skills/frame/SKILL.md`, about 900 words.
- `plugin/skills/frame/startup.md`, about 700 words: the six forcing questions (demand reality, status quo, desperate specificity, narrowest wedge, observation and surprise, future-fit) and how to push back on vague answers. Read only in startup mode.

The user can call it straight: `/acta:frame` in Claude Code, the skill name `frame` in omp. No acta skill sets `user-invocable: false`, so this works once the file exists. A plugincheck test keeps frame's frontmatter free of `user-invocable: false`.

Triggers in the description: frame, frame this idea, is this worth building, should I build this, new product idea, validate an idea, who is this for. The words "office hours" are not used. A feature for a product that already exists stays with `shape`.

Flow:

1. Scratch item. Use the one the user names, else file the user's words with `acta scratch new`. Never set status `brainstorming`: the hook counts that status as the session's one brainstorm.
2. Goal question. A startup or work inside a company goes to startup mode, which reads `startup.md`. A hackathon, learning, open source, research or fun goes to builder mode. When customers or revenue come up in builder mode, switch to startup mode.
3. Questions one at a time, because each push-back builds on the last answer. Skip what the user's prompt already answered. "Just do it" or a full plan jumps to step 4.
4. Premise challenge: is this the right problem, what happens if nothing is built, what already exists (in the repo or the market), and for a new artifact, how users get it. Output numbered premises; the user agrees or disagrees with each.
5. Two or three product directions (what, for whom, why): the narrowest, the ideal and one from a different angle. Recommend one; the user picks. No technical approaches: those belong to shape.
6. One assignment: a concrete action in the real world, such as talking to three possible users. Never "go build it".
7. Write the result with `acta scratch add`: Context (who, the problem, the agreed premises, the chosen direction), Log (each answer), Open questions.
8. Ask "continue with shape SCR-xxxx now?". On yes, load `acta:shape` in the same session. Frame does not use the one-brainstorm slot.

Dropped from gstack: the shared preamble, telemetry, gbrain, web search, the codex second opinion, the founder profile, mockups, the design doc and its hand-off, and the AskUserQuestion format rules.

Also: `frame` joins `hook.Skills` in `internal/hook/hook.go`, and `plugin/NOTICE` credits gstack and mattpocock.

## 2. The `probe` mode of shape

File: `plugin/skills/shape/probe.md`, about 250 words. `shape/SKILL.md` gets one line: when the user says "probe" or "grill me", read `probe.md` and use it in place of asking one question at a time. Probe works on the Bounded and Architectural paths.

Content:

1. Map the decisions as a tree. The frontier is every decision whose prerequisites are settled.
2. One round asks the frontier, five questions at most. The questions that unblock the most go first; the rest wait for the next round. A question that depends on one still open in this round waits too.
3. Each question: `**Q1. title**`, the body, then `Recommended: ...`. No emoji.
4. Facts are looked up by subagents, under shape's subagent model rule. Do not wait on them: only questions that depend on the fact wait. Decisions belong to the user.
5. On the Architectural path, append each round's answers with `acta scratch add SCR-0001 --section log`.
6. Done when the frontier is empty and the user confirms a shared understanding. Then shape goes on to approaches or the design.

Dropped from grilling: the emoji format and the rules between questions.

### The `questions` setting

Probe can be the default for a user. A new user key `questions` in `~/.acta/config.yaml` takes `one` (the default) or `probe`. It is a user key only, not a repo key: how a person likes to be asked is their taste, not a repo rule.

- `internal/config/user.go`: the field and its check. Any other value is an error, like `plan_depth`.
- `internal/cli/config_cmd.go`: `acta config show` prints it with the default mark, and `acta config set --questions probe|one` writes it.
- `internal/hook`: the session-start Voice block gets the line `- Questions: probe.` only when the value is `probe`. The default adds no text.
- `plugin/skills/setup/SKILL.md`: setup asks for it with the other settings.
- `shape/SKILL.md`: the probe line also fires when the session note says `Questions: probe`.

The user can still switch inside a session. "probe" or "grill me" turns probe on. "one at a time" turns it off.

Out of scope: grilling a spec or plan that already exists, outside shape.

## 3. Shape diet

Target: `shape/SKILL.md` from 2696 words (17458 bytes) to about 1300 words. One file, one flow: the checklist per path becomes the only description of the flow.

Kept:

1. Description, shorter. It keeps the word "brainstorm".
2. A short HARD-GATE.
3. The three paths, the "take the heavier path" rule, the one-way ratchet, and a written spec for sensitive changes.
4. One line offering `frame`, one line pointing to `probe.md`.
5. A checklist per path. The Bounded checklist says that "changes requested" goes back to the short spec, never to the Architectural design doc.
6. Architectural only: scratch step 0, Log appends, one Architectural brainstorm per session with the a/b/c choices, the subagent model rule.
7. Spec rules: path, `acta id`, commit on the checked-out branch, `parent:` and `closes:`, debt items, `CONTEXT.md` and ADRs, a self-review of four checks, the user review gate, then `acta:slice`.

Dropped: the dot graph, Red Flags, the Anti-Pattern section, and general design advice the model already knows. `plugin/skills/shape/spec-document-reviewer-prompt.md` is deleted: no skill links it.

Tests:

- `skill_shape_test.go`: the Must list stays, `MaxLines` goes down. `TestShapeBoundedReviewLoopsToShortSpec` now checks the Bounded checklist instead of the graph, so the old bug (Bounded changes led into the Architectural design doc) stays guarded.
- `budget_test.go`: the shape cap goes down to its new size, the reviewer prompt line goes, and caps for `frame/SKILL.md`, `frame/startup.md` and `shape/probe.md` are added.

## Proof of same quality

1. Word counts before and after for shape, frame (against office-hours) and probe (against grilling).
2. The four shape eval cases stay green: `second-brainstorm-choices`, `one-file-fix-no-brainstorm`, `brainstorm-files-scratch-first`, `answers-appended`.
3. New eval `frame-no-brainstorming-status`: a new product idea; the first question is the goal question and the scratch item's status stays `raw`.
4. New eval `probe-round`: "grill me" on a design; the first reply has at most five numbered questions, each with a recommended answer, and no code.
5. Go tests for the `questions` key: a bad value is refused, `config show` marks the default, and the session-start text has the `Questions: probe.` line only for `probe`.
6. `scripts/test --full` and `scripts/eval` green, with the branch binary on PATH.

## Order with SPC-0069

SPC-0069 plans 2 and 3 also edit `internal/plugincheck/budget_test.go`, and plan 3 rewrites skill descriptions. This plan does not run side by side with them. Whichever lands second keeps the shorter shape description, as long as it still has the word "brainstorm". If this plan lands first, plan 3 counts `frame` as one more description to trim.

The plan ends by adding 1 to the patch version in the three manifests.
