---
id: SPC-0078
created: "2026-10-05 14:25:27"
hash: l31a9vj
started: "2026-10-05 14:36:45"
---
# Hand-offs in the repo language, and evals may use Read

Status: Bounded, approved by the user in chat on 2026-10-05.

Why:
- omp wrote a subagent brief in Indonesian. The session voice rules name only two cases: chat to the user (chat language) and files in the repo (repo language). A prompt to a subagent or another agent is neither, so the agent used the chat language.
- The `probe-round` eval fails at random. `scripts/eval` grants only Bash, so when the agent reads `plugin/skills/shape/probe.md` with the Read tool, the read is denied ("Permission to use Read has been denied because Claude Code is running in don't ask mode") and the agent asks plain questions instead of the probe format. With `cat` it passes.

Design:
1. The voice block from `acta hook session-start` (`internal/hook/hook.go` SessionStart) gets one more line after the repo line: `- Write prompts and hand-offs to subagents or other agents in <repo language>.` The fallback `plugin/hooks/default-rules.md` gets the same line (regenerated the way its test does it).
2. `scripts/eval` grants `Read` next to `Bash`. `plugin/evals/FACTS.md` gets one dated line: why Read is granted.
3. The last task adds 1 to the patch version in the three plugin files.

Tests: the session start text holds the new line with the repo language from the config; default-rules.md matches. The eval change is checked by running `scripts/eval --case probe-round` three times with all green.
