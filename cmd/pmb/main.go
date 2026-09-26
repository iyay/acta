// Command pmb shows and changes the planning files of the repo it runs in.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"pm-board/internal/board"
	"pm-board/internal/config"
	"pm-board/internal/editor"
	"pm-board/internal/write"
)

const (
	exitOK       = 0
	exitBadInput = 1
	exitSkipped  = 2
	exitOther    = 3
)

// runTUI opens the TUI. Task 10 replaces this stub.
var runTUI = func(cfg config.Config, stderr io.Writer) int {
	fmt.Fprintln(stderr, "TUI not built yet")
	return exitOther
}

func main() {
	st, _ := os.Stdin.Stat()
	tty := st != nil && st.Mode()&os.ModeCharDevice != 0
	os.Exit(run(os.Args[1:], os.Stdin, tty, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdinIsTTY bool, stdout, stderr io.Writer) int {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		fs, root := flags("pmb", stderr)
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
			fmt.Fprintln(stderr, "usage: pmb bug new <slug> [--ref X] [--title T] < body.md")
			return exitBadInput
		}
		return cmdBugNew(args[2:], stdin, stdinIsTTY, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q; use list, show, set or bug new\n", args[0])
		return exitBadInput
	}
}

func flags(name string, stderr io.Writer) (*flag.FlagSet, *string) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", "", "planning root folder (default .pm, or PM_ROOT, or root in .pm.yaml)")
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

type outItem struct {
	ID           string       `json:"id"`
	Type         string       `json:"type"`
	Title        string       `json:"title"`
	Status       string       `json:"status"`
	StatusSource string       `json:"status_source"`
	Ref          string       `json:"ref"`
	Parent       string       `json:"parent"`
	Children     []string     `json:"children"`
	Progress     jsonProgress `json:"progress"`
	Path         string       `json:"path"`
	Legacy       bool         `json:"legacy"`
	Problems     []string     `json:"problems"`
}

type jsonProgress struct {
	Done  int `json:"done"`
	Total int `json:"total"`
}

func toJSON(cfg config.Config, it *board.Item) outItem {
	rel, err := filepath.Rel(cfg.RepoRoot, it.Path)
	if err != nil {
		rel = it.Path
	}
	j := outItem{ID: it.ID, Type: string(it.Kind), Title: it.Title, Status: it.Status,
		StatusSource: it.StatusSource, Ref: it.Ref, Parent: it.Parent,
		Children: append([]string{}, it.Children...), Progress: jsonProgress{it.Done, it.Total},
		Path: filepath.ToSlash(rel), Legacy: it.Legacy, Problems: append([]string{}, it.Problems...)}
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
	cfg, b, code := loadBoard(*root, stderr)
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
		fmt.Fprintf(stdout, "%-11s %-6s %s  %s\n", it.Status, it.Type, it.ID, it.Title)
	}
	return exitOK
}

func cmdShow(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("show", stderr)
	asJSON := fs.Bool("json", false, "print JSON")
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) != 1 {
		fmt.Fprintln(stderr, "usage: pmb show <id> [--json]")
		return exitBadInput
	}
	cfg, b, code := loadBoard(*root, stderr)
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
	fmt.Fprintf(stdout, "%s  %s\ntype: %s\nstatus: %s (%s)\nprogress: %d/%d\npath: %s\n",
		j.ID, j.Title, j.Type, j.Status, j.StatusSource, j.Progress.Done, j.Progress.Total, j.Path)
	if j.Ref != "" {
		fmt.Fprintf(stdout, "ref: %s\n", j.Ref)
	}
	if j.Parent != "" {
		fmt.Fprintf(stdout, "parent: %s\n", j.Parent)
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

func cmdSet(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("set", stderr)
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) != 3 {
		fmt.Fprintln(stderr, "usage: pmb set <id> status|type <value>")
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
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) != 1 {
		fmt.Fprintln(stderr, "usage: pmb bug new <slug> [--ref X] [--title T] < body.md")
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
		o, err := write.NewBug(cfg, pos[0], *title, *ref, body)
		return report(o, err, stdout, stderr)
	}
	path, tmpl, err := write.StartBug(cfg, pos[0], *ref)
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
	fmt.Fprintln(stdout, filepath.ToSlash(o.Path))
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
