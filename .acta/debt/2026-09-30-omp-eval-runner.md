---
id: DBT-0047
hash: altff5p
parent: plans/2026-09-30-omp-eval-runner
---
# omp eval runner review notes

- [ ] internal/evalomp/run_test.go TestRunCaseScaffold: the test scaffold writes $HOME/.acta/config.yaml with no t.Setenv("HOME", t.TempDir()); a mutant without the HOME override wiped a real user config during review. Isolate HOME in the test.
- [ ] internal/evalomp/run.go: a bad --case glob drops the path.Match error; every case is skipped and the run exits 0 with 0 passed.
- [ ] internal/evalomp/run.go: a --case that matches no case exits 0 with 0 passed; a typo in a gate looks green.
- [ ] internal/evalomp/grade.go: regex flags other than i (m, s) are dropped silently.
- [ ] internal/evalomp/grade.go: file_exists drops the path.Match error; a malformed glob or empty path with exists false passes; no ** support.
- [ ] internal/evalomp/grade.go: the llm judge reads only the file body; a criteria frontmatter key would give an empty rubric.
- [ ] internal/evalomp/run.go: the scaffold bash run has no timeout; a hanging scaffold stalls the run.
- [ ] internal/evalomp/run.go: on timeout only omp is killed, not its process group; children can linger.
- [ ] internal/evalomp/run.go: OmpJudge has no WaitDelay and drops stderr on failure.
- [ ] internal/evalomp/run.go: Ctrl-C during RunAll leaves acta-eval-omp-* folders behind.
- [ ] scripts/eval --omp always appends plugin after "$@", so passing your own plugin dir is bad usage.
- [ ] Spec departures carried by the plan: report.go folded into RunAll, probe facts in task 5, judge adds --no-extensions --no-rules --no-skills, --case is a glob.
- [ ] Real scaffolds put acta in ./bin through .claude/settings.json, which omp ignores; omp runs use the acta on PATH.
