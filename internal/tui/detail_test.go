package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// One board with every kind on it: a spec with a plan, a bug with a plan, a
// plan whose tasks are done, under way and not begun, a task with boxes and a
// task without any, and a debt file with two lines and prose around them.
func boardFiles() map[string]string {
	return map[string]string{
		".acta/specs/2026-09-20-a.md": "---\nid: SPEC-1\n---\n# Spec A\n\nWhy the work matters.\n",
		".acta/plans/2026-09-21-a.md": "---\nid: PLAN-1\n---\n# Plan A\n\n**Spec:** `.acta/specs/2026-09-20-a.md`\n\n### Task 1: First\n- [x] a\n\n### Task 2: Second\n- [x] b\n- [ ] c\n\n### Task 3: Third\n",
		".acta/specs/2026-09-22-e.md": "---\nid: SPEC-2\n---\n# Spec E\n",
		".acta/bugs/2026-09-23-b.md":  "---\nid: BUG-1\n---\n# Bug B\n",
		".acta/plans/2026-09-23-c.md": "---\nid: PLAN-2\nparent: bugs/2026-09-23-b\n---\n# Plan C\n\n### Task 1: Fix one\n- [ ] x\n",
		".acta/debt/2026-09-24-d.md":  "---\nid: DEBT-1\n---\n# Review NOTEs\n\nProse about the review.\n\n- [ ] first note\n- [x] second note\n",
		".acta/.agents.json":          "{\"plans/2026-09-21-a#task-2\": {\"agent\": \"omp\", \"at\": \"2026-09-27T10:00:00Z\"}}\n",
	}
}

// detailLines gives the lines the detail box draws for one item of the board built
// from cfg, with the color codes still on them.
func detailLines(t *testing.T, cfg config.Config, id string) []string {
	t.Helper()
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	it := b.Get(id)
	if it == nil {
		t.Fatalf("the board holds no %s", id)
	}
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	// A finished item sits in a Done pane and a task sits under its plan, so
	// every tab, both panes, every Done sub-tab and every open plan is walked.
	m.openPlans = everyPlanOpen(b)
	for i := range topTabs {
		m.openTab(i)
		for _, p := range m.panes() {
			for d := range max(1, len(topTabs[i].done)) {
				m.done = d
				m.focus = p
				rows, sel, idx := m.slotOf(p)
				for _, r := range rows {
					if r.id == it.ID {
						*sel, *idx = it.ID, 0
						return m.detailLines(100)
					}
				}
			}
		}
	}
	t.Fatalf("%s is on no list of the board", id)
	return nil
}

// detailOf is the detail of one item on a board that is not in git.
func detailOf(t *testing.T, id string) []string {
	t.Helper()
	return detailLines(t, treeCfg(t, boardFiles()), id)
}

func plainLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		out = append(out, strings.TrimRight(plain(ln), " "))
	}
	return out
}

// headerField is one "LABEL: value" line of the header, which is all the
// lines above the rule under it.
var headerField = regexp.MustCompile(`^([A-Z]+)\s*: (.*)$`)

// labelsOf are the field labels of the header, in the order it draws them.
func labelsOf(lines []string) []string {
	var out []string
	for _, ln := range plainLines(lines) {
		if strings.Trim(ln, "─ ") == "" {
			break
		}
		if m := headerField.FindStringSubmatch(ln); m != nil {
			out = append(out, m[1])
		}
	}
	return out
}

// valueOf is what one header field says.
func valueOf(t *testing.T, lines []string, label string) string {
	t.Helper()
	labels := labelsOf(lines)
	at := slices.Index(labels, label)
	if at < 0 {
		t.Fatalf("the detail holds no %s line:\n%s", label, strings.Join(plainLines(lines), "\n"))
	}
	return headerField.FindStringSubmatch(plainLines(lines)[at])[2]
}

// wantInOrder checks that every wanted line is drawn, whole, in the order
// given, which is how a test names what the reader sees.
func wantInOrder(t *testing.T, lines []string, want ...string) {
	t.Helper()
	at := 0
	for _, ln := range plainLines(lines) {
		if at < len(want) && ln == want[at] {
			at++
		}
	}
	if at < len(want) {
		t.Errorf("the detail is missing %q, got:\n%s", want[at], strings.Join(plainLines(lines), "\n"))
	}
}

// lineOf is where the drawn line that holds want sits, so a test can read the
// brush it wears.
func lineOf(t *testing.T, lines []string, want string) int {
	t.Helper()
	for i, ln := range plainLines(lines) {
		if strings.Contains(ln, want) {
			return i
		}
	}
	t.Fatalf("no line holds %q:\n%s", want, strings.Join(plainLines(lines), "\n"))
	return 0
}

// A task counts its steps, so it labels the line SUBTASKS. Every other kind
// counts the tasks under it and keeps TASKS.
func TestDetailLabelsSubtasksOnATask(t *testing.T) {
	for _, c := range []struct{ id, label string }{
		{"PLAN-1", "TASKS"},
		{"SPEC-1", "TASKS"},
		{"BUG-1", "TASKS"},
		{"PLAN-1.2", "SUBTASKS"},
	} {
		got := labelsOf(detailOf(t, c.id))
		if !slices.Contains(got, c.label) {
			t.Errorf("%s: labels %v, want one of them to be %s", c.id, got, c.label)
		}
		if other := "TASKS"; c.label != other && slices.Contains(got, other) {
			t.Errorf("%s: labels %v, want no %s", c.id, got, other)
		}
	}
	// A debt item counts nothing of its own, so its line is left out, but it
	// must never read SUBTASKS.
	if got := labelsOf(detailOf(t, "DEBT-1.1")); slices.Contains(got, "SUBTASKS") {
		t.Errorf("a debt item labels its line SUBTASKS, got %v", got)
	}
}

// The author sits right under the status, comes from the file itself, and is
// left out when git has no name to give.
func TestDetailShowsTheAuthorUnderTheStatus(t *testing.T) {
	lines := detailLines(t, gitTree(t, "Ana", boardFiles()), "PLAN-1")
	labels := labelsOf(lines)
	at := slices.Index(labels, "AUTHOR")
	if at < 0 {
		t.Fatalf("the detail holds no AUTHOR line:\n%s", strings.Join(plainLines(lines), "\n"))
	}
	if got := valueOf(t, lines, "AUTHOR"); got != "Ana" {
		t.Errorf("AUTHOR = %q, want Ana", got)
	}
	if want := slices.Index(labels, "STATUS") + 1; at != want {
		t.Errorf("AUTHOR sits under %q, want it right under STATUS:\n%s", labels[want], strings.Join(plainLines(lines), "\n"))
	}
	if got := labelsOf(detailOf(t, "PLAN-1")); slices.Contains(got, "AUTHOR") {
		t.Errorf("an empty author should leave the line out, got %v", got)
	}
}

// A plan lists its tasks in file order, each with a dot that says how far it
// has come. Only the work under way carries its count and its agent.
func TestDetailPlanListsItsTasksWithDots(t *testing.T) {
	withColors(func() {
		lines := detailOf(t, "PLAN-1")
		wantInOrder(t, lines,
			"✓ PLAN-1.1  First",
			"● PLAN-1.2  Second  1/2 · omp",
			"○ PLAN-1.3  Third")
		if got := plainLines(lines)[lineOf(t, lines, "PLAN-1.2")]; !strings.HasSuffix(got, "1/2 · omp") {
			t.Errorf("the task under way ends on %q, want its 1/2 and its agent", got)
		}
		for _, id := range []string{"PLAN-1.1", "PLAN-1.3"} {
			ln := lines[lineOf(t, lines, id)]
			if !wears(ln, 2) {
				t.Errorf("the line of %s is not dim: %q", id, plain(ln))
			}
			if strings.ContainsAny(plain(ln), "/·") {
				t.Errorf("the line of %s carries a count or an agent: %q", id, plain(ln))
			}
		}
		if i := lineOf(t, lines, "PLAN-1.2"); !wears(lines[i], 38, 5, 111) {
			t.Errorf("the line of the work under way wears no accent: %q", lines[i])
		}
	})
}

// A task lists its own steps with no id in front of them, and the first box
// still open in a task under way is the step being worked on.
func TestDetailTaskListsItsSteps(t *testing.T) {
	withColors(func() {
		lines := detailOf(t, "PLAN-1.2")
		wantInOrder(t, lines, "✓ b", "● c")
		if i := lineOf(t, lines, "● c"); !wears(lines[i], 38, 5, 111) {
			t.Errorf("the step under way wears no accent: %q", lines[i])
		}
		if i := lineOf(t, lines, "✓ b"); !wears(lines[i], 2) {
			t.Errorf("a ticked step is not dim: %q", lines[i])
		}
	})
	// A task with no box at all lists no step and leaves no empty line.
	for _, ln := range detailOf(t, "PLAN-1.3") {
		if strings.HasPrefix(strings.TrimSpace(plain(ln)), "○") {
			t.Errorf("a task with no step listed one: %q", plain(ln))
		}
	}
}

// A spec lists the tasks of every plan under it, one plain line naming each
// plan, and a bug the tasks of the plans that name it as their parent.
func TestDetailSpecAndBugListTheTasksOfTheirPlans(t *testing.T) {
	wantInOrder(t, detailOf(t, "SPEC-1"),
		"PLAN-1  Plan A",
		"✓ PLAN-1.1  First",
		"● PLAN-1.2  Second  1/2 · omp",
		"○ PLAN-1.3  Third")
	wantInOrder(t, detailOf(t, "BUG-1"),
		"PLAN-2  Plan C",
		"○ PLAN-2.1  Fix one")
	// A spec with no plan shows its header and no list, not an empty line.
	lines := detailOf(t, "SPEC-2")
	if len(plainLines(lines)) < 3 || !strings.HasPrefix(plainLines(lines)[0], "ID") {
		t.Errorf("a spec with no plan should still show its header:\n%s", strings.Join(plainLines(lines), "\n"))
	}
	if joined := strings.Join(plainLines(lines), "\n"); strings.ContainsAny(joined, "○●✓") {
		t.Errorf("a spec with no plan lists something:\n%s", joined)
	}
}

// A debt item lists every line of its debt file, and the line on show is the
// bright one while the others are dim.
func TestDetailDebtItemListsEveryLineOfItsFile(t *testing.T) {
	withColors(func() {
		lines := detailOf(t, "DEBT-1.1")
		wantInOrder(t, lines, "○ DEBT-1.1  first note", "✓ DEBT-1.2  second note")
		mine, other := lineOf(t, lines, "DEBT-1.1  first note"), lineOf(t, lines, "DEBT-1.2  second note")
		if wears(lines[mine], 2) {
			t.Errorf("the line on show is dim: %q", lines[mine])
		}
		if !wears(lines[other], 2) {
			t.Errorf("the other line is not dim: %q", lines[other])
		}
	})
}

// The text of a debt file around its checklist renders under the list, once,
// and not as the whole file again.
func TestDetailDebtItemShowsTheTextAroundItsLine(t *testing.T) {
	lines := plainLines(detailOf(t, "DEBT-1.1"))
	below := strings.Join(lines[ruleLine(t, lines):], "\n")
	if !strings.Contains(below, "Prose about the review.") {
		t.Errorf("the debt detail holds no text of its file:\n%s", below)
	}
	if n := strings.Count(below, "first note"); n != 1 {
		t.Errorf("the debt detail shows its checklist line %d times under the header, want once:\n%s", n, below)
	}
}

// ruleLine is where the header ends and the list and body begin.
func ruleLine(t *testing.T, lines []string) int {
	t.Helper()
	for i, ln := range lines {
		if strings.Trim(ln, "─ ") == "" {
			return i
		}
	}
	t.Fatalf("the detail holds no rule under its header:\n%s", strings.Join(lines, "\n"))
	return 0
}

// Every kind of item draws its header, so no detail pane is empty while the
// file behind it has something to say.
func TestDetailIsNeverEmpty(t *testing.T) {
	for _, id := range []string{"SPEC-1", "SPEC-2", "PLAN-1", "BUG-1", "PLAN-1.1", "PLAN-1.3", "DEBT-1.1", "DEBT-1.2"} {
		lines := detailOf(t, id)
		if len(lines) < 3 {
			t.Errorf("%s: the detail pane is empty", id)
			continue
		}
		if head := plainLines(lines)[0]; !strings.HasPrefix(head, "ID") {
			t.Errorf("%s: the detail starts on %q, want its ID line", id, head)
		}
		if blank := plainLines(lines)[2]; strings.TrimSpace(blank) == "" {
			t.Errorf("%s: the detail holds a blank line where the header ends:\n%s", id, strings.Join(plainLines(lines), "\n"))
		}
	}
}

// gitTree writes the files into a fresh checkout and commits them as name, so
// the detail shows the name the first commit of the file carries.
func gitTree(t *testing.T, name string, files map[string]string) config.Config {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME="+name, "GIT_COMMITTER_NAME="+name,
			"GIT_AUTHOR_EMAIL="+name+"@example.com", "GIT_COMMITTER_EMAIL="+name+"@example.com")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	run("init", "-q", "-b", "main")
	for rel, body := range files {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	run("add", ".")
	run("commit", "-q", "-m", "add the board")
	return config.Default(dir)
}

// A tab fills the room up to the next stop of eight, counted with the width
// of the words before it on the same line.
func TestExpandTabsToNextStop(t *testing.T) {
	cases := []struct{ in, want string }{
		{"\tx", "        x"},
		{"ab\tx", "ab      x"},
		{"\t\tx", "                x"},
		{"12345678\tx", "12345678        x"},
		{"a\nb\tc", "a\nb       c"},
		{"界\tx", "界      x"},
		{"no tabs", "no tabs"},
		{"", ""},
	}
	for _, c := range cases {
		if got := expandTabs(c.in); got != c.want {
			t.Errorf("expandTabs(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// One board whose files carry a tab in every place the detail pane draws its
// own words: a plan body with Go code indented by tabs, the steps of a task,
// the line of a debt file, and the prose around it. The titles on the lists
// the panes are showing keep their tabs out, because a row is drawn by
// another pane.
func tabFiles() map[string]string {
	return map[string]string{
		".acta/specs/2026-09-28-t.md": "---\nid: SPEC-1\n---\n# Spec T\n\nWhy the tabs matter.\n",
		".acta/plans/2026-09-28-p.md": "---\nid: PLAN-1\n---\n# Plan P\n\n**Spec:** `.acta/specs/2026-09-28-t.md`\n\n## Code\n\n```go\nfunc main() {\n\tif true {\n\t\tfmt.Println(\"hi\")\n\t}\n}\n```\n\n" +
			strings.Repeat("A line of prose long enough to wrap on its own, so the body has more lines than the pane has rows.\n\n", 20) +
			"### Task 1: Fix the frame\n\n- [x] a step\tthat runs on and on to the edge of the pane\n- [ ] step\ttwo\n",
		// The line runs on, so cutting it to the width of a narrow pane
		// leaves the tab in the line that is drawn.
		".acta/debt/2026-09-28-d.md": "---\nid: DEBT-1\n---\n# Review NOTEs\n\n- [ ] a note\tthat runs on to the edge of the pane\n\nProse\twith a tab around the list.\n",
	}
}

// detailModel is a model over the board of cfg with the real markdown
// renderer, because that is the path a body with tabs really takes.
func detailModel(t *testing.T, cfg config.Config) Model {
	t.Helper()
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return New(cfg, b, false)
}

// onItem walks every tab, every pane and every row until the item is the one
// the screen has selected, so the walk holds whatever shape the bar has, and
// leaves the detail box on that item.
func onItem(t *testing.T, m Model, id string) Model {
	t.Helper()
	it := m.board.Get(id)
	if it == nil {
		t.Fatalf("the board holds no %s", id)
	}
	for i := range topTabs {
		m = press(m, tabKey(i))
		m.openPlans = everyPlanOpen(m.board)
		for _, p := range m.panes() {
			m.focusPane(p)
			for d := range max(1, len(topTabs[i].done)) {
				m.done = d
				for n, r := range m.rowsOf(p) {
					if r.id == it.ID {
						m.moveTo(n)
						m.focus = paneDetail
						return press(m, "g")
					}
				}
			}
		}
	}
	t.Fatalf("no pane of the board lists %s", id)
	return m
}

// frameBreak is the first line of a screen that holds a tab or outgrows the
// terminal, or -1 when the screen is whole. lipgloss counts a tab as nothing
// while the terminal draws it up to the next stop, so a line holding one is a
// line the pane and the screen disagree about.
func frameBreak(view string, w int) int {
	for i, ln := range strings.Split(view, "\n") {
		if strings.Contains(ln, "\t") || lipgloss.Width(ln) > w {
			return i
		}
	}
	return -1
}

// A plan read in the detail pane and scrolled down line by line to the end
// draws no tab and no line wider than the terminal, at both sizes. The frame
// breaks otherwise: the screen counts the cells the pane never drew, and the
// old words stay under the new ones.
func TestDetailWithTabsFitsThePane(t *testing.T) {
	cfg := treeCfg(t, tabFiles())
	for _, size := range [][2]int{{80, 30}, {160, 50}} {
		w, h := size[0], size[1]
		for _, id := range []string{"PLAN-1", "PLAN-1.1"} {
			m := onItem(t, sized(detailModel(t, cfg), w, h), id)
			for range 400 {
				view := m.View()
				if at := frameBreak(view, w); at >= 0 {
					t.Errorf("%s at %dx%d, scrolled to line %d: the frame breaks at line %d:\n%s",
						id, w, h, m.off[paneDetail], at, view)
					break
				}
				off := m.off[paneDetail]
				m = press(m, "j")
				if m.off[paneDetail] == off {
					break
				}
			}
		}
	}
}

// Every item of that board, at every width, not only the ones a view happened
// to show: the header, the work list and the body all end in the same lines,
// and none of them may hold a tab or outgrow the pane it is measured for.
func TestEveryDetailItemHoldsNoTabAndFitsItsWidth(t *testing.T) {
	files := tabFiles()
	// A second plan whose own title holds a tab and whose frontmatter names a
	// kind that does not exist, so the line a spec draws above the tasks of a
	// plan and the warning under the header both carry one. It sits under no
	// pane this test renders, because a row is drawn by the list panes.
	files[".acta/plans/2026-09-28-q.md"] = "---\nid: PLAN-2\ntype: \"a\tb\"\n---\n# a plan\twhose title runs to the edge of the pane\n\n**Spec:** `.acta/specs/2026-09-28-t.md`\n"
	cfg := treeCfg(t, files)
	for _, id := range []string{"SPEC-1", "PLAN-1", "PLAN-1.1", "PLAN-2", "DEBT-1.1"} {
		m := onItem(t, detailModel(t, cfg), id)
		for _, w := range []int{40, 60, 80, 120, 160} {
			for i, ln := range m.detailLines(w) {
				if strings.Contains(plain(ln), "\t") {
					t.Errorf("%s at %d wide: line %d holds a tab: %q", id, w, i, plain(ln))
				}
				if got := lipgloss.Width(ln); got > w {
					t.Errorf("%s at %d wide: line %d is %d cells wide: %q", id, w, i, got, plain(ln))
				}
			}
		}
	}
}

// metaOf is the value of one header line named by its label, a label with a
// space in it like CLOSED BY included, which the header pattern leaves out.
func metaOf(t *testing.T, lines []string, label string) (string, bool) {
	t.Helper()
	for _, ln := range plainLines(lines) {
		head, value, ok := strings.Cut(ln, ":")
		if ok && strings.TrimSpace(head) == label {
			return strings.TrimSpace(value), true
		}
	}
	return "", false
}

// One spec that closes a scratch item, that scratch item, and a spec that
// closes nothing, so the detail can be read on an item with closes, on one
// with closed by and on one with neither.
func closesFiles() map[string]string {
	return map[string]string{
		".acta/specs/2026-09-29-a.md":   "---\nid: SPEC-1\ncloses: [SCRATCH-1, scratch/2026-09-28-j]\n---\n# Spec A\n",
		".acta/scratch/2026-09-28-i.md": "---\nid: SCRATCH-1\n---\n# Idea\n",
		// No short id in the frontmatter, so the detail names this one by its
		// path, the same way the closes list wrote it.
		".acta/scratch/2026-09-28-j.md": "# Another idea\n",
		".acta/specs/2026-09-29-b.md":   "---\nid: SPEC-2\n---\n# Spec B\n",
	}
}

func TestDetailShowsClosesAndClosedBy(t *testing.T) {
	cfg := treeCfg(t, closesFiles())
	for _, c := range []struct{ id, label, value, absent string }{
		{"SPEC-1", "CLOSES", "SCRATCH-1, scratch/2026-09-28-j", "CLOSED BY"},
		{"SCRATCH-1", "CLOSED BY", "SPEC-1", "CLOSES"},
		{"scratch/2026-09-28-j", "CLOSED BY", "SPEC-1", "CLOSES"},
	} {
		lines := detailLines(t, cfg, c.id)
		got, ok := metaOf(t, lines, c.label)
		if !ok {
			t.Errorf("%s drew no %s line:\n%s", c.id, c.label, strings.Join(plainLines(lines), "\n"))
			continue
		}
		if got != c.value {
			t.Errorf("%s %s = %q, want %q", c.id, c.label, got, c.value)
		}
		if _, ok := metaOf(t, lines, c.absent); ok {
			t.Errorf("%s drew a %s line it has no link for:\n%s", c.id, c.absent, strings.Join(plainLines(lines), "\n"))
		}
	}
	// SPEC-2 closes nothing and nothing closes it, so both lines stay out.
	for _, label := range []string{"CLOSES", "CLOSED BY"} {
		if _, ok := metaOf(t, detailLines(t, cfg, "SPEC-2"), label); ok {
			t.Errorf("SPEC-2 drew a %s line with no link", label)
		}
	}
}
