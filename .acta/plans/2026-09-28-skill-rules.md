---
id: PLN-0018
hash: hsnijdg
---
# Skill Rules and First-Run Setup Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `acta doctor [--fix]`, two new voice fields (`build_executor`, `subagent_models`), an `acta: ` description prefix on every skill, a new `acta:scratch` skill, brainstorm/setup/build rules from the spec, and matching session rules, hook text and README.

**Architecture:** Go work first: the voice fields live in `internal/voice` and `internal/cli/voice.go`; doctor's checks live in a new package `internal/doctor` (pure functions over an `Env` struct, so tests use temp folders) with a thin command in `internal/cli/doctor.go`. Skill text changes are guarded by text tests in `internal/plugincheck` (`SkillRule.Must` / `MustNot`), red first. Session-rule text lives in `internal/hook/hook.go` and its fallback copy `plugin/hooks/default-rules.md`.

**Tech Stack:** Go 1.27, gopkg.in/yaml.v3, Markdown skills.

**Spec:** `.acta/specs/2026-09-28-skill-rules-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Voice values, exact: `build_executor` is `subagent`, `dispatch` or `inline` (or empty); `subagent_models` is `split` (or empty). Any other value fails `Validate` with `ErrBad`.
- Doctor line format, exact: `<level> <name>: <message>`, then `fix: <text>` on its own line when level is not `ok`. Levels exactly `ok`, `warn`, `fail`. Check names exactly `binary`, `harness`, `stale-links`, `conflicts`, `repo`, `agents-view`, `setup`, printed in that order.
- Doctor exit code: 0 when no `fail`, 1 when any `fail`.
- Doctor never writes outside the repo. It never writes `~/.claude.json`, `~/.omp`, or `~/.claude/settings*.json`.
- Every skill `description:` starts with `acta: ` (with the space). Frontmatter `name:` never changes.
- CLAUDE.md block markers, exact: `<!-- acta:begin -->` and `<!-- acta:end -->`.
- Comments: plain English a 10-year-old reads, say why. No marker tags, no Latin, no emoji.
- Gates before every commit: `gofmt -l .` prints nothing, `go vet ./...` clean, `go test ./...` passes.
- TDD: write the failing test first, watch it fail, then the minimum code.
- Tests that write files do so only inside `t.TempDir()`. Never run `acta` write commands in the worktree, except the `acta tick` lines for this plan.
- Never `git reset`, `rebase`, `amend` or `push`. Never delete test fixtures to make a test pass. Never file acta bugs for defects in this branch; fix them in the task.

## File map

| File | Tasks |
|---|---|
| `internal/voice/voice.go`, `internal/voice/voice_test.go`, `internal/cli/voice.go`, `internal/cli/cli_test.go` | 1 |
| `internal/plugincheck/check.go`, `internal/plugincheck/check_test.go`, `plugin/skills/*/SKILL.md` (frontmatter `description:` line only, all 11) | 2 |
| `internal/doctor/doctor.go` (new), `internal/doctor/doctor_test.go` (new), `internal/cli/doctor.go` (new), `internal/cli/doctor_test.go` (new), `internal/cli/cli.go` (`doctor` case + unknown-command text) | 3 |
| `plugin/skills/scratch/SKILL.md` (new), `plugin/skills/brainstorm/SKILL.md` (body), `internal/plugincheck/skill_scratch_test.go` (new), `internal/plugincheck/skill_brainstorm_test.go`, `internal/hook/hook.go` (`Skills` entry only), `plugin/hooks/default-rules.md` (index line only) | 4 |
| `plugin/skills/setup/SKILL.md` (body), `plugin/skills/build/SKILL.md` (body), `internal/plugincheck/skill_setup_test.go`, `internal/plugincheck/skill_build_test.go` | 5 |
| `plugin/skills/{plan,brainstorm,debug,build,dispatch,review}/SKILL.md` (one paragraph each), `internal/plugincheck/models_test.go` (new) | 6 |
| `internal/hook/hook.go`, `internal/hook/hook_test.go`, `plugin/hooks/default-rules.md`, `plugin/README.md`, `internal/plugincheck/plugin_test.go`, `internal/plugincheck/no_old_names_test.go` (only if README line numbers it pins move) | 7 |

## Waves

- Wave 1: Task 1, Task 2 (no shared files).
- Wave 2: Task 3 (needs Task 1 voice fields), Task 4, Task 5 (need Task 2 prefix in the same SKILL.md files).
- Wave 3: Task 6 (touches SKILL.md files Tasks 4 and 5 own), Task 7 (touches `hook.go` and `default-rules.md` after Task 4).

---

### Task 1: Voice fields `build_executor` and `subagent_models`

**Files:**
- Modify: `internal/voice/voice.go` (`Voice` struct, `Validate`, `fill`)
- Modify: `internal/cli/voice.go` (usage, `show`, `set` flags)
- Test: `internal/voice/voice_test.go`, `internal/cli/cli_test.go`

**verify:** No path writes a voice file whose `build_executor` or `subagent_models` holds a value outside the allowed set, and no path drops an existing field when another flag is set. List every path checked: `Save`, `Load` of a hand-edited file, `voice set` with each new flag alone and combined with old flags, `--clear-subagent-models`.

**Interfaces:**
- Produces: `voice.Voice.BuildExecutor string` (yaml `build_executor,omitempty`), `voice.Voice.SubagentModels string` (yaml `subagent_models,omitempty`). Flags `--executor`, `--subagent-models`, `--clear-subagent-models`. `voice show` prints `build_executor: <x>` and `subagent_models: <x>` lines only when set; `--json` always has both keys.

- [x] **Step 1: Write the failing tests**

```go
// internal/voice/voice_test.go
func TestValidateExecutorAndModels(t *testing.T) {
	for _, tc := range []struct {
		exec, models string
		ok           bool
	}{
		{"", "", true},
		{"subagent", "", true},
		{"dispatch", "split", true},
		{"inline", "", true},
		{"Subagent", "", false},
		{"omp", "", false},
		{"", "all", false},
		{"", "split ", true}, // fill trims
	} {
		v := Default()
		v.BuildExecutor, v.SubagentModels = tc.exec, tc.models
		err := fill(v).Validate()
		if (err == nil) != tc.ok {
			t.Errorf("exec %q models %q: err %v, want ok=%v", tc.exec, tc.models, err, tc.ok)
		}
		if err != nil && !errors.Is(err, ErrBad) {
			t.Errorf("exec %q models %q: err %v is not ErrBad", tc.exec, tc.models, err)
		}
	}
}

func TestLoadRejectsBadExecutor(t *testing.T) {
	p := filepath.Join(t.TempDir(), "voice.yaml")
	os.WriteFile(p, []byte("chat_language: English\nstyle: adhd\nbuild_executor: robot\n"), 0o644)
	if _, exists, err := Load(p); !exists || err == nil {
		t.Fatalf("got exists %v err %v, want a bad-value error", exists, err)
	}
}
```

```go
// internal/cli/cli_test.go (use the file's existing run helper and a temp PM_VOICE_FILE)
func TestVoiceSetExecutorKeepsOtherFields(t *testing.T) {
	t.Setenv("PM_VOICE_FILE", filepath.Join(t.TempDir(), "voice.yaml"))
	mustRun(t, "voice", "set", "--language", "Korean", "--tone", "short")
	mustRun(t, "voice", "set", "--executor", "dispatch")
	mustRun(t, "voice", "set", "--subagent-models", "split")
	out := mustRun(t, "voice", "show")
	for _, want := range []string{"chat_language: Korean", "tone: short", "build_executor: dispatch", "subagent_models: split"} {
		if !strings.Contains(out, want) {
			t.Errorf("show missing %q:\n%s", want, out)
		}
	}
	mustRun(t, "voice", "set", "--clear-subagent-models")
	if out := mustRun(t, "voice", "show"); strings.Contains(out, "subagent_models") {
		t.Errorf("clear left subagent_models:\n%s", out)
	}
}

func TestVoiceSetBadExecutor(t *testing.T) {
	t.Setenv("PM_VOICE_FILE", filepath.Join(t.TempDir(), "voice.yaml"))
	for _, args := range [][]string{
		{"voice", "set", "--executor", "robot"},
		{"voice", "set", "--subagent-models", "all"},
	} {
		if code := runCode(t, args...); code != exitBadInput {
			t.Errorf("%v: exit %d, want %d", args, code, exitBadInput)
		}
	}
}
```

If `cli_test.go` has no `mustRun` / `runCode` helpers, add them at the top of the test file, calling `Run(args, strings.NewReader(""), false, &out, &errb)`.

- [x] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/voice/ ./internal/cli/ -run 'Executor|Models' -v`
Expected: FAIL (unknown field `BuildExecutor`, unknown flag `-executor`).

- [x] **Step 3: Write minimal implementation**

```go
// voice.go, in Voice
	BuildExecutor  string `yaml:"build_executor,omitempty"`
	SubagentModels string `yaml:"subagent_models,omitempty"`

// in Validate, before return nil
	switch v.BuildExecutor {
	case "", "subagent", "dispatch", "inline":
	default:
		return fmt.Errorf("%w: build_executor must be subagent, dispatch or inline, not %q", ErrBad, v.BuildExecutor)
	}
	if v.SubagentModels != "" && v.SubagentModels != "split" {
		return fmt.Errorf("%w: subagent_models must be split or empty, not %q", ErrBad, v.SubagentModels)
	}

// in fill
	v.BuildExecutor = strings.TrimSpace(v.BuildExecutor)
	v.SubagentModels = strings.TrimSpace(v.SubagentModels)
```

In `internal/cli/voice.go`: add the three flags to `set`, include them in the "nothing to set" guard, apply them after the tone lines (`--clear-subagent-models` before `--subagent-models`), print the two lines in `show` when non-empty, add both keys to the JSON map, and extend `voiceUsage` with `[--executor subagent|dispatch|inline] [--subagent-models split] [--clear-subagent-models]`.

- [x] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/voice/ ./internal/cli/ -v`
Expected: PASS

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/voice internal/cli/voice.go internal/cli/cli_test.go
git commit -m "feat(voice): add build_executor and subagent_models"
```

---

### Task 2: `acta: ` prefix on every skill description

**Files:**
- Modify: `internal/plugincheck/check.go` (`SkillProblems`)
- Modify: all 11 `plugin/skills/*/SKILL.md`, the `description:` line only
- Test: `internal/plugincheck/check_test.go`

**verify:** No skill folder can pass `SkillProblems` without a description that starts with exactly `acta: `, and no `name:` changed. List every skill checked and the forms rejected: missing prefix, `acta:` without a space, `Acta: `, prefix in the middle.

**Interfaces:**
- Produces: problem text exactly `description must start with "acta: "`.

- [x] **Step 1: Write the failing test**

Add cases to the existing table in `check_test.go` (it builds skill folders in a temp dir). One case per form:

```go
	{"no prefix", "---\nname: x\ndescription: Use when...\n---\n", "description must start with \"acta: \""},
	{"no space", "---\nname: x\ndescription: \"acta:Use when...\"\n---\n", "description must start with \"acta: \""},
	{"capital", "---\nname: x\ndescription: \"Acta: Use when...\"\n---\n", "description must start with \"acta: \""},
	{"middle", "---\nname: x\ndescription: \"Use acta: when...\"\n---\n", "description must start with \"acta: \""},
	{"good", "---\nname: x\ndescription: \"acta: Use when...\"\n---\n", ""},
```

Match the table's real field names when adding them.

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/plugincheck/ -run TestSkillProblems -v` (use the real test name in `check_test.go`)
Expected: FAIL, the four bad cases report no problem.

- [x] **Step 3: Minimal implementation**

```go
	if !strings.HasPrefix(desc, "acta: ") {
		probs = append(probs, `description must start with "acta: "`)
	}
```

Run `go test ./internal/plugincheck/` now: every `skill_*_test.go` fails on the real skills. Then edit each of the 11 `SKILL.md` descriptions to start with `acta: ` (quote the value when it holds `: `). Change nothing else in those files.

- [x] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/plugincheck/ -v`
Expected: PASS

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/plugincheck/check.go internal/plugincheck/check_test.go plugin/skills/*/SKILL.md
git commit -m "feat(plugin): prefix every skill description with acta:"
```

---

### Task 3: `acta doctor [--fix]`

**Files:**
- Create: `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go`, `internal/cli/doctor.go`, `internal/cli/doctor_test.go`
- Modify: `internal/cli/cli.go` (add `case "doctor": return cmdDoctor(args[1:], stdout, stderr)`, add `doctor` to the unknown-command list)

**verify:** No input can make doctor crash, write outside the repo, or print `ok` for a broken state. List every check and every broken state tried for it (missing file, unreadable JSON, dangling link, value false, value unset, outside git), and confirm `--fix` run twice leaves the repo unchanged on the second run.

**Interfaces:**
- Consumes: `voice.Voice.BuildExecutor` (Task 1), `hook.EnabledPlugins(claudeDir, repoRoot) []string`, `hook.Conflicts(enabled, known []string) []string`, `hook.LoadKnown(path string) []string`, `hook.EnsureGitignore(root, line string) error`, `hook.ClaudeDir() string`, `config.Load(repo, root string)` (`cfg.RepoRoot`, `cfg.Root`), `gitc.CommitPaths(repo string, paths []string, msg string)`.
- Produces:

```go
package doctor

type Level string

const (
	OK   Level = "ok"
	Warn Level = "warn"
	Fail Level = "fail"
)

type Result struct {
	Name, Msg, Fix string
	Level          Level
}

// Env is everything the checks read, so tests can point it at temp folders.
type Env struct {
	Home        string // user home; ~/.claude.json and ~/.omp live here
	ClaudeDir   string // hook.ClaudeDir()
	RepoRoot    string // "" outside a git repo
	ActaRoot    string // the .acta folder for RepoRoot
	Binary      string // os.Executable()
	Version     string // from runtime/debug build info
	KnownFile   string // workflow-plugins.txt; "" skips the conflicts list
	VoiceExists bool
	Voice       voice.Voice
}

func Run(e Env) []Result            // seven results, fixed order
func Fix(e Env) ([]string, error)   // repo-scope only; returns changed paths
func Failed(rs []Result) bool
func Format(rs []Result) string     // "<level> <name>: <msg>\nfix: <fix>\n"
```

Check rules (from spec section 1):

| Name | Result |
|---|---|
| `binary` | Always `ok`, message `<Binary> <Version>`. `fail` only when `Binary` is empty. |
| `harness` | `ok` when an enabled Claude plugin is `acta` or starts with `acta@`, or `Home/.omp/plugins/node_modules/acta` is a link whose target exists. `fail` when that omp link exists but its target is gone (fix `omp plugin link <path to acta/plugin>`), or when neither is present (fix: install the plugin, see README "## Install"). |
| `stale-links` | For each entry in `Home/.omp/plugins/node_modules/` that is a link with a missing target: `warn`, fix `omp plugin unlink <name>` (one Result, names comma-joined, one fix line per name joined by `; `). Folder missing: `ok`. |
| `conflicts` | `hook.Conflicts(hook.EnabledPlugins(ClaudeDir, RepoRoot), hook.LoadKnown(KnownFile))` non-empty: `warn`, fix names the `.claude/settings.local.json` snippet the session hook already prints. |
| `repo` | `RepoRoot == ""`: `ok`, message `not in a git repo, skipped`. `ActaRoot` missing or `.agents.json` not a line in `ActaRoot/.gitignore`: `fail`, fix `acta doctor --fix`. |
| `agents-view` | Read `Home/.claude.json`. Missing file or key missing: `ok`. Key `leftArrowOpensAgents` is `false`: `warn`, fix `Open /config, turn on '← opens agents'`. Broken JSON: `warn`, message holds the parse error. |
| `setup` | `!VoiceExists` or `Voice.BuildExecutor == ""`: `warn`, fix `/acta:setup`. |

- [x] **Step 1: Write the failing tests**

`internal/doctor/doctor_test.go`, table-driven. A helper builds a temp `Env`:

```go
func env(t *testing.T) Env {
	t.Helper()
	home := t.TempDir()
	repo := t.TempDir()
	if out, err := exec.Command("git", "-C", repo, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	return Env{Home: home, ClaudeDir: filepath.Join(home, ".claude"), RepoRoot: repo,
		ActaRoot: filepath.Join(repo, ".acta"), Binary: "/bin/acta", Version: "test",
		VoiceExists: true, Voice: voice.Voice{BuildExecutor: "subagent"}}
}

func byName(rs []Result, name string) Result {
	for _, r := range rs {
		if r.Name == name {
			return r
		}
	}
	return Result{}
}
```

Cases, each its own `t.Run` asserting `Level`, and `Fix` text where the table above names one:
1. Order: `Run` returns names in exactly `binary, harness, stale-links, conflicts, repo, agents-view, setup`.
2. `harness`: Claude settings `{"enabledPlugins":{"acta@local":true}}` in `ClaudeDir/settings.json` gives `ok`; nothing gives `fail`; omp link to a deleted dir gives `fail`; omp link to a live dir gives `ok`.
3. `stale-links`: `node_modules/pm` links to a missing path gives `warn` with fix containing `omp plugin unlink pm`; a live link gives `ok`; no folder gives `ok`.
4. `conflicts`: known file listing `superpowers` plus enabled `superpowers@x` gives `warn`; `KnownFile == ""` gives `ok`.
5. `repo`: fresh repo gives `fail`; after `Fix` gives `ok`; `RepoRoot == ""` gives `ok` with `skipped`.
6. `agents-view`: no file `ok`; `{}` `ok`; `{"leftArrowOpensAgents":true}` `ok`; `{"leftArrowOpensAgents":false}` `warn` with the `/config` fix; `{bad json` `warn`, no panic.
7. `setup`: `VoiceExists=false` `warn`; `BuildExecutor=""` `warn`; set gives `ok`.
8. `Fix` twice: second call returns no paths, and `.gitignore` holds `.agents.json` exactly once.
9. `Fix` never touches `Home`: snapshot every file under `Home` before and after, they match.
10. `Failed` true only when some `Level == Fail`.

`internal/cli/doctor_test.go`: `acta doctor` in a temp repo with `HOME` set to a temp dir exits 1 (repo check fails) and prints `fail repo:`; `acta doctor --fix` exits 0 on the next plain run for the repo line; `acta doctor extra` exits `exitBadInput`.

- [x] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/doctor/ ./internal/cli/ -run Doctor -v`
Expected: FAIL, package `doctor` does not exist.

- [x] **Step 3: Minimal implementation**

`internal/doctor/doctor.go`: one small function per check (`checkBinary`, `checkHarness`, `checkStaleLinks`, `checkConflicts`, `checkRepo`, `checkAgentsView`, `checkSetup`), `Run` calls them in order. Links: `os.Lstat` for `ModeSymlink`, then `os.Stat` for the target. JSON: `json.Unmarshal` into `struct{ LeftArrowOpensAgents *bool }` with tag `json:"leftArrowOpensAgents"`, so unset and false differ. `Fix`: `os.MkdirAll(ActaRoot, 0o755)` when missing, then `hook.EnsureGitignore(ActaRoot, ".agents.json")`; return `ActaRoot/.gitignore` only when its bytes changed.

`internal/cli/doctor.go`:

```go
const doctorUsage = "usage: acta doctor [--fix] [--known <file>]"

func cmdDoctor(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fix := fs.Bool("fix", false, "fix repo problems (.acta folder, .gitignore)")
	known := fs.String("known", "", "file listing workflow plugins that clash with acta")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		fmt.Fprintln(stderr, doctorUsage)
		return exitBadInput
	}
	e := doctorEnv(*known)
	if *fix {
		paths, err := doctor.Fix(e)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitOther
		}
		if len(paths) > 0 {
			gitc.CommitPaths(e.RepoRoot, paths, "acta: doctor fix")
			fmt.Fprintf(stdout, "fixed repo: %s\n", strings.Join(paths, ", "))
		}
	}
	rs := doctor.Run(e)
	fmt.Fprint(stdout, doctor.Format(rs))
	if doctor.Failed(rs) {
		return 1
	}
	return exitOK
}
```

`doctorEnv` fills `Env` from `os.UserHomeDir`, `hook.ClaudeDir()`, `config.Load(cwd, "")` (on error `RepoRoot` stays `""`), `os.Executable`, `debug.ReadBuildInfo` (`Main.Version` plus the `vcs.revision` setting when present), and `voice.Resolve()`. Check the real `gitc.CommitPaths` return type and handle its failure the way `cmdScratchNew` does.

- [x] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/doctor/ ./internal/cli/ -v`
Expected: PASS

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/doctor internal/cli/doctor.go internal/cli/doctor_test.go internal/cli/cli.go
git commit -m "feat(cli): add acta doctor with repo-only --fix"
```

---

### Task 4: `acta:scratch` skill and brainstorm rules

**Files:**
- Create: `plugin/skills/scratch/SKILL.md`, `internal/plugincheck/skill_scratch_test.go`
- Modify: `plugin/skills/brainstorm/SKILL.md` (body), `internal/plugincheck/skill_brainstorm_test.go`
- Modify: `internal/hook/hook.go` (add one `Skills` entry), `plugin/hooks/default-rules.md` (add the same index line), because `final_test.go` requires every skill folder to be in `hook.Skills`

**verify:** No reading of either skill lets an agent put a raw idea in agent memory, start an Architectural brainstorm without a scratch item, run a second Architectural brainstorm in the same session, or offer the herdr choice without `HERDR_ENV=1`. List each rule from spec 4.1 and 4.2 and the sentence that carries it.

- [x] **Step 1: Write the failing tests**

```go
// skill_scratch_test.go
func TestSkillScratch(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "scratch",
		MaxLines: 60,
		Must: []string{
			"acta scratch new", "acta scratch add", "verbatim", "images", "Filed SCRATCH-",
			"catet", "nanti", "kepikiran", "File this in Scratchpad?", "never", "memory",
			"new session", "main branch", "status dropped",
		},
		MustNot: []string{"superpowers:"},
	})
}
```

In `skill_brainstorm_test.go`, add to `Must`: `"acta scratch new"`, `"acta scratch add"`, `"status brainstorming"`, `"parent: scratch/"`, `"One Architectural brainstorm per session"`, `"claude --bg 'brainstorm SCRATCH-"`, `"HERDR_ENV=1"`, `"pbcopy"`, `"wl-copy"`, `"xclip"`, `"OSC 52"`, `"press ←"`, `"Spike and Bounded"`. Raise `MaxLines` by the lines you add, no more.

- [x] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/plugincheck/ -run 'SkillScratch|SkillBrainstorm' -v`
Expected: FAIL, `missing skills/scratch/SKILL.md` and missing brainstorm phrases.

- [x] **Step 3: Write the skill text**

`plugin/skills/scratch/SKILL.md`, frontmatter `name: scratch`, `description: "acta: Use when the user drops a raw idea (catet, nanti, kepikiran, note this, later) or a side idea shows up during other work. Files it in Scratchpad with acta scratch new; never in agent memory."`. Body carries spec section 4.1 as written, in short plain sentences, plus the exact commands `acta scratch new <slug> [--title T] < body.md`, `acta scratch add SCRATCH-n < text.md`, `acta set scratch/<stem> status dropped`.

`plugin/skills/brainstorm/SKILL.md`: add a section "Architectural path: scratch item and one per session" right after "Three Paths", carrying spec section 4.2 as written, the three choices listed (a), (b), (c) with the exact commands. Add "Step 0: scratch item" as the first item of the Architectural checklist and "append with acta scratch add" to the questions and design steps. The spec frontmatter line `parent: scratch/<stem>` goes in "Documentation".

- [x] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/plugincheck/ -v`
Expected: PASS, including `final_test.go` and any hook parity test, after adding `{"scratch", "raw ideas (\"catet\", \"nanti\", side ideas); file with acta scratch new, never memory"}` after `bug` in `hook.Skills` and the same line `- acta:scratch: raw ideas ("catet", "nanti", side ideas); file with acta scratch new, never memory` after the `acta:bug` line in `default-rules.md`.

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add plugin/skills/scratch plugin/skills/brainstorm/SKILL.md internal/plugincheck/skill_scratch_test.go internal/plugincheck/skill_brainstorm_test.go internal/hook/hook.go plugin/hooks/default-rules.md
git commit -m "feat(plugin): add acta:scratch and brainstorm scratch rules"
```

---

### Task 5: `acta:setup` first run and `acta:build` saved executor

**Files:**
- Modify: `plugin/skills/setup/SKILL.md`, `plugin/skills/build/SKILL.md`
- Test: `internal/plugincheck/skill_setup_test.go`, `internal/plugincheck/skill_build_test.go`

**verify:** No reading of setup lets an agent write CLAUDE.md or AGENTS.md without first showing the exact block and hearing a yes, write outside the markers, create a CLAUDE.md or AGENTS.md that did not exist, offer dispatch without herdr, or ask the model question outside Claude Code. No reading of build asks for an executor when one is saved. List each rule and its sentence.

- [x] **Step 1: Write the failing tests**

`skill_setup_test.go`: replace `"It never edits CLAUDE.md"` with, and add: `"acta doctor"`, `"acta doctor --fix"`, `"--executor"`, `"HERDR_ENV=1"`, `"herdr"`, `"--subagent-models split"`, `"Claude Code only"`, `"<!-- acta:begin -->"`, `"<!-- acta:end -->"`, `"only after a yes"`, `"only to files that already exist"`, `"never edits settings"`, `"which part to change"`. Add `MustNot: "It never edits CLAUDE.md"`. Raise `MaxLines` to fit, no more.

`skill_build_test.go`: add `"acta voice show"` and `"build_executor"` to its `Must`.

- [x] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/plugincheck/ -run 'SkillSetup|SkillBuild' -v`
Expected: FAIL on the new phrases.

- [x] **Step 3: Write the skill text**

`setup/SKILL.md`: description `acta: Use on first run and whenever the user asks to change setup: runs acta doctor, then the chat language, style and tone, the default build executor, split subagent models, and the optional acta block in CLAUDE.md or AGENTS.md.` Body follows spec section 4.3 as written, including the exact block. Keep the existing "Change later" flags and add the new ones.

`build/SKILL.md`: where it asks for the executor, add: read `build_executor` from `acta voice show`; when set, use it without asking; when unset, ask as today.

- [x] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/plugincheck/ -v`
Expected: PASS

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add plugin/skills/setup/SKILL.md plugin/skills/build/SKILL.md internal/plugincheck/skill_setup_test.go internal/plugincheck/skill_build_test.go
git commit -m "feat(plugin): first-run setup and saved build executor"
```

---

### Task 6: Optional subagent-model sentence in six skills

**Files:**
- Modify: `plugin/skills/{plan,brainstorm,debug,build,dispatch,review}/SKILL.md` (one paragraph each)
- Create: `internal/plugincheck/models_test.go`

**verify:** In none of the six skills can the model rule apply without both `subagent_models: split` and Claude Code, and no skill names a model otherwise. List the six skills and confirm each has the same paragraph, word for word.

**Interfaces:**
- The paragraph, exact, placed where each skill first dispatches a subagent:

```markdown
**Subagent models.** Only when `acta voice show` lists `subagent_models: split` and you run in Claude Code: subagents that write code use `model: "sonnet"`; all other subagents (mapping, explore, planning help, debug investigation, spikes) use `model: "opus"`; reviewers use your own model alias. Otherwise name no model and follow the user's own config.
```

- [x] **Step 1: Write the failing test**

```go
// models_test.go
package plugincheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const modelsPara = "**Subagent models.** Only when `acta voice show` lists `subagent_models: split` and you run in Claude Code: subagents that write code use `model: \"sonnet\"`; all other subagents (mapping, explore, planning help, debug investigation, spikes) use `model: \"opus\"`; reviewers use your own model alias. Otherwise name no model and follow the user's own config."

func TestModelsParagraphInSixSkills(t *testing.T) {
	for _, s := range []string{"plan", "brainstorm", "debug", "build", "dispatch", "review"} {
		raw, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", s, "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		if n := strings.Count(string(raw), modelsPara); n != 1 {
			t.Errorf("skills/%s: models paragraph found %d times, want 1", s, n)
		}
	}
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/plugincheck/ -run TestModelsParagraphInSixSkills -v`
Expected: FAIL, found 0 times in all six.

- [x] **Step 3: Add the paragraph**

Paste the paragraph once into each of the six files. If a skill already names a model for a subagent (for example `dispatch` or `review`), make that text defer to the paragraph instead of fixing a model on its own; keep the reviewer line consistent with "your own model alias". Raise each skill's `MaxLines` in its test by the lines added, no more.

- [x] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/plugincheck/ -v`
Expected: PASS

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add plugin/skills internal/plugincheck
git commit -m "feat(plugin): optional split subagent models in six skills"
```

---

### Task 7: Session rules, hook text and README

**Files:**
- Modify: `internal/hook/hook.go` (`Skills`, `coreRules`, `firstRun`, `Prompt` first-run line), `plugin/hooks/default-rules.md`, `plugin/README.md`
- Test: `internal/hook/hook_test.go`, `internal/plugincheck/plugin_test.go`; `internal/plugincheck/no_old_names_test.go` only if the README line it pins moves

**verify:** The session text and its fallback copy cannot drift: every line of `default-rules.md` comes from `SessionStart` for an English adhd voice with no conflicts. The old "never edits your CLAUDE.md" promise cannot come back in the README in any form. List each surface checked: hook text, fallback file, README, omp fallback in `plugin/omp/index.ts` (reads the file, no change).

- [x] **Step 1: Write the failing tests**

`hook_test.go`:
- `SessionStart` output contains `- acta:scratch: raw ideas ("catet", "nanti", side ideas); file with acta scratch new, never memory`, the setup line `- acta:setup: first-run setup and later changes: doctor, voice, build executor, subagent models, CLAUDE.md block`, and rule `8. One Architectural brainstorm per session; a second one becomes a scratch item and the user picks how to open it.`
- With `VoiceExists=false`, output contains `/acta:setup` and no longer contains `1. Which language should chat use?`.
- `Prompt` with `VoiceExists=false` contains `/acta:setup`.
- Parity (skip if a parity test already exists in `plugin_test.go`; extend it instead): `SessionStart(Input{Voice: voice.Default(), VoiceExists: true})` equals the content of `plugin/hooks/default-rules.md`, after trimming both.

`plugin_test.go` (`TestNoticeAndReadme`): add `"## First run"`, `"acta doctor"`, `"/acta:setup"`, `"acta: "` (omp note), `"only between acta markers"` to the README list, and fail if the README contains `never edits your CLAUDE.md` in any case (`strings.Contains(strings.ToLower(readme), "never edits your claude.md")`).

- [x] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/hook/ ./internal/plugincheck/ -v`
Expected: FAIL on the new lines and README phrases.

- [x] **Step 3: Implement**

- `hook.Skills`: the `scratch` entry is already there from Task 4; change the `setup` entry to `first-run setup and later changes: doctor, voice, build executor, subagent models, CLAUDE.md block`.
- `coreRules`: add rule 8 as above.
- `firstRun`: replace the three-question block with: voice is not set up; before other work, run `/acta:setup` (it runs `acta doctor`, then asks the setup questions); until then write in English (or the language CLAUDE.md names), adhd style. Keep the "CLAUDE.md already names a language" sentence.
- `Prompt` first-run string: `acta voice: not set up yet; reply in English, adhd style, and run /acta:setup once (see the session rules).`
- Regenerate `plugin/hooks/default-rules.md` from `SessionStart` for the default voice so the parity test holds.
- README: replace line "It never edits your CLAUDE.md, AGENTS.md or settings." with "It edits CLAUDE.md or AGENTS.md only between acta markers, and only after your yes in /acta:setup. It never edits settings."; add a "## First run" section (install, then `/acta:setup`, which runs `acta doctor`; `acta doctor --fix` fixes repo items); add one omp line: skill names have no prefix in omp, and every description starts with `acta: `.

- [x] **Step 4: Run tests to verify they pass**

Run: `go test ./...`
Expected: PASS

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/hook plugin/hooks/default-rules.md plugin/README.md internal/plugincheck
git commit -m "feat(hook): scratch skill, one brainstorm per session, first run via /acta:setup"
```

## Fix round 1

### Task 8: Make `acta doctor --fix` commit and write like other acta write commands

Review round 1 (range 622a74a..69fd51c) found four BLOCKERs, all in doctor. One task, one commit.

**Files:** `internal/cli/doctor.go`, `internal/doctor/doctor.go`, `internal/cli/doctor_test.go`, `internal/doctor/doctor_test.go`

1. `internal/cli/doctor.go:44` commits with `gitc.CommitPaths` and no dirty-before check. A user's own uncommitted `.acta/.gitignore` lines get committed (probe: append `mysecret.txt`, run `--fix`, commit carries it). Take the dirty state before `Fix` and skip the commit when the file was dirty, the way `write/ops.go:298-315` does (`gitc.Commit(..., wasDirty)`).
2. `internal/cli/doctor.go:41-46` ignores `auto_commit: false`. Carry `cfg.AutoCommit` in `doctor.Env` and do not commit when it is false (same rule as `write/ops.go:307`).
3. `internal/doctor/doctor.go:70-73` runs `MkdirAll` and `EnsureGitignore` on `ActaRoot` even when `root:` in `.acta.yaml` or `ACTA_ROOT` points outside `RepoRoot` (`../x` or an absolute path). Global Constraint: doctor never writes outside the repo. When `ActaRoot` is not inside `RepoRoot`, `Fix` writes nothing and the repo check reports `fail` with a hint to fix `root`, not `acta doctor --fix`.
4. `internal/cli/doctor.go:66-68` drops the `config.Load` error, so a broken `.acta.yaml` (for example `root: [`) prints `ok repo: not in a git repo, skipped`. Carry the error in `doctor.Env` and make the repo check report `fail repo:` with the parse error.

- [x] **Step 1: Write failing tests**, one per item above, each with the concrete probe input: dirty `.gitignore` is not committed; `auto_commit: false` gives no commit; `root: ../escaped` creates nothing outside the repo and reports `fail`; `root: [` reports `fail`, never `ok`. Also one test that a clean `--fix` does make the `acta: doctor fix` commit (the spec reviewer showed removing `CommitPaths` stays green today).
- [x] **Step 2: Run** `go test ./internal/cli/ ./internal/doctor/ -v`. Expected: FAIL on the new tests.
- [x] **Step 3: Implement** the minimum to pass.
- [x] **Step 4: Run** `go test ./...`. Expected: PASS.
- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/cli internal/doctor .acta/plans/2026-09-28-skill-rules.md
git commit -m "fix(doctor): respect dirty files, auto_commit and repo bounds in --fix"
```

## Fix round 2

### Task 9: Doctor repo bounds follow symlinks

Review round 2 (range 69fd51c..17f9a9e) found one BLOCKER, from both reviewers. `insideRepo` (`internal/doctor/doctor.go:87-93`) compares path text only. `RepoRoot` comes from git with symlinks resolved; `ActaRoot` does not. Two wrong outputs:

1. A symlink lets `--fix` write outside the repo and then print `ok repo`. Probes: `.acta -> ../outside`; `root: link` with `link -> ../outside`; a committed `.acta/.gitignore -> ../../outside/victim.conf` (plain `acta doctor` says run `--fix`, and `--fix` appends `.agents.json` to `victim.conf`).
2. Regression from Task 8: a root that is inside the repo but reached through a symlink (macOS `/var` -> `/private/var`, for example `ACTA_ROOT=$REPO/.acta` under `$TMPDIR`) now reports `fail repo: ... is outside the repo` and exits 1. It worked at 69fd51c.

Fix inside doctor only: resolve symlinks with `filepath.EvalSymlinks` on `RepoRoot` and on `ActaRoot` (or its deepest existing parent) before `filepath.Rel`. Treat a `.gitignore` that is itself a symlink as outside the repo: `Fix` writes nothing and the repo check reports `fail`. The same write-through-symlink in `hook.EnsureGitignore` callers (`internal/cli/hook.go:45`, `internal/cli/tick.go:115`) was already on the parent branch; it is a separate bug and is not part of this task.

**Files:** `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go`, `internal/cli/doctor_test.go`

- [x] **Step 1: Write failing tests** with the probe inputs above: `.acta` symlinked outside writes nothing and reports `fail`; `root:` pointing at an in-repo symlink to outside writes nothing and reports `fail`; `.acta/.gitignore` symlinked to an outside file leaves that file unchanged and reports `fail`; a root reached through a symlinked parent folder that really is inside the repo reports `ok` and `--fix` commits.
- [x] **Step 2: Run** `go test ./internal/doctor/ ./internal/cli/ -v`. Expected: FAIL on the new tests.
- [x] **Step 3: Implement** the minimum to pass.
- [x] **Step 4: Run** `go test ./...`. Expected: PASS.
- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/doctor internal/cli .acta/plans/2026-09-28-skill-rules.md
git commit -m "fix(doctor): resolve symlinks before the repo bounds check"
```

## Fix round 3

### Task 10: Refuse a `.gitignore` that is a symlink

Review round 3 (range 17f9a9e..8f7abb9) found that Task 9's rule "a `.gitignore` that is itself a symlink counts as outside the repo" was not built. `realPath` (`internal/doctor/doctor.go:103-113`) falls back to the parent folder when `EvalSymlinks` fails, and the `.gitignore` path is never checked with `os.Lstat`. The user ruled on 2026-09-28 to fix this past the three-round budget.

1. Dangling link: a committed `.acta/.gitignore -> ../../outside/new.conf` (`outside/` exists, `new.conf` does not). `--fix` creates `outside/new.conf`.
2. In-repo link: `.acta/.gitignore -> ../.git/config` (or `../docs/notes.txt`). `--fix` appends `.agents.json` to that file and then reports `ok repo`, while git ignores nothing.

Fix: in both `Fix` (`doctor.go:70`) and `checkRepo` (`doctor.go:266`), call `os.Lstat` on the `.gitignore` path. Any symlink, dangling or not, inside or outside the repo, means `Fix` writes nothing and the repo check reports `fail`.

**Files:** `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go`, `internal/cli/doctor_test.go`

- [x] **Step 1: Write failing tests** for both inputs above: the dangling link target is still missing after `--fix` and the report is `fail`; the in-repo target file is unchanged after `--fix` and the report is `fail`, never `ok`.
- [x] **Step 2: Run** `go test ./internal/doctor/ ./internal/cli/ -v`. Expected: FAIL on the new tests.
- [x] **Step 3: Implement** the minimum to pass.
- [x] **Step 4: Run** `go test ./...`. Expected: PASS.
- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/doctor internal/cli .acta/plans/2026-09-28-skill-rules.md
git commit -m "fix(doctor): refuse a .gitignore that is a symlink"
```
