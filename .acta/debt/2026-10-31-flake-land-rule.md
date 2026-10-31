---
id: DBT-0150
hash: s1auoke
parent: plans/2026-10-31-flake-land-rule
---
# Review NOTEs: Flake land rule implementation plan

- [ ] (medium) The land eval gate reruns each red case alone up to 2 times with no cap on how many cases are red, so a wide red set can cost more quota than one whole-set rerun; flakerun caps Go reruns at more than 5 failed tests, the eval gate has nothing like it.
