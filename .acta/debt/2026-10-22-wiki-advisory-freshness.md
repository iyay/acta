---
id: DBT-0139
hash: srbjgel
parent: plans/2026-10-22-wiki-advisory-freshness
---
# Review NOTEs: Advisory wiki freshness

- [ ] (medium) The existing build-close-no-suite eval fails on both main and the advisory branch: scripts/eval --case build-close-no-suite reports broad-suite-ran present (score 0.67) because the agent runs scripts/test during Close despite the existing no-suite rule. The user explicitly waived this known failure for PLN-0187 landing; no fix or grader change is included.
