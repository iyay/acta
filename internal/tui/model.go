// Package tui is the lazygit-style screen over a Board.
package tui

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"pm-board/internal/board"
	"pm-board/internal/config"
	"pm-board/internal/editor"
	"pm-board/internal/write"
)

// pane is which of the three boxes has the focus.
type pane int

const (
	paneOpen pane = iota
	paneDone
	paneDetail
)

// The tabs of pane [1]. Every kind of file is its own tab, because a plan is
// an item of its own even when it carries the tasks of a spec.
const (
	tabSpecs = iota
	tabPlans
	tabTasks
	tabBugs
)

var (
	tabKinds = [4]board.Kind{board.KindStory, board.KindPlan, board.KindTask, board.KindBug}
	tabNames = [4]string{"Specs", "Plans", "Tasks", "Bugs"}
)

// doneTab is one tab of pane [2]: the name in its title and the status it
// lists. The second entry of a kind with only one finished status is empty
// and never shows.
type doneTab struct {
	name   string
	status string
}

// Pane [2] follows the tab of pane [1]. A task is only ever done, so it has
// no second finished tab.
var doneTabs = [4][2]doneTab{
	{{"Done", "done"}, {"Dropped", "dropped"}},
	{{"Done", "done"}, {"Dropped", "dropped"}},
	{{"Done", "done"}, {}},
	{{"Fixed", "fixed"}, {"Wontfix", "wontfix"}},
}

// rowLines is how many screen lines one row takes: the title, the dim meta
// line under it, and a blank line between rows.
const rowLines = 3

// pageLines is how far ctrl+d, ctrl+u and the wheel jump in pane [3].
const pageLines = 10

// groupRowID marks the folded row that holds the legacy items.
const groupRowID = "\x00untyped"

type row struct {
	id    string
	group bool
	depth int
}

// tabBox is where one tab name sits in the top border of a pane, in screen
// cells, the way the mouse counts them.
type tabBox struct{ x, w int }

// box is where one pane sits on the screen and what it holds. first and rows
// belong to the two list panes; pane [3] scrolls by lines and counts its own.
type box struct {
	x, y, w, h  int // the border rectangle
	inner       int
	first, rows int      // first row on screen, and how many rows fit inside
	tabs        []tabBox // where each tab name sits in the top border
}

// geom is the whole screen measured in cells from its top left corner. The
// view and the mouse both read it, so a click lands where it looks.
type geom struct {
	wide   bool // false below 60 columns, where only the focused pane shows
	leftW  int  // width of the left column
	open   box
	done   box
	detail box
	full   box // the focused pane stretched over the whole screen when not wide
}

type popup struct {
	field   string
	options []string
	idx     int
}

type reloadMsg struct {
	b   *board.Board
	err error
}

type watchFailedMsg struct{ err error }

// clockMsg carries the time once a minute, for the clock in the status line.
type clockMsg time.Time

type editorDoneMsg struct {
	newBug string // path of a new bug file; "" after a plain edit
	tmpl   []byte
	err    error
}

// WatchFailed tells the model live reload could not start.
func WatchFailed(err error) tea.Msg { return watchFailedMsg{err: err} }

// Model is the whole screen state. Bubble Tea copies it on every update.
type Model struct {
	cfg         config.Config
	board       *board.Board
	focus       pane // the pane with the focus
	last        pane // the list pane that had it last, for pane [3]
	tab         int  // which tab pane [1] shows
	doneTab     int  // which tab pane [2] shows
	sel         [4]string
	idx         [4]int // selected row number per tab, used when the id vanishes
	doneSel     [4][2]string
	doneIdx     [4][2]int
	query       string
	searching   bool
	groupOpen   bool
	popup       *popup
	slug        *string // non-nil while typing the slug of a new bug
	help        bool
	status      string
	manual      bool
	width       int
	height      int
	scroll      int // how far the body of pane [3] is scrolled
	now         time.Time
	version     string                 // build version shown on the bottom line
	open        func(url string) error // opens a link in the browser
	statusBoxes []statusPiece          // click boxes of the bottom line links

	load     func() (*board.Board, error)
	setValue func(id, field, value string) (write.Outcome, error)
	render   func(md string, width int) string
}

// New builds a model over b. dark picks the markdown style; ask the terminal
// before the program starts, because asking later fights Bubble Tea for stdin.
func New(cfg config.Config, b *board.Board, dark bool) Model {
	return Model{
		cfg: cfg, board: b, width: 120, height: 40, now: time.Now(), version: "dev",
		open: defaultOpen,
		load: func() (*board.Board, error) { return board.Load(cfg) },
		// Load fresh so the write never checks against a stale board.
		setValue: func(id, field, value string) (write.Outcome, error) {
			fresh, err := board.Load(cfg)
			if err != nil {
				return write.Outcome{}, err
			}
			return write.SetValue(cfg, fresh, id, field, value)
		},
		render: newRenderer(dark),
	}
}

// WithLoad replaces how the model reloads the board, for example to read
// every worktree of the repo.
func (m Model) WithLoad(f func() (*board.Board, error)) Model {
	m.load = f
	return m
}

func (m Model) Init() tea.Cmd { return nextMinute() }

// nextMinute waits for the next minute to start and sends the time then, so
// the clock in the status line moves once a minute.
func nextMinute() tea.Cmd {
	wait := time.Until(time.Now().Truncate(time.Minute).Add(time.Minute))
	return tea.Tick(wait, func(t time.Time) tea.Msg { return clockMsg(t) })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case reloadMsg:
		if msg.err != nil {
			m.status = "reload failed: " + msg.err.Error()
			return m, nil
		}
		m.board = msg.b
		m.moveTo(m.cursor())
	case watchFailedMsg:
		m.manual = true
		m.status = "live reload off: " + msg.err.Error() + " (press r)"
	case clockMsg:
		m.now = time.Time(msg)
		return m, nextMinute()
	case editorDoneMsg:
		return m.afterEditor(msg)
	case tea.KeyMsg:
		return m.key(msg)
	case tea.MouseMsg:
		return m.mouse(msg)
	}
	return m, nil
}

// openRows gives pane [1] the items of its tab that are not finished. The
// files outside .pm/ have no tab of their own, so the Specs tab ends with a
// single row that opens them.
func (m Model) openRows() []row {
	if m.query != "" {
		return toRows(m.board.Search(m.query), 0)
	}
	rows := toRows(m.board.List(tabKinds[m.tab], false), 0)
	if m.tab == tabSpecs {
		if legacy := m.board.Untyped(false); len(legacy) > 0 {
			rows = append(rows, row{id: groupRowID, group: true})
			if m.groupOpen {
				rows = append(rows, toRows(legacy, 1)...)
			}
		}
	}
	return rows
}

// doneRows gives pane [2] the finished items of the tab it follows, the ones
// touched last at the top: the date of the last commit on the file decides, and
// the date in the file name is the fallback for a file git has no commit for.
func (m Model) doneRows() []row {
	d := doneTabs[m.tab][m.doneTab]
	if d.name == "" {
		return nil
	}
	var finished []*board.Item
	for _, it := range m.board.List(tabKinds[m.tab], true) {
		if it.Status == d.status {
			finished = append(finished, it)
		}
	}
	sort.SliceStable(finished, func(i, j int) bool {
		return finished[i].SortTime() > finished[j].SortTime()
	})
	return toRows(finished, 0)
}

// doneTabNames gives the tab names pane [2] shows for the tab of pane [1].
func (m Model) doneTabNames() []string {
	var out []string
	for _, d := range doneTabs[m.tab] {
		if d.name != "" {
			out = append(out, d.name)
		}
	}
	return out
}

func toRows(items []*board.Item, depth int) []row {
	rows := make([]row, 0, len(items))
	for _, it := range items {
		rows = append(rows, row{id: it.ID, depth: depth})
	}
	return rows
}

// listPane is the pane whose selection the screen shows: the focused list
// pane, or the one that had the focus last when pane [3] has it.
func (m Model) listPane() pane {
	if m.focus == paneDetail {
		return m.last
	}
	return m.focus
}

// slotOf hands back a pane's rows and the two places its cursor is kept, so
// every caller moves the same memory.
func (m *Model) slotOf(p pane) ([]row, *string, *int) {
	if p == paneDone {
		return m.doneRows(), &m.doneSel[m.tab][m.doneTab], &m.doneIdx[m.tab][m.doneTab]
	}
	return m.openRows(), &m.sel[m.tab], &m.idx[m.tab]
}

func (m *Model) slot() ([]row, *string, *int) { return m.slotOf(m.listPane()) }

func (m Model) listOf() []row {
	rows, _, _ := m.slot()
	return rows
}

// Selected is the item under the cursor of the list pane the screen shows, or
// nil on the group row and on an empty list.
func (m Model) Selected() *board.Item {
	rows, sel, idx := m.slot()
	i := cursorOf(rows, *sel, *idx)
	if i < 0 || rows[i].group {
		return nil
	}
	return m.board.Get(rows[i].id)
}

// cursorOf finds the selected id; when it is gone it falls back to the old
// row number.
func cursorOf(rows []row, sel string, idx int) int {
	if len(rows) == 0 {
		return -1
	}
	for i, r := range rows {
		if r.id == sel {
			return i
		}
	}
	return clamp(idx, 0, len(rows)-1)
}

func (m Model) cursor() int {
	rows, sel, idx := m.slot()
	return cursorOf(rows, *sel, *idx)
}

func (m *Model) moveTo(i int) {
	rows, sel, idx := m.slot()
	if len(rows) == 0 {
		*sel, *idx = "", 0
		return
	}
	i = clamp(i, 0, len(rows)-1)
	if rows[i].id != *sel {
		// Another item in pane [3] starts at its own top.
		m.scroll = 0
	}
	*sel, *idx = rows[i].id, i
}

func (m *Model) moveBy(lines int) { m.moveTo(m.cursor() + lines) }

// focusPane moves the focus and remembers which list pane had it, so pane [3]
// keeps showing that one.
func (m *Model) focusPane(p pane) {
	if m.focus == p {
		return
	}
	m.focus = p
	if p != paneDetail {
		m.last = p
	}
	m.scroll = 0
}

// cycleTab walks the focused pane along its own tabs and wraps around. Pane
// [3] has no tabs, so nothing happens there.
func (m *Model) cycleTab(step int) {
	switch m.focus {
	case paneOpen:
		m.tab = (m.tab + step + len(tabNames)) % len(tabNames)
	case paneDone:
		n := len(m.doneTabNames())
		m.doneTab = (m.doneTab + step + n) % n
	default:
		return
	}
	// Tasks have one finished tab, so coming from Bugs the second is gone.
	m.doneTab = min(m.doneTab, len(m.doneTabNames())-1)
}

// step moves the cursor in a list pane and scrolls the body in pane [3], so
// one set of keys works wherever the focus is.
func (m *Model) step(lines int) {
	if m.focus == paneDetail {
		m.scroll = max(0, m.scroll+lines)
		return
	}
	m.moveBy(lines)
}

func (m *Model) top() {
	if m.focus == paneDetail {
		m.scroll = 0
		return
	}
	m.moveTo(0)
}

// end goes to the last row, or to the bottom of the body in pane [3]. The
// view cuts that jump to the last line it drew, so the model never needs to
// know how long the body is.
func (m *Model) end() {
	if m.focus == paneDetail {
		m.scroll = math.MaxInt
		return
	}
	m.moveTo(len(m.listOf()) - 1)
}

func (m Model) key(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.help {
		// The help sits over the panes, so it takes the keys until it closes.
		if s := k.String(); s == "?" || s == "esc" {
			m.help = false
		}
		return m, nil
	}
	switch {
	case m.popup != nil:
		return m.popupKey(k)
	case m.slug != nil:
		return m.slugKey(k)
	case m.searching:
		return m.searchKey(k)
	}
	switch k.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "1", "2", "3":
		m.focusPane(pane(k.String()[0] - '1'))
	case "tab":
		m.focusPane(pane((int(m.focus) + 1) % 3))
	case "shift+tab":
		m.focusPane(pane((int(m.focus) + 2) % 3))
	case "]":
		m.cycleTab(1)
	case "[":
		m.cycleTab(-1)
	case "j", "down":
		m.step(1)
	case "k", "up":
		m.step(-1)
	case "g":
		m.top()
	case "G":
		m.end()
	case "ctrl+d":
		m.step(pageLines)
	case "ctrl+u":
		m.step(-pageLines)
	case "/":
		m.searching = true
	case "esc":
		m.query = ""
		m.moveTo(0)
	case "?":
		m.help = true
	case "r":
		return m, m.reloadCmd()
	case "t", "s":
		m.openPopup(k.String())
	case "n":
		s := ""
		m.slug = &s
	case "enter":
		return m.enter()
	}
	return m, nil
}

// mouse follows a click or a wheel turn. A click on a row selects it, a click
// on a tab name switches tab, and the wheel works on the pane under the
// pointer, so the mouse and the keys always leave the same cursor.
func (m Model) mouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.help || m.popup != nil || m.slug != nil || m.searching {
		return m, nil
	}
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft &&
		msg.Y == m.height-1 {
		if url, ok := m.linkAt(msg.X); ok {
			open := m.open
			if open == nil {
				open = defaultOpen
			}
			_ = open(url)
		}
		return m, nil
	}
	p, rowIdx, tabIdx := m.hit(msg.X, msg.Y)
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		m.focusPane(p)
		m.step(-1)
		return m, nil
	case tea.MouseButtonWheelDown:
		m.focusPane(p)
		m.step(1)
		return m, nil
	}
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}
	m.focusPane(p)
	switch {
	case tabIdx >= 0:
		m.clickTab(p, tabIdx)
	case rowIdx >= 0:
		m.moveTo(rowIdx)
	}
	return m, nil
}

// clickTab switches the pane the click landed in to the tab that was hit.
func (m *Model) clickTab(p pane, idx int) {
	switch p {
	case paneDone:
		m.doneTab = clamp(idx, 0, len(m.doneTabNames())-1)
	case paneOpen:
		m.tab = clamp(idx, 0, len(tabNames)-1)
	}
}

// geometry measures the screen. The left column is 30% of the width, held
// between 28 and 48 columns, and below 60 columns only the focused pane is on
// screen because two columns of a small terminal fit nothing.
func (m Model) geometry() geom {
	// The last line belongs to the status line, so the panes share the rest.
	bodyH := max(3, m.height-1)
	top, bottom := split(bodyH)
	g := geom{wide: m.width >= 60, leftW: clamp(m.width*3/10, 28, 48)}
	rest := max(0, m.width-g.leftW)
	g.open = m.box(paneOpen, 0, 0, g.leftW, top)
	g.done = m.box(paneDone, 0, top, g.leftW, bottom)
	// A pane of zero width is not on screen, which is how a narrow terminal
	// leaves two of the three out.
	g.detail = m.box(paneDetail, g.leftW, 0, rest, bodyH)
	if g.wide {
		return g
	}
	g.full = m.box(m.focus, 0, 0, m.width, bodyH)
	// Only the focused pane is on screen now, so the other two are zero boxes
	// and neither the view nor the mouse finds them.
	g.open, g.done, g.detail = box{}, box{}, box{}
	return g
}

// split divides the height between the two left panes: about two thirds for
// the open one, and never fewer than five rows each.
func split(h int) (int, int) {
	top := max(5, h*2/3)
	return top, max(5, h-top)
}

// box measures one pane at the given rectangle and works out which of its
// rows the screen shows, so the view and the mouse count the same ones.
func (m Model) box(p pane, x, y, w, h int) box {
	b := box{x: x, y: y, w: w, h: h, inner: max(0, h-2), tabs: tabX(p, x, m.tabsOf(p))}
	if p == paneDetail {
		return b
	}
	b.rows = b.inner / rowLines
	rows, sel, idx := m.slotOf(p)
	b.first = max(0, cursorOf(rows, *sel, *idx)+1-b.rows)
	return b
}

// tabsOf gives the tab names a pane shows in its title.
func (m Model) tabsOf(p pane) []string {
	switch p {
	case paneDone:
		return m.doneTabNames()
	case paneDetail:
		return nil
	}
	return tabNames[:]
}

// rowAt gives the row under a screen line, or -1 for the border, the title and
// the blank line between rows.
func (b box) rowAt(y int) int {
	off := y - b.y - 1
	if off < 0 || b.rows == 0 {
		return -1
	}
	i := off / rowLines
	if i >= b.rows || off%rowLines > 1 {
		return -1
	}
	return b.first + i
}

// hit says what sits under a mouse cell: the pane, the row of it and the tab
// in its title. A cell on none of them answers the focused pane with no row
// and no tab, which leaves the screen where it was.
func (m Model) hit(x, y int) (pane, int, int) {
	g := m.geometry()
	boxes := []struct {
		p pane
		b box
	}{{paneOpen, g.open}, {paneDone, g.done}, {paneDetail, g.detail}}
	if !g.wide {
		boxes = []struct {
			p pane
			b box
		}{{m.focus, g.full}}
	}
	for _, e := range boxes {
		if x < e.b.x || x >= e.b.x+e.b.w || y < e.b.y || y >= e.b.y+e.b.h {
			continue
		}
		if y == e.b.y {
			for i, t := range e.b.tabs {
				if x >= t.x && x < t.x+t.w {
					return e.p, -1, i
				}
			}
		}
		return e.p, e.b.rowAt(y), -1
	}
	return m.focus, -1, -1
}

func (m *Model) openPopup(key string) {
	it := m.Selected()
	switch {
	case it == nil:
		m.status = "nothing selected"
		return
	case it.Kind == board.KindTask:
		m.status = "tasks take their status from their checkboxes"
		return
	case it.Worktree != "":
		m.status = "shown from worktree " + it.Worktree + "; edit it there"
		return
	case it.Legacy:
		m.status = "legacy file, move it into the root folder first"
		return
	}
	p := popup{field: "status", options: board.Allowed(it.Kind)}
	current := it.Status
	if key == "t" {
		p = popup{field: "type", options: []string{string(board.KindStory), string(board.KindBug)}}
		current = string(it.Kind)
	}
	for i, o := range p.options {
		if o == current {
			p.idx = i
		}
	}
	m.popup = &p
}

func (m Model) popupKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	p := *m.popup
	switch k.String() {
	case "j", "down":
		p.idx = min(p.idx+1, len(p.options)-1)
	case "k", "up":
		p.idx = max(p.idx-1, 0)
	case "esc":
		m.popup = nil
		return m, nil
	case "enter":
		m.popup = nil
		it := m.Selected()
		if it == nil {
			return m, nil
		}
		value := p.options[p.idx]
		o, err := m.setValue(it.ID, p.field, value)
		m.status = outcomeText(fmt.Sprintf("pm: %s %s %s", it.ID, p.field, value), o, err)
		return m, m.reloadCmd()
	}
	m.popup = &p
	return m, nil
}

func (m Model) slugKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	s := *m.slug
	switch k.Type {
	case tea.KeyEsc:
		m.slug = nil
		return m, nil
	case tea.KeyBackspace:
		if s != "" {
			s = s[:len(s)-1]
		}
	case tea.KeyEnter:
		m.slug = nil
		path, tmpl, err := write.StartBug(m.cfg, s, "")
		if err != nil {
			m.status = "error: " + err.Error()
			return m, nil
		}
		return m, tea.ExecProcess(editor.Cmd(path, 1), func(err error) tea.Msg {
			return editorDoneMsg{newBug: path, tmpl: tmpl, err: err}
		})
	case tea.KeyRunes:
		// A slug is lower case letters, digits and dashes; anything else is dropped.
		for _, r := range k.Runes {
			if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
				s += string(r)
			}
		}
	}
	m.slug = &s
	return m, nil
}

func (m Model) searchKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.Type {
	case tea.KeyEsc:
		m.searching, m.query = false, ""
	case tea.KeyEnter:
		m.searching = false
	case tea.KeyBackspace:
		if r := []rune(m.query); len(r) > 0 {
			m.query = string(r[:len(r)-1])
		}
	case tea.KeySpace:
		m.query += " "
	case tea.KeyRunes:
		m.query += string(k.Runes)
	}
	m.moveTo(0)
	return m, nil
}

func (m Model) enter() (tea.Model, tea.Cmd) {
	rows := m.listOf()
	i := m.cursor()
	if i < 0 {
		return m, nil
	}
	if rows[i].group {
		m.groupOpen = !m.groupOpen
		return m, nil
	}
	it := m.board.Get(rows[i].id)
	if !it.OnDisk {
		m.status = "branch " + it.Worktree + " is not checked out; open it with: git worktree add ../<repo>-" + it.Worktree + " " + it.Worktree
		return m, nil
	}
	return m, tea.ExecProcess(editor.Cmd(it.Path, it.Line), func(err error) tea.Msg {
		return editorDoneMsg{err: err}
	})
}

func (m Model) afterEditor(msg editorDoneMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.status = "editor: " + msg.err.Error()
	}
	if msg.newBug != "" {
		o, err := write.FinishBug(m.cfg, msg.newBug, msg.tmpl)
		if errors.Is(err, write.ErrUnchanged) {
			m.status = "bug not saved: template left unchanged"
		} else {
			m.status = outcomeText("pm: new bug", o, err)
		}
	}
	return m, m.reloadCmd()
}

func (m Model) reloadCmd() tea.Cmd {
	load := m.load
	return func() tea.Msg {
		b, err := load()
		return reloadMsg{b: b, err: err}
	}
}

func outcomeText(label string, o write.Outcome, err error) string {
	switch {
	case err != nil:
		return "error: " + err.Error()
	case o.Committed:
		return label + " ✓ committed"
	case o.Skipped:
		return "written, not committed: " + o.Reason
	default:
		return "written (auto_commit off)"
	}
}

func clamp(v, lo, hi int) int { return max(lo, min(v, hi)) }
