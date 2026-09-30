---
created: "2026-09-30"
parent: specs/2026-09-30-omp-harness-design
id: PLN-0056
hash: uavj6ws
---
# omp Eval Runner Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `acta eval-omp` (and `scripts/eval --omp`) runs the same `plugin/evals/*` cases through omp and prints PASS, FAIL or SKIP per case.

**Architecture:** A new package `internal/evalomp` loads the cases, runs each one with `omp -p --mode json` in a throwaway folder, parses the JSON event stream into the final reply plus the tool calls, and grades the result with the four grader types the suite uses. `internal/cli` gets one thin `eval-omp` command. Tests use a fake `omp` script, so `scripts/test` costs no quota.

**Tech Stack:** Go stdlib, `gopkg.in/yaml.v3` (already in `go.mod`).

**Spec:** `.acta/specs/2026-09-30-omp-harness-design.md` (the "Plan B" section)

**Tests:** fast `scripts/test ./internal/evalomp ./internal/cli ./internal/plugincheck`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- omp runs never pin a model: no `--model` on the case run or on the judge.
- The case run command is exactly: `omp -p --mode json --no-session --no-extensions --no-rules --skills=<acta skill folder names, comma separated> --plugin-dir <abs plugin dir> -e <abs plugin dir>/omp/index.ts <prompt>`.
- The run's `PM_VOICE_FILE` is `<throwaway home>/.acta/config.yaml`; the scaffold runs with `HOME=<throwaway home>`. The omp run keeps the user's real `HOME`, because omp needs its auth.
- A case whose `tags` holds `claude-only` is skipped. No new frontmatter field: `claude plugin eval` rejects unknown keys.
- An unsupported grader type, target or match value is a FAIL that names it, never a silent pass.
- `max_turns` and `allowed_tools` are ignored; `timeout_seconds` (default 300) caps a run. The `llm` judge votes once.
- Exit codes: 0 all cases passed or skipped, 1 a case failed, 3 no omp on PATH or the suite cannot load, 1 bad usage (acta's `exitBadInput`).
- Comments are plain English a 10-year-old reads back without stopping. They say why, not what. No absolute user paths in any file under `plugin/` (`internal/plugincheck` rejects them).
- Run `gofmt -l .` and `go vet` on the touched packages before each commit, inside the task commit.

## File Map

| File | Task | Change |
|---|---|---|
| `internal/evalomp/case.go` | 1 | new: `Case`, `Grader`, `Target`, `LoadCases`, `splitFrontmatter` |
| `internal/evalomp/case_test.go` | 1, 5 | new: loader tests; task 5 adds the real-suite test |
| `internal/evalomp/result.go` | 2 | new: `Call`, `Result`, `Workspace`, `ParseStream`, `ListFiles` |
| `internal/evalomp/result_test.go` | 2 | new |
| `internal/evalomp/testdata/stream.jsonl` | 2 | new: a trimmed omp event stream |
| `internal/evalomp/grade.go` | 3 | new: `Outcome`, `Judge`, `Grade`, `Unsupported` |
| `internal/evalomp/grade_test.go` | 3 | new |
| `internal/evalomp/run.go` | 4 | new: `Options`, `ErrTimeout`, `SkillNames`, `RunCase`, `OmpJudge`, `RunAll` |
| `internal/evalomp/run_test.go` | 4 | new: fake `omp` script tests |
| `internal/cli/eval_omp.go` | 5 | new: `cmdEvalOmp` |
| `internal/cli/eval_omp_test.go` | 5 | new |
| `internal/cli/cli.go` | 5 | `case "eval-omp"` and the unknown-command list |
| `scripts/eval` | 5 | `--omp` passes through to `go run ./cmd/acta eval-omp` |
| `plugin/evals/second-brainstorm-choices/prompt.md` | 5 | `tags: [brainstorm, claude-only]` |
| `plugin/evals/FACTS.md` | 5 | new section: running the suite in omp, and the four known gaps |
| `plugin/omp/FACTS.md` | 5 | new section: the probe facts behind the run command |

## Waves

- Wave 1: Task 1, Task 2 (no shared files).
- Wave 2: Task 3, Task 4 (both use Task 1 and Task 2 types; no shared files).
- Wave 3: Task 5.

---

### Task 1: Load eval cases

**Files:**
- Create: `internal/evalomp/case.go`
- Test: `internal/evalomp/case_test.go`

**verify:** Every case folder with a `prompt.md` loads with its prompt body, tags, timeout, scaffold path and every grader file, and nothing else in the eval folder is taken for a case. List each input shape handled (no `case.yaml`, `case.yaml` with a scaffold, grader with scalar target, grader with `{source, path}` target, grader with a body, frontmatter-only grader, missing timeout, folder with no `prompt.md`, broken YAML) and what each gives.

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `type Target struct { Kind string; Path string }` (`Kind` is `""`, `"last_message"`, `"file"`, or whatever else the file says)
  - `type Grader struct { Name, Type, Path string; Exists *bool; Pattern, Flags, Match string; Target Target; Tool, InputMatch string; Min, Max *int; Body string }`
  - `type Case struct { Name, Dir, Prompt string; Tags []string; TimeoutSeconds int; Scaffold string; Graders []Grader }`
  - `func (c Case) ClaudeOnly() bool`
  - `func LoadCases(evalDir string) ([]Case, error)`

- [ ] **Step 1: Write the failing test**

Create `internal/evalomp/case_test.go`:

```go
package evalomp

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadCases(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "alpha/prompt.md"), "---\ntags: [scratch]\n# a comment\ntimeout_seconds: 60\n---\n\nDo the thing.\n")
	write(t, filepath.Join(dir, "alpha/case.yaml"), "schema_version: \"1.1\"\nname: alpha\ncontext:\n  scaffold_script: scaffold.sh\n")
	write(t, filepath.Join(dir, "alpha/graders/file.md"), "---\ntype: file_exists\npath: \".acta/scratch/*.md\"\nexists: true\n---\n")
	write(t, filepath.Join(dir, "alpha/graders/body.md"), "---\n# guards: x\ntype: regex\npattern: \"only here\"\ntarget:\n  source: file\n  path: \"a.md\"\n---\n")
	write(t, filepath.Join(dir, "beta/prompt.md"), "---\ntags: [brainstorm, claude-only]\n---\nSecond.")
	write(t, filepath.Join(dir, "beta/graders/judge.md"), "---\ntype: llm\n---\n\nPASS only when polite.\n")
	write(t, filepath.Join(dir, "beta/graders/tool.md"), "---\ntype: tool_used\ntool: Bash\ninput_match: \"acta scratch new\"\nmin: 2\n---")
	write(t, filepath.Join(dir, "results/report.html"), "not a case")

	cases, err := LoadCases(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) != 2 || cases[0].Name != "alpha" || cases[1].Name != "beta" {
		t.Fatalf("cases = %+v", cases)
	}
	a, b := cases[0], cases[1]
	if a.Prompt != "Do the thing." || a.TimeoutSeconds != 60 || a.ClaudeOnly() {
		t.Errorf("alpha = %+v", a)
	}
	if a.Scaffold != filepath.Join(dir, "alpha", "scaffold.sh") {
		t.Errorf("scaffold = %q", a.Scaffold)
	}
	if len(a.Graders) != 2 || a.Graders[0].Name != "body" || a.Graders[1].Name != "file" {
		t.Fatalf("alpha graders = %+v", a.Graders)
	}
	if g := a.Graders[0]; g.Target != (Target{Kind: "file", Path: "a.md"}) || g.Pattern != "only here" {
		t.Errorf("body grader = %+v", g)
	}
	if g := a.Graders[1]; g.Type != "file_exists" || g.Exists == nil || !*g.Exists || g.Path != ".acta/scratch/*.md" {
		t.Errorf("file grader = %+v", g)
	}
	if !b.ClaudeOnly() || b.TimeoutSeconds != 300 || b.Scaffold != "" || b.Prompt != "Second." {
		t.Errorf("beta = %+v", b)
	}
	if g := b.Graders[0]; g.Type != "llm" || g.Body != "PASS only when polite." {
		t.Errorf("judge grader = %+v", g)
	}
	if g := b.Graders[1]; g.Tool != "Bash" || g.InputMatch != "acta scratch new" || g.Min == nil || *g.Min != 2 || g.Max != nil {
		t.Errorf("tool grader = %+v", g)
	}
}

func TestLoadCasesScalarTarget(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "c/prompt.md"), "---\n---\nHi")
	write(t, filepath.Join(dir, "c/graders/r.md"), "---\ntype: regex\npattern: NONE\nflags: i\ntarget: last_message\n---\n")
	cases, err := LoadCases(dir)
	if err != nil {
		t.Fatal(err)
	}
	if g := cases[0].Graders[0]; g.Target.Kind != "last_message" || g.Flags != "i" {
		t.Errorf("grader = %+v", g)
	}
}

func TestLoadCasesErrors(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"no frontmatter": {"c/prompt.md": "Hi"},
		"bad yaml":       {"c/prompt.md": "---\ntags: [\n---\nHi"},
		"bad grader":     {"c/prompt.md": "---\n---\nHi", "c/graders/g.md": "no frontmatter"},
		"bad case.yaml":  {"c/prompt.md": "---\n---\nHi", "c/case.yaml": "context: ["},
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			for p, text := range files {
				write(t, filepath.Join(dir, p), text)
			}
			if _, err := LoadCases(dir); err == nil {
				t.Error("want an error")
			}
		})
	}
	if _, err := LoadCases(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("a missing eval folder must be an error")
	}
}
```

- [ ] **Step 2: Run the test to see it fail**

Run: `scripts/test ./internal/evalomp -run TestLoadCases`
Expected: FAIL, the package does not build (`LoadCases` undefined).

- [ ] **Step 3: Write the implementation**

Create `internal/evalomp/case.go`:

```go
// Package evalomp runs the plugin's eval cases through omp, so the plugin is
// checked in omp the same way claude plugin eval checks it in Claude Code.
// The cases are the same files; only the runner differs.
package evalomp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// defaultTimeout is claude plugin eval's own default, so a case with no
// timeout gets the same cap in both harnesses.
const defaultTimeout = 300

// Target says what a grader looks at: the final reply, or one file.
type Target struct {
	Kind string
	Path string
}

// UnmarshalYAML takes both forms a grader file uses: a plain word like
// last_message, or a map with source and path.
func (t *Target) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		t.Kind = n.Value
		return nil
	}
	var m struct {
		Source string `yaml:"source"`
		Path   string `yaml:"path"`
	}
	if err := n.Decode(&m); err != nil {
		return err
	}
	t.Kind, t.Path = m.Source, m.Path
	return nil
}

// Grader is one file under graders/. Name is the file name without .md, and
// Body is the text after the frontmatter, which is the rubric of an llm grader.
type Grader struct {
	Name       string `yaml:"-"`
	Type       string `yaml:"type"`
	Path       string `yaml:"path"`
	Exists     *bool  `yaml:"exists"`
	Pattern    string `yaml:"pattern"`
	Flags      string `yaml:"flags"`
	Match      string `yaml:"match"`
	Target     Target `yaml:"target"`
	Tool       string `yaml:"tool"`
	InputMatch string `yaml:"input_match"`
	Min        *int   `yaml:"min"`
	Max        *int   `yaml:"max"`
	Body       string `yaml:"-"`
}

// Case is one folder under the eval folder.
type Case struct {
	Name           string
	Dir            string
	Prompt         string
	Tags           []string
	TimeoutSeconds int
	Scaffold       string
	Graders        []Grader
}

// ClaudeOnly says the case only makes sense in Claude Code. A tag and not a
// field, because claude plugin eval refuses frontmatter keys it does not know.
func (c Case) ClaudeOnly() bool {
	return slices.Contains(c.Tags, "claude-only")
}

// LoadCases reads every case under evalDir, in name order. A folder with no
// prompt.md is not a case, so a results folder next to the cases is skipped.
func LoadCases(evalDir string) ([]Case, error) {
	entries, err := os.ReadDir(evalDir)
	if err != nil {
		return nil, err
	}
	var cases []Case
	for _, e := range entries {
		dir := filepath.Join(evalDir, e.Name())
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, "prompt.md")); err != nil {
			continue
		}
		c, err := loadCase(dir)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		cases = append(cases, c)
	}
	return cases, nil
}

func loadCase(dir string) (Case, error) {
	src, err := os.ReadFile(filepath.Join(dir, "prompt.md"))
	if err != nil {
		return Case{}, err
	}
	fm, body, ok := splitFrontmatter(string(src))
	if !ok {
		return Case{}, errors.New("prompt.md has no frontmatter")
	}
	var meta struct {
		Tags    []string `yaml:"tags"`
		Timeout int      `yaml:"timeout_seconds"`
	}
	if err := yaml.Unmarshal([]byte(fm), &meta); err != nil {
		return Case{}, fmt.Errorf("prompt.md: %w", err)
	}
	c := Case{
		Name:           filepath.Base(dir),
		Dir:            dir,
		Prompt:         strings.TrimSpace(body),
		Tags:           meta.Tags,
		TimeoutSeconds: meta.Timeout,
	}
	if c.TimeoutSeconds <= 0 {
		c.TimeoutSeconds = defaultTimeout
	}
	if y, err := os.ReadFile(filepath.Join(dir, "case.yaml")); err == nil {
		var cy struct {
			Context struct {
				Scaffold string `yaml:"scaffold_script"`
			} `yaml:"context"`
		}
		if err := yaml.Unmarshal(y, &cy); err != nil {
			return Case{}, fmt.Errorf("case.yaml: %w", err)
		}
		if cy.Context.Scaffold != "" {
			c.Scaffold = filepath.Join(dir, cy.Context.Scaffold)
		}
	}
	files, err := filepath.Glob(filepath.Join(dir, "graders", "*.md"))
	if err != nil {
		return Case{}, err
	}
	for _, f := range files {
		g, err := loadGrader(f)
		if err != nil {
			return Case{}, fmt.Errorf("%s: %w", filepath.Base(f), err)
		}
		c.Graders = append(c.Graders, g)
	}
	return c, nil
}

func loadGrader(path string) (Grader, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return Grader{}, err
	}
	fm, body, ok := splitFrontmatter(string(src))
	if !ok {
		return Grader{}, errors.New("no frontmatter")
	}
	var g Grader
	if err := yaml.Unmarshal([]byte(fm), &g); err != nil {
		return Grader{}, err
	}
	g.Name = strings.TrimSuffix(filepath.Base(path), ".md")
	g.Body = strings.TrimSpace(body)
	return g, nil
}

// splitFrontmatter cuts a file into the YAML between the two --- lines and the
// text after them. A grader file is often only frontmatter, so the closing
// line may be the very end of the file.
func splitFrontmatter(src string) (fm, body string, ok bool) {
	rest, found := strings.CutPrefix(src, "---\n")
	if !found {
		return "", "", false
	}
	if fm, body, found = strings.Cut(rest, "\n---\n"); found {
		return fm, body, true
	}
	if rest == "---" || strings.HasPrefix(rest, "---\n") {
		// Empty frontmatter: the closing line comes right away.
		return "", strings.TrimPrefix(strings.TrimPrefix(rest, "---"), "\n"), true
	}
	fm, found = strings.CutSuffix(strings.TrimRight(rest, "\n"), "\n---")
	return fm, "", found
}
```

- [ ] **Step 4: Run the test to see it pass**

Run: `scripts/test ./internal/evalomp -run TestLoadCases`
Expected: PASS (`TestLoadCases`, `TestLoadCasesScalarTarget`, `TestLoadCasesErrors`).

- [ ] **Step 5: Commit**

```bash
gofmt -w internal/evalomp && go vet ./internal/evalomp
git add internal/evalomp/case.go internal/evalomp/case_test.go
git commit -m "evalomp: load the plugin's eval cases and graders"
```

---

### Task 2: Parse the omp event stream

**Files:**
- Create: `internal/evalomp/result.go`
- Create: `internal/evalomp/testdata/stream.jsonl`
- Test: `internal/evalomp/result_test.go`

**verify:** The reply is always the text of the last assistant message that has text, and every `tool_execution_start` event becomes exactly one call, in order, whatever else the stream holds. List each stream shape checked (normal run, assistant message with only a tool call at the end, no `agent_end`, a line that is not JSON, custom messages with string content) and what each gives. Also: `ListFiles` lists every file in the folder except under `.git`.

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `type Call struct { Tool string; Input string }` (`Input` is the raw JSON of the call's `args`)
  - `type Result struct { Reply string; Calls []Call }`
  - `type Workspace struct { Dir string; Before map[string]bool; Result Result }` (`Before` holds slash paths relative to `Dir`)
  - `func ParseStream(r io.Reader) (Result, error)`
  - `func ListFiles(dir string) (map[string]bool, error)`

- [ ] **Step 1: Write the fixture and the failing test**

Create `internal/evalomp/testdata/stream.jsonl` (one JSON object per line; the shapes come from a real `omp --mode json` run on 2026-09-30):

```text
{"type":"session","version":3,"id":"s-1","cwd":"/tmp/x"}
{"type":"agent_start"}
{"type":"message_start","message":{"role":"custom","customType":"acta","content":"acta plugin is active."}}
{"type":"tool_execution_start","toolCallId":"c1","toolName":"bash","args":{"command":"acta scratch new idea --title idea"},"intent":"file it"}
{"type":"tool_execution_end","toolCallId":"c1","toolName":"bash","result":{"content":[{"type":"text","text":"SCR-0001"}]},"isError":false}
{"type":"tool_execution_start","toolCallId":"c2","toolName":"read","args":{"path":"skill://scratch"}}
{"type":"agent_end","messages":[{"role":"custom","customType":"acta","content":"acta plugin is active."},{"role":"assistant","content":[{"type":"thinking","thinking":"hmm"},{"type":"text","text":"Filed as SCR-0001."},{"type":"text","text":"Next: brainstorm it."}]},{"role":"toolResult","toolCallId":"c2","content":[{"type":"text","text":"skill text"}]},{"role":"assistant","content":[{"type":"toolCall","id":"c3","name":"todo"}]}]}
```

Create `internal/evalomp/result_test.go`:

```go
package evalomp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseStream(t *testing.T) {
	f, err := os.Open("testdata/stream.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	res, err := ParseStream(f)
	if err != nil {
		t.Fatal(err)
	}
	if res.Reply != "Filed as SCR-0001.\nNext: brainstorm it." {
		t.Errorf("reply = %q", res.Reply)
	}
	if len(res.Calls) != 2 || res.Calls[0].Tool != "bash" || res.Calls[1].Tool != "read" {
		t.Fatalf("calls = %+v", res.Calls)
	}
	if res.Calls[0].Input != `{"command":"acta scratch new idea --title idea"}` {
		t.Errorf("input = %q", res.Calls[0].Input)
	}
}

func TestParseStreamErrors(t *testing.T) {
	for name, text := range map[string]string{
		"no agent_end": `{"type":"agent_start"}` + "\n",
		"not json":     "hello\n",
		"empty":        "",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseStream(strings.NewReader(text)); err == nil {
				t.Error("want an error")
			}
		})
	}
}

func TestListFiles(t *testing.T) {
	dir := t.TempDir()
	// A local helper, so this task does not wait for case_test.go's write.
	put := func(rel string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	put("a.txt")
	put(".acta/scratch/one.md")
	put(".git/HEAD")
	got, err := ListFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || !got["a.txt"] || !got[".acta/scratch/one.md"] {
		t.Errorf("files = %v", got)
	}
}
```

- [ ] **Step 2: Run the test to see it fail**

Run: `scripts/test ./internal/evalomp -run 'TestParseStream|TestListFiles'`
Expected: FAIL, `ParseStream` and `ListFiles` undefined. Task 1 runs in the same wave in the same worktree: if its half-written `case.go` breaks the build, wait and retry; never reset or delete its files.

- [ ] **Step 3: Write the implementation**

Create `internal/evalomp/result.go`:

```go
package evalomp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

// Call is one tool call the agent made. Input is the call's arguments as raw
// JSON, which is what a tool_used grader matches against.
type Call struct {
	Tool  string
	Input string
}

// Result is what one omp run said and did.
type Result struct {
	Reply string
	Calls []Call
}

// Workspace is what one case run left behind. Before lists the files that were
// there before the agent started, so graders can tell what the agent made.
type Workspace struct {
	Dir    string
	Before map[string]bool
	Result Result
}

type streamEvent struct {
	Type     string          `json:"type"`
	ToolName string          `json:"toolName"`
	Args     json.RawMessage `json:"args"`
	Messages []struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"messages"`
}

// ParseStream reads the events omp prints with --mode json. A run that never
// reaches agent_end did not finish, so it is an error, not an empty reply.
func ParseStream(r io.Reader) (Result, error) {
	var res Result
	ended := false
	dec := json.NewDecoder(r)
	for {
		var ev streamEvent
		err := dec.Decode(&ev)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return res, fmt.Errorf("omp json stream: %w", err)
		}
		switch ev.Type {
		case "tool_execution_start":
			res.Calls = append(res.Calls, Call{Tool: ev.ToolName, Input: string(ev.Args)})
		case "agent_end":
			ended = true
			res.Reply = lastReply(ev)
		}
	}
	if !ended {
		return res, errors.New("omp stopped before its agent_end event")
	}
	return res, nil
}

// lastReply is the text of the last assistant message that has any. The very
// last message can be only a tool call, and that is not what the user saw.
func lastReply(ev streamEvent) string {
	for i := len(ev.Messages) - 1; i >= 0; i-- {
		m := ev.Messages[i]
		if m.Role != "assistant" {
			continue
		}
		var parts []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if json.Unmarshal(m.Content, &parts) != nil {
			continue
		}
		var texts []string
		for _, p := range parts {
			if p.Type == "text" {
				texts = append(texts, p.Text)
			}
		}
		if len(texts) > 0 {
			return strings.Join(texts, "\n")
		}
	}
	return ""
}

// ListFiles gives every file under dir as a slash path relative to dir. The
// .git folder is left out: git's own files are never what a grader asks about.
func ListFiles(dir string) (map[string]bool, error) {
	files := map[string]bool{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = true
		return nil
	})
	return files, err
}
```

- [ ] **Step 4: Run the test to see it pass**

Run: `scripts/test ./internal/evalomp -run 'TestParseStream|TestListFiles'`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gofmt -w internal/evalomp && go vet ./internal/evalomp
git add internal/evalomp/result.go internal/evalomp/result_test.go internal/evalomp/testdata/stream.jsonl
git commit -m "evalomp: parse omp's json event stream into the reply and tool calls"
```

---

### Task 3: Grade a run

**Files:**
- Create: `internal/evalomp/grade.go`
- Test: `internal/evalomp/grade_test.go`

**verify:** No grader ever passes by default: every grader type, target and match value the code does not know gives a FAIL that names it, and every known one passes only on the evidence the grader table in `plugin/evals/FACTS.md` names. List every branch of `Grade` (each type, each target, each match value, bad regex, missing file, judge error, judge answer with no verdict) and what each returns.

**Interfaces:**
- Consumes: `Grader`, `Target` (Task 1); `Workspace`, `Result`, `Call`, `ListFiles` (Task 2).
- Produces:
  - `type Outcome struct { Grader string; Pass bool; Why string }`
  - `type Judge func(prompt string) (string, error)`
  - `func Grade(g Grader, w Workspace, judge Judge) Outcome`
  - `func Unsupported(g Grader) string` (empty when this runner can grade `g`, else the reason)

- [ ] **Step 1: Write the failing test**

Create `internal/evalomp/grade_test.go`:

```go
package evalomp

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func yes() *bool {
	b := true
	return &b
}

func no() *bool {
	b := false
	return &b
}

func n(i int) *int { return &i }

func workspace(t *testing.T) Workspace {
	t.Helper()
	dir := t.TempDir()
	write(t, filepath.Join(dir, "old.md"), "from the scaffold")
	before, err := ListFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, ".acta/scratch/new.md"), "only happens in Safari")
	return Workspace{Dir: dir, Before: before, Result: Result{
		Reply: "Filed it. NONE of the rules leaked.",
		Calls: []Call{
			{Tool: "bash", Input: `{"command":"acta scratch new idea"}`},
			{Tool: "bash", Input: `{"command":"acta set scratch/x status brainstorming"}`},
			{Tool: "read", Input: `{"path":"skill://scratch"}`},
		},
	}}
}

func TestGrade(t *testing.T) {
	w := workspace(t)
	judge := func(p string) (string, error) {
		if !strings.Contains(p, "Filed it.") || !strings.Contains(p, "be polite") {
			return "", errors.New("prompt is missing the reply or the rubric")
		}
		return "**PASS** it is polite", nil
	}
	cases := []struct {
		name string
		g    Grader
		pass bool
	}{
		{"file made by the run", Grader{Type: "file_exists", Path: ".acta/scratch/*.md", Exists: yes()}, true},
		{"file from the scaffold does not count", Grader{Type: "file_exists", Path: "old.md", Exists: yes()}, false},
		{"exists defaults to true", Grader{Type: "file_exists", Path: ".acta/scratch/*.md"}, true},
		{"absent file wanted absent", Grader{Type: "file_exists", Path: "AGENTS.md", Exists: no()}, true},
		{"present file wanted absent", Grader{Type: "file_exists", Path: ".acta/scratch/*.md", Exists: no()}, false},
		{"regex on reply", Grader{Type: "regex", Pattern: "NONE"}, true},
		{"regex needs flag i for case", Grader{Type: "regex", Pattern: "none"}, false},
		{"regex flag i", Grader{Type: "regex", Pattern: "none", Flags: "i", Target: Target{Kind: "last_message"}}, true},
		{"regex not_contains", Grader{Type: "regex", Pattern: "Caveman talk", Match: "not_contains"}, true},
		{"regex not_contains found", Grader{Type: "regex", Pattern: "Filed", Match: "not_contains"}, false},
		{"regex on file", Grader{Type: "regex", Pattern: "only happens in Safari", Target: Target{Kind: "file", Path: ".acta/scratch/new.md"}}, true},
		{"regex on missing file", Grader{Type: "regex", Pattern: "x", Target: Target{Kind: "file", Path: "gone.md"}}, false},
		{"regex bad pattern", Grader{Type: "regex", Pattern: "("}, false},
		{"regex unknown target", Grader{Type: "regex", Pattern: "x", Target: Target{Kind: "trace"}}, false},
		{"regex unknown match", Grader{Type: "regex", Pattern: "Filed", Match: "count:1"}, false},
		{"tool used, case ignored", Grader{Type: "tool_used", Tool: "Bash", InputMatch: "acta scratch new"}, true},
		{"tool used, regex input", Grader{Type: "tool_used", Tool: "Bash", InputMatch: "status brainstorming$|status brainstorming\""}, true},
		{"tool not used", Grader{Type: "tool_used", Tool: "Bash", InputMatch: "acta scratch add"}, false},
		{"tool min", Grader{Type: "tool_used", Tool: "bash", Min: n(3)}, false},
		{"tool max 0 means never", Grader{Type: "tool_used", Tool: "bash", InputMatch: "scratch add", Min: n(0), Max: n(0)}, true},
		{"tool max broken", Grader{Type: "tool_used", Tool: "bash", Max: n(1)}, false},
		{"llm pass", Grader{Type: "llm", Body: "be polite"}, true},
		{"unknown type", Grader{Type: "tool_order"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.g.Name = "g"
			got := Grade(c.g, w, judge)
			if got.Pass != c.pass || got.Grader != "g" {
				t.Errorf("Grade = %+v, want pass %v", got, c.pass)
			}
			if !got.Pass && got.Why == "" {
				t.Error("a FAIL must say why")
			}
		})
	}
}

func TestGradeJudgeFailures(t *testing.T) {
	w := workspace(t)
	g := Grader{Name: "j", Type: "llm", Body: "be polite"}
	for name, judge := range map[string]Judge{
		"error":      func(string) (string, error) { return "", errors.New("down") },
		"no verdict": func(string) (string, error) { return "Looks fine to me.", nil },
		"empty":      func(string) (string, error) { return "", nil },
		"fail":       func(string) (string, error) { return "FAIL: rude", nil },
	} {
		t.Run(name, func(t *testing.T) {
			if got := Grade(g, w, judge); got.Pass {
				t.Errorf("Grade = %+v, want FAIL", got)
			}
		})
	}
}

func TestUnsupported(t *testing.T) {
	for _, g := range []Grader{
		{Type: "file_exists"}, {Type: "tool_used"}, {Type: "llm"},
		{Type: "regex"}, {Type: "regex", Target: Target{Kind: "file", Path: "a"}},
		{Type: "regex", Match: "not_contains", Target: Target{Kind: "last_message"}},
	} {
		if why := Unsupported(g); why != "" {
			t.Errorf("Unsupported(%+v) = %q, want empty", g, why)
		}
	}
	for _, g := range []Grader{
		{Type: "baseline"}, {Type: "regex", Target: Target{Kind: "files"}}, {Type: "regex", Match: "count:2"},
	} {
		if Unsupported(g) == "" {
			t.Errorf("Unsupported(%+v) is empty, want a reason", g)
		}
	}
}
```

- [ ] **Step 2: Run the test to see it fail**

Run: `scripts/test ./internal/evalomp -run 'TestGrade|TestUnsupported'`
Expected: FAIL, `Grade`, `Judge` and `Unsupported` undefined.

- [ ] **Step 3: Write the implementation**

Create `internal/evalomp/grade.go`:

```go
package evalomp

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// Outcome is one grader's verdict. Why is set on every FAIL.
type Outcome struct {
	Grader string
	Pass   bool
	Why    string
}

// Judge asks a model to grade a reply and returns its raw answer.
type Judge func(prompt string) (string, error)

// Unsupported gives the reason this runner cannot grade g, or "" when it can.
// It covers only what the suite uses; the rest is named, never passed.
func Unsupported(g Grader) string {
	switch g.Type {
	case "file_exists", "tool_used", "llm":
		return ""
	case "regex":
		switch g.Target.Kind {
		case "", "last_message", "file":
		default:
			return fmt.Sprintf("regex target %q is not supported in omp", g.Target.Kind)
		}
		switch g.Match {
		case "", "contains", "not_contains":
		default:
			return fmt.Sprintf("regex match %q is not supported in omp", g.Match)
		}
		return ""
	}
	return fmt.Sprintf("grader type %q is not supported in omp", g.Type)
}

// Grade runs one grader over one finished run.
func Grade(g Grader, w Workspace, judge Judge) Outcome {
	if why := Unsupported(g); why != "" {
		return fail(g, "%s", why)
	}
	switch g.Type {
	case "file_exists":
		return gradeFile(g, w)
	case "regex":
		return gradeRegex(g, w)
	case "tool_used":
		return gradeTool(g, w)
	}
	return gradeLLM(g, w, judge)
}

func fail(g Grader, format string, args ...any) Outcome {
	return Outcome{Grader: g.Name, Why: fmt.Sprintf(format, args...)}
}

func pass(g Grader) Outcome {
	return Outcome{Grader: g.Name, Pass: true}
}

// gradeFile counts only files the agent made. A file the scaffold left there
// was in the before list, and claude plugin eval does not count those either.
func gradeFile(g Grader, w Workspace) Outcome {
	now, err := ListFiles(w.Dir)
	if err != nil {
		return fail(g, "cannot list the workspace: %v", err)
	}
	found := false
	for f := range now {
		if w.Before[f] {
			continue
		}
		if ok, _ := path.Match(g.Path, f); ok {
			found = true
			break
		}
	}
	want := g.Exists == nil || *g.Exists
	if found != want {
		return fail(g, "a new file matching %q: want %v, got %v", g.Path, want, found)
	}
	return pass(g)
}

func gradeRegex(g Grader, w Workspace) Outcome {
	text := w.Result.Reply
	if g.Target.Kind == "file" {
		data, err := os.ReadFile(filepath.Join(w.Dir, filepath.FromSlash(g.Target.Path)))
		if err != nil {
			return fail(g, "cannot read %s: %v", g.Target.Path, err)
		}
		text = string(data)
	}
	expr := g.Pattern
	if strings.Contains(g.Flags, "i") {
		expr = "(?i)" + expr
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return fail(g, "bad pattern %q: %v", g.Pattern, err)
	}
	found := re.MatchString(text)
	want := g.Match != "not_contains"
	if found != want {
		return fail(g, "pattern %q: want found %v, got %v", g.Pattern, want, found)
	}
	return pass(g)
}

// gradeTool matches the tool name without case, because the suite says Bash
// and omp calls the same tool bash.
func gradeTool(g Grader, w Workspace) Outcome {
	re, err := regexp.Compile(g.InputMatch)
	if err != nil {
		return fail(g, "bad input_match %q: %v", g.InputMatch, err)
	}
	count := 0
	for _, c := range w.Result.Calls {
		if strings.EqualFold(c.Tool, g.Tool) && re.MatchString(c.Input) {
			count++
		}
	}
	lo := 1
	if g.Min != nil {
		lo = *g.Min
	}
	if count < lo || (g.Max != nil && count > *g.Max) {
		return fail(g, "%s calls matching %q: got %d", g.Tool, g.InputMatch, count)
	}
	return pass(g)
}

const judgePrompt = `You grade one reply from an AI agent against a rubric.

Rubric:
%s

Reply:
%s

Answer with PASS or FAIL as your very first word, then one short sentence why.`

func gradeLLM(g Grader, w Workspace, judge Judge) Outcome {
	ans, err := judge(fmt.Sprintf(judgePrompt, g.Body, w.Result.Reply))
	if err != nil {
		return fail(g, "judge failed: %v", err)
	}
	words := strings.Fields(ans)
	verdict := ""
	if len(words) > 0 {
		verdict = strings.ToUpper(strings.Trim(words[0], "*:.,!"))
	}
	switch verdict {
	case "PASS":
		return pass(g)
	case "FAIL":
		return fail(g, "judge: %s", strings.TrimSpace(ans))
	}
	return fail(g, "judge gave no PASS or FAIL: %q", strings.TrimSpace(ans))
}
```

- [ ] **Step 4: Run the test to see it pass**

Run: `scripts/test ./internal/evalomp -run 'TestGrade|TestUnsupported'`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gofmt -w internal/evalomp && go vet ./internal/evalomp
git add internal/evalomp/grade.go internal/evalomp/grade_test.go
git commit -m "evalomp: grade a run with file_exists, regex, tool_used and llm"
```

---

### Task 4: Run cases through omp

**Files:**
- Create: `internal/evalomp/run.go`
- Test: `internal/evalomp/run_test.go`

**verify:** Every case run calls omp with exactly the command line in Global Constraints, in an empty folder, with `PM_VOICE_FILE` in the throwaway home, and every case ends as exactly one PASS, FAIL or SKIP line whatever goes wrong (scaffold fails, omp exits non-zero, timeout, broken stream, a grader fails, claude-only tag). List each of those paths and the line it prints. Also: no throwaway folder is left behind after `RunAll`.

**Interfaces:**
- Consumes: `Case`, `LoadCases` (Task 1); `ParseStream`, `ListFiles`, `Workspace` (Task 2); `Grade`, `Judge`, `Outcome` (Task 3).
- Produces:
  - `type Options struct { Omp string; PluginDir string; Skills []string }`
  - `var ErrTimeout error`
  - `func SkillNames(pluginDir string) ([]string, error)`
  - `func RunCase(c Case, o Options) (Workspace, error)`
  - `func OmpJudge(omp string) Judge`
  - `func RunAll(cases []Case, o Options, only string, judge Judge, out io.Writer) (failed bool)`

- [ ] **Step 1: Write the failing test**

Create `internal/evalomp/run_test.go`:

```go
package evalomp

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeOmp writes a stand-in omp that records its arguments, its folder and
// PM_VOICE_FILE next to itself, then prints the fixture stream. Tests never
// run the real omp, so they cost no model quota.
func fakeOmp(t *testing.T, body string) (omp, logDir string) {
	t.Helper()
	logDir = t.TempDir()
	omp = filepath.Join(logDir, "omp")
	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$@\" > \"" + logDir + "/args\"\n" +
		"pwd > \"" + logDir + "/pwd\"\n" +
		"ls -A > \"" + logDir + "/ls\"\n" +
		"printf '%s' \"$PM_VOICE_FILE\" > \"" + logDir + "/voice\"\n" +
		body + "\n"
	if err := os.WriteFile(omp, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return omp, logDir
}

func fixture(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs("testdata/stream.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	return "cat \"" + abs + "\""
}

func read(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestSkillNames(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "skills/plan/SKILL.md"), "x")
	write(t, filepath.Join(dir, "skills/brainstorm/SKILL.md"), "x")
	write(t, filepath.Join(dir, "skills/README.md"), "x")
	got, err := SkillNames(dir)
	if err != nil || strings.Join(got, ",") != "brainstorm,plan" {
		t.Errorf("SkillNames = %v, %v", got, err)
	}
}

func TestRunCaseCommandLine(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
	omp, logs := fakeOmp(t, fixture(t))
	o := Options{Omp: omp, PluginDir: "/p/plugin", Skills: []string{"brainstorm", "plan"}}
	w, err := RunCase(Case{Name: "c", Prompt: "Do it.", TimeoutSeconds: 30}, o)
	if err != nil {
		t.Fatal(err)
	}
	want := "-p\n--mode\njson\n--no-session\n--no-extensions\n--no-rules\n--skills=brainstorm,plan\n" +
		"--plugin-dir\n/p/plugin\n-e\n/p/plugin/omp/index.ts\nDo it.\n"
	if got := read(t, filepath.Join(logs, "args")); got != want {
		t.Errorf("args =\n%s\nwant\n%s", got, want)
	}
	if strings.TrimSpace(read(t, filepath.Join(logs, "ls"))) != "" {
		t.Error("omp must start in an empty folder when there is no scaffold")
	}
	pwd, _ := filepath.EvalSymlinks(strings.TrimSpace(read(t, filepath.Join(logs, "pwd"))))
	dir, _ := filepath.EvalSymlinks(w.Dir)
	if pwd != dir {
		t.Errorf("omp ran in %s, workspace is %s", pwd, dir)
	}
	voice := read(t, filepath.Join(logs, "voice"))
	if voice != filepath.Join(filepath.Dir(w.Dir), "home", ".acta", "config.yaml") {
		t.Errorf("PM_VOICE_FILE = %q", voice)
	}
	if w.Result.Reply == "" || len(w.Result.Calls) != 2 {
		t.Errorf("result = %+v", w.Result)
	}
}

func TestRunCaseScaffold(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
	omp, logs := fakeOmp(t, fixture(t))
	dir := t.TempDir()
	scaffold := filepath.Join(dir, "scaffold.sh")
	write(t, scaffold, "set -e\n[ -z \"$(ls -A)\" ]\ntouch pre.txt\nmkdir -p \"$HOME/.acta\"\necho 'chat_language: English' > \"$HOME/.acta/config.yaml\"\n")
	w, err := RunCase(Case{Name: "c", Prompt: "x", TimeoutSeconds: 30, Scaffold: scaffold}, Options{Omp: omp, PluginDir: "/p"})
	if err != nil {
		t.Fatal(err)
	}
	if !w.Before["pre.txt"] {
		t.Errorf("before = %v, want pre.txt", w.Before)
	}
	if got := read(t, read(t, filepath.Join(logs, "voice"))); !strings.Contains(got, "English") {
		t.Errorf("voice file = %q, want what the scaffold wrote in the throwaway home", got)
	}
}

func TestRunCaseFailures(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
	broken := filepath.Join(t.TempDir(), "bad.sh")
	write(t, broken, "exit 4\n")
	cases := map[string]struct {
		body string
		c    Case
	}{
		"scaffold fails": {fixture(t), Case{Prompt: "x", TimeoutSeconds: 30, Scaffold: broken}},
		"omp fails":      {"echo nope >&2; exit 2", Case{Prompt: "x", TimeoutSeconds: 30}},
		"broken stream":  {"echo '{\"type\":\"agent_start\"}'", Case{Prompt: "x", TimeoutSeconds: 30}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			omp, _ := fakeOmp(t, tc.body)
			if _, err := RunCase(tc.c, Options{Omp: omp, PluginDir: "/p"}); err == nil {
				t.Error("want an error")
			}
		})
	}
	t.Run("timeout", func(t *testing.T) {
		omp, _ := fakeOmp(t, "exec sleep 20")
		_, err := RunCase(Case{Prompt: "x", TimeoutSeconds: 1}, Options{Omp: omp, PluginDir: "/p"})
		if !errors.Is(err, ErrTimeout) {
			t.Errorf("err = %v, want ErrTimeout", err)
		}
	})
}

func TestRunAll(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	omp, _ := fakeOmp(t, fixture(t))
	cases := []Case{
		{Name: "good", Prompt: "x", TimeoutSeconds: 30, Graders: []Grader{{Name: "said", Type: "regex", Pattern: "SCR-0001"}}},
		{Name: "bad", Prompt: "x", TimeoutSeconds: 30, Graders: []Grader{{Name: "missing", Type: "regex", Pattern: "NOPE"}}},
		{Name: "claude", Tags: []string{"claude-only"}, Prompt: "x"},
	}
	var out bytes.Buffer
	failed := RunAll(cases, Options{Omp: omp, PluginDir: "/p"}, "", nil, &out)
	got := out.String()
	for _, line := range []string{"PASS good", "FAIL bad: missing (", "SKIP claude (claude-only)", "1 passed, 1 failed, 1 skipped"} {
		if !strings.Contains(got, line) {
			t.Errorf("output lacks %q:\n%s", line, got)
		}
	}
	if !failed {
		t.Error("a failed case must make RunAll report failure")
	}
	left, _ := filepath.Glob(filepath.Join(tmp, "acta-eval-omp-*"))
	if len(left) != 0 {
		t.Errorf("throwaway folders left behind: %v", left)
	}

	out.Reset()
	if RunAll(cases, Options{Omp: omp, PluginDir: "/p"}, "go*", nil, &out) {
		t.Errorf("only the good case ran, so nothing failed:\n%s", out.String())
	}
	if strings.Contains(out.String(), "bad") {
		t.Errorf("--case filter let another case run:\n%s", out.String())
	}
}
```

- [ ] **Step 2: Run the test to see it fail**

Run: `scripts/test ./internal/evalomp -run 'TestSkillNames|TestRunCase|TestRunAll'`
Expected: FAIL, `RunCase`, `RunAll`, `SkillNames`, `Options` and `ErrTimeout` undefined.

- [ ] **Step 3: Write the implementation**

Create `internal/evalomp/run.go`:

```go
package evalomp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Options says which omp to run and which plugin to load into it.
type Options struct {
	Omp       string
	PluginDir string
	Skills    []string
}

// ErrTimeout marks a case that ran past its timeout_seconds.
var ErrTimeout = errors.New("timeout")

// judgeTimeout caps one judge call. A judge only reads one reply.
const judgeTimeout = 120 * time.Second

// SkillNames lists the plugin's skill folders. omp's --skills filter takes
// these names, which keeps acta's skills in and every other skill out.
func SkillNames(pluginDir string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(pluginDir, "skills"))
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// RunCase runs one case in a new throwaway folder and returns what it left.
// The caller removes the folder's parent when done with it.
func RunCase(c Case, o Options) (Workspace, error) {
	base, err := os.MkdirTemp("", "acta-eval-omp-")
	if err != nil {
		return Workspace{}, err
	}
	home, work := filepath.Join(base, "home"), filepath.Join(base, "work")
	for _, d := range []string{home, work} {
		if err := os.Mkdir(d, 0o755); err != nil {
			return Workspace{Dir: work}, err
		}
	}
	if c.Scaffold != "" {
		// The scaffold writes the voice file into HOME. A throwaway home keeps
		// it off the user's real config, which omp itself still needs.
		cmd := exec.Command("bash", c.Scaffold)
		cmd.Dir = work
		cmd.Env = withEnv(os.Environ(), "HOME", home)
		if out, err := cmd.CombinedOutput(); err != nil {
			return Workspace{Dir: work}, fmt.Errorf("scaffold: %v: %s", err, strings.TrimSpace(string(out)))
		}
	}
	before, err := ListFiles(work)
	if err != nil {
		return Workspace{Dir: work}, err
	}
	w := Workspace{Dir: work, Before: before}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.TimeoutSeconds)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, o.Omp, ompArgs(c.Prompt, o)...)
	cmd.Dir = work
	cmd.Env = withEnv(os.Environ(), "PM_VOICE_FILE", filepath.Join(home, ".acta", "config.yaml"))
	// A killed omp can leave a child holding the pipes open. Stop waiting for
	// them soon after, so one stuck case cannot hang the whole run.
	cmd.WaitDelay = 5 * time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	runErr := cmd.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return w, ErrTimeout
	}
	if runErr != nil {
		return w, fmt.Errorf("omp: %v: %s", runErr, strings.TrimSpace(stderr.String()))
	}
	w.Result, err = ParseStream(&stdout)
	return w, err
}

func ompArgs(prompt string, o Options) []string {
	return []string{
		"-p", "--mode", "json", "--no-session", "--no-extensions", "--no-rules",
		"--skills=" + strings.Join(o.Skills, ","),
		"--plugin-dir", o.PluginDir,
		"-e", filepath.Join(o.PluginDir, "omp", "index.ts"),
		prompt,
	}
}

// withEnv sets one variable, dropping any copy already in env so the new
// value is the one the child sees.
func withEnv(env []string, key, value string) []string {
	out := make([]string, 0, len(env)+1)
	for _, kv := range env {
		if !strings.HasPrefix(kv, key+"=") {
			out = append(out, kv)
		}
	}
	return append(out, key+"="+value)
}

// OmpJudge grades with omp's own default model. No plugin, skill or rule is
// loaded: the judge only reads the rubric and the reply.
func OmpJudge(omp string) Judge {
	return func(prompt string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), judgeTimeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, omp, "-p", "--no-session", "--no-extensions", "--no-rules", "--no-skills", prompt)
		cmd.Dir = os.TempDir()
		out, err := cmd.Output()
		return string(out), err
	}
}

// RunAll runs every case whose name matches the only glob (all when only is
// empty), prints one line per case and a total, and says whether any failed.
func RunAll(cases []Case, o Options, only string, judge Judge, out io.Writer) (failed bool) {
	passed, failedN, skipped := 0, 0, 0
	for _, c := range cases {
		if only != "" {
			if ok, _ := path.Match(only, c.Name); !ok {
				continue
			}
		}
		if c.ClaudeOnly() {
			fmt.Fprintf(out, "SKIP %s (claude-only)\n", c.Name)
			skipped++
			continue
		}
		if why := runOne(c, o, judge); why != "" {
			fmt.Fprintf(out, "FAIL %s: %s\n", c.Name, why)
			failedN++
			continue
		}
		fmt.Fprintf(out, "PASS %s\n", c.Name)
		passed++
	}
	fmt.Fprintf(out, "%d passed, %d failed, %d skipped\n", passed, failedN, skipped)
	return failedN > 0
}

// runOne gives "" when every grader passed, else what went wrong.
func runOne(c Case, o Options, judge Judge) string {
	w, err := RunCase(c, o)
	if w.Dir != "" {
		defer os.RemoveAll(filepath.Dir(w.Dir))
	}
	if err != nil {
		return err.Error()
	}
	var fails []string
	for _, g := range c.Graders {
		if r := Grade(g, w, judge); !r.Pass {
			fails = append(fails, fmt.Sprintf("%s (%s)", r.Grader, r.Why))
		}
	}
	return strings.Join(fails, "; ")
}
```

- [ ] **Step 4: Run the test to see it pass**

Run: `scripts/test ./internal/evalomp`
Expected: PASS, every test in the package.

- [ ] **Step 5: Commit**

```bash
gofmt -w internal/evalomp && go vet ./internal/evalomp
git add internal/evalomp/run.go internal/evalomp/run_test.go
git commit -m "evalomp: run each case through omp in a throwaway folder and print one line per case"
```

---

### Task 5: `acta eval-omp`, `scripts/eval --omp`, the claude-only tag and the facts

**Files:**
- Create: `internal/cli/eval_omp.go`
- Test: `internal/cli/eval_omp_test.go`
- Modify: `internal/cli/cli.go` (the `switch args[0]` in `Run` and its `default` message)
- Modify: `scripts/eval`
- Modify: `plugin/evals/second-brainstorm-choices/prompt.md` (the `tags:` line)
- Modify: `internal/evalomp/case_test.go` (add one test)
- Modify: `plugin/evals/FACTS.md` (append a section)
- Modify: `plugin/omp/FACTS.md` (append a section)

**verify:** Every case in the real `plugin/evals` suite either loads with only graders this runner supports, or is tagged `claude-only`; and `acta eval-omp` exits 0, 1 or 3 exactly as Global Constraints says on every path. List each exit path checked (no omp, missing plugin dir, bad flag, extra argument, a failing case, all passing) and its code.

**Interfaces:**
- Consumes: `LoadCases`, `SkillNames`, `RunAll`, `OmpJudge`, `Options`, `Unsupported` (Tasks 1 to 4).
- Produces: `func cmdEvalOmp(args []string, stdout, stderr io.Writer) int` in package `cli`; CLI usage `acta eval-omp [--case <glob>] [plugin-dir]`.

- [ ] **Step 1: Write the failing tests**

Add to the end of `internal/evalomp/case_test.go`:

```go
// TestRealSuiteRunsInOmp keeps the shared suite honest: a new case either
// uses graders this runner can grade, or is tagged claude-only on purpose.
func TestRealSuiteRunsInOmp(t *testing.T) {
	cases, err := LoadCases("../../plugin/evals")
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("no cases found in plugin/evals")
	}
	for _, c := range cases {
		if c.Name == "second-brainstorm-choices" && !c.ClaudeOnly() {
			t.Error("second-brainstorm-choices asks for claude --bg, so it must be tagged claude-only")
		}
		if c.ClaudeOnly() {
			continue
		}
		for _, g := range c.Graders {
			if why := Unsupported(g); why != "" {
				t.Errorf("%s/%s: %s", c.Name, g.Name, why)
			}
		}
	}
}
```

Create `internal/cli/eval_omp_test.go`:

```go
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// evalPlugin builds a tiny plugin with one case whose regex grader looks
// for want, and a fake omp on PATH that prints a stream saying said.
func evalPlugin(t *testing.T, want, said string) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"skills/plan/SKILL.md":            "x",
		"evals/one/prompt.md":             "---\ntimeout_seconds: 30\n---\nHi",
		"evals/one/graders/said.md":       "---\ntype: regex\npattern: \"" + want + "\"\n---\n",
		"bin/omp":                         "#!/bin/sh\necho '{\"type\":\"agent_end\",\"messages\":[{\"role\":\"assistant\",\"content\":[{\"type\":\"text\",\"text\":\"" + said + "\"}]}]}'\n",
	}
	for p, text := range files {
		full := filepath.Join(dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", filepath.Join(dir, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TMPDIR", t.TempDir())
	return dir
}

func TestEvalOmpExitCodes(t *testing.T) {
	var out, errb bytes.Buffer
	dir := evalPlugin(t, "hello", "hello")
	if code := cmdEvalOmp([]string{dir}, &out, &errb); code != exitOK || !strings.Contains(out.String(), "PASS one") {
		t.Errorf("passing suite: code %d, out %q, err %q", code, out.String(), errb.String())
	}

	out.Reset()
	dir = evalPlugin(t, "hello", "bye")
	if code := cmdEvalOmp([]string{"--case", "one", dir}, &out, &errb); code != exitCaseFailed || !strings.Contains(out.String(), "FAIL one") {
		t.Errorf("failing suite: code %d, out %q", code, out.String())
	}

	if code := cmdEvalOmp([]string{filepath.Join(t.TempDir(), "nope")}, &out, &errb); code != exitOther {
		t.Errorf("missing plugin dir: code %d", code)
	}
	if code := cmdEvalOmp([]string{"--bogus"}, &out, &errb); code != exitBadInput {
		t.Errorf("bad flag: code %d", code)
	}
	if code := cmdEvalOmp([]string{"a", "b"}, &out, &errb); code != exitBadInput {
		t.Errorf("two plugin dirs: code %d", code)
	}
}

func TestEvalOmpNoOmp(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	var out, errb bytes.Buffer
	if code := cmdEvalOmp(nil, &out, &errb); code != exitOther || !strings.Contains(errb.String(), "omp is not on PATH") {
		t.Errorf("code %d, err %q", code, errb.String())
	}
}

func TestEvalOmpIsACommand(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	var out, errb bytes.Buffer
	Run([]string{"eval-omp"}, strings.NewReader(""), false, &out, &errb)
	if strings.Contains(errb.String(), "unknown command") {
		t.Errorf("eval-omp is not wired into Run: %q", errb.String())
	}
}
```

Before writing `TestEvalOmpIsACommand`, open `internal/cli/cli.go` and check the exact signature of `Run` (the argument order of stdin, the TTY flag and the writers). Match the call to it; do not change `Run`.

- [ ] **Step 2: Run the tests to see them fail**

Run: `scripts/test ./internal/evalomp ./internal/cli -run 'TestRealSuiteRunsInOmp|TestEvalOmp'`
Expected: FAIL. `cmdEvalOmp` and `exitCaseFailed` are undefined, and `second-brainstorm-choices` is not tagged `claude-only`.

- [ ] **Step 3: Write the implementation**

Create `internal/cli/eval_omp.go`:

```go
package cli

import (
	"flag"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	"github.com/iyay/acta/internal/evalomp"
)

// exitCaseFailed is what claude plugin eval also returns when a case fails,
// so scripts treat both runners the same.
const exitCaseFailed = 1

const evalOmpUsage = "usage: acta eval-omp [--case <glob>] [plugin-dir]"

// cmdEvalOmp runs the plugin's eval cases through omp. Each case costs model
// quota, so no test runs it with the real omp.
func cmdEvalOmp(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("eval-omp", flag.ContinueOnError)
	fs.SetOutput(stderr)
	only := fs.String("case", "", "run only the cases whose name matches this glob")
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) > 1 {
		fmt.Fprintln(stderr, evalOmpUsage)
		return exitBadInput
	}
	plugin := "plugin"
	if len(pos) == 1 {
		plugin = pos[0]
	}
	omp, err := exec.LookPath("omp")
	if err != nil {
		fmt.Fprintln(stderr, "acta eval-omp: omp is not on PATH; install omp first")
		return exitOther
	}
	abs, err := filepath.Abs(plugin)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	skills, err := evalomp.SkillNames(abs)
	if err != nil {
		fmt.Fprintf(stderr, "acta eval-omp: %s is not a plugin folder: %v\n", plugin, err)
		return exitOther
	}
	cases, err := evalomp.LoadCases(filepath.Join(abs, "evals"))
	if err != nil {
		fmt.Fprintf(stderr, "acta eval-omp: %v\n", err)
		return exitOther
	}
	o := evalomp.Options{Omp: omp, PluginDir: abs, Skills: skills}
	if evalomp.RunAll(cases, o, *only, evalomp.OmpJudge(omp), stdout) {
		return exitCaseFailed
	}
	return exitOK
}
```

In `internal/cli/cli.go`, add this case to the `switch args[0]` in `Run`, right after `case "run-one":` and its return line:

```go
	case "eval-omp":
		return cmdEvalOmp(args[1:], stdout, stderr)
```

and in the `default:` message of the same switch, change `reply-back or run-one` to `reply-back, run-one or eval-omp`.

In `plugin/evals/second-brainstorm-choices/prompt.md`, change the line `tags: [brainstorm]` to:

```yaml
tags: [brainstorm, claude-only]
```

In `scripts/eval`, add these lines right after the line `cd "$(dirname "$0")/.."`:

```bash
# --omp runs the same cases through omp instead. The runner comes from this
# checkout; the hooks inside each run still use the acta on PATH.
if [ "${1:-}" = "--omp" ]; then
  shift
  exec go run ./cmd/acta eval-omp "$@" plugin
fi
```

- [ ] **Step 4: Run the tests to see them pass**

Run: `scripts/test ./internal/evalomp ./internal/cli ./internal/plugincheck`
Expected: PASS. `internal/plugincheck` still passes with the new tag.

- [ ] **Step 5: Record the facts**

Append to the end of `plugin/omp/FACTS.md`:

````markdown
## Running the eval cases in omp

Checked 2026-09-30 with omp 18.4.4, from an empty temp folder. `acta eval-omp`
uses exactly these facts.

1. `--mode json` prints one JSON event per line. A tool call is
   `{"type":"tool_execution_start","toolName":"bash","args":{"command":"..."}}`.
   The last line is `{"type":"agent_end","messages":[...]}`, and the final
   reply is the text parts of the last `role: "assistant"` message.
2. `--no-skills` also drops the acta skills that `--plugin-dir` brings. With
   it, the model listed no brainstorm, scratch or plan skill.
3. `--skills=<names>` with the acta skill folder names keeps acta's skills and
   filters out every other one. The model then listed exactly the 12 acta
   skills: brainstorm, bug, build, debug, dispatch, land, migrate, plan,
   review, scratch, setup, tdd.
4. An extension given with `-e` still runs under `--no-extensions`: the run's
   messages held the `acta` custom message, and the model quoted
   "acta plugin is active".
5. omp adds its own built-in messages (`eager-todo-prelude`,
   `eager-task-prelude`) even with every flag above. They are part of omp.

```text
$ command omp --no-session --no-extensions --no-rules \
    --skills="brainstorm,bug,build,debug,dispatch,land,migrate,plan,review,scratch,setup,tdd" \
    --plugin-dir <acta>/plugin -e <acta>/plugin/omp/index.ts --mode json -p \
    "Without running any tool, list the names of every skill available to you, comma separated, then quote the first line of any context message starting with 'acta plugin is active'."
(reply) brainstorm, bug, build, debug, dispatch, land, migrate, plan, review, scratch, setup, tdd
(reply) > acta plugin is active. Before each workflow step, load its acta skill with the Skill tool and follow it. ...
(custom messages) eager-todo-prelude, eager-task-prelude, acta
```
````

Append to the end of `plugin/evals/FACTS.md`:

````markdown
## The same cases in omp

`scripts/eval --omp` (or `acta eval-omp [--case <glob>] [plugin-dir]`) runs
these cases through omp, one at a time, and prints `PASS`, `FAIL` or `SKIP`
per case. It uses omp's own default model for the run and for the `llm` judge.
The omp flags it uses are in `plugin/omp/FACTS.md`.

Each case gets an empty folder and a throwaway home. The scaffold runs with
`HOME` set to that home, and the omp run gets `PM_VOICE_FILE` pointing at the
voice file the scaffold writes there. omp itself keeps the real home, because
it needs the user's login.

Known gaps, on purpose:

1. `max_turns` is ignored. omp has no turn cap flag; only `timeout_seconds`
   stops a run.
2. `allowed_tools` is ignored. omp has no flag for it.
3. The `llm` judge votes once, not best of three, to save quota.
4. A case tagged `claude-only` is skipped. `second-brainstorm-choices` has the
   tag, because its grader asks for `claude --bg`. It is a tag and not a new
   field, because an unknown frontmatter key is an error here.

Only the graders the suite uses are supported: `file_exists`, `regex` on
`last_message` or `{ source: file, path }`, `tool_used`, and `llm`. Anything
else fails with a message that names it, and `TestRealSuiteRunsInOmp` fails
when a new case uses one without the `claude-only` tag.
````

- [ ] **Step 6: Run the checks again and commit**

Run: `scripts/test ./internal/evalomp ./internal/cli ./internal/plugincheck`
Expected: PASS (the FACTS files hold no absolute user path).

```bash
gofmt -w internal/evalomp internal/cli && go vet ./internal/evalomp ./internal/cli
git add internal/cli/eval_omp.go internal/cli/eval_omp_test.go internal/cli/cli.go scripts/eval \
  plugin/evals/second-brainstorm-choices/prompt.md internal/evalomp/case_test.go \
  plugin/evals/FACTS.md plugin/omp/FACTS.md
git commit -m "acta eval-omp runs the plugin evals through omp; scripts/eval --omp; claude-only tag"
```
