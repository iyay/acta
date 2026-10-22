---
parent: scratch/2026-10-22-release-wiki-warning-noise
id: SPC-0173
created: "2026-10-22 23:07:21"
hash: fyy1agu
started: "2026-10-22 23:41:14"
finished: "2026-10-22 23:42:39"
---
# Release commits stop staling two plugin wiki pages

## Why

`scripts/release` commits only three version files:
`plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`
and `plugin/package.json`. The pages `plugin-loads-from-repo` and
`plugin-ships-in-binary` both list `paths: [plugin/]`, so every release makes
them stale. The stale warning names only the newest unchecked commit
(`gitc.WikiHistory.Unseen`), so the release commit also hides the change that
matters. At spec time, `ccaffcf chore(release): v0.1.50` hides
`4d08f32 feat(style)`.

## Design

Change only the `paths` of the two pages. Neither page is about version files.

| Page | New paths |
| --- | --- |
| `plugin-ships-in-binary` | `plugin/embed.go`, `plugin/embed_test.go`, `internal/setup/` |
| `plugin-loads-from-repo` | `plugin/skills/`, `plugin/hooks/`, `plugin/output-styles/`, `plugin/references/` |

`plugin-loads-from-repo` loses `plugin/.claude-plugin/`. A change to the
marketplace source no longer hints this page. We accept that cost.

Any edit to a page file counts as a check commit, frontmatter included. Before
editing either page, read its full body and compare each claim with the code
and with the unchecked commits under its old paths. Fix any wrong fact in the
same edit. One example to check: `plugin-loads-from-repo` says to run
`go install ./cmd/acta`. Record the review in the plan's verification evidence.

Build this after SPC-0172 lands. SPC-0172 changes the rules for page review and
names these two pages in its baseline.

## Verification

1. Red, before the edit: `acta wiki match plugin/.claude-plugin/plugin.json plugin/package.json`
   lists both pages.
2. After the edit, the same command lists neither page.
3. `acta wiki match` lists each page for every one of its new paths, and lists
   `plugin-loads-from-repo` for a file under `plugin/skills/`.
4. In a temporary clone, commit a version bump to the three version files, then
   run `acta wiki check`. It prints no line for either page.
5. `acta wiki check` reports no error for either page.

## Boundaries

Only `.acta/wiki/plugin-ships-in-binary.md` and
`.acta/wiki/plugin-loads-from-repo.md` change. No change to the freshness
engine, the `paths` syntax, the stale message or `scripts/release`. Other pages
keep their paths.
