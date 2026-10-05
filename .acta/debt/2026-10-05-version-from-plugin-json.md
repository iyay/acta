---
id: DBT-0075
hash: rlnhl73
parent: plans/2026-10-05-version-from-plugin-json
---
# Review NOTEs: Version from plugin.json Implementation Plan

- [ ] (low) pluginFileVersion is copied in internal/cli/doctor_test.go and plugin/version_test.go with slightly different bodies; a change to how the test reads plugin.json must be made twice.
- [ ] (low) tuiVersionFor in internal/cli/doctor.go compares against a typed "dev" that must match the fallback in plugin/version.go; if that fallback changes, the TUI shows "v" plus the new text with no error. A shared constant fixes it.
