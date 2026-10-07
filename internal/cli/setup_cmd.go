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
	"github.com/iyay/acta/internal/hook"
	"github.com/iyay/acta/internal/setup"
)

const setupUsage = "usage: acta setup [--plugin-dir <path>]"

// setupRunner runs harness install commands through the real process, so
// the wizard carries out what the user said yes to.
type setupRunner struct{}

func (setupRunner) Run(argv []string) error {
	cmd := exec.Command(argv[0], argv[1:]...)
	// The install tool's own output would break the rail, so it stays off
	// screen. A failure prints the command to run by hand instead. Stdin
	// stays nil: an installer that asks a question gets end of input and
	// fails at once, instead of waiting on a prompt nobody can see.
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
	fmt.Fprint(stdout, setup.RailOpen())
	fmt.Fprint(stdout, setup.DoctorSummary(doctor.Run(doctorEnv(""))))
	answers, err := setup.Ask(env, stdout)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	actions := setup.Plan(answers, env)
	if err := setup.Apply(actions, setupRunner{}, stdout); err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	fmt.Fprint(stdout, setupSummary())
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
	denv := doctor.Env{Home: homeDir(), ClaudeDir: hook.ClaudeDir(), RepoRoot: env.RepoRoot}
	env.Installed = map[string]bool{}
	for _, h := range env.Harnesses {
		env.Installed[h] = doctor.PluginInstalled(denv, h)
	}
	// The form shows the config file's values as its defaults. A first run
	// has no file, so the form starts empty: the old voice files are never
	// read here, or a deleted config would still offer an old language. A
	// broken file reads as empty too, never a failed wizard.
	if path, err := config.UserPath(); err == nil {
		if u, exists, _ := config.LoadUser(path); exists {
			env.Current = u
		}
	}
	return env
}

// setupSummary closes the wizard with the config path and the TUI hint.
func setupSummary() string {
	path, err := config.UserPath()
	if err != nil {
		path = ""
	}
	return setup.SummaryBox(path)
}
