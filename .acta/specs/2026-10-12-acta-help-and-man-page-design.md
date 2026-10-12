---
closes: [DBT-0034.07, DBT-0044.05]
id: SPC-0128
created: "2026-10-12 09:45:52"
hash: xc00wjw
---

# acta Help, Version and Man Page

Status: approved by the user on 2026-10-12 (Bounded, probe mode, every recommendation taken). `--version` added by the user the same day.

## Why

`acta --help` prints only the `-root` flag. `acta help` fails with `unknown command "help"`, and that error lists commands but misses `config`, `hook` and `voice`. `man acta` finds nothing. `acta --version` fails too, so there is no quick way to tell which build is on PATH. Each command has its own usage string spread over `internal/cli/*.go`, each in a different shape. A user or an agent has no single place to learn the commands.

## Design

1. **One command table.** A new file in `internal/cli` (for example `help.go`) holds one entry per command: name, group, a one-sentence summary, usage line, flags (one per line) and one example for commands that write files. The usage strings in `cli.go`, `tick.go`, `wiki.go`, `config_cmd.go` and the other command files read from this table instead of their own strings.
2. **Groups**, in this order:
   - Board: `list`, `show`, `set`, `state`, `tick`, `commits`
   - Write: `bug`, `debt`, `scratch`, `commit`, `id`, `wiki`
   - Setup: `setup`, `config`, `doctor`, `update`, `migrate-root`, `tidy`
   - Agent plumbing: `hook`, `dispatch`, `reply-back`, `run-one`, `eval-omp`
   - `voice` is an old alias and stays out of the list. It still works.
3. **Terminal help.**
   - `acta help`, `acta --help` and `acta -h` print every group to stdout and exit 0.
   - `acta help <cmd>` and `acta <cmd> --help` (and `-h`) print that command's detail to stdout and exit 0.
   - An unknown command prints `unknown command "x"; run acta help` to stderr and exits 1. The long command list in that error goes away.
   - `acta` with no arguments still opens the TUI. `acta -root X` still works.
4. **Man page.**
   - The binary renders `acta.1` in roff from the same table, so the man page and `acta help` always match.
   - `acta setup` and the plugin refresh step of `acta update` (`refreshPlugin`) write it to `<dir of the running binary>/../share/man/man1/acta.1`. For the standard install that is `~/.local/share/man/man1/acta.1`. man on macOS and man-db on Linux search `<prefix>/share/man` for each `<prefix>/bin` on PATH, so no MANPATH change is needed.
   - A failed man page write prints a warning. Setup and update still go on and their exit code does not change for it.
5. **Version.**
   - `acta --version` prints `acta <version>` to stdout and exits 0. The version comes from `plugin.Version()`, the same one `acta update` compares.
   - A binary not built by goreleaser (`update.IsRelease()` is false) prints `acta <version> (source build)`, so a `go install` copy is easy to tell from the release install.
   - `--version` shows in `acta help` and in the man page.
6. **Tests.**
   - Every `case` in the `Run` switch has a table entry; a new command without one fails a test.
   - `acta help`, `acta <cmd> --help`, the unknown-command path: output stream, exit code and that each listed command shows up.
   - `acta --version` output for a release and a source build.
   - The man page lands at the path derived from the binary location, and a write failure only warns.

## Out of scope

Shell completion, help text for the TUI keys.
