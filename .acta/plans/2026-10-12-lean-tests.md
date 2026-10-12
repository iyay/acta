---
parent: specs/2026-10-12-lean-tests-design
depth: minimal
id: PLN-0136
created: "2026-10-12 08:08:00"
hash: baqoidu
started: "2026-10-12 08:10:19"
---
**Goal:** The tdd, slice and build skill text makes implementers write one test per case the plan names, at the layer that owns it, with upper layers testing wiring only and adversarial inputs kept to trust boundaries.

**Spec:** `.acta/specs/2026-10-12-lean-tests-design.md`

**Tests:** fast `scripts/test ./internal/plugincheck ./plugin`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Pinned phrases, verbatim, so skill text and tests agree: `test it once, at the layer that owns it`; `wiring only: one success path and one error passed up from below`; `only where the input crosses a trust boundary`; `Every case the plan names has a test, and no other`; `Missing validation the plan names`; `named once, in the task whose layer owns the behaviour`; `A case the plan does not name is not a test`; `Does every plan case have a test, and nothing beyond?`.
- Removed phrases, never back in any form: `Every new function/method has a test`, `Edge cases and errors covered`, `where the input can take them`, `Missing validation for zero, empty, nil`, `Are tests comprehensive?`.
- Skill text stays plain, short, and free of repo paths and user names (wiki `skill-text-fits-everyone`). Change only the lines the spec names; keep the rest of each file as it is.
- Raise a byte or line cap in `internal/plugincheck` only as far as the new text needs. No version bump.
- Tests never run `rm -rf` or `rm -f`; temp files go under `t.TempDir()`.
- Not a task: noting the test-to-code ratio of the next two plans; it goes in the landing report as the next action.

## Waves

- Wave 1: Task 01
- Wave 2: Task 02
- Wave 3: Task 03
- Wave 4: Task 04

### Task 01: tdd skill tests the plan's cases once

**Files:** Modify `plugin/skills/tdd/SKILL.md`, `plugin/skills/tdd/writing-good-tests.md`, `internal/plugincheck/skill_tdd_test.go`, `internal/plugincheck/budget_test.go`

**verify:** No sentence in the tdd folder asks for a test beyond the plan's cases or for the same case at a second layer, and the adversarial list applies only at trust boundaries. List every sentence in `SKILL.md` and `writing-good-tests.md` that tells the reader which tests to write (cover section, quality bar, checklist, mutation check, rationalization table) and show each agrees.

- [x] **Failing test:** in `TestSkillTDD`, add the tdd pinned phrases and `Missing validation the plan names` to `Must` (or a sibling test reading `writing-good-tests.md` if `CheckSkill` does not read it) and the tdd removed phrases to `MustNot`; `scripts/test ./internal/plugincheck -run 'TestSkillTDD|TestBudget'` fails on the missing phrases.
- [x] **Change:** rewrite `## Cover the plan, not the universe` (plan's cases, tested once at the owning layer, upper layers wiring only, adversarial list only at trust boundaries named as user input, files, network, auth, money, migrations, deletes), swap the two checklist lines for the one pinned line, and change the mutation item in `writing-good-tests.md`.
- [x] **Commit:** `feat(tdd): one test per plan case, at the layer that owns it`

### Task 02: slice names each case once

**Files:** Modify `plugin/skills/slice/SKILL.md`, `internal/plugincheck/skill_slice_test.go`, `internal/plugincheck/budget_test.go`

**verify:** A plan written from the slice skill cannot name one case in two tasks. List every place the skill describes a task's tests or `verify:` line (verify section, full task template, `minimal` task shape) and show each says the case is named once at the owning layer and upper layers are wiring only.

- [x] **Failing test:** add `named once, in the task whose layer owns the behaviour` and `wiring only: one success path and one error passed up from below` to `Must` in `TestSkillSlice`; `scripts/test ./internal/plugincheck -run 'TestSkillSlice|TestBudget'` fails.
- [x] **Change:** add the rule to `## Verify lines are properties` and to the `minimal` task shape sentence; the full template's `verify:` placeholder points at it.
- [x] **Commit:** `feat(slice): name each test case once, at the owning layer`

### Task 03: implementer writes only the plan's cases

**Files:** Modify `plugin/skills/build/implementer-prompt.md`, `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/budget_test.go`

**verify:** The implementer prompt never invites a test for a case the plan does not name. List every line of the prompt that mentions tests (work steps, self-review, report format) and show each agrees, and that an unnamed case goes to the report as one line.

- [x] **Failing test:** a test in `skill_build_test.go` that reads `implementer-prompt.md` on its own (like `TestBuildStartTickPrompt`) requires `A case the plan does not name is not a test` and `Does every plan case have a test, and nothing beyond?` and refuses `Are tests comprehensive?`; `scripts/test ./internal/plugincheck -run 'TestBuild|TestBudget'` fails.
- [x] **Change:** add the unnamed-case rule next to the `acta:tdd` work step (it goes in the report as one line, the orchestrator decides), add one report line for it, and swap the self-review question.
- [x] **Commit:** `feat(build): implementers write only the plan's test cases`

### Task 04: lean-tests eval

**Files:** Create `plugin/evals/lean-tests/case.yaml`, `prompt.md`, `scaffold.sh`, `graders/*.md`; Modify `internal/plugincheck/evals_test.go`

**verify:** The case fails a run that tests a lower-layer case again at the upper layer or adds a case the plan does not name, and passes a run that writes exactly the plan's cases. Name which grader rejects each of those two failures.

- [ ] **Failing test:** register `{"lean-tests", "skills/tdd/SKILL.md", "test it once, at the layer that owns it"}` in `evalCases`; `scripts/test ./internal/plugincheck -run TestEvalCases` fails until the folder is valid.
- [ ] **Change:** copy the shape of `plugin/evals/review-depth-floor` (scaffold guards, `./bin/acta`, `./bin/git`, voice file). The scaffold writes a Python repo with a minimal two-task plan: task 1 `src/age.py` `parse_age(text)` with three cases (a whole number returns it, a negative number raises `ValueError`, a non-number raises `ValueError`); task 2 `src/cli.py` `main(argv)` prints the age, wiring only. The prompt asks for both tasks with acta:tdd, no worktree, no review. Graders: one `llm` on `{ source: file, path: tests/test_cli.py }` (FAIL when it re-tests the negative or non-number rules beyond one passed-up error), one `llm` on `tests/test_age.py` (FAIL on any case beyond the three, like empty, unicode or huge), each with a `# guards:` line; allow `Skill, Read, Write, Edit, Bash`.
- [ ] **Run:** with the branch binary first on PATH (`go build -o "$TMPDIR/acta" ./cmd/acta`), `claude plugin eval` for this one case only; both graders pass.
- [ ] **Commit:** `test(evals): lean-tests case`

## State

### Next

Task 04: lean-tests eval case
