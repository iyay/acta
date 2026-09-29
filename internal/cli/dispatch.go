package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/hook"
)

// dispatchRecord is the note the orchestrator leaves in the worktree so the
// recipient can answer without the user repeating anything.
type dispatchRecord struct {
	Pane  string `json:"pane"`
	Base  string `json:"base"`
	Plan  string `json:"plan"`
	Round string `json:"round"`
}

func recordPath(cfg config.Config) string {
	return filepath.Join(cfg.Root, ".dispatch.json")
}

var (
	panePattern  = regexp.MustCompile(`^[A-Za-z0-9]+:[A-Za-z0-9]+$`)
	shaPattern   = regexp.MustCompile(`^[0-9a-f]{40}$`)
	roundPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)
)

var notSlug = regexp.MustCompile(`[^a-z0-9]+`)

// roundFromBranch turns a branch name into a round, so the default works for
// names like feat/x that the round check would refuse as they are.
func roundFromBranch(branch string) (string, bool) {
	s := strings.Trim(notSlug.ReplaceAllString(strings.ToLower(branch), "-"), "-")
	if len(s) > 64 {
		s = strings.TrimRight(s[:64], "-")
	}
	return s, roundPattern.MatchString(s)
}

// checkRecord treats the record as untrusted, because anyone can edit the
// file. Every value is checked before it is used for anything.
func checkRecord(cfg config.Config, b *board.Board, r dispatchRecord) (*board.Item, error) {
	if !panePattern.MatchString(r.Pane) {
		return nil, fmt.Errorf("bad pane %q", r.Pane)
	}
	if !shaPattern.MatchString(r.Base) {
		return nil, fmt.Errorf("bad base %q", r.Base)
	}
	if !roundPattern.MatchString(r.Round) {
		return nil, fmt.Errorf("bad round %q", r.Round)
	}
	if r.Plan == "" || filepath.IsAbs(r.Plan) {
		return nil, fmt.Errorf("bad plan %q", r.Plan)
	}
	abs := filepath.Clean(filepath.Join(cfg.RepoRoot, r.Plan))
	rel, err := filepath.Rel(cfg.Root, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return nil, fmt.Errorf("plan %q is outside %s", r.Plan, cfg.Root)
	}
	it := b.Get(strings.TrimSuffix(filepath.ToSlash(rel), ".md"))
	if it == nil || it.Kind != board.KindPlan {
		return nil, fmt.Errorf("no plan %s", r.Plan)
	}
	return it, nil
}

// gitIn runs one git command in the repo and returns its trimmed output.
// Arguments stay separate, so nothing in the repo can become shell text.
func gitIn(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func cmdDispatchInit(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("dispatch init", stderr)
	pane := fs.String("pane", "", "orchestrator pane id, for example wM:pH")
	plan := fs.String("plan", "", "plan path relative to the repo root")
	round := fs.String("round", "", "short slug for this round (default: the branch name)")
	if err := fs.Parse(args); err != nil {
		return exitBadInput
	}
	cfg, b, code := loadBoard(*root, stderr)
	if code != exitOK {
		return code
	}
	if !cfg.IsGit {
		fmt.Fprintln(stderr, "dispatch init needs a git repo, so it can read the head commit")
		return exitBadInput
	}
	base, err := gitIn(cfg.RepoRoot, "rev-parse", "HEAD")
	if err != nil {
		fmt.Fprintln(stderr, "cannot read the head commit:", err)
		return exitOther
	}
	slug := *round
	if slug == "" {
		branch, err := gitIn(cfg.RepoRoot, "rev-parse", "--abbrev-ref", "HEAD")
		if err != nil {
			fmt.Fprintln(stderr, "cannot read the branch name:", err)
			return exitOther
		}
		s, ok := roundFromBranch(branch)
		if !ok {
			fmt.Fprintf(stderr, "cannot make a round from branch %q: pass --round <slug>\n", branch)
			return exitBadInput
		}
		slug = s
	}
	r := dispatchRecord{Pane: *pane, Base: base, Plan: *plan, Round: slug}
	if _, err := checkRecord(cfg, b, r); err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	// A missing ignore line would put the record in the next commit, so say it
	// out loud instead of failing: the record itself is already good.
	if err := hook.EnsureGitignore(cfg.Root, ".dispatch.json"); err != nil {
		fmt.Fprintln(stderr, "cannot add .dispatch.json to", filepath.Join(cfg.Root, ".gitignore")+":", err)
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	path := recordPath(cfg)
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	fmt.Fprintf(stdout, "dispatch record: %s\n", path)
	return exitOK
}

// cmdReplyBack tells the orchestrator pane that this round is done, or why it
// is stuck. It sends nothing unless every check passes, so a refused
// reply-back never looks like a finished round.
func cmdReplyBack(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("reply-back", stderr)
	blocked := fs.String("blocked", "", "say the round is blocked and why")
	if err := fs.Parse(args); err != nil {
		return exitBadInput
	}
	// An empty --blocked is a typo, not a reason, so the flag's own presence
	// decides this branch and not the value.
	asked := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "blocked" {
			asked = true
		}
	})
	cfg, b, code := loadBoard(*root, stderr)
	if code != exitOK {
		return code
	}
	data, err := os.ReadFile(recordPath(cfg))
	if err != nil {
		fmt.Fprintln(stderr, "no dispatch record, so there is no pane to answer:", err)
		return exitBadInput
	}
	var r dispatchRecord
	if err := json.Unmarshal(data, &r); err != nil {
		fmt.Fprintln(stderr, "broken dispatch record:", err)
		return exitBadInput
	}
	plan, err := checkRecord(cfg, b, r)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	// The own pane goes into a message a human reads, so a missing or odd
	// value says "unknown" instead of passing junk on.
	own := os.Getenv("HERDR_PANE_ID")
	if !panePattern.MatchString(own) {
		own = "unknown"
	}
	var text string
	if asked {
		if *blocked == "" {
			fmt.Fprintln(stderr, "--blocked needs a reason")
			return exitBadInput
		}
		text = fmt.Sprintf("%s blocked: %s - pane %s", r.Round, *blocked, own)
	} else {
		var open []string
		for _, id := range plan.Children {
			if c := b.Get(id); c == nil || c.Status != "done" {
				open = append(open, id)
			}
		}
		if len(open) > 0 {
			fmt.Fprintln(stderr, "open tasks:")
			for _, id := range open {
				fmt.Fprintln(stderr, id)
			}
			return exitBadInput
		}
		head, err := gitIn(cfg.RepoRoot, "rev-parse", "HEAD")
		if err != nil {
			fmt.Fprintln(stderr, "cannot read the head commit:", err)
			return exitOther
		}
		text = fmt.Sprintf("/acta:review %s..%s - plan %s, round %s, pane %s", r.Base, head, r.Plan, r.Round, own)
	}
	// Arguments stay separate, so a value from the record file can never
	// become shell text.
	cmd := exec.Command("herdr", "agent", "prompt", r.Pane, text)
	var herdrErr strings.Builder
	cmd.Stderr = &herdrErr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(stderr, "herdr: %v %s\n", err, herdrErr.String())
		return exitOther
	}
	fmt.Fprintf(stdout, "sent to %s: %s\n", r.Pane, text)
	return exitOK
}
