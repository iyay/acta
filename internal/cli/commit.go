package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/iyay/acta/internal/gitc"
)

// cmdCommit commits one planning file. An agent that edits a spec or a plan
// calls it instead of git commit, so repeat edits fold into the file's last
// planning commit and leave one commit, not a long run of them.
func cmdCommit(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("commit", stderr)
	msg := fs.String("m", "", "commit message")
	usage := func() int {
		fmt.Fprintln(stderr, `usage: acta commit <path> -m "<message>"`)
		return exitBadInput
	}
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) != 1 {
		return usage()
	}
	if strings.TrimSpace(*msg) == "" {
		fmt.Fprintln(stderr, "commit message is empty: pass -m \"<message>\"")
		return usage()
	}
	cfg, code := loadConfig(*root, stderr)
	if code != exitOK {
		return code
	}
	abs, err := planningFile(cfg.Root, pos[0])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	o := gitc.Result{Reason: "auto_commit is off"}
	if cfg.AutoCommit {
		o = gitc.CommitOrFold(cfg.RepoRoot, abs, *msg, false)
	}
	// Show the path as the caller typed it, not the link-free one git got.
	line := pos[0]
	if typed, err := filepath.Abs(pos[0]); err == nil {
		if cwd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(cwd, typed); err == nil {
				line = rel
			}
		}
	}
	fmt.Fprintln(stdout, filepath.ToSlash(line))
	if !o.Committed {
		fmt.Fprintln(stderr, "written, not committed:", o.Reason)
		return exitSkipped
	}
	return exitOK
}

// planningFile returns the absolute path of a file that exists and sits under
// the planning root. Links are followed first, so a path cannot slip out of
// the root through a link or a dot-dot step.
func planningFile(planRoot, path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("no such file: %s", path)
	}
	st, err := os.Stat(real)
	if err != nil || st.IsDir() {
		return "", fmt.Errorf("not a file: %s", path)
	}
	rootReal, err := filepath.EvalSymlinks(planRoot)
	if err != nil {
		return "", fmt.Errorf("planning root %s: %w", planRoot, err)
	}
	rel, err := filepath.Rel(rootReal, real)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%s is outside the planning root %s", path, planRoot)
	}
	return real, nil
}
