---
parent: scratch/2026-09-29-fast-id-lookup
status: approved
id: PLAN-32
hash: q5u0
---
# Short Id Format Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Short ids become a 3-letter prefix plus 4 digits (`PLN-0030`), tasks and debt items get 2 digits (`PLN-0030.03`), hashes grow to 7 characters, and old ids keep working as aliases.

**Architecture:** The board reads both the old and the new format and always shows the new one. One `Canon` function turns any id into one lookup key, so old ids written in bodies, commits and memory still resolve. `acta id` writes new ids in the new format and rewrites old frontmatter (`id`, `hash`, `closes`) on every run.

**Tech Stack:** Go (stdlib only), `go test`, `gofmt`, `go vet`.

**Spec:** `.acta/specs/2026-09-29-short-id-format-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Prefixes: `PLN` (plan), `SPC` (spec), `BUG` (bug), `DBT` (debt), `SCR` (scratch). Old prefixes: `PLAN`, `SPEC`, `BUG`, `DEBT`, `SCRATCH`.
- Id: prefix, dash, 4 digits padded with zeros. `0000` is rejected. More than 4 digits is accepted only without a leading zero, and then gets a problem line.
- Hash: 7 characters, a lower-case letter first, then a-z or 0-9.
- Task and debt-item numbers show as 2 digits (`.03`). A task number that is not a plain number (`F1`) stays as written. The number is never written to a file.
- A plan with a task number of 100 or more, a debt file with an item number of 100 or more, or an id number over 4 digits gets a problem line.
- Hash lookup matches a unique prefix of at least 4 characters. Two matches mean not found.
- `acta id` rewrites frontmatter only. `parent:` and bodies are never touched.
- Comments in plain English, short words, say why. No marker tags.
- Before each commit: `gofmt -l .` prints nothing and `go vet ./...` passes.
- TDD: write the failing test first, run it, watch it fail, then write the code.

## File Map

| File | Job | Tasks |
|---|---|---|
| `internal/board/ids.go` | prefixes, `IsID`, `IsHash`, reading ids, aliases, `Canon`, padding | 1, 2 |
| `internal/board/board.go` | `Board.Get` | 2 |
| `internal/board/ids_test.go`, `board_test.go`, `closes_test.go`, `parse_test.go`, `scratch_test.go` | board tests and fixtures | 1, 2 |
| `internal/write/ids.go` | new ids, hashes, `AssignIDs`, `FixDuplicates` | 3, 4 |
| `internal/write/scratch.go`, `internal/write/ops.go` | ids for new scratch, bug, debt files | 3 |
| `internal/write/*_test.go` | write tests and fixtures | 3, 4 |
| `internal/tui/*_test.go`, `internal/cli/*_test.go`, `internal/hook/*_test.go` | fixtures, list row test | 5 |
| `plugin/skills/{land,brainstorm,plan,scratch}/SKILL.md`, `plugin/evals/*` | examples, 99-task rule | 6 |

## Waves

- Wave 1: Task 1, Task 6
- Wave 2: Task 2
- Wave 3: Task 3
- Wave 4: Task 4, Task 5

---

### Task 1: Board reads and shows the new id format

**Files:**
- Modify: `internal/board/ids.go` (`Prefix`, `IsHash`, `IsID`, `setIDs`, `aliasTask`, `aliasDebtItem`)
- Test: `internal/board/ids_test.go`; update fixtures in `internal/board/*_test.go` that break

**verify:** Every item the board loads shows its id in the new form, whatever form its file holds. List every input form checked (new id, old id, old 4-char hash, new 7-char hash, bad shapes, task and debt-item numbers, `F1`) and what `ShortID`, `Hash` and `Problems` hold for each. No bad shape ever becomes an id.

**Interfaces:**
- Produces:
  - `func Prefix(k Kind, plan bool) string` returns `PLN`, `SPC`, `BUG`, `DBT`, `SCR`.
  - `func OldPrefix(p string) string` maps `PLN` to `PLAN`, `SPC` to `SPEC`, `DBT` to `DEBT`, `SCR` to `SCRATCH`, `BUG` to `BUG`.
  - `func IsID(id, prefix string) bool`: new format only.
  - `func IsHash(s string) bool`: 7 characters only.
  - `func FormatID(prefix string, n int) string` returns `fmt.Sprintf("%s-%04d", prefix, n)`.
  - `Item.OldForm bool`: a new field, true when the file holds an old id or an old 4-char hash, so `acta id` knows to rewrite it.

- [x] **Step 1: Write the failing tests**

```go
func TestIsIDNewFormat(t *testing.T) {
	for id, want := range map[string]bool{
		"PLN-0030": true, "PLN-0001": true, "PLN-10000": true,
		"PLN-30": false, "PLN-0000": false, "PLN-00300": false, "PLAN-30": false,
		"PLN-": false, "PLN-00a1": false, "PLN-٠٠٣٠": false, "": false,
	} {
		if got := IsID(id, "PLN"); got != want {
			t.Errorf("IsID(%q) = %v, want %v", id, got, want)
		}
	}
}

func TestIsHashSevenChars(t *testing.T) {
	for h, want := range map[string]bool{
		"oxoqk2m": true, "a000000": true,
		"oxoq": false, "oxoqk2mz": false, "1xoqk2m": false, "Oxoqk2m": false, "oxoqk2é": false, "": false,
	} {
		if got := IsHash(h); got != want {
			t.Errorf("IsHash(%q) = %v, want %v", h, got, want)
		}
	}
}

func TestBoardShowsNewFormForOldFiles(t *testing.T) {
	b := boardWith(t, map[string]string{
		"plans/2026-09-21-a.md":  "---\nid: PLAN-12\nhash: k3f2\n---\n# A\n\n**Spec:** none (Bounded, approved in chat on 2026-09-21)\n\n### Task 3: Three\n- [ ] x\n\n### Task F1: Fix\n- [ ] y\n",
		"plans/2026-09-22-b.md":  "---\nid: PLN-0013\nhash: k3f2abc\n---\n# B\n\n**Spec:** none (Bounded, approved in chat on 2026-09-22)\n\n### Task 1: One\n- [ ] z\n",
		"scratch/2026-09-23-c.md": "---\nid: SCRATCH-4\nhash: m2x9\ntitle: C\nstatus: raw\n---\nbody\n",
	})
	cases := map[string]struct{ short, hash string; old bool }{
		"plans/2026-09-21-a":        {"PLN-0012", "PLN-k3f2", true},
		"plans/2026-09-21-a#task-3":  {"PLN-0012.03", "PLN-k3f2.03", true},
		"plans/2026-09-21-a#task-F1": {"PLN-0012.F1", "PLN-k3f2.F1", true},
		"plans/2026-09-22-b":        {"PLN-0013", "PLN-k3f2abc", false},
		"scratch/2026-09-23-c":      {"SCR-0004", "SCR-m2x9", true},
	}
	for id, want := range cases {
		it := b.Get(id)
		if it == nil {
			t.Fatalf("%s not loaded", id)
		}
		if it.ShortID != want.short || it.Hash != want.hash || it.OldForm != want.old {
			t.Errorf("%s: got %q %q old=%v, want %q %q old=%v", id, it.ShortID, it.Hash, it.OldForm, want.short, want.hash, want.old)
		}
		if len(it.Problems) != 0 {
			t.Errorf("%s: unexpected problems %v", id, it.Problems)
		}
	}
}

func TestBadIDShapesStayProblems(t *testing.T) {
	for _, raw := range []string{"PLN-30", "PLN-0000", "PLAN-030", "SPEC-4", "PLN-00300"} {
		b := boardWith(t, map[string]string{
			"plans/2026-09-21-a.md": "---\nid: " + raw + "\n---\n# A\n\n**Spec:** none (Bounded, approved in chat on 2026-09-21)\n",
		})
		it := b.Get("plans/2026-09-21-a")
		if it.ShortID != "" || !hasProblem(it, "bad id "+raw) {
			t.Errorf("%s: ShortID %q problems %v", raw, it.ShortID, it.Problems)
		}
	}
}

func TestOversizeNumbersGetProblems(t *testing.T) {
	var tasks strings.Builder
	tasks.WriteString("### Task 100: Big\n- [ ] x\n")
	b := boardWith(t, map[string]string{
		"plans/2026-09-21-a.md": "---\nid: PLN-10000\nhash: k3f2abc\n---\n# A\n\n**Spec:** none (Bounded, approved in chat on 2026-09-21)\n\n" + tasks.String(),
	})
	plan := b.Get("plans/2026-09-21-a")
	if plan.ShortID != "PLN-10000" || !hasProblem(plan, "id PLN-10000 is wider than 4 digits") {
		t.Errorf("plan: %q %v", plan.ShortID, plan.Problems)
	}
	if !hasProblem(plan, "task 100 is wider than 2 digits") {
		t.Errorf("plan: %v", plan.Problems)
	}
}
```

`hasProblem(it *Item, p string) bool` is a small test helper: true when `it.Problems` holds `p`. Add a debt file case with 100 items the same way, expecting `item 100 is wider than 2 digits` on the debt file.

- [x] **Step 2: Run the tests and watch them fail**

Run: `go test ./internal/board/ -run 'TestIsIDNewFormat|TestIsHashSevenChars|TestBoardShowsNewFormForOldFiles|TestBadIDShapesStayProblems|TestOversizeNumbersGetProblems' -v`
Expected: FAIL (old `IsID` accepts `PLN-30`, `IsHash` accepts 4 chars, `OldForm` does not exist).

- [x] **Step 3: Write the code**

```go
func Prefix(k Kind, plan bool) string {
	switch {
	case plan || k == KindPlan:
		return "PLN"
	case k == KindBug:
		return "BUG"
	case k == KindDebt || k == KindDebtItem:
		return "DBT"
	case k == KindScratch:
		return "SCR"
	default:
		return "SPC"
	}
}

// OldPrefix gives the prefix a kind had before ids were cut to 3 letters.
// Old files and old notes still use it.
func OldPrefix(p string) string {
	switch p {
	case "PLN":
		return "PLAN"
	case "SPC":
		return "SPEC"
	case "DBT":
		return "DEBT"
	case "SCR":
		return "SCRATCH"
	}
	return p
}

// FormatID pads the number to 4 digits so every id in a list is one width.
func FormatID(prefix string, n int) string { return fmt.Sprintf("%s-%04d", prefix, n) }

// IsHash says if s is a hash: 7 characters, a lower-case letter first, then
// lower-case letters or digits. The letter first keeps it apart from numbers.
func IsHash(s string) bool { return isHashLen(s, 7) }

func isHashLen(s string, n int) bool {
	if len(s) != n || s[0] < 'a' || s[0] > 'z' {
		return false
	}
	for _, c := range s[1:] {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

// IsID says if id is a new style short ID for prefix: the prefix, a dash,
// then 4 digits. A wider number is allowed only without a leading zero, so
// number 10000 can still be written down.
func IsID(id, prefix string) bool {
	n, ok := strings.CutPrefix(id, prefix+"-")
	if !ok || len(n) < 4 || strings.Trim(n, "0123456789") != "" || n == "0000" {
		return false
	}
	return len(n) == 4 || n[0] != '0'
}

// isOldID says if id is written the old way: the old prefix and a number
// with no leading zero.
func isOldID(id, prefix string) bool {
	n, ok := strings.CutPrefix(id, OldPrefix(prefix)+"-")
	return ok && n != "" && strings.Trim(n, "0123456789") == "" && n[0] != '0'
}
```

In `setIDs`: when `IsID(raw, prefix)` set `ShortID = raw`; else when `isOldID(raw, prefix)` parse the number, set `ShortID = FormatID(prefix, n)` and `OldForm = true`; else add `bad id`. For the hash: `IsHash` gives `Hash = prefix + "-" + raw`; `isHashLen(raw, 4)` gives the same plus `OldForm = true`; else `bad hash`. When the number part of `ShortID` is wider than 4, add `"id " + ShortID + " is wider than 4 digits"`. Reset `OldForm` to false at the top, next to `ShortID` and `Hash`.

In `aliasTask` and `aliasDebtItem`, add a `pad2(n string) string` helper: a plain number becomes `fmt.Sprintf("%02d", n)`, anything else (`F1`) stays. A number of 100 or more adds `"task N is wider than 2 digits"` to the plan, or `"item N is wider than 2 digits"` to the debt file. Both functions need the parent item to hang the problem on, so pass it in.

- [x] **Step 4: Run the package tests and fix fixtures**

Run: `go test ./internal/board/`
Expected: the new tests PASS. Older tests that assert `PLAN-12`, `SPEC-4` or 4-char hashes as the shown id now fail. Change their expected values to the new form. Keep their old-format input files, so they keep testing the old reading. Do not delete or skip any test.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./... && go test ./internal/board/
git add internal/board
git commit -m "feat(board): read old ids, show 3-letter prefix and fixed width"
```

---

### Task 2: One lookup key for every id form

**Files:**
- Modify: `internal/board/ids.go` (`addAlias`, new `Canon`), `internal/board/board.go:314` (`Get`)
- Test: `internal/board/ids_test.go`

**verify:** Every way a person or a note can write an id finds the same item, and nothing else ever finds it. List every form checked (new, old, lower case, zero padded, unpadded, task, debt item, `F1`, full hash, hash prefix, path id) and the forms that must find nothing (wrong prefix, empty, short or shared hash prefix).

**Interfaces:**
- Consumes: `OldPrefix`, `ShortID` and `Hash` in the new form from Task 1.
- Produces:
  - `func Canon(id string) string`: lower case, old prefix mapped to new, leading zeros dropped from each number part. `Canon("PLAN-30.03") == "pln-30.3"`.
  - `func (b *Board) Get(id string) *Item`: path id, then canon alias, then unique hash prefix.
  - `func (b *Board) Ambiguous(id string) bool`: true when `id` is a hash prefix two items share, so callers can say why nothing was found.

- [x] **Step 1: Write the failing tests**

```go
func TestCanon(t *testing.T) {
	for in, want := range map[string]string{
		"PLAN-30.3": "pln-30.3", "pln-0030.03": "pln-30.3", "PLN-30.3": "pln-30.3",
		"SCRATCH-14": "scr-14", "scr-0014": "scr-14", "DEBT-22.4": "dbt-22.4",
		"SPEC-4": "spc-4", "BUG-0007": "bug-7", "PLN-0030.F1": "pln-30.f1",
		"PLN-k3f2abc": "pln-k3f2abc", "": "",
	} {
		if got := Canon(in); got != want {
			t.Errorf("Canon(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGetFindsEveryForm(t *testing.T) {
	b := boardWith(t, map[string]string{
		"plans/2026-09-21-a.md": "---\nid: PLAN-12\nhash: k3f2abc\n---\n# A\n\n**Spec:** none (Bounded, approved in chat on 2026-09-21)\n\n### Task 3: Three\n- [ ] x\n\n### Task F1: Fix\n- [ ] y\n",
		"plans/2026-09-22-b.md": "---\nid: PLN-0013\nhash: k3f2xyz\n---\n# B\n\n**Spec:** none (Bounded, approved in chat on 2026-09-22)\n",
	})
	for _, id := range []string{"PLN-0012", "pln-0012", "PLAN-12", "PLN-12", "plans/2026-09-21-a", "PLN-k3f2abc", "k3f2a", "k3f2abc"} {
		if it := b.Get(id); it == nil || it.ID != "plans/2026-09-21-a" {
			t.Errorf("Get(%q) = %v", id, it)
		}
	}
	for _, id := range []string{"PLN-0012.03", "PLAN-12.3", "pln-12.03"} {
		if it := b.Get(id); it == nil || it.ID != "plans/2026-09-21-a#task-3" {
			t.Errorf("Get(%q) = %v", id, it)
		}
	}
	if it := b.Get("PLN-0012.F1"); it == nil || it.ID != "plans/2026-09-21-a#task-F1" {
		t.Errorf("Get F1 = %v", it)
	}
	for _, id := range []string{"PLX-0012", "", "k3f", "k3f2", "SPC-0012"} {
		if it := b.Get(id); it != nil {
			t.Errorf("Get(%q) = %s, want nil", id, it.ID)
		}
	}
	if !b.Ambiguous("k3f2") || b.Ambiguous("k3f2a") {
		t.Error("Ambiguous wrong for k3f2 / k3f2a")
	}
}
```

- [x] **Step 2: Run the tests and watch them fail**

Run: `go test ./internal/board/ -run 'TestCanon|TestGetFindsEveryForm' -v`
Expected: FAIL with `undefined: Canon`.

- [x] **Step 3: Write the code**

```go
// Canon turns any way of writing an id into one lookup key. Old prefixes
// map to new ones and zeros in front of a number drop, so PLAN-30.3 and
// PLN-0030.03 are the same key.
func Canon(id string) string {
	s := strings.ToLower(id)
	pre, rest, ok := strings.Cut(s, "-")
	if !ok {
		return s
	}
	for _, p := range []string{"PLN", "SPC", "BUG", "DBT", "SCR"} {
		if pre == strings.ToLower(OldPrefix(p)) {
			pre = strings.ToLower(p)
		}
	}
	parts := strings.Split(rest, ".")
	for i, p := range parts {
		if p != "" && strings.Trim(p, "0123456789") == "" {
			parts[i] = strings.TrimLeft(p, "0")
			if parts[i] == "" {
				parts[i] = "0"
			}
		}
	}
	return pre + "-" + strings.Join(parts, ".")
}
```

`addAlias` stores `Canon(key)` instead of `strings.ToLower(key)`. `Get` becomes: `byID[id]`, then `alias[Canon(id)]`, then a hash prefix scan. The scan takes `id` with any `xxx-` prefix removed, needs at least 4 characters, and walks every item whose `Hash` part after the dash (and before any dot) starts with it. Exactly one file item matches: return it. Zero or two or more: return nil. `Ambiguous` runs the same scan and says if two or more matched. Tasks and debt items are skipped in the scan, so one plan's hash does not match itself many times.

- [x] **Step 4: Run the package tests**

Run: `go test ./internal/board/`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./... && go test ./internal/board/
git add internal/board
git commit -m "feat(board): one lookup key for old and new ids, hash prefix match"
```

---

### Task 3: New ids and hashes are written in the new format

**Files:**
- Modify: `internal/write/ids.go` (`randHash`, `scanIDs`, `AssignIDs`, `FixDuplicates`, `filePrefix`, `prefixOf`), `internal/write/scratch.go:32`, `internal/write/ops.go:118`, `internal/write/ops.go:277`
- Test: `internal/write/ids_test.go`, `internal/write/scratch_test.go`, `internal/write/ops_test.go`, `internal/write/debt_test.go`

**verify:** No path that writes a new id or hash can write the old format. List every writer checked (`AssignIDs`, `FixDuplicates`, `acta scratch new`, `acta bug new`, the debt writer) and the id and hash each writes. Numbering continues past the highest number in either format.

**Interfaces:**
- Consumes: `board.Prefix`, `board.FormatID`, `board.IsID`, `board.IsHash` from Task 1.
- Produces: `randHash()` returns 7 characters; `scanIDs(b)` keys `next` by new prefix (`PLN`, `SPC`, `BUG`, `DBT`, `SCR`) and reads numbers from `ShortID`, which Task 1 already shows in the new form.

- [x] **Step 1: Write the failing tests**

```go
func TestRandHashIsSevenChars(t *testing.T) {
	for i := 0; i < 200; i++ {
		if h := randHash(); !board.IsHash(h) {
			t.Fatalf("randHash() = %q", h)
		}
	}
}

func TestAssignIDsWritesNewFormat(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		"plans/2026-09-21-a.md": "---\nid: PLAN-12\nhash: k3f2\n---\n# A\n\n**Spec:** none (Bounded, approved in chat on 2026-09-21)\n",
		"plans/2026-09-22-b.md": "# B\n\n**Spec:** none (Bounded, approved in chat on 2026-09-22)\n",
	})
	b := loadBoard(t, cfg)
	if _, _, err := AssignIDs(cfg, b, nil); err != nil {
		t.Fatal(err)
	}
	front := readFront(t, cfg, "plans/2026-09-22-b.md")
	if front["id"] != "PLN-0013" || !board.IsHash(front["hash"]) {
		t.Errorf("new plan got %v", front)
	}
}
```

Add the same check for `NewScratch` (expect `SCR-000N`), `NewBug` (expect `BUG-000N`) and the debt writer (expect `DBT-000N`), and for `FixDuplicates` with two files holding `PLN-0005` (expect the later one to become the next free `PLN-000N`). Reuse the helpers the existing `internal/write` tests already have for making a repo and reading frontmatter; if their names differ from `repoWith`, `loadBoard` and `readFront`, use the existing names.

- [x] **Step 2: Run the tests and watch them fail**

Run: `go test ./internal/write/ -run 'TestRandHashIsSevenChars|TestAssignIDsWritesNewFormat|Scratch|Bug|Debt|Duplicate' -v`
Expected: FAIL (4-char hash, `PLAN-13`).

- [x] **Step 3: Write the code**

- `randHash`: `var b [7]byte`.
- `scanIDs`: `next := map[string]int{"SPC": 1, "PLN": 1, "BUG": 1, "DBT": 1, "SCR": 1}`. Parse the number with `strconv.Atoi` after the dash; `Atoi("0012")` gives 12.
- Every `fmt.Sprintf("%s-%d", …)` and `fmt.Sprintf("SCRATCH-%d", …)`, `"BUG-%d"`, `"DEBT-%d"` becomes `board.FormatID(prefix, n)` with the new prefix.
- `filePrefix` and `prefixOf`: `"PLAN"` becomes `"PLN"`.
- `FixDuplicates` groups by `board.Canon(f.id)`, so `PLAN-5` and `PLN-0005` count as one number.

- [x] **Step 4: Run the package tests and fix fixtures**

Run: `go test ./internal/write/`
Expected: new tests PASS. Fix older tests that expect the old output format by changing the expected value, never by deleting the test.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./... && go test ./internal/write/
git add internal/write
git commit -m "feat(write): write new ids as 3-letter prefix, 4 digits, 7-char hash"
```

---

### Task 4: acta id rewrites old frontmatter

**Files:**
- Modify: `internal/write/ids.go` (`candidates`, `AssignIDs`, `idFile`)
- Test: `internal/write/ids_test.go`

**verify:** After one `acta id` run no file on the main tree holds an old id, an old hash or an old id in `closes:`, and a second run changes no byte. List every field checked, and show that `parent:`, the body and other frontmatter keys come out byte for byte the same.

**Interfaces:**
- Consumes: `Item.OldForm`, `board.Canon`, `board.FormatID`, `board.IsHash` from Tasks 1 to 3.
- Produces: nothing new for other tasks. `AssignIDs` keeps its signature.

- [x] **Step 1: Write the failing tests**

```go
func TestAssignIDsRewritesOldFiles(t *testing.T) {
	old := "---\nparent: scratch/2026-09-20-x\nid: PLAN-12\nhash: k3f2\ncloses: [DEBT-17.1, SCRATCH-14]\nstatus: approved\n---\n# A\n\nSee PLAN-12 and SCRATCH-14.\n\n**Spec:** none (Bounded, approved in chat on 2026-09-21)\n"
	cfg := repoWith(t, map[string]string{
		"plans/2026-09-21-a.md":   old,
		"scratch/2026-09-20-x.md": "---\nid: SCRATCH-14\nhash: m2x9abc\ntitle: X\nstatus: raw\n---\nx\n",
		"debt/2026-09-19-d.md":    "---\nid: DBT-0017\nhash: d4d4abc\n---\n# D\n\n- [ ] one\n",
	})
	if _, _, err := AssignIDs(cfg, loadBoard(t, cfg), nil); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, cfg, "plans/2026-09-21-a.md")
	front := readFront(t, cfg, "plans/2026-09-21-a.md")
	if front["id"] != "PLN-0012" || !strings.HasPrefix(front["hash"], "k3f2") || !board.IsHash(front["hash"]) {
		t.Errorf("front %v", front)
	}
	if !strings.Contains(got, "closes: [DBT-0017.01, SCR-0014]") {
		t.Errorf("closes not rewritten:\n%s", got)
	}
	for _, keep := range []string{"parent: scratch/2026-09-20-x\n", "status: approved\n", "See PLAN-12 and SCRATCH-14.\n"} {
		if !strings.Contains(got, keep) {
			t.Errorf("lost %q", keep)
		}
	}
	// A second run finds nothing old and leaves every byte alone.
	changes, _, err := AssignIDs(cfg, loadBoard(t, cfg), nil)
	if err != nil || len(changes) != 0 || readFile(t, cfg, "plans/2026-09-21-a.md") != got {
		t.Errorf("second run changed files: %v %v", changes, err)
	}
}

func TestAssignIDsLeavesNewFilesAlone(t *testing.T) {
	src := "---\nid: PLN-0012\nhash: k3f2abc\ncloses: [SCR-0014]\n---\n# A\n\n**Spec:** none (Bounded, approved in chat on 2026-09-21)\n"
	cfg := repoWith(t, map[string]string{
		"plans/2026-09-21-a.md":   src,
		"scratch/2026-09-20-x.md": "---\nid: SCR-0014\nhash: m2x9abc\ntitle: X\nstatus: raw\n---\nx\n",
	})
	changes, _, err := AssignIDs(cfg, loadBoard(t, cfg), nil)
	if err != nil || len(changes) != 0 || readFile(t, cfg, "plans/2026-09-21-a.md") != src {
		t.Errorf("new file touched: %v %v", changes, err)
	}
}
```

Add one more case: a `closes:` entry that finds no item stays as written, and the run still rewrites the rest.

- [x] **Step 2: Run the tests and watch them fail**

Run: `go test ./internal/write/ -run 'TestAssignIDsRewritesOldFiles|TestAssignIDsLeavesNewFilesAlone' -v`
Expected: FAIL (old id and hash are kept today, since a value already written is never changed).

- [x] **Step 3: Write the code**

- `candidates`: a file whose item has `OldForm` true, or whose `closes:` holds an entry where `board.Canon(entry)` resolves to an item whose `ShortID` differs from the entry, is a candidate even when it has both an id and a hash.
- `AssignIDs`: for an old id, write `it.ShortID` (already new form). For a 4-char hash, write the old hash plus 3 random characters, and loop until `taken` does not hold it, the same way `freeHash` does. For `closes:`, map each entry through `b.Get(entry)` and write that item's `ShortID`; an entry with no match or no `ShortID` stays as written. Write the list back with `SetField` in the same `[a, b]` shape it was read in.
- The commit message stays `acta: assign short ids` when only new ids are written, and becomes `acta: migrate ids to 3-letter prefix` when any old field was rewritten.
- Update the `AssignIDs` doc comment: it now rewrites old ids, and still never replaces a value in the new format.

- [x] **Step 4: Run the package tests**

Run: `go test ./internal/write/`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./... && go test ./internal/write/
git add internal/write
git commit -m "feat(write): acta id rewrites old ids, hashes and closes lists"
```

---

### Task 5: List row test and fixture sweep in tui, cli and hook

**Files:**
- Test: `internal/tui/view_test.go`; update fixtures in `internal/tui/*_test.go`, `internal/cli/*_test.go`, `internal/hook/*_test.go` that break

**verify:** Every list row across kinds starts its title at the same column. List every kind checked (plan, spec, bug, debt, scratch) with a one-digit and a four-digit number each. The whole suite passes, and no test was deleted or skipped.

**Interfaces:**
- Consumes: new `ShortID` form from Task 1; new writers from Tasks 3 and 4.
- Produces: nothing.

- [x] **Step 1: Write the failing test**

```go
func TestListRowsLineUp(t *testing.T) {
	m := modelWith(t, map[string]string{
		"scratch/2026-09-20-a.md": "---\nid: SCRATCH-3\nhash: aaaa\ntitle: Three\nstatus: raw\n---\nx\n",
		"scratch/2026-09-21-b.md": "---\nid: SCRATCH-13\nhash: bbbb\ntitle: Thirteen\nstatus: raw\n---\nx\n",
		"scratch/2026-09-22-c.md": "---\nid: SCR-1234\nhash: ccccccc\ntitle: Big\nstatus: raw\n---\nx\n",
	})
	col := -1
	for _, id := range []string{"scratch/2026-09-20-a", "scratch/2026-09-21-b", "scratch/2026-09-22-c"} {
		it := m.board.Get(id)
		row := m.rowText(row{id: id}, it, 80)
		at := strings.Index(row, it.Title)
		if col == -1 {
			col = at
		}
		if at != col || !strings.HasPrefix(row, it.ShortID+"  ") {
			t.Errorf("%s row %q: title at %d, want %d", id, row, at, col)
		}
	}
}
```

Run the same check for plans, specs, bugs and debt. Use the existing model helper from `internal/tui` tests if it is named differently from `modelWith`.

- [x] **Step 2: Run the test**

Run: `go test ./internal/tui/ -run TestListRowsLineUp -v`
Expected: PASS once Tasks 1 to 4 have landed on this branch. To see it red first, run it on the commit before Task 1 (`git stash; git checkout <base>`) or trust the old output: `SCRATCH-3  Three` and `SCRATCH-13  Thirteen` put the title at different columns. Record which you did.

- [x] **Step 3: Fix every failing test in the suite**

Run: `go test ./...`
For each failure that expects an old shown id, an old written id or a 4-char hash, change the expected value to the new form. Keep old-format input files where they already exist, so the alias path stays tested. Never delete, skip or loosen a test. A failure that is not about the id format is a real bug: stop and report it.

- [x] **Step 4: Run the full suite**

Run: `gofmt -l . && go vet ./... && go test ./...`
Expected: all PASS. Paste the package summary.

- [x] **Step 5: Commit**

```bash
git add internal
git commit -m "test: list rows line up; fixtures move to the new id format"
```

---

### Task 6: Skills and evals use the new format

**Files:**
- Modify: `plugin/skills/land/SKILL.md`, `plugin/skills/brainstorm/SKILL.md`, `plugin/skills/plan/SKILL.md`, `plugin/skills/scratch/SKILL.md`
- Modify: `plugin/evals/second-brainstorm-choices/graders/choices.md`, `plugin/evals/answers-appended/scaffold.sh`, `plugin/evals/answers-appended/prompt.md`

**verify:** No skill or eval text teaches the old id format as the one to write. List every old-format string found by `grep -rnE '\b(PLAN|SPEC|DEBT|SCRATCH)-[0-9n]' plugin` and what each became. No grader got looser: each grader still checks the same behaviour.

**Interfaces:**
- Consumes: nothing (text only).
- Produces: nothing.

- [x] **Step 1: Write the failing check**

Run: `grep -rnE '\b(PLAN|SPEC|DEBT|SCRATCH)-[0-9n]' plugin`
Expected: matches in the 7 files above. This list is the red test.

- [x] **Step 2: Change the text**

- `SCRATCH-n` becomes `SCR-n`, `PLAN-n` becomes `PLN-n`, `SPEC-n` becomes `SPC-n`, `DEBT-17.1` becomes `DBT-0017.01`, and so on. Concrete numbers get 4 digits.
- `plugin/skills/plan/SKILL.md`, section "Task Right-Sizing", add: "A plan holds 99 tasks at most, so task ids stay two digits wide. Work bigger than that is several plans."
- `plugin/skills/land/SKILL.md`, where it runs `acta id` after the merge, add: "`acta id` also rewrites ids still in the old format (`PLAN-30` becomes `PLN-0030`)."
- In the eval scaffold and prompt, change the ids the fixtures write, and change the grader only where it names an id string. The rule it checks stays the same.

- [x] **Step 3: Run the check again**

Run: `grep -rnE '\b(PLAN|SPEC|DEBT|SCRATCH)-[0-9n]' plugin`
Expected: no output, or only lines that on purpose mention the old format as an alias (the land sentence above).

Run: `go test ./internal/plugincheck/`
Expected: PASS.

- [x] **Step 4: Commit**

```bash
git add plugin
git commit -m "docs(plugin): skills and evals use the 3-letter id format, plan caps 99 tasks"
```

## Fix round 1

### Task F1: Merge with PLAN-31 turns four write tests red

**Files:**
- Modify: `internal/write/ids.go`, `internal/write/scratch.go` only if a test proves a real bug
- Test: `internal/write/ids_test.go`, `internal/write/scratch_test.go`

**verify:** Every test PLAN-31 added still checks its own rule after the merge, with fixtures in the new id format where the rule is not about the old format. No test is deleted, skipped or loosened. Name each of the four tests and say whether it was a stale fixture or a real bug.

- [x] **Step 1: Reproduce** `go test ./internal/write/ -run 'TestAssignIDsLeavesAFileThatHasAnIDAlone|TestAssignIDsSkipsASchemaFileThatFails|TestAssignIDsFinishesTheScratchParent|TestAppendScratchOldItem' -v` fails.
- [x] **Step 2: Fix** stale fixtures by moving them to the new format; fix code only where a test shows wrong behaviour.
- [x] **Step 3: Run** `go test ./internal/write/ ./internal/board/`, `gofmt -l .`, `go vet ./...`.
- [x] **Step 4: Commit** `fix(write): PLAN-31 tests follow the 3-letter id format after merge`.
