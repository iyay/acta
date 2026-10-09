package tui

import (
	"os/exec"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/commits"
)

// commitHint is the hint bar of the Commits screen. The board keys do nothing
// there, so it never shows the Help hint.
const commitHint = "j/k pick · tab focus · c chore · o pager · esc back"

// commitWide is the width from where the list and the diff sit side by side.
const commitWide = 80

// showDiff reads the diff of one commit. It is a hook so tests can fake git.
var showDiff = commits.Diff

// pagerCmd builds the command of the o key: git show in the repo root, which
// git sends through the pager of the user.
var pagerCmd = func(repo, sha string) *exec.Cmd {
	c := exec.Command("git", "show", sha, "--")
	c.Dir = repo
	return c
}

// diffLoadedMsg carries the diff of one commit back from its goroutine.
type diffLoadedMsg struct {
	sha  string
	text string
	err  error
}

// pagerDoneMsg says the pager of the o key is closed.
type pagerDoneMsg struct{ err error }

// diffResult is what a diff load gave: the text, or why there is none.
type diffResult struct {
	text string
	err  error
}

// commitRow is one row of the commit list. tasks holds the task numbers that
// name the commit, which only a plan shows.
type commitRow struct {
	c     commits.Commit
	tasks []int
}

// diffPaint keeps the last colored diff. Coloring a long diff on every draw
// would make the wheel feel slow, and the draw only ever asks for the same
// sha at the same width. It sits behind a pointer that every copy of the
// model shares, the way frameCache does.
type diffPaint struct {
	sha   string
	w     int
	lines []string
}

// commitView is the state of the Commits screen. The model keeps a pointer to
// one, and nil means the screen is closed.
type commitView struct {
	itemID    string // the task or plan the screen shows
	plan      bool   // a plan lists its tasks' commits with a #n label
	all       []commitRow
	chore     bool   // true shows the chore commits too
	pick      string // sha of the picked row
	at        int    // row number of the pick, used when its sha is gone
	listOff   int    // first row the list shows
	diffOff   int    // first line the diff shows
	diffFocus bool   // true when j k and the page keys scroll the diff
	paint     *diffPaint
}

// commitRowsOf lists the commits of a task or a plan, newest first, one row for
// each sha. A commit that names two tasks of a plan is one row with two labels.
func commitRowsOf(it *board.Item, found map[string][]commits.Commit) []commitRow {
	if it == nil {
		return nil
	}
	type keyed struct {
		key string
		num int
	}
	var keys []keyed
	for _, k := range commitKeys(it, found) {
		_, n, _ := strings.Cut(k, "#")
		num, _ := strconv.Atoi(n)
		keys = append(keys, keyed{k, num})
	}
	// Lowest task number first, so the labels of a row read in order.
	sort.SliceStable(keys, func(i, j int) bool { return keys[i].num < keys[j].num })
	at := map[string]int{}
	var rows []commitRow
	for _, k := range keys {
		for _, c := range found[k.key] {
			if i, ok := at[c.Sha]; ok {
				rows[i].tasks = append(rows[i].tasks, k.num)
				continue
			}
			at[c.Sha] = len(rows)
			rows = append(rows, commitRow{c: c, tasks: []int{k.num}})
		}
	}
	// The load lists each task oldest first, so the newest ones go on top.
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].c.Date.After(rows[j].c.Date) })
	return rows
}

// text is the row as the list shows it, before it is cut to the box width.
func (r commitRow) text(plan bool) string {
	sha := r.c.Sha
	if len(sha) > 7 {
		sha = sha[:7]
	}
	label := ""
	if plan && len(r.tasks) > 0 {
		parts := make([]string, len(r.tasks))
		for i, n := range r.tasks {
			parts[i] = "#" + strconv.Itoa(n)
		}
		label = strings.Join(parts, ",") + "  "
	}
	return sha + "  " + label + expandTabs(r.c.Subject)
}

// visible is the rows the list shows: all of them, or without the chore commits.
func (v commitView) visible() []commitRow {
	if v.chore {
		return v.all
	}
	out := make([]commitRow, 0, len(v.all))
	for _, r := range v.all {
		if !r.c.Chore {
			out = append(out, r)
		}
	}
	return out
}

// cursor is the row number of the pick: its sha when the sha is still listed,
// else the old row number kept inside the list. It is -1 for an empty list.
func (v commitView) cursor(rows []commitRow) int {
	for i, r := range rows {
		if r.c.Sha == v.pick {
			return i
		}
	}
	if len(rows) == 0 {
		return -1
	}
	return clamp(v.at, 0, len(rows)-1)
}

// pickSha is the sha of the picked row, or "" when the list is empty.
func (v commitView) pickSha() string {
	rows := v.visible()
	if i := v.cursor(rows); i >= 0 {
		return rows[i].c.Sha
	}
	return ""
}

// settled puts the pick on a row of the list as it is now. The diff goes back
// to its top when the pick is another commit than before.
func (v commitView) settled(before string) commitView {
	rows := v.visible()
	i := v.cursor(rows)
	if i < 0 {
		v.pick, v.at = "", 0
	} else {
		v.pick, v.at = rows[i].c.Sha, i
	}
	if v.pick != before {
		v.diffOff = 0
	}
	return v
}

// commitBoxes measures the two boxes of the screen: side by side from 80
// columns on, else the list on top with a third of the height.
func (m Model) commitBoxes() (list, diff box, wide bool) {
	bodyH := max(0, m.height-barRows-1)
	if m.width >= commitWide {
		leftW := m.width * 35 / 100
		list = box{x: 0, y: barRows, w: leftW, h: bodyH, inner: max(0, bodyH-2)}
		diff = box{x: leftW, y: barRows, w: m.width - leftW, h: bodyH, inner: max(0, bodyH-2)}
		return list, diff, true
	}
	listH := bodyH / 3
	list = box{x: 0, y: barRows, w: m.width, h: listH, inner: max(0, listH-2)}
	diff = box{x: 0, y: barRows + listH, w: m.width, h: bodyH - listH, inner: max(0, bodyH-listH-2)}
	return list, diff, false
}

// diffBody gives the lines of the diff box for the picked commit: nothing for
// an empty list, a wait line while the diff loads, the error when it failed,
// else the colored diff.
func (m Model) diffBody(v commitView, w int) []string {
	sha := v.pickSha()
	if sha == "" {
		return nil
	}
	res, ok := m.diffCache[sha]
	switch {
	case !ok:
		return []string{m.styles.faint.Render("loading…")}
	case res.err != nil:
		lines := cut(res.err.Error(), w)
		for i, ln := range lines {
			lines[i] = m.styles.problem.Render(ln)
		}
		return lines
	}
	if v.paint.sha != sha || v.paint.w != w || v.paint.lines == nil {
		*v.paint = diffPaint{sha: sha, w: w, lines: colorDiff(res.text, w, m.styles)}
	}
	return v.paint.lines
}

// diffLast is the largest first line the diff box can have.
func (m Model) diffLast(v commitView) int {
	_, d, _ := m.commitBoxes()
	return max(0, len(m.diffBody(v, d.textW()))-d.inner)
}

// listLast is the largest first row the list can have.
func (m Model) listLast(v commitView) int {
	l, _, _ := m.commitBoxes()
	return max(0, len(v.visible())-l.inner)
}

// keepPickShown slides the list so the picked row is on screen.
func (m Model) keepPickShown(v commitView) commitView {
	l, _, _ := m.commitBoxes()
	if l.inner < 1 {
		return v
	}
	v.listOff = clamp(clamp(v.listOff, 0, m.listLast(v)), v.at+1-l.inner, v.at)
	return v
}

// openCommits opens the Commits screen for the task or plan under the cursor.
// Any other row, and the group row, does nothing.
func (m Model) openCommits() (Model, tea.Cmd) {
	it := m.Selected()
	if it == nil || (it.Kind != board.KindTask && it.Kind != board.KindPlan) {
		return m, nil
	}
	v := commitView{
		itemID: it.ID,
		plan:   it.Kind == board.KindPlan,
		all:    commitRowsOf(it, m.taskCommits),
		paint:  &diffPaint{},
	}
	m.commitScreen = ptr(m.keepPickShown(v.settled("")))
	// Notches gathered on the board must not scroll it from behind the screen.
	m.wheelDelta, m.wheelArmed, m.wheelMoved = 0, false, false
	return m, m.loadPick()
}

func ptr(v commitView) *commitView { return &v }

// loadPick asks for the diff of the picked commit when it is not cached.
func (m Model) loadPick() tea.Cmd {
	if m.commitScreen == nil {
		return nil
	}
	sha := m.commitScreen.pickSha()
	if sha == "" {
		return nil
	}
	if _, ok := m.diffCache[sha]; ok {
		return nil
	}
	// The hook is read now, so a test that puts the real one back is never
	// raced by a load that is still running.
	read, repo := showDiff, m.cfg.RepoRoot
	return func() tea.Msg {
		text, err := read(repo, sha)
		return diffLoadedMsg{sha: sha, text: text, err: err}
	}
}

// storeDiff keeps a loaded diff for its sha. A diff that arrives after the
// screen closed is dropped, and the cache is copied so older copies of the
// model never see it change.
func (m Model) storeDiff(msg diffLoadedMsg) Model {
	if m.commitScreen == nil {
		return m
	}
	next := make(map[string]diffResult, len(m.diffCache)+1)
	for k, r := range m.diffCache {
		next[k] = r
	}
	next[msg.sha] = diffResult{text: msg.text, err: msg.err}
	m.diffCache = next
	return m
}

// reloadCommits runs after a board reload while the screen is open. The
// screen stays on its item and re-reads that item's commits, the pick stays on
// the same sha when it is still listed, and the diff cache starts empty.
func (m Model) reloadCommits() (Model, tea.Cmd) {
	m.diffCache = nil
	if m.commitScreen == nil {
		return m, nil
	}
	v := *m.commitScreen
	before := v.pick
	var it *board.Item
	if m.board != nil {
		it = m.board.Get(v.itemID)
	}
	v.all = commitRowsOf(it, m.taskCommits)
	m.commitScreen = ptr(m.keepPickShown(v.settled(before)))
	return m, m.loadPick()
}

// commitKey takes every key while the screen is open. A key it has no use for
// does nothing, so no board key can act on the item behind it.
func (m Model) commitKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	v := *m.commitScreen
	before := v.pick
	switch k.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		m.commitScreen = nil
		return m, nil
	case "tab":
		v.diffFocus = !v.diffFocus
	case "c":
		v.chore = !v.chore
		v = v.settled(before)
	case "o":
		return m.commitPager()
	case "j", "down":
		v = m.commitMove(v, 1)
	case "k", "up":
		v = m.commitMove(v, -1)
	case "ctrl+d":
		v = m.commitMove(v, pageLines)
	case "ctrl+u":
		v = m.commitMove(v, -pageLines)
	case "g":
		v = m.commitMove(v, -m.commitSpan(v))
	case "G":
		v = m.commitMove(v, m.commitSpan(v))
	}
	// A diff scroll never moves the list, so a wheel turn on the list stays.
	if !v.diffFocus {
		v = m.keepPickShown(v)
	}
	m.commitScreen = &v
	return m, m.loadPick()
}

// commitSpan is a distance longer than the focused box, for the g and G keys.
func (m Model) commitSpan(v commitView) int {
	if v.diffFocus {
		return len(m.diffBody(v, m.diffWidth()))
	}
	return len(v.visible())
}

// diffWidth is how many cells the diff box has for its words.
func (m Model) diffWidth() int {
	_, d, _ := m.commitBoxes()
	return d.textW()
}

// commitMove goes down lines of the focused box: the pick in the list, the
// first line of the diff in the diff box. Both stop at their ends.
func (m Model) commitMove(v commitView, lines int) commitView {
	if v.diffFocus {
		v.diffOff = clamp(v.diffOff+lines, 0, m.diffLast(v))
		return v
	}
	rows := v.visible()
	i := v.cursor(rows)
	if i < 0 {
		return v
	}
	before := v.pick
	v.at = clamp(i+lines, 0, len(rows)-1)
	v.pick = rows[v.at].c.Sha
	if v.pick != before {
		v.diffOff = 0
	}
	return v
}

// commitPager opens git show of the picked commit in the pager. The terminal
// is handed over to it, and pagerDoneMsg turns the mouse back on.
func (m Model) commitPager() (tea.Model, tea.Cmd) {
	sha := m.commitScreen.pickSha()
	// A leading dash would make git read the sha as an option.
	if sha == "" || strings.HasPrefix(sha, "-") {
		return m, nil
	}
	return m, tea.ExecProcess(pagerCmd(m.cfg.RepoRoot, sha), func(err error) tea.Msg {
		return pagerDoneMsg{err: err}
	})
}

// commitMouse takes every mouse event while the screen is open. Only the wheel
// does anything: it scrolls the box under the pointer, never the board.
func (m Model) commitMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	step := wheelStep
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		step = -wheelStep
	case tea.MouseButtonWheelDown:
	default:
		m.same = true
		return m, nil
	}
	list, diff, _ := m.commitBoxes()
	v := *m.commitScreen
	switch {
	case inBox(list, msg.X, msg.Y):
		v.listOff = clamp(v.listOff+step, 0, m.listLast(v))
	case inBox(diff, msg.X, msg.Y):
		v.diffOff = clamp(v.diffOff+step, 0, m.diffLast(v))
	default:
		m.same = true
		return m, nil
	}
	m.commitScreen = &v
	return m, nil
}

func inBox(b box, x, y int) bool {
	return x >= b.x && x < b.x+b.w && y >= b.y && y < b.y+b.h
}

// commitsFrame draws the Commits screen in place of every pane.
func (m Model) commitsFrame() string {
	v := *m.commitScreen
	listB, diffB, wide := m.commitBoxes()
	list := m.commitListBox(v, listB)
	diff := m.commitDiffBox(v, diffB)
	if !wide {
		return m.frameFrom(append(list, diff...), false)
	}
	body := make([]string, max(len(list), len(diff)))
	for i := range body {
		l, r := strings.Repeat(" ", listB.w), strings.Repeat(" ", diffB.w)
		if i < len(list) {
			l = list[i]
		}
		if i < len(diff) {
			r = diff[i]
		}
		body[i] = l + r
	}
	return m.frameFrom(body, true)
}

// commitListBox draws the list of commits with the pick as a band.
func (m Model) commitListBox(v commitView, b box) []string {
	rows := v.visible()
	cur := v.cursor(rows)
	w := b.textW()
	var lines []string
	first := firstOf(v.listOff, len(rows), b.inner)
	switch {
	case len(rows) == 0 && len(v.all) > 0:
		lines = []string{m.styles.faint.Render(truncate("only chore commits · c to show", w))}
	case len(rows) == 0:
		lines = []string{m.styles.faint.Render("no linked commits")}
	default:
		for i := first; i < len(rows) && i < first+b.inner; i++ {
			text := truncate(rows[i].text(v.plan), w)
			if i == cur {
				lines = append(lines, m.styles.selected.Render(pad(text, w)))
				continue
			}
			lines = append(lines, text)
		}
	}
	title := "Commits"
	if it := m.board.Get(v.itemID); it != nil {
		title += " · " + shortRef(it)
	}
	if v.chore {
		title += " · +chore"
	}
	foot := itemCount(cur+1, len(rows))
	return m.commitBox(title, b, !v.diffFocus, lines, scrollbar(len(rows), b.inner, first, b.inner), foot)
}

// commitDiffBox draws the diff of the picked commit.
func (m Model) commitDiffBox(v commitView, b box) []string {
	body := m.diffBody(v, b.textW())
	first := firstOf(v.diffOff, len(body), b.inner)
	title := "Diff"
	if sha := v.pickSha(); sha != "" {
		title += " · " + sha[:min(7, len(sha))]
	}
	return m.commitBox(title, b, v.diffFocus, window(body, first, b.inner), scrollbar(len(body), b.inner, first, b.inner), "")
}

// commitBox draws one bordered box of the screen. The focused one wears the
// accent border, and the scrollbar thumb is the right wall itself, the way the
// panes of the board draw it.
func (m Model) commitBox(title string, b box, focused bool, lines []string, bar []bool, foot string) []string {
	if b.w < 2 || b.h < 1 {
		return nil
	}
	edge, name := m.styles.faint, m.styles.faint
	if focused {
		edge, name = m.styles.accent, m.styles.accent.Bold(true)
	}
	top := topLine(b.w, edge, []segment{{text: "─", style: edge}, {text: title, style: name}})
	if b.h == 1 {
		return []string{top}
	}
	inner := b.textW()
	out := make([]string, 0, b.h)
	out = append(out, top)
	for i := range b.inner {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		wall := "│"
		if i < len(bar) && bar[i] {
			wall = "┃"
		}
		out = append(out, edge.Render("│")+pad(line, inner)+edge.Render(wall))
	}
	return append(out, paneBottom(b, edge, foot))
}
