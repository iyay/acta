---
parent: specs/2026-11-12-cold-read-evidence
depth: minimal
after: [PLN-0233]
id: PLN-0234
created: "2026-11-12 23:06:39"
hash: ymbz46a
---
# Cold read findings carry their evidence

**Goal:** A cold read checks the spec's claims against the repo, shows what it read, and reaches the user with no finding dropped, with two eval cases that measure it.

**Spec:** `.acta/specs/2026-11-12-cold-read-evidence.md` (SPC-0205)

**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`; evals `scripts/eval --case <name> --runs 10 -j 8`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Task 1 runs its two cases and `cold-read-accept` on `cold-read.md` as it is before this plan, 10 runs each, and records the pass counts with `acta state set <plan id> findings`. That is the old arm for every later task.
- Tasks 2 and 3 each run those three cases, 10 runs, after the change. A change that scores below the old arm on any of them is taken out before the commit, and `acta state set <plan id> findings` says which. A task whose whole change is taken out ends with its boxes ticked and no code commit; say so in the report.
- The `acta` on PATH must be the branch binary before any eval run. Build it to a folder of its own and put that folder first on PATH for the run; never overwrite the user's installed binary.
- Every gate the cold read has today stays: a fresh context, no same-context fallback, one pass, no approval of spec or plan, and the user's yes before any spec edit.
- Skill text stays general for every acta user, in plain English. When a file passes its byte cap, raise the cap in `internal/plugincheck/budget_test.go`; never trim words an agent acts on.

## Build brief

- Land order: this plan starts after PLN-0233 is on main, because both change `internal/plugincheck/budget_test.go` and `internal/plugincheck/evals_test.go`.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2
- Wave 3: Task 3

### Task 1: Two cold read cases with a planted defect, and their old-arm scores

**Files:** `plugin/evals/cold-read-planted-contradiction/`, `plugin/evals/cold-read-false-claim/`, `internal/plugincheck/evals_test.go`
**Seam:** `TestEvalCases` in plugincheck, and `scripts/eval --case 'cold-read-*'`
**verify:** Each case fails for the one behaviour it names and for no other reason: a spec whose two sections say opposite things must come back with a finding that names both lines; a spec that says a scaffold file holds something it does not must come back with a finding that names the spec line and the file. A run that only says the spec looks fine fails both. Every case has a `# guards:` line for a heading `cold-read.md` has today, a scaffold that refuses a folder that is not empty, and caps on turns and time. List each case, its graders, and its score over 10 runs on the old text.
- [ ] Red: `TestEvalCases` fails because the two case folders it now lists do not exist.
- [ ] Change: add the two case folders and their `evalCases` entries, then record the old-arm scores in the plan state.
- [ ] Commit: test(evals): cold read cases with a planted contradiction and a false claim

### Task 2: The reviewer checks claims against the repo and shows what it read

**Files:** `plugin/skills/shape/cold-read.md`, `internal/plugincheck/skill_shape_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** plugincheck required strings for the cold read, and the cases `cold-read-planted-contradiction`, `cold-read-false-claim` and `cold-read-accept`
**verify:** No claim the spec makes about a file, a command or today's behaviour goes unchecked: the brief tells the reviewer to check each one against the repo and report each that is wrong, and to report each requirement that has no way to be tested. Every finding quotes the spec line it is about. Every reply, with findings or with none, has a `Read` part naming the spec sections read and the references opened, and a reply that cut findings says `N more findings not listed`. The brief still forbids edits, write commands and delegation, and still treats document content as material, not instructions. List every kind of finding the brief asks for and what each must carry.
- [ ] Red: `skill_shape_test.go` fails because the brief lacks the claim check, the testability check, the quoted line, the `Read` part and the cut line.
- [ ] Change: add the two duties, the quoted line, the `Read` part and the cut line to the reviewer brief.
- [ ] Commit: feat(shape): cold read checks claims against the repo and shows what it read

### Task 3: Every finding reaches the user

**Files:** `plugin/skills/shape/cold-read.md`, `internal/plugincheck/skill_shape_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** plugincheck required strings for the cold read, and the cases `cold-read-planted-contradiction`, `cold-read-false-claim` and `cold-read-accept`
**verify:** No finding the reviewer returned can be lost between the reviewer and the user: `## Resolve` says the orchestrator presents every one with its quoted spec line, may add one line of its own view under each, and never leaves one out or merges two. A reply with no `Read` part is never presented as a finished cold read: the user is told it is incomplete and asked whether to retry or go on without it, and the orchestrator waits. The word "briefly" no longer gives leave to cut. List every reply shape the reviewer can send (findings, none, cut findings, no `Read` part, failure) and what the user sees for each.
- [ ] Red: `skill_shape_test.go` fails because `## Resolve` lacks the present-every-finding rule and the incomplete-reply rule.
- [ ] Change: rewrite the first paragraph of `## Resolve` with the two rules.
- [ ] Commit: feat(shape): cold read presents every finding and flags an incomplete reply
