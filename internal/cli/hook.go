package cli

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/hook"
	"github.com/iyay/acta/internal/telemetry"
	"github.com/iyay/acta/internal/wiki"
	"github.com/iyay/acta/plugin"
)

var hookUsage = usageLine("hook")

// exitBlock is the code a tool hook uses to tell the agent the command is
// stopped. It is the only non-zero code an acta hook ever returns.
const exitBlock = 2

// wikiHints is a variable so a test can make the hint path panic.
var wikiHints = hook.WikiHints

// cmdHook prints hook text. Every hook exits 0 whatever stdin holds, because a
// hook that fails would get in the way of the user's session. The one
// exception is pre-tool stopping a bare go test or a destructive command,
// which exits 2 with the reason the agent has to read.
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
		return exitOK
	case "pre-tool", "post-tool":
		if len(args) != 1 {
			fmt.Fprintln(stderr, hookUsage)
			return exitBadInput
		}
		// Post-tool has nothing left to do. Installed hook files still call
		// it, so it stays and stays quiet.
		if args[0] == "post-tool" {
			return exitOK
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			return exitOK
		}
		ev, ok := hook.ParseEvent(bytes.NewReader(data))
		if !ok {
			return exitOK
		}
		// A Codex patch carries file content in its command field. That text
		// never runs, so the shell guard below must not read it.
		shell := ev.ToolName != "apply_patch"
		// The go test block reads the folder the agent works in, not the
		// planning root, so a repo with no planning root still gets it.
		cwd, _ := os.Getwd()
		if block, msg := hook.GoTestBlock(cwd, ev.ToolInput.Command); shell && block {
			fmt.Fprintln(stderr, msg)
			return exitBlock
		}
		// The Bash tool keeps a cd from one call to the next, so it can work
		// in another folder than this process. The event names that folder.
		// The parse above already passed, so a bad field here just leaves
		// it empty.
		var where struct {
			Cwd string `json:"cwd"`
		}
		_ = json.Unmarshal(data, &where)
		if where.Cwd == "" {
			where.Cwd = cwd
		}
		if block, msg := hook.DestructiveBlock(where.Cwd, ev.ToolInput.Command); shell && block {
			fmt.Fprintln(stderr, msg)
			return exitBlock
		}
		// A hint is only an extra line for the agent, so it comes after the
		// blocks and never changes the exit code. A hook that runs in no
		// checkout has no repository to trust, so it gives none.
		if lines := hintLines(hookConfig(), ev); lines != "" {
			fmt.Fprintln(stdout, hook.HintJSON(lines))
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
		in.Codex = hook.IsCodex()
		in.Update, in.Current = newerRelease(), "v"+plugin.Version()
		in.Notice = sessionTelemetry()
		repo, _ := os.Getwd()
		if cfg, err := config.Load(repo, ""); err == nil {
			repo = cfg.RepoRoot
			// Keep the agent record file out of git. A failure here is not
			// worth a word to the user, so it is dropped.
			_ = hook.EnsureGitignore(cfg.Root, ".agents.json")
			// Pages that cannot load are for `acta wiki check` to report.
			pages, _ := wiki.Load(cfg.Root)
			in.WikiPages = len(pages)
			// Plans with work going on, three lines each. Errors are
			// dropped like the ones above: a session that cannot read the
			// board must still start, just without the summary.
			for _, r := range board.Running(cfg) {
				in.Running = append(in.Running, strings.Join(r.Short(), "\n"))
			}
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

// sessionTelemetry counts the harness this session runs in, starts the
// hourly flush, and gives the notice line the first time on this machine.
// The notice goes into the session text and never to stderr, since nobody
// reads hook stderr. Nothing here may stop a session, so errors are dropped.
func sessionTelemetry() string {
	s := telemetry.Default()
	if !s.Enabled() {
		return ""
	}
	_ = s.Add("sessions", sessionHarness())
	startFlush(s)
	if first, err := s.MarkNoticed(); err == nil && first {
		return telemetryNotice
	}
	return ""
}

// sessionHarness names the agent tool this hook runs under. Only the omp
// extension sets ACTA_HARNESS, so any other value counts as other.
func sessionHarness() string {
	switch h := os.Getenv("ACTA_HARNESS"); {
	case h == "omp":
		return "omp"
	case h != "":
		return "other"
	case hook.IsCodex():
		return "codex"
	}
	return "claude"
}

// hintLines asks for the wiki hints of one tool call. A hint is not worth a
// failed call, and a Go panic exits 2, which is the code that blocks the tool.
// So a panic is dropped here, and the call goes on as if it had no hint.
func hintLines(cfg config.Config, ev hook.ToolEvent) (lines string) {
	defer func() {
		if recover() != nil {
			lines = ""
		}
	}()
	// The checkout the hook runs in is known already, so it is not looked up
	// again. A folder that is no checkout has none, and then no hint comes.
	root, repo := "", ""
	if cfg.IsGit {
		root, repo = cfg.Root, cfg.RepoRoot
	}
	return wikiHints(root, repo, ev)
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
		// The user's own value goes in, before the merge, so each override
		// shows what the repo file replaced. A bad repo file gives no list.
		if in.RepoErr == nil {
			in.Overrides, in.RepoErr = config.Overrides(v, cfg.RepoRoot)
		}
	}
	return in
}

// hookConfig finds the planning folder and the repo root for the wiki hints.
// Outside a project it gives an empty config, so no hint comes.
func hookConfig() config.Config {
	cwd, _ := os.Getwd()
	cfg, err := config.Load(cwd, "")
	if err != nil {
		return config.Config{}
	}
	return cfg
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
