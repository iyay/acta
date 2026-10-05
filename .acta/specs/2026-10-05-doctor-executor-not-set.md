---
id: SPC-0075
created: "2026-10-05 12:25:49"
hash: fpx1t09
started: "2026-10-05 12:50:42"
finished: "2026-10-05 12:52:47"
---
# Doctor treats an unset build executor as fine

Status: Bounded, approved by the user in chat on 2026-10-05.

Why: `acta doctor` warns `no default build executor` in every repo that has not picked one. But no default is a valid choice: `acta:build` asks which executor to run when none is set. The warning makes a healthy setup look broken.

Design:

- In `checkSetup` (`internal/doctor/doctor.go`), an empty build executor gives `ok` with the message `voice is set; build executor not set, build asks each time`, and no fix line.
- A missing voice file stays a `warn` with the `/acta:setup` fix, as today.
- A set executor stays `ok` with the current message.

Tests: the three cases above, each with its level, message and fix line.
