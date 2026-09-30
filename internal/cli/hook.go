package cli

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/hook"
)

const hookUsage = "usage: acta hook session-start [--known <file>] | acta hook prompt | acta hook pre-tool | acta hook post-tool"

// exitBlock is the code a tool hook uses to tell the agent the command is
// stopped. It is the only non-zero code an acta hook ever returns.
const exitBlock = 2

// cmdHook prints hook text. Every hook exits 0 whatever stdin holds, because a
// hook that fails would get in the way of the user's session. The one
// exception is pre-tool stopping a second brainstorm, which exits 2 with the
// reason the agent has to read.
func cmdHook(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, hookUsage)
		return exitBadInput
	}
	switch args[0] {
	case "prompt":
		if len(args) != 1 {
			fmt.Fprintln(stderr, hookUsage)
			return exitBadInput
		}
		fmt.Fprintln(stdout, hook.Prompt(loadVoice()))
		if line := sessionReminder(stdin); line != "" {
			fmt.Fprintln(stdout, line)
		}
		return exitOK
	case "pre-tool", "post-tool":
		if len(args) != 1 {
			fmt.Fprintln(stderr, hookUsage)
			return exitBadInput
		}
		ev, ok := hook.ParseEvent(stdin)
		if !ok {
			return exitOK
		}
		root, ok := hookRoot()
		if !ok {
			return exitOK
		}
		if args[0] == "post-tool" {
			// Both writes are dropped on purpose: a session that cannot
			// remember its brainstorm must not be stopped for it.
			_ = hook.EnsureGitignore(root, "state/")
			_ = hook.RecordBrainstorm(root, ev)
			return exitOK
		}
		if block, msg := hook.PreTool(root, ev); block {
			fmt.Fprintln(stderr, msg)
			return exitBlock
		}
		return exitOK
	case "session-start":
		fs := flag.NewFlagSet("hook session-start", flag.ContinueOnError)
		fs.SetOutput(stderr)
		known := fs.String("known", "", "file listing workflow plugins that clash with pm")
		if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 {
			fmt.Fprintln(stderr, hookUsage)
			return exitBadInput
		}
		in := loadVoice()
		in.Herdr = os.Getenv("HERDR_ENV") == "1"
		repo, _ := os.Getwd()
		if cfg, err := config.Load(repo, ""); err == nil {
			repo = cfg.RepoRoot
			// Keep the agent record file out of git. A failure here is not
			// worth a word to the user, so it is dropped.
			_ = hook.EnsureGitignore(cfg.Root, ".agents.json")
		}
		in.Conflicts = hook.Conflicts(hook.EnabledPlugins(hook.ClaudeDir(), repo), hook.LoadKnown(*known))
		fmt.Fprint(stdout, hook.SessionStart(in))
		return exitOK
	default:
		fmt.Fprintln(stderr, hookUsage)
		return exitBadInput
	}
}

func loadVoice() hook.Input {
	v, exists, err := config.ResolveUser()
	return hook.Input{Voice: v, VoiceExists: exists, VoiceErr: err}
}

// hookRoot finds the planning folder the session state lives in. It says no
// outside a project, because a hook must not make planning folders in a
// folder that has none.
func hookRoot() (string, bool) {
	cwd, _ := os.Getwd()
	cfg, err := config.Load(cwd, "")
	if err != nil {
		return "", false
	}
	fi, err := os.Stat(cfg.Root)
	return cfg.Root, err == nil && fi.IsDir()
}

// sessionReminder is the one line a session that already brainstormed gets
// under the voice line, and nothing for any other session.
func sessionReminder(stdin io.Reader) string {
	ev, ok := hook.ParseEvent(stdin)
	if !ok {
		return ""
	}
	root, ok := hookRoot()
	if !ok {
		return ""
	}
	return hook.Reminder(root, ev.SessionID)
}
