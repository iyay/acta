---
id: DBT-0044
hash: l5yicid
parent: plans/2026-09-30-omp-executor-and-config-command
---
# omp executor and config command review notes

- [ ] (low) skill_build_test Must "On omp, `dispatch` runs as `subagent`" also matches the table row; the agent()/do-not-ask sentence has no test pinning it.
- [ ] (medium) plan/SKILL.md executor question still offers dispatch with no omp caveat; setup/SKILL.md offers dispatch inside omp too, where build runs it as subagent.
- [ ] (low) default-rules.md and hook skill list describe acta:dispatch with no "not from omp" hint; only the Step -3 guard covers it.
- [ ] (low) SessionStart broken-file text says "voice file" while Prompt says "config file"; doctor still says "no voice file yet" / "voice and build executor are set".
- [ ] (low) cli.go unknown-command hint lists neither config nor hook (old gap); acta voice now prints that hint, so the new name is not discoverable there.
- [ ] (low) internal/cli/cli_test.go test names still TestVoice* while calling config; only cmd/acta tests were renamed.
- [ ] (medium) Step -3 omp guard is text only; no Go or eval check exercises an omp session reaching acta:dispatch.
