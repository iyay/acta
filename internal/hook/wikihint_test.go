package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

const (
	tuiLine    = "wiki: .acta/wiki/tui-wrap.md: Wrap can overflow after a dash; use Hardwrap(Wordwrap(...))"
	parserLine = "wiki: .acta/wiki/parser.md: The parser keeps the quotes"
)

// hintPage writes one wiki page under <root>/wiki.
func hintPage(t *testing.T, root, id, desc string, paths ...string) {
	t.Helper()
	quoted := make([]string, len(paths))
	for i, p := range paths {
		quoted[i] = fmt.Sprintf("%q", p)
	}
	src := fmt.Sprintf("---\ntype: Gotcha\ntitle: %s\ndescription: %q\npaths: [%s]\ntimestamp: 2026-10-04T10:00:00Z\n---\nbody words\n",
		id, desc, strings.Join(quoted, ", "))
	file := filepath.Join(root, "wiki", filepath.FromSlash(id)+".md")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}

// hintWiki makes a repo with three pages: one for a folder, one for a file and
// a folder, and one that covers nothing. It gives back the planning root and
// the repo root.
func hintWiki(t *testing.T) (root, repo string) {
	t.Helper()
	repo = t.TempDir()
	root = filepath.Join(repo, ".acta")
	hintPage(t, root, "tui-wrap", "Wrap can overflow after a dash; use Hardwrap(Wordwrap(...))", "internal/tui/")
	hintPage(t, root, "parser", "The parser keeps the quotes", "src/parser.go", "docs/")
	hintPage(t, root, "loose", "Covers no file")
	for _, d := range []string{"internal/tui", "src", "docs"} {
		if err := os.MkdirAll(filepath.Join(repo, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return root, repo
}

func fileEv(tool, session, agent, path string) ToolEvent {
	var e ToolEvent
	e.ToolName, e.SessionID, e.AgentID, e.ToolInput.FilePath = tool, session, agent, path
	return e
}

func bashEv(session, agent, command string) ToolEvent {
	var e ToolEvent
	e.ToolName, e.SessionID, e.AgentID, e.ToolInput.Command = "Bash", session, agent, command
	return e
}

// Every file tool names its file in tool_input.file_path.
func TestWikiHintsFileTools(t *testing.T) {
	for _, tool := range []string{"Read", "Edit", "Write", "MultiEdit"} {
		t.Run(tool, func(t *testing.T) {
			root, repo := hintWiki(t)
			got := WikiHints(root, repo, fileEv(tool, "s1", "", filepath.Join(repo, "internal/tui/model.go")))
			if got != tuiLine {
				t.Errorf("hint = %q, want %q", got, tuiLine)
			}
		})
	}

	cases := []struct{ name, path, want string }{
		// A Write can name a file, and even a folder, that is not there yet.
		{"a file that is not there yet", "internal/tui/new/deep/file.go", tuiLine},
		{"a file entry", "src/parser.go", parserLine},
		{"a file no page covers", "README.md", ""},
		{"a file next to a file entry", "src/parser.go.bak", ""},
		{"a sibling folder", "internal/tuix/model.go", ""},
	}
	for _, c := range cases {
		root, repo := hintWiki(t)
		if got := WikiHints(root, repo, fileEv("Read", "s1", "", filepath.Join(repo, c.path))); got != c.want {
			t.Errorf("%s: hint = %q, want %q", c.name, got, c.want)
		}
	}

	// A tool that gives a relative path means the folder the agent works in.
	root, repo := hintWiki(t)
	t.Chdir(repo)
	if got := WikiHints(root, repo, fileEv("Read", "s1", "", "internal/tui/model.go")); got != tuiLine {
		t.Errorf("relative path: hint = %q, want %q", got, tuiLine)
	}
}

// A Bash command names files as words, in the folder the command runs in.
func TestWikiHintsBashWords(t *testing.T) {
	both := parserLine + "\n" + tuiLine
	cases := []struct{ name, dir, cmd, want string }{
		{"a path word", "", "cat src/parser.go", parserLine},
		{"a go pattern", "", "go vet ./internal/tui/...", tuiLine},
		{"a folder word without its slash", "", "scripts/test ./internal/tui -run TestX", tuiLine},
		{"words after a chain", "", "cd /tmp && cat docs/a.md; echo done", parserLine},
		{"words in and after a subshell", "", "(cat src/parser.go); cat internal/tui/a.go", both},
		{"a quoted word", "", `cat "internal/tui/a b.go"`, tuiLine},
		{"two pages in one command", "", "diff internal/tui/a.go src/parser.go", both},
		{"one page twice in one command", "", "cat internal/tui/a.go internal/tui/b.go", tuiLine},
		{"a word relative to the folder the command runs in", "internal/tui", "cat model.go", tuiLine},
		{"a dot dot word from a subfolder", "src", "cat ../internal/tui/model.go", tuiLine},
		{"no path word", "", "git status", ""},
		{"a sibling folder", "", "cat internal/tuix/a.go", ""},
		{"an absolute word outside the repo", "", "cat /etc/hosts", ""},
		{"a word that climbs out of the repo", "", "cat ../internal/tui/a.go", ""},
		{"empty command", "", "", ""},
		{"blanks only", "", "  \t ", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root, repo := hintWiki(t)
			t.Chdir(filepath.Join(repo, c.dir))
			if got := WikiHints(root, repo, bashEv("s1", "", c.cmd)); got != c.want {
				t.Errorf("hint for %q = %q, want %q", c.cmd, got, c.want)
			}
		})
	}
}

// A page shows once for each context. The main thread and every subagent of a
// session are contexts of their own, so a subagent that never saw the hint gets
// it, and the main thread does not get it twice.
func TestWikiHintsOncePerContext(t *testing.T) {
	root, repo := hintWiki(t)
	t.Chdir(repo)
	tui := filepath.Join(repo, "internal/tui/a.go")
	steps := []struct {
		name string
		ev   ToolEvent
		want string
	}{
		{"first touch", fileEv("Read", "s1", "", tui), tuiLine},
		{"same file again", fileEv("Read", "s1", "", tui), ""},
		{"another file under the same page", fileEv("Edit", "s1", "", filepath.Join(repo, "internal/tui/b.go")), ""},
		{"the same page through bash", bashEv("s1", "", "cat internal/tui/c.go"), ""},
		{"another page in the same context", fileEv("Read", "s1", "", filepath.Join(repo, "src/parser.go")), parserLine},
		{"a subagent is a context of its own", fileEv("Read", "s1", "agent-1", tui), tuiLine},
		{"that subagent again", fileEv("Read", "s1", "agent-1", tui), ""},
		{"a second subagent", fileEv("Read", "s1", "agent-2", tui), tuiLine},
		{"another session", fileEv("Read", "s2", "", tui), tuiLine},
		{"the main thread is still counted once", fileEv("Read", "s1", "", tui), ""},
	}
	for _, s := range steps {
		if got := WikiHints(root, repo, s.ev); got != s.want {
			t.Errorf("%s: hint = %q, want %q", s.name, got, s.want)
		}
	}
}

// Two tool calls of one turn run their hooks at the same moment. Each reads the
// list of shown pages before the other has written it, unless something keeps
// them in line, and then the page shows more than once.
func TestWikiHintsParallelCallsShowAPageOnce(t *testing.T) {
	root, repo := hintWiki(t)
	tui := filepath.Join(repo, "internal/tui/a.go")
	got := make([]string, 8)
	var wg sync.WaitGroup
	for i := range got {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got[i] = WikiHints(root, repo, fileEv("Read", "s1", "", tui))
		}()
	}
	wg.Wait()
	shown := 0
	for _, g := range got {
		switch g {
		case tuiLine:
			shown++
		case "":
		default:
			t.Errorf("unexpected hint %q", g)
		}
	}
	if shown != 1 {
		t.Errorf("the page showed %d times in %d parallel calls, want once", shown, len(got))
	}
}

func TestWikiHintsIgnoreFilesOutsideTheRepo(t *testing.T) {
	root, repo := hintWiki(t)
	elsewhere := t.TempDir()
	t.Chdir(repo)
	cases := map[string]ToolEvent{
		"another folder with the same layout": fileEv("Read", "s1", "", filepath.Join(elsewhere, "internal/tui/a.go")),
		"a system file":                       fileEv("Edit", "s1", "", "/etc/hosts"),
		"a path that climbs out":              fileEv("Write", "s1", "", filepath.Join(repo, "..", "internal/tui/a.go")),
		"a folder with a longer name":         fileEv("MultiEdit", "s1", "", repo+"-old/internal/tui/a.go"),
		"a relative path that climbs out":     fileEv("Read", "s1", "", "../internal/tui/a.go"),
		"a bash word in another folder":       bashEv("s1", "", "cat "+filepath.Join(elsewhere, "internal/tui/a.go")),
	}
	for name, e := range cases {
		if got := WikiHints(root, repo, e); got != "" {
			t.Errorf("%s: hint = %q, want none", name, got)
		}
	}
}

// The folder the agent works in may be reached through a link while git names
// the real folder, and the other way round. macOS does this to /tmp and /var.
func TestWikiHintsFollowLinks(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(dir, link); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(dir, ".acta")
	hintPage(t, root, "tui-wrap", "Wrap can overflow after a dash; use Hardwrap(Wordwrap(...))", "internal/tui/")
	if err := os.MkdirAll(filepath.Join(dir, "internal", "tui"), 0o755); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, repo, cwd string
		ev              ToolEvent
	}{
		{"repo by its real name, file by the link", dir, dir, fileEv("Read", "", "", filepath.Join(link, "internal/tui/a.go"))},
		{"repo by the link, file by its real name", link, dir, fileEv("Read", "", "", filepath.Join(dir, "internal/tui/a.go"))},
		{"a new file under the link", dir, dir, fileEv("Write", "", "", filepath.Join(link, "internal/tui/new/b.go"))},
		{"bash in a folder reached by the link", dir, filepath.Join(link, "internal/tui"), bashEv("", "", "cat a.go")},
		{"bash in the real folder, repo by the link", link, filepath.Join(dir, "internal/tui"), bashEv("", "", "cat a.go")},
	}
	for i, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Chdir(c.cwd)
			c.ev.SessionID = fmt.Sprintf("s%d", i)
			if got := WikiHints(root, c.repo, c.ev); got != tuiLine {
				t.Errorf("hint = %q, want %q", got, tuiLine)
			}
		})
	}
}

func TestWikiHintsNeedAWiki(t *testing.T) {
	setups := map[string]func(root string) error{
		"no planning folder at all":      func(string) error { return nil },
		"a planning folder with no wiki": func(root string) error { return os.MkdirAll(root, 0o755) },
		"a wiki folder with no pages":    func(root string) error { return os.MkdirAll(filepath.Join(root, "wiki"), 0o755) },
	}
	for name, setup := range setups {
		t.Run(name, func(t *testing.T) {
			repo := t.TempDir()
			root := filepath.Join(repo, ".acta")
			if err := setup(root); err != nil {
				t.Fatal(err)
			}
			if got := WikiHints(root, repo, fileEv("Read", "s1", "", filepath.Join(repo, "internal/tui/a.go"))); got != "" {
				t.Errorf("hint = %q, want none", got)
			}
			if _, err := os.Stat(filepath.Join(root, "state")); !os.IsNotExist(err) {
				t.Error("a call with no wiki made a state folder")
			}
		})
	}
}

func TestWikiHintsNeedSomethingToLookAt(t *testing.T) {
	root, repo := hintWiki(t)
	for _, e := range []ToolEvent{{}, fileEv("Read", "s1", "", ""), bashEv("s1", "", "")} {
		if got := WikiHints(root, repo, e); got != "" {
			t.Errorf("hint for %+v = %q, want none", e, got)
		}
	}
}

// A call that shows nothing leaves the checkout as it found it.
func TestWikiHintsWriteNothingWithoutAHint(t *testing.T) {
	root, repo := hintWiki(t)
	t.Chdir(repo)
	for _, e := range []ToolEvent{
		fileEv("Read", "s1", "", filepath.Join(repo, "README.md")),
		bashEv("s1", "", "git status"),
	} {
		if got := WikiHints(root, repo, e); got != "" {
			t.Fatalf("hint = %q, want none", got)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "state")); !os.IsNotExist(err) {
		t.Error("a call with no hint made a state folder")
	}
}

// Most calls touch a file whose page was shown already. They must not take the
// lock or write a thing, since every Read and Edit of a covered file pays for it.
func TestWikiHintsRepeatTouchTakesNoLockAndWritesNothing(t *testing.T) {
	root, repo := hintWiki(t)
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	touch := fileEv("Read", "s1", "", filepath.Join(repo, "internal/tui/a.go"))
	if got := WikiHints(root, repo, touch); got != tuiLine {
		t.Fatalf("first touch: hint = %q, want %q", got, tuiLine)
	}
	// Clear away what the first touch left, and put the state in a shape the
	// code never writes, so a second write could not hide.
	for _, name := range []string{"state/wiki-hints.lock", ".gitignore"} {
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(name))); err != nil {
			t.Fatal(err)
		}
	}
	const handMade = "{\"s1/\": [\"tui-wrap\"]}\n"
	putHintState(t, root, handMade)

	if got := WikiHints(root, repo, touch); got != "" {
		t.Fatalf("repeat touch: hint = %q, want none", got)
	}
	for _, name := range []string{"state/wiki-hints.lock", ".gitignore"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(name))); err == nil {
			t.Errorf("a repeat touch made %s", name)
		}
	}
	if got := mustRead(t, filepath.Join(root, "state", "wiki-hints.json")); got != handMade {
		t.Errorf("a repeat touch rewrote the state to %q", got)
	}
}

func TestWikiHintsKeepShownPagesInState(t *testing.T) {
	root, repo := hintWiki(t)
	// A .git folder is all EnsureGitignore needs to see a repo.
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	WikiHints(root, repo, fileEv("Read", "s1", "", filepath.Join(repo, "internal/tui/a.go")))
	WikiHints(root, repo, fileEv("Read", "s1", "a1", filepath.Join(repo, "src/parser.go")))

	if got := mustRead(t, filepath.Join(root, "state", "wiki-hints.json")); got != `{"s1/":["tui-wrap"],"s1/a1":["parser"]}` {
		t.Errorf("state = %s, want one list of page ids for each session and agent", got)
	}
	if got := mustRead(t, filepath.Join(root, ".gitignore")); !strings.Contains(got, "state/") {
		t.Errorf(".gitignore = %q, want a line for state/", got)
	}
	// The write leaves no temp file behind.
	entries, err := os.ReadDir(filepath.Join(root, "state"))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if want := []string{"wiki-hints.json", "wiki-hints.lock"}; !slices.Equal(names, want) {
		t.Errorf("state folder holds %v, want %v", names, want)
	}
}

// A state file that went wrong must look like a context that was shown
// nothing, and the next hint must put it right.
func TestWikiHintsBrokenStateReadsAsNothingShown(t *testing.T) {
	for _, broken := range []string{"{not json", "", "null", `{"s1/":"tui-wrap"}`, `["tui-wrap"]`} {
		t.Run(broken, func(t *testing.T) {
			root, repo := hintWiki(t)
			putHintState(t, root, broken)
			got := WikiHints(root, repo, fileEv("Read", "s1", "", filepath.Join(repo, "internal/tui/a.go")))
			if got != tuiLine {
				t.Errorf("hint = %q, want %q", got, tuiLine)
			}
			if state := mustRead(t, filepath.Join(root, "state", "wiki-hints.json")); state != `{"s1/":["tui-wrap"]}` {
				t.Errorf("state after the repair = %s", state)
			}
		})
	}
}

// A hook that cannot remember what it showed still shows the hint. The hint
// then comes again on the next touch, which is how the user finds out.
func TestWikiHintsStillShowWhenStateCannotBeSaved(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can write anywhere")
	}
	root, repo := hintWiki(t)
	if err := os.Chmod(root, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(root, 0o755) })
	if got := WikiHints(root, repo, fileEv("Read", "s1", "", filepath.Join(repo, "internal/tui/a.go"))); got != tuiLine {
		t.Errorf("hint = %q, want %q", got, tuiLine)
	}
}

// A description that holds a line break or a tab must not break the one line
// the agent reads for the page.
func TestWikiHintsGiveOneLinePerPage(t *testing.T) {
	root, repo := hintWiki(t)
	hintPage(t, root, "wide", "first\nsecond\twith  gaps", "x.go")
	got := WikiHints(root, repo, fileEv("Read", "s1", "", filepath.Join(repo, "x.go")))
	if want := "wiki: .acta/wiki/wide.md: first second with gaps"; got != want {
		t.Errorf("hint = %q, want %q", got, want)
	}
}

// A planning folder outside the repo has no short name from the repo root, so
// the page is named by its full path, which the agent can open.
func TestWikiHintsNameAPageOutsideTheRepoByItsFullPath(t *testing.T) {
	repo, root := t.TempDir(), t.TempDir()
	hintPage(t, root, "far", "Lives elsewhere", "src/")
	got := WikiHints(root, repo, fileEv("Read", "s1", "", filepath.Join(repo, "src/a.go")))
	if want := "wiki: " + filepath.Join(root, "wiki", "far.md") + ": Lives elsewhere"; got != want {
		t.Errorf("hint = %q, want %q", got, want)
	}
}

func TestParseEventReadsHintFields(t *testing.T) {
	e, ok := ParseEvent(strings.NewReader(`{"session_id":"s1","agent_id":"a1","tool_name":"Read","tool_input":{"file_path":"/r/x.go","offset":3}}`))
	if !ok || e.SessionID != "s1" || e.AgentID != "a1" || e.ToolName != "Read" || e.ToolInput.FilePath != "/r/x.go" {
		t.Fatalf("got %+v %v", e, ok)
	}
	// The main thread sends no agent_id.
	e, ok = ParseEvent(strings.NewReader(`{"session_id":"s1","tool_name":"Bash","tool_input":{"command":"ls"}}`))
	if !ok || e.AgentID != "" || e.ToolName != "Bash" || e.ToolInput.Command != "ls" || e.ToolInput.FilePath != "" {
		t.Fatalf("got %+v %v", e, ok)
	}
}

func TestHintJSON(t *testing.T) {
	got := HintJSON("wiki: a.md: one\nwiki: b.md: two")
	want := `{"hookSpecificOutput":{"hookEventName":"PreToolUse","additionalContext":"wiki: a.md: one\nwiki: b.md: two"}}`
	if got != want {
		t.Errorf("HintJSON = %s, want %s", got, want)
	}
}

// putHintState writes the shown-pages file by hand, for the broken cases.
func putHintState(t *testing.T, root, body string) {
	t.Helper()
	writeJSON(t, filepath.Join(root, "state", "wiki-hints.json"), body)
}
