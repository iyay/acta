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
