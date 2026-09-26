package board

import (
	"path/filepath"
	"time"

	"github.com/iyay/acta/internal/gitc"
)

// fillCommitTimes records when each finished file was last committed, so pane
// [2] can show the most recently touched work first. It asks git once for each
// file while the board loads, never while the view draws a row, and a file git
// has no commit for keeps the date in its name.
func (b *Board) fillCommitTimes() {
	at := map[string]int64{}
	for _, it := range b.Items {
		if !Closed(it.Status) || !it.OnDisk || it.Path == "" {
			continue
		}
		when, seen := at[it.Path]
		if !seen {
			when, _ = gitc.LastCommit(filepath.Dir(it.Path), it.Path)
			at[it.Path] = when
		}
		it.CommitAt = when
	}
}

// SortTime is when the item was last touched, in seconds since the epoch: the
// newest commit on its file, or the date in its file name when git has none.
// Zero means neither is known. Pane [2] shows the newest first.
func (it *Item) SortTime() int64 {
	if it.CommitAt > 0 {
		return it.CommitAt
	}
	at, err := time.Parse("2006-01-02", it.Date)
	if err != nil {
		return 0
	}
	return at.Unix()
}

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
