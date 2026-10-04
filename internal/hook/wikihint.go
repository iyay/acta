package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/iyay/acta/internal/wiki"
)

const (
	// hintStateFile keeps, for each context, the ids of the pages it was shown.
	// A context is the main thread of a session, or one subagent of it.
	hintStateFile = "state/wiki-hints.json"

	// hintLockFile is the lock every reader-then-writer of hintStateFile takes.
	hintLockFile = "state/wiki-hints.lock"
)

// WikiHints gives one `wiki: <path>: <description>` line for each wiki page
// that covers a file this tool call touches, and nothing for a page this
// context has been shown already. root is the planning folder and repo is the
// repo root. A file tool gives its file path; a Bash command gives each of its
// words, since any word may be a path.
//
// Nothing here can fail the hook. A page that cannot load, a state file that
// cannot be read or saved, and a lock that cannot be had only mean fewer hints,
// or the same hint twice.
func WikiHints(root, repo string, ev ToolEvent) string {
	words := hintWords(ev)
	if len(words) == 0 {
		return ""
	}
	// Pages that cannot load are for `acta wiki check` to report.
	pages, _ := wiki.Load(root)
	if len(pages) == 0 {
		return ""
	}
	rel := repoRel(repo)
	hit := map[string]bool{}
	for _, w := range words {
		w = strings.Trim(w, `"'`)
		if w == "" {
			continue
		}
		if path, ok := rel(w); ok {
			for _, p := range wiki.Match(pages, path) {
				hit[p.ID] = true
			}
		}
	}
	// Page order, not word order, so the same call always reads the same.
	var covering []wiki.Page
	for _, p := range pages {
		if hit[p.ID] {
			covering = append(covering, p)
		}
	}
	if len(covering) == 0 {
		return ""
	}
	var lines []string
	for _, p := range claim(root, ev.SessionID+"/"+ev.AgentID, covering) {
		name, ok := rel(p.Path)
		if !ok {
			name = p.Path
		}
		// One line each: a description with a line break would cut the hint.
		lines = append(lines, "wiki: "+name+": "+strings.Join(strings.Fields(p.Description), " "))
	}
	return strings.Join(lines, "\n")
}

// hintOutput is the one JSON object Claude Code reads from a PreToolUse hook
// that exits 0. It puts additionalContext in front of the agent.
type hintOutput struct {
	Specific struct {
		Event   string `json:"hookEventName"`
		Context string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
}

// HintJSON wraps the lines WikiHints gave in that object.
func HintJSON(lines string) string {
	var out hintOutput
	out.Specific.Event, out.Specific.Context = "PreToolUse", lines
	// Two strings always marshal, so there is no error to handle.
	data, _ := json.Marshal(out)
	return string(data)
}

// hintWords lists what one tool call names as a file: the path of a file tool,
// or each word of a Bash command.
func hintWords(ev ToolEvent) []string {
	if ev.ToolInput.FilePath != "" {
		return []string{ev.ToolInput.FilePath}
	}
	return goTestCut(ev.ToolInput.Command, goTestSep+goTestSpace)
}

// repoRel gives a function that turns a path an agent wrote into a slash path
// from the repo root. It says no for a path outside the repo. A path with no
// start is read from the folder the hook runs in, since that is where a shell
// command runs.
//
// One folder can have two names when a link is in the way. macOS keeps /tmp and
// /var behind links, so the agent, the shell and git may not agree on the name.
// So a path is tried as written first, and then with its links followed.
func repoRel(repo string) func(string) (string, bool) {
	realRepo := resolve(repo)
	cwd, _ := os.Getwd()
	realCwd := resolve(cwd)
	return func(p string) (string, bool) {
		if !filepath.IsAbs(p) {
			return under(realRepo, filepath.Join(realCwd, p))
		}
		p = filepath.Clean(p)
		if rel, ok := under(repo, p); ok {
			return rel, true
		}
		return under(realRepo, resolve(p))
	}
}

// under says where p sits inside base, as a slash path from base. It says no
// for a path outside base.
func under(base, p string) (string, bool) {
	rel, err := filepath.Rel(base, p)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return filepath.ToSlash(rel), true
}

// resolve follows the links in the part of p that exists. A Write may name a
// file, or a whole folder, that is not there yet, so the nearest folder that is
// there gets resolved and the rest of the name is put back.
func resolve(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	if dir := filepath.Dir(p); dir != p {
		return filepath.Join(resolve(dir), filepath.Base(p))
	}
	return p
}

// claim marks the pages as shown to one context, and gives back only the ones
// it had not been shown before.
func claim(root, key string, pages []wiki.Page) []wiki.Page {
	// Most calls end here: the pages are known already, and nothing is locked or
	// written. The file is replaced in one move, so this read never sees half a
	// file, and a read that is wrong only sends the call on to the check below.
	if len(unseen(pages, readHints(root)[key])) == 0 {
		return nil
	}
	// Dropped on purpose, like the other state writes: a root that cannot be
	// written to must not stop a hint.
	_ = EnsureGitignore(root, "state/")
	// Two tool calls of one turn run their hooks at the same moment. Without the
	// lock both read the same list and both show the page.
	if unlock, ok := lockHints(root); ok {
		defer unlock()
	}
	shown := readHints(root)
	fresh := unseen(pages, shown[key])
	if len(fresh) == 0 {
		return nil
	}
	for _, p := range fresh {
		shown[key] = append(shown[key], p.ID)
	}
	// A hook that cannot save what it showed still shows it. The hint then comes
	// back on every touch, which is how the user finds out the state is broken.
	_ = writeHints(root, shown)
	return fresh
}

// ResetHints makes one session hear its pages again, in the main thread and in
// every subagent. A clear or a compaction takes the old hints out of the
// context, so the list of what was shown is no longer true. It takes the lock
// the hints take, so a hint saved at the same moment cannot bring a page back.
//
// A state that cannot be saved comes back as an error, and the caller drops it:
// a reset that fails only leaves the pages quiet, and must never fail a session.
func ResetHints(root, sessionID string) error {
	// A key is a session id, a slash and an agent id, and a hint needs a session
	// id. So an empty one matches no key and drops nothing.
	prefix := sessionID + "/"
	// Most session starts follow no hint at all. Look first, so they lock and
	// write nothing.
	if shown := readHints(root); len(withoutSession(shown, prefix)) == len(shown) {
		return nil
	}
	// A lock that cannot be had does not stop the reset: a reset that does
	// nothing would leave every page quiet, which is what it is here to undo.
	if unlock, ok := lockHints(root); ok {
		defer unlock()
	}
	// The look above took no lock, so the file is read again now that it is held.
	return writeHints(root, withoutSession(readHints(root), prefix))
}

// withoutSession gives the shown pages minus every context whose key starts with
// prefix, which is one session and all of its subagents.
func withoutSession(shown map[string][]string, prefix string) map[string][]string {
	kept := make(map[string][]string, len(shown))
	for key, ids := range shown {
		if !strings.HasPrefix(key, prefix) {
			kept[key] = ids
		}
	}
	return kept
}

// unseen keeps the pages whose id is not in ids.
func unseen(pages []wiki.Page, ids []string) []wiki.Page {
	var out []wiki.Page
	for _, p := range pages {
		if !slices.Contains(ids, p.ID) {
			out = append(out, p)
		}
	}
	return out
}

// readHints maps a context to the ids of the pages it was shown. A file that is
// missing, unreadable or broken gives an empty map, so a state file that went
// wrong looks like a context that was shown nothing.
func readHints(root string) map[string][]string {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(hintStateFile)))
	if err != nil {
		return map[string][]string{}
	}
	var shown map[string][]string
	if err := json.Unmarshal(data, &shown); err != nil || shown == nil {
		return map[string][]string{}
	}
	return shown
}

func writeHints(root string, shown map[string][]string) error {
	data, err := json.Marshal(shown)
	if err != nil {
		return err
	}
	return writeAtomic(filepath.Join(root, filepath.FromSlash(hintStateFile)), data)
}

// writeAtomic writes the data next to path and then moves it over path, so a
// hook that dies mid-write cannot leave half a file behind for the next hook to
// choke on.
func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+"-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// lockHints takes the lock on the shown pages and gives the call that lets it
// go. It waits half a second at most and then says no, because a hook must never
// hang a session. The kernel drops the lock when the process ends, even in a
// crash, so no lock is ever left behind.
func lockHints(root string) (func(), bool) {
	path := filepath.Join(root, filepath.FromSlash(hintLockFile))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, false
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, false
	}
	for range 250 {
		if syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) == nil {
			return func() {
				syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
				f.Close()
			}, true
		}
		time.Sleep(2 * time.Millisecond)
	}
	f.Close()
	return nil, false
}
