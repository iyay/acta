package cli

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/hook"
	"github.com/iyay/acta/internal/voice"
)

const hookUsage = "usage: acta hook session-start [--known <file>] | acta hook prompt"

// cmdHook prints hook text. After its arguments parse it always exits 0:
// a hook that fails would get in the way of the user's session.
func cmdHook(args []string, stdout, stderr io.Writer) int {
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
	v, exists, err := voice.Resolve()
	return hook.Input{Voice: v, VoiceExists: exists, VoiceErr: err}
}
