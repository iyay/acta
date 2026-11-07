---
id: SPC-0194
created: "2026-11-07 20:02:46"
hash: pa5knib
---
# Probe copy hardening

Bounded, with a written spec because it touches a shared temp folder. The user approved the design on WhatsApp on 2026-11-07. It works the three items of DBT-0159, from the review of PLN-0217.

## Problem

`plugin/skills/debug/probe-copy.sh` makes the temp copy a debug probe works in. Three gaps remain:

1. Its fallback runs `mkdir -p "$base/acta-probe-$$"`. `-p` accepts a folder that already exists, so a folder or symlink planted at that guessable name in a shared `/tmp` receives a full copy of the repo.
2. It copies the current folder. Run from a subfolder, it copies only that subfolder, and the probe runs on a partial tree.
3. The copy has no `.git`, so git run there walks up parent folders. With `TMPDIR` inside another repo, git in the copy acts on that repo.

## Out of scope

- No change to `plugin/skills/debug/SKILL.md`. The wiki page `skill-escape-lines-skip-proof` shows that skill text changes move agent behaviour, so this fix stays inside the script.
- No eval run: the skill text does not change.

## Design

All in `probe-copy.sh`:

1. The fallback folder is made with `mkdir -m 700 "$base/acta-probe-$$"`, without `-p`. A folder or symlink already there makes the script exit non-zero with no output.
2. The source is the output of `git rev-parse --show-toplevel` (stderr hidden) when it exits 0, else `pwd -P`. The copy and the inside-the-repo guard both use that source. A folder inside a parent repo copies the whole parent work tree on purpose, since it is part of that tree.
3. Only when the source came from `git rev-parse`, the script runs `git init -q` in the copy after the copy, so git there finds the copy's own empty repo and never walks up. A failed `git init` makes the script exit non-zero with no output. A plain folder that is not a git work tree gets no `.git`.

The header comment says what each step guards against, in plain words, and no longer says the copy leaves out `.git`.

## Testing

Red first in `internal/plugincheck/probe_copy_test.go`. The rule "copy must not hold .git" changes to: a `.git` in the copy, when there is one, is a folder (never the copied pointer file), and `git rev-parse --show-toplevel` run in the copy prints the copy's own path. Paths are compared after `filepath.EvalSymlinks` (on macOS `/var` is `/private/var`). Fake repos whose `.git` git cannot read take the `pwd -P` branch and keep "no .git in the copy". Three new cases:

- A folder already sits at the fallback path: the test runs `bash -c 'mkdir "$TMPDIR/acta-probe-$$"; exec bash "$0"' <script>` with mktemp stubbed to fail, so `exec` keeps the PID and the name matches. Non-zero exit, no output, and the planted folder stays empty. The script gets no test-only override.
- Run from a subfolder of a real `git init` repo: the copy holds the root files.
- `TMPDIR` inside another `git init` repo: `git rev-parse --show-toplevel` run in the copy prints the copy's own path.

The existing seven cases still pass under the changed rule. `scripts/test` green; the full suite runs at land.
