---
parent: debt/2026-10-20-close-without-fast-suite
closes: [DBT-0136.02]
depth: minimal
id: PLN-0182
created: "2026-10-20 08:12:31"
hash: cjzy9ko
started: "2026-10-20 08:13:04"
finished: "2026-10-20 08:15:14"
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
- [x] Red: Add `TestBuildCloseScaffoldMarkersIgnoreCwd`, which scaffolds in a temp folder with stub `acta`, `git` and `xcrun`, runs both scripts from a subfolder, and asserts the markers sit at the root; it fails because the current scripts write them in the caller folder.
- [x] Change: Write both generated scripts with the scaffold's `$PWD` baked into the marker path, so the folder the scripts run from no longer matters.
- [x] Commit: `fix(eval): anchor the build-close markers to the workspace root`

## Follow-up: Review note approved in chat

### Task 2: Keep workspace names literal in marker scripts

**Files:** Modify `plugin/evals/build-close-no-suite/scaffold.sh` and `internal/plugincheck/evals_test.go`.
**Seam:** The generated test and type-check scripts, through the existing scaffold regression.
**verify:** Both markers remain at the workspace root when scripts run from a subfolder and the workspace name contains spaces, dollar expansions, quotes or backticks; the path is data, never shell source, and no command hidden in the path executes.
- [x] Red: Extend the existing regression with a literal shell-character workspace name and check that neither marker escapes and no embedded command runs; observe the focused test fail on the current generated scripts.
- [x] Change: Keep the generated script heredocs literal and derive the marker location from each script's quoted path, then pass the focused scaffold checks and shell syntax check.
- [x] Commit: `fix(eval): keep marker paths literal in generated scripts`.

## State

### Next

review and land the marker fix

## Review notes

Deep review of `main..HEAD`: spec and standards reviewers were CLEAN with no BLOCKERs. Notes left: the test asserts only the subfolder run, which already proves the baked root path, so a separate root assertion would re-check the same string; unquoted heredocs would interpolate any future `$` in the generated scripts; a failed redirect exits 0 because the scripts end with `echo`; one test comment repeats "git". None changes what a grader sees.

Follow-up deep review of `3af895f..1fb940e`: Spec and Standards reviewers CLEAN, no new BLOCKERs or NOTEs. Seven workspace-name regressions and independent script probes confirm literal paths and no injected commands; the original shell-character cases fail before the fix. Focused checks, vet, shell syntax and diff checks passed.
