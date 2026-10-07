---
created: "2026-09-30"
id: SPC-0039
hash: u5azafn
started: "2026-09-30"
finished: "2026-09-30"
---
# TUI status colors: help key column, green done, pulsing dot

Status: design approved by the user in chat on 2026-09-30. Bounded: the help popup, the status dots, the Tasks pane and the detail work lines all exist already. This spec changes how they are painted and adds one animation tick.

## Why

The user left four notes after PLN-0039 and PLN-0046 landed:
1. The keys in the `?` popup share the accent with the border and the title, so they do not stand out. The user wants the keys in a column of their own, laid out like the lazygit keybindings popup.
2. Work that is done does not look done. Its title and its tick should be green.
3. In the Tasks pane, rows with work under way are blue. The user wants the text in the foreground color, with the dot pulsing green, fading in and out.
4. The same rules apply to the task and step lines in the detail of a plan or a task.

## Changes

1. **Help key column.** Keys are drawn in cyan (slot 6), not the accent. The key column is right-aligned to the width of the widest key. Each description starts one space after that column and is left-aligned. The words in `helpLines` stay the same; only `helpText` changes how it lays them out. The box size rule (`boxView`) does not change.
2. **Done is green.** When a row or a line names an item that is done (the `✓` state, or a closed status for a list row), its title and its mark are drawn in green (slot 2). The id keeps its kind color. This covers:
   - list rows in every pane, the Done pane and tree task rows included;
   - detail work lines (`workLine`);
   - detail step lines (`stepLines`).

   A done plan row keeps its `+` or `-` mark; only its title turns green. The row under the cursor keeps the selection band, as it does today.
3. **Work under way.** The text of a row or line with work under way is drawn in the foreground color. The `work` brush and `slotWork` are removed. The `●` dot is green and pulses. The pulse has 8 steps that go from green toward the background and back. It moves one step every 120 ms, so one full cycle takes about 1 s. The pulse covers every `●` dot: list tree rows, detail work lines and detail step lines.
4. **Pulse stays readable.** No pulse frame may drop below a contrast of 1.6 against the theme background. Below that, Ghostty's `minimum-contrast` would swap the dot for white (see BUG-0014). For each hex theme, the darkest frame is therefore the color closest to the background that still keeps 1.6. A small mix helper comes back for this. In the `terminal` theme there are no hex values, so the dot alternates between slot 2 and slot 10.
5. **Tick only while there is work.** A pulse message advances the frame. The tick is armed by the first window size message and by a reload, and only when the board holds work under way. It re-arms itself while that work lasts, and stops when none is left. A tick only draws again when a `●` dot was on screen in the last frame. Otherwise it changes nothing and the last frame is reused, so no draw happens. A still TUI with no work under way never ticks at all.
6. **The cursor row.** The row under the cursor keeps its selection band. Its dot is drawn inside the band and does not pulse.

Unchanged: the kind colors, the selection band, the popup dim, the frame reuse rule for messages that change nothing, and the theme file format.

## Order

PLN-0045 (TUI scroll at 60fps) landed as e48bce4 before this plan was written, so the branch starts from a main that already holds it.

## Files

- `internal/tui/styles.go`: pulse frames, the green brush, and removal of `work`.
- `internal/tui/view.go`: `helpText`.
- `internal/tui/scroll.go`: list rows.
- `internal/tui/detail.go`: work lines and step lines.
- `internal/tui/model.go`: the pulse message and tick.

Tests go with each file.

## Testing

- **Help popup.** Every key of `helpLines` is drawn in slot 6 and none in the accent. The key cells end at the same column on every line, and each description starts one cell later. The words match `helpLines` once the extra spaces are collapsed. The box width is unchanged at several window widths.
- **Done.** On every path named in change 2, a done item's title and mark are green and its id is in its kind color. A row that is not done is never green.
- **Under way.** On every path named in change 3, the text is never blue and never faint, and the dot is the current pulse frame.
- **Frames.** Every pulse frame of every built-in hex theme has a contrast of 1.6 or more against its background. The frames rise and fall back to green over 8 steps. The `terminal` theme alternates between slots 2 and 10 with no 24-bit code.
- **Tick.** Tests send the pulse message straight to `Update`, with no sleep:
  - it advances the frame and arms the next tick while a dot is on screen;
  - with work but no dot on screen, it re-arms and changes nothing on screen;
  - with no work, it arms nothing;
  - a resize or a reload with work arms the tick once, never twice.
