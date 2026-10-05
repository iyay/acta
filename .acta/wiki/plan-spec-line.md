---
type: Reference
title: How a plan's Spec line is read
description: First backtick .md span, else first bare .md word with a slash, else no spec; the none form for plans without a spec
paths: [internal/board/]
timestamp: 2026-10-06T00:00:00Z
---

`specPath` in `internal/board/parse.go` takes the first backtick span ending in `.md`, else the first bare word ending in `.md` that has a `/`, else no spec (no warning).

A plan with no spec writes `**Spec:** none (Bounded, approved in chat on <date>)`.

A backticked non-spec `.md` on the Spec line still counts as the spec, so keep that line to the spec path or the none form.
