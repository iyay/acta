---
id: SPC-0155
created: "2026-10-19 07:21:30"
hash: kytulm4
started: "2026-10-19 07:22:35"
finished: "2026-10-19 07:23:25"
---
# One Enter per setup question under curl | sh

## Problem

Under `curl ... | sh`, `scripts/install.sh` runs `acta setup </dev/tty`, and
every setup question needs two Enters. A direct `acta setup` needs one.

Each question runs its own huh form, so its own Bubble Tea program. When a
program ends, its input reader must be cancelled. cancelreader on macOS
picks select only for a file named `/dev/tty`
(`cancelreader_bsd.go:28`). Stdin redirected from `/dev/tty` is named
`/dev/stdin` in Go, so kqueue is used, which does not work on `/dev/tty`.
The old reader stays blocked and eats the first key of the next question.

## Design

Bounded change approved in chat on 2026-10-19.

- Keep one form per question: the answered lines piling up is on purpose.
- When stdin is a terminal, `Ask` opens `/dev/tty` once and passes it to
  every form with `huh.Form.WithInput`, the override questions
  (`runGroup`) included. cancelreader then uses select and the old reader
  stops.
- When `/dev/tty` cannot be opened, keep today's behavior.

## Files and verification

- `internal/setup/form.go`: open the tty, pass it to each form.
- `internal/setup` test: the input given to the forms is a file named
  `/dev/tty` when one can be opened, and none when it cannot.
- Manual: run the branch build through the curl | sh path in herdr; one
  Enter moves to the next question.
