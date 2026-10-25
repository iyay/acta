---
type: Gotcha
title: LLM eval graders need trace focus for tool evidence
description: Use focus trace when an LLM grader checks tool calls; the default only shows the final reply
paths: [plugin/evals/cold-read-accept/graders/, plugin/evals/cold-read-offer/graders/, plugin/evals/cold-read-skip/graders/, plugin/evals/cold-read-unavailable/graders/]
---
In Claude Code 2.1.310, an LLM behavior grader without `focus: trace` saw only the
final reply. Cold read dispatched the right reviewer, but the grader rejected
the run because the final reply could not prove the dispatch inputs.

Use `type: llm` with `focus: trace` when judgment depends on tool calls, their
inputs, or ordering. Keep deterministic checks for facts they can prove.

Native child events appear in the kept root `out/trace.jsonl` with
`parent_tool_use_id`. Inspect the child's initial message and successful reads,
as well as the parent's dispatch input. A tool count or a parent's claim of
isolation alone cannot show that conversation history was excluded.
