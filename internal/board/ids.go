package board

import (
	"fmt"
	"strconv"
	"strings"
)

// Prefix gives the short ID prefix of a kind. A plan file is always PLN,
// whether the board shows it as a plan or as something else.
func Prefix(k Kind, plan bool) string {
	switch {
	case plan || k == KindPlan:
		return "PLN"
	case k == KindBug:
		return "BUG"
	case k == KindDebt || k == KindDebtItem:
		return "DBT"
	case k == KindScratch:
		return "SCR"
	default:
		return "SPC"
	}
}

// OldPrefix gives the prefix a kind had before ids were cut to 3 letters.
// Old files and old notes still use it.
func OldPrefix(p string) string {
	switch p {
	case "PLN":
		return "PLAN"
	case "SPC":
		return "SPEC"
	case "DBT":
		return "DEBT"
	case "SCR":
		return "SCRATCH"
	}
	return p
}

// FormatID pads the number to 4 digits so every id in a list is one width.
func FormatID(prefix string, n int) string { return fmt.Sprintf("%s-%04d", prefix, n) }

// IsHash says if s is a hash: 7 characters, a lower-case letter first, then
// lower-case letters or digits. The letter first keeps it apart from numbers.
func IsHash(s string) bool { return isHashLen(s, 7) }

func isHashLen(s string, n int) bool {
	if len(s) != n || s[0] < 'a' || s[0] > 'z' {
		return false
	}
	for _, c := range s[1:] {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

// IsID says if id is a new style short ID for prefix: the prefix, a dash,
// then 4 digits. A wider number is allowed only without a leading zero, so
// number 10000 can still be written down.
func IsID(id, prefix string) bool {
	n, ok := strings.CutPrefix(id, prefix+"-")
	if !ok || len(n) < 4 || strings.Trim(n, "0123456789") != "" || n == "0000" {
		return false
	}
	return len(n) == 4 || n[0] != '0'
}

// isOldID says if id is written the old way: the old prefix and a number
// with no leading zero.
func isOldID(id, prefix string) bool {
	n, ok := strings.CutPrefix(id, OldPrefix(prefix)+"-")
	return ok && n != "" && strings.Trim(n, "0123456789") == "" && n[0] != '0'
}

// setIDs copies id and hash from frontmatter and shows both the new way, so
// an unmigrated file still lines up with the rest. A value in the wrong shape
// is a problem, not an ID, so a typo cannot steal another item's name. Calling
// it again replaces its own old verdict, so a plan first read as a story and
// then as a plan keeps only the plan reading.
func setIDs(it *Item, prefix string, doc Doc) {
	it.ShortID, it.Hash, it.OldForm = "", "", false
	kept := it.Problems[:0]
	for _, pr := range it.Problems {
		if !strings.HasPrefix(pr, "bad id ") && !strings.HasPrefix(pr, "bad hash ") &&
			!strings.HasSuffix(pr, " is wider than 4 digits") {
			kept = append(kept, pr)
		}
	}
	it.Problems = kept
	it.RawID, it.RawHash = field(doc.Front, "id"), field(doc.Front, "hash")
	if it.RawID != "" {
		switch {
		case IsID(it.RawID, prefix):
			it.ShortID = it.RawID
		case isOldID(it.RawID, prefix):
			n, _ := strconv.Atoi(strings.TrimPrefix(it.RawID, OldPrefix(prefix)+"-"))
			it.ShortID, it.OldForm = FormatID(prefix, n), true
		default:
			it.Problems = append(it.Problems, "bad id "+it.RawID)
		}
		if _, num, _ := strings.Cut(it.ShortID, "-"); len(num) > 4 {
			it.Problems = append(it.Problems, "id "+it.ShortID+" is wider than 4 digits")
		}
	}
	if it.RawHash != "" {
		switch {
		case IsHash(it.RawHash):
			it.Hash = prefix + "-" + it.RawHash
		case isHashLen(it.RawHash, 4):
			it.Hash, it.OldForm = prefix+"-"+it.RawHash, true
		default:
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

// pad2 shows a number as 2 digits so task and debt-item rows line up. A
// number that is not plain, like the fix task F1, stays as written.
func pad2(n string) string {
	v, err := strconv.Atoi(n)
	if err != nil {
		return n
	}
	return fmt.Sprintf("%02d", v)
}

// aliasTask gives a task its plan's IDs plus its own number. A number of 100
// or more stops the columns lining up, so the plan says so.
func (b *Board) aliasTask(t *Item, planID, planHash string, plan *Item) {
	n := pad2(t.TaskNum)
	// The id lives in the plan file, so a task takes its old form from there.
	t.OldForm = plan.OldForm
	if planID != "" {
		t.ShortID = planID + "." + n
	}
	if planHash != "" {
		t.Hash = planHash + "." + n
	}
	if v, err := strconv.Atoi(t.TaskNum); err == nil && v > 99 {
		plan.Problems = append(plan.Problems, "task "+t.TaskNum+" is wider than 2 digits")
	}
	b.aliasItem(t)
}

// aliasDebtItem gives a debt-item its file's IDs plus its own line number. A
// number of 100 or more stops the columns lining up, so the debt file says so.
func (b *Board) aliasDebtItem(it *Item, fileShort, fileHash string, num int, file *Item) {
	n := pad2(strconv.Itoa(num))
	// The id lives in the debt file, so an item takes its old form from there.
	it.OldForm = file.OldForm
	if fileShort != "" {
		it.ShortID = fileShort + "." + n
	}
	if fileHash != "" {
		it.Hash = fileHash + "." + n
	}
	if num > 99 {
		file.Problems = append(file.Problems, "item "+n+" is wider than 2 digits")
	}
	b.aliasItem(it)
}
