package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/hook"
	"github.com/iyay/acta/internal/wiki"
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
		// A repo with no planning root is still a repo with a scripts/test,
		// so the go test check is asked even then. Only the brainstorm state
		// and the wiki hints need the root.
		cfg, hasRoot := hookConfig()
		root := cfg.Root
		if args[0] == "post-tool" {
			if !hasRoot {
				return exitOK
			}
			// Both writes are dropped on purpose: a session that cannot
			// remember its brainstorm must not be stopped for it.
			_ = hook.EnsureGitignore(root, "state/")
			_ = hook.RecordBrainstorm(root, ev)
			return exitOK
		}
		if hasRoot {
			if block, msg := hook.PreTool(root, ev); block {
				fmt.Fprintln(stderr, msg)
				return exitBlock
			}
		}
		// The go test block reads the folder the agent works in, not the
		// planning root, so it comes after the brainstorm one.
		cwd, _ := os.Getwd()
		if block, msg := hook.GoTestBlock(cwd, ev.ToolInput.Command); block {
			fmt.Fprintln(stderr, msg)
			return exitBlock
		}
		// A hint is only an extra line for the agent, so it comes after the
		// blocks and never changes the exit code.
		if hasRoot {
			if lines := hook.WikiHints(root, cfg.RepoRoot, ev); lines != "" {
				fmt.Fprintln(stdout, hook.HintJSON(lines))
			}
		}
		return exitOK
	case "session-start":
		fs := flag.NewFlagSet("hook session-start", flag.ContinueOnError)
		fs.SetOutput(stderr)
		known := fs.String("known", "", "file listing plugins that overlap acta")
		if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 {
			fmt.Fprintln(stderr, hookUsage)
			return exitBadInput
		}
		sessionID, source := readSessionStart(stdin)
		in := loadVoice()
		in.Herdr = os.Getenv("HERDR_ENV") == "1"
		repo, _ := os.Getwd()
		if cfg, err := config.Load(repo, ""); err == nil {
			repo = cfg.RepoRoot
			// Keep the agent record file out of git. A failure here is not
			// worth a word to the user, so it is dropped.
			_ = hook.EnsureGitignore(cfg.Root, ".agents.json")
			// Pages that cannot load are for `acta wiki check` to report.
			pages, _ := wiki.Load(cfg.Root)
			in.WikiPages = len(pages)
			// After a clear or a compaction the old hints are gone from the
			// context, so the session must hear its pages again. The error is
			// dropped on purpose: a state that cannot be saved must not stop
			// a session from starting.
			if source == "clear" || source == "compact" {
				_ = hook.ResetHints(cfg.Root, sessionID)
			}
		}
		in.Conflicts = hook.Conflicts(hook.EnabledPlugins(hook.ClaudeDir(), repo), hook.LoadKnown(*known))
		fmt.Fprint(stdout, hook.SessionStart(in))
		return exitOK
	default:
		fmt.Fprintln(stderr, hookUsage)
		return exitBadInput
	}
}

// loadVoice reads the user config and lays this repo's .acta.yaml over it.
// A bad repo file keeps the user's own values and is named in RepoErr, so
// a repo can never switch the chat language, and the session still sees
// what to fix.
func loadVoice() hook.Input {
	v, exists, err := config.ResolveUser()
	in := hook.Input{Voice: v, VoiceExists: exists, VoiceErr: err}
	if err != nil {
		return in
	}
	cwd, _ := os.Getwd()
	if cfg, lerr := config.Load(cwd, ""); lerr == nil {
		in.Voice, _, in.RepoErr = config.MergeRepo(v, cfg.RepoRoot)
	}
	return in
}

// hookRoot finds the planning folder the session state lives in. It says no
// outside a project, because a hook must not make planning folders in a
// folder that has none.
func hookRoot() (string, bool) {
	cfg, ok := hookConfig()
	return cfg.Root, ok
}

// hookConfig is hookRoot with the whole config, for the hook that also needs
// the repo root. It loads the config once, because every load asks git.
func hookConfig() (config.Config, bool) {
	cwd, _ := os.Getwd()
	cfg, err := config.Load(cwd, "")
	if err != nil {
		return config.Config{}, false
	}
	fi, err := os.Stat(cfg.Root)
	return cfg, err == nil && fi.IsDir()
}

// readSessionStart reads the two fields of the SessionStart payload this hook
// uses: the session, and why it started. Stdin that is empty, broken or not what
// was expected gives two empty strings, because a session must start whatever
// the hook is handed.
func readSessionStart(stdin io.Reader) (sessionID, source string) {
	data, err := io.ReadAll(stdin)
	if err != nil {
		return "", ""
	}
	var ev struct {
		SessionID string `json:"session_id"`
		Source    string `json:"source"`
	}
	if json.Unmarshal(data, &ev) != nil {
		return "", ""
	}
	return ev.SessionID, ev.Source
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
