---
parent: specs/2026-10-19-shape-unslop-design
depth: minimal
id: PLN-0173
created: "2026-10-19 13:48:49"
hash: qk3nub7
---
# Unslop for spec self-review plan

**Goal:** Apply Unslop during the existing Bounded and Architectural spec self-review without changing required content or workflow gates.
**Spec:** `.acta/specs/2026-10-19-shape-unslop-design.md`
**Tests:** `scripts/test` (fast); `scripts/test --full` (full, land only); task loop: `go test ./internal/plugincheck ./plugin`.

## Global Constraints

- Follow acta:lean: understand task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Implement inline in a worktree. Keep repository text in English. Never push.
- Limit behavior changes to shape. Preserve all four existing self-review checks and approval gates, with no extra editing pass.
- Keep stable Unslop rule numbers. Treat patterns as prompts to inspect meaning, not automatic errors. Preserve requirements, constraints, uncertainty, evidence, risks and literal text; never invent sources or measurements.
- Add no setting, invocable skill, dependency or automatic prose fixer. Do not run paid behavior evals. Static checks verify instructions and packaging, not model compliance.

## File map

- `plugin/references/unslop.md`: shared prose guidance adapted from the supplied Unslop rules, including content-preservation boundaries.
- `plugin/skills/shape/SKILL.md`: load that reference in the existing self-review and ensure both spec paths run it before user review.
- `internal/plugincheck/skill_shape_test.go`: guard both paths, a resolvable reference, stable rule ids and preservation boundaries.
- `internal/plugincheck/budget_test.go`: register the reference byte cap and adjust only affected caps as needed.

## Waves

Wave 1 contains Task 1. The reference, its callers and its checks are one deliverable.

### Task 1: Apply Unslop in shape self-review

**Files:** Create `plugin/references/unslop.md`; modify `plugin/skills/shape/SKILL.md`, `internal/plugincheck/skill_shape_test.go` and `internal/plugincheck/budget_test.go`.
**Seam:** Shape's Bounded and Architectural checklists, shared spec self-review, and the reference shipped through `plugin.Files`.
**verify:** Both spec paths reach the existing self-review and load the same shipped reference before requesting user review; the reference keeps all supplied rule ids and protects meaning and literal text; the four checks and approval gates remain; list the paths checked and distinguish instruction checks from observed model behavior.
- [ ] Red: Add focused plugin checks for both paths, reference resolution, stable rule ids and preservation boundaries; observe `go test ./internal/plugincheck -run 'TestShape.*Unslop'` fail because the reference and integration are absent.
- [ ] Change: Add the reference, integrate it into both existing self-review paths, update affected caps and pass `go test ./internal/plugincheck ./plugin`.
- [ ] Commit: `feat(shape): apply unslop during spec self-review` with the plan task trailer.

## Pilot after landing

The landing report asks for the next real spec to measure before expanding scope. Compare draft and edited word counts, concrete edits and semantic preservation as defined in the spec. This is a follow-up pilot, not a build task or automatic rollout.
