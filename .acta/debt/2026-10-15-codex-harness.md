---
id: DBT-0122
hash: yhaovmr
parent: plans/2026-10-15-codex-harness
---
# Codex config parsing in doctor

- [ ] (medium) internal/doctor/codex.go codexHasPlugin reads config.toml line by line: inline comments, tabs, a quoted enabled key, single-quoted headers, inline tables and dotted keys read as "not installed", a false Warn. A first fix was reverted because it dropped TrimSpace and broke CRLF files; the next try needs a CRLF case.
