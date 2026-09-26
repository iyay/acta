import { describe, expect, test } from "bun:test";
import acta, { createState } from "./index";

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
      "acta hook session-start": { stdout: "RULES\n", code: 0 },
      "acta hook prompt": { stdout: "REMINDER\n", code: 0 },
    });
    const s = createState(exec, () => "DEFAULTS");
    expect(await s.contextFor()).toBe("RULES\n\nREMINDER");
    expect(await s.contextFor()).toBe("REMINDER");
    s.reset();
    expect(await s.contextFor()).toBe("RULES\n\nREMINDER");
  });

  test("session-start passes the known-plugins file", async () => {
    const { exec, calls } = fakeExec({ "acta": { stdout: "X", code: 0 } });
    await createState(exec, () => "").contextFor();
    expect(calls[0]).toMatch(/^acta hook session-start --known .*hooks\/workflow-plugins\.txt$/);
  });

  test("missing acta falls back to the default rules and no reminder", async () => {
    const { exec } = fakeExec({});
    const out = await createState(exec, () => "DEFAULTS\n").contextFor();
    expect(out).toContain("DEFAULTS");
    expect(out).toContain("acta binary is not installed or failed");
  });

  test("failing acta counts as missing", async () => {
    const { exec } = fakeExec({ "acta": { stdout: "partial", code: 1 } });
    const out = await createState(exec, () => "DEFAULTS").contextFor();
    expect(out.startsWith("DEFAULTS")).toBe(true);
    expect(out).not.toContain("partial");
  });
});

describe("extension", () => {
  test("registers the three events and injects an acta message", async () => {
    const handlers: Record<string, Function> = {};
    const pi = {
      on: (event: string, fn: Function) => { handlers[event] = fn; },
      exec: async (_cmd: string, args: string[]) => ({ stdout: args[1] === "prompt" ? "REMINDER" : "RULES", code: 0 }),
    };
    acta(pi);
    expect(Object.keys(handlers).sort()).toEqual(["before_agent_start", "session_compact", "session_start"]);
    const r = await handlers["before_agent_start"]({ type: "before_agent_start", prompt: "hi" });
    expect(r.message.customType).toBe("acta");
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
