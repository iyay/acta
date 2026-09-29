---
id: BUG-5
hash: uzbc
---
# acta doctor tells you to run an omp command that does not exist

## Symptom
When omp has a dead plugin link, `acta doctor` warns `stale-links` and prints `fix: omp plugin unlink pm`. Running that fails with `error: Expected action to be one of: install, uninstall, list, link, doctor, ...; got "unlink"`. Running `omp plugin uninstall pm` does not work either. It prints `pm is not installed`, because the link is only a leftover symlink in `node_modules`.

## Root cause
`internal/doctor/doctor.go:220` builds the hint as `"omp plugin unlink "+en.Name()`. omp has no `unlink` action. The tests at `internal/doctor/doctor_test.go:145` and `:167` check for that same wrong text, so they pass.

## Repro
1. `ln -s /nonexistent ~/.omp/plugins/node_modules/pm`
2. `acta doctor` shows `fix: omp plugin unlink pm`
3. Run `omp plugin unlink pm`. It errors because the action does not exist.

The only command that cleared the warning was `rm ~/.omp/plugins/node_modules/pm`.

## Found in
main, while running /acta:setup on 2026-09-29.
