// Package tui is the lazygit-style screen over a Board.
package tui

import (
	"errors"
	"fmt"
	"slices"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/editor"
	"github.com/iyay/acta/internal/theme"
	"github.com/iyay/acta/internal/write"
)

// pane is which of the boxes on screen has the focus. The boxes of a tab and
// the pane numbers come from the table in sidebar.go, so no box is named twice.
type pane int

// rowLines is how many screen lines one row takes. A row is one line, so the
// screen line of a row and the row itself are the same number.
const rowLines = 1

// pageLines is how far ctrl+d, ctrl+u and the wheel jump in the detail box.
const pageLines = 10

// groupRowID marks the folded row that holds the legacy items.
const groupRowID = "\x00untyped"

type row struct {
	id    string
	group bool
	depth int
	tree  bool // a row of a plans tree, which space and enter act on
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

// clearStatusMsg hides a toast. It carries the text it was sent for, so a
// newer message that took the line in the meantime stays.
type clearStatusMsg struct{ text string }

// toastFor is how long any message stays on the status line, errors included,
// so the key hints behind it come back by themselves.
const toastFor = 2 * time.Second

// clearStatusAfter sends the clear message for text once d has passed.
func clearStatusAfter(d time.Duration, text string) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return clearStatusMsg{text: text} })
}

// wheelStep is how many lines one wheel notch moves, the usual terminal step.
const wheelStep = 3

// wheelFrame is how long wheel notches are gathered before one scroll. A
// trackpad sends hundreds of notches a second, and drawing the screen after
// each one left the TUI far behind the wheel. It is shorter than two renderer
// writes at 120fps, so a steady scroll never skips a write.
const wheelFrame = 12 * time.Millisecond

// wheelTickMsg says the frame is over: scroll by what the wheel gathered.
type wheelTickMsg struct{}

// frameCache keeps the last frame View drew, and whether a pulse dot was on
// it, so the pulse tick knows if drawing again would change anything. A
// trackpad sends hundreds of notches a second, and most of them only add to
// the pending delta, so the screen stays the same. Drawing it again each time
// kept the TUI busy. The model is copied on every update, so the frame sits
// behind a pointer that all the copies share.
type frameCache struct {
	s    string
	dots bool
}

// pulseMsg moves the pulse of the dots of work under way one frame on.
type pulseMsg struct{}

// pulseStep is the time one pulse frame stays on screen, so eight frames make
// one breath of about a second.
const pulseStep = 120 * time.Millisecond

// pulseAfter sends the next pulse once d has passed.
func pulseAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return pulseMsg{} })
}

// hasWork says if any item on the board is under way, the only time a dot
// can pulse at all.
func (m Model) hasWork() bool {
	if m.board == nil {
		return false
	}
	for _, it := range m.board.Items {
		if inProgress(it) {
			return true
		}
	}
	return false
}

// armPulse starts the pulse chain when there is work under way and no pulse
// is on its way already, so two chains never run at once.
func (m *Model) armPulse() tea.Cmd {
	if m.pulsing || !m.hasWork() {
		return nil
	}
	m.pulsing = true
	return pulseAfter(pulseStep)
}

// wheelMark is the screen a frame of notches was gathered on: the open tab, its
// Done sub-tab, the search and the row under the cursor. A tick scrolls only
// while all of it still stands, so notches meant for one item never land on
// another.
type wheelMark struct {
	top, done int
	query     string
	sel       string
}

type editorDoneMsg struct {
	newBug string // path of a new bug file; "" after a plain edit
	tmpl   []byte
	err    error
}

// WatchFailed tells the model live reload could not start.
func WatchFailed(err error) tea.Msg { return watchFailedMsg{err: err} }

// Model is the whole screen state. Bubble Tea copies it on every update.
type Model struct {
	cfg        config.Config
	board      *board.Board
	focus      pane                   // the box with the focus
	last       pane                   // the list pane that had it last, for the detail box
	top        int                    // the open tab, an index into topTabs
	tabs       [len(topTabs)]tabState // the saved place of each tab that is not open
	done       int                    // the Done sub-tab the open tab shows
	newest     []bool                 // per list box: true shows the newest file date first
	sel        []string
	idx        []int // selected row number per pane, used when the id vanishes
	query      string
	expanded   int // the list pane that takes the room, -1 when none does
	searching  bool
	groupOpen  bool
	openPlans  map[string]bool // the plans the reader opened in a tree list
	shutActs   map[string]bool // the groups the reader shut on Activities; every group starts open
	popup      *popup
	slug       *string // non-nil while typing the slug of a new bug
	help       bool
	status     string
	manual     bool
	width      int
	height     int
	off        []int // the first line each box shows, the detail box last
	wheelPane  pane  // the pane the gathered notches scroll
	wheelDelta int   // lines gathered from the wheel, not yet scrolled
	wheelArmed bool  // true while a frame tick is on its way
	wheelMoved bool  // a notch scrolled at once, so the closing tick must draw
	wheelMark  wheelMark
	drag       drag // the text a mouse drag selected; empty when none
	now        time.Time
	version    string                  // build version shown on the bottom line
	open       func(url string) error  // opens a link in the browser
	clip       func(text string) error // puts text on the clipboard

	load     func() (*board.Board, error)
	setValue func(id, field, value string) (write.Outcome, error)
	markItem func(id string, done bool) (write.Outcome, error)
	render   func(md string, width int) string
	dcache   *detailCache // the last detail lines, shared by every copy
	frame    *frameCache  // the last frame View drew, shared by every copy
	trace    *Tracer      // notes wheel notches and draws; nil when off
	same     bool         // true when the last message changed nothing on screen
	pulse    int          // the pulse frame the dots of work under way wear now
	pulsing  bool         // true while a pulse message is on its way
	styles   styles       // the brushes every screen is painted with
}

// New builds a model over b. dark picks the markdown style; ask the terminal
// before the program starts, because asking later fights Bubble Tea for stdin.
func New(cfg config.Config, b *board.Board, dark bool) Model {
	s := freshTab()
	t, _ := theme.Builtin(theme.Default)
	return Model{
		cfg: cfg, board: b, width: 120, height: 40, now: time.Now(), version: "dev",
		open:     defaultOpen,
		clip:     copyText,
		top:      tabActivities,
		focus:    s.focus,
		last:     s.last,
		sel:      s.sel,
		idx:      s.idx,
		off:      s.off,
		newest:   make([]bool, sidePanes),
		expanded: -1,
		load:     func() (*board.Board, error) { return board.Load(cfg) },
		// Load fresh so the write never checks against a stale board.
		setValue: func(id, field, value string) (write.Outcome, error) {
			fresh, err := board.Load(cfg)
			if err != nil {
				return write.Outcome{}, err
			}
			return write.SetValue(cfg, fresh, id, field, value)
		},
		// Load fresh so the write never checks against a stale board.
		markItem: func(id string, done bool) (write.Outcome, error) {
			fresh, err := board.Load(cfg)
			if err != nil {
				return write.Outcome{}, err
			}
			return write.MarkItem(cfg, fresh, id, done)
		},
		styles: newStyles(t, dark),
		render: newRenderer(t.Dark(dark)),
		dcache: &detailCache{},
		frame:  &frameCache{},
	}
}

// WithTheme picks the colors. A theme that does not load must not keep the
// board from opening, so it falls back to the default and says why.
func (m Model) WithTheme(name string, dark bool) Model {
	t, err := theme.Load(name)
	if err != nil {
		t, _ = theme.Builtin(theme.Default)
		m.status = err.Error()
	}
	m.styles = newStyles(t, dark)
	m.render = newRenderer(t.Dark(dark))
	m.dcache = &detailCache{}
	m.frame = &frameCache{}
	return m
}

// WithLoad replaces how the model reloads the board, for example to read
// every worktree of the repo.
func (m Model) WithLoad(f func() (*board.Board, error)) Model {
	m.load = f
	return m
}

// WithTrace makes the model note its wheel notches and draws in t. A nil t
// keeps the note off, which is how the model runs when no trace is asked for.
func (m Model) WithTrace(t *Tracer) Model {
	m.trace = t
	return m
}

// Init starts the clock, and the hide timer for a message the model was built
// with, such as a theme that did not load, because no Update saw it.
func (m Model) Init() tea.Cmd {
	if m.status == "" {
		return nextMinute()
	}
	return tea.Batch(nextMinute(), clearStatusAfter(toastFor, m.status))
}

// nextMinute waits for the next minute to start and sends the time then, so
// the clock in the status line moves once a minute.
func nextMinute() tea.Cmd {
	wait := time.Until(time.Now().Truncate(time.Minute).Add(time.Minute))
	return tea.Tick(wait, func(t time.Time) tea.Msg { return clockMsg(t) })
}

// update runs one message and hands back the new model. Update wraps it to
// start the timer for the status line.
func (m Model) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Every message may change the screen. The few that surely do not set
	// this back below, so a path nobody thought about always draws again.
	m.same = false
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width == m.width && msg.Height == m.height {
			break
		}
		m.width, m.height = msg.Width, msg.Height
		// The picked cells were on the old screen, so they point at nothing now.
		m.drag = drag{}
		// A box that just got shorter cannot keep the place it had.
		for p := pane(0); int(p) < boxes; p++ {
			m.clampOff(p)
		}
		// The renderer only paints the cells the new frame uses, so a window
		// that shrinks leaves the rest of the old frame on screen. Wiping it
		// is the only way to take those rows back.
		if c := m.armPulse(); c != nil {
			return m, tea.Batch(tea.ClearScreen, c)
		}
		return m, tea.ClearScreen
	case reloadMsg:
		if msg.err != nil {
			m.status = "reload failed: " + msg.err.Error()
			return m, nil
		}
		// A tab whose cursor was never moved shows the row under it, so put
		// that item down before the board changes. The reload can then tell
		// the item the reader is on from another one, and keeps the scroll
		// they made when the item on show really is the same.
		if rows, sel, idx := m.slot(); *sel == "" && len(rows) > 0 {
			*sel = rows[cursorOf(rows, *sel, *idx)].id
		}
		m.board = msg.b
		m.moveTo(m.cursor())
		return m, m.armPulse()
	case pulseMsg:
		m.pulsing = false
		if !m.hasWork() {
			m.same = true
			return m, nil
		}
		m.pulsing = true
		if m.frame == nil || !m.frame.dots {
			// No dot on screen: nothing to draw, but work may show again.
			m.same = true
			return m, pulseAfter(pulseStep)
		}
		m.pulse = (m.pulse + 1) % pulseSteps
		return m, pulseAfter(pulseStep)
	case watchFailedMsg:
		m.manual = true
		m.status = "live reload off: " + msg.err.Error() + " (press r)"
	case clockMsg:
		m.now = time.Time(msg)
		return m, nextMinute()
	case clearStatusMsg:
		if m.status == msg.text {
			m.status = ""
		}
	case wheelTickMsg:
		// Notches belong to the screen they were gathered on, so a reader who
		// moved the item, the tab, the sub-tab or the search before the frame
		// ended never sees them land on the new one.
		if m.wheelDelta != 0 && m.wheelMark == m.mark() {
			m.scrollPane(m.wheelPane, m.wheelDelta)
			m.wheelDelta = 0
			// The wheel is still turning, so the next frame needs its tick.
			m.wheelMoved = false
			return m, wheelTickCmd()
		}
		// A notch that scrolled at once has not been written down yet, so
		// this tick still has to write it. A tick after a wheel that did
		// nothing may keep the frame the reader already has.
		m.same = !m.wheelMoved
		m.wheelMoved = false
		m.wheelDelta, m.wheelArmed = 0, false
	case editorDoneMsg:
		return m.afterEditor(msg)
	case tea.KeyMsg:
		return m.key(msg)
	case tea.MouseMsg:
		return m.mouse(msg)
	}
	return m, nil
}

// Update runs one message. When the message leaves a new text on the status
// line, a timer is started for it here, in one place, so every toast hides
// by itself and the key hints come back.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	before := m.status
	next, cmd := m.update(msg)
	after := next.(Model).status
	if after == "" || after == before {
		return next, cmd
	}
	return next, tea.Batch(cmd, clearStatusAfter(toastFor, after))
}

// inProgress says the item's work has begun: one of the going statuses, a
// task someone ran tick --start on before ticking a box, or a scratch idea
// someone is still thinking about.
func inProgress(it *board.Item) bool {
	if it == nil {
		return false
	}
	if it.Started {
		return true
	}
	switch it.Status {
	case "in-progress", "fixing":
		return true
	}
	return it.Kind == board.KindScratch && it.Status == "brainstorming"
}

func toRows(items []*board.Item, depth int) []row {
	rows := make([]row, 0, len(items))
	for _, it := range items {
		rows = append(rows, row{id: it.ID, depth: depth})
	}
	return rows
}

// slotOf hands back a pane's rows and the two places its cursor is kept, so
// every caller moves the same memory.
func (m *Model) slotOf(p pane) ([]row, *string, *int) {
	return m.rowsOf(p), &m.sel[p], &m.idx[p]
}

// listPane is the pane whose selection the screen shows: the focused list
// pane, or the one that had the focus last when the detail box has it.
func (m Model) listPane() pane {
	if m.focus == paneDetail {
		return m.last
	}
	return m.focus
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

// mark is the screen as it stands: the open tab, its Done sub-tab, the search
// and the row under the cursor. A wheel frame compares it with the one it
// gathered its notches on, so the notches follow the reader and never land on
// a screen they did not come from. It reads the cursor the model keeps rather
// than the rows on show, so a notch costs no more than the frame it is part of.
func (m Model) mark() wheelMark {
	return wheelMark{top: m.top, done: m.done, query: m.query, sel: m.sel[m.listPane()]}
}

// cursorOf finds the selected id; when it is gone it falls back to the old
// row number.
func cursorOf(rows []row, sel string, idx int) int {
	for i, r := range rows {
		if r.id == sel {
			return i
		}
	}
	if len(rows) == 0 {
		return -1
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
		// Another item starts the detail body at its own top.
		m.off[paneDetail] = 0
	}
	*sel, *idx = rows[i].id, i
	m.keepVisible(m.listPane())
}

func (m *Model) moveBy(lines int) { m.moveTo(m.cursor() + lines) }

// focusPane moves the focus. The detail box shows the list box that had it
// last. A tab with no Done pane never gives it the focus.
func (m *Model) focusPane(p pane) {
	if m.focus == p || !slices.Contains(append(m.panes(), paneDetail), p) {
		return
	}
	// The room belongs to the box that has the focus, so a box that is no
	// longer the one on top of the screen cannot keep it.
	m.expanded = -1
	m.focus = p
	if p != paneDetail {
		m.last = p
	}
	m.keepVisible(p)
}

// onPane says whether a cell sits inside the box a pane is drawn in. hit
// answers the focused pane for a cell that lands on no pane at all, so the
// notch asks the geometry again before it moves a word.
func (m Model) onPane(p pane, x, y int) bool {
	g := m.geometry()
	b := g.full
	if g.wide {
		b = g.at(p)
	}
	return x >= b.x && x < b.x+b.w && y >= b.y && y < b.y+b.h
}

// cycleTab walks the Done pane along its sub-tabs and wraps around. The other
// boxes have no sub-tabs, so [ and ] do nothing there.
func (m *Model) cycleTab(step int) {
	n := len(m.doneTabNames())
	if m.focus != paneDone || n == 0 {
		return
	}
	m.done = (m.done + step + n) % n
	m.keepVisible(paneDone)
}

// step moves the cursor in a list box and scrolls the body in the detail box,
// so one set of keys works wherever the focus is.
func (m *Model) step(lines int) {
	if m.focus == paneDetail {
		m.scrollPane(m.focus, lines)
		return
	}
	m.moveBy(lines)
}

// goTop goes to the first row, or to the top of the body in the detail box.
func (m *Model) goTop() {
	if m.focus == paneDetail {
		m.off[paneDetail] = 0
		return
	}
	m.moveTo(0)
}

// end goes to the last row, or to the bottom of the body in the detail box,
// which is the last line that box has room for.
func (m *Model) end() {
	if m.focus == paneDetail {
		m.off[paneDetail] = m.lastOff(paneDetail)
		return
	}
	m.moveTo(len(m.listOf()) - 1)
}

func (m Model) key(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Any key may move the screen, so the picked cells stop meaning anything.
	m.drag = drag{}
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
	case "1", "2", "3", "4", "5", "6":
		m.openTab(int(k.String()[0] - '1'))
	case "left":
		m.openTab((m.top + len(topTabs) - 1) % len(topTabs))
	case "right":
		m.openTab((m.top + 1) % len(topTabs))
	case "tab":
		m.cyclePane(1)
	case "shift+tab":
		m.cyclePane(-1)
	case "z":
		m.toggleExpand()
	case "]":
		m.cycleTab(1)
	case "[":
		m.cycleTab(-1)
	case "j", "down":
		m.step(1)
	case "k", "up":
		m.step(-1)
	case "g":
		m.goTop()
	case "G":
		m.end()
	case "ctrl+d":
		m.step(pageLines)
	case "ctrl+u":
		m.step(-pageLines)
	case "/":
		m.searching = true
	case "esc":
		if m.focus == paneDetail {
			// Back to the list pane the detail came from, on the same item.
			m.focusPane(m.last)
			return m, nil
		}
		m.query = ""
		m.moveTo(0)
	case "?":
		m.help = true
	case "r":
		return m, m.reloadCmd()
	case "t", "s", "p":
		m.openPopup(k.String())
	case "+":
		return m.markRow(true)
	case "-":
		return m.markRow(false)
	case "o":
		// The detail box has no list to sort.
		if m.focus != paneDetail {
			m.newest[m.focus] = !m.newest[m.focus]
		}
	case "n":
		s := ""
		m.slug = &s
	case "e":
		return m.edit()
	case "y":
		// Copy first, then return, so the status the timer carries is the
		// one the copy just set.
		cmd := m.copyID()
		return m, cmd
	case " ":
		m.toggleRow()
	case "h":
		m.foldRow(false)
	case "l":
		m.foldRow(true)
	case "enter":
		if m.toggleRow() {
			return m, nil
		}
		return m.enter()
	}
	return m, nil
}

// mouse follows a click or a wheel turn. A click on a row selects it, a click
// on a tab name switches tab, and the wheel works on the box under the
// pointer, so the mouse and the keys always leave the same cursor.
func (m Model) mouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.help || m.popup != nil || m.slug != nil || m.searching {
		m.same = true
		return m, nil
	}
	// True once this press has dropped a pick that was on the screen. The bar
	// below has to know it, or it keeps a frame with a band nothing holds.
	pressCleared := false
	switch {
	case msg.Action == tea.MouseActionMotion && m.drag.held:
		next := m.drag.to(msg.X, msg.Y)
		m.same = next == m.drag
		m.drag = next
		return m, nil
	case msg.Action == tea.MouseActionRelease && m.drag.held:
		m.drag.held = false
		if !m.drag.shown() {
			m.drag = drag{}
			m.same = true
			return m, nil
		}
		return m, m.copyDrag()
	case msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft:
		// A new press ends the old pick, and starts a new one when it lands on words.
		pressCleared = m.drag.on
		m.drag = m.anchorAt(msg.X, msg.Y)
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
	// The tab box at the top is one more way to open a tab, the same as the
	// number keys.
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft &&
		msg.Y < barRows && m.width >= 2 {
		_, spans := m.barTabs(m.width - 2)
		for i, s := range spans {
			// The bar's left wall takes the first cell, so names start one in.
			if s.w > 0 && msg.X >= 1+s.x && msg.X < 1+s.x+s.w {
				m.openTab(i)
				return m, nil
			}
		}
		m.same = !pressCleared
		return m, nil
	}
	p, rowIdx, tabIdx := m.hit(msg.X, msg.Y)
	switch msg.Button {
	case tea.MouseButtonWheelUp, tea.MouseButtonWheelDown:
		// The wheel moves the words, so the picked cells stop meaning anything.
		cleared := m.drag.on
		m.drag = drag{}
		// A notch belongs to the box under the pointer, so it takes the focus
		// the way a click does and scrolls the pane it landed on. A pane the
		// tab does not have refuses the focus, so a notch over one is ignored
		// as before.
		if p != m.focus {
			m.focusPane(p)
			if m.focus != p {
				m.same = !cleared
				return m, nil
			}
		}
		// hit answers the focused pane for a cell that lands on no pane at
		// all, like the tab bar or the status line, so the notch asks the
		// geometry before it moves a word.
		if !m.onPane(p, msg.X, msg.Y) {
			m.same = !cleared
			return m, nil
		}
		// The notch got past the pane the reader is looking at, so it is part
		// of the scroll. It is written down here, before any of the branches
		// below change what is armed, so first says what really happened.
		m.trace.Notch(!m.wheelArmed)
		step := wheelStep
		if msg.Button == tea.MouseButtonWheelUp {
			step = -wheelStep
		}
		// The notches belong to the screen they were gathered on. A screen
		// that is gone drops the delta it was given, so the next notch starts
		// a new frame for the screen the reader is on now.
		if m.wheelDelta != 0 && m.wheelMark != m.mark() {
			m.wheelDelta = 0
		}
		// Notches for another pane cannot share one delta. Scroll the old
		// pane now so its notches are not lost.
		flushed := false
		if m.wheelDelta != 0 && m.wheelPane != p {
			m.scrollPane(m.wheelPane, m.wheelDelta)
			m.wheelDelta = 0
			flushed = true
		}
		// With no tick waiting, the wheel had stopped. Scroll now, so the
		// reader sees the pane move on the notch itself, and start a frame
		// for the notches that follow.
		if !m.wheelArmed {
			m.scrollPane(p, step)
			m.wheelPane = p
			m.wheelMoved = true
			m.wheelArmed = true
			return m, wheelTickCmd()
		}
		// The first gathered notch of a frame writes the screen down. The
		// notches after it share that screen and must not write over it.
		if m.wheelDelta == 0 {
			m.wheelMark = m.mark()
		}
		m.wheelPane = p
		m.wheelDelta += step
		// A flush to another pane moved that pane, so only a plain gather
		// keeps the old frame.
		m.same = !flushed && !cleared
		return m, nil
	}
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		m.same = true
		return m, nil
	}
	m.focusPane(p)
	switch {
	case tabIdx >= 0:
		m.clickTab(p, tabIdx)
	case rowIdx >= 0:
		// A click on the space below the last row only takes the focus: it
		// holds no item, so the selection stays where it was.
		rows, _, _ := m.slotOf(p)
		if rowIdx >= len(rows) {
			return m, nil
		}
		m.moveTo(rowIdx)
	}
	return m, nil
}

// wheelTickCmd asks for the end of the frame.
func wheelTickCmd() tea.Cmd {
	return tea.Tick(wheelFrame, func(time.Time) tea.Msg { return wheelTickMsg{} })
}

// clickTab switches the Done box to the sub-tab the click landed on. The
// other boxes hold one name, so a click there only takes the focus.
func (m *Model) clickTab(p pane, idx int) {
	if p == paneDone {
		m.done = clamp(idx, 0, max(0, len(m.doneTabNames())-1))
	}
	m.keepVisible(p)
}

func (m *Model) openPopup(key string) {
	it := m.Selected()
	switch {
	case it == nil:
		m.status = "nothing selected"
		return
	case key == "p" && it.Kind != board.KindBug && it.Kind != board.KindDebtItem:
		m.status = "p sets the priority of a bug or a debt line"
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
	if key == "p" {
		p = popup{field: "priority", options: append(append([]string(nil), board.Priorities...), "none")}
		current = it.Priority
		if current == "" {
			current = "none"
		}
	}
	for i, o := range p.options {
		if o == current {
			p.idx = i
		}
	}
	m.popup = &p
}

// markRow marks the task or debt line under the cursor done or back to open.
// The same rows the status popup refuses are refused here, so a key press
// never writes a file the popup would not.
func (m Model) markRow(done bool) (tea.Model, tea.Cmd) {
	it := m.Selected()
	switch {
	case it == nil:
		m.status = "nothing selected"
		return m, nil
	case it.Kind != board.KindTask && it.Kind != board.KindDebtItem:
		m.status = "+ and - work on a task or a debt line; use s for the status"
		return m, nil
	case it.Worktree != "":
		m.status = "shown from worktree " + it.Worktree + "; edit it there"
		return m, nil
	case it.Legacy:
		m.status = "legacy file, move it into the root folder first"
		return m, nil
	}
	word := "open"
	if done {
		word = "done"
	}
	o, err := m.markItem(it.ID, done)
	m.status = outcomeText(fmt.Sprintf("acta: %s %s", it.ID, word), o, err)
	return m, m.reloadCmd()
}

// copyID puts the id of the row under the cursor on the clipboard. The status
// line says the copy landed, or why it could not, but never the id itself:
// the id is already on screen as row text. Update starts the timer that hides
// the message.
func (m *Model) copyID() tea.Cmd {
	it := m.Selected()
	if it == nil {
		m.status = "nothing selected"
		return nil
	}
	id := shortRef(it)
	if err := m.clip(id); err != nil {
		m.status = "copy failed: " + err.Error()
		return nil
	}
	m.status = "copied to clipboard"
	return nil
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
		m.status = outcomeText(fmt.Sprintf("acta: %s %s %s", it.ID, p.field, value), o, err)
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
		path, tmpl, err := write.StartBug(m.cfg, s, "", "")
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

// toggleRow opens or shuts the head under the cursor of a tree list. It says
// false on a task row, so enter goes on to open the detail of the task.
func (m *Model) toggleRow() bool {
	if m.focus == paneDetail {
		return false
	}
	rows := m.listOf()
	i := m.cursor()
	if i < 0 || !rows[i].tree {
		return false
	}
	if rows[i].depth > 0 {
		return false
	}
	m.setOpen(rows[i].id, !m.isOpen(rows[i].id))
	return true
}

// foldRow is h and l on a tree list. h on a task row shuts its head and puts
// the cursor on the head, h on a head shuts it, and l on a head opens it. l
// on a task row, and any row that is not a tree row, stay as they are.
func (m *Model) foldRow(open bool) {
	if m.focus == paneDetail {
		// The detail box has no tree of its own, so the list behind it must
		// not move while the reader reads there.
		return
	}
	rows := m.listOf()
	i := m.cursor()
	if i < 0 || i >= len(rows) || !rows[i].tree {
		return
	}
	if rows[i].depth > 0 {
		if open {
			return
		}
		// The head sits above its tasks, so walk up to it.
		for i > 0 && rows[i].depth > 0 {
			i--
		}
	}
	m.setOpen(rows[i].id, open)
	m.moveTo(i)
}

// enter moves the focus to the detail pane with the row's item, so it can
// be read and scrolled there. The legacy group row still folds open and shut,
// and an item missing from disk still warns how to check its branch out.
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
	if it == nil {
		return m, nil
	}
	if !it.OnDisk {
		m.status = "branch " + it.Worktree + " is not checked out; open it with: git worktree add ../<repo>-" + it.Worktree + " " + it.Worktree
		m.focusPane(paneDetail)
		return m, nil
	}
	m.focusPane(paneDetail)
	return m, nil
}

// edit opens the selected item in the editor from any pane, the way enter did
// before. Items missing from disk still warn how to check their branch out.
func (m Model) edit() (tea.Model, tea.Cmd) {
	it := m.Selected()
	if it == nil {
		return m, nil
	}
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
			m.status = outcomeText("acta: new bug", o, err)
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
