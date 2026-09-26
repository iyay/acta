package board

import (
	"path/filepath"
	"time"

	"pm-board/internal/gitc"
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
