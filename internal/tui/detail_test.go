package tui

import (
	"fmt"
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
	t.Parallel()

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
	t.Parallel()

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
			"✓ PLN-0001.01  First",
			"● PLN-0001.02  Second  1/2 · omp",
			"○ PLN-0001.03  Third")
		if got := plainLines(lines)[lineOf(t, lines, "PLN-0001.02")]; !strings.HasSuffix(got, "1/2 · omp") {
			t.Errorf("the task under way ends on %q, want its 1/2 and its agent", got)
		}
		for _, id := range []string{"PLN-0001.01", "PLN-0001.03"} {
			ln := lines[lineOf(t, lines, id)]
			if sgrHas(ln, "2") {
				t.Errorf("the line of %s is faint: %q", id, plain(ln))
			}
			if strings.ContainsAny(plain(ln), "/·") {
				t.Errorf("the line of %s carries a count or an agent: %q", id, plain(ln))
			}
		}
		// Work under way wears the foreground, so the text behind the id
		// carries no color of its own. The id itself keeps its kind color.
		if i := lineOf(t, lines, "PLN-0001.02"); sgr.MatchString(lines[i][strings.Index(lines[i], "  Second")+2:]) {
			t.Errorf("the text of the work under way wears a color: %q", lines[i])
		}
	})
}

// A task lists its own steps with no id in front of them, and the first box
// still open in a task under way is the step being worked on.
func TestDetailTaskListsItsSteps(t *testing.T) {
	withColors(func() {
		lines := detailOf(t, "PLAN-1.2")
		wantInOrder(t, lines, "✓ b", "● c")
		// The step under way wears the pulse dot and no color of its own, so
		// what follows the dot carries no code.
		i := lineOf(t, lines, "● c")
		rest := lines[i][strings.Index(lines[i], dotGoing):]
		rest = rest[strings.Index(rest, "\x1b[0m")+len("\x1b[0m"):]
		if sgr.MatchString(rest) {
			t.Errorf("the text of the step under way wears a color: %q", lines[i])
		}
		if i := lineOf(t, lines, "✓ b"); sgrHas(lines[i], "2") {
			t.Errorf("a ticked step is faint: %q", lines[i])
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
	t.Parallel()

	wantInOrder(t, detailOf(t, "SPEC-1"),
		"PLN-0001  Plan A",
		"✓ PLN-0001.01  First",
		"● PLN-0001.02  Second  1/2 · omp",
		"○ PLN-0001.03  Third")
	wantInOrder(t, detailOf(t, "BUG-1"),
		"PLN-0002  Plan C",
		"○ PLN-0002.01  Fix one")
	// A spec with no plan shows its header and no list, not an empty line.
	lines := detailOf(t, "SPEC-2")
	if len(plainLines(lines)) < 3 || !strings.HasPrefix(plainLines(lines)[0], "ID") {
		t.Errorf("a spec with no plan should still show its header:\n%s", strings.Join(plainLines(lines), "\n"))
	}
	if joined := strings.Join(plainLines(lines), "\n"); strings.ContainsAny(joined, "○●✓") {
		t.Errorf("a spec with no plan lists something:\n%s", joined)
	}
}

// A debt item is one note, and the list pane already names the other notes,
// so its detail shows the whole note and nothing else of the file.
func TestDetailDebtItemShowsOnlyItsNote(t *testing.T) {
	t.Parallel()

	lines := plainLines(detailOf(t, "DEBT-1.1"))
	head := strings.Join(lines[:ruleLine(t, lines)], "\n")
	below := strings.Join(lines[ruleLine(t, lines):], "\n")
	if strings.Contains(head, "DEBT ") {
		t.Errorf("the header still has a DEBT line:\n%s", head)
	}
	if n := strings.Count(below, "first note"); n != 1 {
		t.Errorf("the note shows %d times under the header, want once:\n%s", n, below)
	}
	for _, gone := range []string{"second note", "Prose about the review.", "Review NOTEs"} {
		if strings.Contains(below, gone) {
			t.Errorf("the debt item detail still shows %q:\n%s", gone, below)
		}
	}
}

// A note longer than the pane is wrapped, so every word of it can be read and
// no line spills over the wall.
func TestDetailDebtItemWrapsALongNote(t *testing.T) {
	t.Parallel()

	note := strings.TrimSpace(strings.Repeat("a long note word ", 20)) + " last<uid>"
	cfg := treeCfg(t, map[string]string{
		".acta/debt/2026-09-24-d.md": "---\nid: DEBT-1\n---\n# Review NOTEs\n\n- [ ] " + note + "\n",
	})
	// The real renderer, because the note is escaped for it and a stub that
	// hands the text back would show the escapes.
	lines := plainLines(onItem(t, detailModel(t, cfg), "DEBT-1.1").detailLines(100))
	below := lines[ruleLine(t, lines):]
	if got := strings.Join(strings.Fields(strings.Join(below, " ")), " "); !strings.Contains(got, note) {
		t.Errorf("the detail does not hold the whole note %q:\n%s", note, strings.Join(below, "\n"))
	}
	for i, ln := range lines {
		if w := lipgloss.Width(ln); w > 100 {
			t.Errorf("line %d is %d cells wide: %q", i, w, ln)
		}
	}
}

// A debt file still lists every one of its lines, because only the debt
// item's detail changed.
func TestDetailDebtFileListsEveryLine(t *testing.T) {
	t.Parallel()

	m := detailModel(t, treeCfg(t, boardFiles()))
	lines := plainLines(m.debtLines(m.board.Get("DEBT-1"), 100))
	wantInOrder(t, lines, "○ DBT-0001.01  first note", "✓ DBT-0001.02  second note")
}

// A note is drawn by the markdown renderer at every width, and no line is
// cut, so the reader never loses a word, a lone "-" or a tag like <uid>.
func TestDetailDebtItemNoteLosesNothingAtAnyWidth(t *testing.T) {
	t.Parallel()

	notes := []string{
		"internal/write/mark.go: + on a task already done (or - on one already open) changes nothing, and more",
		`a checklist line gives "- [ ] - [ ] text".`,
		"averyveryveryverylongwordwithnospacesatallthatrunsonandon then short",
		"日本語のメモ and ünïcode words in one note",
		"Another local user can pre-create /tmp/pmb-<uid>: a & b",
		"view.go still runs `acta list --json` with *no* agent field",
		`brainCmd: the (?:^|&&|;|\|\|) alternation is dead`,
	}
	// Markdown turns these into styling, so they are not drawn as text.
	styled := strings.NewReplacer("`", "", "*", "")
	for _, note := range notes {
		cfg := treeCfg(t, map[string]string{
			".acta/debt/2026-09-24-d.md": "---\nid: DEBT-1\n---\n# Review NOTEs\n\n- [ ] " + note + "\n",
		})
		m := onItem(t, detailModel(t, cfg), "DEBT-1.1")
		want := strings.Join(strings.Fields(styled.Replace(note)), "")
		for w := 10; w <= 160; w++ {
			_, mid, _ := m.buildDetailParts(w)
			var got strings.Builder
			for i, ln := range mid {
				if lw := lipgloss.Width(ln); lw > w {
					t.Errorf("at %d wide line %d is %d cells: %q", w, i, lw, plain(ln))
				}
				got.WriteString(strings.Join(strings.Fields(plain(ln)), ""))
			}
			if got.String() != want {
				t.Errorf("at %d wide the note reads %q, want %q", w, got.String(), want)
			}
		}
	}
}

// A note sits in the detail the way a spec body does: the same blank line
// above it and the same gap on the left, so the two kinds look alike.
func TestDetailDebtItemNoteHasTheSpecBodyGaps(t *testing.T) {
	t.Parallel()

	const text = "Tests that call lock without the base still leave lock files"
	cfg := treeCfg(t, map[string]string{
		".acta/debt/2026-09-24-d.md":  "---\nid: DEBT-1\n---\n# Review NOTEs\n\n- [ ] " + text + "\n",
		".acta/specs/2026-09-20-a.md": "---\nid: SPEC-1\n---\n" + text + "\n",
	})
	firstText := func(mid []string) (row, col int) {
		for i, ln := range mid {
			if c := strings.Index(plain(ln), "Tests"); c >= 0 {
				return i, c
			}
		}
		t.Fatalf("no line holds the text:\n%s", strings.Join(mid, "\n"))
		return 0, 0
	}
	for _, w := range []int{40, 80, 120} {
		_, debtMid, _ := onItem(t, detailModel(t, cfg), "DEBT-1.1").buildDetailParts(w)
		_, specMid, _ := onItem(t, detailModel(t, cfg), "SPEC-1").buildDetailParts(w)
		dr, dc := firstText(debtMid)
		_, sc := firstText(specMid)
		if dc != sc {
			t.Errorf("at %d wide the note starts at column %d, a spec body at %d", w, dc, sc)
		}
		if dr == 0 || strings.TrimSpace(plain(debtMid[dr-1])) != "" {
			t.Errorf("at %d wide the note has no blank line above it:\n%s", w, strings.Join(debtMid, "\n"))
		}
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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	cfg := treeCfg(t, closesFiles())
	for _, c := range []struct{ id, label, value, absent string }{
		{"SPEC-1", "CLOSES", "SCR-0001, scratch/2026-09-28-j", "CLOSED BY"},
		{"SCRATCH-1", "CLOSED BY", "SPC-0001", "CLOSES"},
		{"scratch/2026-09-28-j", "CLOSED BY", "SPC-0001", "CLOSES"},
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

// Seven scratch items: one with all three dates, one with none, one with each
// date on its own, one whose three values are all junk, and one whose three
// dates carry the time to the second. A good date comes in both forms yaml
// gives: bare, which decodes to a time, and quoted, which decodes to a
// string.
func datedFiles() map[string]string {
	return map[string]string{
		".acta/scratch/2026-09-01-all.md":      "---\nid: SCRATCH-1\ntitle: all\nstatus: raw\ncreated: 2026-09-01\nstarted: \"2026-09-02\"\nfinished: 2026-09-03\n---\n# All\n",
		".acta/scratch/2026-09-01-none.md":     "---\nid: SCRATCH-2\ntitle: none\nstatus: raw\n---\n# None\n",
		".acta/scratch/2026-09-01-created.md":  "---\nid: SCRATCH-3\ntitle: created\nstatus: raw\ncreated: 2026-09-01\n---\n# Created\n",
		".acta/scratch/2026-09-01-started.md":  "---\nid: SCRATCH-4\ntitle: started\nstatus: raw\nstarted: \"2026-09-02\"\n---\n# Started\n",
		".acta/scratch/2026-09-01-finished.md": "---\nid: SCRATCH-5\ntitle: finished\nstatus: raw\nfinished: \"2026-09-03\"\n---\n# Finished\n",
		".acta/scratch/2026-09-01-bad.md":      "---\nid: SCRATCH-6\ntitle: bad\nstatus: raw\ncreated: \"\"\nstarted: 12\nfinished: soon\n---\n# Bad\n",
		".acta/scratch/2026-09-01-timed.md":    "---\nid: SCRATCH-7\ntitle: timed\nstatus: raw\ncreated: \"2026-09-01 08:00:00\"\nstarted: \"2026-09-02 09:30:15\"\nfinished: \"2026-09-03 17:45:59\"\n---\n# Timed\n",
		".acta/scratch/2026-09-01-notime.md":   "---\nid: SCRATCH-8\ntitle: notime\nstatus: raw\ncreated: \"2026-02-30 10:00:00\"\nstarted: \"2026-09-30 25:00:00\"\nfinished: \"2026-09-30 16:14\"\n---\n# No Time\n",
		".acta/scratch/2026-09-01-word.md":     "---\nid: SCRATCH-9\ntitle: word\nstatus: raw\ncreated: yesterday\nstarted: \"2026-02-30 10:00:00\"\nfinished: \"2026-09-30 16:14\"\n---\n# Word\n",
	}
}

// The board keeps a date only when the frontmatter holds a real day, or a real
// day and time down to the second, and nothing at all when the value is
// missing, empty, a number, a word, a day that does not exist, an hour that
// does not exist, or a time that stops before the seconds.
func TestItemDatesFromFrontmatter(t *testing.T) {
	b, err := board.Load(treeCfg(t, datedFiles()))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ id, created, started, finished string }{
		{"SCRATCH-1", "2026-09-01", "2026-09-02", "2026-09-03"},
		{"SCRATCH-2", "", "", ""},
		{"SCRATCH-3", "2026-09-01", "", ""},
		{"SCRATCH-4", "", "2026-09-02", ""},
		{"SCRATCH-5", "", "", "2026-09-03"},
		{"SCRATCH-6", "", "", ""},
		{"SCRATCH-7", "2026-09-01 08:00:00", "2026-09-02 09:30:15", "2026-09-03 17:45:59"},
		{"SCRATCH-8", "", "", ""},
		{"SCRATCH-9", "", "", ""},
	} {
		it := b.Get(c.id)
		if it == nil {
			t.Fatalf("the board holds no %s", c.id)
		}
		if it.Created != c.created || it.StartedOn != c.started || it.Finished != c.finished {
			t.Errorf("%s: got %q %q %q, want %q %q %q",
				c.id, it.Created, it.StartedOn, it.Finished, c.created, c.started, c.finished)
		}
	}
}

// Every item ends with one date line that names all three dates, a dash for
// each one not set, and the header holds no date label at all.
func TestDetailShowsTheDates(t *testing.T) {
	cfg := treeCfg(t, datedFiles())
	for _, c := range []struct{ id, foot string }{
		{"SCRATCH-1", "created 2026-09-01 · started 2026-09-02 · finished 2026-09-03"},
		{"SCRATCH-2", "created - · started - · finished -"},
		{"SCRATCH-3", "created 2026-09-01 · started - · finished -"},
		{"SCRATCH-4", "created - · started 2026-09-02 · finished -"},
		{"SCRATCH-5", "created - · started - · finished 2026-09-03"},
		{"SCRATCH-6", "created - · started - · finished -"},
		{"SCRATCH-7", "created 2026-09-01 08:00:00 · started 2026-09-02 09:30:15 · finished 2026-09-03 17:45:59"},
		{"SCRATCH-8", "created - · started - · finished -"},
		{"SCRATCH-9", "created - · started - · finished -"},
	} {
		lines := detailLines(t, cfg, c.id)
		if got := plain(lines[len(lines)-1]); got != c.foot {
			t.Errorf("%s footer %q, want %q", c.id, got, c.foot)
		}
		for _, l := range labelsOf(lines) {
			if l == "CREATED" || l == "STARTED" || l == "FINISHED" {
				t.Errorf("%s header still holds %s", c.id, l)
			}
		}
	}
}

// The footer always names all three dates: the long form when the box is wide
// enough for it, the short one when it is not, and a cut line only when even
// the short one does not fit. Every width from 5 up to 130 is read, for every
// shape the three dates can have, so no width is left to chance.
func TestDetailFooterNamesEveryDateAtEveryWidth(t *testing.T) {
	t.Parallel()

	cfg := treeCfg(t, datedFiles())
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	base := New(cfg, b, true)
	base.render = func(md string, _ int) string { return md }
	for _, c := range []struct{ id, long, short string }{
		{"SCRATCH-1", "created 2026-09-01 · started 2026-09-02 · finished 2026-09-03", "c 2026-09-01 · s 2026-09-02 · f 2026-09-03"},
		{"SCRATCH-2", "created - · started - · finished -", "c - · s - · f -"},
		{"SCRATCH-3", "created 2026-09-01 · started - · finished -", "c 2026-09-01 · s - · f -"},
		{"SCRATCH-4", "created - · started 2026-09-02 · finished -", "c - · s 2026-09-02 · f -"},
		{"SCRATCH-5", "created - · started - · finished 2026-09-03", "c - · s - · f 2026-09-03"},
		{"SCRATCH-6", "created - · started - · finished -", "c - · s - · f -"},
		{"SCRATCH-7", "created 2026-09-01 08:00:00 · started 2026-09-02 09:30:15 · finished 2026-09-03 17:45:59", "c 2026-09-01 08:00:00 · s 2026-09-02 09:30:15 · f 2026-09-03 17:45:59"},
		{"SCRATCH-8", "created - · started - · finished -", "c - · s - · f -"},
		{"SCRATCH-9", "created - · started - · finished -", "c - · s - · f -"},
	} {
		m := onItem(t, sized(base, 120, 40), c.id)
		for w := 5; w <= 130; w++ {
			want := c.long
			if lipgloss.Width(want) > w {
				want = c.short
				if lipgloss.Width(want) > w {
					want = truncate(c.short, w)
				}
			}
			if got := plain(footAt(m, w)); got != want {
				t.Errorf("%s at width %d: footer %q, want %q", c.id, w, got, want)
			}
		}
		// Every width the short form fits in, all three dates are there whole.
		for w := lipgloss.Width(c.short); w <= 130; w++ {
			if got := plain(footAt(m, w)); strings.Contains(got, "…") {
				t.Errorf("%s at width %d: the short form fits but the footer is cut: %q", c.id, w, got)
			}
		}
	}
}

// footAt is the footer the detail draws in a box w cells wide.
func footAt(m Model, w int) string {
	_, _, foot := m.detailParts(w)
	return foot
}

// said is what a drawn line of a box holds, with the scrollbar thumb cut off
// and the padding on the right dropped, because a line that fills its box
// ends on the thumb and not on the plain wall.
func said(line string) string {
	return strings.TrimRight(cutWalls(line), " ")
}

// stickyModel is a spec with a long body, shown in a detail box of a screen
// h lines tall, with the focus on the detail.
func stickyModel(t *testing.T, h int) Model {
	t.Helper()
	var body strings.Builder
	for i := range 60 {
		fmt.Fprintf(&body, "line %02d\n\n", i)
	}
	cfg := treeCfg(t, map[string]string{
		".acta/specs/2026-09-20-long.md": "---\nid: SPC-0001\ncreated: \"2026-09-20\"\nstarted: \"2026-09-21\"\n---\n# Long spec\n\n" + body.String(),
	})
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	return onItem(t, sized(m, 120, h), "specs/2026-09-20-long")
}

func TestDetailHeaderAndFooterStayWhileTheMiddleScrolls(t *testing.T) {
	t.Parallel()

	m := stickyModel(t, 40)
	b := m.geometry().detail
	head, mid, foot := m.detailParts(b.textW())
	want := "created 2026-09-20 · started 2026-09-21 · finished -"
	if plain(foot) != want {
		t.Fatalf("footer %q, want %q", plain(foot), want)
	}
	for _, l := range labelsOf(head) {
		if l == "CREATED" || l == "STARTED" || l == "FINISHED" {
			t.Errorf("the header still holds %s", l)
		}
	}
	n := stickyMid(len(head), b.inner)
	if n < 3 {
		t.Fatalf("a 40 line screen should stick, got %d middle lines", n)
	}
	// The box counts the lines that move: it scrolls over the middle and shows
	// n of them at a time, so an offset can never leave the middle behind.
	if total, fit := m.detailScroll(b.textW(), b.inner); total != len(mid) || fit != n {
		t.Errorf("the detail scrolls over %d lines and shows %d at a time, want %d and %d",
			total, fit, len(mid), n)
	}
	for _, off := range []int{0, 1, 5, len(mid) - n, 1000} {
		s := m
		s.off[paneDetail] = 0
		s.scrollPane(paneDetail, off)
		lines := innerLines(s, b)
		for i, h := range head {
			if got := said(lines[i]); got != strings.TrimRight(plain(h), " ") {
				t.Errorf("off=%d header line %d: %q, want %q", off, i, got, plain(h))
			}
		}
		if got := said(lines[len(lines)-1]); got != want {
			t.Errorf("off=%d last line %q, want the footer", off, got)
		}
		first := min(off, max(0, len(mid)-n))
		for i := range n {
			if got, want := said(lines[len(head)+i]), strings.TrimRight(plain(mid[first+i]), " "); got != want {
				t.Errorf("off=%d middle line %d: %q, want %q", off, i, got, want)
			}
		}
	}
}

func TestShortDetailScrollsAsOneBlock(t *testing.T) {
	t.Parallel()

	checked := 0
	for h := 8; h < 40; h++ {
		m := stickyModel(t, h)
		b := m.geometry().detail
		head, _, _ := m.detailParts(b.textW())
		if b.inner < 1 || stickyMid(len(head), b.inner) > 0 {
			continue
		}
		checked++
		all := m.detailLines(b.textW())
		m.scrollPane(paneDetail, 1)
		if got := innerLines(m, b); len(got) == 0 || said(got[0]) != strings.TrimRight(plain(all[1]), " ") {
			t.Errorf("h=%d: after one line the box starts %q, want %q", h, got, plain(all[1]))
		}
	}
	if checked == 0 {
		t.Fatal("no screen height was too short to stick")
	}
}

func TestStickyMidNeedsThreeMiddleLines(t *testing.T) {
	t.Parallel()

	for _, c := range []struct{ head, h, want int }{
		{5, 10, 3}, {5, 9, 0}, {0, 5, 3}, {0, 4, 0}, {5, 0, 0}, {5, 31, 24},
	} {
		if got := stickyMid(c.head, c.h); got != c.want {
			t.Errorf("stickyMid(%d, %d) = %d, want %d", c.head, c.h, got, c.want)
		}
	}
}

func TestDetailLinesAreBuiltOncePerChange(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 160, 50)
	calls := 0
	m.render = func(md string, _ int) string { calls++; return md }
	m.dcache = &detailCache{}
	// The width the detail box really has, so the draw below asks for the
	// lines of the same width the calls above built.
	w := scrollBox(m, paneDetail).textW()
	m.detailLines(w)
	m.detailLines(w)
	m.View()
	if calls != 1 {
		t.Errorf("the same item at the same width rendered its body %d times, want 1", calls)
	}
}

func TestDetailCacheFollowsWidthItemAndBoard(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 160, 50)
	m.dcache = &detailCache{}
	wide := strings.Join(m.detailLines(80), "\n")
	narrow := strings.Join(m.detailLines(30), "\n")
	if wide == narrow {
		t.Error("a new width gave the lines of the old width")
	}
	first := strings.Join(m.detailLines(80), "\n")
	// The plans tab lists two plans, so j walks to another item.
	m = press(m, tabKey(tabPlans), "j")
	if strings.Join(m.detailLines(80), "\n") == first {
		t.Error("a new selected item gave the lines of the old item")
	}
	shown := strings.Join(m.detailLines(80), "\n")
	fresh, err := m.load()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range fresh.Items {
		it.Title = it.Title + " renamed"
	}
	next, _ := m.Update(reloadMsg{b: fresh})
	if got := strings.Join(next.(Model).detailLines(80), "\n"); got == shown || !strings.Contains(got, "renamed") {
		t.Error("a reloaded board gave the lines of the old board")
	}
}

// With no item under the cursor the lines come from the board alone, so a
// board that swaps its items for others must give other lines even though
// the selected item is nil both times.
func TestDetailCacheFollowsABoardWithNothingSelected(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 160, 50)
	// The untyped group of the specs tab has no item of its own.
	m = press(m, tabKey(tabSpecs))
	for m.Selected() != nil {
		m = press(m, "j")
	}
	if s := m.Selected(); s != nil {
		t.Fatalf("the cursor is on %s, want a row that selects no item", s.ID)
	}
	m.dcache = &detailCache{}
	before := strings.Join(m.detailLines(80), "\n")
	fresh, err := m.load()
	if err != nil {
		t.Fatal(err)
	}
	// A board with nothing on it at all. The box names that case in its own
	// words, where a board with a group under the cursor does not.
	fresh.Items = nil
	next, _ := m.Update(reloadMsg{b: fresh})
	if after := strings.Join(next.(Model).detailLines(80), "\n"); after == before {
		t.Errorf("a reloaded board with nothing selected gave the lines of the old board: %q", after)
	}
}

func TestDetailWithoutCacheStillDraws(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 160, 50)
	m.dcache = nil
	if len(m.detailLines(80)) == 0 {
		t.Error("a model with no cache must still draw the detail box")
	}
}

// A new theme brings new styles and a new markdown renderer, so the box has
// to be given lines built with them and not keep the ones of the old theme.
func TestDetailCacheFollowsTheTheme(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 160, 50)
	m.dcache = &detailCache{}
	before := strings.Join(m.detailLines(80), "\n")
	next := m.WithTheme("dracula", false)
	after := strings.Join(next.detailLines(80), "\n")
	if after == before {
		t.Error("a new theme gave the lines of the old theme")
	}
	if !strings.Contains(after, "\x1b[") {
		t.Errorf("a themed detail box draws no color at all: %q", after)
	}
}

// modelOf is a model over a board built from files, drawn with the markdown
// left as it is so a test reads the words of a body.
func modelOf(t *testing.T, files map[string]string) Model {
	t.Helper()
	cfg := treeCfg(t, files)
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	return m
}

// groupModel is the Specs tab with the cursor on the untyped group row, the one
// row that selects no item, and empty lists everywhere else.
func groupModel(t *testing.T) Model {
	t.Helper()
	m := modelOf(t, map[string]string{
		".acta/specs/2026-09-20-alpha.md":          "# Alpha\n",
		"docs/superpowers/specs/2026-01-01-old.md": "# Old\n",
	})
	return sized(press(m, tabKey(tabSpecs), "G"), 160, 50)
}

// wontfixModel is the Bugs tab with the Done box on its Wontfix sub-tab, which
// holds nothing, so nothing is selected there either.
func wontfixModel(t *testing.T) Model {
	t.Helper()
	m := modelOf(t, map[string]string{
		".acta/bugs/2026-09-24-crash.md": "---\nstatus: fixed\n---\n# App crashes\n",
	})
	return sized(press(m, tabKey(tabBugs), "tab", "]"), 160, 50)
}

// With no item under the cursor the box reads the list beside it, and the
// cache key is the item, the width and the board, so it cannot see that list.
// Every input the box reads with no item must still reach the drawn lines.
func TestDetailWithNothingSelectedFollowsTheList(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name  string
		start func(t *testing.T) Model
		leave func(m Model) Model
		want  string
	}{
		{"the query", groupModel, func(m Model) Model { return press(m, "/", "zzzq") }, "No items"},
		{"the open tab", groupModel, func(m Model) Model { return press(m, tabKey(tabDebts)) }, "No items"},
		{"the focused box", groupModel, func(m Model) Model { return press(m, "tab") }, "No items"},
		{"the Done sub-tab", wontfixModel, func(m Model) Model { return press(m, "[") }, "App crashes"},
	} {
		t.Run(c.name, func(t *testing.T) {
			m := c.start(t)
			if s := m.Selected(); s != nil {
				t.Fatalf("the test needs a box with no item on show, it holds %s", s.ID)
			}
			// Draw once, so the cache holds what the box says now.
			drawnBox(t, m)
			m = c.leave(m)
			fresh := m
			fresh.dcache = nil
			got, want := drawnBox(t, m), drawnBox(t, fresh)
			if got != want {
				t.Errorf("the box drew the screen it was on before:\ngot  %q\nwant %q", got, want)
			}
			if !strings.Contains(got, c.want) {
				t.Errorf("the box drew %q, want it to name %q", got, c.want)
			}
		})
	}
}

// drawnBox is what the detail box has on screen, its walls and color cut off.
func drawnBox(t *testing.T, m Model) string {
	t.Helper()
	return strings.Join(plainLines(paneRows(m, paneDetail)), "\n")
}

// Every label of the header wears the label color, the id on the ID line
// wears the color of its kind, and nothing in the header is faint.
func TestDetailLabelsAndIDWearColors(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true) // BUG-0002 under the cursor
		head, _, _ := m.detailParts(100)
		var id, status string
		for _, ln := range head {
			switch {
			case strings.HasPrefix(plain(ln), "ID"):
				id = ln
			case strings.HasPrefix(plain(ln), "STATUS"):
				status = ln
			}
		}
		if !strings.HasPrefix(status, m.styles.label.Render("STATUS")) {
			t.Errorf("STATUS label is not cyan: %q", status)
		}
		if !strings.Contains(id, m.styles.kind(board.KindBug).Render("BUG-0002")) {
			t.Errorf("id is not in the bug color: %q", id)
		}
		for _, ln := range head {
			if sgrHas(ln, "2") && !strings.Contains(plain(ln), "──") {
				t.Errorf("header line is faint: %q", ln)
			}
		}
	})
}

// A dot wears the color of its state, a work line is not faint, the line the
// reader is on is bold, and a problem line is red.
func TestDetailDotsProblemsAndWorkLines(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true)
		s := m.styles
		if s.dot(dotDone).GetForeground() != s.done.GetForeground() ||
			s.dot(dotGoing).GetForeground() != s.pulse[0].GetForeground() ||
			s.dot(dotWaiting).GetForeground() != s.waiting.GetForeground() {
			t.Fatal("a dot does not wear the color of its state")
		}
		it := &board.Item{ShortID: "PLN-0003", Kind: board.KindPlan, Title: "Plan"}
		off := workLine(s, it, false, 60)
		if sgrHas(off, "2") {
			t.Errorf("work line is faint: %q", off)
		}
		if !strings.HasPrefix(off, s.dot(dotWaiting).Render(dotWaiting)) {
			t.Errorf("waiting dot is not grey: %q", off)
		}
		if on := workLine(s, it, true, 60); !sgrHas(on, "1") {
			t.Errorf("the line the reader is on is not bold: %q", on)
		}
		bug := m.Selected()
		bug.Problems = []string{"broken"}
		m.dcache = &detailCache{}
		_, mid, _ := m.detailParts(60)
		if !strings.Contains(strings.Join(mid, "\n"), s.problem.Render("! broken")) {
			t.Errorf("problem line is not red: %q", mid)
		}
	})
}

// The footer is two lines tall on both paths: the rule first, then the dates
// with their names in the footer color.
func TestDetailFooterHasARuleAndColoredLabels(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true)
		for _, h := range []int{40, 6} { // sticky, then one block
			lines := m.detailView(60, 0, h)
			if h == 6 {
				lines = m.detailLines(60)
			}
			n := len(lines)
			for n > 0 && strings.TrimSpace(plain(lines[n-1])) == "" {
				n--
			}
			foot, rule := lines[n-1], plain(lines[n-2])
			if strings.Trim(rule, "─") != "" || rule == "" {
				t.Errorf("h=%d: line above the footer is %q, want a rule", h, rule)
			}
			if !strings.Contains(foot, m.styles.footLabel.Render("created")) {
				t.Errorf("h=%d: footer label is not magenta: %q", h, foot)
			}
		}
	})
}

// The sticky middle counts both footer lines, so a box with no room for the
// header, the rule, the dates and 3 middle lines scrolls as one block.
func TestStickyMidCountsBothFooterLines(t *testing.T) {
	t.Parallel()

	for _, c := range []struct{ head, h, want int }{
		{5, 10, 3}, // 5 head + 2 foot + 3 middle fits exactly
		{5, 9, 0},  // one line short: the whole detail scrolls
		{0, 5, 3},
		{0, 4, 0},
	} {
		if got := stickyMid(c.head, c.h); got != c.want {
			t.Errorf("stickyMid(%d, %d) = %d, want %d", c.head, c.h, got, c.want)
		}
	}
}

// Every line of work names its item by id, and the id wears the color of that
// item's kind, on all the routes that draw one: the tasks of a plan, the plans
// and tasks under a spec or a bug, the specs of an idea and the NOTE lines of
// a debt file. The dot keeps its state color, the line the reader is on stays
// bold from the dot to the agent, and an id the pane cut short stays plain.
func TestDetailWorkLinesWearKindColoredIDs(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true)
		s := m.styles
		it := &board.Item{ShortID: "PLN-0003", Kind: board.KindPlan, Title: "Plan"}
		off := workLine(s, it, false, 60)
		if !strings.Contains(off, s.kind(board.KindPlan).Render("PLN-0003")) {
			t.Errorf("work line id is not in the plan color: %q", off)
		}
		if !strings.HasPrefix(off, s.dot(dotWaiting).Render(dotWaiting)) {
			t.Errorf("work line lost its dot color: %q", off)
		}
		on := workLine(s, it, true, 60)
		if !strings.Contains(on, s.kind(board.KindPlan).Bold(true).Render("PLN-0003")) {
			t.Errorf("the line the reader is on has no bold kind-colored id: %q", on)
		}
		if cut := workLine(s, it, false, 6); strings.Contains(cut, s.kind(board.KindPlan).Render("PLN-0003")) {
			t.Errorf("a cut id got a color: %q", cut)
		}

		bug := m.Selected() // BUG-0002, with PLN-0004 under it
		lines := m.planLines(bug, 80)
		var head, task string
		for _, ln := range lines {
			switch p := plain(ln); {
			case strings.Contains(p, "PLN-0004.01"):
				task = ln
			case strings.HasPrefix(p, "PLN-0004"):
				head = ln
			}
		}
		if !strings.Contains(head, s.kind(board.KindPlan).Render("PLN-0004")) {
			t.Errorf("plan header line id is not in the plan color: %q", head)
		}
		if !strings.Contains(task, s.kind(board.KindTask).Render("PLN-0004.01")) {
			t.Errorf("task line id is not in the task color: %q", task)
		}
	})
}

// A work line shows a finished item in green with its id still in the color
// of its kind, while work under way reads in the plain foreground with the
// pulse dot in front of it.
func TestDetailWorkLinesShowDoneAndGoing(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true)
		s := m.styles
		green := sgr.FindString(s.done.Render("x"))
		done := &board.Item{ShortID: "PLN-0003.01", Kind: board.KindTask, Title: "Finished", Status: "done"}
		ln := workLine(s, done, false, 80)
		if !strings.HasPrefix(ln, s.done.Render(dotDone)) {
			t.Errorf("done mark is not green: %q", ln)
		}
		if !strings.Contains(ln, s.kind(board.KindTask).Render("PLN-0003.01")) {
			t.Errorf("done id lost its kind color: %q", ln)
		}
		if !strings.Contains(ln[strings.Index(ln, "PLN-0003.01"):], green) {
			t.Errorf("done title is not green: %q", ln)
		}
		if on := workLine(s, done, true, 80); !sgrHas(on, "1") || !strings.Contains(on, green) {
			t.Errorf("the done line the reader is on is not bold green: %q", on)
		}

		going := m.board.Get("plans/2026-09-23-q#task-1") // PLN-0004.01, under way
		gl := workLine(s, going, false, 80)
		if !strings.HasPrefix(gl, s.goingDot) {
			t.Errorf("work under way has no pulse dot: %q", gl)
		}
		// The id keeps its kind color, and in this theme the kind color and the
		// accent are the same code, so the id comes off before the text is
		// read for the accent.
		rest := strings.TrimPrefix(gl, s.goingDot)
		rest = strings.Replace(rest, s.kind(board.KindTask).Render(going.ShortID), "", 1)
		if strings.Contains(rest, sgr.FindString(s.accent.Render("x"))) {
			t.Errorf("work under way text is still blue: %q", gl)
		}
		if strings.Contains(rest, green) {
			t.Errorf("work under way text is green: %q", gl)
		}
	})
}

// A ticked step is green, the step under way is plain with the pulse dot, and
// the rest of the boxes keep their own marks.
func TestDetailStepLinesShowDoneAndGoing(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true)
		s := m.styles
		task := m.board.Get("plans/2026-09-23-q#task-1") // steps: [x] f, [ ] g, under way
		lines := m.stepLines(task, 60)
		if len(lines) != 2 {
			t.Fatalf("got %d step lines, want 2: %q", len(lines), lines)
		}
		if want := s.done.Render(dotDone); !strings.HasPrefix(lines[0], want) {
			t.Errorf("ticked step mark is not green: %q", lines[0])
		}
		if !strings.Contains(lines[0], sgr.FindString(s.done.Render("x"))+" f") && !strings.Contains(lines[0], s.done.Render(" f")) {
			t.Errorf("ticked step text is not green: %q", lines[0])
		}
		if !strings.HasPrefix(lines[1], s.goingDot) {
			t.Errorf("the step under way has no pulse dot: %q", lines[1])
		}
		if strings.Contains(lines[1], sgr.FindString(s.accent.Render("x"))+" g") {
			t.Errorf("the step under way is still in the accent: %q", lines[1])
		}
	})
}
