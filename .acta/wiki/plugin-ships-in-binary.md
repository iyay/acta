---
type: Reference
title: Plugin ships inside the acta binary
description: The binary embeds the plugin; acta setup unpacks it to ~/.acta/plugin; new plugin files must be embedded
paths: [plugin/, internal/setup/]
timestamp: 2026-10-07T13:39:26Z
---

`plugin/embed.go` embeds every plugin file except `evals/`, `evals-routing/` and Go files. `acta setup` unpacks that tree to `~/.acta/plugin` and installs Claude Code and omp from there, so a user needs only the binary. `--plugin-dir` points at a checkout instead and skips the unpack.

`plugin/embed_test.go` compares the embed with the files on disk: add a plugin file or folder and the test fails until `embed.go` lists it. Hook scripts lose their mode in the embed; the unpacker sets 0755 on extensionless files and `.sh` files, so a new executable with another extension needs that rule changed.

This repo's own Claude Code still loads the plugin in place from `plugin/` (see /plugin-loads-from-repo.md); `~/.acta/plugin` is for users of the binary.
