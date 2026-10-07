// Package tidy folds a branch into one commit per code task and proves the
// result holds the same files.
package tidy

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Options says what to fold. PlanningRoot is relative to the repo, like ".acta".
type Options struct {
	Base, Branch, Onto, PlanningRoot string
}

// Result reports the new chain. Tip is the new last commit. Parent is the tip
// of the base branch that tidy read. Folded is how many chore commits of that
// parent were folded into the new chain.
type Result struct {
	OldCount, NewCount int
	Tip                string
	Parent             string
	Folded             int
}

// tamperTree lets a test hand back a wrong final tree. It stays nil otherwise.
var tamperTree func(tree string) string

// foldTypes are subject types that never stand as a commit of their own.
var foldTypes = map[string]bool{"chore": true, "docs": true, "polish": true, "wiki": true}

type commit struct {
	hash, subject string
	date          time.Time
}

// node is one new commit in the making.
type node struct {
	tree string
	src  commit
	date time.Time
	olds []string
}

// Run builds the tidy chain and writes refs/acta/tidy/<branch> when the proof holds.
func Run(repo string, opt Options) (Result, error) {
	if out, err := git(repo, nil, "", "version"); err != nil {
		return Result{}, err
	} else if err := checkVersion(out); err != nil {
		return Result{}, err
	}
	root := strings.Trim(opt.PlanningRoot, "/")
	if root == "" {
		root = ".acta"
	}
	short := strings.TrimPrefix(opt.Branch, "refs/heads/")
	// The name goes into a ref path, so check it before anything is built.
	if _, err := git(repo, nil, "", "check-ref-format", "--branch", short); err != nil {
		return Result{}, fmt.Errorf("tidy: %q is not a valid branch name", short)
	}
	base, err := resolve(repo, opt.Base)
	if err != nil {
		return Result{}, err
	}
	branchTip, err := resolve(repo, opt.Branch)
	if err != nil {
		return Result{}, err
	}
	// The branch may have been cut before the base moved. Walk only the
	// branch's own commits, and aim at the merge of base and branch so the
	// base's newer work is kept.
	fork, err := git(repo, nil, "", "merge-base", base, branchTip)
	if err != nil {
		return Result{}, err
	}
	fork = strings.TrimSpace(fork)
	target, clashPaths, clash, err := mergeTree(repo, fork, base, branchTip)
	if err != nil {
		return Result{}, err
	}
	if clash {
		return Result{}, fmt.Errorf("tidy: %s clashes with %s at %s", opt.Branch, opt.Base, strings.Join(clashPaths, ", "))
	}
	commits, err := walk(repo, fork+".."+branchTip)
	if err != nil {
		return Result{}, err
	}
	if len(commits) == 0 {
		return Result{}, fmt.Errorf("tidy: %s has no commits on top of %s", opt.Branch, opt.Base)
	}

	// Start point: the base, or the fold point when only chores sit between.
	parent, cur := base, ""
	var carried []commit
	if opt.Onto != "" {
		if parent, carried, err = ontoStart(repo, opt.Onto, base); err != nil {
			return Result{}, err
		}
	}
	if cur, err = resolve(repo, parent+"^{tree}"); err != nil {
		return Result{}, err
	}
	var pendDate time.Time
	for _, c := range carried {
		t, clash, err := replay(repo, c.hash, cur)
		if err != nil {
			return Result{}, err
		}
		if !clash {
			cur = t
		}
		pendDate = later(pendDate, c.date)
	}

	nodes, err := plan(repo, commits, root, cur, pendDate)
	if err != nil {
		return Result{}, err
	}
	nodes[len(nodes)-1].tree = target
	keepTimeForward(nodes)

	// Build every commit but the last, then the map from old to new hashes.
	last := len(nodes) - 1
	newOf := map[string]string{}
	for i := 0; i < last; i++ {
		h, err := commitNode(repo, nodes[i], parent)
		if err != nil {
			return Result{}, err
		}
		parent = h
		for _, o := range nodes[i].olds {
			newOf[o] = h
		}
	}
	// The last commit's hash depends on its tree, so the tree cannot name it.
	// Old hashes that land in the last commit stay as they were.
	olds := make([]string, len(commits))
	for i, c := range commits {
		olds[i] = c.hash
	}
	finalTree, _, err := remapTree(repo, target, root, olds, newOf)
	if err != nil {
		return Result{}, err
	}
	if tamperTree != nil {
		finalTree = tamperTree(finalTree)
	}
	if err := proof(repo, target, finalTree, root, olds, newOf); err != nil {
		return Result{}, err
	}
	nodes[last].tree = finalTree
	tip, err := commitNode(repo, nodes[last], parent)
	if err != nil {
		return Result{}, err
	}
	if _, err := git(repo, nil, "", "update-ref", "refs/acta/tidy/"+short, tip); err != nil {
		return Result{}, err
	}
	return Result{OldCount: len(commits), NewCount: len(nodes), Tip: tip, Parent: base, Folded: len(carried)}, nil
}

// plan walks the branch commits oldest first and decides which ones stand
// alone. cur is the tree built so far.
func plan(repo string, commits []commit, root, cur string, pendDate time.Time) ([]*node, error) {
	var nodes []*node
	var pendOlds []string
	reviewed := false
	for _, c := range commits {
		if isReviewMarker(c.subject) {
			reviewed = true
		}
		next, clash, err := replay(repo, c.hash, cur)
		if err != nil {
			return nil, err
		}
		ok := !clash
		if ok {
			cur = next
		}
		keep := false
		if ok && !reviewed && !foldTypes[subjectType(c.subject)] {
			var err error
			if keep, err = touchesOutside(repo, c.hash, root); err != nil {
				return nil, err
			}
		}
		switch {
		case keep:
			nodes = append(nodes, &node{tree: cur, src: c, date: later(c.date, pendDate), olds: append(pendOlds, c.hash)})
			pendOlds, pendDate = nil, time.Time{}
		case len(nodes) > 0:
			n := nodes[len(nodes)-1]
			n.tree, n.date, n.olds = cur, later(n.date, c.date), append(n.olds, c.hash)
		default:
			// No kept commit yet: ride forward into the next one.
			pendOlds, pendDate = append(pendOlds, c.hash), later(pendDate, c.date)
		}
	}
	if len(nodes) == 0 {
		// Nothing was worth keeping. The whole range becomes one commit.
		lastC := commits[len(commits)-1]
		nodes = append(nodes, &node{tree: cur, src: lastC, date: pendDate, olds: pendOlds})
	}
	return nodes, nil
}

func later(a, b time.Time) time.Time {
	if b.After(a) {
		return b
	}
	return a
}

// keepTimeForward stops a date from going back along the chain.
func keepTimeForward(nodes []*node) {
	for i := 1; i < len(nodes); i++ {
		nodes[i].date = later(nodes[i].date, nodes[i-1].date)
	}
}

// subjectType is the word before "(" or ":" at the start of a subject.
func subjectType(subject string) string {
	end := len(subject)
	for _, sep := range []string{"(", ":", " "} {
		if i := strings.Index(subject, sep); i >= 0 && i < end {
			end = i
		}
	}
	return subject[:end]
}

// isReviewMarker spots the commit that ends the code tasks.
func isReviewMarker(subject string) bool {
	return strings.HasPrefix(subject, "chore(plan): review notes") || strings.HasPrefix(subject, "polish")
}

// touchesOutside says whether a commit changes a file outside the planning root.
func touchesOutside(repo, hash, root string) (bool, error) {
	out, err := git(repo, nil, "", "diff-tree", "--no-commit-id", "-r", "--name-only", "-z", hash)
	if err != nil {
		return false, err
	}
	for _, f := range strings.Split(out, "\x00") {
		if f != "" && !strings.HasPrefix(f, root+"/") {
			return true, nil
		}
	}
	return false, nil
}

// replay applies the change of one commit on top of a tree. It returns the new
// tree, then clash (true when the change does not fit), then err for any other
// git failure.
func replay(repo, hash, cur string) (string, bool, error) {
	tree, _, clash, err := mergeTree(repo, hash+"^", cur, hash)
	return tree, clash, err
}

// mergeTree is a three-way merge that writes no files. Exit code 1 is the one
// code git uses for a clash, and then paths names the clashing files. Any
// other failure comes back as an error and is never read as a clash.
func mergeTree(repo, mergeBase, ours, theirs string) (tree string, paths []string, clash bool, err error) {
	out, err := git(repo, nil, "", "merge-tree", "--write-tree", "--name-only", "--merge-base="+mergeBase, ours, theirs)
	lines := strings.Split(out, "\n")
	if err == nil {
		return strings.TrimSpace(lines[0]), nil, false, nil
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		return "", nil, false, err
	}
	// Line one is the tree. The clashing paths follow, up to a blank line.
	for _, l := range lines[1:] {
		if l == "" {
			break
		}
		paths = append(paths, l)
	}
	return "", paths, true, nil
}

// ontoStart returns the commit to build on, and the chore commits to fold
// forward. Any commit between onto and base that is not a chore cancels it,
// and then the chain is built on base. An unknown ref or a failed git call is
// an error.
func ontoStart(repo, onto, base string) (string, []commit, error) {
	o, err := resolve(repo, onto)
	if err != nil {
		return "", nil, err
	}
	if _, err := git(repo, nil, "", "merge-base", "--is-ancestor", o, base); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 1 {
			// Onto is not behind base, so there is nothing to fold.
			return base, nil, nil
		}
		return "", nil, err
	}
	chores, err := walk(repo, o+".."+base)
	if err != nil {
		return "", nil, err
	}
	for _, c := range chores {
		if !strings.HasPrefix(c.subject, "chore(") {
			return base, nil, nil
		}
	}
	// A merge in between would be skipped by the walk, so refuse that too.
	all, err := git(repo, nil, "", "rev-list", "--count", o+".."+base)
	if err != nil {
		return "", nil, err
	}
	if strings.TrimSpace(all) != strconv.Itoa(len(chores)) {
		return base, nil, nil
	}
	return o, chores, nil
}

// walk lists the non-merge commits of a range, oldest first, first parent only.
func walk(repo, rng string) ([]commit, error) {
	out, err := git(repo, nil, "", "rev-list", "--reverse", "--first-parent", "--no-merges", "--format=%H%x1f%aI%x1f%s", rng)
	if err != nil {
		return nil, err
	}
	var list []commit
	for _, line := range strings.Split(out, "\n") {
		f := strings.SplitN(line, "\x1f", 3)
		if len(f) != 3 {
			continue
		}
		t, err := time.Parse(time.RFC3339, f[1])
		if err != nil {
			return nil, err
		}
		list = append(list, commit{hash: f[0], date: t, subject: f[2]})
	}
	return list, nil
}

// commitNode makes the commit with the identity of the kept commit.
func commitNode(repo string, n *node, parent string) (string, error) {
	f, err := git(repo, nil, "", "log", "-1", "--format=%an%x1f%ae%x1f%cn%x1f%ce", n.src.hash)
	if err != nil {
		return "", err
	}
	id := strings.Split(strings.TrimRight(f, "\n"), "\x1f")
	if len(id) != 4 {
		return "", fmt.Errorf("tidy: cannot read identity of %s", n.src.hash)
	}
	msg, err := git(repo, nil, "", "log", "-1", "--format=%B", n.src.hash)
	if err != nil {
		return "", err
	}
	date := n.date.Format(time.RFC3339)
	env := []string{
		"GIT_AUTHOR_NAME=" + id[0], "GIT_AUTHOR_EMAIL=" + id[1], "GIT_AUTHOR_DATE=" + date,
		"GIT_COMMITTER_NAME=" + id[2], "GIT_COMMITTER_EMAIL=" + id[3], "GIT_COMMITTER_DATE=" + date,
	}
	out, err := git(repo, env, strings.TrimSuffix(msg, "\n"), "commit-tree", n.tree, "-p", parent)
	return strings.TrimSpace(out), err
}

// proof checks that the final tree is the three-way merge of base and branch
// (target) plus hash swaps in the planning root, and nothing else.
func proof(repo, target, finalTree, root string, olds []string, newOf map[string]string) error {
	if target == finalTree {
		return nil
	}
	out, err := git(repo, nil, "", "diff-tree", "-r", "-z", "--raw", target, finalTree)
	if err != nil {
		return err
	}
	parts := strings.Split(out, "\x00")
	var paths, oldShas, newShas []string
	for i := 0; i+1 < len(parts); i += 2 {
		meta := strings.Fields(strings.TrimPrefix(parts[i], ":"))
		path := parts[i+1]
		if len(meta) != 5 || meta[4] != "M" || !strings.HasPrefix(path, root+"/") {
			return fmt.Errorf("tidy: final tree differs from the merge of base and branch at %s (%v)", path, meta)
		}
		paths, oldShas, newShas = append(paths, path), append(oldShas, meta[2]), append(newShas, meta[3])
	}
	if len(paths) == 0 {
		return nil
	}
	blobs, err := readBlobs(repo, append(append([]string{}, oldShas...), newShas...))
	if err != nil {
		return err
	}
	for i, p := range paths {
		want, changed := remapBlob(blobs[oldShas[i]], olds, newOf)
		if !changed || string(want) != string(blobs[newShas[i]]) {
			return fmt.Errorf("tidy: %s differs by more than hash swaps", p)
		}
	}
	return nil
}

var versionRe = regexp.MustCompile(`(\d+)\.(\d+)(?:\.\d+)?`)

// checkVersion refuses a git older than 2.40, which lacks merge-tree --write-tree.
func checkVersion(out string) error {
	m := versionRe.FindStringSubmatch(out)
	if m == nil {
		return fmt.Errorf("tidy: cannot read the git version from %q", strings.TrimSpace(out))
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	if major < 2 || (major == 2 && minor < 40) {
		v := strings.TrimPrefix(strings.TrimSpace(out), "git version ")
		return fmt.Errorf("tidy: git %s is too old, need 2.40 or newer", v)
	}
	return nil
}

func resolve(repo, rev string) (string, error) {
	out, err := git(repo, nil, "", "rev-parse", "--verify", "-q", "--end-of-options", rev)
	if err != nil {
		return "", fmt.Errorf("tidy: unknown revision %q", rev)
	}
	return strings.TrimSpace(out), nil
}

// git runs the real git binary in repo and returns its stdout.
func git(repo string, env []string, stdin string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = strings.NewReader(stdin)
	var out, errOut strings.Builder
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return out.String(), nil
}

// tempIndex returns a path for a throwaway index file.
func tempIndex() (string, error) {
	dir, err := os.MkdirTemp("", "acta-tidy-")
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "index"), nil
}

func removeIndex(idx string) {
	// Best effort. A temp folder left behind costs nothing and cannot change
	// the result, so a failure here is ignored on purpose.
	_ = os.RemoveAll(filepath.Dir(idx))
}
