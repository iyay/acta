// Package doctor checks that the user's acta setup still works: the binary
// that runs, the plugin the harness loads, the links under ~/.omp, the
// planning folder in the repo and the voice file. It only reads, except for
// Fix, which touches the repo and nothing else.
package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iyay/acta/internal/hook"
	"github.com/iyay/acta/internal/voice"
)

// Level is how bad one check found the setup.
type Level string

const (
	OK   Level = "ok"
	Warn Level = "warn"
	Fail Level = "fail"
)

// Result is one check and what the user can do about it. Fix is the exact
// command or step; it is empty when there is nothing to run.
type Result struct {
	Name, Msg, Fix string
	Level          Level
}

// Env is everything the checks read, so tests can point it at temp folders.
type Env struct {
	Home        string // user home; ~/.claude.json and ~/.omp live here
	ClaudeDir   string // hook.ClaudeDir()
	RepoRoot    string // "" outside a git repo
	ActaRoot    string // the .acta folder for RepoRoot
	Binary      string // os.Executable()
	Version     string // from runtime/debug build info
	KnownFile   string // workflow-plugins.txt; "" skips the conflicts list
	AutoCommit  bool   // false when .acta.yaml turns auto_commit off
	ConfigErr   error  // the config.Load error, when the config does not parse
	VoiceExists bool
	Voice       voice.Voice
}

// Run does every check in a fixed order, so two runs print the same lines in
// the same order and a user can diff them.
func Run(e Env) []Result {
	return []Result{
		checkBinary(e),
		checkHarness(e),
		checkStaleLinks(e),
		checkConflicts(e),
		checkRepo(e),
		checkAgentsView(e),
		checkSetup(e),
	}
}

// Fix writes what only the repo can fix: the .acta folder and its .gitignore
// line. It returns the paths it changed, so the caller can commit them. It
// never writes under Home: Claude Code and omp own those files.
func Fix(e Env) ([]string, error) {
	if e.RepoRoot == "" || e.ActaRoot == "" || !insideRepo(e.RepoRoot, e.ActaRoot) {
		return nil, nil
	}
	gi := filepath.Join(e.ActaRoot, ".gitignore")
	before, _ := os.ReadFile(gi)
	if err := os.MkdirAll(e.ActaRoot, 0o755); err != nil {
		return nil, err
	}
	if err := hook.EnsureGitignore(e.ActaRoot, ".agents.json"); err != nil {
		return nil, err
	}
	after, err := os.ReadFile(gi)
	if err != nil || string(after) == string(before) {
		return nil, nil
	}
	return []string{gi}, nil
}

// insideRepo says whether a path sits under the repo root. It walks the
// path relative to the root, so a sibling folder that merely starts with
// the same name, like /repo-other next to /repo, counts as outside.
func insideRepo(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)))
}

// Failed is true when a check is fail, which is what makes the command exit 1.
func Failed(rs []Result) bool {
	for _, r := range rs {
		if r.Level == Fail {
			return true
		}
	}
	return false
}

// Format prints one line per check and a second line with the fix for a
// check that is not ok. A line without a fix gets no second line: an empty
// command would be worse than none.
func Format(rs []Result) string {
	var b strings.Builder
	for _, r := range rs {
		fmt.Fprintf(&b, "%s %s: %s\n", r.Level, r.Name, r.Msg)
		if r.Level != OK && r.Fix != "" {
			fmt.Fprintf(&b, "fix: %s\n", r.Fix)
		}
	}
	return b.String()
}

// checkBinary names the acta that runs, so a stale hand-built binary in the
// path is visible in the report.
func checkBinary(e Env) Result {
	if e.Binary == "" {
		return Result{Name: "binary", Level: Fail, Msg: "no acta binary path known"}
	}
	msg := e.Binary
	if e.Version != "" {
		msg += " " + e.Version
	}
	return Result{Name: "binary", Level: OK, Msg: msg}
}

// checkHarness looks for the plugin where the harness looks for it: enabled
// in Claude Code, or linked into omp's node_modules.
func checkHarness(e Env) Result {
	r := Result{Name: "harness"}
	for _, p := range hook.EnabledPlugins(e.ClaudeDir, e.RepoRoot) {
		if p == "acta" || strings.HasPrefix(p, "acta@") {
			r.Level, r.Msg = OK, "acta plugin enabled in Claude Code"
			return r
		}
	}
	link := filepath.Join(e.Home, ".omp", "plugins", "node_modules", "acta")
	fi, err := os.Lstat(link)
	if err != nil {
		r.Level, r.Msg = Fail, "acta plugin is not installed in Claude Code or omp"
		r.Fix = `install it: see the acta plugin README "## Install"`
		return r
	}
	// A link whose target is gone is the case that hurts: omp still lists the
	// plugin and loads nothing. A plain folder is a copy and works.
	if fi.Mode()&os.ModeSymlink == 0 {
		r.Level, r.Msg = OK, "acta plugin is in "+link
		return r
	}
	target, _ := os.Readlink(link)
	if _, err := os.Stat(link); err != nil {
		r.Level = Fail
		r.Msg = "omp link acta points at " + target + ", which is gone"
		r.Fix = "omp plugin link " + target
		return r
	}
	r.Level, r.Msg = OK, "acta plugin linked in omp, "+target
	return r
}

// checkStaleLinks lists every omp plugin link whose target is gone. omp keeps
// them, so skills load as empty.
func checkStaleLinks(e Env) Result {
	r := Result{Name: "stale-links", Level: OK}
	dir := filepath.Join(e.Home, ".omp", "plugins", "node_modules")
	entries, err := os.ReadDir(dir)
	if err != nil {
		// No folder, or one we cannot read: nothing this check can report.
		r.Msg = "no omp plugin links"
		return r
	}
	var names, fixes []string
	for _, en := range entries {
		p := filepath.Join(dir, en.Name())
		fi, err := os.Lstat(p)
		if err != nil || fi.Mode()&os.ModeSymlink == 0 {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			continue
		}
		names = append(names, en.Name())
		fixes = append(fixes, "omp plugin unlink "+en.Name())
	}
	if len(names) == 0 {
		r.Msg = "no dead omp plugin links"
		return r
	}
	r.Level = Warn
	r.Msg = "link with no target: " + strings.Join(names, ", ")
	r.Fix = strings.Join(fixes, "; ")
	return r
}

// checkConflicts names the other workflow plugins that are enabled here, so
// the user can turn them off for this repo.
func checkConflicts(e Env) Result {
	r := Result{Name: "conflicts", Level: OK}
	if e.KnownFile == "" {
		r.Msg = "no list of clashing plugins given"
		return r
	}
	clashes := hook.Conflicts(hook.EnabledPlugins(e.ClaudeDir, e.RepoRoot), hook.LoadKnown(e.KnownFile))
	if len(clashes) == 0 {
		r.Msg = "no other workflow plugin enabled"
		return r
	}
	r.Level = Warn
	r.Msg = "another workflow plugin is enabled: " + strings.Join(clashes, ", ")
	r.Fix = "turn it off for this repo: add its enabledPlugins entry set to false in .claude/settings.local.json"
	return r
}

// checkRepo looks at the planning folder in the repo: it must exist, and it
// must ignore .agents.json so that file never lands in a commit. Outside a
// git repo there is nothing to check.
func checkRepo(e Env) Result {
	r := Result{Name: "repo"}
	// A config that does not parse is a broken setup, not a folder to skip,
	// so the parse error is the whole message.
	if e.ConfigErr != nil {
		r.Level, r.Msg = Fail, "cannot read the config: "+e.ConfigErr.Error()
		return r
	}
	if e.RepoRoot == "" {
		r.Level, r.Msg = OK, "not in a git repo, skipped"
		return r
	}
	// A root outside the repo is the one problem --fix cannot touch, so the
	// fix line names root instead of sending the user back to --fix.
	if !insideRepo(e.RepoRoot, e.ActaRoot) {
		r.Level, r.Msg = Fail, e.ActaRoot+" is outside the repo "+e.RepoRoot
		r.Fix = "fix root in .acta.yaml so it points inside the repo"
		return r
	}
	fi, err := os.Stat(e.ActaRoot)
	if err != nil || !fi.IsDir() {
		r.Level, r.Msg = Fail, "no "+e.ActaRoot+" folder"
		r.Fix = "acta doctor --fix"
		return r
	}
	raw, err := os.ReadFile(filepath.Join(e.ActaRoot, ".gitignore"))
	if err != nil {
		r.Level, r.Msg = Fail, "no .gitignore in "+e.ActaRoot
		r.Fix = "acta doctor --fix"
		return r
	}
	if !hasLine(string(raw), ".agents.json") {
		r.Level, r.Msg = Fail, ".gitignore in "+e.ActaRoot+" has no .agents.json line"
		r.Fix = "acta doctor --fix"
		return r
	}
	r.Level, r.Msg = OK, e.ActaRoot+" ignores .agents.json"
	return r
}

// hasLine matches the line on its own, so a commented-out one does not pass.
func hasLine(text, line string) bool {
	for _, l := range strings.Split(text, "\n") {
		if strings.TrimSpace(l) == line {
			return true
		}
	}
	return false
}

// checkAgentsView reads the one undocumented key in ~/.claude.json. Claude
// Code rewrites that file while it runs, so acta only reads it: unset counts
// as on, because that is the default.
func checkAgentsView(e Env) Result {
	r := Result{Name: "agents-view"}
	raw, err := os.ReadFile(filepath.Join(e.Home, ".claude.json"))
	if err != nil {
		r.Level, r.Msg = OK, "no ~/.claude.json, left arrow keeps its default"
		return r
	}
	var got struct {
		LeftArrowOpensAgents *bool `json:"leftArrowOpensAgents"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		r.Level = Warn
		r.Msg = "~/.claude.json could not be read: " + err.Error()
		return r
	}
	if got.LeftArrowOpensAgents == nil || *got.LeftArrowOpensAgents {
		r.Level, r.Msg = OK, "left arrow opens the agents view"
		return r
	}
	r.Level = Warn
	r.Msg = "left arrow does not open the agents view"
	r.Fix = "Open /config, turn on '← opens agents'"
	return r
}

// checkSetup says whether the first-run questions were ever answered.
func checkSetup(e Env) Result {
	r := Result{Name: "setup"}
	switch {
	case !e.VoiceExists:
		r.Level, r.Msg = Warn, "no voice file yet"
	case e.Voice.BuildExecutor == "":
		r.Level, r.Msg = Warn, "no default build executor"
	default:
		r.Level, r.Msg = OK, "voice and build executor are set"
	}
	if r.Level != OK {
		r.Fix = "/acta:setup"
	}
	return r
}
