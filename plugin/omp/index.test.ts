import { describe, expect, test } from "bun:test";
import pm, { createState } from "./index";

function fakeExec(outputs: Record<string, { stdout: string; code: number }>) {
  const calls: string[] = [];
  const exec = async (cmd: string, args: string[]) => {
    const key = [cmd, ...args].join(" ");
    calls.push(key);
    for (const [prefix, out] of Object.entries(outputs)) {
      if (key.startsWith(prefix)) return out;
    }
    throw new Error("not found");
  };
  return { exec, calls };
}

describe("createState", () => {
  test("first message gets rules and reminder, later ones only the reminder", async () => {
    const { exec } = fakeExec({
      "pmb hook session-start": { stdout: "RULES\n", code: 0 },
      "pmb hook prompt": { stdout: "REMINDER\n", code: 0 },
    });
    const s = createState(exec, () => "DEFAULTS");
    expect(await s.contextFor()).toBe("RULES\n\nREMINDER");
    expect(await s.contextFor()).toBe("REMINDER");
    s.reset();
    expect(await s.contextFor()).toBe("RULES\n\nREMINDER");
  });

  test("session-start passes the known-plugins file", async () => {
    const { exec, calls } = fakeExec({ "pmb": { stdout: "X", code: 0 } });
    await createState(exec, () => "").contextFor();
    expect(calls[0]).toMatch(/^pmb hook session-start --known .*hooks\/workflow-plugins\.txt$/);
  });

  test("missing pmb falls back to the default rules and no reminder", async () => {
    const { exec } = fakeExec({});
    const out = await createState(exec, () => "DEFAULTS\n").contextFor();
    expect(out).toContain("DEFAULTS");
    expect(out).toContain("pmb binary is not installed or failed");
  });

  test("failing pmb counts as missing", async () => {
    const { exec } = fakeExec({ "pmb": { stdout: "partial", code: 1 } });
    const out = await createState(exec, () => "DEFAULTS").contextFor();
    expect(out.startsWith("DEFAULTS")).toBe(true);
    expect(out).not.toContain("partial");
  });
});

describe("extension", () => {
  test("registers the three events and injects a pm message", async () => {
    const handlers: Record<string, Function> = {};
    const pi = {
      on: (event: string, fn: Function) => { handlers[event] = fn; },
      exec: async (_cmd: string, args: string[]) => ({ stdout: args[1] === "prompt" ? "REMINDER" : "RULES", code: 0 }),
    };
    pm(pi);
    expect(Object.keys(handlers).sort()).toEqual(["before_agent_start", "session_compact", "session_start"]);
    const r = await handlers["before_agent_start"]({ type: "before_agent_start", prompt: "hi" });
    expect(r.message.customType).toBe("pm");
    expect(r.message.content).toBe("RULES\n\nREMINDER");
    expect(r.message.display).toBe(false);
    expect("attribution" in r.message).toBe(false);
    const again = await handlers["before_agent_start"]({ type: "before_agent_start", prompt: "hi" });
    expect(again.message.content).toBe("REMINDER");
    await handlers["session_compact"]({});
    const afterCompact = await handlers["before_agent_start"]({ type: "before_agent_start", prompt: "hi" });
    expect(afterCompact.message.content).toBe("RULES\n\nREMINDER");
  });
});
