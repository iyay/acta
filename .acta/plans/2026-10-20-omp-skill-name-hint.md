---
parent: bugs/2026-10-20-omp-startup-skill-name
depth: minimal
id: PLN-0181
created: "2026-10-20 08:00:07"
hash: mf65kqs
started: "2026-10-20 08:02:07"
finished: "2026-10-20 08:04:17"
---
# Load all acta skills by their OMP names

**Goal:** Give OMP sessions one clear rule for loading every acta workflow skill through its bare-name Read URI.
**Spec:** `.acta/specs/2026-10-20-omp-skill-name-hint-design.md`
**Tests:** `bun test plugin/omp/index.test.ts` (focused regression), `scripts/test` (manual fast), `scripts/test --full` (full, land only); run the Bun checks again after final changes because they directly verify this extension.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- OMP uses Read with `skill://<bare-name>` for every acta workflow skill; `acta:build` maps to `skill://build`, never `skill://acta-build`, and the same rule covers references inside other skills and prompts.
- Add the note only in the OMP extension's session-rule context, including fallback; do not rename skills, add aliases, rewrite arbitrary input, change shared hooks or Claude Code/Codex instructions, or touch user-owned links.
- Retain event handling and prompt reminders; no new setting, CLI flag, dependency, or runtime skill-discovery code.
- Use the inline executor in a worktree, keep task TDD and normal landing gates, write repository content in English, and never push; label live OMP model behavior unverified unless actually exercised.

## Waves

- Wave 1: Task 1. One extension and its existing test seam fit one plan.

### Task 1: Explain the bare-name rule once per OMP session context

**Files:** Modify `plugin/omp/index.ts`, `plugin/omp/index.test.ts`, and `.acta/wiki/omp-skill-names.md`.
**Seam:** Exported `createState(...).contextFor(...)` and `reset(...)`, plus the existing default extension handler test harness in `index.test.ts`; read installed workflow skill frontmatter only in tests to confirm the names used by the mapping.
**verify:** Every acta workflow reference resolves by its bare frontmatter name under one generic Read rule, rather than a build-only exception; test the complete acta workflow name set and explicit examples beyond build, and check first-context delivery on success, missing/non-zero/throwing binary fallback, and compaction. Handler tests cover wiring only: startup, session-switch reasons new/resume/fork, and compaction deliver the note once with the existing rules; ordinary subsequent prompts contain only their existing reminder, and existing style, event source, session ID, and cwd behavior remains intact.
- [x] Red: Add mapping and once-per-context assertions at the existing state and handler seams, update exact context expectations to require the new note, and observe `bun test plugin/omp/index.test.ts` fail because the current extension never explains the Read URI or bare-name mapping.
- [x] Change: Insert one concise OMP-only instruction inside the existing unsent-rules branch after success/fallback selection, preserving style and reminder behavior without rewriting shared rules; finish the focused checks and document the runtime rule in the existing wiki page while retaining its user-link warning.
- [x] Commit: `fix(omp): load all acta skills by bare names`.

## Close

Run the focused Bun checks, the configured review, and normal full landing gates; do not add startup baseline tests. After landing, record the landed commit as `fixed_in` for BUG-0052. Restarting a real OMP session to check its first skill read is a user-visible follow-up unless the executor actually performs that probe.

## State

### Next

Review round 1 is CLEAN at a91c1e1 (deep, both axes). Next: land on main, then set fixed_in for BUG-0052 to the landed commit.

### Findings

main moved to c16e7a3 while this plan ran, so the review base is the merge base 93a27fc, not main's tip c16e7a3. Land merges cleanly; do not force it to a fast-forward.

## Review notes

Deep, no floor raised it. Round 1 covered 93a27fc..a91c1e1, both axes CLEAN with no BLOCKER; both reverted the change in a temp copy and saw the focused tests go red. Notes kept, no fix: the note names only the build example, backed by the generic clause and the all-installed-names test, because the spec names build alone; the skills-frontmatter guard would throw on a folder without a SKILL.md, a shape plugincheck already forbids; the State line above still named task 1 until this close.
