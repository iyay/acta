---
id: SCR-0053
hash: hnfoyjm
title: 'acta update: check and install the latest version'
status: brainstorming
created: "2026-10-07 09:40:36"
schema: "1"
started: "2026-10-09 15:38:28"
---
# acta update: check and install the latest version

## Words

### 2026-10-07

User idea, 2026-10-07: acta needs an update mechanism. `acta update` checks for a newer version and, when there is one, installs the latest version on its own.

Depends on SCR-0052 (curl install script and release pipeline): there must be published releases to check against. Open: what gets updated (the binary, the plugin in each harness, or both), where the version check reads from, and whether a normal command should hint that an update exists.

## Context

Ruling 2026-10-07: the acta plugin folder is embedded in the binary (go:embed) and `acta setup` extracts it to ~/.acta/plugin, then installs into Claude Code and omp from there (PLN-0113 Task 04). So `acta update` = replace the binary, then extract the plugin again. A GitHub marketplace comes later, after the repo is public.

## Log

### 2026-10-09

Q1 scope: binary + plugin. Download new binary, verify checksum, replace running binary, then run acta setup with the new binary so the plugin is extracted again.

### 2026-10-09

Q2 hint: none. Version check runs only in acta update and acta update --check. No background network call in other commands, hooks or TUI.

### 2026-10-09

Q3 source builds: refuse. A binary not built by the release pipeline prints 'built from source, rebuild with go install' and exits 1.

### 2026-10-09

Approach: Go native. Resolve latest tag via releases/latest redirect, download archive + checksums.txt over https, verify SHA-256, extract, temp file next to os.Executable(), rename, then exec new binary's acta setup. Release builds get -X release=true ldflag from goreleaser to tell them from source builds. Flags: --check; ACTA_DOWNLOAD_URL override shared with install.sh.

### 2026-10-09

Section 1 approved: flow. 1 refuse non-release build. 2 latest tag via redirect, equal = exit 0, --check stops here. 3 https download, 60 s timeout, SHA-256 check, extract. 4 temp file next to resolved os.Executable(), chmod 755, rename; old binary intact on failure. 5 exec new binary hidden `acta update --refresh-plugin` = setup.ExtractPlugin only, no wizard.

### 2026-10-09

Section 2 approved: refresh + errors. --refresh-plugin extracts, then claude plugin marketplace update acta-local + claude plugin update acta@acta-local when claude is on PATH (exact commands checked in task 1). Harness failure: binary stays updated, print run acta setup, exit 1. Network/checksum error: binary untouched, exit 1. No write access: message, exit 1, no sudo. Tests: httptest via ACTA_DOWNLOAD_URL, fake claude on PATH.

## Open questions
