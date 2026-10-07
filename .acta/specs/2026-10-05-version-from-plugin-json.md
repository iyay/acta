---
parent: none
id: SPC-0073
created: "2026-10-05 08:36:24"
hash: wpf1svc
started: "2026-10-05 11:55:05"
finished: "2026-10-05 12:02:52"
---
# Show the plugin version in the TUI and doctor

Status: Bounded, approved by the user in chat on 2026-10-05.

Why: the TUI shows `v0.0.0-20261005004419-b607bd600cf9`. Since Go 1.24, `go install` stamps a pseudo-version from git, and this repo has no tags. The real version lives only in `plugin/.claude-plugin/plugin.json` (0.1.5 today), which `internal/plugincheck` already keeps in step with the other two version files.

Design:

- A small Go package in `plugin/` embeds `.claude-plugin/plugin.json` with `go:embed` and exposes `Version() string`, the `version` field as written.
- The TUI version (`internal/cli/cli.go`) becomes `v` plus that version. The Go build info is no longer used there.
- `acta doctor` (`buildVersion` in `internal/cli/doctor.go`) shows the same version, and keeps the commit in brackets when the build stamped one, so a stale binary is still easy to name.
- A bad or empty `version` field shows `dev`, never a crash.

Tests: the package returns the version in plugin.json; the TUI and doctor text carry it; a bad JSON gives `dev`.
