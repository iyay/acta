package board

import (
	"path/filepath"
	"strings"

	"github.com/iyay/acta/internal/gitc"
)

// The two questions FillAuthors asks git, as variables, so a test can count
// them without a real repo: a board loads a lot and a plan with ten tasks is
// still one file.
var (
	gitAuthors  = gitc.Authors
	gitUserName = gitc.UserName
)

// FillAuthors gives every item on disk the name of the person whose commit
// first added the file it lives in, so a task takes the author of its plan
// file and a debt item the author of its debt file. A load leaves this out,
// since the log of the whole history is slow and only the detail box shows an
// author. Git is asked once per checkout, since once per file made every load
// take seconds. A file git has no commit for belongs to whoever commits in the
// main checkout now, and a folder outside git leaves the field empty, so the
// detail leaves the AUTHOR line out there. An item read from a branch is not
// on disk and gets none.
func (b *Board) FillAuthors() {
	if len(b.roots) == 0 {
		return
	}
	byRoot := map[string][]string{}
	seen := map[string]bool{}
	for _, it := range b.Items {
		if !it.OnDisk || it.Path == "" || seen[it.Path] {
			continue
		}
		seen[it.Path] = true
		root := checkoutOf(b.roots, it.Path)
		if root == "" {
			// No checkout of ours holds this folder, so ask where the file sits.
			root = filepath.Dir(it.Path)
		}
		byRoot[root] = append(byRoot[root], it.Path)
	}
	found := map[string]string{}
	for root, paths := range byRoot {
		for p, a := range gitAuthors(root, paths) {
			found[p] = a
		}
	}
	me, asked := "", false
	for _, it := range b.Items {
		if !it.OnDisk || it.Path == "" {
			continue
		}
		a, ok := found[it.Path]
		if !ok {
			if !asked {
				me, asked = gitUserName(b.roots[0]), true
			}
			a = me
		}
		it.Author = a
	}
}

// checkoutOf gives the checkout that holds path, or "" when none does. A
// worktree can sit inside the main folder, so the deepest match wins. Whole
// path parts match, so /r/a does not hold /r/ab/x.md.
func checkoutOf(roots []string, path string) string {
	best := ""
	for _, r := range roots {
		if r == "" || len(r) <= len(best) {
			continue
		}
		if strings.HasPrefix(path, strings.TrimSuffix(r, string(filepath.Separator))+string(filepath.Separator)) {
			best = r
		}
	}
	return best
}
