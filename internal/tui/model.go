// Package tui is the lazygit-style screen over a Board.
package tui

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"pm-board/internal/board"
	"pm-board/internal/config"
	"pm-board/internal/editor"
	"pm-board/internal/write"
)

type tab int

const (
	tabStories tab = iota
	tabTasks
	tabBugs
)

var (
	tabKinds = [3]board.Kind{board.KindStory, board.KindTask, board.KindBug}
	tabNames = [3]string{"Stories", "Tasks", "Bugs"}
)

// groupRowID marks the folded row that holds the legacy items.
const groupRowID = "\x00untyped"

type row struct {
	id    string
	group bool
	depth int
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

type editorDoneMsg struct {
	newBug string // path of a new bug file; "" after a plain edit
	tmpl   []byte
	err    error
}

// WatchFailed tells the model live reload could not start.
func WatchFailed(err error) tea.Msg { return watchFailedMsg{err: err} }

// Model is the whole screen state. Bubble Tea copies it on every update.
type Model struct {
	cfg       config.Config
	board     *board.Board
	tab       tab
	sel       [3]string // selected id per tab
	idx       [3]int    // selected row number per tab, used when the id vanishes
	showAll   bool
	query     string
	searching bool
	groupOpen bool
	popup     *popup
	slug      *string // non-nil while typing the slug of a new bug
	help      bool
	status    string
	manual    bool
	width     int
	height    int
	scroll    int

	load     func() (*board.Board, error)
	setValue func(id, field, value string) (write.Outcome, error)
	render   func(md string, width int) string
}

// New builds a model over b. dark picks the markdown style; ask the terminal
// before the program starts, because asking later fights Bubble Tea for stdin.
func New(cfg config.Config, b *board.Board, dark bool) Model {
	return Model{
		cfg: cfg, board: b, width: 120, height: 40,
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

func (m Model) Init() tea.Cmd { return nil }

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
	case editorDoneMsg:
		return m.afterEditor(msg)
	case tea.KeyMsg:
		return m.key(msg)
	}
	return m, nil
}

// Selected is the item under the cursor, or nil on the group row or an empty list.
func (m Model) Selected() *board.Item {
	rows := m.rows()
	i := m.cursor()
	if i < 0 || rows[i].group {
		return nil
	}
	return m.board.Get(rows[i].id)
}

func (m Model) rows() []row {
	if m.query != "" {
		return toRows(m.board.Search(m.query), 0)
	}
	rows := toRows(m.board.List(tabKinds[m.tab], m.showAll), 0)
	if m.tab == tabStories {
		rows = toRows(m.storyItems(), 0)
		if legacy := m.board.Untyped(m.showAll); len(legacy) > 0 {
			rows = append(rows, row{id: groupRowID, group: true})
			if m.groupOpen {
				rows = append(rows, toRows(legacy, 1)...)
			}
		}
	}
	return rows
}

func toRows(items []*board.Item, depth int) []row {
	rows := make([]row, 0, len(items))
	for _, it := range items {
		rows = append(rows, row{id: it.ID, depth: depth})
	}
	return rows
}

// storyItems gives the Stories tab its rows. A plan that has a spec shows on
// that spec's row, but a plan with none has no other home, so it lists here
// too until plans get a tab of their own.
func (m Model) storyItems() []*board.Item {
	stories := m.board.List(board.KindStory, m.showAll)
	plans := m.board.List(board.KindPlan, m.showAll)
	out := make([]*board.Item, 0, len(stories)+len(plans))
	for _, p := range plans {
		if p.SpecID == "" {
			out = append(out, p)
		}
	}
	// Both lists run newest first, so a merge keeps that order.
	i := 0
	for _, s := range stories {
		for i < len(out) && out[i].Date > s.Date {
			i++
		}
		out = append(out, nil)
		copy(out[i+1:], out[i:])
		out[i] = s
	}
	return out
}

// cursor finds the selected id; when it is gone it falls back to the old row number.
func (m Model) cursor() int {
	rows := m.rows()
	if len(rows) == 0 {
		return -1
	}
	for i, r := range rows {
		if r.id == m.sel[m.tab] {
			return i
		}
	}
	return clamp(m.idx[m.tab], 0, len(rows)-1)
}

func (m *Model) moveTo(i int) {
	rows := m.rows()
	if len(rows) == 0 {
		m.sel[m.tab], m.idx[m.tab] = "", 0
		return
	}
	i = clamp(i, 0, len(rows)-1)
	if rows[i].id != m.sel[m.tab] {
		m.scroll = 0
	}
	m.sel[m.tab], m.idx[m.tab] = rows[i].id, i
}

func (m Model) key(k tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		m.tab = tab(k.String()[0] - '1')
		m.scroll = 0
	case "tab":
		m.tab = (m.tab + 1) % 3
		m.scroll = 0
	case "j", "down":
		m.moveTo(m.cursor() + 1)
	case "k", "up":
		m.moveTo(m.cursor() - 1)
	case "g":
		m.moveTo(0)
	case "G":
		m.moveTo(len(m.rows()) - 1)
	case "ctrl+d":
		m.scroll += 10
	case "ctrl+u":
		m.scroll = max(0, m.scroll-10)
	case "/":
		m.searching = true
	case "esc":
		m.query = ""
		m.moveTo(0)
	case "a":
		m.showAll = !m.showAll
		m.moveTo(m.cursor())
	case "?":
		m.help = !m.help
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
	rows := m.rows()
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
