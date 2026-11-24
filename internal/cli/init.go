package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/iyay/acta/internal/doctor"
	"github.com/iyay/acta/internal/gitc"
	"github.com/iyay/acta/internal/setup"
	"github.com/iyay/acta/internal/wiki"
)

var initUsage = usageLine("init")

// cmdInit sets up the repo it runs in: the .acta folder from doctor --fix
// and the acta block in CLAUDE.md or AGENTS.md, in one commit. It asks
// nothing, since the hooks read the settings from the user's own config.
func cmdInit(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		fmt.Fprintln(stderr, initUsage)
		return exitBadInput
	}
	e := doctorEnv("", true)
	if e.ConfigErr != nil {
		// A broken .acta.yaml also leaves RepoRoot empty, so say the real
		// cause instead of asking for git init in a repo that has git.
		fmt.Fprintln(stderr, "acta init:", e.ConfigErr)
		return exitOther
	}
	if e.RepoRoot == "" {
		// Every acta write is committed, so a folder outside git cannot
		// hold acta at all.
		fmt.Fprintln(stderr, "acta init: not inside a git repo; acta commits every change, so run git init first")
		return exitOther
	}
	targets := setup.BlockTargets(e.RepoRoot,
		fileExists(filepath.Join(e.RepoRoot, "CLAUDE.md")),
		fileExists(filepath.Join(e.RepoRoot, "AGENTS.md")))
	// Check every file before writing any, so a broken one never leaves
	// the repo half set up.
	for _, p := range targets {
		if err := setup.CheckBlock(p); err != nil {
			fmt.Fprintln(stderr, "acta init:", err)
			return exitOther
		}
	}
	watched := append([]string{
		filepath.Join(e.ActaRoot, ".gitignore"),
		filepath.Join(e.RepoRoot, ".gitattributes"),
		filepath.Join(e.ActaRoot, "wiki", wiki.CheckedFile),
	}, targets...)
	// A file the user already changed holds their own work, which is not
	// ours to commit.
	dirty, dirtyErr := false, error(nil)
	for _, p := range watched {
		d, err := gitc.IsDirty(e.RepoRoot, p)
		dirty, dirtyErr = dirty || d, errors.Join(dirtyErr, err)
	}
	before := readAll(targets)
	fixed, err := doctor.Fix(e)
	if err != nil {
		fmt.Fprintln(stderr, "acta init:", err)
		return exitOther
	}
	for _, p := range targets {
		if err := setup.WriteBlock(p); err != nil {
			fmt.Fprintln(stderr, "acta init:", err)
			return exitOther
		}
	}
	// WriteBlock rewrites a file even when its bytes stay the same, so
	// compare them to know what really changed.
	changed := fixed
	for _, p := range targets {
		after, _ := os.ReadFile(p)
		if old, ok := before[p]; !ok || !bytes.Equal(old, after) {
			changed = append(changed, p)
		}
	}
	code := commitInit(e, changed, dirty, dirtyErr, stdout, stderr)
	fmt.Fprint(stdout, doctor.Format(doctor.Run(e)))
	return code
}

// commitInit commits what init changed as one new commit. It never folds
// into HEAD, so a user's own commit is never rewritten.
func commitInit(e doctor.Env, changed []string, dirty bool, dirtyErr error, stdout, stderr io.Writer) int {
	if len(changed) == 0 {
		fmt.Fprintln(stdout, "repo is already set up")
		return exitOK
	}
	fmt.Fprintf(stdout, "set up repo: %s\n", strings.Join(changed, ", "))
	if reason := skipReason(e, dirty, dirtyErr); reason != "" {
		fmt.Fprintln(stdout, "written, not committed:", reason)
		return exitOK
	}
	const msg = "chore: acta init"
	// CommitPaths folds a lone path into HEAD, so one path goes through
	// the plain Commit instead.
	var r gitc.Result
	if len(changed) == 1 {
		r = gitc.Commit(e.RepoRoot, changed[0], msg, false)
	} else {
		r = gitc.CommitPaths(e.RepoRoot, changed, msg)
	}
	if !r.Committed {
		fmt.Fprintln(stderr, "acta init: written, not committed:", r.Reason)
		return exitOther
	}
	return exitOK
}

// readAll gives the bytes of each path that exists, so a path missing from
// the map was not there before.
func readAll(paths []string) map[string][]byte {
	out := map[string][]byte{}
	for _, p := range paths {
		if raw, err := os.ReadFile(p); err == nil {
			out[p] = raw
		}
	}
	return out
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
