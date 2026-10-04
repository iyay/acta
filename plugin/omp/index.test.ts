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

// Only one test reads the real style file. The rest use this stub, which gives no style.
const noStyle = () => "";

// A style file as Claude Code reads it: settings on top, the words below.
const STYLE_FILE = "---\nname: acta\nkeep-coding-instructions: true\n---\n\n## Every reply\nOpen with the answer.\n";
const STYLE_BODY = "## Every reply\nOpen with the answer.";

// A style reader that always fails. Its count shows the code did try to read.
function brokenStyle() {
  const reader = () => {
    reader.reads++;
    throw new Error("no such file");
  };
  reader.reads = 0;
  return reader;
}

// What acta prints when a wiki page covers a file: one JSON object with the text.
const HINT = "wiki: .acta/wiki/tui-wrap.md: Wrap long lines with Hardwrap, not Wrap.";
const hintJson = (text: string) =>
  JSON.stringify({ hookSpecificOutput: { hookEventName: "PreToolUse", additionalContext: text } });

// One input for each tool the extension hands to acta.
const CALLS: [string, Record<string, unknown>][] = [
  ["read", { path: "a.go" }],
  ["edit", { path: "a.go" }],
  ["write", { path: "a.go" }],
  ["bash", { command: "ls" }],
];

describe("payload", () => {
  test("has the shape acta hook reads", () => {
    expect(JSON.parse(payload("S1", "ls"))).toEqual({ session_id: "S1", tool_input: { command: "ls" } });
    expect(JSON.parse(payload("S1"))).toEqual({ session_id: "S1", tool_input: { command: "" } });
  });

  test("carries the tool name, the file path and the reason a session starts, only when given", () => {
    expect(JSON.parse(payload("S1", "", { toolName: "read", filePath: "a.go" }))).toEqual({
      session_id: "S1",
      tool_name: "read",
      tool_input: { command: "", file_path: "a.go" },
    });
    expect(JSON.parse(payload("S1", "", { source: "compact" }))).toEqual({
      session_id: "S1",
      source: "compact",
      tool_input: { command: "" },
    });
  });
});

describe("createState", () => {
  test("first message gets rules and reminder, later ones only the reminder", () => {
    const { run } = fakeRun({
      "acta hook session-start": { stdout: "RULES\n", code: 0 },
      "acta hook prompt": { stdout: "REMINDER\n", code: 0 },
    });
    const s = createState(run, () => "DEFAULTS", noStyle);
    expect(s.contextFor("S1")).toBe("RULES\n\nREMINDER");
    expect(s.contextFor("S1")).toBe("REMINDER");
    s.reset("S1", "compact");
    expect(s.contextFor("S1")).toBe("RULES\n\nREMINDER");
  });

  test("session-start passes the known-plugins file", () => {
    const { run, calls } = fakeRun({ "acta": { stdout: "X", code: 0 } });
    createState(run, () => "", noStyle).contextFor("S1");
    expect(calls[0].key).toMatch(/^acta hook session-start --known .*hooks\/workflow-plugins\.txt$/);
  });

  test("prompt gets the session id on stdin, in the given cwd", () => {
    const { run, calls } = fakeRun({ "acta": { stdout: "X", code: 0 } });
    createState(run, () => "", noStyle).contextFor("S7", "/work");
    const prompt = calls.find((c) => c.key === "acta hook prompt")!;
    expect(JSON.parse(prompt.stdin).session_id).toBe("S7");
    expect(prompt.cwd).toBe("/work");
  });

  test("missing acta falls back to the default rules and no reminder", () => {
    const { run } = fakeRun({});
    const out = createState(run, () => "DEFAULTS\n", noStyle).contextFor("S1");
    expect(out).toContain("DEFAULTS");
    expect(out).toContain("acta binary is not installed or failed");
  });

  test("failing or throwing acta counts as missing", () => {
    for (const out of [{ stdout: "partial", code: 1 }, "throw"] as const) {
      const { run } = fakeRun({ "acta": out });
      const text = createState(run, () => "DEFAULTS", noStyle).contextFor("S1");
      expect(text.startsWith("DEFAULTS")).toBe(true);
      expect(text).not.toContain("partial");
    }
  });

  test("acta ok: the style body follows the rules, with its settings block cut off", () => {
    const { run } = fakeRun({
      "acta hook session-start": { stdout: "RULES\n", code: 0 },
      "acta hook prompt": { stdout: "REMINDER\n", code: 0 },
    });
    const s = createState(run, () => "DEFAULTS", () => STYLE_FILE);
    expect(s.contextFor("S1")).toBe("RULES\n\n" + STYLE_BODY + "\n\nREMINDER");
  });

  // No third argument, so the real reader runs on the committed style file. Every
  // other test stubs the reader, so a wrong path in index.ts would pass them all.
  test("the default reader finds the committed style file", () => {
    const { run } = fakeRun({ "acta hook session-start": { stdout: "RULES\n", code: 0 } });
    const out = createState(run, () => "DEFAULTS").contextFor("S1");
    expect(out).toContain("## Every reply");
    expect(out).not.toContain("force-for-plugin");
    expect(out).not.toContain("name: acta");
  });

  test("acta failed: the style body follows the default rules and their note", () => {
    const { run } = fakeRun({});
    const out = createState(run, () => "DEFAULTS\n", () => STYLE_FILE).contextFor("S1");
    expect(out.startsWith("DEFAULTS")).toBe(true);
    expect(out).toContain("acta binary is not installed or failed");
    expect(out.endsWith("\n\n" + STYLE_BODY)).toBe(true);
    expect(out).not.toContain("keep-coding-instructions");
  });

  test("later messages carry the reminder only, never the style", () => {
    const { run } = fakeRun({
      "acta hook session-start": { stdout: "RULES\n", code: 0 },
      "acta hook prompt": { stdout: "REMINDER\n", code: 0 },
    });
    const s = createState(run, () => "DEFAULTS", () => STYLE_FILE);
    expect(s.contextFor("S1")).toContain(STYLE_BODY);
    expect(s.contextFor("S1")).toBe("REMINDER");
    s.reset("S1", "compact");
    expect(s.contextFor("S1")).toBe("RULES\n\n" + STYLE_BODY + "\n\nREMINDER");
  });

  test("reader throws, acta ok: the rules are as they were and the session goes on", () => {
    const style = brokenStyle();
    const { run } = fakeRun({
      "acta hook session-start": { stdout: "RULES\n", code: 0 },
      "acta hook prompt": { stdout: "REMINDER\n", code: 0 },
    });
    const s = createState(run, () => "DEFAULTS", style);
    expect(s.contextFor("S1")).toBe("RULES\n\nREMINDER");
    expect(style.reads).toBeGreaterThan(0);
    expect(s.contextFor("S1")).toBe("REMINDER");
  });

  test("reader throws, acta failed: the default rules are as they were", () => {
    const style = brokenStyle();
    const before = createState(fakeRun({}).run, () => "DEFAULTS", noStyle).contextFor("S1");
    const out = createState(fakeRun({}).run, () => "DEFAULTS", style).contextFor("S1");
    expect(style.reads).toBeGreaterThan(0);
    expect(out).toBe(before);
  });

  test.each([
    ["no frontmatter: the whole file is used", "BODY\n", "RULES\n\nBODY"],
    ["frontmatter that never closes: the whole file is used", "---\nname: acta\nBODY\n", "RULES\n\n---\nname: acta\nBODY"],
    ["CRLF line ends: the frontmatter is still cut", "---\r\nname: acta\r\n---\r\nBODY\r\n", "RULES\n\nBODY"],
    ["frontmatter and nothing below it: nothing is added", "---\nname: acta\n---\n\n", "RULES"],
    ["empty file: nothing is added", "", "RULES"],
  ])("style file shape, %s", (_name, file, want) => {
    const { run } = fakeRun({ "acta hook session-start": { stdout: "RULES\n", code: 0 } });
    expect(createState(run, () => "DEFAULTS", () => file).contextFor("S1")).toBe(want);
  });

  test("onToolCall blocks only on exit 2, with stderr as the reason", () => {
    const blocked = fakeRun({ "acta hook pre-tool": { stderr: "acta: this session already brainstormed X.\n", code: 2 } });
    expect(
      createState(blocked.run, undefined, noStyle).onToolCall("S1", "bash", { command: "acta set scratch/b status brainstorming" }),
    ).toEqual({
      block: true,
      reason: "acta: this session already brainstormed X.",
    });
    expect(JSON.parse(blocked.calls[0].stdin)).toEqual({
      session_id: "S1",
      tool_name: "bash",
      tool_input: { command: "acta set scratch/b status brainstorming" },
    });
    for (const out of [{ code: 0 }, { code: 1, stderr: "oops" }, { code: 127 }, "throw"] as const) {
      const { run } = fakeRun({ "acta hook pre-tool": out });
      expect(createState(run, undefined, noStyle).onToolCall("S1", "bash", { command: "ls" })).toBeUndefined();
    }
  });

  // This run answers by the file in the payload, so each path can get its own answer.
  function perFile(answer: (file: string) => { stdout: string; stderr: string; code: number }) {
    const asked: string[] = [];
    const run: Run = (_args, stdin) => {
      const file = JSON.parse(stdin).tool_input.file_path;
      asked.push(file);
      return answer(file);
    };
    return { run, asked };
  }

  test("onToolCall asks once for each file, in order, and joins the hints", () => {
    const { run, asked } = perFile((file) => ({ stdout: hintJson("page for " + file), stderr: "", code: 0 }));
    const out = createState(run, undefined, noStyle).onToolCall("S1", "edit", { paths: ["a.go", "b.go"] });
    expect(asked).toEqual(["a.go", "b.go"]);
    expect(out).toEqual({ additionalContext: "page for a.go\npage for b.go" });
  });

  test("onToolCall stops at the first block, and the block wins over hints found before it", () => {
    const { run, asked } = perFile((file) =>
      file === "stop.go"
        ? { stdout: "", stderr: "NO\n", code: 2 }
        : { stdout: hintJson("page for " + file), stderr: "", code: 0 },
    );
    const out = createState(run, undefined, noStyle).onToolCall("S1", "edit", {
      paths: ["a.go", "stop.go", "c.go"],
    });
    expect(out).toEqual({ block: true, reason: "NO" });
    expect(asked).toEqual(["a.go", "stop.go"]);
  });

  test("reset tells acta the session id and why at once, in the given cwd, and the next message reuses the answer", () => {
    const { run, calls } = fakeRun({ "acta hook session-start": { stdout: "RULES", code: 0 } });
    const s = createState(run, () => "", noStyle);
    s.reset("S1", "compact", "/work");
    expect(calls.map((c) => c.key.split(" --known")[0])).toEqual(["acta hook session-start"]);
    expect(JSON.parse(calls[0].stdin)).toEqual({ session_id: "S1", source: "compact", tool_input: { command: "" } });
    expect(calls[0].cwd).toBe("/work");
    expect(s.contextFor("S1", "/work")).toBe("RULES");
    expect(calls.filter((c) => c.key.startsWith("acta hook session-start")).length).toBe(1);
  });

  test("a first message with no reset asks acta with source startup, and only once", () => {
    const { run, calls } = fakeRun({ "acta hook session-start": { stdout: "RULES", code: 0 } });
    const s = createState(run, () => "", noStyle);
    s.contextFor("S1");
    s.contextFor("S1");
    const starts = calls.filter((c) => c.key.startsWith("acta hook session-start")).map((c) => JSON.parse(c.stdin));
    expect(starts).toEqual([{ session_id: "S1", source: "startup", tool_input: { command: "" } }]);
  });

  test("a reset that acta fails to answer still lets the next message carry the default rules", () => {
    const s = createState(fakeRun({}).run, () => "DEFAULTS", noStyle);
    s.reset("S1", "compact");
    expect(s.contextFor("S1")).toContain("DEFAULTS");
  });

  test("onToolResult hands the command to post-tool and never throws", () => {
    const ok = fakeRun({ "acta hook post-tool": { code: 0 } });
    createState(ok.run, undefined, noStyle).onToolResult("S1", "acta set scratch/a status brainstorming", "/work");
    expect(ok.calls[0].key).toBe("acta hook post-tool");
    expect(JSON.parse(ok.calls[0].stdin).tool_input.command).toBe("acta set scratch/a status brainstorming");
    expect(ok.calls[0].cwd).toBe("/work");
    const bad = fakeRun({ "acta hook post-tool": "throw" });
    expect(() => createState(bad.run, undefined, noStyle).onToolResult("S1", "ls")).not.toThrow();
  });
});

describe("extension", () => {
  function load(outputs: Record<string, Out | "throw">, readStyle: () => string = noStyle) {
    const handlers: Record<string, Function> = {};
    const pi = { on: (event: string, fn: Function) => { handlers[event] = fn; } };
    const fake = fakeRun(outputs);
    acta(pi, fake.run, readStyle);
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

  test("the extension hands its style reader on, so the style rides on the first message", async () => {
    const { handlers, ctx } = load(
      { "acta hook session-start": { stdout: "RULES", code: 0 }, "acta hook prompt": { stdout: "REMINDER", code: 0 } },
      () => STYLE_FILE,
    );
    const r = await handlers["before_agent_start"]({ type: "before_agent_start", prompt: "hi" }, ctx);
    expect(r.message.content).toBe("RULES\n\n" + STYLE_BODY + "\n\nREMINDER");
  });

  test("tool_call blocks a bash call when pre-tool exits 2", async () => {
    const { handlers, ctx, calls } = load({ "acta hook pre-tool": { stderr: "NO", code: 2 } });
    const r = await handlers["tool_call"]({ type: "tool_call", toolName: "bash", input: { command: "acta set scratch/b status brainstorming" } }, ctx);
    expect(r).toEqual({ block: true, reason: "NO" });
    expect(calls[0].cwd).toBe("/work");
    expect(JSON.parse(calls[0].stdin).session_id).toBe("S9");
  });

  test.each(["read", "edit", "write"])("tool_call: a %s call hands acta its path and returns the hint as additionalContext", async (tool) => {
    const { handlers, ctx, calls } = load({ "acta hook pre-tool": { stdout: hintJson(HINT) + "\n", code: 0 } });
    const r = await handlers["tool_call"]({ type: "tool_call", toolName: tool, input: { path: "internal/tui/wrap.go" } }, ctx);
    expect(r).toEqual({ additionalContext: HINT });
    expect(calls.map((c) => c.key)).toEqual(["acta hook pre-tool"]);
    expect(calls[0].cwd).toBe("/work");
    expect(JSON.parse(calls[0].stdin)).toEqual({
      session_id: "S9",
      tool_name: tool,
      tool_input: { command: "", file_path: "internal/tui/wrap.go" },
    });
  });

  test("tool_call: a bash call gets the hint too, and acta is handed the command", async () => {
    const { handlers, ctx, calls } = load({ "acta hook pre-tool": { stdout: hintJson(HINT) + "\n", code: 0 } });
    const r = await handlers["tool_call"]({ type: "tool_call", toolName: "bash", input: { command: "go vet ./internal/tui" } }, ctx);
    expect(r).toEqual({ additionalContext: HINT });
    expect(JSON.parse(calls[0].stdin)).toEqual({
      session_id: "S9",
      tool_name: "bash",
      tool_input: { command: "go vet ./internal/tui" },
    });
  });

  test.each([
    ["paths", { paths: ["a.go", "dir/b.go"] }, ["a.go", "dir/b.go"]],
    ["the same file in path and in paths", { path: "a.go", paths: ["a.go"] }, ["a.go"]],
    ["blank and non-string entries", { path: "", paths: ["", 7, null, "c.go"] }, ["c.go"]],
    ["no path at all", {}, []],
  ])("tool_call: an edit with %s asks acta for these files: %j", async (_name, input, files) => {
    const { handlers, ctx, calls } = load({ "acta hook pre-tool": { stdout: hintJson(HINT), code: 0 } });
    const r = await handlers["tool_call"]({ type: "tool_call", toolName: "edit", input }, ctx);
    expect(calls.map((c) => JSON.parse(c.stdin).tool_input.file_path)).toEqual(files);
    expect(r).toEqual(files.length ? { additionalContext: files.map(() => HINT).join("\n") } : undefined);
  });

  // omp reads a.go:10-20 as the file a.go, lines 10 to 20. acta only knows the file.
  test.each([
    ["a line range", "a.go:10-20", "a.go"],
    ["a tail count", "dir/a.go:-30", "dir/a.go"],
    ["a single line with an L", "a.go:L10", "a.go"],
    ["raw", "a.go:raw", "a.go"],
    ["a range and raw", "a.go:10-20:raw", "a.go"],
    ["conflicts, in capitals", "a.go:CONFLICTS", "a.go"],
    ["no selector", "dir/a.go", "dir/a.go"],
    ["a colon that is not a selector", "dir/a:b.go", "dir/a:b.go"],
    ["a name that ends in a colon and a letter L", "dir/a:L", "dir/a:L"],
  ])("tool_call: a read path with %s asks acta about %j", async (_name, path, file) => {
    const { handlers, ctx, calls } = load({ "acta hook pre-tool": { code: 0 } });
    await handlers["tool_call"]({ type: "tool_call", toolName: "read", input: { path } }, ctx);
    expect(calls.map((c) => JSON.parse(c.stdin).tool_input.file_path)).toEqual([file]);
  });

  test("tool_call: edit and write paths are sent whole, since only read takes a selector", async () => {
    for (const toolName of ["edit", "write"]) {
      const { handlers, ctx, calls } = load({ "acta hook pre-tool": { code: 0 } });
      await handlers["tool_call"]({ type: "tool_call", toolName, input: { path: "a.go:10-20" } }, ctx);
      expect(calls.map((c) => JSON.parse(c.stdin).tool_input.file_path)).toEqual(["a.go:10-20"]);
    }
  });

  test.each(CALLS)("tool_call: a %s call that acta blocks stays blocked, with its reason", async (tool, input) => {
    const { handlers, ctx } = load({ "acta hook pre-tool": { stderr: "NO\n", code: 2 } });
    expect(await handlers["tool_call"]({ type: "tool_call", toolName: tool, input }, ctx)).toEqual({ block: true, reason: "NO" });
  });

  test.each([
    ["empty", ""],
    ["text with no JSON around it", "wiki: a hint that is not wrapped"],
    ["JSON null", "null"],
    ["a JSON list", "[]"],
    ["an object with no hint", "{}"],
    ["a hook object with no text", '{"hookSpecificOutput":{"hookEventName":"PreToolUse"}}'],
    ["text that is not a string", '{"hookSpecificOutput":{"additionalContext":42}}'],
    ["blank text", hintJson("  \n")],
  ])("tool_call: acta output that is %s adds nothing, for every tool", async (_name, stdout) => {
    for (const [tool, input] of CALLS) {
      const { handlers, ctx } = load({ "acta hook pre-tool": { stdout, code: 0 } });
      expect(await handlers["tool_call"]({ type: "tool_call", toolName: tool, input }, ctx)).toBeUndefined();
    }
  });

  test("tool_call: a hint counts only on exit 0, and a failing or throwing acta adds nothing", async () => {
    for (const out of [{ stdout: hintJson(HINT), code: 1 }, { stdout: hintJson(HINT), code: 127 }, "throw"] as const) {
      for (const [tool, input] of CALLS) {
        const { handlers, ctx } = load({ "acta hook pre-tool": out });
        expect(await handlers["tool_call"]({ type: "tool_call", toolName: tool, input }, ctx)).toBeUndefined();
      }
    }
  });

  test("tools acta does not watch never call acta", async () => {
    const { handlers, ctx, calls } = load({ "acta": { stderr: "NO", code: 2 } });
    for (const toolName of ["grep", "glob"]) {
      expect(await handlers["tool_call"]({ type: "tool_call", toolName, input: { pattern: "x" } }, ctx)).toBeUndefined();
    }
    await handlers["tool_result"]({ type: "tool_result", toolName: "edit", input: {}, isError: false }, ctx);
    expect(calls.length).toBe(0);
  });

  test("session_start and session_compact tell acta the session id and why at once, so a compaction resets the shown pages", async () => {
    const { handlers, ctx, calls } = load({
      "acta hook session-start": { stdout: "RULES", code: 0 },
      "acta hook prompt": { stdout: "REMINDER", code: 0 },
    });
    const prompt = () => handlers["before_agent_start"]({ type: "before_agent_start", prompt: "hi" }, ctx);
    const starts = () =>
      calls.filter((c) => c.key.startsWith("acta hook session-start")).map((c) => JSON.parse(c.stdin));
    await handlers["session_start"]({ type: "session_start" }, ctx);
    expect(starts().map((e) => [e.session_id, e.source])).toEqual([["S9", "startup"]]);
    const first = await prompt();
    expect(first.message.content).toBe("RULES\n\nREMINDER");
    // The compaction reaches acta before any new message, since omp can go on without one.
    await handlers["session_compact"]({ type: "session_compact" }, ctx);
    expect(starts().map((e) => [e.session_id, e.source])).toEqual([
      ["S9", "startup"],
      ["S9", "compact"],
    ]);
    expect(calls.filter((c) => c.key.startsWith("acta hook session-start")).every((c) => c.cwd === "/work")).toBe(true);
    const after = await prompt();
    expect(after.message.content).toBe("RULES\n\nREMINDER");
    expect(starts().length).toBe(2);
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
    for (const [tool, input] of CALLS) {
      expect(await handlers["tool_call"]({ type: "tool_call", toolName: tool, input }, {})).toBeUndefined();
    }
    expect(calls.length).toBe(0);
  });
});