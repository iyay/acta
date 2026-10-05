package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/hook"
	"github.com/iyay/acta/internal/write"
)

const (
	// exitDelivery means herdr failed after everything was checked.
	exitDelivery = 2
	// exitDrift means the recipient's todo list misses some task ids. Exit 1
	// already means bad input in this CLI, so drift gets its own number.
	exitDrift = 4
)

// herdrNamePattern is what herdr accepts as an agent name.
var herdrNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

// dispatchSlug is the agent name for a branch: the round name cut to the 32
// characters herdr allows. It says no when the result is not a valid name.
func dispatchSlug(branch string) (string, bool) {
	s, ok := roundFromBranch(branch)
	if !ok {
		return "", false
	}
	if len(s) > 32 {
		s = strings.TrimRight(s[:32], "-")
	}
	return s, herdrNamePattern.MatchString(s)
}

// readNote returns the note text: a file, or stdin for "-". No flag, no note.
func readNote(path string, stdin io.Reader) (string, error) {
	switch path {
	case "":
		return "", nil
	case "-":
		if stdin == nil {
			return "", errors.New("--note-file - needs text on stdin")
		}
		raw, err := io.ReadAll(stdin)
		return string(raw), err
	}
	raw, err := os.ReadFile(path)
	return string(raw), err
}

// cmdDispatchSend sends one round to the recipient. Every check, the record
// and the brief come first; herdr is not touched until they all pass.
func cmdDispatchSend(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs, root := flags("dispatch send", stderr)
	plan := fs.String("plan", "", "plan path relative to the repo root")
	round := fs.String("round", "", `empty = first dispatch, "polish", or the name of a fix round`)
	noteFile := fs.String("note-file", "", "file with the note for the recipient, or - for stdin")
	rules := fs.String("rules", "", "absolute path of references/house-rules.md")
	if err := fs.Parse(args); err != nil {
		return exitBadInput
	}
	if os.Getenv("HERDR_ENV") != "1" {
		fmt.Fprintln(stderr, "dispatch send needs HERDR_ENV=1, so run it inside herdr")
		return exitBadInput
	}
	orchestrator := os.Getenv("HERDR_PANE_ID")
	if !panePattern.MatchString(orchestrator) {
		fmt.Fprintf(stderr, "HERDR_PANE_ID %q is not a pane id, so the recipient cannot answer\n", orchestrator)
		return exitBadInput
	}
	if *plan == "" {
		fmt.Fprintln(stderr, "--plan is required")
		return exitBadInput
	}
	if *rules == "" {
		fmt.Fprintln(stderr, "--rules is required: the absolute path of house-rules.md")
		return exitBadInput
	}
	cfg, b, code := loadBoard(*root, stderr)
	if code != exitOK {
		return code
	}
	if !cfg.IsGit {
		fmt.Fprintln(stderr, "dispatch send needs a git repo, so it can read the head commit")
		return exitBadInput
	}
	note, err := readNote(*noteFile, stdin)
	if err != nil {
		fmt.Fprintln(stderr, "cannot read the note:", err)
		return exitBadInput
	}
	base, err := gitIn(cfg.RepoRoot, "rev-parse", "HEAD")
	if err != nil {
		fmt.Fprintln(stderr, "cannot read the head commit:", err)
		return exitOther
	}
	branch, err := gitIn(cfg.RepoRoot, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		fmt.Fprintln(stderr, "cannot read the branch name:", err)
		return exitOther
	}
	common, err := gitIn(cfg.RepoRoot, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		fmt.Fprintln(stderr, "cannot find the main checkout:", err)
		return exitOther
	}
	// The repo records no parent branch, so the parent is the branch the
	// main checkout is on: that is where land merges the worktree.
	mainCheckout := filepath.Dir(common)
	parent, err := gitIn(mainCheckout, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		fmt.Fprintln(stderr, "cannot read the branch of the main checkout:", err)
		return exitOther
	}
	if parent == "HEAD" {
		fmt.Fprintln(stderr, "the main checkout is on a detached head, so there is no parent branch to merge into")
		return exitOther
	}
	slug, ok := dispatchSlug(branch)
	if !ok {
		fmt.Fprintf(stderr, "cannot make an agent slug from branch %q: it must be a-z, 0-9, - or _, start with a letter, 32 characters at most\n", branch)
		return exitBadInput
	}
	roundName := *round
	if roundName == "" {
		roundName, _ = roundFromBranch(branch)
	}
	rec := dispatchRecord{Pane: orchestrator, Base: base, Plan: *plan, Round: roundName}
	_, err = checkRecord(cfg, b, rec)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	src, err := os.ReadFile(filepath.Join(cfg.RepoRoot, rec.Plan))
	if err != nil {
		fmt.Fprintln(stderr, "cannot read the plan:", err)
		return exitBadInput
	}
	// The brief is built before the record is written, so a plan the brief
	// refuses leaves no record behind.
	text, err := buildBrief(rec.Plan, src, briefInput{
		Round: *round, Note: note, Rules: *rules, Worktree: cfg.RepoRoot, Branch: branch,
		Parent: parent, Base: base,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	tasks, err := briefTasks(src, *round)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	var ids []string
	for _, t := range tasks {
		ids = append(ids, t.Num)
	}

	if _, code := writeRecord(cfg, rec, stderr); code != exitOK {
		return code
	}
	briefPath := filepath.Join(cfg.RepoRoot, ".claude", "dispatch", slug+"-brief.md")
	if err := os.MkdirAll(filepath.Dir(briefPath), 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	if err := os.WriteFile(briefPath, []byte(text), 0o644); err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}

	title := strings.Join(strings.Fields(board.Parse(src).Title), " ")
	if title == "" {
		title = "plan"
	}
	goal := "/goal " + title + " ultrathink orchestrate. Read the hand-off at " + briefPath +
		" first and follow every line. REPLY-BACK: after the last commit the build skill runs `acta reply-back`."
	pane, err := findOrMakeTab(slug, cfg.RepoRoot)
	if err == nil {
		err = deliverGoal(slug, goal)
	}
	if err != nil {
		fmt.Fprintln(stderr, "delivery failed:", err)
		return exitDelivery
	}
	// The goal is out, so later ticks in this worktree name the recipient
	// even when the harness sets no agent. A write error only gets a word:
	// the send itself already went through.
	_ = hook.EnsureGitignore(cfg.Root, ".agents.json")
	if err := write.SetDefaultAgent(cfg.Root, "omp"); err != nil {
		fmt.Fprintln(stderr, "cannot record the default agent:", err)
	}
	verdict, missing, paneText := checkpoint(slug, ids)
	line := verdict
	if verdict == checkpointDrift {
		line = "drift: missing " + strings.Join(missing, ", ")
	}
	fmt.Fprintf(stdout, "slug: %s\npane: %s\nbase: %s\nbrief: %s\ncheckpoint: %s\nwatcher: herdr agent wait %s --until idle --until done\n",
		slug, pane, base, briefPath, line, slug)
	if verdict == checkpointDrift {
		fmt.Fprintln(stdout, paneText)
		return exitDrift
	}
	return exitOK
}

// cmdDispatchClose closes the recipient's pane at land time.
func cmdDispatchClose(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("dispatch close", stderr)
	if err := fs.Parse(args); err != nil || fs.NArg() > 0 {
		return exitBadInput
	}
	cfg, code := loadConfig(*root, stderr)
	if code != exitOK {
		return code
	}
	if !cfg.IsGit {
		fmt.Fprintln(stderr, "dispatch close needs a git repo, so it can read the branch name")
		return exitBadInput
	}
	branch, err := gitIn(cfg.RepoRoot, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		fmt.Fprintln(stderr, "cannot read the branch name:", err)
		return exitOther
	}
	slug, ok := dispatchSlug(branch)
	if !ok {
		fmt.Fprintf(stderr, "cannot make an agent slug from branch %q\n", branch)
		return exitBadInput
	}
	out, err := herdr("agent", "get", slug)
	if err != nil {
		if isNotFound(err, out) {
			fmt.Fprintf(stderr, "no agent named %s, nothing to close\n", slug)
			return exitBadInput
		}
		fmt.Fprintln(stderr, err)
		return exitDelivery
	}
	a, err := parseAgent(out)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitDelivery
	}
	if _, err := herdr("pane", "close", a.Pane); err != nil {
		fmt.Fprintln(stderr, err)
		return exitDelivery
	}
	fmt.Fprintf(stdout, "closed: %s (pane %s)\n", slug, a.Pane)
	return exitOK
}
