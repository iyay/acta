---
parent: specs/2026-10-07-acta-setup-install-rows
id: SPC-0104
created: "2026-10-07 13:19:23"
hash: n3djm0a
started: "2026-10-07 13:20:04"
finished: "2026-10-07 13:25:36"
---
Status: Bounded, approved by the user in chat on 2026-10-07.
Why: a real run of `acta setup` showed four rough spots. Without `--plugin-dir` the wizard asks Yes/No for each tool, then ignores the Yes and prints the install commands with `<plugin dir>`; the second line of the Claude command falls off the rail; the acta block shows twice; `Saved your answers.` does not use the rail marks. It also asks about a tool where doctor already sees the plugin installed. Separately, after the user deleted `~/.acta`, the wizard still offered Indonesian: it read its defaults through `config.ResolveUser`, which falls back to the old `~/.pm/voice.yaml`.

# acta setup: honest install step and empty defaults

## Changes

1. **Installed tools are not asked.** A tool where the plugin is already installed shows one line `✓ <tool>  already installed` and gets no Yes/No row. Use the same check doctor uses for "acta plugin enabled in Claude Code"; for omp, the plugin link doctor checks.
2. **No plugin dir, no question.** Without `--plugin-dir`, tools that still need the plugin get no Yes/No row. Each shows `▲ <tool>  run by hand:` and then each command on its own rail line (`│    claude plugin marketplace add <plugin dir>`). No line of a command ever leaves the rail.
3. **One block line.** The block shows once, after it is written: `✓ acta block → <path>` per file. The earlier preview line goes.
4. **Rail marks for the config line.** `Saved your answers.` becomes `✓ config saved`.
5. **Empty defaults on a first run.** The wizard takes its defaults only from `~/.acta/config.yaml` (or `PM_VOICE_FILE`), never from the old `~/.acta/voice.yaml` or `~/.pm/voice.yaml`. With no config file, text questions start empty with a dim placeholder (for example `English`) and selects start on their first option. Chat language and repo language must be typed: an empty answer shows `▲ required` and stays on the question. Tone stays optional. The legacy fallback in `config.ResolveUser` stays as it is for every other caller.

6. **Plugin files ship inside the binary (user ruling 2026-10-07).** The `plugin/` folder is embedded in the `acta` binary with `go:embed` (the `plugin` Go package already lives there). `acta setup` extracts it to `~/.acta/plugin/`, replacing what is there, and uses that folder as the plugin dir for every install command (`claude plugin marketplace add ~/.acta/plugin`, `claude plugin install acta@acta-local`, `omp plugin link ~/.acta/plugin`). So the wizard works from any folder, with no repo clone and no network, and the plugin version always matches the binary. `--plugin-dir` stays as an override for developers who want their own checkout. Evals and Go test files are not embedded. Item 2's run-by-hand path then shows only when an install command fails or extraction fails.

Nothing else changes: values written, order, layout, colors, `Plan`'s block rules, `WriteBlock`, the no-TTY guard.

## Testing

- Install step: tool already installed, not installed with and without `--plugin-dir`, mixed; no Yes/No row for an installed tool or without a plugin dir; every command line on the rail.
- Output: block line once, `✓ config saved`.
- Defaults: no config file plus an old `~/.pm/voice.yaml` and `~/.acta/voice.yaml` present gives empty inputs; a config file gives its values; empty chat or repo language is refused.
