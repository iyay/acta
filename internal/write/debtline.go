package write

import (
	"os"
	"regexp"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// ownReasonRe finds the reason CloseDebtLine put at the end of a line. Its
// status word marks it, so a closing note a person wrote by hand, like
// "(stale)", is never taken for one and never replaced.
var ownReasonRe = regexp.MustCompile(` \((done|wontfix): .*\)$`)

// CloseDebtLine closes one debt line as done or wontfix, puts the reason at
// its end and commits the file. Closing it again swaps the reason, so a line
// never carries two.
func CloseDebtLine(cfg config.Config, b *board.Board, id, status, reason string) (Outcome, error) {
	line, path, err := closedDebtLine(b, id, status, reason)
	if err != nil {
		return Outcome{}, err
	}
	dirty, err := dirtyBefore(cfg, path)
	if err != nil {
		return Outcome{}, err
	}
	if err := rewriteDebtLine(path, line, status, reason); err != nil {
		return Outcome{}, err
	}
	return finish(cfg, path, Subject(cfg, []string{path}, subjectWhat(id, "status", status)), dirty), nil
}

// closedDebtLine checks a close before anything is written and returns the
// line number and the file it lives in.
func closedDebtLine(b *board.Board, id, status, reason string) (int, string, error) {
	it := b.Get(id)
	switch {
	case it == nil:
		return 0, "", bad("unknown id %s", id)
	case it.Kind != board.KindDebtItem:
		return 0, "", bad("%s is not a debt line", id)
	case it.Legacy:
		return 0, "", bad("%s is a legacy file; move it into the root folder first", id)
	case status != "done" && status != "wontfix":
		return 0, "", bad("a debt line closes as done or wontfix, not %q", status)
	}
	if strings.Trim(reason, "() \t") == "" {
		return 0, "", bad("closing a debt line needs a reason: --reason \"<text>\"")
	}
	if strings.ContainsAny(reason, "\r\n") {
		return 0, "", bad("reason must be one line, not %q", reason)
	}
	return it.Line, it.Path, nil
}

// rewriteDebtLine sets the box and the reason of one line under the file
// lock, so a tick running at the same time is not lost.
func rewriteDebtLine(path string, line int, status, reason string) error {
	unlock, err := lock(path)
	if err != nil {
		return err
	}
	defer unlock()
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(src), "\n")
	if line < 1 || line > len(lines) || !tickAnyBoxRe.MatchString(lines[line-1]) {
		return bad("line %d of %s is not a debt line any more; reload and try again", line, path)
	}
	box := "x"
	if status == "wontfix" {
		box = "-"
	}
	// A file with CRLF endings keeps its \r, so only this line changes.
	text, cr := strings.CutSuffix(lines[line-1], "\r")
	text = ownReasonRe.ReplaceAllString(strings.TrimRight(text, " \t"), "")
	text = tickAnyBoxRe.ReplaceAllString(text, "${1}"+box+"${3}")
	lines[line-1] = text + " (" + status + ": " + strings.TrimSpace(reason) + ")"
	if cr {
		lines[line-1] += "\r"
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
