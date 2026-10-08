---
id: DBT-0098
hash: a4z4od9
parent: plans/2026-10-08-omp-eval-stop-at-final-reply
---
# Review NOTEs: omp eval runner stops at the first final reply and runs omp in a clean home

- [ ] (medium) ParseStream ends the run at any assistant message_end with stopReason stop; if omp streams a subagent's message with that shape, an eval case would end early and grade a half-done run. Record omp's stream shape for subagents (task tool) in plugin/omp/FACTS.md and match only the top agent's message.
- [ ] (low) RunCase in internal/evalomp/run.go can lose omp's own error and stderr text when omp exits with an error between the exited check and the group kill; the caller then sees only the parse error.
- [ ] (low) omp now runs with a throwaway HOME that holds only .omp; anything else omp's runtime keeps under HOME (a cache) starts empty each case. Confirm on a real eval run and note it in plugin/omp/FACTS.md.
