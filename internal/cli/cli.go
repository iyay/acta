// Package cli runs the acta command line: it shows and changes the planning
// files of the repo it runs in. Both the acta binary and the pmb alias
// call Run, so the two stay one implementation.
package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/editor"
	"github.com/iyay/acta/internal/trees"
	"github.com/iyay/acta/internal/tui"
	"github.com/iyay/acta/internal/write"
)

const (
	exitOK       = 0
	exitBadInput = 1
	exitSkipped  = 2
	exitOther    = 3
)

// runTUI opens the TUI with live reload. When the watcher cannot start, the
// TUI still opens in manual mode.
var runTUI = func(cfg config.Config, stderr io.Writer) int {
	b, err := trees.Load(cfg)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	// Ask the terminal for its background now; asking inside the program
	// fights Bubble Tea for stdin.
	// The version comes from the build info: a tag for go install, dev for a
	// local build with no version stamped in.
	version := "dev"
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
	dark := lipgloss.HasDarkBackground()
	m := tui.New(cfg, b, dark).WithTheme(voiceTheme(), dark).WithVersion(version).WithLoad(func() (*board.Board, error) { return trees.Load(cfg) })
	// 120fps halves how long a new frame waits to reach the screen.
	opts := []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseCellMotion(), tea.WithFPS(120)}
	tr, traceOpts, traceDone := tuiTrace(os.Getenv("ACTA_TUI_TRACE"), os.Stdout, stderr)
	defer traceDone()
	m = m.WithTrace(tr)
	p := tea.NewProgram(m, append(opts, traceOpts...)...)
	load := func() (*board.Board, error) { return trees.Load(cfg) }
	dirs := func() []string { return append(tui.WatchDirs(cfg), trees.WatchDirs(cfg, tui.WatchDirs)...) }
	if stop, err := tui.Watch(dirs, load, p.Send); err != nil {
		go p.Send(tui.WatchFailed(err))
	} else {
		defer stop()
	}
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	return exitOK
}

// voiceTheme is the theme the TUI paints with. A voice file that does not
// parse still gives the default voice, so the board opens either way and the
// name is simply empty, which the default theme covers.
func voiceTheme() string {
	v, _, _ := config.ResolveUser()
	return v.Theme
}

// Run handles one command line call and reports its exit code, so both
// binaries share it instead of copying the dispatch.
func Run(args []string, stdin io.Reader, stdinIsTTY bool, stdout, stderr io.Writer) int {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		fs, root := flags("acta", stderr)
		if err := fs.Parse(args); err != nil {
			return exitBadInput
		}
		cfg, code := loadConfig(*root, stderr)
		if code != exitOK {
			return code
		}
		return runTUI(cfg, stderr)
	}
	switch args[0] {
	case "list":
		return cmdList(args[1:], stdout, stderr)
	case "show":
		return cmdShow(args[1:], stdout, stderr)
	case "set":
		return cmdSet(args[1:], stdout, stderr)
	case "bug":
		if len(args) < 2 || args[1] != "new" {
			fmt.Fprintln(stderr, "usage: acta bug new <slug> [--ref X] [--title T] [--priority P] < body.md")
			return exitBadInput
		}
		return cmdBugNew(args[2:], stdin, stdinIsTTY, stdout, stderr)
	case "debt":
		if len(args) < 2 || args[1] != "new" {
			fmt.Fprintln(stderr, "usage: acta debt new <plan id> [--title T] < notes.md")
			return exitBadInput
		}
		return cmdDebtNew(args[2:], stdin, stdinIsTTY, stdout, stderr)
	case "scratch":
		if len(args) < 2 || (args[1] != "new" && args[1] != "add") {
			fmt.Fprintln(stderr, "usage: acta scratch new <slug> [--title T] < body.md | acta scratch add <SCRATCH-n> [--section words|context|log|questions] < text.md")
			return exitBadInput
		}
		if args[1] == "new" {
			return cmdScratchNew(args[2:], stdin, stdout, stderr)
		}
		return cmdScratchAdd(args[2:], stdin, stdout, stderr)
	case "config":
		return cmdConfig(args[1:], stdout, stderr)
	case "hook":
		return cmdHook(args[1:], stdin, stdout, stderr)
	case "tick":
		return cmdTick(args[1:], stdout, stderr)
	case "id":
		return cmdID(args[1:], stdout, stderr)
	case "migrate-root":
		return cmdMigrateRoot(args[1:], stdout, stderr)
	case "doctor":
		return cmdDoctor(args[1:], stdout, stderr)
	case "dispatch":
		if len(args) < 2 || (args[1] != "init" && args[1] != "send" && args[1] != "close") {
			fmt.Fprintln(stderr, "usage: acta dispatch init --pane <id> --plan <path> [--round <slug>] | send --plan <path> --rules <path> [--round <slug>] [--note-file <path>|-] | close")
			return exitBadInput
		}
		switch args[1] {
		case "send":
			return cmdDispatchSend(args[2:], stdin, stdout, stderr)
		case "close":
			return cmdDispatchClose(args[2:], stdout, stderr)
		}
		return cmdDispatchInit(args[2:], stdout, stderr)
	case "reply-back":
		return cmdReplyBack(args[1:], stdout, stderr)
	case "run-one":
		return cmdRunOne(args[1:], stdin, stdout, stderr)
	case "eval-omp":
		return cmdEvalOmp(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q; use doctor, id, list, show, set, tick, migrate-root, bug new, debt new, scratch new, scratch add, dispatch init, dispatch send, dispatch close, reply-back, run-one or eval-omp\n", args[0])
		return exitBadInput
	}
}

func flags(name string, stderr io.Writer) (*flag.FlagSet, *string) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", "", "planning root folder (default .acta; also ACTA_ROOT, PM_ROOT, or root in .acta.yaml or .pm.yaml)")
	return fs, root
}

// parseMixed lets flags come before or after the positional arguments.
func parseMixed(fs *flag.FlagSet, args []string) ([]string, error) {
	var pos []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		if fs.NArg() == 0 {
			return pos, nil
		}
		pos = append(pos, fs.Arg(0))
		args = fs.Args()[1:]
	}
}

func loadConfig(root string, stderr io.Writer) (config.Config, int) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return config.Config{}, exitOther
	}
	cfg, err := config.Load(cwd, root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return config.Config{}, exitBadInput
	}
	return cfg, exitOK
}

func loadBoard(root string, stderr io.Writer) (config.Config, *board.Board, int) {
	cfg, code := loadConfig(root, stderr)
	if code != exitOK {
		return cfg, nil, code
	}
	b, err := board.Load(cfg)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return cfg, nil, exitOther
	}
	return cfg, b, exitOK
}

// loadAllTrees reads the main tree plus every worktree and unmerged branch,
// so list and show report progress from anywhere in the repo.
func loadAllTrees(root string, stderr io.Writer) (config.Config, *board.Board, int) {
	cfg, code := loadConfig(root, stderr)
	if code != exitOK {
		return cfg, nil, code
	}
	b, err := trees.Load(cfg)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return cfg, nil, exitOther
	}
	return cfg, b, exitOK
}

type outItem struct {
	ID           string       `json:"id"`
	ShortID      string       `json:"short_id"`
	Hash         string       `json:"hash"`
	Type         string       `json:"type"`
	Title        string       `json:"title"`
	Status       string       `json:"status"`
	StatusSource string       `json:"status_source"`
	Ref          string       `json:"ref"`
	Parent       string       `json:"parent"`
	Closes       []string     `json:"closes"`
	ClosedBy     []string     `json:"closed_by"`
	Children     []string     `json:"children"`
	Progress     jsonProgress `json:"progress"`
	Path         string       `json:"path"`
	Legacy       bool         `json:"legacy"`
	Worktree     string       `json:"worktree"`
	OnDisk       bool         `json:"on_disk"`
	Problems     []string     `json:"problems"`
}

type jsonProgress struct {
	Done  int `json:"done"`
	Total int `json:"total"`
}

func toJSON(cfg config.Config, it *board.Item) outItem {
	rel := it.Path
	if it.OnDisk {
		if r, err := filepath.Rel(cfg.RepoRoot, it.Path); err == nil {
			rel = r
		}
	}
	j := outItem{ID: it.ID, ShortID: it.ShortID, Hash: it.Hash, Type: string(it.Kind), Title: it.Title, Status: it.Status,
		StatusSource: it.StatusSource, Ref: it.Ref, Parent: it.Parent,
		Children: append([]string{}, it.Children...), Progress: jsonProgress{it.Done, it.Total},
		Closes: append([]string{}, it.Closes...), ClosedBy: append([]string{}, it.ClosedBy...),
		Path: filepath.ToSlash(rel), Legacy: it.Legacy, Worktree: it.Worktree, OnDisk: it.OnDisk,
		Problems: append([]string{}, it.Problems...)}
	return j
}

func cmdList(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("list", stderr)
	kind := fs.String("type", "", "story, task or bug")
	all := fs.Bool("all", false, "include closed and legacy items")
	asJSON := fs.Bool("json", false, "print JSON")
	if pos, err := parseMixed(fs, args); err != nil || len(pos) > 0 {
		return exitBadInput
	}
	if *kind != "" && *kind != "story" && *kind != "task" && *kind != "bug" {
		fmt.Fprintf(stderr, "--type must be story, task or bug\n")
		return exitBadInput
	}
	cfg, b, code := loadAllTrees(*root, stderr)
	if code != exitOK {
		return code
	}
	out := []outItem{}
	for _, it := range b.Items {
		if *kind != "" && string(it.Kind) != *kind {
			continue
		}
		if !*all && (it.Legacy || board.Closed(it.Status)) {
			continue
		}
		out = append(out, toJSON(cfg, it))
	}
	if *asJSON {
		return printJSON(stdout, stderr, out)
	}
	for _, it := range out {
		short := it.ShortID
		if short == "" {
			short = "-"
		}
		fmt.Fprintf(stdout, "%-11s %-6s %-10s %s  %s\n", it.Status, it.Type, short, it.ID, it.Title)
	}
	return exitOK
}

func cmdShow(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("show", stderr)
	asJSON := fs.Bool("json", false, "print JSON")
	onlyPath := fs.Bool("path", false, "print the file path only")
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) != 1 {
		fmt.Fprintln(stderr, "usage: acta show <id> [--json] [--path]")
		return exitBadInput
	}
	if *onlyPath {
		return showPath(*root, pos[0], stdout, stderr)
	}
	cfg, b, code := loadAllTrees(*root, stderr)
	if code != exitOK {
		return code
	}
	it := b.Get(pos[0])
	if it == nil {
		fmt.Fprintf(stderr, "unknown id %s\n", pos[0])
		return exitBadInput
	}
	j := toJSON(cfg, it)
	if *asJSON {
		return printJSON(stdout, stderr, j)
	}
	head := j.ID
	if j.ShortID != "" {
		head = j.ShortID + " " + j.Hash + " " + j.ID
	}
	fmt.Fprintf(stdout, "%s  %s\ntype: %s\nstatus: %s (%s)\nprogress: %d/%d\npath: %s\n",
		head, j.Title, j.Type, j.Status, j.StatusSource, j.Progress.Done, j.Progress.Total, j.Path)
	if j.Ref != "" {
		fmt.Fprintf(stdout, "ref: %s\n", j.Ref)
	}
	if j.Parent != "" {
		fmt.Fprintf(stdout, "parent: %s\n", j.Parent)
	}
	if len(j.Closes) > 0 {
		fmt.Fprintf(stdout, "closes: %s\n", strings.Join(shortRefs(b, j.Closes), ", "))
	}
	if len(j.ClosedBy) > 0 {
		fmt.Fprintf(stdout, "closed by: %s\n", strings.Join(shortRefs(b, j.ClosedBy), ", "))
	}
	for _, c := range j.Children {
		ch := b.Get(c)
		fmt.Fprintf(stdout, "  %-6s %s  %s\n", ch.Status, c, ch.Title)
	}
	for _, p := range j.Problems {
		fmt.Fprintf(stdout, "! %s\n", p)
	}
	return exitOK
}

// showPath prints only the file an item lives in. It reads the frontmatter of
// the planning files first, so a plain id or a full hash costs no board load.
// A miss goes the long way round through the board, which knows old ids,
// stems, and items that live only in another worktree.
func showPath(root, id string, stdout, stderr io.Writer) int {
	cfg, code := loadConfig(root, stderr)
	if code != exitOK {
		return code
	}
	if p := scanPath(cfg, id); p != "" {
		fmt.Fprintln(stdout, p)
		return exitOK
	}
	_, b, code := loadAllTrees(root, stderr)
	if code != exitOK {
		return code
	}
	it := b.Get(id)
	if it == nil {
		fmt.Fprintf(stderr, "unknown id %s\n", id)
		return exitBadInput
	}
	fmt.Fprintln(stdout, toJSON(cfg, it).Path)
	return exitOK
}

// scanPath gives the path of the file whose frontmatter names id or hash as
// the whole value, in the same form show prints. A sub-item drops its number,
// so PLN-0040.01 finds the plan file. Nothing means the board has to answer.
func scanPath(cfg config.Config, id string) string {
	want := id
	if i := strings.IndexByte(want, '.'); i >= 0 {
		want = want[:i]
	}
	for _, d := range []string{cfg.Dirs.Specs, cfg.Dirs.Plans, cfg.Dirs.Bugs, cfg.Dirs.Debt, cfg.Dirs.Scratch} {
		if d == "" {
			continue
		}
		files, err := filepath.Glob(filepath.Join(cfg.Root, d, "*.md"))
		if err != nil {
			continue
		}
		for _, f := range files {
			if !namesID(f, want) {
				continue
			}
			rel, err := filepath.Rel(cfg.RepoRoot, f)
			if err != nil {
				rel = f
			}
			return filepath.ToSlash(rel)
		}
	}
	return ""
}

// namesID says if the file's frontmatter carries want as its whole id or hash
// value. It reads only the block between the first two --- lines, so a long
// body costs nothing.
func namesID(path, want string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	if !sc.Scan() || strings.TrimSpace(sc.Text()) != "---" {
		return false
	}
	for sc.Scan() {
		ln := strings.TrimSpace(sc.Text())
		if ln == "---" {
			return false
		}
		k, v, ok := strings.Cut(ln, ":")
		if !ok || (k != "id" && k != "hash") {
			continue
		}
		if strings.Trim(strings.TrimSpace(v), "\"'") == want {
			return true
		}
	}
	return false
}

// shortRefs names linked items the way a reader says them out: the short id
// when the file carries one, else the path id it goes by.
func shortRefs(b *board.Board, ids []string) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if it := b.Get(id); it != nil && it.ShortID != "" {
			id = it.ShortID
		}
		out = append(out, id)
	}
	return out
}

func cmdSet(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("set", stderr)
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) != 3 {
		fmt.Fprintln(stderr, "usage: acta set <id> status|type|title|fixed_in|ref|priority <value>")
		return exitBadInput
	}
	cfg, b, code := loadBoard(*root, stderr)
	if code != exitOK {
		return code
	}
	o, err := write.SetValue(cfg, b, pos[0], pos[1], pos[2])
	return report(o, err, stdout, stderr)
}

func cmdBugNew(args []string, stdin io.Reader, stdinIsTTY bool, stdout, stderr io.Writer) int {
	fs, root := flags("bug new", stderr)
	ref := fs.String("ref", "", "outside ticket code")
	title := fs.String("title", "", "bug title (default: the body's # line, or the slug)")
	priority := fs.String("priority", "", "high, medium or low")
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) != 1 {
		fmt.Fprintln(stderr, "usage: acta bug new <slug> [--ref X] [--title T] [--priority P] < body.md")
		return exitBadInput
	}
	cfg, code := loadConfig(*root, stderr)
	if code != exitOK {
		return code
	}
	if !stdinIsTTY {
		body, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitOther
		}
		o, err := write.NewBug(cfg, pos[0], *title, *ref, *priority, body)
		return report(o, err, stdout, stderr)
	}
	path, tmpl, err := write.StartBug(cfg, pos[0], *ref, *priority)
	if err != nil {
		return report(write.Outcome{}, err, stdout, stderr)
	}
	cmd := editor.Cmd(path, 1)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(stderr, "editor:", err)
	}
	o, err := write.FinishBug(cfg, path, tmpl)
	if errors.Is(err, write.ErrUnchanged) {
		fmt.Fprintln(stderr, "template left unchanged; nothing saved")
		return exitBadInput
	}
	return report(o, err, stdout, stderr)
}

func cmdDebtNew(args []string, stdin io.Reader, stdinIsTTY bool, stdout, stderr io.Writer) int {
	fs, root := flags("debt new", stderr)
	title := fs.String("title", "", "debt heading (default: Review NOTEs: <plan title>)")
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) != 1 {
		fmt.Fprintln(stderr, "usage: acta debt new <plan id> [--title T] < notes.md")
		return exitBadInput
	}
	if stdinIsTTY {
		fmt.Fprintln(stderr, "pipe NOTEs on stdin")
		return exitBadInput
	}
	cfg, b, code := loadBoard(*root, stderr)
	if code != exitOK {
		return code
	}
	body, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	o, err := write.NewDebt(cfg, b, pos[0], *title, body)
	return report(o, err, stdout, stderr)
}

func cmdScratchNew(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs, root := flags("scratch new", stderr)
	title := fs.String("title", "", "scratch title (default: the slug)")
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) != 1 {
		fmt.Fprintln(stderr, "usage: acta scratch new <slug> [--title T] < body.md")
		return exitBadInput
	}
	cfg, code := loadConfig(*root, stderr)
	if code != exitOK {
		return code
	}
	body, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	o, err := write.NewScratch(cfg, pos[0], *title, body)
	return report(o, err, stdout, stderr)
}

func cmdScratchAdd(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs, root := flags("scratch add", stderr)
	section := fs.String("section", "", "words, context, log or questions (default: words)")
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) != 1 {
		fmt.Fprintln(stderr, "usage: acta scratch add <SCRATCH-n> [--section words|context|log|questions] < text.md")
		return exitBadInput
	}
	cfg, b, code := loadBoard(*root, stderr)
	if code != exitOK {
		return code
	}
	text, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	o, err := write.AppendScratch(cfg, b, pos[0], *section, text)
	return report(o, err, stdout, stderr)
}

// report prints the written path and maps the outcome to an exit code.
func report(o write.Outcome, err error, stdout, stderr io.Writer) int {
	switch {
	case errors.Is(err, write.ErrBadInput):
		fmt.Fprintln(stderr, err)
		return exitBadInput
	case err != nil:
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	if cwd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(cwd, o.Path); err == nil {
			o.Path = rel
		}
	}
	// A scratch write names the new number ID first, so the next command can
	// use it as it stands.
	line := filepath.ToSlash(o.Path)
	if o.ShortID != "" {
		line = o.ShortID + "  " + line
	}
	fmt.Fprintln(stdout, line)
	if o.Skipped {
		fmt.Fprintln(stderr, "written, not committed:", o.Reason)
		return exitSkipped
	}
	return exitOK
}

func printJSON(stdout, stderr io.Writer, v any) int {
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	return exitOK
}
