---
id: DEBT-11
hash: rafb
parent: plans/2026-09-28-skill-rules
---
# Review NOTEs: Skill Rules and First-Run Setup Implementation Plan

- [ ] internal/doctor/doctor.go hasLine copies hook/gitignore.go hasLine; export and reuse the hook one.
- [ ] internal/cli/doctor.go linkedKnownFile uses the raw Readlink target, so a relative omp link misses the clash list and conflicts prints a false ok.
- [ ] internal/cli/doctor.go drops the voice.Resolve error; a broken voice file shows as "no default build executor" instead of the parse error.
- [ ] internal/doctor/doctor.go agents-view reads Home/.claude.json and ignores CLAUDE_CONFIG_DIR, unlike the harness check.
- [ ] internal/doctor/doctor.go stale-links scans only the top level of ~/.omp/plugins/node_modules and skips @scope folders.
- [ ] internal/doctor/doctor.go harness fix hint "omp plugin link <target>" re-links the gone path it just reported; plan says path to acta/plugin.
- [ ] internal/hook/hook.go first-run text says "run /acta:setup", which is "setup" in omp; old text worked in both harnesses.
- [ ] internal/cli/doctor.go has a loose comment block about --known between the const and cmdDoctor.
- [ ] plugin/skills/setup/SKILL.md does not say where the acta block goes the first time, when no markers exist.
- [ ] internal/cli/doctor.go plain acta doctor in Claude Code finds no clash list (only the omp link is searched), so conflicts says ok even with superpowers on.
- [ ] internal/cli/cli_test.go lacks --executor tests for subagent, inline and the empty value.
- [ ] plugin/skills/build/implementer-prompt.md changed outside the Task 6 file map and reads as if omp agent="task" applies only with split.
- [ ] plugin/skills/brainstorm/SKILL.md Step 0 says "file the user's words" without "verbatim" from spec 4.2.
- [ ] plugin/skills/review/SKILL.md models paragraph sits between "both on your own model (never lower):" and the list that colon introduces.
- [ ] plugin/README.md First run section lists only language, style and tone, not executor, subagent models or the CLAUDE.md block.
- [ ] internal/cli/doctor.go drops the gitc.IsDirty error and treats the file as clean; write/ops.go stops on that error.
- [ ] internal/cli/doctor.go skipReason copies gitc.Commit's dirty rule and reason text instead of calling gitc.Commit with the dirty flag.
- [ ] internal/doctor/doctor.go root-outside hint always says "fix root in .acta.yaml" even when the root came from ACTA_ROOT or PM_ROOT.
- [ ] internal/cli/doctor_test.go dirty test covers only an untracked .gitignore, not tracked-modified or staged.
- [ ] internal/cli/doctor_test.go commitCount parses digits by hand; strconv.Atoi would fail loudly.
- [ ] When .acta.yaml fails to parse, RepoRoot stays empty and the harness and conflicts checks skip the project settings.
- [ ] internal/doctor/doctor.go on macOS a root spelled in different letter case than git reports gives a false "outside the repo" fail.
- [ ] internal/doctor/doctor.go realPath falls back to the parent on any EvalSymlinks error (ELOOP, EACCES), so a self-loop .gitignore gets a --fix hint that then fails.
- [ ] A dangling .acta link or a .acta/.gitignore folder makes doctor suggest --fix, which then exits 3 with a raw error and no report.
- [ ] Commit failure reason from gitc is empty; git stderr is lost (older code).
- [ ] internal/doctor/doctor.go isLink comment says "the same two words", which is unclear.
- [ ] internal/doctor/doctor.go link fail message says "not a file of this repo" even when the link points inside the repo.
- [ ] internal/doctor/doctor.go checkRepo joins ActaRoot/.gitignore again though gi is already set.
- [ ] Unreadable .acta (mode 000) makes checkRepo say "no .gitignore" and suggest --fix, which fails with permission denied.
- [ ] root: .git passes the bounds check; --fix writes .git/.gitignore and reports ok.
- [ ] .acta symlinked to another in-repo folder: --fix appends to that tracked .gitignore, the commit fails, the report says ok.
- [ ] A hard-linked .gitignore or a link swapped between Lstat and write is not caught; needs local access.
- [ ] Tests cover the in-repo .gitignore link with docs/notes.txt, not .git/config.
