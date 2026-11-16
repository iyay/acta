---
created: "2026-11-16 14:45:43"
id: SPC-0213
hash: yfds0mn
---
# Gate audit test brings its own acta binary

## Why

The v0.1.55 Release run 43825183567 failed in its Test job. `TestGateWorktreeOwnerAudit` runs `plugin/evals/land-tidy-temp-worktree/scaffold.sh`. The scaffold copies the acta on PATH with `cp "$(command -v acta)" bin/acta.real`. The GitHub runner has no acta on PATH, so `cp` gets an empty source and prints `cp: cannot stat ''`. The same failure shows up locally when PATH leaves out `~/.local/bin`. The user approved this design on WhatsApp (message 3AF48A56B34C8B8FCB89, 14:45 on 2026-11-16).

## Design

The test builds `./cmd/acta` once into a temp dir before its subtests run. It puts that dir first on PATH in the env that `run()` uses. The scaffold then copies the binary built from this checkout, on every machine.

The scaffold and its guard stay as they are. Only `internal/plugincheck/gate_worktree_owner_test.go` changes. No new dependency.

## Verification

Red first: run the test with PATH stripped of every dir that holds an acta binary, and watch it fail with `cp: cannot stat ''`. Green: the same command passes after the change. Then the package suite and `go vet`. The full suite runs at land.

## After land

Cut v0.1.56 with `scripts/release patch`. The user typed approval to push main and the new tag (same WhatsApp message).
