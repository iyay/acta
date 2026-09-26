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

	"pm-board/internal/board"
	"pm-board/internal/config"
	"pm-board/internal/gitc"
)

var (
	// ErrBadInput is a request the contract does not allow. Nothing is written.
	ErrBadInput = errors.New("bad input")
	// ErrUnchanged is a new bug file the user left as the bare template.
	ErrUnchanged = errors.New("template unchanged")
	// Now is the clock for new bug file names. Tests replace it.
	Now = time.Now

	slugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

// Outcome says what happened to the file after a write.
type Outcome struct {
	Path      string
	Committed bool
	Skipped   bool // auto-commit was on but the commit did not happen
	Reason    string
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
		if !allowed(board.Allowed(it.Kind), value) {
			return Outcome{}, bad("%q is not a %s status (%s)", value, it.Kind, strings.Join(board.Allowed(it.Kind), ", "))
		}
	default:
		return Outcome{}, bad("unknown field %q; use status or type", field)
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
	if err := os.WriteFile(it.Path, out, 0o644); err != nil {
		return Outcome{}, err
	}
	return finish(cfg, it.Path, fmt.Sprintf("pm: %s %s %s", id, field, value), dirty), nil
}

// NewBug writes a bug file from a body an agent sent and commits it.
func NewBug(cfg config.Config, slug, title, ref string, body []byte) (Outcome, error) {
	path, err := bugPath(cfg, slug)
	if err != nil {
		return Outcome{}, err
	}
	content := BugFile(title, slug, ref, body)
	if !HasSymptom(content) {
		return Outcome{}, bad("a bug needs a ## Symptom section")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return Outcome{}, err
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return Outcome{}, err
	}
	return finish(cfg, path, "pm: new bug "+strings.TrimSuffix(filepath.Base(path), ".md"), false), nil
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
	return finish(cfg, path, "pm: new bug "+strings.TrimSuffix(filepath.Base(path), ".md"), false), nil
}

func bugPath(cfg config.Config, slug string) (string, error) {
	if !slugRe.MatchString(slug) {
		return "", bad("slug %q must be lower case words joined by -", slug)
	}
	path := filepath.Join(cfg.Root, cfg.Dirs.Bugs, Now().Format("2006-01-02")+"-"+slug+".md")
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
