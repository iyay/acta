package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"unicode"

	"github.com/iyay/acta/internal/usage"
)

// cmdUsage reports the Codex usage recorded for this repository. It reads
// local transcripts only: no network, no cache, no writes.
func cmdUsage(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("usage", flag.ContinueOnError)
	// The flag package's own usage text would drift from the help table, so
	// the parse error is reported and the table's line is printed instead.
	fs.SetOutput(io.Discard)
	asJSON := fs.Bool("json", false, "print one JSON report")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(stderr, err)
		fmt.Fprintln(stderr, usageLine("usage"))
		return exitBadInput
	}
	if fs.NArg() > 0 {
		fmt.Fprintln(stderr, usageLine("usage"))
		return exitBadInput
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	sessions, err := codexSessionsDir(cwd)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	report, err := usage.Read(cwd, sessions)
	if err != nil {
		if errors.Is(err, usage.ErrNotRepository) {
			// Outside a repository there is no scope, which is a bad call
			// rather than a failure of the machine or the data.
			fmt.Fprintln(stderr, "acta usage: not inside a Git repository")
			return exitBadInput
		}
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	if *asJSON {
		if code := printJSON(stdout, stderr, report); code != exitOK {
			return code
		}
	} else {
		if err := renderUsage(stdout, report); err != nil {
			fmt.Fprintln(stderr, err)
			return exitOther
		}
		// Human readers need the reasons on stderr; a JSON report carries
		// the same diagnostics as data instead.
		renderDiagnostics(stderr, report.Diagnostics)
	}
	if report.Status == usage.StatusPartial {
		return exitSkipped
	}
	return exitOK
}

// codexSessionsDir is where Codex writes transcripts: $CODEX_HOME/sessions,
// or ~/.codex/sessions when the variable is unset or empty. A relative
// CODEX_HOME is read from the working directory.
func codexSessionsDir(cwd string) (string, error) {
	home := os.Getenv("CODEX_HOME")
	if home == "" {
		dir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("find the Codex sessions folder: %w", err)
		}
		return filepath.Join(dir, ".codex", "sessions"), nil
	}
	if !filepath.IsAbs(home) {
		home = filepath.Join(cwd, home)
	}
	return filepath.Join(home, "sessions"), nil
}

// renderUsage prints the human report: one table of root sessions and
// models, then the per-model summary. An unknown range end reads "unknown"
// instead of an invented time.
func renderUsage(w io.Writer, r usage.Report) error {
	if r.Status == usage.StatusNoData {
		// The limitation matters most here: an empty report may only mean
		// that the checkout a session ran in is gone.
		_, err := fmt.Fprintf(w, "No data: no Codex session in %s started in this repository.\nLimit: %s\n",
			r.Scope.SourceDirectory, strings.Join(r.Scope.Limitations, " "))
		return err
	}
	if _, err := fmt.Fprintf(w, "Repository: %s\nCheckouts: %s\nLogs: %s (%s)\nLimit: %s\n\n",
		r.Scope.Repository, strings.Join(r.Scope.Checkouts, ", "), r.Scope.SourceDirectory, r.Scope.HistoryCoverage,
		strings.Join(r.Scope.Limitations, " ")); err != nil {
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if len(r.Sessions) == 0 {
		fmt.Fprintln(tw, "No responses could be counted.")
	} else {
		fmt.Fprintln(tw, "SESSION\tMODEL\tFIRST\tLAST\tFRESH\tCACHED\tOUTPUT\tRESPONSES")
		for _, s := range r.Sessions {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\n",
				termSafe(s.SessionID), termSafe(s.Model),
				rangeEnd(s.Range.FirstResponseAt), rangeEnd(s.Range.LastResponseAt),
				s.FreshInputTokens, s.CachedInputTokens, s.OutputTokens, s.ResponseCount)
		}
		fmt.Fprintln(tw, "\nMODEL\tFRESH\tCACHED\tOUTPUT\tRESPONSES")
		for _, m := range r.Models {
			fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\n",
				termSafe(m.Model), m.FreshInputTokens, m.CachedInputTokens, m.OutputTokens, m.ResponseCount)
		}
	}
	return tw.Flush()
}

// renderDiagnostics writes one line per recoverable problem, so a human
// reader knows why the report is partial.
func renderDiagnostics(w io.Writer, diags []usage.Diagnostic) {
	for _, d := range diags {
		where := "report"
		if d.File != nil {
			where = *d.File
			if d.Line != nil {
				where = fmt.Sprintf("%s:%d", where, *d.Line)
			}
		}
		fmt.Fprintf(w, "usage: %s %s: %s\n", d.Code, termSafe(where), termSafe(d.Message))
	}
}

// rangeEnd names one end of a response range, or says it is unknown.
func rangeEnd(p *string) string {
	if p == nil {
		return "unknown"
	}
	return *p
}

// termSafe keeps text taken from a transcript from painting the terminal.
// Control characters are shown as escapes instead of being sent as bytes.
func termSafe(s string) string {
	var b strings.Builder
	for _, r := range s {
		if !unicode.IsControl(r) {
			b.WriteRune(r)
			continue
		}
		fmt.Fprintf(&b, `\u%04x`, r)
	}
	return b.String()
}
