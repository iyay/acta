---
id: BUG-0031
hash: q3xv1ma
---
# A mistyped brainstorm command blocks the real one in the same session

## Symptom
In one session the agent ran `acta set scratch/rate-limiter status brainstorming` (wrong stem, acta answered "bad input: unknown id"), then the right command `acta set scratch/2026-10-05-rate-limiter status brainstorming`. The pre-tool hook blocked the second one: "this session already brainstormed rate-limiter. One Architectural brainstorm per session". The brainstorm never starts. Seen in the probe-round eval on 2026-10-05 (score 0.75, trace /private/tmp/e-aGN3ai/out/trace.jsonl).

## Root cause
internal/hook/session.go RecordBrainstorm (around line 71) records the stem straight from the command text on every post-tool call. It does not check that the command succeeded or that the stem names a real scratch item. PreTool (around line 85) then compares stems as text, so the real stem counts as a second, different brainstorm.

## Repro
In a repo with acta and a scratch item 2026-10-05-x: in one session run `acta set scratch/x status brainstorming` (fails, unknown id), then `acta set scratch/2026-10-05-x status brainstorming`. The second is blocked with exit 2.

## Found in
main at 3e7e903, acta:debug of the flaky probe-round eval while reviewing PLN-0088.
