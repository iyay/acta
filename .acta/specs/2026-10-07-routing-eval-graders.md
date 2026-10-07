---
parent: debt/2026-10-07-routing-eval-set
closes: [DBT-0090.01, DBT-0090.02, DBT-0090.03]
id: SPC-0098
created: "2026-10-07 08:29:50"
hash: ld7jqsu
---
Status: Bounded, approved by the user in chat on 2026-10-07.
Why: the routing eval set (SPC-0097) cannot give a trusted baseline yet. A light case passes when the agent only answers in text, the chat no-skill grader can never fail because no case grants Skill, and Indonesian prompts run with chat_language: English.

# Routing eval graders: tell light from chat, and match the user's language

## Changes

1. **Skill for every routing case.** All 18 `plugin/evals/routing-*` cases set `allowed_tools: [Skill]`, as a real user has it. No route gets a tool the others lack.
2. **Light needs a workflow skill.** In acta a small fix takes the Bounded path: the agent loads a workflow skill, shows a short design and waits for a yes, so it does not edit the file yet. A grader on Edit would fail the right route. Each light case gets `tool_used` on `Skill` with `input_match: "acta:"`, `min: 1`. The existing `max: 0` graders stay.
3. **Chat loads no workflow skill, by either way in.** Each chat case keeps `tool_used` on `Skill` `input_match: "acta:"` `max: 0`, and adds `tool_used` on `Read` with `input_match: "/skills/"`, `min: 0`, `max: 0`, because `scripts/eval` grants Read and an agent can open a SKILL.md with it.
4. **Indonesian prompts run in Indonesian.** The scaffold of every case whose prompt is mainly Indonesian writes `chat_language: Indonesian` in place of `English`. English cases keep `English`. The rest of each scaffold stays as it is.
5. **Test.** `TestRoutingEvalCases` checks: every routing case grants `Skill`; every light case has a Skill `acta:` grader with `min: 1`; every chat case has the Skill and the Read `max: 0` graders; each scaffold's chat_language matches the main language of its prompt.
6. **Version** 0.1.30.

## Known limit

An agent on a light case that reads SKILL.md with Read in place of the Skill tool fails the light grader. That makes the light number a floor, not an exact count. Say so next to the baseline in SCR-0004.

## Testing

`scripts/test ./internal/plugincheck/` while building, `scripts/test --full` at land. No `scripts/eval` run in this plan.
