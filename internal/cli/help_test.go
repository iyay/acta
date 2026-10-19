package cli

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/iyay/acta/plugin"
)

// helpRun calls Run the way main does and gives back the exit code and both
// streams.
func helpRun(args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	code := Run(args, strings.NewReader(""), false, &out, &errb)
	return code, out.String(), errb.String()
}

// Every way to ask for help must print table text on stdout, exit 0, say
// nothing on stderr, and never show the Go flag package's own usage text.
func TestHelpEveryAskPrintsTableText(t *testing.T) {
	cmd, _ := findHelpCmd("tick")
	asks := []struct {
		name string
		args []string
		want string
	}{
		{"help", []string{"help"}, "Board:"},
		{"--help", []string{"--help"}, "Board:"},
		{"-h", []string{"-h"}, "Board:"},
		{"-root X --help", []string{"-root", "X", "--help"}, "Board:"},
		{"help tick", []string{"help", "tick"}, cmd.usage[0]},
		{"tick --help", []string{"tick", "--help"}, cmd.usage[0]},
		{"tick -h", []string{"tick", "-h"}, cmd.usage[0]},
		{"tick x --start --help", []string{"tick", "x", "--start", "--help"}, cmd.usage[0]},
		{"bug new --help", []string{"bug", "new", "--help"}, "acta bug new"},
		{"wiki ls -h", []string{"wiki", "ls", "-h"}, "acta wiki ls"},
	}
	for _, a := range asks {
		code, out, errb := helpRun(a.args...)
		if code != exitOK {
			t.Errorf("%s: exit %d, want 0", a.name, code)
		}
		if !strings.Contains(out, a.want) {
			t.Errorf("%s: stdout %q does not hold %q", a.name, out, a.want)
		}
		if errb != "" {
			t.Errorf("%s: stderr %q, want empty", a.name, errb)
		}
		if strings.Contains(out+errb, "Usage of") {
			t.Errorf("%s: printed the Go flag default usage", a.name)
		}
	}
}

// The top help names every command of the table under its group, in the
// group order the spec fixes, and shows the global flags.
func TestHelpTopListsEveryCommandInGroupOrder(t *testing.T) {
	_, out, _ := helpRun("help")
	groups := []string{"Board:", "Write:", "Setup:", "Agent plumbing:"}
	last := -1
	for _, g := range groups {
		i := strings.Index(out, "\n"+g+"\n")
		if i < 0 || i < last {
			t.Fatalf("group %q missing or out of order in:\n%s", g, out)
		}
		last = i
	}
	for _, c := range helpTable {
		if !strings.Contains(out, "\n  "+c.name+" ") {
			t.Errorf("top help does not list %q", c.name)
		}
	}
	for _, want := range []string{"--version", "--help", "-root"} {
		if !strings.Contains(out, want) {
			t.Errorf("top help does not show %q", want)
		}
	}
	if strings.Contains(out, "voice") {
		t.Error("top help lists the old voice alias")
	}
}

// help <cmd> shows the summary, every usage line, every flag and the
// example of that command, for every command of the table.
func TestHelpDetailShowsEveryPartOfTheEntry(t *testing.T) {
	for _, c := range helpTable {
		code, out, errb := helpRun("help", c.name)
		if code != exitOK || errb != "" {
			t.Errorf("%s: exit %d, stderr %q", c.name, code, errb)
		}
		parts := append([]string{c.summary, c.example}, c.usage...)
		parts = append(parts, c.flags...)
		for _, p := range parts {
			if !strings.Contains(out, p) {
				t.Errorf("%s: detail does not show %q", c.name, p)
			}
		}
		if len(c.usage) == 0 || c.summary == "" {
			t.Errorf("%s: entry has no usage or no summary", c.name)
		}
	}
}

// Commands that write files carry an example.
func TestHelpWritingCommandsHaveAnExample(t *testing.T) {
	for _, name := range []string{"bug", "debt", "scratch", "commit", "id", "wiki", "set", "state", "tick", "tidy", "migrate-root", "doctor", "setup", "update"} {
		c, ok := findHelpCmd(name)
		if !ok || c.example == "" {
			t.Errorf("%s: no example in the table", name)
		}
	}
}

// An unknown name goes to stderr with exit 1, never to stdout, and the long
// command list is gone from the message.
func TestUnknownCommandGoesToStderr(t *testing.T) {
	for _, args := range [][]string{
		{"frobnicate"},
		{"help", "frobnicate"},
		{"frobnicate", "--help"},
		{"voice"},
	} {
		code, out, errb := helpRun(args...)
		name := args[len(args)-1]
		if args[0] == "frobnicate" || args[0] == "voice" {
			name = args[0]
		}
		want := "unknown command " + strconv.Quote(name) + "; run acta help\n"
		if code != exitBadInput || out != "" || errb != want {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want 1, empty, %q", args, code, out, errb, want)
		}
	}
}

// A new case in Run's switch with no table entry must fail here. The test
// reads cli.go itself, so nobody can forget the table.
func TestEveryRunCaseHasATableEntry(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "cli.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var cases []string
	ast.Inspect(file, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "Run" {
			return true
		}
		for _, st := range fn.Body.List {
			sw, ok := st.(*ast.SwitchStmt)
			if !ok || !isArgsZero(sw.Tag) {
				continue
			}
			for _, c := range sw.Body.List {
				for _, e := range c.(*ast.CaseClause).List {
					if lit, ok := e.(*ast.BasicLit); ok && lit.Kind == token.STRING {
						s, _ := strconv.Unquote(lit.Value)
						cases = append(cases, s)
					}
				}
			}
		}
		return false
	})
	if len(cases) < 20 {
		t.Fatalf("found only %d cases in Run; the test cannot find the switch", len(cases))
	}
	for _, name := range cases {
		if _, ok := findHelpCmd(name); !ok {
			t.Errorf("Run has case %q but the help table has no entry for it", name)
		}
	}
}

// isArgsZero says whether the expression is args[0].
func isArgsZero(e ast.Expr) bool {
	ix, ok := e.(*ast.IndexExpr)
	if !ok {
		return false
	}
	id, ok := ix.X.(*ast.Ident)
	lit, ok2 := ix.Index.(*ast.BasicLit)
	return ok && ok2 && id.Name == "args" && lit.Value == "0"
}

func TestVersionPrintsReleaseAndSourceBuild(t *testing.T) {
	old := isRelease
	defer func() { isRelease = old }()
	for _, tc := range []struct {
		release bool
		want    string
	}{
		{true, "acta " + plugin.Version() + " (" + wantVersionWord + ")\n"},
		{false, "acta " + plugin.Version() + " (" + wantVersionWord + ", source build)\n"},
	} {
		isRelease = func() bool { return tc.release }
		for _, args := range [][]string{{"--version"}, {"-root", "X", "--version"}} {
			code, out, errb := helpRun(args...)
			if code != exitOK || out != tc.want || errb != "" {
				t.Errorf("release=%v %q: exit %d, stdout %q, stderr %q; want 0, %q, empty", tc.release, args, code, out, errb, tc.want)
			}
		}
	}
}

// Nothing after -- is read as a help flag: run-one hands it to the child.
func TestRunOnePassesHelpFlagToTheChild(t *testing.T) {
	code, out, errb := runOne(t, "", "--", "sh", "-c", `echo "child got: $1"`, "sh", "--help")
	if code != exitOK || out != "child got: --help\n" || errb != "" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, out, errb)
	}
	code, out, _ = runOne(t, "", "--", "echo", "-h", "--help")
	if code != exitOK || out != "-h --help\n" {
		t.Fatalf("exit %d, stdout %q; the child did not get -h --help", code, out)
	}
}

// run-one asked for help before the -- still answers with its own help.
func TestRunOneHelpBeforeTheDoubleDash(t *testing.T) {
	code, out, _ := helpRun("run-one", "--help")
	if code != exitOK || !strings.Contains(out, "acta run-one") {
		t.Fatalf("exit %d, stdout %q", code, out)
	}
}

// Bad-call usage lines come from the help table. A "usage: acta" text
// typed into another file could drift from acta help, so none may be left.
func TestNoUsageLiteral(t *testing.T) {
	files, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		name := f.Name()
		if f.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || name == "help.go" {
			continue
		}
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(src), "\n") {
			if strings.Contains(line, "usage: acta") {
				t.Errorf("%s:%d has a usage literal; read it from the help table with usageLine", name, i+1)
			}
		}
	}
}

// The debt help must say which notes it takes, so a refused batch is no
// surprise.
func TestHelpDebtNamesPriorityRule(t *testing.T) {
	_, out, _ := helpRun("debt", "--help")
	if !strings.Contains(out, "(medium) or (high)") {
		t.Fatalf("debt help does not name the priority rule:\n%s", out)
	}
}

// config help names --review-understanding for both set forms, and says it
// can be unset per repo.
func TestHelpConfigReviewUnderstanding(t *testing.T) {
	cmd, ok := findHelpCmd("config")
	if !ok {
		t.Fatal("no config entry")
	}
	if !strings.Contains(cmd.usage[1], "[--review-understanding true|false]") || !strings.Contains(cmd.usage[2], "[--review-understanding B]") {
		t.Errorf("usage lacks --review-understanding:\n%s", strings.Join(cmd.usage, "\n"))
	}
	var flagLine, unsetLine bool
	for _, f := range cmd.flags {
		// The user dropped the quiz after a land, so the help must not
		// promise a question.
		flagLine = flagLine || (strings.HasPrefix(f, "--review-understanding B  ") && strings.Contains(f, "explain the change") && !strings.Contains(f, "question"))
		unsetLine = unsetLine || (strings.HasPrefix(f, "--unset KEY") && strings.Contains(f, "review_understanding"))
	}
	if !flagLine || !unsetLine {
		t.Errorf("flags lack the --review-understanding line (%v) or the unset key (%v):\n%s", flagLine, unsetLine, strings.Join(cmd.flags, "\n"))
	}
}
