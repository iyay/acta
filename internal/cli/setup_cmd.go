package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/iyay/acta/internal/doctor"
	"github.com/iyay/acta/internal/setup"
)

const setupUsage = "usage: acta setup [--plugin-dir <path>]"

// setupRunner runs harness install commands through the real process, so
// the wizard carries out what the user said yes to.
type setupRunner struct{}

func (setupRunner) Run(argv []string) error {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// stdoutIsTTY says whether stdout is a terminal. os.Stdout is the real one
// in production; tests pass a strings.Builder, which is never a terminal.
func stdoutIsTTY(out io.Writer) bool {
	f, ok := out.(*os.File)
	if !ok {
		return false
	}
	st, err := f.Stat()
	if err != nil {
		return false
	}
	return st.Mode()&os.ModeCharDevice != 0
}

// cmdSetup runs the interactive setup wizard. Without a terminal on both
// stdin and stdout it writes nothing, names acta config set, and exits
// non-zero.
func cmdSetup(args []string, stdin io.Reader, stdinIsTTY bool, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	fs.SetOutput(stderr)
	pluginDir := fs.String("plugin-dir", "", "folder holding the acta plugin, used for harness installs")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		fmt.Fprintln(stderr, setupUsage)
		return exitBadInput
	}
	if !stdinIsTTY || !stdoutIsTTY(stdout) {
		fmt.Fprintln(stderr, "acta setup needs a terminal: nothing was written.")
		fmt.Fprintln(stderr, "Set each value by hand with `acta config set`, or re-run `acta setup` in a terminal.")
		return exitOther
	}
	env := setupEnv(*pluginDir)
	fmt.Fprint(stdout, doctor.Format(doctor.Run(doctorEnv(""))))
	answers, err := setup.Ask(env)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	if env.RepoRoot != "" {
		showBlock(stdout, env)
	}
	if err := setup.Apply(setup.Plan(answers, env), setupRunner{}, stdout); err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	fmt.Fprintln(stdout, "Done. Next: open a repo and run acta to start planning.")
	return exitOK
}

// setupEnv reads what the wizard found: harnesses on PATH, the git repo
// holding the working directory, which of CLAUDE.md and AGENTS.md it holds,
// and the current user config the form shows as defaults.
func setupEnv(pluginDir string) setup.Env {
	env := setup.Env{TTY: true, PluginDir: pluginDir}
	for _, h := range []string{"claude", "omp"} {
		if _, err := exec.LookPath(h); err == nil {
			env.Harnesses = append(env.Harnesses, h)
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		if out, err := exec.Command("git", "-C", cwd, "rev-parse", "--show-toplevel").Output(); err == nil {
			env.RepoRoot = strings.TrimSpace(string(out))
		}
		if env.RepoRoot != "" {
			_, errClaude := os.Stat(filepath.Join(env.RepoRoot, "CLAUDE.md"))
			_, errAgents := os.Stat(filepath.Join(env.RepoRoot, "AGENTS.md"))
			env.HasClaudeMD = errClaude == nil
			env.HasAgentsMD = errAgents == nil
		}
	}
	return env
}

// showBlock prints the mandatory acta block and the file it goes to, before
// Apply writes it. The block is never skipped and never confirmed.
func showBlock(out io.Writer, env setup.Env) {
	fmt.Fprintln(out, "The acta block goes to:")
	switch {
	case env.HasClaudeMD && env.HasAgentsMD:
		fmt.Fprintln(out, filepath.Join(env.RepoRoot, "CLAUDE.md"))
		fmt.Fprintln(out, filepath.Join(env.RepoRoot, "AGENTS.md"))
	case env.HasAgentsMD:
		fmt.Fprintln(out, filepath.Join(env.RepoRoot, "AGENTS.md"))
	default:
		fmt.Fprintln(out, filepath.Join(env.RepoRoot, "CLAUDE.md"))
	}
	fmt.Fprint(out, setup.Block)
}
