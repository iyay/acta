---
created: "2026-09-30"
parent: specs/2026-09-30-setup-subagent-models-default-design
id: PLN-0053
hash: kpuh4vp
started: "2026-09-30"
finished: "2026-09-30"
---
# Setup Remembers a No to Split Subagent Models Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Move the user setting code from `internal/voice` into `internal/config`, then let `subagent_models` hold `default` so a no in `acta:setup` is saved and the question stops coming back.

**Architecture:** Task 1 is a pure move: `internal/voice/voice.go` becomes `internal/config/user.go` with the renamed names from the spec table, and every importer switches over. Task 2 adds `default` to the `subagent_models` check, the `config set` help, and the setup skill line.

**Tech Stack:** Go, the repo's own `internal/plugincheck` skill tests.

**Spec:** `.acta/specs/2026-09-30-setup-subagent-models-default-design.md`

**Tests:** fast `scripts/test <touched packages>`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Task 1 changes no behaviour: same file paths (`~/.acta/config.yaml`, the old `~/.acta/voice.yaml` and `~/.pm/voice.yaml`), same `PM_VOICE_FILE` variable, same YAML keys, same error text.
- Rename table, exact: `voice.Voice` -> `config.User`, `voice.Default` -> `config.UserDefault`, `voice.Path` -> `config.UserPath`, `voice.Resolve` -> `config.ResolveUser`, `voice.SaveResolved` -> `config.SaveUser`, `voice.Load` -> `config.LoadUser`, `voice.Save` -> `config.SaveUserFile`, `voice.ErrBad` -> `config.ErrBadUser`. `Validate`, `voicePath`, `oldPath` and `fill` keep their names.
- Field names on other types stay (for example `hook.Input.Voice`, `VoiceExists`, `VoiceErr`). Only the package and the names in the table change.
- `subagent_models` accepts exactly `""`, `split` and `default`. `default` acts like empty for every skill.
- The other skills (brainstorm, build, plan, review, debug, dispatch) are not edited.
- `plugin/skills/setup/SKILL.md` stays at or under 66 lines (plugincheck cap).
- Comments are plain English a 10-year-old reads back without stopping. They say why, not what.

## File map

| File | Task | Change |
|---|---|---|
| `internal/voice/voice.go` | 1 | removed (moved) |
| `internal/voice/voice_test.go`, `internal/voice/voice_acta_test.go` | 1 | removed (moved) |
| `internal/config/user.go` | 1, 2 | new: the moved code; 2 adds `default` |
| `internal/config/user_test.go`, `internal/config/user_acta_test.go` | 1, 2 | new: the moved tests; 2 adds cases |
| `internal/cli/cli.go`, `internal/cli/doctor.go`, `internal/cli/hook.go` | 1 | import switch |
| `internal/cli/config_cmd.go` | 1, 2 | import switch; 2 changes flag help and usage |
| `internal/cli/cli_test.go` | 2 | new test for `default` |
| `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go` | 1 | import switch |
| `internal/hook/hook.go`, `internal/hook/hook_test.go` | 1 | import switch |
| `internal/plugincheck/no_old_names_test.go` | 1 | import switch |
| `internal/plugincheck/skill_setup_test.go` | 2 | require the `default` line |
| `plugin/skills/setup/SKILL.md` | 2 | a no saves `default` |
| `CLAUDE.md` | 1 | Architecture line |

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2 (edits `internal/config/user.go` and `internal/cli/config_cmd.go` after Task 1 moved them).

---

### Task 1: Move the user setting into internal/config

**Files:**
- Create: `internal/config/user.go`, `internal/config/user_test.go`, `internal/config/user_acta_test.go` (with `git mv` from `internal/voice/`)
- Delete: `internal/voice/` (empty after the moves)
- Modify: `internal/cli/cli.go`, `internal/cli/config_cmd.go`, `internal/cli/doctor.go`, `internal/cli/hook.go`, `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go`, `internal/hook/hook.go`, `internal/hook/hook_test.go`, `internal/plugincheck/no_old_names_test.go`, `CLAUDE.md`

**verify:** No behaviour changes on any path, and no code refers to the old package. List: every exported name in the rename table and where each caller now points; that `grep -rn 'internal/voice' --include='*.go' .` prints nothing; that every moved test still runs under its old test name and passes; that paths, env var, YAML keys and error strings are byte-for-byte the same as before.

**Interfaces:**
- Consumes: nothing.
- Produces, in package `config` (`github.com/iyay/acta/internal/config`):
  - `type User struct` (same fields and yaml tags as `voice.Voice`, including `SubagentModels string \`yaml:"subagent_models,omitempty"\``)
  - `var ErrBadUser error`
  - `func UserDefault() User`
  - `func UserPath() (string, error)`
  - `func ResolveUser() (User, bool, error)`
  - `func SaveUser(v User) error`
  - `func LoadUser(path string) (User, bool, error)`
  - `func SaveUserFile(path string, v User) error`
  - `func (v User) Validate() error`
  - unexported `fill(v User) User`, `voicePath()`, `oldPath()`

- [x] **Step 1: Move the tests first (red)**

```bash
git mv internal/voice/voice_test.go internal/config/user_test.go
git mv internal/voice/voice_acta_test.go internal/config/user_acta_test.go
```

In both moved test files: `package voice` -> `package config`, and apply the rename table to every use (`Voice` -> `User`, `Default()` -> `UserDefault()`, `Path()` -> `UserPath()`, `Resolve()` -> `ResolveUser()`, `SaveResolved(` -> `SaveUser(`, `Load(` -> `LoadUser(`, `Save(` -> `SaveUserFile(`, `ErrBad` -> `ErrBadUser`). Test function names stay as they are. Check first that no test name in `internal/config/config_test.go` or `internal/config/config_acta_test.go` clashes with a moved test name (for example both files may have `TestDefault`, `TestPath`, `TestLoadMissing`). When a name clashes, prefix the moved one with `User`, for example `TestDefault` -> `TestUserDefault`.

- [x] **Step 2: Run to watch it fail**

Run: `scripts/test ./internal/config/`
Expected: FAIL to build with `undefined: User` (and the other new names).

- [x] **Step 3: Move the code**

```bash
git mv internal/voice/voice.go internal/config/user.go
```

In `internal/config/user.go`: change `package voice` to `package config`. `config.go` already holds the package doc comment, so turn the old `// Package voice ...` lines into a plain comment above the `User` type:

```go
// User holds how the agent should talk to the user: the chat language,
// the style, an optional tone, and the language for repo files.
```

Keep the existing comment that says where the file lives. Apply the rename table to the declarations and to every use inside the file. `ErrBadUser` keeps the text `"bad voice setting"`.

- [x] **Step 4: Switch every importer**

In each of `internal/cli/cli.go`, `internal/cli/config_cmd.go`, `internal/cli/doctor.go`, `internal/cli/hook.go`, `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go`, `internal/hook/hook.go`, `internal/hook/hook_test.go`, `internal/plugincheck/no_old_names_test.go`:
- Drop the `"github.com/iyay/acta/internal/voice"` import. Add `"github.com/iyay/acta/internal/config"` where it is not there yet (`cli.go`, `doctor.go` and `hook.go` in `internal/cli` already import it).
- Replace `voice.Voice` -> `config.User`, `voice.Default` -> `config.UserDefault`, `voice.Path` -> `config.UserPath`, `voice.Resolve` -> `config.ResolveUser`, `voice.SaveResolved` -> `config.SaveUser`, `voice.ErrBad` -> `config.ErrBadUser`.
- A local variable or parameter named `config` in these files now shadows the package. Where the compiler complains, rename the local (for example `cfg`), only in the lines that need it.

In `CLAUDE.md` Architecture, replace the line

```
- `internal/voice`: chat language, style, tone, repo language and build executor, stored in `~/.acta/config.yaml` (or `PM_VOICE_FILE`).
```

with

```
- `internal/config/user.go`: the user setting (chat language, style, tone, repo language, build executor, subagent models), stored in `~/.acta/config.yaml` (or `PM_VOICE_FILE`).
```

and in the `internal/board` line, change "`internal/config` finds where the planning files live." to "`internal/config` finds where the planning files live, and also holds the user setting."

- [x] **Step 5: Run to watch it pass**

Run: `scripts/test ./internal/config/ ./internal/cli/ ./internal/doctor/ ./internal/hook/ ./internal/plugincheck/`
Expected: PASS.

Run: `go vet ./... && gofmt -l . && grep -rn 'internal/voice' --include='*.go' .`
Expected: vet clean, no gofmt output, no grep output.

- [x] **Step 6: Commit**

```bash
git add -A internal/ CLAUDE.md
git commit -m "Move the user setting from internal/voice into internal/config (SPC-0045 part 1)"
```

---

### Task 2: subagent_models accepts default; setup saves it on a no

**Files:**
- Modify: `internal/config/user.go` (the `subagent_models` check in `Validate`)
- Modify: `internal/config/user_test.go` (`TestValidateExecutorAndModels`)
- Modify: `internal/cli/config_cmd.go` (`configUsage`, the `subagent-models` flag help)
- Modify: `internal/cli/cli_test.go` (new test)
- Modify: `plugin/skills/setup/SKILL.md` (the Split subagent models section)
- Modify: `internal/plugincheck/skill_setup_test.go` (required strings)

**verify:** A no in setup is saved on every path and never read as "not set", and no other value slips in. List: each accepted value (`""`, `split`, `default`) and each rejected one tried (`all`, `mixed`, `Default`); `config set --subagent-models default` then `config show` in text and `--json`; `--clear-subagent-models` still removes `default`; the setup skill text for the yes and the no; that no other skill file changed.

**Interfaces:**
- Consumes: `config.User`, `config.UserDefault`, `config.ErrBadUser`, `fill`, `(User).Validate` from Task 1.
- Produces: nothing new; `subagent_models: default` becomes a valid stored value.

- [x] **Step 1: Write the failing tests**

In `internal/config/user_test.go`, `TestValidateExecutorAndModels`, add these rows to the table:

```go
		{"", "default", true},
		{"", "mixed", false},
		{"", "Default", false},
```

In `internal/cli/cli_test.go`, next to the existing `config set --subagent-models split` test, add (use the same setup that test uses for `PM_VOICE_FILE` and `mustRun`):

```go
// A no in setup is saved as default, so the next setup run sees it as set.
func TestConfigSetSubagentModelsDefault(t *testing.T) {
	t.Setenv("PM_VOICE_FILE", filepath.Join(t.TempDir(), "config.yaml"))
	mustRun(t, "config", "set", "--subagent-models", "default")
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "subagent_models: default") {
		t.Errorf("show lacks subagent_models: default:\n%s", out)
	}
	if out := mustRun(t, "config", "show", "--json"); !strings.Contains(out, `"subagent_models":"default"`) {
		t.Errorf("json lacks subagent_models default:\n%s", out)
	}
	mustRun(t, "config", "set", "--clear-subagent-models")
	if out := mustRun(t, "config", "show"); strings.Contains(out, "subagent_models") {
		t.Errorf("clear left subagent_models:\n%s", out)
	}
}
```

If the existing split test sets up the env another way (a helper), copy that setup instead of the `t.Setenv` line.

In `internal/plugincheck/skill_setup_test.go`, next to `"--subagent-models split"` in `Must`, add `"--subagent-models default"`.

- [x] **Step 2: Run to watch them fail**

Run: `scripts/test ./internal/config/ -run TestValidateExecutorAndModels`
Expected: FAIL on `models "default": ... want ok=true`.

Run: `scripts/test ./internal/cli/ -run TestConfigSetSubagentModelsDefault`
Expected: FAIL, the set is refused with the subagent_models error.

Run: `scripts/test ./internal/plugincheck/ -run Setup`
Expected: FAIL, the setup skill lacks `--subagent-models default`.

- [x] **Step 3: Implement**

In `internal/config/user.go`, `Validate`, replace the models check with:

```go
	switch v.SubagentModels {
	case "", "split", "default":
	default:
		return fmt.Errorf("%w: subagent_models must be split or default, not %q", ErrBadUser, v.SubagentModels)
	}
```

In `internal/cli/config_cmd.go`: in `configUsage`, change `[--subagent-models split]` to `[--subagent-models split|default]`; change the flag help to `"how models are picked for subagents: split, or default to leave it to your own config"`.

In `plugin/skills/setup/SKILL.md`, replace the line

```
Ask this one in Claude Code only. Default no. A yes saves `acta config set --subagent-models split`; a no saves nothing and the user's own config wins.
```

with

```
Ask this one in Claude Code only, and only while `acta config show` has no `subagent_models` line. Default no. A yes saves `acta config set --subagent-models split`. A no saves `acta config set --subagent-models default`, so the question is not asked again; the user's own config wins.
```

- [x] **Step 4: Run to watch them pass**

Run: `scripts/test ./internal/config/ ./internal/cli/ ./internal/plugincheck/`
Expected: PASS.

Run: `go vet ./... && gofmt -l .`
Expected: clean.

- [x] **Step 5: Commit**

```bash
git add internal/config/user.go internal/config/user_test.go internal/cli/config_cmd.go internal/cli/cli_test.go plugin/skills/setup/SKILL.md internal/plugincheck/skill_setup_test.go
git commit -m "subagent_models accepts default; setup saves it on a no (SPC-0045 part 2)"
```
