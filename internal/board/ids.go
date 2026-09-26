package board

import (
	"strconv"
	"strings"
)

// Prefix gives the short ID prefix of a kind. A plan file is always PLAN,
// whether the board shows it as a plan or as something else.
func Prefix(k Kind, plan bool) string {
	switch {
	case plan || k == KindPlan:
		return "PLAN"
	case k == KindBug:
		return "BUG"
	case k == KindDebt || k == KindDebtItem:
		return "DEBT"
	default:
		return "SPEC"
	}
}

// IsHash says if s is a hash: 4 characters, a lower-case letter first, then
// lower-case letters or digits. The letter first keeps it apart from numbers.
func IsHash(s string) bool {
	if len(s) != 4 || s[0] < 'a' || s[0] > 'z' {
		return false
	}
	for _, c := range s[1:] {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

// IsID says if id is a well formed short ID for prefix: the prefix, a dash,
// then a number that does not start with a zero. A tool uses it to tell a
// value it may fill in from one that is already written down.
func IsID(id, prefix string) bool {
	n, ok := strings.CutPrefix(id, prefix+"-")
	return ok && n != "" && strings.Trim(n, "0123456789") == "" && n[0] != '0'
}

// setIDs copies id and hash from frontmatter. A value in the wrong shape is
// a problem, not an ID, so a typo cannot steal another item's name. Calling
// it again replaces its own old verdict, so a plan first read as a story and
// then as a plan keeps only the plan reading.
func setIDs(it *Item, prefix string, doc Doc) {
	it.ShortID, it.Hash = "", ""
	kept := it.Problems[:0]
	for _, pr := range it.Problems {
		if !strings.HasPrefix(pr, "bad id ") && !strings.HasPrefix(pr, "bad hash ") {
			kept = append(kept, pr)
		}
	}
	it.Problems = kept
	it.RawID, it.RawHash = field(doc.Front, "id"), field(doc.Front, "hash")
	if it.RawID != "" {
		if IsID(it.RawID, prefix) {
			it.ShortID = it.RawID
		} else {
			it.Problems = append(it.Problems, "bad id "+it.RawID)
		}
	}
	if it.RawHash != "" {
		if IsHash(it.RawHash) {
			it.Hash = prefix + "-" + it.RawHash
		} else {
			it.Problems = append(it.Problems, "bad hash "+it.RawHash)
		}
	}
}

// addAlias maps a short ID to an item. Two items with one short ID both get a
// problem line, and the first one keeps the name.
func (b *Board) addAlias(key string, it *Item, what string) {
	if key == "" {
		return
	}
	k := strings.ToLower(key)
	if old := b.alias[k]; old != nil && old != it {
		old.Problems = append(old.Problems, "duplicate "+what+" "+key)
		it.Problems = append(it.Problems, "duplicate "+what+" "+key)
		return
	}
	b.alias[k] = it
}

func (b *Board) aliasItem(it *Item) {
	b.addAlias(it.ShortID, it, "id")
	b.addAlias(it.Hash, it, "hash")
}

// aliasTask gives a task its plan's IDs plus its own number.
func (b *Board) aliasTask(t *Item, planID, planHash string) {
	if planID != "" {
		t.ShortID = planID + "." + t.TaskNum
	}
	if planHash != "" {
		t.Hash = planHash + "." + t.TaskNum
	}
	b.aliasItem(t)
}

// aliasDebtItem gives a debt-item its file's IDs plus its own line number.
func (b *Board) aliasDebtItem(it *Item, fileShort, fileHash string, num int) {
	n := strconv.Itoa(num)
	if fileShort != "" {
		it.ShortID = fileShort + "." + n
	}
	if fileHash != "" {
		it.Hash = fileHash + "." + n
	}
	b.aliasItem(it)
}
