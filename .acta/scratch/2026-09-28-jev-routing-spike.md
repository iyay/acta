---
id: SCR-0004
hash: k8c6x55
title: jev-routing-spike
status: raw
created: "2026-09-28"
---
# Spike: Jev model for routing decisions

User idea, 2026-09-27: use TypeSafe's Jev model (docs.typesafe.ai; typed Choice/Score/Noul questions against a state, returns probabilities and confidence) for routing decisions: scratch vs brainstorm vs bounded fix vs chat, and second-brainstorm detection.

Agent's take: optional hook-level classifier, off by default, needs an API key. Privacy concern: user prompts leave the machine. Low confidence falls back to the skill rules. Measure with the eval suite before and after. Not verified yet: API shape, pricing, data retention.

Ruling, 2026-09-27: Jev is a spike, run after the harness spec lands with its eval suite, reusing the eval scenarios as the benchmark (rules baseline vs Jev). Needs a TypeSafe API key from the user.

Research, 2026-10-07 (docs.typesafe.ai, typesafe.ai/legal):

- API: one endpoint, `POST https://api.typesafe.ai/v1/systemone`, `Authorization: Bearer <key>`. Body is `{state, model, questions:{name:{type, instructions, criteria?}}}`. `state` is text, a JSON object or a text array. Every call carries its whole state, so there is no session.
- Question types: `choice` (up to 255 options) returns `choice`, `confidence` and `probabilities`. `score` returns `score`, `confidence`, `legend` and `probabilities`. `noul` returns only the yes probability, with no confidence field; work it out as |2p-1|.
- Pricing: $0.042 per million input tokens, output is free. No free tier found.
- Data: not used for training. A DPA with EU SCCs exists. Data is stored and processed in the US. Retention is "as long as necessary", with no fixed period. Zero retention is for enterprise only.
- SDKs: Python and JS, both pre-1.0 with recent breaking changes. There is no Go SDK, so acta would call plain HTTP.
- Limits: 64k context (32k for the state plus the longest question), 100K tokens/s and 80 requests/s, which "can change without notice". Errors to handle: 401, 422, 429, 529. No latency numbers are published.
- Model: `jev-1.13.0`. `jev-latest` is an alias that moves, so pin the version before tuning thresholds.

Risks for this spike:
- Privacy: user prompts leave the machine and stay there for no fixed period.
- Language: English is primary, and other languages are "not equally well" handled. Indonesian prompts must be measured in the eval.
- Latency is unknown and rate limits may change. A hook that waits on the call needs a timeout that falls back to the skill rules.
- Second-brainstorm detection only works if the hook puts the earlier session context into `state` itself.

Agent's take: the API fits the idea and the cost is close to zero. Keep it opt-in and off by default, pin the version, and set a short timeout.

2026-10-07: blocked by the routing eval set (see the new scratch item routing-eval-set). The existing eval cases name the route in the prompt, so they cannot benchmark rules against Jev.

2026-10-07: the routing eval set landed (SPC-0097, SPC-0098; version 0.1.30). The baseline run `scripts/eval --tag routing` did not start: all 54 runs refused because ~/.docker holds symlinks into Docker.app, which the Bash sandbox rejects. Parked by the user. To resume: move ~/.docker aside (Docker Desktop closed), run the baseline, move it back, then apply the SPC-0097 decision rule (overall >= 90%, no route below 75%). The light number is a floor.

2026-10-07: the routing cases moved to plugin/evals-routing (PLN-0108, 0.1.31), so the land gate no longer runs them. The baseline command is now `scripts/eval --eval-dir evals-routing` (not --tag routing).
