---
id: SCR-0091
hash: mt707oe
title: debug-race-case-on-sonnet
status: raw
created: "2026-11-07 17:12:38"
schema: "1"
---
# debug-race-case-on-sonnet

## Words

### 2026-11-07

Race debug case fails on Sonnet: in the SPC-0157 runs the race eval passed 0/2 on Sonnet. The proposed regression test starts many threads instead of holding workers between the read and the write. Wording changes over three rounds did not move it. Codex passed 2/2. Split out of SPC-0191 on 2026-11-07.

## Context

## Log

## Open questions
