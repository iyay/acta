---
parent: specs/2026-10-01-plan-depth-config-design
id: PLN-0065
created: "2026-10-01 05:32:58"
hash: iejqcfl
started: "2026-10-01 05:57:35"
---
# Plan Depth Setting and Repo-Level Config Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** A `plan_depth` setting (`minimal` or `full`) that `.acta.yaml` can override per repo, together with `repo_language` and `build_executor`, and plan, build and setup skills that act on it.

**Architecture:** `internal/config` gains `User.PlanDepth` and a new file `repo_user.go` with `MergeRepo` (lays the three repo keys of `.acta.yaml` over the global user config and says which ones it set) and `SaveRepoUser` (writes those keys into `.acta.yaml` through a `yaml.Node`, so comments and other keys stay). `acta config show` and the session hook read through `MergeRepo`; `acta config set --repo` writes through `SaveRepoUser`. The skills and core rule 2 learn the minimal plan.

**Tech Stack:** Go, `gopkg.in/yaml.v3` (already used in `internal/config`).

**Spec:** `.acta/specs/2026-10-01-plan-depth-config-design.md`

**Tests:** fast `scripts/test`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Only three keys may be set per repo: `repo_language`, `build_executor`, `plan_depth`. The personal keys `chat_language`, `style`, `tone`, `theme`, `subagent_models` in `.acta.yaml` are an error that names the key.
- `plan_depth` takes `minimal` or `full`; empty means `full`. Any other value fails with `ErrBadUser`, the same way a bad `build_executor` does.
- `config.Load` does not change and never fails because of these keys; only `MergeRepo` checks them. The TUI and every other command keep working with a bad key in `.acta.yaml`.
- Every test that touches the user config sets `PM_VOICE_FILE` or `HOME` to a temp path, so the real `~/.acta/config.yaml` is never written.
- Comments are plain English a 10-year-old can read. They say why, not what.
- PLN-0063 edits `internal/cli/cli_test.go`, `plugin/skills/build/SKILL.md` and `plugin/skills/brainstorm/SKILL.md` in another branch. New CLI tests go in a new file, `internal/cli/config_repo_test.go`, so the two branches do not clash.

## File map

- Modify: `internal/config/user.go` (`User.PlanDepth`, `Validate`, `fill`)
- Create: `internal/config/repo_user.go` (`RepoKeys`, `MergeRepo`, `SaveRepoUser`)
- Create: `internal/config/repo_user_test.go`
- Modify: `internal/cli/config_cmd.go` (`show` merges, `set --plan-depth`, `set --repo`)
- Create: `internal/cli/config_repo_test.go`
- Modify: `internal/cli/hook.go` (`loadVoice` merges the repo keys)
- Modify: `internal/hook/hook.go` (core rule 2), `internal/hook/hook_test.go`, `plugin/hooks/default-rules.md`
- Modify: `plugin/skills/plan/SKILL.md`, `plugin/skills/build/SKILL.md`, `plugin/skills/setup/SKILL.md`
- Modify: `internal/plugincheck/skill_plan_test.go`, `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/skill_setup_test.go`

## Waves

- Wave 1: Task 1 (config package), Task 4 (skills and plugincheck). Disjoint files.
- Wave 2: Task 2 (config command), Task 3 (hook and rule 2). Both need Task 1; they touch different files.

---

### Task 1: plan_depth and the repo layer in internal/config

**Files:**
- Modify: `internal/config/user.go`
- Create: `internal/config/repo_user.go`
- Test: `internal/config/repo_user_test.go`

**verify:** For every one of the three repo keys, a value in `.acta.yaml` beats the global value, and the returned set names exactly the keys `.acta.yaml` holds. Every personal key in `.acta.yaml` gives an error that names it. A bad `plan_depth` or `build_executor` fails with `ErrBadUser` whether it sits in the global file or in `.acta.yaml`. A missing `.acta.yaml` changes nothing. `SaveRepoUser` keeps every other key and every comment already in `.acta.yaml`, adds or replaces only the keys it was given, and refuses a personal key. List the cases checked.

**Interfaces:**
- Consumes: `User`, `ErrBadUser`, `fill`, `(User).Validate` in `user.go`.
- Produces:
  - `User.PlanDepth string` with yaml tag `plan_depth,omitempty`.
  - `var RepoKeys = []string{"repo_language", "build_executor", "plan_depth"}`
  - `func MergeRepo(v User, repoRoot string) (User, map[string]bool, error)` — on error it returns `v` unchanged.
  - `func SaveRepoUser(repoRoot string, set map[string]string) (string, error)` — returns the path it wrote.

- [x] **Step 1: Write the failing tests**

Create `internal/config/repo_user_test.go`:

```go
package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeRepoYAML(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".acta.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// A value in .acta.yaml wins over the global one, key by key, and the set
// says which keys came from the repo so show can label them.
func TestMergeRepoBeatsGlobal(t *testing.T) {
	global := User{ChatLanguage: "Indonesian", Style: "adhd", RepoLanguage: "English", BuildExecutor: "dispatch", PlanDepth: "full"}
	cases := []struct {
		yaml string
		want User
		from []string
	}{
		{"repo_language: Korean\n", User{ChatLanguage: "Indonesian", Style: "adhd", RepoLanguage: "Korean", BuildExecutor: "dispatch", PlanDepth: "full"}, []string{"repo_language"}},
		{"build_executor: inline\n", User{ChatLanguage: "Indonesian", Style: "adhd", RepoLanguage: "English", BuildExecutor: "inline", PlanDepth: "full"}, []string{"build_executor"}},
		{"plan_depth: minimal\n", User{ChatLanguage: "Indonesian", Style: "adhd", RepoLanguage: "English", BuildExecutor: "dispatch", PlanDepth: "minimal"}, []string{"plan_depth"}},
		{"root: .acta\nplan_depth: minimal\nbuild_executor: inline\n", User{ChatLanguage: "Indonesian", Style: "adhd", RepoLanguage: "English", BuildExecutor: "inline", PlanDepth: "minimal"}, []string{"build_executor", "plan_depth"}},
		{"root: .acta\n", global, nil},
	}
	for _, c := range cases {
		got, from, err := MergeRepo(global, writeRepoYAML(t, c.yaml))
		if err != nil {
			t.Fatalf("%q: %v", c.yaml, err)
		}
		if got != c.want {
			t.Errorf("%q: got %+v, want %+v", c.yaml, got, c.want)
		}
		if len(from) != len(c.from) {
			t.Errorf("%q: from %v, want %v", c.yaml, from, c.from)
		}
		for _, k := range c.from {
			if !from[k] {
				t.Errorf("%q: from lacks %s: %v", c.yaml, k, from)
			}
		}
	}
}

// No .acta.yaml means the global config is the whole answer.
func TestMergeRepoMissingFile(t *testing.T) {
	global := User{ChatLanguage: "English", Style: "adhd", RepoLanguage: "English"}
	got, from, err := MergeRepo(global, t.TempDir())
	if err != nil || got != global || len(from) != 0 {
		t.Errorf("got %+v %v %v, want the global config and nothing from the repo", got, from, err)
	}
}

// Personal settings follow the person, not the repo, so a repo file that
// tries to set one is refused, and the message names the key.
func TestMergeRepoRefusesPersonalKeys(t *testing.T) {
	for _, k := range []string{"chat_language", "style", "tone", "theme", "subagent_models"} {
		_, _, err := MergeRepo(UserDefault(), writeRepoYAML(t, k+": x\n"))
		if err == nil || !strings.Contains(err.Error(), k) {
			t.Errorf("%s in .acta.yaml: err %v, want one naming the key", k, err)
		}
	}
}

// A bad value fails the same way in either file.
func TestPlanDepthBadValue(t *testing.T) {
	for _, body := range []string{"plan_depth: deep\n", "build_executor: robot\n"} {
		if _, _, err := MergeRepo(UserDefault(), writeRepoYAML(t, body)); !errors.Is(err, ErrBadUser) {
			t.Errorf(".acta.yaml %q: err %v, want ErrBadUser", body, err)
		}
		path := filepath.Join(t.TempDir(), "config.yaml")
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, _, err := LoadUser(path); !errors.Is(err, ErrBadUser) {
			t.Errorf("global %q: err %v, want ErrBadUser", body, err)
		}
	}
	for _, ok := range []string{"", "minimal", "full"} {
		v := UserDefault()
		v.PlanDepth = ok
		if err := v.Validate(); err != nil {
			t.Errorf("plan_depth %q: %v", ok, err)
		}
	}
}

// Saving keeps what the user already wrote in .acta.yaml, comments too.
func TestSaveRepoUserKeepsOtherKeys(t *testing.T) {
	dir := writeRepoYAML(t, "# where the planning files live\nroot: docs/acta\nplan_depth: full\n")
	path, err := SaveRepoUser(dir, map[string]string{"plan_depth": "minimal", "build_executor": "inline"})
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(dir, ".acta.yaml") {
		t.Errorf("wrote %s", path)
	}
	raw, _ := os.ReadFile(path)
	got := string(raw)
	for _, want := range []string{"# where the planning files live", "root: docs/acta", "plan_depth: minimal", "build_executor: inline"} {
		if !strings.Contains(got, want) {
			t.Errorf("file lacks %q:\n%s", want, got)
		}
	}
	if strings.Count(got, "plan_depth") != 1 {
		t.Errorf("plan_depth written twice:\n%s", got)
	}
	// A new file is made when there is none.
	fresh := t.TempDir()
	if _, err := SaveRepoUser(fresh, map[string]string{"plan_depth": "minimal"}); err != nil {
		t.Fatal(err)
	}
	if v, from, err := MergeRepo(UserDefault(), fresh); err != nil || v.PlanDepth != "minimal" || !from["plan_depth"] {
		t.Errorf("fresh file: %+v %v %v", v, from, err)
	}
	// A personal key is never written.
	if _, err := SaveRepoUser(fresh, map[string]string{"chat_language": "Korean"}); err == nil {
		t.Error("SaveRepoUser wrote chat_language")
	}
}
```

- [x] **Step 2: Run the tests to see them fail**

Run: `scripts/test ./internal/config -run 'MergeRepo|PlanDepth|SaveRepoUser'`
Expected: FAIL to build: `MergeRepo`, `SaveRepoUser` and `User.PlanDepth` are undefined.

- [x] **Step 3: Write the minimal code**

In `internal/config/user.go`, add the field under `SubagentModels`:

```go
	// PlanDepth says how much a plan spells out: minimal (short steps, no
	// code) or full (real code in every step). Empty means full.
	PlanDepth string `yaml:"plan_depth,omitempty"`
```

In `Validate`, after the `SubagentModels` switch:

```go
	switch v.PlanDepth {
	case "", "minimal", "full":
	default:
		return fmt.Errorf("%w: plan_depth must be minimal or full, not %q", ErrBadUser, v.PlanDepth)
	}
```

In `fill`, next to the other trims:

```go
	v.PlanDepth = strings.TrimSpace(v.PlanDepth)
```

Create `internal/config/repo_user.go`:

```go
package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// RepoKeys are the user settings a repo may set for itself in .acta.yaml,
// in the order config show prints them. Everything else in User belongs to
// the person and follows them from repo to repo.
var RepoKeys = []string{"repo_language", "build_executor", "plan_depth"}

// personalKeys may never sit in .acta.yaml: that file is committed, so a
// value there would change how the agent talks to everyone who clones it.
var personalKeys = []string{"chat_language", "style", "tone", "theme", "subagent_models"}

// MergeRepo lays the repo keys of repoRoot/.acta.yaml over v. The map holds
// each key the repo file set, so show can say where a value came from. On
// any error v comes back as it was, so a hook can still use it.
func MergeRepo(v User, repoRoot string) (User, map[string]bool, error) {
	path := filepath.Join(repoRoot, ".acta.yaml")
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return v, map[string]bool{}, nil
	}
	if err != nil {
		return v, nil, err
	}
	var keys map[string]any
	if err := yaml.Unmarshal(raw, &keys); err != nil {
		return v, nil, fmt.Errorf("%s: %w", path, err)
	}
	for _, k := range personalKeys {
		if _, ok := keys[k]; ok {
			return v, nil, fmt.Errorf("%s: %w: %s is a personal setting; set it with acta config set, not per repo", path, ErrBadUser, k)
		}
	}
	out, from := v, map[string]bool{}
	for _, k := range RepoKeys {
		s, ok := keys[k].(string)
		if !ok || strings.TrimSpace(s) == "" {
			continue
		}
		from[k] = true
		switch k {
		case "repo_language":
			out.RepoLanguage = strings.TrimSpace(s)
		case "build_executor":
			out.BuildExecutor = strings.TrimSpace(s)
		case "plan_depth":
			out.PlanDepth = strings.TrimSpace(s)
		}
	}
	if err := out.Validate(); err != nil {
		return v, nil, fmt.Errorf("%s: %w", path, err)
	}
	return out, from, nil
}

// SaveRepoUser writes the given repo keys into repoRoot/.acta.yaml and
// returns its path. It edits the yaml tree instead of rewriting the file,
// so the user's comments and other keys (root, dirs, links) stay.
func SaveRepoUser(repoRoot string, set map[string]string) (string, error) {
	path := filepath.Join(repoRoot, ".acta.yaml")
	for k := range set {
		if !slices.Contains(RepoKeys, k) {
			return path, fmt.Errorf("%w: %s cannot be set per repo; only %s can", ErrBadUser, k, strings.Join(RepoKeys, ", "))
		}
	}
	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return path, err
	}
	var doc yaml.Node
	if len(bytes.TrimSpace(raw)) > 0 {
		if err := yaml.Unmarshal(raw, &doc); err != nil {
			return path, fmt.Errorf("%s: %w", path, err)
		}
	}
	if doc.Kind == 0 {
		doc = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode}}}
	}
	m := doc.Content[0]
	if m.Kind != yaml.MappingNode {
		return path, fmt.Errorf("%s: the top level is not a list of keys", path)
	}
	// Walk RepoKeys, not the map, so the file comes out the same every time.
	for _, k := range RepoKeys {
		val, ok := set[k]
		if !ok {
			continue
		}
		node := &yaml.Node{Kind: yaml.ScalarNode, Value: val}
		i := slices.IndexFunc(m.Content, func(n *yaml.Node) bool { return n.Value == k })
		// Keys sit at even places and values right after them.
		if i >= 0 && i%2 == 0 {
			m.Content[i+1] = node
			continue
		}
		m.Content = append(m.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: k}, node)
	}
	out, err := yaml.Marshal(&doc)
	if err != nil {
		return path, err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return path, err
	}
	return path, os.Rename(tmp, path)
}
```

- [x] **Step 4: Run the tests to see them pass**

Run: `scripts/test ./internal/config`
Expected: PASS, the older config tests too. Then `go vet ./internal/config && gofmt -l internal/config` prints nothing.

- [x] **Step 5: Commit**

```bash
git add internal/config/user.go internal/config/repo_user.go internal/config/repo_user_test.go
git commit -m "Config gains plan_depth and a repo layer in .acta.yaml"
```

### Task 2: acta config show and set learn the repo layer

**Files:**
- Modify: `internal/cli/config_cmd.go`
- Test: `internal/cli/config_repo_test.go`

**verify:** `acta config show` prints the merged value of every repo key, with ` (repo)` after each value `.acta.yaml` set and nothing after the others, always prints a `plan_depth:` line (`full` when unset), and its `--json` holds `plan_depth` and `from_repo`. `acta config set --repo` writes only `.acta.yaml`, leaves the global file byte for byte as it was, accepts only `--repo-language`, `--executor`, `--plan-depth`, refuses every other flag with the rule named, refuses a bad value before writing anything, and refuses when the repo has a `.pm.yaml` but no `.acta.yaml`. `acta config set --plan-depth` without `--repo` writes the global file. A broken `.acta.yaml` makes `show` fail with the path named. List the cases checked.

**Interfaces:**
- Consumes: `config.MergeRepo`, `config.SaveRepoUser`, `config.RepoKeys`, `User.PlanDepth` (Task 1); `config.Load(cwd, "")` for `RepoRoot`; test helper `mustRun` in `cli_test.go`.
- Produces: nothing new for later tasks.

- [ ] **Step 1: Write the failing tests**

Create `internal/cli/config_repo_test.go`:

```go
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoWithGlobal makes a temp folder to run in and a temp global config, so
// no test here ever reads or writes the real ~/.acta/config.yaml.
func repoWithGlobal(t *testing.T, global, repo string) (dir, globalPath string) {
	t.Helper()
	dir = t.TempDir()
	globalPath = filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("PM_VOICE_FILE", globalPath)
	if global != "" {
		if err := os.WriteFile(globalPath, []byte(global), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if repo != "" {
		if err := os.WriteFile(filepath.Join(dir, ".acta.yaml"), []byte(repo), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)
	return dir, globalPath
}

func runCode(args ...string) (int, string, string) {
	var out, errs strings.Builder
	code := Run(args, strings.NewReader(""), false, &out, &errs)
	return code, out.String(), errs.String()
}

// show labels only the values the repo set, and always names plan_depth.
func TestConfigShowRepoLayer(t *testing.T) {
	repoWithGlobal(t, "chat_language: Indonesian\nstyle: adhd\nrepo_language: English\nbuild_executor: dispatch\n", "plan_depth: minimal\n")
	out := mustRun(t, "config", "show")
	for _, want := range []string{"plan_depth: minimal (repo)\n", "build_executor: dispatch\n", "repo_language: English\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("show lacks %q:\n%s", want, out)
		}
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(mustRun(t, "config", "show", "--json")), &got); err != nil {
		t.Fatal(err)
	}
	if got["plan_depth"] != "minimal" {
		t.Errorf("json plan_depth %v", got["plan_depth"])
	}
	if from, _ := got["from_repo"].([]any); len(from) != 1 || from[0] != "plan_depth" {
		t.Errorf("json from_repo %v", got["from_repo"])
	}
}

// With nothing set anywhere, plan_depth still shows, as full.
func TestConfigShowPlanDepthDefault(t *testing.T) {
	repoWithGlobal(t, "", "")
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "plan_depth: full\n") {
		t.Errorf("show lacks plan_depth: full:\n%s", out)
	}
}

// set --repo writes .acta.yaml and never the global file.
func TestConfigSetRepoWritesRepoFile(t *testing.T) {
	dir, globalPath := repoWithGlobal(t, "chat_language: Indonesian\nstyle: adhd\nrepo_language: English\n", "")
	before, _ := os.ReadFile(globalPath)
	mustRun(t, "config", "set", "--repo", "--plan-depth", "minimal", "--executor", "inline", "--repo-language", "Korean")
	after, _ := os.ReadFile(globalPath)
	if string(before) != string(after) {
		t.Errorf("global file changed:\n%s", after)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, ".acta.yaml"))
	for _, want := range []string{"plan_depth: minimal", "build_executor: inline", "repo_language: Korean"} {
		if !strings.Contains(string(raw), want) {
			t.Errorf(".acta.yaml lacks %q:\n%s", want, raw)
		}
	}
	if out := mustRun(t, "config", "show"); !strings.Contains(out, "plan_depth: minimal (repo)") {
		t.Errorf("show:\n%s", out)
	}
}

// Personal flags, bad values and an old .pm.yaml are refused before any write.
func TestConfigSetRepoRefuses(t *testing.T) {
	for _, args := range [][]string{
		{"config", "set", "--repo", "--language", "Korean"},
		{"config", "set", "--repo", "--style", "plain"},
		{"config", "set", "--repo", "--tone", "short"},
		{"config", "set", "--repo", "--subagent-models", "split"},
		{"config", "set", "--repo", "--theme", "x"},
		{"config", "set", "--repo", "--plan-depth", "deep"},
		{"config", "set", "--repo"},
	} {
		dir, _ := repoWithGlobal(t, "", "")
		if code, _, errs := runCode(args...); code != exitBadInput || errs == "" {
			t.Errorf("%v: exit %d stderr %q, want bad input", args, code, errs)
		}
		if _, err := os.Stat(filepath.Join(dir, ".acta.yaml")); err == nil {
			t.Errorf("%v wrote .acta.yaml", args)
		}
	}
	dir, _ := repoWithGlobal(t, "", "")
	if err := os.WriteFile(filepath.Join(dir, ".pm.yaml"), []byte("root: .pm\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, errs := runCode("config", "set", "--repo", "--plan-depth", "minimal"); code != exitBadInput || !strings.Contains(errs, ".pm.yaml") {
		t.Errorf(".pm.yaml repo: exit %d stderr %q", code, errs)
	}
}

// Without --repo the depth goes to the global file.
func TestConfigSetPlanDepthGlobal(t *testing.T) {
	dir, globalPath := repoWithGlobal(t, "", "")
	mustRun(t, "config", "set", "--plan-depth", "minimal")
	raw, _ := os.ReadFile(globalPath)
	if !strings.Contains(string(raw), "plan_depth: minimal") {
		t.Errorf("global file:\n%s", raw)
	}
	if _, err := os.Stat(filepath.Join(dir, ".acta.yaml")); err == nil {
		t.Error("set without --repo wrote .acta.yaml")
	}
}

// A broken repo file stops show and names the file.
func TestConfigShowBrokenRepoFile(t *testing.T) {
	repoWithGlobal(t, "", "chat_language: Korean\n")
	if code, _, errs := runCode("config", "show"); code != exitBadInput || !strings.Contains(errs, ".acta.yaml") {
		t.Errorf("exit %d stderr %q", code, errs)
	}
}
```

- [ ] **Step 2: Run the tests to see them fail**

Run: `scripts/test ./internal/cli -run 'ConfigShowRepo|ConfigShowPlanDepth|ConfigSetRepo|ConfigSetPlanDepth|ConfigShowBrokenRepo'`
Expected: FAIL: `--plan-depth` and `--repo` are unknown flags, and show has no `plan_depth` line.

- [ ] **Step 3: Write the minimal code**

In `internal/cli/config_cmd.go`:

Replace `configUsage` with:

```go
const configUsage = "usage: acta config show [--json] | acta config set [--language L] [--style adhd|plain] [--tone T] [--clear-tone] [--repo-language L] [--executor subagent|dispatch|inline] [--plan-depth minimal|full] [--subagent-models split|default] [--clear-subagent-models] [--theme NAME] [--clear-theme] | acta config set --repo [--repo-language L] [--executor E] [--plan-depth D]"
```

Add `"os"` and `"path/filepath"` to the imports.

In `show`, right after the `ResolveUserFile` error check, merge the repo layer:

```go
		cwd, _ := os.Getwd()
		cfg, err := config.Load(cwd, "")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitBadInput
		}
		v, from, err := config.MergeRepo(v, cfg.RepoRoot)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitBadInput
		}
		depth := v.PlanDepth
		if depth == "" {
			depth = "full"
		}
		// Name the values the repo set, so the user knows which file to edit.
		mark := func(k string) string {
			if from[k] {
				return " (repo)"
			}
			return ""
		}
		var fromRepo []string
		for _, k := range config.RepoKeys {
			if from[k] {
				fromRepo = append(fromRepo, k)
			}
		}
```

In the JSON map add `"plan_depth": depth, "from_repo": fromRepo,`. Change the text output to:

```go
		fmt.Fprintf(stdout, "file: %s (%s)\nchat_language: %s\nstyle: %s\nrepo_language: %s%s\n",
			read, state, v.ChatLanguage, v.Style, v.RepoLanguage, mark("repo_language"))
		if v.Tone != "" {
			fmt.Fprintf(stdout, "tone: %s\n", v.Tone)
		}
		if v.BuildExecutor != "" {
			fmt.Fprintf(stdout, "build_executor: %s%s\n", v.BuildExecutor, mark("build_executor"))
		}
		fmt.Fprintf(stdout, "plan_depth: %s%s\n", depth, mark("plan_depth"))
```

(the `subagent_models` and `theme` lines stay as they are).

In `set`, add the flags next to the others:

```go
		depth := fs.String("plan-depth", "", "how much a plan spells out: minimal or full")
		repoOnly := fs.Bool("repo", false, "save to .acta.yaml in this repo instead of your own config")
```

Add `*depth == ""` to the "nothing to set" check. Right after that check, before `ResolveUserFile`, handle `--repo`:

```go
		if *repoOnly {
			return setRepo(stdout, stderr, map[string]string{"repo_language": *repo, "build_executor": *executor, "plan_depth": *depth},
				*lang != "" || *style != "" || *tone != "" || *models != "" || *themeName != "" || *clearTone || *clearModels || *clearTheme)
		}
```

After the `*executor` block add:

```go
		if *depth != "" {
			v.PlanDepth = *depth
		}
```

Add this function at the end of the file:

```go
// setRepo saves repo keys to .acta.yaml. Personal flags are refused, since
// that file is committed and would change how the agent talks to everyone
// who clones the repo.
func setRepo(stdout, stderr io.Writer, want map[string]string, personal bool) int {
	if personal {
		fmt.Fprintf(stderr, "--repo takes only --repo-language, --executor and --plan-depth; set the others without --repo\n")
		return exitBadInput
	}
	set := map[string]string{}
	for k, val := range want {
		if val != "" {
			set[k] = val
		}
	}
	if len(set) == 0 {
		fmt.Fprintln(stderr, configUsage)
		return exitBadInput
	}
	cwd, _ := os.Getwd()
	cfg, err := config.Load(cwd, "")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	// With only .pm.yaml, a new .acta.yaml would win and hide every other
	// setting in the old file.
	_, errActa := os.Stat(filepath.Join(cfg.RepoRoot, ".acta.yaml"))
	if _, errPM := os.Stat(filepath.Join(cfg.RepoRoot, ".pm.yaml")); errPM == nil && errActa != nil {
		fmt.Fprintf(stderr, "this repo still uses .pm.yaml; rename it to .acta.yaml first\n")
		return exitBadInput
	}
	// Check the values before the write, so a typo never lands in the file.
	check := config.UserDefault()
	for _, k := range config.RepoKeys {
		switch val := set[k]; {
		case val == "":
		case k == "repo_language":
			check.RepoLanguage = val
		case k == "build_executor":
			check.BuildExecutor = val
		case k == "plan_depth":
			check.PlanDepth = val
		}
	}
	if err := check.Validate(); err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	path, err := config.SaveRepoUser(cfg.RepoRoot, set)
	if err != nil {
		fmt.Fprintln(stderr, err)
		if errors.Is(err, config.ErrBadUser) {
			return exitBadInput
		}
		return exitOther
	}
	fmt.Fprintln(stdout, path)
	return exitOK
}
```

- [ ] **Step 4: Run the tests to see them pass**

Run: `scripts/test ./internal/cli -run Config`
Expected: PASS, the older config tests too. Then `go vet ./internal/cli && gofmt -l internal/cli` prints nothing.

- [ ] **Step 5: Commit**

```bash
git add internal/cli/config_cmd.go internal/cli/config_repo_test.go
git commit -m "acta config show and set read and write the repo layer"
```

### Task 3: The session hook reads the repo layer, and rule 2 knows minimal plans

**Files:**
- Modify: `internal/cli/hook.go` (`loadVoice`)
- Modify: `internal/hook/hook.go` (`Input.RepoErr`, `coreRules`, `SessionStart`)
- Modify: `plugin/hooks/default-rules.md` (regenerated)
- Test: `internal/hook/hook_test.go`, `internal/cli/hook_repo_test.go` (new; `config_repo_test.go` belongs to Task 2, which runs in the same wave)

**verify:** Every hook path that prints the voice (`session-start` and `prompt`) uses the repo's `repo_language` when `.acta.yaml` sets it and the global one when it does not. A bad `.acta.yaml` never changes the chat language or style (they are personal and come from the global file), never crashes the hook and never exits non-zero; `session-start` prints one line naming `.acta.yaml` and the error, and uses the global `repo_language` until it is fixed. Rule 2 in the hook output and in `plugin/hooks/default-rules.md` is exactly `2. No code before an approved design and an approved plan. A plan with depth: minimal in its frontmatter needs no plan yes.` List the paths checked.

**Interfaces:**
- Consumes: `config.MergeRepo` (Task 1), `config.Load`, `hook.Input`.
- Produces: `hook.Input.RepoErr error`.

- [ ] **Step 1: Write the failing tests**

Create `internal/cli/hook_repo_test.go`:

```go
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The repo language in .acta.yaml reaches the session rules, so an agent in
// that repo writes files in the repo's language.
func TestHookUsesRepoLanguage(t *testing.T) {
	dir := t.TempDir()
	global := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("PM_VOICE_FILE", global)
	if err := os.WriteFile(global, []byte("chat_language: Indonesian\nstyle: adhd\nrepo_language: English\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".acta.yaml"), []byte("repo_language: Korean\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	out := mustRun(t, "hook", "session-start")
	if !strings.Contains(out, "(code, comments, commits, specs, plans) in Korean") {
		t.Errorf("session-start ignores the repo language:\n%s", out)
	}
	// A broken repo file is named, and the chat language the user picked
	// stays, since a repo file must never change how the agent talks.
	if err := os.WriteFile(filepath.Join(dir, ".acta.yaml"), []byte("chat_language: Korean\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out = mustRun(t, "hook", "session-start")
	if !strings.Contains(out, ".acta.yaml") || !strings.Contains(out, "in Indonesian") || !strings.Contains(out, "plans) in English") {
		t.Errorf("session-start with a bad .acta.yaml:\n%s", out)
	}
	if out := mustRun(t, "hook", "prompt"); !strings.Contains(out, "reply in Indonesian") {
		t.Errorf("prompt with a bad .acta.yaml: %q", out)
	}
}
```

In `internal/hook/hook_test.go`, change the expected rule 2 line (near line 29) to:

```go
		"No code before an approved design and an approved plan. A plan with depth: minimal in its frontmatter needs no plan yes.",
```

- [ ] **Step 2: Run the tests to see them fail**

Run: `scripts/test ./internal/cli -run TestHookUsesRepoLanguage` and `scripts/test ./internal/hook`
Expected: FAIL: the hook prints `in English`, and rule 2 lacks the minimal sentence.

- [ ] **Step 3: Write the minimal code**

In `internal/hook/hook.go`, add a field to `Input` under `VoiceErr`:

```go
	RepoErr     error    // .acta.yaml could not be read; its repo keys are left out
```

In `SessionStart`, at the end of the `default:` case of the voice switch (after the tone lines), add:

```go
		if in.RepoErr != nil {
			fmt.Fprintf(&b, "- The repo settings could not be read (%v). The ones above come from your own config until .acta.yaml is fixed.\n", in.RepoErr)
		}
```

Change rule 2 in `coreRules` to:

```
2. No code before an approved design and an approved plan. A plan with depth: minimal in its frontmatter needs no plan yes.
```

In `internal/cli/hook.go`, replace `loadVoice`:

```go
// loadVoice reads the user config and lays this repo's .acta.yaml over it.
// A bad repo file keeps the user's own values and is named in RepoErr, so
// a repo can never switch the chat language, and the session still sees
// what to fix.
func loadVoice() hook.Input {
	v, exists, err := config.ResolveUser()
	in := hook.Input{Voice: v, VoiceExists: exists, VoiceErr: err}
	if err != nil {
		return in
	}
	cwd, _ := os.Getwd()
	if cfg, lerr := config.Load(cwd, ""); lerr == nil {
		in.Voice, _, in.RepoErr = config.MergeRepo(v, cfg.RepoRoot)
	}
	return in
}
```

`MergeRepo` returns `v` unchanged on error, so `in.Voice` keeps the global values then. `Prompt` needs no change.

Regenerate the rules file: `go test ./internal/hook -run TestDefaultRulesFile -update`.

- [ ] **Step 4: Run the tests to see them pass**

Run: `scripts/test ./internal/hook ./internal/cli -run 'Hook|DefaultRules'`
Expected: PASS. Then `git diff plugin/hooks/default-rules.md` shows only the rule 2 line, and `go vet ./internal/hook ./internal/cli && gofmt -l internal` prints nothing.

- [ ] **Step 5: Commit**

```bash
git add internal/cli/hook.go internal/cli/hook_repo_test.go internal/hook/hook.go internal/hook/hook_test.go plugin/hooks/default-rules.md
git commit -m "Session hook reads the repo layer; rule 2 lets a minimal plan skip its yes"
```

### Task 4: Plan, build and setup skills learn plan depth

**Files:**
- Modify: `plugin/skills/plan/SKILL.md`, `plugin/skills/build/SKILL.md`, `plugin/skills/setup/SKILL.md`
- Test: `internal/plugincheck/skill_plan_test.go`, `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/skill_setup_test.go`

**verify:** The plan skill tells the agent, on every path, how to pick the depth (argument, then `acta config show`, then `full`), what a minimal plan holds and leaves out, that the code-block and No Placeholders rules are for `full` only, and that a minimal plan is committed and handed to `acta:build` in the same turn with no yes. The build skill never refuses a plan with `depth: minimal` for lacking a yes, and still refuses when the spec is not approved. The setup skill asks for the depth and whether to save for every repo or this repo only, with the exact commands. plugincheck fails if any of these sentences goes missing. List the sentences guarded.

**Interfaces:**
- Consumes: the flags from Task 2 (`--plan-depth minimal|full`, `--repo`), as text only.
- Produces: nothing.

- [x] **Step 1: Write the failing tests**

In `internal/plugincheck/skill_plan_test.go`, add to `Must`:

```go
			"## Plan depth", "`/plan minimal`", "depth: minimal", "`plan_depth`",
			"apply to `full` plans only",
			"invoke `acta:build` in the same turn",
```

In `internal/plugincheck/skill_build_test.go`, add to `Must`:

```go
			"or has `depth: minimal` in its frontmatter",
```

and add to its `MustNot` (create the list if the rule has none):

```go
			"Refuse to start without an approved spec and an approved plan. Say which one is missing.",
```

In `internal/plugincheck/skill_setup_test.go`, add to `Must`:

```go
			"--plan-depth minimal", "--plan-depth full", "acta config set --repo",
			"every repo or this repo only",
```

and change `MaxLines: 66` to `MaxLines: 76`.

- [x] **Step 2: Run the tests to see them fail**

Run: `scripts/test ./internal/plugincheck -run 'TestSkillPlan|TestSkillBuild|TestSkillSetup'`
Expected: FAIL naming each missing sentence.

- [x] **Step 3: Write the skill text**

`plugin/skills/plan/SKILL.md`: add this section right after `## Overview`:

```markdown
## Plan depth

Pick the depth in this order, and do not ask when one answers: the argument of `/plan minimal` or `/plan full` (this run only, never saved), then `plan_depth` from `acta config show`, then `full`.

- `full`: everything below, with real code in every code step, and a wait for the user's yes.
- `minimal`: the frontmatter holds `parent:` and `depth: minimal`. The header is `**Goal:**` (one sentence), `**Spec:**`, `**Tests:**`, `## Global Constraints` (the ponytail-lazy line plus only what this plan needs) and `## Waves`; no Architecture, Tech Stack, File map or Interfaces. Each task has a title, `**Files:**`, a property-shaped `**verify:**` and three one-sentence boxes: the failing test and why it fails, the code change, the commit message. No code blocks. About 6 lines a task.

The code-block rule and No Placeholders apply to `full` plans only. A minimal plan's self-review checks two things: every part of the spec has a task, and every verify line is a property.

A minimal plan needs no yes: save it, run `acta id`, commit it on main, tell the user its path in one line, and invoke `acta:build` in the same turn. The spec still needs its yes first.
```

In `## Hand-off`, change the first line to: `Save the plan, show it to the user, and wait for a yes (a minimal plan skips the wait, see Plan depth). Approval of the design did not approve the plan.`

`plugin/skills/build/SKILL.md`: replace the line `Refuse to start without an approved spec and an approved plan. Say which one is missing.` with:

```markdown
Refuse to start unless the spec is approved (for Bounded work with no spec file, approved in chat) and the plan is approved or has `depth: minimal` in its frontmatter. Say which one is missing.
```

`plugin/skills/setup/SKILL.md`: after the `### Default build executor` block, add:

````markdown
### Plan depth

`full` (default: real code in every plan step, and the plan waits for a yes) or `minimal` (short steps, no code, build starts right away).

```bash
acta config set --plan-depth full
```

Then ask once whether to save the executor and the depth for every repo or this repo only. This repo only: `acta config set --repo --executor inline --plan-depth minimal` writes `.acta.yaml`, which is committed, so it reaches everyone who clones the repo. Use `--plan-depth minimal` the same way in the global command.
````

Keep the setup skill at 76 lines or fewer; shorten the new text, not the old, if it runs over.

- [x] **Step 4: Run the tests to see them pass**

Run: `scripts/test ./internal/plugincheck`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
git add plugin/skills/plan/SKILL.md plugin/skills/build/SKILL.md plugin/skills/setup/SKILL.md internal/plugincheck/skill_plan_test.go internal/plugincheck/skill_build_test.go internal/plugincheck/skill_setup_test.go
git commit -m "Plan, build and setup skills learn plan depth"
```
