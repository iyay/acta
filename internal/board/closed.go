package board

import (
	"path/filepath"

	"github.com/iyay/acta/internal/gitc"
)

// The two questions fillAuthors asks git, as variables, so a test can count
// them without a real repo: a board loads a lot and a plan with ten tasks is
// still one file.
var (
	gitAuthors  = gitc.Authors
	gitUserName = gitc.UserName
)

// fillAuthors gives every item the name of the person whose commit first added
// the file it lives in, so a task takes the author of its plan file and a debt
// item the author of its debt file. Git is asked once per folder, since once
// per file made every load take seconds. A file git has no commit for belongs
// to whoever commits here now, and a folder outside git leaves the field
// empty, so the detail leaves the AUTHOR line out there.
func (b *Board) fillAuthors(root string) {
	byDir := map[string][]string{}
	seen := map[string]bool{}
	for _, it := range b.Items {
		if !it.OnDisk || it.Path == "" || seen[it.Path] {
			continue
		}
		seen[it.Path] = true
		dir := filepath.Dir(it.Path)
		byDir[dir] = append(byDir[dir], it.Path)
	}
	found := map[string]string{}
	for dir, paths := range byDir {
		for p, a := range gitAuthors(dir, paths) {
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
				me, asked = gitUserName(root), true
			}
			a = me
		}
		it.Author = a
	}
}
