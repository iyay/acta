// acta for omp. omp does not run Claude Code's shell hooks, so this extension
// sends the same text through omp's own events: the session rules on the first
// message and after compaction, and the voice reminder on every message.
import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const pluginRoot = join(dirname(fileURLToPath(import.meta.url)), "..");
const knownFile = join(pluginRoot, "hooks", "workflow-plugins.txt");
const defaultsFile = join(pluginRoot, "hooks", "default-rules.md");

type Exec = (cmd: string, args: string[]) => Promise<{ stdout: string; code: number }>;

// runActa returns acta's output, or null when acta is missing or fails.
async function runActa(exec: Exec, args: string[]): Promise<string | null> {
  try {
    const r = await exec("acta", args);
    return r.code === 0 ? r.stdout.trim() : null;
  } catch {
    return null;
  }
}

export function createState(exec: Exec, readDefaults: () => string = () => readFileSync(defaultsFile, "utf8")) {
  let rulesSent = false;
  return {
    reset() {
      rulesSent = false;
    },
    async contextFor(): Promise<string> {
      const parts: string[] = [];
      if (!rulesSent) {
        const rules = await runActa(exec, ["hook", "session-start", "--known", knownFile]);
        parts.push(
          rules ??
            readDefaults().trim() +
              "\n\nacta: the acta binary is not installed or failed, so these are the default rules (English, adhd style).",
        );
        rulesSent = true;
      }
      const reminder = await runActa(exec, ["hook", "prompt"]);
      if (reminder) parts.push(reminder);
      return parts.join("\n\n");
    },
  };
}

export default function acta(pi: any) {
  const exec: Exec = async (cmd, args) => {
    const r = await pi.exec(cmd, args);
    return { stdout: String(r?.stdout ?? ""), code: Number(r?.code ?? 1) };
  };
  const state = createState(exec);
  pi.on("session_start", async () => state.reset());
  pi.on("session_compact", async () => state.reset());
  pi.on("before_agent_start", async () => {
    const content = await state.contextFor();
    if (!content) return;
    // Shape follows BeforeAgentStartEventResult: no attribution field.
    return { message: { customType: "acta", content, display: false } };
  });
}
