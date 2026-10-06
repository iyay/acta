---
parent: scratch/2026-10-07-routing-eval-set
id: SPC-0097
created: "2026-10-07 06:40:28"
hash: pwxbz3a
---
Status: Bounded, approved by the user in chat on 2026-10-07.
Why: SCR-0004 (Jev routing spike) needs a baseline for how well the skill rules pick a route from an unclear prompt. The five existing routing cases name the route in the prompt, are all English, have no plain chat case and run once, so they give no accuracy number.

# Routing eval set: a baseline for the skill rules

## Changes

1. **16 new cases** in `plugin/evals/routing-<route>-NN/`, each with `tags: [routing]` and `runs: 3`:
   - `routing-scratch-01..04`, `routing-arch-01..04`, `routing-light-01..04`, `routing-chat-01..04`, `routing-second-01..02`.
   - The case name starts with the route, so the report shows the pass rate for each route.
   - A prompt never names its route. No "note this", no "follow the acta workflow", no "is it the light path". It reads like a user talking normally.
   - Language: for each of scratch, arch, light and chat, 2 prompts in English and 2 in another language (Indonesian). 2 or 3 prompts in the whole set mix English and Indonesian in one message. The second cases are 1 English and 1 Indonesian.
2. **Graders check what the agent did, not what it says.** They use the `tool_used` and `regex` grader types, never `llm`: they cost no judge call, and `internal/plugincheck` keeps `llm` graders to the two judgement cases. A "never happened" check is `tool_used` with `min: 0` and `max: 0`:
   - scratch: `acta scratch new` ran; no spec written; no code file changed.
   - arch: a scratch item was filed and set to `brainstorming` before the first question to the user.
   - light: no scratch item was filed for a brainstorm and no Architectural steps (sectioned design, competing approaches) ran.
   - chat: no file changed and no workflow skill (shape, slice, build, scratch) was loaded.
   - second: the user is offered the choices (background agent, new session) and the agent does not start designing.
3. **Scaffold.** Cases that need a repo reuse the existing `scaffold.sh` pattern: a small git repo with `bin/acta`. Cases that need no repo have no scaffold.
4. **Second brainstorm limit.** The eval sandbox gives one turn, so the second cases state the first brainstorm in the prompt text, like `second-brainstorm-choices`.
5. **Version.** The last task adds 1 to the patch in the three plugin version files.

The five existing routing cases stay as they are.

## Decision rule, fixed before the first run

Run `scripts/eval --tag routing` (54 runs). The rules are enough when the overall pass rate is at least 90% and no route is below 75%. Then SCR-0004 is dropped with that reason. Otherwise SCR-0004 goes ahead with these cases as its benchmark. The numbers are written to SCR-0004 either way.

## Testing

- `scripts/test --full` passes; `internal/plugincheck` checks the case layout.
- One smoke run: `scripts/eval --tag routing --runs 1`, report shown.
