---
parent: scratch/2026-09-29-config-file-name
closes: [SCR-0026]
created: "2026-09-30"
id: SPC-0032
hash: wevgkfg
started: "2026-09-30"
finished: "2026-09-30"
---
# Config file name, setup init, plan reads executor

Status: design approved by the user in chat on 2026-09-30. Bounded (changes the existing config read in `internal/voice`, and the `setup` and `plan` skills). The config move is a data migration of a user file, so the review is the full one.

## Why

- `~/.acta/voice.yaml` now holds more than the voice: `build_executor`, `subagent_models` and `theme` live there too. The name no longer fits (SCR-0024).
- `acta:setup` skips the acta block when a repo has no `CLAUDE.md` and no `AGENTS.md`. The user wants setup to make one.
- `acta:plan` asks which executor to run and calls `subagent` the default, even when the config says `build_executor: dispatch` (SCR-0026). `acta:build` already reads the config.

## Design

1. **Config path.** `voice.Path()` returns `~/.acta/config.yaml`. `PM_VOICE_FILE` still wins when set.
2. **One-time move.** In `voice.Resolve()`: when `PM_VOICE_FILE` is not set, `config.yaml` is missing and `~/.acta/voice.yaml` exists, rename `voice.yaml` to `config.yaml` with `os.Rename`, then load `config.yaml`.
   - The move happens on the first read, hooks included, so the user sees one file right away.
   - When the rename fails, load `voice.yaml` as it is, so no setting is lost. The next read tries again.
   - Two hooks racing is safe: the rename is atomic, and the loser finds `config.yaml` in place.
   - When `config.yaml` already exists, it wins and `voice.yaml` is not touched.
   - `~/.pm/voice.yaml` stays a read-only fallback, as today. It is never moved.
3. **CLI names stay.** `acta voice show` and `acta voice set` keep their names. `show` prints the new path.
4. **Setup makes the file.** In `plugin/skills/setup/SKILL.md`, the acta block step changes:
   - Only `AGENTS.md` exists: write the block there. No new `CLAUDE.md`.
   - Neither file exists: in Claude Code, run `/init` first, then add the block to the `CLAUDE.md` it made. In other harnesses, create `CLAUDE.md` holding only the block.
   - The user still sees the exact block and says yes before any write.
5. **Plan reads the executor.** `plugin/skills/plan/SKILL.md` runs `acta voice show` before the build hand-off. When it prints `build_executor: <name>`, name that executor and do not ask. When the line is missing, ask as today. The session index line for `build` in `internal/hook/hook.go` says the executor comes from config, else ask, instead of "executor subagent (default)".
6. **Docs.** Every live text naming `~/.acta/voice.yaml` as the home path moves to `config.yaml`: the setup skill, `plugin/README.md`, and the comments in `internal/voice/voice.go`. Old specs, plans and scratch items stay as they are.

## Testing

- Go tests in `internal/voice`: the move happens and the content is the same; an existing `config.yaml` wins and `voice.yaml` stays; `PM_VOICE_FILE` set means no move; no old file gives defaults and no error; a failed rename still loads `voice.yaml`.
- Existing tests that name `voice.yaml` as the home path move to `config.yaml`.
- Text tests for the hook index line and the plan and setup skill wording, in the style the repo already uses for skill text.

## Out of scope

- Renaming `PM_VOICE_FILE` to an `ACTA_` name.
- An `acta config` alias for `acta voice`.
