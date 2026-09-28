package cli

import (
	"encoding/json"
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
		slug = branch
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

// cmdReplyBack is a stub so the build compiles. Wave 2 of the plan replaces
// it with the real command that messages the orchestrator pane.
func cmdReplyBack(args []string, stdout, stderr io.Writer) int {
	fmt.Fprintln(stderr, "usage: acta reply-back [--blocked \"<reason>\"]")
	return exitBadInput
}
