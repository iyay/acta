---
id: DBT-0160
hash: ot72zwx
parent: plans/2026-11-07-probe-copy-hardening
---
# Review NOTEs: Probe copy hardening

- [ ] (medium) After D=$(bash probe-copy.sh) && cd "$D", a GIT_DIR the caller shell still exports (inside a git hook) sends later git calls in the copy to the real repo; the script can only clear it for its own git init.
