package write

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/gitc"
)

var (
	// ErrBadInput is a request the contract does not allow. Nothing is written.
	ErrBadInput = errors.New("bad input")
	// ErrUnchanged is a new bug file the user left as the bare template.
	ErrUnchanged = errors.New("template unchanged")
	// Now is the clock for new bug file names. Tests replace it.
	Now = time.Now

	slugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	shaRe  = regexp.MustCompile(`^[0-9a-f]{7,40}$`)
	refRe  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]{0,39}$`)
)

// Outcome says what happened to the file after a write.
type Outcome struct {
	Path      string
	ShortID   string // number ID like SCRATCH-1, empty when the write has none
	Committed bool
	Skipped   bool // auto-commit was on but the commit did not happen
	Reason    string
	Skips     []string // one "skip <path id>: <reason>" line per file left alone
}

func bad(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrBadInput, fmt.Sprintf(format, args...))
}

// SetValue sets the status or type of one story or bug and commits the file.
func SetValue(cfg config.Config, b *board.Board, id, field, value string) (Outcome, error) {
	it := b.Get(id)
	switch {
	case it == nil:
		return Outcome{}, bad("unknown id %s", id)
	case it.Kind == board.KindTask:
		return Outcome{}, bad("tasks take their status from their checkboxes")
	case it.Legacy:
		return Outcome{}, bad("%s is a legacy file; move it into the root folder first", id)
	}
	switch field {
	case "type":
		if value != string(board.KindStory) && value != string(board.KindBug) {
			return Outcome{}, bad("type must be story or bug, not %q", value)
		}
	case "status":
		// specced is not written anywhere: a spec's parent link gives it.
		if it.Kind == board.KindScratch && value == "specced" {
			return Outcome{}, bad("specced comes from a spec's parent link")
		}
		if !allowed(board.Allowed(it.Kind), value) {
			return Outcome{}, bad("%q is not a %s status (%s)", value, it.Kind, strings.Join(board.Allowed(it.Kind), ", "))
		}
	case "fixed_in":
		if it.Kind != board.KindBug {
			return Outcome{}, bad("fixed_in is only for bugs")
		}
		if !shaRe.MatchString(value) {
			return Outcome{}, bad("fixed_in must be a commit sha of 7 to 40 lower-case hex characters, not %q", value)
		}
	case "ref":
		if !refRe.MatchString(value) {
			return Outcome{}, bad("ref %q must be one word of letters, digits and . _ / -, at most 40 characters", value)
		}
	default:
		return Outcome{}, bad("unknown field %q; use status, type, fixed_in or ref", field)
	}

	dirty, err := dirtyBefore(cfg, it.Path)
	if err != nil {
		return Outcome{}, err
	}
	src, err := os.ReadFile(it.Path)
	if err != nil {
		return Outcome{}, err
	}
	out, err := SetField(src, field, value)
	if err != nil {
		return Outcome{}, bad("%s: %v", id, err)
	}
	switch field {
	case "status":
		out, err = DatesFor(out, value)
	case "fixed_in":
		out, err = MarkFinished(out)
	}
	if err != nil {
		return Outcome{}, bad("%s: %v", id, err)
	}
	if err := os.WriteFile(it.Path, out, 0o644); err != nil {
		return Outcome{}, err
	}
	return finish(cfg, it.Path, fmt.Sprintf("acta: %s %s %s", id, field, value), dirty), nil
}

// NewBug writes a bug file from a body an agent sent and commits it. The new
// file gets the next BUG number and a fresh hash before it is written.
func NewBug(cfg config.Config, slug, title, ref string, body []byte) (Outcome, error) {
	path, err := bugPath(cfg, slug)
	if err != nil {
		return Outcome{}, err
	}
	content := BugFile(title, slug, ref, body)
	if !HasSymptom(content) {
		return Outcome{}, bad("a bug needs a ## Symptom section")
	}
	b, err := board.Load(cfg)
	if err != nil {
		return Outcome{}, err
	}
	next, taken := scanIDs(b)
	bugPrefix := board.Prefix(board.KindBug, false)
	content, err = SetField(content, "id", board.FormatID(bugPrefix, next[bugPrefix]))
	if err != nil {
		return Outcome{}, err
	}
	content, err = SetField(content, "hash", freeHash(taken))
	if err != nil {
		return Outcome{}, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return Outcome{}, err
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return Outcome{}, err
	}
	return finish(cfg, path, "acta: new bug "+strings.TrimSuffix(filepath.Base(path), ".md"), false), nil
}

// StartBug writes the bare template so an editor can open it.
func StartBug(cfg config.Config, slug, ref string) (string, []byte, error) {
	path, err := bugPath(cfg, slug)
	if err != nil {
		return "", nil, err
	}
	tmpl := BugTemplate(strings.ReplaceAll(slug, "-", " "), ref)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", nil, err
	}
	if err := os.WriteFile(path, tmpl, 0o644); err != nil {
		return "", nil, err
	}
	return path, tmpl, nil
}

// FinishBug runs after the editor closes. A file still equal to the template
// is removed, so an aborted bug leaves nothing behind.
func FinishBug(cfg config.Config, path string, tmpl []byte) (Outcome, error) {
	now, err := os.ReadFile(path)
	if err != nil {
		return Outcome{}, err
	}
	if bytes.Equal(now, tmpl) {
		if err := os.Remove(path); err != nil {
			return Outcome{}, err
		}
		return Outcome{Path: path}, ErrUnchanged
	}
	return finish(cfg, path, "acta: new bug "+strings.TrimSuffix(filepath.Base(path), ".md"), false), nil
}

// NewDebt writes review NOTEs onto a plan as a checklist file in the debt
// folder. Running it again for the same plan on the same day appends only
// the lines that are not already on the file, so paging the same NOTEs
// twice never duplicates a line or a commit.
func NewDebt(cfg config.Config, b *board.Board, planID, title string, notes []byte) (Outcome, error) {
	it := b.Get(planID)
	if it == nil {
		return Outcome{}, bad("unknown id %s", planID)
	}
	if it.Kind != board.KindPlan {
		return Outcome{}, bad("%s is not a plan", planID)
	}
	texts := splitNotes(notes)
	if len(texts) == 0 {
		return Outcome{}, bad("no notes on stdin")
	}
	// The debt file names itself after today, not the day the plan was
	// written, so it keeps the plan's slug and drops the plan's own date.
	path := filepath.Join(cfg.Root, cfg.Dirs.Debt, Now().Format("2006-01-02")+"-"+it.Slug+".md")
	fileStem := strings.TrimSuffix(filepath.Base(path), ".md")

	src, err := os.ReadFile(path)
	switch {
	case err == nil:
		return appendDebt(cfg, path, fileStem, src, texts)
	case !os.IsNotExist(err):
		return Outcome{}, err
	}
	return createDebt(cfg, b, path, fileStem, title, it, texts)
}

// appendDebt adds only the lines the debt file does not already hold.
// Nothing new means nothing written and nothing committed. The caller's src
// is not used: the file is read again under the lock, so a tick that landed
// after the caller read it is not thrown away.
func appendDebt(cfg config.Config, path, fileStem string, src []byte, texts []string) (Outcome, error) {
	dirty, added, err := addDebtLines(cfg, path, texts)
	if err != nil {
		return Outcome{}, err
	}
	if !added {
		return Outcome{Path: path}, nil
	}
	return finish(cfg, path, "acta: new debt "+fileStem, dirty), nil
}

// addDebtLines writes the new lines onto the debt file at path, under that
// file's lock, through a temp file and a rename, the way TickLine does it.
// TickLine locks the same file, so the two can never read each other half
// written or throw away each other's change. It says whether the file was
// already dirty in git before we wrote, and whether it wrote anything.
func addDebtLines(cfg config.Config, path string, texts []string) (bool, bool, error) {
	unlock, err := lock(path)
	if err != nil {
		return false, false, err
	}
	defer unlock()
	src, err := os.ReadFile(path)
	if err != nil {
		return false, false, err
	}
	// The git status has to be read before we write, otherwise our own write
	// is the one that makes the file look dirty and the commit is skipped.
	dirty, err := dirtyBefore(cfg, path)
	if err != nil {
		return false, false, err
	}
	have := map[string]bool{}
	for _, line := range board.Parse(src).Items {
		have[line.Text] = true
	}
	var add []string
	for _, t := range texts {
		if !have[t] {
			have[t] = true
			add = append(add, t)
		}
	}
	if len(add) == 0 {
		return false, false, nil
	}
	out := string(src)
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	for _, t := range add {
		out += "- [ ] " + t + "\n"
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(out), 0o644); err != nil {
		return false, false, err
	}
	return dirty, true, os.Rename(tmp, path)
}

// createDebt writes a fresh debt file with its own id, hash and parent link
// back to the plan the NOTEs came from.
func createDebt(cfg config.Config, b *board.Board, path, fileStem, title string, plan *board.Item, texts []string) (Outcome, error) {
	heading := title
	if heading == "" {
		heading = "Review NOTEs: " + plan.Title
	}
	var body strings.Builder
	body.WriteString("# " + heading + "\n\n")
	for _, t := range texts {
		body.WriteString("- [ ] " + t + "\n")
	}
	content := []byte(body.String())
	next, taken := scanIDs(b)
	var err error
	debtPrefix := board.Prefix(board.KindDebt, false)
	if content, err = SetField(content, "id", board.FormatID(debtPrefix, next[debtPrefix])); err != nil {
		return Outcome{}, err
	}
	if content, err = SetField(content, "hash", freeHash(taken)); err != nil {
		return Outcome{}, err
	}
	if content, err = SetField(content, "parent", fileID(cfg, plan.Path)); err != nil {
		return Outcome{}, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return Outcome{}, err
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return Outcome{}, err
	}
	return finish(cfg, path, "acta: new debt "+fileStem, false), nil
}

// splitNotes turns raw stdin into one trimmed line of text per note. A
// leading "- " bullet is dropped, and blank lines are skipped, so the
// command reads either a plain list or a markdown checklist.
func splitNotes(notes []byte) []string {
	var out []string
	for _, ln := range strings.Split(string(notes), "\n") {
		ln = strings.TrimSpace(ln)
		ln = strings.TrimPrefix(ln, "- ")
		ln = strings.TrimSpace(ln)
		if ln != "" {
			out = append(out, ln)
		}
	}
	return out
}

func bugPath(cfg config.Config, slug string) (string, error) {
	return datedPath(cfg, cfg.Dirs.Bugs, slug)
}

// datedPath gives a new planning file its date and slug name, and refuses a
// slug the contract does not allow or a name the folder already holds.
func datedPath(cfg config.Config, dir, slug string) (string, error) {
	if !slugRe.MatchString(slug) {
		return "", bad("slug %q must be lower case words joined by -", slug)
	}
	path := filepath.Join(cfg.Root, dir, Now().Format("2006-01-02")+"-"+slug+".md")
	if _, err := os.Stat(path); err == nil {
		return "", bad("%s already exists", path)
	}
	return path, nil
}

func dirtyBefore(cfg config.Config, path string) (bool, error) {
	if !cfg.IsGit {
		return false, nil
	}
	return gitc.IsDirty(cfg.RepoRoot, path)
}

func finish(cfg config.Config, path, msg string, dirty bool) Outcome {
	o := Outcome{Path: path}
	if !cfg.AutoCommit {
		o.Reason = "auto_commit is off"
		return o
	}
	r := gitc.Commit(cfg.RepoRoot, path, msg, dirty)
	o.Committed, o.Reason = r.Committed, r.Reason
	o.Skipped = !r.Committed
	return o
}

func allowed(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
