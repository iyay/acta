---
id: SPC-0153
created: "2026-10-19 07:11:26"
hash: e6dpc7f
---
# Explain the installer and its optional TUI

## Design

Bounded change approved in chat on 2026-10-19. Before the interactive build
choice, print one blank line and this introduction once:

```text
This will install acta, a development toolkit with a command-line tool
and AI agent skills for tasks such as planning, building, testing, and reviewing code.

You can also install a terminal user interface (TUI) to browse specs, plans, and bugs.
```

On the next line, print `Install the TUI? Adds about 7 MB. [Y/n]` in bold,
then reset the terminal style before reading the answer. Keep the trailing
space for input. Repeat only the question after an invalid answer.

Keep the existing default, environment override, noninteractive behavior,
download checks and setup call. The Herdr input issue is outside this change.

## Files and verification

- `scripts/install.sh`: introduction and bold question on the interactive path.
- `scripts/install_test.go`: exercise the real script through the existing PTY
  helper; check spacing, one introduction, bold/reset, repeated questions,
  and the selected archive for Enter, yes, no and an invalid answer.
- Run `scripts/test ./scripts/ -run TestInstall` and shell syntax checks.
