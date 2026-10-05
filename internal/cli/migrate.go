package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// cmdMigrateRoot runs the move from the folder the user stands in, so the
// command needs no arguments.
func cmdMigrateRoot(args []string, stdout, stderr io.Writer) int {
	if len(args) != 0 {
		fmt.Fprintln(stderr, "usage: acta migrate-root")
		return exitBadInput
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	return migrateRoot(cwd, stdout, stderr)
}

const migrateRootMsg = "chore: move root folder .pm to .acta"

// migrateRoot moves .pm/ to .acta/ with git mv, renames .pm.yaml when it
// exists, drops a root line that only says ".pm", and commits once with a
// pathspec so a file the user had staged stays staged. Everything that could
// make the move ambiguous is refused before the first git mv, and a later
// failure puts the repo back. It never pushes.
func migrateRoot(repo string, stdout, stderr io.Writer) int {
	top, err := gitTop(repo)
	if err != nil {
		fmt.Fprintln(stderr, "migrate-root: not a git repo, nothing moved")
		return exitBadInput
	}
	pmDir := filepath.Join(top, ".pm")
	actaDir := filepath.Join(top, ".acta")
	if _, err := os.Stat(actaDir); err == nil {
		fmt.Fprintln(stderr, "migrate-root: .acta/ already exists, nothing moved")
		return exitBadInput
	}
	if st, err := os.Stat(pmDir); err != nil || !st.IsDir() {
		fmt.Fprintln(stderr, "migrate-root: no .pm/ to move, nothing to do")
		return exitBadInput
	}
	// The checks come first, so a refusal leaves the working tree, the index
	// and HEAD exactly as they were.
	dropRoot, code := checkMovable(top, stderr)
	if code != exitOK {
		return code
	}

	moved := []string{".pm/ -> .acta/"}
	// spec names the paths the move may commit, so a file the user had staged
	// stays staged instead of riding along in the move commit.
	spec := []string{".pm", ".acta"}
	var yamlBody []byte
	var yamlMoved bool
	// undo puts back whatever this run moved, so a later failure never leaves
	// the repo half-moved.
	undo := func() {
		if yamlMoved {
			// Write the old text back and stage it again, because git mv takes
			// what the index holds, and the rewritten yaml is what the index
			// holds now.
			if err := os.WriteFile(filepath.Join(top, ".acta.yaml"), yamlBody, 0o644); err == nil {
				if git(top, "mv", ".acta.yaml", ".pm.yaml") == nil {
					_ = git(top, "add", "--", ".pm.yaml")
				}
			}
		}
		_ = git(top, "mv", ".acta", ".pm")
	}
	if err := git(top, "mv", ".pm", ".acta"); err != nil {
		fmt.Fprintf(stderr, "migrate-root: %v, nothing moved\n", err)
		return exitOther
	}
	if body, err := os.ReadFile(filepath.Join(top, ".pm.yaml")); err == nil {
		yamlBody = body
		if err := git(top, "mv", ".pm.yaml", ".acta.yaml"); err != nil {
			undo()
			fmt.Fprintf(stderr, "migrate-root: %v, move undone\n", err)
			return exitOther
		}
		yamlMoved = true
		moved = append(moved, ".pm.yaml -> .acta.yaml")
		spec = append(spec, ".pm.yaml", ".acta.yaml")
		if dropRoot {
			// A root line that says .pm would point the board at the folder
			// that is gone, so it goes away and the default lookup finds
			// .acta/ on its own. The rest of the file is untouched, comments
			// and spacing included.
			out := filepath.Join(top, ".acta.yaml")
			if err := os.WriteFile(out, withoutRootLine(yamlBody), 0o644); err != nil {
				undo()
				fmt.Fprintf(stderr, "migrate-root: %v, move undone\n", err)
				return exitOther
			}
			if err := git(top, "add", "--", ".acta.yaml"); err != nil {
				undo()
				fmt.Fprintf(stderr, "migrate-root: %v, move undone\n", err)
				return exitOther
			}
		}
	}
	if err := git(top, append([]string{"commit", "-q", "-m", migrateRootMsg, "--"}, spec...)...); err != nil {
		undo()
		fmt.Fprintf(stderr, "migrate-root: %v, move undone\n", err)
		return exitOther
	}
	fmt.Fprintf(stdout, "moved %s\n", strings.Join(moved, ", "))
	return exitOK
}

// checkMovable refuses every repo state where the move would be wrong or
// impossible: a .acta.yaml that is in the way, .pm/ with uncommitted work, and
// a .pm.yaml that git cannot rename as it is or that points the board somewhere
// the move does not reach. It returns whether the root line has to go.
func checkMovable(top string, stderr io.Writer) (bool, int) {
	// The move ends in a file called .acta.yaml, so that name has to be free
	// before anything moves. os.Lstat looks at the name itself and never
	// follows a link, so a link that points nowhere or at itself still counts
	// as a name that is taken, which os.Stat would miss.
	if _, err := os.Lstat(filepath.Join(top, ".acta.yaml")); err == nil {
		fmt.Fprintln(stderr, "migrate-root: .acta.yaml already exists, move or delete it yourself, nothing moved")
		return false, exitBadInput
	} else if !os.IsNotExist(err) {
		// The name could not be looked up, so nobody can promise it is free,
		// and the move must not run into a name that is already taken.
		fmt.Fprintln(stderr, "migrate-root: could not read .acta.yaml, nothing moved")
		return false, exitOther
	}
	dirty, err := pmDirty(top)
	if err != nil {
		fmt.Fprintln(stderr, "migrate-root: could not check git status, nothing moved")
		return false, exitOther
	}
	if dirty {
		fmt.Fprintln(stderr, "migrate-root: .pm/ has uncommitted changes, commit or stash first")
		return false, exitBadInput
	}
	body, err := os.ReadFile(filepath.Join(top, ".pm.yaml"))
	if os.IsNotExist(err) {
		return false, exitOK
	}
	if err != nil {
		fmt.Fprintf(stderr, "migrate-root: %v, nothing moved\n", err)
		return false, exitOther
	}
	// An untracked, modified or staged .pm.yaml cannot be renamed as it is, and
	// finding that out after the move would leave the repo half-moved.
	if s, err := gitStatus(top, "--", ".pm.yaml"); err != nil {
		fmt.Fprintln(stderr, "migrate-root: could not check git status, nothing moved")
		return false, exitOther
	} else if s != "" {
		fmt.Fprintln(stderr, "migrate-root: .pm.yaml has uncommitted changes, commit or stash first")
		return false, exitBadInput
	}
	// An ignored .pm.yaml is invisible to a plain status, so ask git whether it
	// keeps the file at all. Without this the move happens and git mv fails
	// afterwards, and the user reads a raw git error instead of a clear reason.
	if tracked, err := gitTracked(top, ".pm.yaml"); err != nil {
		fmt.Fprintln(stderr, "migrate-root: could not ask git about .pm.yaml, nothing moved")
		return false, exitOther
	} else if !tracked {
		fmt.Fprintln(stderr, "migrate-root: .pm.yaml is not tracked, add and commit it or delete it, then run this again")
		return false, exitBadInput
	}
	value, ok := rootValue(string(body))
	if !ok {
		return false, exitOK
	}
	if !defaultRoot(value) {
		// Rewriting a root the user chose would be a guess, and leaving it
		// alone would send the board to a folder this move did not touch.
		fmt.Fprintf(stderr, "migrate-root: .pm.yaml sets root to %q, move it yourself, nothing moved\n", value)
		return false, exitBadInput
	}
	return true, exitOK
}

// defaultRoot tells whether a root value is just the folder this command
// renames, written as .pm, ./.pm, .pm/ or ./.pm/, or left empty to take the
// default.
func defaultRoot(value string) bool {
	switch value {
	case "", ".pm", "./.pm", ".pm/", "./.pm/":
		return true
	}
	return false
}

// rootValue returns the value of the top level root: line and whether the file
// has one. Only a line that starts at the left margin counts, so a root: under
// another key is left alone.
func rootValue(body string) (string, bool) {
	for _, line := range strings.Split(body, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if ok && strings.TrimSpace(key) == "root" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			return strings.TrimSpace(value), true
		}
	}
	return "", false
}

// withoutRootLine takes the root: line out of the yaml text, so the file that
// ends up as .acta.yaml has no root of its own. Every other byte, comments
// included, is left alone.
func withoutRootLine(body []byte) []byte {
	lines := strings.Split(string(body), "\n")
	for i, line := range lines {
		key, _, ok := strings.Cut(line, ":")
		if ok && strings.TrimSpace(key) == "root" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			return []byte(strings.Join(append(lines[:i:i], lines[i+1:]...), "\n"))
		}
	}
	return body
}

func gitTop(repo string) (string, error) {
	out, err := exec.Command("git", "-C", repo, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func git(dir string, args ...string) error {
	if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
		return fmt.Errorf("git %s: %v %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

// pmDirty is true when a file under .pm/ changed, is untracked or is ignored,
// so the move never hides work in progress. An ignored folder is invisible to
// a plain status, and git mv would fail on it late.
func pmDirty(top string) (bool, error) {
	s, err := gitStatus(top, "--ignored=matching", "--", ".pm")
	return s != "", err
}

// gitStatus returns git's short status for the given arguments, empty when
// git sees no change. An untracked, modified or staged file all show up.
func gitStatus(top string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", top, "status", "--porcelain"}, args...)...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// gitTracked tells whether git keeps the file at the given path in its index.
// A file git does not list is not tracked, whether it is ignored or simply
// new, and git mv cannot rename it.
func gitTracked(top, path string) (bool, error) {
	out, err := exec.Command("git", "-C", top, "ls-files", "--", path).Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}
