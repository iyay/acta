---
parent: debt/2026-10-20-close-without-fast-suite
closes: [DBT-0136.02]
depth: minimal
id: PLN-0182
created: "2026-10-20 08:12:31"
hash: cjzy9ko
---

# Anchor eval markers to the workspace root

**Goal:** Make the build-close eval markers land at the workspace root whatever folder the suite runs from.
**Spec:** none (Bounded, approved in chat on 2026-10-20)
**Tests:** fast `scripts/test ./internal/plugincheck/ -run 'TestBuildCloseScaffoldMarkers'`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Test tooling only: no skill text, no CLI or runtime change, no new case or grader. Keep the scaffold's refusal of a non-empty folder.
- Keep the marker names the graders already read, `broad-suite-ran` and `typecheck-ran`.

## Waves

- Wave 1: Task 1.

### Task 1: Anchor the scaffold markers to the workspace root

**Files:** Modify `plugin/evals/build-close-no-suite/scaffold.sh`; add the regression test to `internal/plugincheck/evals_test.go`.
**Seam:** the generated `scripts/test` and `scripts/typecheck`, run with the scaffold's tools stubbed.
**verify:** Running either generated script from the workspace root or from a subfolder writes its marker at the workspace root and nowhere else, so the `file_exists` graders cannot miss a broad-suite run; the scaffold still refuses a non-empty folder. List the paths checked: root and one subfolder for both scripts.
- [ ] Red: Add `TestBuildCloseScaffoldMarkersIgnoreCwd`, which scaffolds in a temp folder with stub `acta`, `git` and `xcrun`, runs both scripts from a subfolder, and asserts the markers sit at the root; it fails because the current scripts write them in the caller folder.
- [ ] Change: Write both generated scripts with the scaffold's `$PWD` baked into the marker path, so the folder the scripts run from no longer matters.
- [ ] Commit: `fix(eval): anchor the build-close markers to the workspace root`
