# Codex facts for the acta plugin

Checked on 2026-10-15 with codex-cli 0.162.1, in a throwaway `CODEX_HOME` and a scratch git repo. A logger hook saved each hook's stdin and env, then ran the real acta hook.

| Fact | Result | How |
|---|---|---|
| A. Codex reads our existing manifest | yes | `codex plugin marketplace add <plugin dir>` found `.claude-plugin/marketplace.json` and added `acta-local`; no Codex `plugin.json` needed |
| B. Install | yes | `codex plugin add acta@acta-local` installed into `$CODEX_HOME/plugins/cache/acta-local/acta/0.1.45` |
| C. Skills load | yes, all 13 | `codex exec` listed `acta:bug` ... `acta:tdd` |
| D. Plugin hooks fire | yes, all four | logger files for SessionStart, UserPromptSubmit, PreToolUse (Bash and apply_patch), PostToolUse |
| E. Hook trust | needed | plugin hooks are skipped until trusted in `/hooks`; `--dangerously-bypass-hook-trust` skips the check for one run |
| F. SessionStart stdout becomes context | yes | the model quoted "acta plugin is active" |
| G. Second-brainstorm block | works unchanged | second `acta set scratch/... status brainstorming` gave "Command blocked by PreToolUse hook: acta: this session already brainstormed SCR-0001 ..." |
| H. Signal that the hook runs under Codex | env `PLUGIN_ROOT` | Codex sets `PLUGIN_ROOT` and `PLUGIN_DATA` beside `CLAUDE_PLUGIN_ROOT`; Claude Code sets only the `CLAUDE_` ones |
| I. Edits | `tool_name: "apply_patch"` | `tool_input.command` holds the whole patch; header paths are absolute |

## Not a signal

- `turn_id` is missing from SessionStart input, so it cannot tell the harness at session start.
- `AI_AGENT` and `CLAUDECODE` come from the parent process. Codex started from inside Claude Code inherits them.

## Hook input samples

SessionStart keys: `cwd`, `hook_event_name`, `model`, `permission_mode`, `session_id`, `source`, `transcript_path`.

PreToolUse for an edit:

```json
{
  "hook_event_name": "PreToolUse",
  "tool_name": "apply_patch",
  "tool_input": {
    "command": "*** Begin Patch\n*** Update File: /abs/repo/a.txt\n@@\n-hi\n+hello\n*** End Patch"
  },
  "turn_id": "...",
  "tool_use_id": "..."
}
```

PostToolUse for Bash adds `tool_response` as a plain string: `"spike\n"`.

## Smoke run with the branch binary

Checked on 2026-10-15: a scratch repo with one wiki page covering `src/`, the acta plugin added from a local marketplace, and `codex exec --dangerously-bypass-hook-trust`.

```text
$ acta doctor | grep codex
ok codex: acta plugin enabled in Codex; trust its hooks once in /hooks

codex reply:
(a) `wiki: .acta/wiki/parser.md: SMOKE-HINT parser keeps quotes`   <- hint from an apply_patch edit of src/a.go
(b) `Codex: this session runs in Codex.`                            <- Codex block in the session note
(c) Ya.                                                             <- style text "Open with the answer or the result." is in context
```

The second-brainstorm block is row G above.

## Install from GitHub

`codex plugin marketplace add iyay/acta` fails: "marketplace root does not contain a supported manifest". `--sparse plugin` only narrows the checkout; the root stays the repo root. So the install goes through `acta setup`, which adds the unpacked `~/.acta/plugin` as a local marketplace.
