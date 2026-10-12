---
parent: specs/2026-10-12-review-depth-option-design
depth: minimal
closes: [SCR-0061]
id: PLN-0135
created: "2026-10-12 07:03:08"
hash: xego2mn
started: "2026-10-12 07:05:58"
---
# Review depth option

**Goal:** A `review_depth` setting (`quick`, `standard`, `deep`; user, repo and per-run) picks which review passes `acta:review` runs, with floors that keep sensitive paths, fix rounds and dispatch work at `standard` or more.

**Spec:** `.acta/specs/2026-10-12-review-depth-option-design.md`

**Tests:** fast `scripts/test ./internal/config ./internal/cli ./internal/setup ./internal/plugincheck ./plugin`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Key name, verbatim: `review_depth`. Values, verbatim: `quick`, `standard`, `deep`. Empty means `standard`.
- Follow `plan_depth` everywhere it appears: same files, same error shape, same repo mark, same setup form style. Do not invent a new pattern.
- Three rounds at most at every level. No version bump.
- Skill text stays plain, short, and free of repo paths and user names (wiki `skill-text-fits-everyone`).
- Tests never run `rm -rf` or `rm -f`; temp files go under `t.TempDir()`.

## Waves

- Wave 1: Task 01, Task 02
- Wave 2: Task 03

### Task 01: `review_depth` setting

**Files:**
- Modify: `internal/config/user.go`, `internal/config/repo_user.go`, `internal/cli/config_cmd.go`, `internal/setup/form.go`
- Test: `internal/config/repo_user_test.go`, `internal/cli/config_repo_test.go`, `internal/setup/form_test.go`

**verify:** `review_depth` behaves like `plan_depth` on every path that reads or writes `plan_depth`. List each path checked: user file load and validate, repo `.acta.yaml` load and validate, repo beats user, `acta config show` text and JSON with and without the repo mark and the default mark, the repo set flag, the user set flag, the setup form default and save. Any value other than the three (and empty) is refused with an error naming `review_depth`.

- [x] **Failing test:** in `repo_user_test.go`, a repo `review_depth: deep` beats a user `quick` and a bad value fails naming the key; in `config_repo_test.go`, `acta config show` prints `review_depth: standard` with the default mark when unset and `deep` with the repo mark when the repo sets it, and the repo set flag writes it; in `form_test.go`, the form defaults to `standard` and saves the picked value. Run the three packages; they fail because the key does not exist.
- [x] **Change:** add `ReviewDepth string \`yaml:"review_depth,omitempty"\`` with a comment beside `PlanDepth` in `User`, trim and validate it next to `plan_depth`, add it to `RepoKeys` and each `plan_depth` case in `repo_user.go` with default `standard`, and mirror every `plan_depth` line in `config_cmd.go` (show text, show JSON, set flags) and `setup/form.go` (field, default, select with the three values, save).
- [x] **Run:** `scripts/test ./internal/config ./internal/cli ./internal/setup`, then `go vet ./internal/... && gofmt -l internal`; all pass, gofmt prints nothing.
- [x] **Commit:** `feat(config): review_depth setting`

### Task 02: review levels in the review skill

**Files:**
- Create: `plugin/skills/review/depth.md`, `plugin/skills/review/fact-check.md`
- Modify: `plugin/skills/review/SKILL.md`, `plugin/skills/review/code-reviewer.md`, `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/budget_test.go`, `plugin/embed.go` only if `plugin/embed_test.go` asks for the new files
- Test: `internal/plugincheck/skill_review_test.go`, `plugin/embed_test.go`

**verify:** No sentence in the review folder lets a review run below a floor or past three rounds. List every place that names how many reviewers run or how deep a review goes, and show each one agrees with `depth.md`: the level pick order, `quick` (light tier inline, else one combined reviewer), `standard` (today), `deep` (two reviewers then fact-check over BLOCKERs only, never on zero BLOCKERs), the three floors, and the review-notes line that names the level used and any floor that raised it.

- [x] **Failing test:** in `TestSkillReview`, add to `Must`: `review_depth`, `depth.md`, `fact-check.md`, `/acta:review quick|standard|deep`, `Fix rounds run at least \`standard\``; replace the `Must` line `Reading the diff yourself never replaces the two reviewers, however small the diff.` with `Reading the diff yourself never replaces the reviewers its depth names, however small the diff.`. Add a test that `depth.md` holds `quick`, `standard`, `deep`, the three floors, `never a round 4`; that `fact-check.md` holds both grounds, `In doubt, it stays a BLOCKER`, and the never-moved subjects (security, data loss, concurrency, behaviour change); and that `code-reviewer.md` holds `[EARLIER_FINDINGS]` and `every changed file gets its own pass`. Add byte caps for the two new files in `budget_test.go`. Run `scripts/test ./internal/plugincheck`; it fails.
- [x] **Change:** write `depth.md` from spec section 2 and 3, and `fact-check.md` as a read-only subagent prompt from spec section 2 (reasoning before the moved ids, each move names its ground and the line, first line states the model). In `SKILL.md`, add a short `## Review depth` section with the pick order and a link to `depth.md`, make `## Review tiers` say the full tier runs the reviewers the depth names, and keep `A fix round always takes the two reviewers.` In `code-reviewer.md`, add the coverage rule and an `## Earlier Findings` block with `[EARLIER_FINDINGS]` ("do not repeat these; look for others", empty in round 1) plus its placeholder line. Raise the `review/SKILL.md` and `code-reviewer.md` byte caps and the review `MaxLines` only as far as the new text needs.
- [x] **Run:** `scripts/test ./internal/plugincheck ./plugin`; pass.
- [x] **Commit:** `feat(review): review depth levels and fact-check pass`

### Task 03: eval for the sensitive-path floor

**Files:**
- Create: `plugin/evals/review-depth-floor/case.yaml`, `prompt.md`, `scaffold.sh`, `graders/spec-reviewer-dispatched.md`, `graders/standards-reviewer-dispatched.md`
- Test: `internal/plugincheck` eval layout tests

**verify:** The case passes only when both a Spec and a Standards reviewer subagent run on a diff that changes an auth check while `.acta.yaml` sets `review_depth: quick`; a run that does one combined review or an inline review fails at least one grader.

- [ ] **Failing test:** copy the shape of `plugin/evals/polish-full-review` (same scaffold guards, `./bin/acta` and `./bin/git`, `allowed_tools: [Skill, Read, Agent]`). The scaffold writes `.acta.yaml` with `review_depth: quick`, a one-task plan, and one commit that changes an auth check in `src/auth.py`. The two graders are `tool_used` on `Agent` with `input_match` `Spec` and `Standards`, each with a `# guards:` line naming `## Review depth`. Run `scripts/test ./internal/plugincheck`; it fails until the case layout is valid, then passes.
- [ ] **Run:** with the branch binary on PATH (`go build -o "$TMPDIR/acta" ./cmd/acta`, then PATH first), `claude plugin eval` for this one case only; both graders pass.
- [ ] **Commit:** `test(evals): review-depth-floor case`
