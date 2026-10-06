---
id: DBT-0082
hash: zjn6q5z
parent: plans/2026-10-06-setup-block-refresh
---
# Review NOTEs: Setup block refresh Implementation Plan

- [ ] (high) acta state set with nothing on stdin clears that part and auto-commits (internal/write/state.go:47), so an agent that forgets the pipe wipes its own Next; the hook line at internal/hook/hook.go:43 does not say stdin either
- [ ] (medium) No test checks that the repo CLAUDE.md acta block matches the setup skill template, so either can drift without a red test
