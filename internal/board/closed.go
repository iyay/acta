package board

import (
	"path/filepath"

	"github.com/iyay/acta/internal/gitc"
)

// The two questions fillAuthors asks git, as variables, so a test can count
// them without a real repo: a board loads a lot and a plan with ten tasks is
// still one file.
var (
	gitAuthor   = gitc.Author
	gitUserName = gitc.UserName
)

// fillAuthors gives every item the name of the person whose commit first added
// the file it lives in, so a task takes the author of its plan file and a debt
// item the author of its debt file. A file git has no commit for belongs to
// whoever commits here now, and a folder outside git leaves the field empty,
// so the detail leaves the AUTHOR line out there.
func (b *Board) fillAuthors(root string) {
	byFile := map[string]string{}
	me, asked := "", false
	author := func(path string) string {
		if a, seen := byFile[path]; seen {
			return a
		}
		a, ok := gitAuthor(filepath.Dir(path), path)
		if !ok {
			if !asked {
				me, asked = gitUserName(root), true
			}
			a = me
		}
		byFile[path] = a
		return a
	}
	for _, it := range b.Items {
		if !it.OnDisk || it.Path == "" {
			continue
		}
		it.Author = author(it.Path)
	}
}
