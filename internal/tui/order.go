package tui

import (
	"sort"
	"strings"

	"github.com/iyay/acta/internal/board"
)

// ordered gives items sorted by the date in their file name, oldest first, or
// newest first when newest is set. Items on the same date sort by the file
// they live in, so the list never jumps between loads. The sort is stable and
// every task of a plan lives in the same file, so tasks keep their file order.
// An item with no date goes last either way, since there is nothing to place
// it by. It returns a new slice so the board keeps its own order.
func ordered(items []*board.Item, newest bool) []*board.Item {
	out := append([]*board.Item(nil), items...)
	sort.SliceStable(out, func(i, j int) bool {
		a, c := out[i], out[j]
		if (a.Date == "") != (c.Date == "") {
			return c.Date == ""
		}
		if a.Date != c.Date {
			return (a.Date < c.Date) != newest
		}
		fa, fc := fileOf(a.ID), fileOf(c.ID)
		if fa != fc {
			return (fa < fc) != newest
		}
		return false
	})
	return out
}

// fileOf drops the "#task-N" end of a task ID, so a task sorts with its plan.
func fileOf(id string) string {
	file, _, _ := strings.Cut(id, "#")
	return file
}
