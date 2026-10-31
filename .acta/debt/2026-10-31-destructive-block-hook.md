---
id: DBT-0151
hash: y1q0nge
parent: plans/2026-10-31-destructive-block-hook
---
# Review NOTEs: Destructive command block hook implementation plan

- [ ] (medium) Shell keywords hide the command word in DestructiveBlock: `if x; then rm -rf y; fi`, `for d in a; do rm -rf $d; done`, `{ rm -rf x; }` and `! rm -rf x` all pass, because the reader does not skip then, do, else, {, or !. Agents write loops like this often.
- [ ] (medium) DestructiveBlock can be bypassed through other shells and wrappers: `bash -c "rm -rf x"`, `sh -c`, `eval`, `exec`, `timeout 5`, `nice` and backticks all pass; only `$(...)` is caught.
- [ ] (medium) A heredoc that writes a script is blocked by mistake: `cat > s.sh <<'EOF' ... rm -rf build ... EOF` blocks because the body is split on newlines, against the let-through-on-doubt rule.
- [ ] (medium) headIsMerge in internal/hook/destructive.go runs git with no timeout; the only bound is the harness hook timeout (10 s), so a stuck git costs about 10 s per Bash call before it fails open.
