package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/doctor"
	"github.com/iyay/acta/internal/gitc"
	"github.com/iyay/acta/internal/hook"
	"github.com/iyay/acta/internal/voice"
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
		paths, err := doctor.Fix(e)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitOther
		}
		if len(paths) > 0 {
			// A commit that does not happen is worth saying, but the report
			// below is the answer the user asked for.
			if r := gitc.CommitPaths(e.RepoRoot, paths, "acta: doctor fix"); !r.Committed {
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
func doctorEnv(known string) doctor.Env {
	e := doctor.Env{Home: homeDir(), ClaudeDir: hook.ClaudeDir(), KnownFile: known}
	if e.KnownFile == "" {
		e.KnownFile = linkedKnownFile(e.Home)
	}
	if cwd, err := os.Getwd(); err == nil {
		if cfg, err := config.Load(cwd, ""); err == nil && cfg.IsGit {
			e.RepoRoot, e.ActaRoot = cfg.RepoRoot, cfg.Root
		}
	}
	if p, err := os.Executable(); err == nil {
		e.Binary, e.Version = p, buildVersion()
	}
	v, exists, _ := voice.Resolve()
	e.Voice, e.VoiceExists = v, exists
	return e
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
