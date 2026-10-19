---
parent: specs/2026-10-19-installer-introduction-design
depth: minimal
id: PLN-0165
created: "2026-10-19 07:11:53"
hash: p4pw96t
---
# Explain the installer and its optional TUI

**Goal:** Explain acta and the optional TUI before asking a clearly separated, bold install question.
**Spec:** `.acta/specs/2026-10-19-installer-introduction-design.md`
**Tests:** `scripts/test ./scripts/ -run TestInstall` (fast); `scripts/test --full` (full, land only); `sh -n scripts/install.sh`; `gofmt -l scripts/install_test.go`; `go vet ./scripts/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Use the exact introduction and question approved in the spec. Apply bold only to the question and reset it before reading input.
- Keep the introduction once per interactive run, existing archive choices, environment overrides, noninteractive behavior, download checks and setup call. The Herdr input issue is outside this change.
- Build inline in a worktree from main. No release or push.

## Waves

- Wave 1: Task 1.

### Task 1: Explain the interactive installer choice

**Files:** Modify `scripts/install.sh` and `scripts/install_test.go`.
**Seam:** The real installer through the existing `installRig.ptyRun` and release archive assertions.
**verify:** Every interactive choice starts with the agreed separated introduction exactly once and each question is bold with a reset before input; Enter/yes select the full build, no selects CLI, and invalid input repeats only the question. Existing environment and noninteractive cases still pass.
- [ ] Red: Update `TestInstallAsksOnTerminal` to check the introduction and bold question for its existing answer cases; observe `scripts/test ./scripts/ -run TestInstallAsksOnTerminal` fail on the current prompt.
- [ ] Change: Print the introduction once before the interactive loop and the bold question inside it; pass the installer tests, shell syntax, formatter and vet checks.
- [ ] Commit: `fix(install): explain acta and highlight the TUI choice`.
