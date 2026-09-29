package tui

import (
	"sort"
	"strconv"
	"strings"

	"github.com/iyay/acta/internal/board"
)

// ordered gives items sorted by the number in their short id, oldest first,
// or newest first when newest is set. Ids are handed out in the order items
// are made, so this is the order they were made in. Two items with the same
// number sort by the date in their file name, then by the file they live in,
// so the list never jumps between loads. The sort is stable and every task
// of a plan lives in the same file, so tasks keep their file order. An item
// whose id holds no number goes last, and among those the date decides, the
// way it did before ids. It returns a new slice so the board keeps its own
// order.
func ordered(items []*board.Item, newest bool) []*board.Item {
	out := append([]*board.Item(nil), items...)
	sort.SliceStable(out, func(i, j int) bool {
		a, c := out[i], out[j]
		na, hasA := idNum(a)
		nc, hasC := idNum(c)
		if hasA != hasC {
			return hasA
		}
		if hasA && na != nc {
			return (na < nc) != newest
		}
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

// idNum reads the number of a short id: 17 from SCR-0017, and 33 from the
// task id PLN-0033.01, so a task sorts with its plan. An id with no dash and
// an id with no number both leave nothing to read, and it says false.
func idNum(it *board.Item) (int, bool) {
	_, rest, _ := strings.Cut(it.ShortID, "-")
	rest, _, _ = strings.Cut(rest, ".")
	n, err := strconv.Atoi(rest)
	return n, err == nil
}

// fileOf drops the "#task-N" end of a task ID, so a task sorts with its plan.
func fileOf(id string) string {
	file, _, _ := strings.Cut(id, "#")
	return file
}
