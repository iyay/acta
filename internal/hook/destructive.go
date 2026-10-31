package hook

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// destructiveText tells the agent why the command stopped and hands the
// choice back to the user, who can still run it outside the agent.
const destructiveText = "acta never runs destructive commands, so it blocked: %s. Tell the user in full sentences what you wanted to run and why. They can run it themselves outside the agent if they mean it."

// destructiveSep adds a newline to the go test separators, because each line
// of a many-line command runs as a command of its own.
const destructiveSep = goTestSep + "\n"

// destructivePrefix holds the words that run the command behind them, with
// the flags of each that take the next word as their value. Those values are
// skipped too, so `sudo -u root rm -rf x` is still read as rm.
var destructivePrefix = map[string]map[string]bool{
	"sudo":    {"-u": true, "-g": true, "-C": true, "-D": true, "-h": true, "-p": true, "-U": true},
	"env":     {"-u": true, "-C": true, "-S": true},
	"command": {},
	"nohup":   {},
	"xargs":   {"-I": true, "-n": true, "-P": true, "-L": true, "-d": true, "-E": true, "-s": true, "-a": true},
}

// findEnd holds the words that close the command of a find -exec.
var findEnd = map[string]bool{";": true, `\;`: true, "';'": true, `";"`: true, "+": true}

// DestructiveBlock says whether this command would run something that throws
// work away. It says no on every doubt, like GoTestBlock: a folder with no
// repo above it, a quote that never closes, or a git it cannot ask about.
func DestructiveBlock(dir, command string) (bool, string) {
	if _, ok := gitTopDir(dir); !ok {
		return false, ""
	}
	pieces, ok := goTestSplit(command, destructiveSep)
	if !ok {
		return false, ""
	}
	// An amend works on the repo of the last cd before it, so the walk keeps
	// track of the folder each piece runs in.
	cwd := dir
	for _, piece := range pieces {
		words := goTestCut(piece, goTestSpace)
		cmd := commandWords(words)
		if len(cmd) == 2 && cmd[0] == "cd" {
			cwd = joinDir(cwd, cmd[1])
			continue
		}
		if destructive(cmd, cwd) {
			return true, fmt.Sprintf(destructiveText, strings.Join(words, " "))
		}
	}
	return false, ""
}

// commandWords drops the words in front of the command that do not change
// which command runs: NAME=value words, the go test prefixes, and each
// destructive prefix with its flags.
func commandWords(words []string) []string {
	for len(words) > 0 {
		head := words[0]
		switch {
		case destructivePrefix[head] != nil:
			words = skipFlags(words[1:], destructivePrefix[head])
		case goTestPrefix[head] || goTestAssign.MatchString(head):
			words = words[1:]
		default:
			return words
		}
	}
	return nil
}

// skipFlags drops the flags at the start of words, and the value word of any
// flag in takesValue.
func skipFlags(words []string, takesValue map[string]bool) []string {
	for len(words) > 0 && strings.HasPrefix(words[0], "-") {
		if takesValue[words[0]] && len(words) > 1 {
			words = words[2:]
			continue
		}
		words = words[1:]
	}
	return words
}

// destructive judges one command whose prefixes are already gone. cwd is the
// folder it runs in, which only the amend check needs.
func destructive(cmd []string, cwd string) bool {
	if len(cmd) == 0 {
		return false
	}
	// A full path like /bin/rm still runs rm.
	switch filepath.Base(cmd[0]) {
	case "rm":
		return rmRecursiveForce(cmd[1:])
	case "git":
		return gitDestructive(cmd[1:], cwd)
	case "find":
		return findExecDestructive(cmd[1:], cwd)
	}
	return false
}

// rmRecursiveForce wants both a recursive and a force flag, in any spelling,
// because rm -r alone asks before it deletes and rm -f alone keeps folders.
func rmRecursiveForce(args []string) bool {
	recursive, force := false, false
	for _, a := range args {
		switch {
		case a == "--recursive":
			recursive = true
		case a == "--force":
			force = true
		case strings.HasPrefix(a, "-") && !strings.HasPrefix(a, "--"):
			recursive = recursive || strings.ContainsAny(a, "rR")
			force = force || strings.Contains(a, "f")
		}
	}
	return recursive && force
}

// findExecDestructive reads the words after -exec or -execdir as a command
// of their own, since find runs them once for each file it finds.
func findExecDestructive(args []string, cwd string) bool {
	for i, a := range args {
		if a != "-exec" && a != "-execdir" {
			continue
		}
		end := i + 1
		for end < len(args) && !findEnd[args[end]] {
			end++
		}
		if destructive(commandWords(args[i+1:end]), cwd) {
			return true
		}
	}
	return false
}

// gitDestructive skips the global options in front of the git subcommand,
// then judges the subcommand. A -C moves the repo the amend check looks at.
func gitDestructive(args []string, cwd string) bool {
	for _, a := range args {
		if a == "--force" || a == "--force-with-lease" || strings.HasPrefix(a, "--force-with-lease=") {
			return true
		}
	}
	repo := cwd
	i := 0
	for i < len(args) && strings.HasPrefix(args[i], "-") {
		switch args[i] {
		case "-C":
			if i+1 < len(args) {
				repo = joinDir(repo, args[i+1])
			}
			i += 2
		case "-c":
			i += 2
		default:
			// --git-dir=<path>, --work-tree=<path> and the other global
			// flags are one word each.
			i++
		}
	}
	if i >= len(args) {
		return false
	}
	sub, rest := args[i], args[i+1:]
	switch sub {
	case "push", "clean":
		return hasShortFlag(rest, 'f')
	case "worktree":
		return len(rest) > 0 && rest[0] == "remove" && hasShortFlag(rest[1:], 'f')
	case "reset":
		return slices.Contains(rest, "--hard")
	case "commit":
		return slices.Contains(rest, "--amend") && headIsMerge(repo)
	}
	return false
}

// hasShortFlag finds a flag letter alone or joined with others, like the f
// in -f, -fd and -fdx.
func hasShortFlag(args []string, letter byte) bool {
	for _, a := range args {
		if strings.HasPrefix(a, "-") && !strings.HasPrefix(a, "--") && strings.IndexByte(a, letter) > 0 {
			return true
		}
	}
	return false
}

// headIsMerge says whether HEAD of repo has two or more parents. Amending a
// merge commit can drop the work it joined, while amending one's own plain
// commit is fine. Any error says no, so a git the hook cannot run never
// blocks anything.
func headIsMerge(repo string) bool {
	out, err := exec.Command("git", "-C", repo, "rev-list", "--parents", "-n", "1", "HEAD").Output()
	if err != nil {
		return false
	}
	// The line is the commit itself followed by one word per parent.
	return len(strings.Fields(string(out))) >= 3
}

// joinDir resolves a cd or -C path against the folder the command is in. A
// quoted path loses its quotes, because the shell drops them too.
func joinDir(base, path string) string {
	if len(path) >= 2 && (path[0] == '"' || path[0] == '\'') && path[len(path)-1] == path[0] {
		path = path[1 : len(path)-1]
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(base, path)
}
