package hook

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
)

// The block text, written out here on purpose: a change to the words an agent
// reads is a change to the product, not to a test.
const destructiveWant = "acta never runs destructive commands, so it blocked: %s. Tell the user in full sentences what you wanted to run and why. They can run it themselves outside the agent if they mean it."

// destructiveRepo makes a real git repo under the test's temp folder. Without
// merge, HEAD has one parent, like an agent's own plain commit. With merge,
// HEAD is a merge commit with two parents, which is the one HEAD the amend
// check must protect.
func destructiveRepo(t *testing.T, merge bool) string {
	t.Helper()
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		// The name, email and no signing are set here, so the test never
		// depends on the git setup of the machine it runs on.
		full := append([]string{"-C", dir, "-c", "user.name=test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false"}, args...)
		if out, err := exec.Command("git", full...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	git("init", "-q", "-b", "main")
	git("commit", "-q", "--allow-empty", "-m", "first")
	if !merge {
		git("commit", "-q", "--allow-empty", "-m", "second")
		return dir
	}
	git("checkout", "-q", "-b", "side")
	git("commit", "-q", "--allow-empty", "-m", "side")
	git("checkout", "-q", "main")
	git("commit", "-q", "--allow-empty", "-m", "main")
	git("merge", "-q", "--no-ff", "-m", "merge side", "side")
	return dir
}

// TestDestructiveBlock walks every command shape the spec names. A row with
// an empty want must run untouched, because a hook that stops the wrong work
// costs the user more than one that misses.
func TestDestructiveBlock(t *testing.T) {
	plain := destructiveRepo(t, false)
	merged := destructiveRepo(t, true)
	// No .git anywhere above, so there is no repo to speak of.
	noRepo := t.TempDir()
	other := filepath.Join(plain, "..", "other")

	for _, c := range []struct{ name, dir, command, want string }{
		{"rm -rf", plain, "rm -rf x", "rm -rf x"},
		{"rm -fr", plain, "rm -fr x", "rm -fr x"},
		{"rm -Rf", plain, "rm -Rf x", "rm -Rf x"},
		{"rm -r -f", plain, "rm -r -f x", "rm -r -f x"},
		{"rm long flags", plain, "rm --recursive --force x", "rm --recursive --force x"},
		{"git push --force", plain, "git push --force origin main", "git push --force origin main"},
		{"git push --force-with-lease", plain, "git push --force-with-lease", "git push --force-with-lease"},
		{"git push -f", plain, "git push -f origin main", "git push -f origin main"},
		{"git worktree remove -f", plain, "git worktree remove -f x", "git worktree remove -f x"},
		{"git clean -f", plain, "git clean -f", "git clean -f"},
		{"git clean -fdx", plain, "git clean -fdx", "git clean -fdx"},
		{"git reset --hard", plain, "git reset --hard HEAD~1", "git reset --hard HEAD~1"},
		{"amend on a merge found through dir", merged, "git commit --amend --no-edit", "git commit --amend --no-edit"},
		{"amend on a merge found through -C", plain, "git -C " + merged + " commit --amend --no-edit", "git -C " + merged + " commit --amend --no-edit"},
		{"amend on a merge found through cd", plain, "cd " + merged + " && git commit --amend --no-edit", "git commit --amend --no-edit"},
		{"git -C worktree remove -f", plain, "git -C " + other + " worktree remove -f x", "git -C " + other + " worktree remove -f x"},
		{"find -exec rm -rf", plain, "find . -name x -exec rm -rf {} +", "find . -name x -exec rm -rf {} +"},
		{"sudo rm -rf", plain, "sudo rm -rf x", "sudo rm -rf x"},
		{"a newline split command", plain, "echo hi\nrm -rf x", "rm -rf x"},

		{"git branch -D", plain, "git branch -D x", ""},
		{"rm -r alone", plain, "rm -r dir", ""},
		{"rm -f alone", plain, "rm -f file", ""},
		{"amend on a plain HEAD", plain, "git commit --amend --no-edit", ""},
		{"a quoted flag in a message", plain, `git commit -m "no --force here"`, ""},
		{"an echo that mentions rm -rf", plain, `echo "rm -rf x"`, ""},
		{"no repo above dir", noRepo, "rm -rf x", ""},
		{"a quote that never closes", plain, `rm -rf "x`, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			block, msg := DestructiveBlock(c.dir, c.command)
			if block != (c.want != "") {
				t.Fatalf("block = %v for %q, want %v", block, c.command, c.want != "")
			}
			want := ""
			if c.want != "" {
				want = fmt.Sprintf(destructiveWant, c.want)
			}
			if msg != want {
				t.Errorf("message = %q, want %q", msg, want)
			}
		})
	}
}
