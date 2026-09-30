---
created: "2026-09-30"
id: SPC-0048
hash: r77dncs
started: "2026-09-30"
finished: "2026-09-30"
---
# Times in the detail dates, and every toast hides itself

Status: design approved by the user in chat on 2026-09-30. Bounded: the date writers (`internal/write`), the date reader (`internal/board`), the detail footer and the status line toast all exist.

## Why

1. The detail footer reads `created 2026-09-30 · started 2026-09-30 · finished -`. The user wants the time too. The files only hold a day: `internal/write/ids.go` (`created` on a new id), `internal/write/scratch.go` (`created` on a new scratch item) and `internal/write/dates.go` (`started`, `finished`) all write `Now().Format("2006-01-02")`. `dateField` in `internal/board/board.go` only accepts `YYYY-MM-DD`, so a time would be thrown away.
2. Most status line messages never go away. Only the id copy (`copyID`) and the drag copy (`internal/tui/select.go`) start the `clearStatusAfter(toastFor, ...)` timer. Popup results, `+` and `-` results, refusals like "nothing selected", editor errors, reload errors and "bug not saved" stay until the next message, and the key hints never come back.

## Design

1. **Write a time.** The `created`, `started` and `finished` fields are written as `YYYY-MM-DD HH:MM:SS` in local time, for example `2026-09-30 16:14:05`. The user asked for the seconds too. The status line clock is not changed. The value stays quoted, as it is now. The rules about when each field is written do not change: `MarkStarted` keeps an old `started`, `MarkFinishedOnce` keeps an old `finished`, and `MarkFinished` replaces it. File names and the day heading in a new scratch body keep the day only.
2. **Read both shapes.** `dateField` accepts `YYYY-MM-DD` and `YYYY-MM-DD HH:MM:SS`, and returns the text as it is. A value that is neither, or a day or time that does not exist, still reads as "". A YAML timestamp value keeps the day only, as it does now. The `Created`, `StartedOn` and `Finished` comments say both shapes.
3. **Show what is there.** The detail footer prints each value as it is: `created 2026-09-30 16:14:05 · started 2026-09-30 16:20:41 · finished -`. The long, short and cut rules stay. Old files that hold only a day show only a day. There is no migration.
4. **One place starts the toast timer.** `Update` notes `m.status` before it handles a message. When the status after is not empty and differs from the one before, `Update` adds `clearStatusAfter(toastFor, m.status)` to the command it returns. The two calls in `copyID` and in the drag copy are removed, so no toast gets two timers. `clearStatusMsg` already clears only the text it was sent for, so an old timer never clears a newer toast.
5. **Every toast lasts `toastFor`.** That includes errors. `toastFor` stays 2 seconds, and its comment says it covers every status message. When the toast goes, the left of the status line shows the key hints again. The search box is not a toast and is not touched.

## Testing

1. Each writer (new id, new scratch item, `MarkStarted`, `MarkFinished`, `MarkFinishedOnce`) writes `YYYY-MM-DD HH:MM:SS` from a fixed `Now`. Existing write tests that expect a bare day get the new value.
2. `dateField` reads both shapes and drops bad ones (`2026-02-30`, `2026-09-30 25:00:00`, `2026-09-30 16:14`, `yesterday`).
3. The detail footer at a normal width shows the times, and on a narrow box the short and cut forms still keep every date readable.
4. For each kind of message (a popup result, a refusal, an editor error, a reload error, a copy), the status goes back to the hints once the clear message arrives. A newer toast set before the old timer fires stays on screen.
