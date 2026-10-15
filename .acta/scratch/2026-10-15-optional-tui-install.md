---
id: SCR-0068
hash: yv49k55
title: Setup asks whether to install the TUI
status: brainstorming
created: "2026-10-15 09:09:17"
schema: "1"
started: "2026-10-15 09:09:19"
finished: "2026-10-15 09:34:43"
---
# Setup asks whether to install the TUI

## Words

### 2026-10-15

Setup asks whether to install the TUI. Users who only want the skills and hooks get a CLI-only acta binary (no Bubble Tea), since every skill needs the CLI but none needs the TUI. Today acta is one 22.6M binary: no args opens the TUI, subcommands run the CLI.

## Context

## Log

### 2026-10-15

Q1 why no TUI: binary size. User wants a smaller binary (option 1).

### 2026-10-15

Measured 2026-10-15: release build has no -s -w. Plain build 24M, with -s -w 17M (-7M). TUI deps (bubbletea, lipgloss, fsnotify) add about 1.5M stripped, plus internal/tui code (21k lines), so a no-TUI build saves roughly 2-3M only.

### 2026-10-15

Decision 2026-10-15: strip the release binary (-s -w) as a separate Bounded change. The no-TUI build is on hold; reopen only if 17M is still too big.

### 2026-10-15

2026-10-15: user reopened the TUI option after the strip landed (17M). Goal stays: smaller binary for users who skip the TUI.

### 2026-10-15

Q2 where the choice lives: install.sh asks before download (env ACTA_TUI=0 for non-interactive). acta setup does not ask; it only shows the installed variant.

### 2026-10-15

Q3 default with no tty and no ACTA_TUI: full build with TUI (same as today). Defaults taken without asking: binary knows its own variant so acta update fetches the same one; bare acta in the small build prints help plus one line on how to get the TUI.

### 2026-10-15

Measured again 2026-10-15 (stripped, darwin/arm64, in a scratch copy with runTUI stubbed): full 16.6M, no-TUI 9.4M (-7.2M, -43%). Earlier 2-3M guess was wrong: it counted only bubbletea/lipgloss/fsnotify. The bulk is glamour, chroma, goldmark, bluemonday (markdown rendering in the TUI). setup (huh) and doctor (lipgloss) still keep bubbletea and lipgloss in the small build.

### 2026-10-15

Approach chosen: build tag notui (option 1). runTUI in two files by tag; goreleaser adds a notui build with its own asset name; a const set by the tag tells acta update which asset to fetch.

### 2026-10-15

Section 1 approved (build and release): runTUI in tui_full.go (!notui) and tui_none.go (notui, prints help + one line on ACTA_TUI=1, exit 0); cli.Variant full|cli shown by --version and doctor; goreleaser second build acta-cli with tags [notui], archive acta-cli_{{ .Os }}_{{ .Arch }}.tar.gz, old asset name unchanged; tests: plugincheck two builds both stripped, only acta-cli has notui; CI builds -tags notui on 4 pairs; go list -deps -tags notui must not hold internal/tui or glamour.

### 2026-10-15

Section 2 approved (install.sh): ACTA_TUI 1 full, 0 small, other value fails before download; unset + tty asks once "Install the TUI board too? It adds about 7M. [Y/n]" (Enter/y full, n small, else ask again); unset + no tty = full, no prompt; asset acta_ or acta-cli_, file on disk stays acta; install.ps1 untouched (SCR-0058); tests in scripts/install_test.go for env values, no tty, pty prompt.

### 2026-10-15

Section 3 approved (update and doctor): update.Fetch takes asset prefix acta or acta-cli from cli.Variant; doctor info line "build: cli (no TUI)" or "build: full", never warn or fail; switch variant only by rerunning install.sh with ACTA_TUI; wiki release.md gets one sentence; tests: Fetch with acta-cli prefix, release without acta-cli asset fails clearly and replaces nothing, doctor line in both variants.

## Open questions
