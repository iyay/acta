// acta for omp. omp does not run Claude Code's shell hooks, so this extension
// sends the same text through omp's own events: the session rules on the first
// message and after compaction, the voice reminder on every message, and the
// one-brainstorm-per-session check around every bash call. omp has no output
// styles, so the acta style rides along with the session rules.
import { spawnSync } from "node:child_process";
import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const pluginRoot = join(dirname(fileURLToPath(import.meta.url)), "..");
const knownFile = join(pluginRoot, "hooks", "workflow-plugins.txt");
const defaultsFile = join(pluginRoot, "hooks", "default-rules.md");
const styleFile = join(pluginRoot, "output-styles", "acta.md");

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

// styleBody cuts the settings block off the top of the style file. Claude Code
// reads that block, and omp only needs the words below it. A block that never
// closes is not settings, so the text stays whole.
function styleBody(text: string): string {
  const lines = text.split("\n");
  if (lines[0].trimEnd() !== "---") return text.trim();
  const close = lines.findIndex((line, i) => i > 0 && line.trimEnd() === "---");
  return (close < 0 ? text : lines.slice(close + 1).join("\n")).trim();
}

// readStyleBody never throws. A style file that is missing or broken counts as
// empty, so it cannot stop the session or change the rules.
function readStyleBody(readStyle: () => string): string {
  try {
    return styleBody(readStyle());
  } catch {
    return "";
  }
}

export function createState(
  run: Run,
  readDefaults: () => string = () => readFileSync(defaultsFile, "utf8"),
  readStyle: () => string = () => readFileSync(styleFile, "utf8"),
) {
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
        // The style goes once, with the rules. The reminder below never carries it.
        const style = readStyleBody(readStyle);
        if (style) parts.push(style);
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

export default function acta(pi: any, run: Run = realRun, readStyle?: () => string) {
  const state = createState(run, undefined, readStyle);
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
