---
type: Gotcha
title: glamour and chroma theme traps
description: Charm registers once per process, chroma needs hex or ansi names, and Text.Color kills heading colors
paths: [internal/tui/]
timestamp: 2026-10-05T23:18:53Z
---

- `StyleCodeBlock.Chroma != nil` registers code colors once per process under the fixed name `charm`; the first theme to draw a code block owns them. Fix: Chroma nil, set `CodeBlock.Theme` to a per-theme name registered in chroma's registry yourself.
- chroma takes only `#rrggbb` or `#ansi*` names; a bare number panics or reads as hex.
- `#ansigreen` and `#ansiblue` are the bright shades (92/94); slots 2 and 4 are `#ansidarkgreen` and `#ansidarkblue` (32/34).
- Use formatter `terminal16m` for hex themes, `terminal16` for the terminal theme.
- Never set `Styles.Text.Color`; it overwrites every heading color. Leave it nil.
- chroma `styles.Get` takes no lock, so code-block tests must not run `t.Parallel`.
- Only rendered SGR bytes catch these; config-level tests passed while output was wrong.
