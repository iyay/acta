---
id: SCR-0058
hash: e0kbkqn
title: acta builds and runs natively on Windows
status: raw
created: "2026-10-08 12:11:32"
schema: "1"
---
# acta builds and runs natively on Windows

## Words

### 2026-10-08

The v0.1.44 release run (37730081152) failed: GOOS=windows does not compile. Unix-only syscalls in internal/write/runone.go (Flock), internal/write/tick.go (Stat_t, Flock), internal/hook/wikihint.go (Flock) and internal/evalomp/run.go (Setpgid, Kill). User chose on 2026-10-08 to ship macOS and Linux first and port Windows later.

Port idea: split each into _unix.go and _windows.go with build tags; file locks through LockFileEx (golang.org/x/sys/windows or a small wrapper), process group kill through a job object or taskkill /T. Then add windows back to .goreleaser.yaml, ship install.ps1 again, and put the Windows line back in plugin/README.md. Ties to SPC-0113 (curl install) and DBT-0102 (Windows CLAUDE_PLUGIN_ROOT path bug).

## Context

## Log

## Open questions
