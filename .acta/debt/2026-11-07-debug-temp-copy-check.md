---
id: DBT-0159
hash: fri5iad
parent: plans/2026-11-07-debug-temp-copy-check
started: "2026-11-07 20:11:25"
---
# Review NOTEs: Debug temp copy check

- [ ] (medium) probe-copy.sh mkdir -p fallback accepts a folder that already exists at a guessable /tmp name; plain mkdir would refuse a planted folder or symlink.
- [ ] (medium) probe-copy.sh does not check it runs from the repo root; from a subfolder it copies only that subfolder and the probe runs on a partial tree.
- [ ] (medium) The probe copy has no .git, so git there walks up parent folders; with TMPDIR inside another repo, git in the copy acts on that repo. GIT_CEILING_DIRECTORIES would stop it.
