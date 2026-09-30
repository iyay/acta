package board

import (
	"slices"
	"strings"
)

// Priorities are the levels a bug or a debt item can carry, most urgent
// first, so the TUI can sort by their place in this list.
var Priorities = []string{"high", "medium", "low"}

// ValidPriority says whether s is one of the levels.
func ValidPriority(s string) bool { return slices.Contains(Priorities, s) }

// SplitPriority reads a "(high) " tag at the start of a debt line. A line
// with no tag, or with a word in brackets that is not a level, comes back
// whole, because a note may really start with brackets.
func SplitPriority(text string) (level, rest string) {
	for _, p := range Priorities {
		if r, ok := strings.CutPrefix(text, "("+p+") "); ok {
			return p, r
		}
	}
	return "", text
}
