---
type: Gotcha
title: xansi Wrap overflows on dash
description: Wrap can return lines wider than the limit after " -"; use Hardwrap(Wordwrap(...)) instead
paths: [internal/tui/]
timestamp: 2026-10-09T14:31:19Z
---

`xansi.Wrap(s, w, "")` can return a line wider than `w` when a `-` follows a space. The fit step then cuts it and text is lost. Fix: `xansi.Hardwrap(xansi.Wordwrap(s, w, ""), w, true)`.

Why it hid: the unit test note had no " -"; only a width sweep over real notes caught it. New wrap code must sweep widths over text with " -", long words, and wide runes. glamour markdown output overflows too, and drops `<tag>` text and backslashes before punctuation, so escape `\` and `< >` first.
