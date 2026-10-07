---
id: DBT-0093
hash: dh6pvs9
parent: plans/2026-10-07-acta-setup-wizard
---
# acta setup review debt

- [ ] (low) WriteBlock replaces a dangling symlink (CLAUDE.md -> missing file) with a plain file, against its doc line; resolve with os.Readlink and write to the missing target, or refuse. No test covers this shape (internal/setup/block.go:44-48).
- [ ] (low) A file with mixed line endings gets an all-CRLF block, and nothing pins that behaviour (internal/setup/block.go blockText).
