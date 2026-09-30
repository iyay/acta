---
created: "2026-09-30"
parent: scratch/2026-09-29-omp-harness
id: SPC-0047
hash: r7tywsv
---
# omp harness: the brainstorm block in omp, and an omp eval run

Status: design approved by the user in chat on 2026-09-30, section by section. Architectural: a new hook path in the omp extension and a second eval runner. One spec, two plans: plan A (extension) and plan B (eval runner). They do not depend on each other.

## Why

1. omp does not run Claude Code's shell hooks. So the SCRATCH-6 work (remember which scratch item a session brainstormed, block a second one, remind the agent) does nothing in omp. The prompt reminder is also dead there: `plugin/omp/index.ts` calls `acta hook prompt` with no stdin, and `sessionReminder` (`internal/cli/hook.go:110`) needs the session id from stdin.
2. `claude plugin eval` only tests Claude Code. Nothing checks that the plugin behaves the same in omp.

## Facts checked (omp 18.4.4, types from pi-coding-agent 18.4.3)

1. `pi.on("tool_call", ...)` fires before a tool runs. Returning `{ block: true, reason }` stops the tool, and `reason` goes back to the model as an error.
2. `pi.on("tool_result", ...)` fires after a tool runs.
3. A bash call has `toolName: "bash"` and `input.command`.
4. The handler's second argument has `ctx.sessionManager.getSessionId()` and `getCwd()`.
5. `pi.exec` takes no stdin (`ExecOptions` is `signal`, `timeout`, `cwd`).
6. `omp` has `--mode json`, `--no-session`, `--no-extensions`, `--no-skills`, `--plugin-dir`, `-e`. It has no max-turns flag and no allowed-tools flag.

## Plan A: the omp extension

Files: `plugin/omp/index.ts` and `plugin/omp/index.test.ts` only. No Go change. The extension hands acta the same JSON Claude Code hands it, so one Go path serves both harnesses.

1. **`Run`** replaces `Exec`: `(args, stdin) => { stdout, stderr, code }`. The real one is `spawnSync("acta", args, { input, cwd, timeout: 10000 })` from `node:child_process`. `cwd` is `ctx.sessionManager.getCwd()`, because acta finds its root from the working dir.
2. **`payload(sessionId, command)`** builds `{"session_id": ..., "tool_input": {"command": ...}}`, the shape `hook.ParseEvent` reads. With no command it leaves `command` empty.
3. **`createState(run)`** has three calls:
   - `contextFor(sessionId)`: the session rules as today, then `acta hook prompt` with the payload on stdin, so the "already brainstormed X" reminder shows up in omp.
   - `onToolCall(sessionId, command)`: runs `acta hook pre-tool`. Exit 2 returns `{ block: true, reason: stderr.trim() }`. Anything else returns nothing.
   - `onToolResult(sessionId, command)`: runs `acta hook post-tool` and drops the output.
4. **Wiring.** `tool_call` and `tool_result` act only when `toolName === "bash"`. `before_agent_start` passes the session id to `contextFor`.
5. **Errors.** acta missing, a timeout, a throw, or an exit other than 0 and 2 all stay quiet and let the tool run. This is the same rule as the shell hooks: a hook that is unsure must not stop work.
6. **Tests** (`index.test.ts`, fake `Run`): exit 2 blocks with stderr as the reason; exit 0, exit 1 and a throw do not block; a non-bash tool never calls acta; the payload carries the right session id and command; `prompt` gets the payload on stdin.
7. **Real check.** One `omp -p` session that brainstorms two different scratch items in a row: the second `acta set ... status brainstorming` is blocked. The command and its output go in `plugin/omp/FACTS.md`.

## Plan B: `acta eval-omp`

Usage: `acta eval-omp [--case <name>] [plugin-dir]`. `scripts/eval --omp` passes through to it. Cases stay in `plugin/evals/*`, one source for both harnesses.

1. **Probe facts (run while planning, 2026-09-30, omp 18.4.4).** The first task writes them in `plugin/omp/FACTS.md`:
   - `--mode json` prints one JSON event per line. A tool call is `{"type":"tool_execution_start","toolName":"bash","args":{"command":"..."}}`. The last line is `{"type":"agent_end","messages":[...]}`; the final reply is the text parts of the last `role: "assistant"` message.
   - `--no-skills` also drops the acta skills that `--plugin-dir` brings, so it cannot be used. `--skills=<names>` with the acta skill folder names (from `plugin/skills/*`) keeps acta's skills and filters out every other one. The model then lists exactly the 12 acta skills.
   - The extension from `-e` still runs under `--no-extensions`: the `acta` custom message shows up.
2. **New package `internal/evalomp`**, small files:
   - `case.go`: reads `prompt.md` (frontmatter and body), `case.yaml` (`context.scaffold_script`) and `graders/*.md`. A case with the tag `claude-only` in `tags` is skipped. A tag and not a new field, because `claude plugin eval` rejects an unknown frontmatter key (`plugin/evals/FACTS.md`).
   - `run.go`: per case, an empty temp workspace (no `git init`: the scaffolds refuse a folder that is not empty and make their own repo), then the scaffold when there is one, run with `HOME` set to a throwaway home next to the workspace, then a list of the files that exist, then `omp -p --mode json --no-session --no-extensions --no-rules --skills=<acta skill names> --plugin-dir <plugin> -e <plugin>/omp/index.ts "<prompt>"`, killed after `timeout_seconds`. `PM_VOICE_FILE` points at `<throwaway home>/.acta/config.yaml`. A scaffold writes that file the same way it does in the claude sandbox; a case with no scaffold leaves it missing. Either way the user's own chat language and style stay out, the same as the fresh home `claude plugin eval` uses. It returns the final reply and the list of tool calls (name and input).
   - `grade.go`, following the grader table in `plugin/evals/FACTS.md`: `file_exists` counts only files the run created (they match `path` and were not in the list taken after the scaffold) and checks `exists`. `regex` matches `pattern` (with `flags: i`) against `target`: `last_message` (the default) or `{ source: file, path }`; `match: not_contains` asks for absence. `tool_used` compares `tool` without case (`Bash` matches `bash`), counts the calls whose JSON input matches the `input_match` regex, and checks `min` (default 1) and `max`. Any other grader type, target or match value is a FAIL that names it, never a silent pass. `llm` runs `omp -p --no-session` with omp's own default model (no `--model`), sends the grader body plus the final reply, and needs the answer to start with PASS or FAIL.
   - `report.go`: one line per case, `PASS`, `FAIL` or `SKIP`, plus the graders that failed. Exit 1 when any case fails.
3. **Sandbox.** The run uses the user's own omp auth. `--no-extensions --no-rules --skills=<acta skill names>` keeps the user's other omp plugins, skills and rules out, so the run judges acta alone. omp's own built-in preludes (like its todo reminder) still run; they are part of omp.
4. **Known gaps, on purpose.** `max_turns` and `allowed_tools` are ignored, because omp has neither flag; only `timeout_seconds` caps a run. The `llm` judge votes once, not best of three, to save quota. `second-brainstorm-choices` gets the `claude-only` tag, because its grader asks for `claude --bg`. All four go in `plugin/evals/FACTS.md`.
5. **Errors.** No `omp` on PATH exits 3 (acta's code for a failure that is not bad input) with a plain message. A timeout marks that case FAIL with reason `timeout` and the run moves on. A judge answer that does not start with PASS or FAIL is FAIL.
6. **Tests.** Unit tests for the case parser, each grader (the json stream from a fixture), and the `claude-only` skip. The runner test puts a fake `omp` on PATH, so `scripts/test` costs no quota.

## Out of scope

1. Model choice for omp runs. omp uses its own default for cases and the judge.
2. A max-turns cap in omp.
3. Changing `acta hook` itself. Both harnesses feed it the same stdin JSON.
