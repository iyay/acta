---
parent: scratch/2026-09-28-themes
id: SPEC-20
hash: qham
---
# TUI themes

Status: design approved by the user in chat on 2026-09-29. Architectural (new user setting, new package).

## Why

The TUI paints with fixed ANSI-256 numbers (`internal/tui/view.go:57-76`). Those numbers never follow the terminal's theme, so on some themes the colors clash. The user wants theme support: a good default, a set of built-in themes, and room for their own.

Reference: terminalcolors.com. It lists terminal color schemes. Each one is 16 ANSI colors plus a background, a foreground and a selection color. It has no API; it offers files per terminal (Ghostty, Alacritty, iTerm2, and so on).

## Rulings

- Default theme is `tokyo-night`. It is used when no theme is set.
- A built-in theme named `terminal` has no hex colors. It uses ANSI numbers 0-15, so it follows whatever theme the terminal has.
- The choice is a user setting, not a repo setting. It lives in `~/.acta/voice.yaml` as `theme:`.
- Themes come from a small built-in set, plus user files at `~/.acta/themes/<name>.yaml`.
- A theme file has the same shape as a terminalcolors scheme, so a user can copy the colors straight across.

## Theme shape

```yaml
bg: "#1a1b26"
fg: "#c0caf5"
selection_bg: "#283457"   # optional
selection_fg: "#c0caf5"   # optional
ansi: ["#15161e", "#f7768e", ...]   # exactly 16
```

The name is the file name. Every color is `#rrggbb`.

## Components

1. **New package `internal/theme`.**
   - `Theme{Name, BG, FG, SelectionBG, SelectionFG string; ANSI [16]string}`.
   - Built-ins: `terminal`, `tokyo-night`, `tokyo-night-day`, `catppuccin-mocha`, `catppuccin-latte`, `gruvbox-dark`, `dracula`. Hex comes from the Ghostty files on terminalcolors.com.
   - `Load(name)`: empty name means `tokyo-night`. It looks in `~/.acta/themes/<name>.yaml` first, then in the built-ins. A user file wins over a built-in with the same name. A name that is not found returns an error that names it.
2. **Voice setting `theme`.** `acta voice set theme <name>` refuses a name `theme.Load` cannot load. `acta voice show` prints it.
3. **TUI styles move off package globals.** The five styles `accent`, `work`, `faint`, `dim` and `selected` become fields of a `styles` struct. It is built from a `Theme` and kept on `Model`. `tui.New` takes the theme.
4. **Fixed role slots.**

   | Role | Color |
   |------|-------|
   | accent | slot 12 |
   | work | slot 4, faint |
   | faint | faint, no color |
   | dim | slot 8, faint |
   | selected | `selection_fg` on `selection_bg`, bold. When those are missing: slot 15 on slot 0 for a dark theme, slot 0 on slot 7 for a light one |

5. **`terminal` mode.** Slots render as ANSI numbers (`"12"`), and nothing paints the background. Light or dark for the selected row comes from the `dark` flag `tui.New` already takes.
6. **Hex themes paint `bg` and `fg` over the whole frame.** Without that, a light theme on a dark terminal is unreadable.
7. **`acta doctor`** checks that the chosen theme loads.

## Data flow

`acta tui` starts. `voice.Resolve()` reads `theme`. `theme.Load(name)` returns the theme. `tui.New(cfg, board, dark, theme)` builds the styles. Every render reads `m.styles`.

## Errors

- The name is not found, the YAML is bad, a hex is bad, or there are not exactly 16 slots: the TUI still starts with `tokyo-night`. The status line shows `theme "<name>": <reason>`. A color problem must never stop the board from opening.
- `acta voice set theme` refuses the same errors up front, so a typo shows at once.

## Testing

- `theme`: every built-in loads; empty name gives `tokyo-night`; a user file wins over a built-in; unknown name; bad YAML; bad hex; 15 slots and 17 slots; missing selection keys fall back to slots.
- `tui`: `terminal` gives ANSI numbers and no background; a hex theme gives hex and paints the frame; the selected fallback for dark and for light; a bad theme falls back to `tokyo-night` and shows the message.
- `voice`: setting a valid theme is saved; an unknown theme is refused and the file stays as it was.
- `doctor`: a good theme passes; a broken user file is reported.

## Out of scope

- Changing single roles in a theme file.
- Reading Ghostty or Alacritty files directly.
- Switching themes while the TUI is running.
