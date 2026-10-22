---
id: BUG-0054
hash: l0jvbfw
priority: low
---
# Build close sometimes reruns the fast suite to show test output

## Symptom
The eval case build-close-no-suite is flaky: about 2 of 5 runs fail no-fast-suite (broad-suite-ran exists). One main run failed typecheck-ran instead.

## Root cause
plugin/skills/build/SKILL.md, Close section: "finish any outstanding type check and show it with the tests each task ran". An agent reads "show it with the tests" as an order to run the fast suite so its output is shown. Trace quote: "I'm also running the fast tests so they're shown."

## Repro
scripts/eval --case build-close-no-suite --runs 3 --keep-temp, then read out/trace.jsonl of a red run.

## Found in
Land gate of PLN-0185 (split-plan-build-brief), 2026-10-22. The diff did not touch the Close section; the red run read only from "## Close" on.
