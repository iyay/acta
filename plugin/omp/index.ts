// acta for omp. omp does not run Claude Code's shell hooks, so this extension
// sends the same text through omp's own events: the session rules on the first
// message and after compaction, the voice reminder on every message, the
// one-brainstorm-per-session check around every bash call, and the wiki hints
// around every read, edit, write and bash call. omp has no output styles, so
// the acta style rides along with the session rules.
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
// serves both harnesses. A field nobody gave is left out of the JSON.
export function payload(
  sessionId: string,
  command = "",
  extra: { toolName?: string; filePath?: string; source?: string } = {},
): string {
  return JSON.stringify({
    session_id: sessionId,
    tool_name: extra.toolName,
    source: extra.source,
    tool_input: { command, file_path: extra.filePath },
  });
}

// The omp tools acta is asked about. A wiki page covers files, so the three file
// tools can earn a hint, and so can bash, since a command names files too.
const WATCHED = new Set(["bash", "read", "edit", "write"]);

// omp lets a read path end in a selector, like a.go:10-20, a.go:-30 or a.go:raw.
// acta knows plain file paths only, so the selector is cut off. The pattern is a
// little looser than omp's own. The worst it can do is miss a page for a file
// with a very odd name.
const SELECTOR_TAIL = /(?::(?:raw|conflicts|img|-?L?\d[\dL.,+-]*))+$/i;

// filePaths lists the files one file-tool call names. omp puts one in path or
// several in paths, and a one-file edit puts the same file in both. A blank or
// a non-string is dropped, since acta cannot match it.
function filePaths(toolName: string, input: any): string[] {
  const named = [input?.path, ...(Array.isArray(input?.paths) ? input.paths : [])]
    .filter((p): p is string => typeof p === "string")
    .map((p) => (toolName === "read" ? p.replace(SELECTOR_TAIL, "") : p))
    .filter((p) => p !== "");
  return [...new Set(named)];
}

// hintOf reads the wiki hint out of what acta printed: one JSON object with the
// text in hookSpecificOutput.additionalContext. Anything else is no hint, so
// odd output can never put words in front of the agent or break the call.
function hintOf(stdout: string): string {
  try {
    const text = JSON.parse(stdout)?.hookSpecificOutput?.additionalContext;
    return typeof text === "string" ? text.trim() : "";
  } catch {
    return "";
  }
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
  // acta is told at once, not on the next message. omp can retry a request right
  // after a compaction with no new message, and the pages acta showed are gone
  // by then. The session rules can only go out with a message, so acta's answer
  // waits here until the next one.
  let startAnswer: ReturnType<typeof call> | undefined;
  const sessionStart = (sessionId: string, source: string, cwd?: string) =>
    call(run, ["hook", "session-start", "--known", knownFile], payload(sessionId, "", { source }), cwd);
  return {
    reset(sessionId: string, source: "startup" | "compact", cwd?: string) {
      rulesSent = false;
      startAnswer = sessionStart(sessionId, source, cwd);
    },
    contextFor(sessionId: string, cwd?: string): string {
      const parts: string[] = [];
      if (!rulesSent) {
        const r = startAnswer ?? sessionStart(sessionId, "startup", cwd);
        startAnswer = undefined;
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
    // A call that names several files asks acta once for each. Only exit 2
    // blocks, and the first block wins over any hint found before it. A hint
    // counts on exit 0 only. Every other answer lets the call run, because a
    // hook that guesses wrong stops real work the user asked for.
    onToolCall(sessionId: string, toolName: string, input: any, cwd?: string) {
      const bodies =
        toolName === "bash"
          ? [payload(sessionId, String(input?.command ?? ""), { toolName })]
          : filePaths(toolName, input).map((filePath) => payload(sessionId, "", { toolName, filePath }));
      const hints: string[] = [];
      for (const body of bodies) {
        const r = call(run, ["hook", "pre-tool"], body, cwd);
        if (r.code === 2) return { block: true as const, reason: r.stderr.trim() };
        const hint = r.code === 0 ? hintOf(r.stdout) : "";
        if (hint) hints.push(hint);
      }
      return hints.length ? { additionalContext: hints.join("\n") } : undefined;
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
  // Both events tell acta the session id and why they fired. That is how a
  // compaction makes acta forget the pages it showed, so the hints come back.
  const restart = (source: "startup" | "compact") => async (_event: any, ctx: any) => {
    const w = where(ctx);
    state.reset(w.id, source, w.cwd);
  };
  pi.on("session_start", restart("startup"));
  pi.on("session_compact", restart("compact"));
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
    if (!WATCHED.has(event?.toolName) || !w.id) return;
    return state.onToolCall(w.id, event.toolName, event.input, w.cwd);
  });
  pi.on("tool_result", async (event: any, ctx: any) => {
    const w = where(ctx);
    // A failed or blocked call did not brainstorm anything, so it is not
    // recorded. Claude Code's post-tool hook also runs only on success.
    if (event?.toolName !== "bash" || event?.isError || !w.id) return;
    state.onToolResult(w.id, String(event.input?.command ?? ""), w.cwd);
  });
}
