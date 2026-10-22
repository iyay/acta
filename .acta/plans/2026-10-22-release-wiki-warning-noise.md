---
parent: specs/2026-10-22-release-wiki-warning-noise-design
depth: minimal
id: PLN-0189
created: "2026-10-22 23:08:19"
hash: liryzbk
---

# Release commits stop staling two plugin wiki pages

**Goal:** A release version bump no longer makes `plugin-loads-from-repo` or `plugin-ships-in-binary` stale, and both pages still match the files they explain.
**Spec:** .acta/specs/2026-10-22-release-wiki-warning-noise-design.md
**Tests:** fast `acta wiki match <file>...` and `acta wiki check`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Start the build only after every plan of SPC-0172 has landed. SPC-0172 changes the page review rules and names these two pages in its baseline.
- Only `.acta/wiki/plugin-ships-in-binary.md` and `.acta/wiki/plugin-loads-from-repo.md` change. No Go code, no other page, no change to `scripts/release`.
- New paths, exactly: `plugin-ships-in-binary` gets `[plugin/embed.go, plugin/embed_test.go, internal/setup/]`; `plugin-loads-from-repo` gets `[plugin/skills/, plugin/hooks/, plugin/output-styles/, plugin/references/]`.

## Waves

- Wave 1: Task 1.

### Task 1: Narrow the paths of the two plugin pages

**Files:** Modify `.acta/wiki/plugin-ships-in-binary.md`, `.acta/wiki/plugin-loads-from-repo.md`.
**Seam:** `acta wiki match <file>...` and `acta wiki check`, run from the worktree.
**verify:** No file that `scripts/release` commits (`plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`) matches either page, and a commit that touches only those three files never makes either page stale. Each page still matches every one of its new paths, and `plugin-loads-from-repo` matches a file under `plugin/skills/`. Before the edit, every claim in each page body was compared with the code and with the unchecked commits under `plugin/`; each wrong claim (for example `go install ./cmd/acta`) is fixed in the same edit. List each claim checked and its result, and each match and check command run.
- [ ] Red: run `acta wiki check` and save its output as the baseline; run `acta wiki match plugin/.claude-plugin/plugin.json plugin/.claude-plugin/marketplace.json plugin/package.json` and see both pages listed.
- [ ] Change: read both page bodies in full, check each claim against the code and `git log` under `plugin/` since the page's last check, then set the new `paths` and fix any wrong claim. Run the match command again and see neither page listed; run `acta wiki match` on each new path and on `plugin/skills/build/SKILL.md` and see the right page listed; in a temporary clone, commit a version change to the three release files and run `acta wiki check` and see no line for either page; run `acta wiki check` in the worktree and see no error for either page.
- [ ] Commit: `docs(wiki): narrow plugin page paths so releases do not stale them`
