---
created: "2026-09-30"
parent: specs/2026-09-30-omp-harness-design
id: PLN-0055
hash: d2khlda
started: "2026-09-30"
finished: "2026-09-30"
---
# omp Brainstorm Block Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** In omp, a second Architectural brainstorm in one session is blocked, the first one is remembered, and the "already brainstormed" reminder shows up, the same as in Claude Code.

**Architecture:** The omp extension feeds `acta hook pre-tool`, `post-tool` and `prompt` the same stdin JSON Claude Code feeds them (`{"session_id", "tool_input": {"command"}}`). It calls acta with `spawnSync` from `node:child_process`, because `pi.exec` cannot send stdin. No Go code changes.

**Tech Stack:** TypeScript run by omp, `bun:test`.

**Spec:** `.acta/specs/2026-09-30-omp-harness-design.md` (the "Plan A" section)

**Tests:** fast `cd plugin/omp && bun test`, full `scripts/test --full` (Go; it also runs `internal/plugincheck` over `plugin/`)

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Files: `plugin/omp/index.ts`, `plugin/omp/index.test.ts`, `plugin/omp/FACTS.md` only. No Go change.
- Any failure (acta missing, a timeout, a throw, an exit other than 0 and 2, an empty session id) stays quiet and lets the tool run. Only exit 2 from `acta hook pre-tool` blocks.
- `tool_call` and `tool_result` act only when `toolName === "bash"`.
- The `before_agent_start` message keeps its shape: `{ message: { customType: "acta", content, display: false } }`, no `attribution` field.
- Comments are plain English a 10-year-old reads back without stopping. They say why, not what.
- The repo has no TypeScript formatter or typecheck. Keep the file's existing style (2 spaces, double quotes, semicolons).

## File Map

| File | Task | Change |
|---|---|---|
| `plugin/omp/index.ts` | 1 | `Exec` becomes `Run` over `spawnSync`; new `payload`; `createState` gains `onToolCall`, `onToolResult`; `contextFor` takes the session id; extension wires `tool_call` and `tool_result` |
| `plugin/omp/index.test.ts` | 1 | tests move to the fake `Run`; new tests for block, record, payload, bash-only |
| `plugin/omp/FACTS.md` | 1 | new section: the real two-brainstorm check in omp |

## Waves

- Wave 1: Task 1.

---

### Task 1: The omp extension feeds acta's session hooks

**Files:**
- Modify: `plugin/omp/index.ts`
- Test: `plugin/omp/index.test.ts`
- Modify: `plugin/omp/FACTS.md` (append a section at the end)

**verify:** A bash tool call is blocked in omp only when `acta hook pre-tool` exits 2, and then the reason is acta's stderr. List every path that reaches `tool_call` and `tool_result` (non-bash tool, acta missing, throw, timeout, exit 0, exit 1, exit 2, empty session id, failed bash result) and what each returns or records. Also: every acta hook call gets the session id on stdin in the shape `hook.ParseEvent` reads.

**Interfaces:**
- Consumes: `acta hook session-start --known <file>`, `acta hook prompt`, `acta hook pre-tool`, `acta hook post-tool` (existing CLI, `internal/cli/hook.go`). All read `{"session_id": "...", "tool_input": {"command": "..."}}` on stdin; `pre-tool` exits 2 with the reason on stderr to block.
- Produces:
  - `export type Run = (args: string[], stdin: string, cwd?: string) => { stdout: string; stderr: string; code: number }`
  - `export const realRun: Run`
  - `export function payload(sessionId: string, command?: string): string`
  - `export function createState(run: Run, readDefaults?: () => string)` returning `{ reset(): void; contextFor(sessionId: string, cwd?: string): string; onToolCall(sessionId: string, command: string, cwd?: string): { block: true; reason: string } | undefined; onToolResult(sessionId: string, command: string, cwd?: string): void }`
  - `export default function acta(pi: any, run?: Run)`

- [x] **Step 1: Write the failing tests**

Replace the whole of `plugin/omp/index.test.ts` with:

```ts
import { describe, expect, test } from "bun:test";
import acta, { createState, payload, type Run } from "./index";

type Out = { stdout?: string; stderr?: string; code: number };

// fakeRun answers by the first matching prefix of "acta <args>" and records
// every call with its stdin, so tests can check what acta was handed.
function fakeRun(outputs: Record<string, Out | "throw">) {
  const calls: { key: string; stdin: string; cwd?: string }[] = [];
  const run: Run = (args, stdin, cwd) => {
    const key = ["acta", ...args].join(" ");
    calls.push({ key, stdin, cwd });
    for (const [prefix, out] of Object.entries(outputs)) {
      if (!key.startsWith(prefix)) continue;
      if (out === "throw") throw new Error("boom");
      return { stdout: out.stdout ?? "", stderr: out.stderr ?? "", code: out.code };
    }
    return { stdout: "", stderr: "not found", code: 127 };
  };
  return { run, calls };
}

describe("payload", () => {
  test("has the shape acta hook reads", () => {
    expect(JSON.parse(payload("S1", "ls"))).toEqual({ session_id: "S1", tool_input: { command: "ls" } });
    expect(JSON.parse(payload("S1"))).toEqual({ session_id: "S1", tool_input: { command: "" } });
  });
});

describe("createState", () => {
  test("first message gets rules and reminder, later ones only the reminder", () => {
    const { run } = fakeRun({
      "acta hook session-start": { stdout: "RULES\n", code: 0 },
      "acta hook prompt": { stdout: "REMINDER\n", code: 0 },
    });
    const s = createState(run, () => "DEFAULTS");
    expect(s.contextFor("S1")).toBe("RULES\n\nREMINDER");
    expect(s.contextFor("S1")).toBe("REMINDER");
    s.reset();
    expect(s.contextFor("S1")).toBe("RULES\n\nREMINDER");
  });

  test("session-start passes the known-plugins file", () => {
    const { run, calls } = fakeRun({ "acta": { stdout: "X", code: 0 } });
    createState(run, () => "").contextFor("S1");
    expect(calls[0].key).toMatch(/^acta hook session-start --known .*hooks\/workflow-plugins\.txt$/);
  });

  test("prompt gets the session id on stdin, in the given cwd", () => {
    const { run, calls } = fakeRun({ "acta": { stdout: "X", code: 0 } });
    createState(run, () => "").contextFor("S7", "/work");
    const prompt = calls.find((c) => c.key === "acta hook prompt")!;
    expect(JSON.parse(prompt.stdin).session_id).toBe("S7");
    expect(prompt.cwd).toBe("/work");
  });

  test("missing acta falls back to the default rules and no reminder", () => {
    const { run } = fakeRun({});
    const out = createState(run, () => "DEFAULTS\n").contextFor("S1");
    expect(out).toContain("DEFAULTS");
    expect(out).toContain("acta binary is not installed or failed");
  });

  test("failing or throwing acta counts as missing", () => {
    for (const out of [{ stdout: "partial", code: 1 }, "throw"] as const) {
      const { run } = fakeRun({ "acta": out });
      const text = createState(run, () => "DEFAULTS").contextFor("S1");
      expect(text.startsWith("DEFAULTS")).toBe(true);
      expect(text).not.toContain("partial");
    }
  });

  test("onToolCall blocks only on exit 2, with stderr as the reason", () => {
    const blocked = fakeRun({ "acta hook pre-tool": { stderr: "acta: this session already brainstormed X.\n", code: 2 } });
    expect(createState(blocked.run).onToolCall("S1", "acta set scratch/b status brainstorming")).toEqual({
      block: true,
      reason: "acta: this session already brainstormed X.",
    });
    expect(JSON.parse(blocked.calls[0].stdin)).toEqual({
      session_id: "S1",
      tool_input: { command: "acta set scratch/b status brainstorming" },
    });
    for (const out of [{ code: 0 }, { code: 1, stderr: "oops" }, { code: 127 }, "throw"] as const) {
      const { run } = fakeRun({ "acta hook pre-tool": out });
      expect(createState(run).onToolCall("S1", "ls")).toBeUndefined();
    }
  });

  test("onToolResult hands the command to post-tool and never throws", () => {
    const ok = fakeRun({ "acta hook post-tool": { code: 0 } });
    createState(ok.run).onToolResult("S1", "acta set scratch/a status brainstorming", "/work");
    expect(ok.calls[0].key).toBe("acta hook post-tool");
    expect(JSON.parse(ok.calls[0].stdin).tool_input.command).toBe("acta set scratch/a status brainstorming");
    expect(ok.calls[0].cwd).toBe("/work");
    const bad = fakeRun({ "acta hook post-tool": "throw" });
    expect(() => createState(bad.run).onToolResult("S1", "ls")).not.toThrow();
  });
});

describe("extension", () => {
  function load(outputs: Record<string, Out | "throw">) {
    const handlers: Record<string, Function> = {};
    const pi = { on: (event: string, fn: Function) => { handlers[event] = fn; } };
    const fake = fakeRun(outputs);
    acta(pi, fake.run);
    const ctx = { sessionManager: { getSessionId: () => "S9", getCwd: () => "/work" } };
    return { handlers, ctx, calls: fake.calls };
  }

  test("registers the five events", () => {
    const { handlers } = load({});
    expect(Object.keys(handlers).sort()).toEqual([
      "before_agent_start", "session_compact", "session_start", "tool_call", "tool_result",
    ]);
  });

  test("injects an acta message with the session id passed to prompt", async () => {
    const { handlers, ctx, calls } = load({
      "acta hook session-start": { stdout: "RULES", code: 0 },
      "acta hook prompt": { stdout: "REMINDER", code: 0 },
    });
    const r = await handlers["before_agent_start"]({ type: "before_agent_start", prompt: "hi" }, ctx);
    expect(r.message.customType).toBe("acta");
    expect(r.message.content).toBe("RULES\n\nREMINDER");
    expect(r.message.display).toBe(false);
    expect("attribution" in r.message).toBe(false);
    expect(JSON.parse(calls.find((c) => c.key === "acta hook prompt")!.stdin).session_id).toBe("S9");
    const again = await handlers["before_agent_start"]({ type: "before_agent_start", prompt: "hi" }, ctx);
    expect(again.message.content).toBe("REMINDER");
    await handlers["session_compact"]({});
    const afterCompact = await handlers["before_agent_start"]({ type: "before_agent_start", prompt: "hi" }, ctx);
    expect(afterCompact.message.content).toBe("RULES\n\nREMINDER");
  });

  test("tool_call blocks a bash call when pre-tool exits 2", async () => {
    const { handlers, ctx, calls } = load({ "acta hook pre-tool": { stderr: "NO", code: 2 } });
    const r = await handlers["tool_call"]({ type: "tool_call", toolName: "bash", input: { command: "acta set scratch/b status brainstorming" } }, ctx);
    expect(r).toEqual({ block: true, reason: "NO" });
    expect(calls[0].cwd).toBe("/work");
    expect(JSON.parse(calls[0].stdin).session_id).toBe("S9");
  });

  test("tools other than bash never call acta", async () => {
    const { handlers, ctx, calls } = load({ "acta": { stderr: "NO", code: 2 } });
    expect(await handlers["tool_call"]({ type: "tool_call", toolName: "read", input: { path: "x" } }, ctx)).toBeUndefined();
    await handlers["tool_result"]({ type: "tool_result", toolName: "edit", input: {}, isError: false }, ctx);
    expect(calls.length).toBe(0);
  });

  test("tool_result records a bash call that worked, and skips one that failed", async () => {
    const { handlers, ctx, calls } = load({ "acta hook post-tool": { code: 0 } });
    await handlers["tool_result"]({ type: "tool_result", toolName: "bash", input: { command: "acta set scratch/a status brainstorming" }, isError: false }, ctx);
    expect(calls.map((c) => c.key)).toEqual(["acta hook post-tool"]);
    await handlers["tool_result"]({ type: "tool_result", toolName: "bash", input: { command: "acta set scratch/b status brainstorming" }, isError: true }, ctx);
    expect(calls.length).toBe(1);
  });

  test("a missing session id leaves the tool alone", async () => {
    const handlers: Record<string, Function> = {};
    const { run, calls } = fakeRun({ "acta": { stderr: "NO", code: 2 } });
    acta({ on: (e: string, fn: Function) => { handlers[e] = fn; } }, run);
    const r = await handlers["tool_call"]({ type: "tool_call", toolName: "bash", input: { command: "ls" } }, {});
    expect(r).toBeUndefined();
    expect(calls.length).toBe(0);
  });
});
```

- [x] **Step 2: Run the tests to see them fail**

Run: `cd plugin/omp && bun test`
Expected: FAIL. `payload` is not exported, `contextFor` returns a Promise, and the extension registers only three events.

- [x] **Step 3: Write the implementation**

Replace the whole of `plugin/omp/index.ts` with:

```ts
// acta for omp. omp does not run Claude Code's shell hooks, so this extension
// sends the same text through omp's own events: the session rules on the first
// message and after compaction, the voice reminder on every message, and the
// one-brainstorm-per-session check around every bash call.
import { spawnSync } from "node:child_process";
import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const pluginRoot = join(dirname(fileURLToPath(import.meta.url)), "..");
const knownFile = join(pluginRoot, "hooks", "workflow-plugins.txt");
const defaultsFile = join(pluginRoot, "hooks", "default-rules.md");

export type Run = (args: string[], stdin: string, cwd?: string) => { stdout: string; stderr: string; code: number };

// realRun calls acta directly. pi.exec cannot send stdin, and acta's hooks
// read the session id from stdin. A missing acta or a timeout comes back as a
// failed exit, not a throw.
export const realRun: Run = (args, stdin, cwd) => {
  const r = spawnSync("acta", args, { input: stdin, cwd, timeout: 10000, encoding: "utf8" });
  return { stdout: r.stdout ?? "", stderr: r.stderr ?? "", code: r.status ?? 1 };
};

// payload is the same JSON Claude Code hands its hooks, so one acta command
// serves both harnesses.
export function payload(sessionId: string, command = ""): string {
  return JSON.stringify({ session_id: sessionId, tool_input: { command } });
}

// call never throws. A hook that breaks must not stop the user's work.
function call(run: Run, args: string[], stdin: string, cwd?: string) {
  try {
    return run(args, stdin, cwd);
  } catch {
    return { stdout: "", stderr: "", code: 1 };
  }
}

export function createState(run: Run, readDefaults: () => string = () => readFileSync(defaultsFile, "utf8")) {
  let rulesSent = false;
  return {
    reset() {
      rulesSent = false;
    },
    contextFor(sessionId: string, cwd?: string): string {
      const parts: string[] = [];
      if (!rulesSent) {
        const r = call(run, ["hook", "session-start", "--known", knownFile], "", cwd);
        parts.push(
          r.code === 0
            ? r.stdout.trim()
            : readDefaults().trim() +
                "\n\nacta: the acta binary is not installed or failed, so these are the default rules (English, adhd style).",
        );
        rulesSent = true;
      }
      const reminder = call(run, ["hook", "prompt"], payload(sessionId), cwd);
      if (reminder.code === 0 && reminder.stdout.trim()) parts.push(reminder.stdout.trim());
      return parts.join("\n\n");
    },
    // Only exit 2 blocks. Every other answer lets the command run, because a
    // hook that guesses wrong stops real work the user asked for.
    onToolCall(sessionId: string, command: string, cwd?: string) {
      const r = call(run, ["hook", "pre-tool"], payload(sessionId, command), cwd);
      return r.code === 2 ? { block: true as const, reason: r.stderr.trim() } : undefined;
    },
    onToolResult(sessionId: string, command: string, cwd?: string): void {
      call(run, ["hook", "post-tool"], payload(sessionId, command), cwd);
    },
  };
}

export default function acta(pi: any, run: Run = realRun) {
  const state = createState(run);
  // acta finds its planning folder from the working dir, so every call runs
  // where the omp session runs.
  const where = (ctx: any) => ({
    id: String(ctx?.sessionManager?.getSessionId?.() ?? ""),
    cwd: ctx?.sessionManager?.getCwd?.() as string | undefined,
  });
  pi.on("session_start", async () => state.reset());
  pi.on("session_compact", async () => state.reset());
  pi.on("before_agent_start", async (_event: any, ctx: any) => {
    const w = where(ctx);
    const content = state.contextFor(w.id, w.cwd);
    if (!content) return;
    // Shape follows BeforeAgentStartEventResult: no attribution field.
    return { message: { customType: "acta", content, display: false } };
  });
  pi.on("tool_call", async (event: any, ctx: any) => {
    const w = where(ctx);
    // With no session id acta cannot tell sessions apart, so it stays out.
    if (event?.toolName !== "bash" || !w.id) return;
    return state.onToolCall(w.id, String(event.input?.command ?? ""), w.cwd);
  });
  pi.on("tool_result", async (event: any, ctx: any) => {
    const w = where(ctx);
    // A failed or blocked call did not brainstorm anything, so it is not
    // recorded. Claude Code's post-tool hook also runs only on success.
    if (event?.toolName !== "bash" || event?.isError || !w.id) return;
    state.onToolResult(w.id, String(event.input?.command ?? ""), w.cwd);
  });
}
```

- [x] **Step 4: Run the tests to see them pass**

Run: `cd plugin/omp && bun test`
Expected: PASS, every test in `index.test.ts`.

- [x] **Step 5: Run the real check in omp and record it**

This needs the `acta` on PATH (any build since SCRATCH-6 landed has `acta hook pre-tool`). Replace `<worktree>` with the absolute path of this worktree. Run from a fresh temp folder:

```bash
T="$(mktemp -d)" && cd "$T" && git init -q && git commit -q --allow-empty -m init
printf 'first\n' | acta scratch new first-idea --title first
printf 'second\n' | acta scratch new second-idea --title second
D="$(date +%F)"
PM_VOICE_FILE="$T/none.yaml" command omp --no-session --no-extensions --no-rules \
  --plugin-dir <worktree>/plugin -e <worktree>/plugin/omp/index.ts -p \
  "Run these two shell commands, each as its own bash call, one after the other. First: acta set scratch/$D-first-idea status brainstorming. Second: acta set scratch/$D-second-idea status brainstorming. Then quote, word for word, any error the second call gave you."
cat .acta/state/sessions.json
```

Expected: the reply quotes `acta: this session already brainstormed` and names the first item. `sessions.json` maps one session id to `<date>-first-idea` only. If omp answers some other way, stop and report the output; do not change acta's Go code.

Append this section to the end of `plugin/omp/FACTS.md`, with the real commands and output pasted in (paths written as `$T` and `<worktree>`, never an absolute user path, because `internal/plugincheck` rejects those):

````markdown
## Brainstorm block in omp

Checked <date> with omp <version> and acta on PATH. The extension hands
`acta hook pre-tool` and `post-tool` the same stdin JSON Claude Code does, so
the second brainstorm in one omp session is blocked.

```text
<the commands from this step and their real output>
```
````

- [x] **Step 6: Commit**

```bash
git add plugin/omp/index.ts plugin/omp/index.test.ts plugin/omp/FACTS.md
git commit -m "omp extension feeds acta's session hooks: second brainstorm blocked, reminder shown"
```
