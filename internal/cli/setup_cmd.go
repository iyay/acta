package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/doctor"
	"github.com/iyay/acta/internal/setup"
)

const setupUsage = "usage: acta setup [--plugin-dir <path>]"

// setupNext is the step the closing box points to after a clean run.
const setupNext = "open a repo and run acta to start planning"

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
	fmt.Fprint(stdout, setup.DoctorSummary(doctor.Run(doctorEnv(""))))
	answers, err := setup.Ask(env)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	var blockFiles []string
	if env.RepoRoot != "" {
		blockFiles = blockPaths(env)
		fmt.Fprint(stdout, setup.BlockLines(blockFiles))
	}
	actions := setup.Plan(answers, env)
	if err := setup.Apply(actions, setupRunner{}, stdout); err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	fmt.Fprint(stdout, setupSummary(actions, setupNext))
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
	// The form shows the current config as its defaults, so read it here
	// and let the form fall back to the built-ins when there is none. A
	// broken file still yields the defaults from ResolveUser, never a
	// failed wizard, so the error stays ignored.
	u, _, _ := config.ResolveUser()
	env.Current = u
	return env
}

// blockPaths lists the files Apply will write the acta block to: one per
// file that exists, or a new CLAUDE.md when neither does. It mirrors what
// Plan turns into block actions, so the short note before Apply matches the
// files Apply really writes.
func blockPaths(env setup.Env) []string {
	switch {
	case env.HasClaudeMD && env.HasAgentsMD:
		return []string{
			filepath.Join(env.RepoRoot, "CLAUDE.md"),
			filepath.Join(env.RepoRoot, "AGENTS.md"),
		}
	case env.HasAgentsMD:
		return []string{filepath.Join(env.RepoRoot, "AGENTS.md")}
	default:
		return []string{filepath.Join(env.RepoRoot, "CLAUDE.md")}
	}
}

// showBlock names each file the acta block goes to, one short line each,
// before Apply writes it. The block is never skipped and never confirmed.
func showBlock(out io.Writer, env setup.Env) {
	fmt.Fprint(out, setup.BlockLines(blockPaths(env)))
}

// setupSummary closes the wizard: where the config landed, which harnesses
// the plan ran installs for, which block files the plan wrote, and what to
// run next. The names come from the plan actions Apply just carried out, so
// the box matches what really ran.
func setupSummary(actions []setup.Action, next string) string {
	path, err := config.UserPath()
	if err != nil {
		path = ""
	}
	var installed, blocks []string
	for _, a := range actions {
		switch a.Kind {
		case setup.ActionInstall:
			installed = append(installed, a.Harness)
		case setup.ActionBlock:
			blocks = append(blocks, a.Path)
		}
	}
	return setup.SummaryBox(path, installed, blocks, next)
}
