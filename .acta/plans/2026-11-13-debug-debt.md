---
parent: debt/2026-11-13-debug-evidence
depth: minimal
closes: [DBT-0167.01, DBT-0167.02, DBT-0167.03]
after: [PLN-0239]
id: PLN-0240
created: "2026-11-13 06:13:10"
hash: wjfpdle
started: "2026-11-14 05:50:46"
finished: "2026-11-14 06:09:32"
---
# Debug uses private probes of the right checkout

**Goal:** Debug probes copy the checkout being investigated and keep every pre-proof write within private probe storage.

**Spec:** `.acta/specs/2026-11-13-workflow-debt.md` (SPC-0208)
**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`.
**Gates:** eval: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched; whole when the diff touches `plugin/hooks/` or `plugin/references/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Build inline, serially, in a worktree under `~/.wwt/acta/`; never push.
- Keep source text generic across repos and harnesses; adjust byte caps only where needed, without cutting safety instructions.
- Validate changed graders with passing and failing samples before scoring a skill; use `evalomp.LoadCases` and `evalomp.Grade` in plugincheck for fixed graders, and measure judge rubrics on planted traces with the existing eval runner.
- For every skill-text task, run each named case once per call with `scripts/eval --case <name> --runs 10 -j 8` on old and new text under identical corrected graders and scaffolds; use an isolated copy for the old arm and the branch binary on PATH, never replace the installed binary.
- Record raw results with `acta state set`; isolate and exclude any regressing text change and keep its debt open, removing that item from `closes:` before land.
- Full suites run only at land; focused checks run after each change. Run scaffolds and mutation checks with fresh HOME and TMPDIR outside the repo.

## Build brief

- Land order: start only after PLN-0239 lands on main; these serial plans share plugincheck and eval files.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: Probe the checkout under investigation

**Files:** `plugin/skills/debug/SKILL.md`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/evals_test.go`, `internal/plugincheck/budget_test.go`, new `plugin/evals/debug-worktree-source/{prompt.md,case.yaml,scaffold.sh,graders/behavior.md}`, `.acta/wiki/skill-escape-lines-skip-proof.md`
**Seam:** debug control instructions and the `debug-worktree-source` eval with different main/worktree contents.
**verify:** Every copied control uses the absolute root of the tree under investigation, captured before changing directories; invocation from main and later shell resets cannot silently change it. A main-only copy cannot prove a worktree defect. The evidence names the tested checkout and shows the symptom there plus a control with the suspected cause removed. No $OLDPWD source remains in the control recipe.
- [x] Red: replace the old required-copy assertion with `TestDebugControlSource` and add the differing-checkout case, then run `scripts/test ./internal/plugincheck -run 'TestDebugControlSource|TestDebugProof|TestEvalCases'` and the new case against old text.
- [x] Change: capture and use an explicit absolute source root throughout the control recipe, update focused rules and wiki facts, then compare `debug-worktree-source` and `debug-independent-check` for 10 runs per arm.
- [x] Commit: fix(debug): copy controls from the checkout under investigation

### Task 2: Cover every pre-proof write with private probe storage

**Files:** `plugin/skills/debug/SKILL.md`, `plugin/skills/debug/root-cause-tracing.md`, `plugin/skills/debug/race.md`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/evals_test.go`, `internal/plugincheck/budget_test.go`, `plugin/evals/debug-independent-check/graders/no-repo-write.md`, new `plugin/evals/debug-probe-isolation/{prompt.md,case.yaml,scaffold.sh,graders/behavior.md}`
**Seam:** debug write-boundary instructions and trace/file evidence from `debug-probe-isolation`.
**verify:** Tests with caches, polluter hunts, profiler outputs and file-tool writes obey the same pre-proof boundary; only explicit metadata exceptions permit repo planning-file writes, never source edits. Probe storage is newly created with restrictive permissions and exclusive creation: an unset TMPDIR, pre-existing name, symlink, failed creation or shell reset cannot cause writes in the source tree or an old/shared folder. Each shell write checks its folder in the same command; file-tool writes name verified absolute probe paths. The existing no-repo-write grader no longer excuses cache writes in the source checkout.
- [x] Red: add `TestDebugProbeIsolation` and planted write/collision scenarios, then run `scripts/test ./internal/plugincheck -run 'TestDebugProbeIsolation|TestDebugProbeFolder|TestDebugRerunsKnownFlake|TestEvalCases'` and the new case on current text to expose unguarded outputs and shared probe reuse.
- [x] Change: use exclusive private creation supported by the eval sandbox, carry checked paths through all referenced recipes and file tools, state bug/flake metadata exceptions precisely, update the isolation grader, then compare `debug-probe-isolation`, `debug-independent-check`, `debug-order-dependent` and `debug-performance` for 10 runs per arm.
- [x] Commit: fix(debug): isolate all pre-proof writes in private probe folders

## Fix round 1

### Task 3: Keep context gathering separate from diagnostic writers

**verify:** No context-gathering or other-caller command runs app, imports, tests or profiling from the source checkout; native isolation verification grants its declared Write/Edit tools and preserves comparable old/new evidence

- [x] Gate finding: initial eval --tag=debug passed9/10, isolation failed; one isolated retry also failed. Actual raw e-bEXxCS first context command appends source checker before private copy. Candidate2 comparisons showed remaining source execution on another caller too. Keep the existing approved private-copy design; clarify the entry to Phase1 and the existing Phase3 caller check, without new product behavior.
- [x] Files: plugin/skills/debug/SKILL.md, internal/plugincheck/skill_debug_test.go, internal/plugincheck/budget_test.go. First add a failing contract assertion for the context/other-caller gate; watch RED, change only those instructions/cap, then focused/package GREEN. Native verification also supplies --allow-tools Bash Read Write Edit for the isolation case, because scripts/eval's default grant omitted its declared file tools. No runner/CLI/source option changes.
- [x] Keep all earlier results. Run corrected granted isolation10old/10new, and candidate3 independent/order/performance10each under the same existing corrected graders; compare against retained identical old arms. Do not select unchanged reruns or weaken graders. Rebuild matching private binary, commit with Task: PLN-wjfpdle#3, tick/readback, review fix range, then rerun fullsuite and selected gate after all fixes. No push or attachments.
- [x] Commit with the trailer `Task: PLN-wjfpdle#3`

## Review notes

- Review: auto selected standard (full tier: agent instructions); Round 1 CLEAN; BLOCKER 0, NOTE 0, pre-existing 0. Both complete Coverage reports reviewed6d14eae296076a8020169b860f78ef8995e345c3..582fec25ed19c5850c2e681fe1843089324e93df. Spec independentfocused0.049s; Standards0.052s and retained newtests RED0.008s with allthree originalinstruction files restored in privateclone/freshHOME/TMPDIR. No polish. Both axes verified identicalcorrectedgraders and disclosed residualnative failures, no universalcomplianceclaim. Worktrees main+thisbuild only; laterchore(plan)state commit belongsorchestrator, no unexpectedsourcecommits.
- Review: auto selected standard (full tier: agent instructions); Round 2 CLEAN; BLOCKER 0, NOTE 0, pre-existing 0. Cumulative totals remain0/0/0. Both completeCoverage reports cover582fec25ed19c5850c2e681fe1843089324e93df..bacd77c60b2ee80c771da051e8273c1b2243dfeb plus directinstructioncallers; round1range alreadycomplete. Bothaxesindependentfocused0.036s PASS; Standards retainsnewtest/restoresfixBASEskill and observesRED0.007s for context/othercaller clauses. No polish ornewfindings. Onlyorchestratorchore(plan)state commit afterreviewHEAD; worktreesmain+thisbuild only. Finalcandidate3 native grantedisolation0/10->4/10, independent3/10->10/10, order/perf10/10both; residualisolationviolations disclosed, no universalclaim.

## State

### Next

FinalpostTask3 gatesGREEN: fullsuite28packages0flakes exit0, metadata full-run-final.json and raw .git/acta-baseline-20261113T232448Z-85090.log, PASSbaselinepublished. Native eval --allow-tools Bash Read Write Edit --tag=debug -j8 --keep-temp10/10 first-run noflakes reporttmp.AkaebXUP30 logland-final-debug-gate.log. EarlierpreTask3gatefailure is resolved by changedsource/propergrant, not an unchangedpassingretry. Round2CLEAN completeCoverage, no notes/polish, plan3/3done. All enduringlessonsinwiki, benchmarkresidualsdisclosed.
Next tidy main debug-debt nofold (upstreamrangecontainsfeatures), comparetrees, parent1976544 clean, finalplan/per-taskancestorchecks beforecleanup. Then closeDBT0167threeitems and report10/14originalitems/4of5groups, continueapprovedPLN0241. No more sourcechanges/gatereruns unless treecomparisonneedsit. No push/attachments.

### Findings

Tasks1e2959d3,2c0c140d done. Round1 completeCLEAN bothaxes0notes0preexisting over1976544..582fec2; focusedreview0.049/0.052s and actualrevertRED0.008s. Firstfullsuite28packages0flakes PASS tested48dedc75790a9f4d1e78c8aa8184fa2426af932f tree34d2a638f602edffb78d47a338c37435cabebfb7 log.git/acta-baseline-20261113T231340Z-23747.log. PASSbaselinepublished, but laterTask3requiresfreshsuite.
Initial eval--tag=debug9/10 then isolationretry1failed: reports tmp.i5pCP5ei6r/tmp.VTAoEcwxJ0. Actualcontextcommandappendedsourcechecker, plus Write/Editnotgrantedwarning; no merge orflakerecord. Correctedoperatorgrant --allow-tools Bash Read Write Edit now available, nativecaseallowlistslimittools, no runner/CLI/vendorchanges. Earlierisolationmetricsretainedas shell-onlyavailability, notfiletoolcoverageproof.
Task3 committed after scoped RED0.007s->GREEN0.072s/package14.594s; no furthertexttuning. Phase1 entry forbidsappendingruntime tocontextgathering; Phase3 othercallerprobesprivateonly. Matchingprivatebinary bin-candidate3/acta. Actualcandidate3independent10/10 vsoldstrict3/10, reporttmp.Pwj9SclsWR; order10/10 tmp.mcFfewq1hY and performance10/10 tmp.clyqDjZrn4 vsold10each. Grantedisolationold0/10 tmp.mHZ5zSfN2w ->new4/10 tmp.NIo39YV4dR. Samecorrectedgraders/fixtures, no case regressed. Remainingisolationfailures include multilinesteps whose laterwrites aren'tguardedafterpossiblefailedcd; actual e-Y0ol2S writesprivateinobservedrun but recipewouldcontinueonfailure. Do not claim perfectcompliance orweakencriteria.
Earliercandidates preserved: Task2oldstrictiso0/10,ind3/10,order/perf10/10; candidate1iso1/10,ind3/10; candidate2iso3/10,ind7/10,others10/10. Task1worktree10/10both,ind9/10both underoldcacheexception. Calibration5/5control+8/8isolation expected. Actualold source4pycfiles despiteenv recordedinwiki; causeofenvbehaviorunproven. Allrawlogs/privatecopies /Users/iyay/.wwt/acta/debug-debt-evidence-20261114. No benchmarkflakes/rerunselection; no push/attachments.
