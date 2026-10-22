---
parent: specs/2026-10-22-release-wiki-warning-noise-design
depth: minimal
after: [PLN-0187, PLN-0188]
id: PLN-0189
created: "2026-10-22 23:08:19"
hash: liryzbk
started: "2026-10-22 23:41:14"
finished: "2026-10-22 23:42:39"
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
- [x] Red: run `acta wiki check` and save its output as the baseline; run `acta wiki match plugin/.claude-plugin/plugin.json plugin/.claude-plugin/marketplace.json plugin/package.json` and see both pages listed.
- [x] Change: read both page bodies in full, check each claim against the code and `git log` under `plugin/` since the page's last check, then set the new `paths` and fix any wrong claim. Run the match command again and see neither page listed; run `acta wiki match` on each new path and on `plugin/skills/build/SKILL.md` and see the right page listed; in a temporary clone, commit a version change to the three release files and run `acta wiki check` and see no line for either page; run `acta wiki check` in the worktree and see no error for either page.
- [x] Commit: `docs(wiki): narrow plugin page paths so releases do not stale them`

## Verification evidence

Baseline before any edit: `acta wiki check` (branch binary) exited 0 with 15 stale warnings and no errors; both target pages were stale. Red: `acta wiki match` on the three release files listed both pages.

Claim review, `plugin-ships-in-binary`: the embed line in `plugin/embed.go` names every plugin entry except `evals/`, `evals-routing/` and the Go files (true). `ExtractPlugin` writes to `<home>/.acta/plugin`, and setup installs Claude Code, omp and Codex from that folder (true). `pluginDirFor` uses `--plugin-dir` as given and skips the unpack (true). `TestEmbedMatchesDisk` fails when a file on disk is not embedded (true). The mode rule was too broad: `isScript` gives 0755 to `.sh` files and to extensionless files directly in `hooks/` only, so the page now says that.

Claim review, `plugin-loads-from-repo`: the local marketplace `acta-local` is a `directory` source at the repo's `plugin/` folder (true); `plugin/.claude-plugin/` changed only in version and description since the last check, so the loading claims still hold. `go install ./cmd/acta` matches `CLAUDE.md` and `rebuild-binary-after-merge`, so it stays; a machine whose PATH binary lives elsewhere is personal setup, not a repo fact. The cache copy is old and unused (true): `installed_plugins.json` points `acta@acta-local` at `cache/acta-local/acta/0.1.0`, last updated 2026-09-26, while this session loads skills from the repo's `plugin/skills/`.

After the edit: the release-file match lists neither page. Positive matches: `plugin/embed.go`, `plugin/embed_test.go`, `internal/setup/extract.go` list `plugin-ships-in-binary`; `plugin/skills/build/SKILL.md`, `plugin/hooks/post-tool`, `plugin/output-styles/acta.md`, `plugin/references/house-rules.md` list `plugin-loads-from-repo`. In temp clones with a fake `chore(release): v9.9.9` bump of the three release files, `acta wiki check` exited 0 in both; the branch clone printed 13 warnings, none for the two pages, while the main clone printed 15, including both pages stale on the bump commit. No errors in either. In the worktree, `acta wiki check` exited 0 with no error and no line for either page, and `acta wiki check main..HEAD` printed nothing.

## Polish

### Task 2: Review polish

**verify:** NOTE below applied, and nothing else changes.
- [x] `plugin-ships-in-binary` also covers `plugin/hooks/`, so a hook author sees the 0755 rule; release commits never touch `plugin/hooks/`.
- [x] Verification evidence names the cache-copy claim and the worktree `acta wiki check` run.
- [x] Commit: `polish: review notes PLN-0189`

## State

### Next

Task 1 committed with evidence; next is acta:review over main..HEAD, then acta:land.

## Review notes

Deep, full tier (wiki pages agents read as hints), two reviewers; round 1 CLEAN, no BLOCKER, no fact-check needed. User approved one deviation from the spec path table: `plugin-ships-in-binary` also covers `plugin/hooks/` so hook authors see the 0755 rule (polish, reviewed CLEAN by both axes). Notes left: the marketplace-source line has no trigger (accepted in spec); `plugin/codex/`, `plugin/omp/` and `plugin/README.md` match neither page, while `TestEmbedMatchesDisk` still guards new folders; no test covers `isScript`.
