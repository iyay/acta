// Package setup plans and applies the acta setup wizard: it writes the
// user config, installs the plugin into the harnesses it finds, and writes
// the acta block into the current repo. Every decision lives in Plan; the
// form and the runner only carry its actions out.
package setup

import (
	"path/filepath"

	"github.com/iyay/acta/internal/config"
)

// Action kinds. A config action writes the user config, an install action
// runs one harness install command, a print action shows a command for the
// user to run by hand, and a block action writes the acta block to a file.
const (
	ActionConfig  = "config"
	ActionInstall = "install"
	ActionPrint   = "print"
	ActionBlock   = "block"
)

// Env holds what the wizard found: whether stdio is a terminal, which
// harnesses sit on PATH, where the plugin folder is, which repo holds the
// working directory, which of CLAUDE.md and AGENTS.md exist there, and the
// current user config the form shows as defaults.
type Env struct {
	TTY                      bool
	Harnesses                []string
	PluginDir, RepoRoot      string
	HasClaudeMD, HasAgentsMD bool
	Current                  config.User
}

// Answers holds what the user said: the new user config and one install
// answer per harness found. The block paths come from Env, since the block
// is mandatory and never picked.
type Answers struct {
	User    config.User
	Install map[string]bool
}

// Action is one thing the runner carries out. Only the fields its Kind
// needs are set: config sets User, install sets Harness and Argv, print
// sets Harness and Path, block sets Path.
type Action struct {
	Kind          string
	Harness, Path string
	Argv          [][]string
	User          config.User
}

// Plan turns answers and findings into the action list. Without a TTY it
// returns a single print that points to acta config set, and nothing is
// ever written or installed. Inside a repo the block is always written:
// both files get one, a single file gets one, and a repo with neither gets
// a new CLAUDE.md. An AGENTS.md is never created.
func Plan(a Answers, e Env) []Action {
	if !e.TTY {
		return []Action{{Kind: ActionPrint,
			Path: "acta setup needs a terminal: nothing was written. " +
				"Set each value by hand with `acta config set`, " +
				"or re-run `acta setup` in a terminal."}}
	}
	out := []Action{{Kind: ActionConfig, User: a.User}}
	for _, h := range e.Harnesses {
		if !a.Install[h] {
			continue
		}
		if e.PluginDir == "" {
			hint, ok := installHint(h)
			if !ok {
				continue
			}
			out = append(out, Action{Kind: ActionPrint, Harness: h, Path: hint})
			continue
		}
		argv, ok := installArgv(h, e.PluginDir)
		if !ok {
			continue
		}
		out = append(out, Action{Kind: ActionInstall, Harness: h, Argv: argv})
	}
	return appendBlocks(out, e)
}

// appendBlocks adds the mandatory block actions: one per file that exists,
// or one that creates CLAUDE.md when neither does. Outside a repo it adds
// none.
func appendBlocks(out []Action, e Env) []Action {
	if e.RepoRoot == "" {
		return out
	}
	claude := filepath.Join(e.RepoRoot, "CLAUDE.md")
	if e.HasClaudeMD && e.HasAgentsMD {
		return append(out,
			Action{Kind: ActionBlock, Path: claude},
			Action{Kind: ActionBlock, Path: filepath.Join(e.RepoRoot, "AGENTS.md")})
	}
	if e.HasAgentsMD {
		return append(out, Action{Kind: ActionBlock, Path: filepath.Join(e.RepoRoot, "AGENTS.md")})
	}
	return append(out, Action{Kind: ActionBlock, Path: claude})
}

// installArgv is the harness install command for a plugin folder, copied
// from plugin/README.md. An unknown harness has no command, so false.
func installArgv(harness, dir string) ([][]string, bool) {
	switch harness {
	case "claude":
		return [][]string{
			{"claude", "plugin", "marketplace", "add", dir},
			{"claude", "plugin", "install", "acta@acta-local"},
		}, true
	case "omp":
		return [][]string{{"omp", "plugin", "link", dir}}, true
	}
	return nil, false
}

// installHint is the same command with <plugin dir> left for the user to
// fill, shown when the wizard runs without --plugin-dir.
func installHint(harness string) (string, bool) {
	switch harness {
	case "claude":
		return "claude plugin marketplace add <plugin dir>\n" +
			"claude plugin install acta@acta-local", true
	case "omp":
		return "omp plugin link <plugin dir>", true
	}
	return "", false
}
