---
id: PLAN-31
hash: u6sq
status: approved
---
# Body Schema Plan 1: Base Rule, Date Meta and Scratch Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** One schema table for every kind, turned on for scratch first; scratch commands write and fill sections; every kind gets `created` / `started` / `finished` dates written by the commands; the TUI and doctor show both.

**Architecture:** A new `internal/board/schema.go` holds the table and a pure `CheckBody`. Write commands, `acta id` and `acta doctor` call it only for files that carry `schema: 1`, and only kinds listed in `SchemaOn` ever get `schema: 1` (scratch only in this plan). Dates are frontmatter fields written with the existing `SetField` at the events that already write files; a new `RemoveField` clears `finished` on reopen.

**Tech Stack:** Go, `gopkg.in/yaml.v3` (already used), existing test helpers in `internal/write` (`repoWith`, `baseFiles`, `fixNow`, `gitRun`, `refuse`) and `internal/plugincheck` (`CheckSkill`).

**Spec:** `.acta/specs/2026-09-29-body-schema-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Schema table, verbatim from the spec (required marked *): scratch: Words*, Context, Log, Open questions. bug: Symptom*, Root cause, Repro, Found in, Context. debt: Notes*, Context. spec: Why*, Design*, Testing*, Context. plan: Global Constraints*, File Map*, Waves*, Context.
- `SchemaOn` holds scratch only. Bug, debt, spec and plan are in the table but get `schema: 1` only in their own later plans. No file of those kinds is checked in this plan.
- Only files with `schema: 1` in frontmatter are checked. Old files are never changed or backfilled.
- Missing section error text: `<kind> <file name>: missing ## <Section>`. Nothing is written when it fires; exit 1.
- Old scratch item with `--section`: exit 1, text `SCRATCH-n is an old item with no sections` (with the real id).
- Dates are `YYYY-MM-DD` from `write.Now()`. `started` is written once, never overwritten. `finished` is removed when a status goes back to a working status (`in-progress`, `fixing`, `brainstorming`).
- Working statuses: `in-progress`, `fixing`, `brainstorming`. Closed statuses: whatever `board.Closed` returns true for.
- `acta tick` still never commits (the orchestrator commits per wave). Its date writes follow that rule.
- Comments are plain English a 10-year-old can read; they say why.
- Run `gofmt -l .` (must print nothing) and `go vet ./...` before every commit.

## File Map

| file | task | change |
|---|---|---|
| `internal/board/schema.go` (new) | 1 | table, `SchemaOn`, `HasSchema`, `CheckBody` |
| `internal/board/schema_test.go` (new) | 1 | tests |
| `internal/write/frontmatter.go`, `frontmatter_test.go` | 2 | `RemoveField` |
| `internal/write/dates.go` (new), `dates_test.go` (new) | 2 | `MarkStarted`, `MarkFinished`, `ClearFinished` |
| `internal/write/ops.go`, `ops_test.go` | 2 | `SetValue` writes dates for `status` and `fixed_in` |
| `internal/write/scratch.go`, `scratch_test.go` | 3 | skeleton, `schema: 1`, `--section` |
| `internal/cli/cli.go`, `cli_test.go` | 3 | `scratch add --section` flag and usage |
| `internal/write/ids.go`, `ids_test.go` | 4 | `schema: 1` + `created` on first id, check, scratch `finished` on child spec |
| `internal/cli/tick.go`, `tick_test.go` | 5 | `started` / `finished` on task ticks |
| `internal/doctor/doctor.go`, `doctor_test.go`, `internal/cli/doctor.go`, `internal/cli/doctor_test.go` | 6 | `schema` check |
| `internal/board/board.go`, `internal/tui/detail.go`, `internal/tui/detail_test.go` | 7 | date fields on `Item`, detail lines |
| `plugin/skills/scratch/SKILL.md`, `plugin/skills/brainstorm/SKILL.md`, `internal/plugincheck/skill_scratch_test.go`, `internal/plugincheck/skill_brainstorm_test.go` | 8 | skill rules |

## Waves

- Wave 1: Task 1, Task 2
- Wave 2: Task 3, Task 4, Task 5, Task 6, Task 7 (Task 4 and 5 use Task 2's date helpers; Task 3, 4, 6 use Task 1)
- Wave 3: Task 8 (needs Task 3's `--section` flag)

---

### Task 1: Schema table and body check

**Files:**
- Create: `internal/board/schema.go`
- Test: `internal/board/schema_test.go`

**verify:** For every kind in the table and every body shape, `CheckBody` passes exactly when every required section is present as a `## ` line, the schema sections that are present are in table order, and the body's first non-empty line is a `# ` title. Extra sections after the last schema section pass. List each shape tested: all present, one required missing (per kind), optional missing, out of order, extra at end, extra between schema sections, no title, CRLF endings, heading with trailing spaces, `###` line with a section name (must not count), unknown kind.

**Interfaces:**
- Produces:
  - `type Section struct { Name string; Required bool }`
  - `var Schema map[Kind][]Section`
  - `func SchemaOn(k Kind) bool` — true only for `KindScratch`
  - `func HasSchema(front map[string]any) bool` — true when `front["schema"]` is the int 1 or the string "1"
  - `func CheckBody(k Kind, body string) []string` — problems as `missing ## <Name>`, `## <Name> is out of order`, `missing # title`; nil when clean; nil for a kind not in the table
  - `func SectionNames(k Kind) []string` — names in order

- [x] **Step 1: Write the failing test**

```go
package board

import (
	"reflect"
	"testing"
)

func TestCheckBody(t *testing.T) {
	full := "# T\n\n## Words\n\nx\n\n## Context\n\n## Log\n\n## Open questions\n"
	cases := []struct {
		name string
		kind Kind
		body string
		want []string
	}{
		{"scratch full", KindScratch, full, nil},
		{"scratch only required", KindScratch, "# T\n\n## Words\nx\n", nil},
		{"scratch missing words", KindScratch, "# T\n\n## Context\n", []string{"missing ## Words"}},
		{"out of order", KindScratch, "# T\n## Context\n## Words\n", []string{"## Words is out of order"}},
		{"extra at end", KindScratch, full + "\n## Extra\n", nil},
		{"extra between", KindScratch, "# T\n## Words\n## Extra\n## Context\n", []string{"## Extra is out of order"}},
		{"no title", KindScratch, "## Words\n", []string{"missing # title"}},
		{"crlf", KindScratch, "# T\r\n## Words\r\n", nil},
		{"trailing spaces", KindScratch, "# T\n## Words  \n", nil},
		{"h3 does not count", KindScratch, "# T\n### Words\n", []string{"missing ## Words"}},
		{"spec missing testing", KindStory, "# T\n## Why\n## Design\n", []string{"missing ## Testing"}},
		{"bug needs symptom", KindBug, "# T\n## Repro\n", []string{"missing ## Symptom"}},
		{"unknown kind", KindTask, "anything", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := CheckBody(c.kind, c.body); !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %q want %q", got, c.want)
			}
		})
	}
}

func TestSchemaOnlyScratchIsOn(t *testing.T) {
	for _, k := range []Kind{KindStory, KindPlan, KindBug, KindDebt} {
		if SchemaOn(k) {
			t.Errorf("%s must stay off in plan 1", k)
		}
	}
	if !SchemaOn(KindScratch) {
		t.Error("scratch must be on")
	}
}

func TestHasSchema(t *testing.T) {
	for v, want := range map[any]bool{1: true, "1": true, 2: false, "": false, nil: false} {
		if got := HasSchema(map[string]any{"schema": v}); got != want {
			t.Errorf("schema %v: got %v", v, got)
		}
	}
	if HasSchema(nil) {
		t.Error("nil front must be false")
	}
}
```


- [x] **Step 2: Run it and watch it fail**

Run: `go test ./internal/board -run 'TestCheckBody|TestSchemaOnlyScratchIsOn|TestHasSchema'`
Expected: FAIL, `undefined: CheckBody`

- [x] **Step 3: Write the code**

```go
package board

import "strings"

// Section is one "## " part of a body, in the order the kind wants it.
type Section struct {
	Name     string
	Required bool
}

// Schema lists the sections of each kind in order. It comes from SPEC-22.
var Schema = map[Kind][]Section{
	KindScratch: {{"Words", true}, {"Context", false}, {"Log", false}, {"Open questions", false}},
	KindBug:     {{"Symptom", true}, {"Root cause", false}, {"Repro", false}, {"Found in", false}, {"Context", false}},
	KindDebt:    {{"Notes", true}, {"Context", false}},
	KindStory:   {{"Why", true}, {"Design", true}, {"Testing", true}, {"Context", false}},
	KindPlan:    {{"Global Constraints", true}, {"File Map", true}, {"Waves", true}, {"Context", false}},
}

// SchemaOn says which kinds get "schema: 1" on new files. The other kinds
// wait for their own plan, so their files are never checked yet.
func SchemaOn(k Kind) bool { return k == KindScratch }

// HasSchema says whether a file opted in to the check. Files without it are
// old and are left alone.
func HasSchema(front map[string]any) bool {
	switch v := front["schema"].(type) {
	case int:
		return v == 1
	case string:
		return v == "1"
	}
	return false
}

// SectionNames gives the section names of a kind, in order.
func SectionNames(k Kind) []string {
	var out []string
	for _, s := range Schema[k] {
		out = append(out, s.Name)
	}
	return out
}

// CheckBody lists what is wrong with a body. Nil means it is fine.
func CheckBody(k Kind, body string) []string {
	want, ok := Schema[k]
	if !ok {
		return nil
	}
	pos := map[string]int{}
	for i, s := range want {
		pos[s.Name] = i
	}
	var probs []string
	seen := map[string]bool{}
	firstLine, last, extra := true, -1, ""
	for _, ln := range strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n") {
		ln = strings.TrimRight(ln, " \t")
		if firstLine && ln != "" {
			firstLine = false
			if !strings.HasPrefix(ln, "# ") {
				probs = append(probs, "missing # title")
			}
		}
		if !strings.HasPrefix(ln, "## ") {
			continue
		}
		name := strings.TrimPrefix(ln, "## ")
		seen[name] = true
		i, known := pos[name]
		if !known {
			// An extra section is fine only after every schema one, so
			// remember it until a schema section shows up below it.
			extra = name
			continue
		}
		if extra != "" {
			probs = append(probs, "## "+extra+" is out of order")
			extra = ""
		}
		if i < last {
			probs = append(probs, "## "+name+" is out of order")
			continue
		}
		last = i
	}
	for _, s := range want {
		if s.Required && !seen[s.Name] {
			probs = append(probs, "missing ## "+s.Name)
		}
	}
	return probs
}
```

- [x] **Step 4: Run it and watch it pass**

Run: `go test ./internal/board`
Expected: PASS

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./...
git add internal/board/schema.go internal/board/schema_test.go
git commit -m "board: add body schema table and check"
```

### Task 2: Date helpers and dates on `acta set`

**Files:**
- Modify: `internal/write/frontmatter.go`
- Create: `internal/write/dates.go`
- Modify: `internal/write/ops.go` (`SetValue`)
- Test: `internal/write/frontmatter_test.go`, `internal/write/dates_test.go`, `internal/write/ops_test.go`

**verify:** No path ever overwrites an existing `started`. Every `acta set` that lands on a closed status leaves `finished` set to today. Every `acta set` that lands on a working status leaves no `finished`. `fixed_in` on a bug leaves `finished` set. The body and the other fields stay byte for byte. List every status value tested per kind, and every path that writes or clears a date.

**Interfaces:**
- Consumes: `SetField` (existing), `Now` (existing), `board.Closed`
- Produces:
  - `func RemoveField(src []byte, key string) ([]byte, error)` — no-op when the key is missing; keeps other fields' order and the body
  - `func MarkStarted(src []byte) ([]byte, error)` — writes `started: <today>` only when missing
  - `func MarkFinished(src []byte) ([]byte, error)` — writes `finished: <today>` (overwrites, so a re-close gets the new day)
  - `func ClearFinished(src []byte) ([]byte, error)` — `RemoveField(src, "finished")`
  - `func Working(status string) bool` — `in-progress`, `fixing`, `brainstorming`
  - `func DatesFor(src []byte, status string) ([]byte, error)` — working: `MarkStarted` then `ClearFinished`; closed: `MarkFinished`; other: unchanged

- [x] **Step 1: Write the failing tests**

```go
// dates_test.go
package write

import (
	"strings"
	"testing"
)

func TestDatesFor(t *testing.T) {
	fixNow(t) // 2026-09-26
	src := "---\nid: SCRATCH-1\nstatus: raw\n---\n# T\n"
	cases := []struct {
		status, in, want string
	}{
		{"brainstorming", src, "started: 2026-09-26"},
		{"in-progress", src, "started: 2026-09-26"},
		{"fixing", src, "started: 2026-09-26"},
		{"done", src, "finished: 2026-09-26"},
		{"wontfix", src, "finished: 2026-09-26"},
		{"dropped", src, "finished: 2026-09-26"},
		{"fixed", src, "finished: 2026-09-26"},
	}
	for _, c := range cases {
		out, err := DatesFor([]byte(c.in), c.status)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(out), c.want) {
			t.Errorf("%s: %q lacks %q", c.status, out, c.want)
		}
		if !strings.HasSuffix(string(out), "---\n# T\n") {
			t.Errorf("%s: body changed: %q", c.status, out)
		}
	}
}

func TestStartedIsWrittenOnce(t *testing.T) {
	fixNow(t)
	in := "---\nstarted: 2026-01-01\n---\n# T\n"
	out, _ := DatesFor([]byte(in), "in-progress")
	if !strings.Contains(string(out), "started: 2026-01-01") || strings.Contains(string(out), "2026-09-26") {
		t.Errorf("started was overwritten: %q", out)
	}
}

func TestReopenClearsFinished(t *testing.T) {
	fixNow(t)
	in := "---\nstarted: 2026-01-01\nfinished: 2026-02-01\n---\n# T\n"
	out, _ := DatesFor([]byte(in), "fixing")
	if strings.Contains(string(out), "finished") {
		t.Errorf("finished kept: %q", out)
	}
}

func TestOtherStatusChangesNoDate(t *testing.T) {
	fixNow(t)
	for _, s := range []string{"raw", "draft", "approved", "open"} {
		in := "---\nstatus: x\n---\n# T\n"
		out, _ := DatesFor([]byte(in), s)
		if string(out) != in {
			t.Errorf("%s changed the file: %q", s, out)
		}
	}
}
```

In `frontmatter_test.go` add `TestRemoveField`: key present in the middle (the order of the others is kept), key missing (the output is byte-equal to the input), CRLF file (CRLF kept), and no frontmatter (error, nothing written). In `ops_test.go` add `TestSetValueWritesDates`: `SetValue(..., "status", "brainstorming")` on a scratch file leaves `started:` in the committed file; `"dropped"` then leaves `finished:`; `"brainstorming"` again removes `finished:` and keeps the first `started:`; `SetValue(..., "fixed_in", "abc1234")` on a bug leaves `finished:`.

- [x] **Step 2: Run them and watch them fail**

Run: `go test ./internal/write -run 'TestDatesFor|TestStartedIsWrittenOnce|TestReopenClearsFinished|TestOtherStatusChangesNoDate|TestRemoveField|TestSetValueWritesDates'`
Expected: FAIL, `undefined: DatesFor`

- [x] **Step 3: Write the code**

```go
// dates.go
package write

import "github.com/iyay/acta/internal/board"

// The board works most statuses out from other data, so a date is written
// by the command that changes the file, never by the agent.

// Working says whether a status means someone has begun the work.
func Working(status string) bool {
	switch status {
	case "in-progress", "fixing", "brainstorming":
		return true
	}
	return false
}

// MarkStarted writes today as the start day, but only the first time.
func MarkStarted(src []byte) ([]byte, error) {
	if hasField(src, "started") {
		return src, nil
	}
	return SetField(src, "started", Now().Format("2006-01-02"))
}

// MarkFinished writes today as the day the work closed.
func MarkFinished(src []byte) ([]byte, error) {
	return SetField(src, "finished", Now().Format("2006-01-02"))
}

// ClearFinished takes the close day away when the work opens again.
func ClearFinished(src []byte) ([]byte, error) { return RemoveField(src, "finished") }

// DatesFor gives the file the dates its new status calls for.
func DatesFor(src []byte, status string) ([]byte, error) {
	switch {
	case Working(status):
		out, err := MarkStarted(src)
		if err != nil {
			return nil, err
		}
		return ClearFinished(out)
	case board.Closed(status):
		return MarkFinished(src)
	}
	return src, nil
}
```

`hasField` and `RemoveField` go in `frontmatter.go` and use the same yaml.v3 node walk `SetField` uses, so the order and line endings are kept. In `SetValue`, after `out, err := SetField(src, field, value)`: when `field == "status"`, run `out, err = DatesFor(out, value)`; when `field == "fixed_in"`, run `out, err = MarkFinished(out)`. Return `bad("%s: %v", id, err)` on error, like the line above.

- [x] **Step 4: Run them and watch them pass**

Run: `go test ./internal/write`
Expected: PASS

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./...
git add internal/write/frontmatter.go internal/write/frontmatter_test.go internal/write/dates.go internal/write/dates_test.go internal/write/ops.go internal/write/ops_test.go
git commit -m "write: date meta on status and fixed_in"
```

### Task 3: Scratch skeleton and `--section`

**Files:**
- Modify: `internal/write/scratch.go` (`NewScratch`, `AppendScratch`)
- Modify: `internal/cli/cli.go` (`cmdScratchAdd`, usage lines at `cli.go:102` and in `cmdScratchNew` / `cmdScratchAdd`)
- Test: `internal/write/scratch_test.go`, `internal/cli/cli_test.go`

**verify:** Every new scratch file passes `board.CheckBody(KindScratch, body)` and carries `schema: 1` and `created`. Text sent with any `--section` value lands inside that section and nowhere else, and the text around it stays byte for byte. An old item (no `schema: 1`) never gets its body reshaped: no flag appends at the end the way it does today, and any `--section` is refused with nothing written. List every section value, the default, a bad value, and the old/new item branches tested.

**Interfaces:**
- Consumes: `board.CheckBody`, `board.SectionNames`, `board.HasSchema`, `Now`
- Produces:
  - `func NewScratch(cfg config.Config, slug, title string, body []byte) (Outcome, error)` — same signature; writes `schema: 1` after `created`, and a body of `# <title>\n\n## Words\n\n### <today>\n\n<stdin>\n\n## Context\n\n## Log\n\n## Open questions\n`
  - `func AppendScratch(cfg config.Config, b *board.Board, id, section string, text []byte) (Outcome, error)` — `section` is `""` (no flag given), `words`, `context`, `log` or `questions`. `""` means `words` on a new item and "append at the end" on an old item.
  - CLI: `acta scratch add <SCRATCH-n> [--section words|context|log|questions] < text.md`

- [x] **Step 1: Write the failing tests**

```go
func TestNewScratchWritesSkeleton(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, baseFiles)
	o, err := NewScratch(cfg, "idea", "Idea", []byte("kata user\n"))
	if err != nil {
		t.Fatal(err)
	}
	src, _ := os.ReadFile(o.Path)
	doc := board.Parse(src)
	if !board.HasSchema(doc.Front) {
		t.Errorf("no schema: 1 in %q", src)
	}
	if doc.Front["created"] != "2026-09-26" {
		t.Errorf("created = %v", doc.Front["created"])
	}
	want := "# Idea\n\n## Words\n\n### 2026-09-26\n\nkata user\n\n## Context\n\n## Log\n\n## Open questions\n"
	if doc.Body != want {
		t.Errorf("body\n%q\nwant\n%q", doc.Body, want)
	}
	if p := board.CheckBody(board.KindScratch, doc.Body); p != nil {
		t.Errorf("new file fails its own schema: %v", p)
	}
}

func TestAppendScratchSections(t *testing.T) {
	cases := []struct {
		section, want string
	}{
		{"", "## Words\n\n### 2026-09-26\n\nkata user\n\n### 2026-09-26\n\nmore\n\n## Context"},
		{"words", "## Words\n\n### 2026-09-26\n\nkata user\n\n### 2026-09-26\n\nmore\n\n## Context"},
		{"context", "## Context\n\nmore\n\n## Log"},
		{"log", "## Log\n\n### 2026-09-26\n\nmore\n\n## Open questions"},
		{"questions", "## Open questions\n\nmore\n"},
	}
	for _, c := range cases {
		t.Run("section="+c.section, func(t *testing.T) {
			fixNow(t)
			cfg := repoWith(t, baseFiles)
			o, _ := NewScratch(cfg, "idea", "Idea", []byte("kata user\n"))
			b, _ := board.Load(cfg)
			if _, err := AppendScratch(cfg, b, o.ShortID, c.section, []byte("more\n")); err != nil {
				t.Fatal(err)
			}
			src, _ := os.ReadFile(o.Path)
			if !strings.Contains(string(src), c.want) {
				t.Errorf("%q lacks %q", src, c.want)
			}
			if p := board.CheckBody(board.KindScratch, board.Parse(src).Body); p != nil {
				t.Errorf("append broke the schema: %v", p)
			}
		})
	}
}

func TestAppendScratchBadSection(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, baseFiles)
	o, _ := NewScratch(cfg, "idea", "Idea", []byte("x\n"))
	b, _ := board.Load(cfg)
	refuse(t, cfg, `unknown section "notes"; use words, context, log or questions`, func() error {
		_, err := AppendScratch(cfg, b, o.ShortID, "notes", []byte("y\n"))
		return err
	})
}

func TestAppendScratchOldItem(t *testing.T) {
	// An old item has no schema field. No flag keeps today's append; a flag is refused.
	fixNow(t)
	cfg := repoWith(t, map[string]string{
		"scratch/2026-09-01-old.md": "---\nid: SCRATCH-1\nhash: aaaa\ntitle: old\nstatus: raw\n---\nfree text\n",
	})
	b, _ := board.Load(cfg)
	if _, err := AppendScratch(cfg, b, "SCRATCH-1", "", []byte("more\n")); err != nil {
		t.Fatal(err)
	}
	src, _ := os.ReadFile(filepath.Join(cfg.Root, "scratch", "2026-09-01-old.md"))
	if !strings.HasSuffix(string(src), "free text\n\nmore\n") {
		t.Errorf("old append changed: %q", src)
	}
	b, _ = board.Load(cfg)
	refuse(t, cfg, "SCRATCH-1 is an old item with no sections", func() error {
		_, err := AppendScratch(cfg, b, "SCRATCH-1", "context", []byte("y\n"))
		return err
	})
}
```

Every other caller of `AppendScratch` in the existing tests now passes `""` as `section`. Their expected output stays the same, because their fixtures are old items. In `cli_test.go`, add a case for `acta scratch add SCRATCH-1 --section context` (exit 0, the text lands in Context), one for `--section bogus` (exit 1, the error on stderr), and one checking that `-h` prints the new usage line.

- [x] **Step 2: Run them and watch them fail**

Run: `go test ./internal/write -run 'Scratch' && go test ./internal/cli -run 'Scratch'`
Expected: FAIL, because the body has no sections and `AppendScratch` gets too many arguments

- [x] **Step 3: Write the code**

In `NewScratch`, add `{"schema", "1"}` after `{"created", ...}` in the field list. Then build the body:

```go
day := Now().Format("2006-01-02")
words := strings.TrimRight(string(body), "\n")
content = append(content, fmt.Sprintf("# %s\n\n## Words\n\n### %s\n\n%s\n\n## Context\n\n## Log\n\n## Open questions\n", title, day, words)...)
```

In `AppendScratch`, add a `section string` parameter:

```go
// sections maps the --section value to the heading it fills. Words and log
// entries get a dated heading, so a reader sees when each one came in.
var sections = map[string]struct {
	heading string
	dated   bool
}{
	"words":     {"Words", true},
	"context":   {"Context", false},
	"log":       {"Log", true},
	"questions": {"Open questions", false},
}
```

Flow:
1. If the section is not `""` and not a key of `sections`, return `bad("unknown section %q; use words, context, log or questions", section)`.
2. Parse the file. When `!board.HasSchema(doc.Front)`: with a section set, return `bad("%s is an old item with no sections", it.ShortID)`; with none, keep today's append path unchanged.
3. A new item with `section == ""` uses `"words"`.
4. Find the `## <heading>` line and the next `## ` line after it (or the end of the file). Insert the text just before that next heading, with one blank line on each side. A dated section gets `### <today>\n\n` in front of the text.
5. When the heading is missing from a new item, add it at its schema place: before the first later schema heading that is there, or at the end.
6. Run `board.CheckBody` on the result. On any problem, return `bad("scratch %s: %s", filepath.Base(it.Path), p[0])` and write nothing.

In `cli.go`, `cmdScratchAdd` gets `section := fs.String("section", "", "words, context, log or questions (default: words)")` and passes `*section`. Update the three usage strings to `acta scratch add <SCRATCH-n> [--section words|context|log|questions] < text.md`.

- [x] **Step 4: Run them and watch them pass**

Run: `go test ./internal/write ./internal/cli`
Expected: PASS

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./...
git add internal/write/scratch.go internal/write/scratch_test.go internal/cli/cli.go internal/cli/cli_test.go
git commit -m "scratch: write sections and add --section"
```

### Task 4: `acta id` writes schema and created, checks, and closes specced scratch

**Files:**
- Modify: `internal/write/ids.go` (`AssignIDs`)
- Test: `internal/write/ids_test.go`

**verify:** `schema: 1` and `created` are written only on a file's first id, and only `created` for a kind not in `SchemaOn`. No file that already had an id ever gains `schema` or `created`. A `schema: 1` file whose body fails `CheckBody` gets no id, shows up in `Outcome.Skips` with the spec's error text, and does not stop the other files. When a spec gets its first id and its frontmatter `parent:` names a scratch item, that scratch file gets `finished` in the same commit. List every file state tested: new scratch, new spec, file with an id already, schema file that fails, spec with a scratch parent, spec with a non-scratch parent.

**Interfaces:**
- Consumes: `board.SchemaOn`, `board.HasSchema`, `board.CheckBody`, `MarkFinished` (Task 2), `SetField`
- Produces: no new API. `AssignIDs` keeps its signature.

- [x] **Step 1: Write the failing tests**

```go
func TestAssignIDsFirstIDWritesCreated(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, map[string]string{
		"specs/2026-09-26-x-design.md": "# X\n\n## Why\n",
	})
	b, _ := board.Load(cfg)
	if _, _, err := AssignIDs(cfg, b, nil); err != nil {
		t.Fatal(err)
	}
	src, _ := os.ReadFile(filepath.Join(cfg.Root, "specs", "2026-09-26-x-design.md"))
	doc := board.Parse(src)
	if doc.Front["created"] != "2026-09-26" {
		t.Errorf("created = %v", doc.Front["created"])
	}
	if board.HasSchema(doc.Front) {
		t.Error("spec is not on yet, so it must not get schema: 1")
	}
}

func TestAssignIDsLeavesOldFilesAlone(t *testing.T) {
	fixNow(t)
	in := "---\nid: SPEC-1\nhash: abcd\n---\n# X\n"
	cfg := repoWith(t, map[string]string{"specs/2026-09-01-x-design.md": in})
	b, _ := board.Load(cfg)
	AssignIDs(cfg, b, nil)
	src, _ := os.ReadFile(filepath.Join(cfg.Root, "specs", "2026-09-01-x-design.md"))
	if string(src) != in {
		t.Errorf("old file changed: %q", src)
	}
}

func TestAssignIDsSkipsBrokenSchemaFile(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, map[string]string{
		"scratch/2026-09-26-bad.md":  "---\ntitle: bad\nstatus: raw\nschema: 1\n---\n# bad\n\n## Context\n",
		"specs/2026-09-26-ok-design.md": "# OK\n",
	})
	b, _ := board.Load(cfg)
	_, o, err := AssignIDs(cfg, b, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !containsLine(o.Skips, "scratch 2026-09-26-bad.md: missing ## Words") {
		t.Errorf("skips = %v", o.Skips)
	}
	bad, _ := os.ReadFile(filepath.Join(cfg.Root, "scratch", "2026-09-26-bad.md"))
	if strings.Contains(string(bad), "id:") {
		t.Error("broken file got an id")
	}
	ok, _ := os.ReadFile(filepath.Join(cfg.Root, "specs", "2026-09-26-ok-design.md"))
	if !strings.Contains(string(ok), "id: SPEC-") {
		t.Error("the good file next to it got no id")
	}
}

func TestAssignIDsFinishesScratchParent(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, map[string]string{
		"scratch/2026-09-20-idea.md":      "---\nid: SCRATCH-1\nhash: aaaa\ntitle: idea\nstatus: brainstorming\n---\nx\n",
		"specs/2026-09-26-idea-design.md": "---\nparent: scratch/2026-09-20-idea\n---\n# Idea\n",
		"specs/2026-09-26-other-design.md": "---\nparent: bugs/2026-09-20-b\n---\n# Other\n",
	})
	b, _ := board.Load(cfg)
	AssignIDs(cfg, b, nil)
	src, _ := os.ReadFile(filepath.Join(cfg.Root, "scratch", "2026-09-20-idea.md"))
	if !strings.Contains(string(src), "finished: 2026-09-26") {
		t.Errorf("scratch parent not finished: %q", src)
	}
	if n := gitRun(t, cfg.RepoRoot, "log", "--format=%s", "-1"); n != "acta: assign short ids" {
		t.Errorf("finished must ride the same commit, last commit %q", n)
	}
}
```

`containsLine` is a tiny helper in the test file: true when any string in the slice contains the text. The spec's error text is `<kind> <file name>: missing ## <Section>`, so a skip line must contain that exact part.

- [x] **Step 2: Run them and watch them fail**

Run: `go test ./internal/write -run 'TestAssignIDs'`
Expected: FAIL, because `created` is missing and the broken file gets an id

- [x] **Step 3: Write the code**

In the `AssignIDs` loop, before the id is assigned:

1. Parse `orig`. When `board.HasSchema(front)` and `CheckBody(c.it.Kind, doc.Body)` returns problems, add `fmt.Sprintf("skip %s: %s %s: %s", fileID(cfg, c.file), c.it.Kind, filepath.Base(c.file), p[0])` to `skips` and `continue`.
2. Inside the `if id == ""` branch, after the id is set, write `created` (`Now().Format("2006-01-02")`). When `board.SchemaOn(c.it.Kind)` and the file has no `schema` yet, also write `schema: 1`.
3. Also inside `if id == ""`: when the kind is a spec and `front["parent"]` is a string that starts with `scratch/`, look the parent up with `b.Get`. If it is there and on disk, run `MarkFinished` on its file and add the path to `paths`.

Kind names in the skip text: use the same words the spec's error text uses (`scratch`, `bug`, `debt`, `spec`, `plan`). Map a spec's kind constant to `spec` and a plan file to `plan` with a small switch.

- [x] **Step 4: Run them and watch them pass**

Run: `go test ./internal/write`
Expected: PASS

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./...
git add internal/write/ids.go internal/write/ids_test.go
git commit -m "id: write created and schema on first id, check schema files"
```

### Task 5: Dates on task ticks

**Files:**
- Modify: `internal/cli/tick.go` (`cmdTick`)
- Test: `internal/cli/tick_test.go`

**verify:** The first tick or `--start` on any task of a plan leaves `started` on the plan file and on its spec or bug (the item named by the plan's `SpecID`), and no later tick changes either one. `finished` lands on the plan only when every task of that plan is fully ticked, and on the spec only when every plan under it is done. A debt line tick writes no date. `acta tick` still makes no commit. List every tick kind tested (`--start`, `--step`, `--all`, `--wontfix` on debt) and every spec state tested (one plan, two plans with one still open).

**Interfaces:**
- Consumes: `MarkStarted`, `MarkFinished` (Task 2), `board.Load`, `Item.PlanPath`, `Item.SpecID`, `Item.Status`
- Produces: none

- [x] **Step 1: Write the failing tests**

Add them to `tick_test.go`. `tickRepo` and `runTick` already exist there. The new fixture helper `datesRepo` writes one spec and either one or two plans that point at it:

```go
// datesRepo holds one spec and the plans that carry its tasks, so a tick can
// be followed up to the spec. With two, the second plan stays open.
func datesRepo(t *testing.T, plans int) string {
	t.Helper()
	dir := t.TempDir()
	for _, d := range []string{"specs", "plans"} {
		if err := os.MkdirAll(filepath.Join(dir, ".acta", d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"specs/2026-09-20-s-design.md": "---\nid: SPEC-1\nhash: ssss\n---\n# S\n",
		"plans/2026-09-21-a.md": "---\nid: PLAN-1\nhash: aaaa\n---\n# A\n\n**Spec:** `.acta/specs/2026-09-20-s-design.md`\n\n### Task 1: One\n- [ ] a\n\n### Task 2: Two\n- [ ] b\n",
	}
	if plans == 2 {
		files["plans/2026-09-22-b.md"] = "---\nid: PLAN-2\nhash: bbbb\n---\n# B\n\n**Spec:** `.acta/specs/2026-09-20-s-design.md`\n\n### Task 1: One\n- [ ] c\n"
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, ".acta", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func onDay(t *testing.T, day int) {
	t.Helper()
	old := write.Now
	write.Now = func() time.Time { return time.Date(2026, 9, day, 10, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { write.Now = old })
}

func specOf(dir string) string { return filepath.Join(dir, ".acta", "specs", "2026-09-20-s-design.md") }
func planA(dir string) string  { return filepath.Join(dir, ".acta", "plans", "2026-09-21-a.md") }

func TestTickWritesStartedOnPlanAndSpec(t *testing.T) {
	dir := datesRepo(t, 1)
	onDay(t, 26)
	if code, _, errs := runTick(t, dir, "plans/2026-09-21-a#task-1", "--start"); code != 0 {
		t.Fatalf("exit %d: %s", code, errs)
	}
	for _, p := range []string{planA(dir), specOf(dir)} {
		if !strings.Contains(read(t, p), "started: 2026-09-26") {
			t.Errorf("%s lacks started: %q", p, read(t, p))
		}
	}
	onDay(t, 27)
	runTick(t, dir, "plans/2026-09-21-a#task-1", "--all")
	for _, p := range []string{planA(dir), specOf(dir)} {
		if s := read(t, p); !strings.Contains(s, "started: 2026-09-26") || strings.Contains(s, "started: 2026-09-27") {
			t.Errorf("a second tick moved started in %s: %q", p, s)
		}
	}
}

func TestTickStepAlsoStarts(t *testing.T) {
	dir := datesRepo(t, 1)
	onDay(t, 26)
	runTick(t, dir, "plans/2026-09-21-a#task-2", "--step", "1")
	if !strings.Contains(read(t, planA(dir)), "started: 2026-09-26") {
		t.Errorf("a --step tick did not start the plan: %q", read(t, planA(dir)))
	}
}

func TestTickFinishesPlanOnlyWhenAllTasksDone(t *testing.T) {
	dir := datesRepo(t, 1)
	onDay(t, 26)
	runTick(t, dir, "plans/2026-09-21-a#task-1", "--all")
	if strings.Contains(read(t, planA(dir)), "finished:") {
		t.Fatalf("plan finished with a task still open: %q", read(t, planA(dir)))
	}
	runTick(t, dir, "plans/2026-09-21-a#task-2", "--all")
	for _, p := range []string{planA(dir), specOf(dir)} {
		if !strings.Contains(read(t, p), "finished: 2026-09-26") {
			t.Errorf("%s lacks finished: %q", p, read(t, p))
		}
	}
}

func TestTickSpecWaitsForEveryPlan(t *testing.T) {
	dir := datesRepo(t, 2)
	onDay(t, 26)
	runTick(t, dir, "plans/2026-09-21-a#task-1", "--all")
	runTick(t, dir, "plans/2026-09-21-a#task-2", "--all")
	if !strings.Contains(read(t, planA(dir)), "finished: 2026-09-26") {
		t.Errorf("plan A lacks finished: %q", read(t, planA(dir)))
	}
	if strings.Contains(read(t, specOf(dir)), "finished:") {
		t.Errorf("spec finished while plan B is open: %q", read(t, specOf(dir)))
	}
}

func TestTickDebtLineWritesNoDate(t *testing.T) {
	dir := tickRepo(t)
	onDay(t, 26)
	runTick(t, dir, "DEBT-1.1", "--all")
	if s := read(t, debtFile(dir)); strings.Contains(s, "started:") || strings.Contains(s, "finished:") {
		t.Errorf("debt tick wrote a date: %q", s)
	}
}

func TestTickMakesNoCommit(t *testing.T) {
	dir := datesRepo(t, 1)
	gitInit(t, dir) // git init, add and commit everything
	before := gitOut(t, dir, "rev-list", "--count", "HEAD")
	onDay(t, 26)
	runTick(t, dir, "plans/2026-09-21-a#task-1", "--all")
	runTick(t, dir, "plans/2026-09-21-a#task-2", "--all")
	if after := gitOut(t, dir, "rev-list", "--count", "HEAD"); after != before {
		t.Errorf("tick made a commit: %s -> %s", before, after)
	}
}
```

Before writing new ones, check `internal/cli/*_test.go` for helpers that already run git (`grep -n "func git" internal/cli/*_test.go`), and reuse any you find as `gitInit` / `gitOut`. When none exist, write both in `tick_test.go` with `exec.Command("git", ...)`, setting `cmd.Dir = dir`, the user name and email set inline, and `t.Fatal` on error. Check the real task id form with `acta tick -h`, and the debt line id with the test at `tick_test.go:60`. When a form is different from the one above, use the real form.

- [x] **Step 2: Run them and watch them fail**

Run: `go test ./internal/cli -run 'TestTick'`
Expected: FAIL, because the plan and spec files hold no `started`

- [x] **Step 3: Write the code**

In `cmdTick`, after the `switch` that ticks (and before `RecordAgent`), only for `it.Kind == board.KindTask`:

```go
// The plan and the spec above it learn when work began and ended. The
// files stay uncommitted like the tick itself.
if err := markTaskDates(cfg, it); err != nil {
	fmt.Fprintf(stderr, "dates: %v\n", err)
}
```

`markTaskDates` lives in `tick.go`:
1. Run `MarkStarted` on `it.PlanPath`, and on the spec or bug file of `b.Get(plan.SpecID)` when there is one.
2. Reload the board. When the plan item has every task done (`Done == Total` over its `Children`), run `MarkFinished` on the plan.
3. When the spec item's derived `Status` is `done`, run `MarkFinished` on the spec.
4. Read and write each file with `os.ReadFile` / `os.WriteFile`, and only when the bytes changed.

A date failure only gets printed, the same way the agent record does. The tick itself already worked.

- [x] **Step 4: Run them and watch them pass**

Run: `go test ./internal/cli`
Expected: PASS

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./...
git add internal/cli/tick.go internal/cli/tick_test.go
git commit -m "tick: write started and finished on plan and spec"
```

### Task 6: `acta doctor` schema check

**Files:**
- Modify: `internal/doctor/doctor.go` (`Env`, `Run`, new `checkSchema`)
- Modify: `internal/cli/doctor.go` (fill `Env.SchemaProblems`)
- Test: `internal/doctor/doctor_test.go`, `internal/cli/doctor_test.go`

**verify:** Every `schema: 1` file with a body problem shows up in one `warn schema` line, in the spec's error text, and no file without `schema: 1` ever shows up. A clean repo prints `ok schema`. Doctor itself reads no files: the CLI hands it the list. List the repo states tested: no .acta, clean, one bad schema file, a bad file without schema (ignored), two bad files (both named, sorted).

**Interfaces:**
- Consumes: `board.Load`, `board.HasSchema`, `board.CheckBody`
- Produces: `Env.SchemaProblems []string`, and `checkSchema(e Env) Result` with `Name: "schema"`. Put it in `Run` after `checkRepo`.

- [x] **Step 1: Write the failing tests**

```go
func TestCheckSchema(t *testing.T) {
	r := checkSchema(Env{})
	if r.Level != OK || r.Msg != "every schema file has its sections" {
		t.Errorf("clean: %+v", r)
	}
	r = checkSchema(Env{SchemaProblems: []string{"scratch a.md: missing ## Words", "scratch b.md: missing ## Words"}})
	if r.Level != Warn || r.Msg != "scratch a.md: missing ## Words; scratch b.md: missing ## Words" {
		t.Errorf("bad: %+v", r)
	}
	if r.Fix != "add the missing sections, or run acta scratch add --section" {
		t.Errorf("fix: %q", r.Fix)
	}
}
```

In `internal/cli/doctor_test.go`, build a repo with one bad `schema: 1` scratch file and one bad file that has no schema. Want: the doctor output has `warn schema: scratch 2026-...-bad.md: missing ## Words`, and it does not name the file without schema.

- [x] **Step 2: Run them and watch them fail**

Run: `go test ./internal/doctor ./internal/cli -run 'Schema'`
Expected: FAIL, `undefined: checkSchema`

- [x] **Step 3: Write the code**

```go
// checkSchema reports files that opted in to the body schema but lost a
// section. The CLI does the reading, so this stays a pure check.
func checkSchema(e Env) Result {
	if len(e.SchemaProblems) == 0 {
		return Result{Name: "schema", Level: OK, Msg: "every schema file has its sections"}
	}
	return Result{Name: "schema", Level: Warn, Msg: strings.Join(e.SchemaProblems, "; "),
		Fix: "add the missing sections, or run acta scratch add --section"}
}
```

In `internal/cli/doctor.go`, when a repo root is known, load the board. For each item that is not a task, not a debt line and not legacy, and whose file has `schema: 1`, run `CheckBody`. Build `<kind> <base name>: <problem>` strings and sort them. A board that fails to load adds nothing here, because `checkRepo` already covers a broken repo.

- [x] **Step 4: Run them and watch them pass**

Run: `go test ./internal/doctor ./internal/cli`
Expected: PASS

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./...
git add internal/doctor/doctor.go internal/doctor/doctor_test.go internal/cli/doctor.go internal/cli/doctor_test.go
git commit -m "doctor: warn on schema files missing a section"
```

### Task 7: Dates on the board and in the TUI detail

**Files:**
- Modify: `internal/board/board.go` (`Item`, where frontmatter fields are read)
- Modify: `internal/tui/detail.go` (`detailLines` field list)
- Test: `internal/tui/detail_test.go`

**verify:** Each of `created`, `started` and `finished` shows up in the detail exactly when the file's frontmatter holds it, and a file with none of them gets no new lines. Bad values (not a date, a number, empty) never show and never crash. List every combination tested: none, each one alone, all three, a bad value for each.

**Interfaces:**
- Produces: `Item.Created, Item.StartedOn, Item.Finished string`. The name is `StartedOn` because `Item.Started bool` already exists for tasks. Each holds `YYYY-MM-DD` or `""`.

- [x] **Step 1: Write the failing tests**

Both tests go in `internal/tui/detail_test.go`. They use the helpers that are already there: `treeCfg`, `detailLines`, `plainLines`, `labelsOf`. The board fields are checked through `board.Load` in the same test.

```go
func datedFiles() map[string]string {
	return map[string]string{
		".acta/scratch/2026-09-01-all.md":  "---\nid: SCRATCH-1\ntitle: all\nstatus: raw\ncreated: 2026-09-01\nstarted: 2026-09-02\nfinished: 2026-09-03\n---\nx\n",
		".acta/scratch/2026-09-01-none.md": "---\nid: SCRATCH-2\ntitle: none\nstatus: raw\n---\nx\n",
		".acta/scratch/2026-09-01-one.md":  "---\nid: SCRATCH-3\ntitle: one\nstatus: raw\nstarted: 2026-09-02\n---\nx\n",
		".acta/scratch/2026-09-01-bad.md":  "---\nid: SCRATCH-4\ntitle: bad\nstatus: raw\ncreated: \"\"\nstarted: 12\nfinished: soon\n---\nx\n",
	}
}

func TestItemDates(t *testing.T) {
	cfg := treeCfg(t, datedFiles())
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ id, created, started, finished string }{
		{"SCRATCH-1", "2026-09-01", "2026-09-02", "2026-09-03"},
		{"SCRATCH-2", "", "", ""},
		{"SCRATCH-3", "", "2026-09-02", ""},
		{"SCRATCH-4", "", "", ""},
	}
	for _, c := range cases {
		it := b.Get(c.id)
		if it == nil {
			t.Fatalf("no %s", c.id)
		}
		if it.Created != c.created || it.StartedOn != c.started || it.Finished != c.finished {
			t.Errorf("%s: got %q %q %q", c.id, it.Created, it.StartedOn, it.Finished)
		}
	}
}

func TestDetailShowsDates(t *testing.T) {
	cfg := treeCfg(t, datedFiles())
	has := func(id string) map[string]bool {
		got := map[string]bool{}
		for _, l := range labelsOf(plainLines(detailLines(t, cfg, id))) {
			got[l] = true
		}
		return got
	}
	all := has("SCRATCH-1")
	for _, l := range []string{"CREATED", "STARTED", "FINISHED"} {
		if !all[l] {
			t.Errorf("SCRATCH-1 lacks %s", l)
		}
	}
	for _, id := range []string{"SCRATCH-2", "SCRATCH-4"} {
		got := has(id)
		for _, l := range []string{"CREATED", "STARTED", "FINISHED"} {
			if got[l] {
				t.Errorf("%s shows %s with no good date", id, l)
			}
		}
	}
	one := has("SCRATCH-3")
	if !one["STARTED"] || one["CREATED"] || one["FINISHED"] {
		t.Errorf("SCRATCH-3 labels: %v", one)
	}
}
```

Check what `labelsOf` returns for the date values too. When it gives only labels, add one more check that `2026-09-02` shows up on the STARTED line of SCRATCH-1.

- [x] **Step 2: Run them and watch them fail**

Run: `go test ./internal/board ./internal/tui -run 'Dates'`
Expected: FAIL, `Item has no field Created`

- [x] **Step 3: Write the code**

In `board.go`, wherever `Ref` / `FixedIn` are read from the frontmatter, read the three date keys the same way. Keep a value only when it is a string that parses with `time.Parse("2006-01-02", v)`. yaml can decode a bare date into a `time.Time`, so a `time.Time` also counts: format it back to `2006-01-02`.

In `detail.go`, add these after `{"AUTHOR", it.Author}`:

```go
{"CREATED", it.Created},
{"STARTED", it.StartedOn},
{"FINISHED", it.Finished},
```

Empty values are already skipped by the loop that draws these fields. Check this in the test.

- [x] **Step 4: Run them and watch them pass**

Run: `go test ./internal/board ./internal/tui`
Expected: PASS

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./...
git add internal/board/board.go internal/tui/detail.go internal/tui/detail_test.go
git commit -m "tui: show created, started and finished in the detail"
```

### Task 8: Skill rules for scratch and brainstorm

**Files:**
- Modify: `plugin/skills/scratch/SKILL.md`
- Modify: `plugin/skills/brainstorm/SKILL.md`
- Test: `internal/plugincheck/skill_scratch_test.go`, `internal/plugincheck/skill_brainstorm_test.go`

**verify:** The scratch skill cannot lose the rule that context is added right after `new` with `--section context`, the four things the context holds, the `--section questions` rule, or the ban on hand-written `---` and "Agent notes". The brainstorm skill cannot go back to a plain `acta scratch add` for answers and approved sections. The old wording "You may add your own lines below theirs" cannot come back in any form. List every Must and MustNot string added.

**Interfaces:**
- Consumes: the CLI usage from Task 3

- [ ] **Step 1: Write the failing tests**

In `skill_scratch_test.go`, add these to `Must`: `"--section context"`, `"--section questions"`, `"right after"`, `"what work was going on"`, `"what already exists"`, `"file:line"`, `"where the facts came from"`. Add these to `MustNot`: `"You may add your own lines"`, `"below theirs"`. Add one more test that fails when the skill text says to write `---` or `Agent notes` by hand in any form other than the ban line: the only line allowed to hold either one is the ban line itself.

In `skill_brainstorm_test.go`, add this to `Must`: `"acta scratch add SCRATCH-n --section log"`.

- [ ] **Step 2: Run them and watch them fail**

Run: `go test ./internal/plugincheck -run 'TestSkillScratch|TestSkillBrainstorm'`
Expected: FAIL, missing `--section context`

- [ ] **Step 3: Write the skill text**

In the scratch skill, replace the "What goes in" bullets with:

```markdown
## What goes in

- The user's words, verbatim. Do not tidy or rewrite them. `new` puts them under `## Words`.
- Right after `new`, add context with `acta scratch add SCRATCH-n --section context`: what work was going on, what already exists, the file:line spots, and where the facts came from (reading, debug, a run).
- Open questions go in with `--section questions`.
- More words from the user later go in with `--section words`, the default.
- Never write `---` or "Agent notes" by hand; the sections do that job.
- Pasted images are written as their paths, not as descriptions.
```

Update the `How to file` add line to the new usage. Keep the file under 60 lines, the limit `MaxLines` sets.

In the brainstorm skill, change the two "append with `acta scratch add SCRATCH-n`" spots (Step 0 "Append as you go" and checklist items 2 and 4) to `acta scratch add SCRATCH-n --section log`.

- [ ] **Step 4: Run them and watch them pass**

Run: `go test ./internal/plugincheck`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
gofmt -l . && go vet ./...
git add plugin/skills/scratch/SKILL.md plugin/skills/brainstorm/SKILL.md internal/plugincheck/skill_scratch_test.go internal/plugincheck/skill_brainstorm_test.go
git commit -m "skills: scratch context and brainstorm log use --section"
```
