package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/iyay/acta/internal/config"
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
// context has been shown already. A file tool gives its file path; a Bash
// command gives each of its words, since any word may be a path.
//
// Only the session's own repository is trusted: the checkout the hook runs in,
// and the worktrees of that same repository. A Claude Code build keeps its
// session in the main checkout while the work is done in a worktree, and the
// pages of the branch are only in the worktree, so a file of a worktree gets the
// pages of that worktree. A file of any other repository gets nothing, because
// its pages and its settings are not the user's to trust, and the hook must
// write nothing there. A repository inside the session checkout that is not one
// of its worktrees, like a vendored clone, counts as part of the session
// checkout.
//
// The list of what was shown is kept in the session checkout only, so one reset
// covers every worktree. root is the planning folder and repo the repo root of
// the checkout the hook runs in. Both are empty when the hook runs in none, and
// then there are no hints at all.
//
// A path that is not a full one is read from the folder the hook runs in, or
// from the folder of the last `cd` word before it in the same command.
//
// Nothing here can fail the hook. A page that cannot load, a state file that
// cannot be read or saved, and a lock that cannot be had only mean fewer hints,
// or the same hint twice.
func WikiHints(root, repo string, ev ToolEvent) string {
	if root == "" || repo == "" {
		return ""
	}
	cwd, _ := os.Getwd()
	files := hintFiles(ev, cwd)
	if len(files) == 0 {
		return ""
	}
	look := &lookup{root: root, repo: repo, common: commonDir(repo), byTop: map[string]*checkout{}, byDir: map[string]*checkout{}}
	// The checkouts that have a hint, in the order the words reached them.
	var touched []*checkout
	for _, file := range files {
		c := look.of(file)
		if c == nil {
			continue
		}
		rel, ok := c.rel(file)
		if !ok {
			continue
		}
		covering := wiki.Match(c.pages, rel)
		if len(covering) > 0 && !slices.Contains(touched, c) {
			touched = append(touched, c)
		}
		for _, p := range covering {
			c.hit[p.ID] = true
		}
	}
	var lines []string
	for _, c := range touched {
		lines = append(lines, c.lines(root, ev.SessionID+"/"+ev.AgentID)...)
	}
	return strings.Join(lines, "\n")
}

// checkout is one git checkout a tool call touches, with the pages it holds.
type checkout struct {
	repo, root string
	// real is repo with its links followed.
	real  string
	pages []wiki.Page
	// prefix names the pages of this checkout in the shown list. The session
	// checkout has none. A worktree has its own path, since its pages can share
	// an id with the pages of the session checkout.
	prefix string
	// hit holds the ids of the pages that cover a file the call touched.
	hit map[string]bool
}

// rel says where a file sits in the checkout, as a slash path from its repo
// root. It says no for a file outside it.
//
// One folder can have two names when a link is in the way. macOS keeps /tmp and
// /var behind links, so the agent, the shell and git may not agree on the name.
// So a path is tried as written first, and then with its links followed.
func (c *checkout) rel(file string) (string, bool) {
	if rel, ok := under(c.repo, file); ok {
		return rel, true
	}
	return under(c.real, resolve(file))
}

// lines gives the hint line for each page that covers a touched file and that
// this context has not been shown. The page file is named from the repo root,
// so the agent can open it from its checkout. What was shown is kept in the
// session root.
func (c *checkout) lines(sessionRoot, key string) []string {
	// Page order, not word order, so the same call always reads the same.
	var covering []wiki.Page
	for _, p := range c.pages {
		if c.hit[p.ID] {
			covering = append(covering, p)
		}
	}
	var out []string
	for _, p := range claim(sessionRoot, key, c.prefix, covering) {
		name, ok := c.rel(p.Path)
		if !ok {
			name = p.Path
		}
		// One line each: a description with a line break would cut the hint.
		out = append(out, "wiki: "+name+": "+strings.Join(strings.Fields(p.Description), " "))
	}
	return out
}

// lookup finds the checkout of a file. Asking git for one costs a few
// milliseconds, so a call asks once for each checkout it meets. A command can
// name thousands of words, so each folder is walked once too.
type lookup struct {
	// root and repo are the checkout the hook runs in, which is known already.
	root, repo string
	// common is the git folder that checkout shares with its worktrees.
	common string
	// own is the checkout the hook runs in, once it is read.
	own *checkout
	// byTop holds each checkout by the folder its .git was found in. A checkout
	// with no usable pages is nil: it has nothing to hint.
	byTop map[string]*checkout
	// byDir holds the checkout of each folder that was asked about.
	byDir map[string]*checkout
}

// of gives the checkout that holds file, or nil when no git checkout holds it or
// it has no pages. The file can be a path that is not there yet: the walk up for
// the checkout starts at its folder, and a folder that is missing has no .git.
func (l *lookup) of(file string) *checkout {
	dir := filepath.Dir(file)
	c, done := l.byDir[dir]
	if !done {
		c = l.find(dir)
		l.byDir[dir] = c
	}
	return c
}

// find walks up from dir to the checkout that holds it.
func (l *lookup) find(dir string) *checkout {
	top, ok := gitTopDir(dir)
	if !ok {
		return nil
	}
	c, done := l.byTop[top]
	if !done {
		c = l.open(top)
		l.byTop[top] = c
	}
	return c
}

// open reads the pages of the checkout that has its top at top. The checkout the
// hook runs in comes from the caller. A worktree of the same repository is asked
// of config, which knows where its planning folder is. Any other repository has
// no pages: a nested one is read as part of the checkout the hook runs in, and an
// unrelated one is not read at all.
func (l *lookup) open(top string) *checkout {
	if resolve(top) == resolve(l.repo) {
		return l.session()
	}
	if l.common != "" && commonDir(top) == l.common {
		cfg, err := config.Load(top, "")
		if err != nil || !cfg.IsGit {
			return nil
		}
		return load(cfg.Root, cfg.RepoRoot, resolve(top)+"|")
	}
	if _, ok := under(l.repo, top); ok {
		return l.session()
	}
	if _, ok := under(resolve(l.repo), resolve(top)); ok {
		return l.session()
	}
	return nil
}

// session gives the checkout the hook runs in, read once.
func (l *lookup) session() *checkout {
	if l.own == nil {
		l.own = load(l.root, l.repo, "")
	}
	return l.own
}

// load reads the pages of one checkout. A checkout with none is nil.
func load(root, repo, prefix string) *checkout {
	// Pages that cannot load are for `acta wiki check` to report.
	pages, _ := wiki.Load(root)
	if len(pages) == 0 {
		return nil
	}
	return &checkout{repo: repo, root: root, real: resolve(repo), prefix: prefix, pages: pages, hit: map[string]bool{}}
}

// commonDir gives the git folder a checkout shares with its worktrees, read from
// its .git and not from git, so no process is started. A .git folder is the
// common folder. A .git file names the folder of its worktree, and that folder
// names the common one in a commondir file, as git writes it. A folder with no
// commondir file, like a submodule one, has no common folder here. A worktree
// folder that does not point back at this .git is no worktree of it, and gets
// none: a repository must not be able to claim a folder that is not its own.
// Nothing here can be read gives an empty string.
func commonDir(top string) string {
	dotGit := filepath.Join(top, ".git")
	fi, err := os.Lstat(dotGit)
	if err != nil {
		return ""
	}
	if fi.IsDir() {
		return resolve(dotGit)
	}
	data, err := os.ReadFile(dotGit)
	if err != nil {
		return ""
	}
	gitdir, ok := strings.CutPrefix(strings.TrimSpace(string(data)), "gitdir:")
	if !ok {
		return ""
	}
	gitdir = fromDir(top, strings.TrimSpace(gitdir))
	rel, err := os.ReadFile(filepath.Join(gitdir, "commondir"))
	if err != nil {
		// No commondir file means no linked worktree. Anyone can write a .git
		// file that names the session's .git, so it must claim nothing.
		return ""
	}
	back, err := os.ReadFile(filepath.Join(gitdir, "gitdir"))
	if err != nil || resolve(strings.TrimSpace(string(back))) != resolve(dotGit) {
		return ""
	}
	return resolve(fromDir(gitdir, strings.TrimSpace(string(rel))))
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

// hintFiles lists what one tool call names as a file, each as a full path: the
// path of a file tool, or each word of a Bash command. A path that is not a full
// one is read from cwd, since that is where a shell command runs, or from the
// folder of the last `cd` word before it. A `cd` inside parentheses is read as a
// `cd` of the rest of the command. The shell would not, but it costs at most one
// hint too many or too few.
func hintFiles(ev ToolEvent, cwd string) []string {
	if p := ev.ToolInput.FilePath; p != "" {
		return hintWord(nil, cwd, p)
	}
	var files []string
	dir := cwd
	for _, piece := range goTestCut(ev.ToolInput.Command, goTestSep+"\n") {
		words := goTestCut(piece, goTestSpace)
		for _, w := range words {
			files = hintWord(files, dir, w)
		}
		// The next piece runs in the folder this one moved to. A folder that is
		// not a plain name, like $WT or ~, leads to a folder that is not there,
		// so the words after it get no hint, which is better than a wrong one.
		if len(words) > 1 && words[0] == "cd" {
			dir = fromDir(dir, strings.Trim(words[1], `"'`))
		}
	}
	return files
}

// hintWord adds one word to files as a full path, unless the word is empty once
// its quotes are cut.
func hintWord(files []string, dir, w string) []string {
	if w = strings.Trim(w, `"'`); w == "" {
		return files
	}
	return append(files, fromDir(dir, w))
}

// fromDir makes a full path of a word: a full one stays, and any other is read
// from dir.
func fromDir(dir, w string) string {
	if filepath.IsAbs(w) {
		return filepath.Clean(w)
	}
	return filepath.Join(dir, w)
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
//
// The ids are kept as prefix+id, so the pages of two checkouts that share an id
// are two pages.
func claim(root, key, prefix string, pages []wiki.Page) []wiki.Page {
	// Most calls end here: the pages are known already, and nothing is locked or
	// written. The file is replaced in one move, so this read never sees half a
	// file, and a read that is wrong only sends the call on to the check below.
	if len(unseen(pages, prefix, readHints(root)[key])) == 0 {
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
	fresh := unseen(pages, prefix, shown[key])
	if len(fresh) == 0 {
		return nil
	}
	for _, p := range fresh {
		shown[key] = append(shown[key], prefix+p.ID)
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

// unseen keeps the pages whose id, with the prefix in front, is not in ids.
func unseen(pages []wiki.Page, prefix string, ids []string) []wiki.Page {
	var out []wiki.Page
	for _, p := range pages {
		if !slices.Contains(ids, prefix+p.ID) {
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
