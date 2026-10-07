---
id: BUG-0030
hash: a1ivser
started: "2026-10-05 13:06:16"
fixed_in: fa9088668ec3781c40e2bb7689006146edac374f
finished: "2026-10-05 13:29:24"
---
# omp gets no acta session rules

## Symptom
After new (or resume or fork) session in a running omp, the agent gets only the one-line per-turn reminder ("acta config: reply in Indonesian, adhd style.") and never the acta session rules block. It then skips the acta flow: in be-pmis on 2026-10-05 it wrote a plan file straight into .acta/plans with no shape, no slice and no yes.

## Root cause
plugin/omp/index.ts:180-181 listens only to "session_start" and "session_compact". omp emits "session_switch" with reason "new", "resume" or "fork" for those actions (found in @oh-my-pi/pi-coding-agent dist: emit({type:"session_switch",reason:"new",...})). So state.reset never runs, rulesSent (index.ts createState) stays true from the earlier session, and contextFor (index.ts ~121) sends only the reminder.

## Repro
Evidence: in ~/.omp/agent/sessions/-Nayakatara-PMIS-codes-febe-be-pmis/, session 2026-10-05T05-09-43 got the rules (first acta message 4436 chars, no user message), then session 2026-10-05T05-21-00 in the same omp got only 45-char acta messages from its first turn. Session 2026-10-05T03-57-59 shows the same. Fresh omp processes (all dispatch tabs) get the rules (4729-4756 chars).
Steps: start omp in a repo with acta, send one message, run /new, send a message; the acta custom message holds only the reminder.
Test shape: in plugin/omp/index.test.ts, fire session_start, before_agent_start, then session_switch {reason:"new"}, then before_agent_start; the second context must hold the rules again.

## Found in
main at 7ce8c03, acta:debug on a be-pmis omp session the user pasted.
