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

const migrateRootMsg = "acta: move root folder .pm to .acta"

// migrateRoot moves .pm/ to .acta/ with git mv, renames .pm.yaml when it
// exists, and commits once. It refuses when .acta/ already exists, when .pm/
// is missing, or when .pm/ has uncommitted changes. It never pushes.
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
	if dirty, err := pmDirty(top); err != nil {
		fmt.Fprintln(stderr, "migrate-root: could not check git status, nothing moved")
		return exitOther
	} else if dirty {
		fmt.Fprintln(stderr, "migrate-root: .pm/ has uncommitted changes, commit or stash first")
		return exitBadInput
	}
	moved := []string{".pm/ -> .acta/"}
	if err := git(top, "mv", ".pm", ".acta"); err != nil {
		fmt.Fprintf(stderr, "migrate-root: %v, nothing moved\n", err)
		return exitOther
	}
	if _, err := os.Stat(filepath.Join(top, ".pm.yaml")); err == nil {
		if err := git(top, "mv", ".pm.yaml", ".acta.yaml"); err != nil {
			fmt.Fprintf(stderr, "migrate-root: %v, root moved but .pm.yaml kept its name\n", err)
			return exitOther
		}
		moved = append(moved, ".pm.yaml -> .acta.yaml")
	}
	if err := git(top, "commit", "-q", "-m", migrateRootMsg); err != nil {
		fmt.Fprintf(stderr, "migrate-root: %v, files moved but not committed\n", err)
		return exitOther
	}
	fmt.Fprintf(stdout, "moved %s\n", strings.Join(moved, ", "))
	return exitOK
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

// pmDirty is true when a tracked file under .pm/ changed or any file under
// .pm/ is untracked, so the move never hides work in progress.
func pmDirty(top string) (bool, error) {
	out, err := exec.Command("git", "-C", top, "status", "--porcelain", "--", ".pm").Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}
