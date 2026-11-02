---
id: SPC-0187
created: "2026-11-02 11:16:16"
hash: uqief7h
started: "2026-11-02 11:31:36"
finished: "2026-11-02 11:58:37"
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
   migration and no notice. The default resolves by name like any other theme,
   so a user file named `terminal` wins over the built-in, the same precedence
   that lets a user copy and change a built-in.
2. `acta config show` always prints one `theme:` line: the saved name, or the
   default name when nothing is saved. It never adds the `(default)` marker,
   because setup asks about every setting marked that way and theme is not one
   of setup's questions. `--json` reports the same string. A saved theme that
   no longer loads is still printed by name; `acta doctor` reports that it does
   not load.
3. Every fallback that used `theme.Default` keeps using it: a theme file that
   no longer loads falls back to the terminal theme, and `acta doctor` suggests
   `acta config set --theme terminal` as the fix for an unknown name.

Out of scope: new themes, changes to the `terminal` theme itself, a setup
question for the theme, and any migration of saved values. The doctor fix for
a theme file that does not parse stays as it is today: it suggests the default
name, which cannot repair that file.

## Acceptance

- `theme.Load("")` returns the theme named `terminal`; in a clean home that is
  the built-in, with an empty background.
- `acta config show` prints `theme: terminal` on a fresh home, and
  `theme: dracula` after `acta config set --theme dracula`.
- `acta config show --json` reports `"theme": "terminal"` when unset.
- `acta doctor` says `theme terminal loads` on a fresh home.
- The TUI opens with the terminal theme when nothing is saved.
