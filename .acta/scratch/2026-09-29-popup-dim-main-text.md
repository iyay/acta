---
id: SCR-0017
hash: s4h6zmf
title: "TUI notes round 4: popup dim, copy id, tabs panel, sticky detail, expand/collapse, sort"
status: brainstorming
created: "2026-09-29"
started: "2026-09-29"
finished: "2026-09-29"
---
ketika popup muncul, text color di main window harusnya redup. bukan cuman border


---
Agent notes (2026-09-29, found by reading, not by running the TUI):
- The code already tries to dim everything. `cover` in internal/tui/view.go:602 strips the colors from every line behind the popup and repaints each one with `dim`. The status line gets the same treatment at view.go:128-132.
- `dim` at view.go:68 is `Faint(true)` plus foreground 250 (light) / 240 (dark). The user says only the border looks dimmed, so the text most likely ends up looking about the same as before. Not proven yet. Maybe the terminal ignores faint, or 240 is close to the normal text color. Check with acta:debug before changing anything.
- Popups: help `?`, the picker, and the others drawn by `popupBox` (view.go:555).
- Open question: how dim is "redup"? A darker grey, or a fixed low-contrast color from the theme work (SCRATCH-2 / SPEC-20)?

1. Bisa copy item id di setiap list dengan tekan tombol shortcut
2. untuk top tab di atas, harusnya ada di dalam panel baru, setiap tab ber-nomor 
3. Rename `List` di setiap panel jadi Open, kecuali di panel Activities -> Tasks
4. Indicator order dipindah ke bawah kanan, sebelum  total items
5. di panel Activities harus nampilin parent-nya, kalo plan berarti mirip dengan Panel Plans, Kalo Bug, berarti parent-nya Bug ID, expands by default
6. Tanggal created, start, finish, ditampilin di panel detail sebagai Footer. tapi footer-nya bisa gak sticky? begitu juga header, kalo bisa dijadiin sticky
7. Hilangkan 0 pada panel Detail, dan hilangkan juga shortcut-nya
8. Untuk expand Plan sekarang pake enter, tapi kesuilitan untuk collapse kalo selected items-nya ada di bawah, harus cari cara mudah untuk expand/collaps tanpa harus scroll items
9. Order/Sort kayaknya masih ngaco ya, belum berdasarkan created date?

Code map (2026-09-29, reading): keys tui/model.go:348, no clipboard code; tabs tui/view.go:118 not numbered; "List" tui/sidebar.go:151; sort word top border tui/frame.go:116, count bottom right tui/frame.go:160; Activities flat, no parent (tui/sidebar.go:188); plans start collapsed (tui/model.go:82); Detail scrolls as one block (tui/view.go:297), dates in header (tui/detail.go:41), "[0]" from tui/sidebar.go:65; sort key = day from file name, tie by slug (tui/order.go:16), so same-day items look unsorted.

Ruling 2026-09-29 (note 9): sort on the id number (SCR-0017 < SCR-0018), since ids are handed out in creation order. Day from the file name no longer drives order.

Ruling (note 1): key y copies the short id of the selected row (task row gives PLN-n#task-m). OSC 52 first, pbcopy fallback, status line shows "copied <id>".

Ruling (note 2): tabs move into their own bordered box on top, 3 rows tall, shown as "1 Scratches  2 Bugs ... 6 Activities", active tab highlighted, keys 1-6 unchanged.

Ruling (note 5): Activities groups in-progress tasks under their parent, tree like Plans, all groups open. Parent is the plan; when the plan has a bug parent, the group head is the bug only (one level).

Ruling (note 6): Detail header is sticky and holds every label field (id, title, status, parent...). Footer is sticky, one line: created · started · finished, blank as "-". Only the rest scrolls. Very short pane: both scroll with the body.

Ruling (note 8): h on a task row collapses its plan and moves the cursor to the plan row; h on a plan row collapses it; l expands. enter keeps toggling. Arrows stay on tabs.
