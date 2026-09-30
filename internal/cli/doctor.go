package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/doctor"
	"github.com/iyay/acta/internal/gitc"
	"github.com/iyay/acta/internal/hook"
	"github.com/iyay/acta/internal/theme"
)

const doctorUsage = "usage: acta doctor [--fix] [--known <file>]"

// The clashes list names the workflow plugins that pull an agent two ways.
// It lives in the plugin folder, which an installed binary does not carry,
// so --known names it.

func cmdDoctor(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fix := fs.Bool("fix", false, "fix repo problems (.acta folder, .gitignore)")
	known := fs.String("known", "", "file listing workflow plugins that clash with acta")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		fmt.Fprintln(stderr, doctorUsage)
		return exitBadInput
	}
	e := doctorEnv(*known)
	if *fix {
		// The file was dirty before --fix touched it, so its other lines
		// are the user's own and are not ours to commit.
		dirty, _ := gitc.IsDirty(e.RepoRoot, filepath.Join(e.ActaRoot, ".gitignore"))
		paths, err := doctor.Fix(e)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitOther
		}
		if len(paths) > 0 {
			// A commit that does not happen is worth saying, but the report
			// below is the answer the user asked for.
			if reason := skipReason(e, dirty); reason != "" {
				fmt.Fprintln(stderr, "fixed, not committed:", reason)
			} else if r := gitc.CommitPaths(e.RepoRoot, paths, "acta: doctor fix"); !r.Committed {
				fmt.Fprintln(stderr, "fixed, not committed:", r.Reason)
			}
			fmt.Fprintf(stdout, "fixed repo: %s\n", strings.Join(paths, ", "))
		}
	}
	rs := doctor.Run(e)
	fmt.Fprint(stdout, doctor.Format(rs))
	if doctor.Failed(rs) {
		return 1
	}
	return exitOK
}

// doctorEnv reads everything the checks need. A repo that cannot be read
// leaves RepoRoot empty, which makes the repo check skip instead of guess.
// A config that does not parse is carried as ConfigErr, so the repo check
// can report it instead of hiding it behind a skip.
func doctorEnv(known string) doctor.Env {
	e := doctor.Env{Home: homeDir(), ClaudeDir: hook.ClaudeDir(), KnownFile: known, AutoCommit: true}
	if e.KnownFile == "" {
		e.KnownFile = linkedKnownFile(e.Home)
	}
	if cwd, err := os.Getwd(); err == nil {
		cfg, err := config.Load(cwd, "")
		switch {
		case err != nil:
			e.ConfigErr = err
		case cfg.IsGit:
			e.RepoRoot, e.ActaRoot, e.AutoCommit = cfg.RepoRoot, cfg.Root, cfg.AutoCommit
			e.SchemaProblems = schemaProblems(cfg)
		}
	}
	if p, err := os.Executable(); err == nil {
		e.Binary, e.Version = p, buildVersion()
	}
	v, exists, _ := config.ResolveUser()
	e.Voice, e.VoiceExists = v, exists
	// Load the theme the voice file names, so the check reports the same
	// error the TUI will hit.
	_, e.ThemeErr = theme.Load(e.Voice.Theme)
	return e
}

// schemaProblems names every board file that opted in to the body schema and
// lost a section. Doctor only prints the list, so the reading lives here. A
// board that cannot load adds nothing, because the repo check already covers
// a broken repo.
func schemaProblems(cfg config.Config) []string {
	b, err := board.Load(cfg)
	if err != nil {
		return nil
	}
	var out []string
	for _, it := range b.Items {
		switch it.Kind {
		case board.KindTask, board.KindDebtItem:
			continue // a task and a debt line have no sections of their own
		}
		if it.Legacy {
			continue
		}
		raw, err := os.ReadFile(it.Path)
		if err != nil {
			continue // an item read from a branch that is not checked out
		}
		d := board.Parse(raw)
		if !board.HasSchema(d.Front) {
			continue // an old file, left as it is
		}
		for _, p := range board.CheckBody(it.Kind, d.Body) {
			out = append(out, string(it.Kind)+" "+filepath.Base(it.Path)+": "+p)
		}
	}
	sort.Strings(out)
	return out
}

// skipReason says why --fix will not commit, or "" when it will. The rules
// are the ones the other write commands follow: auto_commit off means the
// user wants nothing committed, and a file that was already dirty holds
// their own changes, which are not ours to commit.
func skipReason(e doctor.Env, dirty bool) string {
	switch {
	case !e.AutoCommit:
		return "auto_commit is off"
	case dirty:
		return "file had other uncommitted changes"
	}
	return ""
}

// linkedKnownFile finds the list of clashing plugins inside the plugin folder
// omp links, so the conflicts check works without a flag. A missing link
// gives "", and the check then skips.
func linkedKnownFile(home string) string {
	link := filepath.Join(home, ".omp", "plugins", "node_modules", "acta")
	dir := link
	if target, err := os.Readlink(link); err == nil {
		dir = target
	}
	p := filepath.Join(dir, "hooks", "workflow-plugins.txt")
	if _, err := os.Stat(p); err != nil {
		return ""
	}
	return p
}

// homeDir is the user home, or "" when the system cannot say. Every path
// under it is then simply missing, which the checks already handle.
func homeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

// buildVersion is the version the binary was built with, with the commit
// appended when the build stamped one, so a stale binary is easy to name.
func buildVersion() string {
	version := "dev"
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		version = v
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" && s.Value != "" {
			version += " (" + s.Value + ")"
		}
	}
	return version
}
