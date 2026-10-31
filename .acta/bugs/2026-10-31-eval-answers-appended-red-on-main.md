---
id: BUG-0063
hash: ogbvj4z
priority: medium
---
# Eval answers-appended is red on main

## Symptom
At the PLN-0190 land on 2026-10-23 the eval case answers-appended scored 0.50 on the branch and on main: the answer-in-body grader did not find the answer in .acta/scratch/eval-answer.md. It is a real red, not a flake: it fails on main too.

## Root cause
Not known yet. Next step from SCR-0079: run the case with --keep-temp on main and read trace.jsonl before calling it a skill gap.

## Repro
scripts/eval --case answers-appended on main.

## Found in
main, at the PLN-0190 land (SCR-0079). The other main red there, build-close-no-suite, is already tracked as DBT-0139.01.
