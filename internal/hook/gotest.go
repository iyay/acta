package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// goTestBlockText stops a bare go test in a repo that has scripts/test, and
// names the same run through that script so the agent can just use it.
const goTestBlockText = "acta: run tests with scripts/test, not go test. It adds -short and the machine-wide lock. Use: scripts/test%s"

// goTestCut splits a command line where the shell starts its next command, so
// a go test behind one of them is still a command of its own.
var goTestCut = regexp.MustCompile(`&&|\|\||;|\||\(|\)`)

// goTestPrefix holds the words that can stand in front of a command without
// changing which command runs.
var goTestPrefix = map[string]bool{"rtk": true, "proxy": true, "env": true, "time": true}

// goTestRunOne is the wrapper scripts/test itself uses, so a piece that holds
// it already has the lock and must be left alone.
const goTestRunOne = "run-one --"

// goTestAssign matches a NAME=value word in front of a command.
var goTestAssign = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

// GoTestBlock says whether this command really runs a bare go test in a repo
// that keeps its tests in scripts/test. It says no on every doubt, because a
// hook that stops the wrong work costs the user more than a run that slips
// past: a folder with no repo above it, a repo with no scripts/test and a
// command that cannot be read as a command all go through.
func GoTestBlock(dir, command string) (bool, string) {
	top, ok := gitTopDir(dir)
	if !ok || !hasTestScript(top) {
		return false, ""
	}
	for _, piece := range goTestCut.Split(command, -1) {
		words := goTestWords(piece)
		if len(words) < 2 || words[0] != "go" || words[1] != "test" {
			continue
		}
		if strings.Contains(piece, goTestRunOne) {
			continue
		}
		return true, fmt.Sprintf(goTestBlockText, goTestUse(words[2:]))
	}
	return false, ""
}

// gitTopDir walks up to the first folder holding .git, which is the repo
// root. A worktree has .git as a file, so both shapes count.
func gitTopDir(dir string) (string, bool) {
	if dir == "" {
		return "", false
	}
	cur, err := filepath.Abs(dir)
	if err != nil {
		return "", false
	}
	for {
		if _, err := os.Lstat(filepath.Join(cur, ".git")); err == nil {
			return cur, true
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", false
		}
		cur = parent
	}
}

// hasTestScript says whether this repo keeps its tests in scripts/test. It
// opens the file, so a script the hook cannot read is treated as no script:
// a broken runner must never turn into a hook that stops every Bash call.
func hasTestScript(top string) bool {
	f, err := os.Open(filepath.Join(top, "scripts", "test"))
	if err != nil {
		return false
	}
	defer f.Close()
	fi, err := f.Stat()
	return err == nil && fi.Mode().IsRegular()
}

// goTestWords drops the words that stand in front of a command, so
// `env FOO=1 time go test` is seen as the go test it runs.
func goTestWords(piece string) []string {
	words := strings.Fields(piece)
	for len(words) > 0 {
		head := words[0]
		if goTestPrefix[head] || goTestAssign.MatchString(head) {
			words = words[1:]
			continue
		}
		return words
	}
	return words
}

// goTestUse names the same run through scripts/test: the packages the command
// asked for and its -run value, because those two are what a test loop needs
// to say again. Other flags are left out, so the suggestion stays a command
// that runs.
func goTestUse(args []string) string {
	var keep []string
	for i := range args {
		switch {
		case args[i] == "-run" && i+1 < len(args):
			keep = append(keep, "-run", args[i+1])
			i++
		case strings.HasPrefix(args[i], "-run="):
			keep = append(keep, args[i])
		case isPackageArg(args[i]):
			keep = append(keep, args[i])
		}
	}
	if len(keep) == 0 {
		return ""
	}
	// A subshell closes after the command, so its bracket belongs to the
	// shell, not to the run the agent has to copy.
	return " " + strings.TrimRight(strings.Join(keep, " "), ")]}")
}

// isPackageArg keeps a package pattern and drops a flag value like the 1800s
// of -timeout, which is a number and not something to test.
func isPackageArg(s string) bool {
	return strings.HasPrefix(s, "./") || strings.HasPrefix(s, "../") || strings.Contains(s, "/")
}
