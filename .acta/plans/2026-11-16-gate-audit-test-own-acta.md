---
parent: specs/2026-11-16-gate-audit-test-own-acta
depth: minimal
id: PLN-0245
created: "2026-11-16 14:46:38"
hash: tyqv36a
---
# Gate audit test builds its own acta

**Goal:** `TestGateWorktreeOwnerAudit` passes on a machine with no acta on PATH, using an acta built from this checkout.

**Spec:** `.acta/specs/2026-11-16-gate-audit-test-own-acta.md` (SPC-0213)
**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`.
**Gates:** eval: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched; whole when the diff touches `plugin/hooks/` or `plugin/references/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Leave `plugin/evals/land-tidy-temp-worktree/scaffold.sh` and its guard unchanged.
- Build inline in a worktree; the push after release is approved by the user only for main and tag v0.1.56.

## Waves

- Wave 1: Task 1

### Task 1: Test supplies a freshly built acta on PATH

**Files:** `internal/plugincheck/gate_worktree_owner_test.go`
**Seam:** `go test ./internal/plugincheck -run TestGateWorktreeOwnerAudit`
**verify:** Every subtest's scaffold copies the acta built from this checkout, never an installed one, and no PATH the caller passes (with or without an acta on it, or an old acta first) changes which binary runs. List each PATH case tried.
- [ ] Red: the test run with a PATH that holds no acta fails because the scaffold runs `cp` with an empty source (`cp: cannot stat ''`).
- [ ] Change: build `./cmd/acta` once into a temp dir at the start of the test and put that dir first on PATH in the env that `run()` uses.
- [ ] Commit: test(plugincheck): build acta for the gate worktree audit
