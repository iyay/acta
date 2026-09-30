---
parent: specs/2026-10-01-priority-and-build-worktree-design
id: PLN-0064
created: "2026-10-01 05:22:21"
hash: lwfa6qz
started: "2026-10-01 05:27:35"
finished: "2026-10-01 05:52:33"
---
# Priority for Bugs and Debt Items, and Build Always Uses git worktree add Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Bugs and debt items can carry a priority (high, medium, low) that the CLI writes and the TUI sorts and shows, and `acta:build` always makes its worktree with `git worktree add`.

**Architecture:** `internal/board` gets one small file with the three levels and a parser for the `(high) ` tag, and fills a new `Item.Priority` from a bug's frontmatter or a debt line's tag. `internal/write` writes the field on `bug new` and `set <id> priority`. `internal/tui` sorts the Bugs and Debts open lists by priority after the id order and draws an H/M/L tag and a PRIORITY meta line. The build skill loses its native worktree step.

**Tech Stack:** Go, Bubble Tea + lipgloss, gopkg.in/yaml.v3 (through the existing `SetField` / `RemoveField`).

**Spec:** `.acta/specs/2026-10-01-priority-and-build-worktree-design.md`

**Tests:** fast `scripts/test ./internal/<pkg>`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- The field is `priority`. Levels are exactly `high`, `medium`, `low`. `none` exists only as the `acta set` value that removes it. No severity field.
- Priority is optional everywhere. Old files are not migrated. Unset is a normal state, never a Problem.
- A debt tag is exactly `(high) `, `(medium) ` or `(low) ` (with one space after) right after the checkbox. Anything else, like `(hgh) `, is plain text.
- Priority is only for bugs and debt items. Any other kind is refused with the words `priority is only for bugs and debt items`.
- Comments are plain English a 10-year-old can read. They say why, not what.
- Every task runs only the package it touches. The full suite runs once, in `acta:land`.

## File map

- `plugin/skills/build/SKILL.md`: Step 1 loses 1a; `git worktree add` is the only way. (Task 1)
- `internal/plugincheck/skill_build_test.go`: MustNot guards for the native step. (Task 1)
- `plugin/skills/review/SKILL.md`, `plugin/skills/bug/SKILL.md`: one line each about priority. (Task 2)
- `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/skill_bug_test.go`: Must guards for those lines. (Task 2)
- `internal/board/priority.go` (new): `Priorities`, `ValidPriority`, `SplitPriority`. (Task 3)
- `internal/board/board.go`: `Item.Priority`; bug frontmatter read; debt line tag read in `linkDebt`. (Task 3)
- `internal/board/priority_test.go` (new). (Task 3)
- `internal/write/priority.go` (new): `setPriority`, `setBugPriority`, `setLinePriority`. (Task 4)
- `internal/write/ops.go`: `SetValue` routes `priority`; `NewBug` and `StartBug` take a priority. (Task 4)
- `internal/write/priority_test.go` (new); `internal/write/ops_test.go`: the eight `NewBug(` and two `StartBug(` calls gain the new argument. (Task 4)
- `internal/cli/cli.go`: `--priority` on `bug new`; usage lines. (Task 4)
- `internal/tui/model.go`: the one `write.StartBug(` call gains `""`. (Task 4)
- `internal/tui/order.go`: `byPriority`. (Task 5)
- `internal/tui/sidebar.go`: `openRows` calls `byPriority` for Bugs and Debts. (Task 5)
- `internal/tui/scroll.go`: row tag in `rowText`, tag color in `styles.paintID`. (Task 5)
- `internal/tui/styles.go`: `slotBrightRed` and the `priority` brushes. (Task 5)
- `internal/tui/detail.go`: PRIORITY meta line. (Task 5)
- `internal/tui/priority_test.go` (new). (Task 5)

## Waves

- Wave 1: Task 1, Task 2, Task 3 (no shared files).
- Wave 2: Task 4 (needs `board.ValidPriority` and `board.SplitPriority` from Task 3).
- Wave 3: Task 5 (needs `Item.Priority` from Task 3). It shares no file with Task 4, but Task 4 changes `write.StartBug`, which `internal/tui/model.go` calls, so running them side by side in one worktree would break the `internal/tui` build halfway.

---

### Task 1: Build always uses git worktree add

**Files:**
- Modify: `plugin/skills/build/SKILL.md` (the `### Step 1: Create Isolated Workspace` section)
- Test: `internal/plugincheck/skill_build_test.go`

**verify:** No text in `plugin/skills/build/` tells an agent to use a native or harness worktree tool, in any wording, and `git worktree add "../$REPO-$SLUG" -b "$SLUG" "$PARENT"` is the one way the skill gives to make a worktree. List every file in `plugin/skills/build/` you checked.

**Interfaces:**
- Consumes: nothing.
- Produces: nothing other tasks use.

- [x] **Step 1: Write the failing test**

In `TestSkillBuild`, add to the `MustNot` list:

```go
"EnterWorktree", "Native Worktree Tools", "native worktree tool", "Git Worktree Fallback", "Step 1a",
```

and add to the `Must` list:

```go
"A native worktree tool puts the worktree inside the repo",
```

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/plugincheck -run TestSkillBuild`
Expected: FAIL naming `EnterWorktree`, `Native Worktree Tools` and the missing Must line.

- [x] **Step 3: Edit the skill**

Replace everything from `### Step 1: Create Isolated Workspace` up to (not including) `Sandbox fallback:` with:

````markdown
### Step 1: Create Isolated Workspace

Always use `git worktree add`, even when the harness has its own worktree tool. A native worktree tool puts the worktree inside the repo and starts it from `origin/<default-branch>`. The user pushes by hand, so the spec and plan just committed on local main are often not on origin, and that worktree would start without the plan.

Directory selection, in priority order. Explicit user preference always beats observed filesystem state.

1. Check your instructions for a declared worktree directory preference. If the user has already specified one, use it without asking.
2. Otherwise use `../<repo>-<slug>`, next to the repo, never inside it (`$REPO` is the repo folder name, `$SLUG` the branch name).

Create it. `$PARENT` is the parent branch recorded above, so the branch starts from it and not from whatever is checked out:

```bash
git worktree add "../$REPO-$SLUG" -b "$SLUG" "$PARENT"
cd "../$REPO-$SLUG"
```

````

Then grep the whole folder so no other line still points at a native tool:

```bash
grep -rn -i -E "native|EnterWorktree|1a|1b" plugin/skills/build/
```

Any hit that tells the agent to use a harness worktree tool goes too. Hits about other things stay.

- [x] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/plugincheck`
Expected: PASS (the whole package, since size caps and other skill checks read the same file).

- [x] **Step 5: Commit**

```bash
gofmt -l internal/plugincheck && go vet ./internal/plugincheck
git add plugin/skills/build/SKILL.md internal/plugincheck/skill_build_test.go
git commit -m "Build always makes its worktree with git worktree add"
```

---

### Task 2: Skills name priority

**Files:**
- Modify: `plugin/skills/review/SKILL.md` (the NOTE line that names `acta debt new`)
- Modify: `plugin/skills/bug/SKILL.md` (the `## Record` section, below the `acta bug new` example)
- Test: `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/skill_bug_test.go`

**verify:** Both skills say priority is optional, and neither tells an agent to always set one. Every level word the skills use is one of high, medium, low. List each sentence in both skills that names priority.

**Interfaces:**
- Consumes: nothing.
- Produces: nothing other tasks use.

- [x] **Step 1: Write the failing test**

In `TestSkillReview` `Must`, add:

```go
"A NOTE may start with `(high) `, `(medium) ` or `(low) `",
```

In `TestSkillBug` `Must`, add:

```go
"`--priority high|medium|low` is optional",
```

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/plugincheck -run 'TestSkillReview|TestSkillBug'`
Expected: FAIL, both Must lines missing.

- [x] **Step 3: Edit the skills**

In `plugin/skills/review/SKILL.md`, right after the sentence ending `The debt file merges with the branch.`, add on the same bullet:

```markdown
 A NOTE may start with `(high) `, `(medium) ` or `(low) ` when it matters more or less than the rest; with no tag it is unset, which is fine.
```

In `plugin/skills/bug/SKILL.md`, right after the closing ```` ``` ```` of the `acta bug new` example in `## Record`, add a paragraph:

```markdown
`--priority high|medium|low` is optional. Set it when the user named how urgent the bug is; leave it off otherwise. `acta set <bug id> priority <level>` changes it later, and `none` removes it.
```

- [x] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/plugincheck`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
git add plugin/skills/review/SKILL.md plugin/skills/bug/SKILL.md internal/plugincheck/skill_review_test.go internal/plugincheck/skill_bug_test.go
git commit -m "Review and bug skills name the optional priority"
```

---

### Task 3: Board reads priority

**Files:**
- Create: `internal/board/priority.go`
- Modify: `internal/board/board.go` (the `Item` struct; the item builder right after `it.FixedIn = field(doc.Front, "fixed_in")`; `linkDebt`)
- Test: `internal/board/priority_test.go`

**verify:** For every bug file and every debt line on the board, `Item.Priority` is one of `high`, `medium`, `low` or empty, and a debt item's `Title` never starts with a valid tag. A bad bug value always leaves a Problem and an empty Priority; a bad debt tag always stays in the Title and leaves no Problem. List every way a value can reach `Item.Priority` and what each gives.

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `board.Item.Priority string` — `"high"`, `"medium"`, `"low"` or `""`.
  - `var board.Priorities = []string{"high", "medium", "low"}` — most urgent first.
  - `func board.ValidPriority(s string) bool`
  - `func board.SplitPriority(text string) (level, rest string)` — `("high", "note")` for `"(high) note"`, `("", text)` when there is no valid tag.

- [x] **Step 1: Write the failing test**

Create `internal/board/priority_test.go`:

```go
package board

import "testing"

func TestSplitPriority(t *testing.T) {
	t.Parallel()

	for _, c := range []struct{ in, level, rest string }{
		{"(high) a note", "high", "a note"},
		{"(medium) a note", "medium", "a note"},
		{"(low) a note", "low", "a note"},
		{"(hgh) a note", "", "(hgh) a note"},
		{"(high)a note", "", "(high)a note"},
		{"a note (high) here", "", "a note (high) here"},
		{"", "", ""},
	} {
		level, rest := SplitPriority(c.in)
		if level != c.level || rest != c.rest {
			t.Errorf("SplitPriority(%q) = %q, %q; want %q, %q", c.in, level, rest, c.level, c.rest)
		}
	}
}

func TestBugPriorityFromFrontmatter(t *testing.T) {
	t.Parallel()

	b := boardWith(t, map[string]string{
		"bugs/2026-10-01-high.md":  "---\npriority: high\n---\n# High\n",
		"bugs/2026-10-01-none.md":  "---\nref: X-1\n---\n# None\n",
		"bugs/2026-10-01-wrong.md": "---\npriority: urgent\n---\n# Wrong\n",
	})
	if got := b.Get("bugs/2026-10-01-high").Priority; got != "high" {
		t.Errorf("high bug Priority = %q", got)
	}
	none := b.Get("bugs/2026-10-01-none")
	if none.Priority != "" || len(none.Problems) != 0 {
		t.Errorf("unset bug Priority = %q, Problems = %v; want empty and none", none.Priority, none.Problems)
	}
	wrong := b.Get("bugs/2026-10-01-wrong")
	if wrong.Priority != "" || !hasProblem(wrong, "unknown priority urgent") {
		t.Errorf("bad bug Priority = %q, Problems = %v", wrong.Priority, wrong.Problems)
	}
}

func TestSpecIgnoresPriority(t *testing.T) {
	t.Parallel()

	b := boardWith(t, map[string]string{"specs/2026-10-01-s.md": "---\npriority: high\n---\n# S\n"})
	if got := b.Get("specs/2026-10-01-s").Priority; got != "" {
		t.Errorf("spec Priority = %q, want empty", got)
	}
}

func TestDebtItemPriorityFromTag(t *testing.T) {
	t.Parallel()

	b := boardWith(t, map[string]string{
		"debt/2026-10-01-d.md": "---\nid: DBT-0001\nhash: aaaaaaa\n---\n# Review NOTEs: D\n\n- [ ] (high) first\n- [x] (low) second\n- [ ] (hgh) third\n- [ ] fourth\n",
	})
	for _, c := range []struct{ id, level, title string }{
		{"debt/2026-10-01-d#item-1", "high", "first"},
		{"debt/2026-10-01-d#item-2", "low", "second"},
		{"debt/2026-10-01-d#item-3", "", "(hgh) third"},
		{"debt/2026-10-01-d#item-4", "", "fourth"},
	} {
		it := b.Get(c.id)
		if it == nil {
			t.Fatalf("no %s on the board", c.id)
		}
		if it.Priority != c.level || it.Title != c.title {
			t.Errorf("%s Priority = %q Title = %q; want %q %q", c.id, it.Priority, it.Title, c.level, c.title)
		}
		if len(it.Problems) != 0 {
			t.Errorf("%s Problems = %v, want none", c.id, it.Problems)
		}
	}
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/board -run 'Priority'`
Expected: FAIL to build: `undefined: SplitPriority`, `it.Priority undefined`.

- [x] **Step 3: Write minimal implementation**

Create `internal/board/priority.go`:

```go
package board

import (
	"slices"
	"strings"
)

// Priorities are the levels a bug or a debt item can carry, most urgent
// first, so the TUI can sort by their place in this list.
var Priorities = []string{"high", "medium", "low"}

// ValidPriority says whether s is one of the levels.
func ValidPriority(s string) bool { return slices.Contains(Priorities, s) }

// SplitPriority reads a "(high) " tag at the start of a debt line. A line
// with no tag, or with a word in brackets that is not a level, comes back
// whole, because a note may really start with brackets.
func SplitPriority(text string) (level, rest string) {
	for _, p := range Priorities {
		if r, ok := strings.CutPrefix(text, "("+p+") "); ok {
			return p, r
		}
	}
	return "", text
}
```

In `internal/board/board.go`, add to `Item` after `FixedIn string`:

```go
	Priority     string // bugs and debt items: high, medium, low, or "" when unset
```

Right after `it.FixedIn = field(doc.Front, "fixed_in")`:

```go
	// Only a bug carries a priority in its frontmatter. A word that is not a
	// level is shown as a problem, and the bug counts as unset.
	if p := field(doc.Front, "priority"); p != "" && it.Kind == KindBug {
		if ValidPriority(p) {
			it.Priority = p
		} else {
			it.Problems = append(it.Problems, "unknown priority "+p)
		}
	}
```

In `linkDebt`, inside the loop, before `item := &Item{...}`:

```go
		level, title := SplitPriority(line.Text)
```

and in the `Item` literal change `Title: line.Text` to `Title: title, Priority: level`.

- [x] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/board`
Expected: PASS (whole package, so the old debt and bug tests still hold).

- [x] **Step 5: Commit**

```bash
gofmt -l internal/board && go vet ./internal/board
git add internal/board/priority.go internal/board/priority_test.go internal/board/board.go
git commit -m "Board reads priority from bugs and debt lines"
```

---

### Task 4: CLI writes priority

**Files:**
- Create: `internal/write/priority.go`
- Modify: `internal/write/ops.go` (`SetValue`, `NewBug`, `StartBug`)
- Modify: `internal/cli/cli.go` (`cmdSet`, `cmdBugNew`, the `bug new` usage line near the top of the dispatch too)
- Modify: `internal/tui/model.go` (the one `write.StartBug(m.cfg, s, "")` call becomes `write.StartBug(m.cfg, s, "", "")`)
- Modify: `internal/write/ops_test.go` (existing `NewBug(` and `StartBug(` calls gain `""` for priority)
- Test: `internal/write/priority_test.go`

**verify:** No command ever writes a priority value other than `high`, `medium` or `low`, and no refused command leaves a file changed or a new commit. `set ... priority` on any kind other than bug and debt item is refused. List every command path that can write a priority (bug new from stdin, bug new through the editor, set on a bug, set on a debt item, debt new), and for each the checks it runs before writing.

**Interfaces:**
- Consumes: `board.ValidPriority`, `board.SplitPriority`, `board.Item.Priority` (Task 3).
- Produces:
  - `func write.NewBug(cfg config.Config, slug, title, ref, priority string, body []byte) (Outcome, error)`
  - `func write.StartBug(cfg config.Config, slug, ref, priority string) (string, []byte, error)`
  - `write.SetValue(cfg, b, id, "priority", "high"|"medium"|"low"|"none")`

- [x] **Step 1: Write the failing test**

Create `internal/write/priority_test.go`:

```go
package write

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const prioDebt = "---\nid: DBT-0001\nhash: aaaaaaa\n---\n# Review NOTEs: D\n\n- [ ] first\n- [x] (low) second\n"

func prioRepo(t *testing.T) (cfgRoot string, set func(id, value string) error) {
	t.Helper()
	cfg := repoWith(t, map[string]string{
		".acta/bugs/2026-09-24-crash.md":      "---\nref: B-1\n---\n# Crash\n\n## Symptom\nIt crashes.\n",
		".acta/debt/2026-10-01-d.md":          prioDebt,
		".acta/specs/2026-10-01-s.md":         "---\nid: SPC-0001\n---\n# S\n",
		".acta/plans/2026-10-01-p.md":         "---\nid: PLN-0001\n---\n# P\n\n### Task 1: T\n- [ ] a\n",
		".acta/scratch/2026-10-01-idea.md":    "---\nid: SCR-0001\n---\n# Idea\n",
	})
	return cfg.Root, func(id, value string) error {
		_, err := SetValue(cfg, mustLoad(t, cfg), id, "priority", value)
		return err
	}
}

func TestSetPriorityOnBug(t *testing.T) {
	root, set := prioRepo(t)
	path := filepath.Join(root, "bugs/2026-09-24-crash.md")
	if err := set("bugs/2026-09-24-crash", "high"); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, path); !strings.Contains(got, "priority: high\n") {
		t.Fatalf("bug file = %q", got)
	}
	if err := set("bugs/2026-09-24-crash", "none"); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, path); strings.Contains(got, "priority") {
		t.Fatalf("none left the field: %q", got)
	}
}

func TestSetPriorityOnDebtItem(t *testing.T) {
	root, set := prioRepo(t)
	path := filepath.Join(root, "debt/2026-10-01-d.md")
	if err := set("debt/2026-10-01-d#item-1", "medium"); err != nil {
		t.Fatal(err)
	}
	if err := set("debt/2026-10-01-d#item-2", "high"); err != nil {
		t.Fatal(err)
	}
	want := "- [ ] (medium) first\n- [x] (high) second\n"
	if got := readFile(t, path); !strings.HasSuffix(got, want) {
		t.Fatalf("debt file = %q, want it to end with %q", got, want)
	}
	if err := set("debt/2026-10-01-d#item-2", "none"); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, path); !strings.HasSuffix(got, "- [x] second\n") {
		t.Fatalf("none left the tag: %q", got)
	}
}

func TestSetPriorityRefuses(t *testing.T) {
	root, set := prioRepo(t)
	before := map[string]string{}
	for _, p := range []string{"bugs/2026-09-24-crash.md", "debt/2026-10-01-d.md"} {
		before[p] = readFile(t, filepath.Join(root, p))
	}
	for _, c := range []struct{ id, value, want string }{
		{"specs/2026-10-01-s", "high", "priority is only for bugs and debt items"},
		{"plans/2026-10-01-p", "high", "priority is only for bugs and debt items"},
		{"plans/2026-10-01-p#task-1", "high", "priority is only for bugs and debt items"},
		{"scratch/2026-10-01-idea", "high", "priority is only for bugs and debt items"},
		{"bugs/2026-09-24-crash", "urgent", "priority must be high, medium, low or none"},
		{"debt/2026-10-01-d#item-1", "hgh", "priority must be high, medium, low or none"},
	} {
		err := set(c.id, c.value)
		if !errors.Is(err, ErrBadInput) || !strings.Contains(err.Error(), c.want) {
			t.Errorf("set %s priority %s: err = %v, want %q", c.id, c.value, err, c.want)
		}
	}
	for p, was := range before {
		if got := readFile(t, filepath.Join(root, p)); got != was {
			t.Errorf("%s changed after a refused set", p)
		}
	}
}

func TestNewBugPriority(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, baseFiles)
	body := []byte("# Hang\n\n## Symptom\nIt hangs.\n")
	if _, err := NewBug(cfg, "hang", "", "", "low", body); err != nil {
		t.Fatal(err)
	}
	matches, _ := filepath.Glob(filepath.Join(cfg.Root, "bugs", "*-hang.md"))
	if len(matches) != 1 || !strings.Contains(readFile(t, matches[0]), "priority: low\n") {
		t.Fatalf("bug files %v, want one with priority: low", matches)
	}
	_, err := NewBug(cfg, "stall", "", "", "urgent", []byte("# Stall\n\n## Symptom\nIt stalls.\n"))
	if !errors.Is(err, ErrBadInput) || !strings.Contains(err.Error(), "priority must be high, medium or low") {
		t.Fatalf("bad priority err = %v", err)
	}
	if m, _ := filepath.Glob(filepath.Join(cfg.Root, "bugs", "*-stall.md")); len(m) != 0 {
		t.Fatalf("a refused bug left %v", m)
	}
}

func TestStartBugPriority(t *testing.T) {
	cfg := repoWith(t, baseFiles)
	path, tmpl, err := StartBug(cfg, "slow", "", "high")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(tmpl), "priority: high\n") {
		t.Fatalf("template = %q", tmpl)
	}
	if _, _, err := StartBug(cfg, "slower", "", "urgent"); !errors.Is(err, ErrBadInput) {
		t.Fatalf("bad priority err = %v", err)
	}
	_ = os.Remove(path)
}

func TestNewDebtKeepsTheTag(t *testing.T) {
	fixNowAt(t, "2026-09-27")
	cfg := repoWith(t, map[string]string{".acta/plans/2026-09-26-short-ids.md": debtPlan})
	if _, err := NewDebt(cfg, mustLoad(t, cfg), "PLAN-3", "", []byte("(high) big one\n(hgh) typo one\nplain one\n")); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, filepath.Join(cfg.Root, "debt", "2026-09-27-short-ids.md"))
	for _, want := range []string{"- [ ] (high) big one\n", "- [ ] (hgh) typo one\n", "- [ ] plain one\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	b := mustLoad(t, cfg)
	if it := b.Get("debt/2026-09-27-short-ids#item-1"); it == nil || it.Priority != "high" || it.Title != "big one" {
		t.Errorf("item-1 = %+v", it)
	}
}
```

Update the existing calls in `internal/write/ops_test.go`: every `NewBug(cfg, a, b, c, body)` becomes `NewBug(cfg, a, b, c, "", body)`, and every `StartBug(cfg, a, b)` becomes `StartBug(cfg, a, b, "")`.

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/write -run 'Priority|KeepsTheTag'`
Expected: FAIL to build: too many arguments to `NewBug` / `StartBug`.

- [x] **Step 3: Write minimal implementation**

Create `internal/write/priority.go`:

```go
package write

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// debtBoxRe splits a debt line into its checkbox and the words after it.
var debtBoxRe = regexp.MustCompile(`^(\s*[-*] \[[ xX-]\] )(.*)$`)

// setPriority sets or removes the priority of one bug or debt item and
// commits the file. Every check runs before the file is touched.
func setPriority(cfg config.Config, it *board.Item, id, value string) (Outcome, error) {
	switch {
	case it.Kind != board.KindBug && it.Kind != board.KindDebtItem:
		return Outcome{}, bad("priority is only for bugs and debt items")
	case it.Legacy:
		return Outcome{}, bad("%s is a legacy file; move it into the root folder first", id)
	case it.Worktree != "":
		return Outcome{}, bad("%s is shown from worktree %s; edit it there", id, it.Worktree)
	case value != "none" && !board.ValidPriority(value):
		return Outcome{}, bad("priority must be high, medium, low or none, not %q", value)
	}
	dirty, err := dirtyBefore(cfg, it.Path)
	if err != nil {
		return Outcome{}, err
	}
	if it.Kind == board.KindBug {
		err = setBugPriority(it.Path, value)
	} else {
		err = setLinePriority(it.Path, it.Line, value)
	}
	if err != nil {
		return Outcome{}, bad("%s: %v", id, err)
	}
	return finish(cfg, it.Path, fmt.Sprintf("acta: %s priority %s", id, value), dirty), nil
}

// setBugPriority writes the priority field, or takes it out for none.
func setBugPriority(path, value string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var out []byte
	if value == "none" {
		out, err = RemoveField(src, "priority")
	} else {
		out, err = SetField(src, "priority", value)
	}
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

// setLinePriority swaps the tag on one debt line. An old tag goes first, so
// a line never ends up with two.
func setLinePriority(path string, line int, value string) error {
	unlock, err := lock(path)
	if err != nil {
		return err
	}
	defer unlock()
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(src), "\n")
	if line < 1 || line > len(lines) {
		return fmt.Errorf("line %d is not a checklist box", line)
	}
	m := debtBoxRe.FindStringSubmatch(lines[line-1])
	if m == nil {
		return fmt.Errorf("line %d is not a checklist box", line)
	}
	_, rest := board.SplitPriority(m[2])
	if value != "none" {
		rest = "(" + value + ") " + rest
	}
	lines[line-1] = m[1] + rest
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
```

In `internal/write/ops.go`, `SetValue`: add a case to the first `switch`, right after `case it == nil:`:

```go
	case field == "priority":
		return setPriority(cfg, it, id, value)
```

and change the `default:` message to `unknown field %q; use status, type, fixed_in, ref or priority`.

`NewBug`: add `priority string` after `ref string`. As the first lines of the function:

```go
	if priority != "" && !board.ValidPriority(priority) {
		return Outcome{}, bad("priority must be high, medium or low, not %q", priority)
	}
```

and right after the `hash` `SetField` block:

```go
	if priority != "" {
		if content, err = SetField(content, "priority", priority); err != nil {
			return Outcome{}, err
		}
	}
```

`StartBug`: add `priority string` after `ref string`. The same `ValidPriority` check as its first lines. After `tmpl := BugTemplate(...)`:

```go
	if priority != "" {
		if tmpl, err = SetField(tmpl, "priority", priority); err != nil {
			return "", nil, err
		}
	}
```

In `internal/cli/cli.go`:
- `cmdSet` usage: `usage: acta set <id> status|type|fixed_in|ref|priority <value>`.
- `cmdBugNew`: add `priority := fs.String("priority", "", "high, medium or low")`; pass `*priority` to `write.NewBug(cfg, pos[0], *title, *ref, *priority, body)` and `write.StartBug(cfg, pos[0], *ref, *priority)`.
- Both `bug new` usage lines: `usage: acta bug new <slug> [--ref X] [--title T] [--priority P] < body.md`.

In `internal/tui/model.go`: `write.StartBug(m.cfg, s, "")` becomes `write.StartBug(m.cfg, s, "", "")`.

- [x] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/write ./internal/cli` and `go build ./internal/tui`
Expected: PASS, and the TUI still builds. (Write commands auto-commit; these tests run in temp repos from `repoWith`, never in this checkout.)

- [x] **Step 5: Commit**

```bash
gofmt -l internal/write internal/cli internal/tui && go vet ./internal/write ./internal/cli ./internal/tui
git add internal/write/priority.go internal/write/priority_test.go internal/write/ops.go internal/write/ops_test.go internal/cli/cli.go internal/tui/model.go
git commit -m "bug new --priority and acta set <id> priority"
```

---

### Task 5: TUI sorts and shows priority

**Files:**
- Modify: `internal/tui/order.go` (add `byPriority`)
- Modify: `internal/tui/sidebar.go` (`openRows`)
- Modify: `internal/tui/scroll.go` (`rowText`, `styles.paintID`)
- Modify: `internal/tui/styles.go` (`slotBrightRed`, `priority` brushes in `styles` and `newStyles`)
- Modify: `internal/tui/detail.go` (the header `fields` list)
- Test: `internal/tui/priority_test.go`

**verify:** In the Bugs and Debts open lists, no row ever sits above a row of higher priority, in either `o` direction, and rows of one level keep the id order they had before. The Done pane, search and every other tab keep exactly the order they had before. A row with no priority draws exactly as it did before. List every list builder you checked (`openRows`, `doneRows`, `activityRows`, `searchRows`) and whether it sorts by priority.

**Interfaces:**
- Consumes: `board.Item.Priority`, `board.Priorities` (Task 3).
- Produces: `func byPriority(items []*board.Item) []*board.Item` (new slice, stable).

- [x] **Step 1: Write the failing test**

Create `internal/tui/priority_test.go`:

```go
package tui

import (
	"slices"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// prioCfg is a board with bugs and debt lines of every level, and some with
// none, so one fixture covers every sort group.
func prioCfg(t *testing.T) config.Config {
	t.Helper()
	bug := func(id, prio, extra string) string {
		fm := "---\nid: " + id + "\nhash: " + strings.ToLower(strings.ReplaceAll(id, "-", ""))[:7] + "\n"
		if prio != "" {
			fm += "priority: " + prio + "\n"
		}
		return fm + extra + "---\n# " + id + "\n"
	}
	cfg := treeCfg(t, map[string]string{
		".acta/bugs/2026-10-01-a.md": bug("BUG-0001", "", ""),
		".acta/bugs/2026-10-01-b.md": bug("BUG-0002", "low", ""),
		".acta/bugs/2026-10-01-c.md": bug("BUG-0003", "high", ""),
		".acta/bugs/2026-10-01-d.md": bug("BUG-0004", "high", ""),
		".acta/bugs/2026-10-01-e.md": bug("BUG-0005", "medium", ""),
		".acta/bugs/2026-10-01-f.md": bug("BUG-0006", "low", "fixed_in: abc1234\n"),
		".acta/bugs/2026-10-01-g.md": bug("BUG-0007", "high", "fixed_in: abc1234\n"),
		".acta/debt/2026-10-01-d.md": "---\nid: DBT-0001\nhash: dbt0001\n---\n# Review NOTEs: D\n\n- [ ] plain\n- [ ] (low) l\n- [ ] (high) h\n",
	})
	return cfg
}

func prioModel(t *testing.T) Model {
	t.Helper()
	cfg := prioCfg(t)
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	return m
}

func TestBugsSortByPriorityBothWays(t *testing.T) {
	t.Parallel()

	m := press(prioModel(t), tabKey(tabBugs))
	want := []string{"bugs/2026-10-01-c", "bugs/2026-10-01-d", "bugs/2026-10-01-e", "bugs/2026-10-01-b", "bugs/2026-10-01-a"}
	if got := rowIDs(m); !slices.Equal(got, want) {
		t.Fatalf("Bugs rows %v, want %v", got, want)
	}
	m = press(m, "o")
	want = []string{"bugs/2026-10-01-d", "bugs/2026-10-01-c", "bugs/2026-10-01-e", "bugs/2026-10-01-b", "bugs/2026-10-01-a"}
	if got := rowIDs(m); !slices.Equal(got, want) {
		t.Fatalf("Bugs rows after o %v, want %v", got, want)
	}
}

func TestBugsDonePaneKeepsIDOrder(t *testing.T) {
	t.Parallel()

	m := press(prioModel(t), tabKey(tabBugs), "tab")
	if got, want := doneRowIDs(m), []string{"bugs/2026-10-01-f", "bugs/2026-10-01-g"}; !slices.Equal(got, want) {
		t.Fatalf("Fixed rows %v, want id order %v", got, want)
	}
}

func TestDebtsSortByPriority(t *testing.T) {
	t.Parallel()

	m := press(prioModel(t), tabKey(tabDebts))
	want := []string{"debt/2026-10-01-d#item-3", "debt/2026-10-01-d#item-2", "debt/2026-10-01-d#item-1"}
	if got := rowIDs(m); !slices.Equal(got, want) {
		t.Fatalf("Debts rows %v, want %v", got, want)
	}
}

func TestRowShowsPriorityTag(t *testing.T) {
	t.Parallel()

	m := prioModel(t)
	for _, c := range []struct{ id, want string }{
		{"bugs/2026-10-01-c", "BUG-0003  H BUG-0003"},
		{"bugs/2026-10-01-e", "BUG-0005  M BUG-0005"},
		{"bugs/2026-10-01-b", "BUG-0002  L BUG-0002"},
		{"bugs/2026-10-01-a", "BUG-0001  BUG-0001"},
	} {
		if got := m.rowText(row{id: c.id}, m.board.Get(c.id), 80); got != c.want {
			t.Errorf("row %s = %q, want %q", c.id, got, c.want)
		}
	}
}

func TestDetailShowsPriorityOnlyWhenSet(t *testing.T) {
	t.Parallel()

	cfg := prioCfg(t)
	if got := labelsOf(detailLines(t, cfg, "bugs/2026-10-01-c")); !slices.Contains(got, "PRIORITY") {
		t.Errorf("high bug labels %v, want PRIORITY", got)
	}
	if got := labelsOf(detailLines(t, cfg, "bugs/2026-10-01-a")); slices.Contains(got, "PRIORITY") {
		t.Errorf("unset bug labels %v, want no PRIORITY", got)
	}
}

func TestByPriorityIsStableAndCopies(t *testing.T) {
	t.Parallel()

	mk := func(id, p string) *board.Item { return &board.Item{ID: id, Priority: p} }
	in := []*board.Item{mk("a", ""), mk("b", "low"), mk("c", "high"), mk("d", ""), mk("e", "high")}
	if got, want := itemIDs(byPriority(in)), []string{"c", "e", "b", "a", "d"}; !slices.Equal(got, want) {
		t.Fatalf("byPriority %v, want %v", got, want)
	}
	if got := itemIDs(in); !slices.Equal(got, []string{"a", "b", "c", "d", "e"}) {
		t.Fatalf("input changed to %v", got)
	}
}
```

`detailLines` (in `detail_test.go`) loads its own board from `cfg` and walks every tab, so it finds the bugs.

- [x] **Step 2: Run test to verify it fails**

Run: `scripts/test ./internal/tui -run 'Priority'`
Expected: FAIL to build: `undefined: byPriority`.

- [x] **Step 3: Write minimal implementation**

`internal/tui/order.go`, add:

```go
// byPriority moves high above medium above low above unset. The sort is
// stable, so inside one level the order from ordered stays. It returns a new
// slice so the caller's order is kept.
func byPriority(items []*board.Item) []*board.Item {
	out := append([]*board.Item(nil), items...)
	rank := func(it *board.Item) int {
		if i := slices.Index(board.Priorities, it.Priority); i >= 0 {
			return i
		}
		return len(board.Priorities)
	}
	sort.SliceStable(out, func(i, j int) bool { return rank(out[i]) < rank(out[j]) })
	return out
}
```

(add `"slices"` to the imports).

`internal/tui/sidebar.go`, `openRows`, right after `items := ordered(m.board.List(tab.kind, false), m.newest[paneList])`:

```go
	// Bugs and debt are where the user picks what to fix next, so the most
	// urgent ones come first.
	if tab.kind == board.KindBug || tab.kind == board.KindDebtItem {
		items = byPriority(items)
	}
```

`internal/tui/styles.go`: add `slotBrightRed = 9` to the slot constants. Add a field `priority map[string]lipgloss.Style` to `styles`, and in `newStyles`:

```go
		// High is the one level that should catch the eye, so it gets the
		// bright red. Low only needs to be there, so it is dim.
		priority: map[string]lipgloss.Style{
			"high":   lipgloss.NewStyle().Foreground(slot(slotBrightRed)),
			"medium": lipgloss.NewStyle().Foreground(slot(slotYellow)),
			"low":    lipgloss.NewStyle().Foreground(slot(slotDim)),
		},
```

`internal/tui/scroll.go`, `rowText`: change `head := lead + name + "  " + it.Title` to:

```go
	head := lead + name + "  " + priorityTag(it) + it.Title
```

and add:

```go
// priorityTag is the one letter a row shows for its priority, with a space
// after it. An item with no priority shows nothing, so its row stays as it was.
func priorityTag(it *board.Item) string {
	if it.Priority == "" {
		return ""
	}
	return strings.ToUpper(it.Priority[:1]) + " "
}
```

`styles.paintID`: replace the last line with:

```go
	after := text[i+len(name):]
	tag := "  " + priorityTag(it)
	if it.Priority != "" && strings.HasPrefix(after, tag) {
		letter := s.priority[it.Priority].Bold(base.GetBold()).Render(tag[2:3])
		return base.Render(text[:i]) + brush.Render(name) + base.Render("  ") + letter + base.Render(after[3:])
	}
	return base.Render(text[:i]) + brush.Render(name) + base.Render(after)
```

`internal/tui/detail.go`, in the header `fields` list, right after `{"STATUS", it.Status},`:

```go
		{"PRIORITY", it.Priority},
```

- [x] **Step 4: Run test to verify it passes**

Run: `scripts/test ./internal/tui`
Expected: PASS (whole package, so layout, width and color tests still hold with no priority set).

- [x] **Step 5: Commit**

```bash
gofmt -l internal/tui && go vet ./internal/tui
git add internal/tui/order.go internal/tui/sidebar.go internal/tui/scroll.go internal/tui/styles.go internal/tui/detail.go internal/tui/priority_test.go
git commit -m "TUI sorts bugs and debt by priority and shows the tag"
```
