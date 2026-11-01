---
id: SPC-0187
created: "2026-11-02 11:16:16"
hash: uqief7h
---
# Default theme is the terminal, and config show names it

## Problem

The board opens with `tokyo-night`, a fixed dark blue palette, unless the user
picked a theme. Users who never chose one get colors that do not match their
terminal, and `acta config show` prints no `theme:` line at all when the value
is unset, so there is no way to see which theme is in use.

## Design

1. `theme.Default` becomes `"terminal"`. That built-in has an empty background
   and uses the terminal's own ANSI colors, so the board looks native in any
   terminal and on any palette. Users who set a theme keep it; there is no
   migration and no notice.
2. `acta config show` always prints the theme: the chosen name, or
   `theme: <default> (default)` when unset, the same way `plan_depth`,
   `questions` and the other keys already show their default. `--json` carries
   the effective name too, so a script sees what the TUI will use.
3. Every fallback that used `theme.Default` keeps using it: a theme file that
   no longer loads falls back to the terminal theme, and `acta doctor` suggests
   `acta config set --theme terminal` as the fix.

Out of scope: new themes, changes to the `terminal` theme itself, and any
migration of saved values.

## Acceptance

- `theme.Load("")` returns the terminal theme with an empty background.
- `acta config show` prints `theme: terminal (default)` on a fresh home, and
  `theme: dracula` with no marker after `acta config set --theme dracula`.
- `acta config show --json` reports `"theme": "terminal"` when unset.
- `acta doctor` says `theme terminal loads` when no theme is set.
- The TUI opens with the terminal theme when no theme is set.
