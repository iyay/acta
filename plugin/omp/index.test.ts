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